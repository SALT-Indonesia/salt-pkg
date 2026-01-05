# HTTP Manager - HTTP Server Module

`httpmanager` is a lightweight Go module for quickly setting up HTTP servers with configurable options and type-safe request handling. It simplifies creating HTTP endpoints with JSON request/response processing.

## Key Features

- **Type-safe request handling** with Go generics
- **Automatic query parameter binding** with struct tags (similar to Gin's `ShouldBindQuery`)
- **Path parameter support** with dynamic URL routing
- **Built-in health check endpoint** enabled by default at `/health`
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

## Documentation

For detailed documentation, see the following guides:

| Document | Description |
|----------|-------------|
| [Configuration](docs/CONFIGURATION.md) | Server options, health check, environment settings, SSL |
| [Parameters](docs/PARAMETERS.md) | Query parameters, path parameters, headers, automatic binding |
| [Uploads](docs/UPLOADS.md) | File upload handling, static file serving |
| [Responses](docs/RESPONSES.md) | Custom success/error status codes, ResponseSuccess, ResponseError |
| [Redirects](docs/REDIRECTS.md) | HTTP redirect functionality |

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

## Function Reference

### Server Functions

| Function | Description |
|----------|-------------|
| `NewServer(opts ...Option)` | Creates a new server with optional configuration |
| `server.Start()` | Starts the HTTP server |
| `server.Stop(ctx)` | Gracefully stops the server |
| `server.Handle(path, handler)` | Registers a handler for a path |
| `server.GET/POST/PUT/DELETE/PATCH(path, handler)` | HTTP method shortcuts |
| `server.Use(middleware...)` | Adds global middleware |
| `server.EnableCORS(...)` | Enables CORS with settings |

### Context Functions

| Function | Description |
|----------|-------------|
| `GetQueryParams(ctx)` | Gets query parameters from context |
| `BindQueryParams(ctx, dst)` | Binds query parameters to struct |
| `GetPathParams(ctx)` | Gets path parameters from context |
| `GetHeader(ctx, key)` | Gets a specific header value |
| `GetHeaders(ctx)` | Gets all headers |
| `GetFormValue(form, key)` | Gets first form field value |
| `GetFormValues(form, key)` | Gets all form field values |

### Response Types

| Type | Description |
|------|-------------|
| `ResponseSuccess[T]` | Custom success status codes (201, 202, 204, etc.) |
| `ResponseError[T]` | Custom error responses with status codes |
