// Package breakdown implements the breakdown project type for decomposing
// large features or design documents into implementable work units.
//
// The breakdown project type enables users to:
//   - Gather codebase/design context through discovery
//   - Decompose complex features into discrete, implementable tasks
//   - Specify requirements and acceptance criteria for each work unit
//   - Review and approve work units before publication
//   - Publish approved work units as GitHub issues with proper dependency tracking
//
// Workflow States:
//   - Discovery: Gathering codebase and design context
//   - Active: Decomposing, specifying, and reviewing work units
//   - Publishing: Creating GitHub issues from approved work units
//   - Completed: All work units successfully published
package breakdown

import (
	"github.com/jmgilman/sow/libs/project"
	"github.com/jmgilman/sow/libs/project/state"
	
	projschema "github.com/jmgilman/sow/libs/schemas/project"
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
//  1. breakdown - Contains discovery, decomposition, specification, and review tasks,
//     produces discovery and work_unit_spec artifacts.
//
// Unlike design/exploration projects, there is NO finalization phase.
// The breakdown phase completes after all work units are published as GitHub issues.
//
// Returns the builder to enable method chaining.
func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("breakdown",
			project.WithStartState(project.State(Discovery)),
			project.WithEndState(project.State(Publishing)),
			project.WithOutputs("discovery", "work_unit_spec"),
			project.WithTasks(),
			project.WithMetadataSchema(breakdownMetadataSchema),
		)
}

// configureTransitions adds state machine transitions to the builder.
//
// Configures three transitions:
// 1. Discovery → Active - When discovery document is approved
// 2. Active → Publishing - When all work units approved and dependencies valid
// 3. Publishing → Completed - When all work units published to GitHub
//
// Returns the builder to enable method chaining.
func configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		// Set initial state to Discovery
		SetInitialState(project.State(Discovery)).

		// Transition 1: Discovery → Active (discovery complete)
		AddTransition(
			project.State(Discovery),
			project.State(Active),
			project.Event(EventBeginActive),
			project.WithProjectGuard("discovery document approved", func(p *state.Project) bool {
				return hasApprovedDiscoveryDocument(p)
			}),
			project.WithProjectOnEntry(func(p *state.Project) error {
				// Update breakdown phase status to "active"
				phase := p.Phases["breakdown"]
				phase.Status = "active"
				p.Phases["breakdown"] = phase
				return nil
			}),
		).

		// Transition 2: Active → Publishing (intra-phase transition)
		AddTransition(
			project.State(Active),
			project.State(Publishing),
			project.Event(EventBeginPublishing),
			project.WithProjectGuard("all work units approved and dependencies valid", func(p *state.Project) bool {
				return allWorkUnitsApproved(p) && dependenciesValid(p)
			}),
			project.WithProjectOnEntry(func(p *state.Project) error {
				// Update breakdown phase status to "publishing"
				phase := p.Phases["breakdown"]
				phase.Status = "publishing"
				p.Phases["breakdown"] = phase
				return nil
			}),
		).

		// Transition 3: Publishing → Completed (terminal transition)
		AddTransition(
			project.State(Publishing),
			project.State(Completed),
			project.Event(EventCompleteBreakdown),
			project.WithProjectGuard("all work units published", func(p *state.Project) bool {
				return allWorkUnitsPublished(p)
			}),
			// Note: Breakdown phase completion is automatically managed by FireWithPhaseUpdates
		)
}

// configureEventDeterminers adds event determination logic to the builder.
//
// Maps each advanceable state to its corresponding advance event:
// - Discovery → EventBeginActive
// - Active → EventBeginPublishing
// - Publishing → EventCompleteBreakdown
//
// Returns the builder to enable method chaining.
func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(project.State(Discovery), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventBeginActive), nil
		}).
		OnAdvance(project.State(Active), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventBeginPublishing), nil
		}).
		OnAdvance(project.State(Publishing), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventCompleteBreakdown), nil
		})
}

// configurePrompts adds prompt generators to the builder.
// Registers orchestrator, discovery, active, and publishing prompt generators.
// Returns the builder to enable method chaining.
func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithOrchestratorPrompt(generateOrchestratorPrompt).
		WithPrompt(project.State(Discovery), generateDiscoveryPrompt).
		WithPrompt(project.State(Active), generateActivePrompt).
		WithPrompt(project.State(Publishing), generatePublishingPrompt)
}

// initializeBreakdownProject creates the breakdown phase for a new breakdown project.
// Unlike design/exploration projects, breakdown has only ONE phase (no finalization phase).
//
// The breakdown project type has a single phase:
// - breakdown: Starts in "discovery" status with enabled=true
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

	// Create breakdown phase (starts in_progress, state machine starts in Discovery state)
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "in_progress",
		Enabled:    true,
		Created_at: now,
		Started_at: now,
		Inputs:     inputs,
		Outputs:    []projschema.ArtifactState{},
		Tasks:      []projschema.TaskState{},
		Metadata:   make(map[string]interface{}),
	}

	return nil
}
