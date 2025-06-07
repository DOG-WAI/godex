package task

import (
	"context"
	"godex/internal/conf"
	"godex/internal/service"
	"godex/pkg/logger"
	"godex/pkg/task"
)

var (
	// 全局任务调度器实例
	scheduler task.TaskSchedulerInterface
)

// InitTask 初始化任务调度器
func InitTask() error {
	// 创建任务调度器
	scheduler = task.NewTaskScheduler()

	// 注册所有业务任务
	registerBusinessTasks()

	// 从配置加载任务
	if taskScheduler, ok := scheduler.(*task.TaskScheduler); ok {
		if err := taskScheduler.LoadTasksFromConfig(conf.AppConfig.System.Tasks); err != nil {
			return err
		}
	}

	// 启动调度器
	return scheduler.Start()
}

// registerBusinessTasks 注册业务任务函数
func registerBusinessTasks() {
	scheduler.RegisterTask(CronTestTask)
	scheduler.RegisterTask(LoadPhishingSites2CacheTask)
	scheduler.RegisterTask(OnceTestTask)
}

// LoadPhishingSites2CacheTask 加载到cache
func LoadPhishingSites2CacheTask() error {
	logger.Info("开始执行任务: 导入数据")

	phishingSitesService := service.NewPhishingSitesService()
	err := phishingSitesService.LoadPhishingSites2Cache(context.Background())
	if err != nil {
		logger.Errorf("导入数据失败: %v", err)
		return err
	}

	logger.Info("任务完成: 导入数据")
	return nil
}

// ===== 测试任务 =====
// CronTestTask 简单测试任务
func CronTestTask() error {
	logger.Info("开始执行任务: 简单测试")
	return nil
}
func OnceTestTask() error {
	logger.Info("开始执行任务: 执行一次测试")
	return nil
}
