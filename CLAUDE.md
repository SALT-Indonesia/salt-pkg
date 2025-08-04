# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Structure

This is a Go polyrepo containing multiple independent modules managed with Go workspaces:

- **clientmanager**: HTTP client library with authentication support (JWT, OAuth1/2, Basic, Digest)
- **httpmanager**: Lightweight HTTP server framework with type-safe request handling
- **logmanager**: Structured logging with trace ID tracking and framework integrations
- **eventmanager**: Event publishing/subscribing system
- **examples**: Sample applications demonstrating module usage

## Build and Test Commands

This repository uses Go's standard tooling. Each module can be built and tested independently:

```bash
# Build all modules
go build ./...

# Test all modules  
go test ./...

# Test specific module
go test ./clientmanager/...
go test ./httpmanager/...
go test ./logmanager/...

# Run tests with coverage
go test -cover ./...

# Build examples
go build ./examples/...
```

## Go Workspace Setup

The repository uses Go workspaces (`go.work`) with local module replacements:

```bash
# Sync workspace dependencies
go work sync

# Add new module to workspace
go work use ./new-module

# Download dependencies
go mod download
```

## Module Dependencies

- All modules require Go 1.23+
- logmanager is a core dependency used by clientmanager and httpmanager
- Examples demonstrate integration patterns between modules

## Code Architecture

### Client Manager
- **Purpose**: Type-safe HTTP client with authentication support
- **Authentication**: JWT, OAuth1/2, Basic, Digest, AWS Signature
- **Features**: Request validation, response handling, retry logic

### HTTP Manager  
- **Purpose**: Lightweight HTTP server framework
- **Features**: Type-safe handlers, CORS, static file serving, file uploads
- **Integration**: Uses logmanager for request logging

### Log Manager
- **Purpose**: Structured logging with trace tracking
- **Features**: Trace ID propagation, request/response logging, data masking
- **Integrations**: Gin, Echo, Gorilla Mux, gRPC, RabbitMQ, Resty

### Event Manager
- **Purpose**: Event publishing and subscribing
- **Features**: Event handling, message routing

## Development Patterns

- Each module follows clean architecture principles
- Type-safe request/response handling using Go generics
- Comprehensive error handling with custom error types
- Extensive use of a functional options pattern for configuration
- Integration tests in `examples/` directory demonstrate real-world usage

## Testing

- Unit tests use testify/assert and testify/mock
- Integration tests in the examples directory
- Test files follow `*_test.go` naming convention
- Mock implementations available in test directories
- **Do not use test tables if the input object/struct is large** - create individual test functions instead to improve readability and maintainability

## Commit Message Format

**Format Rules:**
- **type**: feat, fix, chore, docs, refactor, test, etc.
- **description**: Brief description of the change

**Examples:**
```
feat: short description

long description
```
