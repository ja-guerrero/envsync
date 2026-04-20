package backend

import "fmt"

// Error codes for structured error reporting.
const (
	ECodeParseInvalidKey     = "E_PARSE_INVALID_KEY"
	ECodeParseMismatchQuotes = "E_PARSE_MISMATCH_QUOTES"
	ECodeParseDuplicateKey   = "E_PARSE_DUPLICATE_KEY"
	ECodeParseInvalidEscape  = "E_PARSE_INVALID_ESCAPE"

	ECodeSchemaConflict     = "E_SCHEMA_CONFLICT"
	ECodeSchemaUnknownType  = "E_SCHEMA_UNKNOWN_TYPE"
	ECodeSchemaInvalidField = "E_SCHEMA_INVALID_FIELD"

	ECodeConfigMissing = "E_CONFIG_MISSING"
	ECodeConfigInvalid = "E_CONFIG_INVALID"

	ECodeBackendTimeout     = "E_BACKEND_TIMEOUT"
	ECodeBackendAuth        = "E_BACKEND_AUTH"
	ECodeBackendNotFound    = "E_BACKEND_NOT_FOUND"
	ECodeBackendKeyNotFound = "E_BACKEND_KEY_NOT_FOUND"

	ECodeInternal = "E_INTERNAL"
)

// Category classifies errors for CLI output.
type Category int

const (
	CategoryUser    Category = iota // bad config, bad input
	CategoryBackend                 // vault timeout, auth failure
	CategoryBug                     // should never happen
)

// CodedError is an error with a machine-readable code and category.
type CodedError struct {
	Code     string
	Category Category
	Message  string
	Cause    error
}

func (e *CodedError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *CodedError) Unwrap() error {
	return e.Cause
}

// ErrKeyNotFound is returned when a key does not exist in the backend.
type ErrKeyNotFound struct {
	Key     string
	Backend string
}

func (e *ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key %q not found in %s backend", e.Key, e.Backend)
}
