# Basic CLI Example

Simple command-line application demonstrating logmanager core concepts with HTTP client requests.

## Features

- Basic transaction management
- HTTP request logging with Resty integration
- Context propagation
- Error handling and notification

## Key Concepts

### Transaction Lifecycle
```go
app := logmanager.NewApplication(
    logmanager.WithAppName("basic-cli"),
)

txn := app.Start("demo-cli", "cli", logmanager.TxnTypeOther)
ctx := txn.ToContext(context.Background())
defer txn.End()
```

### HTTP Request Logging
```go
client := resty.New()
resp, err := client.R().SetContext(ctx).Post("https://httpbin.org/post")

txn := lmresty.NewTxn(resp)
defer txn.End()

if err != nil {
    txn.NoticeError(err)
}
```

## Running

```bash
go run main.go
```

## Expected Output

The application will make HTTP requests and log transaction details including:
- Request/response timing
- HTTP status codes
- Request/response payloads
- Trace ID correlation

## Next Steps

After understanding basic concepts, explore:
- [HTTP Servers](../02-http-servers/) - Web framework integration
- [Data Masking](../05-masking/) - Sensitive data protection