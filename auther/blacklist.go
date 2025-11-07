package auther

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// BlackListItem 黑名单项
type BlackListItem struct {
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// BlackList 黑名单管理器
type BlackList struct {
    items map[string]*BlackListItem
    mu    sync.RWMutex
}

// NewBlackList 创建新的黑名单
func NewBlackList() *BlackList {
	return &BlackList{
		items: make(map[string]*BlackListItem),
	}
}

// Add 添加令牌到黑名单
func (bl *BlackList) Add(ctx context.Context, token string, userID string, expiresAt time.Time) error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	tokenHash := bl.hashToken(token)
	bl.items[tokenHash] = &BlackListItem{
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	return nil
}


// IsRevoked 检查令牌是否被撤销
func (bl *BlackList) IsRevoked(ctx context.Context, token string) (bool, error) {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	tokenHash := bl.hashToken(token)
	item, exists := bl.items[tokenHash]
	if !exists {
		return false, nil
	}

	// 检查是否已过期
	if time.Now().After(item.ExpiresAt) {
		return false, nil
	}

	return true, nil
}



// Cleanup 清理过期的黑名单项
func (bl *BlackList) Cleanup(ctx context.Context) error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	for tokenHash, item := range bl.items {
		if now.After(item.ExpiresAt) {
			delete(bl.items, tokenHash)
		}
	}

	return nil
}

// Size 获取黑名单大小
func (bl *BlackList) Size() int {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return len(bl.items)
}

// hashToken 生成令牌哈希
func (bl *BlackList) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}