# Task 050: Integration Testing and Validation

## Context

This task implements comprehensive integration tests that verify the complete design project type workflow from creation through completion. While previous tasks focused on unit testing individual components (guards, prompts, configuration), this task validates that all components work together correctly as a cohesive system.

Integration tests are critical for:
- Verifying the complete state machine workflow (Active → Finalizing → Completed)
- Testing phase transitions and state changes
- Validating guard enforcement at transition boundaries
- Ensuring action execution (OnEntry/OnExit handlers)
- Confirming proper timestamp management
- Testing edge cases and error scenarios
- Providing regression protection for the entire workflow

These tests simulate real-world orchestrator operations and ensure zero-context resumability works correctly.

## Requirements

### Integration Test File

Create `integration_test.go` with comprehensive end-to-end test scenarios:

### Test Scenarios

1. **TestDesignLifecycle_SingleDocument**
   - Complete happy path with one design document
   - Tests all states: Active → Finalizing → Completed
   - Verifies phase status updates at each transition
   - Validates timestamp management
   - Confirms guard behavior at each boundary

2. **TestDesignLifecycle_MultipleDocuments**
   - Multiple documents with different types (design, adr, architecture)
   - Tests document planning and approval workflow
   - Verifies all document types are allowed
   - Tests mix of completed and abandoned tasks
   - Ensures at least one completed document required

3. **TestDesignLifecycle_WithInputs**
   - Tests design project with initial input artifacts
   - Verifies inputs are tracked separately from outputs
   - Confirms inputs don't block progression

4. **TestDesignLifecycle_ReviewWorkflow**
   - Tests needs_review → in_progress → completed flow
   - Verifies backward transition (revision workflow)
   - Tests auto-approval on task completion

5. **TestDesignLifecycle_AllAbandoned**
   - Tests that advancement blocked when all tasks abandoned
   - Verifies error message is clear
   - Ensures at least one completed task is required

6. **TestDesignLifecycle_NoTasks**
   - Tests that advancement blocked with no document tasks
   - Verifies error message guides user to create tasks

7. **TestDesignLifecycle_TaskValidation**
   - Tests validateTaskForCompletion enforcement
   - Verifies task completion blocked without artifact
   - Tests error messages are actionable

8. **TestDesignLifecycle_AutoApproval**
   - Tests automatic artifact approval on task completion
   - Verifies artifact.Approved flag is set
   - Confirms approval happens atomically with task completion

### Helper Functions

Implement test helper functions for common operations:

```go
// Setup helpers
func setupDesignProject(t *testing.T) (*state.Project, *sdkstate.Machine)
func newTestProject() *state.Project

// Task management helpers
func addDocumentTask(t *testing.T, p *state.Project, id, name, docType, targetPath string)
func markTaskInProgress(t *testing.T, p *state.Project, taskID string)
func markTaskNeedsReview(t *testing.T, p *state.Project, taskID string)
func markTaskCompleted(t *testing.T, p *state.Project, taskID string)
func markTaskAbandoned(t *testing.T, p *state.Project, taskID string)

// Artifact helpers
func addDesignArtifact(t *testing.T, p *state.Project, path, docType string) string
func linkArtifactToTask(t *testing.T, p *state.Project, taskID, artifactPath string)
func verifyArtifactApproved(t *testing.T, p *state.Project, artifactPath string, approved bool)

// Verification helpers
func verifyPhaseStatus(t *testing.T, p *state.Project, phaseName, expectedStatus string)
func verifyPhaseEnabled(t *testing.T, p *state.Project, phaseName string, expectedEnabled bool)
func verifyTaskStatus(t *testing.T, p *state.Project, phaseName, taskID, expectedStatus string)
func verifyTransitionAllowed(t *testing.T, machine *sdkstate.Machine, event sdkstate.Event, expected bool)
```

### Test Structure Pattern

Follow the exploration project integration test pattern:
- Use subtests with `t.Run()` for logical phases
- Setup project and state machine once per test
- Progress through states sequentially
- Verify state changes after each transition
- Check timestamps are set correctly
- Use descriptive test names that explain scenario

## Acceptance Criteria

### Functional Requirements

- [ ] `integration_test.go` created with all test scenarios
- [ ] All helper functions implemented
- [ ] Tests cover complete workflow (Active → Finalizing → Completed)
- [ ] Tests verify guard enforcement at each transition
- [ ] Tests confirm action execution (phase status updates)
- [ ] Tests validate timestamp management
- [ ] Tests check zero-context resumability (state persisted correctly)
- [ ] Edge cases covered (no tasks, all abandoned, missing artifacts)
- [ ] Error scenarios tested with clear expectations

### Test Coverage Requirements

The integration tests must verify:

**State machine behavior**:
- [ ] Initial state is Active
- [ ] Active → Finalizing transition works when guard passes
- [ ] Active → Finalizing blocked when guard fails
- [ ] Finalizing → Completed transition works when guard passes
- [ ] Finalizing → Completed blocked when guard fails
- [ ] Cannot skip states (no Active → Completed)

**Phase lifecycle**:
- [ ] Design phase starts active and enabled
- [ ] Finalization phase starts pending and disabled
- [ ] Design phase marked completed on exit
- [ ] Finalization phase enabled and activated on entry
- [ ] Finalization phase marked completed on final transition
- [ ] Timestamps set correctly (Created_at, Started_at, Completed_at)

**Task lifecycle**:
- [ ] Tasks can be created with metadata
- [ ] Tasks progress through statuses (pending → in_progress → needs_review → completed)
- [ ] Backward transition works (needs_review → in_progress)
- [ ] Tasks can be abandoned
- [ ] Task completion requires linked artifact
- [ ] Task completion auto-approves artifact

**Artifact management**:
- [ ] Artifacts can be added to phase outputs
- [ ] Artifacts can be linked to tasks via metadata
- [ ] Artifact approval status tracks correctly
- [ ] Multiple artifact types supported (design, adr, architecture, diagram, spec)

**Guard enforcement**:
- [ ] allDocumentsApproved enforces business rules
- [ ] allFinalizationTasksComplete enforces completion
- [ ] Guards prevent invalid transitions
- [ ] Guard error messages are clear

### Test Quality

- [ ] Tests are deterministic (no flaky tests)
- [ ] Tests are isolated (no shared state between tests)
- [ ] Tests have clear assertions with descriptive messages
- [ ] Tests use testify/assert and testify/require
- [ ] Helper functions reduce duplication
- [ ] Test names clearly describe scenario
- [ ] Comments explain complex test logic

## Technical Details

### Test Setup Pattern

```go
func TestDesignLifecycle_SingleDocument(t *testing.T) {
	// Setup: Create project and state machine
	proj, machine := setupDesignProject(t)

	// Phase 1: Active design
	t.Run("plan and draft document", func(t *testing.T) {
		// Verify initial state
		assert.Equal(t, sdkstate.State(Active), machine.State())
		verifyPhaseStatus(t, proj, "design", "active")

		// Create document task
		addDocumentTask(t, proj, "010", "Auth Design", "design", ".sow/knowledge/designs/auth.md")

		// Draft document
		artifactPath := addDesignArtifact(t, proj, "project/auth-design.md", "design")
		linkArtifactToTask(t, proj, "010", artifactPath)

		// Mark complete
		markTaskCompleted(t, proj, "010")

		// Verify auto-approval
		verifyArtifactApproved(t, proj, artifactPath, true)
	})

	// Phase 2: Advance to finalizing
	t.Run("advance to finalizing", func(t *testing.T) {
		// Verify guard passes
		can, err := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)
		assert.True(t, can)

		// Fire event
		err = machine.Fire(sdkstate.Event(EventCompleteDesign))
		require.NoError(t, err)

		// Verify transition
		assert.Equal(t, sdkstate.State(Finalizing), machine.State())
		verifyPhaseStatus(t, proj, "design", "completed")
		verifyPhaseStatus(t, proj, "finalization", "in_progress")
	})

	// ... more phases
}
```

### Helper Implementation Pattern

```go
func setupDesignProject(t *testing.T) (*state.Project, *sdkstate.Machine) {
	t.Helper()

	// Create project
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:       "test-design",
			Type:       "design",
			Branch:     "design/test",
			Created_at: time.Now(),
			Updated_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize project
	config := NewDesignProjectConfig()
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Build state machine
	machine, err := config.BuildMachine(proj)
	require.NoError(t, err)

	return proj, machine
}
```

### Timestamp Verification

```go
// Verify timestamps are set during transitions
phase := proj.Phases["design"]
assert.False(t, phase.Completed_at.IsZero(), "design phase should have completion time")

finPhase := proj.Phases["finalization"]
assert.False(t, finPhase.Started_at.IsZero(), "finalization phase should have start time")
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/.sow/knowledge/designs/project-modes/design-design.md` - Success criteria and workflow specification
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/integration_test.go` - Reference integration test pattern
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/sdks/project/integration_test.go` - SDK integration test examples
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/.sow/project/context/issue-37.md` - Acceptance criteria from issue

## Examples

### Complete Test Scenario

```go
func TestDesignLifecycle_ReviewWorkflow(t *testing.T) {
	proj, machine := setupDesignProject(t)

	t.Run("create and draft document", func(t *testing.T) {
		addDocumentTask(t, proj, "010", "API Design", "design", ".sow/knowledge/designs/api.md")
		artifactPath := addDesignArtifact(t, proj, "project/api-design.md", "design")
		linkArtifactToTask(t, proj, "010", artifactPath)
		markTaskInProgress(t, proj, "010")
	})

	t.Run("request review", func(t *testing.T) {
		markTaskNeedsReview(t, proj, "010")
		verifyTaskStatus(t, proj, "design", "010", "needs_review")

		// Cannot advance with needs_review task
		can, _ := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		assert.False(t, can, "should not advance with task in review")
	})

	t.Run("revise after feedback", func(t *testing.T) {
		// Go back to in_progress
		markTaskInProgress(t, proj, "010")
		verifyTaskStatus(t, proj, "design", "010", "in_progress")

		// Make revisions...
		// Request review again
		markTaskNeedsReview(t, proj, "010")
	})

	t.Run("approve and complete", func(t *testing.T) {
		markTaskCompleted(t, proj, "010")

		// Verify auto-approval
		phase := proj.Phases["design"]
		artifact := findArtifact(phase.Outputs, "project/api-design.md")
		assert.True(t, artifact.Approved, "artifact should be auto-approved")

		// Can now advance
		can, _ := machine.CanFire(sdkstate.Event(EventCompleteDesign))
		assert.True(t, can, "should be able to advance after approval")
	})
}
```

### Edge Case Test

```go
func TestDesignLifecycle_NoArtifactBlocks Completion(t *testing.T) {
	proj, _ := setupDesignProject(t)

	// Create task without artifact
	addDocumentTask(t, proj, "010", "Design Doc", "design", ".sow/knowledge/designs/doc.md")

	// Try to complete task without artifact
	err := validateTaskForCompletion(proj, "010")

	assert.Error(t, err, "should error when completing task without artifact")
	assert.Contains(t, err.Error(), "artifact not found", "error should mention missing artifact")
}
```

## Dependencies

- Task 010: Core Structure and Constants
- Task 020: Guard Functions and Helpers
- Task 030: Prompt Templates and Generators
- Task 040: Project Configuration and SDK Integration

All previous tasks must be complete as integration tests exercise the entire system.

## Constraints

### Test Independence

Each test must:
- Create its own project instance
- Not modify global state
- Not depend on execution order
- Clean up resources (though state is in-memory)

### Test Realism

Tests should simulate realistic orchestrator operations:
- Create tasks before artifacts
- Link artifacts to tasks via metadata
- Progress through statuses in logical order
- Use valid artifact types and paths
- Set metadata correctly

### Performance

Integration tests should:
- Complete quickly (< 100ms each typically)
- Not perform I/O (all in-memory)
- Be suitable for CI/CD pipelines

### Coverage Goals

Aim for comprehensive coverage of:
- All state transitions
- All guard conditions (pass and fail)
- All action executions
- Common and edge case scenarios
- Error paths and validation
