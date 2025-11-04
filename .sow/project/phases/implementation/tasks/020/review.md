# Task 020 Review: Define States and Events

## Requirements Summary

- Create `states.go` with 8 state constants using SDK `state.State` type
- Create `events.go` with 9 event constants using SDK `state.Event` type
- Include comprehensive documentation comments
- Use `internal/sdks/state` NOT `internal/project/statechart`
- Verify compilation

## Changes Made

**Files Created:**
1. `cli/internal/projects/standard/states.go` - 8 state constants
2. `cli/internal/projects/standard/events.go` - 9 event constants

**States Defined:**
- NoProject, PlanningActive, ImplementationPlanning, ImplementationExecuting
- ReviewActive, FinalizeDocumentation, FinalizeChecks, FinalizeDelete

**Events Defined:**
- EventProjectInit, EventCompletePlanning, EventTasksApproved, EventAllTasksComplete
- EventReviewPass, EventReviewFail, EventDocumentationDone, EventChecksDone, EventProjectDelete

## Verification

✅ **Correct Types**: Both files use `internal/sdks/state` types (not old statechart)
✅ **Documentation**: All constants have clear documentation comments
✅ **Event Details**: Events include transition documentation (from/to states, guards)
✅ **Compilation**: Package compiles successfully (verified with `go build`)
✅ **Old Package**: No changes to `cli/internal/project/standard/`
✅ **State Machine Alignment**: States and events match design doc diagram

## Code Quality

- Clean imports (only SDK types)
- Consistent naming conventions
- Comprehensive documentation
- Follows Go best practices

## Assessment

**APPROVED**

Task completed successfully. All acceptance criteria met:
- Correct SDK types used throughout
- All 8 states and 9 events defined
- Excellent documentation with transition details
- Compiles without errors
- Old implementation untouched

Ready for use in subsequent tasks (guards, configuration).
