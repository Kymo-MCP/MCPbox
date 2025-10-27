package i18n

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Format 格式化错误消息
func Format(lang SupportedLanguage, code int, args ...interface{}) string {
	return GetLocalizedMessage(code, lang, args...)
}

// FormatWithContext 使用上下文格式化错误消息
func FormatWithContext(ctx context.Context, code int, args ...interface{}) string {
	lang := GetLanguageFromContext(ctx)
	return GetLocalizedMessage(code, lang, args...)
}

// FormatWithGin 使用 Gin 上下文格式化错误消息
func FormatWithGin(c *gin.Context, code int, args ...interface{}) string {
	lang := GetLanguageFromGin(c)
	return GetLocalizedMessage(code, lang, args...)
}

// ValidateMessageTemplates 验证所有语言的消息模板是否完整
func ValidateMessageTemplates() []string {
	var missing []string

	// 获取所有错误码
	allCodes := getAllErrorCodes()
	loader := GetGlobalMessageLoader()

	// 检查每种语言是否都有对应的消息
	for _, lang := range GetSupportedLanguages() {
		langMessages := loader.GetAllMessages(lang)
		for _, code := range allCodes {
			if message, exists := langMessages[code]; !exists || message == "" {
				missing = append(missing, fmt.Sprintf("Language: %s, Code: %d", lang, code))
			}
		}
	}

	return missing
}

// getAllErrorCodes 获取所有定义的错误码
func getAllErrorCodes() []int {
	codes := []int{
		// 成功
		CodeSuccess,

		// 通用错误 (1000-1999)
		CodeBadRequest, CodeUnauthorized, CodeForbidden, CodeNotFound,
		CodeMethodNotAllowed, CodeRequestTimeout, CodeTooManyRequests,
		CodeInternalError, CodeNotImplemented, CodeServiceUnavailable, CodeGatewayTimeout,

		// 认证相关错误 (2000-2999)
		CodeInvalidToken, CodeTokenExpired, CodeMissingToken, CodeInvalidCredentials,
		CodeUserNotFound, CodePasswordIncorrect, CodeUserDisabled, CodeAccountLocked,

		// 授权相关错误 (3000-3999)
		CodeInsufficientPermissions, CodeAccessDenied, CodeRoleRequired, CodePermissionRequired,

		// 请求签名相关错误 (4000-4999)
		CodeInvalidSignature, CodeSignatureExpired, CodeMissingSignature, CodeReplayAttack,
		CodeTimestampInvalid, CodeKeyNotFound, CodeKeyExpired,

		// 业务逻辑错误 (5000-5999)
		CodeBusinessError, CodeDataValidation, CodeDuplicateEntry, CodeForeignKeyViolation,
		CodeDataConflict, CodeResourceExhausted,

		// 系统错误 (6000-6999)
		CodeDatabaseError, CodeNetworkError, CodeFileSystemError, CodeConfigurationError,
		CodeTimeoutError, CodeDependencyError,

		// 参数化错误码 (7000-7999)
		CodeParameterRequired, CodeParameterInvalid, CodeResourceNotFound, CodeResourceAlreadyExists,
		CodeOperationFailed, CodeServiceError, CodeConnectionFailed, CodeFileOperationFailed,
		CodeParseError, CodeValidationFailed,
	}

	return codes
}

// GetErrorCodeByName 根据错误名称获取错误码（用于配置文件等场景）
func GetErrorCodeByName(name string) (int, bool) {
	codeMap := map[string]int{
		"SUCCESS":                  CodeSuccess,
		"BAD_REQUEST":              CodeBadRequest,
		"UNAUTHORIZED":             CodeUnauthorized,
		"FORBIDDEN":                CodeForbidden,
		"NOT_FOUND":                CodeNotFound,
		"METHOD_NOT_ALLOWED":       CodeMethodNotAllowed,
		"REQUEST_TIMEOUT":          CodeRequestTimeout,
		"TOO_MANY_REQUESTS":        CodeTooManyRequests,
		"INTERNAL_ERROR":           CodeInternalError,
		"NOT_IMPLEMENTED":          CodeNotImplemented,
		"SERVICE_UNAVAILABLE":      CodeServiceUnavailable,
		"GATEWAY_TIMEOUT":          CodeGatewayTimeout,
		"INVALID_TOKEN":            CodeInvalidToken,
		"TOKEN_EXPIRED":            CodeTokenExpired,
		"MISSING_TOKEN":            CodeMissingToken,
		"INVALID_CREDENTIALS":      CodeInvalidCredentials,
		"USER_NOT_FOUND":           CodeUserNotFound,
		"PASSWORD_INCORRECT":       CodePasswordIncorrect,
		"USER_DISABLED":            CodeUserDisabled,
		"ACCOUNT_LOCKED":           CodeAccountLocked,
		"INSUFFICIENT_PERMISSIONS": CodeInsufficientPermissions,
		"ACCESS_DENIED":            CodeAccessDenied,
		"ROLE_REQUIRED":            CodeRoleRequired,
		"PERMISSION_REQUIRED":      CodePermissionRequired,
		"INVALID_SIGNATURE":        CodeInvalidSignature,
		"SIGNATURE_EXPIRED":        CodeSignatureExpired,
		"MISSING_SIGNATURE":        CodeMissingSignature,
		"REPLAY_ATTACK":            CodeReplayAttack,
		"TIMESTAMP_INVALID":        CodeTimestampInvalid,
		"KEY_NOT_FOUND":            CodeKeyNotFound,
		"KEY_EXPIRED":              CodeKeyExpired,
		"BUSINESS_ERROR":           CodeBusinessError,
		"DATA_VALIDATION":          CodeDataValidation,
		"DUPLICATE_ENTRY":          CodeDuplicateEntry,
		"FOREIGN_KEY_VIOLATION":    CodeForeignKeyViolation,
		"DATA_CONFLICT":            CodeDataConflict,
		"RESOURCE_EXHAUSTED":       CodeResourceExhausted,
		"DATABASE_ERROR":           CodeDatabaseError,
		"NETWORK_ERROR":            CodeNetworkError,
		"FILE_SYSTEM_ERROR":        CodeFileSystemError,
		"CONFIGURATION_ERROR":      CodeConfigurationError,
		"TIMEOUT_ERROR":            CodeTimeoutError,
		"DEPENDENCY_ERROR":         CodeDependencyError,
		"PARAMETER_REQUIRED":       CodeParameterRequired,
		"PARAMETER_INVALID":        CodeParameterInvalid,
		"RESOURCE_NOT_FOUND":       CodeResourceNotFound,
		"RESOURCE_ALREADY_EXISTS":  CodeResourceAlreadyExists,
		"OPERATION_FAILED":         CodeOperationFailed,
		"SERVICE_ERROR":            CodeServiceError,
		"CONNECTION_FAILED":        CodeConnectionFailed,
		"FILE_OPERATION_FAILED":    CodeFileOperationFailed,
		"PARSE_ERROR":              CodeParseError,
		"VALIDATION_FAILED":        CodeValidationFailed,
	}

	code, exists := codeMap[name]
	return code, exists
}

// IsClientError 判断是否为客户端错误（4xx类错误）
func IsClientError(code int) bool {
	return code >= 1000 && code < 5000
}

// IsServerError 判断是否为服务器错误（5xx类错误）
func IsServerError(code int) bool {
	return code >= 5000
}

// GetErrorCategory 获取错误类别
func GetErrorCategory(code int) string {
	switch {
	case code == 0:
		return "success"
	case code >= 1000 && code < 2000:
		return "general"
	case code >= 2000 && code < 3000:
		return "authentication"
	case code >= 3000 && code < 4000:
		return "authorization"
	case code >= 4000 && code < 5000:
		return "signature"
	case code >= 5000 && code < 6000:
		return "business"
	case code >= 6000 && code < 7000:
		return "system"
	case code >= 7000 && code < 8000:
		return "parameterized"
	default:
		return "unknown"
	}
}
