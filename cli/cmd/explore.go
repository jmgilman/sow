package cmd

import (
	"fmt"
	"os"

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

	// Create mode runner
	mode := &explorationMode{}
	runner := modes.NewModeRunner(
		mode,
		exploration.Exists,
		exploration.InitExploration,
		generateExplorationPrompt,
	)

	// Run the mode
	result, err := runner.Run(ctx, branchName, initialPrompt)
	if err != nil {
		return err
	}

	// Display appropriate message
	if result.ShouldCreateNew {
		modes.FormatCreationMessage(cmd, mode, result.Topic, result.SelectedBranch)
	} else {
		// Load index for resumption message
		index, err := exploration.LoadIndex(ctx)
		if err != nil {
			return fmt.Errorf("failed to load exploration: %w", err)
		}
		modes.FormatResumptionMessage(cmd, mode, index.Exploration.Topic, index.Exploration.Branch, index.Exploration.Status, map[string]int{
			"Files": len(index.Files),
		})
	}

	// Launch Claude Code
	return launchClaudeCode(cmd, ctx, result.Prompt)
}

// explorationMode implements the modes.Mode interface for exploration mode.
type explorationMode struct{}

func (m *explorationMode) Name() string                    { return "exploration" }
func (m *explorationMode) BranchPrefix() string            { return "explore/" }
func (m *explorationMode) DirectoryName() string           { return "exploration" }
func (m *explorationMode) IndexPath() string               { return "exploration/index.yaml" }
func (m *explorationMode) PromptID() prompts.PromptID      { return prompts.PromptModeExplore }
func (m *explorationMode) ValidStatuses() []string         { return []string{"active", "completed", "abandoned"} }

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
