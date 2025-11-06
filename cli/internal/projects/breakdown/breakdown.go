// Package breakdown implements the breakdown project type for decomposing
// large features or design documents into implementable work units.
//
// The breakdown project type enables users to:
//   - Decompose complex features into discrete, implementable tasks
//   - Specify requirements and acceptance criteria for each work unit
//   - Review and approve work units before publication
//   - Publish approved work units as GitHub issues with proper dependency tracking
//
// Workflow States:
//   - Active: Decomposing, specifying, and reviewing work units
//   - Publishing: Creating GitHub issues from approved work units
//   - Completed: All work units successfully published
package breakdown

import (
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// init registers the breakdown project type on package load.
func init() {
	state.Register("breakdown", NewBreakdownProjectConfig())
}

// NewBreakdownProjectConfig creates the complete configuration for breakdown project type.
// Uses builder pattern to assemble phases, transitions, event determiners, and prompts.
// Returns a fully configured ProjectTypeConfig ready for use.
func NewBreakdownProjectConfig() *project.ProjectTypeConfig {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	builder = configurePhases(builder)
	builder = configureTransitions(builder)
	builder = configureEventDeterminers(builder)
	builder = configurePrompts(builder)
	builder = builder.WithInitializer(initializeBreakdownProject)
	return builder.Build()
}

// configurePhases adds phase definitions to the builder.
//
// The breakdown project type has a single phase:
// 1. breakdown - Contains decomposition, specification, and review tasks, produces work_unit_spec artifacts.
//
// Unlike design/exploration projects, there is NO finalization phase.
// The breakdown phase completes after all work units are published as GitHub issues.
//
// Returns the builder to enable method chaining.
func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("breakdown",
			project.WithStartState(sdkstate.State(Active)),
			project.WithEndState(sdkstate.State(Publishing)),
			project.WithOutputs("work_unit_spec"),
			project.WithTasks(),
			project.WithMetadataSchema(breakdownMetadataSchema),
		)
}

// configureTransitions adds state machine transitions to the builder.
//
// Configures two transitions:
// 1. Active → Publishing - When all work units approved and dependencies valid
// 2. Publishing → Completed - When all work units published to GitHub
//
// Returns the builder to enable method chaining.
func configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		// Set initial state to Active
		SetInitialState(sdkstate.State(Active)).

		// Transition 1: Active → Publishing (intra-phase transition)
		AddTransition(
			sdkstate.State(Active),
			sdkstate.State(Publishing),
			sdkstate.Event(EventBeginPublishing),
			project.WithGuard("all work units approved and dependencies valid", func(p *state.Project) bool {
				return allWorkUnitsApproved(p) && dependenciesValid(p)
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Update breakdown phase status to "publishing"
				phase := p.Phases["breakdown"]
				phase.Status = "publishing"
				p.Phases["breakdown"] = phase
				return nil
			}),
		).

		// Transition 2: Publishing → Completed (terminal transition)
		AddTransition(
			sdkstate.State(Publishing),
			sdkstate.State(Completed),
			sdkstate.Event(EventCompleteBreakdown),
			project.WithGuard("all work units published", func(p *state.Project) bool {
				return allWorkUnitsPublished(p)
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Mark breakdown phase as completed
				phase := p.Phases["breakdown"]
				phase.Status = "completed"
				phase.Completed_at = time.Now()
				p.Phases["breakdown"] = phase
				return nil
			}),
		)
}

// configureEventDeterminers adds event determination logic to the builder.
//
// Maps each advanceable state to its corresponding advance event:
// - Active → EventBeginPublishing
// - Publishing → EventCompleteBreakdown
//
// Returns the builder to enable method chaining.
func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(sdkstate.State(Active), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventBeginPublishing), nil
		}).
		OnAdvance(sdkstate.State(Publishing), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventCompleteBreakdown), nil
		})
}

// configurePrompts adds prompt generators to the builder.
// Registers orchestrator, active, and publishing prompt generators.
// Returns the builder to enable method chaining.
func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithOrchestratorPrompt(generateOrchestratorPrompt).
		WithPrompt(sdkstate.State(Active), generateActivePrompt).
		WithPrompt(sdkstate.State(Publishing), generatePublishingPrompt)
}

// initializeBreakdownProject creates the breakdown phase for a new breakdown project.
// Unlike design/exploration projects, breakdown has only ONE phase (no finalization phase).
//
// The breakdown project type has a single phase:
// - breakdown: Starts immediately in "active" status with enabled=true
//
// Parameters:
//   - p: The project being initialized
//   - initialInputs: Optional map of phase name to initial input artifacts (can be nil)
func initializeBreakdownProject(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
	now := p.Created_at

	// Get initial inputs for breakdown phase
	inputs := []projschema.ArtifactState{}
	if initialInputs != nil {
		if phaseInputs, exists := initialInputs["breakdown"]; exists {
			inputs = phaseInputs
		}
	}

	// Create breakdown phase (starts active immediately)
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: now,
		Inputs:     inputs,
		Outputs:    []projschema.ArtifactState{},
		Tasks:      []projschema.TaskState{},
		Metadata:   make(map[string]interface{}),
	}

	return nil
}
