package project

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// NewDeleteCmd creates the project delete command.
func NewDeleteCmd(accessor SowFSAccessor) *cobra.Command {
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
			return runDelete(cmd, args, accessor)
		},
	}

	// Flags
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(cmd *cobra.Command, _ []string, accessor SowFSAccessor) error {
	force, _ := cmd.Flags().GetBool("force")

	// Get SowFS from context
	sowFS := accessor(cmd.Context())
	if sowFS == nil {
		return fmt.Errorf("not in a sow repository - run 'sow init' first")
	}

	// Get ProjectFS (will error if no project exists)
	projectFS, err := sowFS.Project()
	if err != nil {
		return fmt.Errorf("no active project found - nothing to delete")
	}

	// Read project state to get project name
	state, err := projectFS.State()
	if err != nil {
		return fmt.Errorf("failed to read project state: %w", err)
	}

	// Confirm deletion unless --force
	if !force {
		cmd.Printf("WARNING: This will permanently delete project '%s'\n", state.Project.Name)
		cmd.Printf("This action cannot be undone (though it remains in git history).\n\n")
		cmd.Printf("Delete project? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			cmd.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete the project directory
	if err := projectFS.Delete(); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Print success message
	cmd.Printf("✓ Deleted project '%s'\n", state.Project.Name)
	cmd.Printf("\nNext steps:\n")
	cmd.Printf("  1. Commit the deletion: git add -A && git commit -m \"Delete project state\"\n")
	cmd.Printf("  2. Create PR or merge to main\n")

	return nil
}
