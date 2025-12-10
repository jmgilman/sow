package exploration

import (
	"github.com/jmgilman/sow/libs/project"
	"github.com/jmgilman/sow/libs/project/state"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
)

// init registers the exploration project type on package load.
func init() {
	state.Register("exploration", NewExplorationProjectConfig())
}

// NewExplorationProjectConfig creates the complete configuration for exploration project type.
func NewExplorationProjectConfig() *project.ProjectTypeConfig {
	builder := project.NewProjectTypeConfigBuilder("exploration")
	builder = configurePhases(builder)
	builder = configureTransitions(builder)
	builder = configureEventDeterminers(builder)
	builder = configurePrompts(builder)
	builder = builder.WithInitializer(initializeExplorationProject)
	return builder.Build()
}

// initializeExplorationProject creates all phases for a new exploration project.
// This is called during project creation to set up the phase structure.
//
// The exploration project type has two phases:
// - exploration: Starts immediately in "active" status with enabled=true
// - finalization: Starts in "pending" status with enabled=false
//
// Parameters:
//   - p: The project being initialized
//   - initialInputs: Optional map of phase name to initial input artifacts (can be nil)
func initializeExplorationProject(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
	now := p.Created_at

	// Create exploration phase (starts active)
	explorationInputs := []projschema.ArtifactState{}
	if initialInputs != nil {
		if phaseInputs, exists := initialInputs["exploration"]; exists {
			explorationInputs = phaseInputs
		}
	}

	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: now,
		Inputs:     explorationInputs,
		Outputs:    []projschema.ArtifactState{},
		Tasks:      []projschema.TaskState{},
		Metadata:   make(map[string]interface{}),
	}

	// Create finalization phase (starts pending)
	finalizationInputs := []projschema.ArtifactState{}
	if initialInputs != nil {
		if phaseInputs, exists := initialInputs["finalization"]; exists {
			finalizationInputs = phaseInputs
		}
	}

	p.Phases["finalization"] = projschema.PhaseState{
		Status:     "pending",
		Enabled:    false,
		Created_at: now,
		Inputs:     finalizationInputs,
		Outputs:    []projschema.ArtifactState{},
		Tasks:      []projschema.TaskState{},
		Metadata:   make(map[string]interface{}),
	}

	return nil
}

// configurePhases adds phase definitions to the builder.
//
// The exploration project type has two phases:
// 1. exploration - Contains research tasks, produces summary and findings.
// 2. finalization - Single-state phase for PR creation, no tasks.
func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("exploration",
			project.WithStartState(project.State(Active)),
			project.WithEndState(project.State(Summarizing)),
			project.WithOutputs("summary", "findings"),
			project.WithTasks(),
			project.WithMetadataSchema(explorationMetadataSchema),
		).
		WithPhase("finalization",
			project.WithStartState(project.State(Finalizing)),
			project.WithEndState(project.State(Finalizing)),
			project.WithOutputs("pr"),
			project.WithMetadataSchema(finalizationMetadataSchema),
		)
}

// configureTransitions adds state machine transitions to the builder.
// Configures all 3 transitions for the exploration project type:
// - Active → Summarizing (when all tasks resolved).
// - Summarizing → Finalizing (when summaries approved, crosses phase boundary).
// - Finalizing → Completed (when finalization tasks complete).
func configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		// Set initial state to Active
		SetInitialState(project.State(Active)).

		// Transition 1: Active → Summarizing (intra-phase transition)
		AddTransition(
			project.State(Active),
			project.State(Summarizing),
			project.Event(EventBeginSummarizing),
			project.WithProjectGuard("all tasks resolved", func(p *state.Project) bool {
				return allTasksResolved(p)
			}),
			project.WithProjectOnEntry(func(p *state.Project) error {
				// Update exploration phase status to "summarizing"
				phase := p.Phases["exploration"]
				phase.Status = "summarizing"
				p.Phases["exploration"] = phase
				return nil
			}),
		).

		// Transition 2: Summarizing → Finalizing (inter-phase transition)
		AddTransition(
			project.State(Summarizing),
			project.State(Finalizing),
			project.Event(EventCompleteSummarizing),
			project.WithProjectGuard("all summaries approved", func(p *state.Project) bool {
				return allSummariesApproved(p)
			}),
			project.WithProjectOnEntry(func(p *state.Project) error {
				// Enable finalization phase
				// Note: Phase status and timestamps are automatically managed by FireWithPhaseUpdates
				phase := p.Phases["finalization"]
				phase.Enabled = true
				p.Phases["finalization"] = phase
				return nil
			}),
		).

		// Transition 3: Finalizing → Completed (terminal transition)
		AddTransition(
			project.State(Finalizing),
			project.State(Completed),
			project.Event(EventCompleteFinalization),
			project.WithProjectGuard("all finalization tasks complete", func(p *state.Project) bool {
				return allFinalizationTasksComplete(p)
			}),
			// Note: Finalization phase completion is automatically managed by FireWithPhaseUpdates
		)
}

// configureEventDeterminers adds event determination logic to the builder.
// Maps each advanceable state to its corresponding advance event.
// For exploration project type, the mapping is straightforward:
// - Active → EventBeginSummarizing.
// - Summarizing → EventCompleteSummarizing.
// - Finalizing → EventCompleteFinalization.
func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(project.State(Active), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventBeginSummarizing), nil
		}).
		OnAdvance(project.State(Summarizing), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventCompleteSummarizing), nil
		}).
		OnAdvance(project.State(Finalizing), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventCompleteFinalization), nil
		})
}
