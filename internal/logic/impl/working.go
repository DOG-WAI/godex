package impl

import (
	"context"
	"github.com/kataras/iris/v12"
	"godex/internal/logic"
	"godex/pkg/api"
	"time"
)

var WorkingLogic logic.WorkingLogic = &workingLogic{}

type workingLogic struct {
	/* dependencies */
}

// 编译时检查接口实现
var _ logic.WorkingLogic = (*workingLogic)(nil)

// Working 检查服务状态
func (c *workingLogic) Working(ctx context.Context, req api.WorkingReq) (api.WorkingRsp, error) {
	irisContext, b := api.ToIrisContext(ctx)
	return iris.Map{"req": req, "date": time.Now().Unix(), "isIrisCtx": b, "ctx": irisContext}, nil
}
