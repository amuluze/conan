# pg

面向 GORM 的查询构造辅助工具与单元测试（使用 SQLite DryRun 生成 SQL 进行断言）。

模块路径：`github.com/amuluze/conan/pg`

## 功能说明

- 使用 GORM 的 DryRun 配合 SQLite 方言构造 SQL 并进行断言
- 覆盖 WHERE/ORDER/LIMIT/OFFSET/IN 等常见子句组合
- 适合参考如何将查询选项组合到 GORM 查询中

## 目录结构

```
pg/
├── query.go          # 查询构造器
├── query_test.go     # 单元测试（DryRun + SQLite）
├── config.go         # 连接配置
├── db.go             # 数据库初始化与封装
├── go.mod
├── go.sum
└── README.md         # 本文档
```

## 使用示例

详细的使用示例请参考 `query_test.go` 文件，其中包含了各种查询选项的组合用法。

### AutoMigrate 使用示例

```go
package main

import (
    "log"
    "github.com/amuluze/conan/pg"
)

// User 为示例模型，仅用于展示迁移用法
type User struct {
    ID       string `gorm:"primaryKey"`
    Username string
}

func main() {
    cfg := &pg.Config{
        Debug:        true,
        AutoMigrate:  true, // 开启自动迁移
        SSLMode:      "disable",
        Type:         "postgres",
        Host:         "localhost",
        Port:         "5432",
        Username:     "postgres",
        Password:     "password",
        DBName:       "demo",
        MaxLifetime:  300,
        MaxOpenConns: 100,
        MaxIdleConns: 100,
        TimeZone:     "Asia/Shanghai",
    }

    db, err := pg.NewDB(cfg)
    if err != nil {
        log.Fatalf("init db failed: %v", err)
    }

    // 在 autoMigrate=true 时执行迁移；未传模型或关闭标志则不做任何操作
    if err := db.AutoMigrate(&User{}); err != nil {
        log.Fatalf("auto migrate failed: %v", err)
    }
}
```

## 测试

```bash
cd pg && go test ./...
```

## 主要组件

- **query.go**: 提供查询构造器功能，支持各种 SQL 子句的组合
- **query_test.go**: 使用 SQLite DryRun 模式进行单元测试，验证生成的 SQL 语句
- **config.go**: 数据库连接配置管理
- **db.go**: 数据库初始化和基础封装

## 安全建议（表名白名单）

- 请使用 `WithTableSafe(tableName string, whitelist map[string]struct{})` 设置表名，并提供受控白名单：

```go
wl := map[string]struct{}{ "users": {}, "accounts": {} }
db = OptionDB(db, WithTableSafe("users", wl)) // 仅当 "users" 在白名单中时生效
```

- 旧函数 `WithTable` 已移除，请使用 `WithTableSafe` 以降低 SQL 注入风险。

## 清理状态说明

- 已移除旧接口 `WithTable`，统一通过 `WithTableSafe` 设置表名并进行白名单校验。
- 通过 `go vet` 与 `staticcheck` 检查当前模块未发现未使用的函数、变量或冗余导入。
- 若有对旧接口的历史引用，请按上述安全建议迁移；如需协助请联系维护者。

## 特性

- 支持 WHERE 条件构造
- 支持 ORDER BY 排序
- 支持 LIMIT 和 OFFSET 分页
- 支持 IN 查询
- 基于 GORM 框架，易于集成