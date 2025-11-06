# Task 050: Implement Transition Configuration

## Context

This task implements the state machine transition configuration for the exploration project type. Transitions define how the state machine moves between states, including:
- Which events trigger which transitions
- Guard functions that validate transition preconditions
- OnEntry/OnExit actions that mutate state during transitions

The exploration state machine has 3 transitions:
1. Active → Summarizing (intra-phase transition)
2. Summarizing → Finalizing (inter-phase transition)
3. Finalizing → Completed (terminal transition)

Each transition updates phase status and timestamps appropriately.

## Requirements

### Update exploration.go

In `cli/internal/projects/exploration/exploration.go`, implement two functions:

#### 1. configureTransitions Function

Implement `configureTransitions` to configure all state transitions:

```go
func configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder
```

**Transition 1: Active → Summarizing**
- From: `sdkstate.State(Active)`
- To: `sdkstate.State(Summarizing)`
- Event: `sdkstate.Event(EventBeginSummarizing)`
- Guard: `allTasksResolved(p)` (closure capturing project)
- OnEntry action:
  - Update exploration phase status to "summarizing"
  - Store updated phase back to `p.Phases["exploration"]`

**Transition 2: Summarizing → Finalizing**
- From: `sdkstate.State(Summarizing)`
- To: `sdkstate.State(Finalizing)`
- Event: `sdkstate.Event(EventCompleteSummarizing)`
- Guard: `allSummariesApproved(p)` (closure)
- OnExit action:
  - Mark exploration phase as completed
  - Set `phase.Status = "completed"`
  - Set `phase.Completed_at = time.Now()`
  - Store updated phase
- OnEntry action:
  - Enable and activate finalization phase
  - Set `phase.Enabled = true`
  - Set `phase.Status = "in_progress"`
  - Set `phase.Started_at = time.Now()`
  - Store updated phase

**Transition 3: Finalizing → Completed**
- From: `sdkstate.State(Finalizing)`
- To: `sdkstate.State(Completed)`
- Event: `sdkstate.Event(EventCompleteFinalization)`
- Guard: `allFinalizationTasksComplete(p)` (closure)
- OnEntry action:
  - Mark finalization phase as completed
  - Set `phase.Status = "completed"`
  - Set `phase.Completed_at = time.Now()`
  - Store updated phase

**Initial State**
- Set initial state: `builder.SetInitialState(sdkstate.State(Active))`

#### 2. configureEventDeterminers Function

Implement `configureEventDeterminers` to map states to advance events:

```go
func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder
```

Configure event determiners for each advanceable state:

- **Active state**: Returns `EventBeginSummarizing`
- **Summarizing state**: Returns `EventCompleteSummarizing`
- **Finalizing state**: Returns `EventCompleteFinalization`

Pattern:
```go
return builder.
    OnAdvance(sdkstate.State(Active), func(p *state.Project) (sdkstate.Event, error) {
        return sdkstate.Event(EventBeginSummarizing), nil
    }).
    OnAdvance(sdkstate.State(Summarizing), func(p *state.Project) (sdkstate.Event, error) {
        return sdkstate.Event(EventCompleteSummarizing), nil
    }).
    OnAdvance(sdkstate.State(Finalizing), func(p *state.Project) (sdkstate.Event, error) {
        return sdkstate.Event(EventCompleteFinalization), nil
    })
```

### Add Required Import

Add `"time"` to imports in `exploration.go` for timestamp handling.

## Test-Driven Development

This task follows TDD methodology:

1. **Write tests first** for:
   - `configureTransitions()` returns non-nil builder
   - Transition configuration:
     - All 3 transitions registered
     - Correct from/to states for each transition
     - Guards attached to transitions
   - `configureEventDeterminers()` returns non-nil builder
   - Event determiner behavior:
     - Active state returns EventBeginSummarizing
     - Summarizing state returns EventCompleteSummarizing
     - Finalizing state returns EventCompleteFinalization
   - OnEntry/OnExit actions:
     - Phase status updates correctly
     - Timestamps set appropriately
     - enabled flags updated

2. **Run tests** - they should fail initially (red phase)

3. **Implement functionality** - configure transitions and determiners (green phase)

4. **Refactor** - ensure action logic is clean and maintainable

Add tests to `cli/internal/projects/exploration/exploration_test.go`.

## Acceptance Criteria

- [ ] `configureTransitions` function implemented
- [ ] **Unit tests written before implementation**
- [ ] Tests verify transition configuration and behavior
- [ ] All tests pass
- [ ] All 3 transitions configured with correct from/to states
- [ ] All transitions have correct events
- [ ] All transitions have guard functions via closures
- [ ] Active → Summarizing updates exploration phase status
- [ ] Summarizing → Finalizing has both OnExit and OnEntry actions
- [ ] OnExit marks exploration phase complete with timestamp
- [ ] OnEntry enables finalization phase with timestamp
- [ ] Finalizing → Completed marks finalization complete
- [ ] Initial state set to Active
- [ ] `configureEventDeterminers` function implemented
- [ ] Event determiners configured for all 3 advanceable states
- [ ] Event determiners return correct events
- [ ] `time` package imported
- [ ] Code follows Go formatting standards (gofmt)
- [ ] No compilation errors

## Technical Details

### Transition Configuration Pattern

Transitions use the builder's `AddTransition` method with optional configuration:

```go
builder.AddTransition(
    from,
    to,
    event,
    project.WithGuard(guardFunc),
    project.WithOnEntry(entryAction),
    project.WithOnExit(exitAction),
)
```

Guards and actions are optional. Transitions can have:
- Just a guard (validate but no side effects)
- Just actions (no validation)
- Both guard and actions
- Neither (unconditional transition)

### Guard Closure Pattern

Guards are wrapped in closures that capture the project instance:

```go
project.WithGuard(func(p *state.Project) bool {
    return allTasksResolved(p)
})
```

This binds the guard to the project instance during machine construction.

### Phase Update Pattern

To update a phase:
1. Read phase from map: `phase := p.Phases["exploration"]`
2. Modify phase struct: `phase.Status = "completed"`
3. Write back to map: `p.Phases["exploration"] = phase`

This is necessary because map values are not addressable in Go.

### Intra-Phase vs Inter-Phase Transitions

- **Intra-phase**: Active → Summarizing (both in exploration phase)
  - Only updates phase status, not enabled/timestamps
- **Inter-phase**: Summarizing → Finalizing (crosses phase boundary)
  - Completes old phase, enables new phase
  - Updates timestamps appropriately

### Event Determiners

Event determiners tell the state machine which event to fire when `sow project advance` is called. They examine project state and return the appropriate event.

For exploration, the mapping is simple (one event per state). More complex project types might examine state to determine which of multiple events to fire.

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/standard.go` - Reference transition configuration (lines 84-195)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/builder.go` - Builder API
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/options.go` - Transition options
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/knowledge/designs/project-modes/exploration-design.md` - Design specification (lines 209-306)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/project/context/issue-36.md` - Requirements

## Examples

### Standard Project Transition (Reference)

From `cli/internal/projects/standard/standard.go:109-116`:

```go
AddTransition(
    sdkstate.State(ImplementationExecuting),
    sdkstate.State(ReviewActive),
    sdkstate.Event(EventAllTasksComplete),
    project.WithGuard(func(p *state.Project) bool {
        return allTasksComplete(p)
    }),
)
```

### Transition with OnEntry/OnExit (Reference)

From `cli/internal/projects/standard/standard.go:128-171`:

```go
AddTransition(
    sdkstate.State(ReviewActive),
    sdkstate.State(ImplementationPlanning),
    sdkstate.Event(EventReviewFail),
    project.WithGuard(func(p *state.Project) bool {
        return latestReviewApproved(p)
    }),
    project.WithOnExit(func(p *state.Project) error {
        return state.MarkPhaseFailed(p, "review")
    }),
    project.WithOnEntry(func(p *state.Project) error {
        // Complex rework logic...
        return nil
    }),
)
```

### Phase Update Pattern

```go
project.WithOnEntry(func(p *state.Project) error {
    // Read phase
    phase := p.Phases["finalization"]

    // Modify
    phase.Enabled = true
    phase.Status = "in_progress"
    phase.Started_at = time.Now()

    // Write back
    p.Phases["finalization"] = phase

    return nil
})
```

### Event Determiner Pattern

From `cli/internal/projects/standard/standard.go:197-250`:

```go
OnAdvance(sdkstate.State(ReviewActive), func(p *state.Project) (sdkstate.Event, error) {
    // Complex: examine review assessment
    phase, exists := p.Phases["review"]
    if !exists {
        return "", fmt.Errorf("review phase not found")
    }

    // Find latest review and check assessment
    // Return EventReviewPass or EventReviewFail based on assessment
})
```

For exploration, determiners are simpler (no branching logic needed).

## Dependencies

- Task 010 (Package structure) - Provides `exploration.go`
- Task 020 (States and events) - Provides state and event constants
- Task 030 (Phase configuration) - Defines phases being transitioned
- Task 040 (Guards) - Provides guard functions
- Will be used by state machine construction in SDK

## Constraints

- Must configure all 3 transitions exactly as designed
- Guards must use closures to capture project reference
- Phase updates must follow read-modify-write pattern
- Timestamps must use `time.Now()` at transition time
- Event determiners must not return errors for exploration (simple mapping)
- Initial state must be Active (exploration starts immediately)
- Must handle both intra-phase and inter-phase transitions correctly
- OnEntry/OnExit actions must return error (even if nil)
