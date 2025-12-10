package state

import (
	"context"
	"fmt"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/qmuntal/stateless"
)

// ProjectTypeConfig defines the interface for project type configuration.
// This interface exists to avoid import cycles - the actual implementation
// lives in the parent project package and provides type-specific behavior.
type ProjectTypeConfig interface {
	// Name returns the configured name for this project type.
	Name() string
}

// Project wraps the CUE-generated ProjectState with runtime behavior.
// It provides methods for state machine integration and storage operations.
// The embedded ProjectState is serialized; the runtime fields (backend,
// config, machine) are not persisted.
type Project struct {
	project.ProjectState

	backend Backend
	config  ProjectTypeConfig
	machine *stateless.StateMachine
}

// NewProject creates a new Project wrapper around a ProjectState.
// The config and machine fields are nil until set via SetConfig/SetMachine,
// which is typically done during Load() or Create() operations.
func NewProject(state project.ProjectState, backend Backend) *Project {
	return &Project{
		ProjectState: state,
		backend:      backend,
	}
}

// AllTasksComplete checks if all tasks across all phases are completed.
// Returns true if all tasks have status "completed", or if no tasks exist (vacuous truth).
// This is a read-only method used in state machine guards.
func (p *Project) AllTasksComplete() bool {
	for _, phase := range p.Phases {
		for _, task := range phase.Tasks {
			if task.Status != "completed" {
				return false
			}
		}
	}
	return true
}

// Backend returns the storage backend for this project.
func (p *Project) Backend() Backend {
	return p.backend
}

// Config returns the project type configuration.
// Returns nil if not yet set via SetConfig.
func (p *Project) Config() ProjectTypeConfig {
	return p.config
}

// Machine returns the state machine for this project.
// Returns nil if not yet set via SetMachine.
func (p *Project) Machine() *stateless.StateMachine {
	return p.machine
}

// PhaseMetadataBool reads a boolean value from phase metadata.
// Returns false if phase not found, key not found, or value is not a boolean.
// This is a read-only method used in state machine guards.
func (p *Project) PhaseMetadataBool(phaseName, key string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false
	}

	if phase.Metadata == nil {
		return false
	}

	value, ok := phase.Metadata[key]
	if !ok {
		return false
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false
	}

	return boolValue
}

// PhaseOutputApproved checks if a phase has an approved artifact of the given type.
// Returns false if phase not found, artifact not found, or artifact not approved.
// This is a read-only method used in state machine guards.
func (p *Project) PhaseOutputApproved(phaseName, outputType string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false
	}

	for _, artifact := range phase.Outputs {
		if artifact.Type == outputType && artifact.Approved {
			return true
		}
	}

	return false
}

// Save persists the current project state to the backend.
func (p *Project) Save(ctx context.Context) error {
	if err := p.backend.Save(ctx, &p.ProjectState); err != nil {
		return fmt.Errorf("save project state: %w", err)
	}
	return nil
}

// SetConfig sets the project type configuration.
// This is called during Load/Create to attach the config from the registry.
func (p *Project) SetConfig(config ProjectTypeConfig) {
	p.config = config
}

// SetMachine sets the state machine for this project.
// This is called during Load/Create to attach the built machine.
func (p *Project) SetMachine(machine *stateless.StateMachine) {
	p.machine = machine
}
