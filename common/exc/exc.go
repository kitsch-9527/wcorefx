package exc

import (
	"fmt"
	"strings"
)

// 自定义错误类型：业务错误（带错误码和操作详情）
type CoreError struct {
	Action string   // 错误方法
	Errors []string // 错误信息
}

// GetMessages：获取所有错误信息（方便外部遍历）
func (e *CoreError) GetError() []string {
	return e.Errors
}

// HasError：判断是否包含指定错误信息（模糊匹配）
func (e *CoreError) HasError(target string) bool {
	for _, msg := range e.Errors {
		if strings.Contains(msg, target) {
			return true
		}
	}
	return false
}

func New(action string, errors ...string) error {
	errMassage := make([]string, 0, len(errors))
	errMassage = append(errMassage, errors...)
	return &CoreError{
		Action: action,
		Errors: errMassage,
	}
}

func (e *CoreError) Add(msg string) {
	e.Errors = append(e.Errors, msg)
}

// 实现 error 接口：拼接所有错误信息，格式化为字符串
func (e *CoreError) Error() string {
	return fmt.Sprintf("Action: %s ; Errors: %s ;", e.Action, strings.Join(e.Errors, "\n  "))
}
