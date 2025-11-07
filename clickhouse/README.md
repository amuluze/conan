# clickhouse

面向 GORM 的 ClickHouse 查询构造辅助工具与单元测试（使用 ClickHouse DryRun 生成 SQL 进行断言）。

模块路径：`github.com/amuluze/conan/clickhouse`

## 功能说明

- 使用 GORM 的 DryRun 配合 ClickHouse 方言构造 SQL 并进行断言
- 覆盖 WHERE/ORDER/LIMIT/OFFSET/IN 等常见子句组合
- 适合参考如何将查询选项组合到 GORM 查询中
- 针对 ClickHouse 数据库特性进行了适配
- **详细的错误分类和处理机制**
- **安全的参数验证和 SQL 注入防护**

## 目录结构

```
clickhouse/
├── errors.go         # 错误类型和分类系统
├── errors_test.go    # 错误处理测试
├── query.go          # 查询构造器（增强错误处理）
├── query_test.go     # 单元测试（DryRun + ClickHouse + 错误处理）
├── config.go         # 连接配置（包含验证逻辑）
├── db.go             # 数据库初始化与封装（增强错误处理）
├── db_test.go        # 数据库功能测试
├── go.mod
├── go.sum
└── README.md         # 本文档
```

## 使用示例

详细的使用示例请参考 `query_test.go` 文件，其中包含了各种查询选项的组合用法。

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/amuluze/conan/clickhouse"
)

func main() {
    // 创建配置
    config := &clickhouse.Config{
        Host:     "localhost",
        Port:     "9000",
        Username: "default",
        Password: "",
        DBName:   "test",
        Debug:    true,
    }

    // 创建数据库连接
    db, err := clickhouse.NewDB(config)
    if err != nil {
        // 处理错误
        var chErr *clickhouse.ClickHouseError
        if clickhouse.IsConfigError(err) {
            log.Printf("配置错误: %v", err)
        } else if clickhouse.IsConnectionError(err) {
            log.Printf("连接错误: %v", err)
        } else {
            log.Printf("未知错误: %v", err)
        }
        return
    }

    // 使用新的错误处理版本的 OptionDB
    resultDB, err := clickhouse.OptionDB(db,
        clickhouse.WithTable("users"),
        clickhouse.WithId("123"),
        clickhouse.WithUserName("john"),
    )
    if err != nil {
        log.Printf("查询构造错误: %v", err)
        return
    }

    // 测试连接
    ctx := context.Background()
    if err := db.Ping(ctx); err != nil {
        if clickhouse.IsTimeoutError(err) {
            log.Printf("连接超时: %v", err)
        } else {
            log.Printf("连接失败: %v", err)
        }
        return
    }

    fmt.Println("数据库连接和查询构造成功")
}
```

### 错误处理

```go
// 错误类型判断
db, err := clickhouse.NewDB(config)
if err != nil {
    switch {
    case clickhouse.IsConnectionError(err):
        // 连接相关错误，可以重试
        if clickhouse.IsRetriableError(err) {
            log.Println("连接错误，可重试")
        }
    case clickhouse.IsConfigError(err):
        // 配置错误，需要修正配置
        log.Println("配置错误，请检查配置参数")
    case clickhouse.IsValidationError(err):
        // 参数验证错误
        log.Println("参数验证失败")
    case clickhouse.IsTimeoutError(err):
        // 超时错误，可以重试
        log.Println("操作超时")
    default:
        log.Printf("未知错误: %v", err)
    }
}

// 获取详细错误信息
var chErr *clickhouse.ClickHouseError
if errors.As(err, &chErr) {
    log.Printf("错误类型: %s", chErr.Type)
    log.Printf("错误代码: %s", chErr.Code)
    log.Printf("是否可重试: %v", chErr.Retriable)
    log.Printf("上下文信息: %+v", chErr.Context)
}
```

## 测试

```bash
cd clickhouse && go test ./...
```

### 运行特定测试

```bash
# 运行所有错误处理相关测试
go test -v -run "Error"

# 运行数据库连接测试
go test -v -run "TestNewDB"

# 运行查询构造测试
go test -v -run "TestOptionDB"
```

## 主要组件

- **errors.go**: 错误类型定义和分类系统
- **errors_test.go**: 错误处理测试
- **query.go**: 查询构造器（增强错误处理和参数验证）
- **query_test.go**: 查询构造器测试（包含错误处理场景）
- **config.go**: 数据库连接配置（包含验证逻辑）
- **db.go**: 数据库初始化和连接管理（增强错误处理）
- **db_test.go**: 数据库功能测试

## 特性

### 查询构造
- 支持 WHERE 条件构造
- 支持 ORDER BY 排序
- 支持 LIMIT 和 OFFSET 分页
- 支持 IN 查询
- 基于 GORM 框架，易于集成
- 针对 ClickHouse 进行优化适配

### 错误处理 ⭐ 新增
- **详细的错误分类**: 连接、配置、查询、验证、超时、TLS 等类型
- **错误上下文信息**: 包含详细的上下文数据便于调试
- **重试机制指示**: 自动标识哪些错误可以重试
- **安全验证**: 防止 SQL 注入和无效参数
- **错误包装**: 支持 errors.Is 和 errors.As

### 安全增强 ⭐ 新增
- **参数验证**: 表名、字段名、ID 等参数的严格验证
- **SQL 注入防护**: 白名单机制和参数化查询
- **长度限制**: 防止过长参数导致的性能问题
- **特殊字符过滤**: 防止控制字符和恶意输入

## 配置说明

ClickHouse 配置相比 PostgreSQL 有以下区别：

- 默认端口为 9000（TCP 协议）
- 支持连接池配置
- 兼容标准库驱动和原生驱动
- 支持 ClickHouse 特有的设置项

### 配置验证

新的配置验证会在创建连接时自动检查：

```go
config := &clickhouse.Config{
    Host:     "localhost",
    Port:     "9000",        // 自动验证端口范围
    Username: "default",     // 必填项验证
    DBName:   "test",        // 必填项验证
    SSLMode:  "disable",     // 验证 SSL 模式
}

// Validate() 方法会自动调用，或在需要时手动调用
if err := config.Validate(); err != nil {
    log.Printf("配置验证失败: %v", err)
}
```

## 依赖

- github.com/ClickHouse/clickhouse-go/v2: ClickHouse 官方驱动
- gorm.io/driver/clickhouse: GORM ClickHouse 驱动
- gorm.io/gorm: ORM 框架

## 错误类型

| 错误类型 | 描述 | 可重试 | 示例场景 |
|---------|------|--------|----------|
| `connection` | 数据库连接相关错误 | ✅ | 连接超时、网络中断 |
| `config` | 配置参数错误 | ❌ | 无效主机名、端口范围错误 |
| `query` | SQL 查询执行错误 | ⚠️ | 语法错误、字段不存在 |
| `validation` | 参数验证错误 | ❌ | 无效表名、SQL 注入尝试 |
| `timeout` | 操作超时错误 | ✅ | 查询超时、连接超时 |
| `tls` | TLS/SSL 连接错误 | ❌ | 证书验证失败 |

## 向后兼容性

为了保持向后兼容，提供了 `OptionDBMust` 函数，行为与原来的 `OptionDB` 相同：

```go
// 新版本（推荐）- 返回错误
resultDB, err := clickhouse.OptionDB(db, clickhouse.WithTable("users"))

// 向后兼容 - 不返回错误
resultDB := clickhouse.OptionDBMust(db, clickhouse.WithTable("users"))
```

## 测试说明与 DryRun 机制澄清

本项目的单元测试使用的是「sqlite 内存驱动 + GORM 的 DryRun」来构造 SQL 并做断言。这么做的目的是：
- 不依赖真实 ClickHouse 服务也能验证查询构造逻辑；
- 通过 DryRun 获取最终的 SQL 文本与绑定参数，以确保查询选项组合正确。

在生产环境或集成测试中，如果需要基于 ClickHouse Dialector 验证 SQL 构造，可以使用如下方式：

```go
package main

import (
    "log"
    gormclickhouse "gorm.io/driver/clickhouse"
    "gorm.io/gorm"
)

// newCHDryRun 测试示例：使用 ClickHouse Dialector 并开启 DryRun。
// 函数目的：演示如何在不执行真实查询的情况下，用 ClickHouse 方言构造 SQL。
func newCHDryRun() (*gorm.DB, error) {
    dsn := "clickhouse://user:pass@127.0.0.1:9000/db?dial_timeout=10s&read_timeout=30s"
    gdb, err := gorm.Open(gormclickhouse.Open(dsn), &gorm.Config{DryRun: true})
    if err != nil {
        return nil, err
    }
    return gdb, nil
}

func main() {
    gdb, err := newCHDryRun()
    if err != nil { log.Fatal(err) }
    // 在此处使用 gdb 构造查询并读取 gdb.Statement.SQL 进行断言
}
```

注意：DryRun 模式仅构造 SQL，不会实际执行查询或建立真实的网络连接。

## 安全最佳实践

为防止 SQL 注入和非法参数输入，建议遵循以下实践：

- 表名与字段名白名单
  - 通过预先维护白名单，限制可用的表和字段，避免动态字符串拼接被利用。
  - OrderAsc/OrderDesc 已内置字段名白名单校验；表名可在使用 WithTable 前做额外白名单检查。

```go
package main

import (
    "gorm.io/gorm"
    "github.com/amuluze/conan/clickhouse"
)

// isAllowedTable 检查表名是否在白名单中。
// 函数目的：在调用 WithTable 之前进行业务侧的白名单拦截。
func isAllowedTable(tbl string, wl map[string]struct{}) bool {
    _, ok := wl[tbl]
    return ok
}

func buildSafeQuery(db *clickhouse.DB) (*gorm.DB, error) {
    tableWL := map[string]struct{}{ "users": {}, "accounts": {} }
    fieldWL := map[string]struct{}{ "id": {}, "created_at": {}, "username": {} }

    tbl := "users"
    if !isAllowedTable(tbl, tableWL) {
        // 非白名单表，忽略或返回错误
        return db.DB, nil
    }

    // 字段排序使用白名单校验的 OrderAsc/OrderDesc
    return clickhouse.OptionDB(db,
        clickhouse.WithTable(tbl),
        clickhouse.OrderAsc("created_at", fieldWL),
    )
}
```

- 参数化查询优先
  - IN 条件优先使用安全列封装：`WithIds`/`WithNames`/`WithUsernames`；通用 `WithIn(field IN ?)` 使用前应确保字段名来自白名单。

```go
package main

import (
    "github.com/amuluze/conan/clickhouse"
)

// buildInQuery 构造 IN 查询示例。
// 函数目的：展示在使用通用 WithIn 前先做字段白名单校验，或直接使用安全列的 IN 辅助函数。
func buildInQuery(db *clickhouse.DB, ids []string) (*clickhouse.DB, error) {
    // 直接使用安全列辅助函数（推荐）
    return clickhouse.OptionDB(db,
        clickhouse.WithTable("users"),
        clickhouse.WithIds(ids),
    )
}
```

- 输入长度限制与控制字符过滤
  - WithId/WithUserName/WithName 内部已对长度和控制字符进行校验；建议上游也进行基本过滤。

- 错误处理与审计
  - 使用 `WrapError` 与 `IsValidationError` 等进行错误类型判断。
  - 记录被拒绝的表/字段名，便于审计与安全监控。

## 性能调优示例

ClickHouse 适合高吞吐查询与写入，建议结合连接池与设置项进行调优：

- 连接池参数（Config）
  - 根据负载调整：MaxOpenConns、MaxIdleConns、MaxLifetime。示例：

```go
// newDBWithPool 示例：演示连接池参数设置。
// 函数目的：提供生产环境下常见的连接池配置参考。
func newDBWithPool() (*clickhouse.DB, error) {
    cfg := &clickhouse.Config{
        Host:         "127.0.0.1",
        Port:         "9000",
        Username:     "default",
        Password:     "",
        DBName:       "test",
        MaxOpenConns: 200,
        MaxIdleConns: 100,
        MaxLifetime:  600, // 秒
    }
    return clickhouse.NewDB(cfg)
}
```

- 使用原生驱动 OpenDB 并设置 Settings（全局）
  - 在 `Config.OpenDB = true` 时，可通过 `ch.Options.Settings` 设置全局参数，例如 `max_execution_time`。

```go
// newDBWithSettings 示例：演示通过 OpenDB 设置 ClickHouse 全局参数。
// 函数目的：展示如何在驱动层配置执行与连接超时等设置项。
func newDBWithSettings() (*clickhouse.DB, error) {
    cfg := &clickhouse.Config{
        Host:    "127.0.0.1",
        Port:    "9000",
        Username:"default",
        DBName:  "test",
        OpenDB:  true,
        // 其他连接池参数略
    }
    return clickhouse.NewDB(cfg)
}
```

- 每查询设置（建议作为扩展）
  - 某些场景需为单次查询设置 `max_execution_time`、`max_threads` 等，建议封装专用 QueryOption，或在执行前通过 Session/Clause 注入配置（依据 GORM ClickHouse Dialector 的支持情况进行实现）。

## TLS 部署指南

根据 `Config.SSLMode` 的不同，TLS 配置与安全强度各不相同：

- disable：不启用 TLS（开发测试环境使用）。
- require：启用 TLS，但不校验证书（快速接入，存在中间人风险）。
- verify-ca：启用 TLS，校验证书链但不校验主机名。
- verify-full：启用 TLS，校验证书链与主机名（生产推荐）。

建议的部署步骤：
1. 准备服务端证书与私钥，确保由受信任的 CA 颁发；
2. 客户端安装 CA 根证书或指定信任链；
3. 当使用 `verify-full` 时，确保证书的 CN/SAN 包含访问的主机名，`Config.Host` 与证书的 `ServerName` 匹配；
4. 在生产环境避免使用 `InsecureSkipVerify: true` 的模式（即 require），防止被中间人攻击；
5. 避免在日志中输出敏感信息（如密码、完整 DSN）。

示例（verify-full）：

```go
// newSecureDB 示例：演示开启 verify-full 模式的 TLS 配置。
// 函数目的：展示在生产环境中使用严格的主机名校验与证书链验证。
func newSecureDB() (*clickhouse.DB, error) {
    cfg := &clickhouse.Config{
        Host:    "db.prod.example.com",
        Port:    "9000",
        Username:"default",
        Password:"",
        DBName:  "prod",
        SSLMode: "verify-full",
    }
    return clickhouse.NewDB(cfg)
}
```

常见问题：
- 连接报错 “TLS_UNSUPPORTED_MODE”：检查 `SSLMode` 是否为支持的枚举值。
- 连接报错 “TLS_HOST_REQUIRED”：在 `verify-full` 下确保 `Host` 非空且与证书匹配。
- 本地开发：若仅验证功能，建议使用 `disable` 或自签证书 + 信任根链的方式。