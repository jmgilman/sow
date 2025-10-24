package design

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/design"
	"github.com/spf13/cobra"
)

// NewAddOutputCmd creates the design add-output command.
func NewAddOutputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-output <path>",
		Short: "Add an output to the design index",
		Long: `Add a planned output document to the current design's index.

Each output is tracked with:
- Path relative to .sow/design/
- Description of the document
- Target location where it will be moved when finalized
- Optional document type and tags

Document types:
  adr          - Architecture Decision Record
  architecture - Architecture documentation
  diagram      - Diagrams (Mermaid, PlantUML, etc.)
  spec         - Specifications
  other        - Other design documents

Requirements:
  - Must be in a sow repository with an active design session
  - Output must not already be in the index
  - Target location must not be empty

Examples:
  # Add ADR
  sow design add-output adr-001-oauth-decision.md \
    --description "Decision to use OAuth 2.0" \
    --target .sow/knowledge/adrs/ \
    --type adr \
    --tags "oauth,decision"

  # Add architecture doc
  sow design add-output architecture-overview.md \
    --description "High-level system architecture" \
    --target .sow/knowledge/architecture/ \
    --type architecture

  # Add diagram
  sow design add-output diagrams/auth-flow.mmd \
    --description "Authentication flow diagram" \
    --target docs/diagrams/ \
    --type diagram`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddOutput(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("description", "d", "", "Brief description of the document (required)")
	cmd.Flags().StringP("target", "t", "", "Target location for this document when finalized (required)")
	cmd.Flags().String("type", "", "Document type (optional: adr, architecture, diagram, spec, other)")
	cmd.Flags().StringSlice("tags", nil, "Comma-separated tags for organization")

	_ = cmd.MarkFlagRequired("description")
	_ = cmd.MarkFlagRequired("target")

	return cmd
}

func runAddOutput(cmd *cobra.Command, args []string) error {
	path := args[0]
	description, _ := cmd.Flags().GetString("description")
	target, _ := cmd.Flags().GetString("target")
	docType, _ := cmd.Flags().GetString("type")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	// Trim tags
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	// Validate document type if provided
	if docType != "" {
		validTypes := map[string]bool{
			"adr":          true,
			"architecture": true,
			"diagram":      true,
			"spec":         true,
			"other":        true,
		}
		if !validTypes[docType] {
			return fmt.Errorf("invalid document type: %s (must be adr, architecture, diagram, spec, or other)", docType)
		}
	}

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Add output to index
	if err := design.AddOutput(ctx, path, description, target, docType, tags); err != nil {
		if errors.Is(err, design.ErrNoDesign) {
			return fmt.Errorf("no active design session - run 'sow design <topic>' first")
		}
		if errors.Is(err, design.ErrOutputExists) {
			return fmt.Errorf("output %s already exists in design index", path)
		}
		return fmt.Errorf("failed to add output: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Added output to design index: %s\n", path)
	cmd.Printf("\nOutput Details:\n")
	cmd.Printf("  Path:        %s\n", path)
	cmd.Printf("  Description: %s\n", description)
	cmd.Printf("  Target:      %s\n", target)
	if docType != "" {
		cmd.Printf("  Type:        %s\n", docType)
	}
	if len(tags) > 0 {
		cmd.Printf("  Tags:        %s\n", strings.Join(tags, ", "))
	}

	return nil
}
