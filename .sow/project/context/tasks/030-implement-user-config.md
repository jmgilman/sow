# Implement User Configuration Loading

## Context

This task implements user configuration loading for the `libs/config` module. User configuration is stored in a platform-specific location (`~/.config/sow/config.yaml` on Unix, `%APPDATA%\sow\config.yaml` on Windows) and controls agent executor bindings, executor settings, and other user-level preferences.

The good news is the existing implementation is **already mostly decoupled** from `sow.Context` - it uses `os.ReadFile` directly. This task involves moving the code to the new module and cleaning it up to follow project standards.

## Requirements

Create `libs/config/user.go` with the following public API:

### GetUserConfigPath

```go
// GetUserConfigPath returns the path to the user configuration file.
// Uses XDG-style paths:
//   - Linux/Mac: ~/.config/sow/config.yaml (or $XDG_CONFIG_HOME/sow/config.yaml)
//   - Windows: %APPDATA%\sow\config.yaml
func GetUserConfigPath() (string, error)
```

### LoadUserConfig

```go
// LoadUserConfig loads the user configuration from the standard location.
// Returns default configuration if file doesn't exist (zero-config experience).
// Returns error only for actual failures (parse errors, permission issues).
func LoadUserConfig() (*schemas.UserConfig, error)
```

### LoadUserConfigFromPath

```go
// LoadUserConfigFromPath loads user configuration from a specific path.
// This is useful for testing and for loading configs from non-standard locations.
func LoadUserConfigFromPath(path string) (*schemas.UserConfig, error)
```

### ValidateUserConfig

```go
// ValidateUserConfig validates the user configuration.
// Checks:
//   - Executor types are valid ("claude", "cursor", "windsurf")
//   - Bindings reference defined executors (or default "claude-code")
// Returns nil if valid, error with details if invalid.
func ValidateUserConfig(config *schemas.UserConfig) error
```

### Internal Functions

- `getDefaultUserConfig() *schemas.UserConfig` - returns config with all defaults
- `applyUserConfigDefaults(config *schemas.UserConfig)` - fills in missing values
- `applyEnvOverrides(config *schemas.UserConfig)` - applies environment variable overrides

### ValidExecutorTypes

```go
// ValidExecutorTypes defines the allowed executor types.
var ValidExecutorTypes = map[string]bool{
    "claude":   true,
    "cursor":   true,
    "windsurf": true,
}
```

## Acceptance Criteria

1. [ ] `user.go` implements all public functions listed above
2. [ ] `GetUserConfigPath()` returns correct platform-specific paths
3. [ ] Missing config file returns default config (not error)
4. [ ] Invalid YAML returns wrapped `ErrInvalidYAML`
5. [ ] Invalid executor type returns wrapped `ErrInvalidConfig`
6. [ ] Undefined executor binding returns wrapped `ErrInvalidConfig`
7. [ ] Environment overrides applied correctly (SOW_AGENTS_* vars)
8. [ ] Loading pipeline order: parse -> validate -> defaults -> env overrides
9. [ ] All tests pass with proper behavioral coverage
10. [ ] `golangci-lint run` passes

### Test Requirements (TDD - write tests first)

Create `libs/config/user_test.go` with table-driven tests covering:

1. **GetUserConfigPath behaviors**:
   - Returns ~/.config/sow/config.yaml on Unix when XDG_CONFIG_HOME not set
   - Uses XDG_CONFIG_HOME when set
   - Returns correct Windows path

2. **LoadUserConfig behaviors**:
   - Config file exists and is valid -> returns parsed config
   - Config file doesn't exist -> returns default config
   - Config file is empty -> returns default config with defaults applied

3. **LoadUserConfigFromPath behaviors**:
   - Valid config -> returns parsed config
   - Invalid YAML -> returns error
   - File doesn't exist -> returns default config
   - Empty file -> returns default config
   - Partial config -> applies defaults for missing

4. **ValidateUserConfig behaviors**:
   - Valid config -> returns nil
   - Unknown executor type -> returns error
   - Binding references undefined executor -> returns error
   - "claude-code" binding is always valid (implicit default)
   - nil config -> returns nil

5. **Environment overrides**:
   - SOW_AGENTS_IMPLEMENTER overrides binding
   - Multiple env vars can be set
   - Env vars take precedence over file config

6. **Default values**:
   - Default executor is "claude-code" with type "claude"
   - All role bindings default to "claude-code"
   - Default yolo_mode is false

## Technical Details

### Import Dependencies

```go
import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/jmgilman/sow/libs/schemas"
    "gopkg.in/yaml.v3"
)
```

### The UserConfig Schema

From `libs/schemas`, the `UserConfig` type has nested structures for agents:

```go
type UserConfig struct {
    Agents *struct {
        Executors map[string]struct {
            Type     string `json:"type"`
            Settings *struct {
                Yolo_mode *bool   `json:"yolo_mode,omitempty"`
                Model     *string `json:"model,omitempty"`
            } `json:"settings,omitempty"`
            Custom_args []string `json:"custom_args,omitempty"`
        } `json:"executors,omitempty"`
        Bindings *struct {
            Orchestrator *string `json:"orchestrator,omitempty"`
            Implementer  *string `json:"implementer,omitempty"`
            Architect    *string `json:"architect,omitempty"`
            Reviewer     *string `json:"reviewer,omitempty"`
            Planner      *string `json:"planner,omitempty"`
            Researcher   *string `json:"researcher,omitempty"`
            Decomposer   *string `json:"decomposer,omitempty"`
        } `json:"bindings,omitempty"`
    } `json:"agents,omitempty"`
}
```

### Loading Pipeline Order

The loading pipeline is critical and must follow this order:
1. Read and parse YAML
2. Validate (before applying defaults) - catch invalid executor types/bindings
3. Apply defaults for missing values
4. Apply environment overrides (highest priority)

### Environment Variable Format

Environment variables override agent bindings:
- `SOW_AGENTS_ORCHESTRATOR` - overrides orchestrator binding
- `SOW_AGENTS_IMPLEMENTER` - overrides implementer binding
- etc.

### Test Setup for Environment Variables

Use `t.Setenv()` for testing environment variables:

```go
func TestEnvOverrides(t *testing.T) {
    t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")
    // ... test logic
}
```

## Relevant Inputs

- `cli/internal/sow/user_config.go` - Original implementation to move
- `libs/schemas/user_config.cue` - Schema definition for UserConfig type
- `libs/exec/local_test.go` - Example test patterns
- `.standards/TESTING.md` - Testing requirements
- `.standards/STYLE.md` - Code style requirements

## Examples

### Usage Example

```go
// Load from default location
cfg, err := config.LoadUserConfig()
if err != nil {
    return fmt.Errorf("load user config: %w", err)
}

// Get specific binding
if cfg.Agents != nil && cfg.Agents.Bindings != nil {
    if cfg.Agents.Bindings.Implementer != nil {
        executor := *cfg.Agents.Bindings.Implementer
        fmt.Printf("Implementer uses: %s\n", executor)
    }
}

// Load from custom path (e.g., for testing)
cfg, err := config.LoadUserConfigFromPath("/custom/path/config.yaml")
```

### Test Example

```go
func TestLoadUserConfigFromPath(t *testing.T) {
    tests := []struct {
        name      string
        setup     func(t *testing.T) string // returns path
        want      *schemas.UserConfig
        wantErr   error
    }{
        {
            name: "valid config with custom executor",
            setup: func(t *testing.T) string {
                t.Helper()
                dir := t.TempDir()
                path := filepath.Join(dir, "config.yaml")
                content := `
agents:
  executors:
    my-cursor:
      type: cursor
  bindings:
    implementer: my-cursor
`
                require.NoError(t, os.WriteFile(path, []byte(content), 0644))
                return path
            },
            want: /* expected config */,
        },
        {
            name: "file not found returns defaults",
            setup: func(t *testing.T) string {
                return filepath.Join(t.TempDir(), "nonexistent.yaml")
            },
            want: getDefaultUserConfig(),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            path := tt.setup(t)
            got, err := LoadUserConfigFromPath(path)
            if tt.wantErr != nil {
                require.Error(t, err)
                assert.True(t, errors.Is(err, tt.wantErr))
                return
            }
            require.NoError(t, err)
            // Compare relevant fields
        })
    }
}
```

## Dependencies

- Task 010 (module structure) must be completed first
- Task 020 (repo config) should be completed first for consistency

## Constraints

- Do NOT implement path helpers - that's a separate task
- Must handle all nil pointer checks carefully (deeply nested struct)
- Environment variables take highest precedence
- Functions must be under 80 lines
- Use early returns to reduce nesting
- Platform-specific path logic must work on Linux, macOS, and Windows
