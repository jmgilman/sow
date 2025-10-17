package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newPhaseCompleteCmd creates the command to mark a phase as completed.
//
// Usage:
//   sow project phase complete <phase>
//
// Validates that the phase meets all completion requirements before
// marking it as completed.
func newPhaseCompleteCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete <phase>",
		Short: "Mark a phase as completed",
		Long: `Mark a phase as completed.

Each phase has specific completion requirements:
  - Discovery: All artifacts must be approved
  - Design: All artifacts must be approved
  - Implementation: All tasks must be completed or abandoned
  - Review: Latest review report must have assessment "pass"
  - Finalize: project_deleted must be true

The command validates these requirements before marking the phase complete.

Example:
  # Complete the discovery phase
  sow project phase complete discovery

  # Complete the implementation phase
  sow project phase complete implementation`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]

			// Validate phase name
			if err := project.ValidatePhase(phase); err != nil {
				return err
			}

			// Get SowFS from context
			sowFS := accessor(cmd.Context())
			if sowFS == nil {
				return fmt.Errorf("not in a sow repository - run 'sow init' first")
			}

			// Get project filesystem
			projectFS, err := sowFS.Project()
			if err != nil {
				return err
			}

			// Read current state
			state, err := projectFS.State()
			if err != nil {
				return fmt.Errorf("failed to read project state: %w", err)
			}

			// Complete the phase (validates and updates state)
			if err := project.CompletePhase(state, phase); err != nil {
				return err
			}

			// Write updated state
			if err := projectFS.WriteState(state); err != nil {
				return fmt.Errorf("failed to write project state: %w", err)
			}

			cmd.Printf("✓ Completed %s phase for project '%s'\n", phase, state.Project.Name)
			return nil
		},
	}

	return cmd
}
