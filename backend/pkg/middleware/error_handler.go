package middleware

import (
	"qm-mcp-server/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 只处理有错误的请求
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// 根据错误类型返回相应的错误码
			switch err.Type {
			case gin.ErrorTypeBind:
				// 请求参数绑定错误
				i18n.BadRequest(c, "请求参数格式错误")
			case gin.ErrorTypeRender:
				// 渲染错误
				i18n.InternalServerError(c, "服务器渲染错误")
			case gin.ErrorTypePrivate:
				// 私有错误
				i18n.InternalServerError(c, "服务器内部错误")
			case gin.ErrorTypePublic:
				// 公共错误
				i18n.BadRequest(c, err.Error())
			default:
				// 默认内部服务器错误
				i18n.InternalServerError(c, err.Error())
			}
		}
	}
}

// NotFoundHandler 404 处理器
func NotFoundHandler(c *gin.Context) {
	i18n.NotFound(c, "请求的资源不存在")
}

// MethodNotAllowedHandler 405 处理器
func MethodNotAllowedHandler(c *gin.Context) {
	i18n.ErrorResponse(c, i18n.CodeMethodNotAllowed, "请求方法不允许")
}

// PanicRecovery 恐慌恢复中间件
func PanicRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		i18n.InternalServerError(c, "服务器内部错误")
	})
}
