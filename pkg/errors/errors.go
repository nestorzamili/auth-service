package errors

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrCodeTokenMissing       ErrorCode = "TOKEN_MISSING"

	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"

	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"

	ErrCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
)

type AppError struct {
	Code       ErrorCode         `json:"code"`
	Message    string            `json:"message"`
	HTTPStatus int               `json:"-"`
	Details    map[string]string `json:"details,omitempty"`
	Err        error             `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithDetails(details map[string]string) *AppError {
	e.Details = details
	return e
}

func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

type ErrorResponse struct {
	Error ErrorInfo `json:"error"`
}

type ErrorInfo struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func New(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Details:    make(map[string]string),
	}
}

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func InvalidCredentials() *AppError {
	return New(ErrCodeInvalidCredentials, "Invalid username or password", http.StatusUnauthorized)
}

func TokenExpired() *AppError {
	return New(ErrCodeTokenExpired, "Token has expired", http.StatusUnauthorized)
}

func TokenInvalid() *AppError {
	return New(ErrCodeTokenInvalid, "Token is invalid", http.StatusUnauthorized)
}

func TokenMissing() *AppError {
	return New(ErrCodeTokenMissing, "Authorization token is missing", http.StatusUnauthorized)
}

func ValidationFailed(message string) *AppError {
	return New(ErrCodeValidationFailed, message, http.StatusBadRequest)
}

func InvalidInput(message string) *AppError {
	return New(ErrCodeInvalidInput, message, http.StatusBadRequest)
}

func NotFound(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func AlreadyExists(resource string) *AppError {
	return New(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource), http.StatusConflict)
}

func Internal(message string) *AppError {
	return New(ErrCodeInternal, message, http.StatusInternalServerError)
}

func ServiceUnavailable(message string) *AppError {
	return New(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

func RateLimitExceeded() *AppError {
	return New(ErrCodeRateLimitExceeded, "Rate limit exceeded, please try again later", http.StatusTooManyRequests)
}

func (e *AppError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Error: ErrorInfo{
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
		},
	}
}
