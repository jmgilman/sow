# Task 060: Add Descriptions to All Standard Project Transitions

## Context

This task adds human-readable descriptions to all transitions in the standard project type. This is the first step in the standard project refactoring work, and it should be done BEFORE refactoring the ReviewActive branching logic.

**Why Separate Task**: Adding descriptions is a straightforward, low-risk change that:
- Improves discoverability for all transitions (not just ReviewActive)
- Makes the codebase easier to review (one focused change)
- Provides immediate value for orchestrators using `--list` mode
- Sets up better context for the ReviewActive refactoring

**What Descriptions Do**:
- Explain what each transition accomplishes (the "why")
- Shown by `sow advance --list` to help orchestrators
- Improve code readability and maintainability
- Serve as inline documentation of state machine logic

**Current State**: Most transitions in standard.go lack descriptions. Only guards have descriptions, not transitions.

## Requirements

### Add WithDescription to All Transitions

Modify `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/projects/standard/standard.go` (lines 84-206):

Add `project.WithDescription()` to each `AddTransition()` call. Descriptions should:
- Be concise (one sentence)
- Focus on the outcome/purpose (not mechanics)
- Use consistent voice and style
- Help orchestrators understand what happens

### Transitions to Update

All transitions in the `configureTransitions` function:

1. **NoProject → ImplementationPlanning** (EventProjectInit, line 92)
2. **ImplementationPlanning → ImplementationDraftPRCreation** (EventPlanningComplete, line 99)
3. **ImplementationDraftPRCreation → ImplementationExecuting** (EventDraftPRCreated, line 109)
4. **ImplementationExecuting → ReviewActive** (EventAllTasksComplete, line 119)
5. **ReviewActive → FinalizeChecks** (EventReviewPass, line 129)
6. **ReviewActive → ImplementationPlanning** (EventReviewFail, line 139)
7. **FinalizeChecks → FinalizePRReady** (EventChecksDone, line 178)
8. **FinalizePRReady → FinalizePRChecks** (EventPRReady, line 183)
9. **FinalizePRChecks → FinalizeCleanup** (EventPRChecksPass, line 191)
10. **FinalizeCleanup → NoProject** (EventCleanupComplete, line 199)

### Description Guidelines

**Good descriptions** explain the business logic:
- "Initialize project and begin implementation planning"
- "Task descriptions approved, create draft PR"
- "Review approved, proceed to finalization checks"

**Avoid implementation details**:
- ❌ "Fire EventPlanningComplete and transition to next state"
- ❌ "Call OnEntry handler and update phase status"

**Be consistent**:
- Use imperative mood or present tense consistently
- Start with what's being done or what condition is met
- Keep similar length across descriptions

## Acceptance Criteria

### Functional Tests (TDD)

Write tests BEFORE implementation in `cli/internal/projects/standard/lifecycle_test.go`:

1. **TestStandardProjectDescriptions**:
   - Get config for standard project
   - For each known transition (from state, event):
     - Call `config.GetTransitionDescription(from, event)`
     - Verify: Returns non-empty string
     - Verify: Description is meaningful (not just event name)
   - Ensure all 10 transitions have descriptions

2. **TestDescriptionQuality** (optional but recommended):
   - Verify descriptions are under 100 characters (concise)
   - Verify descriptions don't contain placeholder text like "TODO"
   - Verify descriptions are unique (no copy-paste errors)

### Implementation Verification

1. Tests written and failing (no descriptions yet)
2. Descriptions added to all transitions
3. All tests pass
4. Manual verification using `sow advance --list` in standard project

### Code Quality

- Descriptions are clear and helpful
- Consistent style across all transitions
- No typos or grammatical errors
- Properly formatted (quotes, punctuation)

## Technical Details

### Adding WithDescription

For each transition, add the option:

```go
// Before
AddTransition(
    sdkstate.State(ImplementationPlanning),
    sdkstate.State(ImplementationDraftPRCreation),
    sdkstate.Event(EventPlanningComplete),
    project.WithGuard("task descriptions approved", func(p *state.Project) bool {
        return allTaskDescriptionsApproved(p)
    }),
)

// After
AddTransition(
    sdkstate.State(ImplementationPlanning),
    sdkstate.State(ImplementationDraftPRCreation),
    sdkstate.Event(EventPlanningComplete),
    project.WithDescription("Task descriptions approved, create draft PR"),
    project.WithGuard("task descriptions approved", func(p *state.Project) bool {
        return allTaskDescriptionsApproved(p)
    }),
)
```

Note: `WithDescription` and `WithGuard` are both transition options, can be used together.

### Suggested Descriptions

Based on the standard project workflow:

1. **NoProject → ImplementationPlanning**: "Initialize project and begin implementation planning"
2. **ImplementationPlanning → DraftPRCreation**: "Task descriptions approved, create draft PR"
3. **DraftPRCreation → ImplementationExecuting**: "Draft PR created, begin task execution"
4. **ImplementationExecuting → ReviewActive**: "All implementation tasks completed, ready for review"
5. **ReviewActive → FinalizeChecks**: "Review approved, proceed to finalization checks"
6. **ReviewActive → ImplementationPlanning**: "Review failed, return to implementation planning for rework"
7. **FinalizeChecks → FinalizePRReady**: "Checks completed, prepare PR for final review"
8. **FinalizePRReady → FinalizePRChecks**: "PR body approved, monitoring PR checks"
9. **FinalizePRChecks → FinalizeCleanup**: "All PR checks passed, begin cleanup"
10. **FinalizeCleanup → NoProject**: "Cleanup complete, project finalized"

These are suggestions - implementer can refine based on context.

### Testing Pattern

```go
func TestStandardProjectDescriptions(t *testing.T) {
    config := NewStandardProjectConfig()

    transitions := []struct {
        from  sdkstate.State
        event sdkstate.Event
        name  string
    }{
        {sdkstate.State(NoProject), sdkstate.Event(EventProjectInit), "project init"},
        {sdkstate.State(ImplementationPlanning), sdkstate.Event(EventPlanningComplete), "planning complete"},
        // ... all 10 transitions
    }

    for _, tt := range transitions {
        t.Run(tt.name, func(t *testing.T) {
            desc := config.GetTransitionDescription(tt.from, tt.event)

            if desc == "" {
                t.Errorf("transition %s->%s has no description", tt.from, tt.event)
            }

            if len(desc) > 100 {
                t.Errorf("description too long (%d chars): %s", len(desc), desc)
            }

            // Optional: Check description doesn't just repeat event name
            if strings.ToLower(desc) == strings.ToLower(string(tt.event)) {
                t.Errorf("description is just the event name: %s", desc)
            }
        })
    }
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/projects/standard/standard.go` - Transitions to update (lines 84-206)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/projects/standard/lifecycle_test.go` - Add tests here
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/project/builder.go` - WithDescription option (line ~70-80)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/project/context/issue-78.md` - Description requirements (Section 2, Story 2)

## Examples

### Before (No Descriptions)

```bash
$ sow advance --list
Current state: ImplementationPlanning

Available transitions:

  sow advance planning_complete
    → ImplementationDraftPRCreation
    Requires: task descriptions approved
```

### After (With Descriptions)

```bash
$ sow advance --list
Current state: ImplementationPlanning

Available transitions:

  sow advance planning_complete
    → ImplementationDraftPRCreation
    Task descriptions approved, create draft PR
    Requires: task descriptions approved
```

The description helps the orchestrator understand what happens at a glance.

## Dependencies

- **Tasks 010-050** complete (CLI modes work, can test with `--list`)
- No code dependencies (WithDescription already exists in SDK)

## Constraints

### No Behavior Changes

- CRITICAL: Only adding descriptions, not changing any logic
- Transitions fire the same way
- Guards work the same way
- Actions execute the same way
- Only metadata (descriptions) changes

### Backward Compatibility

- Existing projects continue to work
- State machine behavior unchanged
- Only enhanced discoverability

### Style Consistency

- All descriptions should follow same pattern
- Similar length and detail level
- Professional tone (this is documentation)

## Implementation Notes

### TDD Workflow

1. Write test that checks all transitions have descriptions
2. Run test, verify it fails (currently no descriptions)
3. Add descriptions to all transitions
4. Run test, verify it passes
5. Manually test with `sow advance --list` in real project

### Review Checklist

Before marking complete, verify:
- [ ] All 10 transitions have descriptions
- [ ] Descriptions are clear and helpful
- [ ] No typos or grammatical errors
- [ ] Consistent style across all descriptions
- [ ] Test passes
- [ ] Manual testing with `--list` shows descriptions

### Manual Testing

Create or use existing standard project:

```bash
# In a standard project at ImplementationPlanning state
sow advance --list

# Should show descriptions for all available transitions
# Verify descriptions are shown under each transition
```

### Commit Message

This change deserves a focused commit:

```
Add descriptions to all standard project transitions

- All 10 state transitions now have human-readable descriptions
- Improves discoverability when using `sow advance --list`
- Descriptions explain the purpose/outcome of each transition
- No behavior changes, only enhanced metadata
```

### Next Steps

After this task:
- Task 070 will refactor ReviewActive with AddBranch
- Descriptions already in place make the refactoring easier to review
- Both pass/fail branches will have descriptions (already added here)
