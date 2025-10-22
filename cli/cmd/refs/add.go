package refs

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/refs"
	"strings"

	"github.com/spf13/cobra"
)

// newAddCmd creates the unified add command.
func newAddCmd() *cobra.Command {
	var (
		id          string
		semantic    string
		link        string
		tags        []string
		description string
		branch      string
		path        string
		local       bool
	)

	cmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add a new reference",
		Long: `Add a new reference to external knowledge or code.

The reference type is automatically inferred from the URL scheme:
  git+https://github.com/org/repo
  git+ssh://git@github.com/org/repo
  git@github.com:org/repo (SSH shorthand, auto-converted)
  file:///absolute/path

Type-specific flags:
  --branch, --path  Only valid for git URLs

Examples:
  # Add git ref with subpath
  sow refs add git+https://github.com/acme/style-guides \
    --link python-style \
    --semantic knowledge \
    --tags formatting,naming \
    --description "Python coding standards" \
    --path python/ \
    --branch main

  # Add local file ref
  sow refs add file:///Users/josh/docs \
    --link local-docs \
    --semantic knowledge \
    --description "Local documentation"`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRefsAdd(c, args, id, semantic, link, tags, description, branch, path, local)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Ref ID (auto-generated from URL if not specified)")
	cmd.Flags().StringVar(&semantic, "semantic", "knowledge", "Semantic type: knowledge or code")
	cmd.Flags().StringVar(&link, "link", "", "Workspace symlink name (required)")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Topic tags for categorization")
	cmd.Flags().StringVar(&description, "description", "", "Description of this ref")
	cmd.Flags().StringVar(&branch, "branch", "", "Git branch (only for git URLs)")
	cmd.Flags().StringVar(&path, "path", "", "Subpath within repository (only for git URLs)")
	cmd.Flags().BoolVar(&local, "local", false, "Add to local index only (not shared with team)")

	_ = cmd.MarkFlagRequired("link")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runRefsAdd(
	c *cobra.Command,
	args []string,
	id string,
	semantic string,
	link string,
	tags []string,
	description string,
	branch string,
	path string,
	local bool,
) error {
	rawURL := args[0]
	ctx := c.Context()

	// Get context
	sowCtx := cmdutil.GetContext(ctx)

	// Create refs manager
	mgr := refs.NewManager(sowCtx)

	// Build options
	opts := []refs.RefOption{
		refs.WithRefLink(link),
		refs.WithRefSemantic(semantic),
		refs.WithRefDescription(description),
		refs.WithRefLocal(local),
	}

	if id != "" {
		opts = append(opts, refs.WithRefID(id))
	}

	if len(tags) > 0 {
		opts = append(opts, refs.WithRefTags(tags...))
	}

	if branch != "" {
		opts = append(opts, refs.WithRefBranch(branch))
	}

	if path != "" {
		opts = append(opts, refs.WithRefPath(path))
	}

	// Add ref (handles all validation, type inference, caching, symlinking)
	ref, err := mgr.Add(ctx, rawURL, opts...)
	if err != nil {
		return err
	}

	// Print confirmation
	return printAddConfirmation(c, ref)
}

func printAddConfirmation(c *cobra.Command, ref *refs.Ref) error {
	refID := ref.ID()

	// Get ref details
	source, _ := ref.Source()
	isLocal, _ := ref.IsLocal()
	semanticType, _ := ref.Semantic()
	workspacePath, _ := ref.WorkspacePath()
	config, _ := ref.Config()
	typeName, _ := ref.Type()

	indexType := "committed"
	if isLocal {
		indexType = "local"
	}

	c.Printf("✓ Added %s ref: %s (%s index)\n", typeName, refID, indexType)

	// Strip scheme prefix for display
	displayURL := source
	displayURL = strings.TrimPrefix(displayURL, "git+")
	displayURL = strings.TrimPrefix(displayURL, "file://")

	c.Printf("  Source: %s\n", displayURL)
	if config.Branch != "" {
		c.Printf("  Branch: %s\n", config.Branch)
	}
	if config.Path != "" {
		c.Printf("  Path: %s\n", config.Path)
	}
	c.Printf("  Semantic: %s\n", semanticType)
	c.Printf("  Workspace: %s\n", workspacePath)

	return nil
}
