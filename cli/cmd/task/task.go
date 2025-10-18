// Package task provides commands for managing tasks within a project.
package task

import (
	"context"

	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/spf13/cobra"
)

// SowFSAccessor is a function type that retrieves SowFS from context.
// This allows commands to be tested with different SowFS implementations.
type SowFSAccessor func(ctx context.Context) sowfs.SowFS

// NewTaskCmd creates the root task command.
func NewTaskCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage project tasks",
		Long: `Manage tasks within the implementation phase of a project.

The task commands provide access to task creation, status checking, updates,
and listing. Tasks are the atomic units of work within the implementation phase.

Tasks follow gap numbering (010, 020, 030...) to allow insertion of new tasks
between existing ones if needed. Each task has:
  - A unique ID (gap-numbered: 010, 020, 030...)
  - A status (pending, in_progress, completed, abandoned)
  - An assigned agent type (e.g., "implementer", "reviewer")
  - Dependencies on other tasks (optional)
  - Iteration tracking and feedback history`,
	}

	// Add subcommands
	cmd.AddCommand(NewAddCmd(accessor))
	cmd.AddCommand(NewListCmd(accessor))
	cmd.AddCommand(NewStatusCmd(accessor))
	cmd.AddCommand(NewUpdateCmd(accessor))
	cmd.AddCommand(newStateCmd(accessor))
	cmd.AddCommand(newFeedbackCmd(accessor))

	return cmd
}
