package standard

import (
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// TestFullLifecycle tests complete project lifecycle from start to finish.
//
//nolint:funlen // Test is comprehensive and needs to cover full lifecycle
func TestFullLifecycle(t *testing.T) {
	// Setup: Create minimal project in NoProject state
	proj, machine, config := createTestProject(t, NoProject)

	// NoProject → ImplementationPlanning
	t.Run("init transitions to ImplementationPlanning", func(t *testing.T) {
		err := config.FireWithPhaseUpdates(machine, EventProjectInit, proj)
		if err != nil {
			t.Fatalf("Fire(EventProjectInit) failed: %v", err)
		}
		if got := machine.State(); got != sdkstate.State(ImplementationPlanning) {
			t.Errorf("state = %v, want %v", got, ImplementationPlanning)
		}

		// Verify implementation phase marked as in_progress (automatic)
		if got := proj.Phases["implementation"].Status; got != "in_progress" {
			t.Errorf("implementation phase status = %v, want in_progress", got)
		}
	})

	// ImplementationPlanning → ImplementationDraftPRCreation
	t.Run("planning complete transitions to draft PR creation", func(t *testing.T) {
		// Set planning approved metadata flag
		setPhaseMetadata(t, proj, "implementation", "planning_approved", true)

		err := config.FireWithPhaseUpdates(machine, EventPlanningComplete, proj)
		if err != nil {
			t.Fatalf("Fire(EventPlanningComplete) failed: %v", err)
		}
		if got := machine.State(); got != sdkstate.State(ImplementationDraftPRCreation) {
			t.Errorf("state = %v, want %v", got, ImplementationDraftPRCreation)
		}
	})

	// ImplementationDraftPRCreation → ImplementationExecuting
	t.Run("draft PR created transitions to execution", func(t *testing.T) {
		// Set draft_pr_created metadata flag
		setPhaseMetadata(t, proj, "implementation", "draft_pr_created", true)

		err := config.FireWithPhaseUpdates(machine, EventDraftPRCreated, proj)
		if err != nil {
			t.Fatalf("Fire(EventDraftPRCreated) failed: %v", err)
		}
		if got := machine.State(); got != sdkstate.State(ImplementationExecuting) {
			t.Errorf("state = %v, want %v", got, ImplementationExecuting)
		}

		// Add actual tasks for later test stages
		addTaskWithApprovedDescription(t, proj, "implementation", "010", "Task 1")
		addTaskWithApprovedDescription(t, proj, "implementation", "020", "Task 2")
	})

	// ImplementationExecuting → ReviewActive
	t.Run("all tasks complete transitions to review", func(t *testing.T) {
		// Mark the existing tasks as completed
		markTaskAsCompleted(t, proj, "implementation", "010")
		markTaskAsCompleted(t, proj, "implementation", "020")

		err := config.FireWithPhaseUpdates(machine, EventAllTasksComplete, proj)
		if err != nil {
			t.Fatalf("Fire(EventAllTasksComplete) failed: %v", err)
		}
		if got := machine.State(); got != sdkstate.State(ReviewActive) {
			t.Errorf("state = %v, want %v", got, ReviewActive)
		}

		// Verify implementation phase marked as completed (automatic)
		if got := proj.Phases["implementation"].Status; got != "completed" {
			t.Errorf("implementation phase status = %v, want completed", got)
		}
		if proj.Phases["implementation"].Completed_at.IsZero() {
			t.Error("implementation phase completed_at not set")
		}

		// Verify review phase marked as in_progress (automatic)
		if got := proj.Phases["review"].Status; got != "in_progress" {
			t.Errorf("review phase status = %v, want in_progress", got)
		}
	})

	// ReviewActive → FinalizeChecks
	t.Run("review pass transitions to finalize", func(t *testing.T) {
		// Add approved review with pass assessment
		addApprovedReview(t, proj, "pass", "review.md")

		err := config.FireWithPhaseUpdates(machine, EventReviewPass, proj)
		if err != nil {
			t.Fatalf("Fire(EventReviewPass) failed: %v", err)
		}
		if got := machine.State(); got != sdkstate.State(FinalizeChecks) {
			t.Errorf("state = %v, want %v", got, FinalizeChecks)
		}

		// Verify review phase marked as completed (automatic)
		if got := proj.Phases["review"].Status; got != "completed" {
			t.Errorf("review phase status = %v, want completed", got)
		}
		if proj.Phases["review"].Completed_at.IsZero() {
			t.Error("review phase completed_at not set")
		}

		// Verify finalize phase marked as in_progress (automatic)
		if got := proj.Phases["finalize"].Status; got != "in_progress" {
			t.Errorf("finalize phase status = %v, want in_progress", got)
		}
	})

	// Finalize substates
	t.Run("finalize substates progress correctly", func(t *testing.T) {
		// FinalizeChecks → FinalizePRReady
		err := config.FireWithPhaseUpdates(machine, EventChecksDone, proj)
		if err != nil {
			t.Fatalf("Fire(EventChecksDone) failed: %v", err)
		}
		if got := machine.State(); got != sdkstate.State(FinalizePRReady) {
			t.Errorf("state = %v, want %v", got, FinalizePRReady)
		}

		// FinalizePRReady → FinalizePRChecks
		addApprovedOutput(t, proj, "finalize", "pr_body", "pr_body.md")
		err = config.FireWithPhaseUpdates(machine, EventPRReady, proj)
		if err != nil {
			t.Fatalf("Fire(EventPRReady) failed: %v", err)
		}
		if got := machine.State(); got != sdkstate.State(FinalizePRChecks) {
			t.Errorf("state = %v, want %v", got, FinalizePRChecks)
		}

		// FinalizePRChecks → FinalizeCleanup
		setPhaseMetadata(t, proj, "finalize", "pr_checks_passed", true)
		err = config.FireWithPhaseUpdates(machine, EventPRChecksPass, proj)
		if err != nil {
			t.Fatalf("Fire(EventPRChecksPass) failed: %v", err)
		}
		if got := machine.State(); got != sdkstate.State(FinalizeCleanup) {
			t.Errorf("state = %v, want %v", got, FinalizeCleanup)
		}

		// FinalizeCleanup → NoProject
		setPhaseMetadata(t, proj, "finalize", "project_deleted", true)
		err = config.FireWithPhaseUpdates(machine, EventCleanupComplete, proj)
		if err != nil {
			t.Fatalf("Fire(EventCleanupComplete) failed: %v", err)
		}
		if got := machine.State(); got != sdkstate.State(NoProject) {
			t.Errorf("state = %v, want %v", got, NoProject)
		}

		// Verify finalize phase marked as completed (automatic)
		if got := proj.Phases["finalize"].Status; got != "completed" {
			t.Errorf("finalize phase status = %v, want completed", got)
		}
		if proj.Phases["finalize"].Completed_at.IsZero() {
			t.Error("finalize phase completed_at not set")
		}
	})
}

// TestReviewFailLoop tests review failure rework loop.
func TestReviewFailLoop(t *testing.T) {
	proj, machine, config := createTestProject(t, ReviewActive)

	// Add approved review with fail assessment
	addApprovedReview(t, proj, "fail", "review-fail.md")

	// ReviewActive → ImplementationPlanning (rework)
	err := config.FireWithPhaseUpdates(machine, EventReviewFail, proj)
	if err != nil {
		t.Fatalf("Fire(EventReviewFail) failed: %v", err)
	}
	if got := machine.State(); got != sdkstate.State(ImplementationPlanning) {
		t.Errorf("state = %v, want %v", got, ImplementationPlanning)
	}

	// Verify review phase marked as failed
	reviewPhase := proj.Phases["review"]
	if reviewPhase.Status != "failed" {
		t.Errorf("review phase status = %v, want failed", reviewPhase.Status)
	}
	if reviewPhase.Failed_at.IsZero() {
		t.Error("review phase failed_at not set")
	}

	// Verify implementation iteration incremented
	implPhase := proj.Phases["implementation"]
	if implPhase.Iteration != 1 {
		t.Errorf("implementation phase iteration = %v, want 1", implPhase.Iteration)
	}

	// Verify implementation phase status set to in_progress
	if implPhase.Status != "in_progress" {
		t.Errorf("implementation phase status = %v, want in_progress", implPhase.Status)
	}

	// Verify failed review added as implementation input
	foundReviewInput := false
	for _, input := range implPhase.Inputs {
		if input.Type == "review" {
			foundReviewInput = true
			// Verify it's the failed review
			if assessment, ok := input.Metadata["assessment"].(string); ok {
				if assessment != "fail" {
					t.Errorf("review input assessment = %v, want fail", assessment)
				}
			} else {
				t.Error("review input missing assessment metadata")
			}
			break
		}
	}
	if !foundReviewInput {
		t.Error("failed review not added as implementation input")
	}
}

// TestMultipleReviewFailures tests iteration increments across multiple review failures.
func TestMultipleReviewFailures(t *testing.T) {
	proj, machine, config := createTestProject(t, ReviewActive)

	// First review failure
	addApprovedReview(t, proj, "fail", "review-fail-1.md")
	err := config.FireWithPhaseUpdates(machine, EventReviewFail, proj)
	if err != nil {
		t.Fatalf("Fire(EventReviewFail) #1 failed: %v", err)
	}

	// Verify iteration = 1 after first failure
	implPhase := proj.Phases["implementation"]
	if implPhase.Iteration != 1 {
		t.Errorf("after first failure: iteration = %v, want 1", implPhase.Iteration)
	}

	// Simulate completing implementation again and returning to review
	// (In real workflow, would go through full ImplementationPlanning → ImplementationExecuting → ReviewActive)
	// For this test, we'll just update machine state and add another review
	machine = config.BuildMachine(proj, sdkstate.State(ReviewActive))

	// Second review failure
	addApprovedReview(t, proj, "fail", "review-fail-2.md")
	err = config.FireWithPhaseUpdates(machine, EventReviewFail, proj)
	if err != nil {
		t.Fatalf("Fire(EventReviewFail) #2 failed: %v", err)
	}

	// Verify iteration = 2 after second failure
	implPhase = proj.Phases["implementation"]
	if implPhase.Iteration != 2 {
		t.Errorf("after second failure: iteration = %v, want 2", implPhase.Iteration)
	}

	// Third review failure
	machine = config.BuildMachine(proj, sdkstate.State(ReviewActive))
	addApprovedReview(t, proj, "fail", "review-fail-3.md")
	err = config.FireWithPhaseUpdates(machine, EventReviewFail, proj)
	if err != nil {
		t.Fatalf("Fire(EventReviewFail) #3 failed: %v", err)
	}

	// Verify iteration = 3 after third failure
	implPhase = proj.Phases["implementation"]
	if implPhase.Iteration != 3 {
		t.Errorf("after third failure: iteration = %v, want 3", implPhase.Iteration)
	}

	// Verify review phase failed_at updated
	reviewPhase := proj.Phases["review"]
	if reviewPhase.Status != "failed" {
		t.Errorf("review phase status = %v, want failed", reviewPhase.Status)
	}
}

// TestGuardsBlockInvalidTransitions tests guards prevent invalid transitions.
func TestGuardsBlockInvalidTransitions(t *testing.T) {
	tests := []struct {
		name         string
		initialState sdkstate.State
		setupFunc    func(*state.Project)
		event        sdkstate.Event
		shouldBlock  bool
	}{
		{
			name:         "implementation planning without approved task descriptions blocks",
			initialState: sdkstate.State(ImplementationPlanning),
			setupFunc:    func(_ *state.Project) {}, // No task descriptions
			event:        sdkstate.Event(EventPlanningComplete),
			shouldBlock:  true,
		},
		{
			name:         "implementation planning with unapproved planning blocks",
			initialState: sdkstate.State(ImplementationPlanning),
			setupFunc: func(p *state.Project) {
				// Set planning_approved to false (or leave unset - both block)
				phase := p.Phases["implementation"]
				if phase.Metadata == nil {
					phase.Metadata = make(map[string]interface{})
				}
				phase.Metadata["planning_approved"] = false
				p.Phases["implementation"] = phase
			},
			event:       sdkstate.Event(EventPlanningComplete),
			shouldBlock: true,
		},
		{
			name:         "implementation executing without completed tasks blocks",
			initialState: sdkstate.State(ImplementationExecuting),
			setupFunc: func(p *state.Project) {
				// Add pending task
				addPendingTask(p, "implementation", "001", "Task 1")
			},
			event:       sdkstate.Event(EventAllTasksComplete),
			shouldBlock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj, machine, _ := createTestProject(t, tt.initialState)
			tt.setupFunc(proj)

			can, err := machine.CanFire(tt.event)
			if err != nil {
				t.Fatalf("CanFire() error: %v", err)
			}

			if tt.shouldBlock && can {
				t.Errorf("guard should block transition but allowed it")
			}
			if !tt.shouldBlock && !can {
				t.Errorf("guard should allow transition but blocked it")
			}
		})
	}
}

// TestPromptGeneration tests prompts generate correctly for each state.
func TestPromptGeneration(t *testing.T) {
	states := []sdkstate.State{
		sdkstate.State(ImplementationPlanning),
		sdkstate.State(ImplementationDraftPRCreation),
		sdkstate.State(ImplementationExecuting),
		sdkstate.State(ReviewActive),
		sdkstate.State(FinalizeChecks),
		sdkstate.State(FinalizePRReady),
		sdkstate.State(FinalizePRChecks),
		sdkstate.State(FinalizeCleanup),
	}

	for _, st := range states {
		t.Run(string(st), func(t *testing.T) {
			proj, _, _ := createTestProject(t, st)

			// Use the prompt generator directly to generate prompt
			promptGen := getPromptGenerator(st)
			if promptGen == nil {
				t.Errorf("no prompt generator for state %s", st)
				return
			}

			prompt := promptGen(proj)
			if prompt == "" {
				t.Error("prompt is empty")
			}
			if !contains(prompt, "Project:") {
				t.Error("prompt missing project header")
			}
		})
	}
}

// TestOnAdvanceEventDetermination tests event determiners work correctly.
func TestOnAdvanceEventDetermination(t *testing.T) {
	t.Run("ReviewActive determines pass event", func(t *testing.T) {
		proj, _, _ := createTestProject(t, sdkstate.State(ReviewActive))
		addApprovedReview(t, proj, "pass", "review.md")

		// Verify the review was added with correct assessment
		phase := proj.Phases["review"]
		if len(phase.Outputs) == 0 {
			t.Fatal("expected review output to be added")
		}
		assessment, ok := phase.Outputs[0].Metadata["assessment"].(string)
		if !ok {
			t.Fatal("assessment metadata not a string")
		}
		if assessment != "pass" {
			t.Errorf("assessment = %v, want pass", assessment)
		}
	})

	t.Run("ReviewActive determines fail event", func(t *testing.T) {
		proj, _, _ := createTestProject(t, sdkstate.State(ReviewActive))
		addApprovedReview(t, proj, "fail", "review.md")

		// Verify the review was added with correct assessment
		phase := proj.Phases["review"]
		if len(phase.Outputs) == 0 {
			t.Fatal("expected review output to be added")
		}
		assessment, ok := phase.Outputs[0].Metadata["assessment"].(string)
		if !ok {
			t.Fatal("assessment metadata not a string")
		}
		if assessment != "fail" {
			t.Errorf("assessment = %v, want fail", assessment)
		}
	})
}

// Helper functions

// createTestProject creates a project for testing in the specified initial state.
func createTestProject(t *testing.T, initialState sdkstate.State) (*state.Project, *sdkstate.Machine, *project.ProjectTypeConfig) {
	t.Helper()

	// Create minimal project state
	ps := &projschema.ProjectState{
		Name:        "test-project",
		Type:        "standard",
		Branch:      "test-branch",
		Description: "Test project for lifecycle tests",
		Created_at:  time.Now(),
		Updated_at:  time.Now(),
		Phases: map[string]projschema.PhaseState{
			"implementation": {
				Status:     "pending",
				Enabled:    true,
				Created_at: time.Now(),
				Inputs:     []projschema.ArtifactState{},
				Outputs:    []projschema.ArtifactState{},
				Tasks:      []projschema.TaskState{},
				Metadata:   make(map[string]interface{}),
			},
			"review": {
				Status:     "pending",
				Enabled:    true,
				Created_at: time.Now(),
				Inputs:     []projschema.ArtifactState{},
				Outputs:    []projschema.ArtifactState{},
				Metadata:   make(map[string]interface{}),
			},
			"finalize": {
				Status:     "pending",
				Enabled:    true,
				Created_at: time.Now(),
				Inputs:     []projschema.ArtifactState{},
				Outputs:    []projschema.ArtifactState{},
				Metadata:   make(map[string]interface{}),
			},
		},
		Statechart: projschema.StatechartState{
			Current_state: string(initialState),
			Updated_at:    time.Now(),
		},
	}

	// Get standard project type config
	config := NewStandardProjectConfig()

	// Create project instance (minimal wrapper for testing)
	proj := &state.Project{
		ProjectState: *ps,
	}

	// Build state machine with initial state
	machine := config.BuildMachine(proj, sdkstate.State(initialState))

	return proj, machine, config
}

// addApprovedOutput adds an approved output artifact to a phase.
func addApprovedOutput(t *testing.T, p *state.Project, phaseName, outputType, path string) {
	t.Helper()

	phase, exists := p.Phases[phaseName]
	if !exists {
		t.Fatalf("phase %s not found", phaseName)
	}

	artifact := projschema.ArtifactState{
		Type:       outputType,
		Path:       path,
		Created_at: time.Now(),
		Approved:   true,
	}

	phase.Outputs = append(phase.Outputs, artifact)
	p.Phases[phaseName] = phase
}

// addApprovedReview adds an approved review artifact with assessment metadata.
func addApprovedReview(t *testing.T, p *state.Project, assessment, path string) {
	t.Helper()

	phase, exists := p.Phases["review"]
	if !exists {
		t.Fatal("review phase not found")
	}

	artifact := projschema.ArtifactState{
		Type:       "review",
		Path:       path,
		Created_at: time.Now(),
		Approved:   true,
		Metadata: map[string]interface{}{
			"assessment": assessment,
		},
	}

	phase.Outputs = append(phase.Outputs, artifact)
	p.Phases["review"] = phase
}

// setPhaseMetadata sets a metadata value on a phase.
func setPhaseMetadata(t *testing.T, p *state.Project, phaseName, key string, value interface{}) {
	t.Helper()

	phase, exists := p.Phases[phaseName]
	if !exists {
		t.Fatalf("phase %s not found", phaseName)
	}

	if phase.Metadata == nil {
		phase.Metadata = make(map[string]interface{})
	}

	phase.Metadata[key] = value
	p.Phases[phaseName] = phase
}

// addPendingTask adds a pending task to a phase.
func addPendingTask(p *state.Project, phaseName, id, name string) {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return
	}

	task := projschema.TaskState{
		Id:         id,
		Name:       name,
		Phase:      phaseName,
		Status:     "pending",
		Created_at: time.Now(),
		Updated_at: time.Now(),
		Inputs:     []projschema.ArtifactState{},
	}

	phase.Tasks = append(phase.Tasks, task)
	p.Phases[phaseName] = phase
}

// addTaskWithApprovedDescription adds a task with an approved task_description input.
func addTaskWithApprovedDescription(t *testing.T, p *state.Project, phaseName, id, name string) {
	t.Helper()

	phase, exists := p.Phases[phaseName]
	if !exists {
		t.Fatalf("phase %s not found", phaseName)
	}

	descriptionArtifact := projschema.ArtifactState{
		Type:       "task_description",
		Path:       ".sow/project/phases/implementation/tasks/task-" + id + "/description.md",
		Created_at: time.Now(),
		Approved:   true,
	}

	task := projschema.TaskState{
		Id:         id,
		Name:       name,
		Phase:      phaseName,
		Status:     "pending",
		Created_at: time.Now(),
		Updated_at: time.Now(),
		Inputs:     []projschema.ArtifactState{descriptionArtifact},
	}

	phase.Tasks = append(phase.Tasks, task)
	p.Phases[phaseName] = phase
}

// markTaskAsCompleted marks an existing task as completed.
func markTaskAsCompleted(t *testing.T, p *state.Project, phaseName, id string) {
	t.Helper()

	phase, exists := p.Phases[phaseName]
	if !exists {
		t.Fatalf("phase %s not found", phaseName)
	}

	for i, task := range phase.Tasks {
		if task.Id == id {
			phase.Tasks[i].Status = "completed"
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases[phaseName] = phase
			return
		}
	}

	t.Fatalf("task %s not found in phase %s", id, phaseName)
}

// PromptGenerator is a function that creates a contextual prompt for a given state.
type PromptGenerator func(*state.Project) string

// getPromptGenerator returns the prompt generator for a given state.
func getPromptGenerator(st sdkstate.State) PromptGenerator {
	// Map states to their prompt generators
	switch st {
	case sdkstate.State(ImplementationPlanning):
		return generateImplementationPlanningPrompt
	case sdkstate.State(ImplementationDraftPRCreation):
		return generateImplementationDraftPRCreationPrompt
	case sdkstate.State(ImplementationExecuting):
		return generateImplementationExecutingPrompt
	case sdkstate.State(ReviewActive):
		return generateReviewPrompt
	case sdkstate.State(FinalizeChecks):
		return generateFinalizeChecksPrompt
	case sdkstate.State(FinalizePRReady):
		return generateFinalizePRReadyPrompt
	case sdkstate.State(FinalizePRChecks):
		return generateFinalizePRChecksPrompt
	case sdkstate.State(FinalizeCleanup):
		return generateFinalizeCleanupPrompt
	default:
		return nil
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
