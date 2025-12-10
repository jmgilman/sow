# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `libs/project` module: Consolidated project SDK with Backend interface abstraction
- `libs/project/state`: Project state management with pluggable backends
- `MemoryBackend` for improved testability
- Context-based API with cancellation support

### Changed

- Project state operations now use `Backend` interface instead of `sow.Context`
- Moved project SDK from `cli/internal/sdks/` to `libs/project/`

### Removed

- `cli/internal/sdks/state/` package (consolidated into libs/project)
- `cli/internal/sdks/project/` package (consolidated into libs/project)

## [0.1.1] - 2025-10-20

### Added

- Slash command for installing `sow` CLI

### Changed

- The CLI `start` command now calls `/sow:greet` instead of `/sow-greet`

## [0.1.0] - 2025-10-20

### Added

- Initial release

[unreleased]: https://github.com/jmgilman/sow/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/jmgilman/sow/releases/tag/v0.1.1
[0.1.0]: https://github.com/jmgilman/sow/releases/tag/v0.1.0