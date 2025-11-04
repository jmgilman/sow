package standard

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// Guard helper functions check transition conditions.
// These are called by guard closures defined in standard.go.

// phaseOutputApproved checks if a specific output artifact type is approved.
// Returns false if phase not found, artifact not found, or artifact not approved.
func phaseOutputApproved(p *state.Project, phaseName, outputType string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false
	}

	for _, output := range phase.Outputs {
		if output.Type == outputType && output.Approved {
			return true
		}
	}
	return false
}

// phaseMetadataBool gets a boolean value from phase metadata.
// Returns false if key missing, wrong type, or phase missing.
func phaseMetadataBool(p *state.Project, phaseName, key string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false
	}

	if phase.Metadata == nil {
		return false
	}

	val, ok := phase.Metadata[key]
	if !ok {
		return false
	}

	boolVal, ok := val.(bool)
	if !ok {
		return false
	}

	return boolVal
}

// allTasksComplete checks if all implementation tasks are completed or abandoned.
// Returns false if implementation phase missing or if no tasks exist.
func allTasksComplete(p *state.Project) bool {
	phase, exists := p.Phases["implementation"]
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

// latestReviewApproved checks if the most recent review artifact is approved.
// Returns false if review phase missing or if no review artifacts found.
func latestReviewApproved(p *state.Project) bool {
	phase, exists := p.Phases["review"]
	if !exists {
		return false
	}

	// Find latest review output by iterating backwards
	for i := len(phase.Outputs) - 1; i >= 0; i-- {
		if phase.Outputs[i].Type == "review" {
			return phase.Outputs[i].Approved
		}
	}

	return false
}

// projectDeleted checks if the project_deleted flag is set in finalize metadata.
func projectDeleted(p *state.Project) bool {
	return phaseMetadataBool(p, "finalize", "project_deleted")
}
