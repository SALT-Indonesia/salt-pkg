# HTTP Manager - HTTP Server Module

`httpmanager` is a lightweight Go module for quickly setting up HTTP servers with configurable options and type-safe request handling. It simplifies creating HTTP endpoints with JSON request/response processing.

## Key Features

- **Type-safe request handling** with Go generics
- **Automatic query parameter binding** with struct tags (similar to Gin's `ShouldBindQuery`)
- **Path parameter support** with dynamic URL routing
- **Built-in CORS middleware** with configurable settings
- **File upload handling** with multipart form support
- **Static file serving** with automatic content type detection
- **HTTP redirects** with comprehensive redirect functionality
- **SSL/TLS support** with certificate and key configuration
- **Middleware support** at both server and handler levels
- **Flexible error handling** with custom JSON error responses

## Installation

To use this module in your Go project:

```bash
go get github.com/yourusername/httpmanager
```

## Quick Start

Here's a minimal example to get a server up and running:

```go
package main

import (
	"context"
	"httpserver"
	"log"
	"net/http"
)

// HelloRequest Define request and response types
type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

func main() {
	// Create a new server with default options
	server := httpmanager.NewServer()

	// Create and register a handler
	helloHandler := httpmanager.NewHandler(http.MethodPost, func(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
		return &HelloResponse{
			Message: "Hello, " + req.Name + "!",
		}, nil
	})

	// Register the handler with a route
	server.Handle("/hello", helloHandler)

	// Start the server (blocks until server stops)
	log.Panic(server.Start())
}
```

## Core Components

### Server

The `Server` struct is the main part that manages HTTP request routing and server lifecycle:

```go
server := httpmanager.NewServer()
server.Handle("/path", myHandler)
server.Start() // Blocks until server stops
```

To stop the server gracefully:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Stop(ctx)
```

### Middleware

The module supports middleware for request processing. Middleware can be applied at the server level (global middleware) or at the handler level (handler-specific middleware).

#### CORS Middleware

The module provides built-in CORS (Cross-Origin Resource Sharing) middleware:

```go
// Enable CORS with default settings (allows all origins)
server := httpmanager.NewServer(
    httpmanager.WithCORS(nil, nil, nil, false)
)

// Or with custom settings
server := httpmanager.NewServer(
    httpmanager.WithCORS(
        []string{"https://example.com", "https://api.example.com"}, // allowed origins
        []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},        // allowed methods
        []string{"Content-Type", "Authorization"},                  // allowed headers
        true,                                                       // allow credentials
    )
)

// Enable CORS on an existing server
server.EnableCORS(
    []string{"https://example.com"},
    nil,  // use default methods
    nil,  // use default headers
    true, // allow credentials
)
```

The CORS middleware handles preflight OPTIONS requests automatically and sets the appropriate CORS headers.

#### Server Middleware

Server middleware is applied to all handlers registered with the server:

```go
// Create middleware
loggingMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

// Apply middleware when creating the server
server := httpmanager.NewServer(
    httpmanager.WithMiddleware(loggingMiddleware)
)

// Or add middleware after server creation
server.Use(loggingMiddleware)
```

#### Handler-Specific Middleware

Handler-specific middleware is applied only to specific handlers:

```go
// Create a handler
handler := httpmanager.NewHandler(http.MethodPost, handlerFunc)

// Add middleware to the handler
handler.Use(authMiddleware, validationMiddleware)

// Register the handler with the server
server.Handle("/path", handler.WithMiddleware())
```

#### Combining Server and Handler Middleware

You can also register a handler with specific middleware in addition to the server middleware:

```go
// Register a handler with specific middleware
server.HandleWithMiddleware("/path", handler, authMiddleware, validationMiddleware)
```

This applies both the handler-specific middleware and the server middleware to the handler.

### Handler

The `Handler` provides type-safe request handling with automatic JSON serialization/deserialization:

```go
handler := httpmanager.NewHandler[RequestType, ResponseType](http.MethodPost, handlerFunc)
```

The handler ensures that:
- Only the specified HTTP method is accepted
- Request bodies are automatically decoded into your request type
- Responses are automatically encoded as JSON
- Appropriate HTTP status codes are returned for errors

## Server Configuration

The server can be configured using option functions:

```go
server := httpmanager.NewServer(
	httpmanager.WithAddr(":3000"),
	httpmanager.WithReadTimeout(15 * time.Second),
	httpmanager.WithWriteTimeout(15 * time.Second),
)
```

### Available Options

| Option             | Description                         | Default                                                 |
|--------------------|-------------------------------------|---------------------------------------------------------|
| `WithAddr`         | Sets the server address             | `:PORT` from environment variable or `:8080` if not set |
| `WithPort`         | Sets just the server port           | `PORT` from environment variable or `8080` if not set   |
| `WithReadTimeout`  | Sets the read timeout               | `10s`                                                   |
| `WithWriteTimeout` | Sets the write timeout              | `10s`                                                   |
| `WithSSL`          | Enables/disables SSL                | `false`                                                 |
| `WithCertFile`     | Sets SSL certificate file path      | `""`                                                    |
| `WithKeyFile`      | Sets SSL key file path              | `""`                                                    |
| `WithCertData`     | Sets SSL certificate as string data | `""`                                                    |
| `WithKeyData`      | Sets SSL key as string data         | `""`                                                    |
| `WithCORS`         | Enables CORS with custom settings   | `disabled`                                              |

## Request Handling

The module uses Go generics to provide type-safe request handling:

```go
// Define request and response types
type UserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

// Create a handler function with proper types
handlerFunc := func(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	// Process request...
	return &UserResponse{
		ID:       "user-123",
		Username: req.Username,
		Status:   "active",
	}, nil
}

// Create and register the handler
userHandler := httpmanager.NewHandler(http.MethodPost, handlerFunc)
server.Handle("/users", userHandler)
```

## Query Parameter Handling

The module provides support for handling query parameters from the URL:

```go
// Create a handler function that uses query parameters
handlerFunc := func(ctx context.Context, req *UserRequest) (*UserResponse, error) {
    // Get query parameters from the context
    queryParams := httpmanager.GetQueryParams(ctx)

    // Get a single value for a parameter
    page := queryParams.Get("page")

    // Get all values for a parameter (for parameters with multiple values)
    tags := queryParams.GetAll("tag")

    // Process request with query parameters...
    return &UserResponse{
        ID:       "user-123",
        Username: req.Username,
        Status:   "active",
        Page:     page,
        Tags:     tags,
    }, nil
}

// Create and register the handler
userHandler := httpmanager.NewHandler(http.MethodGet, handlerFunc)
server.Handle("/users", userHandler)
```

### QueryParams Methods

The `QueryParams` type provides methods for accessing query parameters:

| Method                        | Description                                         |
|-------------------------------|-----------------------------------------------------|
| `Get(key string) string`      | Returns the first value for the given parameter key |
| `GetAll(key string) []string` | Returns all values for the given parameter key      |

### Query Parameter Functions

The module provides the following functions for working with query parameters:

| Function                                            | Description                                          |
|----------------------------------------------------|------------------------------------------------------|
| `GetQueryParams(ctx context.Context) QueryParams` | Extracts all query parameters from the context       |
| `BindQueryParams(ctx context.Context, dst interface{}) error` | Automatically binds query parameters to a struct using tags |

Query parameters are automatically extracted from the request URL and added to the context, making them accessible in your handler functions.

## Automatic Query Parameter Binding

The module provides automatic query parameter binding similar to Gin's `ShouldBindQuery`, allowing you to bind query parameters directly to a struct using struct tags. This eliminates the need for manual parameter extraction and provides type-safe query parameter handling.

### Basic Usage

```go
// Define a struct with query parameter tags
type UserSearchQuery struct {
    Name         string   `query:"name"`
    MinAge       int      `query:"min_age"`
    MaxAge       int      `query:"max_age"`
    Active       bool     `query:"active"`
    Tags         []string `query:"tags"`
    IncludeEmail bool     `query:"include_email"`
}

type UserSearchRequest struct{}

type UserSearchResponse struct {
    Users []map[string]interface{} `json:"users"`
    Total int                      `json:"total"`
    Query UserSearchQuery          `json:"query"`
}

// Create a handler that uses automatic binding
handlerFunc := func(ctx context.Context, req *UserSearchRequest) (*UserSearchResponse, error) {
    // Automatically bind query parameters to struct
    var queryParams UserSearchQuery
    if err := httpmanager.BindQueryParams(ctx, &queryParams); err != nil {
        return nil, err
    }

    // Use the bound parameters
    users := []map[string]interface{}{}
    for i := 1; i <= 3; i++ {
        user := map[string]interface{}{
            "id":     fmt.Sprintf("user_%d", i),
            "name":   fmt.Sprintf("%s_%d", queryParams.Name, i),
            "age":    queryParams.MinAge + i,
            "active": queryParams.Active,
            "tags":   queryParams.Tags,
        }

        if queryParams.IncludeEmail {
            user["email"] = fmt.Sprintf("%s_%d@example.com", queryParams.Name, i)
        }

        users = append(users, user)
    }

    return &UserSearchResponse{
        Users: users,
        Total: len(users),
        Query: queryParams, // Return the parsed query parameters
    }, nil
}

// Register the handler
searchHandler := httpmanager.NewHandler("GET", handlerFunc)
server.GET("/users/search", searchHandler)
```

### Supported Data Types

The automatic binding supports the following data types:

| Go Type      | Description                                    | Example Query                    |
|--------------|------------------------------------------------|----------------------------------|
| `string`     | Single string value                            | `?name=john`                     |
| `int`        | Integer value                                  | `?age=25`                        |
| `int64`      | 64-bit integer value                           | `?count=1234567890`              |
| `bool`       | Boolean value (true/false, 1/0, on/off)       | `?active=true`                   |
| `[]string`   | Array of strings                               | `?tags=go&tags=web&tags=api`     |
| `[]int`      | Array of integers                              | `?ids=1&ids=2&ids=3`             |
| `[]int64`    | Array of 64-bit integers                       | `?values=123&values=456`         |
| `[]bool`     | Array of booleans                              | `?flags=true&flags=false`        |

### Example URL Queries

Here are examples of URLs that work with the above struct:

```
GET /users/search?name=john&min_age=18&max_age=65&active=true&tags=developer&tags=golang&include_email=true
GET /users/search?name=alice&min_age=25&active=false
GET /users/search?tags=frontend&tags=react&tags=typescript
GET /users/search?name=bob&active=true&include_email=false
```

### Error Handling

The `BindQueryParams` function gracefully handles various scenarios:

- **Invalid values**: Parameters that cannot be converted to the target type are skipped
- **Missing parameters**: Fields remain at their zero values if no corresponding query parameter exists
- **Invalid input**: Returns early if destination is not a pointer to a struct
- **Empty context**: Returns early if no query parameters are present in the context

```go
// Example with error handling
var queryParams UserSearchQuery
if err := httpmanager.BindQueryParams(ctx, &queryParams); err != nil {
    return nil, &httpmanager.ResponseError[ErrorResponse]{
        Err:        err,
        StatusCode: http.StatusBadRequest,
        Body: ErrorResponse{
            Code:    "QUERY_BIND_ERROR",
            Message: "Failed to bind query parameters",
            Data:    nil,
        },
    }
}
```

### Migration from Manual Extraction

**Before (Manual approach):**
```go
queryParams := httpmanager.GetQueryParams(ctx)
name := queryParams.Get("name")
ageStr := queryParams.Get("min_age")
minAge := 0
if ageStr != "" {
    if val, err := strconv.Atoi(ageStr); err == nil {
        minAge = val
    }
}
activeStr := queryParams.Get("active")
active := false
if activeStr == "true" {
    active = true
}
tags := queryParams.GetAll("tags")
```

**After (Automatic binding):**
```go
type QueryParams struct {
    Name   string   `query:"name"`
    MinAge int      `query:"min_age"`
    Active bool     `query:"active"`
    Tags   []string `query:"tags"`
}

var params QueryParams
err := httpmanager.BindQueryParams(ctx, &params)
```

### Benefits

- **Type Safety**: Automatic conversion to appropriate Go types
- **Reduced Boilerplate**: No need for manual parameter extraction and conversion
- **Better Maintainability**: Query parameters are clearly defined in struct tags
- **Error Resilience**: Invalid values are handled gracefully without causing panics
- **Familiar Syntax**: Similar to JSON binding and Gin's `ShouldBindQuery`

The automatic query parameter binding feature makes it easier to work with complex query parameters while maintaining type safety and reducing repetitive code.

## Path Parameter Handling

The module supports Gin-like path parameters for dynamic URL routing. Path parameters allow you to capture values from URL segments and use them in your handlers:

```go
// Define request and response types
type GetUserRequest struct{}

type GetUserResponse struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Message  string `json:"message"`
}

// Create a handler function that uses path parameters
handlerFunc := func(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    // Get path parameters from the context
    pathParams := httpmanager.GetPathParams(ctx)

    // Get the user ID from the path parameter
    userID := pathParams.Get("id")

    // Get query parameters as well
    queryParams := httpmanager.GetQueryParams(ctx)
    includeEmail := queryParams.Get("include_email")

    response := &GetUserResponse{
        ID:      userID,
        Name:    fmt.Sprintf("User %s", userID),
        Message: fmt.Sprintf("Retrieved user with ID: %s", userID),
    }

    if includeEmail == "true" {
        response.Email = fmt.Sprintf("user%s@example.com", userID)
    }

    return response, nil
}

// Create and register the handler using HTTP method-specific routes
userHandler := httpmanager.NewHandler("GET", handlerFunc)

// Register with path parameters using server method shortcuts
server.GET("/user/{id}", userHandler)                    // Single path parameter
server.GET("/user/{id}/profile/{section}", userHandler)  // Multiple path parameters
```

### HTTP Method Shortcuts

The module provides convenient methods for registering handlers with path parameters:

```go
// HTTP method shortcuts with path parameter support
server.GET("/user/{id}", getUserHandler)
server.POST("/user/{id}", createUserHandler)
server.PUT("/user/{id}", updateUserHandler)
server.DELETE("/user/{id}", deleteUserHandler)
server.PATCH("/user/{id}", patchUserHandler)
```

### Path Parameter Patterns

Path parameters use the Gorilla Mux syntax with curly braces `{}`:

```go
// Basic path parameter
server.GET("/user/{id}", handler)

// Multiple path parameters
server.GET("/user/{id}/profile/{section}", handler)

// Path parameter with regex pattern (numeric only)
server.GET("/user/{id:[0-9]+}", handler)

// Path parameter with regex pattern (alphanumeric)
server.GET("/user/{id:[a-zA-Z0-9]+}", handler)

// Wildcard path parameter (captures remaining path)
server.GET("/files/{path:.+}", handler)
```

### PathParams Methods

The `PathParams` type provides methods for accessing path parameters:

| Method                         | Description                                          |
|--------------------------------|------------------------------------------------------|
| `Get(key string) string`       | Returns the value for the given path parameter key  |
| `Has(key string) bool`         | Checks if a path parameter exists                    |
| `Keys() []string`              | Returns all parameter keys                           |

### Example URLs and Matching

Here are examples of how URLs match path parameter patterns:

| Pattern                        | URL                              | Parameters                        |
|--------------------------------|----------------------------------|-----------------------------------|
| `/user/{id}`                   | `/user/123`                      | `id: "123"`                       |
| `/user/{id}/profile/{section}` | `/user/123/profile/settings`     | `id: "123"`, `section: "settings"` |
| `/user/{id:[0-9]+}`            | `/user/123`                      | `id: "123"` (numeric only)        |
| `/files/{path:.+}`             | `/files/docs/readme.txt`         | `path: "docs/readme.txt"`         |

### Complete Example

```go
func main() {
    server := httpmanager.NewServer(logmanager.NewApplication())
    server.EnableCORS([]string{"*"}, nil, nil, false)

    // Path parameter routes
    server.GET("/user/{id}", user.NewGetUserHandler())
    server.PUT("/user/{id}", user.NewUpdateUserHandler())
    server.GET("/user/{id}/profile/{section}", user.NewGetUserProfileHandler())

    log.Println("Path parameter examples:")
    log.Println("GET http://localhost:8080/user/123")
    log.Println("GET http://localhost:8080/user/123?include_email=true")
    log.Println("PUT http://localhost:8080/user/123 (with JSON body)")
    log.Println("GET http://localhost:8080/user/123/profile/settings")

    log.Panic(server.Start())
}
```

Path parameters are automatically extracted from the request URL and added to the context, making them accessible in your handler functions alongside query parameters and headers.

## HTTP Header Handling

The module provides support for accessing HTTP request headers from the context:

```go
// Create a handler function that uses HTTP headers
handlerFunc := func(ctx context.Context, req *UserRequest) (*UserResponse, error) {
    // Get a specific header value
    requestID := httpmanager.GetHeader(ctx, "X-Request-ID")

    // Get all headers
    headers := httpmanager.GetHeaders(ctx)

    // Access specific headers from the header collection
    contentType := headers.Get("Content-Type")
    authorization := headers.Get("Authorization")

    // Process request with headers...
    return &UserResponse{
        ID:         "user-123",
        Username:   req.Username,
        Status:     "active",
        RequestID:  requestID,
    }, nil
}

// Create and register the handler
userHandler := httpmanager.NewHandler(http.MethodGet, handlerFunc)
server.Handle("/users", userHandler)
```

### Header Methods

The module provides the following methods for accessing HTTP headers:

| Method                                              | Description                                    |
|-----------------------------------------------------|------------------------------------------------|
| `GetHeader(ctx context.Context, key string) string` | Returns the value for the specified header key |
| `GetHeaders(ctx context.Context) http.Header`       | Returns all headers as an http.Header object   |

HTTP headers are automatically extracted from the request and added to the context via the `RequestKey` constant, making them accessible in your handler functions.

### Common HTTP Headers

Here are some common HTTP headers you might want to access:

| Header          | Description                                                      |
|-----------------|------------------------------------------------------------------|
| `Authorization` | Contains authentication credentials                              |
| `Content-Type`  | Indicates the media type of the resource                         |
| `Accept`        | Informs the server about the types of data that can be sent back |
| `X-Request-ID`  | A unique identifier for the request (often used for tracing)     |
| `User-Agent`    | Information about the client making the request                  |

## File Upload Handling

The module provides support for handling file uploads via multipart form data:

```go
// Create an upload handler
uploadHandler := httpmanager.NewUploadHandler(
    http.MethodPost,
    "./uploads", // Directory where files will be saved
    func(ctx context.Context, files map[string][]*httpmanager.UploadedFile, form map[string][]string) (interface{}, error) {
        // Process uploaded files
        for fieldName, uploadedFiles := range files {
            for _, file := range uploadedFiles {
                // Access file metadata
                fmt.Printf("Field: %s, File: %s, Size: %d, Type: %s, Saved at: %s\n",
                    fieldName, file.Filename, file.Size, file.ContentType, file.SavedPath)
            }
        }

        // Access form values
        name := ""
        if values, ok := form["name"]; ok && len(values) > 0 {
            name = values[0]
        }

        // Return a response
        return map[string]string{
            "status": "success",
            "message": "Files uploaded successfully",
        }, nil
    },
)

// Optionally set maximum file size (default is 10MB)
uploadHandler.WithMaxFileSize(20 << 20) // 20MB

// Add middleware if needed
uploadHandler.Use(authMiddleware)

// Register the handler
server.Handle("/upload", uploadHandler.WithMiddleware())
```

### UploadedFile Structure

The `UploadedFile` struct provides metadata about uploaded files:

```go
type UploadedFile struct {
	Filename    string // Original filename
	Size        int64  // File size in bytes
	ContentType string // MIME type
	SavedPath   string // Path where the file was saved
}
```

### HTML Form Example

Here's an example HTML form that works with the upload handler:

```html
<form action="/upload" method="post" enctype="multipart/form-data">
    <input type="text" name="name" placeholder="Your name">
    <input type="file" name="files" multiple>
    <input type="file" name="avatar">
    <button type="submit">Upload</button>
</form>
```

## Static File Serving

The module provides support for serving static files, particularly images, with appropriate content types:

```go
// Create a static file handler
staticHandler := httpmanager.NewStaticHandler(
    http.MethodGet,
    "./static", // Directory containing static files
)

// Add middleware if needed
staticHandler.Use(cacheMiddleware)

// Register the handler
server.Handle("/images/", staticHandler.WithMiddleware())
```

The static handler automatically:
- Serves files from the specified directory
- Sets appropriate content type headers based on file extensions
- Applies cache control headers for better performance
- Prevents directory traversal attacks
- Handles common image formats (JPEG, PNG, GIF, SVG, WebP, etc.)

### Accessing Static Files

Once the static handler is registered, files can be accessed via URLs like:

```
https://yourserver.com/images/logo.png
https://yourserver.com/images/photos/vacation.jpg
```

The handler maps these URLs to files in the static directory:

```
./static/logo.png
./static/photos/vacation.jpg
```

### Supported Image Formats

The static handler supports the following image formats with appropriate content types:

| Extension   | Content Type  |
|-------------|---------------|
| .jpg, .jpeg | image/jpeg    |
| .png        | image/png     |
| .gif        | image/gif     |
| .svg        | image/svg+xml |
| .webp       | image/webp    |
| .ico        | image/x-icon  |
| .bmp        | image/bmp     |
| .tiff, .tif | image/tiff    |

## Custom HTTP Status Codes for Success Responses

The httpmanager module provides `ResponseSuccess` - a generic type for returning custom HTTP status codes in successful responses (e.g., 201 Created, 202 Accepted, 204 No Content) instead of the default 200 OK.

### ResponseSuccess Structure

```go
type ResponseSuccess[T any] struct {
    StatusCode int // HTTP status code (201, 202, 204, etc.)
    Body       T   // Response structure that will be serialized to JSON
}
```

### Common Success Status Codes

| Status Code | Constant                     | Use Case                                    |
|-------------|------------------------------|---------------------------------------------|
| 200         | `http.StatusOK`              | Standard successful response (default)      |
| 201         | `http.StatusCreated`         | Resource successfully created               |
| 202         | `http.StatusAccepted`        | Request accepted for processing             |
| 204         | `http.StatusNoContent`       | Successful request with no response body    |
| 206         | `http.StatusPartialContent`  | Partial resource returned (range requests)  |

### Usage Example

**Default behavior (returns 200 OK):**
```go
func createUserHandler(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    // Create user logic...
    return &CreateUserResponse{
        ID:      "user-123",
        Name:    req.Name,
        Message: "User created successfully",
    }, nil
}
```

**With custom 201 Created status:**
```go
func createUserHandler(ctx context.Context, req *CreateUserRequest) (*httpmanager.ResponseSuccess[CreateUserResponse], error) {
    // Create user logic...
    return &httpmanager.ResponseSuccess[CreateUserResponse]{
        StatusCode: http.StatusCreated,
        Body: CreateUserResponse{
            ID:      "user-123",
            Name:    req.Name,
            Message: "User created successfully",
        },
    }, nil
}
```

**Response (with 201 status code):**
```json
{
  "id": "user-123",
  "name": "John Doe",
  "message": "User created successfully"
}
```

### Additional Examples

**202 Accepted for asynchronous operations:**
```go
func processDataHandler(ctx context.Context, req *ProcessDataRequest) (*httpmanager.ResponseSuccess[ProcessDataResponse], error) {
    // Queue data for processing...
    return &httpmanager.ResponseSuccess[ProcessDataResponse]{
        StatusCode: http.StatusAccepted,
        Body: ProcessDataResponse{
            Status: "Accepted",
            TaskID: "task-12345",
            Message: "Request queued for processing",
        },
    }, nil
}
```

**204 No Content for delete operations:**
```go
func deleteUserHandler(ctx context.Context, req *DeleteUserRequest) (*httpmanager.ResponseSuccess[struct{}], error) {
    // Delete user logic...
    return &httpmanager.ResponseSuccess[struct{}]{
        StatusCode: http.StatusNoContent,
        Body: struct{}{},
    }, nil
}
```

### Backward Compatibility

The `ResponseSuccess` type is optional. Existing handlers that return regular response types will continue to work with the default 200 OK status code. This feature only applies when you explicitly use `ResponseSuccess[T]` as the return type.

## Error Handling with ResponseError

The httpmanager module provides `ResponseError` - a generic error type for custom JSON error responses.

> **Note:** `CustomError` is now deprecated. Please migrate to `ResponseError[T]` for more flexible and type-safe error handling. `ResponseError` allows complete customization of error response structure while preserving original errors for server-side logging.

### ResponseError Structure

```go
type ResponseError[T any] struct {
    Err        error // Original error (for server-side logging only)
    StatusCode int   // HTTP status code (400, 401, 422, 500, etc.)
    Body       T     // Custom JSON response structure
}
```

### Field Descriptions

| Field        | Description                                                                         |
|--------------|-------------------------------------------------------------------------------------|
| `Err`        | Original Go error preserved for logging/debugging. Not sent to client.              |
| `StatusCode` | HTTP status code: `400` (Bad Request), `422` (Business Error), `500` (Server Error) |
| `Body`       | Your custom struct that gets serialized to JSON response                            |

### Basic Handler Usage

```go
// Define your error response format
type ErrorResponse struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

// Use in handler
func createUserHandler(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    // Validation error - 400 status
    if req.Name == "" {
        return nil, &httpmanager.ResponseError[ErrorResponse]{
            Err:        fmt.Errorf("name is required"),
            StatusCode: http.StatusBadRequest,
            Body: ErrorResponse{
                Code:    "VIRB01001",
                Message: "Name field is required",
                Data:    nil,
            },
        }
    }

    // Server error - 500 status
    if req.Name == "database_error" {
        return nil, &httpmanager.ResponseError[ErrorResponse]{
            Err:        fmt.Errorf("database connection failed"),
            StatusCode: http.StatusInternalServerError,
            Body: ErrorResponse{
                Code:    "VISE01001",
                Message: "Database unavailable",
                Data:    nil,
            },
        }
    }

    return &CreateUserResponse{ID: 123, Message: "User created"}, nil
}
```

### Response Examples

**400 Error:**
```json
{
  "code": "VIRB01001",
  "message": "Name field is required",
  "data": null
}
```

**500 Error:**
```json
{
  "code": "VISE01001", 
  "message": "Database unavailable",
  "data": null
}
```

## SSL Support

To enable HTTPS with SSL:

```go
server := httpmanager.NewServer(
	httpmanager.WithSSL(true),
	httpmanager.WithCertFile("server.crt"),
	httpmanager.WithKeyFile("server.key"),
)
```

Alternatively, you can provide a certificate and key as string data:

```go
server := httpmanager.NewServer(
	httpmanager.WithSSL(true),
	httpmanager.WithCertData(certString),
	httpmanager.WithKeyData(keyString),
)
```

## HTTP Redirects

The httpmanager module provides comprehensive HTTP redirect functionality similar to Gin's implementation. You can redirect users to different URLs with various HTTP status codes.

### Basic Redirect Functions

The module provides utility functions for common redirect scenarios:

```go
// Basic redirect with custom status code
httpmanager.Redirect(w, r, http.StatusFound, "http://example.com")

// Redirect with 302 (Found) status code
httpmanager.RedirectToURL(w, r, "http://example.com")

// Redirect with 301 (Moved Permanently) status code
httpmanager.RedirectPermanent(w, r, "http://example.com")
```

### Context-Based Redirects

For more convenient usage, you can use the `Context` type which provides Gin-like redirect methods:

```go
// Using RedirectHandler for context-based redirects
redirectHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
    // Redirect with custom status code
    c.Redirect(http.StatusFound, "http://example.com")
    
    // Or use convenience methods
    c.RedirectToURL("http://example.com")         // 302 Found
    c.RedirectPermanent("http://example.com")     // 301 Moved Permanently
})

server.Handle("/old-path", redirectHandler.WithMiddleware())
```

### Redirect Handler

The `RedirectHandler` provides a specialized handler for redirect operations:

```go
// Create a redirect handler
redirectHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
    // Access query parameters
    targetURL := c.GetQueryParams().Get("redirect_to")
    if targetURL == "" {
        targetURL = "http://default-example.com"
    }
    
    // Redirect to the target URL
    c.RedirectToURL(targetURL)
})

// Add middleware if needed
redirectHandler.Use(authMiddleware, loggingMiddleware)

// Register with the server
server.GET("/redirect", redirectHandler.WithMiddleware())
```

### Dynamic Redirects with Path Parameters

You can create dynamic redirects using path parameters:

```go
// Redirect handler with path parameters
redirectHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
    // Get path parameters
    pathParams := c.GetPathParams()
    userID := pathParams.Get("id")
    
    // Get query parameters
    queryParams := c.GetQueryParams()
    section := queryParams.Get("section")
    
    // Build redirect URL
    redirectURL := fmt.Sprintf("https://newdomain.com/users/%s", userID)
    if section != "" {
        redirectURL += "?section=" + section
    }
    
    c.RedirectPermanent(redirectURL)
})

// Register with path parameter
server.GET("/old-user/{id}", redirectHandler.WithMiddleware())
```

### Redirect Status Codes

The module supports all standard HTTP redirect status codes:

| Status Code | Constant                        | Method              | Description                          |
|-------------|--------------------------------|---------------------|--------------------------------------|
| 301         | `http.StatusMovedPermanently`  | `RedirectPermanent` | Permanent redirect                   |
| 302         | `http.StatusFound`             | `RedirectToURL`     | Temporary redirect (default)         |
| 303         | `http.StatusSeeOther`          | `Redirect`          | See other (POST to GET)              |
| 307         | `http.StatusTemporaryRedirect` | `Redirect`          | Temporary redirect (method preserved) |
| 308         | `http.StatusPermanentRedirect` | `Redirect`          | Permanent redirect (method preserved) |

### Complete Redirect Examples

#### Example 1: Simple Domain Migration

```go
func main() {
    server := httpmanager.NewServer(logmanager.NewApplication())
    server.EnableCORS([]string{"*"}, nil, nil, false)

    // Redirect all old domain traffic to new domain
    migrationHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
        oldPath := c.Request.URL.Path
        c.RedirectPermanent("https://newdomain.com" + oldPath)
    })

    server.GET("/old-api/{path:.*}", migrationHandler.WithMiddleware())
    
    log.Panic(server.Start())
}
```

#### Example 2: Conditional Redirects

```go
func main() {
    server := httpmanager.NewServer(logmanager.NewApplication())
    server.EnableCORS([]string{"*"}, nil, nil, false)

    // Conditional redirect based on user agent
    mobileRedirectHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
        userAgent := c.GetHeader("User-Agent")
        
        if strings.Contains(strings.ToLower(userAgent), "mobile") {
            c.RedirectToURL("https://m.example.com" + c.Request.URL.Path)
        } else {
            c.RedirectToURL("https://www.example.com" + c.Request.URL.Path)
        }
    })

    server.GET("/", mobileRedirectHandler.WithMiddleware())
    
    log.Panic(server.Start())
}
```

#### Example 3: POST to GET Redirect

```go
func main() {
    server := httpmanager.NewServer(logmanager.NewApplication())
    server.EnableCORS([]string{"*"}, nil, nil, false)

    // Handle form submission with redirect
    formHandler := httpmanager.NewRedirectHandler(http.MethodPost, func(c *httpmanager.Context) {
        // Process form data here (if needed)
        // Then redirect to success page
        c.Redirect(http.StatusSeeOther, "/success")
    })

    server.POST("/submit-form", formHandler.WithMiddleware())
    
    log.Panic(server.Start())
}
```

### Context Methods

The `Context` type provides the following redirect methods:

| Method                                  | Description                                    |
|----------------------------------------|------------------------------------------------|
| `Redirect(code int, location string)`  | Redirect with custom HTTP status code         |
| `RedirectToURL(location string)`       | Redirect with 302 status (Found)              |
| `RedirectPermanent(location string)`   | Redirect with 301 status (Moved Permanently)  |
| `GetQueryParams()`                     | Access query parameters for dynamic redirects  |
| `GetPathParams()`                      | Access path parameters for dynamic redirects   |
| `GetHeader(key string)`                | Access request headers                         |

### Error Handling

The redirect functions will panic if an invalid HTTP status code is provided:

```go
// This will panic - status code must be 3xx
c.Redirect(http.StatusOK, "http://example.com") // Panics!

// Valid redirect status codes (3xx)
c.Redirect(http.StatusMovedPermanently, "http://example.com")  // 301 ✓
c.Redirect(http.StatusFound, "http://example.com")             // 302 ✓
c.Redirect(http.StatusSeeOther, "http://example.com")          // 303 ✓
```

HTTP redirects are essential for URL migration, mobile detection, form processing, and API versioning. The httpmanager module provides flexible redirect functionality that integrates seamlessly with the existing middleware and routing system.