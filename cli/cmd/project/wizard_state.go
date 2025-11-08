package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
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

// Wizard manages the interactive project creation/continuation workflow.
type Wizard struct {
	state       WizardState
	ctx         *sow.Context
	choices     map[string]interface{}
	claudeFlags []string
	cmd         *cobra.Command
}

// NewWizard creates a new wizard instance.
func NewWizard(cmd *cobra.Command, ctx *sow.Context, claudeFlags []string) *Wizard {
	return &Wizard{
		state:       StateEntry,
		ctx:         ctx,
		choices:     make(map[string]interface{}),
		claudeFlags: claudeFlags,
		cmd:         cmd,
	}
}

// Run executes the wizard state machine loop.
func (w *Wizard) Run() error {
	for w.state != StateComplete && w.state != StateCancelled {
		if err := w.handleState(); err != nil {
			return err
		}
	}

	if w.state == StateCancelled {
		return nil // User cancelled, not an error
	}

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

// handleIssueSelect allows selecting a GitHub issue (stub for now).
func (w *Wizard) handleIssueSelect() error {
	fmt.Println("Issue select screen (stub)")
	w.state = StateComplete
	return nil
}

// handleTypeSelect allows selecting project type.
func (w *Wizard) handleTypeSelect() error {
	var selectedType string

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

	projectType, ok := w.choices["type"].(string)
	if !ok {
		return fmt.Errorf("type choice not set or invalid")
	}
	branchName, ok := w.choices["branch"].(string)
	if !ok {
		return fmt.Errorf("branch choice not set or invalid")
	}

	contextInfo := fmt.Sprintf(
		"Type: %s\nBranch: %s\n\nPress Ctrl+E to open $EDITOR for multi-line input",
		projectType, branchName)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Enter your task or question for Claude (optional):").
				Description(contextInfo).
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

// handleProjectSelect allows selecting existing project to continue.
func (w *Wizard) handleProjectSelect() error {
	// 1. Discover projects with spinner
	var projects []ProjectInfo
	err := withSpinner("Discovering projects...", func() error {
		var discoverErr error
		projects, discoverErr = listProjects(w.ctx)
		return discoverErr
	})

	if err != nil {
		return fmt.Errorf("failed to discover projects: %w", err)
	}

	// 2. Handle empty list
	if len(projects) == 0 {
		fmt.Fprintln(os.Stderr, "\nNo existing projects found.")
		w.state = StateCancelled
		return nil
	}

	// 3. Build selection options
	var selectedBranch string
	options := make([]huh.Option[string], 0, len(projects)+1)

	for _, proj := range projects {
		progress := formatProjectProgress(proj)
		label := fmt.Sprintf("%s - %s\n    [%s]", proj.Branch, proj.Name, progress)
		options = append(options, huh.NewOption(label, proj.Branch))
	}
	options = append(options, huh.NewOption("Cancel", "cancel"))

	// 4. Show selection
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a project to continue:").
				Options(options...).
				Value(&selectedBranch),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
			return nil
		}
		return fmt.Errorf("project selection error: %w", err)
	}

	// 5. Handle cancellation
	if selectedBranch == "cancel" {
		w.state = StateCancelled
		return nil
	}

	// 6. Validate project still exists
	var selectedProj *ProjectInfo
	for i := range projects {
		if projects[i].Branch == selectedBranch {
			selectedProj = &projects[i]
			break
		}
	}

	if selectedProj == nil {
		return fmt.Errorf("internal error: selected project not found in list")
	}

	// Double-check state file still exists (race condition check)
	worktreePath := sow.WorktreePath(w.ctx.MainRepoRoot(), selectedBranch)
	statePath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")
	if _, err := os.Stat(statePath); err != nil {
		// Project was deleted between discovery and selection
		_ = showError("Project no longer exists (state file missing)\n\nPress Enter to try again")
		return nil // Stay in current state to retry
	}

	// 7. Save selection and transition
	w.choices["project"] = *selectedProj
	w.state = StateContinuePrompt
	return nil
}

// handleContinuePrompt allows entering additional prompt for continuing.
func (w *Wizard) handleContinuePrompt() error {
	// 1. Extract selected project
	proj, ok := w.choices["project"].(ProjectInfo)
	if !ok {
		return fmt.Errorf("internal error: project choice not set or invalid")
	}

	// 2. Build context display
	progress := formatProjectProgress(proj)
	contextInfo := fmt.Sprintf(
		"Project: %s\nBranch: %s\nState: %s",
		proj.Name,
		proj.Branch,
		progress,
	)

	// 3. Show prompt entry form
	var prompt string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("What would you like to work on? (optional):").
				Description(contextInfo + "\n\nPress Ctrl+E to open $EDITOR for multi-line input").
				CharLimit(5000).
				Value(&prompt).
				EditorExtension(".md"),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
			return nil
		}
		return fmt.Errorf("continuation prompt error: %w", err)
	}

	// 4. Save prompt and transition
	w.choices["prompt"] = prompt
	w.state = StateComplete
	return nil
}

// finalize routes to the appropriate finalization method based on the action choice.
func (w *Wizard) finalize() error {
	// Determine which path we're on
	action, ok := w.choices["action"].(string)
	if !ok {
		return fmt.Errorf("internal error: action choice not set")
	}

	switch action {
	case "create":
		return w.finalizeCreation()
	case "continue":
		return w.finalizeContinuation()
	default:
		return fmt.Errorf("internal error: unknown action: %s", action)
	}
}

// finalizeCreation creates the project, initializes it in a worktree, and launches Claude Code.
func (w *Wizard) finalizeCreation() error {
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

	// Step 3: Initialize project in worktree
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to create worktree context: %w", err)
	}

	project, err := initializeProject(worktreeCtx, branch, name, nil)
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

// finalizeContinuation loads an existing project, generates a continuation prompt, and launches Claude Code.
// NOTE: Unlike creation, continuation does NOT check uncommitted changes.
// The worktree already exists, so there's no risk of needing to switch branches in the main repo.
// This is intentional and documented in issue #71.
func (w *Wizard) finalizeContinuation() error {
	// 1. Extract choices
	proj, ok := w.choices["project"].(ProjectInfo)
	if !ok {
		return fmt.Errorf("internal error: project choice not set or invalid")
	}

	userPrompt, ok := w.choices["prompt"].(string)
	if !ok {
		return fmt.Errorf("internal error: prompt choice not set or invalid")
	}

	// 2. Ensure worktree exists (idempotent)
	worktreePath := sow.WorktreePath(w.ctx.MainRepoRoot(), proj.Branch)
	if err := sow.EnsureWorktree(w.ctx, worktreePath, proj.Branch); err != nil {
		return fmt.Errorf("failed to ensure worktree: %w", err)
	}

	// 3. Create worktree context
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to create worktree context: %w", err)
	}

	// 4. Load fresh project state
	projectState, err := state.Load(worktreeCtx)
	if err != nil {
		return fmt.Errorf("failed to load project state: %w", err)
	}

	// 5. Generate 3-layer continuation prompt
	basePrompt, err := generateContinuePrompt(projectState)
	if err != nil {
		return fmt.Errorf("failed to generate continuation prompt: %w", err)
	}

	// 6. Append user prompt if provided
	fullPrompt := basePrompt
	if userPrompt != "" {
		fullPrompt += "\n\nUser request:\n" + userPrompt
	}

	// 7. Success message
	fmt.Fprintf(os.Stderr, "✓ Continuing project '%s' on branch %s\n", proj.Name, proj.Branch)

	// 8. Launch Claude
	if w.cmd != nil {
		if err := launchClaudeCode(w.cmd, worktreeCtx, fullPrompt, w.claudeFlags); err != nil {
			return fmt.Errorf("failed to launch Claude: %w", err)
		}
	}

	return nil
}
