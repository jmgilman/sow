package statechart

import "github.com/jmgilman/sow/cli/schemas"

// PromptGenerator defines the contract for project-owned prompt generation.
// Each project type implements this interface to generate contextual prompts
// for their specific state machine states.
//
// Prompt generators have access to:
//   - Full project state (for context like task lists, artifacts, etc.)
//   - External systems via sow.Context (git, GitHub, filesystem)
//   - PromptComponents for reusable prompt sections
//   - Template rendering via prompts package
//
// Example implementation:
//
//	type StandardPromptGenerator struct {
//	    components *statechart.PromptComponents
//	}
//
//	func (g *StandardPromptGenerator) GeneratePrompt(
//	    state statechart.State,
//	    projectState *schemas.ProjectState,
//	) (string, error) {
//	    switch state {
//	    case statechart.PlanningActive:
//	        return g.generatePlanningPrompt(projectState)
//	    case statechart.ImplementationExecuting:
//	        return g.generateImplementationPrompt(projectState)
//	    default:
//	        return "", fmt.Errorf("unknown state: %s", state)
//	    }
//	}
type PromptGenerator interface {
	// GeneratePrompt generates a contextual prompt for the given state.
	// The prompt should provide guidance appropriate for the current state
	// and include relevant context from the project state.
	//
	// This method is called automatically on state entry by the state machine.
	// Errors are propagated to the caller, preventing the state transition.
	//
	// Parameters:
	//   - state: The state machine state to generate a prompt for
	//   - projectState: The current project state with full context
	//
	// Returns:
	//   - The rendered prompt string ready to display to the user
	//   - An error if prompt generation fails (e.g., template error, external system failure)
	GeneratePrompt(state State, projectState *schemas.ProjectState) (string, error)
}
