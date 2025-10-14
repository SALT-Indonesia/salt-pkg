package httpmanager

import (
	"errors"
)

// CustomError is a custom error type that carries client-provided values for code, title, and desc
// Deprecated: Use ResponseError[T] instead for more flexible and type-safe error handling.
// ResponseError allows complete customization of error response structure while preserving
// original errors for server-side logging. See examples in internal/delivery/validation/.
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
// Deprecated: Use ResponseError[T] instead for more flexible error handling.
func IsCustomError(err error) (*CustomError, bool) {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr, true
	}
	return nil, false
}

// ResponseError is a generic error type that allows clients to define their own response structure
type ResponseError[T any] struct {
	// Err preserves the original Go error for server-side logging, debugging, and error tracking.
	// This field is not included in the JSON response sent to clients, but is available for
	// server-side monitoring, logging middleware, and error tracking systems.
	// Use error wrapping (fmt.Errorf with %w) to maintain error chains for debugging.
	Err error

	// StatusCode specifies the HTTP status code to return to the client (e.g., 400, 401, 422, 500).
	// This determines how clients and intermediate systems (load balancers, proxies, browsers)
	// will handle the response. Common values:
	//   - 400 (Bad Request): Client validation errors, malformed requests
	//   - 401 (Unauthorized): Authentication required or failed
	//   - 403 (Forbidden): Client lacks permission for this resource
	//   - 422 (Unprocessable Entity): Business logic validation failures
	//   - 500 (Internal Server Error): Server-side errors, database issues
	StatusCode int

	// Body is the custom response structure that will be serialized to JSON and sent to the client.
	// This is a generic type T that can be any struct you define, allowing complete customization
	// of your error response format. The httpmanager will automatically serialize this to JSON
	// using Go's json package with proper Content-Type headers. Examples:
	//   - Simple: struct{Code string; Message string; Data interface{}}
	//   - Rich: struct with timestamps, request IDs, field details, suggestions, metadata
	Body T
}

// Error implements the error interface
func (e *ResponseError[T]) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "response error with custom body"
}

// ResponseSuccess is a generic type that allows clients to return custom HTTP status codes
// for successful responses (e.g., 201 Created, 202 Accepted, 204 No Content).
// This provides flexibility when the default 200 OK status code doesn't accurately represent
// the operation's outcome.
type ResponseSuccess[T any] struct {
	// StatusCode specifies the HTTP status code to return to the client for successful responses.
	// Common success status codes:
	//   - 200 (OK): Standard successful response (default if ResponseSuccess is not used)
	//   - 201 (Created): Resource successfully created
	//   - 202 (Accepted): Request accepted for processing but not yet completed
	//   - 204 (No Content): Successful request with no response body
	//   - 206 (Partial Content): Partial resource returned (range requests)
	StatusCode int

	// Body is the response structure that will be serialized to JSON and sent to the client.
	// For 204 No Content responses, this should be nil or an empty struct.
	// The httpmanager will automatically serialize this to JSON with proper Content-Type headers.
	Body T
}
