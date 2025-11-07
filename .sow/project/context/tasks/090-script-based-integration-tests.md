# Task 090: Replace Non-Standard Unit Tests with Script-Based Integration Tests

## Context

The current implementation includes `cli/cmd/advance_test.go` which is non-standard - it contains unit tests that call command handlers directly rather than building and executing the actual CLI binary. All other commands in the codebase use script-based integration tests in `cli/testdata/script/` that test the CLI the way a real user would.

**Problem**: The "integration tests" in `advance_integration_test.go` and the unit tests in `advance_test.go` don't actually build the sow CLI binary and execute it as a subprocess. They call Go functions directly, which means they're not truly testing the full CLI execution path.

**Solution**: Replace `advance_test.go` with proper script-based integration tests in `cli/testdata/script/` that:
1. Build the actual sow CLI binary
2. Execute it as a subprocess
3. Test real user interactions with the CLI
4. Follow the existing pattern used by other commands

## Requirements

### Remove Non-Standard Test File

Delete `cli/cmd/advance_test.go` which contains:
- Unit tests calling command handlers directly (not true integration)
- Test helper functions using reflection/unsafe (code smell)
- Tests that don't build or execute the actual CLI binary

### Create Script-Based Integration Tests

Add new script files in `cli/testdata/script/` following the pattern of existing tests in subdirectories:
- `cli/testdata/script/integration/`
- `cli/testdata/script/phase/`
- `cli/testdata/script/project/`
- `cli/testdata/script/tasks/`

### Test Coverage Requirements

Create comprehensive script-based tests covering all four CLI modes:

1. **Auto-Determination Mode** (`advance_auto.txt`):
   - Test auto-advance through linear states
   - Test backward compatibility (no args, no flags)
   - Test enhanced error messages
   - Test guard enforcement

2. **Discovery Mode** (`advance_list.txt`):
   - Test `--list` showing all transitions
   - Test display of blocked vs permitted transitions
   - Test descriptions shown correctly
   - Test terminal state handling
   - Test read-only behavior (no state changes)

3. **Dry-Run Mode** (`advance_dryrun.txt`):
   - Test `--dry-run [event]` validation
   - Test valid transition success message
   - Test blocked transition error message
   - Test invalid event error message
   - **Critical**: Test zero side effects (no state changes)

4. **Explicit Event Mode** (`advance_explicit.txt`):
   - Test `sow advance [event]` execution
   - Test intent-based branching (multiple options)
   - Test enhanced guard error messages
   - Test invalid event handling

5. **Flag Validation** (`advance_flags.txt`):
   - Test mutual exclusivity rules
   - Test `--list` and `--dry-run` together (error)
   - Test `--list` with event argument (error)
   - Test `--dry-run` without event argument (error)

6. **Standard Project Integration** (`advance_standard_project.txt`):
   - Test full lifecycle with ReviewActive branching
   - Test both pass and fail paths
   - Test that AddBranch refactoring works end-to-end

### Script Test Pattern

Follow the existing testscript pattern used in the codebase:

```
# Test description
# Setup test environment
exec sow init

# Create test project
exec sow project new standard test-feature

# Test the command
exec sow advance --list
stdout 'Available transitions'
stdout 'ImplementationPlanning'

# Verify no side effects
! exists .sow/project/modified.txt

# Test execution
exec sow advance planning_complete
stdout 'Advanced to'

# Verify state changed correctly
exec sow project status
stdout 'ImplementationDraftPRCreation'
```

### Key Differences from Current Tests

**Current (Wrong)**:
- Call `NewAdvanceCmd()` and execute handlers directly
- Use reflection/unsafe to set private fields
- Don't build or execute actual CLI binary
- Don't test subprocess execution
- Don't test actual user experience

**Script-Based (Correct)**:
- Build actual sow CLI binary
- Execute as subprocess (real user experience)
- Test stdin/stdout/stderr
- Test exit codes
- Test file system changes
- No reflection or unsafe code needed

## Acceptance Criteria

### Tests Removed
- ✅ Delete `cli/cmd/advance_test.go`
- ✅ Delete `cli/cmd/advance_integration_test.go`
- ✅ Delete `cli/cmd/advance_compatibility_test.go`

### Script Tests Added
- ✅ At least 6 new `.txt` script files in `cli/testdata/script/`
- ✅ All four CLI modes covered
- ✅ Flag validation covered
- ✅ Standard project integration covered
- ✅ All tests build actual CLI binary and execute as subprocess

### Test Execution
- ✅ All new script tests pass: `go test -C cli ./testdata/script`
- ✅ Tests follow existing testscript patterns
- ✅ Tests are fast (<5s total)
- ✅ Tests are deterministic (no flaky tests)

### Coverage Maintained
- ✅ All scenarios from old tests covered by new script tests
- ✅ Critical no-side-effects test for dry-run mode
- ✅ Intent-based branching test for explicit mode
- ✅ Backward compatibility verified

## Technical Details

### Testscript Framework

The codebase uses https://github.com/rogpeppe/go-internal/tree/master/testscript for script-based testing.

Key commands available in scripts:
- `exec <command>` - Execute command, fail if non-zero exit
- `! exec <command>` - Execute command, expect non-zero exit
- `stdout <pattern>` - Assert stdout contains pattern
- `! stdout <pattern>` - Assert stdout doesn't contain pattern
- `stderr <pattern>` - Assert stderr contains pattern
- `exists <file>` - Assert file exists
- `! exists <file>` - Assert file doesn't exist
- `cmp <file1> <file2>` - Compare files

### Example Script Structure

```txt
# Test: sow advance --list shows all transitions
exec sow init
exec sow project new standard test-project

# Set up prerequisites
exec sow phase set metadata.planning_approved true --phase implementation

# Test list mode
exec sow advance --list
stdout 'Current state: ImplementationPlanning'
stdout 'Available transitions:'
stdout 'sow advance planning_complete'
stdout '→ ImplementationDraftPRCreation'
stdout 'Task descriptions approved'

# Verify no state changes (read-only)
exec sow project status
stdout 'ImplementationPlanning'  # Still in same state
```

### Test Organization

Organize tests by feature area (follow existing pattern):

```
cli/testdata/script/
├── advance/
│   ├── auto.txt
│   ├── list.txt
│   ├── dryrun.txt
│   ├── explicit.txt
│   ├── flags.txt
│   └── standard_project.txt
```

Or integrate into existing directories:
```
cli/testdata/script/
├── integration/
│   ├── advance_modes.txt  # All four modes
│   └── advance_flags.txt  # Flag validation
└── project/
    └── advance_standard.txt  # Full lifecycle
```

## Relevant Inputs

- `cli/testdata/script/integration/` - Existing integration test examples
- `cli/testdata/script/project/` - Existing project test examples
- `cli/cmd/advance.go` - The command being tested
- `cli/internal/projects/standard/standard.go` - Standard project type
- `.sow/project/phases/review/reports/001.md` - Review findings about integration testing

## Examples

### Existing Pattern (Other Commands)

Look at existing script tests for reference:
- How they build the CLI
- How they set up test environment
- How they assert on output and state
- How they clean up after tests

### Test All Four Modes

**Auto Mode**:
```txt
exec sow advance
stdout 'Current state:'
stdout 'Advanced to:'
```

**List Mode**:
```txt
exec sow advance --list
stdout 'Available transitions'
! stdout 'Advanced to'  # Read-only, no execution
```

**Dry-Run Mode**:
```txt
exec sow advance --dry-run planning_complete
stdout '✓ Transition is valid'
stdout 'To execute: sow advance planning_complete'
! stdout 'Advanced to'  # No execution
```

**Explicit Mode**:
```txt
exec sow advance planning_complete
stdout 'Current state:'
stdout 'Advanced to: ImplementationDraftPRCreation'
```

## Dependencies

- Understanding of testscript framework
- Existing script test patterns in `cli/testdata/script/`
- All CLI modes implemented (tasks 010-050)
- Standard project refactored (task 070)

## Constraints

### Must Follow Existing Patterns
- Use same testscript structure as other commands
- Follow existing organization in `cli/testdata/script/`
- Don't introduce new test patterns unique to advance command

### Must Actually Build CLI
- Tests must invoke the actual sow binary
- Must test subprocess execution, not Go function calls
- Must test real user experience

### Performance
- All script tests should complete in <5 seconds total
- Each individual test should be fast (<500ms)
- No slow integration setup (use minimal test projects)

## Implementation Notes

### TDD Approach

1. **Write scripts first** - Create `.txt` files with test scenarios
2. **Run tests** - They will fail (red phase)
3. **Verify** - CLI implementation already exists, tests should pass
4. **Refine** - Adjust scripts if needed for clarity

Since implementation is already complete, tests should mostly pass immediately.

### Migration Strategy

1. **Identify coverage** - List all scenarios in `advance_test.go`
2. **Map to scripts** - Create script test for each scenario
3. **Verify equivalence** - Ensure script tests cover same ground
4. **Delete old tests** - Remove `advance_test.go` files
5. **Run full suite** - Ensure nothing regressed

### Testing the Tests

Before deleting old tests:
1. Run old tests: `go test -C cli/cmd -v`
2. Run new script tests: `go test -C cli ./testdata/script`
3. Compare coverage - ensure all scenarios covered
4. Delete old tests
5. Run script tests again - all should pass

## Next Steps

After this task:
- All advance command testing uses proper script-based approach
- Consistent with rest of codebase
- Truly tests CLI as users experience it
- Review can pass with confidence
