package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/your-org/sow/internal/config"
)

// NewInitCmd creates the init command
func NewInitCmd() *cobra.Command {
	var force bool
	var dir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize sow structure in repository",
		Long: `Initialize sow structure by creating .sow/ directory with required subdirectories.

Creates:
  - .sow/               Main directory
  - .sow/knowledge/     Repository-specific documentation
  - .sow/.version       Version tracking file

If .sow/ already exists, the command will skip initialization unless --force is used.`,
		Example: `  # Initialize sow in current directory
  sow init

  # Force reinitialize (recreate structure)
  sow init --force

  # Initialize in specific directory
  sow init --dir /path/to/repo`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, dir, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force recreate structure if it exists")
	cmd.Flags().StringVar(&dir, "dir", ".", "Directory to initialize (default: current directory)")

	return cmd
}

func runInit(cmd *cobra.Command, dir string, force bool) error {
	sowDir := filepath.Join(dir, ".sow")

	// Check if .sow/ already exists
	if _, err := os.Stat(sowDir); err == nil {
		if !force {
			fmt.Fprintf(cmd.OutOrStdout(), "sow already initialized in this directory\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Use --force to recreate structure\n")
			return nil
		}

		// Remove existing directory if --force
		if err := os.RemoveAll(sowDir); err != nil {
			return fmt.Errorf("failed to remove existing .sow/ directory: %w", err)
		}
	}

	// Create .sow/ directory
	if err := os.MkdirAll(sowDir, 0755); err != nil {
		return fmt.Errorf("failed to create .sow/ directory: %w", err)
	}

	// Create .sow/knowledge/ directory
	knowledgeDir := filepath.Join(sowDir, "knowledge")
	if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .sow/knowledge/ directory: %w", err)
	}

	// Create .sow/.version file
	versionFile := filepath.Join(sowDir, ".version")
	versionContent := config.Version
	if err := os.WriteFile(versionFile, []byte(versionContent), 0644); err != nil {
		return fmt.Errorf("failed to create .sow/.version file: %w", err)
	}

	// Output confirmation
	fmt.Fprintf(cmd.OutOrStdout(), "sow initialized successfully!\n\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Created directories:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - .sow/\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - .sow/knowledge/\n")
	fmt.Fprintf(cmd.OutOrStdout(), "\nCreated files:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - .sow/.version (version: %s)\n", config.Version)

	return nil
}
