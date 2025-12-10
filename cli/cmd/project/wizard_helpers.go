package project

import (
	"errors"
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
	"github.com/jmgilman/sow/libs/git"
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
// In test mode (SOW_TEST=1), this is a no-op to prevent tests from hanging.
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
	// Skip interactive prompts in test mode
	if os.Getenv("SOW_TEST") == "1" {
		debugLog("Error", "%s", message)
		return nil
	}

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
// In test mode (SOW_TEST=1), executes action directly without spinner.
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
	// Skip spinner in test mode - just run action directly
	if os.Getenv("SOW_TEST") == "1" {
		debugLog("Spinner", "%s", title)
		return action()
	}

	var err error

	_ = spinner.New().
		Title(title).
		Action(func() {
			err = action()
		}).
		Run()

	return err
}

// isProtectedBranch checks if a branch name is protected (main or master).
// Convenience wrapper around the logic used in git.IsProtectedBranch().
//
// Protected branches cannot have sow projects created on them to avoid
// accidental commits to the main development line.
//
// Example:
//
//	if isProtectedBranch("main") {
//	    return fmt.Errorf("cannot use protected branch")
//	}
func isProtectedBranch(name string) bool {
	return name == "main" || name == "master"
}

// isValidBranchName checks if a string is a valid git branch name.
// Returns nil if valid, error describing the problem if invalid.
//
// Git branch name rules:
// - Not empty or whitespace-only
// - Not a protected branch (main, master)
// - Cannot start or end with /
// - Cannot contain ..
// - Cannot contain consecutive slashes //
// - Cannot end with .lock
// - Cannot contain special characters: ~, ^, :, ?, *, [, \.
// - Cannot contain whitespace.
//
// Example:
//
//	err := isValidBranchName("feat/add-auth")  // nil (valid)
//	err := isValidBranchName("main")           // error (protected)
//	err := isValidBranchName("has spaces")     // error (spaces)
func isValidBranchName(name string) error {
	// Trim and check empty
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check protected branches
	if isProtectedBranch(name) {
		return fmt.Errorf("cannot use protected branch name")
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

// validateProjectName validates user input for project name entry.
// Called by huh input field validator during name entry screen.
//
// The function:
//  1. Checks for empty input
//  2. Normalizes the name using normalizeName()
//  3. Builds full branch name (prefix + normalized)
//  4. Validates using isValidBranchName()
//
// Returns nil if valid, or error with user-friendly message.
//
// Example:
//
//	err := validateProjectName("Web Agents", "feat/")
//	// Normalizes to "web-agents", validates "feat/web-agents"
func validateProjectName(name string, prefix string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	normalized := normalizeName(name)
	branchName := prefix + normalized

	return isValidBranchName(branchName)
}

// shouldCheckUncommittedChanges determines if uncommitted changes validation is needed.
// Returns true only when current branch == target branch.
//
// Why conditional? Git worktrees can't have the same branch checked out twice.
// If current == target, sow must switch the main repo to master/main first.
// Switching with uncommitted changes fails, so we must check first.
//
// Example:
//
//	shouldCheck, err := shouldCheckUncommittedChanges(ctx, "feat/auth")
//	if err != nil {
//	    return err
//	}
//	if shouldCheck {
//	    // Perform validation
//	}
func shouldCheckUncommittedChanges(ctx *sow.Context, targetBranch string) (bool, error) {
	currentBranch, err := ctx.Git().CurrentBranch()
	if err != nil {
		return false, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Only check if we'll need to switch branches
	return currentBranch == targetBranch, nil
}

// performUncommittedChangesCheckIfNeeded runs uncommitted changes validation conditionally.
// Uses existing sow.CheckUncommittedChanges() but adds enhanced error message.
//
// The check only runs when current branch == target branch, because that's when
// sow needs to switch branches (which fails with uncommitted changes).
//
// Error message follows the 3-part pattern:
//  1. What: "Repository has uncommitted changes"
//  2. How: "You are currently on branch 'X'. Creating a worktree requires switching..."
//  3. Next: "To fix: Commit: ... Or stash: ..."
//
// Example:
//
//	err := performUncommittedChangesCheckIfNeeded(ctx, "feat/auth")
//	if err != nil {
//	    // User sees helpful error with current branch and fix commands
//	    return err
//	}
func performUncommittedChangesCheckIfNeeded(ctx *sow.Context, targetBranch string) error {
	shouldCheck, err := shouldCheckUncommittedChanges(ctx, targetBranch)
	if err != nil {
		return err
	}

	if !shouldCheck {
		return nil // No check needed
	}

	// Use existing validation
	if err := git.CheckUncommittedChanges(ctx.Git()); err != nil {
		// Enhance with user-friendly message
		return fmt.Errorf(
			"repository has uncommitted changes\n\n"+
				"You are currently on branch '%s'.\n"+
				"Creating a worktree requires switching to a different branch first.\n\n"+
				"To fix:\n"+
				"  Commit: git add . && git commit -m \"message\"\n"+
				"  Or stash: git stash",
			targetBranch,
		)
	}

	return nil
}

// BranchState represents the state of a branch in the repository.
type BranchState struct {
	BranchExists   bool
	WorktreeExists bool
	ProjectExists  bool
}

// checkBranchState examines branch, worktree, and project state for a given branch name.
// Used before creation to detect conflicts.
//
// Returns:
//   - BranchState with all three boolean flags set
//   - error if filesystem or git operations fail
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
	worktreePath := git.WorktreePath(ctx.RepoRoot(), branchName)
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

// canCreateProject validates that project creation is allowed on this branch.
// Returns error if:
//   - Branch already has a project (state.ProjectExists == true)
//   - Inconsistent state: worktree exists but project missing
//
// Returns nil if creation is allowed.
func canCreateProject(state *BranchState, branchName string) error {
	// Check if project already exists
	if state.ProjectExists {
		return fmt.Errorf("branch '%s' already has a project", branchName)
	}

	// Check for inconsistent state (worktree without project)
	if state.WorktreeExists && !state.ProjectExists {
		return fmt.Errorf("worktree exists but project missing for branch '%s'", branchName)
	}

	// Creation is allowed
	return nil
}

// validateProjectExists checks that a project at given branch exists.
// Used when continuing projects (Work Unit 004).
//
// Returns error if:
//   - Branch doesn't exist
//   - Worktree doesn't exist
//   - Project doesn't exist in worktree
func validateProjectExists(ctx *sow.Context, branchName string) error {
	state, err := checkBranchState(ctx, branchName)
	if err != nil {
		return err
	}

	if !state.BranchExists {
		return fmt.Errorf("branch '%s' does not exist", branchName)
	}

	if !state.WorktreeExists {
		return fmt.Errorf("worktree for branch '%s' does not exist", branchName)
	}

	if !state.ProjectExists {
		return fmt.Errorf("project for branch '%s' does not exist", branchName)
	}

	return nil
}

// listExistingProjects finds all branches with existing projects.
// Used by "continue existing project" screen to show available options.
//
// Returns:
//   - Slice of branch names that have projects
//   - error if filesystem or git operations fail
func listExistingProjects(ctx *sow.Context) ([]string, error) {
	branches, err := ctx.Git().Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var projects []string
	for _, branch := range branches {
		state, err := checkBranchState(ctx, branch)
		if err != nil {
			return nil, err
		}
		if state.ProjectExists {
			projects = append(projects, branch)
		}
	}

	// Sort alphabetically for consistent UI
	sort.Strings(projects)

	return projects, nil
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
		StateTypeSelect:     {StateFileSelect, StateCancelled},
		StateNameEntry:      {StateFileSelect, StateCancelled},
		StateFileSelect:     {StatePromptEntry, StateCancelled},
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

// formatError formats error messages in the consistent 3-part pattern:
//  1. What went wrong (title)
//  2. How to fix (problem and solution)
//  3. Next steps (what to do now)
//
// The function assembles the parts into a single formatted string suitable
// for display in huh components.
//
// Example:
//
//	msg := formatError(
//	    "Cannot create project on protected branch 'main'",
//	    "Projects must be created on feature branches.",
//	    "Action: Choose a different project name",
//	)
func formatError(problem string, howToFix string, nextSteps string) string {
	var parts []string

	if problem != "" {
		parts = append(parts, problem)
	}
	if howToFix != "" {
		parts = append(parts, howToFix)
	}
	if nextSteps != "" {
		parts = append(parts, nextSteps)
	}

	return strings.Join(parts, "\n\n")
}

// showErrorWithOptions displays an error with multiple action choices.
// Returns the user's selected choice.
//
// Used for errors where multiple recovery paths are available
// (e.g., "continue existing project" vs "change name").
//
// Example:
//
//	choice, err := showErrorWithOptions(
//	    formatError(...),
//	    map[string]string{
//	        "retry": "Change project name",
//	        "continue": "Continue existing project",
//	        "cancel": "Cancel",
//	    },
//	)
//
//nolint:unused // Will be used by wizard screens in subsequent work units
func showErrorWithOptions(message string, options map[string]string) (string, error) {
	// Skip interactive prompts in test mode
	if os.Getenv("SOW_TEST") == "1" {
		debugLog("ErrorWithOptions", "%s", message)
		// Return first option key in test mode
		for key := range options {
			return key, nil
		}
		return "", nil
	}

	var selected string

	// Convert map to huh options
	var huhOptions []huh.Option[string]
	for key, label := range options {
		huhOptions = append(huhOptions, huh.NewOption(label, key))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Error").
				Description(message),
			huh.NewSelect[string]().
				Title("What would you like to do?").
				Options(huhOptions...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	return selected, nil
}

// errorProtectedBranch returns the formatted error message for attempting
// to create a project on a protected branch (main or master).
func errorProtectedBranch(branchName string) string {
	return formatError(
		fmt.Sprintf("Cannot create project on protected branch '%s'", branchName),
		"Projects must be created on feature branches.",
		"Action: Choose a different project name",
	)
}

// errorIssueAlreadyLinked returns the formatted error message when a GitHub
// issue already has a linked branch.
func errorIssueAlreadyLinked(issueNumber int, linkedBranch string) string {
	return formatError(
		fmt.Sprintf("Issue #%d already has a linked branch: %s", issueNumber, linkedBranch),
		"To continue working on this issue:\n  Select \"Continue existing project\" from the main menu",
		"",
	)
}

// errorBranchHasProject returns the formatted error message when attempting
// to create a project on a branch that already has one.
func errorBranchHasProject(branchName string, projectName string) string {
	return formatError(
		fmt.Sprintf("Branch '%s' already has a project", branchName),
		"To continue this project:\n  Select \"Continue existing project\" from the main menu",
		fmt.Sprintf("To create a different project:\n  Choose a different project name (currently: \"%s\")", projectName),
	)
}

// errorUncommittedChanges returns the formatted error message when the
// repository has uncommitted changes and worktree creation requires switching branches.
func errorUncommittedChanges(currentBranch string) string {
	return formatError(
		"Repository has uncommitted changes",
		fmt.Sprintf(
			"You are currently on branch '%s'.\n"+
				"Creating a worktree requires switching to a different branch first.",
			currentBranch,
		),
		"To fix:\n"+
			"  Commit: git add . && git commit -m \"message\"\n"+
			"  Or stash: git stash",
	)
}

// errorInconsistentState returns the formatted error message when a worktree
// exists but the project directory is missing.
func errorInconsistentState(branchName string, worktreePath string) string {
	return formatError(
		"Worktree exists but project missing",
		fmt.Sprintf(
			"Branch '%s' has a worktree at %s\n"+
				"but no .sow/project/ directory.",
			branchName,
			worktreePath,
		),
		fmt.Sprintf(
			"To fix:\n"+
				"  1. Remove worktree: git worktree remove %s\n"+
				"  2. Delete directory: rm -rf %s\n"+
				"  3. Try creating project again",
			branchName,
			worktreePath,
		),
	)
}

// errorGitHubCLIMissing returns the formatted error message when the gh
// command is not installed.
func errorGitHubCLIMissing() string {
	return formatError(
		"GitHub CLI not found",
		"The 'gh' command is required for GitHub issue integration.\n\n"+
			"To install:\n"+
			"  macOS: brew install gh\n"+
			"  Linux: See https://cli.github.com/",
		"Or select \"From branch name\" instead.",
	)
}

// wrapValidationError wraps a validation error with user-friendly formatting.
// If err is nil, returns nil.
// Otherwise, wraps with formatError and returns displayable error.
func wrapValidationError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Wrap the error with context
	if context != "" {
		return fmt.Errorf("%s: %w", context, err)
	}

	return err
}

// formatGitHubError converts GitHub command errors to user-friendly messages.
// Handles specific error cases by parsing stderr output.
//
// Error types handled:
//   - Network errors: "check connection, retry"
//   - Rate limit: "wait or authenticate for higher limit"
//   - Issue not found: "check issue number"
//   - Permission denied: "check repository access"
//   - Unknown errors: show command that failed
//
// Returns formatted error string ready for display.
func formatGitHubError(err error) string {
	var ghErr git.ErrGHCommand
	if !errors.As(err, &ghErr) {
		// Not a GitHub command error, return generic message
		return fmt.Sprintf("GitHub operation failed: %v", err)
	}

	stderr := strings.ToLower(ghErr.Stderr)

	// Check for rate limit (most specific)
	if strings.Contains(stderr, "rate limit") {
		return "GitHub API rate limit exceeded.\n\n" +
			"To fix:\n" +
			"  Wait a few minutes and try again\n" +
			"  Or run: gh auth login (for higher limits)"
	}

	// Check for network errors
	if strings.Contains(stderr, "network") ||
		strings.Contains(stderr, "connection") ||
		strings.Contains(stderr, "timeout") ||
		strings.Contains(stderr, "unreachable") {
		return "Cannot reach GitHub.\n\n" +
			"Check your internet connection and try again."
	}

	// Check for not found
	if strings.Contains(stderr, "not found") ||
		strings.Contains(stderr, "does not exist") {
		return "Resource not found.\n\n" +
			"Check the issue number or repository access."
	}

	// Check for permission denied
	if strings.Contains(stderr, "permission denied") ||
		strings.Contains(stderr, "forbidden") ||
		strings.Contains(stderr, "not authorized") {
		return "Permission denied.\n\n" +
			"Check your GitHub repository access."
	}

	// Unknown error - show command
	return fmt.Sprintf("GitHub command failed: gh %s\n\n"+
		"Check that gh CLI is working correctly:\n"+
		"  Run: gh auth status",
		ghErr.Command)
}

// checkIssueLinkedBranch validates that a GitHub issue doesn't have an
// existing linked branch. Used before creating a new project from an issue.
//
// Returns:
//   - nil if no linked branches (OK to create)
//   - formatted error if linked branch exists
func checkIssueLinkedBranch(github git.GitHubClient, issueNumber int) error {
	branches, err := github.GetLinkedBranches(issueNumber)
	if err != nil {
		// Format the GitHub error for user display
		return errors.New(formatGitHubError(err))
	}

	// If no linked branches, OK to create
	if len(branches) == 0 {
		return nil
	}

	// Issue already has linked branch - show error
	branchName := branches[0].Name
	return errors.New(errorIssueAlreadyLinked(issueNumber, branchName))
}

// filterIssuesBySowLabel filters issues to only include those with 'sow' label.
// Used by issue selection screen to show only sow-related issues.
//
// Returns:
//   - Slice of issues that have the 'sow' label
func filterIssuesBySowLabel(issues []git.Issue) []git.Issue {
	var filtered []git.Issue

	for _, issue := range issues {
		if issue.HasLabel("sow") {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

// discoverKnowledgeFiles walks the knowledge directory tree and returns all file paths.
// Files are returned as relative paths from the knowledge directory (e.g., "designs/api.md").
// Results are sorted alphabetically for consistent presentation.
//
// Edge cases:
//   - If directory doesn't exist: returns empty slice, NOT an error (graceful degradation)
//   - If directory is empty: returns empty slice
//   - If permission errors occur: returns error
//   - Directories are skipped, only files are returned
//
// Example:
//
//	knowledgeDir := filepath.Join(ctx.MainRepoRoot(), ".sow", "knowledge")
//	files, err := discoverKnowledgeFiles(knowledgeDir)
//	if err != nil {
//	    return fmt.Errorf("failed to discover files: %w", err)
//	}
//	// files = ["README.md", "designs/api.md", "adrs/001-decision.md"]
func discoverKnowledgeFiles(knowledgeDir string) ([]string, error) {
	// Check if directory exists
	if _, err := os.Stat(knowledgeDir); err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist - not an error, just return empty slice
			return []string{}, nil
		}
		// Other stat errors (permission denied, etc.) should be returned
		return nil, fmt.Errorf("failed to stat knowledge directory: %w", err)
	}

	var files []string

	// Walk the directory tree
	err := filepath.Walk(knowledgeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Error accessing path - propagate it
			return err
		}

		// Skip directories - we only want files
		if info.IsDir() {
			return nil
		}

		// Get relative path from knowledge directory
		relPath, err := filepath.Rel(knowledgeDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk knowledge directory: %w", err)
	}

	// Sort alphabetically for consistent ordering
	sort.Strings(files)

	return files, nil
}
