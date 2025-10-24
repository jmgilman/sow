package cmd

import (
	"fmt"
	"os"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	designcmd "github.com/jmgilman/sow/cli/cmd/design"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/design"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewDesignCmd creates the design command.
func NewDesignCmd() *cobra.Command {
	var branchName string

	cmd := &cobra.Command{
		Use:   "design [prompt]",
		Short: "Start or resume a design mode session",
		Long: `Start or resume design mode for creating formal design documents.

Design mode provides a guided environment for:
- Synthesizing research findings into formal documents
- Creating ADRs (Architecture Decision Records)
- Documenting architecture and system design
- Producing diagrams and specifications
- Collaborating with stakeholders on design decisions

This command handles design lifecycle based on context:

No arguments:
  - Uses current branch
  - If design exists: continues it
  - If not: creates new design (validates not on protected branch)

With [prompt]:
  - Provides initial context to the orchestrator
  - Useful for scoping the design topic

With --branch:
  - Checks out the branch (creates if doesn't exist)
  - If design exists: continues it
  - If not: creates new design

Directory Structure:
  - .sow/design/              Design workspace
  - .sow/design/index.yaml    Input/output index

The design index tracks:
- Input sources (explorations, files, references) that inform the design
- Planned outputs (documents) with their target locations

Examples:
  sow design                                      # Continue or start in current branch
  sow design "Create auth system architecture"   # Start with context
  sow design --branch design/auth-system          # Work on specific branch`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			initialPrompt := ""
			if len(args) > 0 {
				initialPrompt = args[0]
			}
			return runDesign(cmd, branchName, initialPrompt)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to work on (creates if doesn't exist)")

	// Add management subcommands
	cmd.AddCommand(designcmd.NewAddInputCmd())
	cmd.AddCommand(designcmd.NewRemoveInputCmd())
	cmd.AddCommand(designcmd.NewAddOutputCmd())
	cmd.AddCommand(designcmd.NewRemoveOutputCmd())
	cmd.AddCommand(designcmd.NewSetOutputTargetCmd())
	cmd.AddCommand(designcmd.NewSetStatusCmd())
	cmd.AddCommand(designcmd.NewIndexCmd())

	return cmd
}

func runDesign(cmd *cobra.Command, branchName, initialPrompt string) error {
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
		selectedBranch, topic, shouldCreateNew, err = handleDesignBranchScenario(ctx, branchName)
		if err != nil {
			return err
		}
	} else {
		// Scenario: No flags (current branch)
		selectedBranch, topic, shouldCreateNew, err = handleDesignCurrentBranchScenario(ctx)
		if err != nil {
			return err
		}
	}

	// At this point we know:
	// - selectedBranch: which branch we're on
	// - topic: the design topic
	// - shouldCreateNew: whether to create new or continue

	if shouldCreateNew {
		// Create new design
		if err := design.InitDesign(ctx, topic, selectedBranch); err != nil {
			return fmt.Errorf("failed to initialize design: %w", err)
		}
		cmd.Printf("\n✓ Created new design session: %s\n", topic)
		cmd.Printf("  Branch: %s\n", selectedBranch)
	} else {
		// Continue existing design
		index, err := design.LoadIndex(ctx)
		if err != nil {
			return fmt.Errorf("failed to load design: %w", err)
		}
		cmd.Printf("\n✓ Resuming design: %s\n", index.Design.Topic)
		cmd.Printf("  Branch: %s\n", index.Design.Branch)
		cmd.Printf("  Status: %s\n", index.Design.Status)
		cmd.Printf("  Inputs:  %d\n", len(index.Inputs))
		cmd.Printf("  Outputs: %d\n", len(index.Outputs))

		// Use the topic from the existing design
		topic = index.Design.Topic
	}

	// Generate design mode prompt
	designPrompt, err := generateDesignPrompt(ctx, topic, selectedBranch, initialPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate design prompt: %w", err)
	}

	return launchClaudeCode(cmd, ctx, designPrompt)
}

// handleDesignBranchScenario handles the --branch flag scenario.
// Returns: (branchName, topic, shouldCreateNew, error).
func handleDesignBranchScenario(ctx *sow.Context, branchName string) (string, string, bool, error) {
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
		if err := createDesignBranch(git, branchName); err != nil {
			return "", "", false, fmt.Errorf("failed to create branch %s: %w", branchName, err)
		}
	}

	// Extract topic from branch name
	// If it starts with design/, strip that prefix, otherwise use the whole name
	topic := branchName
	if strings.HasPrefix(branchName, "design/") {
		topic = strings.TrimPrefix(branchName, "design/")
	}

	// Check if design exists in this branch
	designExists := design.Exists(ctx)

	return branchName, topic, !designExists, nil
}

// handleDesignCurrentBranchScenario handles the no-flags scenario (current branch).
// Returns: (branchName, topic, shouldCreateNew, error).
func handleDesignCurrentBranchScenario(ctx *sow.Context) (string, string, bool, error) {
	git := ctx.Git()

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return "", "", false, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if design exists
	designExists := design.Exists(ctx)

	if !designExists {
		// Validate we're not on a protected branch before creating new
		if git.IsProtectedBranch(currentBranch) {
			return "", "", false, fmt.Errorf("cannot create design on protected branch '%s' - create a branch first", currentBranch)
		}
	}

	// Extract topic from branch name
	topic := currentBranch
	if strings.HasPrefix(currentBranch, "design/") {
		topic = strings.TrimPrefix(currentBranch, "design/")
	}

	return currentBranch, topic, !designExists, nil
}

// createDesignBranch creates a new branch and checks it out.
func createDesignBranch(git *sow.Git, branchName string) error {
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

// generateDesignPrompt creates the design mode prompt with context.
func generateDesignPrompt(sowCtx *sow.Context, topic, branch, initialPrompt string) (string, error) {
	// Load design index if it exists
	var inputs []prompts.DesignInput
	var outputs []prompts.DesignOutput
	status := "active"

	if design.Exists(sowCtx) {
		index, err := design.LoadIndex(sowCtx)
		if err != nil {
			return "", fmt.Errorf("failed to load design index: %w", err)
		}

		// Convert schema inputs to prompt inputs
		for _, input := range index.Inputs {
			inputs = append(inputs, prompts.DesignInput{
				Type:        input.Type,
				Path:        input.Path,
				Description: input.Description,
				Tags:        input.Tags,
			})
		}

		// Convert schema outputs to prompt outputs
		for _, output := range index.Outputs {
			outputs = append(outputs, prompts.DesignOutput{
				Path:           output.Path,
				Description:    output.Description,
				TargetLocation: output.Target_location,
				Type:           output.Type,
				Tags:           output.Tags,
			})
		}

		status = index.Design.Status
	}

	// Create design context
	ctx := &prompts.DesignContext{
		Topic:         topic,
		Branch:        branch,
		Status:        status,
		Inputs:        inputs,
		Outputs:       outputs,
		InitialPrompt: initialPrompt,
	}

	// Render prompt
	prompt, err := prompts.Render(prompts.PromptModeDesign, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to render design prompt: %w", err)
	}

	return prompt, nil
}
