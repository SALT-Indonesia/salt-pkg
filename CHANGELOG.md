# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **httpmanager**: Health check endpoint support
  - Added configurable health check endpoint (default: `GET /health`)
  - Health check is enabled by default
  - Returns HTTP 200 with empty body
  - Configurable endpoint path via `WithHealthCheckPath(path string)` option
  - Can be disabled via `WithoutHealthCheck()` option
  - Only accepts GET requests, returns 405 Method Not Allowed for other methods

### Changed

### Deprecated

### Removed

### Fixed

### Security