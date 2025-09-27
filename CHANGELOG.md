# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Module Changelogs

For detailed module-specific changes, see:
- [httpmanager CHANGELOG](./httpmanager/CHANGELOG.md)
- [logmanager CHANGELOG](./logmanager/CHANGELOG.md)

## [Unreleased]

### Fixed
- **[Issue #13]** Fixed multipart/form-data request logging not capturing form fields and file metadata
  - See [httpmanager CHANGELOG](./httpmanager/CHANGELOG.md) for details on the core fix in `upload.go`
  - See [logmanager CHANGELOG](./logmanager/CHANGELOG.md) for details on comprehensive test coverage across framework integrations

### Summary
- **httpmanager**: Fixed `UploadHandler` to properly log multipart/form-data request fields and file metadata
- **logmanager**: Added 7 comprehensive unit tests across lmgorilla, lmgin, and lmecho integrations
- **examples**: Added `/v1/event` endpoint demonstrating multipart/form-data handling
- **Total**: 36 tests passing across all modules and integrations