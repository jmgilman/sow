package project

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newPhaseSkipCmd creates the command to skip an optional phase.
//
// Usage:
//   sow project phase skip <phase>
//
// Only discovery and design phases can be skipped.
// Implementation, review, and finalize are always required.
func newPhaseSkipCmd() *cobra.Command {
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

			// Get Sow from context
			s := sowFromContext(cmd.Context())

			// Get project
			project, err := s.GetProject()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Skip phase (handles validation, state machine transitions)
			if err := project.SkipPhase(phase); err != nil {
				return err
			}

			cmd.Printf("\n✓ Skipped %s phase\n", phase)
			return nil
		},
	}

	return cmd
}
