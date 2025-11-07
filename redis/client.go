// Package redis
// Date: 2022/9/23 00:51
// Author: Amu
// Description:
package redis

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Client struct {
    redis.UniversalClient
}

// ContextClient 是带有默认上下文的轻量封装，用于在不修改原有 API 的基础上
// 为调用方提供统一的上下文传递方式。
//
// 注意：ContextClient 并不复制底层连接，仅保存一个默认的 context.Context 并将调用
// 委托给 Client 的 *Ctx 方法。适用于需要统一取消、截止时间或链路追踪的场景。
type ContextClient struct {
    base *Client
    ctx  context.Context
}

// WithContext 返回一个绑定了默认上下文的 ContextClient。
//
// 参数：
// - ctx: 默认上下文，用于后续所有通过 ContextClient 发起的调用
// 返回：
// - *ContextClient：可以使用与 Client 相同的 API（通过 *Ctx 方法委托）
func (rc *Client) WithContext(ctx context.Context) *ContextClient {
    return &ContextClient{base: rc, ctx: ctx}
}

// =============== ContextClient 常用方法委托（按需补充）================ //

// Set 使用默认上下文设置键值。
func (cc *ContextClient) Set(key, value string) (string, error) {
    return cc.base.SetEXCtx(cc.ctx, key, value, 0)
}

// SetEX 使用默认上下文设置键值并附带过期时间。
func (cc *ContextClient) SetEX(key, value string, duration time.Duration) (string, error) {
    return cc.base.SetEXCtx(cc.ctx, key, value, duration)
}

// SetNX 使用默认上下文在键不存在时设置值。
func (cc *ContextClient) SetNX(key, value string, duration time.Duration) (bool, error) {
    return cc.base.SetNXCtx(cc.ctx, key, value, duration)
}

// Get 使用默认上下文获取键值。
func (cc *ContextClient) Get(key string) (string, error) {
    return cc.base.GetCtx(cc.ctx, key)
}

// Delete 使用默认上下文删除键。
func (cc *ContextClient) Delete(keys ...string) (int64, error) {
    return cc.base.DeleteCtx(cc.ctx, keys...)
}

// Exists 使用默认上下文判断键是否存在。
func (cc *ContextClient) Exists(key string) (int64, error) {
    return cc.base.ExistsCtx(cc.ctx, key)
}

// ScanKeys 使用默认上下文扫描匹配键。
func (cc *ContextClient) ScanKeys(pattern string, count int64) ([]string, error) {
    return cc.base.ScanKeysCtx(cc.ctx, pattern, count)
}

// HSet 使用默认上下文设置哈希字段。
func (cc *ContextClient) HSet(key string, fieldValues ...interface{}) (int64, error) {
    return cc.base.HSetCtx(cc.ctx, key, fieldValues...)
}

// HSetMap 使用默认上下文设置哈希字段（map）。
func (cc *ContextClient) HSetMap(key string, values map[string]interface{}) (int64, error) {
    return cc.base.HSetMapCtx(cc.ctx, key, values)
}

// HGet 使用默认上下文获取哈希字段。
func (cc *ContextClient) HGet(key, field string) (string, error) {
    return cc.base.HGetCtx(cc.ctx, key, field)
}

// HGetAll 使用默认上下文获取哈希全部字段。
func (cc *ContextClient) HGetAll(key string) (map[string]string, error) {
    return cc.base.HGetAllCtx(cc.ctx, key)
}

// SAdd 使用默认上下文向集合添加成员。
func (cc *ContextClient) SAdd(key string, values ...interface{}) (int64, error) {
    return cc.base.SAddCtx(cc.ctx, key, values...)
}

// SIsMember 使用默认上下文判断成员是否在集合中。
func (cc *ContextClient) SIsMember(key string, value interface{}) (bool, error) {
    return cc.base.SIsMemberCtx(cc.ctx, key, value)
}

// ZAdd 使用默认上下文向有序集合添加成员。
func (cc *ContextClient) ZAdd(key string, member interface{}, score float64) (int64, error) {
    return cc.base.ZAddCtx(cc.ctx, key, member, score)
}

// ZRange 使用默认上下文按排名区间返回成员。
func (cc *ContextClient) ZRange(key string, start int64, end int64) ([]string, error) {
    return cc.base.ZRangeCtx(cc.ctx, key, start, end)
}

// LPush 使用默认上下文将值插入列表头部。
func (cc *ContextClient) LPush(key string, values ...interface{}) (int64, error) {
    return cc.base.LPushCtx(cc.ctx, key, values...)
}

// RPop 使用默认上下文从列表尾部弹出元素。
func (cc *ContextClient) RPop(key string) (string, error) {
    return cc.base.RPopCtx(cc.ctx, key)
}

// NewClient 创建并返回一个 Redis 通用客户端（支持单机、集群、哨兵）。
// 默认配置：
// - Addrs: ["localhost:6379"]
// - Password: ""（默认不启用密码）
// - DB: 0
// - PoolSize: 50
// - MasterName: ""（为空表示非哨兵模式）
// - DialTimeout: 5s
// - ReadTimeout: 3s
// - WriteTimeout: 3s
// 你可以通过可选参数 Option 覆盖上述默认配置。
//
// 注意：此函数会尝试连接Redis，如果连接失败会返回错误。如果需要在没有Redis的情况下创建客户端用于测试，请使用NewClientWithoutPing。
func NewClient(opts ...Option) (*Client, error) {
    conf := &option{
        Addrs:                 []string{"localhost:6379"},
        Password:              "",
        DB:                    0,
        PoolSize:              50,
        MasterName:            "",
        DialConnectionTimeout: "5s",
        DialReadTimeout:       "3s",
        DialWriteTimeout:      "3s",
    }
    for _, opt := range opts {
        opt(conf)
    }

    // 验证配置
    if err := validateConfig(conf); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    // 解析超时配置
    dialTimeout, err := time.ParseDuration(conf.DialConnectionTimeout)
    if err != nil {
        return nil, fmt.Errorf("invalid dial timeout %s: %w", conf.DialConnectionTimeout, err)
    }
    readTimeout, err := time.ParseDuration(conf.DialReadTimeout)
    if err != nil {
        return nil, fmt.Errorf("invalid read timeout %s: %w", conf.DialReadTimeout, err)
    }
    writeTimeout, err := time.ParseDuration(conf.DialWriteTimeout)
    if err != nil {
        return nil, fmt.Errorf("invalid write timeout %s: %w", conf.DialWriteTimeout, err)
    }
    c := redis.NewUniversalClient(&redis.UniversalOptions{
        Addrs:        conf.Addrs,
        DB:           conf.DB,
        Password:     conf.Password,
        PoolSize:     conf.PoolSize,
        MasterName:   conf.MasterName,
        DialTimeout:  dialTimeout,
        ReadTimeout:  readTimeout,
        WriteTimeout: writeTimeout,
    })

    // 测试连接
    _, err = c.Ping(ctx).Result()
    if err != nil {
        log.Printf("failed to connect to redis: %v", err)
        return nil, fmt.Errorf("failed to connect to redis at %v: %w", conf.Addrs, err)
    }

    return &Client{c}, nil
}

// NewClientWithoutPing 创建Redis客户端但不进行连接测试
// 此函数用于在测试或离线环境下创建客户端实例
func NewClientWithoutPing(opts ...Option) (*Client, error) {
    conf := &option{
        Addrs:                 []string{"localhost:6379"},
        Password:              "",
        DB:                    0,
        PoolSize:              50,
        MasterName:            "",
        DialConnectionTimeout: "5s",
        DialReadTimeout:       "3s",
        DialWriteTimeout:      "3s",
    }
    for _, opt := range opts {
        opt(conf)
    }

    // 验证配置
    if err := validateConfig(conf); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    // 解析超时配置
    dialTimeout, err := time.ParseDuration(conf.DialConnectionTimeout)
    if err != nil {
        return nil, fmt.Errorf("invalid dial timeout %s: %w", conf.DialConnectionTimeout, err)
    }
    readTimeout, err := time.ParseDuration(conf.DialReadTimeout)
    if err != nil {
        return nil, fmt.Errorf("invalid read timeout %s: %w", conf.DialReadTimeout, err)
    }
    writeTimeout, err := time.ParseDuration(conf.DialWriteTimeout)
    if err != nil {
        return nil, fmt.Errorf("invalid write timeout %s: %w", conf.DialWriteTimeout, err)
    }
    c := redis.NewUniversalClient(&redis.UniversalOptions{
        Addrs:        conf.Addrs,
        DB:           conf.DB,
        Password:     conf.Password,
        PoolSize:     conf.PoolSize,
        MasterName:   conf.MasterName,
        DialTimeout:  dialTimeout,
        ReadTimeout:  readTimeout,
        WriteTimeout: writeTimeout,
    })

    return &Client{c}, nil
}

// validateConfig 验证Redis配置的有效性
func validateConfig(conf *option) error {
    if len(conf.Addrs) == 0 {
        return fmt.Errorf("addrs cannot be empty")
    }

    for _, addr := range conf.Addrs {
        if addr == "" {
            return fmt.Errorf("address cannot be empty")
        }
    }

    if conf.PoolSize <= 0 {
        return fmt.Errorf("pool size must be positive")
    }

    if conf.DB < 0 || conf.DB > 15 {
        return fmt.Errorf("DB must be between 0 and 15")
    }

    return nil
}

// Close 关闭底层的 Redis 客户端连接。
// 注意：该方法会尝试关闭连接并忽略返回错误，如需精细控制可直接使用 rc.UniversalClient.Close()。
func (rc *Client) Close() {
    if rc != nil && rc.UniversalClient != nil {
        _ = rc.UniversalClient.Close()
    }
}
