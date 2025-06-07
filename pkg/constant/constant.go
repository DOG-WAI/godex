package constant

// 配置文件路径优先级 ...
var ConfPaths = []string{"./config.yaml", "./conf/config.yaml", "/etc/conf/config.yaml"}

// MiddlewareLoggerKeys 上下文常量定义，用于中间件日志记录
const (
	TraceIDKey       = "trace_id"
	IPKey            = "ip"
	HostNameKey      = "host_name"
	UserIDKey        = "user_id"
	MethodKey        = "method"
	PathKey          = "path"
	QueryKey         = "query"
	RequestKey       = "request"
	RequestHeaderKey = "request_header"
	UserAgentKey     = "user_agent"
	StatusKey        = "status"
	LatencyKey       = "latency"
)

// Env ...
const (
	EnvLocal   = "local"
	EnvDev     = "dev"
	EnvTesting = "testing"
	EnvPre     = "pre"
	EnvProd    = "prod"
)
