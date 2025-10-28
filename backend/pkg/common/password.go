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
	// Password length check
	if len(password) < PasswordMinLength {
		return fmt.Errorf("password length cannot be less than %d characters", PasswordMinLength)
	}

	if len(password) > PasswordMaxLength {
		return fmt.Errorf("password length cannot exceed %d characters", PasswordMaxLength)
	}

	// Check if contains at least one letter
	hasLetter := false
	// Check if contains at least one digit
	hasDigit := false
	// Check if contains special characters
	hasSpecial := false

	for _, char := range password {
		switch {
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
			hasLetter = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= PasswordMinASCII && char <= PasswordMaxASCII: // Special characters within printable ASCII range
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	// Check letter requirement based on configuration
	if PasswordRequireLetter && !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	// Check digit requirement based on configuration
	if PasswordRequireDigit && !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	// Check special character requirement based on configuration
	if PasswordRequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// If special characters are not required but recommended
	if !PasswordRequireSpecial && !hasSpecial {
		logger.Warn("password is recommended to contain special characters for better security")
	}

	return nil
}

// ValidatePasswordStrengthWithI18n validates password strength and returns i18n error code
func ValidatePasswordStrengthWithI18n(password string) (bool, int) {
	// Password length check
	if len(password) < PasswordMinLength || len(password) > PasswordMaxLength {
		return false, i18n.CodePasswordTooWeak
	}

	// Check if contains at least one letter
	hasLetter := false
	// Check if contains at least one digit
	hasDigit := false
	// Check if contains special characters
	hasSpecial := false

	for _, char := range password {
		switch {
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
			hasLetter = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= PasswordMinASCII && char <= PasswordMaxASCII: // Special characters within printable ASCII range
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	// Check letter requirement based on configuration
	if PasswordRequireLetter && !hasLetter {
		return false, i18n.CodePasswordTooWeak
	}

	// Check digit requirement based on configuration
	if PasswordRequireDigit && !hasDigit {
		return false, i18n.CodePasswordTooWeak
	}

	// Check special character requirement based on configuration
	if PasswordRequireSpecial && !hasSpecial {
		return false, i18n.CodePasswordTooWeak
	}

	// If special characters are not required but recommended
	if !PasswordRequireSpecial && !hasSpecial {
		logger.Warn("password is recommended to contain special characters for better security")
	}

	return true, i18n.CodeSuccess
}

// ValidatePasswordStrengthDetailed validates password strength and returns detailed result
func ValidatePasswordStrengthDetailed(password string) PasswordValidationResult {
	result := PasswordValidationResult{
		IsValid: true,
		Errors:  []string{},
	}

	// Password length check
	if len(password) < PasswordMinLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("password length cannot be less than %d characters", PasswordMinLength))
	}

	if len(password) > PasswordMaxLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("password length cannot exceed %d characters", PasswordMaxLength))
	}

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

	if PasswordRequireLetter && !hasLetter {
		result.IsValid = false
		result.Errors = append(result.Errors, "password must contain at least one letter")
	}

	if PasswordRequireDigit && !hasDigit {
		result.IsValid = false
		result.Errors = append(result.Errors, "password must contain at least one digit")
	}

	if PasswordRequireSpecial && !hasSpecial {
		result.IsValid = false
		result.Errors = append(result.Errors, "password must contain at least one special character")
	}

	return result
}

// IsPasswordStrong checks if password meets all strength requirements
func IsPasswordStrong(password string) bool {
	return ValidatePasswordStrength(password) == nil
}