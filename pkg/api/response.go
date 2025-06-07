package api

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
	TraceId interface{} `json:"trace-id,omitempty"`
}
