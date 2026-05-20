package assistant

import (
	"errors"
	"fmt"
	"strings"
)

const (
	MsgTimeout           = "AI 服务响应超时，请稍后重试"
	MsgAuthFailed        = "AI 服务认证失败，请在设置中检查 API Key"
	MsgRateLimited       = "请求过于频繁，请稍后再试"
	MsgContainerNotFound = "容器 '%s' 不存在"
	MsgGeneralError      = "操作失败，请稍后重试"
)

var ErrRateLimited = errors.New("rate limited")

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func MapErrorToUserMessage(err error) string {
	if err == nil {
		return MsgGeneralError
	}

	if errors.Is(err, ErrRateLimited) {
		return MsgRateLimited
	}

	errStr := err.Error()

	if errors.Is(err, _errRetryTimeout) || strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return MsgTimeout
	}

	if strings.Contains(errStr, "status 401") || strings.Contains(errStr, "status 403") {
		return MsgAuthFailed
	}

	if strings.Contains(errStr, "status 429") {
		return MsgRateLimited
	}

	if strings.Contains(errStr, "not found") && strings.Contains(errStr, "container") {
		idx := strings.Index(errStr, "container '")
		if idx >= 0 {
			start := idx + len("container '")
			end := strings.Index(errStr[start:], "'")
			if end > 0 {
				return fmt.Sprintf(MsgContainerNotFound, errStr[start:start+end])
			}
		}
		return "容器不存在"
	}

	if strings.Contains(errStr, "status 5") || errors.Is(err, _errRetryServer) {
		return MsgTimeout
	}

	return MsgGeneralError
}
