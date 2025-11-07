// Package redis
// Date: 2025/11/06
// Author: Amu
// Description: Redis utility functions
package redis

import (
	"context"
	"fmt"
	"time"
)

// PingRedis 测试Redis连接是否正常
func (rc *Client) PingRedis() error {
	if rc.UniversalClient == nil {
		return fmt.Errorf("redis client is nil")
	}

	// 使用超时context避免长时间阻塞
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := rc.UniversalClient.Ping(timeoutCtx).Result()
	return err
}

// IsRedisAvailable 检查Redis是否可用
func (rc *Client) IsRedisAvailable() bool {
	return rc.PingRedis() == nil
}

// SetWithExpiration 设置带过期时间的key
func (rc *Client) SetWithExpiration(key, value string, expiration time.Duration) error {
	if rc.UniversalClient == nil {
		return fmt.Errorf("redis client is nil")
	}

	return rc.UniversalClient.Set(ctx, key, value, expiration).Err()
}

// GetOrSet 获取key，如果不存在则设置默认值
func (rc *Client) GetOrSet(key, defaultValue string) (string, error) {
	if rc.UniversalClient == nil {
		return defaultValue, fmt.Errorf("redis client is nil")
	}

	val, err := rc.UniversalClient.Get(ctx, key).Result()
	if err != nil {
		// key不存在，设置默认值
		err := rc.UniversalClient.Set(ctx, key, defaultValue, 0).Err()
		if err != nil {
			return defaultValue, fmt.Errorf("failed to set default value: %w", err)
		}
		return defaultValue, nil
	}

	return val, nil
}

// IncrementAtomic 原子性增加计数器
func (rc *Client) IncrementAtomic(key string, increment int64) (int64, error) {
	if rc.UniversalClient == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	return rc.UniversalClient.IncrBy(ctx, key, increment).Result()
}

// BatchDelete 批量删除keys
func (rc *Client) BatchDelete(keys []string) (int64, error) {
	if rc.UniversalClient == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	if len(keys) == 0 {
		return 0, nil
	}

	return rc.UniversalClient.Del(ctx, keys...).Result()
}

// BatchExpire 批量设置过期时间
func (rc *Client) BatchExpire(keys []string, expiration time.Duration) error {
	if rc.UniversalClient == nil {
		return fmt.Errorf("redis client is nil")
	}

	if len(keys) == 0 {
		return nil
	}

	// 使用pipeline提高性能
	pipe := rc.UniversalClient.Pipeline()
	for _, key := range keys {
		pipe.Expire(ctx, key, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetClientInfo 获取客户端信息
func (rc *Client) GetClientInfo() (map[string]string, error) {
	if rc.UniversalClient == nil {
		return nil, fmt.Errorf("redis client is nil")
	}

	info, err := rc.UniversalClient.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	// 简单解析Redis info信息
	result := make(map[string]string)
	result["info"] = info
	return result, nil
}

// GetRedisVersion 获取Redis版本
func (rc *Client) GetRedisVersion() (string, error) {
	if rc.UniversalClient == nil {
		return "", fmt.Errorf("redis client is nil")
	}

	info, err := rc.UniversalClient.Info(ctx, "server").Result()
	if err != nil {
		return "", err
	}

	// 简单解析版本信息
	// 实际实现中可能需要更复杂的解析逻辑
	return info, nil
}