# Task 040: Config Show Command

## Context

The sow CLI needs a `config show` command that displays the effective (merged) configuration. This shows what configuration is actually being used after merging defaults, the config file, and environment variable overrides.

The command helps users understand:
- What configuration is currently in effect
- Where each setting comes from (defaults, file, or environment)
- Whether their config file is being used

## Requirements

### 1. Create Show Command

Create `cli/cmd/config/show.go` with:

```go
package config

import (
    "fmt"
    "os"
    "strings"

    "github.com/jmgilman/sow/cli/internal/sow"
    "gopkg.in/yaml.v3"
)

func newShowCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "show",
        Short: "Show effective configuration",
        Long: `Display the effective configuration after merging:
  1. Built-in defaults
  2. Config file (if exists)
  3. Environment variables (highest priority)

The output shows what configuration is actually being used.`,
        RunE: runShow,
    }
}

func runShow(cmd *cobra.Command, _ []string) error {
    // Load effective config
    config, err := sow.LoadUserConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    // Get config file status
    path, _ := sow.GetUserConfigPath()
    _, fileErr := os.Stat(path)
    fileExists := fileErr == nil

    // Print header with source info
    cmd.Println("# Effective configuration (merged from defaults + file + environment)")
    if fileExists {
        cmd.Printf("# Config file: %s (exists)\n", path)
    } else {
        cmd.Printf("# Config file: %s (not found, using defaults)\n", path)
    }

    // Check for env overrides
    envOverrides := getEnvOverrides()
    if len(envOverrides) > 0 {
        cmd.Printf("# Environment overrides: %s\n", strings.Join(envOverrides, ", "))
    }
    cmd.Println()

    // Output as YAML
    output, err := yaml.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    cmd.Print(string(output))

    return nil
}

// getEnvOverrides returns a list of SOW_AGENTS_* environment variables that are set.
func getEnvOverrides() []string {
    envVars := []string{
        "SOW_AGENTS_ORCHESTRATOR",
        "SOW_AGENTS_IMPLEMENTER",
        "SOW_AGENTS_ARCHITECT",
        "SOW_AGENTS_REVIEWER",
        "SOW_AGENTS_PLANNER",
        "SOW_AGENTS_RESEARCHER",
        "SOW_AGENTS_DECOMPOSER",
    }

    var set []string
    for _, ev := range envVars {
        if os.Getenv(ev) != "" {
            set = append(set, ev)
        }
    }
    return set
}
```

### 2. Register with Parent Command

Update `cli/cmd/config/config.go` to add:
```go
cmd.AddCommand(newShowCmd())
```

## Acceptance Criteria

1. **Shows merged config**: Displays effective configuration after all merging
2. **Indicates file status**: Header shows if config file exists or not
3. **Shows env overrides**: Lists any environment variables that are overriding settings
4. **Valid YAML output**: The config portion is valid parseable YAML
5. **Works without config file**: Returns defaults when no file exists
6. **Works with partial config**: Merges partial config with defaults correctly

### Tests to Write

Create `cli/cmd/config/show_test.go`:

```go
func TestRunShow_NoConfigFile(t *testing.T) {
    // Should show defaults with "(not found, using defaults)"
}

func TestRunShow_WithConfigFile(t *testing.T) {
    // Create temp config, verify "(exists)" in output
}

func TestRunShow_EnvOverrides(t *testing.T) {
    t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")
    // Verify env var is shown in header
}

func TestGetEnvOverrides_ReturnsSetVars(t *testing.T) {
    t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")
    t.Setenv("SOW_AGENTS_ARCHITECT", "windsurf")

    overrides := getEnvOverrides()

    if len(overrides) != 2 {
        t.Errorf("expected 2 overrides, got %d", len(overrides))
    }
}

func TestGetEnvOverrides_IgnoresEmpty(t *testing.T) {
    t.Setenv("SOW_AGENTS_IMPLEMENTER", "")

    overrides := getEnvOverrides()

    // Empty strings should not be included
    for _, o := range overrides {
        if o == "SOW_AGENTS_IMPLEMENTER" {
            t.Error("empty env var should not be in overrides")
        }
    }
}

func TestRunShow_OutputIsValidYAML(t *testing.T) {
    // Capture output and verify the YAML portion parses
}
```

## Technical Details

### Output Format

The output consists of:
1. Comment header with metadata (config file status, env overrides)
2. Blank line
3. YAML representation of the effective config

Example output:
```yaml
# Effective configuration (merged from defaults + file + environment)
# Config file: /Users/josh/.config/sow/config.yaml (exists)
# Environment overrides: SOW_AGENTS_IMPLEMENTER

agents:
  executors:
    claude-code:
      type: claude
      settings:
        yolo_mode: false
  bindings:
    orchestrator: claude-code
    implementer: cursor
    architect: claude-code
    ...
```

### Testing with Command Output

Use `cmd.SetOut()` to capture output:

```go
func TestRunShow_NoConfigFile(t *testing.T) {
    cmd := newShowCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)

    // Ensure no config file exists (use temp dir approach)
    // Then execute and check output
}
```

### File Locations

- New: `cli/cmd/config/show.go`
- New: `cli/cmd/config/show_test.go`
- Modify: `cli/cmd/config/config.go` (add subcommand)

## Relevant Inputs

- `cli/cmd/config/config.go` - Parent command to register with
- `cli/internal/sow/user_config.go` - LoadUserConfig(), GetUserConfigPath()
- `cli/internal/sow/user_config_test.go` - Shows environment override testing patterns

## Examples

### No Config File

```bash
$ sow config show
# Effective configuration (merged from defaults + file + environment)
# Config file: /Users/josh/.config/sow/config.yaml (not found, using defaults)

agents:
  executors:
    claude-code:
      type: claude
      settings:
        yolo_mode: false
  bindings:
    orchestrator: claude-code
    implementer: claude-code
    architect: claude-code
    reviewer: claude-code
    planner: claude-code
    researcher: claude-code
    decomposer: claude-code
```

### With Environment Override

```bash
$ SOW_AGENTS_IMPLEMENTER=cursor sow config show
# Effective configuration (merged from defaults + file + environment)
# Config file: /Users/josh/.config/sow/config.yaml (exists)
# Environment overrides: SOW_AGENTS_IMPLEMENTER

agents:
  executors:
    claude-code:
      type: claude
      settings:
        yolo_mode: false
    cursor:
      type: cursor
  bindings:
    orchestrator: claude-code
    implementer: cursor
    ...
```

## Constraints

- Header lines must start with `#` (YAML comment)
- YAML output must be valid and parseable
- Must use `LoadUserConfig()` for consistency with rest of application
- Must handle all error cases gracefully (unreadable file, etc.)
