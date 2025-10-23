package agent

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// NewSkipCmd creates the command to skip an optional phase.
//
// Usage:
//
//	sow agent skip <phase>
//
// Only discovery and design phases can be skipped (they are optional).
// Implementation, review, and finalize are always required and cannot be skipped.
func NewSkipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skip <phase>",
		Short: "Skip an optional phase",
		Long: `Skip an optional phase (discovery or design).

Only discovery and design phases can be skipped as they are optional.
Implementation, review, and finalize phases are always required and cannot be skipped.

Skipping a phase marks it as "skipped" and transitions to the next phase.

Example:
  # Skip discovery phase
  sow agent skip discovery

  # Skip design phase
  sow agent skip design`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]

			// Validate phase
			if phase != "discovery" && phase != "design" {
				return fmt.Errorf("only discovery and design phases can be skipped (got: %s)", phase)
			}

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent project init' first")
			}

			// Skip the phase
			if err := project.SkipPhase(phase); err != nil {
				return fmt.Errorf("failed to skip %s phase: %w", phase, err)
			}

			cmd.Printf("\n✓ Skipped %s phase\n", phase)
			return nil
		},
	}

	return cmd
}
