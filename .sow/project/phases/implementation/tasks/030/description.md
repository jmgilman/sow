# Task 030: Config Path Command

## Context

The sow CLI needs a `config path` command that displays the configuration file location. This helps users find where their config file should be stored and verify whether it exists.

The command supports an `--exists` flag that outputs `true` or `false` based on whether the config file exists, which is useful for scripting.

## Requirements

### 1. Create Path Command

Create `cli/cmd/config/path.go` with:

```go
package config

import (
    "fmt"
    "os"

    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/spf13/cobra"
)

func newPathCmd() *cobra.Command {
    var existsFlag bool

    cmd := &cobra.Command{
        Use:   "path",
        Short: "Show configuration file path",
        Long: `Show the path to the user configuration file.

The path is platform-specific:
  Linux/Mac: ~/.config/sow/config.yaml
  Windows:   %APPDATA%\sow\config.yaml

Use --exists to check if the file exists (for scripting).`,
        RunE: func(cmd *cobra.Command, _ []string) error {
            return runPath(cmd, existsFlag)
        },
    }

    cmd.Flags().BoolVar(&existsFlag, "exists", false, "Check if config file exists (outputs true/false)")

    return cmd
}

func runPath(cmd *cobra.Command, checkExists bool) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return fmt.Errorf("failed to get config path: %w", err)
    }

    if checkExists {
        _, err := os.Stat(path)
        if err == nil {
            cmd.Println("true")
        } else {
            cmd.Println("false")
        }
        return nil
    }

    cmd.Println(path)
    return nil
}
```

### 2. Register with Parent Command

Update `cli/cmd/config/config.go` to add:
```go
cmd.AddCommand(newPathCmd())
```

## Acceptance Criteria

1. **Shows correct path**: `sow config path` prints platform-appropriate path
2. **--exists flag works**: Outputs "true" if file exists, "false" if not
3. **Exit code is always 0**: Both existing and non-existing files should not error
4. **Path is absolute**: Output should be a full absolute path
5. **No trailing newline issues**: Output should be clean for scripting

### Tests to Write

Create `cli/cmd/config/path_test.go`:

```go
func TestRunPath_ShowsPath(t *testing.T) {
    // Verify output contains expected path components
}

func TestRunPath_ExistsFlag_FileExists(t *testing.T) {
    // Create config file, verify outputs "true"
}

func TestRunPath_ExistsFlag_FileNotExists(t *testing.T) {
    // No config file, verify outputs "false"
}

func TestNewPathCmd_HasExistsFlag(t *testing.T) {
    cmd := newPathCmd()
    flag := cmd.Flags().Lookup("exists")
    if flag == nil {
        t.Error("expected --exists flag")
    }
}
```

## Technical Details

### Testing Approach

For testing, capture command output using a buffer:

```go
func TestRunPath_ShowsPath(t *testing.T) {
    cmd := newPathCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)

    err := cmd.Execute()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    output := buf.String()
    // Verify output contains expected path components
    if !strings.Contains(output, "sow") || !strings.Contains(output, "config.yaml") {
        t.Errorf("unexpected output: %s", output)
    }
}
```

### Testing --exists Flag

For the --exists tests, you'll need to:
1. Get the actual config path using `sow.GetUserConfigPath()`
2. Ensure the directory and file exist/don't exist for your test case
3. Capture and verify output

Note: Tests should not modify the user's actual config file. If the real config path has a file, tests for "file not exists" may need to use a mock or skip.

### File Locations

- New: `cli/cmd/config/path.go`
- New: `cli/cmd/config/path_test.go`
- Modify: `cli/cmd/config/config.go` (add subcommand)

## Relevant Inputs

- `cli/cmd/config/config.go` - Parent command to register with
- `cli/internal/sow/user_config.go` - Contains GetUserConfigPath()
- `cli/internal/sow/user_config_test.go` - Shows how GetUserConfigPath is tested

## Examples

### Basic Usage

```bash
$ sow config path
/Users/josh/.config/sow/config.yaml
```

### Check Existence (File Exists)

```bash
$ sow config path --exists
true
```

### Check Existence (File Does Not Exist)

```bash
$ sow config path --exists
false
```

### Scripting Example

```bash
if [ "$(sow config path --exists)" = "true" ]; then
    echo "Config exists"
else
    sow config init
fi
```

## Constraints

- Output must be clean (no extra formatting, colors, or decorations)
- Must work on all platforms (Linux, Mac, Windows)
- --exists should never return an error, just true/false
- Path output should not include quotes or escaping
