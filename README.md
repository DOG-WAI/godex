# godex

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Status](https://img.shields.io/badge/Status-Active-success)]()


## 🏢 赞助

感谢以下赞助商支持 XUGOU 的开发：

[![Powered by DartNode](https://dartnode.com/branding/DN-Open-Source-sm.png)](https://dartnode.com "Powered by DartNode - Free VPS for Open Source")


## 📋 项目简介

`godex` 是一个Go应用框架。

### 核心特性

- 🛡️ **检测**: 实时检测域名是否为
- 🚀 **高性能缓存**: 将数据加载到内存中，提供毫秒级响应
- ⏰ **定时更新**: 自动从第三方数据源更新列表
- 🎯 **类型安全**: 基于静态声明式API，编译时类型检查
- 📊 **任务调度**: 支持定时任务和一次性任务执行
- 🔧 **模块化架构**: 支持双模式运行(Web/CLI)，组件可选择启用

## 🚀 快速开始

### 环境要求

- Go 1.22+(编译期)

### 安装与运行

```bash
# 克隆项目
git clone <repository-url>
cd godex

# 安装依赖
go mod tidy

# 启动服务
go run cmd/main.go

# 或构建后运行
make build
./app
```

### 配置文件

配置文件加载优先级：
1. `.config.yaml` (项目根目录)
2. `./conf/config.yaml` (推荐)
3. `/etc/conf/config.yaml` (系统级)

## ⚙️ 配置文件说明

项目使用YAML格式配置文件，示例如下，完整配置可以查看`./conf/config-example.yaml`：

```yaml
### 系统配置 ###
system:
  show-conf: false                    # 非develop环境中严禁开启
  log:
    level: INFO                      # 日志级别: DEBUG, INFO, WARN, ERROR
    file: "app.log"                 # 日志文件名
  service:
    name: "godex"      # 服务名称
    port: 8000                      # HTTP服务端口
  tasks:                            # 定时任务配置
    - name: "预加载数据到内存"
      enable: true                  # 是否启用该任务
      cron: "@once"                # 启动时执行一次
      function: "LoadPhishingSites2CacheTask"
      description: "启动时预加载数据"
    - name: "定时加载数据到内存"
      enable: true
      cron: "0 0 2 * * *"          # 每天凌晨2点执行 (秒 分 时 日 月 周)
      function: "LoadPhishingSites2CacheTask"
      description: "定时更新数据"
    - name: "测试任务1"
      enable: false                 # 可设为false禁用
      cron: "*/5 * * * * *"        # 每5秒执行一次
      function: "CronTestTask"
      description: "测试任务"
  report:                          # 数据上报配置
    endpoint: https://sys-test.adspower.net
    enable: true                   # 是否启用上报
    aes-public-key: |             # AES加密公钥
      -----BEGIN PUBLIC KEY-----
      [公钥内容]
      -----END PUBLIC KEY-----

### 数据库配置 ###
databases:
  dsn: "user:password@tcp(host:port)/database?parseTime=true&loc=Local"
  driver-name: mysql               # 支持mysql, postgres等
  max-idle: 4                     # 最大空闲连接数
  max-open: 8                     # 最大打开连接数
  max-life-time: 600s             # 连接最大生命周期
  max-idle-time: 300s             # 空闲连接超时时间
  show-sql: false                 # 是否显示SQL语句(调试用)

### 应用设置 ###
app-setting:
  scam-sniffer: "https://raw.githubusercontent.com/scamsniffer/scam-database/refs/heads/main/blacklist/domains.json"
  batch-upsert-size: 5000         # 批量插入数据大小
  batch-load-size: 100000         # 批量加载数据大小
  fixed-sniffer:                  # 固定的列表
    - phishing-sites-foo.com
    - phishing-sites-bar.com
```

**配置要点说明：**

- **Cron表达式格式**: 支持标准格式`"0 0 2 * * *"`(秒分时日月周)、特殊关键字`@once/@daily/@hourly`、间隔表达式`@every 5s`
- **任务配置**: 通过`enable`字段控制任务启用/禁用，修改后需重启服务生效
- **数据库连接**: DSN格式为标准数据库连接字符串，支持连接池参数调优
- **安全配置**: 生产环境请使用环境变量或K8s Secret管理敏感信息，不要将密码直接写入配置文件
- **日志配置**: `show-sql`建议仅在开发环境启用，生产环境关闭以提高性能

## 🏗️ 架构设计

### API设计原则

本项目采用静态声明式API设计，确保类型安全和代码简洁：

```go
// ✅ 推荐的API声明方式
// CheckSites 检查网站是否为
func (c *phishingSitesLogic) CheckSites(ctx context.Context, req api.CheckSitesReq) (api.CheckSitesRsp, error) {
    phishingSites, err := service.NewPhishingSitesService().CheckPhishingSitesWithCache(ctx, req)
    if err != nil {
        return nil, errs.Newf(errors.RequestParamInvalid, "check phishing sites failed")
    }

    var rsp api.CheckSitesRsp
    if err := copier.Copy(&rsp, &phishingSites); err != nil {
        return nil, errs.Newf(errors.InternalError, "copy response data failed: %v", err)
    }

    return rsp, nil
}
```

### 核心优势

- ✅ **透明化**: 框架自动处理请求绑定、错误处理、响应序列化
- ✅ **类型安全**: 编译时验证请求和响应类型
- ✅ **代码简洁**: 开发者只需关注核心业务逻辑
- ✅ **一致性**: 统一的API声明风格
- ✅ **便捷性**: 一个函数声明即可完成完整的API处理

### 🏗️ 高级特性

#### 模块化启动架构
- **双模式运行**: 同一二进制文件支持Web服务和命令行工具两种模式
- **组件可选**: 可选择性启用数据库、缓存、定时任务等组件  
- **插件式扩展**: 支持自定义初始化函数，便于功能扩展

```go
// 支持多种启动模式
serve.WithWebDefault()     // Web服务模式
serve.WithCommandDefault() // 命令行模式
serve.WithCustomInit()     // 自定义初始化
```

#### 配置热重载
- **零停机更新**: 业务配置文件修改后自动重载，无需重启服务
- **多路径支持**: 支持多个配置文件路径，按优先级自动查找
- **线程安全**: 读写锁保护配置的并发访问

#### 企业级错误处理
- **错误分类管理**: 区分框架错误、业务错误和下游错误
- **堆栈跟踪**: 可配置的错误堆栈跟踪，支持过滤无关信息
- **优雅降级**: 静默错误处理机制，避免非关键错误影响主流程

```go
// 支持堆栈跟踪的错误处理
errs.SetTraceableWithContent("service_name") // 过滤特定内容
logger.IgnoreError(err) // 静默处理非关键错误
```

#### 高性能批处理
- **内存可控**: 分批加载避免内存溢出
- **性能优化**: 批量操作提高数据库性能  
- **灵活配置**: 可根据硬件资源调整批次大小

```yaml
app-setting:
  batch-upsert-size: 5000   # 批量插入大小
  batch-load-size: 100000   # 批量加载大小
```

#### 数据安全传输
- **双重加密**: AES对称加密 + RSA非对称加密保护敏感数据
- **异步上报**: 数据上报不阻塞主业务流程
- **可配置开关**: 支持按环境启用/禁用上报功能

#### 泛型API处理器
- **编译时类型安全**: 避免运行时类型错误
- **自动序列化**: 框架自动处理JSON序列化/反序列化
- **错误统一处理**: 统一的错误响应格式

```go
// 泛型API处理器示例
func Handler[TReq any, TRsp any](handler func(ctx context.Context, req TReq) (TRsp, error))
```

### 开发规范

1. **错误处理**: 必须对所有错误进行适当处理
2. **上下文传递**: Service方法第一个参数必须为`context.Context`
3. **接口限制**: 非特殊情况不允许定义interface类型的API
4. **上下文生命周期**: 向下传递的ctx不允许被结构体引用
5. **非develop环境中严禁开启** system.show-conf: true
6. **错误码管理**: 新增错误码必须定义到 `internal/errors` 目录下，禁止在业务代码中硬编码错误码
7. **配置管理**: 所有配置项必须通过 `internal/conf` 配合 `config.yaml` 进行加载，严禁硬编码到项目内

## 📁 目录结构

```
├── cmd/                          # 程序入口，仅存放main函数
│   └── main.go
├── conf/                         # 配置文件目录
│   └── config.yaml
├── internal/                     # 内部代码，仅供本服务使用
│   ├── cache/                    # 缓存相关
│   ├── command/                  # 命令模式入口
│   ├── configuration/            # 配置文件加载
│   ├── controller/               # HTTP控制器
│   ├── db/                       # 数据库连接
│   ├── entity/                   # 数据传输对象
│   ├── errors/                   # 错误定义
│   ├── logic/                    # 核心业务逻辑
│   ├── model/                    # 数据模型定义
│   ├── repo/                     # 数据访问层
│   ├── resty/                    # 第三方API调用
│   ├── serve/                    # 服务初始化
│   ├── service/                  # 业务服务层
│   ├── task/                     # 定时任务
│   └── utils/                    # 工具函数
├── pkg/                          # 公共组件，可被外部引用
│   ├── cfgs/                     # 通用配置加载器
│   ├── constant/                 # 常量定义
│   ├── errs/                     # 错误处理基础包
│   ├── logger/                   # 日志组件
│   ├── report/                   # 上报组件
│   ├── retcode/                  # 状态码定义
│   └── scheduler/                # 定时任务组件
├── Makefile                      # 构建脚本
└── README.md                     # 项目文档
```

## 🔧 功能模块

### 检测服务

- **数据源**: 从ScamSniffer等第三方源获取列表
- **实时检测**: 提供HTTP API检查域名安全性
- **缓存机制**: 内存缓存提高查询性能
- **自动更新**: 定期同步最新的威胁情报

### 定时任务系统

支持多种类型的定时任务：

- **一次性任务**: 启动时执行的初始化任务
- **定时任务**: 基于Cron表达式的周期性任务
- **手动触发**: 支持命令行手动执行任务

### 命令行工具

```bash
# 手动更新数据
./app importPhishingSites

# 查看可用命令
./app --help
```

## 📊 服务监控

### 启动成功日志示例

```log
INFO[2025-06-05T14:57:38+08:00] 🎉 Logger initialized successfully with daily rotation (max_days: 30, timezone: ) 
INFO[2025-06-05T14:57:38+08:00] [task/scheduler.go:105] 🎉 Scheduling one-time task: 预加载数据到内存 
INFO[2025-06-05T14:57:38+08:00] [task/scheduler.go:137] 🎉 Registered task: 定时加载数据到内存 (cron: 0 0 2 * * *, function: LoadPhishingSites2CacheTask) 
INFO[2025-06-05T14:57:38+08:00] [task/scheduler.go:137] 🎉 Registered task: 测试任务1 (cron: */5 * * * * *, function: CronTestTask) 
INFO[2025-06-05T14:57:38+08:00] [task/scheduler.go:73] 🎉 Available task functions: [CronTestTask LoadPhishingSites2CacheTask OnceTestTask] 
INFO[2025-06-05T14:57:38+08:00] [task/scheduler.go:77] 🎉 Task scheduler started successfully 
INFO[2025-06-05T14:57:38+08:00] [serve/serve.go:242] 🚀 [Web] Application started. listening on :8000. Press CTRL+C to shut down. 
```

## 🛠️ 开发指南

### Make命令

```bash
make          # 构建项目
make check    # 代码检查
make unit-test    # 单元测试
make api-test     # API测试
make upload       # 上传部署
```

### 配置管理

- 开发环境：使用 `./conf/config.yaml`
- 生产环境：建议使用K8s ConfigMap挂载配置文件
- 配置修改后需要重启服务生效（定时任务配置）

### 数据库依赖

根据配置文件中的数据库设置，支持多种数据库类型。具体配置请参考 `./conf/config.yaml` 示例。
