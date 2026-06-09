---
name: salt-logmanager
description: Structured logging with trace-ID propagation, data masking, and segment timing for Go apps using the SALT-Indonesia/salt-pkg logmanager library — plus middleware for Gin, Echo, Gorilla Mux, gRPC, RabbitMQ, and Resty. Use when adding request logging, distributed tracing, or sensitive-data masking with this library.
---

# logmanager — structured logging & tracing

`logmanager` provides JSON structured logging built around a per-request
`Transaction` that carries a trace ID. Framework middleware starts a transaction
automatically and puts it on the request `context`; you then log against the
context and time sub-operations as "segments". It also masks sensitive fields in
logged request/response bodies.

## Install

```bash
go get github.com/SALT-Indonesia/salt-pkg/logmanager
```

```go
import "github.com/SALT-Indonesia/salt-pkg/logmanager"
```

## Create the application

`NewApplication(opts...)` is configured with functional options. Create it once at
startup and share it.

```go
app := logmanager.NewApplication(
    logmanager.WithService("user-service"),
    logmanager.WithEnvironment("production"),
    logmanager.WithDebug(), // verbose; omit in prod
    logmanager.WithMaskingConfig([]logmanager.MaskingConfig{
        {JSONPath: "$..password", Type: logmanager.FullMask},
        {JSONPath: "$..token", Type: logmanager.FullMask},
        {FieldPattern: "apiKey", Type: logmanager.PartialMask, ShowFirst: 4, ShowLast: 4},
        {JSONPath: "$..email", Type: logmanager.EmailMask},
    }),
)
```

Common options: `WithService`, `WithAppName`, `WithEnvironment`, `WithDebug`,
`WithLogDir`, `WithSplitLevelOutput`, `WithTags`, `WithMaskingConfig`,
`WithExposeHeaders`, `WithSkipHeaders`, `WithTraceIDKey` / `WithTraceIDHeaderKey`
/ `WithTraceIDContextKey`, and `WithOpenTelemetry(WithOTelEndpoint(...), ...)`.

## Wire up middleware

Each integration starts a transaction per request and stores it in the context.

```go
import "github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"

r := gin.New()
r.Use(lmgin.Middleware(app))
```

Available integrations (all expose `Middleware(app)` except client/consumer ones):

- `integrations/lmgin` — Gin (`gin.HandlerFunc`)
- `integrations/lmecho` — Echo (`echo.MiddlewareFunc`)
- `integrations/lmgorilla` — Gorilla Mux (`mux.MiddlewareFunc`)
- `integrations/lmgrpc` — gRPC unary/stream server & client interceptors
- `integrations/lmrabbitmq` — RabbitMQ consumer (`NewConsumer`, then `txn.SetConsumer`)
- `integrations/lmresty` — Resty client (`NewTxn(resp)`, plus masking variants)

## Log against the context

Inside a handler, the request context already carries the transaction.

```go
func handler(ctx context.Context) error {
    logmanager.InfoWithContext(ctx, "processing", map[string]string{"user_id": "123"})

    if err := doWork(); err != nil {
        logmanager.ErrorWithContext(ctx, err) // logs with the request's trace ID
        return err
    }
    logmanager.DebugWithContext(ctx, "done")
    return nil
}
```

Get the transaction directly when you need the trace ID:
`tx := logmanager.FromContext(ctx); id := tx.TraceID()`.

## Time sub-operations (segments)

Segments record duration and metadata for downstream calls. Always `defer .End()`.

```go
tx := logmanager.FromContext(ctx)

// Database
db := logmanager.StartDatabaseSegment(tx, logmanager.DatabaseSegment{
    Name:  "get-user",
    Table: "users",
    Query: "SELECT * FROM users WHERE id = $1",
    Host:  "db.internal",
})
defer db.End()

// Outbound HTTP API (pass the *http.Request)
api := logmanager.StartApiSegment(logmanager.ApiSegment{Name: "payment-api", Request: req})
defer api.End()

// Anything else
seg := logmanager.StartOtherSegment(tx, logmanager.OtherSegment{
    Name:  "encode-image",
    Extra: map[string]interface{}{"format": "png"},
})
defer seg.End()
```

A segment can also be marked failed: `seg.NoticeError(err)` or
`seg.SetBusinessError(err)`.

## Goroutines

The transaction isn't safe to share across goroutines — clone it into a fresh
context first:

```go
go func() {
    gctx := logmanager.CloneTransactionToContext(context.Background(), ctx)
    seg := logmanager.StartOtherSegment(logmanager.FromContext(gctx), logmanager.OtherSegment{Name: "async"})
    defer seg.End()
    // ... same trace ID, safe to use
}()
```

## Data masking

`MaskingConfig` selects fields by `JSONPath`, `FieldPattern` (substring,
case-insensitive), or exact `Field`, and applies a `Type`:

- `FullMask` — replace the whole value (`"***"`)
- `PartialMask` — keep `ShowFirst`/`ShowLast` characters
- `EmailMask` — preserve the domain (`ar****ri@salt.id`)
- `HideMask` — single `*`

## More

See [`logmanager/README.md`](../../logmanager/README.md) and
[`logmanager/docs/`](../../logmanager/docs/) (ARCHITECTURE) for the full reference,
including OpenTelemetry export and each integration's setup.
