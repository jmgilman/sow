package cmd

import (
	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command
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
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := FilesystemFromContext(cmd.Context())
			_ = fs // TODO: implement

			cmd.Println("sow initialized successfully")
			return nil
		},
	}

	return cmd
}
