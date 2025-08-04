package httpmanager

import (
	"errors"
)

// CustomError is a custom error type that carries client-provided values for code, title, and desc
type CustomError struct {
	Err        error
	Code       string
	Title      string
	Desc       string
	StatusCode int
}

// Error implements the error interface
func (e *CustomError) Error() string {
	return e.Err.Error()
}

// IsCustomError checks if an error is a CustomError
func IsCustomError(err error) (*CustomError, bool) {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr, true
	}
	return nil, false
}
