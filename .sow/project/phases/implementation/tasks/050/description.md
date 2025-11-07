# Task 050: Introspection Methods for Branch Discovery

## Context

This task implements introspection methods that enable the CLI and other tooling to discover available transitions, branch paths, and state information from a ProjectTypeConfig. These methods are critical for the enhanced `sow advance` command (planned in a future work unit) that will list available transitions and provide context to users.

This is the fifth and final task in the TDD implementation of the AddBranch API. It builds on the complete, tested implementation from Tasks 010-040 to add discoverability.

### Project Goal

Enable programmatic discovery of state machine structure so that:
- The CLI can list available transitions with descriptions
- Orchestrators can understand what options are available in each state
- Developers can debug state machine configurations
- Future tooling can analyze and visualize project type workflows

### Why This Task Last

Introspection depends on:
- Transition descriptions (Task 010)
- Branch configurations (Task 020-030)
- Complete, validated AddBranch implementation (Task 040)

By implementing introspection last, we ensure we're querying a fully functional system.

## Requirements

### 1. Define TransitionInfo Struct

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go`

Add near the top of the file (after imports, before config structs):

```go
// TransitionInfo describes a single available transition from a state.
// Used by introspection methods to provide structured information about
// state machine configuration without exposing internal implementation details.
type TransitionInfo struct {
    Event       sdkstate.Event  // Event that triggers this transition
    From        sdkstate.State  // Source state
    To          sdkstate.State  // Target state
    Description string          // Human-readable description (empty if not provided)
    GuardDesc   string          // Guard description (empty if no guard)
}
```

This struct provides a public API for transition information without exposing internal TransitionConfig details.

### 2. Implement GetAvailableTransitions

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go`

Add after existing methods (around end of file):

```go
// GetAvailableTransitions returns all configured transitions from a state.
//
// This returns the transitions defined in the project type configuration,
// not filtered by guards. To check if a transition is currently allowed,
// use machine.CanFire(event) or machine.PermittedTriggers().
//
// Transitions are returned in a deterministic order:
//   1. First, transitions from branches (if state is a branching state)
//   2. Then, direct AddTransition calls
//   Both sorted by event name for consistency
//
// Returns empty slice if no transitions are defined from the state.
//
// Example:
//   transitions := config.GetAvailableTransitions(sdkstate.State(ReviewActive))
//   for _, t := range transitions {
//       fmt.Printf("%s -> %s: %s\n", t.Event, t.To, t.Description)
//   }
//
func (ptc *ProjectTypeConfig) GetAvailableTransitions(from sdkstate.State) []TransitionInfo {
    var result []TransitionInfo

    // Check if this is a branching state
    if branchConfig, exists := ptc.branches[from]; exists {
        // Add transitions from branch paths
        for _, path := range branchConfig.branches {
            result = append(result, TransitionInfo{
                Event:       path.event,
                From:        from,
                To:          path.to,
                Description: path.description,
                GuardDesc:   path.guardTemplate.Description,
            })
        }
    }

    // Add direct transitions (from AddTransition calls)
    for _, tc := range ptc.transitions {
        if tc.From == from {
            result = append(result, TransitionInfo{
                Event:       tc.Event,
                From:        tc.From,
                To:          tc.To,
                Description: tc.description,
                GuardDesc:   tc.guardTemplate.Description,
            })
        }
    }

    // Sort by event name for deterministic output
    sort.Slice(result, func(i, j int) bool {
        return result[i].Event < result[j].Event
    })

    return result
}
```

### 3. Implement GetTransitionDescription

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go`

```go
// GetTransitionDescription returns the human-readable description for a transition.
//
// Searches both branch paths and direct transitions for the specified from-state
// and event combination. Returns the first matching description found.
//
// Returns empty string if:
//   - No transition exists for the from/event combination
//   - The transition exists but has no description
//
// Note: This searches by from-state and event, not by to-state, since that's
// how transitions are triggered. The same event from different states can have
// different descriptions (context-specific).
//
// Example:
//   desc := config.GetTransitionDescription(
//       sdkstate.State(ReviewActive),
//       sdkstate.Event(EventReviewPass))
//   // Returns: "Review approved - proceed to finalization"
//
func (ptc *ProjectTypeConfig) GetTransitionDescription(from sdkstate.State, event sdkstate.Event) string {
    // Check branch paths first
    if branchConfig, exists := ptc.branches[from]; exists {
        for _, path := range branchConfig.branches {
            if path.event == event {
                return path.description
            }
        }
    }

    // Check direct transitions
    for _, tc := range ptc.transitions {
        if tc.From == from && tc.Event == event {
            return tc.description
        }
    }

    return ""
}
```

### 4. Implement GetTargetState

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go`

```go
// GetTargetState returns the target state for a transition.
//
// Searches both branch paths and direct transitions for the specified from-state
// and event combination. Returns the target state of the first matching transition.
//
// Returns empty State if no transition exists for the from/event combination.
//
// Example:
//   target := config.GetTargetState(
//       sdkstate.State(ReviewActive),
//       sdkstate.Event(EventReviewPass))
//   // Returns: sdkstate.State(FinalizeChecks)
//
func (ptc *ProjectTypeConfig) GetTargetState(from sdkstate.State, event sdkstate.Event) sdkstate.State {
    // Check branch paths first
    if branchConfig, exists := ptc.branches[from]; exists {
        for _, path := range branchConfig.branches {
            if path.event == event {
                return path.to
            }
        }
    }

    // Check direct transitions
    for _, tc := range ptc.transitions {
        if tc.From == from && tc.Event == event {
            return tc.To
        }
    }

    return ""
}
```

### 5. Implement GetGuardDescription

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go`

```go
// GetGuardDescription returns the guard description for a transition.
//
// Searches both branch paths and direct transitions for the specified from-state
// and event combination. Returns the guard description if a guard exists.
//
// Returns empty string if:
//   - No transition exists for the from/event combination
//   - The transition exists but has no guard
//   - The guard exists but has no description
//
// Example:
//   desc := config.GetGuardDescription(
//       sdkstate.State(ImplementationExecuting),
//       sdkstate.Event(EventAllTasksComplete))
//   // Returns: "all tasks complete"
//
func (ptc *ProjectTypeConfig) GetGuardDescription(from sdkstate.State, event sdkstate.Event) string {
    // Check branch paths first
    if branchConfig, exists := ptc.branches[from]; exists {
        for _, path := range branchConfig.branches {
            if path.event == event {
                return path.guardTemplate.Description
            }
        }
    }

    // Check direct transitions
    for _, tc := range ptc.transitions {
        if tc.From == from && tc.Event == event {
            return tc.guardTemplate.Description
        }
    }

    return ""
}
```

### 6. Implement IsBranchingState

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go`

```go
// IsBranchingState checks if a state has branches configured via AddBranch.
//
// Returns true if the state was configured with AddBranch (state-determined branching).
// Returns false if the state:
//   - Has no transitions
//   - Has only direct transitions (via AddTransition)
//   - Has multiple transitions but no AddBranch configuration
//
// This distinction is useful for UI/CLI to:
//   - Show different help text for branching vs non-branching states
//   - Indicate that transition choice is automatic vs manual
//   - Highlight states where discriminator logic determines the path
//
// Example:
//   if config.IsBranchingState(sdkstate.State(ReviewActive)) {
//       fmt.Println("This is a branching state - the system will automatically")
//       fmt.Println("determine which transition to take based on project state")
//   }
//
func (ptc *ProjectTypeConfig) IsBranchingState(state sdkstate.State) bool {
    _, exists := ptc.branches[state]
    return exists
}
```

### 7. Test Coverage

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config_test.go`

Add comprehensive unit tests:

```go
func TestGetAvailableTransitions(t *testing.T) {
    t.Run("returns transitions from branching state", func(t *testing.T) {
        // Create config with AddBranch
        config := createBranchingConfig(t)

        transitions := config.GetAvailableTransitions(sdkstate.State("BranchState"))

        require.Len(t, transitions, 2)
        // Verify both branches are included
        // Verify descriptions are present
        // Verify guard descriptions are present
    })

    t.Run("returns transitions from non-branching state", func(t *testing.T) {
        // Create config with AddTransition
        config := createLinearConfig(t)

        transitions := config.GetAvailableTransitions(sdkstate.State("State1"))

        require.Len(t, transitions, 1)
        // Verify transition info is correct
    })

    t.Run("returns empty slice for state with no transitions", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").Build()

        transitions := config.GetAvailableTransitions(sdkstate.State("NoTransitions"))

        assert.Empty(t, transitions)
    })

    t.Run("combines branch and direct transitions", func(t *testing.T) {
        // State with both AddBranch and additional AddTransition
        config := createMixedConfig(t)

        transitions := config.GetAvailableTransitions(sdkstate.State("MixedState"))

        // Should include both branching and direct transitions
        assert.Greater(t, len(transitions), 2)
    })

    t.Run("returns transitions in sorted order by event", func(t *testing.T) {
        // Create config with multiple transitions
        config := createMultiTransitionConfig(t)

        transitions := config.GetAvailableTransitions(sdkstate.State("MultiState"))

        // Verify sorted by event name
        for i := 1; i < len(transitions); i++ {
            assert.Less(t, transitions[i-1].Event, transitions[i].Event)
        }
    })

    t.Run("includes description when provided", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddBranch(
                sdkstate.State("State1"),
                BranchOn(func(_ *state.Project) string { return "val" }),
                When("val", sdkstate.Event("E1"), sdkstate.State("State2"),
                    WithDescription("Test description")),
            ).
            Build()

        transitions := config.GetAvailableTransitions(sdkstate.State("State1"))

        require.Len(t, transitions, 1)
        assert.Equal(t, "Test description", transitions[0].Description)
    })

    t.Run("includes guard description when provided", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddBranch(
                sdkstate.State("State1"),
                BranchOn(func(_ *state.Project) string { return "val" }),
                When("val", sdkstate.Event("E1"), sdkstate.State("State2"),
                    WithGuard("test guard", func(_ *state.Project) bool { return true })),
            ).
            Build()

        transitions := config.GetAvailableTransitions(sdkstate.State("State1"))

        require.Len(t, transitions, 1)
        assert.Equal(t, "test guard", transitions[0].GuardDesc)
    })
}

func TestGetTransitionDescription(t *testing.T) {
    t.Run("returns description for branch transition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddBranch(
                sdkstate.State("State1"),
                BranchOn(func(_ *state.Project) string { return "val" }),
                When("val", sdkstate.Event("E1"), sdkstate.State("State2"),
                    WithDescription("Branch description")),
            ).
            Build()

        desc := config.GetTransitionDescription(
            sdkstate.State("State1"),
            sdkstate.Event("E1"))

        assert.Equal(t, "Branch description", desc)
    })

    t.Run("returns description for direct transition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddTransition(
                sdkstate.State("State1"),
                sdkstate.State("State2"),
                sdkstate.Event("E1"),
                WithDescription("Direct description"),
            ).
            Build()

        desc := config.GetTransitionDescription(
            sdkstate.State("State1"),
            sdkstate.Event("E1"))

        assert.Equal(t, "Direct description", desc)
    })

    t.Run("returns empty string for non-existent transition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").Build()

        desc := config.GetTransitionDescription(
            sdkstate.State("NoState"),
            sdkstate.Event("NoEvent"))

        assert.Empty(t, desc)
    })

    t.Run("returns empty string when description not provided", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddTransition(
                sdkstate.State("State1"),
                sdkstate.State("State2"),
                sdkstate.Event("E1"),
                // No WithDescription
            ).
            Build()

        desc := config.GetTransitionDescription(
            sdkstate.State("State1"),
            sdkstate.Event("E1"))

        assert.Empty(t, desc)
    })
}

func TestGetTargetState(t *testing.T) {
    t.Run("returns target for branch transition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddBranch(
                sdkstate.State("State1"),
                BranchOn(func(_ *state.Project) string { return "val" }),
                When("val", sdkstate.Event("E1"), sdkstate.State("State2")),
            ).
            Build()

        target := config.GetTargetState(
            sdkstate.State("State1"),
            sdkstate.Event("E1"))

        assert.Equal(t, sdkstate.State("State2"), target)
    })

    t.Run("returns target for direct transition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddTransition(
                sdkstate.State("State1"),
                sdkstate.State("State2"),
                sdkstate.Event("E1"),
            ).
            Build()

        target := config.GetTargetState(
            sdkstate.State("State1"),
            sdkstate.Event("E1"))

        assert.Equal(t, sdkstate.State("State2"), target)
    })

    t.Run("returns empty state for non-existent transition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").Build()

        target := config.GetTargetState(
            sdkstate.State("NoState"),
            sdkstate.Event("NoEvent"))

        assert.Empty(t, target)
    })
}

func TestGetGuardDescription(t *testing.T) {
    t.Run("returns guard description for branch transition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddBranch(
                sdkstate.State("State1"),
                BranchOn(func(_ *state.Project) string { return "val" }),
                When("val", sdkstate.Event("E1"), sdkstate.State("State2"),
                    WithGuard("test guard", func(_ *state.Project) bool { return true })),
            ).
            Build()

        desc := config.GetGuardDescription(
            sdkstate.State("State1"),
            sdkstate.Event("E1"))

        assert.Equal(t, "test guard", desc)
    })

    t.Run("returns guard description for direct transition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddTransition(
                sdkstate.State("State1"),
                sdkstate.State("State2"),
                sdkstate.Event("E1"),
                WithGuard("test guard", func(_ *state.Project) bool { return true }),
            ).
            Build()

        desc := config.GetGuardDescription(
            sdkstate.State("State1"),
            sdkstate.Event("E1"))

        assert.Equal(t, "test guard", desc)
    })

    t.Run("returns empty string for transition without guard", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddTransition(
                sdkstate.State("State1"),
                sdkstate.State("State2"),
                sdkstate.Event("E1"),
            ).
            Build()

        desc := config.GetGuardDescription(
            sdkstate.State("State1"),
            sdkstate.Event("E1"))

        assert.Empty(t, desc)
    })

    t.Run("returns empty string for non-existent transition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").Build()

        desc := config.GetGuardDescription(
            sdkstate.State("NoState"),
            sdkstate.Event("NoEvent"))

        assert.Empty(t, desc)
    })
}

func TestIsBranchingState(t *testing.T) {
    t.Run("returns true for state with AddBranch", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddBranch(
                sdkstate.State("BranchState"),
                BranchOn(func(_ *state.Project) string { return "val" }),
                When("val", sdkstate.Event("E1"), sdkstate.State("State2")),
            ).
            Build()

        assert.True(t, config.IsBranchingState(sdkstate.State("BranchState")))
    })

    t.Run("returns false for state with only AddTransition", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddTransition(
                sdkstate.State("LinearState"),
                sdkstate.State("State2"),
                sdkstate.Event("E1"),
            ).
            Build()

        assert.False(t, config.IsBranchingState(sdkstate.State("LinearState")))
    })

    t.Run("returns false for state with no transitions", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").Build()

        assert.False(t, config.IsBranchingState(sdkstate.State("NoState")))
    })

    t.Run("returns false for state with multiple AddTransition but no AddBranch", func(t *testing.T) {
        config := NewProjectTypeConfigBuilder("test").
            AddTransition(sdkstate.State("MultiState"), sdkstate.State("S1"), sdkstate.Event("E1")).
            AddTransition(sdkstate.State("MultiState"), sdkstate.State("S2"), sdkstate.Event("E2")).
            Build()

        // Has multiple transitions but not configured with AddBranch
        assert.False(t, config.IsBranchingState(sdkstate.State("MultiState")))
    })
}
```

## Acceptance Criteria

- [ ] `TransitionInfo` struct defined with all required fields
- [ ] `GetAvailableTransitions()` implemented:
  - Returns transitions from both branches and direct transitions
  - Sorts results by event name for deterministic output
  - Returns empty slice for states with no transitions
  - Includes descriptions and guard descriptions when available
- [ ] `GetTransitionDescription()` implemented:
  - Searches both branches and direct transitions
  - Returns description or empty string
- [ ] `GetTargetState()` implemented:
  - Searches both branches and direct transitions
  - Returns target state or empty state
- [ ] `GetGuardDescription()` implemented:
  - Searches both branches and direct transitions
  - Returns guard description or empty string
- [ ] `IsBranchingState()` implemented:
  - Returns true only for states configured with AddBranch
  - Returns false for direct transitions and non-existent states
- [ ] All unit tests pass for introspection methods
- [ ] Methods handle edge cases (no transitions, no descriptions, etc.)
- [ ] Code is well-documented with examples
- [ ] No breaking changes to existing functionality

## Technical Details

### Search Strategy

All introspection methods search in this order:
1. Branch paths (if state is in branches map)
2. Direct transitions (from transitions slice)

This ensures branch configurations take precedence, which is correct since AddBranch auto-generates transitions.

### Performance Considerations

- `GetAvailableTransitions()`: O(n) where n = number of transitions from state
- `GetTransitionDescription()`: O(n) where n = number of transitions from state
- `GetTargetState()`: O(n) where n = number of transitions from state
- `GetGuardDescription()`: O(n) where n = number of transitions from state
- `IsBranchingState()`: O(1) map lookup

These are acceptable because:
- Called infrequently (CLI commands, not hot paths)
- n is typically small (2-5 transitions per state)
- Results could be cached if needed

### Sorting for Deterministic Output

`GetAvailableTransitions()` sorts by event name to ensure:
- Consistent output in tests
- Predictable CLI output
- Easier debugging

### Why TransitionInfo?

Instead of exposing internal `TransitionConfig`:
- Provides stable public API
- Hides implementation details
- Can evolve independently from internal structures
- Cleaner for external consumers

## Relevant Inputs

### Implementation File (Modify)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go` - Add TransitionInfo and all introspection methods

### Test File (Extend)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config_test.go` - Add tests for all introspection methods

### Existing Code (Reference)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go` - See existing query methods like GetPhaseForState for patterns
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder.go` - Understanding of branches map and transitions slice

### Documentation

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/.sow/project/context/issue-77.md` - Introspection requirements (section 4.5)
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/.sow/knowledge/designs/sdk-addbranch-api.md` - Introspection methods specification (section 5.4)

## Examples

### Usage Example

```go
// List all available transitions from current state
config := getProjectTypeConfig()
currentState := sdkstate.State(project.Statechart.Current_state)

transitions := config.GetAvailableTransitions(currentState)
fmt.Printf("Available transitions from %s:\n", currentState)

for _, t := range transitions {
    fmt.Printf("  %s -> %s", t.Event, t.To)
    if t.Description != "" {
        fmt.Printf(": %s", t.Description)
    }
    if t.GuardDesc != "" {
        fmt.Printf(" (requires: %s)", t.GuardDesc)
    }
    fmt.Println()
}

// Output:
// Available transitions from ReviewActive:
//   review_pass -> finalize_checks: Review approved - proceed to finalization (requires: latest review approved)
//   review_fail -> implementation_planning: Review failed - return to planning for rework (requires: latest review approved)
```

### CLI Integration Example

```go
// In CLI command
if config.IsBranchingState(currentState) {
    fmt.Println("This is a branching state. The system will automatically")
    fmt.Println("determine which transition to take based on project state.")
    fmt.Println()
    fmt.Println("Possible outcomes:")

    transitions := config.GetAvailableTransitions(currentState)
    for _, t := range transitions {
        fmt.Printf("  - %s\n", t.Description)
    }
} else {
    transitions := config.GetAvailableTransitions(currentState)
    if len(transitions) > 1 {
        fmt.Println("Multiple transitions available. Please specify event:")
        for _, t := range transitions {
            fmt.Printf("  sow advance %s  # %s\n", t.Event, t.Description)
        }
    }
}
```

## Dependencies

- Task 010 (Transition Descriptions) - MUST be complete
  - TransitionInfo includes description field
- Task 020-030 (Branch structures and AddBranch) - MUST be complete
  - Introspection queries branches map
- Task 040 (Error handling) - Should be complete
  - Ensures robust implementation to query

## Constraints

### Public API Stability

- TransitionInfo is a public struct (external consumers may depend on it)
- Method signatures should be stable (avoid breaking changes)
- Field names should be clear and self-documenting

### Backwards Compatibility

- Methods return empty values (not errors) for non-existent data
- Works with both AddBranch and AddTransition configurations
- Handles missing descriptions gracefully

### Memory Efficiency

- TransitionInfo is a small struct (5 fields, no pointers)
- GetAvailableTransitions creates new slice (safe to modify)
- No caching (simplicity over optimization for now)

## Success Criteria

This task is complete when:

1. All five introspection methods are implemented and tested
2. TransitionInfo struct is defined and documented
3. Methods work with both branching and non-branching states
4. Edge cases are handled (no transitions, no descriptions, etc.)
5. All unit tests pass
6. Code is well-documented with examples
7. No breaking changes to existing functionality

After this task, the AddBranch API is complete with full introspection support. The CLI can use these methods to provide rich information to users about available transitions and branching logic.
