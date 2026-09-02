# Changelog

## [0.2.3] - 2026-09-02

### Changed
- Updated logmanager dependency from v1.44.0 to v1.44.1
  - Fixes HTTP redirects (301/302/303/308) being misclassified as `internal server error`

## [0.2.1] - 2026-06-18

### Changed
- Updated logmanager dependency from v1.38.1 to v1.43.1
  - Fixes a data race / `concurrent map writes` panic on `Transaction.txnRecords` under concurrent fanout
  - Rolls up `WithSkipHeaders()`, split-level log output, and the async nil-deref fix
