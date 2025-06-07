package api

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"godex/pkg/errs"
	"godex/pkg/logger"
)

// OK 返回成功响应
func OK(ctx iris.Context, data interface{}, message string) {
	if message == "" {
		message = errs.Success
	}
	resp := APIResponse{
		Code:    0,
		Message: message,
		Data:    data,
		TraceId: ctx.GetID(),
	}
	logger.IgnoreError(ctx.JSON(resp))
}

// Error 返回错误响应
func Error(ctx iris.Context, err error) {
	resp := APIResponse{
		Code:    errs.Code(err),
		Message: errs.Msg(err),
		TraceId: ctx.GetID(),
	}
	if err != nil {
		resp.Message = fmt.Sprintf("%+v(%+v)", resp.Message, err.Error())
	}
	ctx.StatusCode(400)
	logger.IgnoreError(ctx.JSON(resp))
}

// Handler 泛型处理器，自动处理请求绑定、错误处理和响应序列化
// TReq: 请求类型, TRsp: 响应类型
// 会尝试读取请求体JSON，如果失败或无数据则使用零值
func Handler[TReq any, TRsp any](handler func(ctx context.Context, req TReq) (TRsp, error)) iris.Handler {
	return func(ctx iris.Context) {
		var req TReq
		var rsp TRsp

		// 尝试自动绑定请求参数，失败时使用零值
		// 这样既支持POST/PUT等带请求体的请求，也支持GET等不带请求体的请求
		if err := ctx.ReadJSON(&req); err != nil {
			// 如果读取JSON失败，可能是GET请求或空请求体，使用零值
			// 对于GET请求或其他不需要请求体的情况，这是正常行为
			Error(ctx, errs.NewFrameError(errs.RetClientEncodeFail, err.Error()))
			logger.Errorf("req parse json fail, err: %+v", err)
			return
		}

		// 调用业务处理函数
		rsp, err := handler(ctx, req)
		if err != nil {
			Error(ctx, err)
			logger.Errorf("handler error: %+v", err)
			return
		}

		// 自动序列化响应
		OK(ctx, rsp, "")
	}
}

// ToIrisContext 安全地将context.Context转换为iris.Context
// 如果ctx本身就是iris.Context或者包含iris.Context，则返回它
// 否则返回nil和false
func ToIrisContext(ctx context.Context) (iris.Context, bool) {
	if ctx == nil {
		return nil, false
	}

	// 直接类型断言检查是否为iris.Context
	if irisCtx, ok := ctx.(iris.Context); ok {
		return irisCtx, true
	}

	// 检查是否为包装了iris.Context的context
	// 尝试从context的Value中获取iris.Context
	if irisCtx := ctx.Value("iris.context"); irisCtx != nil {
		if ic, ok := irisCtx.(iris.Context); ok {
			return ic, true
		}
	}

	return nil, false
}
