package domain

import "fmt"

// ErrorCode is a machine-readable error identifier sent to clients.
type ErrorCode string

// ErrorCode constants.
const (
	ErrCodeNotFound            ErrorCode = "NOT_FOUND"
	ErrCodeConflict            ErrorCode = "CONFLICT"
	ErrCodeValidation          ErrorCode = "VALIDATION_ERROR"
	ErrCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrCodeRateLimited         ErrorCode = "RATE_LIMITED"
	ErrCodeInsufficientCredits ErrorCode = "INSUFFICIENT_CREDITS"
	ErrCodePaymentFailed       ErrorCode = "PAYMENT_FAILED"
	ErrCodeAIUnavailable       ErrorCode = "AI_UNAVAILABLE"
	ErrCodeInternal            ErrorCode = "INTERNAL_ERROR"
)

// AppError is the standard application error type.
// Handlers map these to HTTP status codes.
type AppError struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	Err     error          `json:"-"` // wrapped original error (never sent to client)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// --- Constructors for common errors ---

// NotFound creates a NOT_FOUND error.
func NotFound(resource string, id any) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: map[string]any{"id": id},
	}
}

// Conflict creates a CONFLICT error (e.g., duplicate email).
func Conflict(resource string, detail string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: fmt.Sprintf("%s already exists", resource),
		Details: map[string]any{"detail": detail},
	}
}

// ValidationError creates a VALIDATION_ERROR.
func ValidationError(message string, details map[string]any) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: details,
	}
}

// Unauthorized creates an UNAUTHORIZED error.
func Unauthorized(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

// Forbidden creates a FORBIDDEN error.
func Forbidden(message string) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: message,
	}
}

// RateLimited creates a RATE_LIMITED error.
func RateLimited() *AppError {
	return &AppError{
		Code:    ErrCodeRateLimited,
		Message: "Too many requests. Please try again later.",
	}
}

// InsufficientCredits creates an INSUFFICIENT_CREDITS error.
func InsufficientCredits(creditType CreditType) *AppError {
	return &AppError{
		Code:    ErrCodeInsufficientCredits,
		Message: fmt.Sprintf("Insufficient %s credits", creditType),
		Details: map[string]any{"credit_type": creditType},
	}
}

// InternalError wraps an unexpected error for logging while returning
// a safe message to the client.
func InternalError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: "An unexpected error occurred",
		Err:     err,
	}
}
