package agent

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// NewFinalizeDocCmd creates the command to track documentation updates.
//
// Usage:
//   sow agent finalize doc <path>
//
// Arguments:
//   <path>: Path to documentation file (relative to repo root)
//
// This command tracks which documentation files were updated during finalization.
func NewFinalizeDocCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc <path>",
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
  sow agent finalize doc README.md

  # Track API docs update
  sow agent finalize doc docs/api/authentication.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			docPath := args[0]

			// Get Sow from context
			ctx := cmdutil.GetContext(cmd.Context())

			// Get project
			proj, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent init' first")
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
