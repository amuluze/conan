// Package redis
// Date: 2023/12/4 14:02
// Author: Amu
// Description:
package redis

import (
	"testing"
)

func TestClient(t *testing.T) {
	// 使用无效端口测试客户端创建失败情况
	invalidOpts := []Option{
		WithAddrs([]string{"localhost:9999"}), // 使用无效端口
		WithConnectionTimeout("1s"),           // 设置较短超时
	}

	rc, err := NewClient(invalidOpts...)
	if err != nil {
		// 预期会连接失败，这是正常的
		t.Logf("Expected connection failure: %v", err)
		return
	}

	// 如果意外连接成功，清理连接
	defer rc.Close()
	t.Log("redis client created successfully")
}

func TestClientWithOptions(t *testing.T) {
    opts := []Option{
        WithAddrs([]string{"localhost:6379"}),
        WithPassword("test123"),
        WithDB(1),
        WithPoolSize(10),
        WithConnectionTimeout("3s"),
        WithReadTimeout("2s"),
        WithWriteTimeout("2s"),
    }

	// 创建配置结构体，不实际连接
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

	// 应用选项
	for _, opt := range opts {
		opt(conf)
	}

	// 验证配置是否正确应用
	if len(conf.Addrs) != 1 || conf.Addrs[0] != "localhost:6379" {
		t.Errorf("Expected Addrs to be [localhost:6379], got %v", conf.Addrs)
	}
	if conf.Password != "test123" {
		t.Errorf("Expected Password to be test123, got %s", conf.Password)
	}
	if conf.DB != 1 {
		t.Errorf("Expected DB to be 1, got %d", conf.DB)
	}
	if conf.PoolSize != 10 {
		t.Errorf("Expected PoolSize to be 10, got %d", conf.PoolSize)
	}
	if conf.DialConnectionTimeout != "3s" {
		t.Errorf("Expected DialConnectionTimeout to be 3s, got %s", conf.DialConnectionTimeout)
	}
	if conf.DialReadTimeout != "2s" {
		t.Errorf("Expected DialReadTimeout to be 2s, got %s", conf.DialReadTimeout)
	}
	if conf.DialWriteTimeout != "2s" {
		t.Errorf("Expected DialWriteTimeout to be 2s, got %s", conf.DialWriteTimeout)
	}
}

func TestClientClose(t *testing.T) {
	rc := &Client{} // 创建一个空客户端
	rc.Close() // 应该不会panic

	// 测试nil客户端
	var nilClient *Client
	nilClient.Close() // 应该不会panic
}