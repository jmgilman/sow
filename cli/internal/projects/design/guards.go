package design

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// Guard functions for design project state transitions.
// Guards are pure functions that examine project state and return boolean values
// indicating whether a transition is allowed.

// allDocumentsApproved checks if all design tasks are completed or abandoned,
// with at least one completed.
// Guards Active → Finalizing transition.
// Returns false if:
//   - Design phase doesn't exist
//   - No tasks exist (must have at least one design document planned)
//   - Any task has status other than "completed" or "abandoned"
//   - All tasks are abandoned (must have at least one completed)
//
// Returns true if all tasks are completed/abandoned AND at least one is completed.
// This ensures the design has meaningful output before advancing.
func allDocumentsApproved(p *state.Project) bool {
	phase, exists := p.Phases["design"]
	if !exists {
		return false
	}

	if len(phase.Tasks) == 0 {
		return false
	}

	hasCompleted := false
	for _, task := range phase.Tasks {
		if task.Status == "completed" {
			hasCompleted = true
		} else if task.Status != "abandoned" {
			// Task is not completed or abandoned, so not all are resolved
			return false
		}
	}

	// Must have at least one completed task
	return hasCompleted
}

// allFinalizationTasksComplete checks if all finalization tasks are completed.
// Guards Finalizing → Completed transition.
// Returns false if:
//   - Finalization phase doesn't exist
//   - No tasks exist
//   - Any task has status != "completed"
//
// Returns true if all tasks are completed.
// Note: Unlike design tasks, finalization tasks must be "completed" - "abandoned" is not accepted.
func allFinalizationTasksComplete(p *state.Project) bool {
	phase, exists := p.Phases["finalization"]
	if !exists {
		return false
	}

	if len(phase.Tasks) == 0 {
		return false
	}

	for _, task := range phase.Tasks {
		if task.Status != "completed" {
			return false
		}
	}
	return true
}

// Helper functions for task lifecycle management.

// countUnresolvedTasks returns count of pending/in_progress tasks in design phase.
// Returns 0 if design phase doesn't exist.
// This helper function can be used for status messages.
func countUnresolvedTasks(p *state.Project) int {
	phase, exists := p.Phases["design"]
	if !exists {
		return 0
	}

	count := 0
	for _, task := range phase.Tasks {
		if task.Status != "completed" && task.Status != "abandoned" {
			count++
		}
	}
	return count
}

// validateTaskForCompletion validates that a task can be marked as completed.
// Checks:
//   - Task exists in design phase
//   - Task has metadata field
//   - Task metadata contains "artifact_path" key
//   - artifact_path is a valid string
//   - Artifact exists at the specified path in phase.Outputs
//
// Returns descriptive error if validation fails.
// Returns nil if validation passes.
func validateTaskForCompletion(p *state.Project, taskID string) error {
	phase, exists := p.Phases["design"]
	if !exists {
		return fmt.Errorf("design phase not found")
	}

	// Find task
	var task *projschema.TaskState
	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			task = &phase.Tasks[i]
			break
		}
	}
	if task == nil {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Check for metadata
	if len(task.Metadata) == 0 {
		return fmt.Errorf("task %s has no metadata - set artifact_path before completing", taskID)
	}

	// Check for artifact_path
	artifactPathRaw, exists := task.Metadata["artifact_path"]
	if !exists {
		return fmt.Errorf("task %s has no artifact_path in metadata - link artifact to task before completing", taskID)
	}

	// Validate artifact_path is a string
	artifactPath, ok := artifactPathRaw.(string)
	if !ok || artifactPath == "" {
		return fmt.Errorf("task %s has no artifact_path in metadata - link artifact to task before completing", taskID)
	}

	// Check artifact exists in phase outputs
	found := false
	for _, artifact := range phase.Outputs {
		if artifact.Path == artifactPath {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("artifact not found at %s - add artifact before completing task", artifactPath)
	}

	return nil
}

// autoApproveArtifact automatically approves the artifact linked to a completed task.
// Called during task completion (status update to "completed").
// Finds task by ID in design phase, reads artifact_path from task metadata,
// finds artifact in phase.Outputs by path, sets artifact.Approved = true,
// and updates project state.
//
// Returns error if task not found, artifact_path invalid, or artifact not found.
func autoApproveArtifact(p *state.Project, taskID string) error {
	phase, exists := p.Phases["design"]
	if !exists {
		return fmt.Errorf("design phase not found")
	}

	// Find task
	var task *projschema.TaskState
	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			task = &phase.Tasks[i]
			break
		}
	}
	if task == nil {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Get artifact_path from metadata
	artifactPathRaw, exists := task.Metadata["artifact_path"]
	if !exists {
		return fmt.Errorf("task %s has invalid artifact_path in metadata", taskID)
	}

	artifactPath, ok := artifactPathRaw.(string)
	if !ok || artifactPath == "" {
		return fmt.Errorf("task %s has invalid artifact_path in metadata", taskID)
	}

	// Find and approve artifact
	found := false
	for i := range phase.Outputs {
		if phase.Outputs[i].Path == artifactPath {
			phase.Outputs[i].Approved = true
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("artifact not found at %s", artifactPath)
	}

	// Update project state
	p.Phases["design"] = phase

	return nil
}
