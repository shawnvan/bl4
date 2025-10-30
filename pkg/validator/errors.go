package validator

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// ErrorCode represents different types of validation and processing errors
type ErrorCode int

const (
	// General errors
	ErrCodeUnknown ErrorCode = iota
	ErrCodeInternal
	ErrCodeNotImplemented

	// Input validation errors
	ErrCodeInvalidInput
	ErrCodeEmptyInput
	ErrCodeInvalidFormat
	ErrCodeInvalidLength
	ErrCodeInvalidCharacter

	// Base85 encoding/decoding errors
	ErrCodeInvalidBase85
	ErrCodeInvalidPrefix
	ErrCodeBase85DecodeFailed
	ErrCodeBase85EncodeFailed

	// Bitstream errors
	ErrCodeBitStreamEOF
	ErrCodeInvalidBitCount
	ErrCodeBitStreamCorrupted
	ErrCodeUnexpectedToken

	// Token parsing errors
	ErrCodeInvalidToken
	ErrCodeInvalidTokenType
	ErrCodeTokenValueInvalid
	ErrCodeTokenStreamInvalid

	// Item serialization/deserialization errors
	ErrCodeInvalidItemData
	ErrCodeInvalidItemLevel
	ErrCodeInvalidItemType
	ErrCodeInvalidManufacturer
	ErrCodeInvalidPartsData

	// HTTP/API errors
	ErrCodeBadRequest
	ErrCodeUnauthorized
	ErrCodeForbidden
	ErrCodeNotFound
	ErrCodeMethodNotAllowed
	ErrCodeRequestTimeout
	ErrCodeTooManyRequests
	ErrCodeInternalServerError
	ErrCodeServiceUnavailable

	// Performance/Resource errors
	ErrCodeProcessingTimeout
	ErrCodeMemoryLimit
	ErrCodeBatchSizeExceeded
	ErrCodeRateLimited

	// Additional validation errors
	ErrCodeValueTooLarge
	ErrCodeInvalidString
)

// String returns the string representation of an error code
func (e ErrorCode) String() string {
	switch e {
	case ErrCodeUnknown:
		return "UNKNOWN"
	case ErrCodeInternal:
		return "INTERNAL"
	case ErrCodeNotImplemented:
		return "NOT_IMPLEMENTED"
	case ErrCodeInvalidInput:
		return "INVALID_INPUT"
	case ErrCodeEmptyInput:
		return "EMPTY_INPUT"
	case ErrCodeInvalidFormat:
		return "INVALID_FORMAT"
	case ErrCodeInvalidLength:
		return "INVALID_LENGTH"
	case ErrCodeInvalidCharacter:
		return "INVALID_CHARACTER"
	case ErrCodeInvalidBase85:
		return "INVALID_BASE85"
	case ErrCodeInvalidPrefix:
		return "INVALID_PREFIX"
	case ErrCodeBase85DecodeFailed:
		return "BASE85_DECODE_FAILED"
	case ErrCodeBase85EncodeFailed:
		return "BASE85_ENCODE_FAILED"
	case ErrCodeBitStreamEOF:
		return "BITSTREAM_EOF"
	case ErrCodeInvalidBitCount:
		return "INVALID_BIT_COUNT"
	case ErrCodeBitStreamCorrupted:
		return "BITSTREAM_CORRUPTED"
	case ErrCodeUnexpectedToken:
		return "UNEXPECTED_TOKEN"
	case ErrCodeInvalidToken:
		return "INVALID_TOKEN"
	case ErrCodeInvalidTokenType:
		return "INVALID_TOKEN_TYPE"
	case ErrCodeTokenValueInvalid:
		return "TOKEN_VALUE_INVALID"
	case ErrCodeTokenStreamInvalid:
		return "TOKEN_STREAM_INVALID"
	case ErrCodeInvalidItemData:
		return "INVALID_ITEM_DATA"
	case ErrCodeInvalidItemLevel:
		return "INVALID_ITEM_LEVEL"
	case ErrCodeInvalidItemType:
		return "INVALID_ITEM_TYPE"
	case ErrCodeInvalidManufacturer:
		return "INVALID_MANUFACTURER"
	case ErrCodeInvalidPartsData:
		return "INVALID_PARTS_DATA"
	case ErrCodeBadRequest:
		return "BAD_REQUEST"
	case ErrCodeUnauthorized:
		return "UNAUTHORIZED"
	case ErrCodeForbidden:
		return "FORBIDDEN"
	case ErrCodeNotFound:
		return "NOT_FOUND"
	case ErrCodeMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case ErrCodeRequestTimeout:
		return "REQUEST_TIMEOUT"
	case ErrCodeTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case ErrCodeInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	case ErrCodeServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case ErrCodeProcessingTimeout:
		return "PROCESSING_TIMEOUT"
	case ErrCodeMemoryLimit:
		return "MEMORY_LIMIT"
	case ErrCodeBatchSizeExceeded:
		return "BATCH_SIZE_EXCEEDED"
	case ErrCodeRateLimited:
		return "RATE_LIMITED"
	case ErrCodeValueTooLarge:
		return "VALUE_TOO_LARGE"
	case ErrCodeInvalidString:
		return "INVALID_STRING"
	default:
		return "UNKNOWN"
	}
}

// HTTPStatus returns the appropriate HTTP status code for this error type
func (e ErrorCode) HTTPStatus() int {
	switch e {
	case ErrCodeInvalidInput, ErrCodeEmptyInput, ErrCodeInvalidFormat,
		ErrCodeInvalidLength, ErrCodeInvalidCharacter, ErrCodeInvalidBase85,
		ErrCodeInvalidPrefix, ErrCodeBase85DecodeFailed, ErrCodeBase85EncodeFailed,
		ErrCodeInvalidBitCount, ErrCodeBitStreamCorrupted, ErrCodeUnexpectedToken,
		ErrCodeInvalidToken, ErrCodeInvalidTokenType, ErrCodeTokenValueInvalid,
		ErrCodeTokenStreamInvalid, ErrCodeInvalidItemData, ErrCodeInvalidItemLevel,
		ErrCodeInvalidItemType, ErrCodeInvalidManufacturer, ErrCodeInvalidPartsData:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case ErrCodeRequestTimeout, ErrCodeProcessingTimeout:
		return http.StatusRequestTimeout
	case ErrCodeTooManyRequests, ErrCodeRateLimited:
		return http.StatusTooManyRequests
	case ErrCodeInternal, ErrCodeInternalServerError:
		return http.StatusInternalServerError
	case ErrCodeNotImplemented:
		return http.StatusNotImplemented
	case ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeMemoryLimit, ErrCodeBatchSizeExceeded:
		return http.StatusRequestTimeout
	case ErrCodeBitStreamEOF:
		return http.StatusBadRequest
	case ErrCodeValueTooLarge, ErrCodeInvalidString:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// ValidationError represents a validation error with detailed information
type ValidationError struct {
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Field     string                 `json:"field,omitempty"`
	Value     interface{}            `json:"value,omitempty"`
	Operation string                 `json:"operation,omitempty"`
	Cause     error                  `json:"cause,omitempty"`
	Stack     []string               `json:"stack,omitempty"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s, value: %v)", e.Code, e.Message, e.Field, e.Value)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a new validation error
func NewValidationError(code ErrorCode, message string) *ValidationError {
	return &ValidationError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
		Stack:   captureStack(),
	}
}

// WithField adds field information to the error
func (e *ValidationError) WithField(field string, value interface{}) *ValidationError {
	e.Field = field
	e.Value = value
	return e
}

// WithOperation adds operation information to the error
func (e *ValidationError) WithOperation(operation string) *ValidationError {
	e.Operation = operation
	return e
}

// WithCause adds an underlying cause to the error
func (e *ValidationError) WithCause(cause error) *ValidationError {
	e.Cause = cause
	return e
}

// WithDetail adds a detail to the error
func (e *ValidationError) WithDetail(key string, value interface{}) *ValidationError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Is checks if this error matches the target error type
func (e *ValidationError) Is(target error) bool {
	if t, ok := target.(*ValidationError); ok {
		return e.Code == t.Code
	}
	return false
}

// captureStack captures the current call stack for debugging
func captureStack() []string {
	var stack []string
	for i := 2; i < 10; i++ { // Skip captureStack and NewValidationError
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// Format: function (file:line)
		stack = append(stack, fmt.Sprintf("%s (%s:%d)", fn.Name(), file, line))
	}
	return stack
}

// Pre-defined error constructors
func NewInvalidInputError(message string) *ValidationError {
	return NewValidationError(ErrCodeInvalidInput, message)
}

func NewEmptyInputError(message string) *ValidationError {
	return NewValidationError(ErrCodeEmptyInput, message)
}

func NewInvalidFormatError(message string) *ValidationError {
	return NewValidationError(ErrCodeInvalidFormat, message)
}

func NewBase85Error(message string) *ValidationError {
	return NewValidationError(ErrCodeInvalidBase85, message)
}

func NewBitStreamError(message string) *ValidationError {
	return NewValidationError(ErrCodeBitStreamCorrupted, message)
}

func NewTokenError(message string) *ValidationError {
	return NewValidationError(ErrCodeInvalidToken, message)
}

func NewItemDataError(message string) *ValidationError {
	return NewValidationError(ErrCodeInvalidItemData, message)
}

func NewInternalError(message string) *ValidationError {
	return NewValidationError(ErrCodeInternal, message)
}

func NewNotImplementedError(message string) *ValidationError {
	return NewValidationError(ErrCodeNotImplemented, message)
}

// ValidationRule represents a validation rule function
type ValidationRule func(interface{}) error

// Validator manages validation rules and performs validation
type Validator struct {
	rules map[string][]ValidationRule
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]ValidationRule),
	}
}

// AddRule adds a validation rule for a field
func (v *Validator) AddRule(field string, rule ValidationRule) {
	if v.rules[field] == nil {
		v.rules[field] = make([]ValidationRule, 0)
	}
	v.rules[field] = append(v.rules[field], rule)
}

// Validate validates a data object against all registered rules
func (v *Validator) Validate(data interface{}) error {
	// Convert data to map for field access (simplified)
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return NewValidationError(ErrCodeInvalidInput, "data must be a map")
	}

	var errors []string
	for field, rules := range v.rules {
		value := dataMap[field]
		for _, rule := range rules {
			if err := rule(value); err != nil {
				if ve, ok := err.(*ValidationError); ok {
					errors = append(errors, fmt.Sprintf("field '%s': %s", field, ve.Message))
				} else {
					errors = append(errors, fmt.Sprintf("field '%s': %s", field, err.Error()))
				}
			}
		}
	}

	if len(errors) > 0 {
		return NewValidationError(ErrCodeInvalidInput, "validation failed").
			WithDetail("errors", strings.Join(errors, "; "))
	}

	return nil
}

// Common validation rules
func RequiredRule() ValidationRule {
	return func(value interface{}) error {
		if value == nil || value == "" {
			return NewValidationError(ErrCodeEmptyInput, "field is required")
		}
		return nil
	}
}

func MinLengthRule(minLength int) ValidationRule {
	return func(value interface{}) error {
		if value == nil {
			return nil // Use RequiredRule for required fields
		}

		str, ok := value.(string)
		if !ok {
			return NewValidationError(ErrCodeInvalidFormat, "field must be a string")
		}

		if len(str) < minLength {
			return NewValidationError(ErrCodeInvalidLength, fmt.Sprintf("field must be at least %d characters", minLength))
		}

		return nil
	}
}

func MaxLengthRule(maxLength int) ValidationRule {
	return func(value interface{}) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return NewValidationError(ErrCodeInvalidFormat, "field must be a string")
		}

		if len(str) > maxLength {
			return NewValidationError(ErrCodeInvalidLength, fmt.Sprintf("field must be at most %d characters", maxLength))
		}

		return nil
	}
}

func PatternRule(pattern string) ValidationRule {
	return func(value interface{}) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return NewValidationError(ErrCodeInvalidFormat, "field must be a string")
		}

		// This is a simplified pattern check - in real implementation, use regexp
		if !strings.Contains(str, pattern) {
			return NewValidationError(ErrCodeInvalidFormat, fmt.Sprintf("field must contain pattern: %s", pattern))
		}

		return nil
	}
}