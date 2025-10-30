package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"qm-mcp-server/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestResponseLoggingMiddleware 详细的请求响应日志中间件
func RequestResponseLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，以便后续处理器可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
		// 准备日志字段
		logFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// 记录请求头
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		logFields = append(logFields, zap.Any("headers", headers))

		// 记录查询参数
		if c.Request.URL.RawQuery != "" {
			logFields = append(logFields, zap.String("query", c.Request.URL.RawQuery))
		}

		// 记录表单参数
		if err := c.Request.ParseForm(); err == nil {
			if len(c.Request.Form) > 0 {
				logFields = append(logFields, zap.Any("form", c.Request.Form))
			}
		}

		// 检查Content-Type是否为JSON，并尝试解析请求体
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") && len(requestBody) > 0 {
			var jsonBody interface{}
			if err := json.Unmarshal(requestBody, &jsonBody); err == nil {
				// 将解析后的 JSON 请求体添加到日志字段中
				logFields = append(logFields, zap.Any("json", jsonBody))
				// 立即使用 logFields 记录请求日志
				logger.Info("收到请求", logFields...)
			}
		}

		// 创建自定义的 ResponseWriter 来捕获响应
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 继续处理请求
		c.Next()

		// 记录响应信息
		latency := time.Since(start)
		// 检查Content-Type是否为流式数据或下载数据
		responseContentType := c.Writer.Header().Get("Content-Type")
		if strings.Contains(responseContentType, "text/event-stream") ||
			strings.Contains(responseContentType, "application/octet-stream") ||
			strings.Contains(c.Writer.Header().Get("Content-Disposition"), "attachment") {
			// 流式数据和下载数据只记录基本信息
			logger.Info("请求完成",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
			)
		} else {
			// 其他数据记录完整响应
			logger.Info("请求完成",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
				zap.String("response_body", blw.body.String()),
			)
		}
	}
}

// bodyLogWriter 自定义的 ResponseWriter，用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
