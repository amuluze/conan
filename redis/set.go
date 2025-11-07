// Package redis
// Date: 2023/12/4 14:01
// Author: Amu
// Description:
package redis

import "context"

// ======================== set 指令 ======================== //

// SAdd 将一个或多个元素添加到集合，返回新增元素的数量。
func (rc *Client) SAdd(key string, values ...interface{}) (int64, error) {
    return rc.UniversalClient.SAdd(ctx, key, values...).Result()
}

// SAddCtx 将一个或多个元素添加到集合（带上下文），返回新增元素的数量。
// 参数：
// - ctx: 上下文
// - key: 集合键名
// - values: 元素列表
func (rc *Client) SAddCtx(ctx context.Context, key string, values ...interface{}) (int64, error) {
    return rc.UniversalClient.SAdd(ctx, key, values...).Result()
}

// SPop 随机获取一个元素，无序，随机
func (rc *Client) SPop(key string) (string, error) {
    return rc.UniversalClient.SPop(ctx, key).Result()
}

// SPopCtx 随机弹出一个元素（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 集合键名
func (rc *Client) SPopCtx(ctx context.Context, key string) (string, error) {
    return rc.UniversalClient.SPop(ctx, key).Result()
}

func (rc *Client) SRem(key string, values ...interface{}) (int64, error) {
    return rc.UniversalClient.SRem(ctx, key, values...).Result()
}

// SRemCtx 从集合中移除一个或多个元素（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 集合键名
// - values: 要移除的元素
func (rc *Client) SRemCtx(ctx context.Context, key string, values ...interface{}) (int64, error) {
    return rc.UniversalClient.SRem(ctx, key, values...).Result()
}

func (rc *Client) SMembers(key string) ([]string, error) {
    return rc.UniversalClient.SMembers(ctx, key).Result()
}

// SMembersCtx 返回集合中的所有成员（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 集合键名
func (rc *Client) SMembersCtx(ctx context.Context, key string) ([]string, error) {
    return rc.UniversalClient.SMembers(ctx, key).Result()
}

// SIsMember 判断元素是否存在于集合中，命名与 go-redis 保持一致。
func (rc *Client) SIsMember(key string, value interface{}) (bool, error) {
    return rc.UniversalClient.SIsMember(ctx, key, value).Result()
}

// SIsMemberCtx 判断元素是否存在于集合中（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 集合键名
// - value: 元素值
func (rc *Client) SIsMemberCtx(ctx context.Context, key string, value interface{}) (bool, error) {
    return rc.UniversalClient.SIsMember(ctx, key, value).Result()
}

func (rc *Client) SCard(key string) (int64, error) {
    return rc.UniversalClient.SCard(ctx, key).Result()
}

// SCardCtx 返回集合的元素数量（带上下文）。
// 参数：
// - ctx: 上下文
// - key: 集合键名
func (rc *Client) SCardCtx(ctx context.Context, key string) (int64, error) {
    return rc.UniversalClient.SCard(ctx, key).Result()
}

func (rc *Client) SUnion(key1, key2 string) ([]string, error) {
    return rc.UniversalClient.SUnion(ctx, key1, key2).Result()
}

// SUnionCtx 返回两个集合的并集（带上下文）。
// 参数：
// - ctx: 上下文
// - key1, key2: 集合键名
func (rc *Client) SUnionCtx(ctx context.Context, key1, key2 string) ([]string, error) {
    return rc.UniversalClient.SUnion(ctx, key1, key2).Result()
}

func (rc *Client) SDiff(key1, key2 string) ([]string, error) {
    return rc.UniversalClient.SDiff(ctx, key1, key2).Result()
}

// SDiffCtx 返回两个集合的差集（带上下文）。
// 参数：
// - ctx: 上下文
// - key1, key2: 集合键名
func (rc *Client) SDiffCtx(ctx context.Context, key1, key2 string) ([]string, error) {
    return rc.UniversalClient.SDiff(ctx, key1, key2).Result()
}

func (rc *Client) SInter(key1, key2 string) ([]string, error) {
    return rc.UniversalClient.SInter(ctx, key1, key2).Result()
}

// SInterCtx 返回两个集合的交集（带上下文）。
// 参数：
// - ctx: 上下文
// - key1, key2: 集合键名
func (rc *Client) SInterCtx(ctx context.Context, key1, key2 string) ([]string, error) {
    return rc.UniversalClient.SInter(ctx, key1, key2).Result()
}
