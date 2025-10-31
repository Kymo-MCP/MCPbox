package proxy

import (
	"net/http"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/logger"

	"github.com/fatedier/golib/pool"
	"go.uber.org/zap"
)

// 定义错误类型
type proxyError struct {
	message string
	status  int
}

func (e *proxyError) Error() string {
	return e.message
}

// errorHandler 处理代理请求过程中的错误
func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	// 检查是否是连接中断相关的错误
	if isProxyConnectionError(err) {
		// 连接中断是正常情况，使用 Debug 级别记录
		logger.Debug("Proxy connection interrupted",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
		)
		// 对于连接中断，不需要向客户端发送错误响应
		return
	}

	// 其他错误使用 Error 级别记录
	logger.Error("Proxy error", zap.Error(err))
	if pe, ok := err.(*proxyError); ok {
		http.Error(w, pe.message, pe.status)
	} else {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
}

type wrapPool struct{}

func newWrapPool() *wrapPool { return &wrapPool{} }

func (p *wrapPool) Get() []byte { return pool.GetBuf(32 * 1024) }

func (p *wrapPool) Put(buf []byte) { pool.PutBuf(buf) }

// isProxyConnectionError 检查代理错误是否与连接中断相关
func isProxyConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := err.Error()
	return strings.Contains(errorStr, "connection reset") ||
		strings.Contains(errorStr, "broken pipe") ||
		strings.Contains(errorStr, "connection refused") ||
		strings.Contains(errorStr, "context canceled") ||
		strings.Contains(errorStr, "context deadline exceeded") ||
		strings.Contains(errorStr, "use of closed network connection") ||
		strings.Contains(errorStr, "client disconnected") ||
		strings.Contains(errorStr, "EOF")
}

// proxyLogger 实现 io.Writer 接口，将日志转发到 zap logger
type proxyLogger struct{}

func (w *proxyLogger) Write(p []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			// 记录 panic 信息，但不抛出
			logger.Error("Recovered from panic in proxy logger",
				zap.Any("panic", r),
				zap.String("message", string(p)),
			)
		}
	}()

	// 检查是否是 context canceled 错误
	msg := string(p)
	if strings.Contains(msg, "context canceled") {
		// 对于 context canceled 错误，使用 Debug 级别记录
		logger.Debug("Client canceled connection",
			zap.String("message", msg),
		)
	} else {
		// 其他错误使用 Error 级别记录
		logger.Error("Proxy error",
			zap.String("message", msg),
		)
	}

	return len(p), nil
}

// 定义 context key
type contextKey string

const (
	IsSSEReqKey     contextKey = "isSSEReq"
	InstanceInfoKey contextKey = "instanceInfo"

	MCP_SERVER_SUBFIX_SSE = "sse"
	MCP_SERVER_SUBFIX_MCP = "mcp"
)

// PathComponents 路径组件结构体
type PathComponents struct {
	Prefix     string // 路径前缀
	InstanceID string // 实例ID
	Type       string // 传输类型
	Suffix     string //路径后缀
}
