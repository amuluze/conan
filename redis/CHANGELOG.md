# Changelog

本文件记录 `github.com/amuluze/conan/redis` 模块的重要变更与升级指南。

## 2025-11-07 不兼容性重构

类型：Breaking Changes / 重构

变更摘要：
- 移除所有为旧版本兼容保留的 API 与逻辑。
- 统一对外接口至基于 go-redis v9 的最新规范。

具体变更：
- 删除哈希旧方法：`Hset`（非规范命名）与 `HMset`（已弃用）。
  - 替代方案：使用 `HSet(key string, fieldValues ...interface{})` 或 `HSetMap(key string, values map[string]interface{})`。
- 删除集合重复方法：`SIsMembers`（命名不规范且与 `SIsMember` 重复）。
  - 保留并推荐使用：`SIsMember(key string, value interface{})`。
- 移除选项字段与函数：`option.IdleTimeout` 与 `WithIdleTimeout`（go-redis v9 已移除该概念）。
- 移除阻塞式键遍历：删除 `Keys()` 与 `KeysByPattern()`，新增非阻塞 `ScanKeys(pattern string, count int64)`。

依赖升级：
- `github.com/redis/go-redis/v9` 升级至 `v9.16.0`。

迁移指南：
1. 哈希：
   - 旧：`HMset(key, map)` → 新：`HSetMap(key, map)`。
   - 旧：`Hset(key, f1, v1, f2, v2)` → 新：`HSet(key, f1, v1, f2, v2)`（支持可变参数）。
2. 集合：
   - 旧：`SIsMembers(key, value)` → 新：`SIsMember(key, value)`。
3. 键遍历：
   - 旧：`Keys()`/`KeysByPattern(pattern)` → 新：`ScanKeys(pattern, count)`。
4. 选项：
   - 移除：`WithIdleTimeout()`；无需在 `option` 设置 `IdleTimeout`，请使用 `DialTimeout`、`ReadTimeout`、`WriteTimeout` 等有效选项。

注意事项：
- `ScanKeys` 是基于游标的非阻塞扫描，`count` 为提示值而非严格限制，扫描结果需合并所有批次。
- 依赖升级后，连接池与重试策略由 go-redis v9 管理，建议根据业务场景设置 `PoolSize`、`DialTimeout` 等。

新增（2025-11-07）：
- 添加上下文支持：为核心方法提供 `*Ctx` 变体（如 `SetEXCtx`、`GetCtx`、`ScanKeysCtx` 等）。
- 添加 `Client.WithContext(ctx)` 封装，返回 `ContextClient`，可使用与原 API 一致的方法名，但默认携带 `ctx` 并委托至 `*Ctx` 方法。

上下文使用示例：
```go
// 直接传入 ctx
_, _ = rc.SetEXCtx(ctx, key, value, ttl)

// 使用封装
w := rc.WithContext(ctx)
_, _ = w.SetEX(key, value, ttl)
```