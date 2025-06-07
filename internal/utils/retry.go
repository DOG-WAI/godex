package utils

import "time"

// Handler ...
type Handler func() (interface{}, error)

// 方便业务逻辑重试，不用到处手动 for 重试
func Do(retryTimes int, duration time.Duration, handler Handler) (rsp interface{}, err error) {

	for i := 0; i < retryTimes; i++ {
		rsp, err = handler()
		if err == nil {
			return
		}

		if i == retryTimes-1 {
			break
		}

		time.Sleep(duration)
	}

	return rsp, err
}
