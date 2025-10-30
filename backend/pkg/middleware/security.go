package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	SecretKey    string        // 签名密钥
	ReplayWindow time.Duration // 重放攻击时间窗口
	EnableReplay bool          // 是否启用防重放
	EnableSign   bool          // 是否启用防篡改
}

// SecurityMiddleware 安全中间件，实现防篡改和防重放攻击
func SecurityMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		config := &SecurityConfig{
			SecretKey:    secret,
			ReplayWindow: common.ReplayWindow,
			EnableReplay: common.EnableReplay,
			EnableSign:   common.EnableSign,
		}

		// 防重放攻击检查
		if config.EnableReplay {
			if err := checkReplayAttack(c, config.ReplayWindow); err != nil {
				logger.Warn("防重放检查失败", zap.Error(err), zap.String("path", c.Request.URL.Path))
				i18n.HandleSignatureError(c, "请求已过期或重复")
				c.Abort()
				return
			}
		}

		// 防篡改检查
		if config.EnableSign {
			if err := checkSignature(c, config.SecretKey); err != nil {
				logger.Warn("签名验证失败", zap.Error(err), zap.String("path", c.Request.URL.Path))
				i18n.HandleSignatureError(c, "签名验证失败")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// checkReplayAttack 检查重放攻击
func checkReplayAttack(c *gin.Context, window time.Duration) error {
	timestampStr := c.GetHeader("X-Timestamp")
	if timestampStr == "" {
		return fmt.Errorf("缺少时间戳")
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("时间戳格式错误")
	}

	requestTime := time.Unix(timestamp, 0)
	now := time.Now()

	// 检查时间窗口
	if now.Sub(requestTime) > window {
		return fmt.Errorf("请求已过期")
	}

	// 检查时间是否过于超前
	if requestTime.Sub(now) > time.Minute {
		return fmt.Errorf("请求时间过于超前")
	}

	// TODO: 这里可以添加nonce检查，防止相同时间戳的重复请求
	// 可以使用Redis或内存缓存存储已使用的nonce
	// redisClient := redis.GetClient()
	// nonceKey := fmt.Sprintf("nonce:%d:%s", timestamp, c.Request.URL.Path)

	// // 检查nonce是否已存在
	// if _, err := redisClient.Get(nonceKey); err == nil {
	// 	return fmt.Errorf("重复请求")
	// }

	// // 缓存nonce，过期时间为窗口长度
	// if err := redisClient.Set(nonceKey, "used", window); err != nil {
	// 	return fmt.Errorf("缓存nonce失败: %v", err)
	// }
	return nil
}

// checkSignature 检查签名
func checkSignature(c *gin.Context, secretKey string) error {
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		return fmt.Errorf("缺少签名")
	}

	// 构建签名字符串
	signString, err := buildSignString(c)
	if err != nil {
		return fmt.Errorf("构建签名字符串失败: %v", err)
	}

	// 计算期望的签名
	expectedSignature := calculateSignature(signString, secretKey)

	// 验证签名
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("签名不匹配")
	}

	return nil
}

// buildSignString 构建签名字符串
func buildSignString(c *gin.Context) (string, error) {
	var parts []string

	// 添加HTTP方法
	parts = append(parts, c.Request.Method)

	// 添加路径
	parts = append(parts, c.Request.URL.Path)

	// 添加时间戳
	timestamp := c.GetHeader("X-Timestamp")
	if timestamp != "" {
		parts = append(parts, timestamp)
	}

	// 添加查询参数（按字母顺序排序）
	if len(c.Request.URL.RawQuery) > 0 {
		queryParams := make([]string, 0)
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
			}
		}
		sort.Strings(queryParams)
		parts = append(parts, strings.Join(queryParams, "&"))
	}

	// 添加请求体（如果是POST/PUT等方法）
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		body := c.GetHeader("X-Body-Hash")
		if body != "" {
			parts = append(parts, body)
		}
	}

	return strings.Join(parts, "|"), nil
}

// calculateSignature 计算签名
func calculateSignature(data, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
