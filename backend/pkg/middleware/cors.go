package middleware

import (
	"net"
	"strings"

	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/utils"

	"github.com/gin-gonic/gin"
)

var defaultDomains = []string{
	"http://localhost",
	"http://127.0.0.1",
	"https://localhost",
	"https://127.0.0.1",
}

// CORSMiddleware CORS跨域中间件
func CORSMiddleware(domains []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 如果没有Origin头，或者Origin为空，则不是跨域请求
		if origin == "" {
			c.Next()
			return
		}

		// 检查Origin是否在允许列表中
		if isAllowedOrigin(origin, domains) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// 如果不在允许列表中, 则返回403 Forbidden
			c.Header("Access-Control-Allow-Origin", "*")
			i18n.Forbidden(c, "跨域请求被拒绝")
			c.Abort()
			return
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 预检请求结果缓存24小时

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			i18n.SuccessResponse(c, nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// isAllowedOrigin 检查Origin是否在允许列表中
func isAllowedOrigin(origin string, domains []string) bool {
	hostIps, _ := utils.GetHostIPs()
	if len(hostIps) > 0 {
		for _, ip := range hostIps {
			if strings.HasPrefix(origin, ip) {
				return true
			}
		}
	}
	for _, domain := range domains {
		if strings.HasPrefix(origin, domain) {
			return true
		}
		if domain == "*" {
			return true
		}
	}
	// 检查默认允许的来源
	for _, allowedOrigin := range defaultDomains {
		if strings.HasPrefix(origin, allowedOrigin) {
			return true
		}
	}

	// 检查是否为本地IP地址
	host := strings.TrimPrefix(origin, "http://")
	host = strings.TrimPrefix(host, "https://")
	host = strings.Split(host, ":")[0] // 移除端口号

	// 检查是否为本地IP地址
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() {
			return true
		}
	}

	// 获取当前服务器公网 IP
	if ip, err := common.GetPublicIP(); err == nil {
		if strings.Contains(origin, ip) {
			return true
		}
	}

	return false
}
