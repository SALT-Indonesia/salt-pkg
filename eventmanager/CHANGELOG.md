# Changelog

## [0.2.5] - 2026-09-24

### Security
- Bumped vulnerable dependencies to resolve `govulncheck` findings without changing the Go version (`go` directive unchanged): `golang.org/x/net` v0.58.0, `golang.org/x/crypto` v0.55.0, `golang.org/x/sys` v0.47.0, `golang.org/x/text` v0.41.0, `google.golang.org/grpc` v1.84.0, `go.opentelemetry.io/otel*` v1.46.0, `github.com/rabbitmq/amqp091-go` v1.15.0

### Changed
- Updated logmanager dependency from v1.46.1 to v1.46.2 (dependency security fixes)

## [0.2.4] - 2026-09-14

### Changed
- Updated logmanager dependency from v1.44.1 to v1.45.0
  - Adds OTLP HTTP/Protobuf transport, TLS CA certificate, and `OTEL_*` environment variable support (#82)

## [0.2.3] - 2026-09-02

### Changed
- Updated logmanager dependency from v1.44.0 to v1.44.1
  - Fixes HTTP redirects (301/302/303/308) being misclassified as `internal server error`

## [0.2.1] - 2026-06-18

### Changed
- Updated logmanager dependency from v1.38.1 to v1.43.1
  - Fixes a data race / `concurrent map writes` panic on `Transaction.txnRecords` under concurrent fanout
  - Rolls up `WithSkipHeaders()`, split-level log output, and the async nil-deref fix
