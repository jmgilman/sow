package breakdown

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// Guard functions for breakdown project state transitions.
// Guards are pure functions that examine project state and return boolean values
// indicating whether a transition is allowed.

// hasApprovedDiscoveryDocument checks if an approved discovery document exists.
// Guards Discovery → Active transition.
// Returns false if:
//   - Breakdown phase doesn't exist
//   - No discovery artifact exists in phase outputs
//   - Discovery artifact exists but is not approved
//
// Returns true if at least one discovery artifact is approved.
// This ensures codebase/design context is gathered and validated before work unit identification.
func hasApprovedDiscoveryDocument(p *state.Project) bool {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return false
	}

	// Check for approved discovery artifact
	for _, artifact := range phase.Outputs {
		if artifact.Type == "discovery" && artifact.Approved {
			return true
		}
	}

	return false
}

// allWorkUnitsApproved checks if all work unit tasks are completed or abandoned,
// with at least one completed.
// Guards Active → Publishing transition (combined with dependenciesValid).
// Returns false if:
//   - Breakdown phase doesn't exist
//   - No tasks exist (must have at least one work unit planned)
//   - Any task has status other than "completed" or "abandoned"
//   - All tasks are abandoned (must have at least one completed)
//
// Returns true if all tasks are completed/abandoned AND at least one is completed.
// This ensures the breakdown has meaningful output before advancing.
func allWorkUnitsApproved(p *state.Project) bool {
	phase, exists := p.Phases["breakdown"]
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

// extractTaskDependencies extracts dependency strings from task metadata.
// Returns nil if no valid dependencies found.
func extractTaskDependencies(task projschema.TaskState) []string {
	if task.Metadata == nil {
		return nil
	}

	depsRaw, ok := task.Metadata["dependencies"]
	if !ok {
		return nil
	}

	deps, ok := depsRaw.([]interface{})
	if !ok {
		return nil
	}

	depStrings := make([]string, 0, len(deps))
	for _, d := range deps {
		if depStr, ok := d.(string); ok {
			depStrings = append(depStrings, depStr)
		}
	}

	if len(depStrings) == 0 {
		return nil
	}

	return depStrings
}

// dependenciesValid checks that task dependencies form a valid directed acyclic graph (DAG).
// Guards Active → Publishing transition (combined with allWorkUnitsApproved).
// Validates that:
//   - All dependency references point to valid completed tasks
//   - No cyclic dependencies exist (including self-references)
//
// Returns true if dependencies are valid or no dependencies exist.
// Returns false if cycles or invalid references detected.
// Only completed tasks are validated; abandoned/pending tasks are ignored.
func dependenciesValid(p *state.Project) bool {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return false
	}

	// Build adjacency list and valid task ID set
	graph := make(map[string][]string)
	taskIDs := make(map[string]bool)

	for _, task := range phase.Tasks {
		// Only validate completed tasks
		if task.Status == "completed" {
			taskIDs[task.Id] = true

			// Extract dependencies from metadata
			if deps := extractTaskDependencies(task); deps != nil {
				graph[task.Id] = deps
			}
		}
	}

	// Check all dependencies point to valid task IDs
	for _, deps := range graph {
		for _, depID := range deps {
			if !taskIDs[depID] {
				return false // Invalid dependency reference
			}
		}
	}

	// Check for cycles using depth-first search
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(taskID string) bool {
		visited[taskID] = true
		recStack[taskID] = true

		for _, depID := range graph[taskID] {
			if !visited[depID] {
				if hasCycle(depID) {
					return true
				}
			} else if recStack[depID] {
				return true // Cycle detected
			}
		}

		recStack[taskID] = false
		return false
	}

	// Check all connected components for cycles
	for taskID := range graph {
		if !visited[taskID] {
			if hasCycle(taskID) {
				return false // Cyclic dependency found
			}
		}
	}

	return true
}

// allWorkUnitsPublished checks if all completed work units have been published to GitHub.
// Guards Publishing → Completed transition.
// Returns false if:
//   - Breakdown phase doesn't exist
//   - No completed tasks exist
//   - Any completed task has metadata.published != true
//
// Returns true if all completed tasks have metadata.published == true.
// Ignores abandoned and non-completed tasks.
func allWorkUnitsPublished(p *state.Project) bool {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return false
	}

	hasCompleted := false

	for _, task := range phase.Tasks {
		// Only check completed tasks
		if task.Status == "completed" {
			hasCompleted = true

			// Check metadata exists
			if task.Metadata == nil {
				return false
			}

			// Check published field exists and is true
			publishedRaw, ok := task.Metadata["published"]
			if !ok {
				return false
			}

			published, ok := publishedRaw.(bool)
			if !ok || !published {
				return false
			}
		}
	}

	// Must have at least one completed task
	return hasCompleted
}

// Helper functions for task lifecycle management.

// countUnresolvedTasks returns count of tasks with status != "completed" and != "abandoned".
// Returns 0 if breakdown phase doesn't exist.
// Used in prompts and error messages.
func countUnresolvedTasks(p *state.Project) int {
	phase, exists := p.Phases["breakdown"]
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

// countUnpublishedTasks returns count of completed tasks where metadata.published != true.
// Returns 0 if breakdown phase doesn't exist.
// Only counts completed tasks; ignores abandoned and other statuses.
// Used in prompts and error messages.
func countUnpublishedTasks(p *state.Project) int {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return 0
	}

	count := 0
	for _, task := range phase.Tasks {
		if task.Status == "completed" {
			// Count as unpublished if no metadata
			if task.Metadata == nil {
				count++
				continue
			}

			// Count as unpublished if published field missing or not true
			if published, ok := task.Metadata["published"].(bool); !ok || !published {
				count++
			}
		}
	}
	return count
}

// validateTaskForCompletion validates that a task can be marked as completed.
// Checks:
//   - Task exists in breakdown phase
//   - Task has metadata field
//   - Task metadata contains "artifact_path" key
//   - artifact_path is a valid non-empty string
//   - Artifact exists at the specified path in phase.Outputs
//
// Returns descriptive error if validation fails.
// Returns nil if validation passes.
func validateTaskForCompletion(p *state.Project, taskID string) error {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return fmt.Errorf("breakdown phase not found")
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
	if task.Metadata == nil {
		return fmt.Errorf("task %s has no metadata - set artifact_path before completing", taskID)
	}

	// Check for artifact_path
	artifactPathRaw, exists := task.Metadata["artifact_path"]
	if !exists {
		return fmt.Errorf("task %s has no artifact_path in metadata - link artifact to task before completing", taskID)
	}

	// Validate artifact_path is a non-empty string
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
// Finds task by ID in breakdown phase, reads artifact_path from task metadata,
// finds artifact in phase.Outputs by path, sets artifact.Approved = true,
// and updates project state.
//
// Returns error if task not found, artifact_path invalid, or artifact not found.
func autoApproveArtifact(p *state.Project, taskID string) error {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return fmt.Errorf("breakdown phase not found")
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
	p.Phases["breakdown"] = phase

	return nil
}
