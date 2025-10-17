package refs

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jmgilman/sow/cli/internal/refs"
	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// newAddCmd creates the unified add command.
func newAddCmd(sowFSFromContext func(context.Context) sowfs.SowFS) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRefsAdd(
				cmd.Context(),
				cmd,
				args[0],
				id,
				semantic,
				link,
				tags,
				description,
				branch,
				path,
				local,
				sowFSFromContext,
			)
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

//nolint:funlen // Command handlers have inherent complexity
func runRefsAdd(
	ctx context.Context,
	cmd *cobra.Command,
	rawURL string,
	id string,
	semantic string,
	link string,
	tags []string,
	description string,
	branch string,
	path string,
	local bool,
	sowFSFromContext func(context.Context) sowfs.SowFS,
) error {
	// Infer type from URL
	typeName, err := refs.InferTypeFromURL(rawURL)
	if err != nil {
		return fmt.Errorf("failed to infer type from URL: %w", err)
	}

	// Get type implementation
	refType, err := refs.GetType(typeName)
	if err != nil {
		return fmt.Errorf("unknown reference type: %s", typeName)
	}

	// Check if type is enabled
	enabled, err := refType.IsEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if type enabled: %w", err)
	}
	if !enabled {
		return fmt.Errorf("type %s is not available on this system", typeName)
	}

	// Normalize URL and validate flags
	normalizedURL, local, err := normalizeURLForType(rawURL, typeName, branch, path, local)
	if err != nil {
		return err
	}

	// Validate semantic type
	if semantic != "knowledge" && semantic != "code" {
		return fmt.Errorf("semantic must be 'knowledge' or 'code', got: %s", semantic)
	}

	// Generate ID if not specified
	if id == "" {
		id = generateIDFromURL(normalizedURL, typeName)
	}

	// Get SowFS from context
	sfs := sowFSFromContext(ctx)
	if sfs == nil {
		return fmt.Errorf(".sow directory not found - run 'sow init' first")
	}

	// Create ref structure
	ref := &schemas.Ref{
		Id:          id,
		Source:      normalizedURL,
		Semantic:    semantic,
		Link:        link,
		Tags:        tags,
		Description: description,
		Config: schemas.RefConfig{
			Branch: branch,
			Path:   path,
		},
	}

	// Validate config for this type
	if err := refType.ValidateConfig(ref.Config); err != nil {
		return fmt.Errorf("invalid config for type %s: %w", typeName, err)
	}

	// Determine which index to use
	refsFS := sfs.Refs()
	var index *schemas.RefsCommittedIndex
	var isLocal bool

	if local {
		// Load or create local index
		localIndex, err := refsFS.LocalIndex()
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load local index: %w", err)
		}
		if localIndex == nil {
			localIndex = &schemas.RefsLocalIndex{
				Version: "1.0.0",
				Refs:    []schemas.Ref{},
			}
		}
		// Convert to committed index structure for unified handling
		index = &schemas.RefsCommittedIndex{
			Version: localIndex.Version,
			Refs:    localIndex.Refs,
		}
		isLocal = true
	} else {
		// Load committed index
		index, err = refsFS.CommittedIndex()
		if err != nil {
			return fmt.Errorf("failed to load committed index: %w", err)
		}
		isLocal = false
	}

	// Check for duplicate ID
	for _, existingRef := range index.Refs {
		if existingRef.Id == ref.Id {
			indexType := "committed"
			if isLocal {
				indexType = "local"
			}
			return fmt.Errorf("ref with ID %q already exists in %s index", ref.Id, indexType)
		}
	}

	// Create manager and install ref
	manager, err := refs.NewManager(sfs.SowDir())
	if err != nil {
		return fmt.Errorf("failed to create refs manager: %w", err)
	}

	workspacePath, err := manager.Install(ctx, ref)
	if err != nil {
		return fmt.Errorf("failed to install ref: %w", err)
	}

	// Add to index and save
	index.Refs = append(index.Refs, *ref)

	if isLocal {
		// Convert back to local index and save
		localIndexToSave := &schemas.RefsLocalIndex{
			Version: "1.0.0", // Ensure version is always set
			Refs:    index.Refs,
		}
		if err := refsFS.WriteLocalIndex(localIndexToSave); err != nil {
			return fmt.Errorf("failed to save local index: %w", err)
		}
	} else {
		// Ensure version is set for committed index too
		index.Version = "1.0.0"
		if err := refsFS.WriteCommittedIndex(index); err != nil {
			return fmt.Errorf("failed to save committed index: %w", err)
		}
	}

	// Print confirmation
	printAddConfirmation(cmd, ref, typeName, workspacePath, isLocal)

	return nil
}

// normalizeURLForType normalizes a URL based on its type and validates type-specific flags.
func normalizeURLForType(rawURL, typeName, branch, path string, local bool) (string, bool, error) {
	normalizedURL := rawURL

	switch typeName {
	case "git":
		normalized, _, err := refs.ParseGitURL(rawURL)
		if err != nil {
			return "", local, fmt.Errorf("failed to parse git URL: %w", err)
		}
		normalizedURL = normalized

	case "file":
		// Convert path to file URL if needed
		if !strings.HasPrefix(rawURL, "file://") {
			fileURL, err := refs.PathToFileURL(rawURL)
			if err != nil {
				return "", local, fmt.Errorf("failed to convert path to file URL: %w", err)
			}
			normalizedURL = fileURL
		}

		// Validate file URL
		if err := refs.ValidateFileURL(normalizedURL); err != nil {
			return "", local, fmt.Errorf("invalid file URL: %w", err)
		}

		// File refs are always local
		local = true

		// File refs don't support branch/path
		if branch != "" {
			return "", local, fmt.Errorf("--branch flag only valid for git URLs")
		}
		if path != "" {
			return "", local, fmt.Errorf("--path flag only valid for git URLs")
		}

	default:
		// For other types, validate they don't use git-specific flags
		if branch != "" {
			return "", local, fmt.Errorf("--branch flag only valid for git URLs")
		}
		if path != "" {
			return "", local, fmt.Errorf("--path flag only valid for git URLs")
		}
	}

	return normalizedURL, local, nil
}

func printAddConfirmation(cmd *cobra.Command, ref *schemas.Ref, typeName, workspacePath string, isLocal bool) {
	indexType := "committed"
	if isLocal {
		indexType = "local"
	}

	cmd.Printf("✓ Added %s ref: %s (%s index)\n", typeName, ref.Id, indexType)

	// Strip scheme prefix for display
	displayURL := ref.Source
	displayURL = strings.TrimPrefix(displayURL, "git+")
	displayURL = strings.TrimPrefix(displayURL, "file://")

	cmd.Printf("  Source: %s\n", displayURL)
	if ref.Config.Branch != "" {
		cmd.Printf("  Branch: %s\n", ref.Config.Branch)
	}
	if ref.Config.Path != "" {
		cmd.Printf("  Path: %s\n", ref.Config.Path)
	}
	cmd.Printf("  Semantic: %s\n", ref.Semantic)
	cmd.Printf("  Workspace: %s\n", workspacePath)
}
