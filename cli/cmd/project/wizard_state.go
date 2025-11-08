package project

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// WizardState represents the current state of the project wizard.
type WizardState string

// Wizard states define the flow through project creation/continuation.
const (
	StateEntry          WizardState = "entry"
	StateCreateSource   WizardState = "create_source"
	StateIssueSelect    WizardState = "issue_select"
	StateTypeSelect     WizardState = "type_select"
	StateNameEntry      WizardState = "name_entry"
	StatePromptEntry    WizardState = "prompt_entry"
	StateProjectSelect  WizardState = "project_select"
	StateContinuePrompt WizardState = "continue_prompt"
	StateComplete       WizardState = "complete"
	StateCancelled      WizardState = "cancelled"
)

// GitHubClient defines the interface for GitHub operations used by the wizard.
// This interface allows for easy mocking in tests.
type GitHubClient interface {
	Ensure() error
	ListIssues(label, state string) ([]sow.Issue, error)
	GetLinkedBranches(number int) ([]sow.LinkedBranch, error)
	CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)
	GetIssue(number int) (*sow.Issue, error)
}

// Wizard manages the interactive project creation/continuation workflow.
type Wizard struct {
	state       WizardState
	ctx         *sow.Context
	choices     map[string]interface{}
	claudeFlags []string
	cmd         *cobra.Command
	github      GitHubClient // GitHub client for issue operations
}

// NewWizard creates a new wizard instance.
func NewWizard(cmd *cobra.Command, ctx *sow.Context, claudeFlags []string) *Wizard {
	ghExec := sowexec.NewLocal("gh")

	return &Wizard{
		state:       StateEntry,
		ctx:         ctx,
		choices:     make(map[string]interface{}),
		claudeFlags: claudeFlags,
		cmd:         cmd,
		github:      sow.NewGitHub(ghExec),
	}
}

// setState transitions the wizard to a new state with optional validation in debug mode.
func (w *Wizard) setState(newState WizardState) {
	oldState := w.state
	w.state = newState

	debugLog("Wizard", "State transition: %s -> %s", oldState, newState)

	// In debug mode, validate state transitions
	if os.Getenv("SOW_DEBUG") == "1" {
		if err := validateStateTransition(oldState, newState); err != nil {
			fmt.Fprintf(os.Stderr, "[WARN] %v\n", err)
		}
	}
}

// Run executes the wizard state machine loop.
func (w *Wizard) Run() error {
	debugLog("Wizard", "Starting wizard in state=%s", w.state)

	for w.state != StateComplete && w.state != StateCancelled {
		if err := w.handleState(); err != nil {
			return err
		}
	}

	if w.state == StateCancelled {
		debugLog("Wizard", "User cancelled wizard")
		return nil // User cancelled, not an error
	}

	debugLog("Wizard", "Wizard complete, finalizing project")
	return w.finalize()
}

// handleState dispatches to the appropriate handler based on current state.
func (w *Wizard) handleState() error {
	switch w.state {
	case StateEntry:
		return w.handleEntry()
	case StateCreateSource:
		return w.handleCreateSource()
	case StateIssueSelect:
		return w.handleIssueSelect()
	case StateTypeSelect:
		return w.handleTypeSelect()
	case StateNameEntry:
		return w.handleNameEntry()
	case StatePromptEntry:
		return w.handlePromptEntry()
	case StateProjectSelect:
		return w.handleProjectSelect()
	case StateContinuePrompt:
		return w.handleContinuePrompt()
	default:
		return fmt.Errorf("unknown state: %s", w.state)
	}
}

// handleEntry shows the main entry screen with create/continue/cancel options.
func (w *Wizard) handleEntry() error {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do?").
				Options(
					huh.NewOption("Create new project", "create"),
					huh.NewOption("Continue existing project", "continue"),
					huh.NewOption("Cancel", "cancel"),
				).
				Value(&action),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
			return nil
		}
		return fmt.Errorf("entry screen error: %w", err)
	}

	w.choices["action"] = action

	switch action {
	case "create":
		w.state = StateCreateSource
	case "continue":
		w.state = StateProjectSelect
	case "cancel":
		w.state = StateCancelled
	}

	return nil
}

// handleCreateSource shows options for creating a project.
func (w *Wizard) handleCreateSource() error {
	var source string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("How would you like to create the project?").
				Options(
					huh.NewOption("From GitHub issue", "issue"),
					huh.NewOption("From branch name", "branch"),
					huh.NewOption("Cancel", "cancel"),
				).
				Value(&source),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
			return nil
		}
		return fmt.Errorf("create source screen error: %w", err)
	}

	w.choices["source"] = source

	switch source {
	case "issue":
		w.state = StateIssueSelect
	case "branch":
		w.state = StateTypeSelect
	case "cancel":
		w.state = StateCancelled
	}

	return nil
}

// handleIssueSelect allows selecting a GitHub issue.
func (w *Wizard) handleIssueSelect() error {
	debugLog("Wizard", "State=%s", w.state)

	// Validate GitHub CLI is available and authenticated
	if err := w.github.Ensure(); err != nil {
		return w.handleGitHubError(err)
	}

	// Fetch issues with 'sow' label using spinner
	var issues []sow.Issue
	var fetchErr error

	debugLog("GitHub", "Calling gh issue list --label sow --state open")
	err := withSpinner("Fetching issues from GitHub...", func() error {
		issues, fetchErr = w.github.ListIssues("sow", "open")
		return fetchErr
	})

	if err != nil {
		debugLog("GitHub", "Failed to fetch issues: %v", err)
		errorMsg := fmt.Sprintf("Failed to fetch issues: %v\n\n"+
			"This may be a network issue or a GitHub API problem.\n"+
			"Please try again or select 'From branch name' instead.", err)
		_ = showError(errorMsg)
		w.state = StateCreateSource
		return nil
	}

	debugLog("GitHub", "Fetched %d issues", len(issues))
	for _, issue := range issues {
		debugLog("GitHub", "Issue #%d: %s", issue.Number, issue.Title)
	}

	// Handle empty issue list
	if len(issues) == 0 {
		errorMsg := "No issues found with 'sow' label\n\n" +
			"To use GitHub issue integration:\n" +
			"  1. Create an issue in your repository\n" +
			"  2. Add the 'sow' label to the issue\n" +
			"  3. Try again\n\n" +
			"Or select 'From branch name' to continue without an issue."
		_ = showError(errorMsg)
		w.state = StateCreateSource
		return nil
	}

	// Store issues in choices for next step (Task 030)
	w.choices["issues"] = issues

	// Proceed to issue selection (next screen)
	return w.showIssueSelectScreen()
}

// handleGitHubError displays GitHub-related errors and offers fallback paths.
// Returns nil to keep wizard running (user can choose fallback).
func (w *Wizard) handleGitHubError(err error) error {
	var errorMsg string
	var fallbackMsg string

	// Determine error type using errors.As for wrapped error support
	var notInstalled sow.ErrGHNotInstalled
	var notAuthenticated sow.ErrGHNotAuthenticated

	if errors.As(err, &notInstalled) {
		errorMsg = "GitHub CLI not found\n\n" +
			"The 'gh' command is required for GitHub issue integration.\n\n" +
			"To install:\n" +
			"  macOS: brew install gh\n" +
			"  Linux: See https://cli.github.com/"
		fallbackMsg = "Or select 'From branch name' instead."

	} else if errors.As(err, &notAuthenticated) {
		errorMsg = "GitHub CLI not authenticated\n\n" +
			"Run the following command to authenticate:\n" +
			"  gh auth login\n\n" +
			"Then try creating your project again."
		fallbackMsg = "Or select 'From branch name' instead."

	} else {
		// Generic GitHub error
		errorMsg = fmt.Sprintf("GitHub CLI error: %v", err)
		fallbackMsg = "Select 'From branch name' to continue without GitHub integration."
	}

	// Show error with fallback option
	fullMessage := errorMsg + "\n\n" + fallbackMsg
	_ = showError(fullMessage)

	// Return to source selection so user can choose "From branch name"
	w.state = StateCreateSource
	return nil
}

// handleTypeSelect allows selecting project type.
func (w *Wizard) handleTypeSelect() error {
	var selectedType string

	// Check if we have issue context
	_, hasIssue := w.choices["issue"].(*sow.Issue)

	// Build form with just type selection
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What type of project?").
				Options(getTypeOptions()...).
				Value(&selectedType),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
			return nil
		}
		return fmt.Errorf("type selection error: %w", err)
	}

	if selectedType == "cancel" {
		w.state = StateCancelled
		return nil
	}

	w.choices["type"] = selectedType

	// Route based on context
	if hasIssue {
		// GitHub issue path: create branch then go to prompt entry
		return w.createLinkedBranch()
	}

	// Branch name path: go to name entry
	w.state = StateNameEntry
	return nil
}

// handleNameEntry allows entering project name with real-time branch preview.
func (w *Wizard) handleNameEntry() error {
	var name string
	projectType, ok := w.choices["type"].(string)
	if !ok {
		return fmt.Errorf("type choice not set or invalid")
	}
	prefix := getTypePrefix(projectType)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter project name:").
				Placeholder("e.g., Web Based Agents").
				Value(&name).
				Validate(func(s string) error {
					// Validation 1: Not empty
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("project name cannot be empty")
					}

					// Validation 2: Not protected branch
					normalized := normalizeName(s)
					branchName := fmt.Sprintf("%s%s", prefix, normalized)

					if w.ctx.Git().IsProtectedBranch(branchName) {
						return fmt.Errorf("cannot use protected branch name")
					}

					// Validation 3: Valid git branch name
					if err := isValidBranchName(branchName); err != nil {
						return err
					}

					return nil
				}),

			// Real-time preview
			huh.NewNote().
				Title("Branch Preview").
				DescriptionFunc(func() string {
					if name == "" {
						return fmt.Sprintf("%s<project-name>", prefix)
					}
					normalized := normalizeName(name)
					return fmt.Sprintf("%s%s", prefix, normalized)
				}, &name), // CRITICAL: Bind to name variable
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateTypeSelect // Go back
			return nil
		}
		return fmt.Errorf("name entry error: %w", err)
	}

	// Post-submit validation: check branch state
	normalized := normalizeName(name)
	branchName := fmt.Sprintf("%s%s", prefix, normalized)

	state, err := checkBranchState(w.ctx, branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch state: %w", err)
	}

	if state.ProjectExists {
		_ = showError(fmt.Sprintf(
			"Error: Branch '%s' already has a project\n\n"+
				"To continue this project:\n"+
				"  Select \"Continue existing project\" from the main menu\n\n"+
				"To create a different project:\n"+
				"  Choose a different project name",
			branchName))
		return nil // Stay in current state to retry
	}

	// Store both original name and full branch name
	w.choices["name"] = name
	w.choices["branch"] = branchName
	w.state = StatePromptEntry

	return nil
}

// handlePromptEntry allows entering initial prompt with external editor support.
func (w *Wizard) handlePromptEntry() error {
	var prompt string

	// Build context display based on project source
	var contextLines []string

	// Check for issue context (GitHub issue path)
	if issue, ok := w.choices["issue"].(*sow.Issue); ok {
		contextLines = append(contextLines,
			fmt.Sprintf("Issue: #%d - %s", issue.Number, issue.Title))
	}

	// Show branch name
	if branchName, ok := w.choices["branch"].(string); ok {
		contextLines = append(contextLines, fmt.Sprintf("Branch: %s", branchName))
	} else if name, ok := w.choices["name"].(string); ok {
		// Branch name path - compute branch name for display
		projectType, ok := w.choices["type"].(string)
		if !ok {
			projectType = "standard" // Default fallback
		}
		prefix := getTypePrefix(projectType)
		normalized := normalizeName(name)
		contextLines = append(contextLines,
			fmt.Sprintf("Branch: %s%s", prefix, normalized))
	}

	// Add project type for clarity
	if projectType, ok := w.choices["type"].(string); ok {
		typeConfig := projectTypes[projectType]
		contextLines = append(contextLines,
			fmt.Sprintf("Type: %s", typeConfig.Description))
	}

	contextDisplay := strings.Join(contextLines, "\n")
	instructionText := contextDisplay + "\n\nPress Ctrl+E to open $EDITOR for multi-line input"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Enter your task or question for Claude (optional):").
				Description(instructionText).
				CharLimit(10000).
				Value(&prompt).
				EditorExtension(".md"),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
			return nil
		}
		return fmt.Errorf("prompt entry error: %w", err)
	}

	w.choices["prompt"] = prompt
	w.state = StateComplete

	return nil
}

// handleProjectSelect allows selecting existing project to continue (stub for now).
func (w *Wizard) handleProjectSelect() error {
	fmt.Println("Project select screen (stub)")
	w.state = StateComplete
	return nil
}

// handleContinuePrompt allows entering additional prompt for continuing (stub for now).
func (w *Wizard) handleContinuePrompt() error {
	fmt.Println("Continue prompt screen (stub)")
	w.state = StateComplete
	return nil
}

// showIssueSelectScreen displays the issue selection prompt.
// Issues are retrieved from w.choices["issues"] (set by handleIssueSelect).
func (w *Wizard) showIssueSelectScreen() error {
	issues, ok := w.choices["issues"].([]sow.Issue)
	if !ok {
		return fmt.Errorf("issues not found in choices")
	}

	var selectedIssueNumber int

	// Build select options
	options := make([]huh.Option[int], 0, len(issues)+1)
	for _, issue := range issues {
		label := fmt.Sprintf("#%d: %s", issue.Number, issue.Title)
		options = append(options, huh.NewOption(label, issue.Number))
	}

	// Add cancel option
	options = append(options, huh.NewOption("Cancel", -1))

	// Create select form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select an issue (filtered by 'sow' label):").
				Options(options...).
				Value(&selectedIssueNumber),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
			return nil
		}
		return fmt.Errorf("issue selection error: %w", err)
	}

	// Handle cancel
	if selectedIssueNumber == -1 {
		w.state = StateCancelled
		return nil
	}

	// NEW: Validate issue doesn't have linked branch
	linkedBranches, err := w.github.GetLinkedBranches(selectedIssueNumber)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to check linked branches: %v\n\n"+
			"Please try again or select 'From branch name' instead.", err)
		_ = showError(errorMsg)
		w.state = StateCreateSource
		return nil
	}

	if len(linkedBranches) > 0 {
		return w.handleAlreadyLinkedError(selectedIssueNumber, linkedBranches[0])
	}

	// NEW: Fetch full issue details
	issue, err := w.github.GetIssue(selectedIssueNumber)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get issue details: %v\n\n"+
			"Please try again.", err)
		_ = showError(errorMsg)
		// Stay in current state to allow retry
		return nil
	}

	// Store issue in choices for next steps
	w.choices["issue"] = issue

	// Task 070: Default GitHub issues to "standard" type and skip type selection
	w.choices["type"] = "standard"

	// Task 070: Proceed directly to branch creation (skip type selection screen)
	return w.createLinkedBranch()
}

// createLinkedBranch generates a branch name from the issue and creates a linked branch.
func (w *Wizard) createLinkedBranch() error {
	issue, ok := w.choices["issue"].(*sow.Issue)
	if !ok {
		return fmt.Errorf("issue not found in choices")
	}

	projectType, ok := w.choices["type"].(string)
	if !ok {
		return fmt.Errorf("type not found in choices")
	}

	// Generate branch name: <prefix><issue-slug>-<number>
	prefix := getTypePrefix(projectType)
	issueSlug := normalizeName(issue.Title)
	branchName := fmt.Sprintf("%s%s-%d", prefix, issueSlug, issue.Number)

	// Create linked branch via gh issue develop with spinner
	var createdBranch string
	err := withSpinner("Creating linked branch...", func() error {
		var err error
		// Pass checkout=false because we use worktrees, not traditional checkout
		createdBranch, err = w.github.CreateLinkedBranch(issue.Number, branchName, false)
		return err
	})

	if err != nil {
		errorMsg := fmt.Sprintf("Failed to create linked branch: %v\n\n"+
			"This may be a GitHub API issue. Please try again.",
			err)
		_ = showError(errorMsg)
		// Stay in current state to allow retry
		return nil
	}

	// Store branch name and use issue title as project name
	w.choices["branch"] = createdBranch
	w.choices["name"] = issue.Title

	// Proceed to prompt entry
	w.state = StatePromptEntry

	return nil
}

// handleAlreadyLinkedError displays error when issue has existing linked branch.
// Returns nil to keep wizard running (user can select different issue).
func (w *Wizard) handleAlreadyLinkedError(issueNumber int, branch sow.LinkedBranch) error {
	errorMsg := fmt.Sprintf(
		"Issue #%d already has a linked branch: %s\n\n"+
			"To continue working on this issue:\n"+
			"  Select \"Continue existing project\" from the main menu\n\n"+
			"To work on a different issue:\n"+
			"  Select a different issue from the list",
		issueNumber,
		branch.Name,
	)

	_ = showError(errorMsg)

	// Return to issue select to let user choose different issue
	// Keep issues list in choices so we don't need to fetch again
	return w.showIssueSelectScreen()
}

// finalize creates the project, initializes it in a worktree, and launches Claude Code.
func (w *Wizard) finalize() error {
	// Extract wizard choices
	name, ok := w.choices["name"].(string)
	if !ok {
		return fmt.Errorf("name choice not set or invalid")
	}
	branch, ok := w.choices["branch"].(string)
	if !ok {
		return fmt.Errorf("branch choice not set or invalid")
	}
	initialPrompt := ""
	if prompt, ok := w.choices["prompt"].(string); ok {
		initialPrompt = prompt
	}

	// Extract issue if present (GitHub issue path)
	var issue *sow.Issue
	if issueData, ok := w.choices["issue"].(*sow.Issue); ok {
		issue = issueData
	}

	// Step 1: Conditional uncommitted changes check
	currentBranch, err := w.ctx.Git().CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Only check if we're on the branch we're trying to create a worktree for
	if currentBranch == branch {
		if err := sow.CheckUncommittedChanges(w.ctx); err != nil {
			return fmt.Errorf("repository has uncommitted changes\n\n"+
				"You are currently on branch '%s'.\n"+
				"Creating a worktree requires switching to a different branch first.\n\n"+
				"To fix:\n"+
				"  Commit: git add . && git commit -m \"message\"\n"+
				"  Or stash: git stash", currentBranch)
		}
	}

	// Step 2: Ensure worktree exists
	worktreePath := sow.WorktreePath(w.ctx.RepoRoot(), branch)
	if err := sow.EnsureWorktree(w.ctx, worktreePath, branch); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Step 3: Initialize project in worktree WITH issue metadata
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to create worktree context: %w", err)
	}

	// Pass issue to initializeProject (will be nil for branch name path)
	project, err := initializeProject(worktreeCtx, branch, name, issue)
	if err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	// Step 4: Generate 3-layer prompt
	prompt, err := generateNewProjectPrompt(project, initialPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Step 5: Display success message
	_, _ = fmt.Fprintf(os.Stdout, "✓ Initialized project '%s' on branch %s\n", name, branch)
	if issue != nil {
		_, _ = fmt.Fprintf(os.Stdout, "✓ Linked to issue #%d: %s\n", issue.Number, issue.Title)
	}
	_, _ = fmt.Fprintf(os.Stdout, "✓ Launching Claude in worktree...\n")

	// Step 6: Launch Claude Code
	// Note: w.cmd may be nil in tests, so we skip launch in that case
	if w.cmd != nil {
		if err := launchClaudeCode(w.cmd, worktreeCtx, prompt, w.claudeFlags); err != nil {
			return fmt.Errorf("failed to launch Claude: %w", err)
		}
	}

	return nil
}
