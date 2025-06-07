package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"godex/pkg/constant"
	"godex/pkg/logger"
	"io"
	"math"
	"time"
)

// LoggerMiddleware API请求日志中间件
func LoggerMiddleware() iris.Handler {
	return func(ctx iris.Context) {
		path := ctx.Request().URL.Path
		query := ctx.Request().URL.RawQuery

		// 读取请求体
		var requestBody []byte
		if ctx.Request().Body != nil {
			requestBody, _ = io.ReadAll(ctx.Request().Body)
			ctx.ResetRequest(ctx.Request().Clone(context.Background()))
			ctx.Request().Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 处理请求
		start := time.Now()
		ctx.Next()
		latency := float64(time.Since(start).Nanoseconds()) / float64(time.Second.Nanoseconds())
		latency = math.Round(latency*10000) / 10000

		// 记录请求头
		requestHeaders := make(map[string]string)
		for k, v := range ctx.Request().Header {
			// 排除大头
			if k == "Cookie" {
				continue
			}
			if len(v) > 0 {
				requestHeaders[k] = v[0]
			}
		}
		requestHeadersStr, _ := json.Marshal(requestHeaders)

		jsonData, _ := json.Marshal(logrus.Fields{
			constant.TraceIDKey:       ctx.GetID(),
			constant.IPKey:            getRealIP(ctx),
			constant.HostNameKey:      ctx.Host(),
			constant.UserIDKey:        0,
			constant.MethodKey:        ctx.Request().Method,
			constant.PathKey:          path,
			constant.QueryKey:         query,
			constant.RequestKey:       string(requestBody),
			constant.RequestHeaderKey: string(requestHeadersStr),
			constant.UserAgentKey:     ctx.Request().UserAgent(),
			constant.StatusKey:        ctx.GetStatusCode(),
			constant.LatencyKey:       latency,
		})
		logger.Info("Request metadata: " + string(jsonData))
	}
}

// 获取真实IP，优先取 X-Forwarded-For/X-Real-IPKey
func getRealIP(ctx iris.Context) string {
	ip := ctx.GetHeader("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For 可能有多个IP，取第一个
		ips := bytes.Split([]byte(ip), []byte(","))
		if len(ips) > 0 {
			return string(bytes.TrimSpace(ips[0]))
		}
	}
	ip = ctx.GetHeader("X-Real-IPKey")
	if ip != "" {
		return ip
	}
	return ctx.RemoteAddr()
}
