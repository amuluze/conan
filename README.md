# conan

一个包含两个子模块的 Go 项目：

- **[auther](./auther/)**：JWT 认证与令牌管理，支持黑名单、令牌撤销、刷新令牌旋转（防重放）等能力
- **[pg](./pg/)**：面向 GORM 的查询构造辅助工具与单元测试（使用 SQLite DryRun 生成 SQL 进行断言）

## 项目结构

```
conan/
├── auther/           # JWT 认证子模块
│   ├── README.md     # auther 详细文档
│   └── ...
├── pg/               # GORM 查询构造子模块
│   ├── README.md     # pg 详细文档
│   └── ├── ...
├── LICENSE
└── README.md         # 本文档
```

## 快速开始

运行所有测试：

```bash
cd auther && go test ./...
cd ../pg && go test ./...
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

License

见仓库根目录 `LICENSE` 文件。
golang tools
