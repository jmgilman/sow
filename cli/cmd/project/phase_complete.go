package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/statechart"
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
				return fmt.Errorf("phase validation failed: %w", err)
			}

			// Get SowFS from context
			sowFS := accessor(cmd.Context())
			if sowFS == nil {
				return fmt.Errorf("not in a sow repository - run 'sow init' first")
			}

			// Verify project exists
			_, err := sowFS.Project()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first: %w", err)
			}

			// === STATECHART INTEGRATION START ===

			// Load machine
			machine, err := statechart.Load()
			if err != nil {
				return fmt.Errorf("failed to load statechart: %w", err)
			}

			state := machine.ProjectState()

			// Validate phase completion requirements
			if err := project.ValidatePhaseCompletion(state, phase); err != nil {
				return fmt.Errorf("cannot complete phase: %w", err)
			}

			// Determine which event to fire based on phase
			var event statechart.Event
			switch phase {
			case "discovery":
				event = statechart.EventCompleteDiscovery
			case "design":
				event = statechart.EventCompleteDesign
			default:
				return fmt.Errorf("phase '%s' completion is managed automatically by statechart", phase)
			}

			// Update phase status before firing event
			if err := project.CompletePhase(state, phase); err != nil {
				return fmt.Errorf("failed to update phase status: %w", err)
			}

			// Fire event (validates transition, outputs prompt)
			if err := machine.Fire(event); err != nil {
				return fmt.Errorf("failed to complete %s phase: %w", phase, err)
			}

			// Save state
			if err := machine.Save(); err != nil {
				return fmt.Errorf("failed to save state: %w", err)
			}

			// === STATECHART INTEGRATION END ===

			cmd.Printf("\n✓ Completed %s phase\n", phase)
			return nil
		},
	}

	return cmd
}
