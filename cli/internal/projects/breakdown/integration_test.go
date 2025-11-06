package breakdown

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBreakdownLifecycle_HappyPath tests the complete breakdown workflow from creation to completion.
// This is the primary happy path test covering all states: Discovery → Active → Publishing → Completed.
//nolint:funlen // Test contains multiple subtests for lifecycle verification
func TestBreakdownLifecycle_HappyPath(t *testing.T) {
	// Setup: Create project and state machine
	proj, machine := setupBreakdownProject(t)

	// Phase 0: Discovery - Add discovery document and transition to Active
	t.Run("discovery phase", func(t *testing.T) {
		// Verify initial state is Discovery
		assert.Equal(t, sdkstate.State(Discovery), machine.State(), "initial state should be Discovery")
		verifyPhaseStatus(t, proj, "discovery")

		// Add and approve discovery document
		addDiscoveryArtifact(t, proj, "project/discovery/analysis.md")

		// Transition to Active
		err := machine.Fire(sdkstate.Event(EventBeginActive))
		require.NoError(t, err, "transition to Active should succeed")
		assert.Equal(t, sdkstate.State(Active), machine.State(), "state should be Active")
		verifyPhaseStatus(t, proj, "active")
	})

	// Phase 1: Active - Create and specify work units
	t.Run("create and specify work units", func(t *testing.T) {
		// Verify we're in Active state
		assert.Equal(t, sdkstate.State(Active), machine.State(), "state should be Active")
		verifyPhaseEnabled(t, proj, "breakdown", true)

		// Create work unit tasks
		addWorkUnit(t, proj, "001", "Feature A", "pending")
		addWorkUnit(t, proj, "002", "Feature B", "pending")

		// Link specifications to tasks
		artifact1 := addWorkUnitArtifact(t, proj, "project/specs/feature-a.md")
		linkArtifactToTask(t, proj, "001", artifact1)
		artifact2 := addWorkUnitArtifact(t, proj, "project/specs/feature-b.md")
		linkArtifactToTask(t, proj, "002", artifact2)

		// Complete tasks (should auto-approve artifacts)
		markTaskCompleted(t, proj, "001")
		markTaskCompleted(t, proj, "002")

		// Verify task completion and artifact approval
		verifyTaskStatus(t, proj, "breakdown", "001", "completed")
		verifyTaskStatus(t, proj, "breakdown", "002", "completed")
		verifyArtifactApproved(t, proj, artifact1, true)
		verifyArtifactApproved(t, proj, artifact2, true)
	})

	// Phase 2: Advance to Publishing
	t.Run("advance to publishing", func(t *testing.T) {
		// Verify guard passes
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err, "CanFire should not error")
		assert.True(t, can, "guard should allow transition with completed work units")

		// Fire event
		err = machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err, "Fire should not error")

		// Verify state transition
		assert.Equal(t, sdkstate.State(Publishing), machine.State(), "state should be Publishing")

		// Verify breakdown phase status
		verifyPhaseStatus(t, proj, "publishing")
	})

	// Phase 3: Publish work units
	t.Run("publish work units to GitHub", func(t *testing.T) {
		// Cannot complete yet - work units not published
		can, err := machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.False(t, can, "guard should block when work units not published")

		// Publish work units
		publishWorkUnit(t, proj, "001", 123, "https://github.com/org/repo/issues/123")
		publishWorkUnit(t, proj, "002", 124, "https://github.com/org/repo/issues/124")

		// Now should be able to complete
		can, err = machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow completion after all units published")
	})

	// Phase 4: Complete breakdown
	t.Run("complete breakdown", func(t *testing.T) {
		// Fire completion event
		err := machine.Fire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err, "Fire should not error")

		// Verify state transition to Completed
		assert.Equal(t, sdkstate.State(Completed), machine.State(), "state should be Completed")

		// Verify breakdown phase completed
		verifyPhaseStatus(t, proj, "completed")
		phase := proj.Phases["breakdown"]
		assert.False(t, phase.Completed_at.IsZero(), "breakdown phase should have completion time")
	})

	// Verify final state
	t.Run("verify final state", func(t *testing.T) {
		// Phase should be completed
		verifyPhaseStatus(t, proj, "completed")

		// State machine should be in Completed state
		assert.Equal(t, sdkstate.State(Completed), machine.State(), "final state should be Completed")

		// Verify all timestamps set
		phase := proj.Phases["breakdown"]
		assert.False(t, phase.Created_at.IsZero(), "breakdown created_at should be set")
		assert.False(t, phase.Completed_at.IsZero(), "breakdown completed_at should be set")

		// Verify all work units published
		for _, task := range phase.Tasks {
			if task.Status == "completed" {
				assert.NotNil(t, task.Metadata, "completed task should have metadata")
				published, ok := task.Metadata["published"].(bool)
				assert.True(t, ok, "published field should be a bool")
				assert.True(t, published, "completed task should be published")
				assert.NotZero(t, task.Metadata["github_issue_number"], "completed task should have issue number")
				assert.NotEmpty(t, task.Metadata["github_issue_url"], "completed task should have issue URL")
			}
		}
	})
}

// TestReviewWorkflow_BackAndForth tests work unit review workflow with status transitions.
//nolint:funlen // Test contains multiple subtests for review workflow verification
func TestReviewWorkflow_BackAndForth(t *testing.T) {
	proj, machine := setupBreakdownProject(t)
	transitionToActive(t, proj, machine)

	t.Run("create and draft work unit", func(t *testing.T) {
		addWorkUnit(t, proj, "001", "Feature A", "pending")
		artifact := addWorkUnitArtifact(t, proj, "project/specs/feature-a.md")
		linkArtifactToTask(t, proj, "001", artifact)

		verifyTaskStatus(t, proj, "breakdown", "001", "pending")
		verifyArtifactApproved(t, proj, artifact, false)
	})

	t.Run("start work and request review", func(t *testing.T) {
		markTaskInProgress(t, proj, "001")
		verifyTaskStatus(t, proj, "breakdown", "001", "in_progress")

		// Request review
		markTaskNeedsReview(t, proj, "001")
		verifyTaskStatus(t, proj, "breakdown", "001", "needs_review")

		// Cannot advance with needs_review task
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.False(t, can, "should not advance with task in review")
	})

	t.Run("revise after feedback", func(t *testing.T) {
		// Go back to in_progress for revisions
		markTaskInProgress(t, proj, "001")
		verifyTaskStatus(t, proj, "breakdown", "001", "in_progress")

		// Still cannot advance
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.False(t, can, "should not advance with in_progress task")

		// Request review again
		markTaskNeedsReview(t, proj, "001")
		verifyTaskStatus(t, proj, "breakdown", "001", "needs_review")
	})

	t.Run("approve and complete", func(t *testing.T) {
		markTaskCompleted(t, proj, "001")
		verifyTaskStatus(t, proj, "breakdown", "001", "completed")

		// Verify auto-approval
		phase := proj.Phases["breakdown"]
		var artifact *projschema.ArtifactState
		for i := range phase.Outputs {
			if phase.Outputs[i].Path == "project/specs/feature-a.md" {
				artifact = &phase.Outputs[i]
				break
			}
		}
		require.NotNil(t, artifact, "artifact should exist")
		assert.True(t, artifact.Approved, "artifact should be auto-approved on completion")

		// Can now advance
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.True(t, can, "should be able to advance after completion")
	})
}

// TestDependencyValidation_Cycles tests that cyclic dependencies block advancement.
func TestDependencyValidation_Cycles(t *testing.T) {
	proj, machine := setupBreakdownProject(t)
	transitionToActive(t, proj, machine)

	t.Run("create tasks with cyclic dependencies", func(t *testing.T) {
		// Create cycle: 001 → 002 → 003 → 001
		addWorkUnitWithDeps(t, proj, "001", "Task A", "completed", []string{"003"})
		addWorkUnitWithDeps(t, proj, "002", "Task B", "completed", []string{"001"})
		addWorkUnitWithDeps(t, proj, "003", "Task C", "completed", []string{"002"})

		// Link artifacts
		artifact1 := addWorkUnitArtifact(t, proj, "project/specs/task-a.md")
		linkArtifactToTask(t, proj, "001", artifact1)
		artifact2 := addWorkUnitArtifact(t, proj, "project/specs/task-b.md")
		linkArtifactToTask(t, proj, "002", artifact2)
		artifact3 := addWorkUnitArtifact(t, proj, "project/specs/task-c.md")
		linkArtifactToTask(t, proj, "003", artifact3)
	})

	t.Run("advancement blocked by cyclic dependencies", func(t *testing.T) {
		// Guard should block due to cycle
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with cyclic dependencies")

		// Verify state unchanged
		assert.Equal(t, sdkstate.State(Active), machine.State(), "state should remain Active")
	})
}

// TestDependencyValidation_InvalidRefs tests that invalid dependency references block advancement.
func TestDependencyValidation_InvalidRefs(t *testing.T) {
	proj, machine := setupBreakdownProject(t)
	transitionToActive(t, proj, machine)

	t.Run("create tasks with invalid dependency references", func(t *testing.T) {
		// Create task that references non-existent task
		addWorkUnitWithDeps(t, proj, "001", "Task A", "completed", []string{"999"}) // 999 doesn't exist

		// Link artifact
		artifact := addWorkUnitArtifact(t, proj, "project/specs/task-a.md")
		linkArtifactToTask(t, proj, "001", artifact)
	})

	t.Run("advancement blocked by invalid references", func(t *testing.T) {
		// Guard should block due to invalid reference
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with invalid dependency references")

		// Verify state unchanged
		assert.Equal(t, sdkstate.State(Active), machine.State(), "state should remain Active")
	})
}

// TestDependencyValidation_ValidChain tests that valid dependency chains work correctly.
func TestDependencyValidation_ValidChain(t *testing.T) {
	proj, machine := setupBreakdownProject(t)
	transitionToActive(t, proj, machine)

	t.Run("create tasks with valid dependency chain", func(t *testing.T) {
		// Create valid chain: 001 → 002 → 003
		addWorkUnit(t, proj, "001", "Base", "completed")
		addWorkUnitWithDeps(t, proj, "002", "Middle", "completed", []string{"001"})
		addWorkUnitWithDeps(t, proj, "003", "Top", "completed", []string{"002"})

		// Link artifacts
		artifact1 := addWorkUnitArtifact(t, proj, "project/specs/base.md")
		linkArtifactToTask(t, proj, "001", artifact1)
		artifact2 := addWorkUnitArtifact(t, proj, "project/specs/middle.md")
		linkArtifactToTask(t, proj, "002", artifact2)
		artifact3 := addWorkUnitArtifact(t, proj, "project/specs/top.md")
		linkArtifactToTask(t, proj, "003", artifact3)
	})

	t.Run("advancement allowed with valid dependencies", func(t *testing.T) {
		// Guard should pass
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow transition with valid dependencies")

		// Successfully advance
		err = machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.Equal(t, sdkstate.State(Publishing), machine.State(), "should advance to Publishing")
	})
}

// TestPublishing_Resumability tests that publishing can be interrupted and resumed.
//nolint:funlen // Test contains multiple subtests for resumability verification
func TestPublishing_Resumability(t *testing.T) {
	proj, machine := setupBreakdownProject(t)
	transitionToActive(t, proj, machine)

	t.Run("setup work units", func(t *testing.T) {
		// Create 3 work units
		addWorkUnit(t, proj, "001", "Feature A", "completed")
		addWorkUnit(t, proj, "002", "Feature B", "completed")
		addWorkUnit(t, proj, "003", "Feature C", "completed")

		// Link artifacts
		artifact1 := addWorkUnitArtifact(t, proj, "project/specs/a.md")
		linkArtifactToTask(t, proj, "001", artifact1)
		artifact2 := addWorkUnitArtifact(t, proj, "project/specs/b.md")
		linkArtifactToTask(t, proj, "002", artifact2)
		artifact3 := addWorkUnitArtifact(t, proj, "project/specs/c.md")
		linkArtifactToTask(t, proj, "003", artifact3)

		// Advance to publishing
		err := machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.Equal(t, sdkstate.State(Publishing), machine.State())
	})

	t.Run("publish some work units (simulate interruption)", func(t *testing.T) {
		// Publish only first task
		publishWorkUnit(t, proj, "001", 123, "https://github.com/org/repo/issues/123")

		// Cannot complete yet - unpublished tasks remain
		can, err := machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.False(t, can, "guard should block when unpublished tasks remain")
	})

	t.Run("resume publishing", func(t *testing.T) {
		// Verify task 001 is already published (would be skipped in real workflow)
		phase := proj.Phases["breakdown"]
		var task001Published bool
		for _, task := range phase.Tasks {
			if task.Id == "001" {
				if task.Metadata != nil {
					if published, ok := task.Metadata["published"].(bool); ok && published {
						task001Published = true
					}
				}
			}
		}
		assert.True(t, task001Published, "task 001 should remain published")

		// Publish remaining tasks
		publishWorkUnit(t, proj, "002", 124, "https://github.com/org/repo/issues/124")
		publishWorkUnit(t, proj, "003", 125, "https://github.com/org/repo/issues/125")

		// Now can complete
		can, err := machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow completion after all units published")

		err = machine.Fire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.Equal(t, sdkstate.State(Completed), machine.State())
	})
}

// TestBreakdownIntegration_DiamondDependencies tests complex dependency graphs.
//nolint:funlen // Test contains multiple subtests for dependency verification
func TestBreakdownIntegration_DiamondDependencies(t *testing.T) {
	proj, machine := setupBreakdownProject(t)
	transitionToActive(t, proj, machine)

	t.Run("create work units with diamond dependency", func(t *testing.T) {
		// Diamond: 001 → 002, 003 → 004
		//         001 → 003
		addWorkUnit(t, proj, "001", "Base", "completed")
		addWorkUnitWithDeps(t, proj, "002", "Branch A", "completed", []string{"001"})
		addWorkUnitWithDeps(t, proj, "003", "Branch B", "completed", []string{"001"})
		addWorkUnitWithDeps(t, proj, "004", "Merge", "completed", []string{"002", "003"})

		// Link artifacts
		artifact1 := addWorkUnitArtifact(t, proj, "project/specs/base.md")
		linkArtifactToTask(t, proj, "001", artifact1)
		artifact2 := addWorkUnitArtifact(t, proj, "project/specs/branch-a.md")
		linkArtifactToTask(t, proj, "002", artifact2)
		artifact3 := addWorkUnitArtifact(t, proj, "project/specs/branch-b.md")
		linkArtifactToTask(t, proj, "003", artifact3)
		artifact4 := addWorkUnitArtifact(t, proj, "project/specs/merge.md")
		linkArtifactToTask(t, proj, "004", artifact4)
	})

	t.Run("advance to publishing with valid diamond", func(t *testing.T) {
		// Guard should pass - diamond is valid DAG
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow transition with valid diamond dependency")

		err = machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.Equal(t, sdkstate.State(Publishing), machine.State())
	})

	t.Run("publish all units", func(t *testing.T) {
		// Publish in dependency order (though order doesn't affect test)
		publishWorkUnit(t, proj, "001", 201, "https://github.com/org/repo/issues/201")
		publishWorkUnit(t, proj, "002", 202, "https://github.com/org/repo/issues/202")
		publishWorkUnit(t, proj, "003", 203, "https://github.com/org/repo/issues/203")
		publishWorkUnit(t, proj, "004", 204, "https://github.com/org/repo/issues/204")

		// Complete breakdown
		err := machine.Fire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.Equal(t, sdkstate.State(Completed), machine.State())
	})
}

// TestBreakdownLifecycle_WithAbandoned tests mix of completed and abandoned tasks.
func TestBreakdownLifecycle_WithAbandoned(t *testing.T) {
	proj, machine := setupBreakdownProject(t)
	transitionToActive(t, proj, machine)

	t.Run("create and complete some tasks, abandon others", func(t *testing.T) {
		// Complete first two tasks
		addWorkUnit(t, proj, "001", "Feature A", "completed")
		artifact1 := addWorkUnitArtifact(t, proj, "project/specs/a.md")
		linkArtifactToTask(t, proj, "001", artifact1)

		addWorkUnit(t, proj, "002", "Feature B", "completed")
		artifact2 := addWorkUnitArtifact(t, proj, "project/specs/b.md")
		linkArtifactToTask(t, proj, "002", artifact2)

		// Abandon third task
		addWorkUnit(t, proj, "003", "Feature C", "abandoned")

		// Verify statuses
		verifyTaskStatus(t, proj, "breakdown", "001", "completed")
		verifyTaskStatus(t, proj, "breakdown", "002", "completed")
		verifyTaskStatus(t, proj, "breakdown", "003", "abandoned")
	})

	t.Run("advance with mix of completed and abandoned", func(t *testing.T) {
		// Should succeed - has completed tasks
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow transition with completed tasks")

		err = machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.Equal(t, sdkstate.State(Publishing), machine.State())
	})

	t.Run("publish only completed tasks", func(t *testing.T) {
		// Publish only completed tasks (abandoned not published)
		publishWorkUnit(t, proj, "001", 301, "https://github.com/org/repo/issues/301")
		publishWorkUnit(t, proj, "002", 302, "https://github.com/org/repo/issues/302")

		// Complete breakdown
		err := machine.Fire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.Equal(t, sdkstate.State(Completed), machine.State())

		// Verify abandoned task not published
		phase := proj.Phases["breakdown"]
		for _, task := range phase.Tasks {
			if task.Id == "003" {
				assert.Equal(t, "abandoned", task.Status)
				if task.Metadata != nil {
					assert.NotEqual(t, true, task.Metadata["published"])
				}
			}
		}
	})
}

// TestBreakdownLifecycle_AllAbandoned tests that advancement blocked when all tasks abandoned.
func TestBreakdownLifecycle_AllAbandoned(t *testing.T) {
	proj, machine := setupBreakdownProject(t)
	transitionToActive(t, proj, machine)

	t.Run("create and abandon all tasks", func(t *testing.T) {
		// Abandon all tasks
		addWorkUnit(t, proj, "001", "Feature A", "abandoned")
		addWorkUnit(t, proj, "002", "Feature B", "abandoned")
		addWorkUnit(t, proj, "003", "Feature C", "abandoned")

		// Verify all abandoned
		verifyTaskStatus(t, proj, "breakdown", "001", "abandoned")
		verifyTaskStatus(t, proj, "breakdown", "002", "abandoned")
		verifyTaskStatus(t, proj, "breakdown", "003", "abandoned")
	})

	t.Run("cannot advance with all abandoned", func(t *testing.T) {
		// Guard should block - needs at least one completed
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.False(t, can, "guard should block when all tasks abandoned")

		// Try to fire anyway - should fail
		err = machine.Fire(sdkstate.Event(EventBeginPublishing))
		assert.Error(t, err, "should error when trying to fire blocked transition")

		// Verify state unchanged
		assert.Equal(t, sdkstate.State(Active), machine.State(), "state should remain Active")
	})
}

// TestBreakdownLifecycle_NoTasks tests that advancement blocked when no tasks exist.
func TestBreakdownLifecycle_NoTasks(t *testing.T) {
	proj, machine := setupBreakdownProject(t)
	transitionToActive(t, proj, machine)

	t.Run("cannot advance without tasks", func(t *testing.T) {
		// No tasks created
		phase := proj.Phases["breakdown"]
		assert.Empty(t, phase.Tasks, "should have no tasks")

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.False(t, can, "guard should block with no tasks")

		// Verify state unchanged
		assert.Equal(t, sdkstate.State(Active), machine.State(), "state should remain Active")
	})
}

// TestBreakdownLifecycle_WithInputs tests breakdown project with initial input artifacts.
func TestBreakdownLifecycle_WithInputs(t *testing.T) {
	t.Run("create project with inputs", func(t *testing.T) {
		// Create project with inputs
		now := time.Now()
		proj := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:       "test-breakdown-with-inputs",
				Branch:     "breakdown/test-inputs",
				Type:       "breakdown",
				Created_at: now,
				Updated_at: now,
				Phases:     make(map[string]projschema.PhaseState),
			},
		}

		// Initialize with inputs
		initialInputs := map[string][]projschema.ArtifactState{
			"breakdown": {
				{
					Type:       "design",
					Path:       ".sow/knowledge/designs/feature.md",
					Created_at: now,
					Approved:   true,
				},
			},
		}

		config := NewBreakdownProjectConfig()
		err := config.Initialize(proj, initialInputs)
		require.NoError(t, err)

		// Build state machine (starts in Discovery)
		machine := config.BuildMachine(proj, sdkstate.State(Discovery))
		require.NotNil(t, machine)

		// Verify inputs are tracked separately
		phase := proj.Phases["breakdown"]
		assert.Len(t, phase.Inputs, 1, "should have 1 input")
		assert.Equal(t, "design", phase.Inputs[0].Type)
		assert.Empty(t, phase.Outputs, "outputs should be empty initially")

		// Transition to Active
		transitionToActive(t, proj, machine)

		// Complete workflow
		addWorkUnit(t, proj, "001", "Implementation", "completed")
		artifact := addWorkUnitArtifact(t, proj, "project/specs/impl.md")
		linkArtifactToTask(t, proj, "001", artifact)

		// Inputs don't block progression
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.True(t, can, "inputs should not block progression")
	})
}

// TestTaskValidation tests validateTaskForCompletion enforcement.
func TestTaskValidation(t *testing.T) {
	proj, _ := setupBreakdownProject(t)

	t.Run("completion blocked without artifact", func(t *testing.T) {
		// Create task without artifact
		addWorkUnit(t, proj, "001", "Feature A", "pending")

		// Try to validate for completion - should fail
		err := validateTaskForCompletion(proj, "001")
		assert.Error(t, err, "should error when completing task without artifact")
		assert.Contains(t, err.Error(), "metadata", "error should mention metadata")
	})

	t.Run("completion blocked with artifact_path but no artifact", func(t *testing.T) {
		// Link task to non-existent artifact
		phase := proj.Phases["breakdown"]
		for i := range phase.Tasks {
			if phase.Tasks[i].Id == "001" {
				phase.Tasks[i].Metadata = map[string]interface{}{
					"artifact_path": "project/nonexistent.md",
				}
			}
		}
		proj.Phases["breakdown"] = phase

		// Try to validate - should fail
		err := validateTaskForCompletion(proj, "001")
		assert.Error(t, err, "should error when artifact doesn't exist")
		assert.Contains(t, err.Error(), "artifact not found", "error should mention missing artifact")
	})

	t.Run("completion allowed with valid artifact", func(t *testing.T) {
		// Add artifact
		artifact := addWorkUnitArtifact(t, proj, "project/valid-spec.md")
		linkArtifactToTask(t, proj, "001", artifact)

		// Should validate successfully
		err := validateTaskForCompletion(proj, "001")
		assert.NoError(t, err, "should allow completion with valid artifact")
	})
}

// TestAutoApproval tests automatic artifact approval on task completion.
func TestAutoApproval(t *testing.T) {
	proj, _ := setupBreakdownProject(t)

	t.Run("artifact auto-approved on task completion", func(t *testing.T) {
		// Create task and artifact
		addWorkUnit(t, proj, "001", "Feature A", "pending")
		artifact := addWorkUnitArtifact(t, proj, "project/spec-a.md")
		linkArtifactToTask(t, proj, "001", artifact)

		// Initially not approved
		verifyArtifactApproved(t, proj, artifact, false)

		// Complete task (should auto-approve)
		markTaskCompleted(t, proj, "001")

		// Verify auto-approval
		verifyArtifactApproved(t, proj, artifact, true)
	})

	t.Run("auto-approval is atomic with completion", func(t *testing.T) {
		// Create second task and artifact
		addWorkUnit(t, proj, "002", "Feature B", "pending")
		artifact := addWorkUnitArtifact(t, proj, "project/spec-b.md")
		linkArtifactToTask(t, proj, "002", artifact)

		// Complete task
		markTaskCompleted(t, proj, "002")

		// Approval should happen immediately
		phase := proj.Phases["breakdown"]
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
	t.Run("Active to Publishing blocked with pending tasks", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)
		transitionToActive(t, proj, machine)

		// Add pending task
		addWorkUnit(t, proj, "001", "Feature A", "pending")

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with pending tasks")
	})

	t.Run("Active to Publishing blocked with in_progress tasks", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)
		transitionToActive(t, proj, machine)

		// Add in_progress task
		addWorkUnit(t, proj, "001", "Feature A", "in_progress")

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with in_progress tasks")
	})

	t.Run("Active to Publishing blocked with needs_review tasks", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)
		transitionToActive(t, proj, machine)

		// Add needs_review task
		addWorkUnit(t, proj, "001", "Feature A", "needs_review")
		artifact := addWorkUnitArtifact(t, proj, "project/spec.md")
		linkArtifactToTask(t, proj, "001", artifact)

		// Guard should block
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with needs_review tasks")
	})

	t.Run("Active to Publishing allowed with completed tasks", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)
		transitionToActive(t, proj, machine)

		// Add completed task
		addWorkUnit(t, proj, "001", "Feature A", "completed")
		artifact := addWorkUnitArtifact(t, proj, "project/spec.md")
		linkArtifactToTask(t, proj, "001", artifact)

		// Guard should allow
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow transition with completed tasks")
	})

	t.Run("Active to Publishing allowed with mix of completed and abandoned", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)
		transitionToActive(t, proj, machine)

		// Add completed task
		addWorkUnit(t, proj, "001", "Feature A", "completed")
		artifact := addWorkUnitArtifact(t, proj, "project/spec-a.md")
		linkArtifactToTask(t, proj, "001", artifact)

		// Add abandoned task
		addWorkUnit(t, proj, "002", "Feature B", "abandoned")

		// Guard should allow
		can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow transition with completed and abandoned tasks")
	})

	t.Run("Publishing to Completed blocked with unpublished tasks", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)
		transitionToActive(t, proj, machine)

		// Setup: advance to Publishing
		addWorkUnit(t, proj, "001", "Feature A", "completed")
		artifact := addWorkUnitArtifact(t, proj, "project/spec.md")
		linkArtifactToTask(t, proj, "001", artifact)
		err := machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)

		// Task not yet published - guard should block
		can, err := machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.False(t, can, "guard should block transition with unpublished tasks")
	})

	t.Run("Publishing to Completed allowed after publishing all", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)
		transitionToActive(t, proj, machine)

		// Setup: advance to Publishing
		addWorkUnit(t, proj, "001", "Feature A", "completed")
		artifact := addWorkUnitArtifact(t, proj, "project/spec.md")
		linkArtifactToTask(t, proj, "001", artifact)
		err := machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)

		// Publish task
		publishWorkUnit(t, proj, "001", 123, "https://github.com/org/repo/issues/123")

		// Guard should allow
		can, err := machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.True(t, can, "guard should allow transition after all tasks published")

		err = machine.Fire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)
		assert.Equal(t, sdkstate.State(Completed), machine.State())
	})
}

// TestStateValidation tests that project state validates correctly at each stage.
//nolint:funlen // Test contains multiple validation subtests
func TestStateValidation(t *testing.T) {
	t.Run("breakdown phase status updates correctly", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)

		// Initial: discovery
		verifyPhaseStatus(t, proj, "discovery")

		// After transitioning to Active: active
		// Add discovery artifact and approve it
		addDiscoveryArtifact(t, proj, "project/discovery/analysis.md")
		err := machine.Fire(sdkstate.Event(EventBeginActive))
		require.NoError(t, err)
		verifyPhaseStatus(t, proj, "active")

		// After advancing to Publishing: publishing
		addWorkUnit(t, proj, "001", "Feature", "completed")
		artifact := addWorkUnitArtifact(t, proj, "project/spec.md")
		linkArtifactToTask(t, proj, "001", artifact)

		err = machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)

		verifyPhaseStatus(t, proj, "publishing")

		// After completing: completed
		publishWorkUnit(t, proj, "001", 123, "https://github.com/org/repo/issues/123")
		err = machine.Fire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)

		verifyPhaseStatus(t, proj, "completed")
	})

	t.Run("timestamps set correctly", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)
		transitionToActive(t, proj, machine)

		// Breakdown created_at should be set
		phase := proj.Phases["breakdown"]
		assert.False(t, phase.Created_at.IsZero(), "breakdown created_at should be set")

		// Complete workflow
		addWorkUnit(t, proj, "001", "Feature", "completed")
		artifact := addWorkUnitArtifact(t, proj, "project/spec.md")
		linkArtifactToTask(t, proj, "001", artifact)

		err := machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)

		publishWorkUnit(t, proj, "001", 123, "https://github.com/org/repo/issues/123")
		err = machine.Fire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)

		// Breakdown completed_at should be set
		phase = proj.Phases["breakdown"]
		assert.False(t, phase.Completed_at.IsZero(), "breakdown completed_at should be set after completion")
	})

	t.Run("phase completion markers", func(t *testing.T) {
		proj, machine := setupBreakdownProject(t)
		transitionToActive(t, proj, machine)

		// Complete full lifecycle
		addWorkUnit(t, proj, "001", "Feature", "completed")
		artifact := addWorkUnitArtifact(t, proj, "project/spec.md")
		linkArtifactToTask(t, proj, "001", artifact)

		err := machine.Fire(sdkstate.Event(EventBeginPublishing))
		require.NoError(t, err)

		publishWorkUnit(t, proj, "001", 123, "https://github.com/org/repo/issues/123")
		err = machine.Fire(sdkstate.Event(EventCompleteBreakdown))
		require.NoError(t, err)

		// Phase should be marked completed
		verifyPhaseStatus(t, proj, "completed")
	})
}

// Helper functions (integration test specific)

// addWorkUnitArtifact adds a work unit specification artifact to breakdown outputs.
// addDiscoveryArtifact adds an approved discovery artifact to the breakdown phase.
func addDiscoveryArtifact(t *testing.T, p *state.Project, path string) {
	t.Helper()

	phase, exists := p.Phases["breakdown"]
	if !exists {
		t.Fatal("breakdown phase not found")
	}

	artifact := projschema.ArtifactState{
		Type:       "discovery",
		Path:       path,
		Created_at: time.Now(),
		Approved:   true, // Discovery artifacts are approved for testing
	}

	phase.Outputs = append(phase.Outputs, artifact)
	p.Phases["breakdown"] = phase
}

func addWorkUnitArtifact(t *testing.T, p *state.Project, path string) string {
	t.Helper()

	phase, exists := p.Phases["breakdown"]
	if !exists {
		t.Fatal("breakdown phase not found")
	}

	artifact := projschema.ArtifactState{
		Type:       "work_unit_spec",
		Path:       path,
		Created_at: time.Now(),
		Approved:   false,
	}

	phase.Outputs = append(phase.Outputs, artifact)
	p.Phases["breakdown"] = phase

	return path
}

// linkArtifactToTask links an artifact to a task via metadata.
func linkArtifactToTask(t *testing.T, p *state.Project, taskID, artifactPath string) {
	t.Helper()

	phase, exists := p.Phases["breakdown"]
	if !exists {
		t.Fatal("breakdown phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			if phase.Tasks[i].Metadata == nil {
				phase.Tasks[i].Metadata = make(map[string]interface{})
			}
			phase.Tasks[i].Metadata["artifact_path"] = artifactPath
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["breakdown"] = phase
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// publishWorkUnit marks a work unit as published with GitHub metadata.
func publishWorkUnit(t *testing.T, p *state.Project, taskID string, issueNum int, issueURL string) {
	t.Helper()

	phase, exists := p.Phases["breakdown"]
	if !exists {
		t.Fatal("breakdown phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			if phase.Tasks[i].Metadata == nil {
				phase.Tasks[i].Metadata = make(map[string]interface{})
			}
			phase.Tasks[i].Metadata["published"] = true
			phase.Tasks[i].Metadata["github_issue_number"] = issueNum
			phase.Tasks[i].Metadata["github_issue_url"] = issueURL
			p.Phases["breakdown"] = phase
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// markTaskInProgress marks a task as in_progress.
func markTaskInProgress(t *testing.T, p *state.Project, taskID string) {
	t.Helper()

	phase, exists := p.Phases["breakdown"]
	if !exists {
		t.Fatal("breakdown phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			phase.Tasks[i].Status = "in_progress"
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["breakdown"] = phase
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// markTaskNeedsReview marks a task as needs_review.
func markTaskNeedsReview(t *testing.T, p *state.Project, taskID string) {
	t.Helper()

	phase, exists := p.Phases["breakdown"]
	if !exists {
		t.Fatal("breakdown phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			phase.Tasks[i].Status = "needs_review"
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["breakdown"] = phase
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// markTaskCompleted marks a task as completed and auto-approves its artifact.
func markTaskCompleted(t *testing.T, p *state.Project, taskID string) {
	t.Helper()

	phase, exists := p.Phases["breakdown"]
	if !exists {
		t.Fatal("breakdown phase not found")
	}

	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			phase.Tasks[i].Status = "completed"
			phase.Tasks[i].Updated_at = time.Now()
			p.Phases["breakdown"] = phase

			// Auto-approve artifact
			err := autoApproveArtifact(p, taskID)
			if err != nil {
				t.Fatalf("failed to auto-approve artifact: %v", err)
			}
			return
		}
	}

	t.Fatalf("task %s not found", taskID)
}

// verifyArtifactApproved verifies the approval status of an artifact.
func verifyArtifactApproved(t *testing.T, p *state.Project, artifactPath string, expectedApproved bool) {
	t.Helper()

	phase, exists := p.Phases["breakdown"]
	if !exists {
		t.Fatal("breakdown phase not found")
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

// verifyPhaseStatus verifies the status of the breakdown phase.
func verifyPhaseStatus(t *testing.T, p *state.Project, expectedStatus string) {
	t.Helper()

	phase, exists := p.Phases["breakdown"]
	if !exists {
		t.Fatalf("phase breakdown not found")
	}

	if phase.Status != expectedStatus {
		t.Errorf("phase breakdown status: got %v, want %v", phase.Status, expectedStatus)
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
