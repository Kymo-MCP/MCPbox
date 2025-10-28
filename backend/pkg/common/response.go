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

	// Parameter binding priority (from low to high): Body < Query < RawQuery < URI
	// Higher priority parameters will override lower priority parameters with the same name

	// 1. First bind Body parameters (lowest priority)
	switch {
	// JSON binding - for POST/PUT/PATCH requests with JSON body
	case strings.HasPrefix(contentType, "application/json") ||
		(method == "POST" || method == "PUT" || method == "PATCH") && contentType == "":
		if err := c.ShouldBindJSON(req); err != nil {
			// JSON binding failure doesn't return error directly, might not have JSON body
		}

	// Form binding - for form submissions
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		strings.HasPrefix(contentType, "multipart/form-data"):
		if err := c.ShouldBind(req); err != nil {
			// Form binding failure doesn't return error directly, might not have form data
		}
	}

	// 2. Bind Query parameters (second lowest priority)
	if err := c.ShouldBindQuery(req); err != nil {
		// Query parameter binding failure doesn't return error directly, continue with other binding methods
	}

	// 3. Handle RawQuery parameters (second highest priority)
	if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
		if err := bindRawQuery(c, rawQuery, req); err != nil {
			// RawQuery binding failure doesn't return error directly, continue with other binding methods
		}
	}

	// 4. Finally bind URI parameters (highest priority, will override all fields with the same name)
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

	// Parse raw query string
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return err
	}

	// Use reflection to set struct fields
	reqValue := reflect.ValueOf(req)
	if reqValue.Kind() != reflect.Ptr || reqValue.Elem().Kind() != reflect.Struct {
		return nil // Not a pointer to struct, skip
	}

	reqElem := reqValue.Elem()
	reqType := reqElem.Type()

	// Iterate through struct fields
	for i := 0; i < reqElem.NumField(); i++ {
		field := reqElem.Field(i)
		fieldType := reqType.Field(i)

		// Skip non-settable fields
		if !field.CanSet() {
			continue
		}

		// Get field tag name (prioritize form tag, then json tag)
		fieldName := getFieldName(fieldType)
		if fieldName == "" {
			continue
		}

		// Get value from query parameters
		if queryValues, exists := values[fieldName]; exists && len(queryValues) > 0 {
			if err := setFieldValue(field, queryValues[0]); err != nil {
				continue // Setting failed, skip this field
			}
		}
	}

	return nil
}

// getFieldName gets the field name for binding, prioritizing json tag
func getFieldName(field reflect.StructField) string {
	// Prioritize form tag
	if tag := field.Tag.Get("form"); tag != "" {
		if idx := strings.Index(tag, ","); idx != -1 {
			return tag[:idx]
		}
		return tag
	}

	// Then use json tag
	if tag := field.Tag.Get("json"); tag != "" {
		if idx := strings.Index(tag, ","); idx != -1 {
			return tag[:idx]
		}
		return tag
	}

	// Finally use lowercase field name
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
		// Handle pointer types (like *bool, *int, etc.)
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setFieldValue(field.Elem(), value)
	default:
		return nil // Unsupported type, skip
	}
	return nil
}
