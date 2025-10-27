package exploration

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/spf13/cobra"
)

// NewListTopicsCmd creates the exploration list-topics command.
func NewListTopicsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-topics",
		Short: "List all topics in the parking lot",
		Long: `List all research topics in the exploration's parking lot with their status.

Shows pending, in_progress, and completed topics.

Requirements:
  - Must be in a sow repository with an active exploration

Example:
  sow exploration list-topics`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListTopics(cmd)
		},
	}

	return cmd
}

func runListTopics(cmd *cobra.Command) error {
	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Load topics
	topics, err := exploration.ListTopics(ctx)
	if err != nil {
		if errors.Is(err, exploration.ErrNoExploration) {
			return fmt.Errorf("no active exploration - run 'sow explore <topic>' first")
		}
		return fmt.Errorf("failed to load topics: %w", err)
	}

	if len(topics) == 0 {
		cmd.Printf("\nNo topics in parking lot yet.\n\n")
		cmd.Printf("Add topics with: sow exploration add-topic \"<topic>\"\n")
		return nil
	}

	// Group by status
	pending := []string{}
	inProgress := []string{}
	completed := []string{}

	for _, t := range topics {
		switch t.Status {
		case "pending":
			pending = append(pending, t.Topic)
		case "in_progress":
			inProgress = append(inProgress, t.Topic)
		case "completed":
			desc := t.Topic
			if len(t.Related_files) > 0 {
				desc += fmt.Sprintf(" [%s]", strings.Join(t.Related_files, ", "))
			}
			completed = append(completed, desc)
		}
	}

	cmd.Printf("\nTopics Parking Lot (%d total):\n\n", len(topics))

	if len(pending) > 0 {
		cmd.Printf("Pending (%d):\n", len(pending))
		for _, topic := range pending {
			cmd.Printf("  • %s\n", topic)
		}
		cmd.Printf("\n")
	}

	if len(inProgress) > 0 {
		cmd.Printf("In Progress (%d):\n", len(inProgress))
		for _, topic := range inProgress {
			cmd.Printf("  • %s\n", topic)
		}
		cmd.Printf("\n")
	}

	if len(completed) > 0 {
		cmd.Printf("Completed (%d):\n", len(completed))
		for _, topic := range completed {
			cmd.Printf("  ✓ %s\n", topic)
		}
		cmd.Printf("\n")
	}

	return nil
}
