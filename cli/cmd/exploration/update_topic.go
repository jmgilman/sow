package exploration

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/spf13/cobra"
)

// NewUpdateTopicCmd creates the exploration update-topic command.
func NewUpdateTopicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-topic <topic>",
		Short: "Update a topic's status in the parking lot",
		Long: `Update a topic's status and optionally associate files with it.

Valid statuses: pending, in_progress, completed

Requirements:
  - Must be in a sow repository with an active exploration
  - Topic must exist in the parking lot

Examples:
  sow exploration update-topic "Integration patterns" --status in_progress
  sow exploration update-topic "Integration patterns" --status completed --files "integration.md"
  sow exploration update-topic "Cleanup strategies" --status completed --files "cleanup.md,strategies.md"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateTopic(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("status", "s", "", "Topic status: pending, in_progress, completed (required)")
	cmd.Flags().StringSliceP("files", "f", nil, "Comma-separated list of related file paths")

	_ = cmd.MarkFlagRequired("status")

	return cmd
}

func runUpdateTopic(cmd *cobra.Command, args []string) error {
	topic := args[0]
	status, _ := cmd.Flags().GetString("status")
	files, _ := cmd.Flags().GetStringSlice("files")

	// Trim files
	for i, f := range files {
		files[i] = strings.TrimSpace(f)
	}

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Update topic
	if err := exploration.UpdateTopicStatus(ctx, topic, status, files); err != nil {
		if errors.Is(err, exploration.ErrNoExploration) {
			return fmt.Errorf("no active exploration - run 'sow explore <topic>' first")
		}
		return fmt.Errorf("failed to update topic: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Updated topic: %s\n", topic)
	cmd.Printf("  Status: %s\n", status)
	if len(files) > 0 {
		cmd.Printf("  Files: %s\n", strings.Join(files, ", "))
	}
	cmd.Printf("\n")

	return nil
}
