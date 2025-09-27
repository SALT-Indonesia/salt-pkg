# Changelog

## [0.15.1] - 2025-09-27

### Fixed
- **[Issue #13]** Fixed multipart/form-data request logging not capturing form fields and file metadata
  - Added `SetWebRequest()` call after `ParseMultipartForm()` in `upload.go` to properly log parsed form data
  - Request logs now include all form fields and file metadata (`_files` array with field, filename, size, headers)
  - Affects: `UploadHandler.ServeHTTP()`

### Added
- Example implementation in `examples/httpmanager/internal/delivery/event/handler.go`
  - Demonstrates proper multipart/form-data handling with `/v1/event` endpoint
  - Shows form field extraction and file upload processing

### Changed
- Enhanced request logging for multipart/form-data requests to include:
  - All form field values in the `request` log field
  - File metadata in `_files` array (field name, filename, size, headers)

## [0.15.0] - 2025-09-27

### Added
- **Automatic Query Parameter Binding**: Added `BindQueryParams` function for automatic binding of query parameters to struct fields using tags
  - Similar to Gin's `ShouldBindQuery` functionality for familiar developer experience
  - Support for multiple data types: `string`, `int`, `int64`, `bool`, and slice types (`[]string`, `[]int`, `[]int64`, `[]bool`)
  - Reflection-based implementation with graceful error handling for invalid values
  - Uses `query` struct tags to map URL parameters to struct fields (e.g., `query:"name"`)
  - Comprehensive test coverage with 12 test cases covering all scenarios and edge cases
  - Complete documentation with usage examples, migration guide, and before/after code comparisons

### Enhanced
- **Query Parameter Functions**: Added `BindQueryParams(ctx context.Context, dst interface{}) error` to function reference table
- **Documentation**: Updated both root and module README files with comprehensive automatic binding documentation
- **Examples**: Added `NewUserSearchHandler` in examples demonstrating real-world usage of automatic query parameter binding

### Technical Details
- Automatic type conversion with validation for supported Go types
- Graceful handling of missing parameters (fields remain at zero values)
- Invalid values are skipped without causing panics or errors
- Maintains backward compatibility with existing manual `GetQueryParams()` approach
- Zero-configuration setup - just add struct tags and call `BindQueryParams()`

### Usage Example
```go
type UserSearchQuery struct {
    Name         string   `query:"name"`
    MinAge       int      `query:"min_age"`
    Active       bool     `query:"active"`
    Tags         []string `query:"tags"`
}

var params UserSearchQuery
err := httpmanager.BindQueryParams(ctx, &params)
```

### Benefits
- **Reduces Boilerplate**: Eliminates repetitive manual parameter extraction and type conversion code
- **Type Safety**: Automatic conversion to appropriate Go types with validation
- **Better Maintainability**: Query parameters clearly defined in struct tags
- **Error Resilience**: Invalid values handled gracefully without breaking request processing
- **Developer Experience**: Familiar syntax for developers coming from Gin framework

## [0.14.0] - 2025-09-27

### Added
- **Health Check Endpoint**: Added configurable health check endpoint functionality
  - Health check is enabled by default at `GET /health` endpoint
  - Returns HTTP 200 status with empty response body
  - Configurable endpoint path via `WithHealthCheckPath(path string)` option function
  - Can be disabled via `WithoutHealthCheck()` option function
  - Only accepts GET requests, returns 405 Method Not Allowed for other HTTP methods
  - Comprehensive test coverage for all configuration scenarios
  - Updated examples to demonstrate health check usage patterns

### Changed
- **Dependencies**: Upgrade logmanager module to v1.35.0

### Improved
- **Test Coverage**: Enhanced test coverage from 84.7% to 93.9% with comprehensive test additions
- Added tests for all HTTP method handlers (GET, POST, PUT, DELETE, PATCH, HandleFunc)
- Added server lifecycle tests including Start/Stop error conditions and validation
- Added ResponseError.Error() method tests for both with and without underlying errors
- Added comprehensive checkCustomErrorV2 function tests covering various error types and edge cases
- Enhanced validation of reflection-based error detection with improved edge case handling

## [0.13.0] - 2025-08-28

### Added
- **ResponseError Generic Error Handling**: Added new `ResponseError[T any]` generic error type for custom JSON error responses
- Added detailed field documentation for `Err`, `StatusCode`, and `Body` fields with usage examples and best practices
- Added reflection-based error detection in handler pipeline using `checkCustomErrorV2()` function
- Added comprehensive examples in `examples/httpmanager/internal/delivery/validation/` and `examples/httpmanager/internal/delivery/customv2/`
- Added concise README documentation focused on practical ResponseError usage
- Enhanced error handling to support both `CustomError` (fixed format) and `ResponseError` (fully customizable)

### Deprecated
- **CustomError**: Marked `CustomError` and `IsCustomError()` as deprecated in favor of `ResponseError[T]`
- Added deprecation notices with migration guidance to use `ResponseError[T]` for more flexible error handling

### Technical Details
- `ResponseError[T]` preserves original errors in `Err` field for server-side logging while allowing custom JSON response structures
- Reflection-based detection handles any `ResponseError` type at runtime without requiring compile-time type knowledge
- Automatic JSON serialization of custom error response structures using Go's `json` package
- Support for different HTTP status codes: 400 (validation), 401 (auth), 422 (business), 500 (server)
- Type-safe implementation using Go generics with full compile-time checking

### Examples
- Simple validation example with standard `{"code": "VIRB01001", "message": "...", "data": null}` format
- Complex order processing example with different error types: ValidationErrorResponse, BusinessErrorResponse, SystemErrorResponse
- Comprehensive curl testing examples demonstrating all error scenarios and status codes

### Benefits
- Complete customization of error response JSON structure
- Error preservation for logging and debugging without exposing internal details to clients
- Multiple error response formats for different scenarios (validation, business, system)
- Backward compatibility with existing `CustomError` implementation

## [0.12.0] - 2025-07-30

### Added
- **HTTP Redirect Support**: Added comprehensive HTTP redirect functionality similar to Gin's implementation
- Added `Context` type with Gin-like redirect methods: `Redirect()`, `RedirectToURL()`, `RedirectPermanent()`
- Added `RedirectHandler` for specialized redirect operations with middleware support
- Added utility functions: `Redirect()`, `RedirectToURL()`, `RedirectPermanent()` for standalone usage
- Added support for all standard HTTP redirect status codes (301, 302, 303, 307, 308)
- Added `GetContextFromStdContext()` helper function for context extraction
- Enhanced `Context` with parameter access methods: `GetQueryParams()`, `GetPathParams()`, `GetHeader()`
- Added comprehensive redirect tests covering all functionality and edge cases
- Added complete redirect documentation with examples for domain migration, conditional redirects, and form handling

### Technical Details
- Redirect functions validate status codes and panic for non-3xx codes
- `Context` type wraps standard context and provides direct access to `http.ResponseWriter` and `*http.Request`
- `RedirectHandler` integrates seamlessly with existing middleware system
- All redirect methods support path parameters, query parameters, and headers
- Utility functions provide compatibility with existing HTTP handlers

### Examples
- Simple domain migration with permanent redirects
- Conditional redirects based on User-Agent headers
- POST to GET redirects for form processing
- Dynamic redirects using path and query parameters

## [0.11.0] - 2025-07-30

### Added
- **Path Parameter Support**: Added Gin-like path parameter handling for dynamic URL routing
- Added `PathParams` type for accessing path parameters from request context
- Added `GetPathParams()` function to extract path parameters from context
- Added HTTP method shortcuts: `GET()`, `POST()`, `PUT()`, `DELETE()`, `PATCH()` for easier route registration with path parameters
- Added support for Gorilla Mux path parameter patterns with regex support
- Added comprehensive path parameter tests and examples
- Updated server implementation to use `gorilla/mux.Router` instead of `http.ServeMux` for path parameter support
- Added new `/user/{id}` and `/user/{id}/profile/{section}` example routes demonstrating path parameter usage

### Technical Details
- Replaced `http.ServeMux` with `gorilla/mux.Router` for advanced routing capabilities
- Added `extractPathParams()` function to parse path variables from requests
- Enhanced handler context to include both query parameters and path parameters
- Path parameters are automatically extracted and added to the request context alongside existing query parameters and headers

### Examples
- Added complete working examples in `examples/httpmanager/internal/delivery/user/` directory
- Examples demonstrate single and multiple path parameters with different data types
- Integration with existing query parameter and header functionality

## [0.10.0] - 2025-07-09

- Upgrade logmanager module to v1.30.0

## [0.9.0] - 2025-07-03

- Upgrade logmanager module to v1.29.0

## [0.8.0] - 2025-06-17

- Added StaticHandler for serving static files, particularly images
- Added support for automatic content type detection based on file extensions
- Added cache control headers for static files
- Added security measures to prevent directory traversal attacks
- Added support for common image formats (JPEG, PNG, GIF, SVG, WebP, etc.)
- Fix duplication logs

## [0.7.0] - 2025-06-12

- Added GetHeader method to extract a specific header value from the context
- Added GetHeaders method to retrieve all headers from the context
- Added RequestKey constant for storing HTTP request in context
- Updated ServeHTTP to store the HTTP request in the context

## [0.6.0] - 2025-05-22

- Added support for configuring server port from PORT environment variable (default: 8080)
- Added WithPort function to set server port directly

## [0.5.0] - 2025-05-21

- Changed error responses to always be in JSON format
- Added support for client-provided values in DetailedErrorResponse for code, title, and description using the new DetailedError type
- Made ErrorResponse implement the error interface, allowing it to be used as an error type

## [0.4.0] - 2025-05-03

- CORS and logging setup in httpmanager is required

## [0.3.1] - 2025-05-02

- upgrade log manager module

## [0.3.0] - 2025-04-28

- first release
