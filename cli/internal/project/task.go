package project

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
)

// Task represents an implementation task with auto-save operations.
// All mutations automatically persist changes to both task state and project state.
type Task struct {
	project *Project
	id      string
}

// ID returns the task ID.
func (t *Task) ID() string {
	return t.id
}

// Name returns the task name from project state.
func (t *Task) Name() string {
	for _, task := range t.project.State().Phases.Implementation.Tasks {
		if task.Id == t.id {
			return task.Name
		}
	}
	return ""
}

// Status returns the current task status.
func (t *Task) Status() string {
	for _, task := range t.project.State().Phases.Implementation.Tasks {
		if task.Id == t.id {
			return task.Status
		}
	}
	return ""
}

// State returns the task state from disk.
func (t *Task) State() (*schemas.TaskState, error) {
	statePath := filepath.Join("project/phases/implementation/tasks", t.id, "state.yaml")

	var taskState schemas.TaskState
	if err := t.project.readYAML(statePath, &taskState); err != nil {
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
		return sow.ErrInvalidStatus
	}

	// Update in project state
	projectState := t.project.State()
	for i := range projectState.Phases.Implementation.Tasks {
		if projectState.Phases.Implementation.Tasks[i].Id == t.id {
			projectState.Phases.Implementation.Tasks[i].Status = status
			break
		}
	}

	// Update task state file
	taskState, err := t.State()
	if err != nil {
		return err
	}

	taskState.Task.Updated_at = time.Now()

	statePath := filepath.Join("project/phases/implementation/tasks", t.id, "state.yaml")
	if err := t.project.writeYAML(statePath, taskState); err != nil {
		return err
	}

	// Check if all tasks are now complete
	allComplete := true
	for _, task := range projectState.Phases.Implementation.Tasks {
		if task.Status != "completed" && task.Status != "abandoned" {
			allComplete = false
			break
		}
	}

	// If all complete, fire state machine event
	if allComplete && status == "completed" {
		_ = t.project.machine.Fire(statechart.EventAllTasksComplete)
		// Don't fail the operation if transition fails
	}

	// Auto-save project state
	return t.project.save()
}

// IncrementIteration increments the task iteration counter.
func (t *Task) IncrementIteration() error {
	taskState, err := t.State()
	if err != nil {
		return err
	}

	taskState.Task.Iteration++
	taskState.Task.Updated_at = time.Now()

	statePath := filepath.Join("project/phases/implementation/tasks", t.id, "state.yaml")
	return t.project.writeYAML(statePath, taskState)
}

// SetAgent sets the assigned agent for the task.
func (t *Task) SetAgent(agent string) error {
	taskState, err := t.State()
	if err != nil {
		return err
	}

	taskState.Task.Assigned_agent = agent
	taskState.Task.Updated_at = time.Now()

	statePath := filepath.Join("project/phases/implementation/tasks", t.id, "state.yaml")
	return t.project.writeYAML(statePath, taskState)
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

	statePath := filepath.Join("project/phases/implementation/tasks", t.id, "state.yaml")
	return t.project.writeYAML(statePath, taskState)
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

	statePath := filepath.Join("project/phases/implementation/tasks", t.id, "state.yaml")
	return t.project.writeYAML(statePath, taskState)
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
	statePath := filepath.Join("project/phases/implementation/tasks", t.id, "state.yaml")
	if err := t.project.writeYAML(statePath, taskState); err != nil {
		return "", err
	}

	// Create feedback file
	feedbackPath := filepath.Join("project/phases/implementation/tasks", t.id, "feedback", feedbackID+".md")
	feedbackContent := []byte(fmt.Sprintf("# Feedback %s\n\n%s\n", feedbackID, message))
	if err := t.project.writeFile(feedbackPath, feedbackContent); err != nil {
		return "", err
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

			statePath := filepath.Join("project/phases/implementation/tasks", t.id, "state.yaml")
			return t.project.writeYAML(statePath, taskState)
		}
	}

	return fmt.Errorf("feedback not found: %s", feedbackID)
}

// AppendLog appends a log entry to the task log file.
func (t *Task) AppendLog(entry string) error {
	logPath := filepath.Join("project/phases/implementation/tasks", t.id, "log.md")

	// Read existing content
	existing, err := t.project.readFile(logPath)
	if err != nil {
		return fmt.Errorf("failed to read task log: %w", err)
	}

	// Append new entry
	updated := append(existing, []byte(entry)...)

	// Write back
	if err := t.project.writeFile(logPath, updated); err != nil {
		return fmt.Errorf("failed to write task log: %w", err)
	}

	return nil
}
