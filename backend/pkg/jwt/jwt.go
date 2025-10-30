package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config JWT配置
type Config struct {
	Secret  string        // JWT密钥
	Expires time.Duration // token过期时间
}

// Manager JWT管理器接口
type Manager interface {
	// GenerateToken 生成JWT token
	GenerateToken(userID int64, username string) (string, error)
	// ValidateToken 验证JWT token
	ValidateToken(tokenString string) (Claims, error)
	// GenerateRefreshToken 生成refresh token
	GenerateRefreshToken() (string, error)
}

// Claims JWT声明
type Claims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// manager JWT管理器实现
type manager struct {
	config *Config
}

// NewManager 创建JWT管理器
func NewManager(config *Config) Manager {
	return &manager{
		config: config,
	}
}

// GenerateToken 生成JWT token
func (m *manager) GenerateToken(userID int64, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.Expires)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// ValidateToken 验证JWT token
func (m *manager) ValidateToken(tokenString string) (Claims, error) {
	return ParseTokenWithClaims(tokenString, m.config.Secret)
}

// GenerateRefreshToken 生成refresh token
func (m *manager) GenerateRefreshToken() (string, error) {
	return generateRandomString(32)
}

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ParseTokenWithClaims 解析JWT token并提取Claims
func ParseTokenWithClaims(tokenString, secret string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return Claims{}, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return Claims{}, fmt.Errorf("invalid token claims")
	}

	return *claims, nil
}
