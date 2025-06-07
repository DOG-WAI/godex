package service

import (
	"context"
	"encoding/json"
	"fmt"
	"godex/internal/cache"
	"godex/internal/conf"
	"godex/internal/entity"
	"godex/internal/resty"
	"godex/pkg/logger"
	"godex/pkg/report"
	"strings"
	"time"
)

const PhishingSitesSourceScamSniffer = "scam-sniffer"
const PhishingSitesSourceFixedSniffer = "fixed-sniffer"

// PhishingSitesService 服务
type PhishingSitesService struct {
	reporter    *report.Reporter
	ossStoreSvc *OssStoresService
}

// NewPhishingSitesService 创建服务实例
func NewPhishingSitesService() *PhishingSitesService {
	return &PhishingSitesService{
		reporter:    report.NewReporter(conf.AppConfig.System.Report),
		ossStoreSvc: NewOssStoresService(),
	}
}

// LoadPhishingSites2Cache 加载到cache
func (s *PhishingSitesService) LoadPhishingSites2Cache(ctx context.Context) error {
	logger.Info("开始加载数据到内存")

	// 1. 先加载固定配置中的
	fixedCount := 0
	if conf.AppConfig.AppSetting.FixedSniffer != nil {
		for _, domain := range conf.AppConfig.AppSetting.FixedSniffer {
			// 将域名标准化：转小写，去空格
			domainStd := strings.ToLower(strings.TrimSpace(domain))
			if domainStd == "" {
				continue
			}

			cacheItem := &entity.PhishingSite{
				Domain: domainStd,
				Source: PhishingSitesSourceFixedSniffer,
			}
			cache.PhishingSitesCache.Store(domainStd, cacheItem)
			fixedCount++
		}
		logger.Infof("Successfully loaded %d fixed phishing sites from config to cache", fixedCount)
	}

	// 2. 再从OSS加载数据
	download, err := s.ossStoreSvc.Download(ctx, fmt.Sprintf("%s-domains.json", PhishingSitesSourceScamSniffer))
	if err != nil {
		logger.Errorf("Download PhishingSites failed: %v", err)
		return err
	}

	sites := []string{}
	err = json.Unmarshal([]byte(download), &sites)

	// 将数据库中的sites写入到缓存 - sync.Map是线程安全的
	for _, site := range sites {
		cacheItem := &entity.PhishingSite{
			Domain: site,
			Source: PhishingSitesSourceScamSniffer,
		}
		// 使用域名作为key，sync.Map的Store方法是线程安全的
		cache.PhishingSitesCache.Store(site, cacheItem)
	}

	ossCount := len(sites)
	logger.Infof("Successfully loaded %d phishing sites from oss to cache", ossCount)

	logger.Infof("Total loaded %d phishing sites to cache (fixed-config: %d, database: %d)",
		fixedCount+ossCount, fixedCount, ossCount)
	return nil
}

// CheckPhishingSitesWithCache 检查是否为
func (s *PhishingSitesService) CheckPhishingSitesWithCache(ctx context.Context, sites []string) ([]*entity.PhishingSiteCheckRet, error) {
	phishingSitesRet := []*entity.PhishingSiteCheckRet{}

	for _, site := range sites {
		// 1. 将site转为小写并去除空格
		siteStd := strings.ToLower(strings.TrimSpace(site))

		// 2. 检查原始值是否存在于cache中
		if val, exists := cache.PhishingSitesCache.Load(siteStd); exists {
			phishingSite := val.(*entity.PhishingSite)
			phishingSitesRet = append(phishingSitesRet, &entity.PhishingSiteCheckRet{
				Query:  site,
				Domain: phishingSite.Domain,
				Source: phishingSite.Source,
			})
			continue // 如果原始值已找到，跳过后续检查避免重复
		}

		// 3. 如果site本身不带www，检查添加www.前缀的值是否存在于cache中
		if !strings.HasPrefix(siteStd, "www.") {
			wwwSite := "www." + siteStd
			if val, exists := cache.PhishingSitesCache.Load(wwwSite); exists {
				phishingSite := val.(*entity.PhishingSite)
				phishingSitesRet = append(phishingSitesRet, &entity.PhishingSiteCheckRet{
					Query:  site,
					Domain: phishingSite.Domain,
					Source: phishingSite.Source,
				})
			}
		} else {
			// 4. 如果site本身带有www，检查去掉www.前缀的值是否存在于cache中
			nonWwwSite := strings.TrimPrefix(siteStd, "www.")
			if val, exists := cache.PhishingSitesCache.Load(nonWwwSite); exists {
				phishingSite := val.(*entity.PhishingSite)
				phishingSitesRet = append(phishingSitesRet, &entity.PhishingSiteCheckRet{
					Query:  site,
					Domain: phishingSite.Domain,
					Source: phishingSite.Source,
				})
			}
		}
	}

	// 上报名中的到webbb平台
	s.ReportWithPhishingSiteCheckRet(phishingSitesRet)
	return phishingSitesRet, nil
}

// ReportWithPhishingSiteCheckRet 上报名中的到webbb平台
func (s *PhishingSitesService) ReportWithPhishingSiteCheckRet(ret []*entity.PhishingSiteCheckRet) {
	if conf.AppConfig.System.Report.Enable && len(ret) > 0 {
		// 构建上报数据
		nowMillis := time.Now().UnixMilli()
		reportPayload := report.ReportPayload{}

		for _, result := range ret {
			reportItem := report.ReportPayloadItem{
				OpRes: report.OpResOK, OpObjType: report.OpObjTypePhishing, // 31表示检测类型
				OpObjValue: map[string]interface{}{"url": result.Query, "source": result.Source}, UserTimestamp: nowMillis, Timestamp: nowMillis,
			}
			reportPayload = append(reportPayload, reportItem)
		}
		// 发送上报
		s.reporter.SyncSend("/conf", report.ReportHead{UserId: 0}, reportPayload)
	}
}

// ImportPhishingSites 导入
func (s *PhishingSitesService) ImportPhishingSites(ctx context.Context) error {
	domains, err := resty.ScamsnifferResty.FetchDomains()
	if err != nil {
		logger.Errorf("fetch scamsniffer failed: %v", err)
		return err
	}

	marshal, err := json.Marshal(domains)
	if err != nil {
		logger.Errorf("marshal scamsniffer failed: %v", err)
		return err
	}

	err = s.ossStoreSvc.Upload(ctx, fmt.Sprintf("%s-domains.json", PhishingSitesSourceScamSniffer), string(marshal))
	if err != nil {
		logger.Errorf("upload scam-sniffer failed: %v", err)
		return err
	}

	logger.Infof("Successfully uploaded scamsniffer to config")
	return nil
}
