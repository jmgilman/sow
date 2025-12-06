package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current project status",
		Long: `Display the current project state in a readable format.

Shows:
  - Project header (name, branch, type, state)
  - Phase list with status and task progress
  - Task list for the current/active phase

Example:
  sow project status`,
		RunE: runStatus,
	}
}

func runStatus(cmd *cobra.Command, _ []string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Load project state
	proj, err := state.Load(ctx)
	if err != nil {
		return fmt.Errorf("no active project")
	}

	// Output to stdout
	out := cmd.OutOrStdout()

	// Project header
	fmt.Fprintf(out, "Project: %s\n", proj.Name)
	fmt.Fprintf(out, "Branch: %s\n", proj.Branch)
	fmt.Fprintf(out, "Type: %s\n", proj.Type)
	fmt.Fprintln(out)
	fmt.Fprintf(out, "State: %s\n", proj.Statechart.Current_state)
	fmt.Fprintln(out)

	// Phases section
	fmt.Fprintln(out, "Phases:")
	phaseOrder := []string{"implementation", "review", "finalize"}
	for _, phaseName := range phaseOrder {
		phase, exists := proj.Phases[phaseName]
		if !exists {
			continue
		}

		// Count completed tasks
		completed := 0
		total := len(phase.Tasks)
		for _, task := range phase.Tasks {
			if task.Status == "completed" {
				completed++
			}
		}

		fmt.Fprintf(out, "  %s  [%s]  %d/%d tasks completed\n", phaseName, phase.Status, completed, total)
	}
	fmt.Fprintln(out)

	// Tasks section for current phase
	currentPhase := inferCurrentPhase(proj.Statechart.Current_state)
	if phase, exists := proj.Phases[currentPhase]; exists && len(phase.Tasks) > 0 {
		fmt.Fprintf(out, "Tasks (%s):\n", currentPhase)
		for _, task := range phase.Tasks {
			fmt.Fprintf(out, "  [%s]  %s  %s\n", task.Status, task.Id, task.Name)
		}
	}

	return nil
}

// inferCurrentPhase extracts the phase name from a statechart state.
// State names typically include the phase (e.g., "ImplementationPlanning", "ReviewActive").
func inferCurrentPhase(stateName string) string {
	// Check for known prefixes
	prefixes := map[string]string{
		"Implementation": "implementation",
		"Review":         "review",
		"Finalize":       "finalize",
	}

	for prefix, phase := range prefixes {
		if len(stateName) >= len(prefix) && stateName[:len(prefix)] == prefix {
			return phase
		}
	}

	// Default to implementation
	return "implementation"
}
