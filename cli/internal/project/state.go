package project

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/project/types"
	_ "github.com/jmgilman/sow/cli/internal/project/types/standard" // Register StandardProject
	"github.com/jmgilman/sow/cli/internal/sow"
)

// Load loads an existing project from disk using the provided context.
// Returns sow.ErrNoProject if no project exists.
func Load(ctx *sow.Context) (*Project, error) {
	// Check if project exists
	exists, err := ctx.FS().Exists("project/state.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}
	if !exists {
		return nil, sow.ErrNoProject
	}

	// Load state from disk
	state, _, err := statechart.LoadProjectState(ctx.FS())
	if err != nil {
		return nil, fmt.Errorf("failed to load project state: %w", err)
	}

	// Detect project type and build state machine using composable phases architecture
	projectType, err := types.DetectProjectType(state)
	if err != nil {
		return nil, fmt.Errorf("failed to detect project type: %w", err)
	}

	// Build state machine from project type
	// The project type's BuildStateMachine() reads state.Statechart.Current_state
	// and creates the machine at the correct state automatically
	machine := projectType.BuildStateMachine()

	// Set filesystem for persistence
	machine.SetFilesystem(ctx.FS())

	return &Project{
		ctx:     ctx,
		machine: machine,
	}, nil
}

// Create creates a new project with the given name and description.
// Returns sow.ErrProjectExists if a project already exists.
func Create(ctx *sow.Context, name, description string) (*Project, error) {
	fs := ctx.FS()

	// Validate project name is kebab-case
	if !isKebabCase(name) {
		return nil, fmt.Errorf("project name must be kebab-case (lowercase letters, digits, and single hyphens only)")
	}

	// Validate description is not empty
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}

	// Get current git branch
	branch, err := ctx.Git().CurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Validate not on protected branch
	if ctx.Git().IsProtectedBranch(branch) {
		return nil, fmt.Errorf("cannot create project on protected branch '%s'", branch)
	}

	// Check if project already exists
	exists, _ := fs.Exists("project/state.yaml")
	if exists {
		// Try to load existing project to get its name
		if existing, err := Load(ctx); err == nil {
			existingName := existing.State().Project.Name
			return nil, fmt.Errorf("project '%s' already exists on branch '%s'", existingName, branch)
		}
		return nil, sow.ErrProjectExists
	}

	// Create project directory structure
	if err := fs.MkdirAll("project", 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	if err := fs.MkdirAll("project/context", 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", err)
	}

	// Create phase directories for required phases
	requiredPhases := []string{"implementation", "review", "finalize"}
	for _, phase := range requiredPhases {
		phaseDir := filepath.Join("project/phases", phase)
		if err := fs.MkdirAll(phaseDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s phase directory: %w", phase, err)
		}

		// Create log.md for each phase
		logPath := filepath.Join(phaseDir, "log.md")
		// Capitalize first letter of phase name
		phaseName := strings.ToUpper(phase[:1]) + phase[1:]
		logContent := []byte(fmt.Sprintf("# %s Phase Log\n\n", phaseName))
		if err := fs.WriteFile(logPath, logContent, 0644); err != nil {
			return nil, fmt.Errorf("failed to create %s log: %w", phase, err)
		}
	}

	// Create tasks directory for implementation
	tasksDir := filepath.Join("project/phases/implementation/tasks")
	if err := fs.MkdirAll(tasksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tasks directory: %w", err)
	}

	// Create review reports directory
	reportsDir := filepath.Join("project/phases/review/reports")
	if err := fs.MkdirAll(reportsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Create project log
	logPath := filepath.Join("project", "log.md")
	logContent := []byte("# Project Log\n\n")
	if err := fs.WriteFile(logPath, logContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to create project log: %w", err)
	}

	// Create initial project state with default values
	state := statechart.NewProjectState(name, description, branch)

	// Detect project type and build state machine using composable phases architecture
	projectType, err := types.DetectProjectType(state)
	if err != nil {
		return nil, fmt.Errorf("failed to detect project type: %w", err)
	}

	// Build state machine from project type
	machine := projectType.BuildStateMachine()

	// Set filesystem for persistence
	machine.SetFilesystem(fs)

	// Fire project init event to transition to DiscoveryDecision
	if err := machine.Fire(statechart.EventProjectInit); err != nil {
		return nil, fmt.Errorf("failed to initialize project state: %w", err)
	}

	// Save the initial state
	if err := machine.Save(); err != nil {
		return nil, fmt.Errorf("failed to save project state: %w", err)
	}

	return &Project{
		ctx:     ctx,
		machine: machine,
	}, nil
}

// CreateFromIssue creates a project linked to a GitHub issue.
// It validates the issue, checks for existing linked branches, creates a branch,
// initializes the project, and links it to the issue.
func CreateFromIssue(ctx *sow.Context, issueNumber int, branchName string) (*Project, error) {
	gh := ctx.GitHub()

	// Fetch issue
	issue, err := gh.GetIssue(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue #%d: %w", issueNumber, err)
	}

	// Validate issue has 'sow' label
	if !issue.HasLabel("sow") {
		return nil, fmt.Errorf("issue #%d does not have the 'sow' label", issueNumber)
	}

	// Check for existing linked branches
	branches, err := gh.GetLinkedBranches(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check linked branches: %w", err)
	}

	if len(branches) > 0 {
		return nil, fmt.Errorf("issue #%d already has a linked branch: %s\nTo work on this: git checkout %s && sow project status",
			issueNumber, branches[0].Name, branches[0].Name)
	}

	// Create branch via gh issue develop
	createdBranchName, err := gh.CreateLinkedBranch(issueNumber, branchName, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create linked branch: %w", err)
	}

	// Generate project name from issue
	// Use the branch name without the issue number prefix
	projectName := createdBranchName
	if idx := strings.Index(projectName, "-"); idx > 0 {
		projectName = projectName[idx+1:]
	}

	// Create project (will detect current branch automatically)
	project, err := Create(ctx, projectName, issue.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Set github_issue field
	state := project.State()
	issueNum64 := int64(issueNumber)
	state.Project.Github_issue = &issueNum64

	// Save the updated state
	if err := project.save(); err != nil {
		return nil, fmt.Errorf("failed to save github_issue link: %w", err)
	}

	return project, nil
}

// Delete deletes the active project.
// Returns sow.ErrNoProject if no project exists.
func Delete(ctx *sow.Context) error {
	fs := ctx.FS()

	// Check if project exists
	exists, _ := fs.Exists("project/state.yaml")
	if !exists {
		return sow.ErrNoProject
	}

	// Load the machine to update state before deletion
	machine, err := statechart.LoadFS(fs) //nolint:staticcheck // Deprecated but needed for backward compatibility
	if err != nil {
		return fmt.Errorf("failed to load project state: %w", err)
	}

	// Set project_deleted flag to true (required by state machine guard)
	state := machine.ProjectState()
	state.Phases.Finalize.Project_deleted = true

	// Save state with flag set
	if err := machine.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Fire transition
	if err := machine.Fire(statechart.EventProjectDelete); err != nil {
		return fmt.Errorf("failed to transition state: %w", err)
	}

	// Remove project directory
	if err := fs.RemoveAll("project"); err != nil {
		return fmt.Errorf("failed to remove project directory: %w", err)
	}

	return nil
}

// Exists checks if a project exists.
func Exists(ctx *sow.Context) bool {
	exists, _ := ctx.FS().Exists("project/state.yaml")
	return exists
}

// isKebabCase validates that a string is in kebab-case format.
func isKebabCase(s string) bool {
	if s == "" {
		return false
	}

	// Must start with lowercase letter or digit
	if (s[0] < 'a' || s[0] > 'z') && (s[0] < '0' || s[0] > '9') {
		return false
	}

	// Check each character
	prevHyphen := false
	for _, r := range s {
		if r == '-' {
			// No consecutive hyphens
			if prevHyphen {
				return false
			}
			prevHyphen = true
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			prevHyphen = false
		} else {
			return false
		}
	}

	// Must not end with hyphen
	return !prevHyphen
}
