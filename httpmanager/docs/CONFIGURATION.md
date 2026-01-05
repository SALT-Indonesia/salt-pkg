# Server Configuration

The server can be configured using option functions:

```go
server := httpmanager.NewServer(
	httpmanager.WithAddr(":3000"),
	httpmanager.WithReadTimeout(15 * time.Second),
	httpmanager.WithWriteTimeout(15 * time.Second),
)
```

## Available Options

| Option             | Description                         | Default                                                 |
|--------------------|-------------------------------------|---------------------------------------------------------|
| `WithAddr`         | Sets the server address             | `:PORT` from environment variable or `:8080` if not set |
| `WithPort`         | Sets just the server port           | `PORT` from environment variable or `8080` if not set   |
| `WithReadTimeout`  | Sets the read timeout               | `10s`                                                   |
| `WithWriteTimeout` | Sets the write timeout              | `10s`                                                   |
| `WithSSL`          | Enables/disables SSL                | `false`                                                 |
| `WithCertFile`     | Sets SSL certificate file path      | `""`                                                    |
| `WithKeyFile`      | Sets SSL key file path              | `""`                                                    |
| `WithCertData`     | Sets SSL certificate as string data | `""`                                                    |
| `WithKeyData`      | Sets SSL key as string data         | `""`                                                    |
| `WithCORS`         | Enables CORS with custom settings   | `disabled`                                              |

## Health Check Endpoint

The module includes a built-in health check endpoint that is **enabled by default**. This endpoint is useful for container orchestration systems like Kubernetes, load balancers, and monitoring tools to verify that your service is running.

### Default Behavior

By default, the health check endpoint is available at `GET /health` and returns:

```json
{"status":"ok"}
```

With HTTP status code `200 OK` and `Content-Type: application/json` header.

```go
// Health check is automatically enabled at /health
server := httpmanager.NewServer(logmanager.NewApplication())
```

### Custom Health Check Path

You can customize the health check endpoint path:

```go
// Custom health check path
server := httpmanager.NewServer(
    logmanager.NewApplication(),
    httpmanager.WithHealthCheckPath("/api/health"),
)
```

### Disabling Health Check

If you don't need the built-in health check (e.g., you want to implement your own), you can disable it:

```go
// Disable health check endpoint
server := httpmanager.NewServer(
    logmanager.NewApplication(),
    httpmanager.WithoutHealthCheck(),
)
```

### Health Check Options

| Option                    | Description                                    | Default     |
|---------------------------|------------------------------------------------|-------------|
| `WithHealthCheck()`       | Explicitly enables health check                | `enabled`   |
| `WithHealthCheckPath()`   | Sets custom path and enables health check      | `/health`   |
| `WithoutHealthCheck()`    | Disables the health check endpoint             | -           |

### HTTP Method Restrictions

The health check endpoint only accepts `GET` requests. Other HTTP methods will receive a `405 Method Not Allowed` response.

### Example Usage

```go
package main

import (
    "log"
    "github.com/yourusername/httpmanager"
    "github.com/yourusername/logmanager"
)

func main() {
    // Create server with custom health check path
    server := httpmanager.NewServer(
        logmanager.NewApplication(),
        httpmanager.WithHealthCheckPath("/api/v1/health"),
        httpmanager.WithAddr(":8080"),
    )

    // Register your application handlers
    // server.Handle("/users", userHandler)

    log.Println("Health check available at: http://localhost:8080/api/v1/health")
    log.Panic(server.Start())
}
```

**Testing the health check:**

```bash
curl http://localhost:8080/health
# Response: {"status":"ok"}
```

## Environment Configuration

The httpmanager module integrates with logmanager's environment-based configuration. The `APP_ENV` environment variable controls debug mode and logging behavior.

### APP_ENV Environment Variable

| APP_ENV Value | Debug Mode | 404 Debug Logging | Description                    |
|---------------|------------|-------------------|--------------------------------|
| `production`  | Disabled   | Disabled          | Production environment         |
| `development` | Enabled    | Enabled           | Development environment        |
| `staging`     | Enabled    | Enabled           | Staging environment            |
| (not set)     | Enabled    | Enabled           | Defaults to development        |

### Debug Mode Features

When debug mode is enabled (non-production environment):
- **404 Debug Logging**: All 404 Not Found responses are logged with method, path, and query parameters
- **Verbose Logging**: Additional debug information is available via `logmanager.DebugWithContext`

### Configuration Examples

**Production environment (debug disabled):**
```bash
export APP_ENV=production
./myapp
```

**Development environment (debug enabled):**
```bash
export APP_ENV=development
./myapp
```

**Explicit debug mode override:**
```go
// Force debug mode even in production
app := logmanager.NewApplication(
    logmanager.WithEnvironment("production"),
    logmanager.WithDebug(), // Explicitly enable debug
)
server := httpmanager.NewServer(app)
```

### 404 Debug Logging

When debug mode is enabled, all 404 responses are automatically logged:

```json
{
  "level": "debug",
  "type": "http",
  "method": "GET",
  "path": "/non-existent-path",
  "query": "foo=bar",
  "trace_id": "abc-123-def",
  "msg": "404 Not Found",
  "time": "2025-12-09T10:00:00+07:00"
}
```

This is useful for:
- Identifying misconfigured client URLs
- Debugging missing routes during development
- Tracking 404 patterns for API versioning decisions

### Non-Production Warning

When the server starts in a non-production environment, a warning message is displayed:

```
[WARNING] Server is running in 'development' environment. Set APP_ENV=production for production deployments.
```

## SSL Support

To enable HTTPS with SSL:

```go
server := httpmanager.NewServer(
	httpmanager.WithSSL(true),
	httpmanager.WithCertFile("server.crt"),
	httpmanager.WithKeyFile("server.key"),
)
```

Alternatively, you can provide a certificate and key as string data:

```go
server := httpmanager.NewServer(
	httpmanager.WithSSL(true),
	httpmanager.WithCertData(certString),
	httpmanager.WithKeyData(keyString),
)
```
