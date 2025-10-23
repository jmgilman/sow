package agent

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewDeleteCmd creates the project delete command.
func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete the project directory",
		Long: `Delete the entire .sow/project/ directory.

This command is typically run during the finalize phase before creating a PR.
The sow design requires no project files to be present in merged code, so the
project directory must be deleted before the PR is merged to main.

By default, prompts for confirmation. Use --force to skip confirmation.

Warning: This action cannot be undone. The project state will be permanently
deleted from the working directory (though it remains in git history).

Example:
  sow agent delete
  sow agent delete --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, args)
		},
	}

	// Flags
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(cmd *cobra.Command, _ []string) error {
	force, _ := cmd.Flags().GetBool("force")

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Load project to retrieve name for confirmation
	proj, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("no active project found - nothing to delete")
	}

	projectName := proj.Name()

	// Confirm deletion unless --force
	if !force {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "WARNING: This will permanently delete project '%s'\n", projectName)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "This action cannot be undone (though it remains in git history).\n\n")
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Delete project? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Deletion cancelled.")
			return nil
		}
	}

	// Delete project (handles state machine transition and cleanup)
	if err := loader.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Print success message
	_, _ = fmt.Fprintf(cmd.OutOrStderr(), "✓ State machine transitioned to NoProject\n")
	_, _ = fmt.Fprintf(cmd.OutOrStderr(), "✓ Deleted project '%s'\n", projectName)
	_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\nNext steps:\n")
	_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  1. Commit the deletion: git add -A && git commit -m \"Delete project state\"\n")
	_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  2. Create PR or merge to main\n")

	return nil
}
