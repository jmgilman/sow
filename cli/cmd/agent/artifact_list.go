package agent

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// NewArtifactListCmd creates the command to list artifacts for the active phase.
//
// Usage:
//
//	sow agent artifact list
//
// Lists all artifacts for the currently active phase.
func NewArtifactListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List artifacts for the active phase",
		Long: `List all artifacts for the currently active phase.

The active phase must support artifacts (discovery and design phases do).

Example:
  sow agent artifact list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent init' first")
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

			// Get artifacts based on phase
			var artifacts []struct {
				Path     string
				Approved bool
			}

			switch activePhase {
			case "discovery":
				for _, a := range state.Phases.Discovery.Artifacts {
					artifacts = append(artifacts, struct {
						Path     string
						Approved bool
					}{Path: a.Path, Approved: a.Approved})
				}
			case "design":
				for _, a := range state.Phases.Design.Artifacts {
					artifacts = append(artifacts, struct {
						Path     string
						Approved bool
					}{Path: a.Path, Approved: a.Approved})
				}
			}

			// Display artifacts
			if len(artifacts) == 0 {
				cmd.Printf("No artifacts in %s phase\n", activePhase)
				return nil
			}

			cmd.Printf("Artifacts in %s phase:\n\n", activePhase)
			for _, a := range artifacts {
				status := "pending"
				if a.Approved {
					status = "approved"
				}
				cmd.Printf("  %s [%s]\n", a.Path, status)
			}

			// Summary
			approved := 0
			for _, a := range artifacts {
				if a.Approved {
					approved++
				}
			}
			cmd.Printf("\nTotal: %d (%d approved, %d pending)\n", len(artifacts), approved, len(artifacts)-approved)

			return nil
		},
	}

	return cmd
}
