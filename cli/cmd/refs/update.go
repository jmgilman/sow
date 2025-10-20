package refs

import (
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

	// Get Sow from context
	s := sowFromContext(ctx)

	// Update specific ref or all refs
	if refID != "" {
		// Get specific ref
		ref, err := s.GetRef(refID)
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
	refs, err := s.ListRefs()
	if err != nil {
		return fmt.Errorf("failed to list refs: %w", err)
	}

	if len(refs) == 0 {
		cmd.Println("No refs to update")
		return nil
	}

	updated := 0
	skipped := 0

	for _, ref := range refs {
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
