package state

import (
	"context"
	"sync"

	"github.com/jmgilman/sow/libs/schemas/project"
)

// MemoryBackend implements Backend using in-memory storage.
// This is primarily intended for unit tests and development.
// It is NOT intended for production use as state is lost when the process exits.
type MemoryBackend struct {
	state *project.ProjectState
	mu    sync.RWMutex
}

// NewMemoryBackend creates an empty in-memory backend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{}
}

// NewMemoryBackendWithState creates a backend pre-populated with state.
// Useful for testing scenarios that require existing project state.
// A deep copy of the provided state is stored to ensure isolation.
func NewMemoryBackendWithState(state *project.ProjectState) *MemoryBackend {
	return &MemoryBackend{
		state: copyProjectState(state),
	}
}

// Load returns the stored project state.
// Returns ErrNotFound if no state is stored.
// Returns a deep copy to prevent external modifications from affecting the backend.
func (b *MemoryBackend) Load(_ context.Context) (*project.ProjectState, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.state == nil {
		return nil, ErrNotFound
	}

	return copyProjectState(b.state), nil
}

// Save stores the project state in memory.
// A deep copy of the input state is stored to ensure isolation.
// Replaces any existing state completely.
func (b *MemoryBackend) Save(_ context.Context, state *project.ProjectState) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.state = copyProjectState(state)
	return nil
}

// Exists returns whether state is currently stored.
// Always succeeds (returns nil error).
func (b *MemoryBackend) Exists(_ context.Context) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.state != nil, nil
}

// Delete clears the stored state.
// Always succeeds even if state was already nil.
func (b *MemoryBackend) Delete(_ context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.state = nil
	return nil
}

// State returns the raw stored state for test assertions.
// This is NOT part of the Backend interface - it's a test helper.
// May return nil if no state is stored.
// WARNING: This returns the actual internal pointer. Do not use in production code.
func (b *MemoryBackend) State() *project.ProjectState {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.state
}

// copyProjectState creates a deep copy of a ProjectState.
// This is used to ensure the backend's internal state is isolated
// from caller modifications.
func copyProjectState(src *project.ProjectState) *project.ProjectState {
	if src == nil {
		return nil
	}

	dst := &project.ProjectState{
		Name:        src.Name,
		Type:        src.Type,
		Branch:      src.Branch,
		Description: src.Description,
		Created_at:  src.Created_at,
		Updated_at:  src.Updated_at,
		Statechart: project.StatechartState{
			Current_state: src.Statechart.Current_state,
			Updated_at:    src.Statechart.Updated_at,
		},
	}

	// Copy phases map
	if src.Phases != nil {
		dst.Phases = make(map[string]project.PhaseState, len(src.Phases))
		for k, v := range src.Phases {
			dst.Phases[k] = copyPhaseState(v)
		}
	}

	// Copy agent_sessions map
	if src.Agent_sessions != nil {
		dst.Agent_sessions = make(map[string]string, len(src.Agent_sessions))
		for k, v := range src.Agent_sessions {
			dst.Agent_sessions[k] = v
		}
	}

	return dst
}

// copyPhaseState creates a deep copy of a PhaseState.
func copyPhaseState(src project.PhaseState) project.PhaseState {
	dst := project.PhaseState{
		Status:       src.Status,
		Enabled:      src.Enabled,
		Created_at:   src.Created_at,
		Started_at:   src.Started_at,
		Completed_at: src.Completed_at,
		Failed_at:    src.Failed_at,
		Iteration:    src.Iteration,
	}

	// Copy inputs
	if src.Inputs != nil {
		dst.Inputs = make([]project.ArtifactState, len(src.Inputs))
		for i, a := range src.Inputs {
			dst.Inputs[i] = copyArtifactState(a)
		}
	}

	// Copy outputs
	if src.Outputs != nil {
		dst.Outputs = make([]project.ArtifactState, len(src.Outputs))
		for i, a := range src.Outputs {
			dst.Outputs[i] = copyArtifactState(a)
		}
	}

	// Copy tasks
	if src.Tasks != nil {
		dst.Tasks = make([]project.TaskState, len(src.Tasks))
		for i, t := range src.Tasks {
			dst.Tasks[i] = copyTaskState(t)
		}
	}

	// Copy metadata (shallow copy of map values since they're `any`)
	if src.Metadata != nil {
		dst.Metadata = make(map[string]any, len(src.Metadata))
		for k, v := range src.Metadata {
			dst.Metadata[k] = v
		}
	}

	return dst
}

// copyArtifactState creates a deep copy of an ArtifactState.
func copyArtifactState(src project.ArtifactState) project.ArtifactState {
	dst := project.ArtifactState{
		Type:       src.Type,
		Path:       src.Path,
		Approved:   src.Approved,
		Created_at: src.Created_at,
	}

	// Copy metadata (shallow copy of map values since they're `any`)
	if src.Metadata != nil {
		dst.Metadata = make(map[string]any, len(src.Metadata))
		for k, v := range src.Metadata {
			dst.Metadata[k] = v
		}
	}

	return dst
}

// copyTaskState creates a deep copy of a TaskState.
func copyTaskState(src project.TaskState) project.TaskState {
	dst := project.TaskState{
		Id:             src.Id,
		Name:           src.Name,
		Phase:          src.Phase,
		Status:         src.Status,
		Created_at:     src.Created_at,
		Started_at:     src.Started_at,
		Updated_at:     src.Updated_at,
		Completed_at:   src.Completed_at,
		Iteration:      src.Iteration,
		Assigned_agent: src.Assigned_agent,
		Session_id:     src.Session_id,
	}

	// Copy inputs
	if src.Inputs != nil {
		dst.Inputs = make([]project.ArtifactState, len(src.Inputs))
		for i, a := range src.Inputs {
			dst.Inputs[i] = copyArtifactState(a)
		}
	}

	// Copy outputs
	if src.Outputs != nil {
		dst.Outputs = make([]project.ArtifactState, len(src.Outputs))
		for i, a := range src.Outputs {
			dst.Outputs[i] = copyArtifactState(a)
		}
	}

	// Copy metadata (shallow copy of map values since they're `any`)
	if src.Metadata != nil {
		dst.Metadata = make(map[string]any, len(src.Metadata))
		for k, v := range src.Metadata {
			dst.Metadata[k] = v
		}
	}

	return dst
}
