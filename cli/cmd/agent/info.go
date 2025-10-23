package agent

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent project init' first")
			}

			// Get state
			state := project.State()

			// Determine which phase to show
			var targetPhase string
			var phaseStatus string

			if phaseName == "" {
				// Use active phase
				targetPhase, phaseStatus = projectpkg.DetermineActivePhase(state)
				if targetPhase == "unknown" {
					return fmt.Errorf("no active phase found - use --phase to specify a phase")
				}
			} else {
				// Use specified phase
				targetPhase = phaseName

				// Get status for specified phase
				switch phaseName {
				case "discovery":
					phaseStatus = state.Phases.Discovery.Status
				case "design":
					phaseStatus = state.Phases.Design.Status
				case "implementation":
					phaseStatus = state.Phases.Implementation.Status
				case "review":
					phaseStatus = state.Phases.Review.Status
				case "finalize":
					phaseStatus = state.Phases.Finalize.Status
				default:
					return fmt.Errorf("unknown phase: %s", phaseName)
				}
			}

			// Get phase metadata
			metadata, err := projectpkg.GetPhaseMetadata(targetPhase)
			if err != nil {
				return fmt.Errorf("failed to get phase metadata: %w", err)
			}

			// Build output
			var output strings.Builder
			output.WriteString(fmt.Sprintf("\n━━━ Phase: %s ━━━\n", targetPhase))
			output.WriteString(fmt.Sprintf("Status: %s\n\n", phaseStatus))

			// Supported operations
			output.WriteString("Supported Operations:\n")
			if metadata.SupportsTasks {
				output.WriteString("  ✓ Tasks\n")
			} else {
				output.WriteString("  ✗ Tasks\n")
			}
			if metadata.SupportsArtifacts {
				output.WriteString("  ✓ Artifacts\n")
			} else {
				output.WriteString("  ✗ Artifacts\n")
			}

			// Custom fields
			if len(metadata.CustomFields) > 0 {
				output.WriteString("\nCustom Fields:\n")
				for _, field := range metadata.CustomFields {
					desc := field.Description
					if desc == "" {
						desc = "No description"
					}
					output.WriteString(fmt.Sprintf("  • %s (%s): %s\n", field.Name, field.Type, desc))
				}
			}

			// Current counts
			output.WriteString("\nCurrent State:\n")
			if metadata.SupportsTasks {
				tasks := project.ListTasks()
				output.WriteString(fmt.Sprintf("  Tasks: %d\n", len(tasks)))
			}
			if metadata.SupportsArtifacts {
				var artifactCount int
				switch targetPhase {
				case "discovery":
					artifactCount = len(state.Phases.Discovery.Artifacts)
				case "design":
					artifactCount = len(state.Phases.Design.Artifacts)
				}
				output.WriteString(fmt.Sprintf("  Artifacts: %d\n", artifactCount))
			}

			cmd.Print(output.String())
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&phaseName, "phase", "", "Specify which phase to show info for (defaults to active phase)")

	return cmd
}
