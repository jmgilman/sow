package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/modes"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// NewProjectCmd creates the unified project command.
func NewProjectCmd() *cobra.Command {
	var branchName string
	var issueNumber int
	var noLaunch bool

	cmd := &cobra.Command{
		Use:   "project",
		Short: "Start or continue a project",
		Long: `Start a new project or continue an existing one.

This unified command handles project lifecycle based on context:

No arguments:
  - Checks current branch for existing project
  - If found: continues the project
  - If not: creates new project (validates not on protected branch)

With --branch:
  - Checks out the branch (creates if doesn't exist)
  - If project exists: continues it
  - If not: creates new project

With --issue:
  - Looks up GitHub issue and checks for linked branch
  - If branch found: checks out branch and continues project
  - If not found: creates new branch from issue and new project

Claude Code Flags:
  Use -- to pass additional flags to the Claude Code CLI:
    sow project -- --model opus --verbose

Examples:
  sow project                    # Continue or start in current branch
  sow project --branch feat/auth # Work on feat/auth branch
  sow project --issue 123        # Work on issue #123
  sow project -- --model opus    # Continue with specific Claude model`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProject(cmd, branchName, issueNumber, noLaunch)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to work on (creates if doesn't exist)")
	cmd.Flags().IntVar(&issueNumber, "issue", 0, "GitHub issue number to work on")
	cmd.Flags().BoolVar(&noLaunch, "no-launch", false, "Setup worktree and project but don't launch Claude Code (for testing)")
	cmd.MarkFlagsMutuallyExclusive("branch", "issue")

	return cmd
}

func runProject(cmd *cobra.Command, branchName string, issueNumber int, noLaunch bool) error {
	// 1. Get main repo context
	mainCtx := cmdutil.GetContext(cmd.Context())

	// Require sow to be initialized
	if !mainCtx.IsInitialized() {
		fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
		fmt.Fprintln(os.Stderr, "Run: sow init")
		return fmt.Errorf("not initialized")
	}

	// Extract Claude Code flags (everything after --)
	var claudeFlags []string
	if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
		claudeFlags = cmd.Flags().Args()[dashIndex:]
	}

	// 2. Check for uncommitted changes BEFORE any branch operations
	if err := sow.CheckUncommittedChanges(mainCtx); err != nil {
		return fmt.Errorf("cannot create worktree: %w", err)
	}

	// 3. Determine target branch BEFORE worktree creation
	var selectedBranch string
	var err error
	var issue *sow.Issue
	var shouldCreateNew bool

	// Scenario 1: --issue flag provided
	if issueNumber > 0 {
		selectedBranch, issue, shouldCreateNew, err = handleIssueScenario(mainCtx, issueNumber)
		if err != nil {
			return err
		}
	} else if branchName != "" {
		// Scenario 2: --branch flag provided
		selectedBranch, shouldCreateNew, err = handleBranchScenario(mainCtx, branchName)
		if err != nil {
			return err
		}
	} else {
		// Scenario 3: No flags (current branch)
		selectedBranch, shouldCreateNew, err = handleCurrentBranchScenario(mainCtx)
		if err != nil {
			return err
		}
	}

	// At this point we know:
	// - selectedBranch: which branch we're on
	// - shouldCreateNew: whether to create new or continue
	// - issue: GitHub issue if applicable (nil otherwise)

	// 4. Validate not on protected branch (if creating new project)
	if shouldCreateNew && mainCtx.Git().IsProtectedBranch(selectedBranch) {
		return fmt.Errorf("cannot create project on protected branch '%s' - use a feature branch", selectedBranch)
	}

	// 5. Generate worktree path and ensure directory exists
	worktreePath := sow.WorktreePath(mainCtx.RepoRoot(), selectedBranch)
	worktreesDir := filepath.Join(mainCtx.RepoRoot(), ".sow", "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// 6. Ensure worktree exists
	if err := sow.EnsureWorktree(mainCtx, worktreePath, selectedBranch); err != nil {
		return fmt.Errorf("failed to ensure worktree: %w", err)
	}

	// 7. Create worktree context
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to create worktree context: %w", err)
	}

	if shouldCreateNew {
		// Create new project in worktree context
		if err := initializeProject(worktreeCtx, issue, ""); err != nil {
			return fmt.Errorf("failed to initialize project: %w", err)
		}

		// Generate new project prompt
		prompt, err := generateNewProjectPrompt(worktreeCtx, issue, "", selectedBranch)
		if err != nil {
			return fmt.Errorf("failed to generate new project prompt: %w", err)
		}

		// 8. Launch Claude Code from worktree directory (unless --no-launch)
		if noLaunch {
			return nil
		}
		return launchClaudeCode(cmd, worktreeCtx, prompt, claudeFlags)
	}

	// Continue existing project in worktree context
	project, err := loader.Load(worktreeCtx)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Generate continue prompt
	prompt, err := generateContinuePrompt(worktreeCtx, project)
	if err != nil {
		return fmt.Errorf("failed to generate continue prompt: %w", err)
	}

	// 8. Launch Claude Code from worktree directory (unless --no-launch)
	if noLaunch {
		return nil
	}
	return launchClaudeCode(cmd, worktreeCtx, prompt, claudeFlags)
}

// handleIssueScenario handles the --issue flag scenario.
// Returns: (branchName, issue, shouldCreateNew, error).
func handleIssueScenario(ctx *sow.Context, issueNumber int) (string, *sow.Issue, bool, error) {
	ghExec := sowexec.NewLocal("gh")
	gh := sow.NewGitHub(ghExec)

	// Fetch issue
	issue, err := gh.GetIssue(issueNumber)
	if err != nil {
		return "", nil, false, fmt.Errorf("failed to fetch issue #%d: %w", issueNumber, err)
	}

	// Validate has 'sow' label
	if !issue.HasLabel("sow") {
		return "", nil, false, fmt.Errorf("issue #%d does not have the 'sow' label - add it with: gh issue edit %d --add-label sow", issueNumber, issueNumber)
	}

	// Check for linked branches
	branches, err := gh.GetLinkedBranches(issueNumber)
	if err != nil {
		return "", nil, false, fmt.Errorf("failed to check linked branches: %w", err)
	}

	if len(branches) > 0 {
		// Branch exists - checkout and continue
		if len(branches) > 1 {
			fmt.Fprintf(os.Stderr, "Warning: issue #%d has multiple linked branches, using first one:\n", issueNumber)
			for _, b := range branches {
				fmt.Fprintf(os.Stderr, "  - %s (%s)\n", b.Name, b.URL)
			}
		}

		branchName := branches[0].Name
		if err := ctx.Git().CheckoutBranch(branchName); err != nil {
			return "", nil, false, fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
		}

		// Check if project exists
		if !loader.Exists(ctx) {
			return "", nil, false, fmt.Errorf("branch %s is linked to issue #%d but has no project - this is unexpected", branchName, issueNumber)
		}

		return branchName, issue, false, nil
	}

	// No linked branch - create new branch and project
	branchName, err := gh.CreateLinkedBranch(issueNumber, "", true)
	if err != nil {
		return "", nil, false, fmt.Errorf("failed to create linked branch: %w", err)
	}

	return branchName, issue, true, nil
}

// handleBranchScenario handles the --branch flag scenario.
// Returns: (branchName, shouldCreateNew, error).
func handleBranchScenario(ctx *sow.Context, branchName string) (string, bool, error) {
	git := ctx.Git()

	// Check if branch exists locally
	branches, err := git.Branches()
	if err != nil {
		return "", false, fmt.Errorf("failed to list branches: %w", err)
	}

	branchExists := false
	for _, b := range branches {
		if b == branchName {
			branchExists = true
			break
		}
	}

	if branchExists {
		// Checkout existing branch
		if err := git.CheckoutBranch(branchName); err != nil {
			return "", false, fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
		}
	} else {
		// Create new branch
		// First check we're on a safe branch to create from
		currentBranch, err := git.CurrentBranch()
		if err != nil {
			return "", false, fmt.Errorf("failed to get current branch: %w", err)
		}

		if !git.IsProtectedBranch(currentBranch) {
			return "", false, fmt.Errorf("cannot create branch %s from %s - please checkout main/master first", branchName, currentBranch)
		}

		// Create and checkout new branch
		if err := modes.CreateBranch(git, branchName); err != nil {
			return "", false, fmt.Errorf("failed to create branch %s: %w", branchName, err)
		}
	}

	// Check if project exists in this branch
	projectExists := loader.Exists(ctx)

	return branchName, !projectExists, nil
}

// handleCurrentBranchScenario handles the no-flags scenario (current branch).
// Returns: (branchName, shouldCreateNew, error).
func handleCurrentBranchScenario(ctx *sow.Context) (string, bool, error) {
	git := ctx.Git()

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return "", false, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if project exists
	projectExists := loader.Exists(ctx)

	if !projectExists {
		// Validate we're not on a protected branch before creating new
		if git.IsProtectedBranch(currentBranch) {
			return "", false, fmt.Errorf("cannot create project on protected branch '%s' - create a feature branch first", currentBranch)
		}
	}

	return currentBranch, !projectExists, nil
}
// initializeProject creates the project structure via statechart.
func initializeProject(ctx *sow.Context, issue *sow.Issue, initialPrompt string) error {
	// Determine project name and description
	var projectName, projectDescription string
	var githubIssueNum *int

	if issue != nil {
		// Use issue title as description, generate name from it
		projectDescription = issue.Title
		projectName = generateProjectName(issue.Title)
		githubIssueNum = &issue.Number
	} else {
		// Use initial prompt or generate generic name
		if initialPrompt != "" {
			projectDescription = initialPrompt
			projectName = generateProjectName(initialPrompt)
		} else {
			projectDescription = "New project"
			projectName = "new-project"
		}
	}

	// Create project using the internal project package
	project, err := loader.Create(ctx, projectName, projectDescription)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Set github_issue field if provided
	if githubIssueNum != nil {
		state := project.Machine().ProjectState()
		issueNum64 := int64(*githubIssueNum)
		state.Project.Github_issue = &issueNum64

		// Save the updated state via the machine
		if err := project.Machine().Save(); err != nil {
			return fmt.Errorf("failed to save github_issue link: %w", err)
		}
	}

	return nil
}

// generateNewProjectPrompt creates the custom prompt for new projects.
// Composes base greeting + new project context.
func generateNewProjectPrompt(ctx *sow.Context, issue *sow.Issue, initialPrompt, branchName string) (string, error) {
	// Render base greeting (orchestrator introduction)
	// Use empty GreetContext since we don't need the state-specific sections
	baseCtx := &prompts.GreetContext{
		SowInitialized: ctx.IsInitialized(),
		HasProject:     false, // Don't include project info in base
	}

	base, err := prompts.Render(prompts.PromptGreetBase, baseCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render base greeting: %w", err)
	}

	// Render new project context
	newProjectCtx := &prompts.NewProjectContext{
		RepoRoot:        ctx.RepoRoot(),
		BranchName:      branchName,
		InitialPrompt:   initialPrompt,
		StatechartState: "DiscoveryDecision",
	}

	if issue != nil {
		newProjectCtx.IssueNumber = &issue.Number
		newProjectCtx.IssueTitle = issue.Title
		newProjectCtx.IssueBody = issue.Body
	}

	newSection, err := prompts.Render(prompts.PromptCommandNew, newProjectCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render new project section: %w", err)
	}

	// Compose: base + new project context
	return base + "\n\n" + newSection, nil
}

// generateContinuePrompt creates the custom prompt for continuing projects.
// Composes base greeting + continue project context.
func generateContinuePrompt(ctx *sow.Context, project domain.Project) (string, error) {
	state := project.Machine().ProjectState()

	// Render base greeting (orchestrator introduction)
	baseCtx := &prompts.GreetContext{
		SowInitialized: ctx.IsInitialized(),
		HasProject:     false, // Don't include project info in base, we'll add it separately
	}

	base, err := prompts.Render(prompts.PromptGreetBase, baseCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render base greeting: %w", err)
	}

	// Build continue context
	promptCtx := &prompts.ContinueProjectContext{
		BranchName:         state.Project.Branch,
		ProjectName:        state.Project.Name,
		ProjectDescription: state.Project.Description,
		StatechartState:    state.Statechart.Current_state,
	}

	// Add issue number if present
	if state.Project.Github_issue != nil {
		issueNum := int(*state.Project.Github_issue)
		promptCtx.IssueNumber = &issueNum
	}

	// Add phase status
	promptCtx.PlanningStatus = state.Phases.Planning.Status
	promptCtx.ImplementationStatus = state.Phases.Implementation.Status
	promptCtx.ReviewStatus = state.Phases.Review.Status
	promptCtx.FinalizeStatus = state.Phases.Finalize.Status

	// Count tasks
	tasks := state.Phases.Implementation.Tasks
	promptCtx.TasksTotal = len(tasks)

	for i := range tasks {
		switch tasks[i].Status {
		case "completed":
			promptCtx.TasksCompleted++
		case "in_progress":
			promptCtx.TasksInProgress++
			if promptCtx.CurrentTaskID == "" {
				promptCtx.CurrentTaskID = tasks[i].Id
				promptCtx.CurrentTaskName = tasks[i].Name
				promptCtx.CurrentTaskStatus = tasks[i].Status
			}
		case "pending":
			promptCtx.TasksPending++
		case "abandoned":
			promptCtx.TasksAbandoned++
		}
	}

	// Generate state-specific guidance and next actions
	promptCtx.StateSpecificGuidance = generateStateGuidance(state)
	promptCtx.NextActions = generateNextActions(state)
	promptCtx.CurrentPhaseDescription = getCurrentPhaseDescription(state)
	promptCtx.NextActionSummary = generateNextActionSummary(state)

	// Render continue section
	continueSection, err := prompts.Render(prompts.PromptCommandContinue, promptCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render continue section: %w", err)
	}

	// Compose: base + continue context
	return base + "\n\n" + continueSection, nil
}

// launchClaudeCode launches Claude Code with the given prompt and optional additional flags.
// Additional flags after the prompt are passed directly to the claude CLI.
func launchClaudeCode(cmd *cobra.Command, ctx *sow.Context, prompt string, claudeFlags []string) error {
	claude := sowexec.NewLocal("claude")
	if !claude.Exists() {
		fmt.Fprintln(os.Stderr, "Error: Claude Code CLI not found")
		fmt.Fprintln(os.Stderr, "Install from: https://claude.com/download")
		return fmt.Errorf("claude not found")
	}

	// Build command args: prompt first, then any additional flags
	args := []string{prompt}
	args = append(args, claudeFlags...)

	claudeCmd := exec.CommandContext(cmd.Context(), claude.Command(), args...)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr
	claudeCmd.Dir = ctx.RepoRoot()

	return claudeCmd.Run()
}

// generateProjectName converts a description to a kebab-case project name.
func generateProjectName(description string) string {
	name := description
	if len(name) > 50 {
		name = name[:50]
	}

	// Convert to kebab-case
	return toKebabCase(name)
}

// toKebabCase converts a string to kebab-case.
func toKebabCase(s string) string {
	// Simple implementation - just lowercase and replace spaces
	result := ""
	for i, r := range s {
		if r == ' ' || r == '_' {
			if i > 0 && result[len(result)-1] != '-' {
				result += "-"
			}
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else if r >= 'A' && r <= 'Z' {
			result += string(r + 32) // Convert to lowercase
		}
	}

	// Remove trailing hyphen
	if len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}

	return result
}

// generateStateGuidance provides context-specific guidance based on current state.
func generateStateGuidance(state *schemas.ProjectState) string {
	switch state.Statechart.Current_state {
	case "DiscoveryDecision":
		return "You are at the Discovery Decision state. Assess whether discovery work is needed for this project using the Discovery Worthiness Rubric."
	case "DiscoveryActive":
		return "You are in the Discovery phase. Research and investigate to build context for this work."
	case "DesignDecision":
		return "You are at the Design Decision state. Determine if design/architecture work is warranted before implementation."
	case "DesignActive":
		return "You are in the Design phase. Create architectural decisions and design documents as needed."
	case "ImplementationPlanning":
		return "You are in Implementation Planning. Break down the work into concrete tasks for execution."
	case "ImplementationExecuting":
		return "You are executing implementation tasks. Spawn implementer agents to complete each task."
	case "ReviewActive":
		return "You are in the Review phase. Perform quality validation of the completed work."
	case "FinalizeDocumentation":
		return "You are in Finalize (Documentation). Update relevant documentation files."
	case "FinalizeChecks":
		return "You are in Finalize (Checks). Run final validation (tests, linters, build)."
	case "FinalizeDelete":
		return "You are in Finalize (Delete). Clean up project structure and create PR."
	default:
		return "Continue working on the current project phase."
	}
}

// generateNextActions provides specific next actions based on current state.
func generateNextActions(state *schemas.ProjectState) string {
	switch state.Statechart.Current_state {
	case "DiscoveryDecision":
		return "Run the Discovery Worthiness Rubric and decide whether to enable or skip discovery. Use `sow agent project phase enable discovery --type <type>` or `sow agent project phase skip discovery`."
	case "DiscoveryActive":
		return "Conduct research, create discovery artifacts, and request human approval when ready. Use `sow agent project artifact add` and `sow agent project phase complete discovery` when done."
	case "DesignDecision":
		return "Determine if design work is needed. Enable design with `sow agent project phase enable design` or skip with `sow agent project phase skip design`."
	case "DesignActive":
		return "Create ADRs and design documents. Use `sow agent project artifact add` and `sow agent project phase complete design` when done."
	case "ImplementationPlanning":
		return "Create a task breakdown using `sow agent task add`. When all tasks are defined, mark them approved to transition to execution."
	case "ImplementationExecuting":
		// Check tasks_approved in metadata
		tasksApproved := false
		if state.Phases.Implementation.Metadata != nil {
			if approved, ok := state.Phases.Implementation.Metadata["tasks_approved"].(bool); ok {
				tasksApproved = approved
			}
		}
		if tasksApproved {
			return "Spawn implementer agents via Task tool to execute pending tasks. Update task status as work completes."
		}
		return "Await human approval of the task plan before execution."
	case "ReviewActive":
		return "Perform quality review of completed work. Create review report with `sow agent project review add-report`. Approve to finalize or fail to loop back to implementation."
	case "FinalizeDocumentation":
		return "Update documentation files. Record updates with `sow agent project finalize add-document`. Transition when complete."
	case "FinalizeChecks":
		return "Run tests, linters, and builds. Verify all checks pass. Transition when validated."
	case "FinalizeDelete":
		return "Delete project folder with `sow agent project delete`. Create PR with `sow agent project create-pr`."
	default:
		return "Continue with the current phase work."
	}
}

// getCurrentPhaseDescription returns a human-readable current phase description.
func getCurrentPhaseDescription(state *schemas.ProjectState) string {
	currentPhase, phaseStatus := determineCurrentPhase(state)
	return fmt.Sprintf("%s phase (%s)", currentPhase, phaseStatus)
}

// generateNextActionSummary provides a brief summary of what's next.
func generateNextActionSummary(state *schemas.ProjectState) string {
	switch state.Statechart.Current_state {
	case "DiscoveryDecision":
		return "I'll assess if discovery is needed and proceed accordingly."
	case "DiscoveryActive":
		return "I'll continue discovery work."
	case "DesignDecision":
		return "I'll determine if design work is warranted."
	case "DesignActive":
		return "I'll continue design work."
	case "ImplementationPlanning":
		return "I'll work on the task breakdown."
	case "ImplementationExecuting":
		if len(state.Phases.Implementation.Tasks) > 0 {
			completed := 0
			for _, t := range state.Phases.Implementation.Tasks {
				if t.Status == "completed" {
					completed++
				}
			}
			if completed == len(state.Phases.Implementation.Tasks) {
				return "All tasks complete! Moving to review."
			}
			return fmt.Sprintf("I'll continue implementing tasks (%d/%d done).", completed, len(state.Phases.Implementation.Tasks))
		}
		return "I'll await task approval to begin execution."
	case "ReviewActive":
		return "I'll perform quality review of the work."
	case "FinalizeDocumentation":
		return "I'll update the necessary documentation."
	case "FinalizeChecks":
		return "I'll run final validation checks."
	case "FinalizeDelete":
		return "I'll clean up and create the PR."
	default:
		return "I'll continue where we left off."
	}
}
