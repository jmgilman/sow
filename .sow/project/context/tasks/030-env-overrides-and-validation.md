# Task 030: Environment Overrides and Configuration Validation

## Context

This task is part of work unit 004: User Configuration System for sow. Building on the core configuration loading from task 020, this task adds:
1. Environment variable overrides for agent bindings
2. Configuration validation against schema constraints
3. Priority order enforcement (env vars > file > defaults)

## Requirements

### 1. Environment Variable Overrides

Add function to `cli/internal/sow/user_config.go`:

```go
// applyEnvOverrides applies environment variable overrides to the configuration.
// Environment variables take precedence over file configuration.
// Format: SOW_AGENTS_{ROLE}={executor_name}
// Example: SOW_AGENTS_IMPLEMENTER=cursor
func applyEnvOverrides(config *schemas.UserConfig)
```

Supported environment variables:
- `SOW_AGENTS_ORCHESTRATOR` - Override orchestrator binding
- `SOW_AGENTS_IMPLEMENTER` - Override implementer binding
- `SOW_AGENTS_ARCHITECT` - Override architect binding
- `SOW_AGENTS_REVIEWER` - Override reviewer binding
- `SOW_AGENTS_PLANNER` - Override planner binding
- `SOW_AGENTS_RESEARCHER` - Override researcher binding
- `SOW_AGENTS_DECOMPOSER` - Override decomposer binding

### 2. Configuration Validation

Add validation function to `cli/internal/sow/user_config.go`:

```go
// ValidateUserConfig validates the user configuration.
// Checks:
// - Executor types are valid ("claude", "cursor", "windsurf")
// - Bindings reference defined executors (or default "claude-code")
// Returns nil if valid, error with details if invalid.
func ValidateUserConfig(config *schemas.UserConfig) error
```

Validation rules:
1. **Executor type validation**: Each executor must have type "claude", "cursor", or "windsurf"
2. **Binding reference validation**: Each binding must reference:
   - An executor defined in the config's executors map, OR
   - The default "claude-code" executor

### 3. Update LoadUserConfig

Modify `LoadUserConfig()` to apply the full loading pipeline:

```go
func LoadUserConfig() (*schemas.UserConfig, error) {
    path, err := GetUserConfigPath()
    if err != nil {
        return getDefaultUserConfig(), nil
    }

    data, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        config := getDefaultUserConfig()
        applyEnvOverrides(config)  // Env vars still apply without file
        return config, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to read config at %s: %w", path, err)
    }

    var config schemas.UserConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config at %s: %w", path, err)
    }

    // Validate before applying defaults
    if err := ValidateUserConfig(&config); err != nil {
        return nil, fmt.Errorf("invalid config at %s: %w", path, err)
    }

    // Apply defaults for missing values
    applyUserConfigDefaults(&config)

    // Apply environment overrides (highest priority)
    applyEnvOverrides(&config)

    return &config, nil
}
```

### Priority Order

Configuration values are resolved in this order (highest priority first):
1. Environment variables (`SOW_AGENTS_*`)
2. User config file (`~/.config/sow/config.yaml`)
3. Built-in defaults

## Acceptance Criteria

- [ ] `applyEnvOverrides()` reads all SOW_AGENTS_* environment variables
- [ ] Environment variables override file configuration
- [ ] Environment variables work even when config file doesn't exist
- [ ] `ValidateUserConfig()` catches invalid executor types
- [ ] `ValidateUserConfig()` catches bindings to undefined executors
- [ ] `ValidateUserConfig()` allows binding to "claude-code" even if not explicitly defined
- [ ] Clear error messages indicate which executor/binding is invalid
- [ ] Unit tests cover all validation scenarios
- [ ] Unit tests cover environment override scenarios

## Technical Details

### Environment Variable Handling

```go
func applyEnvOverrides(config *schemas.UserConfig) {
    // Ensure bindings struct exists
    if config.Agents == nil {
        config.Agents = &AgentsConfig{}
    }
    if config.Agents.Bindings == nil {
        config.Agents.Bindings = &BindingsConfig{}
    }

    envMap := map[string]**string{
        "SOW_AGENTS_ORCHESTRATOR": &config.Agents.Bindings.Orchestrator,
        "SOW_AGENTS_IMPLEMENTER":  &config.Agents.Bindings.Implementer,
        "SOW_AGENTS_ARCHITECT":    &config.Agents.Bindings.Architect,
        "SOW_AGENTS_REVIEWER":     &config.Agents.Bindings.Reviewer,
        "SOW_AGENTS_PLANNER":      &config.Agents.Bindings.Planner,
        "SOW_AGENTS_RESEARCHER":   &config.Agents.Bindings.Researcher,
        "SOW_AGENTS_DECOMPOSER":   &config.Agents.Bindings.Decomposer,
    }

    for envVar, field := range envMap {
        if value := os.Getenv(envVar); value != "" {
            *field = &value
        }
    }
}
```

Note: The exact field access pattern depends on the generated types from task 010. Adjust as needed.

### Validation Error Format

Errors should be clear and actionable:

```go
// Invalid executor type
fmt.Errorf("unknown executor type %q for executor %q; must be one of: claude, cursor, windsurf", exec.Type, name)

// Binding references undefined executor
fmt.Errorf("binding %q references undefined executor %q", agentRole, executorName)
```

### Testing Strategy (TDD)

Add tests to `cli/internal/sow/user_config_test.go`:

**Environment Override Tests:**
1. `TestApplyEnvOverrides_SingleVar` - One env var overrides one binding
2. `TestApplyEnvOverrides_MultipleVars` - Multiple env vars work together
3. `TestApplyEnvOverrides_EmptyVar` - Empty env var is ignored
4. `TestApplyEnvOverrides_NilConfig` - Handles nil config.Agents gracefully
5. `TestLoadUserConfig_EnvOverridesFile` - Env vars take precedence over file
6. `TestLoadUserConfig_EnvOverridesNoFile` - Env vars work without config file

**Validation Tests:**
1. `TestValidateUserConfig_ValidConfig` - Valid config passes
2. `TestValidateUserConfig_InvalidExecutorType` - Catches bad type
3. `TestValidateUserConfig_BindingUndefinedExecutor` - Catches bad binding
4. `TestValidateUserConfig_BindingDefaultExecutor` - Allows "claude-code" even if not defined
5. `TestValidateUserConfig_EmptyConfig` - Empty config is valid
6. `TestLoadUserConfig_InvalidConfig` - Returns validation error

Use `t.Setenv()` for environment variable tests (automatically cleaned up).

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/internal/sow/user_config.go` - File from task 020
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/internal/sow/config.go` - Pattern for error handling
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/internal/agents/agents.go` - List of agent roles
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/.sow/project/context/issue-98.md` - Full requirements

## Examples

### Environment Override Example

```bash
# Override implementer to use cursor
export SOW_AGENTS_IMPLEMENTER=cursor

# In Go code:
config, _ := LoadUserConfig()
// config.Agents.Bindings.Implementer == "cursor"
// Even if file says "claude-code" or doesn't exist
```

### Validation Error Examples

Config file with invalid executor type:
```yaml
agents:
  executors:
    my-executor:
      type: "copilot"  # Invalid - not claude/cursor/windsurf
```

Error: `invalid config at ~/.config/sow/config.yaml: unknown executor type "copilot" for executor "my-executor"; must be one of: claude, cursor, windsurf`

Config file with invalid binding:
```yaml
agents:
  bindings:
    implementer: "nonexistent"  # No executor named "nonexistent"
```

Error: `invalid config at ~/.config/sow/config.yaml: binding "implementer" references undefined executor "nonexistent"`

### Valid Partial Config

```yaml
agents:
  executors:
    cursor:
      type: "cursor"
  bindings:
    implementer: "cursor"
    # Other bindings default to "claude-code" which is always valid
```

This is valid because:
- "cursor" executor is defined with valid type
- "implementer" binding references defined "cursor"
- Other bindings will default to "claude-code" (always valid)

## Dependencies

- Task 010 completed (CUE schema and generated types)
- Task 020 completed (core loading functions)

## Constraints

- Do NOT validate bindings added by env vars against defined executors
  - Env vars might reference executors that will be added dynamically
  - Validation happens before env overrides
- Do NOT fail on empty environment variables
- "claude-code" is ALWAYS a valid executor reference (implicit default)
- Error messages MUST include the config file path
- Follow existing error wrapping patterns in the codebase
