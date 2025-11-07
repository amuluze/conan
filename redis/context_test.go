// Package redis
// Date: 2025/11/07
// Author: Amu
// Description: 针对带上下文的 *Ctx 方法与 WithContext 封装的最小化测试
package redis

import (
    "context"
    "testing"
    "time"
)

// TestContextMethodsWithoutConnection 验证在无有效连接时，*Ctx 方法返回错误且不 panic。
func TestContextMethodsWithoutConnection(t *testing.T) {
    rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
    if err != nil {
        t.Fatalf("failed to create client without ping: %v", err)
    }
    defer rc.Close()

    // 使用带超时的 ctx，避免阻塞
    c, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    // 1) 字符串：SetEXCtx / GetCtx
    if _, err := rc.SetEXCtx(c, "ctx_key", "val", time.Second); err == nil {
        t.Errorf("expected error when SetEXCtx without valid redis, got nil")
    }
    if _, err := rc.GetCtx(c, "ctx_key"); err == nil {
        t.Errorf("expected error when GetCtx without valid redis, got nil")
    }

    // 2) 公共：ScanKeysCtx
    if _, err := rc.ScanKeysCtx(c, "*", 100); err == nil {
        t.Errorf("expected error when ScanKeysCtx without valid redis, got nil")
    }
}

// TestWithContextWrapper 验证 WithContext 封装的委托行为。
func TestWithContextWrapper(t *testing.T) {
    rc, err := NewClientWithoutPing(WithAddrs([]string{"localhost:9999"}))
    if err != nil {
        t.Fatalf("failed to create client without ping: %v", err)
    }
    defer rc.Close()

    c, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    w := rc.WithContext(c)

    // 通过 wrapper 执行常用方法，预期返回错误但不 panic
    if _, err := w.SetEX("ctx_key", "val", time.Second); err == nil {
        t.Errorf("expected error when SetEX via WithContext without valid redis, got nil")
    }
    if _, err := w.Get("ctx_key"); err == nil {
        t.Errorf("expected error when Get via WithContext without valid redis, got nil")
    }
}