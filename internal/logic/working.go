package logic

import (
	"context"
	"godex/pkg/api"
)

// WorkingLogic 服务状态检查逻辑层接口
type WorkingLogic interface {
	// Working 检查服务状态
	Working(ctx context.Context, req api.WorkingReq) (api.WorkingRsp, error)
}
