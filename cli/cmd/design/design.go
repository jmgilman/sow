// Package design provides commands for managing design mode.
package design

import (
	"github.com/spf13/cobra"
)

// NewDesignManagementCmd creates the design management command.
func NewDesignManagementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "design",
		Short: "Manage design index",
		Long: `Manage the design input/output registry for the current session.

The design index tracks:
- Input sources that inform the design (explorations, files, references)
- Output documents to be produced with their target locations

This enables context-aware loading and planning across sessions.

Subcommands:
  add-input         Register a new input source
  remove-input      Remove an input
  add-output        Register a planned output document
  remove-output     Remove an output
  set-output-target Update an output's target location
  set-status        Update design session status
  index             Display the current registry

Example workflow:
  1. Start design:
     sow design "authentication-system"

  2. Register inputs:
     sow design add-input .sow/exploration/ --type exploration \
       --description "OAuth research" --tags "oauth,research"

  3. Plan outputs:
     sow design add-output adr-001-oauth.md \
       --description "OAuth decision" \
       --target .sow/knowledge/adrs/ --type adr

  4. View index:
     sow design index`,
	}

	// Add subcommands
	cmd.AddCommand(NewAddInputCmd())
	cmd.AddCommand(NewRemoveInputCmd())
	cmd.AddCommand(NewAddOutputCmd())
	cmd.AddCommand(NewRemoveOutputCmd())
	cmd.AddCommand(NewSetOutputTargetCmd())
	cmd.AddCommand(NewSetStatusCmd())
	cmd.AddCommand(NewIndexCmd())

	return cmd
}
