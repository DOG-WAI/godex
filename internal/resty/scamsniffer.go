package resty

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"godex/internal/conf"
	"godex/internal/errors"
	"godex/pkg/errs"
	"godex/pkg/logger"
	"time"
)

var ScamsnifferResty = NewScamsnifferResty()

type scamsniffer struct {
	client *resty.Client
}

func NewScamsnifferResty() *scamsniffer {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	return &scamsniffer{client: client}
}

// FetchDomains 读取
func (r *scamsniffer) FetchDomains() ([]string, error) {
	// 获取配置的URL
	url := conf.AppConfig.AppSetting.ScamSniffer

	// 发送HTTP请求获取数据
	resp, err := r.client.R().Get(url)
	if err != nil {
		return nil, errs.Newf(errors.InternalError, "failed to fetch data from %s: %v", url, err)
	}

	// 检查HTTP状态码
	if resp.StatusCode() != 200 {
		return nil, errs.Newf(errors.InternalError, "HTTP request failed with status: %d", resp.StatusCode())
	}
	logger.Infof("Successfully fetched data, response size: %d bytes", len(resp.Body()))

	// 解析JSON数组
	var domains []string
	if err := json.Unmarshal(resp.Body(), &domains); err != nil {
		return nil, errs.Newf(errors.InternalError, "failed to parse JSON: %v", err)
	}
	logger.Infof("Parsed %d domains from JSON", len(domains))
	return domains, nil
}
