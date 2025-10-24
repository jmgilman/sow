package design

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/design"
	"github.com/spf13/cobra"
)

// NewRemoveInputCmd creates the design remove-input command.
func NewRemoveInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-input <path>",
		Short: "Remove an input from the design index",
		Long: `Remove an input source from the current design's index.

The path must match exactly as it was registered.

Example:
  sow design remove-input .sow/exploration/oauth.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Remove input from index
			if err := design.RemoveInput(ctx, path); err != nil {
				if errors.Is(err, design.ErrNoDesign) {
					return fmt.Errorf("no active design session")
				}
				if errors.Is(err, design.ErrInputNotFound) {
					return fmt.Errorf("input %s not found in design index", path)
				}
				return fmt.Errorf("failed to remove input: %w", err)
			}

			// Success
			cmd.Printf("\n✓ Removed input from design index: %s\n", path)

			return nil
		},
	}

	return cmd
}
