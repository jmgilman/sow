package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
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
func newFinalizeAddDocumentCmd() *cobra.Command {
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

			// Get Sow from context
			ctx := cmdutil.GetContext(cmd.Context())

			// Get project
			proj, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Add documentation update (auto-saves)
			if err := proj.AddDocumentation(docPath); err != nil {
				return err
			}

			cmd.Printf("✓ Tracked documentation update\n")
			cmd.Printf("  %s\n", docPath)

			return nil
		},
	}

	return cmd
}
