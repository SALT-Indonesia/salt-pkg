# Changelog

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
