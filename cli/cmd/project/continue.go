// Package project provides commands for managing project lifecycle.
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sdks/project/templates"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

func newContinueCmd() *cobra.Command {
	var branchName string
	var noLaunch bool

	cmd := &cobra.Command{
		Use:   "continue",
		Short: "Continue an existing project",
		Long: `Continue work on an existing project.

Loads the project state, generates a continue prompt, and launches Claude Code.

Examples:
  sow project continue
  sow project continue --branch feat/auth
  sow project continue --no-launch

Claude Code Flags:
  Use -- to pass additional flags to the Claude Code CLI:
    sow project continue -- --model opus --verbose`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runContinue(cmd, branchName, noLaunch)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to continue project on")
	cmd.Flags().BoolVar(&noLaunch, "no-launch", false, "Don't launch Claude Code (for testing)")

	return cmd
}

func runContinue(cmd *cobra.Command, branchName string, noLaunch bool) error {
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

	// 2. Determine target branch
	var selectedBranch string
	var err error

	if branchName != "" {
		selectedBranch, err = handleBranchScenarioContinue(mainCtx, branchName)
		if err != nil {
			return err
		}
	} else {
		selectedBranch, err = handleCurrentBranchScenarioContinue(mainCtx)
		if err != nil {
			return err
		}
	}

	// 3. Generate worktree path
	worktreePath := sow.WorktreePath(mainCtx.RepoRoot(), selectedBranch)
	worktreesDir := filepath.Join(mainCtx.RepoRoot(), ".sow", "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// 4. Ensure worktree exists
	if err := sow.EnsureWorktree(mainCtx, worktreePath, selectedBranch); err != nil {
		return fmt.Errorf("failed to ensure worktree: %w", err)
	}

	// 5. Create worktree context
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to create worktree context: %w", err)
	}

	// 6. Load existing project using SDK
	proj, err := state.Load(worktreeCtx)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Continuing project '%s' on branch %s\n", proj.Name, selectedBranch)

	// 7. Generate continue prompt
	prompt, err := generateContinuePrompt(proj)
	if err != nil {
		return fmt.Errorf("failed to generate continue prompt: %w", err)
	}

	// 8. Launch Claude Code from worktree directory (unless --no-launch)
	if noLaunch {
		return nil
	}
	return launchClaudeCode(cmd, worktreeCtx, prompt, claudeFlags)
}

// handleBranchScenarioContinue handles the --branch flag scenario for continue command.
// Returns: (branchName, error).
func handleBranchScenarioContinue(ctx *sow.Context, branchName string) (string, error) {
	git := ctx.Git()

	// Check if branch exists
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

	if !branchExists {
		return "", fmt.Errorf("branch %s does not exist - create it first with: sow project new --branch %s \"description\"", branchName, branchName)
	}

	// Note: We do NOT checkout the branch in the main repository.
	// The worktree will be created/accessed separately via EnsureWorktree(),
	// which handles the branch checkout in its own isolated working directory.
	// Checking out here would fail if the main repo has uncommitted changes.

	return branchName, nil
}

// handleCurrentBranchScenarioContinue handles the no-flags scenario (current branch) for continue command.
// Returns: (branchName, error).
func handleCurrentBranchScenarioContinue(ctx *sow.Context) (string, error) {
	git := ctx.Git()

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return currentBranch, nil
}

// generateContinuePrompt creates the custom prompt for continuing projects.
// Uses 3-layer structure: Base Orchestrator + Project Type Orchestrator + Current State.
func generateContinuePrompt(proj *state.Project) (string, error) {
	var buf strings.Builder

	// Layer 1: Base Orchestrator Introduction
	baseOrch, err := templates.Render(prompts.FS, "templates/greet/orchestrator.md", nil)
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

	// Layer 3: Current State Prompt
	currentState := proj.Machine().State()
	statePrompt := proj.Config().GetStatePrompt(currentState, proj)
	if statePrompt != "" {
		buf.WriteString(statePrompt)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
