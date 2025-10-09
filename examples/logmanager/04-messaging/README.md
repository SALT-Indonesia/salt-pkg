# Messaging Examples

Message queue integrations demonstrating logmanager usage with RabbitMQ and Apache Kafka for asynchronous processing.

## Message Brokers

| Broker | Features | Use Case |
|--------|----------|----------|
| **RabbitMQ** | Simple consumer | Basic queue processing |
| **Kafka** | Producer/Consumer | High-throughput streaming |

## Prerequisites

### RabbitMQ Setup
```bash
docker run -d --name rabbitmq \
  -p 5672:5672 -p 15672:15672 \
  rabbitmq:3-management

# Access management UI: http://localhost:15672
# Default credentials: guest/guest
```

### Kafka Setup
```bash
# Start Zookeeper
docker run -d --name zookeeper \
  -p 2181:2181 \
  confluentinc/cp-zookeeper:latest

# Start Kafka
docker run -d --name kafka \
  -p 9092:9092 \
  -e KAFKA_ZOOKEEPER_CONNECT=localhost:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  confluentinc/cp-kafka:latest
```

## Examples Overview

### RabbitMQ Consumer (`rabbitmq/`)

Simple message consumer with:
- Queue declaration and setup
- Message deserialization
- Transaction tracking per message
- Error handling and acknowledgment

Key features:
```go
func (h *MessageHandler) processMessage(delivery amqp091.Delivery) {
    txn := h.app.Start(
        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
        "rabbitmq-consumer",
        logmanager.TxnTypeConsumer,
    )
    defer txn.End()

    // Process message...
}
```

### Kafka Producer/Consumer (`kafka/`)

Complete Kafka integration with:
- Producer with message publishing
- Consumer group with parallel processing
- Trace ID propagation via headers
- Graceful shutdown handling

Key features:
```go
// Producer trace header
kafkaMsg := &sarama.ProducerMessage{
    Headers: []sarama.RecordHeader{
        {
            Key:   []byte("trace-id"),
            Value: []byte(txn.TraceID()),
        },
    },
}

// Consumer transaction
txn := h.app.Start(txnName, "kafka-consumer", logmanager.TxnTypeConsumer)
```

## Running Examples

```bash
# RabbitMQ consumer
cd rabbitmq && go run main.go

# Kafka producer/consumer
cd kafka && go run main.go
```

## Message Flow

### RabbitMQ Flow
1. Consumer connects to queue
2. Messages are received and processed
3. Each message gets its own transaction
4. Acknowledgment sent after processing

### Kafka Flow
1. Producer sends messages every 5 seconds
2. Consumer group processes messages in parallel
3. Trace IDs propagated from producer to consumer
4. Graceful shutdown on SIGTERM/SIGINT

## Key Concepts

### Transaction Types
```go
logmanager.TxnTypeConsumer  // For message consumers
logmanager.TxnTypeOther     // For producers/other operations
```

### Error Handling
```go
// RabbitMQ
if err := json.Unmarshal(delivery.Body, &msg); err != nil {
    txn.NoticeError(err)
    return
}

// Kafka
if err := json.Unmarshal(message.Value, &msg); err != nil {
    txn.NoticeError(err)
    session.MarkMessage(message, "")
    return
}
```

### Trace Propagation
```go
// Producer: Add trace ID to headers
Headers: []sarama.RecordHeader{
    {
        Key:   []byte("trace-id"),
        Value: []byte(txn.TraceID()),
    },
}

// Consumer: Extract and use trace ID
var traceID string
for _, header := range message.Headers {
    if string(header.Key) == "trace-id" {
        traceID = string(header.Value)
        break
    }
}
```

## Best Practices

1. **Transaction Per Message**: Create separate transaction for each message
2. **Error Notification**: Always call `txn.NoticeError()` for failures
3. **Graceful Shutdown**: Implement proper shutdown handling
4. **Trace Propagation**: Include trace IDs in message headers
5. **Consumer Groups**: Use consumer groups for scalability
6. **Acknowledgment**: Properly acknowledge processed messages

## Testing

### RabbitMQ
```bash
# Publish test message via management UI
# Or use rabbitmqadmin tool
```

### Kafka
```bash
# The example includes both producer and consumer
# Messages are automatically generated every 5 seconds
```

## Monitoring

Both examples log:
- Message processing time
- Success/failure rates
- Trace ID correlation
- Error details
- Throughput metrics

## Next Steps

- [Data Masking](../05-masking/) - Protect sensitive message data
- [Basic Usage](../01-basic/) - Understand core concepts