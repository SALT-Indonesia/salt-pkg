# Logmanager Examples

This directory contains comprehensive examples demonstrating the logmanager module features, organized by use case for easy discovery and learning.

## 📁 Directory Structure

```
examples/logmanager/
├── 01-basic/              # Simple CLI usage
├── 02-http-servers/       # HTTP framework integrations
│   ├── gin/              # Gin framework
│   ├── echo/             # Echo framework
│   └── gorilla/          # Gorilla Mux
├── 03-grpc/              # gRPC integration
├── 04-messaging/         # Message queue integrations
│   ├── rabbitmq/         # RabbitMQ consumer
│   └── kafka/            # Kafka producer/consumer
├── 05-masking/           # Data masking patterns
├── 06-http-methods/      # HTTP methods & content types
│   ├── methods/          # All HTTP methods (GET, POST, etc.)
│   ├── content-types/    # Content type variations
│   ├── query-headers/    # Query params & headers
│   └── advanced/         # File uploads, streaming
├── 07-resty-client/      # Resty HTTP client integration
├── 08-native-http-client/# Native Go HTTP client integration
└── shared/               # Shared utilities
    ├── models/           # Common data models
    └── config/           # Configuration helpers
```

## 🚀 Quick Start

### Prerequisites
- Go 1.23+
- For messaging examples: Docker (RabbitMQ/Kafka)

### Running Examples

```bash
# Basic CLI example
cd 01-basic && go run main.go

# HTTP servers (choose framework)
cd 02-http-servers/gin && go run main.go      # :8001
cd 02-http-servers/echo && go run main.go     # :8002
cd 02-http-servers/gorilla && go run main.go  # :8003

# gRPC server
cd 03-grpc && go run main.go                  # :50051

# Messaging (requires Docker)
cd 04-messaging/rabbitmq && go run main.go
cd 04-messaging/kafka && go run main.go

# Data masking
cd 05-masking && go run main.go

# HTTP methods and content types
cd 06-http-methods/methods && go run main.go       # :8080
cd 06-http-methods/content-types && go run main.go # :8081
cd 06-http-methods/query-headers && go run main.go # :8082
cd 06-http-methods/advanced && go run main.go      # :8083

# HTTP clients
cd 07-resty-client && go run main.go
cd 08-native-http-client && go run main.go
```

## 📚 Examples Overview

| Feature | Location | Description |
|---------|----------|-------------|
| **Basic Usage** | `01-basic/` | CLI application with HTTP requests |
| **Gin Framework** | `02-http-servers/gin/` | Gin web server with middleware |
| **Echo Framework** | `02-http-servers/echo/` | Echo server with database segments |
| **Gorilla Mux** | `02-http-servers/gorilla/` | Gorilla with data masking |
| **gRPC Server** | `03-grpc/` | gRPC service with interceptors |
| **RabbitMQ** | `04-messaging/rabbitmq/` | Message consumer with logging |
| **Kafka** | `04-messaging/kafka/` | Producer/consumer with tracing |
| **Data Masking** | `05-masking/` | Sensitive data masking patterns |
| **HTTP Methods** | `06-http-methods/methods/` | Complete REST API methods |
| **Content Types** | `06-http-methods/content-types/` | All content type variations |
| **Query & Headers** | `06-http-methods/query-headers/` | Parameters and headers |
| **Advanced HTTP** | `06-http-methods/advanced/` | File uploads, streaming |
| **Resty Client** | `07-resty-client/` | Resty HTTP client with logmanager |
| **Native HTTP Client** | `08-native-http-client/` | Go stdlib HTTP client integration |

## 🔧 Features Demonstrated

- **Transaction Tracking**: Automatic request/response logging
- **Trace ID Propagation**: Cross-service tracing support
- **Error Handling**: Structured error logging and notification
- **Performance Monitoring**: Request timing and segments
- **Data Masking**: Sensitive information protection
- **Framework Integration**: Middleware for popular frameworks

## 🏗️ Architecture Patterns

### Clean Architecture
Examples follow clean architecture principles:
- **Repository Pattern**: Data access abstraction
- **Service Layer**: Business logic separation
- **Handler Layer**: HTTP/gRPC endpoint handling
- **Shared Utilities**: Common functionality reuse

### Error Handling
Consistent error handling across examples:
- Transaction error notification
- Graceful error responses
- Structured error logging

### Configuration Management
Flexible configuration using:
- Environment variables
- Default values
- Functional options pattern

## 📖 Learning Path

1. **Start with Basic** (`01-basic/`) - Understand core concepts
2. **Choose HTTP Framework** (`02-http-servers/`) - Learn middleware integration
3. **Explore gRPC** (`03-grpc/`) - Service-to-service communication
4. **Try Messaging** (`04-messaging/`) - Async processing patterns
5. **Implement Masking** (`05-masking/`) - Data privacy protection
6. **Master HTTP** (`06-http-methods/`) - Complete HTTP protocol coverage
7. **HTTP Clients** (`07-resty-client/`, `08-native-http-client/`) - Client-side request logging

## 🐳 Docker Setup for Messaging

```bash
# RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# Kafka
docker run -d --name kafka -p 9092:9092 \
  -e KAFKA_ZOOKEEPER_CONNECT=localhost:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  confluentinc/cp-kafka:latest
```

## 📝 Best Practices

- **Transaction Management**: Always defer `txn.End()` after starting transactions
- **Context Propagation**: Use `txn.ToContext()` for request context
- **Error Notification**: Call `txn.NoticeError(err)` for error tracking
- **Segment Naming**: Use descriptive names for database/external segments
- **Masking Configuration**: Configure sensitive data masking upfront

## 🔗 Related Documentation

- [Logmanager Module](../../logmanager/) - Core module documentation
- [Integration Guides](../../logmanager/integrations/) - Framework integrations
- [Configuration Options](../../logmanager/config/) - Setup and configuration