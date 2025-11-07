package clickhouse

import (
    "context"
    "crypto/tls"
    "fmt"
    "time"

    ch "github.com/ClickHouse/clickhouse-go/v2"
    gormclickhouse "gorm.io/driver/clickhouse"
    "gorm.io/gorm"
)

type DB struct {
    *gorm.DB
    autoMigrate bool
}

// NewDB 根据配置创建并返回一个 GORM 的 ClickHouse 数据库实例。
// - 验证配置参数
// - 初始化 Dialector 并打开 GORM 连接
// - 根据配置设置连接池参数（最大空闲连接、最大打开连接、连接生命周期）
// - 如果开启 Debug，则返回带有调试信息的 DB
func NewDB(config *Config) (*DB, error) {
    // 验证配置
    if err := config.Validate(); err != nil {
        return nil, WrapError(err, ErrorTypeConfig, "failed to validate database configuration")
    }

    // 构建 Dialector
    dial, err := dial(config)
    if err != nil {
        return nil, WrapError(err, ErrorTypeConnection, "failed to create database dialer").
            WithContext("host", config.Host).
            WithContext("port", config.Port).
            WithContext("database", config.DBName)
    }

    // 打开数据库连接
    db, err := gorm.Open(dial, &gorm.Config{})
    if err != nil {
        return nil, NewConnectionError("failed to open database connection", err).
            WithContext("host", config.Host).
            WithContext("port", config.Port).
            WithContext("database", config.DBName).
            WithCode("CONN_OPEN_FAILED")
    }

    // 设置调试模式
    if config.Debug {
        db = db.Debug()
    }

    // 获取底层 SQL DB 以配置连接池
    sqlDB, err := db.DB()
    if err != nil {
        return nil, NewConnectionError("failed to get underlying sql.DB", err).
            WithContext("database", config.DBName).
            WithCode("SQLDB_GET_FAILED")
    }

    // 配置连接池参数
    if err := configureConnectionPool(sqlDB, config); err != nil {
        return nil, WrapError(err, ErrorTypeConnection, "failed to configure connection pool").
            WithContext("max_open_conns", config.MaxOpenConns).
            WithContext("max_idle_conns", config.MaxIdleConns).
            WithContext("max_lifetime", config.MaxLifetime)
    }

    return &DB{DB: db, autoMigrate: config.AutoMigrate}, nil
}

// dial 构建 ClickHouse 的 GORM Dialector。
// 说明：
// - 同时提供 DSN 与已有 *sql.DB（通过 ch.OpenDB 构建）两种方式，增强兼容性
// - 使用 clickhouse-go v2 的 OpenDB 与 Options/Settings 配置，而不是错误的 stdlib 包路径
func dial(cfg *Config) (gorm.Dialector, error) {
    dbName := cfg.DBName
    if dbName == "" {
        dbName = cfg.Database
    }

    // 构建 TLS 配置
    tlsConf, err := buildTLSConfig(cfg)
    if err != nil {
        return nil, WrapError(err, ErrorTypeTLS, "failed to build TLS configuration").
            WithContext("ssl_mode", cfg.SSLMode).
            WithContext("host", cfg.Host)
    }

    // 根据配置决定是否通过 OpenDB 初始化底层连接
    if cfg.OpenDB {
        options := &ch.Options{
            Addr: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
            Auth: ch.Auth{
                Database: dbName,
                Username: cfg.Username,
                Password: cfg.Password,
            },
            TLS: tlsConf,
            Settings: ch.Settings{
                "max_execution_time": 60,
            },
            DialTimeout: 10 * time.Second,
            ReadTimeout: 30 * time.Second,
        }

        sqlDB := ch.OpenDB(options)

        return gormclickhouse.New(gormclickhouse.Config{
            Conn: sqlDB,
        }), nil
    }

    // 默认仅通过 DSN 初始化，简化配置与依赖
    dsn := buildDSN(cfg, dbName)
    return gormclickhouse.Open(dsn), nil
}

// buildTLSConfig 根据SSL模式构建TLS配置
func buildTLSConfig(cfg *Config) (*tls.Config, error) {
    switch cfg.SSLMode {
    case "", "disable":
        return nil, nil
    case "require":
        return &tls.Config{InsecureSkipVerify: true}, nil
    case "verify-ca":
        return &tls.Config{InsecureSkipVerify: false}, nil
    case "verify-full":
        if cfg.Host == "" {
            return nil, NewTLSError("host cannot be empty when using verify-full SSL mode", nil).
                WithCode("TLS_HOST_REQUIRED")
        }
        return &tls.Config{InsecureSkipVerify: false, ServerName: cfg.Host}, nil
    default:
        return nil, NewTLSError(fmt.Sprintf("unsupported SSL mode: %s", cfg.SSLMode), nil).
            WithContext("ssl_mode", cfg.SSLMode).
            WithCode("TLS_UNSUPPORTED_MODE")
    }
}

// configureConnectionPool 配置数据库连接池
func configureConnectionPool(sqlDB interface{}, cfg *Config) error {
    // 使用类型断言来获取 *sql.DB
    sqlDBType, ok := sqlDB.(interface {
        SetMaxIdleConns(n int)
        SetMaxOpenConns(n int)
        SetConnMaxLifetime(d time.Duration)
    })

    if !ok {
        return NewConnectionError("unexpected database connection type, expected *sql.DB interface", nil).
            WithContext("type", fmt.Sprintf("%T", sqlDB)).
            WithCode("CONN_WRONG_TYPE")
    }

    tryCatch(func() {
        sqlDBType.SetMaxIdleConns(cfg.MaxIdleConns)
        sqlDBType.SetMaxOpenConns(cfg.MaxOpenConns)
        sqlDBType.SetConnMaxLifetime(time.Second * time.Duration(cfg.MaxLifetime))
    }, func(err error) {
        // 这里处理连接池配置过程中可能出现的错误
        _ = NewConnectionError("failed to set connection pool parameters", err).
            WithContext("max_idle_conns", cfg.MaxIdleConns).
            WithContext("max_open_conns", cfg.MaxOpenConns).
            WithContext("max_lifetime", cfg.MaxLifetime)
    })

    return nil
}

// tryCatch 简单的错误处理包装器
func tryCatch(fn func(), errorHandler func(error)) {
    defer func() {
        if r := recover(); r != nil {
            if err, ok := r.(error); ok {
                errorHandler(err)
            } else if errStr, ok := r.(string); ok {
                errorHandler(fmt.Errorf("panic: %s", errStr))
            } else {
                errorHandler(fmt.Errorf("panic: %v", r))
            }
        }
    }()
    fn()
}

// buildDSN 构建 ClickHouse 的 DSN 字符串。
// 形如：clickhouse://username:password@host:port/db?dial_timeout=10s&read_timeout=30s
func buildDSN(cfg *Config, dbName string) string {
    return fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?dial_timeout=10s&read_timeout=30s",
        cfg.Username,
        cfg.Password,
        cfg.Host,
        cfg.Port,
        dbName,
    )
}

// Ping 使用标准库连接执行 Ping 测试，确认数据库连通性。
// 调用方可传入超时时间的 Context 来控制最长等待时间。
func (d *DB) Ping(ctx context.Context) error {
    if d == nil || d.DB == nil {
        return NewConnectionError("database instance is nil", nil).
            WithCode("DB_NIL")
    }

    sqlDB, err := d.DB.DB()
    if err != nil {
        return NewConnectionError("failed to get underlying SQL database connection", err).
            WithCode("SQLDB_GET_FAILED")
    }

    // 执行 Ping 操作，并处理可能的超时
    err = sqlDB.PingContext(ctx)
    if err != nil {
        // 检查是否为上下文超时错误
        if ctx.Err() == context.DeadlineExceeded {
            return NewTimeoutError("database ping timed out", err).
                WithContext("timeout", ctx).
                WithCode("PING_TIMEOUT")
        }

        // 检查是否为上下文取消错误
        if ctx.Err() == context.Canceled {
            return NewConnectionError("database ping was canceled", err).
                WithContext("context", "canceled").
                WithCode("PING_CANCELED")
        }

        // 其他连接错误
        return NewConnectionError("database ping failed", err).
            WithCode("PING_FAILED")
    }

    return nil
}