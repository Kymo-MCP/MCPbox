package i18n

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// SupportedLanguage supported language type
type SupportedLanguage string

const (
	LanguageZhCN SupportedLanguage = "zh-CN" // Simplified Chinese
	LanguageEnUS SupportedLanguage = "en-US" // English
)

// DefaultLanguage default language
const DefaultLanguage = LanguageZhCN

// ContextLanguageKey 上下文中语言的键
const ContextLanguageKey = "language"

// GetLanguageFromContext 从上下文获取语言
func GetLanguageFromContext(ctx context.Context) SupportedLanguage {
	if lang, ok := ctx.Value(ContextLanguageKey).(SupportedLanguage); ok {
		return lang
	}
	return DefaultLanguage
}

// GetLanguageFromGin 从 Gin 上下文获取语言
func GetLanguageFromGin(c *gin.Context) SupportedLanguage {
	// 1. 优先从查询参数获取
	if lang := c.Query("lang"); lang != "" {
		if supportedLang := parseSupportedLanguage(lang); supportedLang != "" {
			return supportedLang
		}
	}

	// 2. 从请求头获取
	if lang := c.GetHeader("Accept-Language"); lang != "" {
		if supportedLang := parseAcceptLanguage(lang); supportedLang != "" {
			return supportedLang
		}
	}

	// 3. 从上下文获取（可能由中间件设置）
	if lang, exists := c.Get(ContextLanguageKey); exists {
		if supportedLang, ok := lang.(SupportedLanguage); ok {
			return supportedLang
		}
	}

	return DefaultLanguage
}

// parseSupportedLanguage 解析支持的语言
func parseSupportedLanguage(lang string) SupportedLanguage {
	lang = strings.ToLower(strings.TrimSpace(lang))

	switch {
	case strings.HasPrefix(lang, "zh"):
		return LanguageZhCN
	case strings.HasPrefix(lang, "en"):
		return LanguageEnUS
	default:
		return ""
	}
}

// parseAcceptLanguage 解析 Accept-Language 头
func parseAcceptLanguage(acceptLang string) SupportedLanguage {
	// 简单解析 Accept-Language，取第一个支持的语言
	languages := strings.Split(acceptLang, ",")
	for _, lang := range languages {
		// 移除权重信息 (如 zh-CN;q=0.9)
		lang = strings.Split(strings.TrimSpace(lang), ";")[0]
		if supportedLang := parseSupportedLanguage(lang); supportedLang != "" {
			return supportedLang
		}
	}
	return ""
}

// SetLanguageToContext 设置语言到上下文
func SetLanguageToContext(ctx context.Context, lang SupportedLanguage) context.Context {
	return context.WithValue(ctx, ContextLanguageKey, lang)
}

// SetLanguageToGin 设置语言到 Gin 上下文
func SetLanguageToGin(c *gin.Context, lang SupportedLanguage) {
	c.Set(ContextLanguageKey, lang)
}

// GetLocalizedMessage 获取本地化消息
func GetLocalizedMessage(code int, lang SupportedLanguage, args ...interface{}) string {
	// 获取消息模板
	template := getMessageTemplate(lang, code)

	// 如果有参数，进行格式化
	if len(args) > 0 {
		return fmt.Sprintf(template, args...)
	}

	return template
}

// GetLocalizedMessageWithContext 使用上下文获取本地化消息
func GetLocalizedMessageWithContext(ctx context.Context, code int, args ...interface{}) string {
	lang := GetLanguageFromContext(ctx)
	return GetLocalizedMessage(code, lang, args...)
}

// GetLocalizedMessageWithGin 使用 Gin 上下文获取本地化消息
func GetLocalizedMessageWithGin(c *gin.Context, code int, args ...interface{}) string {
	lang := GetLanguageFromGin(c)
	return GetLocalizedMessage(code, lang, args...)
}
