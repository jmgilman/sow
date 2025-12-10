package state

import "fmt"

// PhaseCollection provides map-based access to phases by name.
// It uses a map structure for efficient lookup by phase identifier.
type PhaseCollection map[string]*Phase

// Get retrieves a phase by name.
// Returns an error if the phase is not found.
func (pc PhaseCollection) Get(name string) (*Phase, error) {
	phase, exists := pc[name]
	if !exists {
		return nil, fmt.Errorf("phase not found: %s", name)
	}
	return phase, nil
}

// ArtifactCollection provides slice-based access to artifacts.
// It supports indexed access, addition, and removal operations.
type ArtifactCollection []Artifact

// Add appends an artifact to the collection.
// Always succeeds and returns nil.
func (ac *ArtifactCollection) Add(artifact Artifact) error {
	*ac = append(*ac, artifact)
	return nil
}

// Get retrieves an artifact by index.
// Returns an error if the index is out of range (negative or >= length).
func (ac ArtifactCollection) Get(index int) (*Artifact, error) {
	if index < 0 || index >= len(ac) {
		return nil, fmt.Errorf("index out of range: %d (length: %d)", index, len(ac))
	}
	return &ac[index], nil
}

// Remove deletes an artifact at the specified index.
// Returns an error if the index is out of range.
// Remaining elements are shifted left to fill the gap.
func (ac *ArtifactCollection) Remove(index int) error {
	if index < 0 || index >= len(*ac) {
		return fmt.Errorf("index out of range: %d (length: %d)", index, len(*ac))
	}
	*ac = append((*ac)[:index], (*ac)[index+1:]...)
	return nil
}

// TaskCollection provides slice-based access to tasks.
// It supports lookup by task ID, addition, and removal operations.
type TaskCollection []Task

// Add appends a task to the collection.
// Always succeeds and returns nil.
func (tc *TaskCollection) Add(task Task) error {
	*tc = append(*tc, task)
	return nil
}

// Get retrieves a task by its ID.
// Returns an error if no task with the given ID is found.
func (tc TaskCollection) Get(id string) (*Task, error) {
	for i := range tc {
		if tc[i].Id == id {
			return &tc[i], nil
		}
	}
	return nil, fmt.Errorf("task not found: %s", id)
}

// Remove deletes a task with the specified ID.
// Returns an error if no task with the given ID is found.
// Remaining elements are shifted left to fill the gap.
func (tc *TaskCollection) Remove(id string) error {
	for i := range *tc {
		if (*tc)[i].Id == id {
			*tc = append((*tc)[:i], (*tc)[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", id)
}
