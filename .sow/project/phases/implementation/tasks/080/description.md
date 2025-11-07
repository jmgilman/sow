# Task 080: Integration Testing and Final Validation

## Context

This is the final task in the work unit, focusing on end-to-end integration testing and validation that all components work together correctly. While each previous task had unit tests, this task ensures the complete system works as designed.

**Why This Matters**:
- Validates that CLI modes integrate with SDK introspection correctly
- Ensures standard project refactoring works in real-world scenarios
- Catches integration issues that unit tests might miss
- Verifies backward compatibility across the system
- Provides confidence for production use

**What to Test**:
- All four CLI modes against real project types
- Standard project full lifecycle with refactored branching
- CLI modes with standard project (specific integration)
- Edge cases and error scenarios
- Backward compatibility with existing projects

## Requirements

### End-to-End CLI Tests

Create comprehensive integration tests in `cli/cmd/advance_integration_test.go` (new file):

1. **Test Auto-Advance with Standard Project**:
   - Create standard project, advance through full lifecycle
   - Verify auto-determination works for linear states
   - Verify auto-determination works for ReviewActive (branching)
   - Verify error message for intent-based states (if any exist)

2. **Test List Mode with Various States**:
   - Test in linear state (one transition)
   - Test in branching state (multiple transitions)
   - Test with all guards passing
   - Test with some guards blocked
   - Test in terminal state

3. **Test Dry-Run Mode**:
   - Test valid transition (guards pass)
   - Test blocked transition (guards fail)
   - Test invalid event
   - Verify no side effects (state unchanged)

4. **Test Explicit Event Mode**:
   - Test successful transition
   - Test guard failure with enhanced error
   - Test invalid event

### Standard Project Lifecycle Tests

Extend `cli/internal/projects/standard/lifecycle_test.go`:

1. **TestStandardProjectWithEnhancedCLI**:
   - Full lifecycle test using new CLI modes
   - Use `--list` to discover options at key states
   - Use `--dry-run` to validate before advancing
   - Use explicit events where appropriate
   - Verify all modes work correctly

2. **TestReviewActiveBranchingBothPaths**:
   - Test pass path (review approved → FinalizeChecks)
   - Test fail path (review failed → ImplementationPlanning rework)
   - Verify discriminator selects correct branch
   - Verify both paths have descriptions
   - Verify OnEntry actions execute correctly

### Backward Compatibility Tests

Create `cli/cmd/advance_compatibility_test.go` (new file):

1. **TestExistingProjectsContinueWorking**:
   - Load project state from pre-refactor format
   - Verify auto-advance still works
   - Verify no breaking changes to state machine

2. **TestNewCLIModesBackwardCompatible**:
   - Verify `sow advance` (no args) unchanged
   - Verify error messages improved but not breaking
   - Verify new flags don't affect old workflows

### Error Scenario Tests

Test comprehensive error handling:

1. **TestCLIErrorMessages**:
   - Verify all error messages are helpful
   - Verify suggestions are actionable (use --list, etc.)
   - Verify guard descriptions shown when guards fail
   - Verify consistent formatting

2. **TestEdgeCases**:
   - Empty project (no phases)
   - Corrupted state (missing data)
   - Guards with errors (not just false)
   - Invalid state machine configurations

## Acceptance Criteria

### All Integration Tests Pass

- End-to-end CLI tests pass
- Standard project lifecycle tests pass
- Backward compatibility tests pass
- Error scenario tests pass

### Manual Testing Checklist

Perform manual testing with real projects:

- [ ] Create standard project, advance through full lifecycle using auto mode
- [ ] Use `--list` at each state to verify descriptions shown
- [ ] Use `--dry-run` to validate transitions before executing
- [ ] Use explicit events to advance when multiple options exist
- [ ] Test ReviewActive branching (both pass and fail paths)
- [ ] Verify error messages are helpful when guards fail
- [ ] Test with existing project (backward compatibility)

### Performance Validation

- [ ] `sow advance --list` completes in <100ms
- [ ] `sow advance --dry-run [event]` completes in <50ms
- [ ] No performance regression in auto mode
- [ ] Large projects (10+ transitions) perform acceptably

### Code Quality

- [ ] All tests are well-documented
- [ ] Test names clearly indicate what's being tested
- [ ] Test coverage for all modes and scenarios
- [ ] No flaky tests (run multiple times to verify)

## Technical Details

### Integration Test Structure

```go
// cli/cmd/advance_integration_test.go

package cmd

import (
    "testing"
    "github.com/jmgilman/sow/cli/internal/projects/standard"
)

func TestCLIWithStandardProject(t *testing.T) {
    t.Run("auto-advance through full lifecycle", func(t *testing.T) {
        // Create standard project
        project := setupStandardProject(t)

        // Advance through states using auto mode
        states := []string{
            "ImplementationPlanning",
            "ImplementationDraftPRCreation",
            "ImplementationExecuting",
            "ReviewActive",
            "FinalizeChecks",
            // ... etc
        }

        for i, expectedState := range states {
            // Verify current state
            if project.Statechart.Current_state != expectedState {
                t.Fatalf("step %d: expected state %s, got %s", i, expectedState, project.Statechart.Current_state)
            }

            // Setup prerequisites for next transition
            setupPrerequisites(t, project, expectedState)

            // Auto-advance
            cmd := NewAdvanceCmd()
            if err := cmd.Execute(); err != nil {
                t.Fatalf("step %d: auto advance failed: %v", i, err)
            }

            // Reload project
            project = reloadProject(t)
        }
    })

    t.Run("list mode shows all transitions", func(t *testing.T) {
        project := setupStandardProject(t)

        // Capture output
        output := captureOutput(func() {
            cmd := NewAdvanceCmd()
            cmd.SetArgs([]string{"--list"})
            cmd.Execute()
        })

        // Verify output contains expected transitions
        if !strings.Contains(output, "sow advance") {
            t.Error("list mode output missing transition commands")
        }
        if !strings.Contains(output, "→") {
            t.Error("list mode output missing target states")
        }
    })

    t.Run("dry-run validates without side effects", func(t *testing.T) {
        project := setupStandardProject(t)
        initialState := project.Statechart.Current_state

        // Dry-run a valid transition
        cmd := NewAdvanceCmd()
        cmd.SetArgs([]string{"--dry-run", "planning_complete"})
        err := cmd.Execute()

        // Verify validation result
        if err == nil {
            // Should succeed (or fail based on setup)
        }

        // Verify no side effects
        project = reloadProject(t)
        if project.Statechart.Current_state != initialState {
            t.Error("dry-run modified state")
        }
    })

    t.Run("explicit event with ReviewActive branching", func(t *testing.T) {
        // Test both branches explicitly
        testReviewPass := func(t *testing.T) {
            project := setupProjectInReviewActive(t)
            addApprovedReview(t, project, "pass", "review.md")

            cmd := NewAdvanceCmd()
            cmd.SetArgs([]string{"review_pass"})
            if err := cmd.Execute(); err != nil {
                t.Fatalf("explicit review_pass failed: %v", err)
            }

            project = reloadProject(t)
            if project.Statechart.Current_state != "FinalizeChecks" {
                t.Error("review_pass did not advance to FinalizeChecks")
            }
        }

        testReviewFail := func(t *testing.T) {
            project := setupProjectInReviewActive(t)
            addApprovedReview(t, project, "fail", "review.md")

            cmd := NewAdvanceCmd()
            cmd.SetArgs([]string{"review_fail"})
            if err := cmd.Execute(); err != nil {
                t.Fatalf("explicit review_fail failed: %v", err)
            }

            project = reloadProject(t)
            if project.Statechart.Current_state != "ImplementationPlanning" {
                t.Error("review_fail did not return to ImplementationPlanning")
            }
        }

        t.Run("pass path", testReviewPass)
        t.Run("fail path", testReviewFail)
    })
}
```

### Helper Functions for Integration Tests

```go
func setupStandardProject(t *testing.T) *state.Project {
    // Create minimal standard project in ImplementationPlanning state
    // With all required phases initialized
}

func setupPrerequisites(t *testing.T, p *state.Project, state string) {
    // Setup guards to pass for given state
    // E.g., approve task descriptions, complete tasks, etc.
}

func captureOutput(f func()) string {
    // Capture stdout during function execution
}

func reloadProject(t *testing.T) *state.Project {
    // Reload project from disk (verifies persistence)
}
```

### Test Coverage Goals

- **CLI modes**: All four modes tested end-to-end
- **Standard project**: Full lifecycle covered
- **ReviewActive**: Both branches tested
- **Error cases**: All error paths tested
- **Edge cases**: Terminal states, blocked guards, etc.

### Performance Benchmarks

Create benchmarks for critical paths:

```go
func BenchmarkAdvanceList(b *testing.B) {
    project := setupStandardProject(b)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cmd := NewAdvanceCmd()
        cmd.SetArgs([]string{"--list"})
        cmd.Execute()
    }
}

func BenchmarkAdvanceDryRun(b *testing.B) {
    project := setupStandardProject(b)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cmd := NewAdvanceCmd()
        cmd.SetArgs([]string{"--dry-run", "planning_complete"})
        cmd.Execute()
    }
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/cmd/advance.go` - Complete CLI implementation
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/projects/standard/standard.go` - Refactored standard project
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/projects/standard/lifecycle_test.go` - Existing lifecycle tests
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/project/context/issue-78.md` - Acceptance criteria (Section 7)

## Examples

### Successful Integration Test Output

```
=== RUN   TestCLIWithStandardProject
=== RUN   TestCLIWithStandardProject/auto-advance_through_full_lifecycle
--- PASS: TestCLIWithStandardProject/auto-advance_through_full_lifecycle (0.15s)
=== RUN   TestCLIWithStandardProject/list_mode_shows_all_transitions
--- PASS: TestCLIWithStandardProject/list_mode_shows_all_transitions (0.02s)
=== RUN   TestCLIWithStandardProject/dry-run_validates_without_side_effects
--- PASS: TestCLIWithStandardProject/dry-run_validates_without_side_effects (0.01s)
=== RUN   TestCLIWithStandardProject/explicit_event_with_ReviewActive_branching
=== RUN   TestCLIWithStandardProject/explicit_event_with_ReviewActive_branching/pass_path
--- PASS: TestCLIWithStandardProject/explicit_event_with_ReviewActive_branching/pass_path (0.03s)
=== RUN   TestCLIWithStandardProject/explicit_event_with_ReviewActive_branching/fail_path
--- PASS: TestCLIWithStandardProject/explicit_event_with_ReviewActive_branching/fail_path (0.03s)
--- PASS: TestCLIWithStandardProject (0.24s)
PASS
```

### Manual Testing Session

```bash
# Create standard project
$ sow project new standard test-feature
Created standard project: test-feature

# List available transitions
$ sow advance --list
Current state: ImplementationPlanning

Available transitions:

  sow advance planning_complete
    → ImplementationDraftPRCreation
    Task descriptions approved, create draft PR
    Requires: task descriptions approved

# Approve planning (setup prerequisite)
$ sow phase set metadata.planning_approved true --phase implementation

# Dry-run to validate
$ sow advance --dry-run planning_complete
Validating transition: ImplementationPlanning -> planning_complete

✓ Transition is valid and can be executed

Target state: ImplementationDraftPRCreation
Description: Task descriptions approved, create draft PR

To execute: sow advance planning_complete

# Execute transition
$ sow advance planning_complete
Current state: ImplementationPlanning
Advanced to: ImplementationDraftPRCreation

# Continue through lifecycle...
```

## Dependencies

- **All previous tasks** (010-070) must be complete
- All unit tests passing
- CLI modes implemented
- Standard project refactored

## Constraints

### No New Features

- Only testing existing functionality
- No new code except test code
- May add test utilities/helpers

### Test Reliability

- Tests must be deterministic (no flaky tests)
- Tests must clean up after themselves
- Tests must not interfere with each other

### Documentation

- Test names clearly describe what's tested
- Complex test setup should be commented
- Integration test results should be clear

## Implementation Notes

### Test Organization

Create clear separation:
- `advance_test.go` - Unit tests for individual modes
- `advance_integration_test.go` - End-to-end integration tests
- `advance_compatibility_test.go` - Backward compatibility tests

### Test Data

Create reusable test fixtures:
- Standard project in various states
- Projects with different configurations
- Test data for all scenarios

### Debugging Failed Tests

If integration tests fail:
1. Check unit tests first (should all pass)
2. Isolate failing scenario
3. Add logging/output capture
4. Verify test setup is correct
5. Check for race conditions

### Performance Monitoring

Track performance metrics:
- List mode response time
- Dry-run validation time
- Auto-advance execution time
- Compare to baseline (before changes)

### Final Validation Checklist

Before marking work unit complete:

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] All backward compatibility tests pass
- [ ] Manual testing completed successfully
- [ ] Performance acceptable (no regressions)
- [ ] Error messages tested and helpful
- [ ] Code reviewed and clean
- [ ] Documentation updated if needed
- [ ] Ready for production use

## Next Steps

After this task completes:
- **Work unit is complete**
- CLI enhanced advance command ready for use
- Standard project refactored and tested
- Full test coverage across unit and integration levels
- Ready for code review and merge
