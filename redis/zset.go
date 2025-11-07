// Package redis
// Date: 2023/12/4 14:01
// Author: Amu
// Description:
package redis

import (
    "context"
    "github.com/redis/go-redis/v9"
)

// ======================== zset 指令 ======================== //

// ZAdd 向有序集合添加一个元素及其分值，返回新增元素的数量。
func (rc *Client) ZAdd(key string, member interface{}, score float64) (int64, error) {
    return rc.UniversalClient.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Result()
}

// ZAddCtx 向有序集合添加一个元素及其分值（带上下文），返回新增元素的数量。
// 参数：
// - ctx: 上下文
// - key: 有序集合键名
// - member: 成员值
// - score: 分值
func (rc *Client) ZAddCtx(ctx context.Context, key string, member interface{}, score float64) (int64, error) {
    return rc.UniversalClient.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Result()
}

// ZIncrBy 增加 member 的分值，返回更新后的分值
func (rc *Client) ZIncrBy(key string, member string, score float64) (float64, error) {
    return rc.UniversalClient.ZIncrBy(ctx, key, score, member).Result()
}

// ZIncrByCtx 增加 member 的分值（带上下文），返回更新后的分值。
func (rc *Client) ZIncrByCtx(ctx context.Context, key string, member string, score float64) (float64, error) {
    return rc.UniversalClient.ZIncrBy(ctx, key, score, member).Result()
}

// ZRange 按排名区间 [start, end] 返回成员（按分值升序排列）。
func (rc *Client) ZRange(key string, start int64, end int64) ([]string, error) {
    return rc.UniversalClient.ZRange(ctx, key, start, end).Result()
}

// ZRangeCtx 按排名区间 [start, end] 返回成员（按分值升序排列，带上下文）。
func (rc *Client) ZRangeCtx(ctx context.Context, key string, start int64, end int64) ([]string, error) {
    return rc.UniversalClient.ZRange(ctx, key, start, end).Result()
}

// ZRevRange 按排名区间 [start, end] 返回成员（按分值降序排列）。
func (rc *Client) ZRevRange(key string, start int64, end int64) ([]string, error) {
    return rc.UniversalClient.ZRevRange(ctx, key, start, end).Result()
}

// ZRevRangeCtx 按排名区间 [start, end] 返回成员（按分值降序排列，带上下文）。
func (rc *Client) ZRevRangeCtx(ctx context.Context, key string, start int64, end int64) ([]string, error) {
    return rc.UniversalClient.ZRevRange(ctx, key, start, end).Result()
}

func (rc *Client) ZRangeByScore(key string, minScore string, maxScore string) ([]string, error) {
    return rc.UniversalClient.ZRangeByScore(ctx, key, &redis.ZRangeBy{Min: minScore, Max: maxScore}).Result()
}

// ZRangeByScoreCtx 根据分值区间正序获取元素（带上下文）。
func (rc *Client) ZRangeByScoreCtx(ctx context.Context, key string, minScore string, maxScore string) ([]string, error) {
    return rc.UniversalClient.ZRangeByScore(ctx, key, &redis.ZRangeBy{Min: minScore, Max: maxScore}).Result()
}

// ZRevRangeByScore 根据分值区间倒序获取元素（分值从高到低）。
func (rc *Client) ZRevRangeByScore(key string, minScore string, maxScore string) ([]string, error) {
    return rc.UniversalClient.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{Min: minScore, Max: maxScore}).Result()
}

// ZRevRangeByScoreCtx 根据分值区间倒序获取元素（分值从高到低，带上下文）。
func (rc *Client) ZRevRangeByScoreCtx(ctx context.Context, key string, minScore string, maxScore string) ([]string, error) {
    return rc.UniversalClient.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{Min: minScore, Max: maxScore}).Result()
}

func (rc *Client) ZCard(key string) (int64, error) {
    return rc.UniversalClient.ZCard(ctx, key).Result()
}

// ZCardCtx 返回有序集合的成员数量（带上下文）。
func (rc *Client) ZCardCtx(ctx context.Context, key string) (int64, error) {
    return rc.UniversalClient.ZCard(ctx, key).Result()
}

func (rc *Client) ZCount(key string, minScore, maxScore string) (int64, error) {
    return rc.UniversalClient.ZCount(ctx, key, minScore, maxScore).Result()
}

// ZCountCtx 返回分值区间内的成员数量（带上下文）。
func (rc *Client) ZCountCtx(ctx context.Context, key string, minScore, maxScore string) (int64, error) {
    return rc.UniversalClient.ZCount(ctx, key, minScore, maxScore).Result()
}

// ZScore 获取指定成员的分值
func (rc *Client) ZScore(key string, member string) (float64, error) {
    return rc.UniversalClient.ZScore(ctx, key, member).Result()
}

// ZScoreCtx 获取指定成员的分值（带上下文）。
func (rc *Client) ZScoreCtx(ctx context.Context, key string, member string) (float64, error) {
    return rc.UniversalClient.ZScore(ctx, key, member).Result()
}

func (rc *Client) ZRank(key string, value string) (int64, error) {
    return rc.UniversalClient.ZRank(ctx, key, value).Result()
}

// ZRankCtx 获取指定成员的排名（带上下文）。
func (rc *Client) ZRankCtx(ctx context.Context, key string, value string) (int64, error) {
    return rc.UniversalClient.ZRank(ctx, key, value).Result()
}

// ZRevRank 获取指定成员的倒序排名
func (rc *Client) ZRevRank(key string, value string) (int64, error) {
    return rc.UniversalClient.ZRevRank(ctx, key, value).Result()
}

// ZRevRankCtx 获取指定成员的倒序排名（带上下文）。
func (rc *Client) ZRevRankCtx(ctx context.Context, key string, value string) (int64, error) {
    return rc.UniversalClient.ZRevRank(ctx, key, value).Result()
}

func (rc *Client) ZRem(key string, value string) (int64, error) {
    return rc.UniversalClient.ZRem(ctx, key, value).Result()
}

// ZRemCtx 移除指定成员（带上下文）。
func (rc *Client) ZRemCtx(ctx context.Context, key string, value string) (int64, error) {
    return rc.UniversalClient.ZRem(ctx, key, value).Result()
}

func (rc *Client) ZRemRangeByRank(key string, startIndex, endIndex int64) (int64, error) {
    return rc.UniversalClient.ZRemRangeByRank(ctx, key, startIndex, endIndex).Result()
}

// ZRemRangeByRankCtx 按排名区间移除成员（带上下文）。
func (rc *Client) ZRemRangeByRankCtx(ctx context.Context, key string, startIndex, endIndex int64) (int64, error) {
    return rc.UniversalClient.ZRemRangeByRank(ctx, key, startIndex, endIndex).Result()
}

func (rc *Client) ZRemRangeByScore(key string, minScore, maxScore string) (int64, error) {
    return rc.UniversalClient.ZRemRangeByScore(ctx, key, minScore, maxScore).Result()
}

// ZRemRangeByScoreCtx 按分值区间移除成员（带上下文）。
func (rc *Client) ZRemRangeByScoreCtx(ctx context.Context, key string, minScore, maxScore string) (int64, error) {
    return rc.UniversalClient.ZRemRangeByScore(ctx, key, minScore, maxScore).Result()
}
