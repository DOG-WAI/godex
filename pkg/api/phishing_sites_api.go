package api

// CheckSitesReq 检查请求体
type CheckSitesReq = []string

// CheckSitesRsp 检查响应体
type CheckSitesRsp = []struct {
	Query  string `json:"query"`  // 查询的原始域名
	Domain string `json:"domain"` // 匹配到的
	Source string `json:"source"` // 数据来源
}
