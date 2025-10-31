package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
)

// I18nMiddleware 国际化中间件
func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中获取语言代码
		lang := GetLanguageFromRequest(c)

		// 解析为支持的语言类型
		supportedLang := parseSupportedLanguage(lang)

		// 将语言代码存储到上下文中
		i18nresp.SetLanguageToGin(c, supportedLang)

		// 设置响应头，告知客户端当前使用的语言
		c.Header("Accept-Language", string(supportedLang))

		c.Next()
	}
}

// parseSupportedLanguage 解析支持的语言
func parseSupportedLanguage(lang string) i18nresp.SupportedLanguage {
	lang = strings.ToLower(strings.TrimSpace(lang))

	switch {
	case strings.HasPrefix(lang, "zh"):
		return i18nresp.LanguageZhCN
	case strings.HasPrefix(lang, "en"):
		return i18nresp.LanguageEnUS
	default:
		return i18nresp.DefaultLanguage
	}
}

// GetLocalizer 从上下文中获取本地化器（保持向后兼容）
func GetLocalizer(c *gin.Context) interface{} {
	// 这个函数保持向后兼容，但实际上不再使用 go-i18n 的 Localizer
	return nil
}

// GetMessage 从上下文中获取本地化消息
func GetMessage(c *gin.Context, messageID string, templateData map[string]interface{}) string {
	// 这个函数主要用于向后兼容，实际使用 pkg/i18n 的消息系统
	return messageID
}

// GetLanguage 从上下文中获取当前语言
func GetLanguage(c *gin.Context) string {
	lang := i18nresp.GetLanguageFromGin(c)
	return string(lang)
}

// LocalizedError 返回本地化的错误响应
func LocalizedError(c *gin.Context, statusCode int, messageID string, templateData map[string]interface{}) {
	// 根据状态码映射到错误码
	var errorCode int
	switch statusCode {
	case http.StatusBadRequest:
		errorCode = i18nresp.CodeBadRequest
	case http.StatusUnauthorized:
		errorCode = i18nresp.CodeUnauthorized
	case http.StatusForbidden:
		errorCode = i18nresp.CodeForbidden
	case http.StatusNotFound:
		errorCode = i18nresp.CodeNotFound
	case http.StatusMethodNotAllowed:
		errorCode = i18nresp.CodeMethodNotAllowed
	case http.StatusRequestTimeout:
		errorCode = i18nresp.CodeRequestTimeout
	case http.StatusTooManyRequests:
		errorCode = i18nresp.CodeTooManyRequests
	case http.StatusInternalServerError:
		errorCode = i18nresp.CodeInternalError
	case http.StatusNotImplemented:
		errorCode = i18nresp.CodeNotImplemented
	case http.StatusServiceUnavailable:
		errorCode = i18nresp.CodeServiceUnavailable
	case http.StatusGatewayTimeout:
		errorCode = i18nresp.CodeGatewayTimeout
	default:
		errorCode = i18nresp.CodeInternalError
	}

	i18nresp.ErrorResponse(c, errorCode, "")
}

// LocalizedSuccess 返回本地化的成功响应
func LocalizedSuccess(c *gin.Context, data interface{}) {
	i18nresp.SuccessResponse(c, data)
}

// SetLanguageCookie 设置语言Cookie
func SetLanguageCookie(c *gin.Context, lang string) {
	// 设置Cookie，有效期为30天
	c.SetCookie("lang", lang, 60*60*24*30, "/", "", false, true)
}

// GetLanguageFromRequest 从请求中获取语言代码
func GetLanguageFromRequest(c *gin.Context) string {
	// 使用 pkg/i18n 包的语言检测逻辑
	lang := i18nresp.GetLanguageFromGin(c)
	return string(lang)
}
