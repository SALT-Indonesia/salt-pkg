# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## About This Changelog

This is a **monorepo** containing multiple independent Go modules. This root changelog provides a high-level overview of changes across all modules.

### Module-Specific Changelogs

For detailed changes in each module, see:

| Module | Changelog | Description |
|--------|-----------|-------------|
| **httpmanager** | [CHANGELOG](./httpmanager/CHANGELOG.md) | HTTP server framework with type-safe request handling |
| **logmanager** | [CHANGELOG](./logmanager/CHANGELOG.md) | Structured logging with trace ID tracking and framework integrations |
| **clientmanager** | [CHANGELOG](./clientmanager/CHANGELOG.md) | HTTP client library with authentication support |
| **eventmanager** | [CHANGELOG](./eventmanager/CHANGELOG.md) | Event publishing/subscribing system |

### Changelog Format

Each entry follows this format:
```
- **[module]**: Brief description (#PR)
  - Sub-item with additional context
```

---

## [Unreleased]

### Fixed

- **[httpmanager]**: Fixed multipart/form-data request logging not capturing form fields and file metadata (#13)
  - Added `SetWebRequest()` call after `ParseMultipartForm()` in `upload.go` to properly log parsed form data
  - Request logs now include all form fields and file metadata (`_files` array with field, filename, size, headers)
  - Affects: `UploadHandler.ServeHTTP()`
  - See [httpmanager/CHANGELOG.md](./httpmanager/CHANGELOG.md#unreleased) for details

### Added

- **[logmanager]**: Comprehensive unit tests for multipart/form-data logging across framework integrations (#13)
  - Added 7 new test cases across lmgorilla (3), lmgin (2), and lmecho (2) integrations
  - All tests validate form fields and file metadata are properly logged
  - Total: 36 tests passing across all framework integrations
  - See [logmanager/CHANGELOG.md](./logmanager/CHANGELOG.md#unreleased) for details

- **[examples/httpmanager]**: Added `/v1/event` endpoint demonstrating multipart/form-data handling (#13)
  - Shows proper form field extraction and file upload processing
  - Provides reference implementation for handling multipart requests

### Impact Summary

| Module | Changes | Tests Added | Breaking Changes |
|--------|---------|-------------|------------------|
| httpmanager | 1 bug fix | 0 | No |
| logmanager | 7 new tests | 7 | No |
| examples | 1 new endpoint | 0 | No |

**Total**: 1 bug fix, 7 new tests, 1 example endpoint, 0 breaking changes

---

## Previous Releases

See individual module changelogs for historical releases:
- [httpmanager releases](./httpmanager/CHANGELOG.md)
- [logmanager releases](./logmanager/CHANGELOG.md)
- [clientmanager releases](./clientmanager/CHANGELOG.md)
- [eventmanager releases](./eventmanager/CHANGELOG.md)