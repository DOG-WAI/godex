package serve

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"godex/internal/command"
	"godex/internal/conf"
	"godex/internal/controller"
	"godex/internal/errors"
	"godex/internal/service"
	"godex/internal/task"
	"godex/pkg/errs"
	"godex/pkg/logger"
)

// Serve 服务器结构体
type Serve struct {
	app     *iris.Application
	options *ServeOptions
}

// ServeOptions 服务器配置选项
type ServeOptions struct {
	// 初始化选项
	enableConfig     bool
	enableLogger     bool
	enableController bool
	enableWebServer  bool
	enableCommand    bool
	enableTask       bool

	// 配置选项
	configPaths []string
	port        int

	// 自定义初始化函数
	customInitFuncs []func() error
}

// Option 服务器配置选项函数类型
type Option func(*ServeOptions)

// WithConfig 启用配置初始化
func WithConfig() Option {
	return func(opts *ServeOptions) {
		opts.enableConfig = true
	}
}

// WithLogger 启用日志
func WithLogger() Option {
	return func(opts *ServeOptions) {
		opts.enableLogger = true
	}
}

// WithConfigPaths 设置自定义配置文件路径
func WithConfigPaths(paths []string) Option {
	return func(opts *ServeOptions) {
		opts.enableConfig = true
		opts.configPaths = paths
	}
}

// WithController 启用控制器初始化
func WithController() Option {
	return func(opts *ServeOptions) {
		opts.enableController = true
	}
}

// WithWebServer 启用Web服务器
func WithWebServer() Option {
	return func(opts *ServeOptions) {
		opts.enableWebServer = true
	}
}

// WithTask 启用定时任务
func WithTask() Option {
	return func(opts *ServeOptions) {
		opts.enableTask = true
	}
}

// WithPort 设置服务端口
func WithPort(port int) Option {
	return func(opts *ServeOptions) {
		opts.port = port
	}
}

// WithCustomInit 添加自定义初始化函数
func WithCustomInit(initFunc func() error) Option {
	return func(opts *ServeOptions) {
		opts.customInitFuncs = append(opts.customInitFuncs, initFunc)
	}
}

// WithCommandDefault 启用命令执行
func WithCommandDefault() Option {
	return func(opts *ServeOptions) {
		opts.enableConfig = true
		opts.enableLogger = true
		opts.enableCommand = true
	}
}

// WithWebDefault 启用所有默认组件（包括定时任务）
func WithWebDefault() Option {
	return func(opts *ServeOptions) {
		opts.enableConfig = true
		opts.enableLogger = true
		opts.enableController = true
		opts.enableTask = true
		opts.enableWebServer = true
	}
}

// NewServe 创建新的服务器实例
func NewServe(options ...Option) *Serve {
	opts := &ServeOptions{}

	// 应用所有选项
	for _, option := range options {
		option(opts)
	}

	// 创建iris应用
	app := iris.New()
	// 禁用iris的默认启动日志，我们将手动输出
	app.Configure(iris.WithConfiguration(iris.Configuration{
		DisableStartupLog: true,
	}))

	return &Serve{
		app:     app,
		options: opts,
	}
}

// Run 运行应用
func (s *Serve) Run() error {
	// 1. 初始化配置
	if s.options.enableConfig {
		if err := s.initConfiguration(); err != nil {
			return errs.Newf(errors.InternalError, "failed to initialize configuration: %v", err)
		}
	}

	// 2. 初始化配置
	if s.options.enableLogger {
		if err := s.initLogger(); err != nil {
			return errs.Newf(errors.InternalError, "failed to initialize logger: %v", err)
		}
	}

	// 5. 执行自定义初始化函数
	for _, initFunc := range s.options.customInitFuncs {
		if err := initFunc(); err != nil {
			return errs.Newf(errors.InternalError, "failed to execute custom initialization: %v", err)
		}
	}

	// 6. 初始化定时任务
	if s.options.enableTask {
		if err := s.initTask(); err != nil {
			return errs.Newf(errors.InternalError, "failed to initialize task scheduler: %v", err)
		}
	}

	// 7. 初始化控制器
	if s.options.enableController {
		if err := s.initController(); err != nil {
			return errs.Newf(errors.InternalError, "failed to initialize controller: %v", err)
		}
	}

	// 8. 执行命令（如果启用）
	if s.options.enableCommand {
		return s.executeCommand()
	}

	// 9. 启动Web服务（如果启用）
	if s.options.enableWebServer {
		return s.initWeb()
	}

	return nil
}

func (s *Serve) initConfiguration() error {
	if len(s.options.configPaths) > 0 {
		conf.InitConfigWithPaths(s.options.configPaths)
	} else {
		conf.InitConfig()
	}
	return nil
}

func (s *Serve) initLogger() error {
	return logger.InitLogger(conf.AppConfig.System.Log)
}

func (s *Serve) initCache() error {
	// 启动时加载,用于快速响应
	go func() {
		service.NewPhishingSitesService().LoadPhishingSites2Cache(context.Background())
	}()
	return nil
}

func (s *Serve) initController() error {
	// 设置路由
	controller.Routing(s.app)
	return nil
}

func (s *Serve) executeCommand() error {
	// 注册所有命令
	command.RegisterCommands()

	// 执行命令
	return command.Execute()
}

func (s *Serve) initTask() error {
	return task.InitTask()
}

func (s *Serve) initWeb() error {
	var port int
	if s.options.port > 0 {
		port = s.options.port
	} else {
		port = conf.AppConfig.System.Service.Port
	}

	portStr := fmt.Sprintf(":%d", port)

	logger.Infof("🚀 [Web] Application started. listening on %s. Press CTRL+C to shut down.", portStr)

	if err := s.app.Listen(portStr); err != nil {
		logger.Fatalf("Error starting server: %v", err)
		return err
	}
	return nil
}
