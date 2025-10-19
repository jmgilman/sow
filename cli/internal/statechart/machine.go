package statechart

import (
	"context"
	"fmt"

	"github.com/jmgilman/sow/cli/schemas"
	"github.com/qmuntal/stateless"
)

// Machine wraps the stateless state machine with project-specific context.
type Machine struct {
	sm           *stateless.StateMachine
	projectState *schemas.ProjectState
}

// NewMachine creates a new state machine for project lifecycle management.
// The initial state is determined from the project state (or NoProject if nil).
func NewMachine(projectState *schemas.ProjectState) *Machine {
	var initialState State
	if projectState == nil {
		initialState = NoProject
	} else {
		initialState = determineCurrentState(projectState)
	}

	return NewMachineAt(initialState, projectState)
}

// NewMachineAt creates a new state machine starting at a specific state.
// This is useful when loading state from disk where the state is explicitly stored.
func NewMachineAt(initialState State, projectState *schemas.ProjectState) *Machine {
	sm := stateless.NewStateMachine(initialState)
	m := &Machine{
		sm:           sm,
		projectState: projectState,
	}

	m.configure()
	return m
}

// ProjectState returns the machine's project state for modification.
func (m *Machine) ProjectState() *schemas.ProjectState {
	return m.projectState
}

// SetProjectState sets the machine's project state.
func (m *Machine) SetProjectState(state *schemas.ProjectState) {
	m.projectState = state
}

// configure sets up all state transitions, guards, and entry actions.
func (m *Machine) configure() {
	// NoProject state
	m.sm.Configure(NoProject).
		Permit(EventProjectInit, DiscoveryDecision).
		OnEntry(m.onEntry(NoProject))

	// DiscoveryDecision state
	m.sm.Configure(DiscoveryDecision).
		Permit(EventEnableDiscovery, DiscoveryActive).
		Permit(EventSkipDiscovery, DesignDecision).
		OnEntry(m.onEntry(DiscoveryDecision))

	// DiscoveryActive state
	m.sm.Configure(DiscoveryActive).
		Permit(EventCompleteDiscovery, DesignDecision, m.discoveryComplete).
		OnEntry(m.onEntry(DiscoveryActive))

	// DesignDecision state
	m.sm.Configure(DesignDecision).
		Permit(EventEnableDesign, DesignActive).
		Permit(EventSkipDesign, ImplementationPlanning).
		OnEntry(m.onEntry(DesignDecision))

	// DesignActive state
	m.sm.Configure(DesignActive).
		Permit(EventCompleteDesign, ImplementationPlanning, m.designComplete).
		OnEntry(m.onEntry(DesignActive))

	// ImplementationPlanning state
	m.sm.Configure(ImplementationPlanning).
		Permit(EventTaskCreated, ImplementationExecuting, m.hasAtLeastOneTask).
		OnEntry(m.onEntry(ImplementationPlanning))

	// ImplementationExecuting state
	m.sm.Configure(ImplementationExecuting).
		Permit(EventAllTasksComplete, ReviewActive, m.allTasksComplete).
		OnEntry(m.onEntry(ImplementationExecuting))

	// ReviewActive state
	m.sm.Configure(ReviewActive).
		Permit(EventReviewFail, ImplementationPlanning). // Loop back to re-plan
		Permit(EventReviewPass, FinalizeDocumentation).
		OnEntry(m.onEntry(ReviewActive))

	// FinalizeDocumentation state
	m.sm.Configure(FinalizeDocumentation).
		Permit(EventDocumentationDone, FinalizeChecks, m.documentationAssessed).
		OnEntry(m.onEntry(FinalizeDocumentation))

	// FinalizeChecks state
	m.sm.Configure(FinalizeChecks).
		Permit(EventChecksDone, FinalizeDelete, m.checksAssessed).
		OnEntry(m.onEntry(FinalizeChecks))

	// FinalizeDelete state
	m.sm.Configure(FinalizeDelete).
		Permit(EventProjectDelete, NoProject, m.projectDeleted).
		OnEntry(m.onEntry(FinalizeDelete))
}

// onEntry creates an entry action that outputs the contextual prompt for a state.
func (m *Machine) onEntry(state State) func(context.Context, ...any) error {
	return func(_ context.Context, _ ...any) error {
		prompt := GeneratePrompt(PromptContext{
			State:        state,
			ProjectState: m.projectState,
		})
		fmt.Println(prompt)
		return nil
	}
}

// Guard wrapper functions

func (m *Machine) discoveryComplete(_ context.Context, _ ...any) bool {
	if m.projectState == nil {
		return false
	}
	return ArtifactsApproved(m.projectState.Phases.Discovery)
}

func (m *Machine) designComplete(_ context.Context, _ ...any) bool {
	if m.projectState == nil {
		return false
	}
	return ArtifactsApprovedDesign(m.projectState.Phases.Design)
}

func (m *Machine) hasAtLeastOneTask(_ context.Context, _ ...any) bool {
	if m.projectState == nil {
		return false
	}
	return HasAtLeastOneTask(m.projectState)
}

func (m *Machine) allTasksComplete(_ context.Context, _ ...any) bool {
	if m.projectState == nil {
		return false
	}
	return AllTasksComplete(m.projectState)
}

func (m *Machine) documentationAssessed(_ context.Context, _ ...any) bool {
	if m.projectState == nil {
		return false
	}
	return DocumentationAssessed(m.projectState)
}

func (m *Machine) checksAssessed(_ context.Context, _ ...any) bool {
	if m.projectState == nil {
		return false
	}
	return ChecksAssessed(m.projectState)
}

func (m *Machine) projectDeleted(_ context.Context, _ ...any) bool {
	if m.projectState == nil {
		return false
	}
	return ProjectDeleted(m.projectState)
}

// Fire triggers an event, causing a state transition if valid.
func (m *Machine) Fire(event Event) error {
	if err := m.sm.Fire(event); err != nil {
		return fmt.Errorf("failed to fire event %s: %w", event, err)
	}
	return nil
}

// State returns the current state.
func (m *Machine) State() State {
	state := m.sm.MustState()
	if s, ok := state.(State); ok {
		return s
	}
	// This should never happen if the state machine is properly configured
	return NoProject
}

// CanFire checks if an event can be fired from the current state.
func (m *Machine) CanFire(event Event) (bool, error) {
	can, err := m.sm.CanFire(event)
	if err != nil {
		return false, fmt.Errorf("failed to check if event %s can fire: %w", event, err)
	}
	return can, nil
}

// PermittedTriggers returns all events that can be fired from the current state.
func (m *Machine) PermittedTriggers() ([]Event, error) {
	triggers, err := m.sm.PermittedTriggers()
	if err != nil {
		return nil, fmt.Errorf("failed to get permitted triggers: %w", err)
	}
	events := make([]Event, 0, len(triggers))
	for _, t := range triggers {
		if e, ok := t.(Event); ok {
			events = append(events, e)
		}
	}
	return events, nil
}

// determineCurrentState infers the current state from project state.
// This is used when resuming an existing project.
func determineCurrentState(_ *schemas.ProjectState) State {
	// This is a simplified version - real implementation would inspect
	// the actual project state to determine current position in lifecycle.

	// For now, return NoProject as default
	// TODO: Implement proper state inference based on phase status
	return NoProject
}
