package common

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"

	i18nresp "qm-mcp-server/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// GinSuccess returns a successful response with data
func GinSuccess(c *gin.Context, data interface{}) {
	i18nresp.SuccessResponse(c, data)
}

// GinError returns an error response with error code and message
func GinError(c *gin.Context, code int, message string) {
	i18nresp.ErrorResponse(c, code, message)
}

// BindAndValidateUniversal binds request data and performs validation
func BindAndValidateUniversal(c *gin.Context, req interface{}) error {
	contentType := c.GetHeader("Content-Type")
	method := c.Request.Method

	// 参数绑定优先级（从低到高）：Body < Query < RawQuery < URI
	// 优先级高的参数会覆盖优先级低的同名参数

	// 1. 首先绑定Body参数（优先级最低）
	switch {
	// JSON绑定 - 适用于POST/PUT/PATCH等带JSON body的请求
	case strings.HasPrefix(contentType, "application/json") ||
		(method == "POST" || method == "PUT" || method == "PATCH") && contentType == "":
		if err := c.ShouldBindJSON(req); err != nil {
			// JSON绑定失败不直接返回错误，可能没有JSON body
		}

	// Form绑定 - 适用于表单提交
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		strings.HasPrefix(contentType, "multipart/form-data"):
		if err := c.ShouldBind(req); err != nil {
			// Form绑定失败不直接返回错误，可能没有Form数据
		}
	}

	// 2. 绑定Query参数（次低优先级）
	if err := c.ShouldBindQuery(req); err != nil {
		// Query参数绑定失败不直接返回错误，继续尝试其他绑定方式
	}

	// 3. 处理RawQuery参数（次高优先级）
	if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
		if err := bindRawQuery(c, rawQuery, req); err != nil {
			// RawQuery绑定失败不直接返回错误，继续尝试其他绑定方式
		}
	}

	// 4. 最后绑定URI参数（优先级最高，会覆盖所有同名字段）
	if len(c.Params) > 0 {
		if err := c.ShouldBindUri(req); err != nil {
			GinError(c, i18nresp.CodeInternalError, err.Error())
			return err
		}
	}
	return nil
}

// BindAndValidate binds JSON request data and performs validation
func BindAndValidate(c *gin.Context, req interface{}) error {
	return BindAndValidateUniversal(c, req)
}

// BindAndValidateQuery binds query parameters and performs validation
func BindAndValidateQuery(c *gin.Context, req interface{}) error {
	return BindAndValidateUniversal(c, req)
}

// bindRawQuery binds raw query parameters to struct fields
func bindRawQuery(c *gin.Context, rawQuery string, req interface{}) error {
	if rawQuery == "" {
		return nil
	}

	// 解析原始查询字符串
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return err
	}

	// 使用反射设置结构体字段
	reqValue := reflect.ValueOf(req)
	if reqValue.Kind() != reflect.Ptr || reqValue.Elem().Kind() != reflect.Struct {
		return nil // 不是指向结构体的指针，跳过
	}

	reqElem := reqValue.Elem()
	reqType := reqElem.Type()

	// 遍历结构体字段
	for i := 0; i < reqElem.NumField(); i++ {
		field := reqElem.Field(i)
		fieldType := reqType.Field(i)

		// 跳过不可设置的字段
		if !field.CanSet() {
			continue
		}

		// 获取字段的标签名（优先使用form标签，然后json标签）
		fieldName := getFieldName(fieldType)
		if fieldName == "" {
			continue
		}

		// 从查询参数中获取值
		if queryValues, exists := values[fieldName]; exists && len(queryValues) > 0 {
			if err := setFieldValue(field, queryValues[0]); err != nil {
				continue // 设置失败，跳过该字段
			}
		}
	}

	return nil
}

// getFieldName gets the field name for binding, prioritizing json tag
func getFieldName(field reflect.StructField) string {
	// 优先使用form标签
	if tag := field.Tag.Get("form"); tag != "" {
		if idx := strings.Index(tag, ","); idx != -1 {
			return tag[:idx]
		}
		return tag
	}

	// 其次使用json标签
	if tag := field.Tag.Get("json"); tag != "" {
		if idx := strings.Index(tag, ","); idx != -1 {
			return tag[:idx]
		}
		return tag
	}

	// 最后使用字段名的小写形式
	return strings.ToLower(field.Name)
}

// setFieldValue sets the field value based on its type
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(intVal)
		} else {
			return err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(uintVal)
		} else {
			return err
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(floatVal)
		} else {
			return err
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			field.SetBool(boolVal)
		} else {
			return err
		}
	case reflect.Ptr:
		// 处理指针类型（如 *bool, *int 等）
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setFieldValue(field.Elem(), value)
	default:
		return nil // 不支持的类型，跳过
	}
	return nil
}
