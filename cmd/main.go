package main

import (
	"godex/internal/serve"
	"godex/pkg/logger"
	"os"
)

func main() {
	var opt serve.Option

	// 默认启用Web服务模式
	opt = serve.WithWebDefault()
	if len(os.Args) > 1 {
		// 有命令行参数：启用命令执行模式
		opt = serve.WithCommandDefault()
	}

	// 统一通过serve.Run()运行
	if err := serve.NewServe(opt).Run(); err != nil {
		logger.Fatalf("Application run failed: %v", err)
	}
}
