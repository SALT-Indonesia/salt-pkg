# OpenTelemetry Integration Example

This example demonstrates how to use logmanager with OpenTelemetry trace export enabled.

## Prerequisites

1. **Start Jaeger** (to view traces):
```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest
```

2. **Install dependencies**:
```bash
go mod tidy
```

## Running the Example

```bash
go run main.go
```

The server will start on `:8080` and automatically export traces to Jaeger.

## Making Test Requests

```bash
# Get all users (with database query simulation)
curl http://localhost:8080/api/users

# Get a specific user
curl http://localhost:8080/api/users/123

# Create an order (with external API call simulation)
curl -X POST http://localhost:8080/api/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id": 123}'
```

## Viewing Traces

1. Open Jaeger UI: http://localhost:16686
2. Select "otel-example-service" from the service dropdown
3. Click "Search Traces" or "Find Traces"
4. Click on a trace to see the detailed span tree

## What You'll See

Each request creates a trace with multiple spans:

- **Root span**: HTTP request handler
- **Child spans**: Database queries, external API calls
- **Correlation**: Custom trace_id and OTel trace_id linked together

### Example Log Output

```json
{
  "trace_id": "custom-trace-id-123",
  "otel_trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "otel_span_id": "00f067aa0ba902b7",
  "name": "GET /api/users",
  "type": "http",
  "latency": 95,
  "service": "otel-example-service",
  "status": 200
}
```

## Trace ID Correlation

The example demonstrates trace ID linking:
- **Custom trace_id**: Used by logmanager internally
- **otel_trace_id**: OpenTelemetry's trace ID
- **otel_span_id**: OpenTelemetry's span ID

Both IDs are logged, allowing you to correlate logs with traces in Jaeger.

## Architecture

```
┌─────────────────────────────────────┐
│  HTTP Request (GET /api/users)      │
│  ├─ Root OTel Span                 │
│  ├─ Logmanager Transaction         │
│  └─ Child Spans:                   │
│      ├─ Database Query (SELECT)    │
│      └─ Response                    │
└─────────────────────────────────────┘
         │                   │
         ▼                   ▼
   ┌─────────┐        ┌─────────────┐
   │ logrus  │        │ OpenTelemetry│
   │  logs   │        │   (Jaeger)   │
   └─────────┘        └─────────────┘
```

## Clean Up

Stop Jaeger:
```bash
docker stop jaeger
docker rm jaeger
```
