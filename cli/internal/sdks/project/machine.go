package project

import (
	"context"

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

	// NOTE: We pass nil for projectState because the Project SDK uses closures
	// to bind project state to guards and actions. The state machine builder's
	// projectState parameter is optional and only needed for legacy code that
	// doesn't use closures. Since we bind state via closures, we pass nil here.
	builder := stateMachine.NewBuilder(initialState, nil, promptFunc)

	// Add all transitions with guards and actions bound to project instance
	for _, tc := range ptc.transitions {
		var opts []stateMachine.TransitionOption

		// Bind guard template to project instance via closure
		if tc.guardTemplate != nil {
			opts = append(opts, stateMachine.WithGuard(func() bool {
				return tc.guardTemplate(project)
			}))
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
