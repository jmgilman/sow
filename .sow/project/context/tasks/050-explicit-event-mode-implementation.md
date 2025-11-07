# Task 050: Explicit Event Mode Implementation

## Context

This task implements the explicit event execution mode for `sow advance [event]`, which allows orchestrators to directly fire a specific event without auto-determination. This is the primary mode for intent-based branching scenarios.

**Why This Matters**: In intent-based branching (like exploration projects), the orchestrator must make a decision that cannot be auto-determined from project state. They need to explicitly select which transition to take based on their judgment.

**Explicit Event Behavior**:
- Orchestrator specifies exact event to fire: `sow advance finalize`
- Validates event is configured and guards pass
- Fires the event and advances state
- Provides enhanced error messages when guards fail (shows guard description)
- Used when `DetermineEvent()` cannot decide (multiple valid options)

**Relationship to Other Modes**:
- Auto mode: System picks event (linear or state-determined)
- Explicit mode: Orchestrator picks event (intent-based)
- List mode: Discover options
- Dry-run mode: Validate before executing

## Requirements

### Explicit Event Helper Function

Create a new helper function that validates and executes an explicit event:

```go
func executeExplicitTransition(
    ctx *sow.Context,
    project *state.Project,
    machine *sdkstate.Machine,
    currentState sdkstate.State,
    event sdkstate.Event,
) error
```

This function should:
1. Verify event is configured for current state
2. Fire the event (which validates guards automatically)
3. Save project state
4. Display the new state
5. Provide enhanced error messages on failure

### Enhanced Error Messages

When transition fails, provide context using SDK introspection:

**Invalid Event** (not configured):
```
Error: Event 'invalid_event' is not configured for state ReviewActive

Use 'sow advance --list' to see available transitions.
```

**Guard Failure** (configured but blocked):
```
Error: Transition blocked: task descriptions approved

Current state: ImplementationPlanning
Event: planning_complete
Target state: ImplementationExecuting

The guard condition is not satisfied. Ensure task descriptions are approved before advancing.

Use 'sow advance --dry-run planning_complete' to validate prerequisites.
```

### Validation Before Execution

Before calling `FireWithPhaseUpdates`:
1. Check event is configured using `GetTargetState`
2. Provide helpful error if not configured
3. Let `FireWithPhaseUpdates` handle guard validation (it already does this)
4. Catch guard failures and enhance error message

### Mode Switching

In main RunE, delegate to helper when event argument provided (without flags):

```go
if !listFlag && !dryRunFlag && event != "" {
    // Explicit event mode
    return executeExplicitTransition(ctx, project, machine, currentState, sdkstate.Event(event))
}
```

## Acceptance Criteria

### Functional Tests (TDD)

Write tests BEFORE implementation in `cli/cmd/advance_test.go`:

1. **TestAdvanceExplicitSuccess**:
   - Create project with multiple valid transitions
   - Run `sow advance [event]` where guards pass
   - Verify: Event fires successfully
   - Verify: State advances to correct target
   - Verify: Project saved
   - Verify: New state displayed

2. **TestAdvanceExplicitGuardFailure**:
   - Create project where specified event's guard fails
   - Run `sow advance [event]`
   - Verify: Error message includes guard description
   - Verify: Error message shows current state and target state
   - Verify: Suggests using `--dry-run` for validation
   - Verify: Project state unchanged (transaction rolled back)

3. **TestAdvanceExplicitInvalidEvent**:
   - Create project in known state
   - Run `sow advance invalid_event`
   - Verify: Error about unconfigured event
   - Verify: Suggests using `--list`
   - Verify: Project state unchanged

4. **TestAdvanceExplicitIntentBranching**:
   - Create project with intent-based branching (multiple options)
   - Run `sow advance [event1]`
   - Verify: Selects correct branch
   - Run another project with `sow advance [event2]`
   - Verify: Selects different branch
   - Both should succeed based on explicit choice

5. **TestAdvanceExplicitWithDescriptions**:
   - Create project with transitions having descriptions
   - Run successful explicit advance
   - Verify: Description displayed (optional, for user feedback)

### Error Handling Quality

- Clear indication of what went wrong (guard failure vs invalid event)
- Shows guard description when guard fails
- Provides actionable next steps (use `--dry-run`, use `--list`)
- Includes context (current state, target state, event name)

### Transaction Safety

- If guard fails, project state not modified
- If event invalid, project state not modified
- Only successful transitions modify and save state
- No partial state changes

## Technical Details

### Helper Function Implementation

```go
func executeExplicitTransition(
    ctx *sow.Context,
    project *state.Project,
    machine *sdkstate.Machine,
    currentState sdkstate.State,
    event sdkstate.Event,
) error {
    fmt.Printf("Current state: %s\n", currentState)

    config := project.Config()

    // Validate event is configured for this state
    targetState := config.GetTargetState(currentState, event)
    if targetState == "" {
        fmt.Printf("\nError: Event '%s' is not configured for state %s\n\n", event, currentState)
        fmt.Println("Use 'sow advance --list' to see available transitions.")
        return fmt.Errorf("event not configured")
    }

    // Fire the event (this validates guards and executes transition)
    err := config.FireWithPhaseUpdates(machine, event, project)
    if err != nil {
        // Enhanced error handling for guard failures
        return enhanceTransitionError(err, currentState, event, targetState, config)
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

func enhanceTransitionError(
    err error,
    currentState sdkstate.State,
    event sdkstate.Event,
    targetState sdkstate.State,
    config *project.ProjectTypeConfig,
) error {
    // Check if this is a guard failure
    if strings.Contains(err.Error(), "guard condition is not met") {
        guardDesc := config.GetGuardDescription(currentState, event)

        var msg strings.Builder
        msg.WriteString(fmt.Sprintf("Transition blocked: %s\n\n", guardDesc))
        msg.WriteString(fmt.Sprintf("Current state: %s\n", currentState))
        msg.WriteString(fmt.Sprintf("Event: %s\n", event))
        msg.WriteString(fmt.Sprintf("Target state: %s\n\n", targetState))

        if guardDesc != "" {
            msg.WriteString(fmt.Sprintf("The guard condition is not satisfied. %s\n\n", guardDesc))
        }

        msg.WriteString(fmt.Sprintf("Use 'sow advance --dry-run %s' to validate prerequisites.", event))

        return fmt.Errorf("%s", msg.String())
    }

    // Other error types - use default wrapping
    return fmt.Errorf("failed to advance: %w", err)
}
```

### FireWithPhaseUpdates Behavior

From `cli/internal/sdks/project/machine.go`:

```go
// FireWithPhaseUpdates fires an event and automatically updates phase status.
// It:
// 1. Fires the event (validates guards, executes actions, transitions state)
// 2. Updates phase status based on state machine position
//
// Returns error if:
// - Event not configured for current state
// - Guard conditions not met
// - Actions fail during execution
```

**Guard Validation**: Happens inside `Fire()`, which is called by `FireWithPhaseUpdates`.

**Error Types**:
- Guard failure: Error contains "guard condition is not met"
- Invalid event: Error contains "cannot fire event"
- Action failure: Error from OnEntry/OnExit handlers

### Testing Intent-Based Branching

Create test project type with explicit branching:

```go
func createIntentBranchingProject(t *testing.T) (*state.Project, *project.ProjectTypeConfig) {
    builder := project.NewProjectTypeConfigBuilder("test").
        SetInitialState(sdkstate.State("Researching")).
        AddTransition(
            sdkstate.State("Researching"),
            sdkstate.State("Finalizing"),
            sdkstate.Event("finalize"),
            project.WithDescription("Research complete"),
        ).
        AddTransition(
            sdkstate.State("Researching"),
            sdkstate.State("Researching"),
            sdkstate.Event("add_more_research"),
            project.WithDescription("Need more investigation"),
        )

    config := builder.Build()
    project := createProjectInState(t, "Researching", config)

    return project, config
}

func TestAdvanceExplicitIntentBranching(t *testing.T) {
    // Test choosing 'finalize'
    project1, _ := createIntentBranchingProject(t)
    err := executeExplicitTransition(ctx, project1, machine1, "Researching", "finalize")
    if err != nil {
        t.Fatalf("finalize failed: %v", err)
    }
    if project1.Statechart.Current_state != "Finalizing" {
        t.Error("finalize did not reach Finalizing state")
    }

    // Test choosing 'add_more_research'
    project2, _ := createIntentBranchingProject(t)
    err = executeExplicitTransition(ctx, project2, machine2, "Researching", "add_more_research")
    if err != nil {
        t.Fatalf("add_more_research failed: %v", err)
    }
    if project2.Statechart.Current_state != "Researching" {
        t.Error("add_more_research did not stay in Researching state")
    }
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/project/machine.go` - FireWithPhaseUpdates method (lines 112-200)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/project/config.go` - GetTargetState, GetGuardDescription (lines 415-482)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/state/machine.go` - Guard error handling (lines 65-83)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/project/context/issue-78.md` - Explicit event specifications (Section 6, Phase 1E)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/knowledge/designs/cli-enhanced-advance.md` - Explicit event mode spec (Section 6.2)

## Examples

### Successful Explicit Transition

```bash
$ sow advance finalize
Current state: Researching
Advanced to: Finalizing
```

### Guard Failure with Enhanced Error

```bash
$ sow advance planning_complete
Current state: ImplementationPlanning

Error: Transition blocked: task descriptions approved

Current state: ImplementationPlanning
Event: planning_complete
Target state: ImplementationExecuting

The guard condition is not satisfied. task descriptions approved

Use 'sow advance --dry-run planning_complete' to validate prerequisites.
```

### Invalid Event

```bash
$ sow advance invalid_event
Current state: ReviewActive

Error: Event 'invalid_event' is not configured for state ReviewActive

Use 'sow advance --list' to see available transitions.
```

## Dependencies

- **Task 010** complete (flag infrastructure, argument handling)
- **Task 020** complete (auto mode works)
- **Task 030** complete (can suggest `--list` in errors)
- **Task 040** complete (can suggest `--dry-run` in errors)
- SDK methods: `GetTargetState()`, `GetGuardDescription()`, `FireWithPhaseUpdates()`

## Constraints

### Transaction Safety

- State changes are atomic (either complete or rolled back)
- If guard fails, no state modification
- If save fails, state machine rolled back (handled by state machine library)
- No partial transitions

### Error Message Quality

- Must be more helpful than raw state machine errors
- Should guide orchestrator to resolution
- Include enough context to understand the problem
- Suggest appropriate tools (`--list`, `--dry-run`)

### Backward Compatibility

- Does not affect auto-determination mode
- Pure addition (new mode, not modification)
- No breaking changes to existing workflows

## Implementation Notes

### TDD Workflow

1. Write test for successful explicit transition
2. Implement basic event validation and firing
3. Write test for guard failure
4. Enhance error messages for guard failures
5. Write test for invalid event
6. Add invalid event error handling
7. Write test for intent-based branching
8. Verify all scenarios work

### Error Message Design

**Philosophy**: Errors should be educational, not just informational.

- State what happened: "Transition blocked"
- Explain why: Guard description
- Provide context: Current state, event, target state
- Suggest solution: Use `--dry-run` to diagnose

### Guard Error Detection

The state machine library returns errors like:
```
stateless: trigger 'planning_complete' is valid for transition from state 'ImplementationPlanning' but guard condition is not met: task descriptions approved
```

We detect this by:
1. Checking if error contains "guard condition is not met"
2. Extracting guard description from SDK
3. Reformatting with helpful context

### Integration with Other Modes

Explicit mode complements other modes:
- Orchestrator runs `--list` to discover options
- Orchestrator chooses event based on judgment
- Orchestrator runs `--dry-run [event]` to validate (optional)
- Orchestrator runs `sow advance [event]` to execute
- If it fails, error message guides to fix

### Testing Strategy

Create various project scenarios:
- Simple transitions (one guard)
- Complex transitions (multiple guards)
- Intent-based branching (multiple options)
- Invalid events (typos, wrong state)
- Guard failures (various conditions)

### Next Steps

After this task:
- All four CLI modes complete (auto, list, dry-run, explicit)
- Task 060+ will refactor standard project to demonstrate usage
- Orchestrators have full control over state transitions
