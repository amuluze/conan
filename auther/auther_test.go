package auther

import (
    "context"
    "errors"
    "fmt"
    "runtime"
    "testing"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

// newTestAuther 构造一个用于单元测试的 Auther 实例，
// 通过传入的配置参数控制令牌过期与清理协程的频率，便于观察行为。
func newTestAuther(t *testing.T, cfg AutherConfig) Auther {
    t.Helper()
    a, err := NewAuther(&cfg)
    if err != nil {
        t.Fatalf("NewAuther failed: %v", err)
    }
    return a
}

// TestCloseStopsCleanup 验证 Close 可以安全停止后台清理协程，
// 并且重复调用不会产生 panic（以 goroutine 数量作为近似观测指标）。
func TestCloseStopsCleanup(t *testing.T) {
    base := runtime.NumGoroutine()
    cfg := AutherConfig{
        SecretKey:               "secret",
        AccessTokenExp:          2 * time.Hour,
        RefreshTokenExp:         24 * time.Hour,
        Issuer:                  "test-issuer",
        BlackListEnabled:        true,
        BlackListCleanupInterval: 20 * time.Millisecond,
    }
    a := newTestAuther(t, cfg)
    // 等待一会儿，让协程启动并运行至少一次
    time.Sleep(60 * time.Millisecond)

    // 关闭协程
    if ja, ok := a.(*jwtAuther); ok {
        ja.Close()
        // 重复关闭也不应 panic
        ja.Close()
    } else {
        t.Fatalf("unexpected auther type")
    }

    // 再等待一会儿，观察 goroutine 数量不再持续增长
    time.Sleep(50 * time.Millisecond)
    after := runtime.NumGoroutine()
    // 允许存在少量噪声，但不应比基线高太多（此断言为近似值）
    if after > base+5 {
        t.Fatalf("goroutines not stabilized after Close: base=%d after=%d", base, after)
    }
}

// TestValidateTokenExpired 验证过期令牌返回 ErrExpiredToken，且实现不依赖错误字符串匹配。
func TestValidateTokenExpired(t *testing.T) {
    cfg := AutherConfig{
        SecretKey:               "secret",
        AccessTokenExp:          50 * time.Millisecond,
        RefreshTokenExp:         1 * time.Hour,
        Issuer:                  "test-issuer",
        BlackListEnabled:        false,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    a := newTestAuther(t, cfg)

    // 仅生成访问令牌（MintAccessToken 禁止生成刷新令牌）
    ti, err := a.MintAccessToken(context.Background(), "u1", "user1", "role1", cfg.AccessTokenExp, map[string]string{"k": "v"})
    if err != nil {
        t.Fatalf("MintAccessToken failed: %v", err)
    }
    // 等待超过过期时间
    time.Sleep(80 * time.Millisecond)
    _, err = a.ValidateToken(context.Background(), ti.Token)
    if err == nil {
        t.Fatalf("expected expired error, got nil")
    }
    if !errorsIs(err, ErrExpiredToken) {
        t.Fatalf("expected ErrExpiredToken, got: %v", err)
    }
}

// TestRevokeTokenAndValidation 验证黑名单撤销后的令牌在验证时返回 ErrRevokedToken。
func TestRevokeTokenAndValidation(t *testing.T) {
    cfg := AutherConfig{
        SecretKey:               "secret",
        AccessTokenExp:          1 * time.Hour,
        RefreshTokenExp:         2 * time.Hour,
        Issuer:                  "test-issuer",
        BlackListEnabled:        true,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    a := newTestAuther(t, cfg)
    ti, err := a.MintAccessToken(context.Background(), "u2", "user2", "role2", cfg.AccessTokenExp, nil)
    if err != nil {
        t.Fatalf("MintAccessToken failed: %v", err)
    }

    if err := a.RevokeToken(context.Background(), ti.Token); err != nil {
        t.Fatalf("RevokeToken failed: %v", err)
    }

    _, err = a.ValidateToken(context.Background(), ti.Token)
    if err == nil || !errorsIs(err, ErrRevokedToken) {
        t.Fatalf("expected ErrRevokedToken, got: %v", err)
    }
}

// TestBlacklistDisabledBehavior 验证当 BlackListEnabled=false 时，撤销操作不会生效，验证仍通过。
func TestBlacklistDisabledBehavior(t *testing.T) {
    cfg := AutherConfig{
        SecretKey:               "secret",
        AccessTokenExp:          1 * time.Hour,
        RefreshTokenExp:         2 * time.Hour,
        Issuer:                  "test-issuer",
        BlackListEnabled:        false,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    a := newTestAuther(t, cfg)
    ja := a.(*jwtAuther)

    ti, err := a.MintAccessToken(context.Background(), "u3", "user3", "role3", cfg.AccessTokenExp, nil)
    if err != nil {
        t.Fatalf("MintAccessToken failed: %v", err)
    }

    // 尝试撤销（应被忽略）
    if err := a.RevokeToken(context.Background(), ti.Token); err != nil {
        t.Fatalf("RevokeToken should be noop when blacklist disabled, got err: %v", err)
    }
    if ja.blackList.Size() != 0 {
        t.Fatalf("blacklist should remain empty when disabled")
    }

    // 验证应通过（未被撤销）
    if _, err := a.ValidateToken(context.Background(), ti.Token); err != nil {
        t.Fatalf("ValidateToken should pass when blacklist disabled, got: %v", err)
    }
}

// errorsIs 是对 errors.Is 的轻量封装，便于在测试中断言错误类型。
func errorsIs(err error, target error) bool { return err != nil && target != nil && (func() bool { return errorsIsStd(err, target) })() }

// errorsIsStd 使用标准库 errors.Is 进行错误比较的封装。
func errorsIsStd(err error, target error) bool { return errors.Is(err, target) }

// TestRefreshTokenRotationEnabled 验证当黑名单启用时，刷新令牌旋转会撤销旧刷新令牌，并签发新的令牌对。
func TestRefreshTokenRotationEnabled(t *testing.T) {
    cfg := AutherConfig{
        SecretKey:               "secret",
        AccessTokenExp:          1 * time.Hour,
        RefreshTokenExp:         2 * time.Hour,
        Issuer:                  "test-issuer",
        BlackListEnabled:        true,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    a := newTestAuther(t, cfg)

    pair, err := a.GenerateTokenPair(context.Background(), "u10", "user10", "role10", map[string]string{"x": "y"})
    if err != nil {
        t.Fatalf("GenerateTokenPair failed: %v", err)
    }

    // 执行刷新令牌旋转
    newPair, err := a.RefreshTokenRotate(context.Background(), pair.RefreshToken.Token)
    if err != nil {
        t.Fatalf("RefreshTokenRotate failed: %v", err)
    }

    // 旧刷新令牌应被撤销
    if _, err := a.ValidateToken(context.Background(), pair.RefreshToken.Token); err == nil || !errorsIs(err, ErrRevokedToken) {
        t.Fatalf("expected old refresh token revoked, got: %v", err)
    }

    // 新刷新令牌与访问令牌应有效
    if _, err := a.ValidateToken(context.Background(), newPair.RefreshToken.Token); err != nil {
        t.Fatalf("new refresh token should be valid, got: %v", err)
    }
    if _, err := a.ValidateToken(context.Background(), newPair.AccessToken.Token); err != nil {
        t.Fatalf("new access token should be valid, got: %v", err)
    }

    // 再次使用旧刷新令牌进行旋转应失败（被撤销）
    if _, err := a.RefreshTokenRotate(context.Background(), pair.RefreshToken.Token); err == nil || !errorsIs(err, ErrRevokedToken) {
        t.Fatalf("expected rotation with old token to fail with ErrRevokedToken, got: %v", err)
    }
}

// TestRefreshTokenRotationDisabled 验证当黑名单禁用时，刷新令牌旋转仍会生成新的令牌对，但无法撤销旧刷新令牌。
func TestRefreshTokenRotationDisabled(t *testing.T) {
    cfg := AutherConfig{
        SecretKey:               "secret",
        AccessTokenExp:          1 * time.Hour,
        RefreshTokenExp:         2 * time.Hour,
        Issuer:                  "test-issuer",
        BlackListEnabled:        false,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    a := newTestAuther(t, cfg)

    pair, err := a.GenerateTokenPair(context.Background(), "u11", "user11", "role11", nil)
    if err != nil {
        t.Fatalf("GenerateTokenPair failed: %v", err)
    }

    newPair, err := a.RefreshTokenRotate(context.Background(), pair.RefreshToken.Token)
    if err != nil {
        t.Fatalf("RefreshTokenRotate failed: %v", err)
    }

    // 旧刷新令牌仍然有效（因为未启用黑名单）
    if _, err := a.ValidateToken(context.Background(), pair.RefreshToken.Token); err != nil {
        t.Fatalf("old refresh token should still be valid when blacklist disabled, got: %v", err)
    }
    // 新刷新令牌与访问令牌也应有效
    if _, err := a.ValidateToken(context.Background(), newPair.RefreshToken.Token); err != nil {
        t.Fatalf("new refresh token should be valid, got: %v", err)
    }
    if _, err := a.ValidateToken(context.Background(), newPair.AccessToken.Token); err != nil {
        t.Fatalf("new access token should be valid, got: %v", err)
    }
}

// TestValidateTokenIssuerMismatch_SignatureValid 验证签名正确但 Issuer 不匹配时返回 ErrInvalidToken。
func TestValidateTokenIssuerMismatch_SignatureValid(t *testing.T) {
    cfgA := AutherConfig{
        SecretKey:                "secret",
        AccessTokenExp:           1 * time.Hour,
        RefreshTokenExp:          2 * time.Hour,
        Issuer:                   "issuerA",
        BlackListEnabled:         false,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    cfgB := AutherConfig{
        SecretKey:                "secret", // 相同密钥保证签名可通过
        AccessTokenExp:           1 * time.Hour,
        RefreshTokenExp:          2 * time.Hour,
        Issuer:                   "issuerB",
        BlackListEnabled:         false,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    aA := newTestAuther(t, cfgA)
    aB := newTestAuther(t, cfgB)

    // 使用 issuerB 签发令牌，但用 issuerA 的认证器验证，应失败（Issuer/Audience 均不匹配）。
    ti, err := aB.MintAccessToken(context.Background(), "uX", "userX", "roleX", cfgB.AccessTokenExp, nil)
    if err != nil {
        t.Fatalf("MintAccessToken failed: %v", err)
    }
    if _, err := aA.ValidateToken(context.Background(), ti.Token); err == nil || !errorsIs(err, ErrInvalidToken) {
        t.Fatalf("expected ErrInvalidToken due to issuer mismatch, got: %v", err)
    }
}

// TestValidateTokenAudienceMismatch_SignatureValid 验证签名正确且 Issuer 匹配但 Audience 不包含配置 Issuer 时返回 ErrInvalidToken。
func TestValidateTokenAudienceMismatch_SignatureValid(t *testing.T) {
    cfg := AutherConfig{
        SecretKey:                "secret",
        AccessTokenExp:           1 * time.Hour,
        RefreshTokenExp:          2 * time.Hour,
        Issuer:                   "issuerA",
        BlackListEnabled:         false,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    a := newTestAuther(t, cfg)

    now := time.Now()
    claims := &TokenClaims{
        UserID:   "uY",
        Username: "userY",
        Role:     "roleY",
        Type:     AccessToken,
        Metadata: nil,
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        "test-jti",
            Issuer:    cfg.Issuer,                  // Issuer 匹配
            Subject:   "uY",
            Audience:  []string{"other-aud"},      // Audience 不包含配置 Issuer
            ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
            NotBefore: jwt.NewNumericDate(now),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenStr, err := token.SignedString([]byte(cfg.SecretKey))
    if err != nil {
        t.Fatalf("failed to sign token: %v", err)
    }

    if _, err := a.ValidateToken(context.Background(), tokenStr); err == nil || !errorsIs(err, ErrInvalidToken) {
        t.Fatalf("expected ErrInvalidToken due to audience mismatch, got: %v", err)
    }
}

// TestValidateTokenIssuerMismatch_ParseErrorPath 验证在解析失败路径中，Issuer/Audience 不匹配也返回 ErrInvalidToken。
func TestValidateTokenIssuerMismatch_ParseErrorPath(t *testing.T) {
    cfgA := AutherConfig{
        SecretKey:                "secretA",
        AccessTokenExp:           1 * time.Hour,
        RefreshTokenExp:          2 * time.Hour,
        Issuer:                   "issuerA",
        BlackListEnabled:         false,
        BlackListCleanupInterval: 1 * time.Hour,
    }
    aA := newTestAuther(t, cfgA)

    now := time.Now()
    claims := &TokenClaims{
        UserID:   "uZ",
        Username: "userZ",
        Role:     "roleZ",
        Type:     AccessToken,
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        "test-jti-z",
            Issuer:    "issuerB",                 // 与 aA 不匹配
            Subject:   "uZ",
            Audience:  []string{"issuerB"},       // 不包含 issuerA
            ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
            NotBefore: jwt.NewNumericDate(now),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }
    // 使用不同密钥签名，触发 ParseWithClaims 失败，从而进入回退校验路径
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenStr, err := token.SignedString([]byte("secretB"))
    if err != nil {
        t.Fatalf("failed to sign token: %v", err)
    }

    if _, err := aA.ValidateToken(context.Background(), tokenStr); err == nil || !errorsIs(err, ErrInvalidToken) {
        t.Fatalf("expected ErrInvalidToken in parse-error path due to issuer/audience mismatch, got: %v", err)
    }

    // 附加：若改为匹配 Issuer/Audience 但密钥仍不同，应返回解析错误而非 ErrInvalidToken
    claims.RegisteredClaims.Issuer = cfgA.Issuer
    claims.RegisteredClaims.Audience = []string{cfgA.Issuer}
    token2 := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenStr2, err := token2.SignedString([]byte("secretB"))
    if err != nil {
        t.Fatalf("failed to sign token2: %v", err)
    }
    if _, err := aA.ValidateToken(context.Background(), tokenStr2); err == nil || errorsIs(err, ErrInvalidToken) {
        t.Fatalf("expected parse error (signature invalid) without issuer/audience mismatch, got: %v", err)
    } else {
        // 打印一下具体错误类型，便于定位（不做强断言）
        _ = fmt.Sprintf("parse error: %v", err)
    }
}