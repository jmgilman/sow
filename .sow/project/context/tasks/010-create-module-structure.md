# Create libs/config Module Structure

## Context

This is the first task in creating a new `libs/config` Go module that extracts configuration loading logic from `cli/internal/sow/`. The `libs/` directory contains standalone Go modules that can be used independently of the CLI. Other modules in `libs/` include `libs/exec` and `libs/schemas`.

The goal is to create a properly structured Go module that follows project standards and patterns established by existing libs modules.

## Requirements

Create the following directory and file structure:

```
libs/config/
├── go.mod              # Module definition
├── go.sum              # Dependencies
├── doc.go              # Package documentation
├── README.md           # User-facing documentation
├── errors.go           # Sentinel errors
├── defaults.go         # Default configuration values
└── testdata/           # Test fixtures directory
```

### go.mod

Create a Go module with:
- Module path: `github.com/jmgilman/sow/libs/config`
- Go version: `1.25.3` (match existing libs/exec)
- Dependencies:
  - `github.com/jmgilman/sow/libs/schemas` - For Config and UserConfig types
  - `github.com/stretchr/testify v1.11.1` - For testing (testify/assert, testify/require)
  - `gopkg.in/yaml.v3` - For YAML parsing

### doc.go

Create package documentation following the pattern in `libs/exec/doc.go`:
- Package name: `config`
- Description: Configuration loading for sow repositories and user settings
- Document the two main configuration types: repo config and user config
- Provide usage examples showing the primary API
- List the main implementations/functions

### README.md

Follow the READMES.md standard:
1. **Overview**: 1-3 sentences about configuration loading for sow repositories
2. **Quick Start**: Install and basic usage
3. **Usage**: Show 2-4 common tasks:
   - Load repo config from filesystem
   - Load repo config from bytes
   - Load user config
   - Get path helpers (ADRs path, design docs path)
4. **Configuration**: Environment variables, config file locations
5. **Links**: Link to Go package documentation

### errors.go

Define sentinel errors for common failure cases:
```go
var (
    ErrConfigNotFound  = errors.New("config file not found")
    ErrInvalidConfig   = errors.New("invalid configuration")
    ErrInvalidYAML     = errors.New("invalid YAML syntax")
)
```

### defaults.go

Move the default constants from `cli/internal/sow/config.go`:
```go
const (
    DefaultADRsPath         = "adrs"
    DefaultDesignDocsPath   = "design"
    DefaultExplorationsPath = "explorations"
    DefaultExecutorName     = "claude-code"
)
```

Include the default config factory functions that will be used by loading functions.

## Acceptance Criteria

1. [ ] `libs/config/` directory exists with all required files
2. [ ] `go.mod` compiles successfully with `go mod tidy`
3. [ ] `doc.go` follows the pattern from `libs/exec/doc.go`
4. [ ] `README.md` follows READMES.md standard structure
5. [ ] `errors.go` defines sentinel errors with descriptive messages
6. [ ] `defaults.go` contains all default values moved from source
7. [ ] `testdata/` directory exists (empty for now)
8. [ ] `golangci-lint run` passes with no issues
9. [ ] Module can be imported by other packages

## Technical Details

### Dependencies

The module imports:
- `github.com/jmgilman/sow/libs/schemas` - defines `Config` and `UserConfig` types
- Standard library: `errors`, `fmt`

### File Organization (per STYLE.md)

Within each file, organize code:
1. Imports
2. Constants
3. Type declarations
4. Interfaces (none for this task)
5. Structs (none for this task)
6. Public functions
7. Private functions

### Error Design (per STYLE.md)

- Use sentinel errors at package level
- Error variable names start with `Err`
- Errors should be wrapped with context when used: `fmt.Errorf("load repo config: %w", ErrConfigNotFound)`

## Relevant Inputs

- `libs/exec/go.mod` - Example go.mod structure for libs modules
- `libs/exec/doc.go` - Example package documentation pattern
- `libs/exec/README.md` - Example README following standards
- `libs/schemas/go.mod` - Dependency version reference
- `cli/internal/sow/config.go` - Source of default constants
- `cli/internal/sow/user_config.go` - Source of DefaultExecutorName
- `.standards/STYLE.md` - Code style requirements
- `.standards/READMES.md` - README structure requirements

## Examples

### doc.go example (pattern from libs/exec):

```go
// Package config provides configuration loading for sow repositories and user settings.
//
// This package provides functions to load both repository-level configuration
// (from .sow/config.yaml) and user-level configuration (from ~/.config/sow/config.yaml).
// The loading functions are decoupled from the CLI's Context type, accepting explicit
// dependencies like filesystem interfaces or raw bytes.
//
// # Repository Configuration
//
// Repository configuration controls artifact paths and other repo-specific settings.
// Load it using either a filesystem interface or raw bytes:
//
//	// From filesystem
//	cfg, err := config.LoadRepoConfig(fs)
//
//	// From bytes (more flexible)
//	cfg, err := config.LoadRepoConfigFromBytes(data)
//
// # User Configuration
//
// User configuration controls agent executor bindings and settings:
//
//	cfg, err := config.LoadUserConfig()
//	path, err := config.GetUserConfigPath()
//
// # Path Helpers
//
// Get absolute paths to configuration directories:
//
//	adrsPath := config.GetADRsPath(repoRoot, cfg)
//	designPath := config.GetDesignDocsPath(repoRoot, cfg)
//
// See README.md for more examples.
package config
```

## Dependencies

None - this is the foundational task.

## Constraints

- Do NOT implement the actual loading logic yet - that's in subsequent tasks
- Do NOT add files beyond the required structure
- Use exact Go version 1.25.3 (project standard)
- Follow all linting rules in `.golangci.yml`
