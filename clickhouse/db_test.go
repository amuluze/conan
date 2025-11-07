package clickhouse

import (
    "context"
    "errors"
    "testing"
    "time"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// TestBuildDSN 验证 buildDSN 能按预期格式构造 DSN 字符串。
// 重点检查：用户名、密码、主机、端口、数据库名与超时参数是否出现在字符串中。
func TestBuildDSN(t *testing.T) {
    cfg := &Config{
        Username: "u",
        Password: "p",
        Host:     "127.0.0.1",
        Port:     "9000",
    }
    dsn := buildDSN(cfg, "db1")
    wants := []string{
        "clickhouse://u:p@127.0.0.1:9000/db1",
        "dial_timeout=10s",
        "read_timeout=30s",
    }
    for _, w := range wants {
        if !contains(dsn, w) {
            t.Fatalf("dsn should contain %q, got: %s", w, dsn)
        }
    }
}

// TestPingWithSQLite 使用 sqlite 内存数据库构造 *DB，并调用 Ping 验证方法逻辑。
// 说明：此测试不依赖真实 ClickHouse，仅验证 Ping 的流程与返回值。
func TestPingWithSQLite(t *testing.T) {
    gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite memory: %v", err)
    }
    db := &DB{DB: gdb}
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if err := db.Ping(ctx); err != nil {
        t.Fatalf("Ping should succeed with sqlite memory DB, got error: %v", err)
    }
}

// TestNewDBErrorHandling 测试 NewDB 的错误处理
func TestNewDBErrorHandling(t *testing.T) {
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
            name: "missing username",
            config: &Config{
                Host:   "localhost",
                DBName: "db",
            },
            expectError: true,
            errorType:   ErrorTypeConfig,
        },
        {
            name: "missing database name",
            config: &Config{
                Host:     "localhost",
                Username: "user",
            },
            expectError: true,
            errorType:   ErrorTypeConfig,
        },
        {
            name: "valid minimal config",
            config: &Config{
                Host:     "localhost",
                Username: "user",
                DBName:   "db",
            },
            expectError: false, // 这个测试会在连接阶段失败，但不是配置错误
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewDB(tt.config)
            if tt.expectError {
                if err == nil {
                    t.Fatal("Expected error but got none")
                }
                // 检查错误类型
                var chErr *ClickHouseError
                if !errors.As(err, &chErr) || chErr.Type != tt.errorType {
                    t.Errorf("Expected %s error, got: %T", tt.errorType, err)
                }
            } else {
                // 对于有效的配置，可能会在连接阶段失败
                if err != nil && !IsConnectionError(err) {
                    t.Errorf("Expected connection error or success, got: %v", err)
                }
            }
        })
    }
}

// TestBuildTLSConfig 测试 TLS 配置构建
func TestBuildTLSConfig(t *testing.T) {
    tests := []struct {
        name        string
        config      *Config
        expectError bool
    }{
        {
            name: "disable SSL",
            config: &Config{
                SSLMode: "disable",
                Host:    "localhost",
            },
            expectError: false,
        },
        {
            name: "require SSL",
            config: &Config{
                SSLMode: "require",
                Host:    "localhost",
            },
            expectError: false,
        },
        {
            name: "verify-ca SSL",
            config: &Config{
                SSLMode: "verify-ca",
                Host:    "localhost",
            },
            expectError: false,
        },
        {
            name: "verify-full SSL with host",
            config: &Config{
                SSLMode: "verify-full",
                Host:    "localhost",
            },
            expectError: false,
        },
        {
            name: "verify-full SSL without host",
            config: &Config{
                SSLMode: "verify-full",
                Host:    "",
            },
            expectError: true,
        },
        {
            name: "unsupported SSL mode",
            config: &Config{
                SSLMode: "unsupported",
                Host:    "localhost",
            },
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := buildTLSConfig(tt.config)
            if tt.expectError && err == nil {
                t.Error("Expected error but got none")
            }
            if !tt.expectError && err != nil {
                t.Errorf("Expected no error but got: %v", err)
            }
        })
    }
}

// TestGetDSNInfo 测试获取 DSN 信息
func TestGetDSNInfo(t *testing.T) {
    config := &Config{
        Host:         "localhost",
        Port:         "9000",
        Username:     "user",
        Password:     "password", // 不应出现在 DSNInfo 中
        DBName:       "testdb",
        SSLMode:      "require",
        MaxLifetime:  300,
        MaxOpenConns: 100,
        MaxIdleConns: 50,
        OpenDB:       true,
    }

    info := config.GetDSNInfo()

    // 验证所有必需的字段都存在
    expectedFields := []string{"host", "port", "database", "username", "ssl_mode", "max_lifetime", "max_open_conns", "max_idle_conns", "use_native_db"}
    for _, field := range expectedFields {
        if _, exists := info[field]; !exists {
            t.Errorf("Expected field %q in DSNInfo", field)
        }
    }

    // 验证密码不在 DSNInfo 中
    if _, exists := info["password"]; exists {
        t.Error("Password should not be included in DSNInfo")
    }

    // 验证字段值
    if info["host"] != "localhost" {
        t.Errorf("Expected host=localhost, got %v", info["host"])
    }
    if info["username"] != "user" {
        t.Errorf("Expected username=user, got %v", info["username"])
    }
}

// TestConnectionPoolConfiguration 测试连接池配置
func TestConnectionPoolConfiguration(t *testing.T) {
    gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite memory: %v", err)
    }

    sqlDB, err := gdb.DB()
    if err != nil {
        t.Fatalf("failed to get sql.DB: %v", err)
    }

    config := &Config{
        MaxLifetime:  300,
        MaxOpenConns: 100,
        MaxIdleConns: 50,
    }

    // 测试连接池配置
    err = configureConnectionPool(sqlDB, config)
    if err != nil {
        t.Errorf("Failed to configure connection pool: %v", err)
    }

    // 验证配置是否生效
    if sqlDB.Stats().MaxOpenConnections != 100 {
        t.Errorf("Expected MaxOpenConnections=100, got %d", sqlDB.Stats().MaxOpenConnections)
    }
}

// TestPingTimeout 测试 Ping 超时处理
func TestPingTimeout(t *testing.T) {
    gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite memory: %v", err)
    }
    db := &DB{DB: gdb}

    // 使用极短的超时时间
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
    defer cancel()

    // 等待一小段时间确保超时
    time.Sleep(1 * time.Millisecond)

    err = db.Ping(ctx)
    if err == nil {
        t.Error("Expected timeout error but got none")
    }

    // 检查是否为超时错误
    if !IsTimeoutError(err) && !IsConnectionError(err) {
        t.Errorf("Expected timeout or connection error, got: %v", err)
    }
}

// TestPingCancellation 测试 Ping 取消处理
func TestPingCancellation(t *testing.T) {
    gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite memory: %v", err)
    }
    db := &DB{DB: gdb}

    ctx, cancel := context.WithCancel(context.Background())
    cancel() // 立即取消

    err = db.Ping(ctx)
    if err == nil {
        t.Error("Expected cancellation error but got none")
    }

    // 检查错误类型
    var chErr *ClickHouseError
    if errors.As(err, &chErr) && chErr.Code == "PING_CANCELED" {
        // 这是期望的取消错误
    } else if IsConnectionError(err) {
        // 或者是连接错误（也是可以接受的）
    } else {
        t.Errorf("Expected cancellation or connection error, got: %v", err)
    }
}