package sow

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/schemas"
)

// Phase name constants used for formatting.
const (
	phaseDiscovery = "discovery"
	phaseDesign    = "design"
)

// FormatStatus generates a human-readable status summary for a project.
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
func FormatStatus(state *schemas.ProjectState) string {
	var b strings.Builder

	// Project header
	fmt.Fprintf(&b, "Project: %s (on %s)\n", state.Project.Name, state.Project.Branch)
	fmt.Fprintf(&b, "Description: %s\n", state.Project.Description)

	// GitHub issue link (if present)
	if state.Project.Github_issue != nil {
		if issueNum, ok := state.Project.Github_issue.(int); ok && issueNum > 0 {
			fmt.Fprintf(&b, "GitHub Issue: #%d\n", issueNum)
		} else if issueNum, ok := state.Project.Github_issue.(float64); ok && issueNum > 0 {
			// Handle JSON unmarshaling as float64
			fmt.Fprintf(&b, "GitHub Issue: #%d\n", int(issueNum))
		}
	}

	// Pull request URL (if created)
	if state.Phases.Finalize.Pr_url != nil {
		if prURL, ok := state.Phases.Finalize.Pr_url.(string); ok && prURL != "" {
			fmt.Fprintf(&b, "Pull Request: %s\n", prURL)
		}
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
func formatTaskSummary(b *strings.Builder, tasks []schemas.Task) {
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

// FormatPhaseStatus generates a human-readable phase status table.
//
// This is similar to FormatStatus but focuses only on phases without
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
func FormatPhaseStatus(state *schemas.ProjectState) string {
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

// FormatArtifactList generates a human-readable artifact list.
//
// Parameters:
//   - state: Project state to format
//   - phase: Phase to show artifacts for (empty string = all phases)
//
// Returns:
//   - Formatted string ready for display
func FormatArtifactList(state *schemas.ProjectState, phase string) string {
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
func formatPhaseArtifacts(b *strings.Builder, phaseName string, enabled bool, artifacts []schemas.Artifact) {
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

// ============================================================================
// Task Formatting
// ============================================================================

// FormatTaskList generates a human-readable task list.
//
// Output format:
//   Tasks:
//     ID   Status        Name
//     010  pending       Add authentication
//     020  in_progress   Create database schema
//     030  completed     Setup project structure
//
// Parameters:
//   - tasks: List of tasks to format
//
// Returns:
//   - Formatted string ready for display
func FormatTaskList(tasks []schemas.Task) string {
	if len(tasks) == 0 {
		return "No tasks yet\n"
	}

	var b strings.Builder

	// Header
	fmt.Fprintln(&b, "Tasks:")
	fmt.Fprintln(&b, "  ID   Status        Name")

	// Sort tasks by ID
	sorted := make([]schemas.Task, len(tasks))
	copy(sorted, tasks)
	// Simple bubble sort for 3-digit IDs
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Id > sorted[j+1].Id {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	// Format each task
	for _, task := range sorted {
		// Pad status to 13 characters for alignment
		fmt.Fprintf(&b, "  %s  %-13s %s\n", task.Id, task.Status, task.Name)
	}

	return b.String()
}

// FormatTaskStatus generates a detailed human-readable task status.
//
// Output format:
//   Task: 010 - Add authentication
//   Status: in_progress
//   Phase: implementation
//   Created: 2024-01-15 10:30:00
//   Started: 2024-01-15 11:00:00
//   Iteration: 1
//   Assigned Agent: implementer
//   Parallel: false
//   Dependencies: none
//
// Parameters:
//   - taskState: Task state to format
//
// Returns:
//   - Formatted string ready for display
func FormatTaskStatus(taskState *schemas.TaskState) string {
	var b strings.Builder

	// Task header
	fmt.Fprintf(&b, "Task: %s - %s\n", taskState.Task.Id, taskState.Task.Name)
	fmt.Fprintf(&b, "Status: %s\n", taskState.Task.Status)
	fmt.Fprintf(&b, "Phase: %s\n\n", taskState.Task.Phase)

	// Timestamps
	fmt.Fprintln(&b, "Timestamps:")
	fmt.Fprintf(&b, "  Created:   %s\n", taskState.Task.Created_at.Format("2006-01-02 15:04:05"))
	if taskState.Task.Started_at != nil {
		if startedStr, ok := taskState.Task.Started_at.(string); ok {
			fmt.Fprintf(&b, "  Started:   %s\n", startedStr)
		}
	} else {
		fmt.Fprintln(&b, "  Started:   not started")
	}
	if taskState.Task.Completed_at != nil {
		if completedStr, ok := taskState.Task.Completed_at.(string); ok {
			fmt.Fprintf(&b, "  Completed: %s\n", completedStr)
		}
	} else {
		fmt.Fprintln(&b, "  Completed: not completed")
	}
	fmt.Fprintf(&b, "  Updated:   %s\n\n", taskState.Task.Updated_at.Format("2006-01-02 15:04:05"))

	// Task metadata
	fmt.Fprintf(&b, "Iteration: %d\n", taskState.Task.Iteration)
	fmt.Fprintf(&b, "Assigned Agent: %s\n\n", taskState.Task.Assigned_agent)

	// References
	if len(taskState.Task.References) > 0 {
		fmt.Fprintln(&b, "References:")
		for _, ref := range taskState.Task.References {
			fmt.Fprintf(&b, "  - %s\n", ref)
		}
		fmt.Fprintln(&b)
	}

	// Feedback
	if len(taskState.Task.Feedback) > 0 {
		fmt.Fprintf(&b, "Feedback: %d items\n\n", len(taskState.Task.Feedback))
	}

	// Files modified
	if len(taskState.Task.Files_modified) > 0 {
		fmt.Fprintln(&b, "Files Modified:")
		for _, file := range taskState.Task.Files_modified {
			fmt.Fprintf(&b, "  - %s\n", file)
		}
	}

	return b.String()
}
