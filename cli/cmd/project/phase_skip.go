package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/statechart"
	"github.com/spf13/cobra"
)

// newPhaseSkipCmd creates the command to skip an optional phase.
//
// Usage:
//   sow project phase skip <phase>
//
// Only discovery and design phases can be skipped.
// Implementation, review, and finalize are always required.
func newPhaseSkipCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skip <phase>",
		Short: "Skip an optional phase (discovery or design)",
		Long: `Skip an optional phase.

Only discovery and design phases can be skipped.
Implementation, review, and finalize phases are always required and cannot be skipped.

Skipping a phase marks it as "skipped" and transitions to the next phase decision or planning state.

Example:
  # Skip discovery phase
  sow project phase skip discovery

  # Skip design phase
  sow project phase skip design`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]

			// Only discovery and design can be skipped
			if phase != "discovery" && phase != "design" {
				return fmt.Errorf("only discovery and design phases can be skipped")
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

			// Validate current state allows skipping phases
			currentState := machine.State()
			if phase == "discovery" && currentState != statechart.DiscoveryDecision {
				return fmt.Errorf("cannot skip discovery in current state: %s (expected DiscoveryDecision)", currentState)
			}
			if phase == "design" && currentState != statechart.DesignDecision {
				return fmt.Errorf("cannot skip design in current state: %s (expected DesignDecision)", currentState)
			}

			// Determine which event to fire
			var event statechart.Event
			if phase == "discovery" {
				state.Phases.Discovery.Enabled = false
				state.Phases.Discovery.Status = "skipped"
				event = statechart.EventSkipDiscovery
			} else { // design
				state.Phases.Design.Enabled = false
				state.Phases.Design.Status = "skipped"
				event = statechart.EventSkipDesign
			}

			// Fire event (validates transition, outputs prompt)
			if err := machine.Fire(event); err != nil {
				return fmt.Errorf("failed to skip %s phase: %w", phase, err)
			}

			// Save state
			if err := machine.Save(); err != nil {
				return fmt.Errorf("failed to save state: %w", err)
			}

			// === STATECHART INTEGRATION END ===

			cmd.Printf("\n✓ Skipped %s phase\n", phase)
			return nil
		},
	}

	return cmd
}
