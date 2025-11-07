// Package redis
// Date: 2023/12/4 14:01
// Author: Amu
// Description:
package redis

import "context"

// ======================== hash 指令 ======================== //

// HSet 使用可变参数设置多个字段的值（底层使用 HSET）。
// 用法示例：HSet(key, "f1", "v1", "f2", "v2")。
// 返回成功设置的字段数量。
// 参数：
// - key: 哈希键名
// - fieldValues: 交替出现的字段与值列表
func (rc *Client) HSet(key string, fieldValues ...interface{}) (int64, error) {
    return rc.UniversalClient.HSet(ctx, key, fieldValues...).Result()
}

// HSetCtx 使用可变参数设置多个字段的值（带上下文，底层使用 HSET）。
// 参数：
// - ctx: 上下文
// - key: 哈希键名
// - fieldValues: 交替出现的字段与值列表
func (rc *Client) HSetCtx(ctx context.Context, key string, fieldValues ...interface{}) (int64, error) {
    return rc.UniversalClient.HSet(ctx, key, fieldValues...).Result()
}

// HGet 获取指定字段的值。
// 若字段不存在，返回 redis.Nil 错误。
// 参数：
// - key: 哈希键名
// - field: 字段名
func (rc *Client) HGet(key, field string) (string, error) {
    return rc.UniversalClient.HGet(ctx, key, field).Result()
}

// HGetCtx 获取指定字段的值（带上下文）。
// 若字段不存在，返回 redis.Nil 错误。
// 参数：
// - ctx: 上下文
// - key: 哈希键名
// - field: 字段名
func (rc *Client) HGetCtx(ctx context.Context, key, field string) (string, error) {
    return rc.UniversalClient.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希中所有字段与其值。
// 返回 map[string]string，其中键为字段名，值为对应的字符串值。
// 参数：
// - key: 哈希键名
func (rc *Client) HGetAll(key string) (map[string]string, error) {
    return rc.UniversalClient.HGetAll(ctx, key).Result()
}

// HGetAllCtx 获取哈希中所有字段与其值（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 哈希键名
func (rc *Client) HGetAllCtx(ctx context.Context, key string) (map[string]string, error) {
    return rc.UniversalClient.HGetAll(ctx, key).Result()
}

// HDel 删除一个或多个字段，返回成功删除的字段数量。
// 参数：
// - key: 哈希键名
// - field: 要删除的字段列表
func (rc *Client) HDel(key string, field ...string) (int64, error) {
    return rc.UniversalClient.HDel(ctx, key, field...).Result()
}

// HDelCtx 删除一个或多个字段（带上下文），返回成功删除的字段数量。
// 参数：
// - ctx: 上下文
// - key: 哈希键名
// - field: 要删除的字段列表
func (rc *Client) HDelCtx(ctx context.Context, key string, field ...string) (int64, error) {
    return rc.UniversalClient.HDel(ctx, key, field...).Result()
}

// HExists 判断某字段是否存在。
// 参数：
// - key: 哈希键名
// - field: 字段名
func (rc *Client) HExists(key, field string) (bool, error) {
    return rc.UniversalClient.HExists(ctx, key, field).Result()
}

// HExistsCtx 判断某字段是否存在（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 哈希键名
// - field: 字段名
func (rc *Client) HExistsCtx(ctx context.Context, key, field string) (bool, error) {
    return rc.UniversalClient.HExists(ctx, key, field).Result()
}

// HLen 返回哈希中字段的数量。
// 参数：
// - key: 哈希键名
func (rc *Client) HLen(key string) (int64, error) {
    return rc.UniversalClient.HLen(ctx, key).Result()
}

// HLenCtx 返回哈希中字段的数量（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 哈希键名
func (rc *Client) HLenCtx(ctx context.Context, key string) (int64, error) {
    return rc.UniversalClient.HLen(ctx, key).Result()
}

// HSetMap 使用 map 设置多个字段的值（底层使用 HSET），替代已弃用的 HMSET。
// 返回成功设置的字段数量。
// 参数：
// - key: 哈希键名
// - values: 字段-值映射
func (rc *Client) HSetMap(key string, values map[string]interface{}) (int64, error) {
    return rc.UniversalClient.HSet(ctx, key, values).Result()
}

// HSetMapCtx 使用 map 设置多个字段的值（带上下文，底层使用 HSET），替代已弃用的 HMSET。
// 参数：
// - ctx: 上下文
// - key: 哈希键名
// - values: 字段-值映射
func (rc *Client) HSetMapCtx(ctx context.Context, key string, values map[string]interface{}) (int64, error) {
    return rc.UniversalClient.HSet(ctx, key, values).Result()
}
