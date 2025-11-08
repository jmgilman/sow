package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// debugLog prints debug messages to stderr when SOW_DEBUG=1 is set.
// This helps users and developers troubleshoot issues without modifying code.
//
// Format: [DEBUG] <component>: <message>
//
// Example:
//
//	debugLog("GitHub", "Fetched %d issues", len(issues))
//	// Output: [DEBUG] GitHub: Fetched 3 issues
func debugLog(component, format string, args ...interface{}) {
	if os.Getenv("SOW_DEBUG") == "1" {
		message := fmt.Sprintf(format, args...)
		fmt.Fprintf(os.Stderr, "[DEBUG] %s: %s\n", component, message)
	}
}

// ProjectTypeConfig defines the configuration for a project type.
type ProjectTypeConfig struct {
	Prefix      string
	Description string
}

// projectTypes maps project type names to their configuration.
// These are the four project types currently supported by sow.
var projectTypes = map[string]ProjectTypeConfig{
	"standard": {
		Prefix:      "feat/",
		Description: "Feature work and bug fixes",
	},
	"exploration": {
		Prefix:      "explore/",
		Description: "Research and investigation",
	},
	"design": {
		Prefix:      "design/",
		Description: "Architecture and design documents",
	},
	"breakdown": {
		Prefix:      "breakdown/",
		Description: "Decompose work into tasks",
	},
}

// normalizeName converts user-friendly project names into valid git branch names.
//
// The function applies the following transformations:
//  1. Trim leading/trailing whitespace
//  2. Convert to lowercase
//  3. Replace spaces with hyphens
//  4. Remove invalid characters (keep only: a-z, 0-9, -, _)
//  5. Collapse multiple consecutive hyphens into single hyphen
//  6. Remove leading/trailing hyphens
//  7. Return normalized name
//
// Example transformations:
//   - "Web Based Agents" → "web-based-agents"
//   - "API V2" → "api-v2"
//   - "feature--name" → "feature-name"
//   - "-leading-trailing-" → "leading-trailing"
//   - "With!Invalid@Chars#" → "withinvalidchars"
//   - "UPPERCASE" → "uppercase"
//   - "  spaces  " → "spaces"
//
// Edge cases:
//   - Empty string → "" (empty string)
//   - Only spaces → "" (empty string)
//   - Only special characters → "" (empty string)
//   - Unicode characters → removed (only ASCII alphanumeric allowed)
func normalizeName(name string) string {
	// 1. Trim leading/trailing whitespace
	name = strings.TrimSpace(name)

	// 2. Convert to lowercase
	name = strings.ToLower(name)

	// 3. Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// 4. Remove invalid characters (keep only: a-z, 0-9, -, _)
	// This regex matches anything that is NOT alphanumeric, hyphen, or underscore
	invalidCharsRegex := regexp.MustCompile(`[^a-z0-9\-_]+`)
	name = invalidCharsRegex.ReplaceAllString(name, "")

	// 5. Collapse multiple consecutive hyphens into single hyphen
	multipleHyphensRegex := regexp.MustCompile(`-+`)
	name = multipleHyphensRegex.ReplaceAllString(name, "-")

	// 6. Remove leading/trailing hyphens
	name = strings.Trim(name, "-")

	// 7. Return normalized name
	return name
}

// getTypePrefix returns the branch prefix for a given project type.
// If the project type is not recognized, returns "feat/" as the default.
//
// Examples:
//   - getTypePrefix("standard") → "feat/"
//   - getTypePrefix("exploration") → "explore/"
//   - getTypePrefix("unknown") → "feat/" (fallback)
//   - getTypePrefix("") → "feat/" (fallback)
func getTypePrefix(projectType string) string {
	if config, exists := projectTypes[projectType]; exists {
		return config.Prefix
	}
	return "feat/" // Default fallback
}

// getTypeOptions converts the projectTypes map into huh-compatible options
// for select prompts.
//
// The options are returned in a consistent order:
//  1. standard
//  2. exploration
//  3. design
//  4. breakdown
//  5. cancel
//
// Each option displays the type's description as the label and uses
// the type name as the value.
//
// Returns a slice of huh.Option[string] ready to use in a select prompt.
func getTypeOptions() []huh.Option[string] {
	// Return options in consistent order
	return []huh.Option[string]{
		huh.NewOption(projectTypes["standard"].Description, "standard"),
		huh.NewOption(projectTypes["exploration"].Description, "exploration"),
		huh.NewOption(projectTypes["design"].Description, "design"),
		huh.NewOption(projectTypes["breakdown"].Description, "breakdown"),
		huh.NewOption("Cancel", "cancel"),
	}
}

// previewBranchName shows what the branch name will be for a given project type and name.
// It combines the project type's prefix with the normalized name.
//
// Example:
//   - previewBranchName("standard", "Web Based Agents") → "feat/web-based-agents"
//   - previewBranchName("exploration", "API Research") → "explore/api-research"
func previewBranchName(projectType, name string) string {
	prefix := getTypePrefix(projectType)
	normalizedName := normalizeName(name)
	return prefix + normalizedName
}

// showError displays an error message to the user in a formatted way using huh forms.
// The user must press Enter to acknowledge the error.
//
// Returns nil after the error is shown (error is not propagated).
//
// Example usage:
//
//	if isProtectedBranch(branchName) {
//	    return showError("Cannot use protected branch name: " + branchName)
//	}
//
//nolint:unused,unparam // Will be used by wizard screens in subsequent work units
func showError(message string) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Error").
				Description(message),
		),
	)

	// Run the form (user presses Enter to acknowledge)
	_ = form.Run()

	// Return nil - error is shown, not propagated
	return nil
}

// withSpinner wraps a long-running operation with a loading spinner.
// The spinner displays the provided title while the action is running.
//
// If the action returns an error, it is propagated to the caller.
// If the action succeeds, nil is returned.
//
// Example usage:
//
//	var issues []*sow.Issue
//	err := withSpinner("Fetching GitHub issues...", func() error {
//	    var err error
//	    issues, err = sow.ListIssues(ctx)
//	    return err
//	})
//	if err != nil {
//	    return fmt.Errorf("failed to fetch issues: %w", err)
//	}
func withSpinner(title string, action func() error) error {
	var err error

	_ = spinner.New().
		Title(title).
		Action(func() {
			err = action()
		}).
		Run()

	return err
}

// isValidBranchName checks if a string is a valid git branch name.
// Returns nil if valid, error describing the problem if invalid.
//
// Git branch name rules:
// - Cannot start or end with /
// - Cannot contain ..
// - Cannot contain consecutive slashes //
// - Cannot end with .lock
// - Cannot contain special characters: ~, ^, :, ?, *, [, \.
// - Cannot contain whitespace.
func isValidBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check for invalid patterns
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
		return fmt.Errorf("branch name cannot start or end with /")
	}

	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name cannot contain double dots")
	}

	if strings.Contains(name, "//") {
		return fmt.Errorf("branch name cannot contain consecutive slashes")
	}

	if strings.HasSuffix(name, ".lock") {
		return fmt.Errorf("branch name cannot end with .lock")
	}

	// Check for invalid characters
	invalidChars := []string{"~", "^", ":", "?", "*", "[", "\\", " "}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("branch name contains invalid character: %s", char)
		}
	}

	return nil
}

// BranchState represents the state of a branch in the repository.
type BranchState struct {
	BranchExists   bool
	WorktreeExists bool
	ProjectExists  bool
}

// checkBranchState checks if a branch exists, has a worktree, and has an existing project.
func checkBranchState(ctx *sow.Context, branchName string) (*BranchState, error) {
	state := &BranchState{}

	// Check if branch exists
	branches, err := ctx.Git().Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	for _, branch := range branches {
		if branch == branchName {
			state.BranchExists = true
			break
		}
	}

	// Check if worktree exists
	worktreePath := sow.WorktreePath(ctx.RepoRoot(), branchName)
	if _, err := os.Stat(worktreePath); err == nil {
		state.WorktreeExists = true

		// If worktree exists, check for project
		projectStatePath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")
		if _, err := os.Stat(projectStatePath); err == nil {
			state.ProjectExists = true
		}
	}

	return state, nil
}

// ProjectInfo holds metadata about a project for display in the wizard.
type ProjectInfo struct {
	Branch         string    // Git branch name (e.g., "feat/auth")
	Name           string    // Project name from state.yaml
	Type           string    // Project type (standard, exploration, design, breakdown)
	Phase          string    // Current phase/state from state machine
	TasksCompleted int       // Number of completed tasks (0 if phase has no tasks)
	TasksTotal     int       // Total number of tasks (0 if phase has no tasks)
	ModTime        time.Time // State file modification time for sorting
}

// listProjects discovers all active projects by scanning the worktrees directory.
// Returns projects sorted by modification time (most recent first).
func listProjects(ctx *sow.Context) ([]ProjectInfo, error) {
	// Construct worktrees directory path using MainRepoRoot
	// This works whether we're in a worktree or the main repo
	worktreesDir := filepath.Join(ctx.MainRepoRoot(), ".sow", "worktrees")

	// Check if worktrees directory exists
	if _, err := os.Stat(worktreesDir); err != nil {
		if os.IsNotExist(err) {
			// Missing worktrees directory is not an error - just means no projects exist
			return []ProjectInfo{}, nil
		}
		return nil, fmt.Errorf("failed to stat worktrees directory: %w", err)
	}

	var projects []ProjectInfo

	// Walk the worktrees directory tree to find all state.yaml files
	err := filepath.Walk(worktreesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for state.yaml files in .sow/project/ subdirectories
		if info.IsDir() || info.Name() != "state.yaml" {
			return nil
		}

		// Check if this is a project state file (path ends with .sow/project/state.yaml)
		if filepath.Base(filepath.Dir(path)) != "project" ||
			filepath.Base(filepath.Dir(filepath.Dir(path))) != ".sow" {
			return nil
		}

		// Process this project state file
		projectInfo := processProjectState(path, worktreesDir, info)
		if projectInfo != nil {
			projects = append(projects, *projectInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk worktrees directory: %w", err)
	}

	// Sort by modification time, most recent first
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ModTime.After(projects[j].ModTime)
	})

	return projects, nil
}

// processProjectState processes a project state.yaml file and returns project info.
// Returns nil if the project cannot be loaded (with warning to stderr).
func processProjectState(path, worktreesDir string, info os.FileInfo) *ProjectInfo {
	// Extract branch name from path
	// Path structure: worktreesDir/branchName/.sow/project/state.yaml
	worktreePath := filepath.Dir(filepath.Dir(filepath.Dir(path)))
	branchName, err := filepath.Rel(worktreesDir, worktreePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to extract branch name from %s: %v\n", path, err)
		return nil
	}

	// Create a context for this worktree
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create context for %s: %v\n", branchName, err)
		return nil
	}

	// Load project state
	proj, err := state.Load(worktreeCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load project state for %s: %v\n", branchName, err)
		return nil
	}

	// Count tasks across ALL phases
	var tasksCompleted, tasksTotal int
	for _, phase := range proj.Phases {
		for _, task := range phase.Tasks {
			tasksTotal++
			if task.Status == "completed" {
				tasksCompleted++
			}
		}
	}

	// Extract metadata
	return &ProjectInfo{
		Branch:         branchName,
		Name:           proj.Name,
		Type:           proj.Type,
		Phase:          proj.Machine().State().String(),
		TasksCompleted: tasksCompleted,
		TasksTotal:     tasksTotal,
		ModTime:        info.ModTime(),
	}
}

// formatProjectProgress formats progress information for display in project selection.
// Returns a string like "Standard: implementation, 3/5 tasks completed" or "Design: active".
func formatProjectProgress(proj ProjectInfo) string {
	// Capitalize first letter of type
	typeName := strings.ToUpper(proj.Type[:1]) + proj.Type[1:]

	if proj.TasksTotal > 0 {
		return fmt.Sprintf("%s: %s, %d/%d tasks completed",
			typeName, proj.Phase, proj.TasksCompleted, proj.TasksTotal)
	}

	return fmt.Sprintf("%s: %s", typeName, proj.Phase)
}

// validateStateTransition checks if a state transition is valid.
// This helps catch logic errors during development and debugging.
//
// Returns nil if the transition is valid, error describing the problem if invalid.
func validateStateTransition(from, to WizardState) error {
	// Define valid transitions
	validTransitions := map[WizardState][]WizardState{
		StateEntry:          {StateCreateSource, StateProjectSelect, StateCancelled},
		StateCreateSource:   {StateIssueSelect, StateTypeSelect, StateCancelled},
		StateIssueSelect:    {StateTypeSelect, StateCreateSource, StateCancelled},
		StateTypeSelect:     {StateNameEntry, StatePromptEntry, StateCancelled},
		StateNameEntry:      {StatePromptEntry, StateTypeSelect, StateCancelled},
		StatePromptEntry:    {StateComplete, StateCancelled},
		StateProjectSelect:  {StateContinuePrompt, StateCancelled},
		StateContinuePrompt: {StateComplete, StateCancelled},
	}

	allowed, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("unknown source state: %s", from)
	}

	for _, validTo := range allowed {
		if validTo == to {
			return nil // Transition is valid
		}
	}

	return fmt.Errorf("invalid transition from %s to %s", from, to)
}
