package cmd

import (
	"fmt"
	"os"

	breakdowncmd "github.com/jmgilman/sow/cli/cmd/breakdown"
	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/modes"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewBreakdownCmd creates the breakdown command.
func NewBreakdownCmd() *cobra.Command {
	var branchName string

	cmd := &cobra.Command{
		Use:   "breakdown [prompt]",
		Short: "Start or resume a breakdown mode session",
		Long: `Start or resume breakdown mode for decomposing designs into work units.

Breakdown mode provides a guided environment for:
- Ingesting design and exploration documents
- Decomposing them into logical units of work
- Creating detailed specifications for each unit
- Publishing units as GitHub issues for sow projects

This command handles breakdown lifecycle based on context:

No arguments:
  - Uses current branch
  - If breakdown exists: continues it
  - If not: creates new breakdown (validates not on protected branch)

With [prompt]:
  - Provides initial context to the orchestrator
  - Useful for scoping the breakdown topic

With --branch:
  - Checks out the branch (creates if doesn't exist)
  - If breakdown exists: continues it
  - If not: creates new breakdown

Directory Structure:
  - .sow/breakdown/              Breakdown workspace
  - .sow/breakdown/index.yaml    Inputs and work units index
  - .sow/breakdown/units/        Detailed markdown for each work unit

The breakdown index tracks input sources, proposed work units with dependencies,
and published GitHub issues for zero-context resumability.

Claude Code Flags:
  Use -- to pass additional flags to the Claude Code CLI:
    sow breakdown "topic" -- --model opus --verbose

Examples:
  sow breakdown                                          # Continue or start in current branch
  sow breakdown "Break down auth system design"         # Start with context
  sow breakdown --branch breakdown/auth-implementation   # Work on specific branch
  sow breakdown "topic" -- --model opus                  # Start with specific Claude model`,
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
			return runBreakdown(cmd, branchName, initialPrompt)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to work on (creates if doesn't exist)")

	// Add management subcommands
	cmd.AddCommand(breakdowncmd.NewAddInputCmd())
	cmd.AddCommand(breakdowncmd.NewRemoveInputCmd())
	cmd.AddCommand(breakdowncmd.NewAddUnitCmd())
	cmd.AddCommand(breakdowncmd.NewUpdateUnitCmd())
	cmd.AddCommand(breakdowncmd.NewRemoveUnitCmd())
	cmd.AddCommand(breakdowncmd.NewCreateDocumentCmd())
	cmd.AddCommand(breakdowncmd.NewApproveUnitCmd())
	cmd.AddCommand(breakdowncmd.NewPublishCmd())
	cmd.AddCommand(breakdowncmd.NewSetStatusCmd())
	cmd.AddCommand(breakdowncmd.NewIndexCmd())

	return cmd
}

func runBreakdown(cmd *cobra.Command, branchName, initialPrompt string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Require sow to be initialized
	if !ctx.IsInitialized() {
		fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
		fmt.Fprintln(os.Stderr, "Run: sow init")
		return fmt.Errorf("not initialized")
	}

	// Extract Claude Code flags (everything after --)
	var claudeFlags []string
	if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
		claudeFlags = cmd.Flags().Args()[dashIndex:]
	}

	// Create mode runner
	mode := &breakdownMode{}
	runner := modes.NewModeRunner(
		mode,
		breakdown.Exists,
		breakdown.InitBreakdown,
		generateBreakdownPrompt,
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
		index, err := breakdown.LoadIndex(ctx)
		if err != nil {
			return fmt.Errorf("failed to load breakdown: %w", err)
		}
		modes.FormatResumptionMessage(cmd, mode, index.Breakdown.Topic, index.Breakdown.Branch, index.Breakdown.Status, map[string]int{
			"Inputs":     len(index.Inputs),
			"Work Units": len(index.Work_units),
		})
	}

	// Launch Claude Code
	return launchClaudeCode(cmd, ctx, result.Prompt, claudeFlags)
}

// breakdownMode implements the modes.Mode interface for breakdown mode.
type breakdownMode struct{}

func (m *breakdownMode) Name() string                    { return "breakdown" }
func (m *breakdownMode) BranchPrefix() string            { return "breakdown/" }
func (m *breakdownMode) DirectoryName() string           { return "breakdown" }
func (m *breakdownMode) IndexPath() string               { return "breakdown/index.yaml" }
func (m *breakdownMode) PromptID() prompts.PromptID      { return prompts.PromptModeBreakdown }
func (m *breakdownMode) ValidStatuses() []string         { return []string{"active", "completed", "abandoned"} }

// generateBreakdownPrompt creates the breakdown mode prompt with context.
func generateBreakdownPrompt(sowCtx *sow.Context, topic, branch, initialPrompt string) (string, error) {
	// Load breakdown index if it exists
	var inputs []prompts.BreakdownInput
	var workUnits []prompts.BreakdownWorkUnit
	status := "active"

	if breakdown.Exists(sowCtx) {
		index, err := breakdown.LoadIndex(sowCtx)
		if err != nil {
			return "", fmt.Errorf("failed to load breakdown index: %w", err)
		}

		// Convert schema inputs to prompt inputs
		for _, input := range index.Inputs {
			inputs = append(inputs, prompts.BreakdownInput{
				Type:        input.Type,
				Path:        input.Path,
				Description: input.Description,
				Tags:        input.Tags,
			})
		}

		// Convert schema work units to prompt work units
		for _, unit := range index.Work_units {
			workUnits = append(workUnits, prompts.BreakdownWorkUnit{
				ID:                unit.Id,
				Title:             unit.Title,
				Description:       unit.Description,
				DocumentPath:      unit.Document_path,
				Status:            unit.Status,
				DependsOn:         unit.Depends_on,
				GithubIssueURL:    unit.Github_issue_url,
				GithubIssueNumber: int(unit.Github_issue_number),
			})
		}

		status = index.Breakdown.Status
	}

	// Create breakdown context
	ctx := &prompts.BreakdownContext{
		Topic:         topic,
		Branch:        branch,
		Status:        status,
		Inputs:        inputs,
		WorkUnits:     workUnits,
		InitialPrompt: initialPrompt,
	}

	// Render prompt
	prompt, err := prompts.Render(prompts.PromptModeBreakdown, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to render breakdown prompt: %w", err)
	}

	return prompt, nil
}
