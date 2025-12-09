package exploration

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
)

// Guard functions for exploration project state transitions.
// Guards are pure functions that examine project state and return boolean values
// indicating whether a transition is allowed.

// allTasksResolved checks if all research topics in exploration phase are completed or abandoned.
// Guards Active → Summarizing transition.
// Returns false if:
//   - Exploration phase doesn't exist
//   - No tasks exist (must have at least one research topic)
//   - Any task has status other than "completed" or "abandoned"
//
// Returns true if all tasks are completed or abandoned.
func allTasksResolved(p *state.Project) bool {
	phase, exists := p.Phases["exploration"]
	if !exists {
		return false
	}

	if len(phase.Tasks) == 0 {
		return false
	}

	for _, task := range phase.Tasks {
		if task.Status != "completed" && task.Status != "abandoned" {
			return false
		}
	}
	return true
}

// allSummariesApproved checks if at least one summary artifact exists and all summaries are approved.
// Guards Summarizing → Finalizing transition.
// Summary artifacts are identified by type == "summary".
// Returns false if:
//   - Exploration phase doesn't exist
//   - No summary artifacts exist
//   - Any summary artifact has Approved == false
//
// Returns true if at least one summary exists and all are approved.
func allSummariesApproved(p *state.Project) bool {
	phase, exists := p.Phases["exploration"]
	if !exists {
		return false
	}

	// Collect all summary artifacts
	summaries := []projschema.ArtifactState{}
	for _, artifact := range phase.Outputs {
		if artifact.Type == "summary" {
			summaries = append(summaries, artifact)
		}
	}

	// Must have at least one summary
	if len(summaries) == 0 {
		return false
	}

	// All summaries must be approved
	for _, summary := range summaries {
		if !summary.Approved {
			return false
		}
	}

	return true
}

// allFinalizationTasksComplete checks if all finalization tasks are completed.
// Guards Finalizing → Completed transition.
// Returns false if:
//   - Finalization phase doesn't exist
//   - No tasks exist
//   - Any task has status != "completed"
//
// Returns true if all tasks are completed.
// Note: Unlike exploration tasks, finalization tasks must be "completed" - "abandoned" is not accepted.
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

// countUnresolvedTasks returns count of pending/in_progress tasks in exploration phase.
// Returns 0 if exploration phase doesn't exist.
// This helper function can be used for status messages.
func countUnresolvedTasks(p *state.Project) int {
	phase, exists := p.Phases["exploration"]
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

// countUnapprovedSummaries returns count of unapproved summary artifacts in exploration phase.
// Returns 0 if exploration phase doesn't exist.
// This helper function can be used for status messages.
func countUnapprovedSummaries(p *state.Project) int {
	phase, exists := p.Phases["exploration"]
	if !exists {
		return 0
	}

	count := 0
	for _, artifact := range phase.Outputs {
		if artifact.Type == "summary" && !artifact.Approved {
			count++
		}
	}
	return count
}
