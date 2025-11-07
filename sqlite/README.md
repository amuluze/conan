# sqlite

面向 GORM 的查询构造辅助工具与单元测试（使用 SQLite DryRun 生成 SQL 进行断言）。

模块路径：`github.com/amuluze/conan/sqlite`

## 功能说明

- 使用 GORM 的 DryRun 配合 SQLite 方言构造 SQL 并进行断言
- 覆盖 WHERE/ORDER/LIMIT/OFFSET/IN 等常见子句组合
- 适合参考如何将查询选项组合到 GORM 查询中
- 专为 SQLite 数据库优化配置

## 目录结构

```
sqlite/
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

## 配置说明

```go
config := &sqlite.Config{
    Debug:        true,           // 开启调试模式
    AutoMigrate:  false,          // 自动迁移数据库结构
    DatabasePath: "app.db",       // 数据库文件路径，默认 :memory:
    MaxLifetime:  300,            // 连接最大生命周期（秒）
    MaxOpenConns: 100,            // 最大打开连接数
    MaxIdleConns: 100,            // 最大空闲连接数
    BusyTimeout:  5,              // SQLite 忙等待超时时间（秒）
}
```

## 测试

```bash
cd sqlite && go test ./...
```

## 主要组件

- **query.go**: 提供查询构造器功能，支持各种 SQL 子句的组合
- **query_test.go**: 使用 SQLite DryRun 模式进行单元测试，验证生成的 SQL 语句
- **config.go**: 数据库连接配置管理，针对 SQLite 特性优化
- **db.go**: 数据库初始化和基础封装

## 特性

- 支持 WHERE 条件构造
- 支持 ORDER BY 排序
- 支持 LIMIT 和 OFFSET 分页
- 支持 IN 查询
- 基于 GORM 框架，易于集成
- SQLite 专用配置和优化
- 支持内存数据库和文件数据库模式

## 与 pg 模块的区别

- 使用 SQLite 驱动
- 配置项针对 SQLite 特性进行调整（如数据库路径、忙等待超时）
- 简化的连接字符串配置
- 适合嵌入式应用或轻量级数据库需求
