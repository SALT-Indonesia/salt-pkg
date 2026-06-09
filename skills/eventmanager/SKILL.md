---
name: salt-eventmanager
description: RabbitMQ pub/sub for Go microservices with the SALT-Indonesia/salt-pkg eventmanager library — generic message handlers, dual-error dispatch (domain vs infrastructure), automatic reconnection, dedup, and retry with configurable limits. Use when publishing or subscribing to RabbitMQ events with this library.
---

# eventmanager — RabbitMQ pub/sub for Go

`eventmanager` wraps RabbitMQ publishing and subscribing behind a minimal
type-safe API. Messages are validated with struct tags before publishing,
deserialized with Go generics on the consumer side, and traced through
`logmanager` correlation IDs. A dual-error handler separates business failures
from infrastructure failures — domain errors are logged and acknowledged,
while infrastructure errors trigger automatic requeue up to a configurable
retry limit. Connection loss is handled with automatic reconnection.

## Install

```bash
go get github.com/SALT-Indonesia/salt-pkg/eventmanager
# pulls in logmanager — used internally for tracing and logging
```

```go
import "github.com/SALT-Indonesia/salt-pkg/eventmanager"
```

Set `RABBITMQ_URL` in your environment (e.g. `amqp://guest:guest@localhost:5672`).

## Publish events

Define a struct with `json` and `validate` tags. `Publish` validates the body
before sending, so invalid messages never reach the broker.

```go
type OrderCreated struct {
    OrderID string  `json:"order_id" validate:"required"`
    Amount  float64 `json:"amount" validate:"required"`
}

msg := &OrderCreated{OrderID: "abc-123", Amount: 49.99}

if err := eventmanager.Publish(context.Background(), "orders", msg); err != nil {
    log.Fatal(err)
}
```

The exchange is declared as `direct` and durable. The message is published as
persistent JSON with a correlation ID derived from `logmanager` trace context.

## Subscribe to events

Define a handler with the dual-error signature, then call `Subscribe` with a
type parameter matching your message struct. The call blocks until the context
is cancelled.

```go
type OrderCreated struct {
    OrderID string  `json:"order_id"`
    Amount  float64 `json:"amount"`
}

func handleOrder(m OrderCreated) (domainErr, infrastructureErr error) {
    fmt.Printf("processing order %s\n", m.OrderID)
    // business logic here
    return nil, nil
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    eventmanager.Subscribe[OrderCreated](
        ctx,
        "order-service", // queue name
        "orders",        // exchange name
        []eventmanager.Handler[OrderCreated]{handleOrder},
    )
}
```

Each message is deduplicated by correlation ID across all channels, so
concurrent consumers won't process the same message twice.

## Dual-error handling

The `Handler` signature returns two errors:

```go
type Handler[Message any] func(Message) (domainErr, infrastructureErr error)
```

- **Domain error** (`domainErr`): a business-logic failure (e.g. "order not
  found"). Logged and the message is acknowledged — no retry.
- **Infrastructure error** (`infrastructureErr`): a transient failure (e.g.
  database unavailable). Logged and the message is requeued after a delay,
  up to `RABBITMQ_MAX_RETRY` times (default 5).

```go
func handleOrder(m OrderCreated) (domainErr, infrastructureErr error) {
    order, err := db.FindOrder(m.OrderID)
    if err != nil {
        // database is down — infrastructure error, will retry
        return nil, err
    }
    if order == nil {
        // business rule violation — domain error, no retry
        return fmt.Errorf("order %s not found", m.OrderID), nil
    }
    return nil, nil
}
```

## Configuration

All tuning is done through environment variables:

| Variable | Default | Description |
|---|---|---|
| `RABBITMQ_URL` | *(required)* | AMQP connection string |
| `RABBITMQ_CHANNELS` | `1` (max 10) | Concurrent consumer goroutines |
| `RABBITMQ_PREFETCH_COUNT` | `10` | QoS prefetch per channel |
| `RABBITMQ_RECONNECT_DELAY` | `2s` | Wait before reconnect attempt |
| `RABBITMQ_MAX_RETRY` | `5` | Max requeue attempts for infrastructure errors |
| `RABBITMQ_RETRY_DELAY` | `1m` | Wait before requeuing a failed message |
| `RABBITMQ_REFRESH_DELAY` | `1h` | Interval to reset dedup and retry counters |

## More

See [`eventmanager/README.md`](../../eventmanager/README.md) for the full
reference, including struct validation tags and advanced usage patterns.
