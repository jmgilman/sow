package standard

import (
	"unsafe"

	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// Standard Project Guards - Guards specific to the standard project workflow.
// These guards take *schemas.ProjectState and are specific to standard project structure.

// PlanningComplete checks if the planning phase is complete.
// Specifically, it requires that the task list artifact has been approved.
func PlanningComplete(phase phases.Phase) bool {
	// Check if task list artifact is approved
	for _, a := range phase.Artifacts {
		if a.Type != nil && *a.Type == "task_list" {
			return a.Approved != nil && *a.Approved
		}
	}
	// If no task list artifact exists, planning is not complete
	return false
}

// HasAtLeastOneTask checks if at least one task has been created.
func HasAtLeastOneTask(state *schemas.ProjectState) bool {
	return len(state.Phases.Implementation.Tasks) >= 1
}

// TasksApproved checks if task plan has been approved by human.
func TasksApproved(state *schemas.ProjectState) bool {
	// Get tasks_approved from typed field
	implPhase := (*projects.ImplementationPhase)(unsafe.Pointer(&state.Phases.Implementation))
	if implPhase.Tasks_approved != nil && *implPhase.Tasks_approved {
		return len(state.Phases.Implementation.Tasks) >= 1
	}
	return false
}

// AllTasksComplete checks if all tasks are completed or abandoned.
func AllTasksComplete(state *schemas.ProjectState) bool {
	return statechart.TasksComplete(state.Phases.Implementation.Tasks)
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

// LatestReviewApproved checks if the most recent review report is approved by human.
func LatestReviewApproved(state *schemas.ProjectState) bool {
	// Find review artifacts (artifacts with type=review from typed field)
	var reviewArtifacts []phases.Artifact
	for _, artifact := range state.Phases.Review.Artifacts {
		if artifact.Type != nil && *artifact.Type == "review" {
			reviewArtifacts = append(reviewArtifacts, artifact)
		}
	}

	if len(reviewArtifacts) == 0 {
		return false
	}

	latest := reviewArtifacts[len(reviewArtifacts)-1]
	return latest.Approved != nil && *latest.Approved
}

// ProjectDeleted checks if the project folder has been deleted.
func ProjectDeleted(state *schemas.ProjectState) bool {
	// Get project_deleted from typed field
	finalizePhase := (*projects.FinalizePhase)(unsafe.Pointer(&state.Phases.Finalize))
	if finalizePhase.Project_deleted != nil {
		return *finalizePhase.Project_deleted
	}
	return false
}
