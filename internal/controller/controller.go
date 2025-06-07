package controller

import (
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	recovermw "github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/middleware/requestid"
	"godex/internal/errors"
	"godex/internal/logic/impl"
	"godex/internal/middleware"
	"godex/pkg/api"
	"godex/pkg/errs"
)

// Routing ...
func Routing(app *iris.Application) {
	// 1. 全局中间件
	{
		app.UseRouter(recovermw.New())               // panic保护
		app.UseRouter(middleware.LoggerMiddleware()) // 自定义日志中间件
		app.AllowMethods(iris.MethodOptions)         // OPTIONS预检
		app.UseRouter(requestid.New())               // 自动生成/传递 X-Request-ID
		app.UseRouter(cors.AllowAll())               // 开启跨域
		app.Use(iris.Compression)                    // 启用数据压缩
	}

	// 2. 健康检查(仅用于观察服务是否在运行)
	{
		app.Get("/health", func(ctx iris.Context) {
			ctx.JSON(iris.Map{"status": "ok"})
		})
		app.Get("/working", api.Handler[api.WorkingReq, api.WorkingRsp](impl.WorkingLogic.Working))
	}

	// 3. 全局错误捕获
	{
		app.OnErrorCode(iris.StatusInternalServerError, func(ctx iris.Context) {
			// err := ctx.GetErr() // 生产环境下禁止输出这个err
			api.Error(ctx, errs.Newf(errors.InternalError, "Internal Server Error"))
		})
		app.OnErrorCode(iris.StatusNotFound, func(ctx iris.Context) {
			api.Error(ctx, errs.Newf(errors.CallFail, "Not Found"))
		})
	}

	// 4. 业务路由
	{
		browserextAPI := app.Party("/browserext")
		phishingSitesAPI := browserextAPI.Party("/phishing_sites")
		phishingSitesAPI.Post("/check", api.Handler[api.CheckSitesReq, api.CheckSitesRsp](impl.PhishingSitesLogic.CheckSites))
	}
}
