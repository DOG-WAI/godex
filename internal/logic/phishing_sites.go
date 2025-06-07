package logic

import (
	"context"
	"godex/pkg/api"
)

// PhishingSitesLogic 逻辑层接口
type PhishingSitesLogic interface {
	// CheckSites 检查网站是否为
	CheckSites(ctx context.Context, req api.CheckSitesReq) (api.CheckSitesRsp, error)
}
