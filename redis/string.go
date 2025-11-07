// Package redis
// Date: 2023/12/4 14:00
// Author: Amu
// Description:
package redis

import (
    "context"
    "time"
)

// ======================== string 指令 ======================== //

// Set 设置字符串键的值，若键已存在则覆盖。
// 返回设置结果的状态字符串，例如 "OK"。
// 参数：
// - key: 键名
// - value: 键值
func (rc *Client) Set(key, value string) (string, error) {
    return rc.UniversalClient.Set(ctx, key, value, 0).Result()
}

// SetCtx 设置字符串键的值（带上下文），若键已存在则覆盖。
// 该方法允许调用方传入上下文以控制超时/取消等行为。
// 参数：
// - ctx: 上下文
// - key: 键名
// - value: 键值
func (rc *Client) SetCtx(ctx context.Context, key, value string) (string, error) {
    return rc.UniversalClient.Set(ctx, key, value, 0).Result()
}

// SetNX 当键不存在时设置值并可选设置过期时间。
// 返回：
// - true 表示成功设置（键原本不存在）
// - false 表示未设置（键已存在）
// 参数：
// - key: 键名
// - value: 键值
// - duration: 过期时长
func (rc *Client) SetNX(key, value string, duration time.Duration) (bool, error) {
    return rc.UniversalClient.SetNX(ctx, key, value, duration).Result()
}

// SetNXCtx 当键不存在时设置值并可选设置过期时间（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 键名
// - value: 键值
// - duration: 过期时长
func (rc *Client) SetNXCtx(ctx context.Context, key, value string, duration time.Duration) (bool, error) {
    return rc.UniversalClient.SetNX(ctx, key, value, duration).Result()
}

// SetEX 设置键的值并附带过期时间（EX）。
// 返回设置结果的状态字符串，例如 "OK"。
// 参数：
// - key: 键名
// - value: 键值
// - duration: 过期时长
func (rc *Client) SetEX(key, value string, duration time.Duration) (string, error) {
    return rc.UniversalClient.SetEx(ctx, key, value, duration).Result()
}

// SetEXCtx 设置键的值并附带过期时间（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 键名
// - value: 键值
// - duration: 过期时长
func (rc *Client) SetEXCtx(ctx context.Context, key, value string, duration time.Duration) (string, error) {
    return rc.UniversalClient.SetEx(ctx, key, value, duration).Result()
}

// Get 获取字符串键的值。
// 若键不存在会返回 redis.Nil 错误。
// 参数：
// - key: 键名
func (rc *Client) Get(key string) (string, error) {
    return rc.UniversalClient.Get(ctx, key).Result()
}

// GetCtx 获取字符串键的值（带上下文）。
// 若键不存在会返回 redis.Nil 错误。
// 参数：
// - ctx: 上下文
// - key: 键名
func (rc *Client) GetCtx(ctx context.Context, key string) (string, error) {
    return rc.UniversalClient.Get(ctx, key).Result()
}

// GetRange 按区间 [startIndex, endIndex] 获取子串。
// 索引支持负数，-1 表示最后一个字符。
// 参数：
// - key: 键名
// - startIndex: 起始索引
// - endIndex: 结束索引
func (rc *Client) GetRange(key string, startIndex int64, endIndex int64) (string, error) {
    return rc.UniversalClient.GetRange(ctx, key, startIndex, endIndex).Result()
}

// GetRangeCtx 按区间 [startIndex, endIndex] 获取子串（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 键名
// - startIndex: 起始索引
// - endIndex: 结束索引
func (rc *Client) GetRangeCtx(ctx context.Context, key string, startIndex int64, endIndex int64) (string, error) {
    return rc.UniversalClient.GetRange(ctx, key, startIndex, endIndex).Result()
}

// Incr 将键的数值递增 1。
// 若键不存在则当作 0，然后递增为 1。
// 返回递增后的数值。
// 参数：
// - key: 键名
func (rc *Client) Incr(key string) (int64, error) {
    return rc.UniversalClient.Incr(ctx, key).Result()
}

// IncrCtx 将键的数值递增 1（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 键名
func (rc *Client) IncrCtx(ctx context.Context, key string) (int64, error) {
    return rc.UniversalClient.Incr(ctx, key).Result()
}

// IncrBy 将键的数值按步长递增。
// 若键不存在则当作 0，然后递增。
// 返回递增后的数值。
// 参数：
// - key: 键名
// - step: 递增步长
func (rc *Client) IncrBy(key string, step int64) (int64, error) {
    return rc.UniversalClient.IncrBy(ctx, key, step).Result()
}

// IncrByCtx 将键的数值按步长递增（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 键名
// - step: 递增步长
func (rc *Client) IncrByCtx(ctx context.Context, key string, step int64) (int64, error) {
    return rc.UniversalClient.IncrBy(ctx, key, step).Result()
}

// Decr 将键的数值递减 1。
// 若键不存在则当作 0，然后递减为 -1。
// 返回递减后的数值。
// 参数：
// - key: 键名
func (rc *Client) Decr(key string) (int64, error) {
    return rc.UniversalClient.Decr(ctx, key).Result()
}

// DecrCtx 将键的数值递减 1（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 键名
func (rc *Client) DecrCtx(ctx context.Context, key string) (int64, error) {
    return rc.UniversalClient.Decr(ctx, key).Result()
}

// DecrBy 将键的数值按步长递减。
// 若键不存在则当作 0，然后递减。
// 返回递减后的数值。
// 参数：
// - key: 键名
// - step: 递减步长
func (rc *Client) DecrBy(key string, step int64) (int64, error) {
    return rc.UniversalClient.DecrBy(ctx, key, step).Result()
}

// DecrByCtx 将键的数值按步长递减（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 键名
// - step: 递减步长
func (rc *Client) DecrByCtx(ctx context.Context, key string, step int64) (int64, error) {
    return rc.UniversalClient.DecrBy(ctx, key, step).Result()
}

// Append 追加字符串到键的现有值末尾。
// 返回追加后字符串的总长度。
// 参数：
// - key: 键名
// - appendString: 要追加的字符串
func (rc *Client) Append(key string, appendString string) (int64, error) {
    return rc.UniversalClient.Append(ctx, key, appendString).Result()
}

// AppendCtx 追加字符串到键的现有值末尾（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 键名
// - appendString: 要追加的字符串
func (rc *Client) AppendCtx(ctx context.Context, key string, appendString string) (int64, error) {
    return rc.UniversalClient.Append(ctx, key, appendString).Result()
}

// StrLen 返回字符串键的长度。
// 若键不存在返回 0。
// 参数：
// - key: 键名
func (rc *Client) StrLen(key string) (int64, error) {
    return rc.UniversalClient.StrLen(ctx, key).Result()
}

// StrLenCtx 返回字符串键的长度（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 键名
func (rc *Client) StrLenCtx(ctx context.Context, key string) (int64, error) {
    return rc.UniversalClient.StrLen(ctx, key).Result()
}
