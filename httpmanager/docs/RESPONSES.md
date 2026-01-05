# Response Handling

This document covers custom HTTP status codes for success and error responses.

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

### Common HTTP Status Codes for Errors

| Status Code | Constant                          | Use Case                                |
|-------------|-----------------------------------|-----------------------------------------|
| 400         | `http.StatusBadRequest`           | Validation errors, malformed requests   |
| 401         | `http.StatusUnauthorized`         | Authentication required                 |
| 403         | `http.StatusForbidden`            | Access denied                           |
| 404         | `http.StatusNotFound`             | Resource not found                      |
| 422         | `http.StatusUnprocessableEntity`  | Business logic errors                   |
| 500         | `http.StatusInternalServerError`  | Server errors                           |
