// Package exploration provides commands for managing exploration mode.
package exploration

import (
	"github.com/spf13/cobra"
)

// NewExplorationCmd creates the exploration command.
func NewExplorationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exploration",
		Short: "Manage exploration index",
		Long: `Manage the exploration file registry, topics, and journal for the current session.

The exploration index tracks:
- Files with descriptions and tags for discoverability
- Topics "parking lot" for agreed-upon research areas
- Journal for session context and zero-context recovery

Subcommands:
  File Management:
    add-file     Register a new file
    update-file  Update file metadata
    remove-file  Remove a file

  Topic Management:
    add-topic     Add topic to parking lot
    update-topic  Update topic status
    list-topics   List all topics

  Session Journal:
    journal      Add journal entry

  Display:
    index        Display complete index

Example workflow:
  1. Start exploration:
     sow explore "authentication-approaches"

  2. Add topics to parking lot:
     sow exploration add-topic "OAuth vs JWT comparison"
     sow exploration add-topic "Session management strategies"

  3. Create and register files:
     echo "# OAuth Research" > .sow/exploration/oauth.md
     sow exploration add-file oauth.md --description "OAuth 2.0 research" --tags "oauth,research"

  4. Track progress:
     sow exploration update-topic "OAuth vs JWT comparison" --status completed --files "oauth.md"
     sow exploration journal "Decided to focus on OAuth for API security"

  5. View index:
     sow exploration index`,
	}

	// Add subcommands
	// File management
	cmd.AddCommand(NewAddFileCmd())
	cmd.AddCommand(NewUpdateFileCmd())
	cmd.AddCommand(NewRemoveFileCmd())

	// Topic management
	cmd.AddCommand(NewAddTopicCmd())
	cmd.AddCommand(NewUpdateTopicCmd())
	cmd.AddCommand(NewListTopicsCmd())

	// Journal
	cmd.AddCommand(NewJournalCmd())

	// Display
	cmd.AddCommand(NewIndexCmd())

	return cmd
}
