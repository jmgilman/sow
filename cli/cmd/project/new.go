package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/modes"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sow"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	var branchName string
	var issueNumber int
	var noLaunch bool

	cmd := &cobra.Command{
		Use:   "new [description]",
		Short: "Create a new project",
		Long: `Create a new project on a branch.

Creates a worktree for the branch if it doesn't exist, initializes a new project,
and launches Claude Code.

Examples:
  sow project new --branch feat/auth "Add JWT authentication"
  sow project new --issue 123
  sow project new "Quick bugfix" --no-launch

Claude Code Flags:
  Use -- to pass additional flags to the Claude Code CLI:
    sow project new "Description" -- --model opus --verbose`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Allow at most 1 arg before --, unlimited after --
			dashIndex := cmd.ArgsLenAtDash()
			if dashIndex < 0 {
				// No -- separator, check total args
				if len(args) > 1 {
					return fmt.Errorf("accepts at most 1 arg(s), received %d", len(args))
				}
			} else {
				// Has -- separator, only validate args before --
				if dashIndex > 1 {
					return fmt.Errorf("accepts at most 1 arg before --, received %d", dashIndex)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var description string
			// Only use args before the -- separator (args after -- are Claude Code flags)
			dashIndex := cmd.ArgsLenAtDash()
			if dashIndex < 0 {
				// No -- separator, use all args
				if len(args) > 0 {
					description = args[0]
				}
			} else {
				// Has -- separator, only use args before it
				if dashIndex > 0 {
					description = args[0]
				}
			}
			return runNew(cmd, branchName, issueNumber, description, noLaunch)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to create project on (creates if doesn't exist)")
	cmd.Flags().IntVar(&issueNumber, "issue", 0, "GitHub issue number to link to project")
	cmd.Flags().BoolVar(&noLaunch, "no-launch", false, "Setup project but don't launch Claude Code (for testing)")
	cmd.MarkFlagsMutuallyExclusive("branch", "issue")

	return cmd
}

func runNew(cmd *cobra.Command, branchName string, issueNumber int, description string, noLaunch bool) error {
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
		allArgs := cmd.Flags().Args()
		if dashIndex < len(allArgs) {
			claudeFlags = allArgs[dashIndex:]
		}
	}

	// 2. Check for uncommitted changes BEFORE any branch operations
	if err := sow.CheckUncommittedChanges(mainCtx); err != nil {
		return fmt.Errorf("cannot create worktree: %w", err)
	}

	// 3. Determine target branch and issue
	selectedBranch, issue, description, err := determineBranchAndDescription(mainCtx, issueNumber, branchName, description)
	if err != nil {
		return err
	}

	// 4. Validate not on protected branch
	if mainCtx.Git().IsProtectedBranch(selectedBranch) {
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

	// Validate worktree is initialized
	if !worktreeCtx.IsInitialized() {
		return fmt.Errorf("worktree at %s is not initialized - .sow/ directory not found.\n"+
			"This usually means .sow/ is not committed on branch %s.\n"+
			"Run 'git add .sow && git commit' before creating a project",
			worktreePath, selectedBranch)
	}

	// 8. Ensure .sow/project directory exists in worktree
	projectDir := filepath.Join(worktreePath, ".sow", "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// 9. Always create context directory for project files
	contextDir := filepath.Join(projectDir, "context")
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return fmt.Errorf("failed to create context directory: %w", err)
	}

	// 10. Prepare initial inputs if issue is provided
	var initialInputs map[string][]projschema.ArtifactState
	if issue != nil {

		// Write issue body to file
		issueFileName := fmt.Sprintf("issue-%d.md", issue.Number)
		issuePath := filepath.Join(contextDir, issueFileName)
		issueContent := fmt.Sprintf("# Issue #%d: %s\n\n**URL**: %s\n**State**: %s\n\n## Description\n\n%s\n",
			issue.Number, issue.Title, issue.URL, issue.State, issue.Body)

		if err := os.WriteFile(issuePath, []byte(issueContent), 0644); err != nil {
			return fmt.Errorf("failed to write issue file: %w", err)
		}

		// Create github_issue artifact for implementation phase
		issueArtifact := projschema.ArtifactState{
			Type:       "github_issue",
			Path:       fmt.Sprintf("context/%s", issueFileName),
			Approved:   true, // Auto-approved
			Created_at: time.Now(),
			Metadata: map[string]interface{}{
				"issue_number": issue.Number,
				"issue_url":    issue.URL,
				"issue_title":  issue.Title,
			},
		}

		initialInputs = map[string][]projschema.ArtifactState{
			"implementation": {issueArtifact},
		}
	}

	// 11. Create project using SDK with initial inputs
	proj, err := state.Create(worktreeCtx, selectedBranch, description, initialInputs)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Initialized project '%s' on branch %s\n", proj.Name, selectedBranch)

	// 12. Generate new project prompt
	prompt, err := generateNewProjectPrompt(worktreeCtx, proj, description)
	if err != nil {
		return fmt.Errorf("failed to generate new project prompt: %w", err)
	}

	// 13. Launch Claude Code from worktree directory (unless --no-launch)
	if noLaunch {
		return nil
	}
	return launchClaudeCode(cmd, worktreeCtx, prompt, claudeFlags)
}

// determineBranchAndDescription determines the target branch and description based on flags.
// Returns: (selectedBranch, issue, description, error).
func determineBranchAndDescription(
	mainCtx *sow.Context,
	issueNumber int,
	branchName string,
	description string,
) (string, *sow.Issue, string, error) {
	var selectedBranch string
	var issue *sow.Issue
	var err error

	// Scenario 1: --issue flag provided
	if issueNumber > 0 {
		selectedBranch, issue, err = handleIssueScenarioNew(mainCtx, issueNumber)
		if err != nil {
			return "", nil, "", err
		}
		// Use issue title as description if not provided
		if description == "" {
			description = issue.Title
		}
		return selectedBranch, issue, description, nil
	}

	// Scenario 2: --branch flag provided
	if branchName != "" {
		selectedBranch, err = handleBranchScenarioNew(mainCtx, branchName)
		if err != nil {
			return "", nil, "", err
		}
		if description == "" {
			return "", nil, "", fmt.Errorf("description required: provide as argument or via --issue flag")
		}
		return selectedBranch, nil, description, nil
	}

	// Scenario 3: No flags (current branch)
	selectedBranch, err = handleCurrentBranchScenarioNew(mainCtx)
	if err != nil {
		return "", nil, "", err
	}
	if description == "" {
		return "", nil, "", fmt.Errorf("description required: provide as argument or via --issue flag")
	}
	return selectedBranch, nil, description, nil
}

// handleIssueScenarioNew handles the --issue flag scenario for new command.
// Returns: (branchName, issue, error).
func handleIssueScenarioNew(_ *sow.Context, issueNumber int) (string, *sow.Issue, error) {
	ghExec := sowexec.NewLocal("gh")
	gh := sow.NewGitHub(ghExec)

	// Fetch issue
	issue, err := gh.GetIssue(issueNumber)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch issue #%d: %w", issueNumber, err)
	}

	// Validate has 'sow' label
	if !issue.HasLabel("sow") {
		return "", nil, fmt.Errorf("issue #%d does not have the 'sow' label - add it with: gh issue edit %d --add-label sow", issueNumber, issueNumber)
	}

	// Check for linked branches
	branches, err := gh.GetLinkedBranches(issueNumber)
	if err != nil {
		return "", nil, fmt.Errorf("failed to check linked branches: %w", err)
	}

	if len(branches) > 0 {
		// Branch already exists - user should use 'continue' instead
		return "", nil, fmt.Errorf("issue #%d already has a linked branch: %s\nTo continue this project, use: sow project continue --branch %s", issueNumber, branches[0].Name, branches[0].Name)
	}

	// No linked branch - create new branch
	branchName, err := gh.CreateLinkedBranch(issueNumber, "", true)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create linked branch: %w", err)
	}

	return branchName, issue, nil
}

// handleBranchScenarioNew handles the --branch flag scenario for new command.
// Returns: (branchName, error).
func handleBranchScenarioNew(ctx *sow.Context, branchName string) (string, error) {
	git := ctx.Git()

	// Check if branch exists locally
	branches, err := git.Branches()
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %w", err)
	}

	branchExists := false
	for _, b := range branches {
		if b == branchName {
			branchExists = true
			break
		}
	}

	if branchExists {
		// Branch exists - checkout
		if err := git.CheckoutBranch(branchName); err != nil {
			return "", fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
		}
		return branchName, nil
	}

	// Create new branch from current branch
	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	// Must be on protected branch to create new branch
	if !git.IsProtectedBranch(currentBranch) {
		return "", fmt.Errorf("cannot create branch %s from %s - please checkout main/master first", branchName, currentBranch)
	}

	// Create and checkout new branch
	if err := modes.CreateBranch(git, branchName); err != nil {
		return "", fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	return branchName, nil
}

// handleCurrentBranchScenarioNew handles the no-flags scenario (current branch) for new command.
// Returns: (branchName, error).
func handleCurrentBranchScenarioNew(ctx *sow.Context) (string, error) {
	git := ctx.Git()

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	// Validate we're not on a protected branch
	if git.IsProtectedBranch(currentBranch) {
		return "", fmt.Errorf("cannot create project on protected branch '%s' - create a feature branch first", currentBranch)
	}

	return currentBranch, nil
}

// generateNewProjectPrompt creates the custom prompt for new projects.
// Uses 3-layer structure: Base Orchestrator + Project Type Orchestrator + Initial State.
func generateNewProjectPrompt(ctx *sow.Context, proj *state.Project, initialPrompt string) (string, error) {
	var buf strings.Builder

	// Layer 1: Base Orchestrator Introduction
	baseCtx := &prompts.GreetContext{
		SowInitialized: ctx.IsInitialized(),
		HasProject:     true,
	}

	baseOrch, err := prompts.Render(prompts.PromptGreetOrchestrator, baseCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render base orchestrator prompt: %w", err)
	}
	buf.WriteString(baseOrch)
	buf.WriteString("\n\n---\n\n")

	// Layer 2: Project Type Orchestrator Prompt
	projectTypePrompt := proj.Config().OrchestratorPrompt(proj)
	if projectTypePrompt != "" {
		buf.WriteString(projectTypePrompt)
		buf.WriteString("\n\n---\n\n")
	}

	// Layer 3: Initial State Prompt
	initialState := proj.Machine().State()
	statePrompt := proj.Config().GetStatePrompt(initialState, proj)
	if statePrompt != "" {
		buf.WriteString(statePrompt)
		buf.WriteString("\n\n---\n\n")
	}

	// Add initial user prompt if provided
	if initialPrompt != "" {
		buf.WriteString("## User's Initial Request\n\n")
		buf.WriteString(initialPrompt)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// launchClaudeCode is a placeholder that should call the shared helper.
// For now, inline the implementation to avoid import cycle.
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
