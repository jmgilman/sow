package refs

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	var (
		force      bool
		pruneCache bool
	)

	cmd := &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove a reference",
		Long: `Remove a reference from the repository.

Removes the workspace symlink and the index entry. Optionally prunes
the cache if no other repositories use it (--prune-cache flag).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRefsRemove(cmd, args[0], force, pruneCache)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&pruneCache, "prune-cache", false, "Remove from cache if no other repos use it")

	return cmd
}

func runRefsRemove(cmd *cobra.Command, refID string, force bool, pruneCache bool) error {
	ctx := cmd.Context()

	// Get Sow from context
	s := cmdutil.SowFromContext(ctx)

	// Get the ref
	ref, err := s.GetRef(refID)
	if err != nil {
		return fmt.Errorf("ref not found: %w", err)
	}

	// Confirm unless forced
	if !force {
		source, _ := ref.Source()
		link, _ := ref.Link()
		isLocal, _ := ref.IsLocal()

		indexType := "committed"
		if isLocal {
			indexType = "local"
		}

		// Strip scheme for display
		displayURL := source
		displayURL = strings.TrimPrefix(displayURL, "git+")
		displayURL = strings.TrimPrefix(displayURL, "file://")

		cmd.Printf("Remove ref '%s' from %s index?\n", refID, indexType)
		cmd.Printf("  Source: %s\n", displayURL)
		cmd.Printf("  Link: %s\n", link)

		if pruneCache {
			cmd.Println("  This will also prune the cache if no other repos use it.")
		} else {
			cmd.Println("  The cache will be preserved.")
		}

		cmd.Print("\nContinue? [y/N]: ")
		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			cmd.Println("Cancelled")
			return nil
		}
	}

	// Remove the ref
	if err := s.RemoveRef(ctx, refID, pruneCache); err != nil {
		return fmt.Errorf("failed to remove ref: %w", err)
	}

	isLocal, _ := ref.IsLocal()
	indexType := "committed"
	if isLocal {
		indexType = "local"
	}
	cmd.Printf("✓ Removed %s from %s index\n", refID, indexType)

	if pruneCache {
		cmd.Println("Note: Cache pruning will be implemented with cache commands")
	}

	return nil
}
