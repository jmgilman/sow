package design

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/design"
	"github.com/spf13/cobra"
)

// NewRemoveOutputCmd creates the design remove-output command.
func NewRemoveOutputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-output <path>",
		Short: "Remove an output from the design index",
		Long: `Remove a planned output document from the current design's index.

The path must match exactly as it was registered (relative to .sow/design/).

Example:
  sow design remove-output adr-001-oauth-decision.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Remove output from index
			if err := design.RemoveOutput(ctx, path); err != nil {
				if errors.Is(err, design.ErrNoDesign) {
					return fmt.Errorf("no active design session")
				}
				if errors.Is(err, design.ErrOutputNotFound) {
					return fmt.Errorf("output %s not found in design index", path)
				}
				return fmt.Errorf("failed to remove output: %w", err)
			}

			// Success
			cmd.Printf("\n✓ Removed output from design index: %s\n", path)

			return nil
		},
	}

	return cmd
}
