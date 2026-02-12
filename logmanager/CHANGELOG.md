# Changelog

## [Unreleased]

## [1.41.0] - 2026-02-12
- **Fix nil pointer dereference when using transactions in async goroutines (#54)**
  - Add nil checks for `Attributes` and `Attributes.value` in all attribute functions
  - Prevents panic when transactions are accessed after `End()` is called
  - Fixes race condition in high-traffic scenarios with concurrent goroutine access
  - Add double nil checks (`nil == a || nil == a.value`) to handle race window between check and access
  - Updated functions: `ResponseBodyAttribute`, `RequestBodyConsumerAttributes`, `RequestAgentAttributes`, `headerAttributes`, `RequestBodyAttributes`, `extractMultipartFormData`, `CaptureMultipartFormDataIfParsed`, `parseFormData`, `RequestBodyAttribute`, `ResponseBodyAttributes`, `ResponseCodeAttribute`
  - Add `gin-async-test` example demonstrating async goroutine usage with Gin framework
  - Tested with 20+ concurrent requests without panics
  - Root cause: middleware calls `tx.End()` → `reset()` → `attrs = nil` while goroutines still hold transaction reference

## [1.40.0] - 2026-02-11
- **Add split-level log output routing for containerized environments (#52)**
  - Add `WithSplitLevelOutput()` option following Twelve-Factor App principles
  - DEBUG, INFO, TRACE levels are written to `os.Stdout`
  - WARN, ERROR, FATAL, PANIC levels are written to `os.Stderr`
  - Implemented via a custom logrus hook (`splitLevelOutputHook`) that formats and routes entries by level
  - Option is ignored when `WithLogDir()` is set (file-based logging takes precedence)
  - Resolves false positives in container log collectors (Docker, Kubernetes) that treat stderr as error state
  - Add 10 unit tests covering all log levels, hook behavior, and option configuration
- **Fix multipart/form-data request logging in middleware (#13)**
  - Add `CaptureMultipartFormData(r *http.Request)` method to `TxnRecord` for post-handler form data capture
  - Add `CaptureMultipartFormDataIfParsed()` internal function that safely extracts parsed form data
  - Update `lmecho` middleware to capture multipart form data after handler processes the request
  - Update `lmgorilla` middleware to capture multipart form data after handler processes the request
  - Update `lmgin` middleware to capture multipart form data after handler processes the request
  - Form fields and file metadata (field name, filename, size, headers) are now logged for multipart requests
  - Only captures data when form has been parsed by the handler and request body wasn't already recorded
  - Add 6 unit tests covering capture scenarios, no-op conditions, and file upload metadata
  - Add `echo-multipart` example demonstrating form data and file upload logging with Echo framework

## [1.39.0] - 2026-02-11
- **Add HTTP Flusher, Hijacker, and Pusher support for middleware (#49)**
  - Add `Flush()` method to `replacementResponseWriter` for Server-Sent Events (SSE) and chunked transfer encoding
  - Add `Hijack()` method to `replacementResponseWriter` for WebSocket upgrades
  - Add `Push()` method to `replacementResponseWriter` for HTTP/2 server push
  - Add same optional interface methods to `lmecho`'s `customWriter` for consistency
  - All methods gracefully delegate to underlying writer when supported
  - Headers are automatically written before flushing or hijacking
  - Add comprehensive unit tests (12 test cases) for all optional interfaces
  - Add integration tests for streaming support in `lmgorilla` middleware
  - Add `gorilla-streaming` example demonstrating SSE, chunked transfer, and NDJSON streaming
  - Fix resolves middleware breaking streaming responses by properly implementing optional HTTP interfaces
  - 100% backward compatible with existing code

## [1.38.2] - 2025-12-09
- **Add Debug() method to Application (#40)**
  - New public method `Debug() bool` returns whether debug mode is enabled for the application
  - Returns `false` if called on a nil Application pointer
  - Required by httpmanager for conditional 404 debug logging

## [1.38.1] - 2025-12-01
- **Fix StructMaskWithConfig to apply JSONPath masking for non-struct types**
  - Previously, masking would skip JSONPath/FieldPattern configs when input was a map
  - Now correctly applies all masking configurations regardless of input type
  - Fixes EmailMask and other mask types not working with `map[string]interface{}` inputs

## [1.38.0] - 2025-12-01
- **Add EmailMask type for proper email address masking (#34)**
  - New `EmailMask` mask type that preserves the domain and masks only the username portion
  - Example: `arfan.azhari@salt.id` → `ar********ri@salt.id`
  - Configurable `ShowFirst` and `ShowLast` for username (defaults: 2 and 2)
  - Domain is always fully preserved regardless of length
  - Handles edge cases: short usernames, single character usernames, invalid emails
  - Update `NewTxnWithEmailMasking` convenience function to use `EmailMask` type
  - Add 9 comprehensive unit tests covering various email masking scenarios
  - Works with field patterns, JSONPath expressions, and recursive patterns
- **BREAKING CHANGE: Remove deprecated masking APIs**
  - Remove `MaskConfig` type alias (use `MaskingConfig` instead)
  - Remove `MaskConfigs` type (use `[]MaskingConfig` instead)
  - Remove `GetMaskConfigs()` method (use `ConvertMaskingConfigs()` instead)
  - Remove `WithMaskConfigs()` option (use `WithMaskingConfig()` instead)

## [1.37.0] - 2025-10-10
- **Fix Gin middleware to propagate transaction to c.Request.Context() (#24)**
  - Transaction is now accessible from `c.Request.Context()` using `logmanager.FromContext(ctx)`
  - Enables downstream layers (service, repository, domain) to access transaction without Gin context
  - Uses existing `RequestWithTransactionContext` helper for proper context propagation
  - Add comprehensive test case `TestMiddleware_TransactionInRequestContext` to verify fix
  - Maintains backward compatibility: transaction still accessible via Gin context
  - All existing tests pass with no regressions

## [1.36.0] - 2025-10-07
- **Implement missing gRPC client and stream interceptors (#19)**
  - Add `UnaryClientInterceptor` for client-side unary RPC logging with automatic trace ID propagation
  - Add `StreamClientInterceptor` for client-side streaming RPC logging with message-level tracking
  - Add `StreamServerInterceptor` for server-side streaming RPC logging
  - Implement automatic trace ID extraction from context and injection into gRPC metadata
  - Add stream wrapper types (`wrappedClientStream`, `wrappedServerStream`) for proper lifecycle management
  - Add request/response logging for both client and server interceptors
  - Implement error handling with gRPC status code conversion to HTTP status
  - Add client example demonstrating all interceptor types
  - Update gRPC example with StreamServerInterceptor usage
  - Enhance documentation with trace ID propagation examples and best practices
  - All tests passing with no regressions

## [1.35.0] - 2025-09-27
- **Fix multipart/form-data and application/x-www-form-urlencoded request logging (#11)**
  - Add support for logging multipart form data with form fields and file metadata
  - Add support for logging URL-encoded form data
  - Implement intelligent form parsing: only extracts already-parsed forms to avoid consuming request body on client-side
  - Add `parseMultipartFormData()` for explicit parse-and-log scenarios (server-side handlers)
  - Add `extractMultipartFormData()` to extract data from pre-parsed forms
  - Add `parseFormData()` with body preservation for URL-encoded forms
  - Fix body consumption issue that broke client-side HTTP requests (e.g., clientmanager)
  - File uploads logged with metadata only: field name, filename, size, and headers (not file content)
  - Add 8 comprehensive unit tests for multipart and URL-encoded form parsing
  - Add 5 integration tests for end-to-end transaction logging with forms
  - Add 4 body preservation tests to verify forms remain accessible to downstream handlers
  - All existing tests pass with no regressions

## [1.34.0] - 2025-08-22
- **Environment-based configuration with automatic debug mode control**
  - Add `WithEnvironment` option to set application environment programmatically
  - Add automatic environment detection from `APP_ENV` environment variable
  - Implement intelligent debug mode control: disabled in production, enabled in development/staging
  - Add `Environment()` getter method to retrieve current environment setting
  - Support for custom environments with sensible defaults
- **Improved context-based logging methods with consistent naming**
  - Add new `InfoWithContext` function to replace `LogInfoWithContext` (deprecated)
  - Add new `ErrorWithContext` function to replace `LogErrorWithContext` (deprecated)
  - Add new `DebugWithContext` function with environment-aware debug logging
  - Maintain full backward compatibility with deprecated methods
- **Enhanced debug logging with environment awareness**
  - Debug logs are automatically suppressed in production environments
  - Debug logging respects application debug mode settings from transactions
  - Support for explicit debug override in production using `WithDebug()` option
  - Backward compatibility for contexts without transactions
- **Comprehensive testing and documentation**
  - Add 90%+ test coverage for all new functionality
  - Add comprehensive environment configuration tests covering all scenarios
  - Add debug logging integration tests with transaction support
  - Update README with detailed environment configuration documentation
  - Add migration guide for new logging methods

## [1.33.0] - 2025-08-13
- **Enhanced JSONPath masking with recursive support and array handling**
- Add recursive JSONPath pattern support (`$..field`) for comprehensive field masking across all nesting levels
- Implement case-insensitive substring matching for flexible field name matching (e.g., `$..token` matches `token`, `Token`, `authToken`, `usertoken`, etc.)
- Fix array handling at root level to prevent empty object logging in request/response bodies
- Add comprehensive unit tests for recursive patterns, array handling, and all masking types
- Update README with detailed JSONPath masking documentation including syntax reference, examples, and best practices
- Fix compilation error in Gorilla Mux middleware example by removing undefined function calls
- **Technical improvements:**
  - Added `recursiveMaskField` function for `$..` pattern support with deep traversal
  - Modified `toObj` function to handle both objects and arrays properly at root level
  - Enhanced field matching algorithm with case-insensitive substring comparison
  - Added 60+ new unit tests covering edge cases, recursive patterns, and all masking scenarios
  - Improved error handling for invalid JSONPath expressions and edge cases

## [1.32.0] - 2025-08-07
- Add `LogInfoWithContext` function for structured info logging with context support
- The new feature supports automatic trace ID extraction from context or transaction
- The optional third parameter allows adding custom fields to log entries
- Graceful handling of nil contexts and empty messages
- Comprehensive test coverage with six test scenarios covering all edge cases
- Consistent with existing `LogErrorWithContext` pattern and functionality

## [1.31.0] - 2025-07-23
- Add comprehensive test coverage for transaction logging across all transaction types
- Add test cases for HTTP, Consumer, Cron, gRPC, and Other transaction types
- Add test coverage for transaction tags functionality
- Add test coverage for complex multi-segment transaction workflows
- Improve test coverage for trace ID propagation across nested transactions
- Add JSONPath-based masking for advanced field filtering with support for complex JSON structures
- Introduce `MaskingConfig` with JSONPath expressions and field pattern matching
- Add advanced masking types: `FullMask`, `PartialMask`, and `HideMask` with configurable character visibility
- Add `StructMask` and `StructMaskWithConfig` for flexible struct-based masking with go-masker integration
- Add support for combining struct tags with JSONPath masking configurations
- Enhance transaction methods with comprehensive masking support in `TxnRecord`
- Add convenience functions for password, email, and credit card masking
- Update integration (`lmresty`) with comprehensive masking options and examples
- Deprecate `MaskConfigs` in favor of `MaskingConfig` with extended capabilities and backward compatibility

## [1.30.0] - 2025-07-09
- Fix `lmecho` middleware to properly update request context before calling handlers
- Add Echo framework integration documentation with `lmecho` middleware example
- Ensure logmanager.FromContext(ctx) works correctly in downstream operations

## [1.29.0] - 2025-07-03
- Add StartOtherSegmentWithContext function to create other segments from context.
- Add StartOtherSegmentWithMessage function to create other segments from context and a message.
- Modify StartOtherSegmentWithMessage to handle txn.End() internally and remove the return value.

## [1.28.0] - 2025-06-12
- Add support for http.StatusTemporaryRedirect (307) as a success status code

## [1.27.0] - 2025-05-16
- Add simple logs an error with the trace ID from the context.

## [1.26.0] - 2025-05-05
- Add TraceID propagation and X-Trace-Id header support

## [1.25.0] - 2025-05-05
- Refactor transaction handling with mutex for concurrency safety

## [1.24.0] - 2025-05-04
- Add transaction cloning and async request handling

## [1.23.0] - 2025-05-02
- Add query parameters logging for http api and http client native and resty.

## [1.22.0] - 2025-03-14
- Added the ability to skip logging for request and response bodies that exceed a specified size limit.
- Added a feature to expose all headers in HTTP requests.
