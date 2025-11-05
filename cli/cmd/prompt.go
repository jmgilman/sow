package cmd

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/templates"
	"github.com/spf13/cobra"
)

// NewPromptCmd creates the prompt command.
func NewPromptCmd() *cobra.Command {
	var listFlag bool

	cmd := &cobra.Command{
		Use:   "prompt <type>",
		Short: "Output guidance prompts for AI agents",
		Long: `Output guidance prompts to stdout for AI agent consumption.

This command provides on-demand context and guidance for specific tasks.
AI agents can invoke this during exploration or other modes to get
detailed best practices and methodology guidance.

Examples:
  sow prompt guidance/research
  sow prompt guidance/design/adr
  sow prompt guidance/implementer/base
  sow prompt guidance/implementer/tdd

To see all available prompts:
  sow prompt --list

The output is designed to be consumed by AI agents, providing focused
guidance without overwhelming the initial context window.`,
		Args: func(_ *cobra.Command, args []string) error {
			// If --list flag is set, no args required
			if listFlag {
				return nil
			}
			// Otherwise require exactly one arg
			if len(args) != 1 {
				return fmt.Errorf("requires exactly 1 arg when not using --list")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if listFlag {
				return listPrompts(cmd)
			}
			return runPrompt(cmd, args[0])
		},
	}

	cmd.Flags().BoolVar(&listFlag, "list", false, "List all available prompt types")

	return cmd
}

func runPrompt(cmd *cobra.Command, promptType string) error {
	// User provides the path exactly as shown in --list
	// ListTemplates already strips "templates/" prefix and ".md" suffix
	// So we just need to add them back
	// "guidance/research" -> "templates/guidance/research.md"
	templatePath := fmt.Sprintf("templates/%s.md", promptType)

	// Render the template with no context
	output, err := templates.Render(prompts.FS, templatePath, nil)
	if err != nil {
		// Provide helpful error with suggestion to list prompts
		return fmt.Errorf("prompt type '%s' not found\n\nRun 'sow prompt --list' to see available prompts", promptType)
	}

	cmd.Println(output)
	return nil
}

func listPrompts(cmd *cobra.Command) error {
	availablePrompts, err := templates.ListTemplates(prompts.FS, "templates")
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	cmd.Println("Available prompt types:")
	cmd.Println()

	// Only show guidance templates (not command/greet templates which require context)
	for _, p := range availablePrompts {
		// Filter to only guidance and modes (templates meant for on-demand fetching)
		if strings.HasPrefix(p, "guidance/") || strings.HasPrefix(p, "modes/") {
			cmd.Printf("  %s\n", p)
		}
	}

	return nil
}
