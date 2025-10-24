package cmd

import (
	"fmt"
	"os"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewExploreCmd creates the explore command.
func NewExploreCmd() *cobra.Command {
	var branchName string

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

Examples:
  sow explore                                    # Continue or start in current branch
  sow explore "Research JWT vs OAuth"            # Start with context
  sow explore --branch explore/auth-research     # Work on specific branch`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			initialPrompt := ""
			if len(args) > 0 {
				initialPrompt = args[0]
			}
			return runExplore(cmd, branchName, initialPrompt)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to work on (creates if doesn't exist)")

	return cmd
}

func runExplore(cmd *cobra.Command, branchName, initialPrompt string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Require sow to be initialized
	if !ctx.IsInitialized() {
		fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
		fmt.Fprintln(os.Stderr, "Run: sow init")
		return fmt.Errorf("not initialized")
	}

	var selectedBranch string
	var topic string
	var shouldCreateNew bool
	var err error

	if branchName != "" {
		// Scenario: --branch flag provided
		selectedBranch, topic, shouldCreateNew, err = handleExploreBranchScenario(ctx, branchName)
		if err != nil {
			return err
		}
	} else {
		// Scenario: No flags (current branch)
		selectedBranch, topic, shouldCreateNew, err = handleExploreCurrentBranchScenario(ctx)
		if err != nil {
			return err
		}
	}

	// At this point we know:
	// - selectedBranch: which branch we're on
	// - topic: the exploration topic
	// - shouldCreateNew: whether to create new or continue

	if shouldCreateNew {
		// Create new exploration
		if err := exploration.InitExploration(ctx, topic, selectedBranch); err != nil {
			return fmt.Errorf("failed to initialize exploration: %w", err)
		}
		cmd.Printf("\n✓ Created new exploration: %s\n", topic)
		cmd.Printf("  Branch: %s\n", selectedBranch)
	} else {
		// Continue existing exploration
		index, err := exploration.LoadIndex(ctx)
		if err != nil {
			return fmt.Errorf("failed to load exploration: %w", err)
		}
		cmd.Printf("\n✓ Resuming exploration: %s\n", index.Exploration.Topic)
		cmd.Printf("  Branch: %s\n", index.Exploration.Branch)
		cmd.Printf("  Status: %s\n", index.Exploration.Status)
		cmd.Printf("  Files:  %d\n", len(index.Files))

		// Use the topic from the existing exploration
		topic = index.Exploration.Topic
	}

	// Generate exploration mode prompt
	explorationPrompt, err := generateExplorationPrompt(ctx, topic, selectedBranch, initialPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate exploration prompt: %w", err)
	}

	return launchClaudeCode(cmd, ctx, explorationPrompt)
}

// handleExploreBranchScenario handles the --branch flag scenario.
// Returns: (branchName, topic, shouldCreateNew, error).
func handleExploreBranchScenario(ctx *sow.Context, branchName string) (string, string, bool, error) {
	git := ctx.Git()

	// Check if branch exists locally
	branches, err := git.Branches()
	if err != nil {
		return "", "", false, fmt.Errorf("failed to list branches: %w", err)
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
			return "", "", false, fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
		}
	} else {
		// Create new branch
		// First check we're on a safe branch to create from
		currentBranch, err := git.CurrentBranch()
		if err != nil {
			return "", "", false, fmt.Errorf("failed to get current branch: %w", err)
		}

		if !git.IsProtectedBranch(currentBranch) {
			return "", "", false, fmt.Errorf("cannot create branch %s from %s - please checkout main/master first", branchName, currentBranch)
		}

		// Create and checkout new branch
		if err := createBranch(git, branchName); err != nil {
			return "", "", false, fmt.Errorf("failed to create branch %s: %w", branchName, err)
		}
	}

	// Extract topic from branch name
	// If it starts with explore/, strip that prefix, otherwise use the whole name
	topic := branchName
	if strings.HasPrefix(branchName, "explore/") {
		topic = strings.TrimPrefix(branchName, "explore/")
	}

	// Check if exploration exists in this branch
	explorationExists := exploration.Exists(ctx)

	return branchName, topic, !explorationExists, nil
}

// handleExploreCurrentBranchScenario handles the no-flags scenario (current branch).
// Returns: (branchName, topic, shouldCreateNew, error).
func handleExploreCurrentBranchScenario(ctx *sow.Context) (string, string, bool, error) {
	git := ctx.Git()

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return "", "", false, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if exploration exists
	explorationExists := exploration.Exists(ctx)

	if !explorationExists {
		// Validate we're not on a protected branch before creating new
		if git.IsProtectedBranch(currentBranch) {
			return "", "", false, fmt.Errorf("cannot create exploration on protected branch '%s' - create a branch first", currentBranch)
		}
	}

	// Extract topic from branch name
	topic := currentBranch
	if strings.HasPrefix(currentBranch, "explore/") {
		topic = strings.TrimPrefix(currentBranch, "explore/")
	}

	return currentBranch, topic, !explorationExists, nil
}


// createBranch creates a new branch and checks it out.
func createBranch(git *sow.Git, branchName string) error {
	// Use underlying go-git to create branch
	wt, err := git.Repository().Underlying().Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get current HEAD
	head, err := git.Repository().Underlying().Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create branch reference
	branchRef := "refs/heads/" + branchName
	if err := git.Repository().Underlying().Storer.SetReference(
		plumbing.NewHashReference(plumbing.ReferenceName(branchRef), head.Hash()),
	); err != nil {
		return fmt.Errorf("failed to create branch reference: %w", err)
	}

	// Checkout the new branch
	if err := wt.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.ReferenceName(branchRef),
	}); err != nil {
		return fmt.Errorf("failed to checkout new branch: %w", err)
	}

	return nil
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
	prompt, err := prompts.Render(prompts.PromptModeExplore, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to render exploration prompt: %w", err)
	}

	return prompt, nil
}
