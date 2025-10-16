package cmd

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/refs"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// NewRefsGitCmd creates the refs git command group.
func NewRefsGitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Manage git repository references",
		Long: `Manage references to git repositories.

Git refs are cloned to ~/.cache/sow/refs/git/ and symlinked into
.sow/refs/ for easy access by AI agents.

Commands:
  add     - Add a new git repository reference
  update  - Update existing git reference
  remove  - Remove a git reference
  list    - List all git references`,
	}

	// Subcommands
	cmd.AddCommand(newRefsGitAddCmd())
	cmd.AddCommand(newRefsGitUpdateCmd())
	cmd.AddCommand(newRefsGitRemoveCmd())
	cmd.AddCommand(newRefsGitListCmd())

	return cmd
}

func newRefsGitAddCmd() *cobra.Command {
	var (
		id       string
		semantic string
		link     string
		branch   string
		path     string
		tags     []string
		desc     string
	)

	cmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add a new git repository reference",
		Long: `Add a new git repository reference.

URL formats:
  git+https://github.com/org/repo
  git+ssh://git@github.com/org/repo
  git@github.com:org/repo (SSH shorthand)
  https://github.com/org/repo (will be converted to git+https://)

The repository will be cloned to the cache and symlinked into the workspace.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGitAdd(cmd, args[0], id, semantic, link, branch, path, tags, desc)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Ref ID (auto-generated from URL if not specified)")
	cmd.Flags().StringVar(&semantic, "semantic", "knowledge", "Semantic type: knowledge or code")
	cmd.Flags().StringVar(&link, "link", "", "Workspace symlink name (defaults to ID)")
	cmd.Flags().StringVar(&branch, "branch", "", "Git branch to checkout")
	cmd.Flags().StringVar(&path, "path", "", "Subpath within repository")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Topic tags for categorization")
	cmd.Flags().StringVar(&desc, "description", "", "Description of this ref")

	_ = cmd.MarkFlagRequired("semantic")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runGitAdd(cmd *cobra.Command, rawURL, id, semantic, link, branch, path string, tags []string, desc string) error {
	ctx := cmd.Context()

	// Normalize git URL
	normalizedURL, _, err := refs.ParseGitURL(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse git URL: %w", err)
	}

	// Generate ID if not specified
	if id == "" {
		id = generateIDFromURL(normalizedURL)
	}

	// Validate semantic type
	if semantic != "knowledge" && semantic != "code" {
		return fmt.Errorf("semantic must be 'knowledge' or 'code', got: %s", semantic)
	}

	// Default link to ID if not specified
	if link == "" {
		link = id
	}

	// Get SowFS from context
	sfs := SowFSFromContext(ctx)
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
		Description: desc,
		Config: schemas.RefConfig{
			Branch: branch,
			Path:   path,
		},
	}

	// Check for duplicate ID before doing expensive install work
	refsFS := sfs.Refs()
	committedIndex, err := refsFS.CommittedIndex()
	if err != nil {
		return fmt.Errorf("failed to load committed index: %w", err)
	}

	for _, existingRef := range committedIndex.Refs {
		if existingRef.Id == ref.Id {
			return fmt.Errorf("ref with ID %q already exists", ref.Id)
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
	committedIndex.Refs = append(committedIndex.Refs, *ref)
	committedIndex.Version = "1.0.0"

	if err := refsFS.WriteCommittedIndex(committedIndex); err != nil {
		return fmt.Errorf("failed to save committed index: %w", err)
	}

	// Print confirmation
	printGitAddConfirmation(cmd, id, normalizedURL, branch, path, semantic, workspacePath)

	return nil
}

func printGitAddConfirmation(cmd *cobra.Command, id, normalizedURL, branch, path, semantic, workspacePath string) {
	cmd.Printf("✓ Added git ref: %s\n", id)
	cmd.Printf("  Source: %s\n", normalizedURL)
	if branch != "" {
		cmd.Printf("  Branch: %s\n", branch)
	}
	if path != "" {
		cmd.Printf("  Path: %s\n", path)
	}
	cmd.Printf("  Semantic: %s\n", semantic)
	cmd.Printf("  Workspace: %s\n", workspacePath)
}

func newRefsGitUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a git repository reference",
		Long: `Update a git repository reference by pulling latest changes.

This will fetch the latest commits from the remote repository and
update the local cache. The workspace symlink will be verified.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			refID := args[0]

			// Get SowFS from context
			sfs := SowFSFromContext(ctx)
			if sfs == nil {
				return fmt.Errorf(".sow directory not found - run 'sow init' first")
			}

			refsFS := sfs.Refs()

			// Load committed index
			committedIndex, err := refsFS.CommittedIndex()
			if err != nil {
				return fmt.Errorf("failed to load committed index: %w", err)
			}

			// Find ref by ID
			var ref *schemas.Ref
			for i := range committedIndex.Refs {
				if committedIndex.Refs[i].Id == refID {
					ref = &committedIndex.Refs[i]
					break
				}
			}

			if ref == nil {
				return fmt.Errorf("ref with ID %q not found", refID)
			}

			// Verify it's a git ref
			typeName, err := refs.InferTypeFromURL(ref.Source)
			if err != nil {
				return fmt.Errorf("failed to infer type from URL: %w", err)
			}
			if typeName != "git" {
				return fmt.Errorf("ref %q is not a git ref (type: %s)", refID, typeName)
			}

			// Create manager and update ref
			manager, err := refs.NewManager(sfs.SowDir())
			if err != nil {
				return fmt.Errorf("failed to create refs manager: %w", err)
			}
			if err := manager.Update(ctx, ref); err != nil {
				return fmt.Errorf("failed to update ref: %w", err)
			}

			cmd.Printf("✓ Updated git ref: %s\n", refID)

			return nil
		},
	}

	return cmd
}

func newRefsGitRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove a git repository reference",
		Long: `Remove a git repository reference.

This will remove the workspace symlink and clean up the cache.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			refID := args[0]

			// Get SowFS from context
			sfs := SowFSFromContext(ctx)
			if sfs == nil {
				return fmt.Errorf(".sow directory not found - run 'sow init' first")
			}

			refsFS := sfs.Refs()

			// Load committed index
			committedIndex, err := refsFS.CommittedIndex()
			if err != nil {
				return fmt.Errorf("failed to load committed index: %w", err)
			}

			// Find ref by ID
			var ref *schemas.Ref
			var refIndex int
			for i := range committedIndex.Refs {
				if committedIndex.Refs[i].Id == refID {
					ref = &committedIndex.Refs[i]
					refIndex = i
					break
				}
			}

			if ref == nil {
				return fmt.Errorf("ref with ID %q not found", refID)
			}

			// Verify it's a git ref
			typeName, err := refs.InferTypeFromURL(ref.Source)
			if err != nil {
				return fmt.Errorf("failed to infer type from URL: %w", err)
			}
			if typeName != "git" {
				return fmt.Errorf("ref %q is not a git ref (type: %s)", refID, typeName)
			}

			// Create manager and remove ref
			manager, err := refs.NewManager(sfs.SowDir())
			if err != nil {
				return fmt.Errorf("failed to create refs manager: %w", err)
			}
			if err := manager.Remove(ctx, ref); err != nil {
				return fmt.Errorf("failed to remove ref: %w", err)
			}

			// Remove from index
			committedIndex.Refs = append(committedIndex.Refs[:refIndex], committedIndex.Refs[refIndex+1:]...)

			// Save updated index
			if err := refsFS.WriteCommittedIndex(committedIndex); err != nil {
				return fmt.Errorf("failed to save committed index: %w", err)
			}

			cmd.Printf("✓ Removed git ref: %s\n", refID)

			return nil
		},
	}

	return cmd
}

func newRefsGitListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all git repository references",
		Long: `List all configured git repository references.

Shows:
  - Ref ID
  - Source URL
  - Branch (if specified)
  - Path (if specified)
  - Semantic type
  - Workspace symlink`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			// Get SowFS from context
			sfs := SowFSFromContext(ctx)
			if sfs == nil {
				return fmt.Errorf(".sow directory not found - run 'sow init' first")
			}

			refsFS := sfs.Refs()

			// Load committed index
			committedIndex, err := refsFS.CommittedIndex()
			if err != nil {
				return fmt.Errorf("failed to load committed index: %w", err)
			}

			// Filter for git refs only
			var gitRefs []schemas.Ref
			for _, ref := range committedIndex.Refs {
				// Check if it's a git URL
				if typeName, err := refs.InferTypeFromURL(ref.Source); err == nil && typeName == "git" {
					gitRefs = append(gitRefs, ref)
				}
			}

			if len(gitRefs) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No git refs configured")
				return nil
			}

			// Display refs in table format
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "%-20s %-10s %-15s %s\n", "ID", "SEMANTIC", "LINK", "SOURCE")
			_, _ = fmt.Fprintf(out, "%-20s %-10s %-15s %s\n", "──", "────────", "────", "──────")
			for _, ref := range gitRefs {
				// Strip git+ prefix for display
				displayURL := strings.TrimPrefix(ref.Source, "git+")
				_, _ = fmt.Fprintf(out, "%-20s %-10s %-15s %s\n", ref.Id, ref.Semantic, ref.Link, displayURL)
				if ref.Config.Branch != "" {
					_, _ = fmt.Fprintf(out, "  └─ branch: %s\n", ref.Config.Branch)
				}
				if ref.Config.Path != "" {
					_, _ = fmt.Fprintf(out, "  └─ path: %s\n", ref.Config.Path)
				}
			}

			return nil
		},
	}

	return cmd
}

// generateIDFromURL creates a ref ID from a git URL.
// Example: git+https://github.com/golang/go → golang-go.
func generateIDFromURL(url string) string {
	// Remove scheme prefix
	url = strings.TrimPrefix(url, "git+https://")
	url = strings.TrimPrefix(url, "git+ssh://")
	url = strings.TrimPrefix(url, "git+http://")

	// Remove git@ prefix if present
	url = strings.TrimPrefix(url, "git@")

	// Remove domain (everything up to and including first /)
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		// Take org/repo or just repo
		url = strings.Join(parts[len(parts)-2:], "-")
	}

	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Convert to kebab-case
	url = strings.ToLower(url)
	url = strings.ReplaceAll(url, "/", "-")
	url = strings.ReplaceAll(url, "_", "-")

	return url
}
