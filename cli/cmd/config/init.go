package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/libs/config"
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
	path, err := config.GetUserConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	return runInitWithPath(cmd, path)
}

// runInitWithPath is a helper that allows testing with custom paths.
func runInitWithPath(cmd *cobra.Command, path string) error {
	if err := initConfigAtPath(path); err != nil {
		return err
	}

	cmd.Printf("Created configuration at %s\n", path)
	return nil
}

// initConfigAtPath creates a config file at the specified path with the template content.
// Returns an error if the file already exists.
func initConfigAtPath(path string) error {
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

	return nil
}
