# auther

JWT 认证与令牌管理模块，支持黑名单、令牌撤销、刷新令牌旋转（防重放）等能力。

模块路径：`github.com/amuluze/conan/auther`

## 功能概览

- 生成访问令牌与刷新令牌对（HS256）
- 验证令牌有效性与过期（稳定的过期判定逻辑）
 - 黑名单撤销令牌
- 刷新令牌旋转：使用成功后自动撤销旧刷新令牌并签发新令牌对，防止重放
- 显式关闭后台清理协程，避免资源泄漏

## 配置

配置选项：

- `AccessTokenExp`：访问令牌过期时间，默认 2h
- `RefreshTokenExp`：刷新令牌过期时间，默认 7d
- `Issuer`：令牌签发者（JWT `iss`），默认 "conan"
- `BlackListEnabled`：是否启用黑名单，默认 true
- `BlackListCleanupInterval`：黑名单清理间隔，默认 1h

## 示例代码

```go
package main

import (
    "context"
    "fmt"
    auther "github.com/amuluze/conan/auther"
    "time"
)

// demo 使用示例：创建认证器、生成令牌对、验证、刷新（旋转）与关闭资源。
func main() {
    // 1) 创建认证器
    cfg := auther.AutherConfig{
        SecretKey:                "your-secret",
        AccessTokenExp:           2 * time.Hour,
        RefreshTokenExp:          7 * 24 * time.Hour,
        Issuer:                   "conan-demo",
        BlackListEnabled:         true,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    a, err := auther.NewAuther(&cfg)
    if err != nil {
        panic(err)
    }

    // 2) 生成令牌对
    pair, err := a.GenerateTokenPair(context.Background(), "u1", "user1", "role1", map[string]string{"env": "dev"})
    if err != nil {
        panic(err)
    }
    fmt.Println("access:", pair.AccessToken.Token)
    fmt.Println("refresh:", pair.RefreshToken.Token)

    // 3) 验证访问令牌
    if claims, err := a.ValidateToken(context.Background(), pair.AccessToken.Token); err == nil {
        fmt.Println("validated user:", claims.UserID)
    }

    // 4) 刷新令牌（旋转）——推荐：返回新令牌对
    newPair, err := a.RefreshTokenRotate(context.Background(), pair.RefreshToken.Token)
    if err != nil {
        panic(err)
    }
    fmt.Println("new access:", newPair.AccessToken.Token)
    fmt.Println("new refresh:", newPair.RefreshToken.Token)

    // 5) 撤销令牌（黑名单启用）
    _ = a.RevokeToken(context.Background(), newPair.AccessToken.Token)

    // 6) 关闭后台清理协程，避免资源泄漏（接口已提供 Close 方法，重复调用安全）
    _ = a.Close()
}
```

## 注意事项

- **强制旋转**：业务层统一使用 `RefreshTokenRotate`，以保证统一且安全的刷新策略。
- **关闭资源**：如启用黑名单清理协程，请在应用退出时调用 Close，防止 goroutine 泄漏（库已实现安全的"重复关闭不 panic"）。
- **过期判断**：过期判断依赖 claims 的 `ExpiresAt` 字段而非错误字符串匹配，行为更稳定。

### 令牌的生效时间（NotBefore）与过期时间（ExpiresAt）

- 生效时间（NotBefore）：当当前时间早于令牌的 `NotBefore` 时，令牌将被视为“尚未生效”，`ValidateToken` 会返回 `ErrInvalidToken`。
- 过期时间（ExpiresAt）：当当前时间晚于 `ExpiresAt` 时，令牌将被视为“已过期”，`ValidateToken` 会返回 `ErrExpiredToken`。
- 稳定性：即使 `ParseWithClaims` 解析失败，模块也会使用 `ParseUnverified` 提取 claims 并进行上述两个时间判定，以提供稳定且可预期的错误类型。
- 签发策略：当前实现中，访问令牌与刷新令牌的 `NotBefore` 均设置为签发时刻（`time.Now()`），不支持设置“未来生效”的令牌；因此正常情况下不会遇到 NotBefore 未到的情况。
- 时钟同步建议：为避免由于机器时间偏差导致的误判，建议生产环境保持服务器与客户端的时钟同步（如配置 NTP）。

## 错误类型

- `ErrInvalidToken`：令牌无效
- `ErrExpiredToken`：令牌过期
- `ErrRevokedToken`：令牌已撤销
- `ErrInvalidTokenType`：令牌类型不符（例如用访问令牌执行刷新）
- `ErrSecretKeyEmpty`：密钥为空

## 测试

```bash
cd auther && go test ./...
```

## 目录结构

```
auther/
├── auther.go         # 认证器实现
├── auther_test.go    # 单元测试
├── blacklist.go      # 简易黑名单实现（内存）
├── types.go          # 接口与类型定义
├── go.mod            # 模块定义
├── go.sum
└── README.md         # 本文档
```
