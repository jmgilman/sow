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

func newUpdateCmd(sowFSFromContext func(context.Context) sowfs.SowFS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update reference(s)",
		Long: `Update references by pulling latest changes.

If ID specified, updates that specific ref.
If no ID specified, updates all refs that support updates (e.g., git refs).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			refID := ""
			if len(args) > 0 {
				refID = args[0]
			}
			return runRefsUpdate(cmd.Context(), cmd, refID, sowFSFromContext)
		},
	}

	return cmd
}

func runRefsUpdate(ctx context.Context, cmd *cobra.Command, refID string, sowFSFromContext func(context.Context) sowfs.SowFS) error {
	// Get SowFS from context
	sfs := sowFSFromContext(ctx)
	if sfs == nil {
		return fmt.Errorf(".sow directory not found - run 'sow init' first")
	}

	refsFS := sfs.Refs()

	// Load all refs from both indexes
	var refsToUpdate []schemas.Ref

	// Load committed refs
	committedIndex, err := refsFS.CommittedIndex()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load committed index: %w", err)
	}
	if committedIndex != nil {
		refsToUpdate = append(refsToUpdate, committedIndex.Refs...)
	}

	// Load local refs
	localIndex, err := refsFS.LocalIndex()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load local index: %w", err)
	}
	if localIndex != nil {
		refsToUpdate = append(refsToUpdate, localIndex.Refs...)
	}

	// Filter to specific ref if ID provided
	if refID != "" {
		found := false
		for _, ref := range refsToUpdate {
			if ref.Id == refID {
				refsToUpdate = []schemas.Ref{ref}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ref with ID %q not found", refID)
		}
	}

	if len(refsToUpdate) == 0 {
		cmd.Println("No refs to update")
		return nil
	}

	// Create manager
	manager, err := refs.NewManager(sfs.SowDir())
	if err != nil {
		return fmt.Errorf("failed to create refs manager: %w", err)
	}

	// Update each ref
	updated := 0
	skipped := 0
	for _, ref := range refsToUpdate {
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

		// Update the ref
		if err := manager.Update(ctx, &ref); err != nil {
			cmd.Printf("✗ Failed to update %s: %v\n", ref.Id, err)
			continue
		}

		cmd.Printf("✓ Updated %s\n", ref.Id)
		updated++
	}

	// Summary
	cmd.Printf("\nUpdated %d ref(s)", updated)
	if skipped > 0 {
		cmd.Printf(", skipped %d", skipped)
	}
	cmd.Println()

	return nil
}
