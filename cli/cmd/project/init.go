package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/statechart"
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
		return fmt.Errorf("branch validation failed: %w", err)
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

	// === STATECHART INTEGRATION START ===

	// Load statechart (should be NoProject state)
	machine, err := statechart.Load()
	if err != nil {
		return fmt.Errorf("failed to load statechart: %w", err)
	}

	// Verify we're in NoProject state
	if machine.State() != statechart.NoProject {
		return fmt.Errorf("unexpected state: %s (expected NoProject)", machine.State())
	}

	// Create initial project state
	state := project.NewProjectState(name, branch, description)

	// Set state in machine
	machine.SetProjectState(state)

	// Create file structure BEFORE firing event (easier rollback)
	projectFS = sowFS.ProjectUnchecked()
	if err := projectFS.AppendLog("# Project Log\n\nOrchestrator actions will be logged here.\n"); err != nil {
		return fmt.Errorf("failed to create project log: %w", err)
	}

	// Fire event (validates transition, outputs prompt to stdout)
	if err := machine.Fire(statechart.EventProjectInit); err != nil {
		return fmt.Errorf("failed to transition to DiscoveryDecision: %w", err)
	}

	// Save state atomically (writes state.yaml with statechart field)
	if err := machine.Save(); err != nil {
		return fmt.Errorf("failed to save project state: %w", err)
	}

	// === STATECHART INTEGRATION END ===

	// Success message (prompt already output by statechart)
	cmd.Printf("\n✓ Initialized project '%s' on branch '%s'\n", name, branch)

	return nil
}
