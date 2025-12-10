package issue

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/libs/exec"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <number>",
		Short: "Show details of a specific issue",
		Long: `Show detailed information about a GitHub issue.

Displays the issue title, state, labels, and body. Useful for reviewing
an issue before creating a project from it.

Examples:
  # Show issue #123
  sow issue show 123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			number, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid issue number: %s", args[0])
			}

			// Create GitHub client
			ghExec := exec.NewLocalExecutor("gh")
			gh := sow.NewGitHubCLI(ghExec)

			issue, err := gh.GetIssue(number)
			if err != nil {
				return err
			}

			printIssueDetails(cmd, issue)
			return nil
		},
	}

	return cmd
}

// printIssueDetails prints detailed issue information.
func printIssueDetails(cmd *cobra.Command, issue *sow.Issue) {
	out := cmd.OutOrStdout()

	// Header
	_, _ = fmt.Fprintf(out, "Issue #%d: %s\n", issue.Number, issue.Title)
	_, _ = fmt.Fprintf(out, "%s\n\n", strings.Repeat("=", 60))

	// State
	_, _ = fmt.Fprintf(out, "State: %s\n", issue.State)

	// Labels
	var labels []string
	for _, l := range issue.Labels {
		labels = append(labels, l.Name)
	}
	_, _ = fmt.Fprintf(out, "Labels: %s\n", strings.Join(labels, ", "))

	// URL
	_, _ = fmt.Fprintf(out, "URL: %s\n\n", issue.URL)

	// Body
	if issue.Body != "" {
		_, _ = fmt.Fprintf(out, "Description:\n")
		_, _ = fmt.Fprintf(out, "%s\n", strings.Repeat("-", 60))
		_, _ = fmt.Fprintf(out, "%s\n", issue.Body)
	} else {
		_, _ = fmt.Fprintf(out, "Description: (none)\n")
	}

	// Check for sow label
	if !issue.HasLabel("sow") {
		_, _ = fmt.Fprintf(out, "\n⚠️  Warning: This issue does not have the 'sow' label.\n")
		_, _ = fmt.Fprintf(out, "   Add it via: gh issue edit %d --add-label sow\n", issue.Number)
	}
}
