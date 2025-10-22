// Package issue implements commands for managing GitHub issues as sow projects.
package issue

import (
	"github.com/spf13/cobra"
)

// NewIssueCmd creates the issue command with subcommands.
func NewIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage GitHub issues as sow projects",
		Long: `Manage GitHub issues as sow projects.

Issues with the 'sow' label represent potential sow projects. This command
provides tools to discover, check, and manage these issues.

GitHub CLI Integration:
  Requires the GitHub CLI (gh) to be installed and authenticated.
  Install: https://cli.github.com/
  Authenticate: gh auth login

Commands:
  list   - List issues with 'sow' label
  show   - Show details of a specific issue
  check  - Check if an issue has linked branches (claimed or available)`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newCheckCmd())

	return cmd
}
