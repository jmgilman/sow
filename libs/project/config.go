package project

import (
	"context"
	"sort"

	"github.com/jmgilman/sow/libs/project/state"
	"github.com/jmgilman/sow/libs/schemas/project"
)

// Initializer is a function that initializes a new project with phases and initial state.
// It is called during project creation to set up type-specific initial state.
type Initializer func(p *state.Project, initialInputs map[string][]project.ArtifactState) error

// ProjectTypeConfig holds the complete configuration for a project type.
// It defines phases, state transitions, prompts, and initialization logic.
//
//nolint:revive // "Project" prefix distinguishes this from generic machine configs
type ProjectTypeConfig struct {
	// name is the project type identifier (e.g., "standard", "exploration")
	name string

	// phaseConfigs are the phase configurations indexed by phase name
	phaseConfigs map[string]*PhaseConfig

	// initialState is the starting state of the state machine
	initialState State

	// transitions are all state transitions for this project type
	transitions []TransitionConfig

	// onAdvance are event determiners mapped by state
	// These determine which event to use for the Advance() command
	onAdvance map[State]EventDeterminer

	// prompts are prompt generators mapped by state
	// These generate contextual prompts for users in each state
	prompts map[State]PromptGenerator

	// orchestratorPrompt generates project-type-specific orchestrator guidance
	// This explains how the project type works and how orchestrator coordinates work
	orchestratorPrompt PromptGenerator

	// initializer is called during Create() to initialize the project
	// with phases, metadata, and any type-specific initial state
	initializer Initializer

	// branches are branch configurations mapped by state
	// Stored for introspection and debugging
	branches map[State]*BranchConfig
}

// Name returns the project type name (e.g., "standard", "exploration").
func (ptc *ProjectTypeConfig) Name() string {
	return ptc.name
}

// InitialState returns the initial state for new projects of this type.
func (ptc *ProjectTypeConfig) InitialState() State {
	return ptc.initialState
}

// Phases returns the phase configurations.
func (ptc *ProjectTypeConfig) Phases() map[string]*PhaseConfig {
	return ptc.phaseConfigs
}

// GetPhaseForState returns the phase name that owns the given state.
// Returns empty string if the state doesn't belong to any phase's start or end state.
func (ptc *ProjectTypeConfig) GetPhaseForState(s State) string {
	for name, config := range ptc.phaseConfigs {
		if config.startState == s || config.endState == s {
			return name
		}
	}
	return ""
}

// IsPhaseStartState returns true if the state is the phase's start state.
// Returns false if the phase doesn't exist or the state doesn't match.
func (ptc *ProjectTypeConfig) IsPhaseStartState(phaseName string, s State) bool {
	config, exists := ptc.phaseConfigs[phaseName]
	if !exists {
		return false
	}
	return config.startState == s
}

// IsPhaseEndState returns true if the state is the phase's end state.
// Returns false if the phase doesn't exist or the state doesn't match.
func (ptc *ProjectTypeConfig) IsPhaseEndState(phaseName string, s State) bool {
	config, exists := ptc.phaseConfigs[phaseName]
	if !exists {
		return false
	}
	return config.endState == s
}

// GetTransition returns the transition config for the given state change.
// Returns nil if no matching transition is found.
func (ptc *ProjectTypeConfig) GetTransition(from, to State, event Event) *TransitionConfig {
	for i := range ptc.transitions {
		tc := &ptc.transitions[i]
		if tc.From == from && tc.To == to && tc.Event == event {
			return tc
		}
	}
	return nil
}

// Initialize sets up a new project with phases and initial state.
// Returns nil if no initializer is configured.
func (ptc *ProjectTypeConfig) Initialize(p *state.Project, initialInputs map[string][]project.ArtifactState) error {
	if ptc.initializer == nil {
		return nil
	}
	return ptc.initializer(p, initialInputs)
}

// GetStatePrompt returns the prompt for a specific state.
// Returns empty string if no prompt is configured for the state.
func (ptc *ProjectTypeConfig) GetStatePrompt(s State, p *state.Project) string {
	gen, exists := ptc.prompts[s]
	if !exists {
		return ""
	}
	return gen(p)
}

// OrchestratorPrompt returns the orchestrator prompt for this project type.
// This explains how the project type works and how the orchestrator should coordinate work.
// Returns empty string if no orchestrator prompt is configured.
func (ptc *ProjectTypeConfig) OrchestratorPrompt(p *state.Project) string {
	if ptc.orchestratorPrompt == nil {
		return ""
	}
	return ptc.orchestratorPrompt(p)
}

// GetTaskSupportingPhases returns the names of all phases that support tasks.
// Returns an empty slice if no phases support tasks.
// Phase names are returned in sorted order for deterministic behavior.
func (ptc *ProjectTypeConfig) GetTaskSupportingPhases() []string {
	var phases []string
	for name, config := range ptc.phaseConfigs {
		if config.supportsTasks {
			phases = append(phases, name)
		}
	}
	sort.Strings(phases)
	return phases
}

// PhaseSupportsTasks checks if a specific phase supports tasks.
// Returns false if the phase doesn't exist or doesn't support tasks.
func (ptc *ProjectTypeConfig) PhaseSupportsTasks(phaseName string) bool {
	config, exists := ptc.phaseConfigs[phaseName]
	if !exists {
		return false
	}
	return config.supportsTasks
}

// GetDefaultTaskPhase returns the default phase for task operations based on current state.
// Returns empty string if no phase supports tasks or state mapping is ambiguous.
//
// Logic:
//  1. Check if current state maps to a phase's start or end state
//  2. If that phase supports tasks, return it
//  3. Otherwise return first task-supporting phase alphabetically
func (ptc *ProjectTypeConfig) GetDefaultTaskPhase(currentState State) string {
	// Try to map current state to a phase
	for name, config := range ptc.phaseConfigs {
		if (config.startState == currentState || config.endState == currentState) && config.supportsTasks {
			return name
		}
	}

	// Fallback: return first task-supporting phase
	phases := ptc.GetTaskSupportingPhases()
	if len(phases) > 0 {
		return phases[0]
	}
	return ""
}

// IsBranchingState checks if a state has branches configured via AddBranch.
// Returns true if the state was configured with AddBranch (state-determined branching).
func (ptc *ProjectTypeConfig) IsBranchingState(s State) bool {
	_, exists := ptc.branches[s]
	return exists
}

// DetermineEvent determines which event to fire from the current state.
// Returns the event to fire, or an error if no determiner is configured.
func (ptc *ProjectTypeConfig) DetermineEvent(p *state.Project) (Event, error) {
	currentState := State(p.Statechart.Current_state)
	determiner, exists := ptc.onAdvance[currentState]
	if !exists {
		return "", &ErrNoDeterminer{State: currentState}
	}
	return determiner(p)
}

// GetAvailableTransitions returns all configured transitions from a state.
// This returns the transitions defined in the project type configuration,
// not filtered by guards.
func (ptc *ProjectTypeConfig) GetAvailableTransitions(from State) []TransitionInfo {
	var result []TransitionInfo

	// Check if this is a branching state
	branchConfig, isBranching := ptc.branches[from]
	if isBranching {
		// Add transitions from branch paths
		for _, path := range branchConfig.branches {
			result = append(result, TransitionInfo{
				Event:       path.event,
				From:        from,
				To:          path.to,
				Description: path.description,
				GuardDesc:   path.guardTemplate.Description,
			})
		}
	}

	// Add direct transitions (from AddTransition calls)
	for _, tc := range ptc.transitions {
		if tc.From != from {
			continue
		}
		// Skip if this transition is part of a branch (already included above)
		if isBranching && ptc.isBranchTransition(branchConfig, tc) {
			continue
		}
		result = append(result, TransitionInfo{
			Event:       tc.Event,
			From:        tc.From,
			To:          tc.To,
			Description: tc.description,
			GuardDesc:   tc.guardTemplate.Description,
		})
	}

	// Sort by event name for deterministic output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Event < result[j].Event
	})

	return result
}

// isBranchTransition checks if a transition is part of a branch configuration.
func (ptc *ProjectTypeConfig) isBranchTransition(bc *BranchConfig, tc TransitionConfig) bool {
	for _, path := range bc.branches {
		if tc.Event == path.event && tc.To == path.to {
			return true
		}
	}
	return false
}

// TransitionInfo describes a single available transition from a state.
// Used by introspection methods to provide structured information about
// state machine configuration without exposing internal implementation details.
type TransitionInfo struct {
	Event       Event  // Event that triggers this transition
	From        State  // Source state
	To          State  // Target state
	Description string // Human-readable description (empty if not provided)
	GuardDesc   string // Guard description (empty if no guard)
}

// BuildProjectMachine builds a state machine for a project using this project type's configuration.
// It creates a state machine with all configured transitions, binding guard templates and
// actions to the project instance via closures.
//
// Guard templates (func(*Project) bool) are bound to guard functions (func() bool) by
// capturing the project in a closure. This allows guards to access live project state
// while matching the state machine SDK's expected signature.
//
// Similarly, onEntry and onExit actions are bound via closures to match the SDK's
// expected signature (func(context.Context, ...any) error).
//
// Parameters:
//   - project: The project instance to bind to guards and actions
//   - initialState: The starting state for the machine
//
// Returns:
//   - A configured state machine ready to handle events
//
// Example usage:
//
//	config := NewProjectTypeConfigBuilder("standard").
//	    AddTransition(
//	        StateActive,
//	        StateComplete,
//	        EventFinish,
//	        WithProjectGuard("all tasks complete", func(p *state.Project) bool {
//	            return p.AllTasksComplete()
//	        }),
//	    ).
//	    Build()
//
//	machine := config.BuildProjectMachine(project, StateActive)
func (ptc *ProjectTypeConfig) BuildProjectMachine(
	proj *state.Project,
	initialState State,
) *Machine {
	// Create prompt function that uses project type config prompts
	var promptFunc PromptFunc
	if len(ptc.prompts) > 0 {
		promptFunc = func(s State) string {
			gen := ptc.prompts[s]
			if gen == nil {
				return "" // No prompt configured for this state
			}
			return gen(proj) // Call project type config prompt generator
		}
	}

	builder := NewBuilder(initialState, promptFunc)

	// Add all transitions with guards and actions bound to project instance
	for _, tc := range ptc.transitions {
		var opts []TransitionOption

		// Bind guard template to project instance via closure
		// Need to capture tc.guardTemplate in a local variable for the closure
		if tc.guardTemplate.Func != nil {
			guardFunc := tc.guardTemplate.Func
			if tc.guardTemplate.Description != "" {
				guardDesc := tc.guardTemplate.Description
				// Use WithGuardDescription if description is provided
				opts = append(opts, WithGuardDescription(guardDesc, func() bool {
					return guardFunc(proj)
				}))
			} else {
				// Fallback to WithGuard without description
				opts = append(opts, WithGuard(func() bool {
					return guardFunc(proj)
				}))
			}
		}

		// Bind onExit action to project instance via closure
		if tc.onExit != nil {
			onExitFunc := tc.onExit // Capture for closure
			opts = append(opts, WithOnExit(func(_ context.Context, _ ...any) error {
				return onExitFunc(proj)
			}))
		}

		// Bind onEntry action to project instance via closure
		if tc.onEntry != nil {
			onEntryFunc := tc.onEntry // Capture for closure
			opts = append(opts, WithOnEntry(func(_ context.Context, _ ...any) error {
				return onEntryFunc(proj)
			}))
		}

		builder.AddTransition(
			tc.From,
			tc.To,
			tc.Event,
			opts...,
		)
	}

	return builder.Build()
}

// FireWithPhaseUpdates fires an event and automatically updates phase statuses
// based on the state transition. This wraps the standard Fire() call with
// automatic phase status management.
//
// Phase status updates:
//   - When exiting a phase's end state -> mark phase "completed" (unless WithProjectFailedPhase specified)
//   - When entering a phase's start state -> mark phase "in_progress" (only if "pending")
//
// This approach respects explicit status changes (like MarkPhaseFailed) while
// automating the common case of successful phase progression.
//
// Example usage:
//
//	err := config.FireWithPhaseUpdates(machine, EventPlanningComplete, project)
func (ptc *ProjectTypeConfig) FireWithPhaseUpdates(
	machine *Machine,
	event Event,
	proj *state.Project,
) error {
	// Capture old state before transition
	oldState := machine.State()

	// Fire the event (executes user-defined guards and actions)
	if err := machine.Fire(event); err != nil {
		return &ErrTransitionFailed{Cause: err}
	}

	// Capture new state after transition
	newState := machine.State()

	// Look up the transition config to check for special phase status handling
	transitionConfig := ptc.GetTransition(oldState, newState, event)

	// Update phase status when exiting a phase's end state
	if err := ptc.updateExitingPhaseStatus(oldState, transitionConfig, proj); err != nil {
		return err
	}

	// Update phase status when entering a phase's start state
	if err := ptc.updateEnteringPhaseStatus(newState, proj); err != nil {
		return err
	}

	return nil
}

// updateExitingPhaseStatus updates the status of a phase when exiting its end state.
// Marks the phase as "failed" if configured in the transition, otherwise "completed".
func (ptc *ProjectTypeConfig) updateExitingPhaseStatus(
	oldState State,
	transitionConfig *TransitionConfig,
	proj *state.Project,
) error {
	phaseName := ptc.GetPhaseForState(oldState)
	if phaseName == "" {
		return nil
	}

	if !ptc.IsPhaseEndState(phaseName, oldState) {
		return nil
	}

	// Check if this transition explicitly marks the phase as failed
	if transitionConfig != nil && transitionConfig.failedPhase == phaseName {
		// Mark as failed instead of completed
		if err := state.MarkPhaseFailed(proj, phaseName); err != nil {
			return &ErrPhaseStatusUpdate{Phase: phaseName, Operation: "mark failed", Cause: err}
		}
		return nil
	}

	// Normal case: mark as completed
	if err := state.MarkPhaseCompleted(proj, phaseName); err != nil {
		return &ErrPhaseStatusUpdate{Phase: phaseName, Operation: "mark completed", Cause: err}
	}
	return nil
}

// updateEnteringPhaseStatus updates the status of a phase when entering its start state.
// Marks the phase as "in_progress" if currently "pending".
func (ptc *ProjectTypeConfig) updateEnteringPhaseStatus(
	newState State,
	proj *state.Project,
) error {
	phaseName := ptc.GetPhaseForState(newState)
	if phaseName == "" {
		return nil
	}

	if !ptc.IsPhaseStartState(phaseName, newState) {
		return nil
	}

	// MarkPhaseInProgress only updates if status is "pending"
	if err := state.MarkPhaseInProgress(proj, phaseName); err != nil {
		return &ErrPhaseStatusUpdate{Phase: phaseName, Operation: "mark in_progress", Cause: err}
	}
	return nil
}
