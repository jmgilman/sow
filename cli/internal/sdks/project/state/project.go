package state

import (
	"fmt"

	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/libs/schemas/project"
)

// Initializer is a function that initializes a newly created project.
// It is called during Create() to set up phases, metadata, and any
// project-type-specific initial state.
//
// Parameters:
//   - project: The project being initialized
//   - initialInputs: Optional map of phase name to initial input artifacts (can be nil)
type Initializer func(*Project, map[string][]project.ArtifactState) error

// Event represents a trigger that causes state transitions.
// This is a type alias for SDK state machine events.
type Event = sdkstate.Event

// EventDeterminer is a function that determines which event to fire from a state.
// It examines the project state and returns the appropriate event to advance.
type EventDeterminer func(*Project) (Event, error)

// ProjectTypeConfig is an interface that avoids import cycles.
// The actual implementation is *sdks/project.ProjectTypeConfig.
type ProjectTypeConfig interface {
	// InitialState returns the configured initial state for this project type
	InitialState() State

	// Initialize calls the configured initializer function if present
	Initialize(project *Project, initialInputs map[string][]project.ArtifactState) error

	// BuildMachine builds a state machine for the project
	BuildMachine(project *Project, initialState State) *sdkstate.Machine

	// Validate validates project state against configuration
	Validate(project *Project) error

	// GetTaskSupportingPhases returns the names of all phases that support tasks
	GetTaskSupportingPhases() []string

	// PhaseSupportsTasks checks if a specific phase supports tasks
	PhaseSupportsTasks(phaseName string) bool

	// GetDefaultTaskPhase returns the default phase for task operations based on current state
	GetDefaultTaskPhase(currentState State) string

	// OrchestratorPrompt generates the orchestrator-level prompt for this project type
	OrchestratorPrompt(project *Project) string

	// GetStatePrompt generates the state-specific prompt for a given state
	GetStatePrompt(state State, project *Project) string

	// DetermineEvent determines which event to fire from the current state
	DetermineEvent(project *Project) (Event, error)

	// FireWithPhaseUpdates fires an event and automatically updates phase status
	FireWithPhaseUpdates(machine *sdkstate.Machine, event Event, project *Project) error
}

// Project wraps the CUE-generated ProjectState with runtime behavior.
// It embeds ProjectState for serialization and adds runtime-only fields
// that are not persisted (config, machine, and ctx).
type Project struct {
	project.ProjectState

	// Runtime-only fields (not serialized)
	config  ProjectTypeConfig
	machine *sdkstate.Machine
	ctx     *sow.Context // Context for FS operations
}

// Config returns the project type configuration for this project.
// This provides access to phase configurations, transitions, and other
// type-specific behavior.
// The returned interface{} should be type-asserted to *sdkproject.ProjectTypeConfig.
func (p *Project) Config() ProjectTypeConfig {
	return p.config
}

// Machine returns the state machine for this project.
// This provides access to the current state and state transition logic.
func (p *Project) Machine() *sdkstate.Machine {
	return p.machine
}

// Helper methods for common guard patterns

// PhaseOutputApproved checks if a phase has an approved artifact of the given type.
// Returns false if phase not found, artifact not found, or artifact not approved.
// This is a read-only method used in state machine guards.
func (p *Project) PhaseOutputApproved(phaseName, outputType string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false // Phase not found
	}

	for _, artifact := range phase.Outputs {
		if artifact.Type == outputType && artifact.Approved {
			return true
		}
	}

	return false
}

// PhaseMetadataBool reads a boolean value from phase metadata.
// Returns false if phase not found, key not found, or value is not a boolean.
// This is a read-only method used in state machine guards.
func (p *Project) PhaseMetadataBool(phaseName, key string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false // Phase not found
	}

	if phase.Metadata == nil {
		return false // No metadata
	}

	value, ok := phase.Metadata[key]
	if !ok {
		return false // Key not found
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false // Wrong type
	}

	return boolValue
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
	return true // All tasks completed (or no tasks exist)
}

// Advance progresses the project to its next state using the state machine.
// This is now handled by the SDK state machine, so we delegate to it.
func (p *Project) Advance() error {
	if p.machine == nil {
		return fmt.Errorf("state machine not initialized")
	}
	// The SDK machine handles event determination, guard evaluation, and firing
	// This method is kept for backward compatibility
	return fmt.Errorf("Advance() is deprecated - use state machine directly")
}
