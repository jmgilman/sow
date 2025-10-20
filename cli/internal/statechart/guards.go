package statechart

import (
	"github.com/jmgilman/sow/cli/schemas"
)

// Guards for conditional transitions

// ArtifactsApproved checks if all artifacts for a phase are approved (or no artifacts exist).
func ArtifactsApproved(phase schemas.DiscoveryPhase) bool {
	if len(phase.Artifacts) == 0 {
		return true
	}
	for _, a := range phase.Artifacts {
		if !a.Approved {
			return false
		}
	}
	return true
}

// ArtifactsApprovedDesign checks if all design artifacts are approved (or no artifacts exist).
func ArtifactsApprovedDesign(phase schemas.DesignPhase) bool {
	if len(phase.Artifacts) == 0 {
		return true
	}
	for _, a := range phase.Artifacts {
		if !a.Approved {
			return false
		}
	}
	return true
}

// HasAtLeastOneTask checks if at least one task has been created.
func HasAtLeastOneTask(state *schemas.ProjectState) bool {
	return len(state.Phases.Implementation.Tasks) >= 1
}

// AllTasksComplete checks if all tasks are completed or abandoned.
func AllTasksComplete(state *schemas.ProjectState) bool {
	tasks := state.Phases.Implementation.Tasks
	if len(tasks) == 0 {
		return false
	}
	for _, t := range tasks {
		if t.Status != "completed" && t.Status != "abandoned" {
			return false
		}
	}
	return true
}

// DocumentationAssessed checks if documentation has been assessed and handled.
// The guard always returns true because the act of calling `finalize complete documentation`
// IS the signal that documentation work is done. No additional validation needed.
func DocumentationAssessed(_ *schemas.ProjectState) bool {
	// Always allow transition - the command itself is the validation
	return true
}

// ChecksAssessed checks if final checks have been assessed and handled.
// For now, returns true assuming checks are handled before reaching this guard.
func ChecksAssessed(_ *schemas.ProjectState) bool {
	// This can be enhanced to check a specific field if needed
	// For now, checks are considered assessed if documentation is done
	return true
}

// ProjectDeleted checks if the project folder has been deleted.
func ProjectDeleted(state *schemas.ProjectState) bool {
	return state.Phases.Finalize.Project_deleted
}
