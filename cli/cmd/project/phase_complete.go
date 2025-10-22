package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newPhaseCompleteCmd creates the command to mark a phase as completed.
//
// Usage:
//   sow project phase complete <phase>
//
// Validates that the phase meets all completion requirements before
// marking it as completed.
func newPhaseCompleteCmd() *cobra.Command {
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

			// Get Sow from context
			ctx := cmdutil.GetContext(cmd.Context())

			// Get project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Complete phase (handles validation, state machine transitions)
			if err := project.CompletePhase(phase); err != nil {
				return err
			}

			cmd.Printf("\n✓ Completed %s phase\n", phase)
			return nil
		},
	}

	return cmd
}
