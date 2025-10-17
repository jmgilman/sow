package refs

import (
	"context"
	"fmt"
	"os"

	"github.com/jmgilman/sow/cli/internal/refs"
	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

func newStatusCmd(sowFSFromContext func(context.Context) sowfs.SowFS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [id]",
		Short: "Check reference staleness",
		Long: `Check if references are up to date with their sources.

If ID specified, checks that specific ref.
If no ID specified, checks all refs that support staleness checking (e.g., git refs).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			refID := ""
			if len(args) > 0 {
				refID = args[0]
			}
			return runRefsStatus(cmd.Context(), cmd, refID, sowFSFromContext)
		},
	}

	return cmd
}

//nolint:funlen // Command handlers have inherent complexity
func runRefsStatus(ctx context.Context, cmd *cobra.Command, refID string, sowFSFromContext func(context.Context) sowfs.SowFS) error {
	// Get SowFS from context
	sfs := sowFSFromContext(ctx)
	if sfs == nil {
		return fmt.Errorf(".sow directory not found - run 'sow init' first")
	}

	refsFS := sfs.Refs()

	// Load all refs from both indexes
	var refsToCheck []schemas.Ref

	// Load committed refs
	committedIndex, err := refsFS.CommittedIndex()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load committed index: %w", err)
	}
	if committedIndex != nil {
		refsToCheck = append(refsToCheck, committedIndex.Refs...)
	}

	// Load local refs
	localIndex, err := refsFS.LocalIndex()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load local index: %w", err)
	}
	if localIndex != nil {
		refsToCheck = append(refsToCheck, localIndex.Refs...)
	}

	// Filter to specific ref if ID provided
	if refID != "" {
		found := false
		for _, ref := range refsToCheck {
			if ref.Id == refID {
				refsToCheck = []schemas.Ref{ref}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ref with ID %q not found", refID)
		}
	}

	if len(refsToCheck) == 0 {
		cmd.Println("No refs to check")
		return nil
	}

	// Get cache dir for checking staleness
	cacheDir, err := refs.DefaultCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache dir: %w", err)
	}

	// Check each ref
	current := 0
	stale := 0
	skipped := 0

	for _, ref := range refsToCheck {
		// Infer type
		typeName, err := refs.InferTypeFromURL(ref.Source)
		if err != nil {
			cmd.Printf("⚠ Skipped %s: failed to infer type\n", ref.Id)
			skipped++
			continue
		}

		// Get type implementation
		refType, err := refs.GetType(typeName)
		if err != nil {
			cmd.Printf("⚠ Skipped %s: unknown type %s\n", ref.Id, typeName)
			skipped++
			continue
		}

		// Check if type is enabled
		enabled, err := refType.IsEnabled(ctx)
		if err != nil || !enabled {
			cmd.Printf("⚠ Skipped %s: type %s not enabled\n", ref.Id, typeName)
			skipped++
			continue
		}

		// Check staleness
		// Note: We need a cached ref structure. For now, we'll pass nil and
		// let the type implementation handle it
		isStale, err := refType.IsStale(ctx, cacheDir, &ref, nil)
		if err != nil {
			cmd.Printf("✗ Error checking %s: %v\n", ref.Id, err)
			skipped++
			continue
		}

		if isStale {
			cmd.Printf("⚠ %s is STALE (updates available)\n", ref.Id)
			stale++
		} else {
			cmd.Printf("✓ %s is current\n", ref.Id)
			current++
		}
	}

	// Summary
	cmd.Printf("\nStatus: %d current", current)
	if stale > 0 {
		cmd.Printf(", %d stale", stale)
	}
	if skipped > 0 {
		cmd.Printf(", %d skipped", skipped)
	}
	cmd.Println()

	if stale > 0 {
		cmd.Println("\nRun 'sow refs update' to update stale refs")
	}

	return nil
}
