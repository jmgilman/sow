package project

import (
	"encoding/json"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newArtifactListCmd creates the command to list artifacts.
//
// Usage:
//   sow project artifact list [--phase <phase>] [--format <format>]
//
// Flags:
//   --phase: Phase to filter by (optional, shows all if not specified)
//   --format: Output format (text or json, default: text)
func newArtifactListCmd() *cobra.Command {
	var phaseName string
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List artifacts for a phase",
		Long: `List artifacts tracked in discovery and design phases.

By default, shows artifacts from all phases. Use --phase to filter to a specific phase.

Output format:
  - Default: Human-readable formatted text
  - --format json: Machine-readable JSON

Examples:
  # List all artifacts
  sow project artifact list

  # List artifacts for discovery phase only
  sow project artifact list --phase discovery

  # List in JSON format
  sow project artifact list --format json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Validate format
			if format != "text" && format != "json" {
				return fmt.Errorf("invalid format '%s': must be 'text' or 'json'", format)
			}

			// Validate phase name if provided
			if phaseName != "" {
				// Artifact can only exist in discovery or design
				if phaseName != "discovery" && phaseName != "design" {
					return fmt.Errorf("artifacts can only exist in discovery or design phases, got: %s", phaseName)
				}
			}

			// Get Sow from context
			s := sowFromContext(cmd.Context())

			// Get project
			proj, err := s.GetProject()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Get state
			state := proj.State()

			// Output based on format
			if format == "json" {
				// JSON output: serialize artifacts
				var artifacts interface{}

				switch phaseName {
				case "discovery":
					artifacts = state.Phases.Discovery.Artifacts
				case "design":
					artifacts = state.Phases.Design.Artifacts
				default:
					// All artifacts
					artifacts = map[string]interface{}{
						"discovery": state.Phases.Discovery.Artifacts,
						"design":    state.Phases.Design.Artifacts,
					}
				}

				jsonData, err := json.MarshalIndent(artifacts, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal artifacts to JSON: %w", err)
				}
				cmd.Println(string(jsonData))
			} else {
				// Text output: use formatted display
				output := project.FormatArtifactList(state, phaseName)
				cmd.Print(output)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&phaseName, "phase", "", "Phase to filter by (optional)")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text or json)")

	return cmd
}
