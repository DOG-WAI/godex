package impl

import (
	"context"
	"github.com/jinzhu/copier"
	"godex/internal/errors"
	"godex/internal/logic"
	"godex/internal/service"
	"godex/pkg/api"
	"godex/pkg/errs"
)

var PhishingSitesLogic logic.PhishingSitesLogic = &phishingSitesLogic{}

type phishingSitesLogic struct {
	/* dependencies */
}

// 编译时检查接口实现
var _ logic.PhishingSitesLogic = (*phishingSitesLogic)(nil)

// CheckSites 检查网站是否为
func (c *phishingSitesLogic) CheckSites(ctx context.Context, req api.CheckSitesReq) (api.CheckSitesRsp, error) {
	phishingSites, err := service.NewPhishingSitesService().CheckPhishingSitesWithCache(ctx, req)
	if err != nil {
		return nil, errs.Newf(errors.InternalError, "check phishing sites failed")
	}

	var rsp api.CheckSitesRsp
	if err = copier.Copy(&rsp, &phishingSites); err != nil {
		return nil, errs.Newf(errors.InternalError, "copy response data failed: %v", err)
	}

	return rsp, nil
}
