package statechart

import (
	"context"
	"fmt"

	"github.com/qmuntal/stateless"
)

// Machine wraps the stateless state machine with project-specific context.
type Machine struct {
	sm           *stateless.StateMachine
	projectState *ProjectState
}

// NewMachine creates a new state machine for project lifecycle management.
// The initial state is determined from the project state (or NoProject if nil).
func NewMachine(projectState *ProjectState) *Machine {
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
func NewMachineAt(initialState State, projectState *ProjectState) *Machine {
	sm := stateless.NewStateMachine(initialState)
	m := &Machine{
		sm:           sm,
		projectState: projectState,
	}

	m.configure()
	return m
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
	return func(ctx context.Context, args ...any) error {
		prompt := GeneratePrompt(PromptContext{
			State:        state,
			ProjectState: m.projectState,
		})
		fmt.Println(prompt)
		return nil
	}
}

// Guard wrapper functions

func (m *Machine) discoveryComplete(ctx context.Context, args ...any) bool {
	if m.projectState == nil {
		return false
	}
	return ArtifactsApproved(m.projectState.Phases.Discovery)
}

func (m *Machine) designComplete(ctx context.Context, args ...any) bool {
	if m.projectState == nil {
		return false
	}
	return ArtifactsApproved(m.projectState.Phases.Design)
}

func (m *Machine) hasAtLeastOneTask(ctx context.Context, args ...any) bool {
	if m.projectState == nil {
		return false
	}
	return HasAtLeastOneTask(m.projectState)
}

func (m *Machine) allTasksComplete(ctx context.Context, args ...any) bool {
	if m.projectState == nil {
		return false
	}
	return AllTasksComplete(m.projectState)
}

func (m *Machine) documentationAssessed(ctx context.Context, args ...any) bool {
	if m.projectState == nil {
		return false
	}
	return DocumentationAssessed(m.projectState)
}

func (m *Machine) checksAssessed(ctx context.Context, args ...any) bool {
	if m.projectState == nil {
		return false
	}
	return ChecksAssessed(m.projectState)
}

func (m *Machine) projectDeleted(ctx context.Context, args ...any) bool {
	if m.projectState == nil {
		return false
	}
	return ProjectDeleted(m.projectState)
}

// Fire triggers an event, causing a state transition if valid.
func (m *Machine) Fire(event Event) error {
	return m.sm.Fire(event)
}

// State returns the current state.
func (m *Machine) State() State {
	return m.sm.MustState().(State)
}

// CanFire checks if an event can be fired from the current state.
func (m *Machine) CanFire(event Event) (bool, error) {
	return m.sm.CanFire(event)
}

// PermittedTriggers returns all events that can be fired from the current state.
func (m *Machine) PermittedTriggers() ([]Event, error) {
	triggers, err := m.sm.PermittedTriggers()
	if err != nil {
		return nil, err
	}
	events := make([]Event, len(triggers))
	for i, t := range triggers {
		events[i] = t.(Event)
	}
	return events, nil
}

// determineCurrentState infers the current state from project state.
// This is used when resuming an existing project.
func determineCurrentState(state *ProjectState) State {
	// This is a simplified version - real implementation would inspect
	// the actual project state to determine current position in lifecycle.

	// For now, return NoProject as default
	// TODO: Implement proper state inference based on phase status
	return NoProject
}
