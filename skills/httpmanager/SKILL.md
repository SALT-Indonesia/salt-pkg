---
name: salt-httpmanager
description: Build type-safe Go HTTP REST APIs with the SALT-Indonesia/salt-pkg httpmanager library — generic handlers, automatic query/path binding, file uploads, CORS, and middleware. Use when creating an HTTP server or handlers with this library.
---

# httpmanager — type-safe Go HTTP server

`httpmanager` is a lightweight HTTP server framework built on Gorilla Mux. Handlers
are generic over their request/response types, so JSON (de)serialization, query/path
binding, and error responses are handled for you. It logs requests through
`logmanager`, which `NewServer` requires.

## Install

```bash
go get github.com/SALT-Indonesia/salt-pkg/httpmanager
# pulls in logmanager too — NewServer needs a *logmanager.Application
```

```go
import (
    "github.com/SALT-Indonesia/salt-pkg/httpmanager"
    "github.com/SALT-Indonesia/salt-pkg/logmanager"
)
```

## Minimal server

```go
type HelloRequest struct {
    Name string `json:"name"`
}
type HelloResponse struct {
    Message string `json:"message"`
}

func main() {
    app := logmanager.NewApplication()          // required dependency
    server := httpmanager.NewServer(app)        // optional: WithPort(...), etc.

    hello := httpmanager.NewHandler[HelloRequest, HelloResponse](
        http.MethodPost,
        func(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
            return &HelloResponse{Message: "Hello, " + req.Name}, nil
        },
    )

    server.Handle("/hello", hello)              // or server.POST("/hello", hello)
    log.Fatal(server.Start())                   // defaults to :8080
}
```

The handler signature is always:
`func(ctx context.Context, req *Req) (*Resp, error)`. The request body is decoded
into `*Req`; the returned `*Resp` is encoded to JSON with status `200`.

Register handlers with `server.Handle(pattern, h)` or the method shortcuts
`server.GET / POST / PUT / PATCH / DELETE(pattern, h)`.

## Query parameter binding

Tag a struct with `query:"..."` and bind it inside the handler. Supported field
types include `string`, `int`, `int64`, `bool`, and slices of those.

```go
type ListQuery struct {
    Limit int      `query:"limit"`
    Sort  string   `query:"sort"`
    Tags  []string `query:"tags"`
}

func(ctx context.Context, req *Empty) (*ListResponse, error) {
    var q ListQuery
    if err := httpmanager.BindQueryParams(ctx, &q); err != nil {
        return nil, err
    }
    // GET /products?limit=10&sort=price&tags=a&tags=b
    ...
}
```

## Path parameters

Use `{name}` (optionally with a regex constraint) in the route and read them from
the context.

```go
server.GET("/users/{id:[0-9]+}", handler)

func(ctx context.Context, req *Empty) (*UserResponse, error) {
    id := httpmanager.GetPathParams(ctx).Get("id")
    ...
}
```

Other context helpers: `httpmanager.GetHeader(ctx, key)`, `GetHeaders(ctx)`,
`GetQueryParams(ctx)`.

## Custom status codes and errors

Wrap the response type to control the status code, or return a `ResponseError[T]`
for typed error bodies.

```go
// 201 Created
func(ctx context.Context, req *CreateReq) (*httpmanager.ResponseSuccess[Created], error) {
    return &httpmanager.ResponseSuccess[Created]{
        StatusCode: http.StatusCreated,
        Body:       Created{ID: "123"},
    }, nil
}

// 422 with a typed error body
type ErrBody struct{ Code, Message string }

return nil, &httpmanager.ResponseError[ErrBody]{
    StatusCode: http.StatusUnprocessableEntity,
    Body:       ErrBody{Code: "VALIDATION", Message: "email required"},
    Err:        fmt.Errorf("missing email"), // logged server-side, not sent to client
}
```

## File uploads

```go
up := httpmanager.NewUploadHandler(
    http.MethodPost,
    "/tmp/uploads", // saved here
    func(ctx context.Context, files map[string][]*httpmanager.UploadedFile, form map[string][]string) (interface{}, error) {
        title := httpmanager.GetFormValue(form, "title")
        for _, f := range files["file"] {
            _ = f.Filename // also: f.Size, f.ContentType, f.SavedPath
        }
        return map[string]any{"saved": len(files)}, nil
    },
).WithMaxFileSize(50 << 20) // 50 MB

server.POST("/upload", up)
```

## CORS and middleware

CORS is enabled with a method (not an option):

```go
server.EnableCORS(
    []string{"https://example.com"}, // origins
    nil,                              // methods (nil = defaults)
    nil,                              // headers (nil = defaults)
    true,                            // allow credentials
)
```

Middleware are `func(http.Handler) http.Handler`. Add them globally:

```go
server := httpmanager.NewServer(app, httpmanager.WithMiddleware(mw1, mw2))
// or after construction:
server.Use(mw1)
// or per route:
server.HandleWithMiddleware("/protected", handler, authMW)
```

## Key server options (`NewServer(app, opts...)`)

- `WithPort("3000")` / `WithAddr(":9000")` — listen address
- `WithReadTimeout(d)` / `WithWriteTimeout(d)` — timeouts
- `WithSSL(true)` + `WithCertFile(p)` / `WithKeyFile(p)` (or `WithCertData` / `WithKeyData`) — TLS
- `WithMiddleware(mw...)` — global middleware
- `WithHealthCheckPath("/status")` / `WithoutHealthCheck()` — health endpoint (default `/health`)

Shut down gracefully with `server.Stop(ctx)`.

## More

See [`httpmanager/README.md`](../../httpmanager/README.md) and the
[`httpmanager/docs/`](../../httpmanager/docs/) directory (CONFIGURATION,
PARAMETERS, UPLOADS, RESPONSES, REDIRECTS) for the full reference.
