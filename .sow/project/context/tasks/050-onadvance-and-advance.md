# Task 050: OnAdvance Configuration and Project.Advance()

## Objective

Implement OnAdvance configuration and the generic `Project.Advance()` method. This enables the generic `sow advance` command to work across all project types by delegating event determination to project-type-specific logic.

## Context

**Design Reference:** See `.sow/knowledge/designs/project-sdk-implementation.md`:
- Section "State Management - Event Determination (OnAdvance)" (lines 305-327) for concept
- Section "Event Determination" (lines 623-627 for builder, 777-827 for usage example)
- Section "Data Flow - State Machine Flow" (lines 1123-1175) for complete flow

**Why OnAdvance is Needed:**
Different states require different logic to determine the next event:
- **Simple states:** Single possible event (e.g., PlanningActive → EventCompletePlanning)
- **Complex states:** Conditional logic based on state (e.g., ReviewActive → EventReviewPass OR EventReviewFail based on assessment)

OnAdvance allows project types to configure per-state event determination logic, enabling the generic `sow advance` command.

**Prerequisite:** Task 040 completed (BuildMachine creates machine with bound guards)

**What This Task Builds:**
- `GetEventDeterminer()` method to retrieve configured event determiners
- `Project.Advance()` method that uses determiners to advance state generically

## Requirements

### 1. GetEventDeterminer Method

Modify `cli/internal/sdks/project/state/registry.go`:

Add method to ProjectTypeConfig:
```go
// GetEventDeterminer returns the configured event determiner for the given state.
// Returns nil if no determiner is configured for the state.
//
// Event determiners are configured via builder's OnAdvance() method:
//   .OnAdvance(ReviewActive, func(p *Project) (Event, error) {
//       // Examine state and determine event
//   })
func (ptc *ProjectTypeConfig) GetEventDeterminer(state State) EventDeterminer {
    return ptc.onAdvance[state]
}
```

### 2. Project.Advance Method

Modify `cli/internal/sdks/project/state/project.go`:

Add method to Project:
```go
// Advance progresses the project to its next state using configured event determination.
//
// The Advance flow:
//  1. Get current state from machine
//  2. Look up event determiner for current state (from OnAdvance config)
//  3. Call determiner to get next event
//  4. Check if transition is allowed (guard evaluation via machine.CanFire)
//  5. Fire event if allowed (executes OnExit, transition, OnEntry)
//
// Returns error if:
//  - No event determiner configured for current state
//  - Determiner fails to determine event
//  - Guard prevents transition (CanFire returns false)
//  - Event firing fails
func (p *Project) Advance() error {
    // 1. Get current state
    currentState := p.machine.State()

    // 2. Get event determiner for current state
    determiner := p.config.GetEventDeterminer(currentState)
    if determiner == nil {
        return fmt.Errorf("no event determiner for state: %s", currentState)
    }

    // 3. Determine next event based on project state
    event, err := determiner(p)
    if err != nil {
        return fmt.Errorf("failed to determine event: %w", err)
    }

    // 4. Check if transition is allowed (guard evaluation)
    can, err := p.machine.CanFire(event)
    if err != nil {
        return err
    }
    if !can {
        return fmt.Errorf("cannot fire event %s from state %s", event, currentState)
    }

    // 5. Fire the event (executes OnExit, transition, OnEntry)
    if err := p.machine.Fire(event); err != nil {
        return fmt.Errorf("failed to fire event: %w", err)
    }

    return nil
}
```

## Files to Modify

1. `cli/internal/sdks/project/state/registry.go` - Add GetEventDeterminer() method
2. `cli/internal/sdks/project/state/project.go` - Add Advance() method
3. `cli/internal/sdks/project/state/advance_test.go` (create) - Behavioral tests

## Testing Requirements (TDD)

Create `cli/internal/sdks/project/state/advance_test.go`:

**GetEventDeterminer Tests:**
- GetEventDeterminer() returns configured determiner for state
- GetEventDeterminer() returns nil for state without determiner
- Multiple states can have different determiners

**Advance Tests:**
- Advance() calls event determiner for current state
- Advance() fires determined event
- Advance() transitions to new state
- Advance() returns error if no determiner configured
- Advance() returns error if determiner fails
- Advance() returns error if guard prevents transition
- Advance() returns error if event firing fails

**Integration Test - Full Advance Flow:**
```go
func TestAdvanceFullFlow(t *testing.T) {
    // Setup: Create project with config
    project := &Project{
        ProjectState: schemas.ProjectState{
            Phases: map[string]schemas.PhaseState{
                "test": {
                    Outputs: []schemas.ArtifactState{
                        {Type: "result", Approved: true},
                    },
                },
            },
        },
    }

    config := &ProjectTypeConfig{
        transitions: []TransitionConfig{
            {
                From:  State("start"),
                To:    State("middle"),
                Event: Event("advance"),
                guardTemplate: func(p *Project) bool {
                    return p.PhaseOutputApproved("test", "result")
                },
            },
        },
        onAdvance: map[State]EventDeterminer{
            State("start"): func(p *Project) (Event, error) {
                return Event("advance"), nil
            },
        },
    }

    project.config = config
    project.machine = config.BuildMachine(project, State("start"))

    // Execute: Call Advance()
    err := project.Advance()

    // Verify: No error, state changed
    if err != nil {
        t.Fatalf("Advance() failed: %v", err)
    }
    if project.machine.State() != State("middle") {
        t.Errorf("expected state=middle, got %s", project.machine.State())
    }
}
```

**Test Pattern - Determiner Failure:**
```go
func TestAdvanceDeterminerError(t *testing.T) {
    config := &ProjectTypeConfig{
        onAdvance: map[State]EventDeterminer{
            State("start"): func(p *Project) (Event, error) {
                return "", fmt.Errorf("cannot determine event")
            },
        },
    }

    project := &Project{config: config}
    project.machine = config.BuildMachine(project, State("start"))

    err := project.Advance()
    if err == nil {
        t.Error("expected error when determiner fails")
    }
}
```

**Test Pattern - Guard Blocks:**
```go
func TestAdvanceGuardBlocks(t *testing.T) {
    config := &ProjectTypeConfig{
        transitions: []TransitionConfig{
            {
                From:  State("start"),
                To:    State("end"),
                Event: Event("advance"),
                guardTemplate: func(p *Project) bool {
                    return false // Always block
                },
            },
        },
        onAdvance: map[State]EventDeterminer{
            State("start"): func(p *Project) (Event, error) {
                return Event("advance"), nil
            },
        },
    }

    project := &Project{config: config}
    project.machine = config.BuildMachine(project, State("start"))

    err := project.Advance()
    if err == nil {
        t.Error("expected error when guard blocks transition")
    }
}
```

## Acceptance Criteria

- [ ] GetEventDeterminer() returns configured determiner for state
- [ ] GetEventDeterminer() returns nil for unconfigured state
- [ ] Advance() gets current state from machine
- [ ] Advance() calls event determiner for current state
- [ ] Advance() calls determiner function with project
- [ ] Advance() checks if transition allowed via CanFire()
- [ ] Advance() fires event if allowed
- [ ] Advance() executes complete transition (OnExit, transition, OnEntry)
- [ ] Advance() returns error if no determiner configured
- [ ] Advance() returns error if determiner fails
- [ ] Advance() returns error if guard blocks
- [ ] All tests pass (100% coverage of advance behavior)
- [ ] Code compiles without errors

## Dependencies

**Required:** Task 040 (BuildMachine creates machine with guards)

## Technical Notes

- Import Event and State from `cli/internal/sdks/state`
- The machine field on Project is already populated by Load() → BuildMachine()
- GetEventDeterminer is simple map lookup (O(1))
- Advance() is the generic command implementation - all project types use same logic
- Error messages should include state name for debugging
- CanFire() already evaluates guards (implemented in state machine SDK)

**Error Handling Strategy:**
- No determiner = user error (config incomplete)
- Determiner fails = business logic error (invalid state, missing data)
- Guard blocks = expected flow (conditions not met yet)
- Fire fails = unexpected error (state machine bug)

**OnAdvance Configuration Example:**
```go
// Simple case - single event from state
.OnAdvance(PlanningActive, func(p *Project) (Event, error) {
    return EventCompletePlanning, nil
})

// Complex case - conditional logic
.OnAdvance(ReviewActive, func(p *Project) (Event, error) {
    phase, _ := p.Phases.Get("review")
    assessment := phase.Outputs[0].Metadata["assessment"]

    if assessment == "pass" {
        return EventReviewPass, nil
    }
    return EventReviewFail, nil
})
```

## Estimated Time

2 hours
