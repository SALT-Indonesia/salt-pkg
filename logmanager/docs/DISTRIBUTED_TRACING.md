# Distributed Tracing

LogManager can emit OpenTelemetry spans alongside its structured log lines. This
guide explains how to make a LogManager transaction join a trace that was
started by another service, so a single request can be followed end to end.

## The two trace IDs

A LogManager log line carries two different identifiers, and they are **not**
interchangeable:

| Field           | Format                        | Meaning                                                            |
| --------------- | ----------------------------- | ------------------------------------------------------------------ |
| `trace_id`      | UUID v4, e.g. `3ca1ad80-...`  | The application-level correlation ID. It travels in the `X-Trace-Id` header and is what LogManager has always propagated between services. |
| `otel_trace_id` | W3C hex, 32 chars, e.g. `5473c7194d4f5328f2c61f0b403b0eaf` | The OpenTelemetry trace ID. It travels in the W3C `traceparent` header and is what the tracing backend groups spans by. |

Because `trace_id` is propagated by LogManager while `otel_trace_id` is derived
from the OpenTelemetry context, a service can have a perfectly correct
`trace_id` and still emit a brand new, unrelated `otel_trace_id`. When that
happens the backend shows two separate traces instead of one.

To correlate the two, LogManager exports the application-level ID as a span
attribute:

```
logmanager.trace_id = 3ca1ad80-dc0b-44d6-9d10-36340e256bbd
```

Search that attribute in your tracing backend to jump from a log line to its
span (and back).

## Why the trace used to break

`Application.start` created the root span with a hardcoded
`context.Background()` and a `nil` parent, and the `Start*` constructors had no
way to accept a context. An inbound `traceparent` therefore had nowhere to go:
the transaction always started a new trace.

Three things had to change for a transaction to be part of an inbound trace:

1. the root span must be created from a context that carries the caller's span;
2. the caller's `traceparent` must be extracted from the inbound carrier;
3. outgoing calls must inject `traceparent` so the next service can continue
   the trace.

## Wiring it up

### 1. Reuse one `TracerProvider`

Pass the application's own provider to LogManager so both emit spans through the
same pipeline:

```go
provider := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter),
    sdktrace.WithResource(res),
)

app := logmanager.NewApplication(
    logmanager.WithAppName("my-service"),
    logmanager.WithService("my-service"),
    logmanager.WithTracerProvider(provider),
)
```

`WithTracerProvider` takes precedence over the built-in exporter
(`WithOpenTelemetry(...)`), which is what makes LogManager's spans and the
application's own spans land in the same trace. If you do not supply a provider,
LogManager keeps building its own exporter exactly as before.

As soon as tracing is enabled, LogManager installs the W3C Trace Context
propagator if none is configured yet (see `otel.EnsureW3CPropagator`). Without
it the OpenTelemetry global propagator is a no-op and `Inject` writes nothing,
which is a silent failure. If you use another propagator (B3, Jaeger, ...), call
`otel.SetTextMapPropagator` yourself — LogManager leaves it untouched.

### 2. Inbound HTTP

Extract the incoming context and hand it to the context-aware constructor:

```go
ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

tx := app.StartHttpWithContext(ctx, r.Header.Get("X-Trace-Id"), r.Method+" "+r.URL.Path)
defer tx.End()
```

When the context carries no valid span, `StartHttpWithContext` behaves exactly
like `StartHttp`: it starts a new trace.

### 3. Inbound and outbound gRPC

Nothing to do: the `lmgrpc` interceptors do both halves for you.

```go
server := grpc.NewServer(
    grpc.UnaryInterceptor(lmgrpc.UnaryServerInterceptor(app)),
    grpc.StreamInterceptor(lmgrpc.StreamServerInterceptor(app)),
)

conn, err := grpc.NewClient(target,
    grpc.WithUnaryInterceptor(lmgrpc.UnaryClientInterceptor(app)),
    grpc.WithStreamInterceptor(lmgrpc.StreamClientInterceptor(app)),
)
```

The server interceptor extracts `traceparent` from the incoming metadata and
starts the transaction as a child of the caller's span. The client interceptor
injects `traceparent` into the outgoing metadata, so the callee continues the
same trace. The existing `X-Trace-Id` header keeps being sent as before, so
application-level correlation is unaffected.

### 4. Propagating through your own code

`Transaction.ToContext` embeds both the transaction and its OpenTelemetry span,
so the usual context-based flow works:

```go
req = req.WithContext(tx.ToContext(ctx))
// ... and downstream:
propagator := otel.GetTextMapPropagator()
propagator.Inject(outCtx, propagation.HeaderCarrier(outReq.Header))
```

## Verifying

1. Send a request carrying a `traceparent` header (or call the service through
   another instrumented service).
2. Query the backend for the trace ID you sent: the service's spans should
   appear under it, with the inbound span as the parent of the transaction's
   root span.
3. Search for the `logmanager.trace_id` attribute to confirm the application
   level ID is attached to the same span.

## Compatibility

All changes are additive. `StartHttp`, `StartConsumer` and `Start` keep their
signatures and behaviour and simply delegate with `context.Background()`, so
existing callers continue to start a new trace until they opt in by passing a
context.
