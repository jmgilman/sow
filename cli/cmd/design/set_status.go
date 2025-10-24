package design

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/design"
	"github.com/spf13/cobra"
)

// NewSetStatusCmd creates the design set-status command.
func NewSetStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-status <status>",
		Short: "Update design session status",
		Long: `Update the status of the current design session.

Valid statuses:
  active    - Design work is in progress
  in_review - Design is ready for review
  completed - Design work is complete

Examples:
  sow design set-status in_review
  sow design set-status completed`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			status := args[0]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Update status
			if err := design.UpdateStatus(ctx, status); err != nil {
				if errors.Is(err, design.ErrNoDesign) {
					return fmt.Errorf("no active design session")
				}
				return fmt.Errorf("failed to update status: %w", err)
			}

			// Success
			cmd.Printf("\n✓ Updated design status: %s\n", status)

			return nil
		},
	}

	return cmd
}
