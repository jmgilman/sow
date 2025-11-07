package project

import (
	"context"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	stateMachine "github.com/jmgilman/sow/cli/internal/sdks/state"
)

// BuildMachine builds a state machine for a project using this project type's configuration.
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
//	        WithGuard(func(p *state.Project) bool {
//	            return p.AllTasksComplete()
//	        }),
//	    ).
//	    Build()
//
//	machine := config.BuildMachine(project, StateActive)
func (ptc *ProjectTypeConfig) BuildMachine(
	project *state.Project,
	initialState stateMachine.State,
) *stateMachine.Machine {
	// Create prompt function that uses project type config prompts
	// This closure captures the project wrapper and looks up prompt generators
	// from the project type config's prompts map
	var promptFunc stateMachine.PromptFunc
	if len(ptc.prompts) > 0 {
		promptFunc = func(state stateMachine.State) string {
			gen := ptc.prompts[state]
			if gen == nil {
				return "" // No prompt configured for this state
			}
			return gen(project) // Call project type config prompt generator
		}
	}

	builder := stateMachine.NewBuilder(initialState, promptFunc)

	// Add all transitions with guards and actions bound to project instance
	for _, tc := range ptc.transitions {
		var opts []stateMachine.TransitionOption

		// Bind guard template to project instance via closure
		if tc.guardTemplate.Func != nil {
			if tc.guardTemplate.Description != "" {
				// Use WithGuardDescription if description is provided
				opts = append(opts, stateMachine.WithGuardDescription(tc.guardTemplate.Description, func() bool {
					return tc.guardTemplate.Func(project)
				}))
			} else {
				// Fallback to WithGuard without description
				opts = append(opts, stateMachine.WithGuard(func() bool {
					return tc.guardTemplate.Func(project)
				}))
			}
		}

		// Bind onExit action to project instance via closure
		if tc.onExit != nil {
			opts = append(opts, stateMachine.WithOnExit(func(_ context.Context, _ ...any) error {
				return tc.onExit(project)
			}))
		}

		// Bind onEntry action to project instance via closure
		if tc.onEntry != nil {
			opts = append(opts, stateMachine.WithOnEntry(func(_ context.Context, _ ...any) error {
				return tc.onEntry(project)
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
//   - When exiting a phase's end state → mark phase "completed" (unless already "failed")
//   - When entering a phase's start state → mark phase "in_progress" (only if "pending")
//
// This approach respects explicit status changes (like MarkPhaseFailed) while
// automating the common case of successful phase progression.
//
// Example usage:
//
//	err := config.FireWithPhaseUpdates(machine, EventPlanningComplete, project)
func (ptc *ProjectTypeConfig) FireWithPhaseUpdates(
	machine *stateMachine.Machine,
	event stateMachine.Event,
	project *state.Project,
) error {
	// Capture old state before transition
	oldState := machine.State()

	// Fire the event (executes user-defined guards and actions)
	if err := machine.Fire(event); err != nil {
		return fmt.Errorf("transition failed: %w", err)
	}

	// Capture new state after transition
	newState := machine.State()

	// Look up the transition config to check for special phase status handling
	transitionConfig := ptc.GetTransition(oldState, newState, event)

	// Update phase status when exiting a phase's end state
	if err := ptc.updateExitingPhaseStatus(oldState, transitionConfig, project); err != nil {
		return err
	}

	// Update phase status when entering a phase's start state
	if err := ptc.updateEnteringPhaseStatus(newState, project); err != nil {
		return err
	}

	return nil
}

// updateExitingPhaseStatus updates the status of a phase when exiting its end state.
// Marks the phase as "failed" if configured in the transition, otherwise "completed".
func (ptc *ProjectTypeConfig) updateExitingPhaseStatus(
	oldState stateMachine.State,
	transitionConfig *TransitionConfig,
	project *state.Project,
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
		if err := state.MarkPhaseFailed(project, phaseName); err != nil {
			return fmt.Errorf("failed to mark phase %s as failed: %w", phaseName, err)
		}
		return nil
	}

	// Normal case: mark as completed
	if err := state.MarkPhaseCompleted(project, phaseName); err != nil {
		return fmt.Errorf("failed to mark phase %s completed: %w", phaseName, err)
	}
	return nil
}

// updateEnteringPhaseStatus updates the status of a phase when entering its start state.
// Marks the phase as "in_progress" if currently "pending".
func (ptc *ProjectTypeConfig) updateEnteringPhaseStatus(
	newState stateMachine.State,
	project *state.Project,
) error {
	phaseName := ptc.GetPhaseForState(newState)
	if phaseName == "" {
		return nil
	}

	if !ptc.IsPhaseStartState(phaseName, newState) {
		return nil
	}

	// MarkPhaseInProgress only updates if status is "pending"
	if err := state.MarkPhaseInProgress(project, phaseName); err != nil {
		return fmt.Errorf("failed to mark phase %s in_progress: %w", phaseName, err)
	}
	return nil
}
