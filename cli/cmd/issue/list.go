package issue

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmgilman/sow/libs/git"
)

func newListCmd() *cobra.Command {
	var state string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues with 'sow' label",
		Long: `List all issues with the 'sow' label.

These issues represent available sow projects. To check if a specific issue
has already been claimed, use 'sow issue check <number>'.

Examples:
  # List all open sow issues
  sow issue list

  # List all sow issues (open and closed)
  sow issue list --state all

  # List only closed sow issues
  sow issue list --state closed`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Create GitHub client
			gh, err := git.NewGitHubClient()
			if err != nil {
				return err
			}

			issues, err := gh.ListIssues("sow", state)
			if err != nil {
				return err
			}

			if len(issues) == 0 {
				cmd.Printf("No %s issues with 'sow' label found.\n", state)
				return nil
			}

			printIssuesTable(cmd, issues)
			return nil
		},
	}

	cmd.Flags().StringVar(&state, "state", "open", "Filter by state: open, closed, or all")

	return cmd
}

// printIssuesTable prints issues in a table format.
func printIssuesTable(cmd *cobra.Command, issues []git.Issue) {
	out := cmd.OutOrStdout()

	// Header
	_, _ = fmt.Fprintf(out, "%-8s %-6s %s\n", "NUMBER", "STATE", "TITLE")
	_, _ = fmt.Fprintf(out, "%-8s %-6s %s\n",
		strings.Repeat("─", 8),
		strings.Repeat("─", 6),
		strings.Repeat("─", 50),
	)

	// Rows
	for _, issue := range issues {
		_, _ = fmt.Fprintf(out, "%-8d %-6s %s\n", issue.Number, issue.State, issue.Title)
	}

	_, _ = fmt.Fprintf(out, "\nTotal: %d issue(s)\n", len(issues))
	_, _ = fmt.Fprintf(out, "Use 'sow issue check <number>' to see if an issue is available or claimed.\n")
}
