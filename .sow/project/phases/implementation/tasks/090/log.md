# Task Log

## Analysis Phase

### Examined Existing Tests (advance_*.go)
- **advance_test.go**: Contains unit tests calling command handlers directly
  - Tests command signature (arguments, flags)
  - Tests flag validation (mutual exclusivity)
  - Tests all four CLI modes (auto, list, dry-run, explicit)
  - Tests various scenarios (linear states, branching, terminal states, blocked transitions)
  - Uses reflection/unsafe to set private fields (code smell)
  - Does NOT build or execute actual CLI binary

- **advance_integration_test.go**: Integration tests against standard project
  - Tests full lifecycle with all CLI modes
  - Tests ReviewActive branching (pass/fail paths)
  - Tests dry-run has no side effects
  - Uses real project on disk but calls Go functions directly

- **advance_compatibility_test.go**: Backward compatibility tests
  - Tests auto-advance without flags (original behavior)
  - Tests error messages remain helpful
  - Tests new flags don't break existing workflows
  - Tests guard conditions still enforced

### Test Coverage to Migrate

All scenarios from old tests need script-based equivalents:

1. **Auto Mode** (backward compatible, no args/flags)
   - Linear state advancement
   - Guard enforcement
   - Enhanced error messages
   - Terminal state handling

2. **List Mode** (`--list`)
   - Show all transitions
   - Display permitted vs blocked transitions
   - Show descriptions
   - Terminal state handling
   - Read-only behavior (no state changes)

3. **Dry-Run Mode** (`--dry-run [event]`)
   - Validate valid transitions
   - Show blocked transition errors
   - Show invalid event errors
   - CRITICAL: Zero side effects

4. **Explicit Event Mode** (`sow advance [event]`)
   - Execute specific event
   - Intent-based branching (multiple options)
   - Enhanced guard error messages
   - Invalid event handling

5. **Flag Validation**
   - Mutual exclusivity rules
   - `--list` and `--dry-run` together (error)
   - `--list` with event argument (error)
   - `--dry-run` without event (error)

6. **Standard Project Integration**
   - Full lifecycle with ReviewActive branching
   - Both pass and fail paths
   - AddBranch refactoring works end-to-end

### Reviewed Existing Script Test Patterns

Examined `cli/testdata/script/integration/full_lifecycle.txtar`:
- Uses txtar format for embedded test data
- Sets up git repo with init/config/commit
- Creates .sow/project structure manually
- Uses YAML files for state
- Tests commands by executing `sow` binary via testscript
- Verifies output with `stdout` assertions
- Verifies file changes with `exists` and `cat` commands
- Cleans up automatically (temp dir per test)

Test organization:
- `cli/testdata/script/integration/` - end-to-end tests
- `cli/testdata/script/project/` - project-specific tests
- `cli/testdata/script/phase/` - phase-specific tests
- `cli/cli_test.go` - test runner with testscript.Run()

## Implementation Phase

Following TDD refactor workflow:
1. Create script tests for each CLI mode (tests first)
2. Run tests to verify they pass (CLI already implemented)
3. Delete old unit tests once script tests pass
4. Verify full test suite passes

### Action 1: Created advance_auto.txtar
Created script-based test for auto-determination mode covering:
- Linear state advancement with guard passing
- Guard blocking transition (guard fails)
- State advancement after guard satisfied
- Backward compatibility (no flags, no args)

File: `cli/testdata/script/integration/advance_auto.txtar`

### Action 2: Created advance_list.txtar
Created script-based test for discovery mode covering:
- List shows permitted transitions (guard passes)
- List shows blocked transitions with `[BLOCKED]` marker
- Read-only behavior (no side effects, state unchanged)
- Terminal state handling ("No transitions available")
- Description display

File: `cli/testdata/script/integration/advance_list.txtar`

### Action 3: Created advance_dryrun.txtar
Created script-based test for dry-run mode covering:
- Validate valid transition (guard passes)
- Show validation success message with execution hint
- Validate blocked transition (guard fails)
- Show guard error message
- Validate invalid event (event not configured)
- **CRITICAL**: Zero side effects test - verifies OnEntry actions NOT executed
- State always unchanged after dry-run

File: `cli/testdata/script/integration/advance_dryrun.txtar`

### Action 4: Created advance_explicit.txtar
Created script-based test for explicit event mode covering:
- Execute specific event in linear state (success)
- Guard blocking execution (error with helpful message)
- Invalid event error (event not configured)
- Intent-based branching: ReviewActive with pass path
- Intent-based branching: ReviewActive with fail path
- Iteration increment on review fail

File: `cli/testdata/script/integration/advance_explicit.txtar`

### Action 5: Created advance_flags.txtar
Created script-based test for flag validation covering:
- Valid modes (baseline: auto, list, dry-run, explicit)
- Invalid flag combinations:
  - `--list` and `--dry-run` together (error)
  - `--list` with event argument (error)
  - `--dry-run` without event argument (error)
- State unchanged during validation errors

File: `cli/testdata/script/integration/advance_flags.txtar`

### Action 6: Created advance_standard_lifecycle.txtar
Created script-based test for full standard project lifecycle covering:
- Complete success path: Planning → Implementation → Review (pass) → Finalize → NoProject
- Review pass path: advances to FinalizeChecks
- Review fail path: returns to ImplementationPlanning
- Iteration increment on review fail
- Phase status transitions (completed, in_progress)
- AddBranch discriminator working correctly

File: `cli/testdata/script/integration/advance_standard_lifecycle.txtar`

### Action 7: Verified New Tests Pass
Ran `go test -run TestScripts_Integration` in cli/ directory:
- All 6 new advance script tests: PASS
- All existing script tests: PASS
- Total test time: 0.757s (fast, meets <5s requirement)
- Tests are deterministic (no flaky behavior observed)

### Action 8: Deleted Old Unit Test Files
Removed non-standard test files:
- `cli/cmd/advance_test.go` (1120 lines)
- `cli/cmd/advance_integration_test.go` (638 lines)
- `cli/cmd/advance_compatibility_test.go` (341 lines)

Total removed: 2099 lines of reflection/unsafe-based unit tests

### Action 9: Verified Full Test Suite
Ran `go test ./...` in cli/ directory:
- All cmd package tests: PASS
- All schema tests: PASS
- All integration tests: PASS
- No regressions detected

## Coverage Verification

### All Original Test Scenarios Covered

**From advance_test.go**:
- Command signature (arguments, flags) - ✓ advance_flags.txtar
- Flag validation (mutual exclusivity) - ✓ advance_flags.txtar
- Auto mode linear states - ✓ advance_auto.txtar
- Auto mode terminal state - ✓ advance_list.txtar
- List mode (available, blocked, terminal) - ✓ advance_list.txtar
- List mode descriptions - ✓ advance_list.txtar
- Dry-run valid transition - ✓ advance_dryrun.txtar
- Dry-run blocked transition - ✓ advance_dryrun.txtar
- Dry-run invalid event - ✓ advance_dryrun.txtar
- **Dry-run no side effects** - ✓ advance_dryrun.txtar (CRITICAL test)
- Explicit event success - ✓ advance_explicit.txtar
- Explicit event guard failure - ✓ advance_explicit.txtar
- Explicit event invalid event - ✓ advance_explicit.txtar
- Intent-based branching - ✓ advance_explicit.txtar

**From advance_integration_test.go**:
- Full lifecycle auto-advance - ✓ advance_standard_lifecycle.txtar
- List mode with standard project - ✓ advance_list.txtar
- Dry-run with standard project - ✓ advance_dryrun.txtar
- ReviewActive branching (pass path) - ✓ advance_standard_lifecycle.txtar
- ReviewActive branching (fail path) - ✓ advance_standard_lifecycle.txtar

**From advance_compatibility_test.go**:
- Auto-advance without flags (backward compat) - ✓ advance_auto.txtar
- Error messages remain helpful - ✓ advance_explicit.txtar, advance_dryrun.txtar
- New flags optional - ✓ advance_flags.txtar
- Guard conditions enforced - ✓ advance_auto.txtar, advance_explicit.txtar

### New Tests Actually Build and Execute CLI Binary

Unlike old tests which called Go functions directly, new tests:
- Build actual `sow` CLI binary via testscript
- Execute as subprocess (real user experience)
- Test stdin/stdout/stderr
- Test exit codes
- Test file system changes
- No reflection or unsafe code needed

## Summary

Successfully replaced 2099 lines of non-standard unit tests with 6 comprehensive script-based integration tests totaling approximately 600 lines. The new tests:

1. **Build and execute actual CLI binary** - Test real user experience
2. **Cover all original scenarios** - No coverage loss
3. **Follow existing patterns** - Consistent with codebase conventions
4. **Run fast** - Complete in <1 second
5. **Are deterministic** - No flaky tests
6. **Test critical behaviors** - Especially dry-run zero side effects

The script-based tests are superior because they:
- Test the CLI as users experience it (subprocess execution)
- Don't use reflection/unsafe hacks
- Are simpler and more maintainable
- Catch integration issues that unit tests miss
- Follow the testscript pattern used throughout the codebase

All acceptance criteria met:
- ✓ Deleted 3 non-standard test files
- ✓ Created 6 script-based test files
- ✓ All four CLI modes covered
- ✓ Flag validation covered
- ✓ Standard project integration covered
- ✓ All tests build actual CLI binary and execute as subprocess
- ✓ All tests pass
- ✓ Tests follow existing patterns
- ✓ Tests are fast (<1s)
- ✓ Tests are deterministic
- ✓ Coverage maintained (all original scenarios covered)
- ✓ Critical no-side-effects test for dry-run mode included
