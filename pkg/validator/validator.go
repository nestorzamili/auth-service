package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate      *validator.Validate
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
)

func init() {
	validate = validator.New()

	_ = validate.RegisterValidation("username", validateUsername)
	_ = validate.RegisterValidation("password", validatePassword)
}

func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return formatValidationError(err)
	}
	return nil
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	return usernameRegex.MatchString(username)
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func formatValidationError(err error) error {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, e := range validationErrs {
			messages = append(messages, formatFieldError(e))
		}
		return fmt.Errorf("%s", strings.Join(messages, "; "))
	}
	return err
}

func formatFieldError(e validator.FieldError) string {
	field := strings.ToLower(e.Field())

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "username":
		return fmt.Sprintf("%s must be 3-30 characters and contain only letters, numbers, underscores, or hyphens", field)
	case "password":
		return fmt.Sprintf("%s must be at least 8 characters and contain uppercase, lowercase, number, and special character", field)
	case "eqfield":
		return fmt.Sprintf("%s must match %s", field, e.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidateUsername(username string) bool {
	return usernameRegex.MatchString(username)
}
