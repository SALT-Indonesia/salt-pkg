# Changelog

## [0.2.1] - 2026-06-18

### Changed
- Updated logmanager dependency from v1.38.1 to v1.43.1
  - Fixes a data race / `concurrent map writes` panic on `Transaction.txnRecords` under concurrent fanout
  - Rolls up `WithSkipHeaders()`, split-level log output, and the async nil-deref fix
