package cmd

import (
	"github.com/spf13/cobra"
)

// NewRefsCmd creates the refs command
func NewRefsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refs",
		Short: "Manage external knowledge and code references",
		Long: `Manage external references to knowledge and code repositories.

Refs provide access to external documentation, style guides, and code
examples. They are cached locally and symlinked into .sow/refs/ for
easy access by AI agents.

Two types of refs:
  - Committed refs (team-shared, in .sow/refs/index.json)
  - Local refs (per-developer, in .sow/refs/index.local.json)`,
	}

	// Subcommands
	cmd.AddCommand(newRefsAddCmd())
	cmd.AddCommand(newRefsInitCmd())
	cmd.AddCommand(newRefsStatusCmd())
	cmd.AddCommand(newRefsUpdateCmd())
	cmd.AddCommand(newRefsListCmd())
	cmd.AddCommand(newRefsRemoveCmd())

	return cmd
}

func newRefsAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <source>",
		Short: "Add a new reference",
		Long: `Add a new external reference.

Source formats:
  - Git repository: https://github.com/org/repo
  - Git with branch: https://github.com/org/repo#branch-name
  - Local directory: file:///absolute/path

The command will clone the repository (or link the directory),
interrogate its contents, and add an entry to the appropriate index.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := FilesystemFromContext(cmd.Context())
			source := args[0]
			_, _ = fs, source // TODO: implement

			cmd.Printf("Added ref: %s\n", source)
			return nil
		},
	}

	cmd.Flags().StringP("type", "t", "", "Ref type (knowledge|code)")
	cmd.Flags().StringP("link", "l", "", "Symlink name in .sow/refs/")
	cmd.Flags().StringP("path", "p", "", "Subpath within repository")
	cmd.Flags().StringSliceP("tags", "", []string{}, "Topic tags for categorization")
	cmd.Flags().Bool("local", false, "Add as local-only ref (not committed)")

	return cmd
}

func newRefsInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize refs structure",
		Long: `Initialize the refs directory structure and index files.

Creates:
  .sow/refs/              - Symlink directory
  .sow/refs/index.json    - Committed refs index (empty)
  ~/.cache/sow/           - Cache directory
  ~/.cache/sow/index.json - Cache index (empty)

Also creates .gitignore entries for cache and local refs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := FilesystemFromContext(cmd.Context())
			_ = fs // TODO: implement

			cmd.Println("Refs structure initialized")
			return nil
		},
	}
}

func newRefsStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [ref-id]",
		Short: "Show refs status",
		Long: `Display status of refs and their caches.

Shows:
  - Ref metadata (source, branch, type)
  - Cache status (current, behind, ahead, diverged)
  - Commits behind remote (if behind)
  - Last update timestamp

If ref-id is specified, shows detailed status for that ref.
Otherwise, shows summary for all refs.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := FilesystemFromContext(cmd.Context())
			_ = fs // TODO: implement

			cmd.Println("Refs status:")
			cmd.Println("  All refs up to date")
			return nil
		},
	}

	cmd.Flags().Bool("cache", false, "Show cache details")

	return cmd
}

func newRefsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [ref-id]",
		Short: "Update refs from remote",
		Long: `Update refs by fetching latest changes from remote repositories.

If ref-id is specified, updates only that ref.
Otherwise, updates all refs that are behind their remotes.

Does not update local-only refs (file:// sources).`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := FilesystemFromContext(cmd.Context())
			_ = fs // TODO: implement

			cmd.Println("Refs updated")
			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Update all refs (even if current)")
	cmd.Flags().Bool("prune", false, "Prune unused cache entries after update")

	return cmd
}

func newRefsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all refs",
		Long: `List all configured refs with their metadata.

Output includes:
  - Ref ID
  - Type (knowledge/code)
  - Source URL or path
  - Symlink name
  - Tags
  - Description`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := FilesystemFromContext(cmd.Context())
			_ = fs // TODO: implement

			cmd.Println("No refs configured")
			return nil
		},
	}

	cmd.Flags().StringP("type", "t", "", "Filter by type (knowledge|code)")
	cmd.Flags().StringSliceP("tags", "", []string{}, "Filter by tags")
	cmd.Flags().StringP("format", "f", "table", "Output format (table|json|yaml)")

	return cmd
}

func newRefsRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <ref-id>",
		Short: "Remove a reference",
		Long: `Remove a reference from the index and clean up symlinks.

This removes the ref entry from the index and deletes the symlink
from .sow/refs/. The cached repository is preserved unless --prune-cache
is specified.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := FilesystemFromContext(cmd.Context())
			refID := args[0]
			_, _ = fs, refID // TODO: implement

			cmd.Printf("Removed ref: %s\n", refID)
			return nil
		},
	}

	cmd.Flags().Bool("prune-cache", false, "Also remove cached repository if unused")

	return cmd
}
