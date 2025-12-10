# Issue #116: libs/config module

**URL**: https://github.com/jmgilman/sow/issues/116
**State**: OPEN

## Description

# Work Unit 003: libs/config Module

Create a new `libs/config` Go module for loading and managing sow configuration (repo and user config).

## Objective

Move config loading logic from `cli/internal/sow/` to `libs/config/`, creating a standalone Go module that does NOT depend on `sow.Context`, with full standards compliance and proper documentation.

**This is NOT just a file move.** The key change is decoupling config loading from `sow.Context` - functions should accept explicit parameters (`core.FS`, file paths, or raw bytes) rather than the CLI-specific Context type.

## Scope

### What's Moving
- `config.go` - Repo config loading
- `user_config.go` - User config loading
- Path helper functions

### Target Structure
```
libs/config/
├── go.mod
├── go.sum
├── README.md                 # Per READMES.md standard
├── doc.go                    # Package documentation
│
├── repo.go                   # LoadRepoConfig(fs) / LoadRepoConfigFromBytes(data)
├── repo_test.go              # Tests for repo config
├── user.go                   # LoadUserConfig(), GetUserConfigPath()
├── user_test.go              # Tests for user config
├── paths.go                  # GetADRsPath(repoRoot, config), etc.
├── paths_test.go             # Tests for path helpers
└── defaults.go               # Default configuration values
```

## Standards Requirements

### Go Code Standards (STYLE.md)
- Accept interfaces, return concrete types
- Error handling with proper wrapping (`%w`)
- No global mutable state
- Parameter order: ctx first (if used), required before optional
- Functions under 80 lines

### Testing Standards (TESTING.md)
- Behavioral test coverage for all loading scenarios
- Table-driven tests with `t.Run()`
- Use `testify/assert` and `testify/require`
- Test fixtures in `testdata/` directory
- No external dependencies (use temp directories)

### README Standards (READMES.md)
- Overview: Configuration loading for sow repositories
- Quick Start: Load repo config, load user config
- Usage: Path helpers, defaults
- Configuration: Environment variables, config file locations

### Linting
- Must pass `golangci-lint run` with project's `.golangci.yml`
- Proper error wrapping for external errors

## API Design Requirements

### Decoupled from Context

**Before (coupled to Context):**
```go
func LoadConfig(ctx *Context) (*schemas.Config, error)
func GetADRsPath(ctx *Context, config *schemas.Config) string
```

**After (explicit dependencies):**
```go
// Option A: Accept filesystem
func LoadRepoConfig(fs core.FS) (*schemas.Config, error)

// Option B: Accept raw bytes (most flexible)
func LoadRepoConfigFromBytes(data []byte) (*schemas.Config, error)

// Path helpers take explicit root
func GetADRsPath(repoRoot string, config *schemas.Config) string
```

### User Config (Already Standalone)
```go
// Already doesn't use Context - just clean up and standardize
func LoadUserConfig() (*schemas.UserConfig, error)
func GetUserConfigPath() string
```

### Error Handling
- Wrap all errors with context: `fmt.Errorf("load repo config: %w", err)`
- Define sentinel errors for common cases: `ErrConfigNotFound`, `ErrInvalidConfig`

## Consumer Impact

~12 files currently use config loading. After this work:
- All imports change to `github.com/jmgilman/sow/libs/config`
- Callers pass `fs` or bytes instead of `Context`

## Dependencies

- `libs/schemas` - For `Config` and `UserConfig` types

## Acceptance Criteria

1. [ ] New `libs/config` Go module exists and compiles
2. [ ] Config loading decoupled from `sow.Context`
3. [ ] Both `LoadRepoConfig(fs)` and `LoadRepoConfigFromBytes(data)` available
4. [ ] Path helpers accept explicit `repoRoot` parameter
5. [ ] Proper sentinel errors defined
6. [ ] All tests pass with proper behavioral coverage
7. [ ] `golangci-lint run` passes with no issues
8. [ ] README.md follows READMES.md standard
9. [ ] Package documentation in doc.go
10. [ ] All 12 consumer files updated to new imports and signatures
11. [ ] Old config code removed from `cli/internal/sow/`

## Out of Scope

- Adding new configuration options
- Config file watching/reloading
- Remote configuration sources
