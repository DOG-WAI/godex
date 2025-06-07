package cfgs

import (
	"github.com/fsnotify/fsnotify"
	"godex/pkg/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"sync"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
	configPaths []string     // 配置文件路径列表
	configPath  string       // 实际使用的配置文件路径
	config      interface{}  // 配置实例指针
	configLock  sync.RWMutex // 配置读写锁
}

// NewConfigLoader 创建新的配置加载器
func NewConfigLoader(configPaths []string, config interface{}) *ConfigLoader {
	return &ConfigLoader{
		configPaths: configPaths,
		config:      config,
	}
}

// InitConfig 初始化配置
func (c *ConfigLoader) InitConfig() {
	path := c.findConfigFile()
	if path == "" {
		logger.Fatalf("未找到配置文件，请确保以下路径之一存在配置文件：%v", c.configPaths)
	}
	c.configPath = path
	logger.IgnoreError(c.loadConfig(c.configPath))
	go c.watchConfigFile(c.configPath)
}

// findConfigFile 按优先级查找配置文件
func (c *ConfigLoader) findConfigFile() string {
	for _, path := range c.configPaths {
		if _, err := os.Stat(path); err == nil {
			logger.Infof("使用配置文件: %s", path)
			return path
		}
	}
	return ""
}

// loadConfig 加载配置文件
func (c *ConfigLoader) loadConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	c.configLock.Lock()
	defer c.configLock.Unlock()

	if err = yaml.Unmarshal(data, c.config); err != nil {
		return err
	}

	logger.Infof("配置文件加载成功: %s", path)
	return nil
}

// watchConfigFile 监控配置文件变化并重新加载
func (c *ConfigLoader) watchConfigFile(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Errorf("创建文件监控失败: %+v", err)
		panic(err)
	}
	defer func() {
		logger.IgnoreError(watcher.Close())
	}()

	// 添加需要监控的配置文件
	if err := watcher.Add(path); err != nil {
		logger.Errorf("添加监控文件失败: %+v", err)
		panic(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// 只处理写入或创建事件
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				logger.Infof("检测到配置文件变动，重新加载配置...")
				if err := c.loadConfig(path); err != nil {
					logger.Errorf("重新加载配置失败: %+v", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Errorf("监控错误: %+v", err)
		}
	}
}

// GetConfig 获取配置（线程安全）
func (c *ConfigLoader) GetConfig() interface{} {
	c.configLock.RLock()
	defer c.configLock.RUnlock()
	return c.config
}

// ReloadConfig 手动重新加载配置
func (c *ConfigLoader) ReloadConfig() error {
	return c.loadConfig(c.configPath)
}
