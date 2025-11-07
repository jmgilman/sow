# CLI Design: Enhanced `sow advance` Command with Event Selection

**Status**: Design
**Date**: 2025-11-06
**Context**: Supporting explicit event selection and discovery for branching state transitions
**Related**: SDK Design (AddBranch API), Implementation Guide

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Problem Statement](#problem-statement)
3. [Design Principles](#design-principles)
4. [Solution Overview](#solution-overview)
5. [Command Specification](#command-specification)
6. [Behavior Modes](#behavior-modes)
7. [Implementation Details](#implementation-details)
8. [Usage Examples](#usage-examples)
9. [Orchestrator Integration](#orchestrator-integration)
10. [Error Handling](#error-handling)
11. [Backward Compatibility](#backward-compatibility)
12. [Testing Strategy](#testing-strategy)

---

## Executive Summary

This document specifies enhancements to the `sow advance` command to support:

1. **Explicit event selection** - For intent-based branching where orchestrator must choose
2. **Transition discovery** - List available transitions with `--list` flag
3. **Dry-run validation** - Validate transitions without executing with `--dry-run` flag

**Key Changes**:
- Add optional `[event]` positional argument
- Add `--list` flag for discovering available transitions
- Add `--dry-run` flag for pre-flight validation
- Maintain backward compatibility (no argument = auto-determine)

**Integration with SDK**:
- Auto-determination works for linear states and state-determined branching (AddBranch)
- Explicit events work for intent-based branching (multiple AddTransition)
- Discovery leverages new SDK introspection methods

---

## Problem Statement

### Current Limitations

The current `sow advance` command only supports auto-determination:

```bash
sow advance  # Always calls DetermineEvent(), fires that event
```

**Works for**:
- Linear states (one transition from each state)
- State-determined branching (AddBranch with discriminator)

**Fails for**:
- Intent-based branching (orchestrator must choose between valid options)
- Discovery (no way to see available transitions)
- Debugging (can't validate transition before executing)

### Example: Intent-Based Branching

Exploration project in `Researching` state:
- Orchestrator can choose to `finalize` (research complete)
- Or choose to `add_more_research` (need more investigation)
- Decision cannot be determined from project state
- Requires external orchestrator judgment

**Current behavior**: `sow advance` would error (no DetermineEvent configured)

**Needed behavior**: Orchestrator explicitly selects: `sow advance finalize`

---

## Design Principles

1. **Backward Compatible**: Existing workflows unchanged
2. **Discoverable**: CLI can show available options
3. **Explicit over Implicit**: Clear intent for orchestrators
4. **Safe**: Validate before executing
5. **Consistent**: Works across all project types
6. **Informative**: Clear error messages with suggestions

---

## Solution Overview

### Three Operation Modes

1. **Auto-determination** (existing): `sow advance`
   - For linear states and state-determined branching
   - Calls `DetermineEvent()`, fires that event

2. **Explicit event** (new): `sow advance [event]`
   - For intent-based branching
   - Fires specified event directly

3. **Discovery** (new): `sow advance --list`
   - Shows all available transitions
   - Includes descriptions, target states, guard status

4. **Dry-run** (new): `sow advance --dry-run [event]`
   - Validates transition without executing
   - Shows what would happen

---

## Command Specification

### Signature

```bash
sow advance [event] [flags]
```

### Positional Arguments

- `[event]` (optional): Event name to fire explicitly
  - Used for intent-based branching
  - Validates event is permitted before firing
  - Example: `sow advance finalize`

### Flags

- `--list`: List available transitions without executing
  - Shows: event name, target state, description, guard description
  - Filters by guard status (only shows permitted events)
  - Cannot be combined with `[event]` argument

- `--dry-run`: Validate transition without executing
  - Requires `[event]` argument
  - Checks if event can be fired (guard passes)
  - Shows target state and what would happen
  - Does not modify project state

### Mutual Exclusivity

- `--list` cannot be combined with `[event]`
- `--dry-run` requires `[event]`
- `[event]` and no flags = execute transition

---

## Behavior Modes

### Mode 1: Auto-Determination (Backward Compatible)

**Command**: `sow advance`

**For linear states**:
```
Current state: ImplementationPlanning
Auto-selected event: planning_complete
Advanced to: ImplementationExecuting
```

**For state-determined branching** (AddBranch):
```
Current state: ReviewActive
Auto-selected event: review_pass  (discriminator examined state)
Advanced to: FinalizeChecks
```

**For intent-based branching** (no DetermineEvent):
```
Current state: Researching

Cannot auto-determine transition: no event determiner configured for state Researching

Available transitions (choose one):
  sow advance finalize  # → Finalizing (Complete research and move to finalization phase)
  sow advance add_more_research  # → Planning (Return to planning to add more research topics)

Error: specify event explicitly
```

### Mode 2: Explicit Event Selection

**Command**: `sow advance [event]`

**Success case**:
```bash
$ sow advance finalize

Current state: Researching
Firing event: finalize
Advanced to: Finalizing
```

**Guard failure case**:
```bash
$ sow advance finalize

Current state: Researching
Firing event: finalize

Error: transition blocked: stateless: trigger 'finalize' is valid for
transition from state 'Researching' but guard condition is not met:
all tasks complete

Use --list to see available transitions
```

**Invalid event case**:
```bash
$ sow advance invalid_event

Current state: Researching
Error: event invalid_event not configured from state Researching
```

### Mode 3: Discovery (List Available Transitions)

**Command**: `sow advance --list`

**Output**:
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

**Terminal state**:
```bash
$ sow advance --list

Current state: Completed

No transitions available from current state.
This may be a terminal state.
```

**Guards blocking all transitions**:
```bash
$ sow advance --list

Current state: ImplementationExecuting

Available transitions:

(All configured transitions are currently blocked by guard conditions)

  sow advance all_tasks_complete  [BLOCKED]
    → ReviewActive
    Requires: all tasks completed or abandoned
```

### Mode 4: Dry-Run Validation

**Command**: `sow advance --dry-run [event]`

**Success**:
```bash
$ sow advance --dry-run finalize

Validating transition: finalize
Current state: Researching

✓ Transition is valid
  Event: finalize
  From: Researching
  To: Finalizing

Transition would succeed if executed.
```

**Failure**:
```bash
$ sow advance --dry-run finalize

Validating transition: finalize
Current state: Researching

✗ Transition blocked
  Reason: all tasks complete

Error: guard condition not met
```

**Missing event argument**:
```bash
$ sow advance --dry-run

Error: --dry-run requires an event argument

Usage: sow advance --dry-run [event]
```

---

## Implementation Details

### File Structure

**Location**: `cli/cmd/advance.go`

**Dependencies**:
- `cli/internal/sdks/project/state` - Project loading
- `cli/internal/sdks/state` - State machine
- `cli/cmdutil` - Context utilities

### Command Implementation

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
        RunE: runAdvance,
    }

    cmd.Flags().BoolVar(&listFlag, "list", false, "List available transitions without executing")
    cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Validate transition without executing")

    return cmd
}
```

### Main Run Function

```go
func runAdvance(cmd *cobra.Command, args []string) error {
    // Get flags
    listFlag, _ := cmd.Flags().GetBool("list")
    dryRunFlag, _ := cmd.Flags().GetBool("dry-run")

    // Load context and project
    ctx := cmdutil.GetContext(cmd.Context())
    project, err := state.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load project: %w", err)
    }

    machine := project.Machine()
    currentState := project.Statechart.Current_state

    // Handle list mode
    if listFlag {
        if len(args) > 0 {
            return fmt.Errorf("--list cannot be combined with event argument")
        }
        return listAvailableTransitions(machine, project, currentState)
    }

    // Get event from argument (if provided)
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

    // Execute transition (auto or explicit)
    if event != "" {
        return executeExplicitTransition(machine, project, event, currentState)
    } else {
        return executeAutoTransition(machine, project, currentState)
    }
}
```

### List Available Transitions

```go
func listAvailableTransitions(
    machine *sdkstate.Machine,
    project *state.Project,
    currentState string,
) error {
    fmt.Printf("Current state: %s\n\n", currentState)

    // Get permitted events (guards evaluated by stateless library)
    events, err := machine.PermittedTriggers()
    if err != nil {
        return fmt.Errorf("failed to get permitted triggers: %w", err)
    }

    if len(events) == 0 {
        // Check if any transitions configured (might be blocked by guards)
        allTransitions := project.Config().GetAvailableTransitions(
            sdkstate.State(currentState))

        if len(allTransitions) == 0 {
            fmt.Println("No transitions available from current state.")
            fmt.Println("This may be a terminal state.")
            return nil
        }

        // Transitions exist but all blocked by guards
        fmt.Println("Available transitions:")
        fmt.Println("\n(All configured transitions are currently blocked by guard conditions)\n")
        for _, t := range allTransitions {
            fmt.Printf("  sow advance %s  [BLOCKED]\n", t.Event)
            fmt.Printf("    → %s\n", t.To)
            if t.Description != "" {
                fmt.Printf("    %s\n", t.Description)
            }
            if t.GuardDesc != "" {
                fmt.Printf("    Requires: %s\n", t.GuardDesc)
            }
            fmt.Println()
        }
        return nil
    }

    fmt.Println("Available transitions:")
    for _, event := range events {
        // Get transition info from config
        targetState := project.Config().GetTargetState(
            sdkstate.State(currentState), event)
        description := project.Config().GetTransitionDescription(
            sdkstate.State(currentState), event)
        guardDesc := project.Config().GetGuardDescription(
            sdkstate.State(currentState), event)

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

### Validate Transition (Dry Run)

```go
func validateTransition(
    machine *sdkstate.Machine,
    project *state.Project,
    event sdkstate.Event,
    currentState string,
) error {
    fmt.Printf("Validating transition: %s\n", event)
    fmt.Printf("Current state: %s\n\n", currentState)

    // Check if transition is configured
    targetState := project.Config().GetTargetState(
        sdkstate.State(currentState), event)
    if targetState == "" {
        return fmt.Errorf("event %s not configured from state %s", event, currentState)
    }

    // Check if event is permitted (guard passes)
    canFire, err := machine.CanFire(event)
    if err != nil {
        return fmt.Errorf("failed to check event: %w", err)
    }

    if !canFire {
        // Get guard description for better error message
        guardDesc := project.Config().GetGuardDescription(
            sdkstate.State(currentState), event)
        fmt.Printf("✗ Transition blocked\n")
        if guardDesc != "" {
            fmt.Printf("  Reason: %s\n", guardDesc)
        }
        return fmt.Errorf("guard condition not met")
    }

    // Transition is valid
    description := project.Config().GetTransitionDescription(
        sdkstate.State(currentState), event)

    fmt.Printf("✓ Transition is valid\n")
    fmt.Printf("  Event: %s\n", event)
    fmt.Printf("  From: %s\n", currentState)
    fmt.Printf("  To: %s\n", targetState)
    if description != "" {
        fmt.Printf("  Description: %s\n", description)
    }
    fmt.Println("\nTransition would succeed if executed.")

    return nil
}
```

### Execute Explicit Transition

```go
func executeExplicitTransition(
    machine *sdkstate.Machine,
    project *state.Project,
    event sdkstate.Event,
    currentState string,
) error {
    fmt.Printf("Current state: %s\n", currentState)

    // Get target state for display
    targetState := project.Config().GetTargetState(
        sdkstate.State(currentState), event)
    if targetState == "" {
        // Check if event exists for any state
        return fmt.Errorf("event %s not configured from state %s", event, currentState)
    }

    fmt.Printf("Firing event: %s\n", event)

    // Fire the event with phase updates
    if err := project.Config().FireWithPhaseUpdates(machine, event, project); err != nil {
        // Enhanced error message
        if strings.Contains(err.Error(), "cannot fire event") ||
           strings.Contains(err.Error(), "guard condition") {
            guardDesc := project.Config().GetGuardDescription(
                sdkstate.State(currentState), event)
            fmt.Fprintf(os.Stderr, "\nTransition blocked")
            if guardDesc != "" {
                fmt.Fprintf(os.Stderr, ": %s\n", guardDesc)
            } else {
                fmt.Fprintln(os.Stderr)
            }
            fmt.Fprintln(os.Stderr, "\nUse --list to see available transitions")
            return err
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

### Execute Auto Transition

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
        // Auto-determination failed - expected for intent-based branching
        fmt.Fprintf(os.Stderr, "\nCannot auto-determine transition: %s\n\n", err)

        // Show available options
        events, _ := machine.PermittedTriggers()
        if len(events) > 0 {
            fmt.Fprintln(os.Stderr, "Available transitions (choose one):")
            for _, e := range events {
                targetState := project.Config().GetTargetState(
                    sdkstate.State(currentState), e)
                description := project.Config().GetTransitionDescription(
                    sdkstate.State(currentState), e)

                fmt.Fprintf(os.Stderr, "  sow advance %s", e)
                if targetState != "" {
                    fmt.Fprintf(os.Stderr, "  # → %s", targetState)
                }
                if description != "" {
                    fmt.Fprintf(os.Stderr, " (%s)", description)
                }
                fmt.Fprintln(os.Stderr)
            }
            return fmt.Errorf("specify event explicitly")
        }

        return fmt.Errorf("no valid transitions from state %s", currentState)
    }

    fmt.Printf("Auto-selected event: %s\n", event)

    // Fire the event with phase updates
    if err := project.Config().FireWithPhaseUpdates(machine, event, project); err != nil {
        if strings.Contains(err.Error(), "cannot fire event") ||
           strings.Contains(err.Error(), "guard condition") {
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

---

## Usage Examples

### Linear State Progression

```bash
$ sow advance
Current state: ImplementationPlanning
Auto-selected event: planning_complete
Advanced to: ImplementationExecuting
```

### State-Determined Branching

```bash
# Review assessment already set to "pass" in artifact metadata
$ sow advance
Current state: ReviewActive
Auto-selected event: review_pass
Advanced to: FinalizeChecks
```

### Intent-Based Branching

```bash
# Auto-advance fails (no determiner)
$ sow advance
Current state: Researching

Cannot auto-determine transition: no event determiner configured for state Researching

Available transitions (choose one):
  sow advance finalize  # → Finalizing (Complete research and move to finalization phase)
  sow advance add_more_research  # → Planning (Return to planning to add more research topics)

Error: specify event explicitly

# Orchestrator explicitly chooses
$ sow advance finalize
Current state: Researching
Firing event: finalize
Advanced to: Finalizing
```

### Discovery Workflow

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

### Validation Before Execution

```bash
# Check if finalize is valid
$ sow advance --dry-run finalize
Validating transition: finalize
Current state: Researching

✗ Transition blocked
  Reason: all tasks complete

Error: guard condition not met

# Complete remaining tasks...

# Check again
$ sow advance --dry-run finalize
Validating transition: finalize
Current state: Researching

✓ Transition is valid
  Event: finalize
  From: Researching
  To: Finalizing
  Description: Complete research and move to finalization phase

Transition would succeed if executed.

# Now execute
$ sow advance finalize
Current state: Researching
Firing event: finalize
Advanced to: Finalizing
```

---

## Orchestrator Integration

### Pattern 1: Simple Linear Progression

```python
# Orchestrator implementation pattern
run("sow advance")  # Auto-determines, no decision needed
```

### Pattern 2: State-Determined Branching

```python
# Set review assessment in artifact
create_review_artifact(assessment="pass")
run("sow output add --type review --path review.md")
run("sow output set --index 0 metadata.assessment pass")
run("sow output set --index 0 approved true")

# Auto-determines based on assessment
run("sow advance")  # Goes to FinalizeChecks
```

### Pattern 3: Intent-Based Branching with Discovery

```python
# Discover available options
result = run("sow advance --list")
options = parse_transitions(result)  # Extract events from output

# AI decides based on conversation context
if user_wants_more_research():
    run("sow advance add_more_research")
elif all_research_complete():
    run("sow advance finalize")
```

### Pattern 4: Validation Before Action

```python
# Check prerequisites before attempting transition
dry_run_result = run("sow advance --dry-run finalize")

if dry_run_result.success:
    # Prerequisites met, safe to proceed
    run("sow advance finalize")
else:
    # Handle missing prerequisites
    complete_remaining_tasks()
    run("sow advance finalize")
```

### Pattern 5: Error Recovery

```python
try:
    run("sow advance")
except CommandError as e:
    if "specify event explicitly" in str(e):
        # Intent-based branching - need to choose
        options = run("sow advance --list")
        event = ai_choose_event(options, context)
        run(f"sow advance {event}")
    elif "guard condition not met" in str(e):
        # Prerequisites not satisfied
        handle_prerequisites(e)
        retry_advance()
```

---

## Error Handling

### Invalid Event

**Input**: `sow advance invalid_event`

**Output**:
```
Current state: Researching
Error: event invalid_event not configured from state Researching
```

### Guard Failure

**Input**: `sow advance finalize` (when tasks not complete)

**Output**:
```
Current state: Researching
Firing event: finalize

Transition blocked: all tasks complete

Use --list to see available transitions

Error: transition blocked: stateless: trigger 'finalize' is valid for
transition from state 'Researching' but guard condition is not met:
all tasks complete
```

### Terminal State

**Input**: `sow advance`

**Output**:
```
Current state: Completed
Error: no valid transitions from state Completed

This may be a terminal state
```

### Conflicting Flags

**Input**: `sow advance finalize --list`

**Output**:
```
Error: --list cannot be combined with event argument

Usage: sow advance [event] [flags]
```

**Input**: `sow advance --dry-run`

**Output**:
```
Error: --dry-run requires an event argument

Usage: sow advance --dry-run [event]
```

---

## Backward Compatibility

### No Breaking Changes

All existing workflows continue working:

```bash
# Existing usage - unchanged
sow advance  # Auto-determines event, fires it, saves
```

### Behavior Preserved

- Linear states: Auto-determination works as before
- State-determined branching: Works via DetermineEvent (from AddBranch)
- Error messages enhanced but not breaking

### Migration

No migration needed. New flags and explicit events are opt-in features.

---

## Testing Strategy

### Unit Tests

```go
// cli/cmd/advance_test.go

func TestAdvanceAutoSelection(t *testing.T) {
    // Setup project in linear state
    // Execute: sow advance
    // Assert: auto-determines and transitions correctly
}

func TestAdvanceExplicitEvent(t *testing.T) {
    // Setup project in branching state
    // Execute: sow advance finalize
    // Assert: fires explicit event, transitions correctly
}

func TestAdvanceListMode(t *testing.T) {
    // Setup project in branching state
    // Execute: sow advance --list
    // Assert: shows all available transitions with descriptions
    // Assert: does not modify project state
}

func TestAdvanceDryRun(t *testing.T) {
    // Setup project
    // Execute: sow advance --dry-run finalize
    // Assert: validates without executing
    // Assert: does not modify project state
}

func TestAdvanceGuardFailure(t *testing.T) {
    // Setup project with unmet guard
    // Execute: sow advance finalize
    // Assert: returns error with guard description
    // Assert: does not modify project state
}

func TestAdvanceInvalidEvent(t *testing.T) {
    // Setup project
    // Execute: sow advance invalid_event
    // Assert: returns error about event not configured
}

func TestAdvanceTerminalState(t *testing.T) {
    // Setup project in terminal state
    // Execute: sow advance
    // Assert: returns error about no transitions
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

```go
func TestStandardProjectReview(t *testing.T) {
    // Standard project ReviewActive (state-determined)
    // Test auto-advance with pass/fail assessments
}

func TestExplorationProjectResearching(t *testing.T) {
    // Exploration project Researching (intent-based)
    // Test explicit event selection
}

func TestListWithGuardsBlocked(t *testing.T) {
    // Setup project where guards block all transitions
    // Verify --list shows transitions but marks as blocked
}
```

### CLI E2E Tests

```bash
# test_advance_cli.sh

# Test auto-advance
sow project new --branch test/advance --description "Test advance"
sow advance  # Should work

# Test explicit event
# ... setup project in branching state ...
sow advance finalize  # Should work

# Test --list
sow advance --list  # Should show options

# Test --dry-run
sow advance --dry-run finalize  # Should validate

# Test error cases
sow advance invalid_event  # Should error
sow advance finalize --list  # Should error
sow advance --dry-run  # Should error
```

---

## Conclusion

The enhanced `sow advance` command supports both branching patterns:

- **Auto-determination**: For linear states and state-determined branching (AddBranch)
- **Explicit events**: For intent-based branching (orchestrator choice)
- **Discovery**: `--list` shows all options
- **Validation**: `--dry-run` pre-flight checks

Benefits:
- Backward compatible (zero breaking changes)
- Discoverable (CLI introspection)
- Orchestrator-friendly (AI can discover and choose)
- Safe (validation before execution)
- Clear intent (explicit events in logs)

Next steps:
1. Review and approve this design
2. Implement command enhancements (2-3 days)
3. Add SDK introspection methods (required for --list)
4. Update tests
5. Update orchestrator prompts to use explicit events for intent-based branching
