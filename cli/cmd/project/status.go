package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
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
	proj, err := cmdutil.LoadProject(cmd.Context(), ctx)
	if err != nil {
		return fmt.Errorf("no active project")
	}

	// Output to stdout
	out := cmd.OutOrStdout()

	// Project header
	_, _ = fmt.Fprintf(out, "Project: %s\n", proj.Name)
	_, _ = fmt.Fprintf(out, "Branch: %s\n", proj.Branch)
	_, _ = fmt.Fprintf(out, "Type: %s\n", proj.Type)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "State: %s\n", proj.Statechart.Current_state)
	_, _ = fmt.Fprintln(out)

	// Phases section
	_, _ = fmt.Fprintln(out, "Phases:")
	phaseOrder := []string{"implementation", "review", "finalize"}
	for _, phaseName := range phaseOrder {
		phase, exists := proj.Phases[phaseName]
		if !exists {
			continue
		}

		// Count completed tasks
		completed, total := countTasksByStatus(phase)

		_, _ = fmt.Fprint(out, formatPhaseLine(phaseName, phase.Status, completed, total))
	}
	_, _ = fmt.Fprintln(out)

	// Tasks section for current phase
	currentPhase := inferCurrentPhase(proj.Statechart.Current_state)
	if phase, exists := proj.Phases[currentPhase]; exists && len(phase.Tasks) > 0 {
		_, _ = fmt.Fprintf(out, "Tasks (%s):\n", currentPhase)
		for _, task := range phase.Tasks {
			_, _ = fmt.Fprint(out, formatTaskLine(task.Status, task.Id, task.Name))
		}
	}

	return nil
}

// countTasksByStatus counts completed and total tasks in a phase.
func countTasksByStatus(phase projschema.PhaseState) (completed, total int) {
	for _, task := range phase.Tasks {
		total++
		if task.Status == "completed" {
			completed++
		}
	}
	return
}

// formatPhaseLine formats a phase line with consistent alignment.
// Format: "  {phase_name}  [{status}]  {X}/{Y} tasks completed\n"
// Phase names are padded to 15 chars (longest is "implementation" at 14).
// Status is padded to 11 chars (longest is "in_progress" at 11).
func formatPhaseLine(phaseName, status string, completed, total int) string {
	return fmt.Sprintf("  %-15s  [%-11s]  %d/%d tasks completed\n",
		phaseName, status, completed, total)
}

// formatTaskLine formats a task line with consistent alignment.
// Format: "  [{status}]  {id}  {name}\n"
// Status is padded to 12 chars (longest is "needs_review" at 12).
// ID is always 3 digits.
func formatTaskLine(status, id, name string) string {
	return fmt.Sprintf("  [%-12s]  %s  %s\n", status, id, name)
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
