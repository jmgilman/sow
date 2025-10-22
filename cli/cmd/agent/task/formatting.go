package task

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/schemas"
)

// formatTaskList generates a human-readable task list.
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
func formatTaskList(tasks []schemas.Task) string {
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

// formatTaskStatus generates a detailed human-readable task status.
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
func formatTaskStatus(taskState *schemas.TaskState) string {
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
