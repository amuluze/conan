package clickhouse

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestClickHouseError_Error 测试 ClickHouseError 的 Error() 方法
func TestClickHouseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ClickHouseError
		expected string
	}{
		{
			name:     "basic error",
			err:      NewConnectionError("connection failed", errors.New("network error")),
			expected: "[connection] connection failed (code: CONN_ERROR): network error",
		},
		{
			name:     "error with code",
			err:      NewConfigError("invalid config", nil).WithCode("CONFIG_001"),
			expected: "[config] invalid config (code: CONFIG_001)",
		},
		{
			name:     "error with context",
			err:      NewValidationError("invalid field", nil).WithContext("field", "username").WithContext("value", "test"),
			expected: "[validation] invalid field (code: VALIDATION_ERROR) | context: field=username, value=test",
		},
		{
			name:     "full error with all fields",
			err:      NewQueryError("query failed", errors.New("syntax error")).WithCode("QUERY_001").WithContext("table", "users").WithRetriable(true),
			expected: "[query] query failed (code: QUERY_001): syntax error | context: table=users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("ClickHouseError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestErrorTypes 测试各种错误类型的创建函数
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		errFunc  func() *ClickHouseError
		expected ErrorType
	}{
		{
			name:     "NewConnectionError",
			errFunc:  func() *ClickHouseError { return NewConnectionError("test", nil) },
			expected: ErrorTypeConnection,
		},
		{
			name:     "NewConfigError",
			errFunc:  func() *ClickHouseError { return NewConfigError("test", nil) },
			expected: ErrorTypeConfig,
		},
		{
			name:     "NewQueryError",
			errFunc:  func() *ClickHouseError { return NewQueryError("test", nil) },
			expected: ErrorTypeQuery,
		},
		{
			name:     "NewValidationError",
			errFunc:  func() *ClickHouseError { return NewValidationError("test", nil) },
			expected: ErrorTypeValidation,
		},
		{
			name:     "NewTimeoutError",
			errFunc:  func() *ClickHouseError { return NewTimeoutError("test", nil) },
			expected: ErrorTypeTimeout,
		},
		{
			name:     "NewTLSError",
			errFunc:  func() *ClickHouseError { return NewTLSError("test", nil) },
			expected: ErrorTypeTLS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc()
			if err.Type != tt.expected {
				t.Errorf("Error type = %q, want %q", err.Type, tt.expected)
			}
		})
	}
}

// TestErrorIs 测试 errors.Is 功能
func TestErrorIs(t *testing.T) {
	err1 := NewConnectionError("connection failed", nil).WithCode("CONN_001")
	err2 := NewConnectionError("connection failed", nil).WithCode("CONN_001")
	err3 := NewConnectionError("connection failed", nil).WithCode("CONN_002")

	// 相同类型和代码的错误应该匹配
	if !errors.Is(err1, err2) {
		t.Error("errors.Is should return true for errors with same type and code")
	}

	// 不同代码的错误不应该匹配
	if errors.Is(err1, err3) {
		t.Error("errors.Is should return false for errors with different codes")
	}
}

// TestErrorUnwrap 测试 errors.Unwrap 功能
func TestErrorUnwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewConnectionError("wrapper error", cause)

	if !errors.Is(err, cause) {
		t.Error("errors.Is should return true for underlying cause")
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("errors.Unwrap = %v, want %v", unwrapped, cause)
	}
}

// TestErrorTypeCheckers 测试错误类型检查函数
func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		isConnection bool
		isConfig     bool
		isQuery      bool
		isValidation bool
		isTimeout    bool
		isRetriable  bool
	}{
		{
			name:         "connection error",
			err:          NewConnectionError("network error", nil),
			isConnection: true,
			isRetriable:  true,
		},
		{
			name:     "config error",
			err:      NewConfigError("invalid config", nil),
			isConfig: true,
		},
		{
			name:    "query error",
			err:     NewQueryError("syntax error", nil),
			isQuery: true,
		},
		{
			name:         "validation error",
			err:          NewValidationError("invalid field", nil),
			isValidation: true,
		},
		{
			name:      "timeout error",
			err:       NewTimeoutError("operation timed out", nil),
			isTimeout: true,
			isRetriable: true,
		},
		{
			name: "regular error with timeout text",
			err:  errors.New("context deadline exceeded"),
			isTimeout: true,
			isRetriable: true,
		},
		{
			name: "regular error with connection text",
			err:  errors.New("connection refused"),
			isConnection: true,
			isRetriable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsConnectionError(tt.err) != tt.isConnection {
				t.Errorf("IsConnectionError = %v, want %v", IsConnectionError(tt.err), tt.isConnection)
			}
			if IsConfigError(tt.err) != tt.isConfig {
				t.Errorf("IsConfigError = %v, want %v", IsConfigError(tt.err), tt.isConfig)
			}
			if IsQueryError(tt.err) != tt.isQuery {
				t.Errorf("IsQueryError = %v, want %v", IsQueryError(tt.err), tt.isQuery)
			}
			if IsValidationError(tt.err) != tt.isValidation {
				t.Errorf("IsValidationError = %v, want %v", IsValidationError(tt.err), tt.isValidation)
			}
			if IsTimeoutError(tt.err) != tt.isTimeout {
				t.Errorf("IsTimeoutError = %v, want %v", IsTimeoutError(tt.err), tt.isTimeout)
			}
			if IsRetriableError(tt.err) != tt.isRetriable {
				t.Errorf("IsRetriableError = %v, want %v", IsRetriableError(tt.err), tt.isRetriable)
			}
		})
	}
}

// TestWrapError 测试错误包装功能
func TestWrapError(t *testing.T) {
	originalErr := NewConnectionError("original", nil)
	wrappedErr := WrapError(originalErr, ErrorTypeQuery, "wrapped message")

	// 检查包装后的错误类型
	// 注意：WrapError 对于 ClickHouseError 会保持原有类型
	if !IsConnectionError(wrappedErr) {
		t.Error("Wrapped ClickHouseError should preserve original type")
	}

	// 检查消息是否包含原始消息
	errorStr := wrappedErr.Error()
	if !contains(errorStr, "original") {
		t.Error("Wrapped error should contain original message")
	}

	// 包装 nil 错误
	nilErr := WrapError(nil, ErrorTypeConfig, "test")
	if nilErr == nil || nilErr.Type != ErrorTypeConfig {
		t.Error("Wrapping nil error should create new error")
	}

	// 包装普通错误
	regularErr := errors.New("regular error")
	wrappedRegularErr := WrapError(regularErr, ErrorTypeQuery, "wrapped")
	if !IsQueryError(wrappedRegularErr) {
		t.Error("Wrapped regular error should be query error type")
	}
}

// TestConfigValidate 测试配置验证
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorType:   ErrorTypeConfig,
		},
		{
			name: "empty host",
			config: &Config{
				Host:     "",
				Username: "user",
				DBName:   "db",
			},
			expectError: true,
			errorType:   ErrorTypeConfig,
		},
		{
			name: "invalid port",
			config: &Config{
				Host:     "localhost",
				Port:     "invalid",
				Username: "user",
				DBName:   "db",
			},
			expectError: true,
			errorType:   ErrorTypeConfig,
		},
		{
			name: "valid config with defaults",
			config: &Config{
				Host:     "localhost",
				Username: "user",
				DBName:   "db",
			},
			expectError: false,
		},
		{
			name: "invalid SSL mode",
			config: &Config{
				Host:     "localhost",
				Username: "user",
				DBName:   "db",
				SSLMode:  "invalid-mode",
			},
			expectError: true,
			errorType:   ErrorTypeConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !IsConfigError(err) {
					t.Errorf("Expected config error, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestPingErrorHandling 测试 Ping 方法的错误处理
func TestPingErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		db          *DB
		ctx         context.Context
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "nil DB",
			db:          nil,
			ctx:         context.Background(),
			expectError: true,
			errorType:   ErrorTypeConnection,
		},
		{
			name: "nil GORM DB",
			db: &DB{
				DB: nil,
			},
			ctx:         context.Background(),
			expectError: true,
			errorType:   ErrorTypeConnection,
		},
		{
			name: "timeout context",
			db: &DB{
				DB: nil, // 这里会导致错误
			},
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				return ctx
			}(),
			expectError: true,
			errorType:   ErrorTypeConnection,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.db.Ping(tt.ctx)
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				// 检查错误类型（某些测试可能因为 nil GORM DB 而不是超时而返回连接错误）
				if !IsConnectionError(err) {
					t.Errorf("Expected connection error, got %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidationFunctions 测试验证函数
func TestValidationFunctions(t *testing.T) {
	t.Run("validateTableName", func(t *testing.T) {
		tests := []struct {
			name        string
			tableName   string
			expectError bool
		}{
			{"empty table", "", false}, // 空表名是允许的（会被忽略）
			{"too long table", string(make([]byte, 65)), true},
			{"invalid characters", "invalid-table", true},
			{"SQL keyword", "SELECT", true},
			{"valid table", "valid_table", false},
			{"valid table with numbers", "table_123", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateTableName(tt.tableName)
				if tt.expectError && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})

	t.Run("validateID", func(t *testing.T) {
		tests := []struct {
			name        string
			id          string
			expectError bool
		}{
			{"empty ID", "", false}, // empty ID 是允许的（会被忽略）
			{"too long ID", string(make([]byte, 256)), true},
			{"ID with control chars", "id\x00", true},
			{"valid ID", "valid-id-123", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateID(tt.id)
				if tt.expectError && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})

	t.Run("validateFieldName", func(t *testing.T) {
		whitelist := map[string]struct{}{
			"id":         {},
			"name":       {},
			"created_at": {},
		}

		tests := []struct {
			name        string
			field       string
			expectError bool
		}{
			{"empty field", "", false}, // 空字段名是允许的（会被忽略）
			{"too long field", string(make([]byte, 65)), true},
			{"invalid characters", "invalid-field", true},
			{"field not in whitelist", "email", true},
			{"valid field", "name", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateFieldName(tt.field, whitelist)
				if tt.expectError && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})
}