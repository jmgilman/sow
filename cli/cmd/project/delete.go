package project

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
  sow project delete
  sow project delete --force`,
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

	// Get Sow from context
	sow := sowFromContext(cmd.Context())

	// Get project to retrieve name for confirmation
	project, err := sow.GetProject()
	if err != nil {
		return fmt.Errorf("no active project found - nothing to delete")
	}

	projectName := project.Name()

	// Confirm deletion unless --force
	if !force {
		fmt.Fprintf(cmd.OutOrStdout(), "WARNING: This will permanently delete project '%s'\n", projectName)
		fmt.Fprintf(cmd.OutOrStdout(), "This action cannot be undone (though it remains in git history).\n\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Delete project? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Fprintln(cmd.OutOrStdout(), "Deletion cancelled.")
			return nil
		}
	}

	// Delete project (handles state machine transition and cleanup)
	if err := sow.DeleteProject(); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Print success message
	fmt.Fprintf(cmd.OutOrStderr(), "✓ State machine transitioned to NoProject\n")
	fmt.Fprintf(cmd.OutOrStderr(), "✓ Deleted project '%s'\n", projectName)
	fmt.Fprintf(cmd.OutOrStderr(), "\nNext steps:\n")
	fmt.Fprintf(cmd.OutOrStderr(), "  1. Commit the deletion: git add -A && git commit -m \"Delete project state\"\n")
	fmt.Fprintf(cmd.OutOrStderr(), "  2. Create PR or merge to main\n")

	return nil
}
