// Package project provides commands for managing sow projects.
package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete the current project",
		Long: `Delete the project directory (.sow/project).

This removes all project state, tasks, and artifacts. Use this when:
  - Project is complete and PR is merged
  - Project is abandoned
  - Starting over from scratch

Note: This does not delete the worktree or branch. Use git commands for that.

Example:
  sow project delete`,
		RunE: runDelete,
	}
}

func runDelete(cmd *cobra.Command, _ []string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if project exists
	projectDir := filepath.Join(ctx.RepoRoot(), ".sow", "project")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("no project exists")
	}

	// Delete project directory
	if err := os.RemoveAll(projectDir); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	fmt.Fprintln(os.Stderr, "✓ Deleted project")
	return nil
}
