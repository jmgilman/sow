package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	designcmd "github.com/jmgilman/sow/cli/cmd/design"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/design"
	"github.com/jmgilman/sow/cli/internal/modes"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewDesignCmd creates the design command.
func NewDesignCmd() *cobra.Command {
	var branchName string
	var noLaunch bool

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

Claude Code Flags:
  Use -- to pass additional flags to the Claude Code CLI:
    sow design "topic" -- --model opus --verbose

Examples:
  sow design                                      # Continue or start in current branch
  sow design "Create auth system architecture"   # Start with context
  sow design --branch design/auth-system          # Work on specific branch
  sow design "topic" -- --model opus              # Start with specific Claude model`,
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
			return runDesign(cmd, branchName, initialPrompt, noLaunch)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to work on (creates if doesn't exist)")
	cmd.Flags().BoolVar(&noLaunch, "no-launch", false, "Setup worktree and design but don't launch Claude Code (for testing)")

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

func runDesign(cmd *cobra.Command, branchName, initialPrompt string, noLaunch bool) error {
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

	mode := &designMode{}
	runner := modes.NewModeRunner(
		mode,
		design.Exists,
		design.InitDesign,
		generateDesignPrompt,
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
	modeExistsInWorktree := design.Exists(worktreeCtx)

	// 8. Extract topic from branch name
	topic := modes.ExtractTopicFromBranch(mode, targetBranch)

	// 9. Initialize mode IN WORKTREE (if doesn't exist)
	if !modeExistsInWorktree {
		if err := runner.Initialize(worktreeCtx, topic, targetBranch); err != nil {
			return fmt.Errorf("failed to initialize design: %w", err)
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
		index, err := design.LoadIndex(worktreeCtx)
		if err != nil {
			return fmt.Errorf("failed to load design: %w", err)
		}
		modes.FormatResumptionMessage(cmd, mode, index.Design.Topic, index.Design.Branch, index.Design.Status, map[string]int{
			"Inputs":  len(index.Inputs),
			"Outputs": len(index.Outputs),
		})
	}

	// 12. Launch Claude Code from worktree directory (unless --no-launch)
	if noLaunch {
		return nil
	}
	return launchClaudeCode(cmd, worktreeCtx, prompt, claudeFlags)
}

// designMode implements the modes.Mode interface for design mode.
type designMode struct{}

func (m *designMode) Name() string               { return "design" }
func (m *designMode) BranchPrefix() string       { return "design/" }
func (m *designMode) DirectoryName() string      { return "design" }
func (m *designMode) IndexPath() string          { return "design/index.yaml" }
func (m *designMode) PromptID() prompts.PromptID { return prompts.PromptModeDesign }
func (m *designMode) ValidStatuses() []string    { return []string{"active", "in_review", "completed"} }

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
