# Design Proposal: Enhanced `sow advance` Command

**Status:** Proposal (Refined)
**Date:** 2025-11-06
**Updated:** 2025-11-06
**Context:** Supporting explicit event selection and discovery for branching state transitions

## Problem Statement

The current `sow advance` command assumes linear state progression:
- Always calls `DetermineEvent()` to auto-select the next event
- No way to explicitly choose between multiple transitions (intent-based branching)
- No discovery mechanism for available transitions
- Makes intent-based branching impossible

**Current behavior:**
```bash
sow advance  # Always auto-determines event
```

**Issues:**
- Works fine for linear states and state-determined branching
- Fails for intent-based branching where orchestrator must choose
- No way to see what transitions are available from current state
- Cannot explicitly fire a specific event when needed

## Two Types of Branching (Context)

See `DESIGN_BRANCHING_SDK.md` for full details. Summary:

### Type 1: State-Determined Branching
**Decision discoverable from project state** (e.g., review assessment already set)
- **SDK Solution:** `AddBranch()` with discriminator
- **CLI Behavior:** `sow advance` (no argument) - auto-determines

### Type 2: Intent-Based Branching
**Decision requires orchestrator/user choice** (e.g., "add more research" vs "finalize")
- **SDK Solution:** Multiple `AddTransition()`
- **CLI Behavior:** `sow advance [event]` (explicit event) - orchestrator chooses

## Proposed Solution: Optional Explicit Event Selection

Allow explicit event specification via positional argument, with auto-determination as fallback.

### Command Signature

```bash
sow advance [event]    # Optional positional arg for event
sow advance --list     # List available transitions
sow advance --dry-run [event]  # Validate without executing
```

### Behavior

#### 1. No Argument (Backward Compatible)

```bash
sow advance
```

**For linear states:**
- Calls `DetermineEvent()` to auto-select event
- Fires event and saves
- Works exactly as today

**For state-determined branching (AddBranch):**
- Calls discriminator to examine project state
- Auto-determines which branch to take
- Fires appropriate event and saves

**For intent-based branching (multiple AddTransition):**
- No `DetermineEvent()` registered (returns error)
- Shows available options
- Prompts user to specify event explicitly

#### 2. Explicit Event (New)

```bash
sow advance finalize
sow advance add_more_research
```

**Purpose:** Intent-based branching
- Fires the specified event directly
- Validates event is permitted (guards must pass)
- Saves if successful

**Use cases:**
- Orchestrator decides between multiple valid transitions
- Override auto-determination when needed
- Recovery/debugging

#### 3. List Mode (New)

```bash
sow advance --list
```

**Purpose:** Discovery
- Shows all available transitions from current state
- Uses `machine.PermittedTriggers()` (evaluates guards)
- Includes: event name, target state, description, guard status
- Does not execute any transition

**Use cases:**
- Orchestrator discovers options
- Understanding current state
- Debugging

#### 4. Dry-Run Mode (New)

```bash
sow advance --dry-run finalize
```

**Purpose:** Validation
- Validates the event can be fired (guards pass)
- Shows what would happen (target state, actions)
- Does not execute the transition

**Use cases:**
- Pre-flight checks
- Understanding consequences
- Debugging guard failures

## Implementation

### 1. Command Structure

```go
func NewAdvanceCmd() *cobra.Command {
    var (
        listFlag   bool
        dryRunFlag bool
    )

    cmd := &cobra.Command{
        Use:   "advance [event]",
        Short: "Progress project to next state",
        Long: `Progress the project through its state machine.

Without arguments, automatically determines the next transition using
the project type's event determiner logic (for linear states and
state-determined branching).

With an event argument, explicitly fires that event if valid (for
intent-based branching where orchestrator must choose).

Examples:
  # Auto-advance (linear states, state-determined branching)
  sow advance

  # Explicit event (intent-based branching)
  sow advance finalize
  sow advance add_more_research

  # Discovery
  sow advance --list

  # Validation
  sow advance --dry-run finalize`,
        Args: cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Get context and project
            ctx := cmdutil.GetContext(cmd.Context())
            project, err := state.Load(ctx)
            if err != nil {
                return fmt.Errorf("failed to load project: %w", err)
            }

            machine := project.Machine()
            currentState := project.Statechart.Current_state

            // Handle list mode
            if listFlag {
                return listAvailableTransitions(machine, project, currentState)
            }

            // Get event from argument
            var event sdkstate.Event
            if len(args) > 0 {
                event = sdkstate.Event(args[0])
            }

            // Handle dry-run mode
            if dryRunFlag {
                if event == "" {
                    return fmt.Errorf("--dry-run requires an event argument")
                }
                return validateTransition(machine, project, event, currentState)
            }

            // Execute transition
            if event != "" {
                return executeExplicitTransition(machine, project, event, currentState)
            } else {
                return executeAutoTransition(machine, project, currentState)
            }
        },
    }

    cmd.Flags().BoolVar(&listFlag, "list", false, "List available transitions without executing")
    cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Validate transition without executing")

    return cmd
}
```

### 2. List Available Transitions

Uses `machine.PermittedTriggers()` (already exists in SDK):

```go
func listAvailableTransitions(
    machine *sdkstate.Machine,
    project *state.Project,
    currentState string,
) error {
    fmt.Printf("Current state: %s\n\n", currentState)

    // Get all permitted events (guards evaluated by stateless)
    events, err := machine.PermittedTriggers()
    if err != nil {
        return fmt.Errorf("failed to get permitted triggers: %w", err)
    }

    if len(events) == 0 {
        fmt.Println("No transitions available from current state.")
        fmt.Println("This may be a terminal state.")
        return nil
    }

    fmt.Println("Available transitions:")
    for _, event := range events {
        // Get transition info from config
        targetState := project.Config().GetTargetState(sdkstate.State(currentState), event)
        description := project.Config().GetTransitionDescription(sdkstate.State(currentState), event)
        guardDesc := project.Config().GetGuardDescription(sdkstate.State(currentState), event)

        fmt.Printf("\n  sow advance %s\n", event)
        fmt.Printf("    → %s\n", targetState)
        if description != "" {
            fmt.Printf("    %s\n", description)
        }
        if guardDesc != "" {
            fmt.Printf("    Requires: %s\n", guardDesc)
        }
    }

    fmt.Println()
    return nil
}
```

### 3. Validate Transition (Dry Run)

```go
func validateTransition(
    machine *sdkstate.Machine,
    project *state.Project,
    event sdkstate.Event,
    currentState string,
) error {
    fmt.Printf("Validating transition: %s\n", event)
    fmt.Printf("Current state: %s\n\n", currentState)

    // Check if event is permitted (guard passes)
    canFire, err := machine.CanFire(event)
    if err != nil {
        return fmt.Errorf("failed to check event: %w", err)
    }

    if !canFire {
        // Get guard description for better error message
        guardDesc := project.Config().GetGuardDescription(sdkstate.State(currentState), event)
        fmt.Printf("✗ Transition blocked\n")
        if guardDesc != "" {
            fmt.Printf("  Reason: %s\n", guardDesc)
        }
        return fmt.Errorf("guard condition not met")
    }

    // Get target state
    targetState := project.Config().GetTargetState(sdkstate.State(currentState), event)

    fmt.Printf("✓ Transition is valid\n")
    fmt.Printf("  Event: %s\n", event)
    fmt.Printf("  From: %s\n", currentState)
    fmt.Printf("  To: %s\n", targetState)
    fmt.Println("\nTransition would succeed if executed.")

    return nil
}
```

### 4. Execute Explicit Transition

```go
func executeExplicitTransition(
    machine *sdkstate.Machine,
    project *state.Project,
    event sdkstate.Event,
    currentState string,
) error {
    fmt.Printf("Current state: %s\n", currentState)

    // Get target state for display
    targetState := project.Config().GetTargetState(sdkstate.State(currentState), event)
    if targetState == "" {
        return fmt.Errorf("event %s not configured from state %s", event, currentState)
    }

    fmt.Printf("Firing event: %s\n", event)

    // Fire the event with phase updates
    if err := project.Config().FireWithPhaseUpdates(machine, event, project); err != nil {
        // Enhanced error message
        if strings.Contains(err.Error(), "cannot fire event") {
            return fmt.Errorf("transition blocked: %w\n\nUse --list to see available transitions", err)
        }
        return fmt.Errorf("failed to fire event: %w", err)
    }

    // Save updated state
    if err := project.Save(); err != nil {
        return fmt.Errorf("failed to save state: %w", err)
    }

    fmt.Printf("Advanced to: %s\n", targetState)
    return nil
}
```

### 5. Execute Auto Transition

```go
func executeAutoTransition(
    machine *sdkstate.Machine,
    project *state.Project,
    currentState string,
) error {
    fmt.Printf("Current state: %s\n", currentState)

    // Try to determine event automatically
    event, err := project.Config().DetermineEvent(project)
    if err != nil {
        // Auto-determination failed - this is expected for intent-based branching
        fmt.Printf("\nCannot auto-determine transition: %s\n\n", err)

        // Show available options
        events, _ := machine.PermittedTriggers()
        if len(events) > 0 {
            fmt.Println("Available transitions (choose one):")
            for _, e := range events {
                targetState := project.Config().GetTargetState(sdkstate.State(currentState), e)
                description := project.Config().GetTransitionDescription(sdkstate.State(currentState), e)
                fmt.Printf("  sow advance %s  # → %s", e, targetState)
                if description != "" {
                    fmt.Printf(" (%s)", description)
                }
                fmt.Println()
            }
            return fmt.Errorf("specify event explicitly")
        }

        return fmt.Errorf("no valid transitions from state %s", currentState)
    }

    fmt.Printf("Auto-selected event: %s\n", event)

    // Fire the event with phase updates
    if err := project.Config().FireWithPhaseUpdates(machine, event, project); err != nil {
        if strings.Contains(err.Error(), "cannot fire event") {
            return fmt.Errorf("transition blocked: %w\n\nCheck that prerequisites are met", err)
        }
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

### 6. New ProjectTypeConfig Methods (SDK)

These need to be added to support the CLI:

```go
// TransitionInfo describes a single transition
type TransitionInfo struct {
    Event       Event
    To          State
    Description string
    GuardDesc   string
}

// GetAvailableTransitions returns transitions from state
// Note: This returns configured transitions, not filtered by guards
// Use machine.PermittedTriggers() to get guard-filtered list
func (ptc *ProjectTypeConfig) GetAvailableTransitions(from State) []TransitionInfo

// GetTransitionDescription returns human-readable description
func (ptc *ProjectTypeConfig) GetTransitionDescription(from State, event Event) string

// GetTargetState returns the target state for a transition
func (ptc *ProjectTypeConfig) GetTargetState(from State, event Event) State

// GetGuardDescription returns the guard description for a transition
func (ptc *ProjectTypeConfig) GetGuardDescription(from State, event Event) string
```

## Usage Examples

### Linear State (Unchanged)

```bash
$ sow advance
Current state: ImplementationPlanning
Auto-selected event: planning_complete
Advanced to: ImplementationExecuting
```

### State-Determined Branching (AddBranch)

```bash
# Review assessment is "pass" (already set in artifact metadata)
$ sow advance
Current state: ReviewActive
Auto-selected event: review_pass
Advanced to: FinalizeChecks
```

### Intent-Based Branching (Explicit Event)

```bash
# Orchestrator must decide
$ sow advance
Current state: Researching
Cannot auto-determine transition: no event determiner configured for state Researching

Available transitions (choose one):
  sow advance finalize  # → Finalizing (Complete research and move to finalization phase)
  sow advance add_more_research  # → Planning (Return to planning to add more research topics)
Error: specify event explicitly

# Orchestrator decides to add more research
$ sow advance add_more_research
Current state: Researching
Firing event: add_more_research
Advanced to: Planning
```

### Discovery

```bash
$ sow advance --list
Current state: Researching

Available transitions:

  sow advance finalize
    → Finalizing
    Complete research and move to finalization phase
    Requires: all tasks complete

  sow advance add_more_research
    → Planning
    Return to planning to add more research topics
```

### Validation

```bash
$ sow advance --dry-run finalize
Validating transition: finalize
Current state: Researching

✗ Transition blocked
  Reason: all tasks complete

Error: guard condition not met

# After completing tasks
$ sow advance --dry-run finalize
Validating transition: finalize
Current state: Researching

✓ Transition is valid
  Event: finalize
  From: Researching
  To: Finalizing

Transition would succeed if executed.
```

### Error Cases

```bash
# Guard not met
$ sow advance finalize
Current state: Researching
Firing event: finalize
Error: transition blocked: stateless: trigger 'finalize' is valid for
transition from state 'Researching' but guard condition is not met:
all tasks complete

Use --list to see available transitions

# Invalid event
$ sow advance invalid_event
Current state: Researching
Error: event invalid_event not configured from state Researching

# Terminal state
$ sow advance
Current state: Completed
Error: no valid transitions from state Completed

This may be a terminal state
```

## Backward Compatibility

**Existing workflows unchanged:**
- `sow advance` continues to work for linear states
- `sow advance` continues to work for state-determined branching (AddBranch)
- Auto-determination via `DetermineEvent()` still used when no argument provided
- Error messages enhanced but not breaking

**New capabilities:**
- Explicit event selection for intent-based branching
- Discovery via `--list`
- Validation via `--dry-run`

## Orchestrator Workflow Patterns

### Pattern 1: Linear States

```python
# Simple progression
run("sow advance")
```

### Pattern 2: State-Determined Branching

```python
# Decision already made via artifact metadata
create_review_artifact(assessment="pass")
run("sow output approve --index 0")

# Auto-determines based on assessment
run("sow advance")
```

### Pattern 3: Intent-Based Branching

```python
# Discover options
result = run("sow advance --list")
parse_options(result)

# Orchestrator decides based on conversation, context, etc.
if user_wants_more_research:
    run("sow advance add_more_research")
else:
    run("sow advance finalize")
```

### Pattern 4: Validation Before Action

```python
# Check if transition is valid before executing
result = run("sow advance --dry-run finalize")
if result.success:
    run("sow advance finalize")
else:
    # Handle prerequisites
    complete_remaining_tasks()
    run("sow advance finalize")
```

## Integration with AddBranch

The CLI naturally integrates with both branching patterns:

**State-Determined (AddBranch):**
- Discriminator examines state
- Returns value ("pass", "fail", etc.)
- Auto-determines event
- CLI: `sow advance` (no argument)

**Intent-Based (Multiple AddTransition):**
- No discriminator (can't auto-determine)
- Orchestrator must choose
- CLI: `sow advance [event]` (explicit)

**Discovery works for both:**
- `machine.PermittedTriggers()` returns valid events
- CLI formats with descriptions from transitions
- Shows guard status

## Benefits

1. **Supports both branching types** - Auto for state-determined, explicit for intent-based
2. **Backward compatible** - Zero changes to existing code
3. **Discoverable** - `--list` shows all options with context
4. **Validatable** - `--dry-run` checks before execution
5. **Orchestrator-friendly** - AI can discover and choose options
6. **Scriptable** - Works in automation/CI
7. **Clear intent** - Explicit events make logs readable
8. **Natural for backward** - Just another event to fire

## Testing Strategy

### Unit Tests

```go
func TestAdvanceExplicitEvent(t *testing.T) {
    // Setup project in branching state
    // Execute: sow advance finalize
    // Assert: state transitions correctly
}

func TestAdvanceAutoSelection(t *testing.T) {
    // Setup project in linear state
    // Execute: sow advance
    // Assert: auto-determines and transitions
}

func TestAdvanceList(t *testing.T) {
    // Setup project in branching state
    // Execute: sow advance --list
    // Assert: shows all available transitions with descriptions
}

func TestAdvanceDryRun(t *testing.T) {
    // Setup project
    // Execute: sow advance --dry-run finalize
    // Assert: validates without executing
}

func TestAdvanceGuardFailure(t *testing.T) {
    // Setup project with unmet guard
    // Execute: sow advance finalize
    // Assert: returns helpful error with guard description
}

func TestAdvanceIntentBasedBranching(t *testing.T) {
    // Setup project with multiple AddTransition, no OnAdvance
    // Execute: sow advance (no arg)
    // Assert: shows available options, prompts for explicit event
    // Execute: sow advance [event]
    // Assert: transitions correctly
}
```

### Integration Tests

Test against actual project types:
- Standard project ReviewActive (state-determined)
- Exploration project Researching (intent-based)
- Custom project types with N-way branches
- Terminal states
- Guard failures

## Next Steps

1. ✅ Define command signature and flags
2. ✅ Design behavior for auto vs explicit vs list vs dry-run
3. ✅ Clarify integration with both branching types
4. ✅ Specify required SDK methods (introspection)
5. Update `advance.go` with new implementation
6. Add helper functions for listing, validation, execution
7. Add new methods to ProjectTypeConfig
8. Update tests to cover new functionality
9. Update documentation with examples
10. Coordinate with SDK changes (AddBranch, WithDescription)
11. Test with orchestrator agents (Claude Code)
12. Update agent prompts to use explicit events for intent-based branching
