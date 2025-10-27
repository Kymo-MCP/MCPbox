package common

import (
	"fmt"
	"qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
)

// PasswordValidationResult represents the result of password validation
type PasswordValidationResult struct {
	IsValid   bool
	ErrorCode int
	Errors    []string
}

// ValidatePasswordStrength validates password strength based on configured rules
func ValidatePasswordStrength(password string) error {
	// 密码长度检查
	if len(password) < PasswordMinLength {
		return fmt.Errorf("密码长度不能少于%d位", PasswordMinLength)
	}

	if len(password) > PasswordMaxLength {
		return fmt.Errorf("密码长度不能超过%d位", PasswordMaxLength)
	}

	// 检查是否包含至少一个字母
	hasLetter := false
	// 检查是否包含至少一个数字
	hasDigit := false
	// 检查是否包含特殊字符
	hasSpecial := false

	for _, char := range password {
		switch {
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
			hasLetter = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= PasswordMinASCII && char <= PasswordMaxASCII: // 可打印ASCII字符范围内的特殊字符
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	// 根据配置检查字母要求
	if PasswordRequireLetter && !hasLetter {
		return fmt.Errorf("密码必须包含至少一个字母")
	}

	// 根据配置检查数字要求
	if PasswordRequireDigit && !hasDigit {
		return fmt.Errorf("密码必须包含至少一个数字")
	}

	// 根据配置检查特殊字符要求
	if PasswordRequireSpecial && !hasSpecial {
		return fmt.Errorf("密码必须包含至少一个特殊字符")
	}

	// 如果不强制要求特殊字符，但推荐包含
	if !PasswordRequireSpecial && !hasSpecial {
		logger.Warn("密码建议包含特殊字符以提高安全性")
	}

	return nil
}

// ValidatePasswordStrengthWithI18n validates password strength and returns i18n error code
func ValidatePasswordStrengthWithI18n(password string) (bool, int) {
	// 密码长度检查
	if len(password) < PasswordMinLength || len(password) > PasswordMaxLength {
		return false, i18n.CodePasswordTooWeak
	}

	// 检查是否包含至少一个字母
	hasLetter := false
	// 检查是否包含至少一个数字
	hasDigit := false
	// 检查是否包含特殊字符
	hasSpecial := false

	for _, char := range password {
		switch {
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
			hasLetter = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= PasswordMinASCII && char <= PasswordMaxASCII: // 可打印ASCII字符范围内的特殊字符
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	// 根据配置检查字母要求
	if PasswordRequireLetter && !hasLetter {
		return false, i18n.CodePasswordTooWeak
	}

	// 根据配置检查数字要求
	if PasswordRequireDigit && !hasDigit {
		return false, i18n.CodePasswordTooWeak
	}

	// 根据配置检查特殊字符要求
	if PasswordRequireSpecial && !hasSpecial {
		return false, i18n.CodePasswordTooWeak
	}

	// 如果不强制要求特殊字符，但推荐包含
	if !PasswordRequireSpecial && !hasSpecial {
		logger.Warn("密码建议包含特殊字符以提高安全性")
	}

	return true, i18n.CodeSuccess
}

// ValidatePasswordStrengthDetailed returns detailed validation result
func ValidatePasswordStrengthDetailed(password string) PasswordValidationResult {
	result := PasswordValidationResult{
		IsValid:   true,
		ErrorCode: i18n.CodeSuccess,
		Errors:    make([]string, 0),
	}

	// 密码长度检查
	if len(password) < PasswordMinLength {
		result.IsValid = false
		result.ErrorCode = i18n.CodePasswordTooWeak
		result.Errors = append(result.Errors, fmt.Sprintf("密码长度不能少于%d位", PasswordMinLength))
	}

	if len(password) > PasswordMaxLength {
		result.IsValid = false
		result.ErrorCode = i18n.CodePasswordTooWeak
		result.Errors = append(result.Errors, fmt.Sprintf("密码长度不能超过%d位", PasswordMaxLength))
	}

	// 检查字符类型
	hasLetter := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
			hasLetter = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= PasswordMinASCII && char <= PasswordMaxASCII:
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	// 根据配置检查各种要求
	if PasswordRequireLetter && !hasLetter {
		result.IsValid = false
		result.ErrorCode = i18n.CodePasswordTooWeak
		result.Errors = append(result.Errors, "密码必须包含至少一个字母")
	}

	if PasswordRequireDigit && !hasDigit {
		result.IsValid = false
		result.ErrorCode = i18n.CodePasswordTooWeak
		result.Errors = append(result.Errors, "密码必须包含至少一个数字")
	}

	if PasswordRequireSpecial && !hasSpecial {
		result.IsValid = false
		result.ErrorCode = i18n.CodePasswordTooWeak
		result.Errors = append(result.Errors, "密码必须包含至少一个特殊字符")
	}

	return result
}

// IsPasswordStrong checks if password meets all strength requirements
func IsPasswordStrong(password string) bool {
	return ValidatePasswordStrength(password) == nil
}