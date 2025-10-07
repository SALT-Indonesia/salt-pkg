# gRPC Integration Example

gRPC server and client implementation demonstrating logmanager integration with Protocol Buffers and interceptors.

## Features

- **Unary Interceptor**: Automatic request/response logging (server & client)
- **Stream Interceptor**: Streaming RPC logging support (server & client)
- **Error Handling**: gRPC status code integration
- **Trace ID Headers**: Cross-service tracing support
- **Protocol Buffers**: Structured message logging

## gRPC Service

The example implements a simple Greeter service:

```protobuf
service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}
```

## Key Components

### Server Setup
```go
app := logmanager.NewApplication(
    logmanager.WithAppName("grpc-server"),
    logmanager.WithTraceIDHeaderKey("X-Trace-Id"),
)

grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(
        lmgrpc.UnaryServerInterceptor(app),
    ),
    grpc.StreamInterceptor(
        lmgrpc.StreamServerInterceptor(app),
    ),
)
```

### Client Setup
```go
app := logmanager.NewApplication(
    logmanager.WithAppName("grpc-client"),
    logmanager.WithTraceIDHeaderKey("X-Trace-Id"),
)

conn, err := grpc.NewClient(
    serverAddress,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithUnaryInterceptor(lmgrpc.UnaryClientInterceptor(app)),
    grpc.WithStreamInterceptor(lmgrpc.StreamClientInterceptor(app)),
)
```

### Error Handling
```go
// Simulate error for demonstration
if req.GetName() == "error" {
    return nil, status.New(codes.InvalidArgument, "invalid name provided").Err()
}
```

## Protocol Buffers

The example includes proto files and generated Go code:
- `proto/hello_world.proto` - Service definition
- `proto/proto/` - Generated Go files

## Running the Examples

### Start the Server
```bash
go run main.go
```

Server runs on `localhost:50051`

### Run the Client
In a separate terminal:
```bash
go run main.go client.go -c
```

Or using a custom function:
```go
// In client.go
func main() {
    runClient()
}
```

## Testing with grpcurl

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Test successful request
grpcurl -plaintext -d '{"name":"World"}' \
  localhost:50051 proto.Greeter/SayHello

# Test error case
grpcurl -plaintext -d '{"name":"error"}' \
  localhost:50051 proto.Greeter/SayHello

# Test with custom trace ID
grpcurl -plaintext -H "X-Trace-Id: custom-trace-123" \
  -d '{"name":"Zazin"}' \
  localhost:50051 proto.Greeter/SayHello
```

## Expected Logging

The interceptor automatically logs:
- Request method and parameters
- Response data and status
- Processing time
- Error details (if any)
- Trace ID correlation

## Configuration Options

```go
logmanager.WithAppName("grpc-server")           // Service name
logmanager.WithTraceIDHeaderKey("X-Trace-Id")  // Trace header
logmanager.WithDebug()                          // Debug logging
```

## Best Practices

1. **Interceptor Setup**: Use both unary and stream interceptors for comprehensive logging
2. **Error Handling**: Return proper gRPC status codes
3. **Trace Headers**: Include trace ID in metadata for distributed tracing
4. **Client Logging**: Add client interceptors to track outgoing requests
5. **Service Naming**: Use descriptive service names
6. **Proto Structure**: Keep proto files organized

## Trace ID Propagation

The interceptors automatically handle trace ID propagation:
- **Server**: Extracts trace ID from incoming metadata or generates new one
- **Client**: Injects trace ID into outgoing metadata from context
- **Context**: Trace ID is stored in context for cross-service correlation

Example:
```go
// Server side - trace ID is automatically extracted
func (s *GreeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
    // Trace ID is already in context, accessible via:
    // traceID := ctx.Value(app.TraceIDContextKey())
    return &pb.HelloReply{Message: "Hello!"}, nil
}

// Client side - trace ID is automatically injected
ctx := context.WithValue(ctx, app.TraceIDContextKey(), "custom-trace-123")
resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "World"})
```

## Extending the Example

To add more features:
- Custom metadata handling
- Authentication/authorization
- Connection pooling
- Load balancing
- Retry mechanisms

## Next Steps

- [Messaging](../04-messaging/) - Async communication patterns
- [Data Masking](../05-masking/) - Sensitive data protection