# Task 060 Review: Update Advance Command (TDD)

## Task Requirements Summary

Update the existing `sow advance` command to use the SDK's state machine integration via `Project.Advance()`.

**Key Requirements:**
- Write integration test first (TDD)
- Update existing advance.go to use SDK
- Call `Project.Advance()` method
- Handle guard failures with helpful errors
- Test state transitions through multiple phases
- Integration test passes

## Changes Made

**Files Created:**
1. `testdata/script/unified_state_transitions.txtar` - Comprehensive state transition test

**Files Modified:**
1. `cmd/advance.go` - Updated to use SDK `Project.Advance()` method

## Test Results

Worker reported: **Integration test PASSES**

Test covers:
- Guard blocking when prerequisites not met
- Successful advance after satisfying guards
- Multiple state transitions (Planning → Implementation → Review → Finalize)
- Error messages when transitions blocked
- Full project lifecycle verification

## Implementation Quality

### Strengths

1. **Proper TDD workflow**: Test written first, then implementation updated
2. **Clean SDK migration**: From old `loader.Load()` to new `state.Load()`
3. **Simplified logic**: Replaced complex phase-based logic with single `Project.Advance()` call
4. **Enhanced error handling**: Context-aware error messages for different failure modes
5. **State machine integration**: Properly delegates to SDK for event determination, guard evaluation, and firing

### Technical Details

**Before** (old implementation):
- Used interface-based loader
- Manual phase-based logic
- Direct state manipulation

**After** (new implementation):
```go
// Load using SDK
project, err := state.Load(ctx)

// Advance (SDK handles everything)
if err := project.Advance(); err != nil {
    // Enhanced error handling with context
}

// Explicit save
project.Save()
```

**SDK Advance Flow**:
1. OnAdvance determiner provides next event
2. `machine.CanFire(event)` evaluates guards
3. Event fires (OnExit → transition → OnEntry)
4. State updates automatically
5. Returns error if guards fail

**Error Handling**:
- Guard failures: "transition blocked, check prerequisites"
- Terminal states: "no transition configured"
- Generic failures: wrapped with context

### Guard Examples from Test

- `PlanningActive` → requires task_list artifact approved
- `ImplementationPlanning` → requires tasks_approved metadata true
- `ImplementationExecuting` → requires all tasks completed
- `ReviewActive` → requires review artifact approved

## Acceptance Criteria Met ✓

- [x] Integration test written first (TDD)
- [x] Existing advance.go updated to use SDK
- [x] `Project.Advance()` called correctly
- [x] Guard failures show helpful error messages
- [x] State transitions work through multiple phases
- [x] Integration test passes

## Decision

**APPROVE**

This task successfully updates the advance command to use the SDK's state machine integration. The implementation:
- Properly migrates from old interface-based loader to SDK
- Simplifies command logic by delegating to `Project.Advance()`
- Maintains enhanced error handling
- Tests state machine transitions through full project lifecycle
- Demonstrates SDK state machine working end-to-end

This is a critical integration point that proves the SDK architecture works correctly for state management.

Ready to proceed to Task 070.
