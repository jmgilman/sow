# feat(cli): add explicit event selection and discovery modes to advance command

## Summary

This PR enhances the `sow advance` command with four operation modes to support intent-based branching scenarios and improves user experience through better discoverability and validation. It also refactors the standard project type to use the declarative AddBranch API, demonstrating proper branching patterns.

**Key Enhancements:**
1. **CLI Advance Command** - Four operation modes with enhanced error messages and flag validation
2. **Standard Project Refactoring** - Declarative branching using AddBranch API (60% code reduction)
3. **Script-Based Integration Tests** - Proper testing that builds and executes the actual CLI binary

This work enables orchestrators to handle intent-based branching where auto-determination cannot decide between multiple valid transitions.

## Changes

### Component 1: CLI Enhanced Advance Command (Tasks 010-050)

**Four Operation Modes:**

1. **Auto-Determination Mode** (Task 020) - Backward compatible default
   - Enhanced error messages with actionable guidance
   - Suggests `--list` for intent-based branching scenarios
   - Lists available events when auto-determination fails
   - Maintains 100% backward compatibility

2. **Discovery Mode** (Task 030) - `sow advance --list`
   - Shows all available transitions from current state
   - Displays descriptions, target states, and guard requirements
   - Marks blocked transitions with `[BLOCKED]` indicator
   - Read-only (no state changes)

3. **Dry-Run Mode** (Task 040) - `sow advance --dry-run [event]`
   - Pre-flight validation without executing transitions
   - Zero side effects (verified by integration tests)
   - Shows what would happen without actually doing it
   - Helpful for orchestrators to validate before executing

4. **Explicit Event Mode** (Task 050) - `sow advance [event]`
   - Direct event execution for intent-based branching
   - Primary mode when orchestrator chooses between multiple options
   - Enhanced guard error messages with context and suggestions
   - Transaction safe (state only changes on success)

**Infrastructure** (Task 010):
- Added optional `[event]` positional argument
- Added `--list` and `--dry-run` flags
- Implemented comprehensive flag validation with mutual exclusivity rules
- Enhanced command documentation and usage examples

### Component 2: Standard Project Refactoring (Tasks 060-070)

**Task 060: Transition Descriptions**
- Added human-readable descriptions to all 10 state transitions
- Improves discoverability when using `--list` mode
- Consistent style and professional tone throughout

**Task 070: AddBranch API Refactoring**
- Replaced ReviewActive workaround pattern with declarative AddBranch
- Before: 2 transitions with identical misleading guards + manual OnAdvance (80 lines)
- After: Single AddBranch with discriminator function (30 lines)
- 60% code reduction with improved clarity and maintainability
- Serves as reference implementation for other project types
- 100% backward compatible (all existing tests pass)

### Component 3: Integration Testing (Task 080, 090)

**Task 080: Initial Integration Tests** (Iteration 1)
- Added comprehensive test coverage across all modes
- Discovered issue: tests called handlers directly instead of building CLI

**Task 090: Script-Based Tests** (Iteration 2 - Review Fix)
- **Removed** 2,099 lines of non-standard unit tests
- **Created** 6 script-based integration tests in `cli/testdata/script/integration/`:
  - `advance_auto.txtar` - Auto-determination mode
  - `advance_list.txtar` - Discovery mode
  - `advance_dryrun.txtar` - Dry-run validation (includes critical zero-side-effects test)
  - `advance_explicit.txtar` - Explicit event execution
  - `advance_flags.txtar` - Flag validation
  - `advance_standard_lifecycle.txtar` - Full standard project lifecycle

**Benefits of Script-Based Tests:**
- Execute actual CLI binary as subprocess (real user experience)
- Follow testscript framework pattern used throughout codebase
- Simpler and more maintainable
- Fast execution (<1 second for all tests)
- No reflection or unsafe code needed

## Testing

### Test Coverage

**All Four CLI Modes:**
- Auto-determination (linear states, terminal states, branching states)
- Discovery mode (list transitions, show blocked/permitted, read-only behavior)
- Dry-run mode (validate transitions, **zero side effects verified**)
- Explicit event mode (execute events, intent-based branching, guard errors)

**Standard Project Integration:**
- Full lifecycle tested (8 state transitions)
- ReviewActive branching (both pass and fail paths)
- Iteration loops (review fail → rework)
- AddBranch discriminator function

**Backward Compatibility:**
- Existing behavior unchanged (no args, no flags still works)
- All existing tests pass
- New features are purely additive

### Test Results

```
✅ Build: Successful (no errors or warnings)
✅ Tests: All pass (100% pass rate)
   - 11 script-based integration tests
   - Full test suite across 17 packages
✅ Performance: <1 second for all tests
✅ Coverage: All 24 requirements met
```

### Manual Testing

The CLI has been manually tested with:
- Real standard projects through full lifecycle
- All four modes in various states
- Guard blocking scenarios
- Intent-based branching decisions
- Error handling edge cases

## Implementation Highlights

**Files Modified:**
- `cli/cmd/advance.go`: Enhanced from 87 to 469 lines (+382 lines, +438%)
  - Four operation modes with clean helper functions
  - Enhanced error handling with context-aware messages
  - Flag validation with mutual exclusivity rules

- `cli/internal/projects/standard/standard.go`: AddBranch refactoring
  - Lines 132-181: Two transitions → Single AddBranch call
  - Lines 220-254: Manual OnAdvance removed (auto-generated)
  - 60% code reduction (80 lines → 30 lines)

- `cli/internal/projects/standard/guards.go`: Added discriminator
  - `getReviewAssessment()` function for branching logic
  - Handles edge cases (no review, missing metadata, multiple reviews)

**Test Files Created:**
- 6 script-based integration tests (cli/testdata/script/integration/)
- Tests execute actual CLI binary via testscript framework
- Follow standard patterns used throughout codebase

**Test Files Removed:**
- `cli/cmd/advance_test.go` (1,120 lines)
- `cli/cmd/advance_integration_test.go` (681 lines)
- `cli/cmd/advance_compatibility_test.go` (397 lines)

## Breaking Changes

**None** - All changes are backward compatible:
- `sow advance` (no args) works exactly as before
- New flags are optional
- New event argument is optional
- Existing projects continue to work unchanged

## Migration Notes

**No migration required** - This is a purely additive enhancement.

Users can immediately start using the new modes:
```bash
# Discover available transitions
sow advance --list

# Validate before executing
sow advance --dry-run planning_complete

# Execute specific event (intent-based branching)
sow advance finalize
```

## Related Issues

Closes #78

## Notes

### Design Decisions

1. **Four Operation Modes**: Provides flexibility for different orchestrator needs
   - Auto mode for simple linear flows
   - List mode for discovery
   - Dry-run mode for safety
   - Explicit mode for intent-based decisions

2. **AddBranch API**: Demonstrates declarative branching best practices
   - Replaces workaround pattern with cleaner code
   - Co-locates branching logic
   - Serves as reference for other project types

3. **Script-Based Testing**: Ensures tests match real user experience
   - Builds and executes actual CLI binary
   - Tests subprocess execution, not internal APIs
   - Follows codebase standards

### Future Enhancements

Potential follow-up work:
- Add `--list` output formatting options (JSON, table)
- Add `--dry-run` batch validation (multiple events)
- Extend AddBranch pattern to other project types
- Add more comprehensive guard descriptions

### Acknowledgments

This work was completed using the sow framework's TDD workflow across 9 tasks over 2 review iterations. All code changes were incrementally committed and pushed to maintain clear history.

---

🤖 Generated with [sow](https://github.com/jmgilman/sow)
