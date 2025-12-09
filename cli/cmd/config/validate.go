package config

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/libs/schemas"
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
	return runValidateWithPath(cmd, path)
}

// runValidateWithPath is a helper that allows testing with custom paths.
func runValidateWithPath(cmd *cobra.Command, path string) error {
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

	// Handle empty file or comment-only file
	if len(data) == 0 || isCommentOnly(data) {
		cmd.Println("OK YAML syntax valid")
		cmd.Println("OK Schema valid")
		cmd.Println("OK Executor types valid")
		cmd.Println("OK Bindings reference defined executors")
		cmd.Println("\nConfiguration is valid.")
		return nil
	}

	// Validate YAML syntax
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		cmd.Printf("X YAML syntax error: %v\n", err)
		return fmt.Errorf("validation failed")
	}
	cmd.Println("OK YAML syntax valid")

	// Parse into config struct
	var config schemas.UserConfig
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
	warnings := checkExecutorBinariesFromConfig(&config)
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

// isCommentOnly checks if YAML content is just comments or whitespace.
func isCommentOnly(data []byte) bool {
	for _, line := range splitLines(data) {
		trimmed := trimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		if trimmed[0] != '#' {
			return false
		}
	}
	return true
}

// splitLines splits data into lines.
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

// trimSpace trims whitespace from a byte slice.
func trimSpace(data []byte) []byte {
	start := 0
	end := len(data)
	for start < end && isSpace(data[start]) {
		start++
	}
	for end > start && isSpace(data[end-1]) {
		end--
	}
	return data[start:end]
}

// isSpace checks if a byte is whitespace.
func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r'
}

// checkExecutorBinariesFromConfig checks if the configured executors have their binaries available.
// Returns a list of warnings for missing binaries.
func checkExecutorBinariesFromConfig(config *schemas.UserConfig) []string {
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

	for name, executor := range config.Agents.Executors {
		binary, ok := binaries[executor.Type]
		if !ok {
			continue
		}

		// Check if binary is on PATH
		if _, err := exec.LookPath(binary); err != nil {
			warnings = append(warnings, fmt.Sprintf(
				"%s executor '%s' requires '%s' binary, but it was not found on PATH",
				executor.Type, name, binary,
			))
		}
	}

	return warnings
}

