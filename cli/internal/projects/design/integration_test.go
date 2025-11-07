package design

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDesignLifecycle_SingleDocument tests complete workflow with one design document.
// This is the primary happy path test covering all states: Active → Finalizing → Completed.
//nolint:funlen // Test contains multiple subtests for lifecycle verification
func TestDesignLifecycle_SingleDocument(t *testing.T) {
	// Setup: Create project and state machine
	proj, machine, config := setupDesignProject(t)

	// Phase 1: Active design - plan and draft document
	t.Run("plan and draft document", func(t *testing.T) {
		// Verify initial state
		assert.Equal(t, sdkstate.State(Active), machine.State(), "initial state should be Active")
		verifyPhaseStatus(t, proj, "design", "active")
		verifyPhaseEnabled(t, proj, "design", true)

		// Create document task
		addDocumentTask(t, proj, "010", "Auth Design", "design", ".sow/knowledge/designs/auth.md")

		// Start working on document
		markTaskInProgress(t, proj, "010")
		verifyTaskStatus(t, proj, "design", "010", "in_progress")

		// Draft document and link to task
		artifactPath := addDesignArtifact(t, proj, "project/phases/design/auth-design.md", "design")
		linkArtifactToTask(t, proj, "010", artifactPath)

		// Mark complete (should auto-approve artifact)
		markTaskCompleted(t, proj, "010")

		// Verify task completed
		verifyTaskStatus(t, proj, "design", "010", "completed")

		// Verify auto-approval
		verifyArtifactApproved(t, proj, artifactPath, true)
	})

	// Phase 2: Advance to finalizing
	t.Run("advance to finalizing", func(t *testing.T) {
		// Verify guard passes
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err, "CanFire should not error")
		assert.True(t, can, "guard should allow transition with completed document")

		// Fire event
		err = config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		require.NoError(t, err, "Fire should not error")

		// Verify state transition
		assert.Equal(t, sdkstate.State(Finalizing), machine.State(), "state should be Finalizing")

		// Verify design phase completed
		verifyPhaseStatus(t, proj, "design", "completed")
		designPhase := proj.Phases["design"]
		assert.False(t, designPhase.Completed_at.IsZero(), "design phase should have completion time")

		// Verify finalization phase enabled and started
		verifyPhaseStatus(t, proj, "finalization", "in_progress")
		verifyPhaseEnabled(t, proj, "finalization", true)
		finPhase := proj.Phases["finalization"]
		assert.False(t, finPhase.Started_at.IsZero(), "finalization phase should have start time")
	})

	// Phase 3: Complete finalization
	t.Run("complete finalization", func(t *testing.T) {
		// Add finalization task
		addFinalizationTask(t, proj, "100", "Move documents to knowledge base", "pending")

		// Complete task
		markFinalizationTaskCompleted(t, proj, "100")

		// Verify guard passes
		can, err := machine.CanFire(sdkstate.Event(EventCompleteFinalization))
		require.NoError(t, err, "CanFire should not error")
		assert.True(t, can, "guard should allow transition with completed task")

		// Fire event
		err = config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteFinalization), proj)
		require.NoError(t, err, "Fire should not error")

		// Verify state transition to Completed
		assert.Equal(t, sdkstate.State(Completed), machine.State(), "state should be Completed")

		// Verify finalization phase completed
		verifyPhaseStatus(t, proj, "finalization", "completed")
		finPhase := proj.Phases["finalization"]
		assert.False(t, finPhase.Completed_at.IsZero(), "finalization phase should have completion time")
	})

	// Verify final state
	t.Run("verify final state", func(t *testing.T) {
		// Both phases should be completed
		verifyPhaseStatus(t, proj, "design", "completed")
		verifyPhaseStatus(t, proj, "finalization", "completed")

		// State machine should be in Completed state
		assert.Equal(t, sdkstate.State(Completed), machine.State(), "final state should be Completed")

		// Verify all timestamps set
		designPhase := proj.Phases["design"]
		assert.False(t, designPhase.Created_at.IsZero(), "design created_at should be set")
		assert.False(t, designPhase.Completed_at.IsZero(), "design completed_at should be set")

		finPhase := proj.Phases["finalization"]
		assert.False(t, finPhase.Created_at.IsZero(), "finalization created_at should be set")
		assert.False(t, finPhase.Started_at.IsZero(), "finalization started_at should be set")
		assert.False(t, finPhase.Completed_at.IsZero(), "finalization completed_at should be set")
	})
}

// TestDesignLifecycle_MultipleDocuments tests workflow with multiple design documents of different types.
//nolint:funlen // Test contains multiple subtests for comprehensive document type coverage
func TestDesignLifecycle_MultipleDocuments(t *testing.T) {
	proj, machine, config := setupDesignProject(t)

	t.Run("create multiple document tasks", func(t *testing.T) {
		// Create tasks for different document types
		addDocumentTask(t, proj, "010", "API Design", "design", ".sow/knowledge/designs/api.md")
		addDocumentTask(t, proj, "020", "Authentication ADR", "adr", ".sow/knowledge/adrs/001-auth.md")
		addDocumentTask(t, proj, "030", "System Architecture", "architecture", ".sow/knowledge/architecture/system.md")

		// Verify all tasks created
		phase := proj.Phases["design"]
		assert.Len(t, phase.Tasks, 3, "should have 3 tasks")
	})

	t.Run("complete some tasks and abandon one", func(t *testing.T) {
		// Complete first task
		markTaskInProgress(t, proj, "010")
		artifact1 := addDesignArtifact(t, proj, "project/api-design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact1)
		markTaskCompleted(t, proj, "010")
		verifyArtifactApproved(t, proj, artifact1, true)

		// Complete second task
		markTaskInProgress(t, proj, "020")
		artifact2 := addDesignArtifact(t, proj, "project/adr-auth.md", "adr")
		linkArtifactToTask(t, proj, "020", artifact2)
		markTaskCompleted(t, proj, "020")
		verifyArtifactApproved(t, proj, artifact2, true)

		// Abandon third task (decided not needed)
		markTaskAbandoned(t, proj, "030")

		// Verify statuses
		verifyTaskStatus(t, proj, "design", "010", "completed")
		verifyTaskStatus(t, proj, "design", "020", "completed")
		verifyTaskStatus(t, proj, "design", "030", "abandoned")
	})

	t.Run("advance to finalizing with mix of completed and abandoned", func(t *testing.T) {
		// Should succeed - has at least one completed task
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow transition with completed tasks")

		err = config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		require.NoError(t, err)

		assert.Equal(t, sdkstate.State(Finalizing), machine.State(), "state should be Finalizing")
	})

	t.Run("complete finalization", func(t *testing.T) {
		addFinalizationTask(t, proj, "100", "Move documents", "completed")

		err := config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteFinalization), proj)
		require.NoError(t, err)

		assert.Equal(t, sdkstate.State(Completed), machine.State(), "state should be Completed")
	})

	t.Run("verify all artifacts handled correctly", func(t *testing.T) {
		// All completed task artifacts should be approved
		phase := proj.Phases["design"]
		approvedCount := 0
		for _, artifact := range phase.Outputs {
			if artifact.Approved {
				approvedCount++
			}
		}
		assert.Equal(t, 2, approvedCount, "should have 2 approved artifacts")
	})
}

// TestDesignLifecycle_WithInputs tests design project with initial input artifacts.
func TestDesignLifecycle_WithInputs(t *testing.T) {
	t.Run("create project with inputs", func(t *testing.T) {
		// Create project with inputs
		now := time.Now()
		proj := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:       "test-design-with-inputs",
				Branch:     "design/test-inputs",
				Type:       "design",
				Created_at: now,
				Updated_at: now,
				Phases:     make(map[string]projschema.PhaseState),
			},
		}

		// Initialize with inputs
		initialInputs := map[string][]projschema.ArtifactState{
			"design": {
				{
					Type:       "requirement",
					Path:       ".sow/context/requirements.md",
					Created_at: now,
					Approved:   true,
				},
			},
		}

		config := NewDesignProjectConfig()
		err := config.Initialize(proj, initialInputs)
		require.NoError(t, err)

		// Build state machine
		machine := config.BuildMachine(proj, sdkstate.State(Active))
		require.NotNil(t, machine)

		// Verify inputs are tracked separately
		designPhase := proj.Phases["design"]
		assert.Len(t, designPhase.Inputs, 1, "should have 1 input")
		assert.Equal(t, "requirement", designPhase.Inputs[0].Type)
		assert.Empty(t, designPhase.Outputs, "outputs should be empty initially")

		// Complete workflow
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")
		markTaskInProgress(t, proj, "010")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")

		// Inputs don't block progression
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.True(t, can, "inputs should not block progression")
	})
}

// TestDesignLifecycle_ReviewWorkflow tests needs_review → in_progress → completed workflow.
//nolint:funlen // Test contains multiple subtests for review workflow verification
func TestDesignLifecycle_ReviewWorkflow(t *testing.T) {
	proj, machine, _ := setupDesignProject(t)

	t.Run("create and draft document", func(t *testing.T) {
		addDocumentTask(t, proj, "010", "API Design", "design", ".sow/knowledge/designs/api.md")
		markTaskInProgress(t, proj, "010")

		artifactPath := addDesignArtifact(t, proj, "project/api-design.md", "design")
		linkArtifactToTask(t, proj, "010", artifactPath)

		verifyTaskStatus(t, proj, "design", "010", "in_progress")
		verifyArtifactApproved(t, proj, artifactPath, false)
	})

	t.Run("request review", func(t *testing.T) {
		markTaskNeedsReview(t, proj, "010")
		verifyTaskStatus(t, proj, "design", "010", "needs_review")

		// Cannot advance with needs_review task
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.False(t, can, "should not advance with task in review")
	})

	t.Run("revise after feedback", func(t *testing.T) {
		// Go back to in_progress for revisions
		markTaskInProgress(t, proj, "010")
		verifyTaskStatus(t, proj, "design", "010", "in_progress")

		// Still cannot advance
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.False(t, can, "should not advance with in_progress task")

		// Request review again
		markTaskNeedsReview(t, proj, "010")
		verifyTaskStatus(t, proj, "design", "010", "needs_review")
	})

	t.Run("approve and complete", func(t *testing.T) {
		markTaskCompleted(t, proj, "010")
		verifyTaskStatus(t, proj, "design", "010", "completed")

		// Verify auto-approval
		phase := proj.Phases["design"]
		var artifact *projschema.ArtifactState
		for i := range phase.Outputs {
			if phase.Outputs[i].Path == "project/api-design.md" {
				artifact = &phase.Outputs[i]
				break
			}
		}
		require.NotNil(t, artifact, "artifact should exist")
		assert.True(t, artifact.Approved, "artifact should be auto-approved on completion")

		// Can now advance
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.True(t, can, "should be able to advance after completion")
	})
}

// TestDesignLifecycle_AllAbandoned tests that advancement is blocked when all tasks are abandoned.
func TestDesignLifecycle_AllAbandoned(t *testing.T) {
	proj, machine, config := setupDesignProject(t)

	t.Run("create and abandon all tasks", func(t *testing.T) {
		// Create multiple tasks
		addDocumentTask(t, proj, "010", "Design Doc 1", "design", ".sow/knowledge/designs/doc1.md")
		addDocumentTask(t, proj, "020", "Design Doc 2", "design", ".sow/knowledge/designs/doc2.md")
		addDocumentTask(t, proj, "030", "Design Doc 3", "design", ".sow/knowledge/designs/doc3.md")

		// Abandon all tasks
		markTaskAbandoned(t, proj, "010")
		markTaskAbandoned(t, proj, "020")
		markTaskAbandoned(t, proj, "030")

		// Verify all abandoned
		verifyTaskStatus(t, proj, "design", "010", "abandoned")
		verifyTaskStatus(t, proj, "design", "020", "abandoned")
		verifyTaskStatus(t, proj, "design", "030", "abandoned")
	})

	t.Run("cannot advance with all abandoned", func(t *testing.T) {
		// Guard should block - needs at least one completed
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.False(t, can, "should not advance when all tasks abandoned")

		// Try to fire anyway - should fail
		err = config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		assert.Error(t, err, "should error when trying to fire blocked transition")

		// Verify state unchanged
		assert.Equal(t, sdkstate.State(Active), machine.State(), "state should remain Active")
	})
}

// TestDesignLifecycle_NoTasks tests that advancement is blocked when no tasks exist.
func TestDesignLifecycle_NoTasks(t *testing.T) {
	proj, machine, _ := setupDesignProject(t)

	t.Run("cannot advance without tasks", func(t *testing.T) {
		// No tasks created
		phase := proj.Phases["design"]
		assert.Empty(t, phase.Tasks, "should have no tasks")

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.False(t, can, "should not advance with no tasks")

		// Verify state unchanged
		assert.Equal(t, sdkstate.State(Active), machine.State(), "state should remain Active")
	})
}

// TestDesignLifecycle_TaskValidation tests validateTaskForCompletion enforcement.
func TestDesignLifecycle_TaskValidation(t *testing.T) {
	proj, _, _ := setupDesignProject(t)

	t.Run("completion blocked without artifact", func(t *testing.T) {
		// Create task without artifact
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")

		// Try to validate for completion - should fail
		err := validateTaskForCompletion(proj, "010")
		assert.Error(t, err, "should error when completing task without artifact")
		assert.Contains(t, err.Error(), "metadata", "error should mention metadata")
	})

	t.Run("completion blocked with artifact_path but no artifact", func(t *testing.T) {
		// Link task to non-existent artifact
		phase := proj.Phases["design"]
		for i := range phase.Tasks {
			if phase.Tasks[i].Id == "010" {
				phase.Tasks[i].Metadata = map[string]interface{}{
					"artifact_path": "project/nonexistent.md",
				}
			}
		}
		proj.Phases["design"] = phase

		// Try to validate - should fail
		err := validateTaskForCompletion(proj, "010")
		assert.Error(t, err, "should error when artifact doesn't exist")
		assert.Contains(t, err.Error(), "artifact not found", "error should mention missing artifact")
	})

	t.Run("completion allowed with valid artifact", func(t *testing.T) {
		// Add artifact
		artifact := addDesignArtifact(t, proj, "project/valid-design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)

		// Should validate successfully
		err := validateTaskForCompletion(proj, "010")
		assert.NoError(t, err, "should allow completion with valid artifact")
	})
}

// TestDesignLifecycle_AutoApproval tests automatic artifact approval on task completion.
func TestDesignLifecycle_AutoApproval(t *testing.T) {
	proj, _, _ := setupDesignProject(t)

	t.Run("artifact auto-approved on task completion", func(t *testing.T) {
		// Create task and artifact
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)

		// Initially not approved
		verifyArtifactApproved(t, proj, artifact, false)

		// Complete task (should auto-approve)
		markTaskCompleted(t, proj, "010")

		// Verify auto-approval
		verifyArtifactApproved(t, proj, artifact, true)
	})

	t.Run("auto-approval is atomic with completion", func(t *testing.T) {
		// Create second task and artifact
		addDocumentTask(t, proj, "020", "ADR Doc", "adr", ".sow/knowledge/adrs/001.md")
		artifact := addDesignArtifact(t, proj, "project/adr-001.md", "adr")
		linkArtifactToTask(t, proj, "020", artifact)

		// Complete task
		markTaskCompleted(t, proj, "020")

		// Approval should happen immediately
		phase := proj.Phases["design"]
		for _, a := range phase.Outputs {
			if a.Path == artifact {
				assert.True(t, a.Approved, "artifact should be approved immediately after completion")
			}
		}
	})
}

// TestGuardFailures tests that guards properly block invalid transitions.
//nolint:funlen // Test contains multiple guard validation subtests
func TestGuardFailures(t *testing.T) {
	t.Run("Active to Finalizing blocked with pending tasks", func(t *testing.T) {
		proj, machine, _ := setupDesignProject(t)

		// Add pending task
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with pending tasks")
	})

	t.Run("Active to Finalizing blocked with in_progress tasks", func(t *testing.T) {
		proj, machine, _ := setupDesignProject(t)

		// Add in_progress task
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")
		markTaskInProgress(t, proj, "010")

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with in_progress tasks")
	})

	t.Run("Active to Finalizing blocked with needs_review tasks", func(t *testing.T) {
		proj, machine, _ := setupDesignProject(t)

		// Add needs_review task
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskNeedsReview(t, proj, "010")

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with needs_review tasks")
	})

	t.Run("Active to Finalizing allowed with completed tasks", func(t *testing.T) {
		proj, machine, _ := setupDesignProject(t)

		// Add completed task
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")

		// Guard should allow
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow transition with completed tasks")
	})

	t.Run("Active to Finalizing allowed with mix of completed and abandoned", func(t *testing.T) {
		proj, machine, _ := setupDesignProject(t)

		// Add completed task
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")

		// Add abandoned task
		addDocumentTask(t, proj, "020", "Other Doc", "design", ".sow/knowledge/designs/other.md")
		markTaskAbandoned(t, proj, "020")

		// Guard should allow
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow transition with completed and abandoned tasks")
	})

	t.Run("Finalizing to Completed blocked with pending tasks", func(t *testing.T) {
		proj, machine, config := setupDesignProject(t)

		// Advance to Finalizing
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")
		err := config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		require.NoError(t, err)

		// Add pending finalization task
		addFinalizationTask(t, proj, "100", "Move docs", "pending")

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventCompleteFinalization))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with pending tasks")
	})

	t.Run("Finalizing to Completed blocked with no tasks", func(t *testing.T) {
		proj, machine, config := setupDesignProject(t)

		// Advance to Finalizing
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")
		err := config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		require.NoError(t, err)

		// No finalization tasks
		phase := proj.Phases["finalization"]
		assert.Empty(t, phase.Tasks, "should have no finalization tasks")

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventCompleteFinalization))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with no tasks")
	})

	t.Run("Finalizing rejects abandoned tasks", func(t *testing.T) {
		proj, machine, config := setupDesignProject(t)

		// Advance to Finalizing
		addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")
		err := config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		require.NoError(t, err)

		// Add abandoned finalization task
		addFinalizationTask(t, proj, "100", "Move docs", "abandoned")

		// Guard should block (finalization requires completed)
		can, err := machine.CanFire(sdkstate.Event(EventCompleteFinalization))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with abandoned tasks")
	})
}

// TestStateValidation tests that project state validates correctly at each stage.
//nolint:funlen // Test contains multiple validation subtests
func TestStateValidation(t *testing.T) {
	t.Run("design phase status updates correctly", func(t *testing.T) {
		proj, machine, config := setupDesignProject(t)

		// Initial: active
		verifyPhaseStatus(t, proj, "design", "active")

		// After advancing to Finalizing: completed
		addDocumentTask(t, proj, "010", "Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")

		err := config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		require.NoError(t, err)

		verifyPhaseStatus(t, proj, "design", "completed")
	})

	t.Run("finalization phase enabled at correct time", func(t *testing.T) {
		proj, machine, config := setupDesignProject(t)

		// Initially: disabled
		verifyPhaseEnabled(t, proj, "finalization", false)
		verifyPhaseStatus(t, proj, "finalization", "pending")

		// After entering Finalizing state: enabled and in_progress
		addDocumentTask(t, proj, "010", "Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")

		err := config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		require.NoError(t, err)

		verifyPhaseEnabled(t, proj, "finalization", true)
		verifyPhaseStatus(t, proj, "finalization", "in_progress")
	})

	t.Run("timestamps set correctly", func(t *testing.T) {
		proj, machine, config := setupDesignProject(t)

		// Design phase created_at should be set
		designPhase := proj.Phases["design"]
		assert.False(t, designPhase.Created_at.IsZero(), "design created_at should be set")

		// Complete workflow
		addDocumentTask(t, proj, "010", "Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")

		err := config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		require.NoError(t, err)

		// Design completed_at should be set
		designPhase = proj.Phases["design"]
		assert.False(t, designPhase.Completed_at.IsZero(), "design completed_at should be set")

		// Finalization started_at should be set
		finPhase := proj.Phases["finalization"]
		assert.False(t, finPhase.Started_at.IsZero(), "finalization started_at should be set")

		// Complete finalization
		addFinalizationTask(t, proj, "100", "Move docs", "completed")
		err = config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteFinalization), proj)
		require.NoError(t, err)

		// Finalization completed_at should be set
		finPhase = proj.Phases["finalization"]
		assert.False(t, finPhase.Completed_at.IsZero(), "finalization completed_at should be set")
	})

	t.Run("phase completion markers", func(t *testing.T) {
		proj, machine, config := setupDesignProject(t)

		// Complete full lifecycle
		addDocumentTask(t, proj, "010", "Doc", "design", ".sow/knowledge/designs/doc.md")
		artifact := addDesignArtifact(t, proj, "project/design.md", "design")
		linkArtifactToTask(t, proj, "010", artifact)
		markTaskCompleted(t, proj, "010")

		err := config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteDesign), proj)
		require.NoError(t, err)

		addFinalizationTask(t, proj, "100", "Move docs", "completed")
		err = config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteFinalization), proj)
		require.NoError(t, err)

		// Both phases should be marked completed
		verifyPhaseStatus(t, proj, "design", "completed")
		verifyPhaseStatus(t, proj, "finalization", "completed")
	})
}

// Helper functions

// setupDesignProject creates a project for testing in Active state.
func setupDesignProject(t *testing.T) (*state.Project, *sdkstate.Machine, *project.ProjectTypeConfig) {
	t.Helper()

	// Create project
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:        "test-design",
			Branch:      "design/test",
			Type:        "design",
			Description: "Test design project",
			Created_at:  now,
			Updated_at:  now,
			Phases:      make(map[string]projschema.PhaseState),
			Statechart: projschema.StatechartState{
				Current_state: string(Active),
				Updated_at:    now,
			},
		},
	}

	// Initialize phases using the config's initializer
	config := NewDesignProjectConfig()
	err := initializeDesignProject(proj, nil)
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Build state machine
	machine := config.BuildMachine(proj, sdkstate.State(Active))
	if machine == nil {
		t.Fatal("BuildMachine returned nil")
	}

	return proj, machine, config
}

// addDocumentTask adds a document task to the design phase.
func addDocumentTask(t *testing.T, p *state.Project, id, name, docType, targetPath string) {
	t.Helper()

	phase, exists := p.Phases["design"]
	if !exists {
		t.Fatal("design phase not found")
	}

	task := projschema.TaskState{
		Id:         id,
		Name:       name,
		Phase:      "design",
		Status:     "pending",
		Created_at: time.Now(),
		Updated_at: time.Now(),
		Inputs:     []projschema.ArtifactState{},
		Metadata: map[string]interface{}{
			"document_type": docType,
			"target_path":   targetPath,
		},
	}

	phase.Tasks = append(phase.Tasks, task)
	p.Phases["design"] = phase
}

// markTaskInProgress marks a task as in_progress.
func markTaskInProgress(t *testing.T, p *state.Project, taskID string) {
	t.Helper()

	phase, exists := p.Phases["design"]
	if !exists {
		t.Fatal("design phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			phase.Tasks[i].Status = "in_progress"
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["design"] = phase
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// markTaskNeedsReview marks a task as needs_review.
func markTaskNeedsReview(t *testing.T, p *state.Project, taskID string) {
	t.Helper()

	phase, exists := p.Phases["design"]
	if !exists {
		t.Fatal("design phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			phase.Tasks[i].Status = "needs_review"
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["design"] = phase
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// markTaskCompleted marks a task as completed and auto-approves its artifact (design tasks only).
func markTaskCompleted(t *testing.T, p *state.Project, taskID string) {
	t.Helper()

	phase, exists := p.Phases["design"]
	if !exists {
		t.Fatal("design phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			phase.Tasks[i].Status = "completed"
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["design"] = phase

			// Auto-approve artifact (only for design tasks)
			err := autoApproveArtifact(p, taskID)
			if err != nil {
				t.Fatalf("failed to auto-approve artifact: %v", err)
			}
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// markFinalizationTaskCompleted marks a finalization task as completed.
func markFinalizationTaskCompleted(t *testing.T, p *state.Project, taskID string) {
	t.Helper()

	phase, exists := p.Phases["finalization"]
	if !exists {
		t.Fatal("finalization phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			phase.Tasks[i].Status = "completed"
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["finalization"] = phase
			return
		}
	}

	t.Fatalf("task %s not found in finalization phase", taskID)
}

// markTaskAbandoned marks a task as abandoned.
func markTaskAbandoned(t *testing.T, p *state.Project, taskID string) {
	t.Helper()

	phase, exists := p.Phases["design"]
	if !exists {
		t.Fatal("design phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			phase.Tasks[i].Status = "abandoned"
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["design"] = phase
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// addDesignArtifact adds an artifact to design phase outputs.
func addDesignArtifact(t *testing.T, p *state.Project, path, docType string) string {
	t.Helper()

	phase, exists := p.Phases["design"]
	if !exists {
		t.Fatal("design phase not found")
	}

	artifact := projschema.ArtifactState{
		Type:       docType,
		Path:       path,
		Created_at: time.Now(),
		Approved:   false,
	}

	phase.Outputs = append(phase.Outputs, artifact)
	p.Phases["design"] = phase

	return path
}

// linkArtifactToTask links an artifact to a task via metadata.
func linkArtifactToTask(t *testing.T, p *state.Project, taskID, artifactPath string) {
	t.Helper()

	phase, exists := p.Phases["design"]
	if !exists {
		t.Fatal("design phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			if phase.Tasks[i].Metadata == nil {
				phase.Tasks[i].Metadata = make(map[string]interface{})
			}
			phase.Tasks[i].Metadata["artifact_path"] = artifactPath
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["design"] = phase
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// verifyArtifactApproved verifies the approval status of an artifact.
func verifyArtifactApproved(t *testing.T, p *state.Project, artifactPath string, expectedApproved bool) {
	t.Helper()

	phase, exists := p.Phases["design"]
	if !exists {
		t.Fatal("design phase not found")
	}

	for _, artifact := range phase.Outputs {
		if artifact.Path == artifactPath {
			if artifact.Approved != expectedApproved {
				t.Errorf("artifact %s approval: got %v, want %v", artifactPath, artifact.Approved, expectedApproved)
			}
			return
		}
	}

	t.Fatalf("artifact %s not found", artifactPath)
}

// verifyPhaseStatus verifies the status of a phase.
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

// verifyPhaseEnabled verifies the enabled flag of a phase.
func verifyPhaseEnabled(t *testing.T, p *state.Project, phaseName string, expectedEnabled bool) {
	t.Helper()

	phase, exists := p.Phases[phaseName]
	if !exists {
		t.Fatalf("phase %s not found", phaseName)
	}

	if phase.Enabled != expectedEnabled {
		t.Errorf("phase %s enabled: got %v, want %v", phaseName, phase.Enabled, expectedEnabled)
	}
}

// verifyTaskStatus verifies the status of a task.
//nolint:unparam // phaseName parameter kept for consistency with test helper pattern
func verifyTaskStatus(t *testing.T, p *state.Project, phaseName, taskID, expectedStatus string) {
	t.Helper()

	phase, exists := p.Phases[phaseName]
	if !exists {
		t.Fatalf("phase %s not found", phaseName)
	}

	for _, task := range phase.Tasks {
		if task.Id == taskID {
			if task.Status != expectedStatus {
				t.Errorf("task %s status: got %v, want %v", taskID, task.Status, expectedStatus)
			}
			return
		}
	}

	t.Fatalf("task %s not found in phase %s", taskID, phaseName)
}

// addFinalizationTask adds a task to the finalization phase.
//nolint:unparam // id parameter kept for consistency with test helper pattern
func addFinalizationTask(t *testing.T, p *state.Project, id, name, status string) {
	t.Helper()

	phase, exists := p.Phases["finalization"]
	if !exists {
		t.Fatal("finalization phase not found")
	}

	task := projschema.TaskState{
		Id:         id,
		Name:       name,
		Phase:      "finalization",
		Status:     status,
		Created_at: time.Now(),
		Updated_at: time.Now(),
		Inputs:     []projschema.ArtifactState{},
		Metadata:   make(map[string]interface{}),
	}

	phase.Tasks = append(phase.Tasks, task)
	p.Phases["finalization"] = phase
}
