# Task 060: Update Advance Command (TDD)

# Task 060: Update Advance Command (TDD)

## Overview

Update the existing `sow advance` command to use the SDK's state machine integration via `Project.Advance()`.

## Context

The advance command currently exists at `cli/cmd/advance.go`. It needs to be updated to use the SDK's `Project.Advance()` method, which:
1. Determines the next event using OnAdvance determiners
2. Evaluates guards via `machine.CanFire()`
3. Fires the event (executes OnExit, transition, OnEntry)
4. Updates state

## Design References

- **SDK Advance flow**: `.sow/knowledge/designs/project-sdk-implementation.md` lines 217-262
- **Command spec**: `.sow/knowledge/designs/command-hierarchy-design.md` lines 192-200, 440-446, 877-892
- **Existing implementation**: `cli/cmd/advance.go` (review before modifying)

## Requirements

### Command Behavior

**Syntax**: `sow advance`

**Flow**:
1. Load project using SDK
2. Call `project.Advance()` (handles event determination and guard evaluation)
3. Save updated state
4. Display new state to user

**Error handling**:
- No project exists → clear error
- Guard fails → explain why transition blocked
- No event determiner for current state → explain missing configuration
- Event firing fails → show error details

## TDD Approach

### Step 1: Write Integration Test First

Create `cli/testdata/script/unified_commands/integration/state_transitions.txtar`:

```txtar
# Test: State Machine Transitions via Advance
# Coverage: advance through states, guard evaluation

exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test
exec sow init
exec sow project new --branch feat/test --no-launch "Test state transitions"

# Verify initial state
exec cat .sow/project/state.yaml
stdout 'current_state: PlanningActive'

# Test: Advance fails without required artifact
! exec sow advance
stderr 'cannot fire event'
stderr 'task_list.*not approved'

# Add and approve required artifact
exec mkdir -p .sow/project/planning
exec sh -c 'echo "# Tasks\n- Task 1" > .sow/project/planning/tasks.md'
exec sow output add --type task_list --path planning/tasks.md --phase planning
exec sow output set --index 0 approved true --phase planning

# Test: Advance succeeds with guard satisfied
exec sow advance
stdout 'Advanced to: ImplementationPlanning'
exec cat .sow/project/state.yaml
stdout 'current_state: ImplementationPlanning'

# Test: Advance to next state
exec sow phase set metadata.tasks_approved true --phase implementation
exec sow advance
stdout 'Advanced to: ImplementationExecuting'
exec cat .sow/project/state.yaml
stdout 'current_state: ImplementationExecuting'

# Test: Cannot advance without tasks completed
! exec sow advance
stderr 'cannot fire event'

# Add and complete a task
exec sow task add "Test task" --agent implementer
exec sow task set --id 010 status completed

# Test: Advance with all tasks complete
exec sow advance
stdout 'Advanced to: ReviewActive'
exec cat .sow/project/state.yaml
stdout 'current_state: ReviewActive'
```

### Step 2: Update Command Implementation

Modify `cli/cmd/advance.go` to use SDK.

### Step 3: Run Integration Test

Verify test passes.

## Implementation Details

### Current Command Structure

Review `cli/cmd/advance.go` to understand existing implementation before modifying.

### New Implementation

```go
package cmd

import (
    "fmt"
    "github.com/jmgilman/sow/cli/internal/cmdutil"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/spf13/cobra"
)

func NewAdvanceCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "advance",
        Short: "Progress project to next state",
        Long: `Progress the project through its state machine.

The advance command:
1. Determines the next event based on current state
2. Evaluates guards to ensure transition is allowed
3. Fires the event if guards pass
4. Saves the updated state

Guards may prevent transitions. Common guard failures:
- Planning → Implementation: task_list output not approved
- Implementation Planning → Executing: tasks not approved (metadata.tasks_approved)
- Implementation Executing → Review: not all tasks completed
- Review → Finalize: review not approved or assessment not set`,
        RunE: runAdvance,
    }
}

func runAdvance(cmd *cobra.Command, _ []string) error {
    ctx := cmdutil.GetContext(cmd.Context())

    // Load project
    project, err := state.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load project: %w", err)
    }

    // Get current state for display
    currentState := project.Statechart.Current_state
    fmt.Printf("Current state: %s\n", currentState)

    // Advance (calls OnAdvance determiner, evaluates guards, fires event)
    if err := project.Advance(); err != nil {
        return fmt.Errorf("failed to advance: %w", err)
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

### Error Message Improvements

The SDK's `Project.Advance()` may return generic errors. Consider wrapping them with more helpful context:

```go
if err := project.Advance(); err != nil {
    // Check for specific error types
    if strings.Contains(err.Error(), "cannot fire event") {
        // Guard failure - explain what's missing
        return fmt.Errorf("transition blocked: %w\n\nCheck guards for this state transition", err)
    }
    if strings.Contains(err.Error(), "no event determiner") {
        return fmt.Errorf("cannot advance from state %s: no transition configured\n\nThis may be a terminal state", currentState)
    }
    return fmt.Errorf("failed to advance: %w", err)
}
```

### State Machine Flow

The `Project.Advance()` method (from SDK) handles:

1. **Event Determination**: Calls OnAdvance determiner registered for current state
2. **Guard Evaluation**: Calls `machine.CanFire(event)` to check guards
3. **Event Firing**: If guards pass, calls `machine.Fire(event)` which:
   - Executes OnExit action (if configured)
   - Transitions to new state
   - Executes OnEntry action (if configured)

See `.sow/knowledge/designs/project-sdk-implementation.md` lines 217-262 for detailed flow.

### Testing Guard Failures

Integration test should cover guard failures:

- Planning → Implementation requires task_list output approved
- Implementation Planning → Executing requires metadata.tasks_approved
- Implementation Executing → Review requires all tasks completed
- Review → Finalize requires review approved with assessment

## Files to Modify

### `cli/cmd/advance.go`

Update to use SDK `Project.Advance()`.

## Files to Create

### `cli/testdata/script/unified_commands/integration/state_transitions.txtar`

Integration test covering state transitions and guard evaluation.

## Acceptance Criteria

- [ ] Integration test written first
- [ ] Advance command uses `project.Advance()`
- [ ] State transitions work correctly
- [ ] Guards evaluated properly
- [ ] Guard failures return clear error messages
- [ ] New state displayed after successful advance
- [ ] State persisted via `project.Save()`
- [ ] Integration test passes

## Testing Strategy

**Integration test only** - No unit tests for command logic.

Test scenarios:
1. Advance blocked by guard (missing approval)
2. Advance succeeds after guard satisfied
3. Multiple state transitions in sequence
4. Guard failure for tasks_approved
5. Guard failure for all_tasks_complete
6. Error: no project exists

## Dependencies

- Task 020 (Project Commands) - Required for project state loading

## References

- **Existing implementation**: `cli/cmd/advance.go`
- **SDK Advance method**: `cli/internal/sdks/project/state/project.go` (Advance method)
- **State machine**: `cli/internal/sdks/project/state/project.go` (Machine type)
- **Standard project states**: `cli/internal/projects/standard/states.go`
- **Standard project guards**: `cli/internal/projects/standard/guards.go`
- **Design reference**: `.sow/knowledge/designs/project-sdk-implementation.md` lines 217-262

## Notes

- The advance command is intentionally simple - all complexity lives in the SDK
- Guards are defined in project type configurations (e.g., `cli/internal/projects/standard/`)
- OnAdvance determiners map states to events (single event per state, or conditional)
- State machine handles all phase status updates automatically (OnEntry/OnExit actions)
- No manual phase status management needed - state machine does it
