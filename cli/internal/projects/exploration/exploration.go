package exploration

import (
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
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
			project.WithStartState(state.State(Active)),
			project.WithEndState(state.State(Summarizing)),
			project.WithOutputs("summary", "findings"),
			project.WithTasks(),
			project.WithMetadataSchema(explorationMetadataSchema),
		).
		WithPhase("finalization",
			project.WithStartState(state.State(Finalizing)),
			project.WithEndState(state.State(Finalizing)),
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
		SetInitialState(sdkstate.State(Active)).

		// Transition 1: Active → Summarizing (intra-phase transition)
		AddTransition(
			sdkstate.State(Active),
			sdkstate.State(Summarizing),
			sdkstate.Event(EventBeginSummarizing),
			project.WithGuard("all tasks resolved", func(p *state.Project) bool {
				return allTasksResolved(p)
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Update exploration phase status to "summarizing"
				phase := p.Phases["exploration"]
				phase.Status = "summarizing"
				p.Phases["exploration"] = phase
				return nil
			}),
		).

		// Transition 2: Summarizing → Finalizing (inter-phase transition)
		AddTransition(
			sdkstate.State(Summarizing),
			sdkstate.State(Finalizing),
			sdkstate.Event(EventCompleteSummarizing),
			project.WithGuard("all summaries approved", func(p *state.Project) bool {
				return allSummariesApproved(p)
			}),
			project.WithOnExit(func(p *state.Project) error {
				// Mark exploration phase as completed
				phase := p.Phases["exploration"]
				phase.Status = "completed"
				phase.Completed_at = time.Now()
				p.Phases["exploration"] = phase
				return nil
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Enable and activate finalization phase
				phase := p.Phases["finalization"]
				phase.Enabled = true
				phase.Status = "in_progress"
				phase.Started_at = time.Now()
				p.Phases["finalization"] = phase
				return nil
			}),
		).

		// Transition 3: Finalizing → Completed (terminal transition)
		AddTransition(
			sdkstate.State(Finalizing),
			sdkstate.State(Completed),
			sdkstate.Event(EventCompleteFinalization),
			project.WithGuard("all finalization tasks complete", func(p *state.Project) bool {
				return allFinalizationTasksComplete(p)
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Mark finalization phase as completed
				phase := p.Phases["finalization"]
				phase.Status = "completed"
				phase.Completed_at = time.Now()
				p.Phases["finalization"] = phase
				return nil
			}),
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
		OnAdvance(sdkstate.State(Active), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventBeginSummarizing), nil
		}).
		OnAdvance(sdkstate.State(Summarizing), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventCompleteSummarizing), nil
		}).
		OnAdvance(sdkstate.State(Finalizing), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventCompleteFinalization), nil
		})
}
