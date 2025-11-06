package design

import (
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// init registers the design project type on package load.
func init() {
	state.Register("design", NewDesignProjectConfig())
}

// NewDesignProjectConfig creates the complete configuration for design project type.
// Uses builder pattern to assemble phases, transitions, event determiners, and prompts.
// Returns a fully configured ProjectTypeConfig ready for use.
func NewDesignProjectConfig() *project.ProjectTypeConfig {
	builder := project.NewProjectTypeConfigBuilder("design")
	builder = configurePhases(builder)
	builder = configureTransitions(builder)
	builder = configureEventDeterminers(builder)
	builder = configurePrompts(builder)
	builder = builder.WithInitializer(initializeDesignProject)
	return builder.Build()
}

// initializeDesignProject creates all phases for a new design project.
// This is called during project creation to set up the phase structure.
//
// The design project type has two phases:
// - design: Starts immediately in "active" status with enabled=true
// - finalization: Starts in "pending" status with enabled=false
//
// Parameters:
//   - p: The project being initialized
//   - initialInputs: Optional map of phase name to initial input artifacts (can be nil)
func initializeDesignProject(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
	now := p.Created_at

	// Create design phase (starts active)
	designInputs := []projschema.ArtifactState{}
	if initialInputs != nil {
		if phaseInputs, exists := initialInputs["design"]; exists {
			designInputs = phaseInputs
		}
	}

	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: now,
		Inputs:     designInputs,
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
// The design project type has two phases:
// 1. design - Contains document planning tasks, produces design artifacts.
// 2. finalization - Contains tasks for moving documents, creating PR, and cleanup.
//
// Returns the builder to enable method chaining.
func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("design",
			project.WithStartState(sdkstate.State(Active)),
			project.WithEndState(sdkstate.State(Active)),
			project.WithOutputs("design", "adr", "architecture", "diagram", "spec"),
			project.WithTasks(),
			project.WithMetadataSchema(designMetadataSchema),
		).
		WithPhase("finalization",
			project.WithStartState(sdkstate.State(Finalizing)),
			project.WithEndState(sdkstate.State(Finalizing)),
			project.WithOutputs("pr"),
			project.WithTasks(),
			project.WithMetadataSchema(finalizationMetadataSchema),
		)
}

// configureTransitions adds state machine transitions to the builder.
// Configures both transitions for the design project type:
// - Active → Finalizing (when all documents approved, crosses phase boundary).
// - Finalizing → Completed (when finalization tasks complete).
//
// Returns the builder to enable method chaining.
func configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		// Set initial state to Active
		SetInitialState(sdkstate.State(Active)).

		// Transition 1: Active → Finalizing (inter-phase transition)
		AddTransition(
			sdkstate.State(Active),
			sdkstate.State(Finalizing),
			sdkstate.Event(EventCompleteDesign),
			project.WithGuard("all documents approved", func(p *state.Project) bool {
				return allDocumentsApproved(p)
			}),
			project.WithOnExit(func(p *state.Project) error {
				// Mark design phase as completed
				phase := p.Phases["design"]
				phase.Status = "completed"
				phase.Completed_at = time.Now()
				p.Phases["design"] = phase
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

		// Transition 2: Finalizing → Completed (terminal transition)
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
// For design project type, the mapping is:
// - Active → EventCompleteDesign.
// - Finalizing → EventCompleteFinalization.
//
// Returns the builder to enable method chaining.
func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(sdkstate.State(Active), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventCompleteDesign), nil
		}).
		OnAdvance(sdkstate.State(Finalizing), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventCompleteFinalization), nil
		})
}
