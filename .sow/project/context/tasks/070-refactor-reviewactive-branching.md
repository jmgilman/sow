# Task 070: Refactor ReviewActive with AddBranch API

## Context

This task refactors the standard project's ReviewActive state to use the declarative AddBranch API instead of the current workaround pattern. This is the primary goal of the standard project refactoring work.

**Current Problem**: ReviewActive uses a workaround where:
- Two transitions (review_pass and review_fail) share **identical guards** ("latest review approved")
- The guard just checks if ANY review is approved (misleading)
- Real branching logic lives in the OnAdvance discriminator (lines 220-254)
- The discriminator examines review assessment metadata and returns the appropriate event
- This split logic is hard to understand and maintain

**Better Pattern**: Use AddBranch to express this declaratively:
- One branch configuration with discriminator function
- Two branch paths (pass/fail) with proper descriptions
- Guards can be refined or removed (discriminator handles logic)
- All branching logic in one place (co-located)

**Benefits**:
- Clearer code (branching logic is explicit)
- Better discoverability (descriptions on both paths)
- Proper reference implementation for other project types
- Demonstrates AddBranch best practices

## Requirements

### Remove Current Workaround

Delete or replace these parts of `standard.go`:

**Two transitions with identical guards** (lines 129-175):
```go
AddTransition(
    sdkstate.State(ReviewActive),
    sdkstate.State(FinalizeChecks),
    sdkstate.Event(EventReviewPass),
    project.WithGuard("latest review approved", func(p *state.Project) bool {
        return latestReviewApproved(p)
    }),
).
AddTransition(
    sdkstate.State(ReviewActive),
    sdkstate.State(ImplementationPlanning),
    sdkstate.Event(EventReviewFail),
    project.WithGuard("latest review approved", func(p *state.Project) bool {
        return latestReviewApproved(p)
    }),
    project.WithFailedPhase("review"),
    project.WithOnEntry(/* rework setup */),
)
```

**OnAdvance discriminator** (lines 220-254):
```go
OnAdvance(sdkstate.State(ReviewActive), func(p *state.Project) (sdkstate.Event, error) {
    // Find latest approved review
    // Extract assessment metadata
    // Return EventReviewPass or EventReviewFail
})
```

### Implement AddBranch Pattern

Replace the above with a single AddBranch call:

```go
AddBranch(
    sdkstate.State(ReviewActive),
    project.BranchOn(getReviewAssessment),  // Discriminator function
    project.When("pass",
        sdkstate.Event(EventReviewPass),
        sdkstate.State(FinalizeChecks),
        project.WithDescription("Review approved, proceed to finalization checks"),
        // Optional: guard if additional validation needed
    ),
    project.When("fail",
        sdkstate.Event(EventReviewFail),
        sdkstate.State(ImplementationPlanning),
        project.WithDescription("Review failed, return to implementation planning for rework"),
        project.WithFailedPhase("review"),
        project.WithOnEntry(func(p *state.Project) error {
            // Increment implementation iteration
            if err := state.IncrementPhaseIteration(p, "implementation"); err != nil {
                return fmt.Errorf("failed to increment implementation iteration: %w", err)
            }

            // Add failed review as implementation input
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
)
```

### Create Discriminator Function

Extract the discriminator logic to a helper function (can go in `guards.go` or inline):

```go
// getReviewAssessment extracts the assessment ("pass" or "fail") from the latest approved review.
// Returns empty string if no approved review found or assessment missing.
func getReviewAssessment(p *state.Project) string {
    phase, exists := p.Phases["review"]
    if !exists {
        return ""
    }

    // Find latest approved review
    for i := len(phase.Outputs) - 1; i >= 0; i-- {
        artifact := phase.Outputs[i]
        if artifact.Type == "review" && artifact.Approved {
            if assessment, ok := artifact.Metadata["assessment"].(string); ok {
                return assessment  // "pass" or "fail"
            }
        }
    }

    return ""  // No approved review or assessment missing
}
```

### Refactor or Remove latestReviewApproved Guard

The current `latestReviewApproved` guard in `guards.go` (lines 93-109) only checks if a review is approved (binary check). With AddBranch, the discriminator handles the assessment logic.

**Options**:
1. Keep the guard for documentation (it's already descriptive)
2. Remove it if no longer used
3. Refine it to check assessment specifically (if used elsewhere)

Decision: Up to implementer based on whether guard adds value.

## Acceptance Criteria

### Functional Tests (TDD)

Write tests BEFORE implementation in `cli/internal/projects/standard/lifecycle_test.go`:

1. **TestReviewActiveBranchingRefactored**:
   - Test pass path:
     - Create project in ReviewActive with approved review (assessment="pass")
     - Call `config.DetermineEvent(project)`
     - Verify: Returns EventReviewPass
     - Fire event, verify: Transitions to FinalizeChecks
   - Test fail path:
     - Create project in ReviewActive with approved review (assessment="fail")
     - Call `config.DetermineEvent(project)`
     - Verify: Returns EventReviewFail
     - Fire event, verify: Transitions to ImplementationPlanning
     - Verify: Implementation iteration incremented
     - Verify: Failed review added as input

2. **TestReviewActiveBranchDescriptions**:
   - Get config
   - Verify: GetTransitionDescription(ReviewActive, EventReviewPass) returns meaningful description
   - Verify: GetTransitionDescription(ReviewActive, EventReviewFail) returns meaningful description

3. **TestReviewActiveIsBranchingState**:
   - Get config
   - Call `config.IsBranchingState(sdkstate.State(ReviewActive))`
   - Verify: Returns true (it's now a branching state)

### Backward Compatibility Tests

4. **TestExistingReviewWorkflowContinues**:
   - Use existing test: `TestFullLifecycle` (lines 14-173)
   - Verify: Review pass workflow still works (line 98)
   - Use existing test: `TestReviewFailLoop` (lines 175-230)
   - Verify: Review fail workflow still works (line 183)

All existing lifecycle tests must pass unchanged.

### Implementation Verification

1. Tests written (will fail initially)
2. AddBranch implemented
3. Old transitions removed
4. Old OnAdvance removed
5. All tests pass (new and existing)
6. Manual testing with real standard project

### Code Quality

- Discriminator logic is clear and well-commented
- Branch paths are properly configured
- OnEntry action for fail path preserved
- No regression in existing functionality

## Technical Details

### AddBranch Auto-Generation

AddBranch automatically creates:
1. **Two transitions** (one per When clause)
2. **OnAdvance handler** that calls discriminator and returns corresponding event
3. **Guards** for each branch (if provided via WithGuard in When clauses)

The generated OnAdvance is equivalent to the current manual implementation (lines 220-254).

### Discriminator Contract

The discriminator function must:
- Return string matching one of the When clause values ("pass" or "fail")
- Return empty string if cannot determine (will cause DetermineEvent to fail)
- Be deterministic (same input = same output)
- Not modify project state

### OnEntry Action Preservation

The fail path's OnEntry action (increment iteration, add failed review as input) must be preserved exactly. This is critical for the rework loop to function.

**Current OnEntry** (lines 147-173):
```go
project.WithOnEntry(func(p *state.Project) error {
    // Only execute rework logic if review phase exists and has failed
    reviewPhase, hasReview := p.Phases["review"]
    if !hasReview || len(reviewPhase.Outputs) == 0 {
        return nil
    }

    // Increment implementation iteration
    if err := state.IncrementPhaseIteration(p, "implementation"); err != nil {
        return fmt.Errorf("failed to increment implementation iteration: %w", err)
    }

    // Add failed review as implementation input
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
```

This exact logic goes into the fail When clause's WithOnEntry.

### File Changes Summary

**standard.go**:
- Lines 129-175: Replace two AddTransition calls with one AddBranch call
- Lines 220-254: Remove OnAdvance for ReviewActive (auto-generated by AddBranch)

**guards.go** (optional):
- Add `getReviewAssessment` helper function
- Possibly remove `latestReviewApproved` if no longer used

**lifecycle_test.go**:
- Add tests for refactored branching
- Verify existing tests still pass

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/projects/standard/standard.go` - Current implementation (lines 129-175, 220-254)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/projects/standard/guards.go` - Guard helpers (lines 93-109)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/project/builder.go` - AddBranch API (lines 101-200)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/project/branch.go` - BranchOn and When functions
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/knowledge/designs/sdk-addbranch-api.md` - Binary branch example (lines 380-433)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/project/context/issue-78.md` - Refactoring specification (Section 2, Story 2)

## Examples

### Before (Workaround)

```go
// Two transitions with misleading identical guards
AddTransition(ReviewActive, FinalizeChecks, EventReviewPass,
    project.WithGuard("latest review approved", latestReviewApproved))

AddTransition(ReviewActive, ImplementationPlanning, EventReviewFail,
    project.WithGuard("latest review approved", latestReviewApproved),
    project.WithFailedPhase("review"),
    project.WithOnEntry(/* rework setup */))

// Real logic hidden in OnAdvance
OnAdvance(ReviewActive, func(p *state.Project) (sdkstate.Event, error) {
    assessment := getReviewAssessment(p)
    switch assessment {
    case "pass": return EventReviewPass, nil
    case "fail": return EventReviewFail, nil
    default: return "", fmt.Errorf("invalid assessment")
    }
})
```

### After (Declarative)

```go
// Single branch configuration with explicit paths
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
        project.WithFailedPhase("review"),
        project.WithOnEntry(/* rework setup */),
    ),
)
```

Much clearer: discriminator function, two explicit paths, all logic co-located.

## Dependencies

- **Task 060** complete (descriptions already added, easier to see what's preserved)
- SDK AddBranch API available (already implemented)

## Constraints

### Backward Compatibility

- CRITICAL: Existing standard projects must continue to work
- Projects already in ReviewActive state must advance correctly
- Both pass and fail workflows must function identically
- No breaking changes to behavior

### Testing Rigor

- Must test both pass and fail paths explicitly
- Must test backward compatibility (existing tests pass)
- Must test that OnEntry actions still execute
- Must test that iteration increments correctly

### Code Clarity

- Discriminator function should be clear and well-named
- Comments should explain the branching logic
- Code should be easier to understand than before

## Implementation Notes

### TDD Workflow

1. Write test for pass path (using AddBranch)
2. Implement AddBranch with pass When clause
3. Test passes
4. Write test for fail path
5. Add fail When clause with OnEntry
6. Test passes
7. Run all existing tests
8. Verify no regressions
9. Remove old code (transitions, OnAdvance)
10. All tests still pass

### Discriminator Placement

Two options:
1. **Inline in BranchOn**: Co-locates with branch configuration
2. **Helper function**: Reusable, testable separately

Recommendation: Helper function in `guards.go` for consistency with other guard helpers.

### Testing the Discriminator

Can test discriminator independently:

```go
func TestGetReviewAssessment(t *testing.T) {
    t.Run("returns pass for pass assessment", func(t *testing.T) {
        project := createTestProject(t, ReviewActive)
        addApprovedReview(t, project, "pass", "review.md")

        assessment := getReviewAssessment(project)
        if assessment != "pass" {
            t.Errorf("expected 'pass', got '%s'", assessment)
        }
    })

    t.Run("returns fail for fail assessment", func(t *testing.T) {
        project := createTestProject(t, ReviewActive)
        addApprovedReview(t, project, "fail", "review.md")

        assessment := getReviewAssessment(project)
        if assessment != "fail" {
            t.Errorf("expected 'fail', got '%s'", assessment)
        }
    })

    t.Run("returns empty for no approved review", func(t *testing.T) {
        project := createTestProject(t, ReviewActive)

        assessment := getReviewAssessment(project)
        if assessment != "" {
            t.Errorf("expected empty, got '%s'", assessment)
        }
    })
}
```

### Manual Testing

Test with real standard project:

```bash
# Create standard project, get to ReviewActive
cd standard-project

# Add approved review with pass assessment
sow phase add output review review.md --phase review
sow phase set metadata.assessment pass --artifact review.md --phase review
sow artifact approve review.md --phase review

# Advance (should go to FinalizeChecks)
sow advance
# Verify state is FinalizeChecks

# Test fail path with another project
# Add approved review with fail assessment
sow phase add output review review-fail.md --phase review
sow phase set metadata.assessment fail --artifact review-fail.md --phase review
sow artifact approve review-fail.md --phase review

# Advance (should go to ImplementationPlanning)
sow advance
# Verify state is ImplementationPlanning
# Verify implementation iteration incremented
```

### Commit Strategy

This is a significant refactoring, deserves focused commit:

```
refactor: use AddBranch for ReviewActive state branching

Replaces the workaround pattern (two transitions with identical guards
+ OnAdvance discriminator) with declarative AddBranch API.

Changes:
- Single AddBranch call replaces two AddTransition calls
- Discriminator function explicitly examines review assessment
- Both branches have clear descriptions
- OnAdvance auto-generated by AddBranch
- No behavior changes, identical functionality

Benefits:
- Clearer code (branching logic is explicit)
- Better discoverability (descriptions on branches)
- Serves as reference implementation for AddBranch
- Easier to maintain and understand
```

### Next Steps

After this task:
- Standard project fully refactored (uses AddBranch)
- Serves as reference for other project types
- CLI enhanced advance command complete
- All work unit goals achieved
