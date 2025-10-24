package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewExploreCmd creates the explore command.
func NewExploreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explore [topic]",
		Short: "Start or resume an exploration mode session",
		Long: `Start or resume exploration mode for research and design work.

Exploration mode provides a guided environment for:
- Researching technologies and approaches
- Creating design documents and ADRs
- Documenting findings and comparisons
- Formalizing ideas into structured artifacts

Branch Management:
  - Creates branch: explore/{topic-kebab}
  - Smart resume: checks local → remote → creates new
  - One exploration per branch

Directory Structure:
  - .sow/exploration/        Exploration workspace
  - .sow/exploration/index.yaml    File index

The exploration index tracks all files with descriptions and tags for
context-aware loading and discoverability.

Examples:
  sow explore "authentication-approaches"
  sow explore --branch explore/auth-research    # Resume existing`,
		Args: cobra.MaximumNArgs(1),
		RunE: runExplore,
	}

	cmd.Flags().String("branch", "", "Branch name to resume (e.g., explore/auth-research)")

	return cmd
}

func runExplore(cmd *cobra.Command, args []string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized - run 'sow init' first")
	}

	branchFlag, _ := cmd.Flags().GetString("branch")

	var branchName string
	var topic string

	if branchFlag != "" {
		// Resume mode: use provided branch
		branchName = branchFlag

		// Validate branch name format
		if !strings.HasPrefix(branchName, "explore/") {
			return fmt.Errorf("branch name must start with 'explore/' (got: %s)", branchName)
		}

		// Extract topic from branch name
		topic = strings.TrimPrefix(branchName, "explore/")
	} else {
		// New exploration: require topic argument
		if len(args) == 0 {
			return fmt.Errorf("topic argument required when not using --branch")
		}

		topic = args[0]

		// Convert topic to kebab-case
		kebab := toKebabCase(topic)
		branchName = "explore/" + kebab
	}

	// Smart branch resolution
	if err := resolveBranch(ctx, branchName); err != nil {
		return err
	}

	// Initialize exploration if needed
	if !exploration.Exists(ctx) {
		if err := exploration.InitExploration(ctx, topic, branchName); err != nil {
			return fmt.Errorf("failed to initialize exploration: %w", err)
		}
		cmd.Printf("\n✓ Created new exploration: %s\n", topic)
		cmd.Printf("  Branch: %s\n", branchName)
	} else {
		// Load existing exploration to show info
		index, err := exploration.LoadIndex(ctx)
		if err != nil {
			return fmt.Errorf("failed to load exploration: %w", err)
		}
		cmd.Printf("\n✓ Resuming exploration: %s\n", index.Exploration.Topic)
		cmd.Printf("  Branch: %s\n", index.Exploration.Branch)
		cmd.Printf("  Status: %s\n", index.Exploration.Status)
		cmd.Printf("  Files:  %d\n", len(index.Files))
	}

	// Check claude CLI available
	claude := sowexec.NewLocal("claude")
	if !claude.Exists() {
		fmt.Fprintln(os.Stderr, "\nError: Claude Code CLI not found")
		fmt.Fprintln(os.Stderr, "Install from: https://claude.com/download")
		return fmt.Errorf("claude not found")
	}

	// Generate exploration mode prompt
	explorationPrompt, err := generateExplorationPrompt(ctx, topic, branchName)
	if err != nil {
		return fmt.Errorf("failed to generate exploration prompt: %w", err)
	}

	// Launch Claude Code with the exploration prompt
	cmd.Printf("\nLaunching Claude Code in exploration mode...\n\n")
	claudeCmd := exec.CommandContext(cmd.Context(), claude.Command(), explorationPrompt)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr
	claudeCmd.Dir = ctx.RepoRoot()

	// Run and wait for completion
	return claudeCmd.Run()
}

// resolveBranch implements smart branch resolution:
// 1. Check if local branch exists → checkout
// 2. Check if remote branch exists → fetch and create local tracking branch
// 3. Create new branch if neither exists
func resolveBranch(ctx *sow.Context, branchName string) error {
	git := ctx.Git()

	// Get current branch
	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Already on the branch
	if currentBranch == branchName {
		return nil
	}

	// Check if local branch exists
	branches, err := git.Branches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	localExists := false
	for _, b := range branches {
		if b == branchName {
			localExists = true
			break
		}
	}

	if localExists {
		// Local branch exists, just checkout
		if err := git.CheckoutBranch(branchName); err != nil {
			return fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
		}
		return nil
	}

	// Check if remote branch exists
	// TODO: This requires fetching remote refs
	// For now, just create new branch from current position

	// Get main branch to branch from
	mainBranch := "main"
	if currentBranch == "master" {
		mainBranch = "master"
	}

	// Check if we're on a protected branch
	if git.IsProtectedBranch(currentBranch) {
		// Create from current position (which is main/master)
		if err := createBranch(git, branchName); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branchName, err)
		}
		return nil
	}

	// Not on main/master, ask user to checkout main first
	return fmt.Errorf("branch %s not found. Please checkout %s first, then run this command again", branchName, mainBranch)
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
func generateExplorationPrompt(sowCtx *sow.Context, topic, branch string) (string, error) {
	// Load exploration index if it exists
	var files []prompts.ExplorationFile
	var status string = "active"

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
		Topic:  topic,
		Branch: branch,
		Status: status,
		Files:  files,
	}

	// Render prompt
	prompt, err := prompts.Render(prompts.PromptModeExplore, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to render exploration prompt: %w", err)
	}

	return prompt, nil
}
