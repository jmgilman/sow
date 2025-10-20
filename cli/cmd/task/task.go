// Package task provides commands for managing tasks within a project.
package task

import (
	"context"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// sowFromContext retrieves the Sow instance from the command context.
// Panics if not found (should always be available via root command setup).
func sowFromContext(ctx context.Context) *sow.Sow {
	s, ok := ctx.Value("sow").(*sow.Sow)
	if !ok {
		panic("sow instance not found in context")
	}
	return s
}

// resolveTaskID resolves the task ID from args or infers it.
// If args contains a task ID, it's returned.
// Otherwise, the task ID is inferred from the project.
func resolveTaskID(project *sow.Project, args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	// Infer task ID
	return project.InferTaskID()
}

// NewTaskCmd creates the root task command.
func NewTaskCmd() *cobra.Command {
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
	cmd.AddCommand(NewAddCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(newStateCmd())
	cmd.AddCommand(newFeedbackCmd())

	return cmd
}
