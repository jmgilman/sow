# Issue #100: Config CLI Commands

**URL**: https://github.com/jmgilman/sow/issues/100
**State**: OPEN

## Description

# Work Unit 005: Config CLI Commands

## Behavioral Goal

As a sow user, I need CLI commands to manage my agent configuration, so that I can initialize a config file, view effective settings, validate my configuration, and customize which executors handle which agents.

## Scope

### In Scope
- `sow config init` - create config file with commented template
- `sow config path [--exists]` - show config file location
- `sow config show` - display effective configuration (merged)
- `sow config validate` - validate config file
- `sow config edit` - open config in $EDITOR
- `sow config reset` - remove config file (with backup)
- Integration tests

### Out of Scope
- Config hot-reloading (not needed)
- GUI configuration (out of scope)

## Existing Code Context

**Command package pattern** (`cli/cmd/refs/refs.go`):
```go
func NewRefsCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "refs", ...}
    cmd.AddCommand(newAddCmd())
    // ...
    return cmd
}
```

**User config loading** (from work unit 004):
- `LoadUserConfig()` - loads and merges config
- `GetUserConfigPath()` - returns platform-specific path
- `validateUserConfig()` - validates against schema

## Documentation Context

**Design doc** (`.sow/knowledge/designs/multi-agent-architecture.md`):
- Section "Configuration Initialization" (lines 574-630) shows `sow config init`
- Section "Configuration Discovery" (lines 632-676) shows path/help commands
- Section "Configuration Validation" (lines 678-705) shows validate command
- Section "Additional Config Commands" (lines 707-756) shows show/edit/reset

## File Structure

```
cli/cmd/config/
├── config.go           # Parent command
├── init.go             # sow config init
├── init_test.go
├── path.go             # sow config path
├── path_test.go
├── show.go             # sow config show
├── show_test.go
├── validate.go         # sow config validate
├── validate_test.go
├── edit.go             # sow config edit
├── edit_test.go
├── reset.go            # sow config reset
└── reset_test.go
```

Update root:
```
cli/cmd/root.go         # Add config.NewConfigCmd()
```

## Implementation Approach

### Parent Command

```go
// cli/cmd/config/config.go
func NewConfigCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config",
        Short: "Manage user configuration",
        Long: `Manage sow user configuration for agent preferences.

Configuration is stored at ~/.config/sow/config.yaml (Linux/Mac)
or %APPDATA%\sow\config.yaml (Windows).

If no configuration exists, sow uses defaults (Claude Code for all agents).`,
    }

    cmd.AddCommand(newInitCmd())
    cmd.AddCommand(newPathCmd())
    cmd.AddCommand(newShowCmd())
    cmd.AddCommand(newValidateCmd())
    cmd.AddCommand(newEditCmd())
    cmd.AddCommand(newResetCmd())

    return cmd
}
```

### Init Command

```go
// cli/cmd/config/init.go
func newInitCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "init",
        Short: "Create configuration file with template",
        RunE:  runInit,
    }
}

func runInit(cmd *cobra.Command, _ []string) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return err
    }

    // Check if exists
    if _, err := os.Stat(path); err == nil {
        return fmt.Errorf("config already exists at %s\nUse 'sow config edit' to modify", path)
    }

    // Create directory
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }

    // Write template
    if err := os.WriteFile(path, []byte(configTemplate), 0644); err != nil {
        return err
    }

    fmt.Printf("Created configuration at %s\n", path)
    return nil
}

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

### Path Command

```go
// cli/cmd/config/path.go
func newPathCmd() *cobra.Command {
    var existsFlag bool

    cmd := &cobra.Command{
        Use:   "path",
        Short: "Show configuration file path",
        RunE: func(cmd *cobra.Command, _ []string) error {
            return runPath(cmd, existsFlag)
        },
    }

    cmd.Flags().BoolVar(&existsFlag, "exists", false, "Check if config file exists")
    return cmd
}

func runPath(cmd *cobra.Command, checkExists bool) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return err
    }

    if checkExists {
        _, err := os.Stat(path)
        fmt.Println(err == nil)
        return nil
    }

    fmt.Println(path)
    return nil
}
```

### Show Command

```go
// cli/cmd/config/show.go
func newShowCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "show",
        Short: "Show effective configuration",
        RunE:  runShow,
    }
}

func runShow(cmd *cobra.Command, _ []string) error {
    config, err := sow.LoadUserConfig()
    if err != nil {
        return err
    }

    path, _ := sow.GetUserConfigPath()
    _, fileErr := os.Stat(path)
    fileExists := fileErr == nil

    // Print header with source info
    fmt.Println("# Effective configuration (merged from defaults + file + environment)")
    if fileExists {
        fmt.Printf("# Config file: %s (exists)\n", path)
    } else {
        fmt.Printf("# Config file: %s (not found, using defaults)\n", path)
    }

    // Check for env overrides
    envOverrides := getEnvOverrides()
    if len(envOverrides) > 0 {
        fmt.Printf("# Environment overrides: %s\n", strings.Join(envOverrides, ", "))
    }
    fmt.Println()

    // Output as YAML
    output, _ := yaml.Marshal(config)
    fmt.Print(string(output))

    return nil
}
```

### Validate Command

```go
// cli/cmd/config/validate.go
func newValidateCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "validate",
        Short: "Validate configuration file",
        RunE:  runValidate,
    }
}

func runValidate(cmd *cobra.Command, _ []string) error {
    path, _ := sow.GetUserConfigPath()
    fmt.Printf("Validating configuration at %s...\n\n", path)

    // Check file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        fmt.Println("No configuration file found (using defaults)")
        return nil
    }

    // Try to load (includes validation)
    _, err := sow.LoadUserConfigStrict()  // Strict mode - don't apply defaults
    if err != nil {
        fmt.Printf("✗ Validation failed: %v\n", err)
        return err
    }

    fmt.Println("✓ YAML syntax valid")
    fmt.Println("✓ Schema valid")
    fmt.Println("✓ Executor types valid")
    fmt.Println("✓ Bindings reference defined executors")

    // Check executor binaries (warnings only)
    warnings := checkExecutorBinaries()
    for _, w := range warnings {
        fmt.Printf("⚠ Warning: %s\n", w)
    }

    if len(warnings) > 0 {
        fmt.Printf("\nConfiguration is valid with %d warning(s).\n", len(warnings))
    } else {
        fmt.Println("\nConfiguration is valid.")
    }

    return nil
}
```

### Edit Command

```go
// cli/cmd/config/edit.go
func newEditCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "edit",
        Short: "Open configuration in editor",
        RunE:  runEdit,
    }
}

func runEdit(cmd *cobra.Command, _ []string) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return err
    }

    // Create with template if doesn't exist
    if _, err := os.Stat(path); os.IsNotExist(err) {
        if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
            return err
        }
        if err := os.WriteFile(path, []byte(configTemplate), 0644); err != nil {
            return err
        }
    }

    // Get editor
    editor := os.Getenv("EDITOR")
    if editor == "" {
        editor = "vi"  // Fallback
    }

    // Open editor
    editCmd := exec.Command(editor, path)
    editCmd.Stdin = os.Stdin
    editCmd.Stdout = os.Stdout
    editCmd.Stderr = os.Stderr

    return editCmd.Run()
}
```

### Reset Command

```go
// cli/cmd/config/reset.go
func newResetCmd() *cobra.Command {
    var forceFlag bool

    cmd := &cobra.Command{
        Use:   "reset",
        Short: "Remove configuration file",
        RunE: func(cmd *cobra.Command, _ []string) error {
            return runReset(cmd, forceFlag)
        },
    }

    cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation")
    return cmd
}

func runReset(cmd *cobra.Command, force bool) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return err
    }

    if _, err := os.Stat(path); os.IsNotExist(err) {
        fmt.Println("No configuration file to reset")
        return nil
    }

    if !force {
        fmt.Printf("This will remove %s\n", path)
        fmt.Print("Continue? [y/N] ")
        var response string
        fmt.Scanln(&response)
        if response != "y" && response != "Y" {
            fmt.Println("Cancelled")
            return nil
        }
    }

    // Create backup
    backupPath := path + ".backup"
    if err := os.Rename(path, backupPath); err != nil {
        return err
    }

    fmt.Printf("Configuration removed (backup at %s)\n", backupPath)
    fmt.Println("Using built-in defaults")
    return nil
}
```

## Dependencies

- **Work Unit 004** (User Configuration System): `LoadUserConfig()`, `GetUserConfigPath()`, config types

## Acceptance Criteria

1. **`sow config init`**:
   - Creates config file with commented template
   - Errors if file already exists
   - Creates parent directories if needed
2. **`sow config path`**:
   - Shows platform-appropriate path
   - `--exists` flag returns true/false
3. **`sow config show`**:
   - Displays effective (merged) configuration
   - Shows source annotations (file, defaults, env)
   - Works even if file doesn't exist
4. **`sow config validate`**:
   - Reports YAML syntax errors
   - Reports schema validation errors
   - Reports undefined executor references
   - Warns about missing CLI binaries
   - Shows success message when valid
5. **`sow config edit`**:
   - Opens config in $EDITOR
   - Creates file with template if missing
   - Falls back to vi if $EDITOR not set
6. **`sow config reset`**:
   - Prompts for confirmation (unless --force)
   - Creates backup before removal
   - Reports when no file to reset
7. **Integration tests** cover all commands

## Testing Strategy

- Test each command with various scenarios
- Use temporary directories for config files
- Mock editor invocation for edit command
- Test confirmation prompts for reset
- Verify backup creation on reset
- Test cross-platform path handling
