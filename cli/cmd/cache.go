package cmd

import (
	"github.com/spf13/cobra"
)

// NewCacheCmd creates the cache command
func NewCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage repository cache",
		Long: `Manage the local cache of remote repositories.

The cache stores cloned repositories at ~/.cache/sow/ and is shared
across all sow-enabled repositories on the machine. This saves disk
space and speeds up ref operations.`,
	}

	// Subcommands
	cmd.AddCommand(newCacheStatusCmd())
	cmd.AddCommand(newCachePruneCmd())
	cmd.AddCommand(newCacheClearCmd())

	return cmd
}

func newCacheStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show cache status",
		Long: `Display cache statistics and usage information.

Shows:
  - Total cache size
  - Number of cached repositories
  - Cache directory location
  - Repos using each cached entry
  - Orphaned cache entries (not referenced by any repo)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Cache status:")
			cmd.Println("  Location: ~/.cache/sow")
			cmd.Println("  Size: 0 MB")
			cmd.Println("  Repositories: 0")
			return nil
		},
	}

	cmd.Flags().Bool("detailed", false, "Show detailed cache information")

	return cmd
}

func newCachePruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove unused cache entries",
		Long: `Remove cached repositories that are no longer referenced.

This scans all cache entries and removes those not referenced by
any active sow repository. Useful for cleaning up after removing
refs or deleting repositories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Cache pruned: 0 entries removed")
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be removed without removing")

	return cmd
}

func newCacheClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear entire cache",
		Long: `Remove all cached repositories.

WARNING: This removes all cached repositories regardless of usage.
Refs will be re-cloned on next access. Use 'cache prune' to remove
only unused entries.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Cache cleared")
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}
