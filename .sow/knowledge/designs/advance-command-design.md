# State Progression via Advance Command Design

**Author**: Design Orchestrator
**Date**: 2025-11-01
**Status**: Draft
**Type**: Standard Design Doc

## Overview

Implement a unified state progression mechanism where the `sow advance` command becomes the sole entry point for firing state machine events. This replaces the current `PhaseOperationResult` pattern where multiple phase methods can trigger state transitions.

Current state: Phase methods like `Complete()`, `ApproveTasks()`, and others return `PhaseOperationResult` objects containing optional events. The CLI checks for events and fires them, leading to inconsistent event triggering across different operations. This design consolidates all state transitions through a single `Advance()` method per phase, making state progression predictable and explicit.

## Goals and Non-Goals

**Goals**:
- G1: Single command (`sow advance`) triggers all state machine transitions
- G2: Phase `Advance()` methods examine state and determine which event to fire
- G3: Remove `PhaseOperationResult` type and event-returning pattern from all non-Advance methods
- G4: Clear separation between state examination (Advance) and transition validation (Guards)
- G5: Predictable error handling when advance fails due to invalid state or guard failure

**Non-Goals** (explicitly out of scope):
- Command hierarchy simplification (covered in separate design doc)
- Artifact management changes (covered in command hierarchy design)
- Guard logic changes (guards remain as-is, only caller changes)
- State machine transition graph changes (same states and events, different trigger mechanism)

## Background

### Current State

The project lifecycle uses a state machine (via stateless library) with states, events, and guards. Phase methods can trigger state transitions by returning `PhaseOperationResult` objects:

```go
// Current pattern
func (p *PlanningPhase) Complete() (*domain.PhaseOperationResult, error) {
    // ... validation logic ...
    p.state.Status = "completed"
    if err := p.project.Save(); err != nil {
        return nil, err
    }
    // Return event to fire
    return domain.WithEvent(EventCompletePlanning), nil
}

// CLI fires the event
result, err := phase.Complete()
if result.Event != "" {
    machine.Fire(result.Event)
}
```

This pattern is used in:
- `PlanningPhase.Complete()` → fires `EventCompletePlanning`
- `ImplementationPhase.ApproveTasks()` → fires `EventTasksApproved`
- `ImplementationPhase.Complete()` → fires `EventAllTasksComplete`
- `ReviewPhase.Complete()` → fires `EventReviewPass` or `EventReviewFail`
- `FinalizePhase.Complete()` → fires `EventDocumentationDone`, `EventChecksDone`, or no event

### Problems with Current Approach

1. **Inconsistent event sources**: Events triggered from multiple methods (`Complete`, `ApproveTasks`, etc.)
2. **Hidden state transitions**: Not obvious which operations trigger transitions
3. **Tight coupling**: CLI must know which methods return events
4. **Complex decision logic**: Some methods (e.g., `ReviewPhase.Complete`) have conditional event logic
5. **Difficult to reason about**: State transitions scattered across codebase

### Requirements

**Functional Requirements**:
- FR1: All state transitions must go through `Advance()` method
- FR2: `Advance()` must determine correct event based on current state
- FR3: `Advance()` must return clear errors when transition impossible
- FR4: Other phase methods must not trigger state transitions

**Non-Functional Requirements**:
- NFR1: No performance degradation (same number of Redis/file operations)
- NFR2: Clear error messages for debugging
- NFR3: Backward compatibility during migration (gradual rollout)

## Design

### Architecture Overview

```
┌─────────────┐
│ Orchestrator│
│   Agent     │
└──────┬──────┘
       │
       │ sow advance
       │
       v
┌─────────────────────────────────────┐
│     CLI: advance command            │
│  1. Load project                    │
│  2. Get current phase               │
│  3. Call phase.Advance()            │
└──────┬──────────────────────────────┘
       │
       v
┌─────────────────────────────────────┐
│   Phase.Advance() method            │
│  1. Examine current state           │
│  2. Determine event to fire         │
│  3. Call machine.Fire(event)        │
│  4. Return error if failed          │
└──────┬──────────────────────────────┘
       │
       v
┌─────────────────────────────────────┐
│    State Machine                    │
│  1. Check guards                    │
│  2. Execute transition              │
│  3. Update state file               │
│  4. Trigger entry actions           │
└─────────────────────────────────────┘
```

The `Advance()` method becomes the single routing point for state transitions. It examines the current state (both phase state and state machine state) to determine which event should be fired, then fires it. Guards validate whether the transition is allowed (separate concern).

**Key Principle**: Advance() determines WHICH event to fire (state examination). Guards determine WHETHER the transition is allowed (validation).

### Component Breakdown

#### CLI Advance Command

**Location**: `cli/cmd/advance.go` (currently `cli/cmd/agent/advance.go`)

**Responsibility**: Load project, delegate to phase, handle errors

**Implementation**:
```go
func NewAdvanceCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "advance",
        Short: "Advance project to next state",
        RunE: func(cmd *cobra.Command, _ []string) error {
            ctx := cmdutil.GetContext(cmd.Context())

            // Load project
            proj, err := loader.Load(ctx)
            if err != nil {
                return fmt.Errorf("failed to load project: %w", err)
            }

            // Get current phase
            phase := proj.CurrentPhase()
            if phase == nil {
                return fmt.Errorf("no active phase")
            }

            // Delegate to phase
            if err := phase.Advance(); err != nil {
                return fmt.Errorf("failed to advance: %w", err)
            }

            cmd.Println("✓ Advanced to next state")
            return nil
        },
    }
    return cmd
}
```

**Changes from current**:
- No more `PhaseOperationResult` handling
- No event firing at CLI level (moved to phase)
- Simpler error propagation

#### Phase.Advance() Interface

**Location**: `cli/internal/project/domain/phase.go`

**New interface signature**:
```go
type Phase interface {
    // ... other methods ...
    Advance() error
}
```

**Responsibility**: Examine state, determine event, fire it, return error if failed

**Contract**:
- MUST examine current state to determine event
- MUST fire event via `project.Machine().Fire(event)`
- MUST save project state if transition succeeds
- MUST return error if unexpected state detected
- MUST return error if guard fails (event firing failed)
- MAY have decision logic for conditional events

#### PlanningPhase.Advance()

**Location**: `cli/internal/project/standard/planning.go`

**Implementation**:
```go
func (p *PlanningPhase) Advance() error {
    // Planning phase has only one possible transition
    // Fire EventCompletePlanning
    machine := p.project.Machine()
    if err := machine.Fire(EventCompletePlanning); err != nil {
        // Guard failed - task list not approved
        return fmt.Errorf("cannot complete planning: %w", err)
    }

    // Transition succeeded, save state
    if err := p.project.Save(); err != nil {
        return fmt.Errorf("failed to save state: %w", err)
    }

    return nil
}
```

**Logic**: Single possible event, no state examination needed. Guard validates task list approved.

#### ImplementationPhase.Advance()

**Location**: `cli/internal/project/standard/implementation.go`

**Implementation**:
```go
func (p *ImplementationPhase) Advance() error {
    machine := p.project.Machine()
    currentState := machine.State()

    // Determine event based on current state machine state
    var event Event
    switch currentState {
    case ImplementationPlanning:
        // In planning substate - advance to executing
        event = EventTasksApproved
    case ImplementationExecuting:
        // In executing substate - advance to review
        event = EventAllTasksComplete
    default:
        return fmt.Errorf("unexpected state: %s", currentState)
    }

    // Fire determined event
    if err := machine.Fire(event); err != nil {
        return fmt.Errorf("cannot advance from %s: %w", currentState, err)
    }

    // Save state
    if err := p.project.Save(); err != nil {
        return fmt.Errorf("failed to save state: %w", err)
    }

    return nil
}
```

**Logic**: Examines state machine state to determine which substate we're in, fires appropriate event. Guards validate prerequisites (tasks approved or all tasks complete).

#### ReviewPhase.Advance()

**Location**: `cli/internal/project/standard/review.go`

**Implementation**:
```go
func (p *ReviewPhase) Advance() error {
    // Find latest approved review artifact
    var latestReview *Artifact
    for i := len(p.state.Artifacts) - 1; i >= 0; i-- {
        artifact := &p.state.Artifacts[i]
        if artifact.Type == "review" && artifact.Approved {
            latestReview = artifact
            break
        }
    }

    if latestReview == nil {
        return fmt.Errorf("no approved review artifact found")
    }

    // Determine event based on assessment
    var event Event
    switch latestReview.Assessment {
    case "pass":
        event = EventReviewPass
    case "fail":
        event = EventReviewFail
    default:
        return fmt.Errorf("invalid assessment: %s", latestReview.Assessment)
    }

    // Fire determined event
    machine := p.project.Machine()
    if err := machine.Fire(event); err != nil {
        return fmt.Errorf("cannot complete review: %w", err)
    }

    // Save state
    if err := p.project.Save(); err != nil {
        return fmt.Errorf("failed to save state: %w", err)
    }

    return nil
}
```

**Logic**: Examines review assessment to determine pass/fail, fires corresponding event. Guard validates review is approved.

#### FinalizePhase.Advance()

**Location**: `cli/internal/project/standard/finalize.go`

**Implementation**:
```go
func (p *FinalizePhase) Advance() error {
    machine := p.project.Machine()
    currentState := machine.State()

    // Determine event based on finalize substate
    var event Event
    switch currentState {
    case FinalizeDocumentation:
        event = EventDocumentationDone
    case FinalizeChecks:
        event = EventChecksDone
    case FinalizeDelete:
        event = EventProjectDelete
    default:
        return fmt.Errorf("unexpected state: %s", currentState)
    }

    // Fire determined event
    if err := machine.Fire(event); err != nil {
        return fmt.Errorf("cannot advance from %s: %w", currentState, err)
    }

    // Save state
    if err := p.project.Save(); err != nil {
        return fmt.Errorf("failed to save state: %w", err)
    }

    return nil
}
```

**Logic**: Examines finalize substate, fires appropriate progression event. Guards validate each step is complete.

#### Other Phase Methods (Complete, ApproveTasks, etc.)

**Changes**: Remove `PhaseOperationResult` return type, return only errors

**Example - PlanningPhase.Complete()** (REMOVED):
```go
// Old implementation - REMOVE THIS METHOD
func (p *PlanningPhase) Complete() (*domain.PhaseOperationResult, error) {
    // ... validation ...
    return domain.WithEvent(EventCompletePlanning), nil
}
```

The `Complete()` method is removed entirely. Completing planning is now done by:
1. Human/agent approves task list: `sow output set --index 0 approved true`
2. Orchestrator calls: `sow advance`

**Example - ImplementationPhase.ApproveTasks()**:
```go
// Old implementation - returned event
func (p *ImplementationPhase) ApproveTasks() (*domain.PhaseOperationResult, error) {
    // ... set tasks_approved flag ...
    return domain.WithEvent(EventTasksApproved), nil
}

// New implementation - no event
func (p *ImplementationPhase) ApproveTasks() error {
    if p.state.Metadata == nil {
        p.state.Metadata = make(map[string]interface{})
    }
    p.state.Metadata["tasks_approved"] = true

    if err := p.project.Save(); err != nil {
        return fmt.Errorf("failed to save: %w", err)
    }

    return nil  // No event fired
}
```

### Data Flow

**Scenario: Complete Planning Phase**

1. **Setup**:
   - Planning phase active (state: `PlanningActive`)
   - Task list artifact created and approved

2. **User action**:
   - Orchestrator runs: `sow advance`

3. **CLI processing**:
   - Load project from disk
   - Get current phase (returns PlanningPhase)
   - Call `phase.Advance()`

4. **Phase.Advance() execution**:
   - PlanningPhase examines state (no complex logic needed)
   - Determines event: `EventCompletePlanning`
   - Calls `machine.Fire(EventCompletePlanning)`

5. **State machine processing**:
   - Checks guard: `PlanningComplete()` returns true (task list approved)
   - Executes transition: `PlanningActive` → `ImplementationPlanning`
   - Updates statechart in state file
   - Triggers entry action (generates prompt for implementation planning)

6. **Completion**:
   - Phase saves state to disk
   - Returns nil (success)
   - CLI prints success message

**Error Flow: Advance Before Task List Approved**

1. User runs: `sow advance` (task list NOT approved)
2. CLI calls `phase.Advance()`
3. Phase fires `EventCompletePlanning`
4. State machine checks guard: `PlanningComplete()` returns false
5. Guard failure → `machine.Fire()` returns error
6. Phase propagates error: `"cannot complete planning: guard failed"`
7. CLI displays error to user
8. State unchanged (still in `PlanningActive`)

## Error Handling

**Error Categories**:

1. **Unexpected State Error**
   - Occurs when: Phase in state not handled by Advance() logic
   - Example: `ImplementationPhase.Advance()` called but state is `ReviewActive`
   - Message: `"unexpected state: ReviewActive"`
   - Cause: State file corruption or programming error

2. **Guard Failure Error**
   - Occurs when: Transition prerequisites not met
   - Example: `sow advance` in planning but task list not approved
   - Message: `"cannot complete planning: guard PlanningComplete failed"`
   - Cause: User called advance prematurely

3. **State Save Error**
   - Occurs when: File write fails after successful transition
   - Example: Filesystem I/O error
   - Message: `"failed to save state: permission denied"`
   - Cause: Infrastructure issue

**Error Handling Strategy**:
- Advance() returns error immediately on guard failure (no state change)
- Transition atomic: guard check + state update + save (rollback on save failure)
- Clear error messages distinguish between categories
- No silent failures - all errors propagated to CLI

## Implementation Notes

**Type Removal**: Delete `PhaseOperationResult` type from `domain/phase.go`

**Method Signature Changes**: Update all phase methods:
- `Complete()` → removed (replaced by Advance)
- `ApproveTasks()` → returns `error` only (was `(*PhaseOperationResult, error)`)
- `ApproveArtifact()` → returns `error` only
- `Set()` → returns `error` only

**Guard Unchanged**: Guard functions remain as-is. They validate prerequisites but don't determine events.

**Prompt Updates**: Update orchestrator prompts to instruct calling `sow advance` at appropriate times instead of `sow agent complete`

**Command Removal**: Delete `cli/cmd/agent/complete.go` command

## Testing Strategy

**Unit Tests**:

1. **Phase.Advance() Tests** (per phase):
   - Test correct event determination for each state
   - Test error on unexpected state
   - Test guard failure error propagation
   - Test state save after successful transition

2. **Integration Tests**:
   - Full lifecycle test calling only `advance`
   - Test review pass/fail conditional logic
   - Test finalize substates progression
   - Test error recovery (advance fails, retry after fix)

3. **Migration Tests**:
   - Verify old PhaseOperationResult code removed
   - Verify no methods return events except Advance()

**Test Example**:
```go
func TestPlanningPhaseAdvance(t *testing.T) {
    // Setup
    proj := setupTestProject(t)
    phase := proj.Phase("planning")

    // Add and approve task list
    phase.AddArtifact("tasks.md", WithType("task_list"))
    phase.ApproveArtifact("tasks.md")

    // Execute
    err := phase.Advance()

    // Verify
    assert.NoError(t, err)
    assert.Equal(t, ImplementationPlanning, proj.Machine().State())
}

func TestPlanningAdvance_NotApproved_ReturnsError(t *testing.T) {
    // Setup
    proj := setupTestProject(t)
    phase := proj.Phase("planning")
    phase.AddArtifact("tasks.md", WithType("task_list"))
    // Don't approve

    // Execute
    err := phase.Advance()

    // Verify
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "cannot complete planning")
    assert.Equal(t, PlanningActive, proj.Machine().State()) // unchanged
}
```

## Alternatives Considered

### Option 1: Keep PhaseOperationResult, Add Advance() Alongside

**Description**: Add `Advance()` method but keep `PhaseOperationResult` return type for existing methods. Both patterns coexist.

**Pros**:
- Gradual migration path
- Backward compatibility
- Less code churn

**Cons**:
- Two ways to trigger transitions (confusing)
- Still requires CLI to handle events from some methods
- Doesn't solve the fundamental inconsistency problem
- Technical debt accumulates

**Why not chosen**: Fails to achieve goal G3 (remove PhaseOperationResult). Complexity increases instead of decreasing.

### Option 2: CLI Determines Events (No Phase.Advance())

**Description**: CLI examines state and fires events directly. No phase involvement.

**Example**:
```go
// CLI logic
phase := proj.CurrentPhase()
if phase.Name() == "planning" {
    // CLI checks if task list approved
    if taskListApproved(phase) {
        machine.Fire(EventCompletePlanning)
    }
}
```

**Pros**:
- Single point of control in CLI
- No changes to Phase interface
- Explicit state transition logic

**Cons**:
- CLI becomes phase-aware (violates abstraction)
- Decision logic scattered across CLI commands
- Difficult to extend with new project types
- Phase-specific logic leaks into generic CLI layer

**Why not chosen**: Violates separation of concerns. Phases should own their transition logic. CLI should remain generic across project types.

### Option 3: Event Auto-Detection via Guards

**Description**: Remove explicit event firing. State machine automatically detects which event should fire based on guards that return true.

**Example**:
```go
// State machine auto-checks all permitted events
// Fires first event whose guard returns true
machine.AutoAdvance()
```

**Pros**:
- No phase logic needed
- Guards do double duty (validation + detection)
- Simplest phase implementation

**Cons**:
- Guard order matters (non-deterministic if multiple guards true)
- Guards can't distinguish detection vs. validation
- Complex conditional logic (e.g., review pass/fail) impossible
- Debugging difficult (which guard fired which event?)

**Why not chosen**: Conflates detection (which event) with validation (is transition allowed). Loses explicit control over event firing.

## Open Questions

- [x] Should Advance() save state or return event for CLI to save? (Resolved: Advance saves after transition)
- [x] How to handle finalize substates that don't need guard validation? (Resolved: Guards return true, Advance determines event)
- [ ] Should we add `CanAdvance() bool` method to preview if advance would succeed?
- [ ] Error message standardization across all phases

## References

- [Exploration: Simplified Command Hierarchy](../../knowledge/explorations/simplified_command_hierarchy.md)
- [Arc42 Section 8: Cross-cutting Concepts - State Management](../../docs/architecture/08-crosscutting-concepts.md)
- [stateless library documentation](https://github.com/qmuntal/stateless)

## Future Considerations

- **Dry-run mode**: `sow advance --dry-run` previews next state without transitioning
- **Event history**: Log all fired events for debugging
- **Advance suggestions**: CLI suggests when to call advance based on state
- **Multi-step advance**: Advance multiple states in single command (with confirmation)
