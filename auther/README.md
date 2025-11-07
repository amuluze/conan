# auther

JWT 认证与令牌管理模块，基于 go-cache 管理令牌，支持令牌对（Access/Refresh）、刷新令牌旋转与会话级撤销（登出场景）。

模块路径：`github.com/amuluze/conan/auther`

## 功能概览

- 生成访问令牌与刷新令牌对（HS256），令牌对共享 SessionID
- 验证令牌：签名校验、存在性（白名单）与会话撤销标记检查
- 刷新令牌旋转：删除旧 Refresh，签发新令牌对（沿用 SessionID），防重放
- 生成单枚访问令牌：`MintAccessToken`
- 删除单枚令牌：`RemoveToken`
- 会话级撤销（登出）：`RemoveTokenPair`，仅有 AccessToken 也可撤销本会话下的 Refresh
- 查询令牌存在性：`HasToken`
- 从令牌提取信息（不验签）：`GetTokenInfo`
- 释放资源：`Release`

支持可插拔存储（Storage 接口），可替换为 Redis 等实现。

## 安装

```bash
go get github.com/amuluze/conan/auther
```

## 配置

配置结构：

```go
type Config struct {
    SecretKey       string
    AccessTokenExp  time.Duration // 默认 2h
    RefreshTokenExp time.Duration // 默认 24h
    Issuer          string        // 默认 "conan"
}
```

默认配置：

```go
var DefaultConfig = Config{
    AccessTokenExp:  2 * time.Hour,
    RefreshTokenExp: 24 * time.Hour,
    Issuer:          "conan",
}
```

## API 一览

```go
type Auther interface {
    // GenerateTokenPair 生成访问令牌和刷新令牌对
    GenerateTokenPair(ctx context.Context, userID, username, role string, metadata map[string]string) (*TokenPair, error)

    // MintAccessToken 生成访问令牌（仅限 AccessToken）
    MintAccessToken(ctx context.Context, userID, username, role string, exp time.Duration, metadata map[string]string) (*TokenInfo, error)

    // ValidateToken 验证令牌
    ValidateToken(ctx context.Context, token string) (*TokenClaims, error)

    // RefreshTokenRotate 刷新令牌旋转（删除旧 Refresh、沿用 SessionID 生成新令牌对）
    RefreshTokenRotate(ctx context.Context, refreshToken string) (*TokenPair, error)

    // RemoveToken 删除单枚令牌
    RemoveToken(ctx context.Context, token string) error

    // RemoveTokenPair 基于 AccessToken 或 SessionID 撤销整个会话
    RemoveTokenPair(ctx context.Context, sessionOrAccessToken string) error

    // HasToken 检查令牌是否存在
    HasToken(ctx context.Context, token string) bool

    // GetTokenInfo 从令牌中提取信息（不验证签名）
    GetTokenInfo(token string) (*TokenClaims, error)

    // Release 释放资源
    Release() error
}
```

## 使用示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    auther "github.com/amuluze/conan/auther"
)

// newAuther 示例：创建并返回一个 Auther 实例（包含 go-cache 存储）
// 说明：演示如何构建 Config、初始化默认存储并创建认证器。
func newAuther() (auther.Auther, error) { // 函数：构建并返回 Auther
    cfg := &auther.Config{ // 函数内：配置JWT密钥与基本信息
        SecretKey:       "your-secret",
        Issuer:          "conan-demo",
        AccessTokenExp:  2 * time.Hour,
        RefreshTokenExp: 24 * time.Hour,
    }
    store := auther.NewCacheStorage(cfg) // 函数内：使用配置创建默认存储
    return auther.NewAuther(cfg, store)  // 函数内：返回认证器
}

func main() {
    // 创建认证器
    // 说明：完整演示生成、验证、旋转与会话撤销流程。
    a, err := newAuther() // 函数：获取认证器实例
    if err != nil { panic(err) }

    // 生成令牌对
    pair, _ := a.GenerateTokenPair(context.Background(), "u1", "user1", "role1", map[string]string{"env": "dev"}) // 函数：生成令牌对
    fmt.Println("access:", pair.AccessToken.Token)
    fmt.Println("refresh:", pair.RefreshToken.Token)

    // 验证访问令牌
    if claims, err := a.ValidateToken(context.Background(), pair.AccessToken.Token); err == nil { // 函数：验证令牌并解析Claims
        fmt.Println("validated user:", claims.UserID, "sid:", claims.SessionID)
    }

    // 刷新令牌旋转
    newPair, _ := a.RefreshTokenRotate(context.Background(), pair.RefreshToken.Token) // 函数：旋转刷新令牌
    fmt.Println("new access:", newPair.AccessToken.Token)
    fmt.Println("new refresh:", newPair.RefreshToken.Token)

    // 单枚删除
    _ = a.RemoveToken(context.Background(), newPair.AccessToken.Token) // 函数：删除单枚令牌

    // 会话级撤销（登出）：仅有 AccessToken 即可撤销整个会话
    _ = a.RemoveTokenPair(context.Background(), pair.AccessToken.Token) // 函数：基于Access撤销整个会话

    _ = a.Release() // 函数：释放资源
}
```

## 注意事项

- 会话撤销标记：`RemoveTokenPair` 会设置 `sid:<SessionID>:revoked` 标记，`ValidateToken` 检查到该标记即拒绝同会话下所有令牌。
- 刷新令牌旋转：`RefreshTokenRotate` 会先删除旧 Refresh，再签发新令牌对，防止重放。
- 令牌白名单：模块使用存储的存在性检查作为白名单，删除或过期后验证失败。
- 存储键前缀：默认存储会以 `Issuer` 作为键前缀，例如撤销标记实际写入为 `<Issuer>:sid:<SessionID>:revoked`。
- 安全建议：请妥善管理 `SecretKey`（建议从环境变量加载），并为不同环境设置不同的 `Issuer`。避免将令牌暴露在 URL 等日志中。

## 可插拔存储（Storage 接口）

```go
type Storage interface {
    Set(tokenString string, expiration time.Duration) error
    Check(tokenString string) (bool, error)
    Delete(tokenString string) error
    Close() error
}
```

默认实现：`CacheStorage`（基于 `github.com/patrickmn/go-cache`）。

使用默认存储：

```go
// NewCacheStorage 使用传入的 Config 初始化 go-cache 存储
// 说明：以 Config.AccessTokenExp 作为默认过期时长，并使用 Config.Issuer 作为键前缀。
store := auther.NewCacheStorage(cfg)
```

## 错误类型

- `ErrInvalidToken`：令牌无效
- `ErrExpiredToken`：令牌过期
- `ErrInvalidTokenType`：令牌类型不符（例如用访问令牌执行刷新）
- `ErrSecretKeyEmpty`：密钥为空

## 测试

```bash
cd auther && go test ./...
```

## 目录结构

```
auther/
├── auther.go      # 认证器实现
├── auther_test.go # 单元测试
├── types.go       # 接口与类型定义
├── storage.go     # 默认存储实现（go-cache）
├── go.mod         # 模块定义
├── go.sum
└── README.md      # 本文档
```
