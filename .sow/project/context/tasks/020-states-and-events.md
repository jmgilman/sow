# Task 020: Define States and Events

## Context

This task defines the state machine states and events for the exploration project type. The exploration workflow progresses through 4 states across 2 phases:

1. **Active** (exploration phase) - Active research
2. **Summarizing** (exploration phase) - Synthesizing findings
3. **Finalizing** (finalization phase) - Moving artifacts and cleanup
4. **Completed** (terminal state) - Exploration finished

Events trigger transitions between these states. The state machine is built using the SDK state machine, which uses typed constants for type safety.

## Requirements

### Create States File

Create `cli/internal/projects/exploration/states.go` with:

1. **Package declaration and imports**:
   ```go
   package exploration

   import "github.com/jmgilman/sow/cli/internal/sdks/state"
   ```

2. **State constants**:
   ```go
   const (
       // Active indicates active research phase
       Active = state.State("Active")

       // Summarizing indicates synthesis/summarizing phase
       Summarizing = state.State("Summarizing")

       // Finalizing indicates finalization in progress
       Finalizing = state.State("Finalizing")

       // Completed indicates exploration finished
       Completed = state.State("Completed")
   )
   ```

### Create Events File

Create `cli/internal/projects/exploration/events.go` with:

1. **Package declaration and imports**:
   ```go
   package exploration

   import "github.com/jmgilman/sow/cli/internal/sdks/state"
   ```

2. **Event constants with documentation**:
   ```go
   const (
       // EventBeginSummarizing transitions from Active to Summarizing
       // Fired when all research topics are resolved
       EventBeginSummarizing = state.Event("begin_summarizing")

       // EventCompleteSummarizing transitions from Summarizing to Finalizing
       // Fired when all summary artifacts are approved
       EventCompleteSummarizing = state.Event("complete_summarizing")

       // EventCompleteFinalization transitions from Finalizing to Completed
       // Fired when all finalization tasks are completed
       EventCompleteFinalization = state.Event("complete_finalization")
   )
   ```

## Test-Driven Development

This task follows TDD methodology:

1. **Write tests first** for:
   - State constants have correct values
   - Event constants have correct values
   - Constants are the correct types (state.State and state.Event)

2. **Run tests** - they should fail initially (red phase)

3. **Implement functionality** - define constants (green phase)

4. **Refactor** - ensure proper documentation and formatting

Place tests in `cli/internal/projects/exploration/states_test.go` and `events_test.go`.

## Acceptance Criteria

- [ ] File `states.go` exists with 4 state constants
- [ ] File `events.go` exists with 3 event constants
- [ ] **Unit tests written before implementation**
- [ ] Tests verify constant values and types
- [ ] All tests pass
- [ ] All constants use correct types (`state.State` and `state.Event`)
- [ ] State names match design: "Active", "Summarizing", "Finalizing", "Completed"
- [ ] Event names are descriptive and follow naming convention
- [ ] All constants have clear documentation comments
- [ ] Code follows Go formatting standards (gofmt)
- [ ] No compilation errors

## Technical Details

### State vs Event Types

The SDK uses typed wrappers around strings for type safety:
- `state.State` - Represents a state machine state
- `state.Event` - Represents a state machine event

These are type aliases defined in `cli/internal/sdks/state/` that provide compile-time type checking.

### State Naming Convention

States use PascalCase without prefixes:
- "Active" (not "ExplorationActive")
- "Summarizing" (not "exploration_summarizing")

The state machine context provides the "exploration" namespace. This keeps state names clean and readable.

### Event Naming Convention

Events use descriptive verbs with snake_case:
- "begin_summarizing" (transition action)
- "complete_summarizing" (completion action)
- "complete_finalization" (completion action)

Event names describe what triggers the transition, not the outcome.

### State Machine Flow

```
Active
  │ EventBeginSummarizing
  ▼
Summarizing
  │ EventCompleteSummarizing
  ▼
Finalizing
  │ EventCompleteFinalization
  ▼
Completed
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/states.go` - Reference states implementation
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/events.go` - Reference events implementation
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/knowledge/designs/project-modes/exploration-design.md` - Design specification (lines 310-354 for states, events)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/project/context/issue-36.md` - Requirements

## Examples

### Standard Project States (Reference)

From `cli/internal/projects/standard/states.go`:

```go
package standard

import (
    "github.com/jmgilman/sow/cli/internal/sdks/state"
)

const (
    NoProject = state.State("NoProject")
    ImplementationPlanning = state.State("ImplementationPlanning")
    ImplementationExecuting = state.State("ImplementationExecuting")
    ReviewActive = state.State("ReviewActive")
    FinalizeChecks = state.State("FinalizeChecks")
    FinalizePRCreation = state.State("FinalizePRCreation")
    FinalizeCleanup = state.State("FinalizeCleanup")
)
```

### Standard Project Events (Reference)

From `cli/internal/projects/standard/events.go`:

```go
package standard

import (
    "github.com/jmgilman/sow/cli/internal/sdks/state"
)

const (
    // EventProjectInit creates new project.
    // Transition: NoProject → ImplementationPlanning.
    EventProjectInit = state.Event("project_init")

    // EventPlanningComplete completes planning and task breakdown.
    // Transition: ImplementationPlanning → ImplementationExecuting.
    // Guard: all task description files approved.
    EventPlanningComplete = state.Event("planning_complete")

    // ... more events
)
```

## Dependencies

- Task 010 (Package structure) must be complete
- These constants will be used in Task 030 (Phase configuration) and Task 040 (Transitions)

## Constraints

- State names must exactly match design specification
- Event names must be descriptive and follow snake_case convention
- Must not add extra states or events beyond the design
- Comments must clearly explain when each event fires
- Type safety is critical - use correct SDK types
