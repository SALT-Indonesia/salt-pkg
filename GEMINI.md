# Project Overview

This is a Go monorepo project containing several modules:

- `clientmanager`: Manages API client interactions.
- `eventmanager`: Handles event publishing and subscribing.
- `httpmanager`: Provides HTTP server and client functionalities.
- `logmanager`: Manages logging and tracing.
- `examples`: Contains example usage of the modules.

## Conventions

- Standard Go project structure.
- Each module has its own `go.mod` and `go.sum`.

## Testing

To run all tests in the project, navigate to the root directory and execute:

```bash
go test ./...
```

## Linting/Static Analysis

To run static analysis checks, navigate to the root directory and execute:

```bash
go vet ./...
```

## Commit Message Format

**Format Rules:**
- **type**: feat, fix, chore, docs, refactor, test, etc.
- **description**: Brief description of the change

**Examples:**
```
feat: short description

long description
```
