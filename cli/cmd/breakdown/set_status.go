package breakdown

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewSetStatusCmd creates the breakdown set-status command.
func NewSetStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-status <status>",
		Short: "Update the breakdown session status",
		Long: `Update the status of the current breakdown session.

Valid statuses:
  active    - Currently working on the breakdown
  completed - All work units have been published
  abandoned - Breakdown session was abandoned

Requirements:
  - Must be in a sow repository with an active breakdown session

Examples:
  # Mark breakdown as active
  sow breakdown set-status active

  # Mark breakdown as completed
  sow breakdown set-status completed

  # Mark breakdown as abandoned
  sow breakdown set-status abandoned`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetStatus(cmd, args)
		},
	}

	return cmd
}

func runSetStatus(cmd *cobra.Command, args []string) error {
	status := args[0]

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Update status
	if err := breakdown.UpdateStatus(ctx, status); err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return fmt.Errorf("no active breakdown session")
		}
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Updated breakdown status: %s\n", status)

	return nil
}
