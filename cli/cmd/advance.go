package cmd

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewAdvanceCmd creates the top-level command to advance project state.
//
// Usage:
//
//	sow advance
//
// This command examines the current phase and state, determines the appropriate
// transition event, validates prerequisites via guards, and advances the state
// machine if all conditions are met.
func NewAdvanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advance",
		Short: "Advance project to next state",
		Long: `Advance the current phase to its next state.

This command examines the current phase and state, determines the appropriate
transition event, validates prerequisites via guards, and advances the state
machine if all conditions are met.

The command will fail if:
  - No project exists
  - The current phase is in an unexpected state
  - Prerequisites for advancement are not met (guard failure)
  - State cannot be saved

Example:
  sow advance`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project via loader to get interface
			proj, err := loader.Load(ctx)
			if err != nil {
				if errors.Is(err, project.ErrNoProject) {
					return fmt.Errorf("no active project - run 'sow agent init' first")
				}
				return fmt.Errorf("failed to load project: %w", err)
			}

			// Get current phase
			phase := proj.CurrentPhase()
			if phase == nil {
				return fmt.Errorf("no active phase found - project may be complete")
			}

			// Advance the phase
			// Phase.Advance() now handles event firing and state saving internally
			if err := phase.Advance(); err != nil {
				// Provide specific error messages for different error types
				if errors.Is(err, project.ErrNotSupported) {
					return fmt.Errorf("phase %s does not support state advancement", phase.Name())
				}
				if errors.Is(err, project.ErrCannotAdvance) {
					return fmt.Errorf("cannot advance: %w\n\nEnsure all prerequisites are met before advancing", err)
				}
				if errors.Is(err, project.ErrUnexpectedState) {
					return fmt.Errorf("unexpected state detected: %w\n\nThis may indicate state file corruption", err)
				}
				return fmt.Errorf("failed to advance: %w", err)
			}

			cmd.Println("✓ Advanced to next state")
			return nil
		},
	}

	return cmd
}
