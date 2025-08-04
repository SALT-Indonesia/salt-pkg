# LogManager Architecture

## Overview

LogManager is a structured logging library designed for Go applications that provides standardized logging with trace ID tracking and seamless integration with popular frameworks. It follows clean architecture principles and provides a consistent logging interface across different types of applications.

## Core Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Application Layer                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────────┐ │
│  │ HTTP Server │  │ gRPC Service │  │ Message Queue Consumer │ │
│  └──────┬──────┘  └──────┬───────┘  └───────────┬────────────┘ │
│         │                 │                       │              │
├─────────┴─────────────────┴───────────────────────┴─────────────┤
│                      Framework Middlewares                       │
├──────────────────────────────────────────────────────────────────┤
│  ┌────────┐  ┌────────┐  ┌─────────┐  ┌──────┐  ┌───────────┐ │
│  │ lmgin  │  │ lmecho │  │lmgorilla│  │lmgrpc│  │lmrabbitmq │ │
│  └────┬───┘  └────┬───┘  └────┬────┘  └───┬──┘  └─────┬─────┘ │
│       └───────────┴───────────┴────────────┴───────────┘       │
├──────────────────────────────────────────────────────────────────┤
│                        Core Components                           │
├──────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────────┐ │
│  │ Transaction │  │  Application │  │      Segments          │ │
│  │             │  │              │  │ (API, DB, gRPC, etc)  │ │
│  └─────────────┘  └──────────────┘  └────────────────────────┘ │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────────┐ │
│  │   Context   │  │    Logger    │  │      Masking           │ │
│  │  Manager    │  │   (Logrus)   │  │    (Full/Partial/Hide) │ │
│  └─────────────┘  └──────────────┘  └────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

## Component Descriptions

### 1. Application (`app.go`)

The Application struct is the central configuration object that manages global settings:

```go
type Application struct {
    service         string              // Service name
    debug           bool                // Debug mode flag
    traceIDKey      string              // Key for trace ID in headers
    contextKey      string              // Key for storing in context
    maskingConfigs  []MaskingConfig     // Field masking rules
    tags            map[string]string   // Custom tags
    exposeHeaders   bool                // Header exposure control
}
```

**Key Responsibilities:**
- Service identification
- Debug mode management
- Trace ID configuration
- Masking rules storage
- Custom tag management

### 2. Transaction (`transaction.go`)

The Transaction is the core logging unit that tracks a complete operation lifecycle:

```go
type Transaction struct {
    mutex              sync.Mutex
    Application        *Application
    TraceID           string
    Type              string
    TransactionRecords []*TransactionRecord
    exposeHeaders      bool
}
```

**Key Features:**
- Thread-safe operations with mutex
- Hierarchical structure with transaction records
- Trace ID propagation
- Cloneable for async operations

### 3. Segments

Segments represent different types of operations within a transaction:

#### API Segment (`segment.go`)
```go
type ApiSegment struct {
    BasicSegment
    Method         string
    URL            string
    Request        interface{}
    Response       interface{}
    RequestHeader  http.Header
    ResponseHeader http.Header
    HTTPStatus     int
}
```

#### Database Segment
```go
type DatabaseSegment struct {
    BasicSegment
    Engine    string
    Operation string
    Query     string
    Result    interface{}
}
```

#### gRPC Segment
```go
type GrpcSegment struct {
    BasicSegment
    Service  string
    Method   string
    Request  interface{}
    Response interface{}
    Metadata metadata.MD
    Status   codes.Code
}
```

### 4. Context Management (`context.go`)

Provides seamless transaction propagation through context:

```go
// Store transaction in context
ctx = lm.ToContext(ctx, transaction)

// Retrieve transaction from context
transaction := lm.FromContext(ctx)

// Clone for goroutines
ctx = lm.CloneTransactionToContext(ctx, transaction)
```

### 5. Masking System (`mask.go`)

Three-tier masking system for sensitive data protection:

```go
type MaskingConfig struct {
    JSONPath string      // Path to field (e.g., "$.password")
    Type     MaskingType // FULL, PARTIAL, or HIDE
}
```

**Masking Types:**
- **FullMask**: Replaces entire value with asterisks
- **PartialMask**: Shows first/last characters only
- **HideMask**: Removes field from output entirely

## Data Flow

### 1. HTTP Request Flow (Gin Example)

```
HTTP Request
    │
    ▼
Gin Middleware (lmgin)
    │
    ├─> Extract/Generate Trace ID
    ├─> Create Transaction
    ├─> Capture Request Data
    │
    ▼
Handler Execution
    │
    ├─> Business Logic
    ├─> External API Calls (ApiSegment)
    ├─> Database Operations (DatabaseSegment)
    │
    ▼
Response Capture
    │
    ├─> Capture Response Data
    ├─> Calculate Latency
    ├─> Apply Masking Rules
    │
    ▼
Log Output (JSON)
```

### 2. Async Operation Flow

```
Main Transaction
    │
    ├─> Clone Transaction
    │       │
    │       ▼
    │   Goroutine
    │       │
    │       ├─> Async Operation
    │       ├─> Create Segment
    │       └─> Log with Same Trace ID
    │
    └─> Continue Main Flow
```

## Integration Patterns

### 1. HTTP Framework Integration

All HTTP framework integrations follow a similar pattern:

1. **Middleware Creation**: Accept Application configuration
2. **Request Interception**: Extract trace ID, create transaction
3. **Context Injection**: Store transaction in request context
4. **Response Capture**: Log complete request/response cycle

### 2. Client Integration (Resty)

```go
// Configure client with logging
client := resty.New()
client.OnBeforeRequest(lmresty.OnBeforeRequest(app, transaction))
client.OnAfterResponse(lmresty.OnAfterResponse(app, transaction))
```

### 3. Message Queue Integration

```go
// Wrap consumer with logging
consumer := lmrabbitmq.WrapConsumer(app, originalConsumer)
```

## Configuration Options

### Application Options

```go
app := lm.NewApplication(
    lm.WithService("my-service"),
    lm.WithDebug(true),
    lm.WithTraceIDKey("X-Request-Id"),
    lm.WithMaskingConfig([]lm.MaskingConfig{
        {JSONPath: "$.password", Type: lm.FullMask},
        {JSONPath: "$.credit_card", Type: lm.PartialMask},
    }),
    lm.WithTags(map[string]string{
        "environment": "production",
        "version": "1.0.0",
    }),
)
```

### Logger Configuration

```go
logger := lm.InitLogger(
    lm.WithLoggerDebug(true),
    lm.WithLoggerFile("app.log"),
    lm.WithLoggerMaxSize(100),    // MB
    lm.WithLoggerMaxBackups(3),
    lm.WithLoggerMaxAge(7),       // days
)
```

## Best Practices

### 1. Transaction Lifecycle

- Always create transactions at entry points (HTTP handlers, consumers)
- Use `CloneTransaction()` for goroutines
- Close segments properly with `defer segment.End()`

### 2. Context Propagation

- Pass context through all function calls
- Use `FromContext()` to retrieve transactions
- Never pass transactions as parameters when context is available

### 3. Masking Configuration

- Define masking rules at application startup
- Use JSONPath for nested field access
- Test masking rules with sample data

### 4. Performance Considerations

- Transaction cloning is lightweight (shallow copy)
- Masking is applied only during serialization
- Use debug mode sparingly in production

## Extension Points

### 1. Custom Segments

Create custom segments by embedding BasicSegment:

```go
type CustomSegment struct {
    lm.BasicSegment
    CustomField1 string
    CustomField2 int
}
```

### 2. Custom Middleware

Implement middleware for unsupported frameworks:

```go
func MyFrameworkMiddleware(app *lm.Application) MyMiddleware {
    return func(next Handler) Handler {
        return func(ctx Context) {
            transaction := lm.NewHTTPTransaction(app, traceID)
            ctx = lm.ToContext(ctx, transaction)
            // ... handle request/response
        }
    }
}
```

### 3. Custom Loggers

Replace the default Logrus logger:

```go
type CustomLogger interface {
    Info(args ...interface{})
    Debug(args ...interface{})
    Error(args ...interface{})
}
```

## Thread Safety

- **Transaction**: Thread-safe with mutex protection
- **Application**: Immutable after creation
- **Segments**: Not thread-safe, use within single goroutine
- **Context**: Safe for concurrent read, not for write

## Memory Management

- Transactions are garbage collected after handler completion
- Large request/response bodies are referenced, not copied
- Circular references avoided through careful design
- Context cleanup happens automatically

## Error Handling

- Panics in segments are recovered and logged
- Malformed JSON paths in masking fail silently
- Missing trace IDs are auto-generated
- Invalid configurations panic at startup