package issue

import (
	"fmt"
	"strconv"

	"github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <number>",
		Short: "Check if an issue has linked branches (claimed or available)",
		Long: `Check if an issue has linked branches to determine if it's available or claimed.

Issues with linked branches are considered "claimed" - someone is already
working on them. Issues without linked branches are "available" for claiming.

Examples:
  # Check if issue #123 is available
  sow issue check 123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			number, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid issue number: %s", args[0])
			}

			// Create GitHub client
			ghExec := exec.NewLocal("gh")
			gh := sow.NewGitHub(ghExec)

			// Get issue details
			issue, err := gh.GetIssue(number)
			if err != nil {
				return err
			}

			// Check for sow label
			if !issue.HasLabel("sow") {
				cmd.Printf("⚠️  Warning: Issue #%d does not have the 'sow' label.\n\n", number)
			}

			// Get linked branches
			branches, err := gh.GetLinkedBranches(number)
			if err != nil {
				return err
			}

			printCheckStatus(cmd, issue, branches)
			return nil
		},
	}

	return cmd
}

// printCheckStatus prints the issue check status.
func printCheckStatus(cmd *cobra.Command, issue *sow.Issue, branches []sow.LinkedBranch) {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Issue #%d: %s\n", issue.Number, issue.Title)
	fmt.Fprintf(out, "State: %s\n", issue.State)
	fmt.Fprintf(out, "URL: %s\n\n", issue.URL)

	if len(branches) == 0 {
		// Available
		fmt.Fprintf(out, "Status: ✓ Available\n")
		fmt.Fprintf(out, "No linked branches found. This issue is available for claiming.\n\n")
		fmt.Fprintf(out, "To create a project from this issue:\n")
		fmt.Fprintf(out, "  sow project init --issue %d\n", issue.Number)
	} else {
		// Claimed
		fmt.Fprintf(out, "Status: ✗ Claimed\n")
		fmt.Fprintf(out, "This issue has %d linked branch(es):\n\n", len(branches))

		for i, branch := range branches {
			fmt.Fprintf(out, "%d. Branch: %s\n", i+1, branch.Name)
			fmt.Fprintf(out, "   URL: %s\n", branch.URL)
		}

		fmt.Fprintf(out, "\nTo work on this project:\n")
		if len(branches) == 1 {
			fmt.Fprintf(out, "  git checkout %s && sow project status\n", branches[0].Name)
		} else {
			fmt.Fprintf(out, "  git checkout <branch-name> && sow project status\n")
		}
	}
}
