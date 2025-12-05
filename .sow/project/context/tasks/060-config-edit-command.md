# Task 060: Config Edit Command

## Context

The sow CLI needs a `config edit` command that opens the configuration file in the user's preferred editor. If the file doesn't exist, it creates one with the default template first.

This provides a convenient way for users to modify their configuration without needing to remember the file path.

## Requirements

### 1. Create Edit Command

Create `cli/cmd/config/edit.go` with:

```go
package config

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "edit",
        Short: "Open configuration in editor",
        Long: `Open the configuration file in your preferred editor.

Uses $EDITOR environment variable, falling back to 'vi' if not set.

If no configuration file exists, creates one with the default template first.`,
        RunE: runEdit,
    }
}

func runEdit(cmd *cobra.Command, _ []string) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return fmt.Errorf("failed to get config path: %w", err)
    }

    // Create file with template if it doesn't exist
    if _, err := os.Stat(path); os.IsNotExist(err) {
        if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
            return fmt.Errorf("failed to create config directory: %w", err)
        }
        if err := os.WriteFile(path, []byte(configTemplate), 0644); err != nil {
            return fmt.Errorf("failed to create config: %w", err)
        }
        cmd.Printf("Created new configuration at %s\n", path)
    }

    // Get editor from environment
    editor := os.Getenv("EDITOR")
    if editor == "" {
        editor = "vi" // Fallback
    }

    // Open editor
    editCmd := exec.Command(editor, path)
    editCmd.Stdin = os.Stdin
    editCmd.Stdout = os.Stdout
    editCmd.Stderr = os.Stderr

    if err := editCmd.Run(); err != nil {
        return fmt.Errorf("editor failed: %w", err)
    }

    return nil
}
```

### 2. Share Config Template

The `configTemplate` constant defined in `init.go` should be shared. Options:
1. Export it as `ConfigTemplate` (capital C)
2. Move it to a separate file like `template.go`
3. Keep it in `init.go` and reference from `edit.go`

Recommended: Create `cli/cmd/config/template.go`:

```go
package config

// configTemplate is the default configuration file template.
// Used by both init and edit commands.
var configTemplate = `# Sow Agent Configuration
...
`
```

### 3. Register with Parent Command

Update `cli/cmd/config/config.go` to add:
```go
cmd.AddCommand(newEditCmd())
```

## Acceptance Criteria

1. **Opens editor**: Running `sow config edit` opens $EDITOR with config file
2. **Creates file if missing**: Creates config with template if it doesn't exist
3. **Respects $EDITOR**: Uses user's preferred editor
4. **Falls back to vi**: Uses `vi` if $EDITOR is not set
5. **Handles editor errors**: Reports error if editor fails
6. **Interactive I/O**: Correctly passes stdin/stdout/stderr to editor

### Tests to Write

Create `cli/cmd/config/edit_test.go`:

```go
func TestRunEdit_CreatesFileIfMissing(t *testing.T) {
    // Use temp directory, verify file created with template
    // Note: Don't actually run editor in test
}

func TestRunEdit_UsesEditorEnvVar(t *testing.T) {
    t.Setenv("EDITOR", "echo")  // Use echo as a dummy editor
    // Verify the right editor is invoked
}

func TestRunEdit_FallsBackToVi(t *testing.T) {
    // Unset EDITOR, verify vi is used
    // (may need to mock exec.Command)
}

func TestGetEditor_RespectsEnvVar(t *testing.T) {
    t.Setenv("EDITOR", "nano")
    editor := getEditor()
    if editor != "nano" {
        t.Errorf("expected 'nano', got '%s'", editor)
    }
}

func TestGetEditor_FallsBackToVi(t *testing.T) {
    // Ensure EDITOR is unset
    editor := getEditor()
    if editor != "vi" {
        t.Errorf("expected 'vi', got '%s'", editor)
    }
}
```

## Technical Details

### Testing Editor Invocation

Testing editor commands is tricky because:
1. You don't want to actually open an editor in tests
2. You need to verify the correct command is constructed

Options:
1. Extract editor logic into testable functions
2. Use dependency injection for command execution
3. Test the file creation separately from editor invocation

Recommended approach - extract a helper:

```go
// getEditor returns the editor command to use.
func getEditor() string {
    if editor := os.Getenv("EDITOR"); editor != "" {
        return editor
    }
    return "vi"
}

// runEditorCmd executes the editor command.
// Extracted for testing.
var runEditorCmd = func(editor, path string) error {
    cmd := exec.Command(editor, path)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

Then in tests, you can replace `runEditorCmd` with a mock.

### File Creation

The file creation logic should:
1. Check if file exists
2. Create parent directory if needed
3. Write template with 0644 permissions
4. Print message about creation

This is similar to `init` command but doesn't error on existing file.

### File Locations

- New: `cli/cmd/config/edit.go`
- New: `cli/cmd/config/edit_test.go`
- New (optional): `cli/cmd/config/template.go` (if extracting template)
- Modify: `cli/cmd/config/config.go` (add subcommand)
- Modify: `cli/cmd/config/init.go` (if moving template)

## Relevant Inputs

- `cli/cmd/config/config.go` - Parent command to register with
- `cli/cmd/config/init.go` - Contains configTemplate constant to share
- `cli/internal/sow/user_config.go` - Contains GetUserConfigPath()

## Examples

### New Config File

```bash
$ sow config edit
Created new configuration at /Users/josh/.config/sow/config.yaml
# (editor opens with template)
```

### Existing Config File

```bash
$ sow config edit
# (editor opens with existing content)
```

### Custom Editor

```bash
$ EDITOR=code sow config edit
# (VS Code opens)
```

### Editor Error

```bash
$ EDITOR=nonexistent sow config edit
Error: editor failed: exec: "nonexistent": executable file not found in $PATH
```

## Constraints

- Must correctly pass stdin/stdout/stderr to editor (interactive)
- Must handle both GUI editors (code, atom) and terminal editors (vi, nano)
- Must create parent directories if they don't exist
- Must share template with init command (DRY)
- Template file permissions must be 0644, directory 0755
