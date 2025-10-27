package exploration

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/spf13/cobra"
)

// NewJournalCmd creates the exploration journal command.
func NewJournalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "journal <message>",
		Short: "Add an entry to the exploration journal",
		Long: `Add an entry to the exploration's session journal for context tracking.

The journal helps with zero-context recovery by tracking decisions, insights,
questions, and conversation flow.

Valid types: decision, insight, question, note (default: note)

Requirements:
  - Must be in a sow repository with an active exploration

Examples:
  sow exploration journal "Decided to focus on branch isolation over performance"
  sow exploration journal "Found limitation: worktrees share reflog" --type insight
  sow exploration journal "How to handle cleanup on exploration deletion?" --type question`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJournal(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("type", "t", "note", "Entry type: decision, insight, question, note")

	return cmd
}

func runJournal(cmd *cobra.Command, args []string) error {
	message := args[0]
	entryType, _ := cmd.Flags().GetString("type")

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Add journal entry
	if err := exploration.AddJournalEntry(ctx, entryType, message); err != nil {
		if errors.Is(err, exploration.ErrNoExploration) {
			return fmt.Errorf("no active exploration - run 'sow explore <topic>' first")
		}
		return fmt.Errorf("failed to add journal entry: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Added journal entry (%s)\n\n", entryType)

	return nil
}
