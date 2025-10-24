package design

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/design"
	"github.com/spf13/cobra"
)

// NewSetOutputTargetCmd creates the design set-output-target command.
func NewSetOutputTargetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-output-target <path> <target>",
		Short: "Update an output's target location",
		Long: `Update the target location for a specific output document.

This changes where the document will be moved when the design is finalized.

Arguments:
  path   - Output path (as registered in index)
  target - New target location

Example:
  sow design set-output-target adr-001-oauth.md .sow/knowledge/decisions/`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			target := args[1]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Update output target
			if err := design.UpdateOutputTarget(ctx, path, target); err != nil {
				if errors.Is(err, design.ErrNoDesign) {
					return fmt.Errorf("no active design session")
				}
				if errors.Is(err, design.ErrOutputNotFound) {
					return fmt.Errorf("output %s not found in design index", path)
				}
				return fmt.Errorf("failed to update output target: %w", err)
			}

			// Success
			cmd.Printf("\n✓ Updated target for %s: %s\n", path, target)

			return nil
		},
	}

	return cmd
}
