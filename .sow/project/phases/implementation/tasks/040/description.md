# Task 040: Dry-Run Mode Implementation

## Context

This task implements the validation mode for `sow advance --dry-run [event]`, which validates whether a specific transition can be executed without actually executing it. This is a safety feature allowing orchestrators to check if a transition will succeed before committing to it.

**Why This Matters**: Orchestrators want to validate transitions before execution to:
- Avoid failed transitions that leave project in inconsistent state
- Verify guards will pass before attempting transition
- Understand what will happen (target state) before committing
- Debugging and troubleshooting state machine issues

**Dry-Run Behavior**:
- Validates that the specified event is configured for current state
- Checks if guards pass (using `machine.CanFire()`)
- Shows what would happen (target state, description)
- **Does not** modify project state or execute actions
- Returns success/failure without side effects

## Requirements

### Dry-Run Mode Helper Function

Create a new helper function that validates a transition:

```go
func validateTransition(
    ctx *sow.Context,
    project *state.Project,
    machine *sdkstate.Machine,
    currentState sdkstate.State,
    event sdkstate.Event,
) error
```

This function should:
1. Verify event is configured for current state (using `GetTargetState`)
2. Check if event can fire (using `machine.CanFire()`)
3. Display validation result with details
4. Return error if event invalid, nil if validation passes

### Output Format

#### Valid Transition (Would Succeed)

```
Validating transition: ReviewActive -> review_pass

✓ Transition is valid and can be executed

Target state: FinalizeChecks
Description: Review approved - proceed to finalization

To execute: sow advance review_pass
```

#### Blocked by Guard

```
Validating transition: ImplementationPlanning -> planning_complete

✗ Transition blocked by guard condition

Guard description: task descriptions approved
Current status: Guard not satisfied

Fix the guard condition, then try again.
```

#### Event Not Configured

```
Validating transition: ReviewActive -> invalid_event

✗ Event 'invalid_event' is not configured for state ReviewActive

Use 'sow advance --list' to see available transitions.
```

### Validation Logic

1. **Check if event is configured**:
   - Use `config.GetTargetState(currentState, event)`
   - If returns empty string, event not configured

2. **Check if event can fire**:
   - Use `machine.CanFire(event)`
   - Returns (bool, error)
   - true = guards pass, false = guards block

3. **Get transition details**:
   - `config.GetTransitionDescription(currentState, event)` - What it does
   - `config.GetGuardDescription(currentState, event)` - What's required
   - Use for informative output

### Mode Switching

In main RunE, delegate to helper when `--dry-run` flag is set:

```go
if dryRunFlag {
    if event == "" {
        return fmt.Errorf("--dry-run requires an event argument")
    }
    return validateTransition(ctx, project, machine, currentState, sdkstate.Event(event))
}
```

Note: Flag validation (event required) already done in Task 010.

## Acceptance Criteria

### Functional Tests (TDD)

Write tests BEFORE implementation in `cli/cmd/advance_test.go`:

1. **TestAdvanceDryRunValid**:
   - Create project where transition would succeed (guard passes)
   - Run `sow advance --dry-run [event]`
   - Verify: Success message shown
   - Verify: Target state displayed
   - Verify: Description displayed if present
   - Verify: Project state unchanged (no side effects)

2. **TestAdvanceDryRunBlocked**:
   - Create project where guard blocks transition
   - Run `sow advance --dry-run [event]`
   - Verify: Blocked message shown
   - Verify: Guard description displayed
   - Verify: Helpful fix message shown
   - Verify: Project state unchanged

3. **TestAdvanceDryRunInvalidEvent**:
   - Create project in known state
   - Run `sow advance --dry-run invalid_event`
   - Verify: Error message about unconfigured event
   - Verify: Suggests using `--list`
   - Verify: Project state unchanged

4. **TestAdvanceDryRunNoSideEffects**:
   - Create project with OnEntry actions (that would modify state)
   - Run `sow advance --dry-run [event]` (valid, would succeed)
   - Verify: OnEntry actions NOT executed
   - Verify: Project state file not modified
   - Verify: Phase status not updated

5. **TestAdvanceDryRunWithoutEvent**:
   - Run `sow advance --dry-run` (no event argument)
   - Verify: Error about missing event
   - Note: This is validated in flag validation (Task 010)

### Validation Completeness

- Event configuration check works
- Guard check works
- Displays enough info for orchestrator to understand result
- No false positives (doesn't say "valid" when it would fail)
- No false negatives (doesn't say "blocked" when it would succeed)

### Side Effect Prevention

- **Critical**: Project state never modified during dry-run
- OnEntry actions never executed
- OnExit actions never executed
- Phase status never updated
- No files written
- No state machine state changes

## Technical Details

### Helper Function Implementation

```go
func validateTransition(
    ctx *sow.Context,
    project *state.Project,
    machine *sdkstate.Machine,
    currentState sdkstate.State,
    event sdkstate.Event,
) error {
    fmt.Printf("Validating transition: %s -> %s\n\n", currentState, event)

    config := project.Config()

    // Check if event is configured for this state
    targetState := config.GetTargetState(currentState, event)
    if targetState == "" {
        fmt.Printf("✗ Event '%s' is not configured for state %s\n\n", event, currentState)
        fmt.Println("Use 'sow advance --list' to see available transitions.")
        return fmt.Errorf("event not configured")
    }

    // Check if event can fire (guard passes)
    canFire, err := machine.CanFire(event)
    if err != nil {
        return fmt.Errorf("failed to validate transition: %w", err)
    }

    if !canFire {
        // Blocked by guard
        fmt.Println("✗ Transition blocked by guard condition")
        fmt.Println()

        guardDesc := config.GetGuardDescription(currentState, event)
        if guardDesc != "" {
            fmt.Printf("Guard description: %s\n", guardDesc)
        }
        fmt.Println("Current status: Guard not satisfied")
        fmt.Println()
        fmt.Println("Fix the guard condition, then try again.")

        return fmt.Errorf("transition blocked by guard")
    }

    // Valid - would succeed
    fmt.Println("✓ Transition is valid and can be executed")
    fmt.Println()
    fmt.Printf("Target state: %s\n", targetState)

    description := config.GetTransitionDescription(currentState, event)
    if description != "" {
        fmt.Printf("Description: %s\n", description)
    }

    fmt.Println()
    fmt.Printf("To execute: sow advance %s\n", event)

    return nil
}
```

### Testing No Side Effects

Critical test to ensure dry-run is truly dry:

```go
func TestAdvanceDryRunNoSideEffects(t *testing.T) {
    // Create project with OnEntry action that sets metadata
    project := createTestProject(t, "StateA")

    // Add transition with OnEntry that would modify project
    // (in real scenario, loaded from config)

    // Capture initial state
    initialState := project.Statechart.Current_state
    initialMetadata := cloneMetadata(project.Phases["test"].Metadata)

    // Run dry-run
    cmd := NewAdvanceCmd()
    cmd.SetArgs([]string{"--dry-run", "event_with_action"})
    err := cmd.Execute()

    if err != nil {
        t.Fatalf("dry-run failed: %v", err)
    }

    // Reload project from disk
    reloaded := loadProject(t)

    // Verify state unchanged
    if reloaded.Statechart.Current_state != initialState {
        t.Error("dry-run modified state machine state")
    }

    // Verify metadata unchanged
    if !metadataEqual(reloaded.Phases["test"].Metadata, initialMetadata) {
        t.Error("dry-run executed OnEntry action (metadata changed)")
    }
}
```

### CanFire Behavior

From `cli/internal/sdks/state/machine.go` (lines 41-48):

```go
// CanFire checks if an event can be fired from the current state.
func (m *Machine) CanFire(event Event) (bool, error) {
    can, err := m.sm.CanFire(event)
    if err != nil {
        return false, fmt.Errorf("failed to check if event %s can fire: %w", event, err)
    }
    return can, nil
}
```

**Behavior**:
- Returns `true` if event configured and guards pass
- Returns `false` if event configured but guards fail
- Returns `false` if event not configured for current state
- Never modifies state

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/state/machine.go` - CanFire method (lines 41-48)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/project/config.go` - GetTargetState, GetGuardDescription (lines 415-482)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/project/context/issue-78.md` - Dry-run specifications (Section 6, Phase 1D)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/knowledge/designs/cli-enhanced-advance.md` - Dry-run mode detailed spec (Section 6.4)

## Examples

### Valid Transition Check

```bash
$ sow advance --dry-run review_pass
Validating transition: ReviewActive -> review_pass

✓ Transition is valid and can be executed

Target state: FinalizeChecks
Description: Review approved - proceed to finalization

To execute: sow advance review_pass

$ echo $?
0
```

### Blocked Transition Check

```bash
$ sow advance --dry-run planning_complete
Validating transition: ImplementationPlanning -> planning_complete

✗ Transition blocked by guard condition

Guard description: task descriptions approved
Current status: Guard not satisfied

Fix the guard condition, then try again.

$ echo $?
1
```

### Invalid Event Check

```bash
$ sow advance --dry-run nonexistent_event
Validating transition: ReviewActive -> nonexistent_event

✗ Event 'nonexistent_event' is not configured for state ReviewActive

Use 'sow advance --list' to see available transitions.

$ echo $?
1
```

## Dependencies

- **Task 010** complete (flag infrastructure, validation)
- **Task 020** complete (can test together with auto mode)
- **Task 030** complete (error messages suggest using `--list`)
- SDK methods: `CanFire()`, `GetTargetState()`, `GetGuardDescription()`

## Constraints

### Side Effect Prevention

- **CRITICAL**: Must never modify project state
- Must not execute any actions (OnEntry, OnExit)
- Must not save project file
- Must not update phase status
- Must not trigger state machine transitions

### Performance

- Should complete in <50ms (just checking guards)
- No I/O except initial project load
- Guard evaluation only (no state changes)

### Error Handling

- Clear distinction between "blocked" (configured but guard fails) and "invalid" (not configured)
- Helpful error messages for each case
- Non-zero exit code for failures

## Implementation Notes

### TDD Workflow

1. Write test for valid transition
2. Implement basic validation (GetTargetState check)
3. Write test for blocked transition
4. Add CanFire check and blocked output
5. Write test for invalid event
6. Add error handling for unconfigured events
7. Write test for no side effects (critical)
8. Verify implementation is truly dry

### Guard Evaluation Timing

Guards are evaluated by `CanFire()` which:
- Does NOT fire the event
- Does NOT change state
- Does NOT execute actions
- Just evaluates guard functions and returns boolean

This is safe for dry-run mode.

### Output Design

Use visual indicators:
- ✓ for success (can execute)
- ✗ for failure (blocked or invalid)
- Clear sections (validation result, details, next action)

### Testing Strategy

Create test projects with:
- Valid transitions (guards pass)
- Blocked transitions (guards fail)
- Invalid events (not configured)
- Transitions with OnEntry/OnExit actions (verify not executed)

### Next Steps

After this task:
- Task 050 will implement explicit event execution mode
- Orchestrators can now validate before executing
- Provides safety for state transitions
