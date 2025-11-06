# Task 110: Create Integration Tests

## Context

This task creates end-to-end integration tests that verify the complete exploration project lifecycle works correctly. Integration tests exercise the full state machine, including transitions, guards, and state updates.

The tests simulate real workflows:
1. **Single summary workflow** - Simple exploration with one summary document
2. **Multiple summaries workflow** - Complex exploration with multiple summary documents
3. **Edge cases** - Invalid transitions, guard failures, state validation

These tests ensure all components work together correctly and catch integration issues that unit tests might miss.

## Requirements

### Create Integration Test File

File: `cli/internal/projects/exploration/integration_test.go`

Create comprehensive lifecycle tests:

### 1. Single Summary Workflow Test

Test complete workflow with one summary:

1. Create exploration project
2. Add 3 research topics (tasks)
3. Mark all topics as completed
4. Advance to Summarizing state (verify guard passes)
5. Create and approve single summary artifact
6. Advance to Finalizing state (verify guard passes)
7. Add and complete finalization tasks
8. Advance to Completed state (verify guard passes)
9. Verify final project state

### 2. Multiple Summaries Workflow Test

Test workflow with multiple summary documents:

1. Create exploration project
2. Add and complete research topics
3. Advance to Summarizing
4. Create multiple summary artifacts (including summary.md)
5. Approve all summaries
6. Advance to Finalizing
7. Verify summary.md requirement validated
8. Complete finalization
9. Verify all summaries handled correctly

### 3. Guard Failure Tests

Test that guards properly block invalid transitions:

1. **Active → Summarizing guard failure**:
   - Try to advance with pending tasks
   - Verify transition blocked
   - Verify state unchanged

2. **Summarizing → Finalizing guard failure**:
   - Try to advance without approved summaries
   - Try to advance with no summaries
   - Verify transitions blocked

3. **Finalizing → Completed guard failure**:
   - Try to advance with incomplete finalization tasks
   - Verify transition blocked

### 4. State Validation Test

Test that project state validates correctly at each stage:

1. Verify exploration phase status updates correctly
2. Verify finalization phase enabled at correct time
3. Verify timestamps set correctly
4. Verify phase completion markers

### Test Structure

```go
func TestExplorationLifecycle_SingleSummary(t *testing.T) {
    // Setup: Create project and state machine
    proj, machine := setupExplorationProject(t)

    // Phase 1: Active research
    t.Run("active research phase", func(t *testing.T) {
        // Add research topics
        // Complete all topics
        // Verify state
    })

    // Phase 2: Advance to summarizing
    t.Run("advance to summarizing", func(t *testing.T) {
        // Fire event
        // Verify transition
        // Verify phase status
    })

    // Phase 3: Create summary
    t.Run("create and approve summary", func(t *testing.T) {
        // Add summary artifact
        // Approve it
    })

    // Phase 4: Advance to finalizing
    t.Run("advance to finalizing", func(t *testing.T) {
        // Fire event
        // Verify exploration complete
        // Verify finalization enabled
    })

    // Phase 5: Complete finalization
    t.Run("complete finalization", func(t *testing.T) {
        // Add finalization tasks
        // Complete them
        // Advance to completed
    })

    // Verify final state
    t.Run("verify final state", func(t *testing.T) {
        // Check all phases completed
        // Check state machine in Completed
        // Check timestamps
    })
}
```

### Test Utilities

Create helper functions:

1. **setupExplorationProject()** - Create project with state machine:
   ```go
   func setupExplorationProject(t *testing.T) (*state.Project, *sdkstate.Machine)
   ```

2. **addResearchTopic()** - Add task to exploration phase:
   ```go
   func addResearchTopic(t *testing.T, p *state.Project, id, name, status string)
   ```

3. **addSummaryArtifact()** - Add summary to outputs:
   ```go
   func addSummaryArtifact(t *testing.T, p *state.Project, path string, approved bool)
   ```

4. **verifyPhaseStatus()** - Assert phase in expected state:
   ```go
   func verifyPhaseStatus(t *testing.T, p *state.Project, phaseName, expectedStatus string)
   ```

5. **verifyState()** - Assert state machine in expected state:
   ```go
   func verifyState(t *testing.T, machine *sdkstate.Machine, expected sdkstate.State)
   ```

## Acceptance Criteria

- [ ] File `integration_test.go` created
- [ ] Single summary workflow test implemented
- [ ] Multiple summaries workflow test implemented
- [ ] Guard failure tests implemented for all transitions
- [ ] State validation tests verify correct state updates
- [ ] Test utilities implemented
- [ ] All tests pass
- [ ] Tests exercise complete state machine lifecycle
- [ ] Tests verify phase status transitions
- [ ] Tests verify timestamps set correctly
- [ ] Tests use descriptive subtests
- [ ] Test failures provide clear error messages

## Technical Details

### State Machine Construction

Integration tests need to build the actual state machine from config:

```go
func setupExplorationProject(t *testing.T) (*state.Project, *sdkstate.Machine) {
    t.Helper()

    // Create project
    proj := &state.Project{
        Name:   "test-exploration",
        Branch: "explore/test",
        Type:   "exploration",
        // ... initialize fields
    }

    // Initialize phases
    config := NewExplorationProjectConfig()
    err := config.Initialize(proj, nil)
    if err != nil {
        t.Fatalf("failed to initialize: %v", err)
    }

    // Build state machine
    machine, err := project.BuildMachine(config, proj)
    if err != nil {
        t.Fatalf("failed to build machine: %v", err)
    }

    return proj, machine
}
```

### Event Firing

Use state machine's Fire method to trigger transitions:

```go
err := machine.Fire(EventBeginSummarizing)
if err != nil {
    t.Fatalf("Fire(EventBeginSummarizing) failed: %v", err)
}

// Verify new state
if machine.State() != sdkstate.State(Summarizing) {
    t.Errorf("expected Summarizing, got %v", machine.State())
}
```

### Guard Testing

Guards should prevent invalid transitions:

```go
// Try to advance with pending tasks (should fail)
err := machine.Fire(EventBeginSummarizing)
if err == nil {
    t.Error("expected error firing event with pending tasks, got nil")
}

// State should be unchanged
if machine.State() != sdkstate.State(Active) {
    t.Errorf("state changed unexpectedly: %v", machine.State())
}
```

### Phase Verification

Check phase status after transitions:

```go
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
```

### Timestamp Verification

Ensure timestamps set during transitions:

```go
func verifyTimestamps(t *testing.T, p *state.Project) {
    t.Helper()

    exploration := p.Phases["exploration"]

    // Started at should be set
    if exploration.Started_at.IsZero() {
        t.Error("exploration phase started_at not set")
    }

    // Completed at should be set after completion
    if exploration.Status == "completed" && exploration.Completed_at.IsZero() {
        t.Error("exploration phase completed_at not set")
    }
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/lifecycle_test.go` - Reference integration tests
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/integration_test.go` - SDK integration tests
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/machine.go` - State machine construction
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/knowledge/designs/project-modes/exploration-design.md` - Design specification (testing section)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/project/context/issue-36.md` - Requirements

## Examples

### Standard Project Lifecycle Test (Reference)

From `cli/internal/projects/standard/lifecycle_test.go:14-107`:

```go
func TestFullLifecycle(t *testing.T) {
    // Setup: Create minimal project in NoProject state
    proj, machine := createTestProject(t, NoProject)

    // NoProject → ImplementationPlanning
    t.Run("init transitions to ImplementationPlanning", func(t *testing.T) {
        err := machine.Fire(EventProjectInit)
        if err != nil {
            t.Fatalf("Fire(EventProjectInit) failed: %v", err)
        }
        if got := machine.State(); got != sdkstate.State(ImplementationPlanning) {
            t.Errorf("state = %v, want %v", got, ImplementationPlanning)
        }
    })

    // ImplementationPlanning → ImplementationExecuting
    t.Run("planning complete transitions to execution", func(t *testing.T) {
        addApprovedOutput(t, proj, "implementation", "task_description", "task1.md")
        err := machine.Fire(EventPlanningComplete)
        if err != nil {
            t.Fatalf("Fire(EventPlanningComplete) failed: %v", err)
        }
        // ... verify state
    })

    // ... more transition tests
}
```

### Guard Failure Test Pattern

```go
func TestGuardPreventsInvalidTransition(t *testing.T) {
    proj, machine := setupExplorationProject(t)

    // Add pending task (guard should fail)
    addResearchTopic(t, proj, "010", "Research Topic", "pending")

    // Try to advance (should fail)
    err := machine.Fire(EventBeginSummarizing)
    if err == nil {
        t.Error("expected error with pending tasks, got nil")
    }

    // Verify state unchanged
    if machine.State() != sdkstate.State(Active) {
        t.Errorf("state changed: %v", machine.State())
    }

    // Complete task and try again (should succeed)
    markTaskCompleted(t, proj, "010")
    err = machine.Fire(EventBeginSummarizing)
    if err != nil {
        t.Fatalf("transition failed: %v", err)
    }

    // Verify transition succeeded
    if machine.State() != sdkstate.State(Summarizing) {
        t.Errorf("expected Summarizing, got %v", machine.State())
    }
}
```

## Dependencies

- Task 100 (Unit tests) - Should pass before integration tests
- All implementation tasks (010-090) - Must be complete
- Integration tests verify end-to-end behavior
- These tests catch issues unit tests might miss

## Constraints

- Tests must not depend on file system (use in-memory state)
- Tests must be deterministic (no time-dependent behavior)
- Tests must clean up after themselves (though in-memory tests auto-cleanup)
- Tests should cover both happy path and error cases
- Test names should clearly describe what's being tested
- Use subtests for logical grouping and better error reporting
- Tests must verify both state machine state AND project state
- Tests should verify phase status transitions occur correctly
