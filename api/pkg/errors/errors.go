package errors

import (
	"fmt"
	"net/http"
)

// APIError represents a structured API error
type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common error codes
const (
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeInternalServer = "INTERNAL_SERVER_ERROR"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// HTTP status mapping
var ErrorStatusMap = map[string]int{
	ErrCodeValidation:         http.StatusBadRequest,
	ErrCodeNotFound:           http.StatusNotFound,
	ErrCodeUnauthorized:       http.StatusUnauthorized,
	ErrCodeForbidden:          http.StatusForbidden,
	ErrCodeConflict:           http.StatusConflict,
	ErrCodeInternalServer:     http.StatusInternalServerError,
	ErrCodeBadRequest:         http.StatusBadRequest,
	ErrCodeServiceUnavailable: http.StatusServiceUnavailable,
}

// NewAPIError creates a new API error
func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *APIError {
	return &APIError{
		Code:    ErrCodeValidation,
		Message: message,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *APIError {
	return &APIError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *APIError {
	return &APIError{
		Code:    ErrCodeInternalServer,
		Message: message,
	}
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if apiErr, ok := err.(*APIError); ok {
		if status, exists := ErrorStatusMap[apiErr.Code]; exists {
			return status
		}
	}
	return http.StatusInternalServerError
} 