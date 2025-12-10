# feat(config): create libs/config module for standalone configuration loading

## Summary

This PR creates a new `libs/config` Go module that extracts configuration loading logic from `cli/internal/sow/` into a standalone, reusable module. The module provides configuration loading functions that are **decoupled from `sow.Context`** by accepting explicit `core.FS` filesystem dependencies.

The key architectural change is that config loading functions now accept a `core.FS` interface from `github.com/jmgilman/go/fs/core` instead of depending on the CLI's Context type. This enables the config package to be used in contexts beyond the CLI while maintaining full testability through the `billy.NewMemory()` in-memory filesystem.

## Changes

### New Module: `libs/config`

- **Module Structure** (`go.mod`, `doc.go`, `README.md`)
  - New Go module at `github.com/jmgilman/sow/libs/config`
  - Comprehensive package documentation with usage examples
  - README following project standards

- **Repository Configuration** (`repo.go`, `repo_test.go`)
  - `LoadRepoConfig(fsys core.FS)` - loads config from filesystem
  - `LoadRepoConfigFromBytes(data []byte)` - parses from raw bytes
  - 14 test cases with table-driven tests

- **User Configuration** (`user.go`, `user_test.go`)
  - `LoadUserConfig(fsys core.FS)` - loads from standard XDG location
  - `LoadUserConfigFromPath(fsys core.FS, path string)` - loads from custom path
  - `GetUserConfigPath()` - returns platform-specific config path
  - `ValidateUserConfig()` - validates executor types and bindings
  - Environment variable overrides (`SOW_AGENTS_*`)
  - 51 test cases covering all functionality

- **Path Helpers** (`paths.go`, `paths_test.go`)
  - `GetADRsPath(repoRoot, config)` - returns ADRs directory path
  - `GetDesignDocsPath(repoRoot, config)` - returns design docs path
  - `GetExplorationsPath(repoRoot)` - returns explorations path
  - `GetKnowledgePath(repoRoot)` - returns knowledge base path
  - 14 test cases

- **Supporting Files**
  - `defaults.go` - default configuration values
  - `errors.go` - sentinel errors (`ErrConfigNotFound`, `ErrInvalidConfig`, `ErrInvalidYAML`)

### CLI Consumer Updates

Updated 8 consumer files to use the new `libs/config` module:
- `cli/cmd/config/validate.go`
- `cli/cmd/config/show.go`
- `cli/cmd/config/reset.go`
- `cli/cmd/config/path.go`
- `cli/cmd/config/init.go`
- `cli/cmd/config/edit.go`
- `cli/cmd/agent/spawn.go`
- `cli/cmd/agent/resume.go`

Commands with `sow.Context` access use `ctx.FS()`, while pre-init commands create `billy.NewLocal()` directly.

### Cleanup

- Removed `cli/internal/sow/config.go` (moved to libs/config)
- Removed `cli/internal/sow/user_config_test.go` (moved to libs/config)

## Testing

- **79 test cases** across all files
- **90.6% code coverage**
- All tests pass (`go test ./...`)
- Tests use `billy.NewMemory()` for in-memory filesystem (no custom mocks)
- `golangci-lint run` reports 0 issues for both modules

## Notes

- This is part of the libs/ modularization effort to create reusable Go modules
- The config module follows the same patterns as `libs/exec` and `libs/schemas`
- The `core.FS` interface provides full filesystem abstraction from `github.com/jmgilman/go/fs/core`
- No breaking changes to CLI user experience

Closes #116
