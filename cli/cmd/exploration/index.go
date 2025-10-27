package exploration

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/spf13/cobra"
)

// NewIndexCmd creates the exploration index command.
func NewIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Display the exploration index",
		Long: `Display the current exploration's index with all registered files.

Shows the exploration metadata (topic, branch, status) and all files
with their descriptions and tags.

Requirements:
  - Must be in a sow repository with an active exploration

Example:
  sow exploration index`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIndex(cmd, args)
		},
	}

	return cmd
}

func runIndex(cmd *cobra.Command, _ []string) error {
	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Load index
	index, err := exploration.LoadIndex(ctx)
	if err != nil {
		if errors.Is(err, exploration.ErrNoExploration) {
			return fmt.Errorf("no active exploration - run 'sow explore <topic>' first")
		}
		return fmt.Errorf("failed to load exploration index: %w", err)
	}

	// Display exploration metadata
	cmd.Printf("\nExploration: %s\n", index.Exploration.Topic)
	cmd.Printf("Branch:      %s\n", index.Exploration.Branch)
	cmd.Printf("Status:      %s\n", index.Exploration.Status)
	cmd.Printf("Created:     %s\n\n", index.Exploration.Created_at.Format("2006-01-02 15:04:05"))

	// Display topics
	if len(index.Topics) > 0 {
		cmd.Printf("Topics (%d):\n\n", len(index.Topics))
		for _, topic := range index.Topics {
			statusIcon := "○"
			if topic.Status == "in_progress" {
				statusIcon = "◐"
			} else if topic.Status == "completed" {
				statusIcon = "●"
			}
			cmd.Printf("  %s %s [%s]\n", statusIcon, topic.Topic, topic.Status)
			if len(topic.Related_files) > 0 {
				cmd.Printf("      Files: %s\n", strings.Join(topic.Related_files, ", "))
			}
		}
		cmd.Printf("\n")
	}

	// Display recent journal entries (last 5)
	if len(index.Journal) > 0 {
		displayCount := len(index.Journal)
		if displayCount > 5 {
			displayCount = 5
		}
		start := len(index.Journal) - displayCount

		cmd.Printf("Recent Journal (%d of %d):\n\n", displayCount, len(index.Journal))
		for _, entry := range index.Journal[start:] {
			cmd.Printf("  [%s] %s: %s\n",
				entry.Timestamp.Format("2006-01-02 15:04"),
				entry.Type,
				entry.Content)
		}
		cmd.Printf("\n")
	}

	// Display files
	if len(index.Files) == 0 {
		cmd.Printf("Files (0):\n\n")
		cmd.Printf("  No files registered yet.\n")
		cmd.Printf("  Add files with: sow exploration add-file <path> --description \"...\" --tags \"...\"\n\n")
		return nil
	}

	cmd.Printf("Files (%d):\n\n", len(index.Files))
	for i, file := range index.Files {
		cmd.Printf("  [%d] %s\n", i+1, file.Path)
		cmd.Printf("      Description: %s\n", file.Description)
		if len(file.Tags) > 0 {
			cmd.Printf("      Tags:        %s\n", strings.Join(file.Tags, ", "))
		}
		cmd.Printf("      Created:     %s\n", file.Created_at.Format("2006-01-02 15:04:05"))
		if i < len(index.Files)-1 {
			cmd.Printf("\n")
		}
	}
	cmd.Printf("\n")

	return nil
}
