package refs

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
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
			return runRefsStatus(cmd, refID)
		},
	}

	return cmd
}

func runRefsStatus(cmd *cobra.Command, refID string) error {
	ctx := cmd.Context()

	// Get Sow from context
	s := sowFromContext(ctx)

	// Check specific ref or all refs
	if refID != "" {
		// Get specific ref
		ref, err := s.GetRef(refID)
		if err != nil {
			return fmt.Errorf("ref not found: %w", err)
		}

		// Check status
		isStale, err := ref.Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to check status: %w", err)
		}

		if isStale {
			cmd.Printf("⚠ %s is STALE (updates available)\n", refID)
			cmd.Println("\nRun 'sow refs update " + refID + "' to update")
		} else {
			cmd.Printf("✓ %s is current\n", refID)
		}

		return nil
	}

	// Check all refs
	refs, err := s.ListRefs()
	if err != nil {
		return fmt.Errorf("failed to list refs: %w", err)
	}

	if len(refs) == 0 {
		cmd.Println("No refs to check")
		return nil
	}

	current := 0
	stale := 0
	skipped := 0

	for _, ref := range refs {
		id := ref.ID()

		// Check staleness
		isStale, err := ref.Status(ctx)
		if err != nil {
			cmd.Printf("✗ Error checking %s: %v\n", id, err)
			skipped++
			continue
		}

		if isStale {
			cmd.Printf("⚠ %s is STALE (updates available)\n", id)
			stale++
		} else {
			cmd.Printf("✓ %s is current\n", id)
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
