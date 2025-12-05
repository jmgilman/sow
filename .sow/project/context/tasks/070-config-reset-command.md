# Task 070: Config Reset Command

## Context

The sow CLI needs a `config reset` command that removes the configuration file (with a backup). This allows users to revert to default configuration.

The command includes safety features:
- Confirmation prompt (unless --force is used)
- Creates a backup before deletion
- Handles gracefully when no file exists

## Requirements

### 1. Create Reset Command

Create `cli/cmd/config/reset.go` with:

```go
package config

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/spf13/cobra"
)

func newResetCmd() *cobra.Command {
    var forceFlag bool

    cmd := &cobra.Command{
        Use:   "reset",
        Short: "Remove configuration file",
        Long: `Remove the configuration file and revert to defaults.

A backup is created at config.yaml.backup before removal.
Use --force to skip the confirmation prompt.`,
        RunE: func(cmd *cobra.Command, _ []string) error {
            return runReset(cmd, forceFlag)
        },
    }

    cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation prompt")

    return cmd
}

func runReset(cmd *cobra.Command, force bool) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return fmt.Errorf("failed to get config path: %w", err)
    }

    // Check if file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        cmd.Println("No configuration file to reset")
        return nil
    }

    // Confirm unless --force
    if !force {
        cmd.Printf("This will remove %s\n", path)
        cmd.Print("Continue? [y/N] ")

        reader := bufio.NewReader(cmd.InOrStdin())
        response, err := reader.ReadString('\n')
        if err != nil {
            return fmt.Errorf("failed to read response: %w", err)
        }

        response = strings.TrimSpace(strings.ToLower(response))
        if response != "y" && response != "yes" {
            cmd.Println("Cancelled")
            return nil
        }
    }

    // Create backup
    backupPath := path + ".backup"
    if err := os.Rename(path, backupPath); err != nil {
        return fmt.Errorf("failed to create backup: %w", err)
    }

    cmd.Printf("Configuration removed (backup at %s)\n", backupPath)
    cmd.Println("Using built-in defaults")

    return nil
}
```

### 2. Register with Parent Command

Update `cli/cmd/config/config.go` to add:
```go
cmd.AddCommand(newResetCmd())
```

## Acceptance Criteria

1. **Removes config file**: File is removed after confirmation
2. **Creates backup**: Backup created at `config.yaml.backup` before removal
3. **Prompts for confirmation**: Asks user to confirm (unless --force)
4. **--force skips prompt**: With -f or --force, no confirmation needed
5. **Handles missing file**: Reports "no file to reset" without error
6. **Accepts yes/y**: Accepts both "y" and "yes" as confirmation
7. **Cancels on other input**: Any other input cancels the operation
8. **Reports success**: Confirms removal and backup location

### Tests to Write

Create `cli/cmd/config/reset_test.go`:

```go
func TestRunReset_NoFile(t *testing.T) {
    // Verify "No configuration file to reset" message
}

func TestRunReset_WithForce(t *testing.T) {
    // Create temp config, run with --force
    // Verify file removed, backup created
}

func TestRunReset_ConfirmYes(t *testing.T) {
    // Create temp config, simulate "y" input
    // Verify file removed
}

func TestRunReset_ConfirmNo(t *testing.T) {
    // Create temp config, simulate "n" input
    // Verify file NOT removed, "Cancelled" message
}

func TestRunReset_CreatesBackup(t *testing.T) {
    // Create temp config with content
    // Run reset, verify backup contains original content
}

func TestRunReset_AcceptsYes(t *testing.T) {
    // Test "yes" (full word) is accepted
}

func TestNewResetCmd_HasForceFlag(t *testing.T) {
    cmd := newResetCmd()
    flag := cmd.Flags().Lookup("force")
    if flag == nil {
        t.Error("expected --force flag")
    }
    // Also check -f shorthand
    if flag.Shorthand != "f" {
        t.Error("expected -f shorthand")
    }
}
```

## Technical Details

### Testing Interactive Prompts

To test confirmation prompts, use `cmd.SetIn()` with a strings.Reader:

```go
func TestRunReset_ConfirmYes(t *testing.T) {
    tempDir := t.TempDir()
    configPath := filepath.Join(tempDir, "config.yaml")

    // Create config file
    if err := os.WriteFile(configPath, []byte("agents: {}"), 0644); err != nil {
        t.Fatal(err)
    }

    cmd := newResetCmd()
    cmd.SetIn(strings.NewReader("y\n"))
    var buf bytes.Buffer
    cmd.SetOut(&buf)

    // Need a way to make runReset use the temp path
    // See note below about testing with custom paths
}
```

### Testing with Custom Paths

Since the command uses `sow.GetUserConfigPath()`, you'll need a way to test with custom paths. Options:

1. **Extract core logic**: Create a helper that accepts path as parameter
2. **Use environment**: If `os.UserConfigDir()` respects env vars on your platform
3. **Interface injection**: Create an interface for path resolution

Recommended approach - extract helper:

```go
func runReset(cmd *cobra.Command, force bool) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return err
    }
    return resetConfigAtPath(cmd, path, force)
}

// resetConfigAtPath is the core logic, testable with any path.
func resetConfigAtPath(cmd *cobra.Command, path string, force bool) error {
    // Implementation here
}
```

### Backup Behavior

The backup is created using `os.Rename()`, which:
- Is atomic on most filesystems
- Overwrites any existing backup
- Preserves file metadata

If you want to preserve multiple backups, that would be a future enhancement.

### File Locations

- New: `cli/cmd/config/reset.go`
- New: `cli/cmd/config/reset_test.go`
- Modify: `cli/cmd/config/config.go` (add subcommand)

## Relevant Inputs

- `cli/cmd/config/config.go` - Parent command to register with
- `cli/internal/sow/user_config.go` - Contains GetUserConfigPath()

## Examples

### No Config File

```bash
$ sow config reset
No configuration file to reset
```

### With Confirmation (Accept)

```bash
$ sow config reset
This will remove /Users/josh/.config/sow/config.yaml
Continue? [y/N] y
Configuration removed (backup at /Users/josh/.config/sow/config.yaml.backup)
Using built-in defaults
```

### With Confirmation (Decline)

```bash
$ sow config reset
This will remove /Users/josh/.config/sow/config.yaml
Continue? [y/N] n
Cancelled
```

### With --force Flag

```bash
$ sow config reset --force
Configuration removed (backup at /Users/josh/.config/sow/config.yaml.backup)
Using built-in defaults
```

### Using Short Flag

```bash
$ sow config reset -f
Configuration removed (backup at /Users/josh/.config/sow/config.yaml.backup)
Using built-in defaults
```

## Constraints

- Must always create backup before removal (no data loss)
- Must handle stdin correctly for both interactive and piped input
- Must accept both "y" and "yes" (case insensitive)
- Default (no input / Enter) should cancel, not proceed
- Backup overwrites any existing backup (simple behavior)
