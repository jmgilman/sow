package project

import (
	"github.com/spf13/cobra"
)

// newPhaseCmd creates the phase command for managing project phases.
//
// Usage:
//   sow project phase <subcommand>
//
// Subcommands:
//   - enable: Enable an optional phase (discovery or design)
//   - status: Show phase status
//   - complete: Mark a phase as completed
func newPhaseCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "phase",
		Short: "Manage project phases",
		Long: `Manage project phases.

All projects have the same 5 phases:
  - Discovery (optional): Human-led exploration and requirements gathering
  - Design (optional): Architecture planning and design documentation
  - Implementation (required): AI-autonomous task execution
  - Review (required): AI-autonomous quality validation
  - Finalize (required): Documentation, cleanup, and PR creation

The implementation, review, and finalize phases are always enabled.
Only discovery and design can be optionally enabled.`,
	}

	// Add subcommands
	cmd.AddCommand(newPhaseEnableCmd(accessor))
	cmd.AddCommand(newPhaseStatusCmd(accessor))
	cmd.AddCommand(newPhaseCompleteCmd(accessor))

	return cmd
}
