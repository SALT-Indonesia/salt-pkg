# Log Manager - Structured Logging for Go Applications

## Overview

LogManager is a comprehensive structured logging library for Go applications that provides:

- 🔍 **Automatic Trace ID Management** - Seamless request tracking across distributed systems
- 🛡️ **Built-in Data Masking** - Protect sensitive information in logs
- 🚀 **Framework Integrations** - Ready-to-use middleware for Gin, Echo, Gorilla Mux, gRPC, and more
- 📊 **Structured Logging** - JSON-formatted logs with consistent schema
- ⚡ **Performance Focused** - Minimal overhead with efficient implementations
- 🔗 **Context Propagation** - Automatic trace ID propagation through context

### Key Features

- **Transaction-based Logging**: Track complete request lifecycles with automatic timing
- **Segment Tracking**: Monitor individual operations (API calls, database queries, etc.)
- **Flexible Masking**: Full, partial, or complete hiding of sensitive fields
- **Async Support**: Safe transaction cloning for goroutines
- **Customizable**: Extensive configuration options for different use cases

### Log Schema

| Property    | Data Type | Description                                                                                                                                     |
|-------------|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| service     | string    | The service name configured for the application                                                                                                 |
| trace_id    | string    | Unique request identifier for distributed tracing. Auto-generated if not provided via headers (`X-Request-Id`, `X-Trace-Id`, etc.)              |
| method      | string    | HTTP method (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) for HTTP requests                                                                         |
| url         | string    | Request path (e.g., `/api/users` for HTTP endpoints, `/external/payment` for API calls)                                                         |
| latency     | number    | Total execution time in milliseconds                                                                                                            |
| status      | number    | HTTP response status code                                                                                                                       |
| request     | object    | Request payload (automatically captured for POST, PUT, DELETE, PATCH)                                                                           |
| query_param | object    | Query parameters (automatically captured for GET requests)                                                                                      |
| response    | object    | Response payload (with masking applied if configured)                                                                                           |
| msg         | string    | Log message for debug, info, or error level content                                                                                             |
| level       | string    | Log level: `error`, `info`, or `debug`                                                                                                          |
| type        | string    | Transaction type: `http` (HTTP endpoints), `api` (external calls), `database`, `grpc`, `consumer` (message queues), `other` (custom operations) |
| time        | timestamp | Log timestamp in RFC3339 format (e.g., `2023-07-14T11:20:22+07:00`)                                                                             |
| stack_trace | string    | Error stack trace with file and line information                                                                                                |
| tags        | array     | Custom tags for categorization (e.g., `["order", "payment"]`)                                                                                   |
| headers     | object    | HTTP headers (exposed selectively via `WithExposeHeaders` configuration)                                                                        |

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/SALT-Indonesia/salt-pkg/logmanager"
    "github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
)

func main() {
    // Configure LogManager
    app := logmanager.NewApplication(
        logmanager.WithService("user-service"),
        logmanager.WithDebug(),
        logmanager.WithMaskingConfig([]logmanager.MaskingConfig{
            {JSONPath: "$.password", Type: logmanager.FullMask},
            {JSONPath: "$.credit_card", Type: logmanager.PartialMask},
        }),
    )
	
    // Setup Gin with automatic logging
    router := gin.New()
    router.Use(lmgin.Middleware(app))
    
    router.POST("/users", createUser)
    router.Run(":8080")
}

func createUser(c *gin.Context) {
    c.JSON(201, gin.H{"id": "123", "status": "created"})
}
```

## Documentation

- 📖 [Architecture Guide](docs/ARCHITECTURE.md) - Understanding LogManager's design
- 🔧 [API Reference](docs/API_REFERENCE.md) - Complete API documentation
- 🚀 [Migration Guide](docs/MIGRATION_GUIDE.md) - Migrating from other logging libraries
- ❓ [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues and solutions

## Installation

```shell
go get -u github.com/SALT-Indonesia/salt-pkg/logmanager
```

## Initialize Application
Initialize the log manager by configuring the `Config` options and `App` settings within the `main` function or an `init` block:

### Basic Configuration

```go
// Create application with service name
app := logmanager.NewApplication(
    logmanager.WithService("your-service-name"),
    logmanager.WithDebug(true), // Enable debug mode (disable in production)
)

```

### Data Masking Configuration

Protect sensitive information in logs with flexible masking strategies:

```go
app := logmanager.NewApplication(
    logmanager.WithMaskingConfig([]logmanager.MaskingConfig{
        // Partial masking - show first/last characters
        {
            JSONPath: "$.credit_card",
            Type: logmanager.PartialMask,
        },
        // Full masking - replace entire value
        {
            JSONPath: "$.phone_number",
            Type: logmanager.FullMask,
        },
        // Hide completely - remove from logs
        {
            JSONPath: "$.password",
            Type: logmanager.HideMask,
        },
        // Nested field masking
        {
            JSONPath: "$.user.ssn",
            Type: logmanager.PartialMask,
        },
        // Array element masking
        {
            JSONPath: "$.users[*].password",
            Type: logmanager.FullMask,
        },
    }),
)
```

**Example output:**
```json
{
    "request": {
        "credit_card": "1234****3456",
        "phone_number": "************",
        "user": {
            "name": "John Doe",
            "ssn": "123****789"
        }
    },
    "response": {
        "status": "success"
        // password field completely removed
    }
}
```

### Custom Tags

Add custom tags for log categorization and filtering:

```go
app := logmanager.NewApplication(
    logmanager.WithTags(map[string]string{
        "environment": "production",
        "region": "us-east-1",
        "version": "1.2.0",
    }),
)
```

**Output:**
```json
{
    "service": "user-service",
    "level": "info",
    "tags": {
        "environment": "production",
        "region": "us-east-1",
        "version": "1.2.0"
    },
    "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Trace ID Configuration

Customize trace ID header extraction:

```go
app := logmanager.NewApplication(
    // Set the header key for trace ID extraction
    logmanager.WithTraceIDKey("X-Request-Id"), // Default: "trace_id"
    
    // Common alternatives:
    // logmanager.WithTraceIDKey("X-Trace-Id"),
    // logmanager.WithTraceIDKey("X-Correlation-Id"),
    // logmanager.WithTraceIDKey("X-Amzn-Trace-Id"),
)
```

### Header Exposure

Selectively expose HTTP headers in logs:

```go
app := logmanager.NewApplication(
    logmanager.WithExposeHeaders("X-User-Id"), // Enable header exposure
)

// Or expose all headers for specific endpoints
func handler(c *gin.Context) {
    transaction := logmanager.FromContext(c.Request.Context())
    transaction.ExposeAllHeaders() // Expose all headers for this request
    
    c.JSON(200, gin.H{"status": "ok"})
}
```

## Framework Integrations

### Gorilla Mux

```go
import (
    "net/http"
    "github.com/gorilla/mux"
    lm "github.com/salt-pkg/salt-pkg/logmanager"
    "github.com/salt-pkg/salt-pkg/logmanager/lmgorilla"
)

func main() {
    app := logmanager.NewApplication(
        logmanager.WithService("api-service"),
    )
    
    router := mux.NewRouter()
    router.Use(lmgorilla.Middleware(app))
    
    router.HandleFunc("/api/users", handleUsers).Methods("GET")
    
    http.ListenAndServe(":8080", router)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    transaction := logmanager.FromContext(ctx)
    
    // Track database operation
    ctx, dbSegment := transaction.StartDatabaseSegment(
		    "fetch-users",
            logmanager.DatabaseSegment{
                Query: "SELECT * FROM users WHERE active = $1"	
            }
		)
    defer dbSegment.End()
    
    // Perform query...
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"users": []}`))
}
```

### Gin

```go
import (
    "github.com/gin-gonic/gin"
    lm "github.com/salt-pkg/salt-pkg/logmanager"
    "github.com/salt-pkg/salt-pkg/logmanager/lmgin"
)

func main() {
    app := logmanager.NewApplication(
        logmanager.WithService("api-service"),
        logmanager.WithTraceIDKey("X-Request-Id"),
    )
    
    router := gin.New()
    router.Use(lmgin.Middleware(app))
    
    router.POST("/api/orders", createOrder)
    router.Run(":8080")
}

func createOrder(c *gin.Context) {
    ctx := c.Request.Context()
    transaction := logmanager.FromContext(ctx)
    
    // Automatic request/response logging
    // Request body with masked fields
    var req struct {
        UserID     string `json:"user_id"`
        CreditCard string `json:"credit_card"` // Will be masked
        Amount     float64 `json:"amount"`
    }
    c.BindJSON(&req)
    
    // Process order...
    
    c.JSON(201, gin.H{
        "order_id": "ORD-12345",
        "status": "created",
    })
}
```

### Echo

```go
import (
    "github.com/labstack/echo/v4"
    lm "github.com/salt-pkg/salt-pkg/logmanager"
    "github.com/salt-pkg/salt-pkg/logmanager/lmecho"
)

func main() {
    app := logmanager.NewApplication(
        logmanager.WithService("api-service"),
    )
    
    e := echo.New()
    e.Use(lmecho.Middleware(app))
    
    e.GET("/api/health", handleHealth)
    e.Logger.Fatal(e.Start(":8080"))
}

func handleHealth(c echo.Context) error {
    ctx := c.Request().Context()
    
    // Log custom message
    logmanager.LogErrorWithContext(ctx, nil) // Will log with trace ID
    
    return c.JSON(200, map[string]string{
        "status": "healthy",
    })
}
```

### gRPC

```go
import (
    "google.golang.org/grpc"
    lm "github.com/salt-pkg/salt-pkg/logmanager"
    "github.com/salt-pkg/salt-pkg/logmanager/lmgrpc"
)

func main() {
    app := logmanager.NewApplication(
        logmanager.WithService("grpc-service"),
        logmanager.WithTraceIDKey("X-Trace-Id"),
    )
    
    // Server setup
    server := grpc.NewServer(
        grpc.UnaryInterceptor(lmgrpc.UnaryServerInterceptor(app)),
        grpc.StreamInterceptor(lmgrpc.StreamServerInterceptor(app)),
    )
    
    // Client setup
    conn, _ := grpc.Dial("localhost:50051",
        grpc.WithUnaryInterceptor(lmgrpc.UnaryClientInterceptor(app)),
        grpc.WithStreamInterceptor(lmgrpc.StreamClientInterceptor(app)),
    )
}
```

### RabbitMQ Consumer

```go
import (
    amqp "github.com/rabbitmq/amqp091-go"
    lm "github.com/salt-pkg/salt-pkg/logmanager"
    "github.com/salt-pkg/salt-pkg/logmanager/lmrabbitmq"
)

func processMessage(app *logmanager.Application) func(amqp.Delivery) error {
    return lmrabbitmq.WrapConsumer(app, func(msg amqp.Delivery) error {
        ctx := context.Background()
        
        // Transaction automatically created with correlation ID
        transaction := logmanager.FromContext(ctx)
        
        // Process message...
        
        return nil
    })
}
```

## Segment Tracking

### API Calls

```go
func callExternalAPI(ctx context.Context) error {
    // Make API call
    req, _ := http.NewRequestWithContext(ctx, "POST",
    "https://api.payment.com/charge",
    bytes.NewReader([]byte(`{"amount": 100}`)))
	
    // Start API segment
    segment := logmanager.StartAPISegment(
		"payment-gateway",
        logmanager.ApiSegment{Request: req}
    )
    defer segment.End()
	
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        segment.NoticeError(err)
        return err
    }
	
    return nil
}
```

### Resty Client Integration

```go
import (
    "context"
    "github.com/go-resty/resty/v2"
    "github.com/SALT-Indonesia/salt-pkg/logmanager"
    "github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"
)

func callWithResty(ctx context.Context) error {
    client := resty.New()
    
    // Make the request with context that contains transaction
    resp, err := client.R().
        SetContext(ctx).
        SetBody(map[string]string{"key": "value"}).
        Post("https://api.example.com/endpoint")
    
    if err != nil {
        return err
    }
    
    // Create a transaction record from response
    txn := lmresty.NewTxn(resp)
    if txn != nil {
        defer txn.End()
    }
    
    return nil
}
```

### Database Operations

```go
import (
    "context"
    "database/sql"
    "github.com/SALT-Indonesia/salt-pkg/logmanager"
)

func getUserByID(ctx context.Context, userID string, db *sql.DB) (*User, error) {
    transaction := logmanager.FromContext(ctx)
    
    // Start database segment
    segment := logmanager.StartDatabaseSegment(transaction, logmanager.DatabaseSegment{
        Name:  "get-user",
        Table: "users",
        Query: "SELECT * FROM users WHERE id = $1",
        Host:  "localhost", // database host
    })
    defer segment.End()
    
    // Execute query
    var user User
    err := db.QueryRowContext(ctx, 
        "SELECT * FROM users WHERE id = $1", 
        userID).Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

### Custom Operations

```go
import (
    "context"
    "github.com/SALT-Indonesia/salt-pkg/logmanager"
)

func processData(ctx context.Context, data []byte) error {
    transaction := logmanager.FromContext(ctx)
    
    // Track custom operation with additional data
    segment := logmanager.StartOtherSegment(transaction, logmanager.OtherSegment{
        Name: "data-processing",
        Extra: map[string]interface{}{
            "input_size": len(data),
            "type":       "batch",
        },
    })
    defer segment.End()
    
    // Process data...
    result, err := transform(data)
    if err != nil {
        return err
    }
    
    // Custom processing completed successfully
    return nil
}

// Alternative using context-based method
func processDataWithContext(ctx context.Context, data []byte) error {
    // Track custom operation using context
    segment := logmanager.StartOtherSegmentWithContext(ctx, logmanager.OtherSegment{
        Name: "data-processing-v2",
        Extra: map[string]interface{}{
            "input_size": len(data),
            "type":       "batch",
        },
    })
    defer segment.End()
    
    // Process data...
    result, err := transform(data)
    if err != nil {
        return err
    }
    
    return nil
}
```

## Advanced Features

### Error Logging with Context

Quickly log errors with trace ID from context:

```go
func processData(ctx context.Context, data []byte) error {
    if err := validateData(data); err != nil {
        // Automatically includes trace ID
        logmanager.LogErrorWithContext(ctx, err)
        return err
    }
    
    return nil
}
```

### Info Logging with Context

Log informational messages with trace ID from context and optional additional fields:

```go
func processUserLogin(ctx context.Context, userID string) error {
    // Basic info logging with trace ID
    logmanager.LogInfoWithContext(ctx, "User login attempt started")
    
    // Info logging with additional fields
    fields := map[string]string{
        "user_id":    userID,
        "session_id": "session-abc123",
        "action":     "login",
    }
    logmanager.LogInfoWithContext(ctx, "User authenticated successfully", fields)
    
    return nil
}
```

**Function Signature:**
```go
func LogInfoWithContext(ctx context.Context, msg string, fields ...map[string]string)
```

**Parameters:**
- `ctx` - Context containing trace ID or transaction
- `msg` - Info message to log
- `fields` - Optional additional fields to include in the log (variadic parameter)

**Features:**
- Automatically extracts trace ID from context or transaction
- Supports optional additional fields for structured logging
- Handles nil contexts gracefully
- Uses JSON formatter for consistent output
- Safe to use in concurrent environments

### Goroutine Support

Safely use transactions in goroutines:

```go
func handleRequest(ctx context.Context) {
    transaction := logmanager.FromContext(ctx)
    
    // Synchronous operation
    doSyncWork(ctx)
    
    // Asynchronous operation
    go func() {
        // Clone transaction for goroutine
        ctx := logmanager.CloneTransactionToContext(context.Background(), transaction)
        transaction := logmanager.FromContext(ctx)
        
        ctx, segment := transaction.StartOtherSegment(
			"async-work",
            logmanager.OtherSegment{Name: "OtherSegment"}
        )
        defer segment.End()
        
        // Async work with same trace ID
        doAsyncWork(ctx)
    }()
}
```

### Skip Large Payloads

Prevent logging of large request/response bodies:

```go
func handleLargeUpload(c *gin.Context) {
    transaction := logmanager.FromContext(c.Request.Context())
    
    // Skip logging large payloads
    transaction.SkipRequest()  // Skip request body
    transaction.SkipResponse() // Skip response body
    
    // Process a large file...
    c.JSON(200, gin.H{"status": "uploaded"})
}
```

## Example Log Output

### HTTP Request Log
```json
{
  "service": "user-service",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "http",
  "method": "POST",
  "url": "/api/users",
  "status": 201,
  "latency": 145,
  "request": {
    "name": "John Doe",
    "email": "john@example.com",
    "password": "*******"
  },
  "response": {
    "id": "usr_123",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "level": "info",
  "time": "2024-01-15T10:30:00+07:00"
}
```

### API Segment Log
```json
{
  "service": "user-service",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "api",
  "name": "payment-gateway",
  "method": "POST",
  "url": "https://api.payment.com/charge",
  "status": 200,
  "latency": 523,
  "request": {
    "amount": 100,
    "currency": "USD"
  },
  "response": {
    "transaction_id": "TXN_123",
    "status": "success"
  },
  "level": "info",
  "time": "2024-01-15T10:30:01+07:00"
}
```

### Database Segment Log
```json
{
  "service": "user-service",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "database",
  "name": "create-user",
  "engine": "postgresql",
  "operation": "INSERT",
  "query": "INSERT INTO users (name, email) VALUES ($1, $2)",
  "latency": 15,
  "level": "info",
  "time": "2024-01-15T10:30:00+07:00"
}
```

## Best Practices

1. **Always use middleware** for automatic request/response logging
2. **Configure masking** for sensitive fields at application startup
3. **Use segments** to track individual operations within requests
4. **Clone transactions** for goroutines to maintain trace continuity
5. **Set appropriate log levels** - use debug mode only in development
6. **Handle errors properly** - use `segment.NoticeError()` or `LogErrorWithContext()`
7. **Skip large payloads** to avoid log bloat
8. **Use descriptive segment names** for better observability

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
