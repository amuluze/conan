package auther

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "fmt"
    "sync"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrRevokedToken     = errors.New("token revoked")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrSecretKeyEmpty   = errors.New("secret key is required")
)

// jwtAuther JWT 认证器实现
type jwtAuther struct {
    config    AutherConfig
    blackList *BlackList
    // stopChan 用于控制后台黑名单清理协程的生命周期，实现显式关闭资源。
    stopChan  chan struct{}
    // closeOnce 确保 Close 只执行一次，避免重复关闭通道导致的 panic。
    closeOnce sync.Once
}

// NewAuther 创建新的认证器
// 修复点：
// 1) 统一覆盖传入配置到内部 config（包含 BlackListEnabled），避免默认值与调用方意图不一致；
// 2) 使用内部 config 决定是否启动清理协程，并初始化 stopChan，便于后续 Close 停止协程。
func NewAuther(authConfig *AutherConfig) (Auther, error) {
    if authConfig == nil || strings.TrimSpace(authConfig.SecretKey) == "" {
        return nil, ErrSecretKeyEmpty
    }

    // 从默认配置出发，覆盖调用方提供的配置
    config := DefaultAutherConfig
    config.SecretKey = strings.TrimSpace(authConfig.SecretKey)

    // 设置默认值
    if authConfig.AccessTokenExp > 0 {
        config.AccessTokenExp = authConfig.AccessTokenExp
    }
    if authConfig.RefreshTokenExp > 0 {
        config.RefreshTokenExp = authConfig.RefreshTokenExp
    }
    if authConfig.Issuer != "" {
        config.Issuer = strings.TrimSpace(authConfig.Issuer)
    }
    // 修复：覆盖 BlackListEnabled，保持与调用方配置一致
    config.BlackListEnabled = authConfig.BlackListEnabled
    if authConfig.BlackListCleanupInterval > 0 {
        config.BlackListCleanupInterval = authConfig.BlackListCleanupInterval
    }

    auther := &jwtAuther{
        config:    config,
        blackList: NewBlackList(),
        stopChan:  make(chan struct{}),
    }

    // 启动黑名单清理协程
    if auther.config.BlackListEnabled {
        go auther.startBlackListCleanup()
    }

    return auther, nil
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (a *jwtAuther) GenerateTokenPair(ctx context.Context, userID, username, role string, metadata map[string]string) (*TokenPair, error) {
    // 生成访问令牌（使用 MintAccessToken，仅允许 AccessToken）
    accessToken, err := a.MintAccessToken(ctx, userID, username, role, a.config.AccessTokenExp, metadata)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }

    // 生成刷新令牌（仅内部允许生成）
    refreshToken, err := a.generateTokenInternal(ctx, userID, username, role, RefreshToken, a.config.RefreshTokenExp, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }

    return &TokenPair{
        AccessToken:  *accessToken,
        RefreshToken: *refreshToken,
    }, nil
}

// generateTokenInternal 生成指定类型的令牌（内部方法）。
// 该方法不暴露在接口中，用于在内部生成 RefreshToken，避免业务层误用。
func (a *jwtAuther) generateTokenInternal(ctx context.Context, userID, username, role string, tokenType TokenType, exp time.Duration, metadata map[string]string) (*TokenInfo, error) {
    now := time.Now()
    claims := &TokenClaims{
        UserID:   userID,
        Username: username,
        Role:     role,
        Type:     tokenType,
        Metadata: metadata,
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        generateJTI(),
            Issuer:    a.config.Issuer,
            Subject:   userID,
            Audience:  []string{a.config.Issuer},
            ExpiresAt: jwt.NewNumericDate(now.Add(exp)),
            NotBefore: jwt.NewNumericDate(now),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(a.config.SecretKey))
    if err != nil {
        return nil, fmt.Errorf("failed to sign token: %w", err)
    }

    return &TokenInfo{
        Token:     tokenString,
        Type:      tokenType,
        ExpiresAt: claims.ExpiresAt.Time,
        UserID:    userID,
        Username:  username,
        Role:      role,
        Metadata:  metadata,
    }, nil
}

// MintAccessToken 生成访问令牌（仅限 AccessToken）。
// 注意：此方法不接受 tokenType 参数，始终生成 AccessToken；禁止生成 RefreshToken。
func (a *jwtAuther) MintAccessToken(ctx context.Context, userID, username, role string, exp time.Duration, metadata map[string]string) (*TokenInfo, error) {
    return a.generateTokenInternal(ctx, userID, username, role, AccessToken, exp, metadata)
}

// ValidateToken 验证令牌
// 职责：
// 1) 在启用黑名单时检查令牌是否被撤销；
// 2) 校验签名与标准声明；
// 3) 稳定判定过期与 NotBefore；
// 4) 显式校验 Issuer 与 Audience，避免跨签发者令牌被误接受。
func (a *jwtAuther) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
    // 检查是否在黑名单中
    if a.config.BlackListEnabled {
        revoked, err := a.blackList.IsRevoked(ctx, token)
        if err != nil {
            return nil, fmt.Errorf("failed to check token revocation status: %w", err)
        }
        if revoked {
            return nil, ErrRevokedToken
        }
    }

    // 解析和验证令牌
    claims := &TokenClaims{}
    parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
        // 验证签名方法
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(a.config.SecretKey), nil
    })

    if err != nil {
        // 改进：避免依赖错误字符串，改用不验证签名/时间的解析提取 claims，进行稳定的过期与 NotBefore 判定
        tmpClaims := &TokenClaims{}
        _, _, _ = jwt.NewParser(jwt.WithoutClaimsValidation()).ParseUnverified(token, tmpClaims)
        // 过期判定
        if tmpClaims.ExpiresAt != nil && time.Now().After(tmpClaims.ExpiresAt.Time) {
            return nil, ErrExpiredToken
        }
        // NotBefore 判定：当前时间早于 NotBefore 的令牌视为尚未生效
        if tmpClaims.NotBefore != nil && time.Now().Before(tmpClaims.NotBefore.Time) {
            return nil, ErrInvalidToken
        }
        // Issuer/Audience 近似校验（在解析失败时尽力而为）：不匹配视为无效
        if tmpClaims.Issuer != "" && a.config.Issuer != "" && tmpClaims.Issuer != a.config.Issuer {
            return nil, ErrInvalidToken
        }
        if len(tmpClaims.Audience) > 0 && a.config.Issuer != "" {
            audOk := false
            for _, aud := range tmpClaims.Audience {
                if aud == a.config.Issuer {
                    audOk = true
                    break
                }
            }
            if !audOk {
                return nil, ErrInvalidToken
            }
        }
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

    // 补充 NotBefore 判定：尽管库已校验标准声明，这里显式检查以提升健壮性
    if claims.NotBefore != nil && time.Now().Before(claims.NotBefore.Time) {
        return nil, ErrInvalidToken
    }

    // 显式校验 Issuer：必须与配置一致
    if a.config.Issuer != "" && claims.Issuer != a.config.Issuer {
        return nil, ErrInvalidToken
    }

    // 显式校验 Audience：必须包含配置中的 Issuer（生成时设置为该值）
    if a.config.Issuer != "" {
        audOk := false
        for _, aud := range claims.Audience {
            if aud == a.config.Issuer {
                audOk = true
                break
            }
        }
        if !audOk {
            return nil, ErrInvalidToken
        }
    }

	return claims, nil
}

// RefreshTokenRotate 刷新令牌旋转
// 简介：验证刷新令牌并（在启用黑名单时）撤销旧令牌，随后签发新的访问令牌与刷新令牌对并返回。

// RefreshTokenRotate 刷新令牌旋转
// 功能：
// 1) 验证传入的刷新令牌有效性与类型；
// 2) 当启用黑名单时，立即撤销旧刷新令牌，防止后续重放；
// 3) 基于现有 claims 签发新的访问令牌与刷新令牌，并返回令牌对；
// 4) 当 BlackListEnabled=false 时无法撤销旧刷新令牌，但仍会生成新的令牌对。
func (a *jwtAuther) RefreshTokenRotate(ctx context.Context, refreshToken string) (*TokenPair, error) {
    // 验证刷新令牌
    claims, err := a.ValidateToken(ctx, refreshToken)
    if err != nil {
        return nil, fmt.Errorf("invalid refresh token: %w", err)
    }
    if claims.Type != RefreshToken {
        return nil, ErrInvalidTokenType
    }

    // 撤销旧刷新令牌（若启用黑名单）
    if a.config.BlackListEnabled {
        if err := a.RevokeToken(ctx, refreshToken); err != nil {
            return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
        }
    }

    // 生成新的访问令牌（通过 MintAccessToken）与刷新令牌（内部生成）
    newAccess, err := a.MintAccessToken(ctx, claims.UserID, claims.Username, claims.Role, a.config.AccessTokenExp, claims.Metadata)
    if err != nil {
        return nil, fmt.Errorf("failed to generate new access token: %w", err)
    }
    newRefresh, err := a.generateTokenInternal(ctx, claims.UserID, claims.Username, claims.Role, RefreshToken, a.config.RefreshTokenExp, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
    }

    return &TokenPair{AccessToken: *newAccess, RefreshToken: *newRefresh}, nil
}

// RevokeToken 撤销令牌（加入黑名单）
func (a *jwtAuther) RevokeToken(ctx context.Context, token string) error {
	if !a.config.BlackListEnabled {
		return nil // 如果黑名单未启用，则无法撤销令牌
	}

	// 解析令牌获取信息
	claims, err := a.GetTokenInfo(token)
	if err != nil {
		return fmt.Errorf("failed to get token info: %w", err)
	}

	// 添加到黑名单
	return a.blackList.Add(ctx, token, claims.UserID, claims.ExpiresAt.Time)
}


// IsTokenRevoked 检查令牌是否被撤销
func (a *jwtAuther) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	if !a.config.BlackListEnabled {
		return false, nil
	}

	return a.blackList.IsRevoked(ctx, token)
}

// CleanupExpiredTokens 清理过期的黑名单令牌
func (a *jwtAuther) CleanupExpiredTokens(ctx context.Context) error {
	if !a.config.BlackListEnabled {
		return nil
	}

	return a.blackList.Cleanup(ctx)
}

// GetTokenInfo 从令牌中提取信息（不验证签名）
func (a *jwtAuther) GetTokenInfo(token string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return claims, nil
}

// startBlackListCleanup 启动黑名单清理协程（周期清理过期项，支持显式关闭）。
func (a *jwtAuther) startBlackListCleanup() {
    ticker := time.NewTicker(a.config.BlackListCleanupInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            ctx := context.Background()
            if err := a.CleanupExpiredTokens(ctx); err != nil {
                // TODO: 可接入日志记录错误，但不中断清理过程
            }
        case <-a.stopChan:
            // 收到关闭信号，退出协程
            return
        }
    }
}

// Close 关闭后台黑名单清理协程（使用 sync.Once 保证幂等）。
func (a *jwtAuther) Close() error {
    if a.stopChan == nil {
        return nil
    }
    a.closeOnce.Do(func() {
        close(a.stopChan)
    })
    return nil
}

// generateJTI 生成随机 JWT ID（熵不足时回退为时间戳）。
func generateJTI() string {
    var b [16]byte
    if _, err := rand.Read(b[:]); err == nil {
        return hex.EncodeToString(b[:])
    }
    // 回退：使用纳秒时间戳，避免因系统熵不足导致失败
    return fmt.Sprintf("jti_%d", time.Now().UnixNano())
}
