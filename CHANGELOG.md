# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Go Orchestrator
- Dependency injection container with support for singletons, scoped, and transient lifetimes
- Lifecycle management with automatic startup/shutdown ordering
- Application orchestration with feature registration
- Health checking for all components
- Generic helpers for type-safe resolution
- DAG-based dependency resolution
- Retry configuration for component operations

### Changed
- Restructured project to follow Go standard project layout
- Moved implementation packages to `internal/` directory
- Moved public API to `pkg/` directory
- Added example application in `cmd/example/`

### Security
- No security issues reported
