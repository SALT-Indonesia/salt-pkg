# Changelog

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
