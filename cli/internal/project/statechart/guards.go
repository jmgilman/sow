package statechart

import (
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// Common Guards - Reusable guard predicates for state machine transitions.
// These guards can be used by any project type and operate on typed domain objects.

// TasksComplete checks if all tasks in a task list are completed or abandoned.
// Returns false if the task list is empty.
func TasksComplete(tasks []phases.Task) bool {
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

// ArtifactsApproved checks if all artifacts in a list are approved.
// Returns false if the artifact list is empty.
func ArtifactsApproved(artifacts []phases.Artifact) bool {
	if len(artifacts) == 0 {
		return false
	}
	for _, a := range artifacts {
		if !a.Approved {
			return false
		}
	}
	return true
}

// MinTaskCount checks if there are at least 'minCount' tasks in the task list.
func MinTaskCount(tasks []phases.Task, minCount int) bool {
	return len(tasks) >= minCount
}

// HasArtifactWithType checks if any artifact has the specified type in its metadata.
func HasArtifactWithType(artifacts []phases.Artifact, artifactType string) bool {
	for _, artifact := range artifacts {
		if artifact.Metadata != nil {
			if typ, ok := artifact.Metadata["type"].(string); ok && typ == artifactType {
				return true
			}
		}
	}
	return false
}

// AnyTaskInProgress checks if any task has "in_progress" status.
func AnyTaskInProgress(tasks []phases.Task) bool {
	for _, t := range tasks {
		if t.Status == "in_progress" {
			return true
		}
	}
	return false
}

