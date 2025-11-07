// Package redis
// Date: 2023/12/4 13:59
// Author: Amu
// Description:
package redis

import (
    "context"
    "time"
)

// ======================== 基本指令 ======================== //

// ScanKeys 使用 SCAN 遍历匹配的键，避免 KEYS 阻塞。
// 参数：
// - pattern: 匹配模式，如 "user:*"
// - count: 每批次扫描的提示数量（非严格限制），通常设置为较大的值以提高扫描效率
// 返回：匹配到的所有键切片
func (rc *Client) ScanKeys(pattern string, count int64) ([]string, error) {
    var (
        cursor uint64 = 0
        results       []string
    )
    for {
        keys, nextCursor, err := rc.UniversalClient.Scan(ctx, cursor, pattern, count).Result()
        if err != nil {
            return nil, err
        }
        results = append(results, keys...)
        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }
    return results, nil
}

// ScanKeysCtx 使用 SCAN 遍历匹配的键（带上下文），避免 KEYS 阻塞。
// 参数：
// - ctx: 上下文
// - pattern: 匹配模式，如 "user:*"
// - count: 每批次扫描的提示数量（非严格限制）
// 返回：匹配到的所有键切片
func (rc *Client) ScanKeysCtx(ctx context.Context, pattern string, count int64) ([]string, error) {
    var (
        cursor  uint64 = 0
        results        []string
    )
    for {
        keys, nextCursor, err := rc.UniversalClient.Scan(ctx, cursor, pattern, count).Result()
        if err != nil {
            return nil, err
        }
        results = append(results, keys...)
        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }
    return results, nil
}

func (rc *Client) Type(key string) (string, error) {
    return rc.UniversalClient.Type(ctx, key).Result()
}

// TypeCtx 返回键的类型（带上下文）。
func (rc *Client) TypeCtx(ctx context.Context, key string) (string, error) {
    return rc.UniversalClient.Type(ctx, key).Result()
}

func (rc *Client) Delete(keys ...string) (int64, error) {
    return rc.UniversalClient.Del(ctx, keys...).Result()
}

// DeleteCtx 删除一个或多个键（带上下文）。
func (rc *Client) DeleteCtx(ctx context.Context, keys ...string) (int64, error) {
    return rc.UniversalClient.Del(ctx, keys...).Result()
}

func (rc *Client) Exists(key string) (int64, error) {
    return rc.UniversalClient.Exists(ctx, key).Result()
}

// ExistsCtx 判断键是否存在（带上下文）。
func (rc *Client) ExistsCtx(ctx context.Context, key string) (int64, error) {
    return rc.UniversalClient.Exists(ctx, key).Result()
}

func (rc *Client) Expire(key string, expireDuration time.Duration) (bool, error) {
    return rc.UniversalClient.Expire(ctx, key, expireDuration).Result()
}

// ExpireCtx 设置键过期时间（带上下文）。
func (rc *Client) ExpireCtx(ctx context.Context, key string, expireDuration time.Duration) (bool, error) {
    return rc.UniversalClient.Expire(ctx, key, expireDuration).Result()
}

func (rc *Client) ExpireAt(key string, expireTime time.Time) (bool, error) {
    return rc.UniversalClient.ExpireAt(ctx, key, expireTime).Result()
}

// ExpireAtCtx 在指定时间点过期（带上下文）。
func (rc *Client) ExpireAtCtx(ctx context.Context, key string, expireTime time.Time) (bool, error) {
    return rc.UniversalClient.ExpireAt(ctx, key, expireTime).Result()
}

func (rc *Client) TTL(key string) (time.Duration, error) {
    return rc.UniversalClient.TTL(ctx, key).Result()
}

// TTLCtx 获取键剩余过期时间（带上下文）。
func (rc *Client) TTLCtx(ctx context.Context, key string) (time.Duration, error) {
    return rc.UniversalClient.TTL(ctx, key).Result()
}

func (rc *Client) PTTL(key string) (time.Duration, error) {
    return rc.UniversalClient.PTTL(ctx, key).Result()
}

// PTTLCtx 获取键剩余过期时间（毫秒）（带上下文）。
func (rc *Client) PTTLCtx(ctx context.Context, key string) (time.Duration, error) {
    return rc.UniversalClient.PTTL(ctx, key).Result()
}

func (rc *Client) DBSize() (int64, error) {
    return rc.UniversalClient.DBSize(ctx).Result()
}

// DBSizeCtx 返回当前数据库键数量（带上下文）。
func (rc *Client) DBSizeCtx(ctx context.Context) (int64, error) {
    return rc.UniversalClient.DBSize(ctx).Result()
}

// FlushDB 清空当前数据库
func (rc *Client) FlushDB() (string, error) {
    return rc.UniversalClient.FlushDB(ctx).Result()
}

// FlushDBCtx 清空当前数据库（带上下文）。
func (rc *Client) FlushDBCtx(ctx context.Context) (string, error) {
    return rc.UniversalClient.FlushDB(ctx).Result()
}

// FlushAll 清空所有数据库
func (rc *Client) FlushAll() (string, error) {
    return rc.UniversalClient.FlushAll(ctx).Result()
}

// FlushAllCtx 清空所有数据库（带上下文）。
func (rc *Client) FlushAllCtx(ctx context.Context) (string, error) {
    return rc.UniversalClient.FlushAll(ctx).Result()
}
