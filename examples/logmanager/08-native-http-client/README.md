# Native HTTP Client with Logmanager

Comprehensive example demonstrating Go's standard `net/http` client integration with logmanager for HTTP operations and logging.

## Features

- Complete HTTP method coverage (GET, POST, PUT, DELETE)
- Query parameters and custom headers
- Request/response logging with trace IDs
- Timeout and custom client configuration
- Error handling and notification
- Native Go stdlib implementation

## Key Concepts

### Basic Integration with net/http

```go
app := logmanager.NewApplication(
    logmanager.WithAppName("native-http-client-demo"),
)

txn := app.Start("native-http-examples", "cli", logmanager.TxnTypeOther)
ctx := txn.ToContext(context.Background())
defer txn.End()
```

### Making HTTP Requests

```go
// Get transaction from context
tx := logmanager.FromContext(ctx)
startTime := time.Now()

// Create and execute request
req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Do(req)

// Read response
body, _ := io.ReadAll(resp.Body)
defer resp.Body.Close()

// Log the transaction
txn := tx.AddTxnNow("API-call", logmanager.TxnTypeApi, startTime)
txn.SetWebRequest(req)
txn.SetResponseBodyAndCode(body, resp.StatusCode)
defer txn.End()

if err != nil {
    txn.NoticeError(err)
}
```

### Query Parameters

```go
req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)

q := req.URL.Query()
q.Add("page", "1")
q.Add("limit", "10")
req.URL.RawQuery = q.Encode()
```

### Custom Headers

```go
req.Header.Set("User-Agent", "My-App/1.0")
req.Header.Set("Accept", "application/json")
req.Header.Set("X-Request-ID", "req-123")
```

### Custom Client Configuration

```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        10,
        MaxIdleConnsPerHost: 5,
        IdleConnTimeout:     30 * time.Second,
    },
}
```

## Examples Included

1. **Basic GET Request** - Simple HTTP GET with logging
2. **POST with JSON Body** - Sending JSON payloads
3. **PUT Request** - Update operations
4. **DELETE Request** - Resource deletion
5. **Query Parameters** - URL query parameter handling
6. **Custom Headers** - Setting multiple custom headers
7. **Timeout Handling** - Client timeout configuration
8. **Custom Client** - Advanced client configuration with transport settings

## Running the Example

```bash
cd examples/logmanager/08-native-http-client
go run main.go
```

## Expected Output

The application will execute 8 different HTTP request examples, each demonstrating:
- Request method and URL
- Response status codes
- Success/failure indicators
- Trace ID correlation across requests
- Timing information

## Log Output Features

- **Trace ID Propagation**: All requests within a transaction share the same trace ID
- **Request Details**: Method, URL, headers, query parameters, body
- **Response Metrics**: Status code, response size, timing
- **Error Tracking**: Automatic error logging with stack traces
- **Context Propagation**: Context carries trace information across calls

## Transaction Management Pattern

The native HTTP client example follows this pattern:

1. **Get transaction from context**: `tx := logmanager.FromContext(ctx)`
2. **Record start time**: `startTime := time.Now()`
3. **Execute HTTP request**: Standard `http.Client.Do(req)`
4. **Create transaction record**: `txn := tx.AddTxnNow(name, type, startTime)`
5. **Log request/response**: `txn.SetWebRequest()` and `txn.SetResponseBodyAndCode()`
6. **End transaction**: `defer txn.End()`

## Key Differences from Resty

Unlike Resty, the native HTTP client requires manual:
- Request body creation with `bytes.NewBuffer()`
- Response body reading with `io.ReadAll()`
- Transaction record creation with proper timing
- Explicit request/response logging setup

## Best Practices Demonstrated

1. **Always use context**: Create requests with `http.NewRequestWithContext(ctx)`
2. **Record timing**: Capture `startTime` before making requests
3. **Transaction per request**: Create transaction records for each HTTP call
4. **Error handling**: Use `txn.NoticeError(err)` for error tracking
5. **Deferred cleanup**: Always defer `txn.End()` and `resp.Body.Close()`
6. **Timeout configuration**: Set appropriate timeouts on HTTP clients
7. **Resource management**: Close response bodies to prevent leaks

## When to Use Native HTTP Client

Choose the native HTTP client when:
- You want minimal dependencies (stdlib only)
- You need fine-grained control over requests
- Working with legacy code using `net/http`
- Building libraries that should avoid external dependencies
- Performance is critical (fewer abstractions)

## Integration with Other Frameworks

This example shows client-side HTTP requests using stdlib. For other options, see:
- [Resty Client](../07-resty-client/) - Feature-rich HTTP client library
- [HTTP Servers](../02-http-servers/) - Server-side integrations
- [gRPC](../03-grpc/) - Service-to-service communication

## Common Use Cases

- **API Integration**: Calling external REST APIs
- **Microservices**: Service-to-service HTTP communication
- **Webhooks**: Sending HTTP callbacks
- **Data Fetching**: Retrieving data from remote endpoints
- **Health Checks**: Monitoring endpoint availability

## Advanced Topics

### Connection Pooling

```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}
client := &http.Client{Transport: transport}
```

### Custom Timeout Strategies

```go
// Per-request timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
```

### Error Retry Logic

Implement retry logic with exponential backoff for transient failures:

```go
for i := 0; i < maxRetries; i++ {
    resp, err := client.Do(req)
    if err == nil && resp.StatusCode < 500 {
        // Success or client error
        break
    }
    time.Sleep(backoffDuration * time.Duration(i+1))
}
```

## Troubleshooting

### Transaction not logged
- Ensure you call `txn.SetWebRequest(req)` before ending transaction
- Verify transaction record is created with proper timing

### Missing trace IDs
- Confirm context contains transaction: `tx := logmanager.FromContext(ctx)`
- Check that parent transaction context is propagated

### Memory leaks
- Always close response bodies: `defer resp.Body.Close()`
- End transactions: `defer txn.End()`

### Context cancellation
- Use `http.NewRequestWithContext()` for cancellation support
- Handle context deadline exceeded errors properly

## Performance Considerations

- Reuse HTTP clients (don't create new clients per request)
- Configure connection pooling appropriately
- Set reasonable timeouts to prevent hanging
- Close response bodies to return connections to pool
- Use context for request cancellation

## Next Steps

- Compare with [Resty Client](../07-resty-client/) for feature-rich alternative
- Explore [Data Masking](../05-masking/) for sensitive data protection
- Review [HTTP Servers](../02-http-servers/) for server-side patterns
