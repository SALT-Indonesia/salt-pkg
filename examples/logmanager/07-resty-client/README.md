# Resty Client with Logmanager

Comprehensive example demonstrating Resty HTTP client integration with logmanager for various HTTP operations and logging patterns.

## Features

- Complete HTTP method coverage (GET, POST, PUT, DELETE)
- Query parameters and custom headers
- Request/response logging with trace IDs
- Sensitive data masking for security
- Authentication token handling
- Error handling and notification
- Timeout and retry patterns

## Key Concepts

### Basic Resty Integration

```go
app := logmanager.NewApplication(
    logmanager.WithAppName("resty-client-demo"),
)

txn := app.Start("resty-examples", "cli", logmanager.TxnTypeOther)
ctx := txn.ToContext(context.Background())
defer txn.End()
```

### Making HTTP Requests

```go
client := resty.New()
resp, err := client.R().
    SetContext(ctx).
    SetBody(payload).
    Post("https://api.example.com/users")

txn := lmresty.NewTxn(resp)
defer txn.End()

if err != nil {
    txn.NoticeError(err)
}
```

### Data Masking for Sensitive Information

```go
// Automatically mask password, token, and secret fields
txn := lmresty.NewTxnWithPasswordMasking(resp)
defer txn.End()
```

### Custom Headers and Query Parameters

```go
resp, err := client.R().
    SetContext(ctx).
    SetHeaders(map[string]string{
        "User-Agent": "My-App/1.0",
        "X-Request-ID": "req-123",
    }).
    SetQueryParams(map[string]string{
        "page": "1",
        "limit": "10",
    }).
    Get("https://api.example.com/data")
```

## Examples Included

1. **Basic GET Request** - Simple HTTP GET with logging
2. **POST with JSON Body** - Sending JSON payloads
3. **PUT with Headers** - Update requests with custom headers
4. **DELETE Request** - Resource deletion with logging
5. **Query Parameters** - URL query parameter handling
6. **Custom Headers** - Setting multiple custom headers
7. **Data Masking** - Protecting sensitive information in logs
8. **Authentication** - Bearer token authentication
9. **Timeout Handling** - Dealing with delayed responses

## Running the Example

```bash
cd examples/logmanager/07-resty-client
go run main.go
```

## Expected Output

The application will execute 9 different HTTP request examples, each demonstrating:
- Request method and URL
- Response status codes
- Success/failure indicators
- Trace ID correlation across requests
- Masked sensitive data in logs

## Log Output Features

- **Trace ID Propagation**: All requests within a transaction share the same trace ID
- **Request Details**: Method, URL, headers, query parameters
- **Response Metrics**: Status code, response size, timing
- **Error Tracking**: Automatic error logging with stack traces
- **Data Masking**: Sensitive fields like passwords automatically masked

## Best Practices Demonstrated

1. **Always use context**: Pass context with trace information to all requests
2. **Transaction management**: Create transaction records for all HTTP calls
3. **Error handling**: Use `txn.NoticeError(err)` for proper error tracking
4. **Deferred cleanup**: Always defer `txn.End()` after creating transactions
5. **Sensitive data**: Use masking functions for requests containing passwords, tokens, or PII
6. **Custom headers**: Include request IDs and version information for better tracing

## Integration with Other Frameworks

This example shows client-side HTTP requests. For server-side integration, see:
- [HTTP Servers](../02-http-servers/) - Gin, Echo, Gorilla middleware
- [gRPC](../03-grpc/) - Service-to-service communication

## Resty Features Used

- Client configuration
- Request builder pattern
- Context propagation
- Header and query parameter management
- JSON body marshaling
- Authentication helpers
- Response handling

## Common Use Cases

- **API Integration**: Making calls to external REST APIs
- **Microservices**: Service-to-service HTTP communication
- **Data Fetching**: Retrieving data from remote endpoints
- **Webhooks**: Sending notifications to external services
- **Authentication Flows**: OAuth, JWT token management

## Troubleshooting

### Request not logged
- Ensure context is set: `.SetContext(ctx)`
- Verify transaction is created: `lmresty.NewTxn(resp)`

### Sensitive data visible in logs
- Use `NewTxnWithPasswordMasking()` or `NewTxnWithMasking()`
- Configure custom masking for specific fields

### Missing trace IDs
- Confirm transaction context: `ctx := txn.ToContext(context.Background())`
- Check that context is passed to all requests

## Next Steps

- Explore [Native HTTP Client](../08-native-http-client/) for standard library usage
- Review [Data Masking](../05-masking/) for advanced masking patterns
- Check [HTTP Methods](../06-http-methods/) for server-side examples
