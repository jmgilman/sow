package project

import (
	"github.com/spf13/cobra"
)

// NewProjectCmd creates the project command with subcommands.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
		Long: `Manage sow projects through subcommands.

Subcommands:
  wizard    Create or continue a project (interactive)
  new       Create a new project
  continue  Continue an existing project
  set       Set project field values
  delete    Delete the current project

Examples:
  sow project wizard
  sow project new --branch feat/auth "Add authentication"
  sow project continue
  sow project set description "Updated description"
  sow project set metadata.custom_field value
  sow project delete`,
	}

	cmd.AddCommand(newWizardCmd())
	cmd.AddCommand(newNewCmd())
	cmd.AddCommand(newContinueCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}
