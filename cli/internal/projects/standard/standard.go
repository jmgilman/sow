package standard

import (
	"fmt"

	"github.com/jmgilman/sow/libs/project"
	"github.com/jmgilman/sow/libs/project/state"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
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
			project.WithStartState(project.State(ImplementationPlanning)),
			project.WithEndState(project.State(ImplementationExecuting)),
			project.WithOutputs("task_list"),
			project.WithTasks(),
			project.WithMetadataSchema(implementationMetadataSchema),
		).
		WithPhase("review",
			project.WithStartState(project.State(ReviewActive)),
			project.WithEndState(project.State(ReviewActive)),
			project.WithOutputs("review"),
			project.WithMetadataSchema(reviewMetadataSchema),
		).
		WithPhase("finalize",
			project.WithStartState(project.State(FinalizeChecks)),
			project.WithEndState(project.State(FinalizeCleanup)),
			project.WithOutputs("pr_body"),
			project.WithMetadataSchema(finalizeMetadataSchema),
		)
}

func configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.

		// ===== STATE MACHINE =====

		SetInitialState(project.State(ImplementationPlanning)).

		// Project initialization
		AddTransition(
			project.State(NoProject),
			project.State(ImplementationPlanning),
			project.Event(EventProjectInit),
			project.WithProjectDescription("Initialize project and begin implementation planning"),
		).

		// Implementation planning → draft PR creation
		AddTransition(
			project.State(ImplementationPlanning),
			project.State(ImplementationDraftPRCreation),
			project.Event(EventPlanningComplete),
			project.WithProjectDescription("Task descriptions approved, create draft PR"),
			project.WithProjectGuard("task descriptions approved", func(p *state.Project) bool {
				return allTaskDescriptionsApproved(p)
			}),
		).

		// Draft PR creation → execution
		AddTransition(
			project.State(ImplementationDraftPRCreation),
			project.State(ImplementationExecuting),
			project.Event(EventDraftPRCreated),
			project.WithProjectDescription("Draft PR created, begin task execution"),
			project.WithProjectGuard("draft PR created", func(p *state.Project) bool {
				return draftPRCreated(p)
			}),
		).

		// Implementation → Review
		AddTransition(
			project.State(ImplementationExecuting),
			project.State(ReviewActive),
			project.Event(EventAllTasksComplete),
			project.WithProjectDescription("All implementation tasks completed, ready for review"),
			project.WithProjectGuard("all tasks complete", func(p *state.Project) bool {
				return allTasksComplete(p)
			}),
		).

		// Review → Finalize/Implementation (branching on review assessment)
		// Uses AddBranch to declaratively define pass/fail paths based on review assessment
		AddBranch(
			project.State(ReviewActive),
			project.BranchOn(getReviewAssessment),
			project.When("pass",
				project.Event(EventReviewPass),
				project.State(FinalizeChecks),
				project.WithProjectDescription("Review approved, proceed to finalization checks"),
			),
			project.When("fail",
				project.Event(EventReviewFail),
				project.State(ImplementationPlanning),
				project.WithProjectDescription("Review failed, return to implementation planning for rework"),
				project.WithProjectFailedPhase("review"), // Mark review as failed instead of completed
				project.WithProjectOnEntry(func(p *state.Project) error {
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
			project.State(FinalizeChecks),
			project.State(FinalizePRReady),
			project.Event(EventChecksDone),
			project.WithProjectDescription("Checks completed, prepare PR for final review"),
		).
		AddTransition(
			project.State(FinalizePRReady),
			project.State(FinalizePRChecks),
			project.Event(EventPRReady),
			project.WithProjectDescription("PR body approved, monitoring PR checks"),
			project.WithProjectGuard("PR body approved", func(p *state.Project) bool {
				return prBodyApproved(p)
			}),
		).
		AddTransition(
			project.State(FinalizePRChecks),
			project.State(FinalizeCleanup),
			project.Event(EventPRChecksPass),
			project.WithProjectDescription("All PR checks passed, begin cleanup"),
			project.WithProjectGuard("all PR checks passed", func(p *state.Project) bool {
				return prChecksPassed(p)
			}),
		).
		AddTransition(
			project.State(FinalizeCleanup),
			project.State(NoProject),
			project.Event(EventCleanupComplete),
			project.WithProjectDescription("Cleanup complete, project finalized"),
			project.WithProjectGuard("project deleted", func(p *state.Project) bool {
				return projectDeleted(p)
			}),
		)
}

func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(project.State(ImplementationPlanning), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventPlanningComplete), nil
		}).
		OnAdvance(project.State(ImplementationDraftPRCreation), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventDraftPRCreated), nil
		}).
		OnAdvance(project.State(ImplementationExecuting), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventAllTasksComplete), nil
		}).
		// NOTE: ReviewActive OnAdvance is auto-generated by AddBranch (see configureTransitions)
		// The discriminator function getReviewAssessment determines pass/fail branching
		OnAdvance(project.State(FinalizeChecks), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventChecksDone), nil
		}).
		OnAdvance(project.State(FinalizePRReady), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventPRReady), nil
		}).
		OnAdvance(project.State(FinalizePRChecks), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventPRChecksPass), nil
		}).
		OnAdvance(project.State(FinalizeCleanup), func(_ *state.Project) (project.Event, error) {
			return project.Event(EventCleanupComplete), nil
		})
}

func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		// Orchestrator-level prompt (how standard projects work)
		WithOrchestratorPrompt(generateOrchestratorPrompt).

		// State-level prompts (what to do in each state)
		WithPrompt(project.State(ImplementationPlanning), generateImplementationPlanningPrompt).
		WithPrompt(project.State(ImplementationDraftPRCreation), generateImplementationDraftPRCreationPrompt).
		WithPrompt(project.State(ImplementationExecuting), generateImplementationExecutingPrompt).
		WithPrompt(project.State(ReviewActive), generateReviewPrompt).
		WithPrompt(project.State(FinalizeChecks), generateFinalizeChecksPrompt).
		WithPrompt(project.State(FinalizePRReady), generateFinalizePRReadyPrompt).
		WithPrompt(project.State(FinalizePRChecks), generateFinalizePRChecksPrompt).
		WithPrompt(project.State(FinalizeCleanup), generateFinalizeCleanupPrompt)
}
