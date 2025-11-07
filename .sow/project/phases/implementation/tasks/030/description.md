# Task 030: List Mode Implementation

## Context

This task implements the discovery mode for `sow advance --list`, which shows all available transitions from the current state without executing any transition. This is critical for orchestrators to understand their options, especially in intent-based branching scenarios.

**Why This Matters**: Orchestrators need to discover what transitions are possible before making decisions. In intent-based branching (e.g., exploration project's "finalize" vs "add_more_research"), the orchestrator cannot proceed without knowing the available options.

**List Mode Behavior**:
- Shows all configured transitions from current state
- Filters by guard status (highlights which are currently permitted)
- Displays descriptions, target states, and guard requirements
- Helps orchestrators make informed decisions

## Requirements

### List Mode Helper Function

Create a new helper function that displays all available transitions:

```go
func listAvailableTransitions(
    ctx *sow.Context,
    project *state.Project,
    machine *sdkstate.Machine,
    currentState sdkstate.State,
) error
```

This function should:
1. Get guard-filtered events using `machine.PermittedTriggers()`
2. Get all configured transitions using `config.GetAvailableTransitions(currentState)`
3. For each transition, get description and guard description
4. Display formatted output showing permitted vs blocked transitions

### Output Format

#### All Transitions Permitted

```
Current state: ReviewActive

Available transitions:

  sow advance review_pass
    → FinalizeChecks
    Review approved - proceed to finalization
    Requires: latest review approved

  sow advance review_fail
    → ImplementationPlanning
    Review failed - return to planning for rework
    Requires: latest review approved
```

#### Some Transitions Blocked

```
Current state: ImplementationExecuting

Available transitions:

  sow advance all_tasks_complete
    → ReviewActive
    All implementation tasks finished
    Requires: all tasks complete

(All configured transitions are currently blocked by guard conditions)

  sow advance skip_review  [BLOCKED]
    → FinalizeChecks
    Skip review for minor changes
    Requires: all tasks complete AND skip_review flag set
```

#### Terminal State

```
Current state: NoProject

No transitions available from current state.
This may be a terminal state.
```

### Integration with SDK Introspection

Use these SDK methods:
- `machine.PermittedTriggers()` - Guard-filtered events (what can fire now)
- `config.GetAvailableTransitions(state)` - All configured transitions
- `config.GetTransitionDescription(from, event)` - Human-readable description
- `config.GetTargetState(from, event)` - Where transition leads
- `config.GetGuardDescription(from, event)` - Guard requirement description

### Mode Switching

In main RunE, delegate to helper when `--list` flag is set:

```go
if listFlag {
    return listAvailableTransitions(ctx, project, machine, currentState)
}
```

## Acceptance Criteria

### Functional Tests (TDD)

Write tests BEFORE implementation in `cli/cmd/advance_test.go`:

1. **TestAdvanceListAvailable**:
   - Create project with multiple permitted transitions (all guards pass)
   - Run `sow advance --list`
   - Verify: Shows all transitions with descriptions
   - Verify: Shows target states
   - Verify: Shows guard descriptions
   - Verify: No `[BLOCKED]` markers

2. **TestAdvanceListBlocked**:
   - Create project where some guards fail
   - Run `sow advance --list`
   - Verify: Shows permitted transitions normally
   - Verify: Shows blocked transitions with `[BLOCKED]` marker
   - Verify: Note about blocked transitions appears

3. **TestAdvanceListAllBlocked**:
   - Create project where all guards fail (transitions configured but none permitted)
   - Run `sow advance --list`
   - Verify: Shows all transitions with `[BLOCKED]` markers
   - Verify: Message: "(All configured transitions are currently blocked by guard conditions)"

4. **TestAdvanceListTerminal**:
   - Create project in terminal state (no transitions configured)
   - Run `sow advance --list`
   - Verify: Message: "No transitions available from current state"
   - Verify: Message: "This may be a terminal state"

5. **TestAdvanceListWithDescriptions**:
   - Create project with transitions that have descriptions
   - Run `sow advance --list`
   - Verify: Descriptions are displayed
   - Verify: Guard descriptions are displayed under "Requires:"

6. **TestAdvanceListNoDescriptions**:
   - Create project with transitions lacking descriptions
   - Run `sow advance --list`
   - Verify: Transitions shown without description line
   - Verify: Still shows event name and target state

### Output Quality

- Consistent formatting and indentation
- Clear visual hierarchy (event → state → description → requirements)
- Human-readable (orchestrator can understand options)
- Machine-parseable if needed (consistent structure)

### Edge Cases Handled

- No transitions configured (terminal state)
- All transitions blocked (guards failing)
- Mix of permitted and blocked transitions
- Transitions with and without descriptions
- Transitions with and without guard descriptions

## Technical Details

### Helper Function Implementation

```go
func listAvailableTransitions(
    ctx *sow.Context,
    project *state.Project,
    machine *sdkstate.Machine,
    currentState sdkstate.State,
) error {
    fmt.Printf("Current state: %s\n\n", currentState)

    config := project.Config()

    // Get all configured transitions
    allTransitions := config.GetAvailableTransitions(currentState)
    if len(allTransitions) == 0 {
        fmt.Println("No transitions available from current state.")
        fmt.Println("This may be a terminal state.")
        return nil
    }

    // Get guard-filtered events (what can fire now)
    permittedEvents, err := machine.PermittedTriggers()
    if err != nil {
        return fmt.Errorf("failed to get permitted triggers: %w", err)
    }

    // Build set of permitted events for quick lookup
    permitted := make(map[sdkstate.Event]bool)
    for _, event := range permittedEvents {
        permitted[event] = true
    }

    // Display transitions
    fmt.Println("Available transitions:")

    if len(permittedEvents) == 0 {
        fmt.Println("\n(All configured transitions are currently blocked by guard conditions)\n")
    } else {
        fmt.Println()
    }

    for _, transition := range allTransitions {
        // Check if permitted
        blocked := !permitted[transition.Event]
        blockedMarker := ""
        if blocked {
            blockedMarker = "  [BLOCKED]"
        }

        // Display event and target
        fmt.Printf("  sow advance %s%s\n", transition.Event, blockedMarker)
        fmt.Printf("    → %s\n", transition.To)

        // Display description if present
        if transition.Description != "" {
            fmt.Printf("    %s\n", transition.Description)
        }

        // Display guard description if present
        if transition.GuardDesc != "" {
            fmt.Printf("    Requires: %s\n", transition.GuardDesc)
        }

        fmt.Println()
    }

    return nil
}
```

### TransitionInfo Structure

The SDK's `GetAvailableTransitions` returns:

```go
type TransitionInfo struct {
    Event       sdkstate.Event
    From        sdkstate.State
    To          sdkstate.State
    Description string
    GuardDesc   string
}
```

### Testing Utilities

Create helper to build test projects with various configurations:

```go
func createTestProjectWithTransitions(t *testing.T, state string, transitions []testTransition) *state.Project {
    // Build minimal project type config with specified transitions
    // Return project in specified state
}

type testTransition struct {
    event       string
    to          string
    description string
    guardDesc   string
    guard       func(*state.Project) bool
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/project/config.go` - Introspection methods (lines 301-506)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/internal/sdks/state/machine.go` - PermittedTriggers method (lines 50-63)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/project/context/issue-78.md` - Output format specifications (Section 8, Example 2)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/knowledge/designs/cli-enhanced-advance.md` - List mode detailed spec (Section 6.3)

## Examples

### Discovery for Intent-Based Branching

```bash
$ cd exploration-project
$ sow advance --list
Current state: Researching

Available transitions:

  sow advance finalize
    → Finalizing
    Research complete, ready to finalize findings

  sow advance add_more_research
    → Researching
    Need more investigation, add another research task
```

Orchestrator can now choose: `sow advance finalize` or `sow advance add_more_research`

### Discovery with Blocked Transitions

```bash
$ sow advance --list
Current state: ImplementationPlanning

Available transitions:

  sow advance planning_complete
    → ImplementationExecuting
    Task descriptions approved, begin execution
    Requires: task descriptions approved

(All configured transitions are currently blocked by guard conditions)

  sow advance planning_complete  [BLOCKED]
    → ImplementationExecuting
    Task descriptions approved, begin execution
    Requires: task descriptions approved
```

Orchestrator knows: Need to approve task descriptions first

## Dependencies

- **Task 010** complete (flag infrastructure)
- **Task 020** complete (auto mode implemented, can test together)
- SDK introspection methods available (GetAvailableTransitions, etc.)

## Constraints

### Performance

- Should complete in <100ms for typical projects
- Introspection methods are O(n) where n = transitions from state
- Typical projects: 5-10 transitions per state, negligible performance impact

### Output Format

- Must be human-readable (primary audience: orchestrators)
- Should be consistent (predictable structure)
- Can be machine-parseable (future enhancement)

### Error Handling

- Handle missing descriptions gracefully (don't show empty lines)
- Handle missing guard descriptions gracefully
- Clear error if PermittedTriggers fails

## Implementation Notes

### TDD Workflow

1. Write test for all-permitted case
2. Implement basic listing (no guard filtering)
3. Write test for blocked transitions case
4. Add guard filtering and `[BLOCKED]` markers
5. Write test for terminal state case
6. Add terminal state handling
7. Write test for descriptions
8. Ensure descriptions display properly

### Output Design Philosophy

**Prioritize clarity**:
- Most important info first (event name)
- Clear visual hierarchy (indentation)
- Highlight blocked transitions (but don't hide them)
- Help text when needed (terminal state, all blocked)

**Avoid clutter**:
- Don't show empty fields (missing descriptions)
- Don't repeat info (state is in header)
- Keep formatting consistent

### Testing Strategy

Use table-driven tests for various scenarios:

```go
func TestAdvanceListScenarios(t *testing.T) {
    tests := []struct {
        name           string
        state          string
        transitions    []testTransition
        expectedOutput string
    }{
        {
            name:  "all permitted",
            state: "Start",
            transitions: []testTransition{
                {event: "go", to: "End", guard: alwaysTrue},
            },
            expectedOutput: "sow advance go",
        },
        // ... more scenarios
    }
    // ... run tests
}
```

### Next Steps

After this task:
- Task 040 will implement dry-run mode
- Task 050 will implement explicit event mode
- Orchestrators can now discover options with `--list`
