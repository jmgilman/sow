package agent

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewInfoCmd creates the command to show phase information.
//
// Usage:
//
//	sow agent info [--phase <name>]
//
// Shows detailed information about a phase including supported operations and custom fields.
func NewInfoCmd() *cobra.Command {
	var phaseName string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show phase information",
		Long: `Show detailed information about a phase.

Displays:
  - Phase name and status
  - Supported operations (tasks, artifacts)
  - Custom fields available for this phase
  - Current task/artifact counts

If --phase is not specified, shows information for the currently active phase.

Example:
  # Show active phase info
  sow agent info

  # Show specific phase info
  sow agent info --phase discovery`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project
			project, err := loader.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Determine which phase to show
			var phase domain.Phase
			var targetPhase string

			if phaseName == "" {
				// Use current active phase
				phase = project.CurrentPhase()
				if phase == nil {
					return fmt.Errorf("no active phase found - use --phase to specify a phase")
				}
				targetPhase = phase.Name()
			} else {
				// Use specified phase
				phase, err = project.Phase(phaseName)
				if err != nil {
					return fmt.Errorf("unknown phase: %s", phaseName)
				}
				targetPhase = phaseName
			}

			// Build output
			var output strings.Builder
			output.WriteString(fmt.Sprintf("\n━━━ Phase: %s ━━━\n", targetPhase))
			output.WriteString(fmt.Sprintf("Status: %s\n\n", phase.Status()))

			// Supported operations (determined by checking what operations are available)
			output.WriteString("Supported Operations:\n")

			// Check if phase supports tasks by trying to list them
			tasks := phase.ListTasks()
			if tasks != nil && len(tasks) >= 0 {
				output.WriteString(fmt.Sprintf("  ✓ Tasks (%d)\n", len(tasks)))
			} else {
				output.WriteString("  ✗ Tasks\n")
			}

			// Check if phase supports artifacts by trying to list them
			artifacts := phase.ListArtifacts()
			if artifacts != nil && len(artifacts) >= 0 {
				output.WriteString(fmt.Sprintf("  ✓ Artifacts (%d)\n", len(artifacts)))
			} else {
				output.WriteString("  ✗ Artifacts\n")
			}

			cmd.Print(output.String())
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&phaseName, "phase", "", "Specify which phase to show info for (defaults to active phase)")

	return cmd
}
