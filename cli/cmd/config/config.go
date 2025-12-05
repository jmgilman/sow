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

	// Add subcommands
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newPathCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newEditCmd())
	cmd.AddCommand(newResetCmd())

	return cmd
}
