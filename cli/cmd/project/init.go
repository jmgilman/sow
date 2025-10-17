package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
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
			return runInit(cmd, args, accessor)
		},
	}

	// Flags
	cmd.Flags().StringP("description", "d", "", "Project description (required)")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runInit(cmd *cobra.Command, args []string, accessor SowFSAccessor) error {
	name := args[0]
	description, _ := cmd.Flags().GetString("description")

	// Get SowFS from context
	sowFS := accessor(cmd.Context())
	if sowFS == nil {
		return fmt.Errorf("not in a sow repository - run 'sow init' first")
	}

	// Get git repository to determine current branch
	repo, err := sowFS.Repo()
	if err != nil {
		return fmt.Errorf("failed to access git repository: %w", err)
	}

	// Get current branch
	branch, err := repo.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Validate branch is not protected (main/master)
	if err := project.ValidateBranch(branch); err != nil {
		return err
	}

	// Check if project already exists
	projectFS, err := sowFS.Project()
	if err == nil {
		// Project exists - check if it's on the same branch
		state, err := projectFS.State()
		if err != nil {
			return fmt.Errorf("failed to read existing project state: %w", err)
		}
		return fmt.Errorf("project '%s' already exists on branch '%s'", state.Project.Name, state.Project.Branch)
	}

	// Create initial project state
	state := project.NewProjectState(name, branch, description)

	// Get ProjectFS without existence check (since we're creating it)
	projectFS = sowFS.ProjectUnchecked()

	// Write state (this will create the project directory structure)
	if err := projectFS.WriteState(state); err != nil {
		return fmt.Errorf("failed to write project state: %w", err)
	}

	// Create empty log file
	if err := projectFS.AppendLog("# Project Log\n\nOrchestrator actions will be logged here.\n"); err != nil {
		return fmt.Errorf("failed to create project log: %w", err)
	}

	// Print success message
	cmd.Printf("✓ Initialized project '%s' on branch '%s'\n", name, branch)
	cmd.Printf("\nNext steps:\n")
	cmd.Printf("  1. Run '/project:new' to start the project workflow\n")
	cmd.Printf("  2. The orchestrator will guide you through phase selection\n")

	return nil
}
