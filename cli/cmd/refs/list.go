package refs

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// newListCmd creates the unified list command.
func newListCmd() *cobra.Command {
	var (
		typeFilter     string
		semanticFilter string
		tagsFilter     []string
		showLocal      bool
		showCommitted  bool
		format         string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured references",
		Long: `List all configured references across all types.

Supports filtering by:
  --type      Structural type (git, file)
  --semantic  Semantic type (knowledge, code)
  --tags      Topic tags (comma-separated)
  --local     Show only local refs
  --committed Show only committed refs (default: show both)

Output formats:
  table  Table format (default)
  json   JSON output
  yaml   YAML output`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRefsList(cmd, typeFilter, semanticFilter, tagsFilter, showLocal, showCommitted, format)
		},
	}

	cmd.Flags().StringVar(&typeFilter, "type", "", "Filter by structural type (git, file)")
	cmd.Flags().StringVar(&semanticFilter, "semantic", "", "Filter by semantic type (knowledge, code)")
	cmd.Flags().StringSliceVar(&tagsFilter, "tags", []string{}, "Filter by tags")
	cmd.Flags().BoolVar(&showLocal, "local", false, "Show only local refs")
	cmd.Flags().BoolVar(&showCommitted, "committed", false, "Show only committed refs")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, json, yaml")

	return cmd
}

func runRefsList(
	cmd *cobra.Command,
	typeFilter string,
	semanticFilter string,
	tagsFilter []string,
	showLocal bool,
	showCommitted bool,
	format string,
) error {
	ctx := cmd.Context()

	// Get Sow from context
	s := sowFromContext(ctx)

	// Build filter options
	var opts []sow.RefListOption

	if typeFilter != "" {
		opts = append(opts, sow.WithRefTypeFilter(typeFilter))
	}
	if semanticFilter != "" {
		opts = append(opts, sow.WithRefSemanticFilter(semanticFilter))
	}
	if len(tagsFilter) > 0 {
		opts = append(opts, sow.WithRefTagsFilter(tagsFilter...))
	}

	// Handle local/committed filters
	if showLocal && !showCommitted {
		opts = append(opts, sow.WithRefLocalOnly())
	} else if showCommitted && !showLocal {
		opts = append(opts, sow.WithRefCommittedOnly())
	}
	// If both true or both false, show both (default behavior)

	// List refs
	refs, err := s.ListRefs(opts...)
	if err != nil {
		return fmt.Errorf("failed to list refs: %w", err)
	}

	if len(refs) == 0 {
		cmd.Println("No refs configured")
		return nil
	}

	// Convert to refWithSource for printing
	var refsList []refWithSource
	for _, ref := range refs {
		schema, err := ref.Schema()
		if err != nil {
			return fmt.Errorf("failed to get ref schema: %w", err)
		}
		isLocal, err := ref.IsLocal()
		if err != nil {
			return fmt.Errorf("failed to check if ref is local: %w", err)
		}
		source := "committed"
		if isLocal {
			source = "local"
		}
		refsList = append(refsList, refWithSource{Ref: *schema, Source: source})
	}

	// Output in requested format
	switch format {
	case "table":
		printRefsTable(cmd, refsList)
	case "json":
		if err := printRefsJSON(cmd, refsList); err != nil {
			return err
		}
	case "yaml":
		if err := printRefsYAML(cmd, refsList); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown format: %s (valid: table, json, yaml)", format)
	}

	return nil
}
