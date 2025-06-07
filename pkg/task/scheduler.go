package task

import (
	"github.com/robfig/cron/v3"
	"godex/internal/errors"
	"godex/pkg/errs"
	"godex/pkg/logger"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// TaskConfig 任务配置结构体
type TaskConfig struct {
	Name        string `yaml:"name" json:"name"`
	Enable      bool   `yaml:"enable" json:"enable"`
	Once        bool   `yaml:"once" json:"once"`
	Cron        string `yaml:"cron" json:"cron"`
	Function    string `yaml:"function" json:"function"`
	Description string `yaml:"description" json:"description"`
}

// TaskRegistry 任务注册表
type TaskRegistry map[string]TaskFunc

// TaskScheduler 任务调度器
type TaskScheduler struct {
	cron     *cron.Cron
	registry TaskRegistry
}

// NewTaskScheduler 创建新的任务调度器
func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		cron:     cron.New(cron.WithSeconds()),
		registry: make(TaskRegistry),
	}
}

// RegisterTask 注册任务函数，自动使用函数名作为任务名称
func (ts *TaskScheduler) RegisterTask(taskFunc TaskFunc) {
	name := getFunctionName(taskFunc)
	ts.registry[name] = taskFunc
	logger.Debugf("Registered task: %s", name)
}

// getFunctionName 获取函数名称
func getFunctionName(taskFunc TaskFunc) string {
	funcPtr := runtime.FuncForPC(reflect.ValueOf(taskFunc).Pointer())
	fullName := funcPtr.Name()

	// 提取函数名，去掉包路径
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}

// GetAvailableTaskNames 获取所有可用的任务名称
func (ts *TaskScheduler) GetAvailableTaskNames() []string {
	functions := make([]string, 0, len(ts.registry))
	for name := range ts.registry {
		functions = append(functions, name)
	}
	return functions
}

// Start 启动任务调度器
func (ts *TaskScheduler) Start() error {
	// 输出所有可用的任务函数
	logger.Infof("🎉 Available task functions: %v", ts.GetAvailableTaskNames())

	// 启动调度器
	ts.cron.Start()
	logger.Info("🎉 Task scheduler started successfully")
	return nil
}

// Stop 停止任务调度器
func (ts *TaskScheduler) Stop() {
	ts.cron.Stop()
	logger.Info("Task scheduler stopped")
}

// LoadTasksFromConfig 从配置加载任务
func (ts *TaskScheduler) LoadTasksFromConfig(taskConfigs []TaskConfig) error {
	for _, taskConfig := range taskConfigs {
		if !taskConfig.Enable {
			logger.Debugf("Task '%s' is disabled, skipping", taskConfig.Name)
			continue
		}

		// 从注册表中获取任务函数
		taskFunc, exists := ts.registry[taskConfig.Function]
		if !exists {
			logger.Errorf("Task function '%s' not found in registry. Available functions: %v",
				taskConfig.Function, ts.GetAvailableTaskNames())
			continue
		}

		// 处理 @once 类型的任务
		if taskConfig.Cron == "@once" {
			logger.Infof("🎉 Scheduling one-time task: %s", taskConfig.Name)

			// 创建闭包，避免循环变量问题
			taskName := taskConfig.Name
			go func() {
				// 稍微延迟执行，确保系统完全启动
				time.Sleep(5 * time.Second)
				logger.Infof("Executing one-time task: %s", taskName)
				if err := taskFunc(); err != nil {
					logger.Errorf("One-time task '%s' execution failed: %v", taskName, err)
				} else {
					logger.Infof("One-time task '%s' executed successfully", taskName)
				}
			}()
			continue
		}

		// 创建闭包，避免循环变量问题
		taskName := taskConfig.Name
		_, err := ts.cron.AddFunc(taskConfig.Cron, func() {
			logger.Infof("Executing task: %s", taskName)
			if err := taskFunc(); err != nil {
				logger.Errorf("Task '%s' execution failed: %v", taskName, err)
			} else {
				logger.Infof("Task '%s' executed successfully", taskName)
			}
		})

		if err != nil {
			return errs.Newf(errors.InternalError, "failed to add cron job for task '%s': %v", taskConfig.Name, err)
		}

		logger.Infof("🎉 Registered task: %s (cron: %s, function: %s)",
			taskConfig.Name, taskConfig.Cron, taskConfig.Function)
	}

	return nil
}
