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

func newRemoveCmd(sowFSFromContext func(context.Context) sowfs.SowFS) *cobra.Command {
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
			return runRefsRemove(cmd.Context(), cmd, args[0], force, pruneCache, sowFSFromContext)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&pruneCache, "prune-cache", false, "Remove from cache if no other repos use it")

	return cmd
}

//nolint:funlen // Command handlers have inherent complexity
func runRefsRemove(ctx context.Context, cmd *cobra.Command, refID string, force bool, pruneCache bool, sowFSFromContext func(context.Context) sowfs.SowFS) error {
	// Get SowFS from context
	sfs := sowFSFromContext(ctx)
	if sfs == nil {
		return fmt.Errorf(".sow directory not found - run 'sow init' first")
	}

	refsFS := sfs.Refs()

	// Try to find ref in committed index first
	committedIndex, err := refsFS.CommittedIndex()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load committed index: %w", err)
	}

	var ref *schemas.Ref
	var refIndex int
	var isLocal bool

	if committedIndex != nil {
		for i := range committedIndex.Refs {
			if committedIndex.Refs[i].Id == refID {
				ref = &committedIndex.Refs[i]
				refIndex = i
				isLocal = false
				break
			}
		}
	}

	// If not found in committed, try local
	if ref == nil {
		localIndex, err := refsFS.LocalIndex()
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load local index: %w", err)
		}

		if localIndex != nil {
			for i := range localIndex.Refs {
				if localIndex.Refs[i].Id == refID {
					ref = &localIndex.Refs[i]
					refIndex = i
					isLocal = true
					break
				}
			}
		}
	}

	if ref == nil {
		return fmt.Errorf("ref with ID %q not found", refID)
	}

	// Confirm unless forced
	if !force {
		indexType := "committed"
		if isLocal {
			indexType = "local"
		}

		// Strip scheme for display
		displayURL := ref.Source
		displayURL = strings.TrimPrefix(displayURL, "git+")
		displayURL = strings.TrimPrefix(displayURL, "file://")

		cmd.Printf("Remove ref '%s' from %s index?\n", refID, indexType)
		cmd.Printf("  Source: %s\n", displayURL)
		cmd.Printf("  Link: %s\n", ref.Link)

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

	// Create manager and remove ref
	manager, err := refs.NewManager(sfs.SowDir())
	if err != nil {
		return fmt.Errorf("failed to create refs manager: %w", err)
	}

	if err := manager.Remove(ctx, ref); err != nil {
		return fmt.Errorf("failed to remove ref: %w", err)
	}

	// Remove from index
	if err := removeRefFromIndex(refsFS, refIndex, isLocal); err != nil {
		return err
	}

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

// removeRefFromIndex removes a ref from the appropriate index file.
func removeRefFromIndex(refsFS sowfs.RefsFS, refIndex int, isLocal bool) error {
	if isLocal {
		return removeRefFromLocalIndex(refsFS, refIndex)
	}
	return removeRefFromCommittedIndex(refsFS, refIndex)
}

func removeRefFromLocalIndex(refsFS sowfs.RefsFS, refIndex int) error {
	localIndex, err := refsFS.LocalIndex()
	if err != nil {
		return fmt.Errorf("failed to reload local index: %w", err)
	}
	localIndex.Refs = append(localIndex.Refs[:refIndex], localIndex.Refs[refIndex+1:]...)
	if err := refsFS.WriteLocalIndex(localIndex); err != nil {
		return fmt.Errorf("failed to save local index: %w", err)
	}
	return nil
}

func removeRefFromCommittedIndex(refsFS sowfs.RefsFS, refIndex int) error {
	committedIndex, err := refsFS.CommittedIndex()
	if err != nil {
		return fmt.Errorf("failed to reload committed index: %w", err)
	}
	committedIndex.Refs = append(committedIndex.Refs[:refIndex], committedIndex.Refs[refIndex+1:]...)
	if err := refsFS.WriteCommittedIndex(committedIndex); err != nil {
		return fmt.Errorf("failed to save committed index: %w", err)
	}
	return nil
}
