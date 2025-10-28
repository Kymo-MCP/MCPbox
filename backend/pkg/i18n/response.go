package i18n

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response unified response structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// SuccessResponse success response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: GetLocalizedMessageWithGin(c, CodeSuccess),
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, code)
	}
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithCode 错误响应
func ErrorWithCode(c *gin.Context, code int) {
	ErrorResponse(c, code, "")
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, code)
	}
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// ErrorResponseWithArgs 带参数的错误响应
func ErrorResponseWithArgs(c *gin.Context, code int, args ...interface{}) {
	message := GetLocalizedMessageWithGin(c, code, args...)
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	ErrorResponse(c, CodeBadRequest, message)
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, message string) {
	ErrorResponse(c, CodeUnauthorized, message)
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, message string) {
	ErrorResponse(c, CodeForbidden, message)
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	ErrorResponse(c, CodeNotFound, message)
}

// InternalServerError 500 错误
func InternalServerError(c *gin.Context, message string) {
	ErrorResponse(c, CodeInternalError, message)
}

// ServiceUnavailable 503 错误
func ServiceUnavailable(c *gin.Context, message string) {
	ErrorResponse(c, CodeServiceUnavailable, message)
}

// GatewayTimeout 504 错误
func GatewayTimeout(c *gin.Context, message string) {
	ErrorResponse(c, CodeGatewayTimeout, message)
}

// HandleGinError 处理 Gin 错误
func HandleGinError(c *gin.Context, err error) {
	ErrorResponse(c, CodeInternalError, err.Error())
}

// HandleValidationError 处理验证错误
func HandleValidationError(c *gin.Context, message string) {
	ErrorResponse(c, CodeDataValidation, message)
}

// HandleAuthError 处理认证错误
func HandleAuthError(c *gin.Context, message string) {
	ErrorResponse(c, CodeUnauthorized, message)
}

// HandlePermissionError 处理权限错误
func HandlePermissionError(c *gin.Context, message string) {
	ErrorResponse(c, CodeInsufficientPermissions, message)
}

// HandleSignatureError 处理签名错误
func HandleSignatureError(c *gin.Context, message string) {
	ErrorResponse(c, CodeInvalidSignature, message)
}
