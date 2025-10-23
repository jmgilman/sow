package statechart

import (
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// Guards for conditional transitions

// ArtifactsApproved checks if all artifacts for a phase are approved (or no artifacts exist).
func ArtifactsApproved(phase phases.Phase) bool {
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
func ArtifactsApprovedDesign(phase phases.Phase) bool {
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

// TasksApproved checks if task plan has been approved by human.
func TasksApproved(state *schemas.ProjectState) bool {
	// Check metadata for tasks_approved
	tasksApproved := false
	if state.Phases.Implementation.Metadata != nil {
		if approved, ok := state.Phases.Implementation.Metadata["tasks_approved"].(bool); ok {
			tasksApproved = approved
		}
	}
	return tasksApproved && len(state.Phases.Implementation.Tasks) >= 1
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

// LatestReviewApproved checks if the most recent review report is approved by human.
func LatestReviewApproved(state *schemas.ProjectState) bool {
	// Find review artifacts (artifacts with type=review metadata)
	var reviewArtifacts []phases.Artifact
	for _, artifact := range state.Phases.Review.Artifacts {
		if artifact.Metadata != nil {
			if artifactType, ok := artifact.Metadata["type"].(string); ok && artifactType == "review" {
				reviewArtifacts = append(reviewArtifacts, artifact)
			}
		}
	}

	if len(reviewArtifacts) == 0 {
		return false
	}

	latest := reviewArtifacts[len(reviewArtifacts)-1]
	return latest.Approved
}

// ProjectDeleted checks if the project folder has been deleted.
func ProjectDeleted(state *schemas.ProjectState) bool {
	// Check metadata for project_deleted
	if state.Phases.Finalize.Metadata != nil {
		if deleted, ok := state.Phases.Finalize.Metadata["project_deleted"].(bool); ok {
			return deleted
		}
	}
	return false
}
