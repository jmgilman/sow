# Task 020: Auto-Determination Mode (Enhanced)

## Context

This task implements the enhanced auto-determination mode for the `sow advance` command. Auto-determination is the backward-compatible default behavior when no event argument is provided: `sow advance`.

The current implementation works but provides poor error messages when auto-determination fails (particularly for intent-based branching). This task enhances the error messages to guide orchestrators to the appropriate next action.

**What Auto-Determination Does**:
- Calls `project.Config().DetermineEvent(project)` to automatically select the next event
- Works for linear states (one transition) and state-determined branching (AddBranch with discriminator)
- **Fails** for intent-based branching (multiple valid transitions with no discriminator)

**Enhancement Goal**: When auto-determination fails, suggest using `--list` to discover available options.

## Requirements

### Extract Current Logic to Helper Function

Move the existing auto-determination logic (lines 42-82 in advance.go) to a new helper function:

```go
func executeAutoTransition(ctx *sow.Context, project *state.Project, machine *sdkstate.Machine, currentState sdkstate.State) error
```

This function should:
1. Call `project.Config().DetermineEvent(project)`
2. Fire the event with `project.Config().FireWithPhaseUpdates(machine, event, project)`
3. Save the project state
4. Display the new state

### Enhanced Error Messages

When `DetermineEvent()` fails, provide helpful guidance:

**For terminal states** (no transitions configured):
```
Cannot advance from state ReviewActive: no transitions configured

This may be a terminal state.
```

**For intent-based branching** (multiple transitions, no discriminator):
```
Cannot advance from state Researching: multiple transitions available but no auto-determination configured

Use 'sow advance --list' to see available transitions and select one explicitly.
Available events: finalize, add_more_research
```

### Mode Switching Logic

In the main RunE function, delegate to the helper when appropriate:

```go
// After flag validation, before loading project...
if !listFlag && !dryRunFlag && event == "" {
    // Auto-determination mode
    return executeAutoTransition(ctx, project, machine, currentState)
}
```

## Acceptance Criteria

### Functional Tests (TDD)

Write tests BEFORE implementation in `cli/cmd/advance_test.go`:

1. **TestAdvanceAutoLinear**:
   - Create test project in linear state (one transition)
   - Call `sow advance` (no args)
   - Verify: Transition occurs, state advances, project saved

2. **TestAdvanceAutoBranching**:
   - Create test project with state-determined branching (uses AddBranch)
   - Call `sow advance` (no args)
   - Verify: Discriminator selects correct event, transition occurs

3. **TestAdvanceAutoIntentBased**:
   - Create test project with intent-based branching (multiple transitions, no discriminator)
   - Call `sow advance` (no args)
   - Verify: Error message suggests using `--list`
   - Verify: Error message lists available events

4. **TestAdvanceAutoTerminalState**:
   - Create test project in terminal state (no transitions)
   - Call `sow advance` (no args)
   - Verify: Error indicates terminal state

### Implementation Verification

1. Tests written and failing
2. Helper function extracted
3. Error message enhancements implemented
4. All tests pass
5. Backward compatibility maintained (existing projects work unchanged)

### Error Message Quality

- Clear indication of what went wrong
- Actionable next steps (use `--list`)
- Shows available events when multiple options exist
- Consistent formatting with existing error messages

## Technical Details

### Helper Function Signature

```go
// executeAutoTransition performs automatic event determination and transition.
// This is the backward-compatible default mode when no event is specified.
//
// Returns error if:
// - DetermineEvent fails (terminal state, intent-based branching)
// - Transition fails (guard blocked, invalid event)
// - Save fails (I/O error)
func executeAutoTransition(
    ctx *sow.Context,
    project *state.Project,
    machine *sdkstate.Machine,
    currentState sdkstate.State,
) error {
    fmt.Printf("Current state: %s\n", currentState)

    // Determine which event to fire from current state
    event, err := project.Config().DetermineEvent(project)
    if err != nil {
        // Enhanced error handling
        return enhanceAutoTransitionError(err, project, currentState)
    }

    // Fire the event with automatic phase status updates
    if err := project.Config().FireWithPhaseUpdates(machine, event, project); err != nil {
        return handleTransitionError(err, currentState, event)
    }

    // Save updated state
    if err := project.Save(); err != nil {
        return fmt.Errorf("failed to save state: %w", err)
    }

    // Display new state
    newState := project.Statechart.Current_state
    fmt.Printf("Advanced to: %s\n", newState)

    return nil
}
```

### Error Enhancement Logic

```go
func enhanceAutoTransitionError(err error, project *state.Project, currentState sdkstate.State) error {
    // Check if this is a terminal state (no transitions configured)
    transitions := project.Config().GetAvailableTransitions(currentState)
    if len(transitions) == 0 {
        return fmt.Errorf(
            "cannot advance from state %s: %w\n\nThis may be a terminal state",
            currentState,
            err,
        )
    }

    // Intent-based branching case (multiple transitions, no discriminator)
    if len(transitions) > 1 {
        // Extract event names
        events := make([]string, len(transitions))
        for i, t := range transitions {
            events[i] = string(t.Event)
        }

        return fmt.Errorf(
            "cannot advance from state %s: %w\n\n"+
                "Use 'sow advance --list' to see available transitions and select one explicitly.\n"+
                "Available events: %s",
            currentState,
            err,
            strings.Join(events, ", "),
        )
    }

    // Default error wrapping
    return fmt.Errorf("cannot advance from state %s: %w", currentState, err)
}
```

### Testing Pattern

```go
func TestAdvanceAutoLinear(t *testing.T) {
    // Create test project with linear state machine
    project := createTestProject(t, "linear")

    // Execute command
    cmd := NewAdvanceCmd()
    err := cmd.Execute()

    // Verify success
    if err != nil {
        t.Fatalf("auto advance failed: %v", err)
    }

    // Verify state changed
    reloadedProject := loadProject(t)
    expectedState := "NextState"
    if reloadedProject.Statechart.Current_state != expectedState {
        t.Errorf("state = %v, want %v", reloadedProject.Statechart.Current_state, expectedState)
    }
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/cmd/advance.go` - Current implementation to refactor
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/project/config.go` - GetAvailableTransitions method (lines 301-374)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/project/context/issue-78.md` - Error message specifications (Section 10)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/knowledge/designs/cli-enhanced-advance.md` - Auto-determination mode spec (Section 6.1)

## Examples

### Current Behavior (Poor Error)

```bash
$ sow advance
Error: cannot advance from state Researching: no event determiner configured

This may be a terminal state
```

### Enhanced Behavior (Helpful Error)

```bash
$ sow advance
Error: cannot advance from state Researching: no event determiner configured

Use 'sow advance --list' to see available transitions and select one explicitly.
Available events: finalize, add_more_research
```

### Successful Auto-Advance

```bash
$ sow advance
Current state: ImplementationPlanning
Advanced to: ImplementationExecuting
```

## Dependencies

- **Task 010** must be complete (flag infrastructure exists)
- Uses SDK methods: `DetermineEvent()`, `GetAvailableTransitions()`, `FireWithPhaseUpdates()`

## Constraints

### Backward Compatibility

- CRITICAL: Existing `sow advance` workflows must work unchanged
- No changes to successful transition behavior
- Only error messages are enhanced
- Must not break any existing projects

### Performance

- Error enhancement should be fast (introspection methods are O(n) where n = transitions)
- Acceptable overhead: <10ms for error cases

### Code Organization

- Keep helper function focused and single-purpose
- Error handling logic separate from happy path
- Clear function names and documentation

## Implementation Notes

### TDD Workflow

1. Write failing test for linear state auto-advance
2. Implement helper function, test passes
3. Write failing test for enhanced error messages
4. Implement error enhancement, test passes
5. Write tests for all error cases
6. Implement remaining error handling
7. Verify backward compatibility with manual testing

### Integration Testing

Test against real project types:
- Standard project (state-determined branching with AddBranch)
- Exploration project (intent-based branching)
- Linear project (simple progression)

### Error Message Design

Follow these principles:
1. State what happened ("cannot advance from X")
2. Explain why (terminal state, multiple options, etc.)
3. Suggest action ("use --list to see options")
4. Provide context (list available events)

### Next Steps

After this task:
- Task 030 will implement `--list` mode (discovery)
- Task 040 will implement `--dry-run` mode (validation)
- Task 050 will implement explicit event mode
