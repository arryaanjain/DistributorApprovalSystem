// Package apperrors defines typed application errors so handlers can map
// domain errors to appropriate HTTP responses without coupling services to HTTP.
package apperrors

import "fmt"

// Code represents a machine-readable error category.
type Code string

const (
	CodeNotFound           Code = "NOT_FOUND"
	CodeConflict           Code = "CONFLICT"
	CodeValidation         Code = "VALIDATION_ERROR"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodeInternal           Code = "INTERNAL_ERROR"
	CodeDuplicate          Code = "DUPLICATE"
	CodeExpired            Code = "EXPIRED"
	CodeRateLimited        Code = "RATE_LIMITED"
	CodeExternalService    Code = "EXTERNAL_SERVICE_ERROR"
	CodeCreditBlocked      Code = "CREDIT_BLOCKED"
	CodeInsufficientCredit Code = "INSUFFICIENT_CREDIT"
	CodePaymentRequired    Code = "PAYMENT_REQUIRED"
	CodeHardFlag           Code = "HARD_RISK_FLAG"
)

// AppError is a structured domain error that carries an HTTP-mappable code.
type AppError struct {
	Code    Code
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

// New creates a new AppError.
func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap wraps an existing error with an AppError.
func Wrap(code Code, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Cause: cause}
}

// Is allows errors.Is to match by code.
func Is(err error, code Code) bool {
	if ae, ok := err.(*AppError); ok {
		return ae.Code == code
	}
	return false
}

// Convenience constructors

func NotFound(msg string) *AppError          { return New(CodeNotFound, msg) }
func Conflict(msg string) *AppError          { return New(CodeConflict, msg) }
func Duplicate(msg string) *AppError         { return New(CodeDuplicate, msg) }
func Validation(msg string) *AppError        { return New(CodeValidation, msg) }
func Unauthorized(msg string) *AppError      { return New(CodeUnauthorized, msg) }
func Forbidden(msg string) *AppError         { return New(CodeForbidden, msg) }
func Internal(msg string, err error) *AppError { return Wrap(CodeInternal, msg, err) }
func Expired(msg string) *AppError           { return New(CodeExpired, msg) }
func RateLimited() *AppError                 { return New(CodeRateLimited, "too many requests") }
func CreditBlocked(msg string) *AppError     { return New(CodeCreditBlocked, msg) }
func HardFlag(msg string) *AppError          { return New(CodeHardFlag, msg) }
func InsufficientCredit(available, requested int64) *AppError {
	return New(CodeInsufficientCredit, fmt.Sprintf(
		"insufficient credit: available ₹%d, requested ₹%d", available, requested,
	))
}
