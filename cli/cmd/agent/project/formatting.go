package project

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// Phase name constants used for formatting.
const (
	phaseDiscovery = "discovery"
	phaseDesign    = "design"
)

// formatStatus generates a human-readable status summary for a project.
//
// Output format:
//   Project: {name} (on {branch})
//   Description: {description}
//
//   Phases:
//     [enabled/disabled] Phase    Status
//     [x] Discovery               skipped
//     [x] Design                  skipped
//     [✓] Implementation          in_progress
//     [✓] Review                  pending
//     [✓] Finalize                pending
//
//   Tasks: 3 total (1 completed, 1 in_progress, 1 pending)
//
// Parameters:
//   - state: Project state to format
//
// Returns:
//   - Formatted string ready for display
func formatStatus(state *schemas.ProjectState) string {
	var b strings.Builder

	// Project header
	fmt.Fprintf(&b, "Project: %s (on %s)\n", state.Project.Name, state.Project.Branch)
	fmt.Fprintf(&b, "Description: %s\n", state.Project.Description)

	// GitHub issue link (if present)
	if state.Project.Github_issue != nil && *state.Project.Github_issue > 0 {
		fmt.Fprintf(&b, "GitHub Issue: #%d\n", *state.Project.Github_issue)
	}

	// Pull request URL (if created)
	if state.Phases.Finalize.Pr_url != nil && *state.Phases.Finalize.Pr_url != "" {
		fmt.Fprintf(&b, "Pull Request: %s\n", *state.Phases.Finalize.Pr_url)
	}
	fmt.Fprintln(&b)

	// Phases table
	fmt.Fprintln(&b, "Phases:")
	fmt.Fprintln(&b, "  [enabled] Phase              Status")

	phases := []struct {
		name    string
		enabled bool
		status  string
	}{
		{"Discovery", state.Phases.Discovery.Enabled, state.Phases.Discovery.Status},
		{"Design", state.Phases.Design.Enabled, state.Phases.Design.Status},
		{"Implementation", state.Phases.Implementation.Enabled, state.Phases.Implementation.Status},
		{"Review", state.Phases.Review.Enabled, state.Phases.Review.Status},
		{"Finalize", state.Phases.Finalize.Enabled, state.Phases.Finalize.Status},
	}

	for _, p := range phases {
		enabledMark := " "
		if p.enabled {
			enabledMark = "✓"
		}
		// Pad phase name to 20 characters for alignment
		fmt.Fprintf(&b, "  [%s] %-20s %s\n", enabledMark, p.name, p.status)
	}

	// Task summary
	formatTaskSummary(&b, state.Phases.Implementation.Tasks)

	return b.String()
}

// formatTaskSummary formats the task summary section.
func formatTaskSummary(b *strings.Builder, tasks []phases.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(b, "\nTasks: none yet")
		return
	}

	total := len(tasks)
	completed := 0
	inProgress := 0
	pending := 0
	abandoned := 0

	for _, task := range tasks {
		switch task.Status {
		case "completed":
			completed++
		case "in_progress":
			inProgress++
		case "pending":
			pending++
		case "abandoned":
			abandoned++
		}
	}

	fmt.Fprintf(b, "\nTasks: %d total", total)
	details := []string{}
	if completed > 0 {
		details = append(details, fmt.Sprintf("%d completed", completed))
	}
	if inProgress > 0 {
		details = append(details, fmt.Sprintf("%d in_progress", inProgress))
	}
	if pending > 0 {
		details = append(details, fmt.Sprintf("%d pending", pending))
	}
	if abandoned > 0 {
		details = append(details, fmt.Sprintf("%d abandoned", abandoned))
	}
	if len(details) > 0 {
		fmt.Fprintf(b, " (%s)", strings.Join(details, ", "))
	}
	fmt.Fprintln(b)
}

// formatPhaseStatus generates a human-readable phase status table.
//
// This is similar to formatStatus but focuses only on phases without
// the project metadata and task summary.
//
// Output format:
//   Phases:
//     [enabled] Phase              Status
//     [ ] Discovery                skipped
//     [ ] Design                   skipped
//     [✓] Implementation           in_progress
//     [✓] Review                   pending
//     [✓] Finalize                 pending
//
// Parameters:
//   - state: Project state to format
//
// Returns:
//   - Formatted string ready for display
func formatPhaseStatus(state *schemas.ProjectState) string {
	var b strings.Builder

	// Phases table
	fmt.Fprintln(&b, "Phases:")
	fmt.Fprintln(&b, "  [enabled] Phase              Status")

	phases := []struct {
		name    string
		enabled bool
		status  string
	}{
		{"Discovery", state.Phases.Discovery.Enabled, state.Phases.Discovery.Status},
		{"Design", state.Phases.Design.Enabled, state.Phases.Design.Status},
		{"Implementation", state.Phases.Implementation.Enabled, state.Phases.Implementation.Status},
		{"Review", state.Phases.Review.Enabled, state.Phases.Review.Status},
		{"Finalize", state.Phases.Finalize.Enabled, state.Phases.Finalize.Status},
	}

	for _, p := range phases {
		enabledMark := " "
		if p.enabled {
			enabledMark = "✓"
		}
		// Pad phase name to 20 characters for alignment
		fmt.Fprintf(&b, "  [%s] %-20s %s\n", enabledMark, p.name, p.status)
	}

	return b.String()
}

// formatArtifactList generates a human-readable artifact list.
//
// Parameters:
//   - state: Project state to format
//   - phase: Phase to show artifacts for (empty string = all phases)
//
// Returns:
//   - Formatted string ready for display
func formatArtifactList(state *schemas.ProjectState, phase string) string {
	var b strings.Builder

	if phase == "" || phase == phaseDiscovery {
		formatPhaseArtifacts(&b, "Discovery", state.Phases.Discovery.Enabled, state.Phases.Discovery.Artifacts)
	}

	if phase == "" || phase == phaseDesign {
		if phase == "" && b.Len() > 0 {
			fmt.Fprintln(&b)
		}
		formatPhaseArtifacts(&b, "Design", state.Phases.Design.Enabled, state.Phases.Design.Artifacts)
	}

	if b.Len() == 0 {
		return "No artifacts\n"
	}

	return b.String()
}

// formatPhaseArtifacts formats artifacts for a single phase.
func formatPhaseArtifacts(b *strings.Builder, phaseName string, enabled bool, artifacts []phases.Artifact) {
	if !enabled {
		fmt.Fprintf(b, "%s Phase (disabled)\n", phaseName)
		return
	}

	fmt.Fprintf(b, "%s Phase:\n", phaseName)

	if len(artifacts) == 0 {
		fmt.Fprintln(b, "  No artifacts")
		return
	}

	fmt.Fprintln(b, "  [approved] Path")
	for _, artifact := range artifacts {
		approvedMark := " "
		if artifact.Approved {
			approvedMark = "✓"
		}
		fmt.Fprintf(b, "  [%s] %s\n", approvedMark, artifact.Path)
	}
}
