package project

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"fmt"

	"github.com/spf13/cobra"
)

// newFinalizeMoveArtifactCmd creates the command to record moved artifacts.
//
// Usage:
//   sow project finalize move-artifact <from> <to>
//
// Arguments:
//   <from>: Source path (relative to .sow/project/)
//   <to>: Destination path (relative to .sow/)
//
// This command tracks artifacts moved from project directory to knowledge directory.
func newFinalizeMoveArtifactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move-artifact <from> <to>",
		Short: "Record an artifact moved to knowledge",
		Long: `Record an artifact that was moved from project to knowledge directory.

During finalization, some artifacts (like ADRs and design documents) are moved
from .sow/project/ to permanent locations in .sow/knowledge/. This command
tracks those moves for audit purposes.

The destination path must be under knowledge/ directory.

Common artifact moves:
  - ADRs: phases/design/adrs/001.md → knowledge/adrs/001.md
  - Design docs: phases/design/design-docs/auth.md → knowledge/architecture/auth.md

Examples:
  # Move ADR to knowledge
  sow project finalize move-artifact phases/design/adrs/001-use-jwt.md knowledge/adrs/001-use-jwt.md

  # Move design doc to architecture
  sow project finalize move-artifact phases/design/design-docs/auth-system.md knowledge/architecture/auth-system.md`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromPath := args[0]
			toPath := args[1]

			// Get Sow from context
			s := cmdutil.SowFromContext(cmd.Context())

			// Get project
			proj, err := s.GetProject()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Record moved artifact (auto-saves)
			if err := proj.MoveArtifact(fromPath, toPath); err != nil {
				return err
			}

			cmd.Printf("✓ Recorded artifact move\n")
			cmd.Printf("  %s → %s\n", fromPath, toPath)

			return nil
		},
	}

	return cmd
}
