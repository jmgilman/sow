package task

import (
	"github.com/spf13/cobra"
)

// newStateCmd creates the state command for managing task state.
//
// Usage:
//   sow task state <subcommand>
//
// Subcommands:
//   - increment: Increment the task iteration counter
//   - set-agent: Change the assigned agent
//   - add-reference: Add a context reference
//   - add-file: Track a modified file
func newStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage task state properties",
		Long: `Manage fine-grained task state properties.

These commands allow direct manipulation of task state for orchestrator
and worker coordination. Most users should use higher-level commands instead.

The state commands update the task's state.yaml file and provide precise
control over:
  - Iteration counter (for retry tracking)
  - Assigned agent (for agent reassignment)
  - Context references (for compilation)
  - Modified files (for change tracking)

Task ID inference:
  - If task ID is not provided, it will be inferred from:
    1. Current directory (if inside a task directory)
    2. Active task (if exactly one task has status "in_progress")`,
	}

	// Add subcommands
	cmd.AddCommand(newStateIncrementCmd())
	cmd.AddCommand(newStateSetAgentCmd())
	cmd.AddCommand(newStateAddReferenceCmd())
	cmd.AddCommand(newStateAddFileCmd())

	return cmd
}
