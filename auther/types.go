package auther

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType Token 类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// TokenInfo Token 信息
type TokenInfo struct {
	Token     string            `json:"token"`
	Type      TokenType         `json:"type"`
	ExpiresAt time.Time         `json:"expires_at"`
	UserID    string            `json:"user_id"`
	Username  string            `json:"username"`
	Role      string            `json:"role"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// TokenPair Token 对（访问令牌 + 刷新令牌）
type TokenPair struct {
	AccessToken  TokenInfo `json:"access_token"`
	RefreshToken TokenInfo `json:"refresh_token"`
}

// TokenClaims JWT Claims
type TokenClaims struct {
	UserID   string            `json:"user_id"`
	Username string            `json:"username"`
	Role     string            `json:"role"`
	Type     TokenType         `json:"type"`
	Metadata map[string]string `json:"metadata,omitempty"`
	jwt.RegisteredClaims
}

// AutherConfig Auther 配置
type AutherConfig struct {
	// SecretKey JWT 密钥
	SecretKey string
	// AccessTokenExp 访问令牌过期时间（例如：2*time.Hour）
	AccessTokenExp time.Duration
	// RefreshTokenExp 刷新令牌过期时间（例如：7*24*time.Hour）
	RefreshTokenExp time.Duration
	// Issuer 签发者
	Issuer string
	// BlackListEnabled 是否启用黑名单
	BlackListEnabled bool
	// BlackListCleanupInterval 黑名单清理间隔（例如：1*time.Hour）
	BlackListCleanupInterval time.Duration
}

// DefaultAutherConfig 默认配置
var DefaultAutherConfig = AutherConfig{
	AccessTokenExp:           2 * time.Hour,
	RefreshTokenExp:          7 * 24 * time.Hour,
	Issuer:                   "conan",
	BlackListEnabled:         true,
	BlackListCleanupInterval: 1 * time.Hour,
}

// Auther Token 管理器接口
type Auther interface {
	// GenerateTokenPair 生成访问令牌和刷新令牌对
	GenerateTokenPair(ctx context.Context, userID, username, role string, metadata map[string]string) (*TokenPair, error)

	// MintAccessToken 生成访问令牌（仅限 AccessToken）
	// 注意：此方法不接受 tokenType 参数，始终生成 AccessToken；禁止生成 RefreshToken。
	MintAccessToken(ctx context.Context, userID, username, role string, exp time.Duration, metadata map[string]string) (*TokenInfo, error)

	// ValidateToken 验证令牌
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)

	// RefreshTokenRotate 刷新令牌旋转：
	// 每次使用刷新令牌成功后，立即撤销旧刷新令牌并签发新的访问令牌与刷新令牌对，防止刷新令牌被重放。
	// 注意：当 BlackListEnabled=false 时无法撤销旧刷新令牌，旋转仅会生成新的令牌对。
	RefreshTokenRotate(ctx context.Context, refreshToken string) (*TokenPair, error)

	// RevokeToken 撤销令牌（加入黑名单）
	RevokeToken(ctx context.Context, token string) error

	// IsTokenRevoked 检查令牌是否被撤销
	IsTokenRevoked(ctx context.Context, token string) (bool, error)

	// CleanupExpiredTokens 清理过期的黑名单令牌
	CleanupExpiredTokens(ctx context.Context) error

	// GetTokenInfo 从令牌中提取信息（不验证签名）
	GetTokenInfo(token string) (*TokenClaims, error)

	// Close 关闭后台资源（如黑名单清理协程）。
	// 说明：若创建时未启用黑名单或未启动协程，Close 将安全地执行空操作；重复调用不会产生 panic。
	Close() error
}
