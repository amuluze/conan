// Package redis
// Date: 2025/11/06
// Author: Amu
// Description: Redis utility functions tests
package redis

import (
	"testing"
	"time"
)

func TestPingRedis(t *testing.T) {
	rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer rc.Close()

	err = rc.PingRedis()
	if err == nil {
		t.Error("Expected ping to fail with invalid port")
	}
	t.Logf("Expected ping failure: %v", err)
}

func TestIsRedisAvailable(t *testing.T) {
	rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer rc.Close()

	if rc.IsRedisAvailable() {
		t.Error("Expected Redis to be unavailable with invalid port")
	}
}

func TestSetWithExpiration(t *testing.T) {
	rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer rc.Close()

	err = rc.SetWithExpiration("test", "value", time.Hour)
	if err == nil {
		t.Error("Expected set to fail with invalid Redis connection")
	}
}

func TestGetOrSet(t *testing.T) {
	rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer rc.Close()

	value, err := rc.GetOrSet("test", "default")
	if err == nil {
		t.Error("Expected get-or-set to fail with invalid Redis connection")
	}
	if value != "default" {
		t.Errorf("Expected default value, got %s", value)
	}
}

func TestIncrementAtomic(t *testing.T) {
	rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer rc.Close()

	_, err = rc.IncrementAtomic("counter", 1)
	if err == nil {
		t.Error("Expected increment to fail with invalid Redis connection")
	}
}

func TestBatchDelete(t *testing.T) {
	rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer rc.Close()

	// Test empty keys
	count, err := rc.BatchDelete([]string{})
	if err != nil {
		t.Errorf("Unexpected error with empty keys: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 deletions, got %d", count)
	}

    // Test with keys
    keys := []string{"key1", "key2", "key3"}
    _, err = rc.BatchDelete(keys)
    if err == nil {
        t.Error("Expected batch delete to fail with invalid Redis connection")
    }
}

func TestBatchExpire(t *testing.T) {
	rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer rc.Close()

	// Test empty keys
	err = rc.BatchExpire([]string{}, time.Hour)
	if err != nil {
		t.Errorf("Unexpected error with empty keys: %v", err)
	}

	// Test with keys
	keys := []string{"key1", "key2"}
	err = rc.BatchExpire(keys, time.Hour)
	if err == nil {
		t.Error("Expected batch expire to fail with invalid Redis connection")
	}
}

func TestGetClientInfo(t *testing.T) {
	rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer rc.Close()

	info, err := rc.GetClientInfo()
	if err == nil {
		t.Error("Expected get client info to fail with invalid Redis connection")
	}
	if info != nil {
		t.Error("Expected nil info on failure")
	}
}

func TestGetRedisVersion(t *testing.T) {
	rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer rc.Close()

	version, err := rc.GetRedisVersion()
	if err == nil {
		t.Error("Expected get Redis version to fail with invalid Redis connection")
	}
	if version != "" {
		t.Error("Expected empty version on failure")
	}
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name    string
		config  *option
		wantErr bool
	}{
		{
			name: "valid config",
			config: &option{
				Addrs:    []string{"localhost:6379"},
				DB:       0,
				PoolSize: 10,
			},
			wantErr: false,
		},
		{
			name: "empty addrs",
			config: &option{
				Addrs:    []string{},
				DB:       0,
				PoolSize: 10,
			},
			wantErr: true,
		},
		{
			name: "negative pool size",
			config: &option{
				Addrs:    []string{"localhost:6379"},
				DB:       0,
				PoolSize: -1,
			},
			wantErr: true,
		},
		{
			name: "DB out of range",
			config: &option{
				Addrs:    []string{"localhost:6379"},
				DB:       16,
				PoolSize: 10,
			},
			wantErr: true,
		},
		{
			name: "valid DB range",
			config: &option{
				Addrs:    []string{"localhost:6379"},
				DB:       15,
				PoolSize: 10,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfig(tc.config)
			if tc.wantErr && err == nil {
				t.Errorf("Expected error for config %+v", tc.config)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error for valid config: %v", err)
			}
		})
	}
}