# Branching State Transitions: Design Summary

**Status:** Proposal (Refined)
**Date:** 2025-11-06

## Overview

This design adds support for branching state transitions in the sow project SDK and CLI, enabling projects to transition to multiple possible states based on either project state or orchestrator intent.

## Key Decisions

### 1. Two Types of Branching

We identified two distinct branching patterns that require different solutions:

**State-Determined Branching:**
- Decision can be discovered by examining project state
- Example: Review assessment = "pass" or "fail" (already in artifact metadata)
- **Solution:** New `AddBranch()` SDK API with discriminator function
- **CLI:** `sow advance` (auto-determines)

**Intent-Based Branching:**
- Decision requires external orchestrator/user choice
- Example: "Add more research" vs "Finalize research"
- **Solution:** Multiple `AddTransition()` calls (existing API)
- **CLI:** `sow advance [event]` (explicit event selection)

### 2. Events vs States

**Decision:** Keep semantic events (not state-based transitions)
- Events preserve meaning (e.g., `review_pass` vs `review_fail`)
- Better for logging and understanding what happened
- Aligns with underlying stateless library design
- Both event and target state specified in API (self-documenting despite redundancy)

### 3. Transition Descriptions

**Decision:** Add `WithDescription()` option to transitions
- Provides human-readable explanation of what transition does
- Context-specific (same event from different states can have different meanings)
- Used by CLI `--list` to help orchestrators understand options
- Co-located with guards and actions

## High-Level Changes

### SDK (DESIGN_BRANCHING_SDK.md)

**New API:**
```go
// For state-determined branching
AddBranch(state,
    BranchOn(discriminatorFunc),
    When("value", event, targetState, options...),
)

// New transition option
WithDescription("Human-readable description")
```

**New Methods:**
```go
GetAvailableTransitions(from State) []TransitionInfo
GetTransitionDescription(from State, event Event) string
GetTargetState(from State, event Event) State
GetGuardDescription(from State, event Event) string
```

### CLI (DESIGN_ADVANCE_CLI.md)

**Enhanced Command:**
```bash
sow advance [event]           # Optional event for intent-based branching
sow advance --list            # Discover available transitions
sow advance --dry-run [event] # Validate without executing
```

**Behavior:**
- No argument: Auto-determines (existing behavior, backward compatible)
- With event: Explicitly fires that event (new, for intent-based branching)
- `--list`: Shows all permitted transitions with descriptions
- `--dry-run`: Validates transition without executing

## Examples

### State-Determined Branching (AddBranch)

```go
// SDK
AddBranch(ReviewActive,
    BranchOn(getReviewAssessment),  // Returns "pass" or "fail"
    When("pass", EventReviewPass, FinalizeChecks,
        WithDescription("Review approved - proceed to finalization")),
    When("fail", EventReviewFail, ImplementationPlanning,
        WithDescription("Review failed - return to planning for rework")),
)

// CLI
$ sow advance  # Auto-determines based on assessment
```

### Intent-Based Branching (Multiple AddTransition)

```go
// SDK
AddTransition(Researching, Finalizing, EventFinalize,
    WithDescription("Complete research and move to finalization phase"))
AddTransition(Researching, Planning, EventAddMoreResearch,
    WithDescription("Return to planning to add more research topics"))

// CLI
$ sow advance --list  # Shows options
$ sow advance finalize  # Orchestrator chooses explicitly
```

## Benefits

1. **Two-pattern approach** - Right tool for each type of branching
2. **Backward compatible** - Existing code unchanged, opt-in features
3. **Discoverable** - CLI can show available transitions with context
4. **Orchestrator-friendly** - AI can discover and choose options
5. **Self-documenting** - Descriptions explain what transitions do
6. **Natural backward transitions** - Just another branch destination
7. **Scales to N-way** - Not limited to binary branches

## Implementation Order

1. **SDK: WithDescription**
   - Add `description` field to `TransitionConfig`
   - Add `WithDescription()` option function
   - Update existing transitions to include descriptions

2. **SDK: Introspection Methods**
   - `GetTransitionDescription()`
   - `GetTargetState()`
   - `GetGuardDescription()`

3. **CLI: Enhanced Advance**
   - Add `[event]` positional argument
   - Add `--list` flag with formatted output
   - Add `--dry-run` flag
   - Update error messages

4. **SDK: AddBranch**
   - Implement `BranchConfig` and `BranchPath` structs
   - Add `AddBranch()` builder method
   - Implement `BranchOn()` and `When()` options
   - Auto-generate transitions and event determiner

5. **Migration**
   - Update standard project to use AddBranch
   - Update exploration project with descriptions
   - Test with orchestrator agents

## Related Documents

- **DESIGN_BRANCHING_SDK.md** - Full SDK design (AddBranch API)
- **DESIGN_ADVANCE_CLI.md** - Full CLI design (enhanced advance command)

## Next Steps

See individual design documents for detailed implementation plans and next steps.
