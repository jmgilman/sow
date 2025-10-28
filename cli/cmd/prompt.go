package cmd

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/spf13/cobra"
)

// NewPromptCmd creates the prompt command.
func NewPromptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompt <type>",
		Short: "Output guidance prompts for AI agents",
		Long: `Output guidance prompts to stdout for AI agent consumption.

This command provides on-demand context and guidance for specific tasks.
AI agents can invoke this during exploration or other modes to get
detailed best practices and methodology guidance.

Available prompt types:
  research              Deep research methodology and best practices
  design/prd            Product Requirements Document template
  design/arc42          Arc42 architecture documentation structure
  design/design-doc     Design document templates (mini/standard/comprehensive)
  design/adr            Architecture Decision Record template
  design/c4-diagrams    C4 model diagrams with Mermaid examples

The output is designed to be consumed by AI agents, providing focused
guidance without overwhelming the initial context window.

Examples:
  sow prompt research        # Output research methodology guidance
  sow prompt design/adr      # Output ADR template and best practices
  sow prompt design/arc42    # Output Arc42 structure guidance`,
		Args: cobra.ExactArgs(1),
		RunE: runPrompt,
	}

	return cmd
}

func runPrompt(cmd *cobra.Command, args []string) error {
	promptType := args[0]

	// Map prompt type to prompt ID
	var promptID prompts.PromptID
	switch promptType {
	case "research":
		promptID = prompts.PromptGuidanceResearch

	// Design guidance prompts
	case "design/prd":
		promptID = prompts.PromptGuidanceDesignPRD
	case "design/arc42":
		promptID = prompts.PromptGuidanceDesignArc42
	case "design/design-doc":
		promptID = prompts.PromptGuidanceDesignDoc
	case "design/adr":
		promptID = prompts.PromptGuidanceDesignADR
	case "design/c4-diagrams":
		promptID = prompts.PromptGuidanceDesignC4Diagrams

	default:
		return fmt.Errorf("unknown prompt type: %s\n\nAvailable types:\n  research\n  design/prd\n  design/arc42\n  design/design-doc\n  design/adr\n  design/c4-diagrams", promptType)
	}

	// Create context (guidance prompts currently don't need context)
	ctx := &prompts.GuidanceContext{}

	// Render prompt
	output, err := prompts.Render(promptID, ctx)
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	// Output to stdout (for agent consumption)
	cmd.Println(output)

	return nil
}
