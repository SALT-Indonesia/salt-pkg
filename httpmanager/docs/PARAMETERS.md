# Request Parameters

This document covers query parameters, path parameters, and HTTP header handling.

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
