# Statechart Summary

**Created**: 2025-10-17
**Purpose**: Formal state machine for project lifecycle management

---

## Overview

This package implements the sow project lifecycle as a formal statechart using the `github.com/qmuntal/stateless` library. The statechart acts as the **authoritative workflow engine**, guiding the orchestrator through project phases and outputting contextual prompts at each step.

## Design Philosophy

### State-Driven Workflow

Rather than having the orchestrator make complex decisions about "what comes next," the statechart **tells it**. Each state outputs a prompt that:
- Explains the current context
- Describes responsibilities (subservient vs autonomous mode)
- Provides specific next actions
- References relevant documentation

The orchestrator simply **responds to the current prompt** and fires events to advance.

### Railroad Architecture

The orchestrator walks a guided railroad:
1. CLI command triggers state machine event
2. State machine transitions (if guards pass)
3. Entry action outputs contextual prompt
4. Orchestrator reads prompt and executes instructions
5. Next CLI command fires next event
6. Repeat

No "master planning" - just step-by-step progression through defined states.

---

## State Tree

```
NoProject
  ↓ [project init]
DiscoveryDecision
  ↓ [enable discovery] OR [skip discovery]
DiscoveryActive → DesignDecision
  ↓ [enable design] OR [skip design]
DesignActive → ImplementationPlanning
  ↓ [task created]
ImplementationExecuting
  ↓ [all tasks complete]
ReviewActive
  ↓ [review pass]                    ↓ [review fail]
FinalizeDocumentation              ImplementationPlanning (loop back)
  ↓ [documentation done]
FinalizeChecks
  ↓ [checks done]
FinalizeDelete
  ↓ [project delete]
NoProject (cycle complete)
```

### Key Characteristics

- **11 states total**
- **Optional phases**: Discovery and Design (can be skipped)
- **Required phases**: Implementation, Review, Finalize (always execute)
- **One loop-back**: Review can return to ImplementationPlanning
- **Mode transitions**: Subservient (Discovery/Design) → Autonomous (Implementation/Review/Finalize)

---

## File Structure

```
internal/statechart/
├── SUMMARY.md              # This file
├── states.go               # State definitions with mode helpers
├── events.go               # Event/trigger definitions
├── guards.go               # Guard functions + ProjectState struct
├── prompts.go              # Contextual prompt generation
├── machine.go              # State machine configuration
├── persistence.go          # Load/Save to state.yaml
├── machine_test.go         # Comprehensive tests
```

## Components

### States (`states.go`)

Defines 11 states as constants with helper methods:
- `IsSubservientMode()` - Discovery/Design phases
- `IsAutonomousMode()` - Implementation/Review/Finalize phases

### Events (`events.go`)

Defines triggers that cause transitions:
- User commands: `EventProjectInit`, `EventEnableDiscovery`, etc.
- Internal transitions: `EventTaskCreated`, `EventAllTasksComplete`, etc.

### Guards (`guards.go`)

Conditional logic that controls transitions:
- `ArtifactsApproved()` - Discovery/Design completion requires approval
- `HasAtLeastOneTask()` - Can't execute without tasks
- `AllTasksComplete()` - Review requires completed work
- `DocumentationAssessed()` - Finalize progression gates
- `ProjectDeleted()` - Mandatory cleanup before completion

Also defines `ProjectState` struct (minimal representation for guards).

### Prompts (`prompts.go`)

Generates contextual prompts for each state using template data:
- Current state context
- Task/artifact counts
- Mode (subservient vs autonomous)
- Specific next actions
- Documentation references

### Machine (`machine.go`)

Configures the state machine:
- `NewMachine()` - Creates machine (infers initial state)
- `NewMachineAt()` - Creates machine at specific state (for persistence)
- `configure()` - Defines all transitions, guards, entry actions
- `Fire()` - Triggers events
- `State()` - Returns current state

### Persistence (`persistence.go`)

Simple file-based persistence:
- `Load()` - Reads `.sow/project/state.yaml`, creates machine at stored state
- `Save()` - Writes state atomically (temp + rename)

State is explicitly stored in YAML:
```yaml
_statechart:
  current_state: "ImplementationExecuting"

phases:
  # ... project data
```

---

## Integration with CLI

### Command Pattern

Every CLI command that affects project state follows this pattern:

```go
func someCommand(args) error {
    // 1. Load machine from disk
    machine, err := statechart.Load()
    if err != nil {
        return err
    }

    // 2. Update project state data (if needed)
    machine.projectState.Phases.Something = newValue

    // 3. Fire event (triggers transition + outputs prompt)
    if err := machine.Fire(statechart.EventSomething); err != nil {
        return err
    }

    // 4. Save updated state
    return machine.Save()
}
```

### Example: Project Initialization

```go
func projectInit(name, desc string) error {
    machine, err := statechart.Load() // Returns NoProject state
    if err != nil {
        return err
    }

    // Create initial project state
    machine.projectState = &statechart.ProjectState{
        Phases: /* initialize phases */,
    }

    // Fire event - outputs "Discovery Decision" prompt
    if err := machine.Fire(statechart.EventProjectInit); err != nil {
        return err
    }

    return machine.Save()
}
```

### Example: Phase Completion

```go
func completeDiscovery() error {
    machine, err := statechart.Load()
    if err != nil {
        return err
    }

    // Fire event - guard checks artifacts approved
    if err := machine.Fire(statechart.EventCompleteDiscovery); err != nil {
        return fmt.Errorf("cannot complete: %w", err)
    }

    return machine.Save()
}
```

### Output to User

When events fire, entry actions output prompts to stdout. Example:

```
$ sow project init "add-auth" --description "Add JWT authentication"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DISCOVERY PHASE DECISION

Determine if discovery phase is warranted for this work.

APPROACH:
  1. Consider the Discovery Worthiness Rubric:
     - Context availability (0-2)
     - Problem clarity (0-2)
     ...

NEXT ACTION:
  If discovery needed:
    sow project phase enable discovery --type <type>

  If discovery not needed:
    (Will auto-transition to design decision)

Reference: PROJECT_LIFECYCLE.md (Discovery Rubric)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

The orchestrator reads this prompt and knows exactly what to do next.

---

## Benefits

### 1. Single Source of Truth

The statechart is the definitive workflow. No ambiguity about "what state am I in?" or "what comes next?"

### 2. No Complex Logic in Orchestrator

The orchestrator doesn't need complex conditionals like:
```go
if discovery.enabled && discovery.completed && !design.enabled {
    // What do I do now?
}
```

It just reads the prompt and follows instructions.

### 3. Testable Workflow

The entire lifecycle can be tested in isolation:
- Verify state transitions
- Test guards prevent invalid transitions
- Ensure prompts are contextual
- Validate persistence round-trips

### 4. Easy to Modify

Want to add a new phase? Change the loop-back behavior? Modify the statechart configuration - all logic is centralized.

### 5. Self-Documenting

The statechart configuration reads like documentation:
```go
m.sm.Configure(ReviewActive).
    Permit(EventReviewFail, ImplementationPlanning). // Loop back to re-plan
    Permit(EventReviewPass, FinalizeDocumentation).
```

Clear and explicit.

### 6. Explicit State Storage

State is explicitly tracked in `_statechart.current_state` field. No complex inference logic needed.

---

## Future Considerations

### Additional States

If workflow becomes more complex, new states can be added:
- `ImplementationAwaitingApproval` (for pending task additions)
- `ReviewAwaitingHumanInput` (waiting for approval)
- Substates for finalize steps

### Hierarchical States

The `stateless` library supports hierarchical states (substates) for shared behavior:
```go
m.sm.Configure(ImplementationPlanning).
    SubstateOf(ImplementationPhase)

m.sm.Configure(ImplementationExecuting).
    SubstateOf(ImplementationPhase)
```

### State Visualization

Generate diagrams from the statechart:
```bash
sow project diagram > workflow.dot
dot -Tpng workflow.dot > workflow.png
```

The `stateless` library's `ToGraph()` method can export to Graphviz format.

---

## References

- **[stateless library](https://github.com/qmuntal/stateless)** - Go port of C# Stateless
- **[docs/PROJECT_LIFECYCLE.md](../../../docs/PROJECT_LIFECYCLE.md)** - Phase specifications
- **[docs/PHASES/](../../../docs/PHASES/)** - Individual phase documentation
- **[docs/CLI_REFERENCE.md](../../../docs/CLI_REFERENCE.md)** - CLI commands

---

## Testing

Run tests:
```bash
go test ./internal/statechart/... -v
```

Tests cover:
- Full lifecycle (NoProject → Finalize → NoProject)
- Discovery/Design phase workflows
- Review loop-back behavior
- Guard validation
- Persistence (save/load round-trips)

All tests use the actual state machine - no mocks. This ensures the real workflow is tested.
