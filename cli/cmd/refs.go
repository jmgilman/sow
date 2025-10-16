package cmd

import (
	"github.com/spf13/cobra"
)

// NewRefsCmd creates the refs command.
func NewRefsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refs",
		Short: "Manage external knowledge and code references",
		Long: `Manage external references to knowledge and code repositories.

Refs provide access to external documentation, style guides, and code
examples. They are cached locally and symlinked into .sow/refs/ for
easy access by AI agents.

Type-specific commands:
  sow refs git    - Manage git repository references
  sow refs file   - Manage local file references

Each ref has a semantic type (knowledge or code) for categorization.`,
	}

	// Type-specific subcommand groups
	cmd.AddCommand(NewRefsGitCmd())
	cmd.AddCommand(NewRefsFileCmd())

	return cmd
}
