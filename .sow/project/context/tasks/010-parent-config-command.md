# Task 010: Parent Config Command and Root Integration

## Context

The sow CLI needs a `config` command group for managing user configuration. This task creates the parent command (`sow config`) and integrates it into the root command.

The user configuration system already exists in `cli/internal/sow/user_config.go`, which provides:
- `GetUserConfigPath()` - returns platform-specific config path
- `LoadUserConfig()` - loads and merges config with defaults
- `ValidateUserConfig()` - validates config structure

This task creates the command structure that will serve as the container for all config subcommands (init, path, show, validate, edit, reset).

## Requirements

### 1. Create Parent Config Command

Create `cli/cmd/config/config.go` with:

```go
// Package config implements commands for managing sow user configuration.
package config

import (
    "github.com/spf13/cobra"
)

// NewConfigCmd creates the config command with all subcommands.
func NewConfigCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config",
        Short: "Manage user configuration",
        Long: `Manage sow user configuration for agent preferences.

Configuration is stored at ~/.config/sow/config.yaml (Linux/Mac)
or %APPDATA%\sow\config.yaml (Windows).

If no configuration exists, sow uses defaults (Claude Code for all agents).

Commands:
  init      Create configuration file with template
  path      Show configuration file path
  show      Display effective configuration (merged)
  validate  Validate configuration file
  edit      Open configuration in editor
  reset     Remove configuration file`,
    }

    // Subcommands will be added in subsequent tasks
    // cmd.AddCommand(newInitCmd())
    // cmd.AddCommand(newPathCmd())
    // etc.

    return cmd
}
```

### 2. Wire Into Root Command

Update `cli/cmd/root.go` to:
1. Add import for `"github.com/jmgilman/sow/cli/cmd/config"`
2. Add `cmd.AddCommand(config.NewConfigCmd())` in `NewRootCmd()`

### 3. Follow Existing Patterns

The command structure must match existing patterns in the codebase:
- Use `NewConfigCmd()` as the exported constructor (matching `NewRefsCmd()`, `NewProjectCmd()`)
- Use lowercase unexported `newXxxCmd()` for subcommands
- Return `*cobra.Command`

## Acceptance Criteria

1. **Parent command exists**: `sow config` displays help text
2. **Help text is informative**: Shows location info and available subcommands
3. **Root integration works**: `sow --help` shows `config` in command list
4. **No runtime errors**: Command does not panic or error when invoked alone
5. **Code follows patterns**: Matches existing command patterns in codebase

### Tests to Write

Create `cli/cmd/config/config_test.go`:

```go
func TestNewConfigCmd_Structure(t *testing.T) {
    cmd := NewConfigCmd()

    // Verify command properties
    if cmd.Use != "config" {
        t.Errorf("expected Use='config', got '%s'", cmd.Use)
    }

    if cmd.Short == "" {
        t.Error("expected non-empty Short description")
    }

    if cmd.Long == "" {
        t.Error("expected non-empty Long description")
    }
}

func TestNewConfigCmd_NoErrorWhenRun(t *testing.T) {
    cmd := NewConfigCmd()
    cmd.SetArgs([]string{})

    // Running without subcommand should show help (not error)
    err := cmd.Execute()
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

## Technical Details

### File Locations

- New: `cli/cmd/config/config.go`
- New: `cli/cmd/config/config_test.go`
- Modify: `cli/cmd/root.go`

### Import Path

The package import path will be:
```go
"github.com/jmgilman/sow/cli/cmd/config"
```

### Dependencies

- `github.com/spf13/cobra` - CLI framework

## Relevant Inputs

- `cli/cmd/root.go` - Where to add the config command; shows existing command registration pattern
- `cli/cmd/refs/refs.go` - Example of a parent command with subcommands; pattern to follow
- `cli/cmd/project/project.go` - Another example parent command pattern
- `cli/internal/sow/user_config.go` - Existing user config functions that subcommands will use

## Examples

### Expected Command Output

```bash
$ sow config
Manage sow user configuration for agent preferences.

Configuration is stored at ~/.config/sow/config.yaml (Linux/Mac)
or %APPDATA%\sow\config.yaml (Windows).

If no configuration exists, sow uses defaults (Claude Code for all agents).

Commands:
  init      Create configuration file with template
  path      Show configuration file path
  show      Display effective configuration (merged)
  validate  Validate configuration file
  edit      Open configuration in editor
  reset     Remove configuration file

Use "sow config [command] --help" for more information about a command.
```

## Constraints

- Do NOT implement any subcommands in this task (they are separate tasks)
- Keep the command structure minimal and focused
- Match the style and patterns of existing commands exactly
