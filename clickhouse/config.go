package clickhouse

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Validate 验证配置参数的有效性
func (c *Config) Validate() error {
	if c == nil {
		return NewConfigError("config cannot be nil", nil)
	}

	var errs []string

	// 验证主机名
	if c.Host == "" {
		errs = append(errs, "host cannot be empty")
	} else {
		// 检查主机名格式
		if net.ParseIP(c.Host) == nil && !isValidHostname(c.Host) {
			errs = append(errs, fmt.Sprintf("invalid host format: %s", c.Host))
		}
	}

	// 验证端口
	if c.Port == "" {
		c.Port = "9000" // 设置默认值
	} else {
		port, err := strconv.Atoi(c.Port)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid port format: %s", c.Port))
		} else if port < 1 || port > 65535 {
			errs = append(errs, fmt.Sprintf("port must be between 1 and 65535, got: %d", port))
		}
	}

	// 验证用户名
	if strings.TrimSpace(c.Username) == "" {
		errs = append(errs, "username cannot be empty")
	}

	// 验证数据库名称
	dbName := c.DBName
	if dbName == "" {
		dbName = c.Database
	}
	if strings.TrimSpace(dbName) == "" {
		errs = append(errs, "database name cannot be empty")
	}

	// 验证 SSL 模式
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full", ""}
	if c.SSLMode != "" && !containsValidMode(c.SSLMode, validSSLModes) {
		errs = append(errs, fmt.Sprintf("invalid SSL mode: %s, valid modes: %v", c.SSLMode, validSSLModes))
	}

	// 验证连接池参数
	if c.MaxLifetime <= 0 {
		c.MaxLifetime = 300 // 默认5分钟
	} else if c.MaxLifetime > 3600 {
		errs = append(errs, "MaxLifetime should not exceed 3600 seconds (1 hour)")
	}

	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 100 // 默认值
	} else if c.MaxOpenConns > 1000 {
		errs = append(errs, "MaxOpenConns should not exceed 1000")
	}

	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 100 // 默认值
	} else if c.MaxIdleConns > c.MaxOpenConns {
		errs = append(errs, "MaxIdleConns should not be greater than MaxOpenConns")
	}

	if len(errs) > 0 {
		return NewConfigError(fmt.Sprintf("config validation failed: %s", strings.Join(errs, "; ")), nil).
			WithContext("host", c.Host).
			WithContext("port", c.Port).
			WithContext("database", dbName)
	}

	// 确保数据库名称一致
	if c.DBName == "" {
		c.DBName = c.Database
	} else {
		c.Database = c.DBName
	}

	return nil
}

// isValidHostname 检查是否为有效的主机名
func isValidHostname(host string) bool {
	if len(host) > 253 {
		return false
	}

	// 简单的主机名验证
	host = strings.Trim(host, " ")
	if host == "" {
		return false
	}

	// 检查是否包含有效字符
	for _, char := range host {
		if !((char >= 'a' && char <= 'z') ||
			 (char >= 'A' && char <= 'Z') ||
			 (char >= '0' && char <= '9') ||
			 char == '-' || char == '.') {
			return false
		}
	}

	return true
}

// containsValidMode 检查SSL模式是否有效
func containsValidMode(mode string, validModes []string) bool {
	for _, validMode := range validModes {
		if mode == validMode {
			return true
		}
	}
	return false
}

// GetDSNInfo 获取DSN信息（不包含敏感信息）
func (c *Config) GetDSNInfo() map[string]interface{} {
	if c == nil {
		return nil
	}

	return map[string]interface{}{
		"host":           c.Host,
		"port":           c.Port,
		"database":       c.DBName,
		"username":       c.Username,
		"ssl_mode":       c.SSLMode,
		"max_lifetime":   c.MaxLifetime,
		"max_open_conns": c.MaxOpenConns,
		"max_idle_conns": c.MaxIdleConns,
		"use_native_db":  c.OpenDB,
	}
}

type Config struct {
	Debug        bool   // 是否开启调试模式，默认 false
	AutoMigrate  bool   // 是否自动迁移数据库结构，默认 false
	SSLMode      string // disable, require, verify-ca, verify-full
	Type         string // 数据库类型，默认 clickhouse
	Host         string // 数据库主机，默认 localhost
	Port         string // 数据库端口，默认 9000
	Username     string // 数据库用户名
	Password     string // 数据库密码
	DBName       string // 数据库名称
	MaxLifetime  int    // 连接最大生命周期，默认 5 分钟
	MaxOpenConns int    // 最大打开连接数，默认 100
	MaxIdleConns int    // 最大空闲连接数，默认 100
	Database     string // 数据库名称，兼容性字段
	OpenDB       bool   // 是否使用标准库数据库驱动，默认 false
}