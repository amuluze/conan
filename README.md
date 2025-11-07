# conan

一个包含多个子模块的 Go 项目：

- **[auther](./auther/)**：JWT 认证与令牌管理，支持黑名单、令牌撤销、刷新令牌旋转（防重放）等能力
- **[pg](./pg/)**：面向 GORM 的查询构造与单元测试（使用 SQLite DryRun 生成 SQL 进行断言）
- **[clickhouse](./clickhouse/)**：ClickHouse 查询构造、安全校验与 TLS 连接配置，并附带文档与测试示例
- **[sqlite](./sqlite/)**：SQLite 查询构造与测试辅助（用于 DryRun 与本地快速验证）
- **[redis](./redis/)**：Redis 客户端封装，提供常用数据结构操作与完善的单元测试

## 项目结构

```
conan/
├── auther/           # JWT 认证子模块
│   ├── README.md
│   └── ...
├── pg/               # GORM 查询构造子模块
│   ├── README.md
│   └── ...
├── clickhouse/       # ClickHouse 模块（查询、安全校验、TLS）
│   ├── README.md
│   └── ...
├── sqlite/           # SQLite 模块（查询构造与测试辅助）
│   ├── README.md
│   └── ...
├── redis/            # Redis 客户端与数据结构操作
│   ├── README.md
│   └── ...
├── LICENSE
└── README.md         # 本文档
```

## 快速开始

运行所有测试：

```bash
cd auther && go test ./...
cd ../pg && go test ./...
cd ../clickhouse && go test ./...
cd ../sqlite && go test ./...
cd ../redis && go test ./...
```

## 模块说明

### auther
JWT 认证与令牌管理模块，支持：
- 访问令牌与刷新令牌生成
- 令牌验证与过期处理
- 黑名单与令牌撤销
- 刷新令牌旋转（防重放）

详细文档：[auther/README.md](./auther/README.md)

### pg
面向 GORM 的查询构造辅助工具，支持：
- 查询构造器
- DryRun + SQLite 单元测试
- WHERE/ORDER/LIMIT/OFFSET/IN 等子句组合

详细文档：[pg/README.md](./pg/README.md)

### clickhouse
ClickHouse 查询与安全模块，支持：
- 规范化查询构造与输入校验（表名/字段名校验、SQL 关键字拦截）
- DSN 构建与 TLS 配置（含生产部署建议）
- 单元测试与 DryRun 示例（文档中包含说明）

详细文档：[clickhouse/README.md](./clickhouse/README.md)

### sqlite
SQLite 查询与测试辅助模块，支持：
- 轻量查询构造与 DryRun 校验
- 作为 GORM 测试驱动，便于生成 SQL 进行断言

详细文档：[sqlite/README.md](./sqlite/README.md)

### redis
Redis 客户端封装模块，支持：
- String/Hash/List/Set/ZSet 等常用数据结构操作
- 统一的选项与工具方法，配套单元测试

详细文档：[redis/README.md](./redis/README.md)

License

见仓库根目录 `LICENSE` 文件。
