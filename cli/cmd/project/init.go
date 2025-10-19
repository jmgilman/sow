package project

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewInitCmd creates the project init command.
func NewInitCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Initialize a new project",
		Long: `Initialize a new project on the current branch.

Creates the initial project state with all 5 phases. By default, only the
required phases (Implementation, Review, Finalize) are enabled. The truth
table in /project:new will later modify phase enablement based on requirements.

Requirements:
  - Must be in a sow repository (.sow directory exists)
  - Must be on a feature branch (not main or master)
  - No existing project can be present

Example:
  sow project init my-feature --description "Add authentication feature"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("description", "d", "", "Project description (required)")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	name := args[0]
	description, _ := cmd.Flags().GetString("description")

	// Get Sow from context
	sow := sowFromContext(cmd.Context())

	// Create project (handles all validation, state machine transitions, file creation)
	project, err := sow.CreateProject(name, description)
	if err != nil {
		return err
	}

	// Success message (statechart prompt already output)
	fmt.Fprintf(cmd.OutOrStderr(), "\n✓ Initialized project '%s' on branch '%s'\n",
		project.Name(), project.Branch())

	return nil
}
