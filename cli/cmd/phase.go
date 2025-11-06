package cmd

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/spf13/cobra"
)

// NewPhaseCmd creates the phase command.
func NewPhaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "phase",
		Short: "Manage project phases",
		Long: `Manage project phases.

Phase commands allow you to inspect and modify phase state, including
direct fields (status, enabled) and metadata fields used by state machine
guards.`,
	}

	cmd.AddCommand(newPhaseSetCmd())

	return cmd
}

// newPhaseSetCmd creates the phase set subcommand.
func newPhaseSetCmd() *cobra.Command {
	var phaseName string

	cmd := &cobra.Command{
		Use:   "set <field-path> <value>",
		Short: "Set phase field value",
		Long: `Set a phase field value using dot notation.

Supports:
  - Direct fields: status, enabled
  - Metadata fields: metadata.* (any custom field)

The --phase flag specifies which phase to modify. If omitted, defaults to
the currently active phase based on the state machine state.

Examples:
  # Set direct field on active phase
  sow phase set status in_progress

  # Set direct field with explicit phase
  sow phase set enabled false --phase planning

  # Set metadata field (used by state machine guards)
  sow phase set metadata.tasks_approved true --phase implementation

  # Set nested metadata
  sow phase set metadata.complexity.level high`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPhaseSet(cmd, args, phaseName)
		},
	}

	cmd.Flags().StringVarP(&phaseName, "phase", "p", "", "Target phase (defaults to active phase)")

	return cmd
}

// runPhaseSet implements the phase set command logic.
func runPhaseSet(cmd *cobra.Command, args []string, phaseName string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project
	project, err := state.Load(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Determine target phase
	if phaseName == "" {
		phaseName = getActivePhase(project)
		if phaseName == "" {
			return fmt.Errorf("could not determine active phase")
		}
	}

	// Get phase from map
	phaseState, exists := project.Phases[phaseName]
	if !exists {
		return fmt.Errorf("phase not found: %s", phaseName)
	}

	// Wrap in Phase type for field path mutation
	phase := &state.Phase{
		PhaseState: phaseState,
	}

	// Set field using field path parser
	fieldPath := args[0]
	value := args[1]

	if err := cmdutil.SetField(phase, fieldPath, value); err != nil {
		return fmt.Errorf("failed to set field: %w", err)
	}

	// Update the phase back in the map
	project.Phases[phaseName] = phase.PhaseState

	// Save project state
	if err := project.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	return nil
}

// getActivePhase determines the active phase by finding the phase with status="in_progress".
// This works for all project types (standard, breakdown, design, exploration).
// Returns the phase name, or empty string if no phase is in progress.
func getActivePhase(project *state.Project) string {
	// Find the phase with status="in_progress"
	// The state machine design ensures only one phase is in_progress at a time
	for phaseName, phase := range project.Phases {
		if phase.Status == "in_progress" {
			return phaseName
		}
	}
	return ""
}
