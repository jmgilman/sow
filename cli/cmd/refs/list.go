package refs

import (
	"context"
	"fmt"
	"os"

	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/spf13/cobra"
)

// newListCmd creates the unified list command.
func newListCmd(sowFSFromContext func(context.Context) sowfs.SowFS) *cobra.Command {
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
			return runRefsList(
				cmd.Context(),
				cmd,
				typeFilter,
				semanticFilter,
				tagsFilter,
				showLocal,
				showCommitted,
				format,
				sowFSFromContext,
			)
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
	ctx context.Context,
	cmd *cobra.Command,
	typeFilter string,
	semanticFilter string,
	tagsFilter []string,
	showLocal bool,
	showCommitted bool,
	format string,
	sowFSFromContext func(context.Context) sowfs.SowFS,
) error {
	// Default: show both if neither flag specified
	if !showLocal && !showCommitted {
		showLocal = true
		showCommitted = true
	}

	// Get SowFS from context
	sfs := sowFSFromContext(ctx)
	if sfs == nil {
		return fmt.Errorf(".sow directory not found - run 'sow init' first")
	}

	refsFS := sfs.Refs()

	// Collect all refs
	var allRefs []refWithSource

	if showCommitted {
		committedIndex, err := refsFS.CommittedIndex()
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load committed index: %w", err)
		}
		if committedIndex != nil {
			for _, ref := range committedIndex.Refs {
				allRefs = append(allRefs, refWithSource{Ref: ref, Source: "committed"})
			}
		}
	}

	if showLocal {
		localIndex, err := refsFS.LocalIndex()
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load local index: %w", err)
		}
		if localIndex != nil {
			for _, ref := range localIndex.Refs {
				allRefs = append(allRefs, refWithSource{Ref: ref, Source: "local"})
			}
		}
	}

	// Apply filters
	filtered := filterRefs(ctx, allRefs, typeFilter, semanticFilter, tagsFilter)

	if len(filtered) == 0 {
		cmd.Println("No refs configured")
		return nil
	}

	// Output in requested format
	switch format {
	case "table":
		printRefsTable(cmd, filtered)
	case "json":
		if err := printRefsJSON(cmd, filtered); err != nil {
			return err
		}
	case "yaml":
		if err := printRefsYAML(cmd, filtered); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown format: %s (valid: table, json, yaml)", format)
	}

	return nil
}
