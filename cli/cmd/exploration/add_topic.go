package exploration

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/spf13/cobra"
)

// NewAddTopicCmd creates the exploration add-topic command.
func NewAddTopicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-topic <topic>",
		Short: "Add a topic to the exploration's parking lot",
		Long: `Add a research topic to the exploration's parking lot for future exploration.

This helps track topics that were agreed upon but not yet explored, preventing
them from being forgotten during the session.

Requirements:
  - Must be in a sow repository with an active exploration
  - Topic must not already be in the parking lot

Example:
  sow exploration add-topic "Integration patterns for sow"
  sow exploration add-topic "Cleanup strategies"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddTopic(cmd, args)
		},
	}

	return cmd
}

func runAddTopic(cmd *cobra.Command, args []string) error {
	topic := args[0]

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Add topic to index
	if err := exploration.AddTopic(ctx, topic); err != nil {
		if errors.Is(err, exploration.ErrNoExploration) {
			return fmt.Errorf("no active exploration - run 'sow explore <topic>' first")
		}
		return fmt.Errorf("failed to add topic: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Added topic to parking lot: %s\n", topic)
	cmd.Printf("  Status: pending\n\n")

	return nil
}
