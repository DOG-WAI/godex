package conf

import (
	"godex/pkg/cfgs"
	"godex/pkg/constant"
	"godex/pkg/logger"
	"godex/pkg/report"
	"godex/pkg/task"
)

// 全局配置及加载器
var (
	AppConfig    *Config // 请不要修改改对象
	configLoader *cfgs.ConfigLoader
)

// Config 是整个配置文件的入口结构体
type Config struct {
	System              SystemConfig              `yaml:"system" json:"system"`
	AppSetting          AppSettingConfig          `yaml:"app-setting" json:"app-setting"`
	EnvironmentVariable EnvironmentVariableConfig `yaml:"environment-variable" json:"environment-variable"`
}

// EnvironmentVariableConfig 环境变量集合
type EnvironmentVariableConfig struct {
	OssAccessKey       string `yaml:"oss-access-key" json:"oss-access-key"`
	OssAccessKeySecret string `yaml:"oss-access-key-secret" json:"oss-access-key-secret"`
}

type AppSettingConfig struct {
	ScamSniffer     string   `yaml:"scam-sniffer"`
	BatchUpsertSize int      `yaml:"batch-upsert-size"` // 批量插入,根据实际情况或 DB 参数调节
	BatchLoadSize   int      `yaml:"batch-load-size"`   // 批量加载
	FixedSniffer    []string `yaml:"fixed-sniffer"`
	BucketName      string   `yaml:"bucket-name"`
	BucketEndpoint  string   `yaml:"bucket-endpoint"`
}

// SystemConfig 包含其他相关的配置
type SystemConfig struct {
	ShowConf bool                `yaml:"show-conf"`
	Env      string              `yaml:"env"`
	Service  ServiceConfig       `yaml:"service" json:"service"`
	Log      logger.LogConfig    `yaml:"log" json:"log"`
	Tasks    []task.TaskConfig   `yaml:"tasks" json:"tasks"`
	Report   report.ReportConfig `yaml:"report" json:"report"`
}

// ServiceConfig 是服务相关的配置
type ServiceConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

// InitConfig 初始化配置，使用默认的配置路径
func InitConfig() {
	InitConfigWithPaths(constant.ConfPaths)
}

// InitConfigWithPaths 使用指定的配置路径初始化配置
func InitConfigWithPaths(configPaths []string) {
	AppConfig = &Config{}
	configLoader = cfgs.NewConfigLoader(configPaths, AppConfig)
	configLoader.InitConfig()
}
