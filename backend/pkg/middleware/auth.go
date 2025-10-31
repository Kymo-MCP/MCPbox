package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/jwt"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
)

var SkipPaths = []string{
	"/health",
	"/authz/encryption-key",
	"/authz/login",
	"/authz/logout",
	"/authz/register",
	"/authz/refresh",
	"/authz/validate",
	"/market/code/download",
}

// AuthTokenMiddleware 用户token验证中间件
func AuthTokenMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过登录等不需要认证的接口
		if shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		tokenString := extractToken(c)
		if tokenString == "" {
			i18n.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		claims, err := jwt.ParseTokenWithClaims(tokenString, secret)
		if err != nil {
			logger.Error("JWT令牌验证失败", zap.Error(err))
			i18n.Unauthorized(c, "无效的认证令牌")
			c.Abort()
			return
		}

		// 检查令牌是否过期
		if time.Now().Unix() > int64(claims.ExpiresAt.Unix()) {
			i18n.Unauthorized(c, "认证令牌已过期")
			c.Abort()
			return
		}

		if claims.UserID == 0 || len(claims.Username) == 0 {
			i18n.Unauthorized(c, "user id or username is empty")
			c.Abort()
			return
		}

		userToken, err := redis.GetUserTokenByToken(tokenString)
		if err != nil {
			logger.Error("获取用户令牌失败", zap.Error(err))
			i18n.Unauthorized(c, "无效的认证令牌")
			c.Abort()
			return
		}
		if userToken == nil || userToken.UserID != uint(claims.UserID) {
			i18n.Unauthorized(c, "user id not match")
			c.Abort()
			return
		}

		// 检查令牌是否有效
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// shouldSkipAuth 判断是否跳过认证
func shouldSkipAuth(path string) bool {
	for _, skipPath := range SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// extractToken 提取令牌
func extractToken(c *gin.Context) string {
	// 从Authorization头提取
	auth := c.GetHeader("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// 从查询参数提取
	token := c.Query("token")
	if token != "" {
		return token
	}

	return ""
}
