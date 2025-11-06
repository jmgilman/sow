package exploration

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// TestExplorationLifecycle_SingleSummary tests complete workflow with one summary document.
//nolint:funlen // Test contains multiple subtests for lifecycle verification
func TestExplorationLifecycle_SingleSummary(t *testing.T) {
	// Setup: Create project and state machine
	proj, machine := setupExplorationProject(t)

	// Phase 1: Active research
	t.Run("active research phase", func(t *testing.T) {
		// Verify initial state
		if machine.State() != sdkstate.State(Active) {
			t.Errorf("expected initial state Active, got %v", machine.State())
		}
		verifyPhaseStatus(t, proj, "exploration", "active")

		// Add 3 research topics (tasks)
		addResearchTopic(t, proj, "010", "Research Topic 1", "pending")
		addResearchTopic(t, proj, "020", "Research Topic 2", "pending")
		addResearchTopic(t, proj, "030", "Research Topic 3", "pending")

		// Complete all topics
		markTaskCompleted(t, proj, "exploration", "010")
		markTaskCompleted(t, proj, "exploration", "020")
		markTaskCompleted(t, proj, "exploration", "030")
	})

	// Phase 2: Advance to summarizing
	t.Run("advance to summarizing", func(t *testing.T) {
		// Verify guard passes
		can, err := machine.CanFire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("CanFire(EventBeginSummarizing) error: %v", err)
		}
		if !can {
			t.Error("guard should allow transition with all tasks complete")
		}

		// Fire event
		err = machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("Fire(EventBeginSummarizing) failed: %v", err)
		}

		// Verify transition
		if machine.State() != sdkstate.State(Summarizing) {
			t.Errorf("expected Summarizing state, got %v", machine.State())
		}

		// Verify phase status updated
		verifyPhaseStatus(t, proj, "exploration", "summarizing")
	})

	// Phase 3: Create and approve summary
	t.Run("create and approve summary", func(t *testing.T) {
		// Add summary artifact
		addSummaryArtifact(t, proj, "summary.md", true)

		// Verify summary added
		phase := proj.Phases["exploration"]
		if len(phase.Outputs) == 0 {
			t.Fatal("summary artifact not added")
		}
	})

	// Phase 4: Advance to finalizing
	t.Run("advance to finalizing", func(t *testing.T) {
		// Verify guard passes
		can, err := machine.CanFire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("CanFire(EventCompleteSummarizing) error: %v", err)
		}
		if !can {
			t.Error("guard should allow transition with approved summary")
		}

		// Fire event
		err = machine.Fire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("Fire(EventCompleteSummarizing) failed: %v", err)
		}

		// Verify transition
		if machine.State() != sdkstate.State(Finalizing) {
			t.Errorf("expected Finalizing state, got %v", machine.State())
		}

		// Verify exploration phase completed
		verifyPhaseStatus(t, proj, "exploration", "completed")
		phase := proj.Phases["exploration"]
		if phase.Completed_at.IsZero() {
			t.Error("exploration phase completed_at not set")
		}

		// Verify finalization phase enabled and started
		verifyPhaseStatus(t, proj, "finalization", "in_progress")
		finPhase := proj.Phases["finalization"]
		if !finPhase.Enabled {
			t.Error("finalization phase not enabled")
		}
		if finPhase.Started_at.IsZero() {
			t.Error("finalization phase started_at not set")
		}
	})

	// Phase 5: Complete finalization
	t.Run("complete finalization", func(t *testing.T) {
		// Add finalization task
		addFinalizationTask(t, proj, "100", "Create PR", "pending")

		// Complete task
		markTaskCompleted(t, proj, "finalization", "100")

		// Verify guard passes
		can, err := machine.CanFire(sdkstate.Event(EventCompleteFinalization))
		if err != nil {
			t.Fatalf("CanFire(EventCompleteFinalization) error: %v", err)
		}
		if !can {
			t.Error("guard should allow transition with completed task")
		}

		// Fire event
		err = machine.Fire(sdkstate.Event(EventCompleteFinalization))
		if err != nil {
			t.Fatalf("Fire(EventCompleteFinalization) failed: %v", err)
		}

		// Verify transition to Completed
		if machine.State() != sdkstate.State(Completed) {
			t.Errorf("expected Completed state, got %v", machine.State())
		}

		// Verify finalization phase completed
		verifyPhaseStatus(t, proj, "finalization", "completed")
		finPhase := proj.Phases["finalization"]
		if finPhase.Completed_at.IsZero() {
			t.Error("finalization phase completed_at not set")
		}
	})

	// Verify final state
	t.Run("verify final state", func(t *testing.T) {
		// Check both phases completed
		verifyPhaseStatus(t, proj, "exploration", "completed")
		verifyPhaseStatus(t, proj, "finalization", "completed")

		// Check state machine in Completed
		if machine.State() != sdkstate.State(Completed) {
			t.Errorf("expected Completed state, got %v", machine.State())
		}

		// Check timestamps
		expPhase := proj.Phases["exploration"]
		if expPhase.Completed_at.IsZero() {
			t.Error("exploration completed_at not set")
		}
		finPhase := proj.Phases["finalization"]
		if finPhase.Started_at.IsZero() {
			t.Error("finalization started_at not set")
		}
		if finPhase.Completed_at.IsZero() {
			t.Error("finalization completed_at not set")
		}
	})
}

// TestExplorationLifecycle_MultipleSummaries tests workflow with multiple summary documents.
func TestExplorationLifecycle_MultipleSummaries(t *testing.T) {
	proj, machine := setupExplorationProject(t)

	t.Run("complete research", func(t *testing.T) {
		// Add and complete research topics
		addResearchTopic(t, proj, "010", "Topic 1", "completed")
		addResearchTopic(t, proj, "020", "Topic 2", "completed")

		// Advance to Summarizing
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("Fire(EventBeginSummarizing) failed: %v", err)
		}
	})

	t.Run("create multiple summaries", func(t *testing.T) {
		// Create multiple summary artifacts including summary.md
		addSummaryArtifact(t, proj, "summary.md", true)
		addSummaryArtifact(t, proj, "technical-notes.md", true)
		addSummaryArtifact(t, proj, "findings.md", true)

		// Verify all summaries added
		phase := proj.Phases["exploration"]
		summaryCount := 0
		for _, artifact := range phase.Outputs {
			if artifact.Type == "summary" {
				summaryCount++
			}
		}
		if summaryCount != 3 {
			t.Errorf("expected 3 summaries, got %d", summaryCount)
		}
	})

	t.Run("advance to finalizing", func(t *testing.T) {
		// Should succeed with multiple approved summaries
		err := machine.Fire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("Fire(EventCompleteSummarizing) failed: %v", err)
		}

		if machine.State() != sdkstate.State(Finalizing) {
			t.Errorf("expected Finalizing state, got %v", machine.State())
		}
	})

	t.Run("complete finalization", func(t *testing.T) {
		// Add and complete finalization task
		addFinalizationTask(t, proj, "100", "Create PR", "completed")

		// Advance to Completed
		err := machine.Fire(sdkstate.Event(EventCompleteFinalization))
		if err != nil {
			t.Fatalf("Fire(EventCompleteFinalization) failed: %v", err)
		}

		if machine.State() != sdkstate.State(Completed) {
			t.Errorf("expected Completed state, got %v", machine.State())
		}
	})

	t.Run("verify all summaries handled correctly", func(t *testing.T) {
		// All summaries should still be present and approved
		phase := proj.Phases["exploration"]
		for _, artifact := range phase.Outputs {
			if artifact.Type == "summary" {
				if !artifact.Approved {
					t.Errorf("summary %s not approved", artifact.Path)
				}
			}
		}
	})
}

// TestGuardFailures tests that guards properly block invalid transitions.
//nolint:funlen // Test contains multiple guard validation subtests
func TestGuardFailures(t *testing.T) {
	t.Run("Active to Summarizing blocked with pending tasks", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Add pending task (guard should fail)
		addResearchTopic(t, proj, "010", "Research Topic", "pending")

		// Try to advance (should fail)
		can, err := machine.CanFire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if can {
			t.Error("guard should block transition with pending tasks")
		}

		// Verify state unchanged
		if machine.State() != sdkstate.State(Active) {
			t.Errorf("state should remain Active, got %v", machine.State())
		}
	})

	t.Run("Active to Summarizing blocked with no tasks", func(t *testing.T) {
		_, machine := setupExplorationProject(t)

		// No tasks added - guard should fail
		can, err := machine.CanFire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if can {
			t.Error("guard should block transition with no tasks")
		}
	})

	t.Run("Active to Summarizing allowed with completed tasks", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Add completed task
		addResearchTopic(t, proj, "010", "Research Topic", "completed")

		// Should succeed
		can, err := machine.CanFire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if !can {
			t.Error("guard should allow transition with completed tasks")
		}

		err = machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}

		if machine.State() != sdkstate.State(Summarizing) {
			t.Errorf("expected Summarizing state, got %v", machine.State())
		}
	})

	t.Run("Active to Summarizing allowed with abandoned tasks", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Add abandoned task (should be allowed)
		addResearchTopic(t, proj, "010", "Research Topic", "abandoned")

		// Should succeed
		can, err := machine.CanFire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if !can {
			t.Error("guard should allow transition with abandoned tasks")
		}
	})

	t.Run("Summarizing to Finalizing blocked without summaries", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Setup: advance to Summarizing first
		addResearchTopic(t, proj, "010", "Topic", "completed")
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// No summaries added - guard should fail
		can, err := machine.CanFire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if can {
			t.Error("guard should block transition with no summaries")
		}

		// Verify state unchanged
		if machine.State() != sdkstate.State(Summarizing) {
			t.Errorf("state should remain Summarizing, got %v", machine.State())
		}
	})

	t.Run("Summarizing to Finalizing blocked with unapproved summaries", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Setup: advance to Summarizing
		addResearchTopic(t, proj, "010", "Topic", "completed")
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Add unapproved summary
		addSummaryArtifact(t, proj, "summary.md", false)

		// Guard should fail
		can, err := machine.CanFire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if can {
			t.Error("guard should block transition with unapproved summaries")
		}
	})

	t.Run("Summarizing to Finalizing allowed after approving summaries", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Setup: advance to Summarizing with unapproved summary
		addResearchTopic(t, proj, "010", "Topic", "completed")
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		addSummaryArtifact(t, proj, "summary.md", false)

		// Approve the summary
		phase := proj.Phases["exploration"]
		for i := range phase.Outputs {
			if phase.Outputs[i].Type == "summary" {
				phase.Outputs[i].Approved = true
			}
		}
		proj.Phases["exploration"] = phase

		// Now should succeed
		can, err := machine.CanFire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if !can {
			t.Error("guard should allow transition with approved summaries")
		}

		err = machine.Fire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}

		if machine.State() != sdkstate.State(Finalizing) {
			t.Errorf("expected Finalizing state, got %v", machine.State())
		}
	})

	t.Run("Finalizing to Completed blocked with incomplete tasks", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Setup: advance to Finalizing
		addResearchTopic(t, proj, "010", "Topic", "completed")
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		addSummaryArtifact(t, proj, "summary.md", true)
		err = machine.Fire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Add pending finalization task
		addFinalizationTask(t, proj, "100", "Create PR", "pending")

		// Guard should fail
		can, err := machine.CanFire(sdkstate.Event(EventCompleteFinalization))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if can {
			t.Error("guard should block transition with pending tasks")
		}

		// Verify state unchanged
		if machine.State() != sdkstate.State(Finalizing) {
			t.Errorf("state should remain Finalizing, got %v", machine.State())
		}
	})

	t.Run("Finalizing to Completed blocked with no tasks", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Setup: advance to Finalizing without tasks
		addResearchTopic(t, proj, "010", "Topic", "completed")
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		addSummaryArtifact(t, proj, "summary.md", true)
		err = machine.Fire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// No tasks - guard should fail
		can, err := machine.CanFire(sdkstate.Event(EventCompleteFinalization))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if can {
			t.Error("guard should block transition with no tasks")
		}
	})

	t.Run("Finalizing rejects abandoned tasks", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Setup: advance to Finalizing
		addResearchTopic(t, proj, "010", "Topic", "completed")
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		addSummaryArtifact(t, proj, "summary.md", true)
		err = machine.Fire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Add abandoned task (finalization doesn't accept abandoned)
		addFinalizationTask(t, proj, "100", "Create PR", "abandoned")

		// Guard should fail
		can, err := machine.CanFire(sdkstate.Event(EventCompleteFinalization))
		if err != nil {
			t.Fatalf("CanFire error: %v", err)
		}
		if can {
			t.Error("guard should block transition with abandoned tasks (finalization requires completed)")
		}
	})
}

// TestStateValidation tests that project state validates correctly at each stage.
func TestStateValidation(t *testing.T) {
	t.Run("exploration phase status updates correctly", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Initial: active
		verifyPhaseStatus(t, proj, "exploration", "active")

		// After advancing to Summarizing: summarizing
		addResearchTopic(t, proj, "010", "Topic", "completed")
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}
		verifyPhaseStatus(t, proj, "exploration", "summarizing")

		// After advancing to Finalizing: completed
		addSummaryArtifact(t, proj, "summary.md", true)
		err = machine.Fire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}
		verifyPhaseStatus(t, proj, "exploration", "completed")
	})

	t.Run("finalization phase enabled at correct time", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Initially: disabled
		phase := proj.Phases["finalization"]
		if phase.Enabled {
			t.Error("finalization should start disabled")
		}
		if phase.Status != "pending" {
			t.Errorf("finalization should start pending, got %s", phase.Status)
		}

		// After entering Finalizing state: enabled and in_progress
		addResearchTopic(t, proj, "010", "Topic", "completed")
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}
		addSummaryArtifact(t, proj, "summary.md", true)
		err = machine.Fire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}

		phase = proj.Phases["finalization"]
		if !phase.Enabled {
			t.Error("finalization should be enabled after entering Finalizing state")
		}
		if phase.Status != "in_progress" {
			t.Errorf("finalization status should be in_progress, got %s", phase.Status)
		}
	})

	t.Run("timestamps set correctly", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Exploration started_at should not be set initially (we're already in active)
		// But created_at should be set
		expPhase := proj.Phases["exploration"]
		if expPhase.Created_at.IsZero() {
			t.Error("exploration created_at should be set")
		}

		// Complete workflow
		addResearchTopic(t, proj, "010", "Topic", "completed")
		err := machine.Fire(sdkstate.Event(EventBeginSummarizing))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}
		addSummaryArtifact(t, proj, "summary.md", true)
		err = machine.Fire(sdkstate.Event(EventCompleteSummarizing))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}

		// Exploration completed_at should be set
		expPhase = proj.Phases["exploration"]
		if expPhase.Completed_at.IsZero() {
			t.Error("exploration completed_at should be set after completion")
		}

		// Finalization started_at should be set
		finPhase := proj.Phases["finalization"]
		if finPhase.Started_at.IsZero() {
			t.Error("finalization started_at should be set after entering Finalizing")
		}

		// Complete finalization
		addFinalizationTask(t, proj, "100", "PR", "completed")
		err = machine.Fire(sdkstate.Event(EventCompleteFinalization))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}

		// Finalization completed_at should be set
		finPhase = proj.Phases["finalization"]
		if finPhase.Completed_at.IsZero() {
			t.Error("finalization completed_at should be set after completion")
		}
	})

	t.Run("phase completion markers", func(t *testing.T) {
		proj, machine := setupExplorationProject(t)

		// Complete full lifecycle
		addResearchTopic(t, proj, "010", "Topic", "completed")
		if err := machine.Fire(sdkstate.Event(EventBeginSummarizing)); err != nil {
			t.Fatalf("Failed to fire EventBeginSummarizing: %v", err)
		}
		addSummaryArtifact(t, proj, "summary.md", true)
		if err := machine.Fire(sdkstate.Event(EventCompleteSummarizing)); err != nil {
			t.Fatalf("Failed to fire EventCompleteSummarizing: %v", err)
		}
		addFinalizationTask(t, proj, "100", "PR", "completed")
		if err := machine.Fire(sdkstate.Event(EventCompleteFinalization)); err != nil {
			t.Fatalf("Failed to fire EventCompleteFinalization: %v", err)
		}

		// Both phases should be marked completed
		expPhase := proj.Phases["exploration"]
		if expPhase.Status != "completed" {
			t.Errorf("exploration status should be completed, got %s", expPhase.Status)
		}

		finPhase := proj.Phases["finalization"]
		if finPhase.Status != "completed" {
			t.Errorf("finalization status should be completed, got %s", finPhase.Status)
		}
	})
}

// Helper functions

// setupExplorationProject creates a project for testing in Active state.
func setupExplorationProject(t *testing.T) (*state.Project, *sdkstate.Machine) {
	t.Helper()

	// Create project
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:        "test-exploration",
			Branch:      "explore/test",
			Type:        "exploration",
			Description: "Test exploration project",
			Created_at:  time.Now(),
			Updated_at:  time.Now(),
			Phases:      make(map[string]projschema.PhaseState),
			Statechart: projschema.StatechartState{
				Current_state: string(Active),
				Updated_at:    time.Now(),
			},
		},
	}

	// Initialize phases using the config's initializer
	config := NewExplorationProjectConfig()
	err := initializeExplorationProject(proj, nil)
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Build state machine
	machine := config.BuildMachine(proj, sdkstate.State(Active))
	if machine == nil {
		t.Fatal("BuildMachine returned nil")
	}

	return proj, machine
}

// addResearchTopic adds a task to exploration phase.
func addResearchTopic(t *testing.T, p *state.Project, id, name, status string) {
	t.Helper()

	phase, exists := p.Phases["exploration"]
	if !exists {
		t.Fatal("exploration phase not found")
	}

	task := projschema.TaskState{
		Id:         id,
		Name:       name,
		Phase:      "exploration",
		Status:     status,
		Created_at: time.Now(),
		Updated_at: time.Now(),
		Inputs:     []projschema.ArtifactState{},
	}

	phase.Tasks = append(phase.Tasks, task)
	p.Phases["exploration"] = phase
}

// addSummaryArtifact adds a summary artifact to exploration outputs.
func addSummaryArtifact(t *testing.T, p *state.Project, path string, approved bool) {
	t.Helper()

	phase, exists := p.Phases["exploration"]
	if !exists {
		t.Fatal("exploration phase not found")
	}

	artifact := projschema.ArtifactState{
		Type:       "summary",
		Path:       path,
		Created_at: time.Now(),
		Approved:   approved,
	}

	phase.Outputs = append(phase.Outputs, artifact)
	p.Phases["exploration"] = phase
}

// addFinalizationTask adds a task to finalization phase.
func addFinalizationTask(t *testing.T, p *state.Project, _ /* id */, name, status string) {
	t.Helper()

	phase, exists := p.Phases["finalization"]
	if !exists {
		t.Fatal("finalization phase not found")
	}

	task := projschema.TaskState{
		Id:         "100", // Always use 100 for finalization tasks in tests
		Name:       name,
		Phase:      "finalization",
		Status:     status,
		Created_at: time.Now(),
		Updated_at: time.Now(),
		Inputs:     []projschema.ArtifactState{},
	}

	phase.Tasks = append(phase.Tasks, task)
	p.Phases["finalization"] = phase
}

// verifyPhaseStatus asserts phase is in expected state.
func verifyPhaseStatus(t *testing.T, p *state.Project, phaseName, expectedStatus string) {
	t.Helper()

	phase, exists := p.Phases[phaseName]
	if !exists {
		t.Fatalf("phase %s not found", phaseName)
	}

	if phase.Status != expectedStatus {
		t.Errorf("phase %s status: got %v, want %v", phaseName, phase.Status, expectedStatus)
	}
}

// markTaskCompleted marks an existing task as completed.
func markTaskCompleted(t *testing.T, p *state.Project, phaseName, id string) {
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
