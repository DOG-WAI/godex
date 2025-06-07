package cfgs

// ConfigLoaderInterface 配置加载器接口
type ConfigLoaderInterface interface {
	// InitConfig 初始化配置
	InitConfig()

	// GetConfig 获取配置实例
	GetConfig() interface{}

	// ReloadConfig 重新加载配置
	ReloadConfig() error
}

// 确保ConfigLoader实现了ConfigLoaderInterface接口
var _ ConfigLoaderInterface = (*ConfigLoader)(nil)
