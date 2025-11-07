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
//
// Parameters:
//   - p: The project being initialized
//   - initialInputs: Optional map of phase name to initial input artifacts (can be nil)
func initializeStandardProject(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
	now := p.Created_at
	phaseNames := []string{"implementation", "review", "finalize"}

	for _, phaseName := range phaseNames {
		// Get initial inputs for this phase (empty slice if none provided)
		inputs := []projschema.ArtifactState{}
		if initialInputs != nil {
			if phaseInputs, exists := initialInputs[phaseName]; exists {
				inputs = phaseInputs
			}
		}

		p.Phases[phaseName] = projschema.PhaseState{
			Status:     "pending",
			Enabled:    false,
			Created_at: now,
			Inputs:     inputs, // Use provided initial inputs
			Outputs:    []projschema.ArtifactState{},
			Tasks:      []projschema.TaskState{},
			Metadata:   make(map[string]interface{}),
		}
	}

	return nil
}

func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("implementation",
			project.WithStartState(sdkstate.State(ImplementationPlanning)),
			project.WithEndState(sdkstate.State(ImplementationExecuting)),
			project.WithOutputs("task_list"),
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

		SetInitialState(sdkstate.State(ImplementationPlanning)).

		// Project initialization
		AddTransition(
			sdkstate.State(NoProject),
			sdkstate.State(ImplementationPlanning),
			sdkstate.Event(EventProjectInit),
			project.WithDescription("Initialize project and begin implementation planning"),
		).

		// Implementation planning → draft PR creation
		AddTransition(
			sdkstate.State(ImplementationPlanning),
			sdkstate.State(ImplementationDraftPRCreation),
			sdkstate.Event(EventPlanningComplete),
			project.WithDescription("Task descriptions approved, create draft PR"),
			project.WithGuard("task descriptions approved", func(p *state.Project) bool {
				return allTaskDescriptionsApproved(p)
			}),
		).

		// Draft PR creation → execution
		AddTransition(
			sdkstate.State(ImplementationDraftPRCreation),
			sdkstate.State(ImplementationExecuting),
			sdkstate.Event(EventDraftPRCreated),
			project.WithDescription("Draft PR created, begin task execution"),
			project.WithGuard("draft PR created", func(p *state.Project) bool {
				return draftPRCreated(p)
			}),
		).

		// Implementation → Review
		AddTransition(
			sdkstate.State(ImplementationExecuting),
			sdkstate.State(ReviewActive),
			sdkstate.Event(EventAllTasksComplete),
			project.WithDescription("All implementation tasks completed, ready for review"),
			project.WithGuard("all tasks complete", func(p *state.Project) bool {
				return allTasksComplete(p)
			}),
		).

		// Review → Finalize/Implementation (branching on review assessment)
		// Uses AddBranch to declaratively define pass/fail paths based on review assessment
		AddBranch(
			sdkstate.State(ReviewActive),
			project.BranchOn(getReviewAssessment),
			project.When("pass",
				sdkstate.Event(EventReviewPass),
				sdkstate.State(FinalizeChecks),
				project.WithDescription("Review approved, proceed to finalization checks"),
			),
			project.When("fail",
				sdkstate.Event(EventReviewFail),
				sdkstate.State(ImplementationPlanning),
				project.WithDescription("Review failed, return to implementation planning for rework"),
				project.WithFailedPhase("review"), // Mark review as failed instead of completed
				project.WithOnEntry(func(p *state.Project) error {
					// Only execute rework logic if review phase exists and has failed
					// (this prevents executing on NoProject → ImplementationPlanning transition)
					reviewPhase, hasReview := p.Phases["review"]
					if !hasReview || len(reviewPhase.Outputs) == 0 {
						// First time entering implementation - no rework needed
						return nil
					}

					// Increment implementation iteration for rework
					if err := state.IncrementPhaseIteration(p, "implementation"); err != nil {
						return fmt.Errorf("failed to increment implementation iteration: %w", err)
					}

					// Add failed review as implementation input
					// Note: Implementation phase status will be automatically set to "in_progress"
					// by FireWithPhaseUpdates when entering ImplementationPlanning state
					return state.AddPhaseInputFromOutput(
						p,
						"review",
						"implementation",
						"review",
						func(a *projschema.ArtifactState) bool {
							assessment, ok := a.Metadata["assessment"].(string)
							return ok && assessment == "fail" && a.Approved
						},
					)
				}),
			),
		).

		// Finalize substates
		AddTransition(
			sdkstate.State(FinalizeChecks),
			sdkstate.State(FinalizePRReady),
			sdkstate.Event(EventChecksDone),
			project.WithDescription("Checks completed, prepare PR for final review"),
		).
		AddTransition(
			sdkstate.State(FinalizePRReady),
			sdkstate.State(FinalizePRChecks),
			sdkstate.Event(EventPRReady),
			project.WithDescription("PR body approved, monitoring PR checks"),
			project.WithGuard("PR body approved", func(p *state.Project) bool {
				return prBodyApproved(p)
			}),
		).
		AddTransition(
			sdkstate.State(FinalizePRChecks),
			sdkstate.State(FinalizeCleanup),
			sdkstate.Event(EventPRChecksPass),
			project.WithDescription("All PR checks passed, begin cleanup"),
			project.WithGuard("all PR checks passed", func(p *state.Project) bool {
				return prChecksPassed(p)
			}),
		).
		AddTransition(
			sdkstate.State(FinalizeCleanup),
			sdkstate.State(NoProject),
			sdkstate.Event(EventCleanupComplete),
			project.WithDescription("Cleanup complete, project finalized"),
			project.WithGuard("project deleted", func(p *state.Project) bool {
				return projectDeleted(p)
			}),
		)
}

func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(sdkstate.State(ImplementationPlanning), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventPlanningComplete), nil
		}).
		OnAdvance(sdkstate.State(ImplementationDraftPRCreation), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventDraftPRCreated), nil
		}).
		OnAdvance(sdkstate.State(ImplementationExecuting), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventAllTasksComplete), nil
		}).
		// NOTE: ReviewActive OnAdvance is auto-generated by AddBranch (see configureTransitions)
		// The discriminator function getReviewAssessment determines pass/fail branching
		OnAdvance(sdkstate.State(FinalizeChecks), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventChecksDone), nil
		}).
		OnAdvance(sdkstate.State(FinalizePRReady), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventPRReady), nil
		}).
		OnAdvance(sdkstate.State(FinalizePRChecks), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventPRChecksPass), nil
		}).
		OnAdvance(sdkstate.State(FinalizeCleanup), func(_ *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventCleanupComplete), nil
		})
}

func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		// Orchestrator-level prompt (how standard projects work)
		WithOrchestratorPrompt(generateOrchestratorPrompt).

		// State-level prompts (what to do in each state)
		WithPrompt(sdkstate.State(ImplementationPlanning), generateImplementationPlanningPrompt).
		WithPrompt(sdkstate.State(ImplementationDraftPRCreation), generateImplementationDraftPRCreationPrompt).
		WithPrompt(sdkstate.State(ImplementationExecuting), generateImplementationExecutingPrompt).
		WithPrompt(sdkstate.State(ReviewActive), generateReviewPrompt).
		WithPrompt(sdkstate.State(FinalizeChecks), generateFinalizeChecksPrompt).
		WithPrompt(sdkstate.State(FinalizePRReady), generateFinalizePRReadyPrompt).
		WithPrompt(sdkstate.State(FinalizePRChecks), generateFinalizePRChecksPrompt).
		WithPrompt(sdkstate.State(FinalizeCleanup), generateFinalizeCleanupPrompt)
}
