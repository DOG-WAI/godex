package task

// TaskSchedulerInterface 任务调度器接口
type TaskSchedulerInterface interface {
	// RegisterTask 注册任务函数
	RegisterTask(taskFunc TaskFunc)

	// Start 启动任务调度器
	Start() error

	// Stop 停止任务调度器
	Stop()

	// GetAvailableTaskNames 获取所有可用的任务名称
	GetAvailableTaskNames() []string
}

// TaskFunc 任务函数类型
type TaskFunc func() error

// 确保TaskScheduler实现了TaskSchedulerInterface接口
var _ TaskSchedulerInterface = (*TaskScheduler)(nil)
