package agent

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
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

The active phase must support artifacts (discovery, design, and review phases do).

Example:
  sow agent artifact list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project via loader to get interface
			proj, err := loader.Load(ctx)
			if err != nil {
				if errors.Is(err, project.ErrNoProject) {
					return fmt.Errorf("no active project - run 'sow agent init' first")
				}
				return fmt.Errorf("failed to load project: %w", err)
			}

			// Get current phase
			phase := proj.CurrentPhase()
			if phase == nil {
				return fmt.Errorf("no active phase found")
			}

			// Get artifacts via Phase interface
			artifacts := phase.ListArtifacts()

			// Display artifacts
			if len(artifacts) == 0 {
				cmd.Printf("No artifacts in %s phase\n", phase.Name())
				return nil
			}

			cmd.Printf("Artifacts in %s phase:\n\n", phase.Name())
			for _, a := range artifacts {
				status := "pending"
				if a.Approved != nil && *a.Approved {
					status = "approved"
				}
				cmd.Printf("  %s [%s]", a.Path, status)

				// Show metadata if present
				if len(a.Metadata) > 0 {
					cmd.Printf(" - metadata: %v", a.Metadata)
				}
				cmd.Printf("\n")
			}

			// Summary
			approved := 0
			for _, a := range artifacts {
				if a.Approved != nil && *a.Approved {
					approved++
				}
			}
			cmd.Printf("\nTotal: %d (%d approved, %d pending)\n", len(artifacts), approved, len(artifacts)-approved)

			return nil
		},
	}

	return cmd
}
