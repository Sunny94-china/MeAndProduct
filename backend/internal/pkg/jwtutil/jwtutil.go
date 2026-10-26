package jwtutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"meandproduct/internal/config"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	UserID    uint64 `json:"uid"`
	Role      string `json:"role"`
	TokenType string `json:"ttp"`
	JWTID     string `json:"jti"`
	Exp       int64  `json:"exp"`
	Iat       int64  `json:"iat"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken AccessToken
	RefreshToken
}

type AccessToken struct {
	Token     string
	ExpiresAt time.Time
}
type RefreshToken struct {
	Token     string
	ExpiresAt time.Time
}

// IssueTokens 生成 access + refresh token 对
func IssueTokens(cfg *config.Config, userID uint64, role string) (*TokenPair, error) {
	if cfg.JWT.Secret == "" {
		return nil, errors.New("jwt secret is empty")
	}
	now := time.Now()
	accessJti := uuid.NewString()
	refreshJti := uuid.NewString()

	accessClaims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: TokenTypeAccess,
		JWTID:     accessJti,
		Iat:       now.Unix(),
		Exp:       now.Add(cfg.JWT.AccessTTL).Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JWT.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        accessJti,
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWT.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshClaims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: TokenTypeRefresh,
		JWTID:     refreshJti,
		Iat:       now.Unix(),
		Exp:       now.Add(cfg.JWT.RefreshTTL).Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JWT.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        refreshJti,
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWT.RefreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	sign := jwt.SigningMethodHS256
	secret := []byte(cfg.JWT.Secret)

	accessTok, err := jwt.NewWithClaims(sign, accessClaims).SignedString(secret)
	if err != nil {
		return nil, fmt.Errorf("sign access: %w", err)
	}
	refreshTok, err := jwt.NewWithClaims(sign, refreshClaims).SignedString(secret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh: %w", err)
	}

	return &TokenPair{
		AccessToken:  AccessToken{Token: accessTok, ExpiresAt: now.Add(cfg.JWT.AccessTTL)},
		RefreshToken: RefreshToken{Token: refreshTok, ExpiresAt: now.Add(cfg.JWT.RefreshTTL)},
	}, nil
}

// Parse 验证并解析 token，expectType 传空字符串表示不校验类型
func Parse(cfg *config.Config, tokenStr string, expectType string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid claims")
	}
	if expectType != "" && c.TokenType != expectType {
		return nil, fmt.Errorf("expect token type %s but got %s", expectType, c.TokenType)
	}
	return c, nil
}

// Blacklist 将一个jti放入Redis
func Blacklist(ctx context.Context, cfg *config.Config, rd *redis.Client, jti string, expire time.Duration) error {
	key := cfg.JWT.BlacklistKeyPrefix + jti
	return rd.Set(ctx, key, 1, expire).Err()
}

// IsBlacklisted 检查 jti 是否已加入黑名单
func IsBlacklisted(ctx context.Context, cfg *config.Config, rd *redis.Client, jti string) (bool, error) {
	key := cfg.JWT.BlacklistKeyPrefix + jti
	n, err := rd.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
