package state

// PromptFunc is an optional callback that generates contextual prompts for state entry.
// It receives the current state and returns a prompt string to display to the user.
//
// Prompts are pure string transformations - they should not perform I/O operations
// or return errors. If prompt generation requires complex logic or external state,
// that logic should be encapsulated within the function via closures.
//
// Usage:
//
//	promptFunc := func(state State) string {
//	    switch state {
//	    case PlanningActive:
//	        return "Planning phase: Create and approve task list"
//	    case ImplementationActive:
//	        return "Implementation phase: Execute tasks"
//	    default:
//	        return ""
//	    }
//	}
//
//	builder := NewBuilder(initialState, projectState, promptFunc)
//
// Passing nil for the prompt function is allowed and will skip prompt generation:
//
//	builder := NewBuilder(initialState, projectState, nil)
type PromptFunc func(State) string
