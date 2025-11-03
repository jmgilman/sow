package project

import (
	"context"
	"unsafe"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	stateMachine "github.com/jmgilman/sow/cli/internal/sdks/state"
	"github.com/jmgilman/sow/cli/schemas"
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
	// Create prompt function that uses project SDK prompts
	// This closure captures the project wrapper and looks up prompts from the prompts map
	var promptFunc stateMachine.PromptFunc
	if len(ptc.prompts) > 0 {
		promptFunc = func(state stateMachine.State) string {
			gen := ptc.prompts[state]
			if gen == nil {
				return "" // No prompt configured for this state
			}
			return gen(project) // Call project SDK prompt with Project wrapper
		}
	}

	// The embedded ProjectState can be cast to *schemas.ProjectState since
	// schemas.ProjectState is a type alias for project.ProjectState
	// We use unsafe.Pointer as an intermediate to convert between the types
	projectStatePtr := unsafe.Pointer(&project.ProjectState)
	schemasProjectState := (*schemas.ProjectState)(projectStatePtr)

	builder := stateMachine.NewBuilder(initialState, schemasProjectState, promptFunc)

	// Add all transitions with guards and actions bound to project instance
	for _, tc := range ptc.transitions {
		var opts []stateMachine.TransitionOption

		// Bind guard template to project instance via closure
		if tc.guardTemplate != nil {
			// Closure captures project - guard can access live state
			guardTemplate := tc.guardTemplate // Capture for closure
			boundGuard := func() bool {
				return guardTemplate(project)
			}
			opts = append(opts, stateMachine.WithGuard(boundGuard))
		}

		// Bind onExit action to project instance via closure
		if tc.onExit != nil {
			// Closure captures project and action
			action := tc.onExit // Capture for closure
			boundOnExit := func(_ context.Context, _ ...any) error {
				return action(project)
			}
			opts = append(opts, stateMachine.WithOnExit(boundOnExit))
		}

		// Bind onEntry action to project instance via closure
		if tc.onEntry != nil {
			// Closure captures project and action
			action := tc.onEntry // Capture for closure
			boundOnEntry := func(_ context.Context, _ ...any) error {
				return action(project)
			}
			opts = append(opts, stateMachine.WithOnEntry(boundOnEntry))
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
