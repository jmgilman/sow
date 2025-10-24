package breakdown

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewRemoveInputCmd creates the breakdown remove-input command.
func NewRemoveInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-input <path>",
		Short: "Remove an input from the breakdown index",
		Long: `Remove an input source from the current breakdown's index.

The path must match exactly the path used when adding the input.

Requirements:
  - Must be in a sow repository with an active breakdown session
  - Input must exist in the index

Examples:
  # Remove design document
  sow breakdown remove-input .sow/design/auth-architecture.md

  # Remove exploration directory
  sow breakdown remove-input .sow/exploration/`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemoveInput(cmd, args)
		},
	}

	return cmd
}

func runRemoveInput(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Remove input from index
	if err := breakdown.RemoveInput(ctx, path); err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return fmt.Errorf("no active breakdown session")
		}
		if errors.Is(err, breakdown.ErrInputNotFound) {
			return fmt.Errorf("input %s not found in breakdown index", path)
		}
		return fmt.Errorf("failed to remove input: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Removed input from breakdown index: %s\n", path)

	return nil
}
