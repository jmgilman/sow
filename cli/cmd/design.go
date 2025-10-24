package cmd

import (
	"fmt"
	"os"

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

	// Create mode runner
	mode := &designMode{}
	runner := modes.NewModeRunner(
		mode,
		design.Exists,
		design.InitDesign,
		generateDesignPrompt,
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
		index, err := design.LoadIndex(ctx)
		if err != nil {
			return fmt.Errorf("failed to load design: %w", err)
		}
		modes.FormatResumptionMessage(cmd, mode, index.Design.Topic, index.Design.Branch, index.Design.Status, map[string]int{
			"Inputs":  len(index.Inputs),
			"Outputs": len(index.Outputs),
		})
	}

	// Launch Claude Code
	return launchClaudeCode(cmd, ctx, result.Prompt)
}

// designMode implements the modes.Mode interface for design mode.
type designMode struct{}

func (m *designMode) Name() string                    { return "design" }
func (m *designMode) BranchPrefix() string            { return "design/" }
func (m *designMode) DirectoryName() string           { return "design" }
func (m *designMode) IndexPath() string               { return "design/index.yaml" }
func (m *designMode) PromptID() prompts.PromptID      { return prompts.PromptModeDesign }
func (m *designMode) ValidStatuses() []string         { return []string{"active", "in_review", "completed"} }

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
