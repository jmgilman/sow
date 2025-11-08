package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/sow"
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

	// 8. Initialize project using shared utility
	proj, err := initializeProject(worktreeCtx, selectedBranch, description, issue)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "✓ Initialized project '%s' on branch %s\n", proj.Name, selectedBranch)

	// 12. Generate new project prompt
	prompt, err := generateNewProjectPrompt(proj, description)
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
		// Branch exists - return it (worktree will handle the checkout)
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
	if err := createBranch(git, branchName); err != nil {
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

// generateNewProjectPrompt, generateContinuePrompt, launchClaudeCode, and initializeProject
// have been moved to shared.go to support both the existing commands and the new wizard.

func createBranch(git *sow.Git, branchName string) error {
	// Get current HEAD
	head, err := git.Repository().Underlying().Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create branch reference (but don't check it out - worktree will do that)
	branchRef := "refs/heads/" + branchName
	if err := git.Repository().Underlying().Storer.SetReference(
		plumbing.NewHashReference(plumbing.ReferenceName(branchRef), head.Hash()),
	); err != nil {
		return fmt.Errorf("failed to create branch reference: %w", err)
	}

	return nil
}
