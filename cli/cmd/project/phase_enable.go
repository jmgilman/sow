package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newPhaseEnableCmd creates the command to enable an optional phase.
//
// Usage:
//   sow project phase enable <phase>
//   sow project phase enable discovery --type <type>
//
// Only discovery and design phases can be enabled.
// Implementation, review, and finalize are always enabled.
func newPhaseEnableCmd(accessor SowFSAccessor) *cobra.Command {
	var discoveryType string

	cmd := &cobra.Command{
		Use:   "enable <phase>",
		Short: "Enable an optional phase (discovery or design)",
		Long: `Enable an optional phase.

Only discovery and design phases can be enabled.
Implementation, review, and finalize phases are always enabled.

Discovery phase requires a type:
  - bug: Bug discovery and analysis
  - feature: Feature exploration and requirements
  - docs: Documentation research
  - refactor: Refactoring analysis
  - general: General investigation

Example:
  # Enable discovery phase for a feature
  sow project phase enable discovery --type feature

  # Enable design phase
  sow project phase enable design`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]

			// Validate phase name
			if err := project.ValidatePhase(phase); err != nil {
				return err
			}

			// Get SowFS from context
			sowFS := accessor(cmd.Context())
			if sowFS == nil {
				return fmt.Errorf("not in a sow repository - run 'sow init' first")
			}

			// Get project filesystem
			projectFS, err := sowFS.Project()
			if err != nil {
				return err
			}

			// Read current state
			state, err := projectFS.State()
			if err != nil {
				return fmt.Errorf("failed to read project state: %w", err)
			}

			// Prepare discovery type pointer if needed
			var discoveryTypePtr *string
			if discoveryType != "" {
				discoveryTypePtr = &discoveryType
			}

			// Enable the phase
			if err := project.EnablePhase(state, phase, discoveryTypePtr); err != nil {
				return err
			}

			// Write updated state
			if err := projectFS.WriteState(state); err != nil {
				return fmt.Errorf("failed to write project state: %w", err)
			}

			cmd.Printf("✓ Enabled %s phase for project '%s'\n", phase, state.Project.Name)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&discoveryType, "type", "", "Discovery type (required for discovery phase): bug, feature, docs, refactor, general")

	return cmd
}
