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
  research    Deep research methodology and best practices

The output is designed to be consumed by AI agents, providing focused
guidance without overwhelming the initial context window.

Example:
  sow prompt research    # Output research methodology guidance`,
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
	default:
		return fmt.Errorf("unknown prompt type: %s\nAvailable types: research", promptType)
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
