package agent

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewSkipCmd creates the command to skip the active phase.
//
// Usage:
//
//	sow agent skip
//
// Skips the currently active phase and transitions to the next phase.
// Only optional phases (discovery, design) can be skipped.
func NewSkipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skip",
		Short: "Skip the active phase",
		Long: `Skip the currently active phase and transition to the next phase.

This command automatically detects which phase is currently active and skips it.
You don't need to specify the phase name - it's determined from the project state.

Only optional phases can be skipped:
  - discovery: Skip discovery and move to design decision
  - design: Skip design and move to implementation planning

Required phases cannot be skipped:
  - implementation: Required for all projects
  - review: Required for all projects
  - finalize: Required for all projects

Example:
  # Skip whichever phase is currently active
  sow agent skip`,
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

			// Skip the phase via Phase interface
			err = phase.Skip()
			if errors.Is(err, project.ErrNotSupported) {
				return fmt.Errorf("phase %s cannot be skipped (required phase)", phase.Name())
			}
			if err != nil {
				return fmt.Errorf("failed to skip phase: %w", err)
			}

			cmd.Printf("\n✓ Skipped %s phase\n", phase.Name())
			return nil
		},
	}

	return cmd
}
