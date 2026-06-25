package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// AccessTokenExpire Access Token 过期时间（管理后台短 TTL，禁用/改密后最长 15min 失效）
	AccessTokenExpire = 15 * time.Minute
	// RefreshTokenExpire Refresh Token 过期时间
	RefreshTokenExpire = 12 * time.Hour
	// Issuer JWT 签发者
	Issuer = "svr-admin"
	// Audience JWT 受众
	Audience = "nostack-admin"

	// RoleSuperAdmin 超级管理员角色码，拥有全部权限（跳过权限点校验）
	RoleSuperAdmin = "super_admin"
)

// Claims JWT 自定义声明
type Claims struct {
	AdminID     uint     `json:"admin_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"` // 角色码
	Permissions []string `json:"permissions,omitempty"`
	TokenType   string   `json:"token_type"` // access / refresh
	jwt.RegisteredClaims
}

// TokenPair 登录后返回的 Token 对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access Token 剩余秒数
	TokenType    string `json:"token_type"` // Bearer
}

// JWTManager JWT 管理器
type JWTManager struct {
	secret []byte
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{secret: []byte(secret)}
}

// GenerateTokenPair 生成 Access Token + Refresh Token 对
func (m *JWTManager) GenerateTokenPair(adminID uint, username, role string, permissions []string) (*TokenPair, error) {
	now := time.Now()

	accessClaims := Claims{
		AdminID:     adminID,
		Username:    username,
		Role:        role,
		Permissions: permissions,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateJTI(),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    Issuer,
			Audience:  []string{Audience},
		},
	}
	accessStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(m.secret)
	if err != nil {
		return nil, err
	}

	refreshClaims := Claims{
		AdminID:   adminID,
		Username:  username,
		Role:      role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateJTI(),
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    Issuer,
			Audience:  []string{Audience},
		},
	}
	refreshStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(m.secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    int64(AccessTokenExpire.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ParseToken 解析 JWT Token
func (m *JWTManager) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ParseAccessToken 解析并校验 Access Token
func (m *JWTManager) ParseAccessToken(tokenStr string) (*Claims, error) {
	claims, err := m.ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "access" {
		return nil, errors.New("not an access token")
	}
	return claims, nil
}

// ParseRefreshToken 解析并校验 Refresh Token
func (m *JWTManager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	claims, err := m.ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "refresh" {
		return nil, errors.New("not a refresh token")
	}
	return claims, nil
}

// generateJTI 生成唯一 Token ID
func generateJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
