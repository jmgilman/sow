package breakdown

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewApproveUnitCmd creates the breakdown approve-unit command.
func NewApproveUnitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve-unit <id>",
		Short: "Approve a work unit for publishing",
		Long: `Mark a work unit as approved for publishing to GitHub.

Once approved, the work unit can be published as a GitHub issue using
the 'sow breakdown publish' command.

The work unit's status will be updated to "approved".

Requirements:
  - Must be in a sow repository with an active breakdown session
  - Work unit must exist in the index
  - Work unit must not already be published

Examples:
  # Approve a work unit
  sow breakdown approve-unit unit-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApproveUnit(cmd, args)
		},
	}

	return cmd
}

func runApproveUnit(cmd *cobra.Command, args []string) error {
	id := args[0]

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Approve work unit
	if err := breakdown.ApproveWorkUnit(ctx, id); err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return fmt.Errorf("no active breakdown session")
		}
		if errors.Is(err, breakdown.ErrWorkUnitNotFound) {
			return fmt.Errorf("work unit %s not found in breakdown index", id)
		}
		if errors.Is(err, breakdown.ErrAlreadyPublished) {
			return fmt.Errorf("work unit %s is already published", id)
		}
		return fmt.Errorf("failed to approve work unit: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Approved work unit: %s\n", id)
	cmd.Printf("  Status: approved (ready for publishing)\n")
	cmd.Printf("\nTo publish this work unit as a GitHub issue:\n")
	cmd.Printf("  sow breakdown publish %s\n", id)

	return nil
}
