package state

import (
	"github.com/jmgilman/sow/cli/schemas/project"
)

// convertArtifacts converts CUE-generated ArtifactState slice to ArtifactCollection.
// This is a simple field-by-field copy with no transformations.
// Nil metadata is preserved as nil.
func convertArtifacts(stateArtifacts []project.ArtifactState) ArtifactCollection {
	coll := make(ArtifactCollection, len(stateArtifacts))
	for i, sa := range stateArtifacts {
		coll[i] = Artifact{
			ArtifactState: sa,
		}
	}
	return coll
}

// convertArtifactsToState converts ArtifactCollection back to CUE-generated slice.
// This is the reverse of convertArtifacts.
func convertArtifactsToState(artifacts ArtifactCollection) []project.ArtifactState {
	stateArtifacts := make([]project.ArtifactState, len(artifacts))
	for i, a := range artifacts {
		stateArtifacts[i] = a.ArtifactState
	}
	return stateArtifacts
}

// convertTasks converts CUE-generated TaskState slice to TaskCollection.
// Nested artifacts (inputs/outputs) are recursively converted.
// All fields including optional timestamps are preserved exactly.
func convertTasks(stateTasks []project.TaskState) TaskCollection {
	coll := make(TaskCollection, len(stateTasks))
	for i, st := range stateTasks {
		coll[i] = Task{
			TaskState: st,
		}
	}
	return coll
}

// convertTasksToState converts TaskCollection back to CUE-generated slice.
// This is the reverse of convertTasks.
func convertTasksToState(tasks TaskCollection) []project.TaskState {
	stateTasks := make([]project.TaskState, len(tasks))
	for i, t := range tasks {
		stateTasks[i] = t.TaskState
	}
	return stateTasks
}

// convertPhases converts CUE-generated PhaseState map to PhaseCollection.
// Nested collections (inputs, outputs, tasks) are recursively converted.
// Nil metadata is preserved as nil.
func convertPhases(statePhases map[string]project.PhaseState) PhaseCollection {
	coll := make(PhaseCollection)
	for name, ps := range statePhases {
		coll[name] = &Phase{
			PhaseState: ps,
		}
	}
	return coll
}

// convertPhasesToState converts PhaseCollection back to CUE-generated map.
// This is the reverse of convertPhases.
func convertPhasesToState(phases PhaseCollection) map[string]project.PhaseState {
	statePhases := make(map[string]project.PhaseState)
	for name, p := range phases {
		statePhases[name] = p.PhaseState
	}
	return statePhases
}
