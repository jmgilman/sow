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
	return resetConfigAtPath(cmd, path, force)
}

// resetConfigAtPath is the core logic, testable with any path.
func resetConfigAtPath(cmd *cobra.Command, path string, force bool) error {
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
