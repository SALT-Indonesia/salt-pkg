# HTTP Server Examples

Web framework integrations demonstrating logmanager middleware for popular Go HTTP frameworks.

## Frameworks Covered

| Framework | Port | Features |
|-----------|------|----------|
| **Gin** | 8001 | Trace ID middleware, segments |
| **Echo** | 8002 | Clean architecture, database segments |
| **Gorilla Mux** | 8003 | Data masking, header exposure |

## Common Features

- **Automatic Request Logging**: Request/response capture
- **Trace ID Generation**: Unique request tracking
- **Error Handling**: Structured error responses
- **Performance Monitoring**: Request timing
- **Middleware Integration**: Framework-specific middleware

## Framework-Specific Features

### Gin (`gin/`)
- Custom trace ID middleware
- Header exposure configuration
- Segment-based operation tracking

### Echo (`echo/`)
- Repository pattern implementation
- Database segment logging
- Clean architecture demonstration

### Gorilla Mux (`gorilla/`)
- JSON masking configuration
- Tag-based categorization
- Content type handling

## Running Examples

```bash
# Run all servers in separate terminals
cd gin && go run main.go       # http://localhost:8001
cd echo && go run main.go      # http://localhost:8002
cd gorilla && go run main.go   # http://localhost:8003
```

## Testing Endpoints

```bash
# Gin server
curl -X POST http://localhost:8001/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test"}'

# Echo server
curl http://localhost:8002/users

# Gorilla Mux server
curl -X POST http://localhost:8003/post/json \
  -H "Content-Type: application/json" \
  -d '{"password":"secret","apiKey":"key123456789"}'
```

## Key Concepts

### Middleware Integration
Each framework requires specific middleware setup:

```go
// Gin
r.Use(traceIDMiddleware(), lmgin.Middleware(app))

// Echo
e.Use(lmecho.Middleware(app))

// Gorilla Mux
router.Use(lmgorilla.Middleware(app))
```

### Configuration Options
```go
app := logmanager.NewApplication(
    logmanager.WithAppName("http-server"),
    logmanager.WithTraceIDContextKey("xid"),
    logmanager.WithExposeHeaders("Custom-Header"),
    logmanager.WithMaskingConfig(maskingConfigs),
)
```

## Best Practices

1. **Trace ID Management**: Generate unique trace IDs per request
2. **Context Propagation**: Pass context through service layers
3. **Error Notification**: Use `txn.NoticeError()` for error tracking
4. **Segment Usage**: Track database and external service calls
5. **Configuration**: Set up masking for sensitive data

## Next Steps

- [gRPC Integration](../03-grpc/) - Service-to-service communication
- [Messaging](../04-messaging/) - Async processing patterns