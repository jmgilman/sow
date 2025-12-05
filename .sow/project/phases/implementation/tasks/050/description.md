# Task 050: Config Validate Command

## Context

The sow CLI needs a `config validate` command that validates the user's configuration file. This helps users identify configuration errors before they cause problems during agent execution.

The validation system already exists in `cli/internal/sow/user_config.go` with `ValidateUserConfig()`, which checks:
- Executor types are valid ("claude", "cursor", "windsurf")
- Bindings reference defined executors

This command provides a user-friendly interface for that validation.

## Requirements

### 1. Create Validate Command

Create `cli/cmd/config/validate.go` with:

```go
package config

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"
)

func newValidateCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "validate",
        Short: "Validate configuration file",
        Long: `Validate the configuration file for syntax and semantic errors.

Checks performed:
  - YAML syntax is valid
  - Executor types are valid (claude, cursor, windsurf)
  - Bindings reference defined executors
  - (Optional) Executor binaries are available on PATH`,
        RunE: runValidate,
    }
}

func runValidate(cmd *cobra.Command, _ []string) error {
    path, err := sow.GetUserConfigPath()
    if err != nil {
        return fmt.Errorf("failed to get config path: %w", err)
    }

    cmd.Printf("Validating configuration at %s...\n\n", path)

    // Check if file exists
    data, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        cmd.Println("No configuration file found (using defaults)")
        cmd.Println("Run 'sow config init' to create one.")
        return nil
    }
    if err != nil {
        return fmt.Errorf("failed to read config: %w", err)
    }

    // Validate YAML syntax
    var raw interface{}
    if err := yaml.Unmarshal(data, &raw); err != nil {
        cmd.Printf("X YAML syntax error: %v\n", err)
        return fmt.Errorf("validation failed")
    }
    cmd.Println("OK YAML syntax valid")

    // Parse into config struct
    var config sow.schemas.UserConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        cmd.Printf("X Failed to parse config: %v\n", err)
        return fmt.Errorf("validation failed")
    }

    // Validate schema/semantics
    if err := sow.ValidateUserConfig(&config); err != nil {
        cmd.Printf("X Validation error: %v\n", err)
        return fmt.Errorf("validation failed")
    }
    cmd.Println("OK Schema valid")
    cmd.Println("OK Executor types valid")
    cmd.Println("OK Bindings reference defined executors")

    // Check executor binaries (warnings only)
    warnings := checkExecutorBinaries(&config)
    for _, w := range warnings {
        cmd.Printf("WARN Warning: %s\n", w)
    }

    if len(warnings) > 0 {
        cmd.Printf("\nConfiguration is valid with %d warning(s).\n", len(warnings))
    } else {
        cmd.Println("\nConfiguration is valid.")
    }

    return nil
}

// checkExecutorBinaries checks if the configured executors have their binaries available.
// Returns a list of warnings for missing binaries.
func checkExecutorBinaries(config *schemas.UserConfig) []string {
    if config == nil || config.Agents == nil {
        return nil
    }

    var warnings []string

    // Map executor types to their binary names
    binaries := map[string]string{
        "claude":   "claude",
        "cursor":   "cursor",
        "windsurf": "windsurf",
    }

    for name, exec := range config.Agents.Executors {
        binary, ok := binaries[exec.Type]
        if !ok {
            continue
        }

        // Check if binary is on PATH
        if _, err := exec.LookPath(binary); err != nil {
            warnings = append(warnings, fmt.Sprintf(
                "%s executor '%s' requires '%s' binary, but it was not found on PATH",
                exec.Type, name, binary,
            ))
        }
    }

    return warnings
}
```

### 2. Register with Parent Command

Update `cli/cmd/config/config.go` to add:
```go
cmd.AddCommand(newValidateCmd())
```

## Acceptance Criteria

1. **Reports missing file**: When no config file exists, says so without error
2. **Catches YAML syntax errors**: Reports line/column of syntax errors
3. **Catches invalid executor types**: Reports unknown types like "copilot"
4. **Catches undefined executor references**: Reports when binding uses undefined executor
5. **Warns about missing binaries**: Non-fatal warning if executor binary not found
6. **Reports success clearly**: Shows checkmarks/OK for each validation step
7. **Returns non-zero exit on error**: Validation failures return error

### Tests to Write

Create `cli/cmd/config/validate_test.go`:

```go
func TestRunValidate_NoConfigFile(t *testing.T) {
    // Should report "No configuration file found" without error
}

func TestRunValidate_ValidConfig(t *testing.T) {
    // Create valid temp config, verify success messages
}

func TestRunValidate_InvalidYAML(t *testing.T) {
    // Create config with YAML syntax error, verify error reported
}

func TestRunValidate_InvalidExecutorType(t *testing.T) {
    // Create config with invalid executor type, verify error reported
}

func TestRunValidate_UndefinedExecutorBinding(t *testing.T) {
    // Create config with binding to undefined executor, verify error
}

func TestCheckExecutorBinaries_MissingBinary(t *testing.T) {
    // Create config with executor whose binary doesn't exist
    // Verify warning is returned
}

func TestCheckExecutorBinaries_AllPresent(t *testing.T) {
    // Mock or use real binaries, verify no warnings
}
```

## Technical Details

### Output Format

The output uses consistent prefixes:
- `OK` - Validation passed
- `X` - Validation failed (causes non-zero exit)
- `WARN` - Warning (does not cause failure)

Example successful output:
```
Validating configuration at /Users/josh/.config/sow/config.yaml...

OK YAML syntax valid
OK Schema valid
OK Executor types valid
OK Bindings reference defined executors

Configuration is valid.
```

Example failure output:
```
Validating configuration at /Users/josh/.config/sow/config.yaml...

OK YAML syntax valid
X Validation error: unknown executor type "copilot" for executor "my-copilot"; must be one of: claude, cursor, windsurf

Error: validation failed
```

### Import Adjustment

Note: You'll need to import the schemas package to use `schemas.UserConfig`:
```go
import "github.com/jmgilman/sow/cli/schemas"
```

And adjust the code to parse into the correct type.

### Testing Binary Checks

For testing `checkExecutorBinaries`, you have options:
1. Accept that the test behavior depends on what's installed
2. Create a testable interface for `exec.LookPath`
3. Skip binary checks in unit tests, cover in integration tests

Recommended: Test the logic with known-missing binaries (e.g., type "windsurf" which is unlikely to be installed).

### File Locations

- New: `cli/cmd/config/validate.go`
- New: `cli/cmd/config/validate_test.go`
- Modify: `cli/cmd/config/config.go` (add subcommand)

## Relevant Inputs

- `cli/cmd/config/config.go` - Parent command to register with
- `cli/internal/sow/user_config.go` - Contains ValidateUserConfig() function
- `cli/internal/sow/user_config_test.go` - Shows validation test patterns
- `cli/schemas/user_config.cue` - Schema definition for reference

## Examples

### No Config File

```bash
$ sow config validate
Validating configuration at /Users/josh/.config/sow/config.yaml...

No configuration file found (using defaults)
Run 'sow config init' to create one.
```

### Valid Config

```bash
$ sow config validate
Validating configuration at /Users/josh/.config/sow/config.yaml...

OK YAML syntax valid
OK Schema valid
OK Executor types valid
OK Bindings reference defined executors

Configuration is valid.
```

### Invalid YAML Syntax

```bash
$ sow config validate
Validating configuration at /Users/josh/.config/sow/config.yaml...

X YAML syntax error: yaml: line 5: could not find expected ':'

Error: validation failed
```

### Invalid Executor Type

```bash
$ sow config validate
Validating configuration at /Users/josh/.config/sow/config.yaml...

OK YAML syntax valid
X Validation error: unknown executor type "copilot" for executor "my-copilot"; must be one of: claude, cursor, windsurf

Error: validation failed
```

### Valid with Warnings

```bash
$ sow config validate
Validating configuration at /Users/josh/.config/sow/config.yaml...

OK YAML syntax valid
OK Schema valid
OK Executor types valid
OK Bindings reference defined executors
WARN Warning: cursor executor 'cursor' requires 'cursor' binary, but it was not found on PATH

Configuration is valid with 1 warning(s).
```

## Constraints

- Must use existing `ValidateUserConfig()` for semantic validation
- Binary warnings should not cause exit code failure
- YAML and semantic errors should cause exit code 1
- Output should be clear and actionable for users
