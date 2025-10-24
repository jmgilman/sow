// Package breakdown provides commands for managing breakdown mode.
package breakdown

import (
	"github.com/spf13/cobra"
)

// NewBreakdownManagementCmd creates the breakdown management command.
func NewBreakdownManagementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "breakdown",
		Short: "Manage breakdown index",
		Long: `Manage the breakdown input/work unit registry for the current session.

The breakdown index tracks:
- Input sources that inform the breakdown (design docs, exploration artifacts)
- Work units to be created as GitHub issues

This enables context-aware loading and zero-context resumability across sessions.

Subcommands:
  add-input         Register a new input source
  remove-input      Remove an input
  add-unit          Add a proposed work unit
  update-unit       Update work unit metadata
  remove-unit       Remove a work unit
  create-document   Create detailed markdown for a work unit
  approve-unit      Approve a work unit for publishing
  publish           Publish work unit(s) as GitHub issues
  set-status        Update breakdown session status
  index             Display the current registry

Example workflow:
  1. Start breakdown:
     sow breakdown "auth-implementation"

  2. Register inputs:
     sow breakdown add-input .sow/design/auth-architecture.md \
       --type design --description "Auth system design" --tags "auth,design"

  3. Add work units:
     sow breakdown add-unit --id unit-001 --title "JWT token service" \
       --description "Implement JWT token generation and validation"

  4. Create detailed documents:
     sow breakdown create-document unit-001

  5. Approve and publish:
     sow breakdown approve-unit unit-001
     sow breakdown publish unit-001

  6. View index:
     sow breakdown index`,
	}

	// Add subcommands
	cmd.AddCommand(NewAddInputCmd())
	cmd.AddCommand(NewRemoveInputCmd())
	cmd.AddCommand(NewAddUnitCmd())
	cmd.AddCommand(NewUpdateUnitCmd())
	cmd.AddCommand(NewRemoveUnitCmd())
	cmd.AddCommand(NewCreateDocumentCmd())
	cmd.AddCommand(NewApproveUnitCmd())
	cmd.AddCommand(NewPublishCmd())
	cmd.AddCommand(NewSetStatusCmd())
	cmd.AddCommand(NewIndexCmd())

	return cmd
}
