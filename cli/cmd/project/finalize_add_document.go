package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newFinalizeAddDocumentCmd creates the command to track documentation updates.
//
// Usage:
//   sow project finalize add-document <path>
//
// Arguments:
//   <path>: Path to documentation file (relative to repo root)
//
// This command tracks which documentation files were updated during finalization.
func newFinalizeAddDocumentCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-document <path>",
		Short: "Track a documentation file update",
		Long: `Track a documentation file that was updated during finalization.

This command records which documentation files were modified during the finalize
phase. The orchestrator uses this to track what changed before creating the PR.

Common documentation files:
  - README.md
  - CHANGELOG.md
  - docs/api.md
  - docs/architecture.md

Examples:
  # Track README update
  sow project finalize add-document README.md

  # Track API docs update
  sow project finalize add-document docs/api/authentication.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			docPath := args[0]

			// Get SowFS from context
			sowFS := accessor(cmd.Context())
			if sowFS == nil {
				return fmt.Errorf("not in a sow repository - run 'sow init' first")
			}

			// Get project filesystem
			projectFS, err := sowFS.Project()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first: %w", err)
			}

			// Read current state
			state, err := projectFS.State()
			if err != nil {
				return fmt.Errorf("failed to read project state: %w", err)
			}

			// Add documentation update
			if err := project.AddDocumentationUpdate(state, docPath); err != nil {
				return fmt.Errorf("failed to add documentation update: %w", err)
			}

			// Write updated state
			if err := projectFS.WriteState(state); err != nil {
				return fmt.Errorf("failed to write project state: %w", err)
			}

			cmd.Printf("✓ Tracked documentation update\n")
			cmd.Printf("  %s\n", docPath)

			return nil
		},
	}

	return cmd
}
