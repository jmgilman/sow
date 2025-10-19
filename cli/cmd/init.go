package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize sow structure",
		Long: `Initialize the .sow/ directory structure in the current repository.

Creates:
  .sow/knowledge/     - Repository-specific documentation (committed)
  .sow/refs/          - External knowledge and code references (symlinks)
  .sow/.version       - sow structure version file

This command must be run from a git repository root.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sow := SowFromContext(cmd.Context())

			// Initialize .sow structure
			if err := sow.Init(); err != nil {
				return fmt.Errorf("initialization failed: %w", err)
			}

			cmd.Println("✓ sow initialized successfully")
			return nil
		},
	}

	return cmd
}
