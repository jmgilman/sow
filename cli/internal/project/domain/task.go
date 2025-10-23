package domain

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/schemas"
)

// Task represents an implementation task.
// This concrete type works with any project type through the Project interface.
type Task struct {
	Project Project // Interface, not concrete type
	ID      string
}

// Name returns the task name from project state.
func (t *Task) Name() string {
	state := t.Project.Machine().ProjectState()
	for _, task := range state.Phases.Implementation.Tasks {
		if task.Id == t.ID {
			return task.Name
		}
	}
	return ""
}

// Status returns the current task status from project state.
func (t *Task) Status() string {
	state := t.Project.Machine().ProjectState()
	for _, task := range state.Phases.Implementation.Tasks {
		if task.Id == t.ID {
			return task.Status
		}
	}
	return ""
}

// State returns the task state from disk.
func (t *Task) State() (*schemas.TaskState, error) {
	statePath := filepath.Join("project/phases/implementation/tasks", t.ID, "state.yaml")

	var taskState schemas.TaskState
	if err := t.Project.ReadYAML(statePath, &taskState); err != nil {
		return nil, fmt.Errorf("failed to read task state: %w", err)
	}

	return &taskState, nil
}

// SetStatus updates the task status and persists changes.
func (t *Task) SetStatus(status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
		"abandoned":   true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Update in project state
	projectState := t.Project.Machine().ProjectState()
	for i := range projectState.Phases.Implementation.Tasks {
		if projectState.Phases.Implementation.Tasks[i].Id == t.ID {
			projectState.Phases.Implementation.Tasks[i].Status = status
			break
		}
	}

	// Update task state file
	taskState, err := t.State()
	if err != nil {
		return err
	}

	taskState.Task.Status = status
	taskState.Task.Updated_at = time.Now()

	// Set timestamps based on status
	now := time.Now()
	if status == "in_progress" && taskState.Task.Started_at == nil {
		taskState.Task.Started_at = &now
	}
	if (status == "completed" || status == "abandoned") && taskState.Task.Completed_at == nil {
		taskState.Task.Completed_at = &now
		if taskState.Task.Started_at == nil {
			taskState.Task.Started_at = &now
		}
	}

	statePath := filepath.Join("project/phases/implementation/tasks", t.ID, "state.yaml")
	if err := t.Project.WriteYAML(statePath, taskState); err != nil {
		return fmt.Errorf("failed to write task state: %w", err)
	}

	// Check if all tasks are now complete
	allComplete := true
	for _, task := range projectState.Phases.Implementation.Tasks {
		if task.Status != "completed" && task.Status != "abandoned" {
			allComplete = false
			break
		}
	}

	// If all complete, mark implementation as complete and fire state machine event
	if allComplete && status == "completed" {
		now := time.Now()
		projectState.Phases.Implementation.Status = "completed"
		projectState.Phases.Implementation.Completed_at = &now

		// Fire state machine event to transition to review
		if err := t.Project.Machine().Fire(statechart.EventAllTasksComplete); err == nil {
			// Transition succeeded - set review phase status
			projectState.Phases.Review.Status = "in_progress"
			if projectState.Phases.Review.Started_at == nil {
				projectState.Phases.Review.Started_at = &now
			}
		}
		// Silently ignore transition errors - task status update should still succeed
	}

	// Auto-save project state
	if err := t.Project.Save(); err != nil {
		return fmt.Errorf("failed to save project after updating task status: %w", err)
	}
	return nil
}

// IncrementIteration increments the task iteration counter.
func (t *Task) IncrementIteration() error {
	taskState, err := t.State()
	if err != nil {
		return err
	}

	taskState.Task.Iteration++
	taskState.Task.Updated_at = time.Now()

	statePath := filepath.Join("project/phases/implementation/tasks", t.ID, "state.yaml")
	if err := t.Project.WriteYAML(statePath, taskState); err != nil {
		return fmt.Errorf("failed to write task state after incrementing iteration: %w", err)
	}
	return nil
}

// SetAgent sets the assigned agent for the task.
func (t *Task) SetAgent(agent string) error {
	taskState, err := t.State()
	if err != nil {
		return err
	}

	taskState.Task.Assigned_agent = agent
	taskState.Task.Updated_at = time.Now()

	statePath := filepath.Join("project/phases/implementation/tasks", t.ID, "state.yaml")
	if err := t.Project.WriteYAML(statePath, taskState); err != nil {
		return fmt.Errorf("failed to write task state after setting agent: %w", err)
	}
	return nil
}

// AddReference adds a reference path to the task.
func (t *Task) AddReference(path string) error {
	taskState, err := t.State()
	if err != nil {
		return err
	}

	// Check if reference already exists
	for _, ref := range taskState.Task.References {
		if ref == path {
			return nil // Already exists
		}
	}

	// Add reference
	taskState.Task.References = append(taskState.Task.References, path)
	taskState.Task.Updated_at = time.Now()

	statePath := filepath.Join("project/phases/implementation/tasks", t.ID, "state.yaml")
	if err := t.Project.WriteYAML(statePath, taskState); err != nil {
		return fmt.Errorf("failed to write task state after adding reference: %w", err)
	}
	return nil
}

// AddFile adds a modified file path to the task.
func (t *Task) AddFile(path string) error {
	taskState, err := t.State()
	if err != nil {
		return err
	}

	// Check if file already exists
	for _, file := range taskState.Task.Files_modified {
		if file == path {
			return nil // Already exists
		}
	}

	// Add file
	taskState.Task.Files_modified = append(taskState.Task.Files_modified, path)
	taskState.Task.Updated_at = time.Now()

	statePath := filepath.Join("project/phases/implementation/tasks", t.ID, "state.yaml")
	if err := t.Project.WriteYAML(statePath, taskState); err != nil {
		return fmt.Errorf("failed to write task state after adding file: %w", err)
	}
	return nil
}

// AddFeedback creates a new feedback entry for the task.
// Returns the generated feedback ID.
func (t *Task) AddFeedback(message string) (string, error) {
	taskState, err := t.State()
	if err != nil {
		return "", err
	}

	// Generate feedback ID (001, 002, 003...)
	feedbackCount := len(taskState.Task.Feedback)
	feedbackID := fmt.Sprintf("%03d", feedbackCount+1)

	// Create feedback entry
	now := time.Now()
	feedback := schemas.Feedback{
		Id:         feedbackID,
		Status:     "pending",
		Created_at: now,
	}

	// Add to task state
	taskState.Task.Feedback = append(taskState.Task.Feedback, feedback)
	taskState.Task.Updated_at = time.Now()

	// Save task state
	statePath := filepath.Join("project/phases/implementation/tasks", t.ID, "state.yaml")
	if err := t.Project.WriteYAML(statePath, taskState); err != nil {
		return "", fmt.Errorf("failed to write task state: %w", err)
	}

	// Create feedback file
	feedbackPath := filepath.Join("project/phases/implementation/tasks", t.ID, "feedback", feedbackID+".md")
	feedbackContent := []byte(fmt.Sprintf("# Feedback %s\n\n%s\n", feedbackID, message))
	if err := t.Project.WriteFile(feedbackPath, feedbackContent); err != nil {
		return "", fmt.Errorf("failed to write feedback file: %w", err)
	}

	return feedbackID, nil
}

// MarkFeedbackAddressed marks a feedback entry as addressed.
func (t *Task) MarkFeedbackAddressed(feedbackID string) error {
	taskState, err := t.State()
	if err != nil {
		return err
	}

	// Find and update feedback
	for i := range taskState.Task.Feedback {
		if taskState.Task.Feedback[i].Id == feedbackID {
			taskState.Task.Feedback[i].Status = "addressed"
			taskState.Task.Updated_at = time.Now()

			statePath := filepath.Join("project/phases/implementation/tasks", t.ID, "state.yaml")
			if err := t.Project.WriteYAML(statePath, taskState); err != nil {
				return fmt.Errorf("failed to write task state after marking feedback addressed: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("feedback not found: %s", feedbackID)
}

// Log creates and appends a structured log entry to the task log.
// Automatically determines agent ID from task state (assigned_agent + iteration).
func (t *Task) Log(action, result string, opts ...LogOption) error {
	// Read task state to get agent info
	state, err := t.State()
	if err != nil {
		return fmt.Errorf("failed to read task state: %w", err)
	}

	// Build agent ID from state
	agentID := buildAgentID(state.Task.Assigned_agent, int(state.Task.Iteration))

	entry := &LogEntry{
		Timestamp: time.Now(),
		AgentID:   agentID,
		Action:    action,
		Result:    result,
	}

	// Apply options
	for _, opt := range opts {
		opt(entry)
	}

	// Validate
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("invalid log entry: %w", err)
	}

	// Format and append
	formatted := entry.Format()
	return t.appendLog(formatted)
}

// appendLog appends a raw log entry to the task log file.
func (t *Task) appendLog(entry string) error {
	logPath := filepath.Join("project/phases/implementation/tasks", t.ID, "log.md")

	// Read existing content
	existing, err := t.Project.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("failed to read task log: %w", err)
	}

	// Append new entry
	updated := append(existing, []byte(entry)...)

	// Write back
	if err := t.Project.WriteFile(logPath, updated); err != nil {
		return fmt.Errorf("failed to write task log: %w", err)
	}

	return nil
}

// buildAgentID constructs an agent ID from role and iteration.
func buildAgentID(role string, iteration int) string {
	if role == "orchestrator" || iteration == 0 {
		return role
	}
	return fmt.Sprintf("%s-%d", role, iteration)
}

// Validate checks if the log entry has valid values.
func (e *LogEntry) Validate() error {
	if e.AgentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if e.Action == "" {
		return fmt.Errorf("action is required")
	}
	if e.Result == "" {
		return fmt.Errorf("result is required")
	}
	return nil
}

// Format renders the log entry as structured markdown.
func (e *LogEntry) Format() string {
	var b []byte

	// Front matter
	b = append(b, "---\n"...)
	b = append(b, fmt.Sprintf("timestamp: %s\n", e.Timestamp.Format("2006-01-02 15:04:05"))...)
	b = append(b, fmt.Sprintf("agent: %s\n", e.AgentID)...)
	b = append(b, fmt.Sprintf("action: %s\n", e.Action)...)
	b = append(b, fmt.Sprintf("result: %s\n", e.Result)...)

	// Optional files list
	if len(e.Files) > 0 {
		b = append(b, "files:\n"...)
		for _, file := range e.Files {
			b = append(b, fmt.Sprintf("  - %s\n", file)...)
		}
	}

	b = append(b, "---\n"...)

	// Optional notes section
	if e.Notes != "" {
		b = append(b, "\n"...)
		b = append(b, e.Notes...)
		b = append(b, "\n"...)
	}

	return string(b)
}
