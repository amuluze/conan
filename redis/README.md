# Redis

Redis 客户端封装库，基于 go-redis v9 构建，支持单机、集群和哨兵模式。

模块路径：`github.com/amuluze/conan/redis`

## 功能特性

- ✅ **多模式支持**: 单机、集群、哨兵模式
- ✅ **配置灵活**: 支持多种连接选项配置
- ✅ **错误处理**: 完善的错误处理和验证
- ✅ **测试友好**: 支持无连接测试和Mock
- ✅ **版本更新**: 升级到 go-redis v9 最新稳定版
- ✅ **类型安全**: 统一的API接口，类型安全
- ✅ **工具函数**: 提供常用的Redis操作工具

## 主要修复内容

### 1. 依赖升级
- 使用 `github.com/redis/go-redis/v9` 最新稳定版
- 适配 v9 API变化（如 `SetEX` → `SetEx`）
- 移除已废弃的 `IdleTimeout` 配置

### 2. 测试改进
- 修复测试代码中的panic问题，改用 `t.Errorf`/`t.Fatalf`
- 添加跳过机制，避免Redis连接失败导致测试失败
- 新增配置验证测试
- 增加工具函数测试

### 3. 代码增强
- 添加 `NewClientWithoutPing` 支持离线测试
- 增强配置验证逻辑
- 修复类型不一致问题（`HGet` 返回类型）
- 添加更多实用工具函数

### 4. 错误处理
- 完善nil指针检查
- 增加更详细的错误信息
- 优化连接失败处理

## 安装

```bash
go get github.com/amuluze/conan/redis
```

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "log"

    "github.com/amuluze/conan/redis"
)

func main() {
    // 创建客户端
    rc, err := redis.NewClient()
    if err != nil {
        log.Fatalf("Failed to create redis client: %v", err)
    }
    defer rc.Close()

    // 设置键值
    err = rc.SetWithExpiration("mykey", "myvalue", time.Hour)
    if err != nil {
        log.Printf("Failed to set key: %v", err)
        return
    }

    // 获取值
    value, err := rc.Get("mykey")
    if err != nil {
        log.Printf("Failed to get key: %v", err)
        return
    }

    fmt.Printf("Value: %s\n", value)
}
```

### 自定义配置

```go
rc, err := redis.NewClient(
    redis.WithAddrs([]string{"localhost:6379", "localhost:6380"}),
    redis.WithPassword("yourpassword"),
    redis.WithDB(1),
    redis.WithPoolSize(100),
    redis.WithConnectionTimeout("10s"),
    redis.WithReadTimeout("5s"),
    redis.WithWriteTimeout("5s"),
)
```

### 哨兵模式

```go
rc, err := redis.NewClient(
    redis.WithAddrs([]string{"sentinel1:26379", "sentinel2:26379"}),
    redis.WithMasterName("mymaster"),
    redis.WithPassword("yourpassword"),
)
```

### 测试环境

```go
// 在测试环境中使用不进行连接检查的客户端
rc, err := redis.NewClientWithoutPing(redis.WithAddrs([]string{"localhost:9999"}))
```

### 上下文（Context）支持

为所有核心方法新增了带上下文的 *Ctx 版本，并提供 `WithContext(ctx)` 封装，方便统一传递取消与超时。

示例：直接使用 *Ctx 方法

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

// 设置带过期时间
_, _ = rc.SetEXCtx(ctx, "k", "v", time.Minute)

// 获取键值
_, _ = rc.GetCtx(ctx, "k")

// 扫描键（替代 KEYS）
_, _ = rc.ScanKeysCtx(ctx, "user:*", 1000)
```

示例：使用 WithContext 封装统一上下文

```go
ctx := context.Background()
w := rc.WithContext(ctx)

// 使用与原 API 相同的方法名，自动使用默认 ctx
_, _ = w.SetEX("k", "v", time.Minute)
_, _ = w.Get("k")
_, _ = w.ScanKeys("user:*", 1000)
```

## API文档

### 基础操作

```go
// 字符串操作
rc.Set("key", "value")
rc.SetEX("key", "value", time.Hour)      // 带过期时间设置（底层调用 SetEx）
rc.SetNX("key", "value", time.Hour)      // 不存在才设置
rc.Get("key")
rc.Incr("counter")
rc.IncrBy("counter", 10)

// 哈希操作
rc.HSet("hash", "field1", "value1", "field2", "value2") // 或 rc.HSetMap("hash", map[string]interface{"field1":"value1"})
rc.HGet("hash", "field1")
rc.HGetAll("hash")
rc.HDel("hash", "field1")

// 列表操作
rc.LPush("list", "item1", "item2")
rc.RPop("list")
rc.LRange("list", 0, -1)

// 集合操作
rc.SAdd("set", "member1", "member2")
rc.SMembers("set")
rc.SIsMember("set", "member1")

// 有序集合操作
rc.ZAdd("zset", "member1", 100.0)
rc.ZRange("zset", 0, -1)
rc.ZRank("zset", "member1")
```

### 工具函数

```go
// 连接测试
err := rc.PingRedis()
available := rc.IsRedisAvailable()

// 批量操作
rc.BatchDelete([]string{"key1", "key2"})
rc.BatchExpire([]string{"key1", "key2"}, time.Hour)

// 原子操作
count, err := rc.IncrementAtomic("counter", 1)

// 获取或设置
value, err := rc.GetOrSet("key", "default_value")
```

## 配置选项

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `WithAddrs` | `[]string` | `["localhost:6379"]` | Redis服务器地址 |
| `WithPassword` | `string` | `""` | Redis密码 |
| `WithDB` | `int` | `0` | 数据库编号 (0-15) |
| `WithPoolSize` | `int` | `50` | 连接池大小 |
| `WithMasterName` | `string` | `""` | 哨兵模式下主节点名称 |
| `WithConnectionTimeout` | `string` | `"5s"` | 连接超时 |
| `WithReadTimeout` | `string` | `"3s"` | 读取超时 |
| `WithWriteTimeout` | `string` | `"3s"` | 写入超时 |
（已移除）

### 上下文版本（以下均提供 *Ctx 变体）

- 字符串：`SetCtx`、`SetEXCtx`、`SetNXCtx`、`GetCtx`、`GetRangeCtx`、`IncrCtx`、`IncrByCtx`、`DecrCtx`、`DecrByCtx`、`AppendCtx`、`StrLenCtx`
- 哈希：`HSetCtx`、`HSetMapCtx`、`HGetCtx`、`HGetAllCtx`、`HDelCtx`、`HExistsCtx`、`HLenCtx`
- 集合：`SAddCtx`、`SPopCtx`、`SRemCtx`、`SMembersCtx`、`SIsMemberCtx`、`SCardCtx`、`SUnionCtx`、`SDiffCtx`、`SInterCtx`
- 有序集合：`ZAddCtx`、`ZIncrByCtx`、`ZRangeCtx`、`ZRevRangeCtx`、`ZRangeByScoreCtx`、`ZRevRangeByScoreCtx`、`ZCardCtx`、`ZCountCtx`、`ZScoreCtx`、`ZRankCtx`、`ZRevRankCtx`、`ZRemCtx`、`ZRemRangeByRankCtx`、`ZRemRangeByScoreCtx`
- 通用：`ScanKeysCtx`、`TypeCtx`、`DeleteCtx`、`ExistsCtx`、`ExpireCtx`、`ExpireAtCtx`、`TTLCtx`、`PTTLCtx`、`DBSizeCtx`、`FlushDBCtx`、`FlushAllCtx`

## 测试

```bash
cd redis && go test -v
```

所有测试都设计为可以在没有Redis实例的情况下运行，会跳过需要实际连接的测试。

## 目录结构

```
redis/
├── client.go          # 客户端创建和基础配置
├── common.go          # 基础Redis操作（提供 ScanKeys 替代 KEYS）
├── string.go          # 字符串类型操作
├── hash.go            # 哈希类型操作
├── list.go            # 列表类型操作
├── set.go             # 集合类型操作
├── zset.go            # 有序集合操作
├── option.go          # 配置选项定义
├── utils.go           # 工具函数
├── client_test.go     # 客户端测试
├── common_test.go     # 基础操作测试
├── utils_test.go      # 工具函数测试
├── hash_test.go       # 哈希操作测试
├── list_test.go       # 列表操作测试
├── set_test.go        # 集合操作测试
├── string_test.go     # 字符串操作测试
├── zset_test.go       # 有序集合操作测试
├── go.mod
├── go.sum
└── README.md          # 本文档
```

## 注意事项

1. **版本兼容性**: 使用 go-redis v9，API与v8有部分变化
2. **配置验证**: 客户端创建时会验证配置的有效性
3. **连接池**: 建议根据应用需求调整连接池大小
4. **超时设置**: 生产环境中建议设置合理的超时时间
5. **错误处理**: 所有操作都应该检查错误返回值

## 最佳实践

以下最佳实践基于实际生产场景总结，结合本库提供的 *Ctx 方法与 `WithContext(ctx)` 封装，帮助你更稳健地使用 Redis。

- 统一传递上下文：服务入口生成 `context.Context`，业务层统一使用 `WithContext(ctx)` 或显式 *Ctx 方法，确保超时与取消能正确传播。
- 键空间遍历使用 SCAN：避免使用 KEYS 全盘扫描带来的阻塞，使用 `ScanKeys`/`ScanKeysCtx` 进行分页扫描。
- 区分不存在错误：使用 `github.com/redis/go-redis/v9` 的 `Nil` 常量来区分“键不存在”与其他错误。
- 合理设置过期：优先在写入时设置 TTL（如 `SetEX`、`SetNX` 带过期），避免无过期的大键与僵尸键。
- 连接池与超时：根据并发与延迟调整 PoolSize、Read/WriteTimeout，服务端口高并发下建议加大 PoolSize 并设置超时保护。
- 键命名规范：使用分层命名 `app:{env}:domain:{id}`，方便隔离与扫描；避免过长的 key 与 value。
- 性能优化：批量写入使用 Pipeline；热点数据优先使用字符串与哈希；避免大列表一次性全量读取。
- 原子与并发：分布式锁使用 `SET NX + EX`（本库用 `SetNX` + 过期即可），释放锁使用 Lua 保证只释放自己的锁；限流可用 `INCR` + `EX`。
- 测试策略：使用 `NewClientWithoutPing` 与 *Ctx 方法构建可控的超时与错误路径测试；禁止在单测中依赖真实 KEYS 扫描。

示例一：在服务入口统一传递上下文（推荐使用 `WithContext` 封装）

```go
package service

import (
    "context"
    "time"

    pkgredis "github.com/amuluze/conan/redis"
)

// FetchUserProfile 从 Redis 获取用户画像，并在 200ms 超时后自动取消。
// 说明：使用 Client.WithContext(ctx) 保持原 API 名称，同时具备上下文超时与取消能力。
func FetchUserProfile(ctx context.Context, rc *pkgredis.Client, userID string) (string, error) {
    // 函数级注释：为服务层提供统一的上下文管理与 Redis 访问入口，便于调用链超时与取消传播。
    ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
    defer cancel()

    w := rc.WithContext(ctx)
    key := "app:prod:user:profile:" + userID
    return w.Get(key)
}
```

示例二：使用 SCAN 进行键空间分页遍历（替代 KEYS）

```go
package scanexample

import (
    "context"
    "fmt"
    "time"

    pkgredis "github.com/amuluze/conan/redis"
)

// ScanUsersPrefix 按前缀分批扫描用户相关键，避免阻塞与超时。
// 说明：使用 ScanKeysCtx 进行分页扫描，可在大键空间下渐进处理。
func ScanUsersPrefix(ctx context.Context, rc *pkgredis.Client, prefix string, pageSize int64) error {
    // 函数级注释：在批处理任务中使用上下文控制总时长，并支持外部取消。
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    var (
        cursor uint64 = 0
        pattern        = prefix + "*"
    )

    for {
        keys, next, err := rc.ScanKeysCtx(ctx, pattern, pageSize)
        if err != nil {
            return err
        }
        cursor = next

        for _, k := range keys {
            fmt.Println("found:", k)
        }
        if cursor == 0 {
            break
        }
    }
    return nil
}
```

示例三：区分键不存在错误（redis.Nil）

```go
package errhandling

import (
    "context"
    goredis "github.com/redis/go-redis/v9"
    pkgredis "github.com/amuluze/conan/redis"
)

// GetOrDefault 获取键，不存在时返回默认值；其他错误向上抛出。
// 说明：通过判断 goredis.Nil 来区分“键不存在”。
func GetOrDefault(ctx context.Context, rc *pkgredis.Client, key, defaultVal string) (string, error) {
    // 函数级注释：封装通用错误处理逻辑，降低调用方重复判断成本。
    v, err := rc.GetCtx(ctx, key)
    if err == nil {
        return v, nil
    }
    if err == goredis.Nil {
        return defaultVal, nil
    }
    return "", err
}
```

示例四：分布式锁（SET NX + EX）与安全释放

```go
package distlock

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "time"

    pkgredis "github.com/amuluze/conan/redis"
)

// AcquireLock 尝试获取分布式锁，返回唯一 token；若失败返回空字符串。
// 说明：使用 SetNXCtx 原子设置并附带过期时间，避免死锁。
func AcquireLock(ctx context.Context, rc *pkgredis.Client, key string, ttl time.Duration) (string, bool, error) {
    // 函数级注释：用于协调共享资源的访问，保障并发安全。
    // 生成随机 token 作为锁值，用于安全释放
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        return "", false, err
    }
    token := hex.EncodeToString(b)

    ok, err := rc.SetNXCtx(ctx, key, token, ttl)
    if err != nil || !ok {
        return "", false, err
    }
    return token, true, nil
}

// ReleaseLock 仅当给定 token 与锁中存储的值一致时才释放锁。
// 说明：通过 Lua 脚本保证释放操作的原子性，避免误删他人锁。
func ReleaseLock(ctx context.Context, rc *pkgredis.Client, key, token string) (bool, error) {
    // 函数级注释：确保锁释放的原子性与安全性。
    script := `
        if redis.call('GET', KEYS[1]) == ARGV[1] then
            return redis.call('DEL', KEYS[1])
        else
            return 0
        end
    `
    // 直接使用底层 UniversalClient 执行脚本
    res, err := rc.UniversalClient.Eval(ctx, script, []string{key}, token).Int()
    if err != nil {
        return false, err
    }
    return res == 1, nil
}
```

示例五：固定窗口限流（INCR + EX）

```go
package ratelimit

import (
    "context"
    "fmt"
    "time"

    pkgredis "github.com/amuluze/conan/redis"
)

// Allow 尝试在窗口期内通过限流；返回是否允许。
// 说明：在首次创建计数器时设置过期，后续仅递增。
func Allow(ctx context.Context, rc *pkgredis.Client, key string, limit int64, window time.Duration) (bool, error) {
    // 函数级注释：适用于接口级别的简单限流，避免过载。
    // 递增计数
    count, err := rc.IncrCtx(ctx, key)
    if err != nil {
        return false, err
    }
    if count == 1 {
        // 首次创建，设置窗口过期
        if _, err := rc.ExpireCtx(ctx, key, window); err != nil {
            return false, err
        }
    }
    allowed := count <= limit
    if !allowed {
        fmt.Printf("rate limited: key=%s count=%d limit=%d\n", key, count, limit)
    }
    return allowed, nil
}
```

示例六：Pipeline 批量写入优化

```go
package pipeexample

import (
    "context"
    "time"

    pkgredis "github.com/amuluze/conan/redis"
)

// BatchWrite 使用 Pipeline 批量写入，减少 RTT，提升吞吐。
// 说明：直接通过底层 UniversalClient.Pipeline 使用上下文。
func BatchWrite(ctx context.Context, rc *pkgredis.Client, pairs map[string]string) error {
    // 函数级注释：适用于批量缓存预热、迁移等场景。
    ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
    defer cancel()

    pipe := rc.UniversalClient.Pipeline()
    for k, v := range pairs {
        pipe.Set(ctx, k, v, time.Hour)
    }
    _, err := pipe.Exec(ctx)
    return err
}
```

示例七：Pipeline 批量读取与错误聚合

```go
package pipeget

import (
    "context"
    "fmt"
    "time"

    goredis "github.com/redis/go-redis/v9"
    pkgredis "github.com/amuluze/conan/redis"
)

// BatchGet 使用 Pipeline 批量读取多个键，降低 RTT；并对每条命令错误进行聚合，便于统一处理。
// 说明：返回值包含 values（成功读取的键值）、perKeyErr（每个键的错误，不含 redis.Nil）、aggErr（Exec 与命令错误的综合）。
func BatchGet(ctx context.Context, rc *pkgredis.Client, keys []string) (values map[string]string, perKeyErr map[string]error, aggErr error) {
    // 函数级注释：适用于批量查询场景，支持超时与取消，并提供多维度错误汇总。
    ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
    defer cancel()

    pipe := rc.UniversalClient.Pipeline()

    // 预创建每个键的命令指针，Exec 之后统一提取结果
    cmdMap := make(map[string]*goredis.StringCmd, len(keys))
    for _, k := range keys {
        cmdMap[k] = pipe.Get(ctx, k)
    }

    // 执行 Pipeline
    cmders, execErr := pipe.Exec(ctx)

    values = make(map[string]string, len(keys))
    perKeyErr = make(map[string]error)

    // 提取每个键的结果与错误，Nil 视为“键不存在”不计为错误（可按需自定义）
    for k, cmd := range cmdMap {
        v, err := cmd.Result()
        if err != nil {
            if err != goredis.Nil {
                perKeyErr[k] = err
            }
            continue
        }
        values[k] = v
    }

    // 聚合 Exec 错误与每条命令错误（排除 Nil）
    agg := execErr
    for _, c := range cmders {
        if e := c.Err(); e != nil && e != goredis.Nil {
            if agg == nil {
                agg = e
            } else {
                agg = fmt.Errorf("%v; %v", agg, e)
            }
        }
    }
    return values, perKeyErr, agg
}
```

示例八：Pipeline 混合读写与部分失败处理（非事务）

```go
package pipemixed

import (
    "context"
    "fmt"
    "time"

    goredis "github.com/redis/go-redis/v9"
    pkgredis "github.com/amuluze/conan/redis"
)

// BatchSetAndGet 在同一 Pipeline 中混合批量写入与读取，保持命令顺序执行，并进行错误聚合与部分失败处理。
// 注意：Pipeline 非事务，若需原子性请使用 TxPipeline 或 WATCH/MULTI/EXEC。
func BatchSetAndGet(ctx context.Context, rc *pkgredis.Client, pairs map[string]string, ttl time.Duration) (map[string]string, error) {
    // 函数级注释：适用于缓存预热与读后校验的场景，统一控制上下文超时与错误处理策略。
    ctx, cancel := context.WithTimeout(ctx, time.Second)
    defer cancel()

    pipe := rc.UniversalClient.Pipeline()

    // 1) 批量写入（设置过期）
    for k, v := range pairs {
        pipe.Set(ctx, k, v, ttl)
    }

    // 2) 批量读取（按刚写入的 key）
    resultCmds := make(map[string]*goredis.StringCmd, len(pairs))
    for k := range pairs {
        resultCmds[k] = pipe.Get(ctx, k)
    }

    // 3) 执行并聚合错误
    cmders, execErr := pipe.Exec(ctx)
    if execErr != nil {
        var agg error = execErr
        for _, c := range cmders {
            if e := c.Err(); e != nil && e != goredis.Nil {
                agg = fmt.Errorf("%v; %v", agg, e)
            }
        }
        return nil, agg
    }

    // 4) 组装读取结果；对个别失败（如 Nil）可选择忽略或记录
    out := make(map[string]string, len(pairs))
    for k, cmd := range resultCmds {
        v, err := cmd.Result()
        if err != nil {
            if err != goredis.Nil {
                // 非 Nil 的错误直接返回，或按需改为收集后统一返回
                return nil, err
            }
            continue
        }
        out[k] = v
    }
    return out, nil
}
```

补充说明：

- Pipeline 可显著减少网络往返次数（RTT），但不具备事务原子性；如需原子性，可使用 TxPipeline 或 WATCH/MULTI/EXEC。
- 错误聚合建议同时考虑 Exec 返回值与每条命令的 Err（排除 redis.Nil），并结合业务语义决定是否将 Nil 作为正常情况处理或转化为默认值。
- 高并发场景下建议为 Pipeline 设置合理的超时，并结合连接池（PoolSize）与服务超时策略统一规划。

更多建议：

- 避免大键与超长列表/集合，必要时进行分片或分页。
- 使用哈希存储结构化对象字段，便于部分更新与读取（如 `HSetMap`）。
- 在热点场景下考虑本地缓存与双写一致性策略（按需引入）。
- 对 Scan 的 `count` 做压测调优，避免过小导致慢、过大导致阻塞风险。
- 生产环境开启慢查询监控，结合连接池指标观察瓶颈与超时错误。

## 贡献

欢迎提交Issue和Pull Request来改进这个模块。

## 许可证

本项目采用MIT许可证。