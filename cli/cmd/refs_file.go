package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/cli/internal/refs"
	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// NewRefsFileCmd creates the refs file command group.
func NewRefsFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Manage local file references",
		Long: `Manage references to local files and directories.

File refs are symlinked from the source location into .sow/refs/
for easy access by AI agents.

Commands:
  add     - Add a new file reference
  remove  - Remove a file reference
  list    - List all file references`,
	}

	// Subcommands
	cmd.AddCommand(newRefsFileAddCmd())
	cmd.AddCommand(newRefsFileRemoveCmd())
	cmd.AddCommand(newRefsFileListCmd())

	return cmd
}

func newRefsFileAddCmd() *cobra.Command {
	var (
		id       string
		semantic string
		link     string
		tags     []string
		desc     string
	)

	cmd := &cobra.Command{
		Use:   "add <path>",
		Short: "Add a new file reference",
		Long: `Add a new file reference to a local directory or file.

The path must be an absolute path to an existing file or directory.
A symlink will be created in the workspace pointing to this location.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileAdd(cmd, args[0], id, semantic, link, tags, desc)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Ref ID (auto-generated from path if not specified)")
	cmd.Flags().StringVar(&semantic, "semantic", "knowledge", "Semantic type: knowledge or code")
	cmd.Flags().StringVar(&link, "link", "", "Workspace symlink name (defaults to ID)")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Topic tags for categorization")
	cmd.Flags().StringVar(&desc, "description", "", "Description of this ref")

	_ = cmd.MarkFlagRequired("semantic")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runFileAdd(cmd *cobra.Command, rawPath, id, semantic, link string, tags []string, desc string) error {
	ctx := cmd.Context()

	// Convert path to absolute path
	absPath, err := filepath.Abs(rawPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Verify path exists
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absPath)
		}
		return fmt.Errorf("failed to access path: %w", err)
	}

	// Convert to file URL
	fileURL, err := refs.PathToFileURL(absPath)
	if err != nil {
		return fmt.Errorf("failed to convert path to URL: %w", err)
	}

	// Generate ID if not specified
	if id == "" {
		id = generateIDFromPath(absPath)
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
		Source:      fileURL,
		Semantic:    semantic,
		Link:        link,
		Tags:        tags,
		Description: desc,
	}

	// Install and save ref
	workspacePath, err := installAndSaveRef(ctx, sfs, ref)
	if err != nil {
		return err
	}

	// Print confirmation
	printFileAddConfirmation(cmd, id, absPath, semantic, workspacePath)

	return nil
}

// installAndSaveRef checks for duplicates, installs the ref, and saves to index.
func installAndSaveRef(ctx context.Context, sfs sowfs.SowFS, ref *schemas.Ref) (string, error) {
	// Check for duplicate ID before doing expensive install work
	refsFS := sfs.Refs()
	committedIndex, err := refsFS.CommittedIndex()
	if err != nil {
		return "", fmt.Errorf("failed to load committed index: %w", err)
	}

	for _, existingRef := range committedIndex.Refs {
		if existingRef.Id == ref.Id {
			return "", fmt.Errorf("ref with ID %q already exists", ref.Id)
		}
	}

	// Create manager and install ref
	manager, err := refs.NewManager(sfs.SowDir())
	if err != nil {
		return "", fmt.Errorf("failed to create refs manager: %w", err)
	}
	workspacePath, err := manager.Install(ctx, ref)
	if err != nil {
		return "", fmt.Errorf("failed to install ref: %w", err)
	}

	// Add to index and save
	committedIndex.Refs = append(committedIndex.Refs, *ref)
	committedIndex.Version = "1.0.0"

	if err := refsFS.WriteCommittedIndex(committedIndex); err != nil {
		return "", fmt.Errorf("failed to save committed index: %w", err)
	}

	return workspacePath, nil
}

func printFileAddConfirmation(cmd *cobra.Command, id, absPath, semantic, workspacePath string) {
	cmd.Printf("✓ Added file ref: %s\n", id)
	cmd.Printf("  Source: %s\n", absPath)
	cmd.Printf("  Semantic: %s\n", semantic)
	cmd.Printf("  Workspace: %s\n", workspacePath)
}

func newRefsFileRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove a file reference",
		Long: `Remove a file reference.

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

			// Verify it's a file ref
			typeName, err := refs.InferTypeFromURL(ref.Source)
			if err != nil {
				return fmt.Errorf("failed to infer type from URL: %w", err)
			}
			if typeName != "file" {
				return fmt.Errorf("ref %q is not a file ref (type: %s)", refID, typeName)
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

			cmd.Printf("✓ Removed file ref: %s\n", refID)

			return nil
		},
	}

	return cmd
}

func newRefsFileListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all file references",
		Long: `List all configured file references.

Shows:
  - Ref ID
  - Source path
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

			// Filter for file refs only
			var fileRefs []schemas.Ref
			for _, ref := range committedIndex.Refs {
				// Check if it's a file URL
				if typeName, err := refs.InferTypeFromURL(ref.Source); err == nil && typeName == "file" {
					fileRefs = append(fileRefs, ref)
				}
			}

			if len(fileRefs) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No file refs configured")
				return nil
			}

			// Display refs in table format
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "%-20s %-10s %-15s %s\n", "ID", "SEMANTIC", "LINK", "SOURCE")
			_, _ = fmt.Fprintf(out, "%-20s %-10s %-15s %s\n", "──", "────────", "────", "──────")
			for _, ref := range fileRefs {
				// Convert file URL back to path for display
				sourcePath, err := refs.FileURLToPath(ref.Source)
				if err != nil {
					sourcePath = ref.Source // Fallback to URL if conversion fails
				}
				_, _ = fmt.Fprintf(out, "%-20s %-10s %-15s %s\n", ref.Id, ref.Semantic, ref.Link, sourcePath)
			}

			return nil
		},
	}

	return cmd
}

// generateIDFromPath creates a ref ID from a file path.
// Example: /home/user/style-guide → style-guide.
func generateIDFromPath(path string) string {
	// Get the base name
	base := filepath.Base(path)

	// Remove common extensions
	base = removeCommonExtensions(base)

	// Convert to lowercase kebab-case
	id := filepath.ToSlash(base)
	id = filepath.Base(id)

	return id
}

// removeCommonExtensions removes common file extensions.
func removeCommonExtensions(name string) string {
	exts := []string{".md", ".txt", ".pdf", ".doc", ".docx"}
	for _, ext := range exts {
		if len(name) > len(ext) && name[len(name)-len(ext):] == ext {
			return name[:len(name)-len(ext)]
		}
	}
	return name
}
