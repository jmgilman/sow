package agent

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewEnableCmd creates the command to enable a specific phase.
//
// Usage:
//
//	sow agent enable <phase-name>
//
// Enables an optional phase that would otherwise be skipped.
func NewEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <phase-name>",
		Short: "Enable an optional phase",
		Long: `Enable an optional phase that would otherwise be skipped.

This command enables optional phases in the project workflow:
  - discovery: Enable discovery phase for research and planning
  - design: Enable design phase for architecture and design work

Required phases are always enabled and cannot be explicitly enabled:
  - implementation: Always enabled
  - review: Always enabled
  - finalize: Always enabled

The phase will be activated and its status set to 'in_progress'.

Example:
  # Enable discovery phase
  sow agent enable discovery

  # Enable design phase
  sow agent enable design`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phaseName := args[0]

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

			// Get the phase by name
			phase, err := proj.Phase(phaseName)
			if err != nil {
				if errors.Is(err, project.ErrPhaseNotFound) {
					return fmt.Errorf("phase %s not found - available phases depend on project type", phaseName)
				}
				return fmt.Errorf("failed to get phase: %w", err)
			}

			// Enable the phase via Phase interface
			err = phase.Enable()
			if errors.Is(err, project.ErrNotSupported) {
				return fmt.Errorf("phase %s cannot be enabled (always enabled or not applicable)", phaseName)
			}
			if err != nil {
				return fmt.Errorf("failed to enable phase: %w", err)
			}

			cmd.Printf("\n✓ Enabled %s phase\n", phaseName)
			return nil
		},
	}

	return cmd
}
