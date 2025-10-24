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
		Long: `Manage the exploration file registry for the current session.

The file registry tracks all files in the exploration workspace with
their descriptions and tags. This enables context-aware file loading and
discoverability.

Subcommands:
  add-file     Register a new file
  update-file  Update file metadata
  remove-file  Remove a file
  index        Display the current registry

Example workflow:
  1. Start exploration:
     sow explore "authentication-approaches"

  2. Create and register files:
     echo "# OAuth Research" > .sow/exploration/oauth.md
     sow exploration add-file oauth.md --description "OAuth 2.0 research" --tags "oauth,research"

  3. View index:
     sow exploration index`,
	}

	// Add subcommands
	cmd.AddCommand(NewAddFileCmd())
	cmd.AddCommand(NewUpdateFileCmd())
	cmd.AddCommand(NewRemoveFileCmd())
	cmd.AddCommand(NewIndexCmd())

	return cmd
}
