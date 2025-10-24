package breakdown

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewRemoveUnitCmd creates the breakdown remove-unit command.
func NewRemoveUnitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-unit <id>",
		Short: "Remove a work unit from the breakdown index",
		Long: `Remove a work unit from the current breakdown's index.

This will remove the work unit entry from the index but will NOT delete
any associated markdown document file.

Requirements:
  - Must be in a sow repository with an active breakdown session
  - Work unit must exist in the index

Examples:
  # Remove a work unit
  sow breakdown remove-unit unit-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemoveUnit(cmd, args)
		},
	}

	return cmd
}

func runRemoveUnit(cmd *cobra.Command, args []string) error {
	id := args[0]

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Remove work unit from index
	if err := breakdown.RemoveWorkUnit(ctx, id); err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return fmt.Errorf("no active breakdown session")
		}
		if errors.Is(err, breakdown.ErrWorkUnitNotFound) {
			return fmt.Errorf("work unit %s not found in breakdown index", id)
		}
		return fmt.Errorf("failed to remove work unit: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Removed work unit from breakdown index: %s\n", id)
	cmd.Printf("  Note: Associated markdown document (if any) was not deleted\n")

	return nil
}
