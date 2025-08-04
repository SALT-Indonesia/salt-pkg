# Salt Package (salt-pkg)

A collection of Go modules for building robust and efficient applications. This repository contains multiple Go modules that can be used independently or together to enhance your Go applications.

## Go Version Compatibility

All modules in this repository support Go versions 1.23.

## Modules

### Client Manager

The Client Manager module provides a robust HTTP client for making requests to external APIs. It simplifies the process of making HTTP requests with features like:

- Type-safe request and response handling
- Request validation
- Support for various formats (JSON, XML, form data, file uploads)
- Error handling

[Learn more about Client Manager](./clientmanager/README.md)

### HTTP Manager

The HTTP Manager module is a lightweight solution for quickly setting up HTTP servers with configurable options and type-safe request handling. Key features include:

- Type-safe request handling with automatic JSON serialization/deserialization
- Configurable server options (timeouts, SSL, etc.)
- Clean architecture support
- Appropriate HTTP status codes for errors

[Learn more about HTTP Manager](./httpmanager/README.md)

### Log Manager

The Log Manager module provides comprehensive structured logging with features like:

- Trace ID tracking
- Request/response logging
- Masking sensitive data
- Integrations with various frameworks (Gin, Gorilla Mux, gRPC, RabbitMQ, etc.)
- Segment-based transaction tracking

[Learn more about Log Manager](./logmanager/README.md)

## Examples

The repository includes example applications demonstrating the use of these modules. Check the `examples` directory for sample implementations.

## Installation

Each module can be installed independently using Go modules:

```bash
# Install Client Manager
go get github.com/SALT-Indonesia/salt-pkg/clientmanager

# Install HTTP Manager
go get github.com/SALT-Indonesia/salt-pkg/httpmanager

# Install Log Manager
go get github.com/SALT-Indonesia/salt-pkg/logmanager
```
