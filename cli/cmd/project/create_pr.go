package project

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// newCreatePRCmd creates the command to create a pull request.
func newCreatePRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pr",
		Short: "Create a pull request for the project",
		Long: `Create a pull request for the project using GitHub CLI.

The PR body should be provided via --body flag or stdin. The command will:
  - Generate PR title from project name and description
  - Add "Closes #<number>" if project linked to GitHub issue
  - Add sow footer
  - Create PR via gh CLI
  - Store PR URL in project state

The orchestrator should write a summary of the changes and provide it as the body.

Examples:
  # Provide body via flag
  sow project create-pr --body "## Summary\n\nImplemented authentication system..."

  # Provide body via stdin
  echo "## Summary\n\nImplemented authentication..." | sow project create-pr

  # Provide body from file
  cat pr-description.md | sow project create-pr

Prerequisites:
  - GitHub CLI (gh) installed and authenticated
  - Current branch pushed to remote
  - Project in finalize phase`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreatePR(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("body", "b", "", "PR body (use - to read from stdin)")

	return cmd
}

func runCreatePR(cmd *cobra.Command, _ []string) error {
	bodyFlag, _ := cmd.Flags().GetString("body")

	// Get Sow from context
	sowInstance := cmdutil.SowFromContext(cmd.Context())

	// Get project
	project, err := sowInstance.GetProject()
	if err != nil {
		return fmt.Errorf("no active project found")
	}

	// Get PR body
	var body string
	if bodyFlag == "" || bodyFlag == "-" {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			// Interactive mode - prompt user
			_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Enter PR description (end with Ctrl+D):")
		}

		reader := bufio.NewReader(os.Stdin)
		var lines []string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if line != "" {
						lines = append(lines, line)
					}
					break
				}
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			lines = append(lines, line)
		}
		body = strings.Join(lines, "")
	} else {
		body = bodyFlag
	}

	// Validate body is not empty
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("PR body cannot be empty")
	}

	// Create PR
	prURL, err := project.CreatePullRequest(body)
	if err != nil {
		return err
	}

	// Success message
	_, _ = fmt.Fprintf(cmd.OutOrStderr(), "✓ Pull request created: %s\n", prURL)

	return nil
}
