# Implement Repository Configuration Loading

## Context

This task implements the core repository configuration loading functionality for the `libs/config` module. Repository configuration is stored in `.sow/config.yaml` within a repository and controls artifact paths (ADRs, design docs) and other repo-level settings.

The key change from the original implementation is **decoupling from `sow.Context`**. The original `LoadConfig(ctx *Context)` function accepted the CLI's Context type. The new API accepts explicit dependencies:
- `LoadRepoConfig(fs core.FS)` - accepts a filesystem interface
- `LoadRepoConfigFromBytes(data []byte)` - accepts raw bytes (most flexible)

This enables the config package to be used outside the CLI context.

## Requirements

Create `libs/config/repo.go` with the following public API:

### LoadRepoConfig

```go
// LoadRepoConfig loads the repository configuration from .sow/config.yaml.
// It accepts a core.FS filesystem rooted at the .sow directory.
// Returns the config with defaults applied for any unspecified values.
// If config.yaml doesn't exist, returns default configuration (not an error).
func LoadRepoConfig(fs core.FS) (*schemas.Config, error)
```

### LoadRepoConfigFromBytes

```go
// LoadRepoConfigFromBytes parses repository configuration from raw YAML bytes.
// Returns the config with defaults applied for any unspecified values.
// This is the most flexible API for testing and non-filesystem use cases.
func LoadRepoConfigFromBytes(data []byte) (*schemas.Config, error)
```

### Internal Functions

- `getDefaultConfig() *schemas.Config` - returns config with all defaults
- `applyDefaults(config *schemas.Config)` - fills in missing values with defaults

### Error Handling

- Wrap all errors with context: `fmt.Errorf("load repo config: %w", err)`
- Use `ErrInvalidYAML` for YAML parse errors
- Use `ErrInvalidConfig` for validation errors
- Missing config file is NOT an error - return defaults

## Acceptance Criteria

1. [ ] `repo.go` implements `LoadRepoConfig(fs core.FS)`
2. [ ] `repo.go` implements `LoadRepoConfigFromBytes(data []byte)`
3. [ ] Missing config file returns default config (not error)
4. [ ] Invalid YAML returns wrapped `ErrInvalidYAML`
5. [ ] All unspecified fields get defaults applied
6. [ ] Functions under 80 lines per STYLE.md
7. [ ] All errors wrapped with context using `%w`
8. [ ] All tests pass with proper behavioral coverage
9. [ ] `golangci-lint run` passes

### Test Requirements (TDD - write tests first)

Create `libs/config/repo_test.go` with table-driven tests covering:

1. **LoadRepoConfig behaviors**:
   - Config file exists with all fields -> returns parsed config
   - Config file exists with partial fields -> applies defaults for missing
   - Config file doesn't exist -> returns default config
   - Config file has invalid YAML -> returns error wrapping ErrInvalidYAML
   - Empty config file -> returns default config

2. **LoadRepoConfigFromBytes behaviors**:
   - Valid complete YAML -> returns parsed config
   - Valid partial YAML -> applies defaults
   - Empty bytes -> returns default config
   - Invalid YAML bytes -> returns error wrapping ErrInvalidYAML
   - nil bytes -> returns default config

3. **Default values**:
   - Default ADRs path is "adrs"
   - Default design docs path is "design"

Use `testify/assert` and `testify/require` per TESTING.md standards.

## Technical Details

### Import Dependencies

```go
import (
    "fmt"

    "github.com/jmgilman/go/fs/core"
    "github.com/jmgilman/sow/libs/schemas"
    "gopkg.in/yaml.v3"
)
```

### The core.FS Interface

The `core.FS` interface from `github.com/jmgilman/go/fs/core` provides filesystem abstraction. The key method used is:

```go
type FS interface {
    ReadFile(name string) ([]byte, error)
    // ... other methods
}
```

The filesystem passed to `LoadRepoConfig` should be rooted at the `.sow/` directory, so reading `config.yaml` reads `.sow/config.yaml`.

### Schema Types

The `schemas.Config` type from `libs/schemas` looks like:

```go
type Config struct {
    Artifacts *struct {
        Adrs        *string `json:"adrs,omitempty"`
        Design_docs *string `json:"design_docs,omitempty"`
    } `json:"artifacts,omitempty"`
}
```

Fields use pointers to distinguish "not set" from "empty string".

### Test Fixtures

Create test fixtures in `testdata/`:
- `testdata/valid_config.yaml` - complete valid config
- `testdata/partial_config.yaml` - config with only some fields
- `testdata/invalid_config.yaml` - malformed YAML

### Testing with core.FS

For testing, use `os.DirFS` or create a simple mock filesystem:

```go
// Example using testdata directory
fs := os.DirFS("testdata")

// Or use afero for in-memory testing
memfs := afero.NewMemMapFs()
afero.WriteFile(memfs, "config.yaml", []byte("..."), 0644)
```

## Relevant Inputs

- `cli/internal/sow/config.go` - Original implementation to refactor
- `libs/schemas/config.cue` - Schema definition for Config type
- `libs/exec/local_test.go` - Example test patterns
- `.standards/TESTING.md` - Testing requirements
- `.standards/STYLE.md` - Code style requirements
- `.sow/knowledge/explorations/libs-refactoring-summary.md` - Architecture decisions

## Examples

### Usage Example

```go
// Using filesystem
sowFS, err := sow.NewFS(repoRoot)
if err != nil {
    return err
}
cfg, err := config.LoadRepoConfig(sowFS)
if err != nil {
    return fmt.Errorf("load config: %w", err)
}

// Using raw bytes (e.g., from remote source)
data, err := fetchConfigFromSomewhere()
if err != nil {
    return err
}
cfg, err := config.LoadRepoConfigFromBytes(data)
```

### Test Example

```go
func TestLoadRepoConfigFromBytes(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        want    *schemas.Config
        wantErr error
    }{
        {
            name:  "valid complete config",
            input: []byte("artifacts:\n  adrs: custom-adrs\n  design_docs: docs"),
            want: &schemas.Config{
                Artifacts: &struct{...}{
                    Adrs: ptr("custom-adrs"),
                    Design_docs: ptr("docs"),
                },
            },
        },
        {
            name:  "empty bytes returns defaults",
            input: []byte{},
            want:  getDefaultConfig(),
        },
        {
            name:    "invalid yaml",
            input:   []byte("invalid: [yaml: without: closing"),
            wantErr: ErrInvalidYAML,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := LoadRepoConfigFromBytes(tt.input)
            if tt.wantErr != nil {
                require.Error(t, err)
                assert.True(t, errors.Is(err, tt.wantErr))
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Dependencies

- Task 010 (module structure) must be completed first

## Constraints

- Do NOT implement user config loading - that's a separate task
- Do NOT implement path helpers - that's a separate task
- Accept `core.FS` interface, not concrete filesystem type
- Functions must be under 80 lines
- Use early returns to reduce nesting
