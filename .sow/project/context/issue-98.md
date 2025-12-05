# Issue #98: User Configuration System

**URL**: https://github.com/jmgilman/sow/issues/98
**State**: OPEN

## Description

# Work Unit 004: User Configuration System

## Behavioral Goal

As a sow user, I need to configure which AI CLI tools handle which agent roles, so that I can use my preferred tools (Claude Code for orchestration, Cursor for implementation) based on my subscriptions and preferences.

## Scope

### In Scope
- CUE schema for user configuration
- Configuration loading from `~/.config/sow/config.yaml`
- Cross-platform path resolution (`os.UserConfigDir()`)
- Default configuration when file doesn't exist (zero-config)
- Merging user config with defaults
- Environment variable overrides (`SOW_AGENTS_*`)
- Configuration validation against CUE schema
- Unit tests

### Out of Scope
- CLI commands for config management (work unit 005)
- Actually using config to select executors (integration with work unit 003)

## Existing Code Context

**Repository config loading** (`cli/internal/sow/config.go:18-43`):
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
Follow similar pattern for user config, but from `~/.config/sow/`.

**CUE schema pattern** (`cli/schemas/config.cue`):
```cue
#Config: {
    artifacts?: {
        adrs?: string
        design_docs?: string
    }
}
```

**CUE code generation** (`cli/schemas/project/cue_types_gen.go`):
Generated Go types from CUE schemas.

## Documentation Context

**Design doc** (`.sow/knowledge/designs/multi-agent-architecture.md`):
- Section "Configuration Location" (lines 486-491) specifies `~/.config/sow/config.yaml`
- Section "Configuration Structure" (lines 493-528) shows YAML structure
- Section "Configuration Priority" (lines 546-551) defines override order
- Section "Configuration File Schema (CUE)" (lines 780-813) shows schema
- Section "Cross-Platform Configuration Paths" (lines 759-778) shows path resolution

## File Structure

```
cli/schemas/
├── user_config.cue           # CUE schema for user config

cli/internal/sow/
├── user_config.go            # Loading, merging, validation
├── user_config_test.go       # Unit tests
```

## Implementation Approach

### CUE Schema

```cue
// cli/schemas/user_config.cue
package schemas

#UserConfig: {
    agents?: {
        executors?: [string]: {
            type: "claude" | "cursor" | "windsurf"
            settings?: {
                yolo_mode?: bool
                model?: string  // Only meaningful for claude
            }
            custom_args?: [...string]
        }

        bindings?: {
            orchestrator?: string
            implementer?: string
            architect?: string
            reviewer?: string
            planner?: string
            researcher?: string
            decomposer?: string
        }
    }
}
```

### Go Types

```go
// cli/internal/sow/user_config.go

type UserConfig struct {
    Agents AgentsConfig `yaml:"agents"`
}

type AgentsConfig struct {
    Executors map[string]ExecutorConfig `yaml:"executors"`
    Bindings  map[string]string         `yaml:"bindings"`
}

type ExecutorConfig struct {
    Type       string           `yaml:"type"`
    Settings   ExecutorSettings `yaml:"settings"`
    CustomArgs []string         `yaml:"custom_args"`
}

type ExecutorSettings struct {
    YoloMode bool   `yaml:"yolo_mode"`
    Model    string `yaml:"model"`
}
```

### Configuration Loading

```go
func GetUserConfigPath() (string, error) {
    configDir, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(configDir, "sow", "config.yaml"), nil
}

func LoadUserConfig() (*UserConfig, error) {
    path, err := GetUserConfigPath()
    if err != nil {
        return getDefaultUserConfig(), nil
    }

    data, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        return getDefaultUserConfig(), nil  // Zero-config experience
    }
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    var config UserConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    // Validate against CUE schema
    if err := validateUserConfig(&config); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    // Merge with defaults
    applyUserConfigDefaults(&config)

    // Apply environment overrides
    applyEnvOverrides(&config)

    return &config, nil
}
```

### Default Configuration

```go
func getDefaultUserConfig() *UserConfig {
    return &UserConfig{
        Agents: AgentsConfig{
            Executors: map[string]ExecutorConfig{
                "claude-code": {
                    Type: "claude",
                    Settings: ExecutorSettings{
                        YoloMode: false,
                    },
                },
            },
            Bindings: map[string]string{
                "orchestrator": "claude-code",
                "implementer":  "claude-code",
                "architect":    "claude-code",
                "reviewer":     "claude-code",
                "planner":      "claude-code",
                "researcher":   "claude-code",
                "decomposer":   "claude-code",
            },
        },
    }
}
```

### Environment Overrides

```go
func applyEnvOverrides(config *UserConfig) {
    // SOW_AGENTS_IMPLEMENTER=cursor overrides binding
    agentEnvs := map[string]string{
        "SOW_AGENTS_ORCHESTRATOR": "orchestrator",
        "SOW_AGENTS_IMPLEMENTER":  "implementer",
        "SOW_AGENTS_ARCHITECT":    "architect",
        "SOW_AGENTS_REVIEWER":     "reviewer",
        "SOW_AGENTS_PLANNER":      "planner",
        "SOW_AGENTS_RESEARCHER":   "researcher",
        "SOW_AGENTS_DECOMPOSER":   "decomposer",
    }

    for envVar, agentName := range agentEnvs {
        if value := os.Getenv(envVar); value != "" {
            config.Agents.Bindings[agentName] = value
        }
    }
}
```

### Validation

```go
func validateUserConfig(config *UserConfig) error {
    // Validate executor types
    validTypes := map[string]bool{"claude": true, "cursor": true, "windsurf": true}
    for name, exec := range config.Agents.Executors {
        if !validTypes[exec.Type] {
            return fmt.Errorf("unknown executor type %q for %q", exec.Type, name)
        }
    }

    // Validate bindings reference defined executors
    for agent, executor := range config.Agents.Bindings {
        if _, ok := config.Agents.Executors[executor]; !ok {
            // Check if it's a default executor
            if executor != "claude-code" {
                return fmt.Errorf("binding %q references undefined executor %q", agent, executor)
            }
        }
    }

    return nil
}
```

## Dependencies

None - this work unit is independent and can be developed in parallel with work units 001-003.

## Acceptance Criteria

1. **CUE schema** defined for user config
2. **Config path resolution** works cross-platform:
   - Linux/Mac: `~/.config/sow/config.yaml`
   - Windows: `%APPDATA%\sow\config.yaml`
3. **Zero-config experience**: Missing file returns defaults silently
4. **Config loading** parses YAML correctly
5. **Validation** catches:
   - Unknown executor types
   - Bindings referencing undefined executors
   - Invalid YAML syntax
6. **Default merging** fills missing values
7. **Environment overrides** work:
   - `SOW_AGENTS_IMPLEMENTER=cursor` overrides binding
8. **Priority order** enforced: env vars > file > defaults
9. **Unit tests** cover:
   - Path resolution on different platforms
   - Loading valid config
   - Loading with missing file (defaults)
   - Validation errors
   - Default merging
   - Environment overrides
   - Priority order

## Testing Strategy

- Unit tests with temporary config files
- Test missing file returns defaults (no error)
- Test invalid YAML returns error
- Test validation catches schema violations
- Test environment variable overrides
- Mock `os.UserConfigDir()` for cross-platform testing
