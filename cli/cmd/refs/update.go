package refs

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/refs"
	"fmt"

	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
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
			return runRefsUpdate(cmd, refID)
		},
	}

	return cmd
}

func runRefsUpdate(cmd *cobra.Command, refID string) error {
	ctx := cmd.Context()

	// Get context
	sowCtx := cmdutil.GetContext(ctx)

	// Create refs manager
	mgr := refs.NewManager(sowCtx)

	// Update specific ref or all refs
	if refID != "" {
		// Get specific ref
		ref, err := mgr.Get(refID)
		if err != nil {
			return fmt.Errorf("ref not found: %w", err)
		}

		// Update the ref
		if err := ref.Update(ctx); err != nil {
			return fmt.Errorf("failed to update ref: %w", err)
		}

		cmd.Printf("✓ Updated %s\n", refID)
		return nil
	}

	// Update all refs
	refsList, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list refs: %w", err)
	}

	if len(refsList) == 0 {
		cmd.Println("No refs to update")
		return nil
	}

	updated := 0
	skipped := 0

	for _, ref := range refsList {
		id := ref.ID()

		// Attempt update
		if err := ref.Update(ctx); err != nil {
			cmd.Printf("⚠ Skipped %s: %v\n", id, err)
			skipped++
			continue
		}

		cmd.Printf("✓ Updated %s\n", id)
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
