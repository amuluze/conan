package clickhouse

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorType 定义错误的分类类型
type ErrorType string

const (
	// ErrorTypeConnection 连接相关错误
	ErrorTypeConnection ErrorType = "connection"
	// ErrorTypeConfig 配置相关错误
	ErrorTypeConfig ErrorType = "config"
	// ErrorTypeQuery 查询相关错误
	ErrorTypeQuery ErrorType = "query"
	// ErrorTypeValidation 参数验证错误
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeTimeout 超时错误
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeTLS TLS相关错误
	ErrorTypeTLS ErrorType = "tls"
	// ErrorTypeUnknown 未知错误
	ErrorTypeUnknown ErrorType = "unknown"
)

// ClickHouseError 自定义错误结构
type ClickHouseError struct {
	Type     ErrorType
	Message  string
	Cause    error
	Context  map[string]interface{}
	Code     string // 可选的错误代码
	Retriable bool  // 是否可重试
}

// Error 实现 error 接口
func (e *ClickHouseError) Error() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("[%s] %s", e.Type, e.Message))

	if e.Code != "" {
		builder.WriteString(fmt.Sprintf(" (code: %s)", e.Code))
	}

	if e.Cause != nil {
		builder.WriteString(fmt.Sprintf(": %v", e.Cause))
	}

	if len(e.Context) > 0 {
		builder.WriteString(" | context: ")
		first := true
		for k, v := range e.Context {
			if !first {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
	}

	return builder.String()
}

// Unwrap 支持 errors.Unwrap
func (e *ClickHouseError) Unwrap() error {
	return e.Cause
}

// Is 支持 errors.Is
func (e *ClickHouseError) Is(target error) bool {
	if chErr, ok := target.(*ClickHouseError); ok {
		return e.Type == chErr.Type && e.Code == chErr.Code
	}
	return false
}

// NewClickHouseError 创建新的 ClickHouse 错误
func NewClickHouseError(errType ErrorType, message string, cause error) *ClickHouseError {
	return &ClickHouseError{
		Type:      errType,
		Message:   message,
		Cause:     cause,
		Context:   make(map[string]interface{}),
		Retriable: isRetriableError(errType, cause),
	}
}

// WithContext 添加上下文信息
func (e *ClickHouseError) WithContext(key string, value interface{}) *ClickHouseError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCode 设置错误代码
func (e *ClickHouseError) WithCode(code string) *ClickHouseError {
	e.Code = code
	return e
}

// WithRetriable 设置是否可重试
func (e *ClickHouseError) WithRetriable(retriable bool) *ClickHouseError {
	e.Retriable = retriable
	return e
}

// isRetriableError 判断错误是否可重试
func isRetriableError(errType ErrorType, cause error) bool {
	switch errType {
	case ErrorTypeConnection, ErrorTypeTimeout:
		return true
	case ErrorTypeQuery:
		// 检查是否是网络相关错误
		if cause != nil {
			causeStr := strings.ToLower(cause.Error())
			networkErrors := []string{
				"connection refused",
				"connection reset",
				"timeout",
				"network is unreachable",
				"no such host",
				"temporary failure",
			}
			for _, netErr := range networkErrors {
				if strings.Contains(causeStr, netErr) {
					return true
				}
			}
		}
		return false
	case ErrorTypeConfig, ErrorTypeValidation, ErrorTypeTLS:
		return false
	default:
		return false
	}
}

// IsConnectionError 判断是否为连接错误
func IsConnectionError(err error) bool {
	var chErr *ClickHouseError
	if errors.As(err, &chErr) {
		return chErr.Type == ErrorTypeConnection
	}

	// 检查常见的连接错误字符串
	errStr := strings.ToLower(err.Error())
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"no connection available",
		"max open connections",
		"connection exhausted",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(errStr, connErr) {
			return true
		}
	}

	return false
}

// IsConfigError 判断是否为配置错误
func IsConfigError(err error) bool {
	var chErr *ClickHouseError
	if errors.As(err, &chErr) {
		return chErr.Type == ErrorTypeConfig
	}
	return false
}

// IsQueryError 判断是否为查询错误
func IsQueryError(err error) bool {
	var chErr *ClickHouseError
	if errors.As(err, &chErr) {
		return chErr.Type == ErrorTypeQuery
	}
	return false
}

// IsValidationError 判断是否为验证错误
func IsValidationError(err error) bool {
	var chErr *ClickHouseError
	if errors.As(err, &chErr) {
		return chErr.Type == ErrorTypeValidation
	}
	return false
}

// IsTimeoutError 判断是否为超时错误
func IsTimeoutError(err error) bool {
	var chErr *ClickHouseError
	if errors.As(err, &chErr) {
		return chErr.Type == ErrorTypeTimeout
	}

	// 检查常见的超时错误字符串
	errStr := strings.ToLower(err.Error())
	timeoutErrors := []string{
		"timeout",
		"deadline exceeded",
		"context deadline exceeded",
	}

	for _, timeoutErr := range timeoutErrors {
		if strings.Contains(errStr, timeoutErr) {
			return true
		}
	}

	return false
}

// IsRetriableError 判断错误是否可重试
func IsRetriableError(err error) bool {
	var chErr *ClickHouseError
	if errors.As(err, &chErr) {
		return chErr.Retriable
	}

	// 默认判断逻辑
	return IsConnectionError(err) || IsTimeoutError(err)
}

// WrapError 包装现有错误为 ClickHouseError
func WrapError(err error, errType ErrorType, message string) *ClickHouseError {
	if err == nil {
		return NewClickHouseError(errType, message, nil)
	}

	// 如果已经是 ClickHouseError，则保留原有类型，只添加信息
	if chErr, ok := err.(*ClickHouseError); ok {
		return &ClickHouseError{
			Type:      chErr.Type,
			Message:   message + ": " + chErr.Message,
			Cause:     chErr.Cause,
			Context:   chErr.Context,
			Code:      chErr.Code,
			Retriable: chErr.Retriable,
		}
	}

	return NewClickHouseError(errType, message, err)
}

// Common error creation functions

// NewConnectionError 创建连接错误
func NewConnectionError(message string, cause error) *ClickHouseError {
	return NewClickHouseError(ErrorTypeConnection, message, cause).
		WithCode("CONN_ERROR").
		WithRetriable(true)
}

// NewConfigError 创建配置错误
func NewConfigError(message string, cause error) *ClickHouseError {
	return NewClickHouseError(ErrorTypeConfig, message, cause).
		WithCode("CONFIG_ERROR").
		WithRetriable(false)
}

// NewQueryError 创建查询错误
func NewQueryError(message string, cause error) *ClickHouseError {
	return NewClickHouseError(ErrorTypeQuery, message, cause).
		WithCode("QUERY_ERROR")
}

// NewValidationError 创建验证错误
func NewValidationError(message string, cause error) *ClickHouseError {
	return NewClickHouseError(ErrorTypeValidation, message, cause).
		WithCode("VALIDATION_ERROR").
		WithRetriable(false)
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(message string, cause error) *ClickHouseError {
	return NewClickHouseError(ErrorTypeTimeout, message, cause).
		WithCode("TIMEOUT_ERROR").
		WithRetriable(true)
}

// NewTLSError 创建 TLS 错误
func NewTLSError(message string, cause error) *ClickHouseError {
	return NewClickHouseError(ErrorTypeTLS, message, cause).
		WithCode("TLS_ERROR").
		WithRetriable(false)
}