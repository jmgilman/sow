package standard

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// init registers the standard project type on package load.
func init() {
	state.Register("standard", NewStandardProjectConfig())
}

// NewStandardProjectConfig creates the complete configuration for standard project type.
func NewStandardProjectConfig() *project.ProjectTypeConfig {
	builder := project.NewProjectTypeConfigBuilder("standard")
	builder = configurePhases(builder)
	builder = configureTransitions(builder)
	builder = configureEventDeterminers(builder)
	builder = configurePrompts(builder)
	builder = builder.WithInitializer(initializeStandardProject)
	return builder.Build()
}

// initializeStandardProject creates all phases for a new standard project.
// This is called during project creation to set up the phase structure.
func initializeStandardProject(p *state.Project) error {
	now := p.Created_at
	phaseNames := []string{"planning", "implementation", "review", "finalize"}

	for _, phaseName := range phaseNames {
		p.Phases[phaseName] = projschema.PhaseState{
			Status:     "pending",
			Enabled:    false,
			Created_at: now,
			Inputs:     []projschema.ArtifactState{},
			Outputs:    []projschema.ArtifactState{},
			Tasks:      []projschema.TaskState{},
			Metadata:   make(map[string]interface{}),
		}
	}

	return nil
}

func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("planning",
			project.WithStartState(sdkstate.State(PlanningActive)),
			project.WithEndState(sdkstate.State(PlanningActive)),
			project.WithInputs("context"),
			project.WithOutputs("task_list"),
		).
		WithPhase("implementation",
			project.WithStartState(sdkstate.State(ImplementationPlanning)),
			project.WithEndState(sdkstate.State(ImplementationExecuting)),
			project.WithTasks(),
			project.WithMetadataSchema(implementationMetadataSchema),
		).
		WithPhase("review",
			project.WithStartState(sdkstate.State(ReviewActive)),
			project.WithEndState(sdkstate.State(ReviewActive)),
			project.WithOutputs("review"),
			project.WithMetadataSchema(reviewMetadataSchema),
		).
		WithPhase("finalize",
			project.WithStartState(sdkstate.State(FinalizeChecks)),
			project.WithEndState(sdkstate.State(FinalizeCleanup)),
			project.WithOutputs("pr_body"),
			project.WithMetadataSchema(finalizeMetadataSchema),
		)
}

func configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.

		// ===== STATE MACHINE =====

		SetInitialState(sdkstate.State(PlanningActive)).

		// Project initialization
		AddTransition(
			sdkstate.State(NoProject),
			sdkstate.State(PlanningActive),
			sdkstate.Event(EventProjectInit),
		).

		// Planning → Implementation
		AddTransition(
			sdkstate.State(PlanningActive),
			sdkstate.State(ImplementationPlanning),
			sdkstate.Event(EventCompletePlanning),
			project.WithGuard(func(p *state.Project) bool {
				return phaseOutputApproved(p, "planning", "task_list")
			}),
		).

		// Implementation planning → execution
		AddTransition(
			sdkstate.State(ImplementationPlanning),
			sdkstate.State(ImplementationExecuting),
			sdkstate.Event(EventTasksApproved),
			project.WithGuard(func(p *state.Project) bool {
				return phaseMetadataBool(p, "implementation", "tasks_approved")
			}),
		).

		// Implementation → Review
		AddTransition(
			sdkstate.State(ImplementationExecuting),
			sdkstate.State(ReviewActive),
			sdkstate.Event(EventAllTasksComplete),
			project.WithGuard(func(p *state.Project) bool {
				return allTasksComplete(p)
			}),
		).

		// Review → Finalize (pass)
		AddTransition(
			sdkstate.State(ReviewActive),
			sdkstate.State(FinalizeChecks),
			sdkstate.Event(EventReviewPass),
			project.WithGuard(func(p *state.Project) bool {
				return latestReviewApproved(p)
			}),
		).

		// Review → Implementation (fail - rework)
		AddTransition(
			sdkstate.State(ReviewActive),
			sdkstate.State(ImplementationPlanning),
			sdkstate.Event(EventReviewFail),
			project.WithGuard(func(p *state.Project) bool {
				return latestReviewApproved(p)
			}),
		).

		// Finalize substates
		AddTransition(
			sdkstate.State(FinalizeChecks),
			sdkstate.State(FinalizePRCreation),
			sdkstate.Event(EventChecksDone),
		).
		AddTransition(
			sdkstate.State(FinalizePRCreation),
			sdkstate.State(FinalizeCleanup),
			sdkstate.Event(EventPRCreated),
			project.WithGuard(func(p *state.Project) bool {
				return prBodyApproved(p)
			}),
		).
		AddTransition(
			sdkstate.State(FinalizeCleanup),
			sdkstate.State(NoProject),
			sdkstate.Event(EventCleanupComplete),
			project.WithGuard(func(p *state.Project) bool {
				return projectDeleted(p)
			}),
		)
}

func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(sdkstate.State(PlanningActive), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventCompletePlanning), nil
		}).
		OnAdvance(sdkstate.State(ImplementationPlanning), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventTasksApproved), nil
		}).
		OnAdvance(sdkstate.State(ImplementationExecuting), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventAllTasksComplete), nil
		}).
		OnAdvance(sdkstate.State(ReviewActive), func(p *state.Project) (sdkstate.Event, error) {
			// Complex: examine review assessment
			phase, exists := p.Phases["review"]
			if !exists {
				return "", fmt.Errorf("review phase not found")
			}

			// Find latest approved review
			var latestReview *projschema.ArtifactState
			for i := len(phase.Outputs) - 1; i >= 0; i-- {
				if phase.Outputs[i].Type == "review" && phase.Outputs[i].Approved {
					artifact := phase.Outputs[i]
					latestReview = &artifact
					break
				}
			}

			if latestReview == nil {
				return "", fmt.Errorf("no approved review found")
			}

			// Check assessment metadata
			assessment, ok := latestReview.Metadata["assessment"].(string)
			if !ok {
				return "", fmt.Errorf("review missing assessment")
			}

			switch assessment {
			case "pass":
				return sdkstate.Event(EventReviewPass), nil
			case "fail":
				return sdkstate.Event(EventReviewFail), nil
			default:
				return "", fmt.Errorf("invalid assessment: %s", assessment)
			}
		}).
		OnAdvance(sdkstate.State(FinalizeChecks), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventChecksDone), nil
		}).
		OnAdvance(sdkstate.State(FinalizePRCreation), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventPRCreated), nil
		}).
		OnAdvance(sdkstate.State(FinalizeCleanup), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventCleanupComplete), nil
		})
}

func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPrompt(sdkstate.State(PlanningActive), generatePlanningPrompt).
		WithPrompt(sdkstate.State(ImplementationPlanning), generateImplementationPlanningPrompt).
		WithPrompt(sdkstate.State(ImplementationExecuting), generateImplementationExecutingPrompt).
		WithPrompt(sdkstate.State(ReviewActive), generateReviewPrompt).
		WithPrompt(sdkstate.State(FinalizeChecks), generateFinalizeChecksPrompt).
		WithPrompt(sdkstate.State(FinalizePRCreation), generateFinalizePRCreationPrompt).
		WithPrompt(sdkstate.State(FinalizeCleanup), generateFinalizeCleanupPrompt)
}
