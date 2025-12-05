# Task 020: Config Init Command

## Context

The sow CLI needs a `config init` command that creates a configuration file with a well-documented template. This allows users to customize which AI executor handles each agent role.

The user configuration system exists in `cli/internal/sow/user_config.go`, which provides `GetUserConfigPath()` to determine where the config file should be located. The init command creates the file at this location with a comprehensive template.

## Requirements

### 1. Create Init Command

Create `cli/cmd/config/init.go` with:

```go
package config

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "init",
        Short: "Create configuration file with template",
        Long: `Create a configuration file with a documented template.

The configuration file includes:
- Executor definitions (Claude Code, Cursor, Windsurf)
- Agent role bindings
- All available settings with documentation

If the file already exists, use 'sow config edit' to modify it.`,
        RunE: runInit,
    }
}

func runInit(cmd *cobra.Command, _ []string) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return fmt.Errorf("failed to get config path: %w", err)
    }

    // Check if file exists
    if _, err := os.Stat(path); err == nil {
        return fmt.Errorf("config already exists at %s\nUse 'sow config edit' to modify", path)
    }

    // Create parent directories
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return fmt.Errorf("failed to create config directory: %w", err)
    }

    // Write template
    if err := os.WriteFile(path, []byte(configTemplate), 0644); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }

    cmd.Printf("Created configuration at %s\n", path)
    return nil
}
```

### 2. Define Config Template

Add the config template constant:

```go
var configTemplate = `# Sow Agent Configuration
# Location: ~/.config/sow/config.yaml
#
# This file configures which AI CLI tools handle which agent roles.
# If this file doesn't exist, all agents use Claude Code by default.
#
# Configuration priority:
#   1. Environment variables (SOW_AGENTS_IMPLEMENTER=cursor)
#   2. This config file
#   3. Built-in defaults (Claude Code)

agents:
  # Executor definitions
  executors:
    claude-code:
      type: "claude"
      settings:
        yolo_mode: false    # Set true to skip permission prompts
        # model: "sonnet"   # or "opus", "haiku"

    # Uncomment to enable Cursor
    # cursor:
    #   type: "cursor"
    #   settings:
    #     yolo_mode: false

    # Uncomment to enable Windsurf
    # windsurf:
    #   type: "windsurf"
    #   settings:
    #     yolo_mode: false

  # Bindings: which executor handles which agent role
  bindings:
    orchestrator: "claude-code"
    implementer: "claude-code"
    architect: "claude-code"
    reviewer: "claude-code"
    planner: "claude-code"
    researcher: "claude-code"
    decomposer: "claude-code"
`
```

### 3. Register with Parent Command

Update `cli/cmd/config/config.go` to add:
```go
cmd.AddCommand(newInitCmd())
```

## Acceptance Criteria

1. **Creates config file**: Running `sow config init` creates file at platform-specific path
2. **Template is valid YAML**: The generated file can be parsed by `LoadUserConfig()`
3. **Error on existing file**: Returns error if file already exists with helpful message
4. **Creates parent directories**: Works even if `~/.config/sow/` doesn't exist
5. **Correct permissions**: File created with 0644, directory with 0755
6. **Output message**: Prints confirmation with file path

### Tests to Write

Create `cli/cmd/config/init_test.go`:

```go
func TestRunInit_CreatesFile(t *testing.T) {
    tempDir := t.TempDir()
    // Test that file is created with correct content
}

func TestRunInit_ErrorsOnExistingFile(t *testing.T) {
    tempDir := t.TempDir()
    // Create file first, then verify init returns error
}

func TestRunInit_CreatesParentDirectories(t *testing.T) {
    tempDir := t.TempDir()
    // Test with nested path where parent doesn't exist
}

func TestConfigTemplate_ValidYAML(t *testing.T) {
    // Parse configTemplate as YAML and verify no errors
}

func TestConfigTemplate_PassesValidation(t *testing.T) {
    // Write template to temp file, load with LoadUserConfig,
    // verify it loads successfully
}
```

## Technical Details

### Testing with Custom Paths

Since `GetUserConfigPath()` uses `os.UserConfigDir()`, tests need to either:
1. Create a helper that accepts a path parameter for testing
2. Use dependency injection
3. Work with the actual path in temp directories

Recommended approach: Create a helper function for the core logic:

```go
func initConfigAtPath(path string) error {
    // Core logic here
}

func runInit(cmd *cobra.Command, _ []string) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return err
    }
    return initConfigAtPath(path)
}
```

This allows tests to call `initConfigAtPath` directly with temp paths.

### File Locations

- New: `cli/cmd/config/init.go`
- New: `cli/cmd/config/init_test.go`
- Modify: `cli/cmd/config/config.go` (add subcommand)

## Relevant Inputs

- `cli/cmd/config/config.go` - Parent command to register with
- `cli/internal/sow/user_config.go` - Contains GetUserConfigPath(), LoadUserConfig()
- `cli/schemas/user_config.cue` - Config schema for reference
- `.sow/project/context/issue-100.md` - Full requirements including template content

## Examples

### Success Case

```bash
$ sow config init
Created configuration at /Users/josh/.config/sow/config.yaml
```

### Error Case (File Exists)

```bash
$ sow config init
Error: config already exists at /Users/josh/.config/sow/config.yaml
Use 'sow config edit' to modify
```

## Constraints

- Template must be valid YAML that passes `ValidateUserConfig()`
- Template should include comments explaining all options
- Must not overwrite existing configuration
- Error messages must be user-friendly and actionable
