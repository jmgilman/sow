package breakdown

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewAddInputCmd creates the breakdown add-input command.
func NewAddInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-input <path>",
		Short: "Add an input to the breakdown index",
		Long: `Add an input source to the current breakdown's index with description and tags.

The path can be:
- A specific file path
- A directory path
- A glob pattern (e.g., "docs/*.md")

This registers the input in the index for context management and resumability.

Input types:
  design      - Design documents (ADRs, architecture docs)
  exploration - Exploration artifacts
  file        - Existing codebase or documentation files
  reference   - External references or examples
  url         - Web resources
  git         - Other repositories or projects

Requirements:
  - Must be in a sow repository with an active breakdown session
  - Input must not already be in the index

Examples:
  # Add design document
  sow breakdown add-input .sow/design/auth-architecture.md \
    --type design \
    --description "Authentication system architecture" \
    --tags "auth,design"

  # Add exploration directory
  sow breakdown add-input .sow/exploration/ \
    --type exploration \
    --description "OAuth research findings" \
    --tags "oauth,research"

  # Add specific file
  sow breakdown add-input docs/current-implementation.md \
    --type file \
    --description "Current implementation documentation" \
    --tags "current-state"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddInput(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("type", "t", "", "Input type (required: design, exploration, file, reference, url, git)")
	cmd.Flags().StringP("description", "d", "", "Brief description of what this input provides (required)")
	cmd.Flags().StringSliceP("tags", "", nil, "Comma-separated tags for organization")

	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runAddInput(cmd *cobra.Command, args []string) error {
	path := args[0]
	inputType, _ := cmd.Flags().GetString("type")
	description, _ := cmd.Flags().GetString("description")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	// Trim tags
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	// Validate input type
	validTypes := map[string]bool{
		"design":      true,
		"exploration": true,
		"file":        true,
		"reference":   true,
		"url":         true,
		"git":         true,
	}
	if !validTypes[inputType] {
		return fmt.Errorf("invalid input type: %s (must be design, exploration, file, reference, url, or git)", inputType)
	}

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Add input to index
	if err := breakdown.AddInput(ctx, inputType, path, description, tags); err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return fmt.Errorf("no active breakdown session - run 'sow breakdown <topic>' first")
		}
		if errors.Is(err, breakdown.ErrInputExists) {
			return fmt.Errorf("input %s already exists in breakdown index", path)
		}
		return fmt.Errorf("failed to add input: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Added input to breakdown index: %s\n", path)
	cmd.Printf("\nInput Details:\n")
	cmd.Printf("  Type:        %s\n", inputType)
	cmd.Printf("  Path:        %s\n", path)
	cmd.Printf("  Description: %s\n", description)
	if len(tags) > 0 {
		cmd.Printf("  Tags:        %s\n", strings.Join(tags, ", "))
	} else {
		cmd.Printf("  Tags:        none\n")
	}

	return nil
}
