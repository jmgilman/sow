package agent

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// NewArtifactAddCmd creates the command to add an artifact to the active phase.
//
// Usage:
//
//	sow agent artifact add <path>
//
// Adds an artifact to the currently active phase. The phase must support artifacts.
func NewArtifactAddCmd() *cobra.Command {
	var approved bool

	cmd := &cobra.Command{
		Use:   "add <path>",
		Short: "Add an artifact to the active phase",
		Long: `Add an artifact to the currently active phase.

The active phase must support artifacts (discovery and design phases do).
The path should be relative to the .sow/ directory.

Artifacts track important files created during a phase, such as:
  - Research documents during discovery
  - Architecture diagrams during design
  - Design notes and proposals

Use --approved to mark the artifact as approved immediately.

Example:
  # Add artifact (needs approval later)
  sow agent artifact add project/phases/discovery/research/001-jwt.md

  # Add artifact and mark as approved
  sow agent artifact add project/phases/design/diagram.png --approved`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent project init' first")
			}

			// Determine active phase
			state := project.State()
			activePhase, phaseStatus := projectpkg.DetermineActivePhase(state)

			if activePhase == "unknown" {
				return fmt.Errorf("no active phase found")
			}

			// Get phase metadata to validate artifacts are supported
			metadata, err := projectpkg.GetPhaseMetadata(activePhase)
			if err != nil {
				return fmt.Errorf("failed to get phase metadata: %w", err)
			}

			if !metadata.SupportsArtifacts {
				return fmt.Errorf("phase %s does not support artifacts", activePhase)
			}

			// Validate we're in an active state (not a decision state)
			if phaseStatus == "pending" {
				return fmt.Errorf("phase %s is in decision state - enable it first", activePhase)
			}

			// Add artifact
			if err := project.AddArtifact(activePhase, path, approved); err != nil {
				return fmt.Errorf("failed to add artifact: %w", err)
			}

			if approved {
				cmd.Printf("\n✓ Added artifact to %s phase (approved)\n", activePhase)
			} else {
				cmd.Printf("\n✓ Added artifact to %s phase\n", activePhase)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&approved, "approved", false, "Mark artifact as approved immediately")

	return cmd
}
