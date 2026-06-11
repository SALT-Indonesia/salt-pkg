# Changelog

## [0.7.1] - 2026-06-11

### Added
- `WithResponseHeaderTimeout(timeout time.Duration) Option` to set the HTTP transport `ResponseHeaderTimeout`.

### Fixed
- `WithTimeout` now also raises `Transport.ResponseHeaderTimeout` when the requested timeout exceeds the current value.
  Previously `ResponseHeaderTimeout` was hardcoded at 5s in the default transport, causing requests to slow upstreams
  (LLM, report generation, long DB queries) to fail with `timeout awaiting response headers` regardless of `WithTimeout` value.

## [0.5.2] - 2026-02-12

### Changed
- Updated logmanager dependency from v1.38.1 to v1.41.0

## [0.5.1]

### Previous Changes
- See git history for previous changes
