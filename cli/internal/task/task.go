// Package task provides business logic for task operations.
//
// This package handles task-specific concerns like initializing task state,
// generating gap-numbered IDs, validating status transitions, and formatting
// task information for display.
package task

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/schemas"
)

// Task status constants.
const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusAbandoned  = "abandoned"
)

// validStatuses maps status names to their validity.
var validStatuses = map[string]bool{
	StatusPending:    true,
	StatusInProgress: true,
	StatusCompleted:  true,
	StatusAbandoned:  true,
}

// GenerateNextTaskID generates the next gap-numbered task ID.
//
// Gap numbering uses increments of 10 (010, 020, 030...) to allow
// insertion of tasks between existing ones if needed.
//
// Parameters:
//   - existingTasks: Current list of tasks
//
// Returns:
//   - Next available gap-numbered ID (e.g., "010", "020", "030")
func GenerateNextTaskID(existingTasks []schemas.Task) string {
	if len(existingTasks) == 0 {
		return "010"
	}

	// Find the highest ID
	maxID := 0
	for _, task := range existingTasks {
		// Parse ID as integer (remove leading zeros)
		id, err := strconv.Atoi(task.Id)
		if err != nil {
			continue
		}
		if id > maxID {
			maxID = id
		}
	}

	// Generate next ID with gap of 10
	nextID := maxID + 10
	return fmt.Sprintf("%03d", nextID)
}

// ValidateTaskID validates a task ID follows gap numbering format.
//
// Valid IDs are 3-digit zero-padded numbers:
// 010, 011, 012, ..., 990
//
// Auto-generated IDs use increments of 10 (010, 020, 030) but manual
// IDs can use any number to allow insertion between existing tasks.
//
// Parameters:
//   - id: Task ID to validate
//
// Returns:
//   - nil if ID is valid
//   - error if ID is invalid
func ValidateTaskID(id string) error {
	// Must be exactly 3 characters
	if len(id) != 3 {
		return fmt.Errorf("invalid task ID '%s': must be 3 digits (e.g., '010', '020')", id)
	}

	// Must be numeric
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid task ID '%s': must be numeric", id)
	}

	// Must be between 010 and 990
	if idNum < 10 || idNum > 990 {
		return fmt.Errorf("invalid task ID '%s': must be between 010 and 990", id)
	}

	return nil
}

// NewTaskState creates an initial TaskState for a new task.
//
// Parameters:
//   - id: Gap-numbered task ID (e.g., "010")
//   - name: Human-readable task name
//   - assignedAgent: Agent type to execute this task (e.g., "implementer")
//
// Returns:
//   - Fully initialized TaskState ready to be written to disk
func NewTaskState(id, name, assignedAgent string) *schemas.TaskState {
	now := time.Now()

	state := &schemas.TaskState{}

	state.Task.Id = id
	state.Task.Name = name
	state.Task.Phase = "implementation" // Always implementation in 5-phase model
	state.Task.Status = StatusPending
	state.Task.Created_at = now
	state.Task.Started_at = nil
	state.Task.Updated_at = now
	state.Task.Completed_at = nil
	state.Task.Iteration = 1
	state.Task.Assigned_agent = assignedAgent
	state.Task.References = []string{}
	state.Task.Feedback = []schemas.Feedback{}
	state.Task.Files_modified = []string{}

	return state
}

// AddTaskToProjectState adds a task to the project's implementation phase.
//
// This creates a lightweight Task entry in the project state. The detailed
// TaskState should be created separately using NewTaskState() and written
// via TaskFS.WriteState().
//
// Parameters:
//   - projectState: Project state to modify
//   - id: Gap-numbered task ID
//   - name: Task name
//   - parallel: Whether task can run in parallel with others
//   - dependencies: List of task IDs this task depends on (nil for none)
//
// Returns:
//   - nil on success
//   - error if task ID already exists or validation fails
func AddTaskToProjectState(projectState *schemas.ProjectState, id, name string, parallel bool, dependencies []string) error {
	// Validate ID format
	if err := ValidateTaskID(id); err != nil {
		return err
	}

	// Check if task ID already exists
	for _, existingTask := range projectState.Phases.Implementation.Tasks {
		if existingTask.Id == id {
			return fmt.Errorf("task '%s' already exists", id)
		}
	}

	// Validate dependencies exist
	if len(dependencies) > 0 {
		for _, depID := range dependencies {
			found := false
			for _, task := range projectState.Phases.Implementation.Tasks {
				if task.Id == depID {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("dependency task '%s' not found", depID)
			}
		}
	}

	// Create task entry
	var deps any
	if len(dependencies) > 0 {
		deps = dependencies
	} else {
		deps = nil
	}

	task := schemas.Task{
		Id:           id,
		Name:         name,
		Status:       StatusPending,
		Parallel:     parallel,
		Dependencies: deps,
	}

	// Add to project state
	projectState.Phases.Implementation.Tasks = append(projectState.Phases.Implementation.Tasks, task)
	projectState.Project.Updated_at = time.Now()

	return nil
}

// FindTaskByID finds a task in the project state by ID.
//
// Parameters:
//   - projectState: Project state to search
//   - id: Task ID to find
//
// Returns:
//   - Pointer to task if found
//   - nil if not found
func FindTaskByID(projectState *schemas.ProjectState, id string) *schemas.Task {
	for i := range projectState.Phases.Implementation.Tasks {
		if projectState.Phases.Implementation.Tasks[i].Id == id {
			return &projectState.Phases.Implementation.Tasks[i]
		}
	}
	return nil
}

// UpdateTaskStatusInProject updates a task's status in the project state.
//
// This updates the lightweight Task entry. The detailed TaskState should
// be updated separately via TaskFS.
//
// Parameters:
//   - projectState: Project state to modify
//   - id: Task ID to update
//   - newStatus: New status value
//
// Returns:
//   - nil on success
//   - error if task not found or status is invalid
func UpdateTaskStatusInProject(projectState *schemas.ProjectState, id, newStatus string) error {
	// Validate status
	if err := ValidateStatus(newStatus); err != nil {
		return err
	}

	// Find and update task
	task := FindTaskByID(projectState, id)
	if task == nil {
		return fmt.Errorf("task '%s' not found", id)
	}

	task.Status = newStatus
	projectState.Project.Updated_at = time.Now()

	return nil
}

// ValidateStatus validates a task status value.
//
// Parameters:
//   - status: Status to validate
//
// Returns:
//   - nil if status is valid
//   - error if status is invalid
func ValidateStatus(status string) error {
	if !validStatuses[status] {
		return fmt.Errorf("invalid status '%s': must be one of pending, in_progress, completed, abandoned", status)
	}
	return nil
}

// UpdateTaskStatus updates the status in a TaskState with timestamps.
//
// Sets appropriate timestamps based on status transition:
//   - in_progress: Sets started_at if not already set
//   - completed/abandoned: Sets completed_at and started_at if not set
//
// Parameters:
//   - taskState: Task state to modify
//   - newStatus: New status value
//
// Returns:
//   - nil on success
//   - error if status is invalid
func UpdateTaskStatus(taskState *schemas.TaskState, newStatus string) error {
	// Validate status
	if err := ValidateStatus(newStatus); err != nil {
		return err
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	// Update status
	taskState.Task.Status = newStatus
	taskState.Task.Updated_at = now

	// Set timestamps based on status
	switch newStatus {
	case StatusInProgress:
		// Set started_at if not already set
		if taskState.Task.Started_at == nil {
			taskState.Task.Started_at = nowStr
		}

	case StatusCompleted, StatusAbandoned:
		// Set completed_at
		taskState.Task.Completed_at = nowStr
		// Set started_at if not already set
		if taskState.Task.Started_at == nil {
			taskState.Task.Started_at = nowStr
		}
	}

	return nil
}

// RemoveTaskFromProject removes a task from the project state.
//
// Parameters:
//   - projectState: Project state to modify
//   - id: Task ID to remove
//
// Returns:
//   - nil on success
//   - error if task not found or has dependents
func RemoveTaskFromProject(projectState *schemas.ProjectState, id string) error {
	// Check if any other tasks depend on this one
	for _, task := range projectState.Phases.Implementation.Tasks {
		if task.Dependencies != nil {
			// Dependencies can be stored as []string or []interface{}
			switch deps := task.Dependencies.(type) {
			case []string:
				for _, depID := range deps {
					if depID == id {
						return fmt.Errorf("cannot remove task '%s': task '%s' depends on it", id, task.Id)
					}
				}
			case []interface{}:
				for _, dep := range deps {
					if depStr, ok := dep.(string); ok && depStr == id {
						return fmt.Errorf("cannot remove task '%s': task '%s' depends on it", id, task.Id)
					}
				}
			}
		}
	}

	// Find and remove task
	for i, task := range projectState.Phases.Implementation.Tasks {
		if task.Id == id {
			projectState.Phases.Implementation.Tasks = append(
				projectState.Phases.Implementation.Tasks[:i],
				projectState.Phases.Implementation.Tasks[i+1:]...,
			)
			projectState.Project.Updated_at = time.Now()
			return nil
		}
	}

	return fmt.Errorf("task '%s' not found", id)
}

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
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Id < sorted[j].Id
	})

	// Format each task
	for _, task := range sorted {
		// Pad status to 13 characters for alignment
		fmt.Fprintf(&b, "  %s  %-13s %s\n", task.Id, task.Status, task.Name)
	}

	return b.String()
}

// IncrementTaskIteration increments the task's iteration counter.
//
// Increments the iteration field and updates the updated_at timestamp.
// This is typically called when an implementer needs to retry a task
// with human feedback.
//
// Parameters:
//   - taskState: Task state to modify
//
// Returns:
//   - nil on success
func IncrementTaskIteration(taskState *schemas.TaskState) error {
	taskState.Task.Iteration++
	taskState.Task.Updated_at = time.Now()
	return nil
}

// SetTaskAgent updates the assigned agent for a task.
//
// Changes which agent type should execute this task and updates
// the updated_at timestamp.
//
// Parameters:
//   - taskState: Task state to modify
//   - agent: New agent name (e.g., "implementer", "reviewer", "architect")
//
// Returns:
//   - error if agent name is empty
func SetTaskAgent(taskState *schemas.TaskState, agent string) error {
	if agent == "" {
		return fmt.Errorf("agent name cannot be empty")
	}
	taskState.Task.Assigned_agent = agent
	taskState.Task.Updated_at = time.Now()
	return nil
}

// AddTaskReference adds a context reference path to the task.
//
// References are paths relative to .sow/ that the agent should read
// when executing the task (e.g., refs, knowledge, or other context).
// Duplicates are automatically filtered out.
//
// Parameters:
//   - taskState: Task state to modify
//   - path: Path relative to .sow/ (e.g., "refs/python-style/conventions.md")
//
// Returns:
//   - error if path is empty
func AddTaskReference(taskState *schemas.TaskState, path string) error {
	if path == "" {
		return fmt.Errorf("reference path cannot be empty")
	}

	// Check for duplicates
	for _, existing := range taskState.Task.References {
		if existing == path {
			// Already exists, no-op
			return nil
		}
	}

	// Add reference
	taskState.Task.References = append(taskState.Task.References, path)
	taskState.Task.Updated_at = time.Now()
	return nil
}

// AddModifiedFile tracks a file that was modified during task execution.
//
// Workers use this to record which files they changed. Paths should be
// relative to the repository root. Duplicates are automatically filtered out.
//
// Parameters:
//   - taskState: Task state to modify
//   - path: Path relative to repo root (e.g., "src/auth/jwt.py")
//
// Returns:
//   - error if path is empty
func AddModifiedFile(taskState *schemas.TaskState, path string) error {
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check for duplicates
	for _, existing := range taskState.Task.Files_modified {
		if existing == path {
			// Already exists, no-op
			return nil
		}
	}

	// Add file
	taskState.Task.Files_modified = append(taskState.Task.Files_modified, path)
	taskState.Task.Updated_at = time.Now()
	return nil
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
	fmt.Fprintf(&b, "  Created:   %s\n", taskState.Task.Created_at.Format(time.RFC3339))
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
	fmt.Fprintf(&b, "  Updated:   %s\n\n", taskState.Task.Updated_at.Format(time.RFC3339))

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

// GenerateNextFeedbackID generates the next feedback ID for a task.
//
// Feedback IDs are zero-padded 3-digit numbers (001, 002, 003...) that
// increment sequentially for each piece of feedback.
//
// Parameters:
//   - taskState: Task state to examine
//
// Returns:
//   - Next available feedback ID (e.g., "001", "002", "003")
func GenerateNextFeedbackID(taskState *schemas.TaskState) string {
	if len(taskState.Task.Feedback) == 0 {
		return "001"
	}

	// Find the highest ID
	maxID := 0
	for _, feedback := range taskState.Task.Feedback {
		// Parse ID as integer
		id, err := strconv.Atoi(feedback.Id)
		if err != nil {
			continue
		}
		if id > maxID {
			maxID = id
		}
	}

	// Generate next ID
	nextID := maxID + 1
	return fmt.Sprintf("%03d", nextID)
}

// AddFeedback creates a new feedback entry in the task state.
//
// Adds a Feedback struct to the task's feedback array with status "pending"
// and updates the updated_at timestamp.
//
// Parameters:
//   - taskState: Task state to modify
//   - feedbackID: Feedback ID (e.g., "001", "002")
//
// Returns:
//   - error if feedback ID is invalid or already exists
func AddFeedback(taskState *schemas.TaskState, feedbackID string) error {
	// Validate feedback ID format (must be 3 digits)
	if len(feedbackID) != 3 {
		return fmt.Errorf("invalid feedback ID '%s': must be 3 digits (e.g., '001')", feedbackID)
	}
	if _, err := strconv.Atoi(feedbackID); err != nil {
		return fmt.Errorf("invalid feedback ID '%s': must be numeric", feedbackID)
	}

	// Check for duplicates
	for _, existing := range taskState.Task.Feedback {
		if existing.Id == feedbackID {
			return fmt.Errorf("feedback '%s' already exists", feedbackID)
		}
	}

	// Create feedback entry
	feedback := schemas.Feedback{
		Id:         feedbackID,
		Created_at: time.Now(),
		Status:     "pending",
	}

	// Add to task state
	taskState.Task.Feedback = append(taskState.Task.Feedback, feedback)
	taskState.Task.Updated_at = time.Now()

	return nil
}

// MarkFeedbackAddressed updates a feedback's status to "addressed".
//
// Finds the feedback by ID and changes its status from "pending" to "addressed".
// Updates the task's updated_at timestamp.
//
// Parameters:
//   - taskState: Task state to modify
//   - feedbackID: Feedback ID to mark as addressed
//
// Returns:
//   - error if feedback not found
func MarkFeedbackAddressed(taskState *schemas.TaskState, feedbackID string) error {
	// Find feedback by ID
	found := false
	for i := range taskState.Task.Feedback {
		if taskState.Task.Feedback[i].Id == feedbackID {
			taskState.Task.Feedback[i].Status = "addressed"
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("feedback '%s' not found", feedbackID)
	}

	taskState.Task.Updated_at = time.Now()
	return nil
}
