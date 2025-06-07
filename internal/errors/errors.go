package errors

import (
	"godex/pkg/errs"
	"godex/pkg/retcode"
)

// Errors ...
var (
	// Success 0
	Success = errs.New(0, "Success.")

	// RequestParamInvalid 参数错误
	RequestParamInvalid = errorCode(retcode.ErrorTypeParamsInvalid, 1)

	// InternalConfigErr 内部配置错误，如配置文件中缺少某种配置
	InternalConfigErr = errorCode(retcode.ErrorTypeBusinessErr, 2)

	// InternalError 服务内部错误
	InternalError = errorCode(retcode.ErrorTypeRPCFail, 3)

	// CallFail 调用错误
	CallFail = errorCode(retcode.ErrorTypeRPCFail, 4)
)

// ErrorCode ...
func ErrorCode(bizType retcode.BizType, errorType retcode.ErrorType, customCode int) int32 {
	return int32(int(bizType)*100000 + int(errorType)*1000 + customCode)
}

// errorCode 组装errcode
func errorCode(errorType retcode.ErrorType, customCode int) int {
	return int(ErrorCode(retcode.BizTypeBrowserExt, errorType, customCode))
}
