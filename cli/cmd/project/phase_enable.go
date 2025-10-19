package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/statechart"
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
				return fmt.Errorf("phase validation failed: %w", err)
			}

			// Only discovery and design can be manually enabled
			if phase != "discovery" && phase != "design" {
				return fmt.Errorf("only discovery and design phases can be manually enabled")
			}

			// Get SowFS from context
			sowFS := accessor(cmd.Context())
			if sowFS == nil {
				return fmt.Errorf("not in a sow repository - run 'sow init' first")
			}

			// Verify project exists
			_, err := sowFS.Project()
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first: %w", err)
			}

			// === STATECHART INTEGRATION START ===

			// Load machine
			machine, err := statechart.Load()
			if err != nil {
				return fmt.Errorf("failed to load statechart: %w", err)
			}

			state := machine.ProjectState()

			// Validate current state allows enabling phases
			currentState := machine.State()
			if phase == "discovery" && currentState != statechart.DiscoveryDecision {
				return fmt.Errorf("cannot enable discovery in current state: %s (expected DiscoveryDecision)", currentState)
			}
			if phase == "design" && currentState != statechart.DesignDecision {
				return fmt.Errorf("cannot enable design in current state: %s (expected DesignDecision)", currentState)
			}

			// Update state based on phase
			var event statechart.Event
			if phase == "discovery" {
				if discoveryType == "" {
					return fmt.Errorf("discovery type required (use --type flag)")
				}
				// Validate discovery type
				validTypes := map[string]bool{
					"bug": true, "feature": true, "docs": true, "refactor": true, "general": true,
				}
				if !validTypes[discoveryType] {
					return fmt.Errorf("invalid discovery type: %s", discoveryType)
				}

				state.Phases.Discovery.Enabled = true
				state.Phases.Discovery.Status = "pending"
				state.Phases.Discovery.Discovery_type = &discoveryType
				event = statechart.EventEnableDiscovery

			} else { // design
				state.Phases.Design.Enabled = true
				state.Phases.Design.Status = "pending"
				event = statechart.EventEnableDesign
			}

			// Fire event (validates transition, outputs prompt)
			if err := machine.Fire(event); err != nil {
				return fmt.Errorf("failed to enable %s phase: %w", phase, err)
			}

			// Save state
			if err := machine.Save(); err != nil {
				return fmt.Errorf("failed to save state: %w", err)
			}

			// === STATECHART INTEGRATION END ===

			cmd.Printf("\n✓ Enabled %s phase\n", phase)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&discoveryType, "type", "", "Discovery type (required for discovery phase): bug, feature, docs, refactor, general")

	return cmd
}
