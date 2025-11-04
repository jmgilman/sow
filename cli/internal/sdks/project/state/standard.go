package state

import (
	"time"

	"github.com/jmgilman/sow/cli/schemas/project"
)

// Standard project states.
const (
	PlanningActive          = State("PlanningActive")
	ImplementationPlanning  = State("ImplementationPlanning")
	ImplementationExecuting = State("ImplementationExecuting")
	ReviewActive            = State("ReviewActive")
	FinalizeDocumentation   = State("FinalizeDocumentation")
	FinalizeChecks          = State("FinalizeChecks")
	FinalizeDelete          = State("FinalizeDelete")
)

// Standard project events.
const (
	EventProjectInit       = Event("ProjectInit")
	EventCompletePlanning  = Event("CompletePlanning")
	EventTaskCreated       = Event("TaskCreated")
	EventTasksApproved     = Event("TasksApproved")
	EventAllTasksComplete  = Event("AllTasksComplete")
	EventReviewPass        = Event("ReviewPass")
	EventReviewFail        = Event("ReviewFail")
	EventDocumentationDone = Event("DocumentationDone")
	EventChecksDone        = Event("ChecksDone")
	EventProjectDelete     = Event("ProjectDelete")
)

// initStandardProject creates and registers the standard project type configuration.
// This is called during package initialization to register the standard project type.
func initStandardProject() *ProjectTypeConfig {
	// Create config directly to avoid import cycle with project package
	config := &ProjectTypeConfig{
		phaseConfigs: make(map[string]*PhaseConfig),
		transitions:  []TransitionConfig{},
		onAdvance:    make(map[State]EventDeterminer),
	}

	// Set initial state
	config.initialState = PlanningActive

	// Set initializer function
	config.initializer = func(p *Project) error {
		// Initialize all phases with pending status
		now := time.Now()
		phaseNames := []string{"planning", "implementation", "review", "finalize"}

		for _, phaseName := range phaseNames {
			p.Phases[phaseName] = project.PhaseState{
				Status:     "pending",
				Enabled:    false,
				Created_at: now,
				Inputs:     []project.ArtifactState{},
				Outputs:    []project.ArtifactState{},
				Tasks:      []project.TaskState{},
			}
		}

		return nil
	}

	// TODO: Add phase configurations, transitions, guards, actions, and event determiners
	// This will be expanded as the standard project type is fully migrated to the SDK

	return config
}
