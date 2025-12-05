# Task 020: User Configuration Loading Core

## Context

This task is part of work unit 004: User Configuration System for sow. The system needs to load user configuration from `~/.config/sow/config.yaml` with cross-platform path resolution, zero-config defaults, and proper error handling.

This task builds on the CUE schema from task 010 and implements the core configuration loading logic in Go.

## Requirements

### Create Configuration Loading File

Create `cli/internal/sow/user_config.go` with the following functionality:

### 1. Cross-Platform Path Resolution

```go
// GetUserConfigPath returns the path to the user configuration file.
// Uses os.UserConfigDir() for cross-platform compatibility:
// - Linux/Mac: ~/.config/sow/config.yaml
// - Windows: %APPDATA%\sow\config.yaml
func GetUserConfigPath() (string, error)
```

### 2. Configuration Loading

```go
// LoadUserConfig loads the user configuration from the standard location.
// Returns default configuration if file doesn't exist (zero-config experience).
// Returns error only for actual failures (parse errors, permission issues).
func LoadUserConfig() (*schemas.UserConfig, error)
```

The loading should:
1. Get config path via `GetUserConfigPath()`
2. If file doesn't exist: return defaults silently (zero-config)
3. If file exists: read, parse YAML, apply defaults for missing values
4. Return error for actual failures (bad YAML, permission denied)

### 3. Default Configuration

```go
// getDefaultUserConfig returns a UserConfig with all default values.
// Default: All agents use claude-code executor with safe settings.
func getDefaultUserConfig() *schemas.UserConfig
```

Default values:
- Single executor: "claude-code" with type "claude", yolo_mode false
- All bindings point to "claude-code"
- Agent roles: orchestrator, implementer, architect, reviewer, planner, researcher, decomposer

### 4. Default Merging

```go
// applyUserConfigDefaults fills in missing configuration values with defaults.
// This allows partial configuration - user only specifies what they want to change.
func applyUserConfigDefaults(config *schemas.UserConfig)
```

Merging rules:
- If `agents` is nil, set to default agents config
- If `executors` is nil or empty, add default "claude-code" executor
- If `bindings` is nil, set to default bindings
- For each binding field that is nil, set to "claude-code"

### Implementation Pattern

Follow the existing pattern in `cli/internal/sow/config.go`:

```go
func LoadConfig(ctx *Context) (*schemas.Config, error) {
    fs := ctx.FS()
    data, err := fs.ReadFile("config.yaml")
    if err != nil {
        return getDefaultConfig(), nil  // Return defaults if missing
    }
    var config schemas.Config
    yaml.Unmarshal(data, &config)
    applyDefaults(&config)
    return &config, nil
}
```

Key differences from repo config:
- User config reads from filesystem directly (`os.ReadFile`)
- User config path is from `os.UserConfigDir()`, not repo context
- User config is global, not per-repository

## Acceptance Criteria

- [ ] `cli/internal/sow/user_config.go` created
- [ ] `GetUserConfigPath()` returns correct path for current platform
- [ ] `LoadUserConfig()` returns defaults when file doesn't exist
- [ ] `LoadUserConfig()` parses valid YAML correctly
- [ ] `LoadUserConfig()` returns error for invalid YAML
- [ ] `applyUserConfigDefaults()` merges partial configs correctly
- [ ] Default config uses "claude-code" for all bindings
- [ ] Unit tests cover all scenarios (see Testing Strategy)

## Technical Details

### File Location

```
cli/internal/sow/user_config.go
cli/internal/sow/user_config_test.go
```

### Imports

```go
import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/jmgilman/sow/cli/schemas"
    "gopkg.in/yaml.v3"
)
```

### Generated Types from Task 010

The `schemas.UserConfig` type (generated from CUE) will have nested pointer types. Your code needs to handle nil checks appropriately.

Example structure:
```go
type UserConfig struct {
    Agents *struct {
        Executors map[string]ExecutorConfig
        Bindings *struct {
            Orchestrator *string
            // ... etc
        }
    }
}
```

### Testing Strategy (TDD)

Write tests FIRST, then implement. Create `cli/internal/sow/user_config_test.go`:

1. **TestGetUserConfigPath** - Verify path construction
2. **TestLoadUserConfig_MissingFile** - Returns defaults without error
3. **TestLoadUserConfig_ValidYAML** - Parses config correctly
4. **TestLoadUserConfig_InvalidYAML** - Returns parse error
5. **TestApplyUserConfigDefaults_NilAgents** - Sets all defaults
6. **TestApplyUserConfigDefaults_PartialBindings** - Fills missing bindings
7. **TestApplyUserConfigDefaults_ExistingExecutors** - Preserves user executors
8. **TestGetDefaultUserConfig** - All bindings present and point to claude-code

For testing:
- Use `t.TempDir()` for temporary config directories
- Mock path resolution by using temp directories
- Use `os.Setenv`/`os.Unsetenv` for environment manipulation in tests
- Test table-driven style following `cli/internal/sow/context_test.go` pattern

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/internal/sow/config.go` - Existing repo config loading pattern
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/internal/sow/context_test.go` - Test patterns with temp directories
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/schemas/cue_types_gen.go` - Generated type examples
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/.sow/project/context/issue-98.md` - Full requirements

## Examples

### Example: Loading missing config

```go
config, err := LoadUserConfig()
// err is nil
// config.Agents.Bindings.Orchestrator points to "claude-code"
```

### Example: Loading partial config

Config file:
```yaml
agents:
  bindings:
    implementer: "cursor"
```

Result:
```go
config, err := LoadUserConfig()
// config.Agents.Bindings.Implementer = "cursor"
// config.Agents.Bindings.Orchestrator = "claude-code" (default)
// config.Agents.Executors["claude-code"] exists with defaults
```

### Example: Invalid YAML

Config file:
```yaml
agents:
  bindings:
    implementer: [invalid: yaml
```

Result:
```go
config, err := LoadUserConfig()
// err contains "failed to parse config"
// config is nil
```

## Dependencies

- Task 010 must be completed (CUE schema and generated types)

## Constraints

- Do NOT read from repository `.sow/config.yaml` - that's repo config
- Do NOT require the config file to exist
- Do NOT create the config file automatically
- Return clear error messages that include the config path
- All agent bindings MUST default to "claude-code"
- Use `gopkg.in/yaml.v3` for YAML parsing (matches existing code)
