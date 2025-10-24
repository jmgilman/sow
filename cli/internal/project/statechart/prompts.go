package statechart

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/schemas"
)

// statePrompts maps state machine states to their corresponding prompt IDs.
var statePrompts = map[State]prompts.PromptID{
	NoProject:               prompts.PromptNoProject,
	PlanningActive:          prompts.PromptPlanningActive,
	ImplementationPlanning:  prompts.PromptImplementationPlanning,
	ImplementationExecuting: prompts.PromptImplementationExecuting,
	ReviewActive:            prompts.PromptReviewActive,
	FinalizeDocumentation:   prompts.PromptFinalizeDocumentation,
	FinalizeChecks:          prompts.PromptFinalizeChecks,
	FinalizeDelete:          prompts.PromptFinalizeDelete,
}

// PromptContext contains all information needed to generate contextual prompts.
type PromptContext struct {
	State        State
	ProjectState *schemas.ProjectState
}

// GeneratePrompt generates a contextual prompt for the current state using templates.
func GeneratePrompt(ctx PromptContext) string {
	promptID, ok := statePrompts[ctx.State]
	if !ok {
		return fmt.Sprintf("Unknown state: %s", ctx.State)
	}

	// Convert to prompts.StatechartContext
	promptCtx := &prompts.StatechartContext{
		State:        string(ctx.State), // Convert State to string
		ProjectState: ctx.ProjectState,
	}

	// Render using central prompts package
	output, err := prompts.Render(promptID, promptCtx)
	if err != nil {
		return fmt.Sprintf("Error rendering prompt for state %s: %v", ctx.State, err)
	}

	return output
}
