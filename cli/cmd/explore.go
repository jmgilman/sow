package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/jmgilman/sow/cli/internal/modes"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewExploreCmd creates the explore command.
func NewExploreCmd() *cobra.Command {
	var branchName string
	var noLaunch bool

	cmd := &cobra.Command{
		Use:   "explore [prompt]",
		Short: "Start or resume an exploration mode session",
		Long: `Start or resume exploration mode for research and design work.

Exploration mode provides a guided environment for:
- Researching technologies and approaches
- Creating design documents and ADRs
- Documenting findings and comparisons
- Formalizing ideas into structured artifacts

This command handles exploration lifecycle based on context:

No arguments:
  - Uses current branch
  - If exploration exists: continues it
  - If not: creates new exploration (validates not on protected branch)

With [prompt]:
  - Provides initial context to the orchestrator
  - Useful for scoping the exploration topic

With --branch:
  - Checks out the branch (creates if doesn't exist)
  - If exploration exists: continues it
  - If not: creates new exploration

Directory Structure:
  - .sow/exploration/        Exploration workspace
  - .sow/exploration/index.yaml    File index

The exploration index tracks all files with descriptions and tags for
context-aware loading and discoverability.

Claude Code Flags:
  Use -- to pass additional flags to the Claude Code CLI:
    sow explore "topic" -- --model opus --verbose

Examples:
  sow explore                                    # Continue or start in current branch
  sow explore "Research JWT vs OAuth"            # Start with context
  sow explore --branch explore/auth-research     # Work on specific branch
  sow explore "topic" -- --model opus            # Start with specific Claude model`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Only validate args before -- separator
			argsBeforeDash := args
			if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
				argsBeforeDash = args[:dashIndex]
			}
			if len(argsBeforeDash) > 1 {
				return fmt.Errorf("accepts at most 1 arg(s), received %d", len(argsBeforeDash))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract prompt from args before -- separator
			argsBeforeDash := args
			if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
				argsBeforeDash = args[:dashIndex]
			}

			initialPrompt := ""
			if len(argsBeforeDash) > 0 {
				initialPrompt = argsBeforeDash[0]
			}
			return runExplore(cmd, branchName, initialPrompt, noLaunch)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to work on (creates if doesn't exist)")
	cmd.Flags().BoolVar(&noLaunch, "no-launch", false, "Setup worktree and exploration but don't launch Claude Code (for testing)")

	return cmd
}

func runExplore(cmd *cobra.Command, branchName, initialPrompt string, noLaunch bool) error {
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

	// 3. Determine target branch name (no git operations, just string logic)
	var targetBranch string
	if branchName != "" {
		targetBranch = branchName
	} else {
		// Use current branch
		currentBranch, err := mainCtx.Git().CurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		targetBranch = currentBranch
	}

	mode := &explorationMode{}
	runner := modes.NewModeRunner(
		mode,
		exploration.Exists,
		exploration.InitExploration,
		generateExplorationPrompt,
	)

	// 4. Generate worktree path and ensure directory exists
	worktreePath := sow.WorktreePath(mainCtx.RepoRoot(), targetBranch)
	worktreesDir := filepath.Join(mainCtx.RepoRoot(), ".sow", "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// 5. Ensure worktree exists
	if err := sow.EnsureWorktree(mainCtx, worktreePath, targetBranch); err != nil {
		return fmt.Errorf("failed to ensure worktree: %w", err)
	}

	// 6. Create worktree context
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to create worktree context: %w", err)
	}

	// 7. Check if mode exists in worktree (the definitive check)
	modeExistsInWorktree := exploration.Exists(worktreeCtx)

	// 8. Extract topic from branch name
	topic := modes.ExtractTopicFromBranch(mode, targetBranch)

	// 9. Initialize mode IN WORKTREE (if doesn't exist)
	if !modeExistsInWorktree {
		if err := runner.Initialize(worktreeCtx, topic, targetBranch); err != nil {
			return fmt.Errorf("failed to initialize exploration: %w", err)
		}
	}

	// 10. Generate prompt with worktree context
	prompt, err := runner.GeneratePrompt(worktreeCtx, topic, targetBranch, initialPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate prompt: %w", err)
	}

	// 11. Display appropriate message
	if !modeExistsInWorktree {
		modes.FormatCreationMessage(cmd, mode, topic, targetBranch)
	} else {
		// Load index for resumption message using worktree context
		index, err := exploration.LoadIndex(worktreeCtx)
		if err != nil {
			return fmt.Errorf("failed to load exploration: %w", err)
		}
		modes.FormatResumptionMessage(cmd, mode, index.Exploration.Topic, index.Exploration.Branch, index.Exploration.Status, map[string]int{
			"Files": len(index.Files),
		})
	}

	// 12. Launch Claude Code from worktree directory (unless --no-launch)
	if noLaunch {
		return nil
	}
	return launchClaudeCode(cmd, worktreeCtx, prompt, claudeFlags)
}

// explorationMode implements the modes.Mode interface for exploration mode.
type explorationMode struct{}

func (m *explorationMode) Name() string               { return "exploration" }
func (m *explorationMode) BranchPrefix() string       { return "explore/" }
func (m *explorationMode) DirectoryName() string      { return "exploration" }
func (m *explorationMode) IndexPath() string          { return "exploration/index.yaml" }
func (m *explorationMode) PromptID() prompts.PromptID { return prompts.PromptModeExplore }
func (m *explorationMode) ValidStatuses() []string {
	return []string{"active", "completed", "abandoned"}
}

// generateExplorationPrompt creates the exploration mode prompt with context.
func generateExplorationPrompt(sowCtx *sow.Context, topic, branch, initialPrompt string) (string, error) {
	// Load exploration index if it exists
	var files []prompts.ExplorationFile
	status := "active"

	if exploration.Exists(sowCtx) {
		index, err := exploration.LoadIndex(sowCtx)
		if err != nil {
			return "", fmt.Errorf("failed to load exploration index: %w", err)
		}

		// Convert schema files to prompt files
		for _, f := range index.Files {
			files = append(files, prompts.ExplorationFile{
				Path:        f.Path,
				Description: f.Description,
				Tags:        f.Tags,
			})
		}
		status = index.Exploration.Status
	}

	// Create exploration context
	ctx := &prompts.ExplorationContext{
		Topic:         topic,
		Branch:        branch,
		Status:        status,
		Files:         files,
		InitialPrompt: initialPrompt,
	}

	// Render prompt
	//nolint:staticcheck // Using legacy API during transition period
	prompt, err := prompts.Render(prompts.PromptModeExplore, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to render exploration prompt: %w", err)
	}

	return prompt, nil
}
