# Task 070: Verify Old Implementation Untouched

# Task 070: Verify Old Implementation Untouched

## Overview

Final verification task to ensure the migration didn't affect the existing `internal/project/standard/` implementation. Both implementations must coexist cleanly until Unit 5 migrates CLI commands.

## Context

**Why This Matters**: We created a parallel implementation in `internal/projects/standard/` while leaving `internal/project/standard/` untouched. This task verifies that promise was kept.

**Future Work**: Unit 5 will migrate CLI commands to use the new SDK implementation. Unit 6 will delete the old `internal/project/` package entirely.

## Requirements

### Verification Checks

1. **No File Modifications**
   - Old package files unchanged (git diff shows nothing)
   - Template files unchanged
   - Guard functions unchanged
   - Prompt generation unchanged

2. **Existing Tests Still Pass**
   - All old tests run successfully
   - No test failures or warnings
   - Performance not degraded

3. **No New Dependencies**
   - Old package imports unchanged
   - No circular dependencies introduced
   - Package still compiles independently

4. **No Naming Conflicts**
   - Both packages can coexist
   - No duplicate registrations
   - No symbol collisions

## Acceptance Criteria

### Git Status Check
- [ ] Run `git diff cli/internal/project/standard/` shows no changes
- [ ] Run `git status` shows no untracked files in old package
- [ ] Old package directory structure unchanged

### Test Execution
- [ ] Old tests pass: `go test ./cli/internal/project/standard/... -v`
- [ ] Old tests complete in reasonable time (< 10 seconds)
- [ ] No new test failures or warnings
- [ ] Test coverage unchanged or improved

### Compilation Check
- [ ] Old package compiles: `go build ./cli/internal/project/standard/...`
- [ ] No new compilation warnings
- [ ] No import cycle errors

### Coexistence Check
- [ ] Both packages compile together: `go build ./cli/internal/...`
- [ ] No naming conflicts between packages
- [ ] New package properly namespaced (`internal/projects/standard`)
- [ ] Old package path unchanged (`internal/project/standard`)

### Import Analysis
- [ ] Old package imports unchanged: `go list -f '{{.Imports}}' ./cli/internal/project/standard`
- [ ] Old package not importing new package
- [ ] No circular dependencies

## Validation Commands

```bash
# Git diff check
git diff cli/internal/project/standard/
# Should show: (no output)

# Git status check
git status cli/internal/project/standard/
# Should show: nothing to commit, working tree clean

# Old tests pass
go test ./cli/internal/project/standard/... -v
# Should show: PASS

# Old package compiles
go build ./cli/internal/project/standard/...
# Should show: (no output, success)

# Both packages compile together
go build ./cli/internal/...
# Should show: (no output, success)

# Check for naming conflicts
go list -m all | grep "internal/project"
# Should show both packages separately

# Verify imports unchanged
go list -f '{{.Imports}}' ./cli/internal/project/standard
# Compare to expected import list

# Check for circular dependencies
go list -f '{{.ImportPath}} {{.DepsErrors}}' ./cli/internal/...
# Should show: (no import cycle errors)
```

## Documentation

Create a brief summary document: `cli/internal/projects/standard/MIGRATION.md`

```markdown
# Standard Project SDK Migration

This package (`internal/projects/standard`) is the SDK-based implementation of the standard project type.

## Status

- **Current**: Fully implemented and tested
- **Usage**: Not yet used by CLI commands
- **Old Implementation**: `internal/project/standard` (still active)

## Migration Timeline

- **Unit 4** (this unit): SDK implementation created ✓
- **Unit 5** (next): CLI commands migrated to use SDK
- **Unit 6** (final): Old `internal/project` package deleted

## Key Differences

### Old Implementation
- Uses `*schemas.ProjectState` directly
- Uses `statechart.PromptComponents`
- Manual state machine wiring
- Separate guard files per phase

### New Implementation
- Uses `*state.Project` (SDK wrapper)
- Direct prompt functions (`func(*state.Project) string`)
- SDK builder configuration
- Single `guards.go` with helper functions

## Running Tests

```bash
# New implementation tests
go test ./cli/internal/projects/standard/... -v

# Old implementation tests (still working)
go test ./cli/internal/project/standard/... -v
```

## References

- Design: `.sow/knowledge/designs/project-sdk-implementation.md`
- Issue: #49
```

## Acceptance Criteria Summary

- [ ] Git diff shows no changes to old package
- [ ] All old tests pass
- [ ] Old package compiles independently
- [ ] Both packages coexist without conflicts
- [ ] No circular dependencies
- [ ] Migration summary document created
- [ ] All verification commands pass
- [ ] Report created with findings

## Report Template

Create a brief report in the task log:

```
VERIFICATION REPORT

1. Git Status
   - Changed files: [NONE]
   - Untracked files: [NONE]
   - Status: PASS

2. Old Tests
   - Tests run: [X tests]
   - Failures: [0]
   - Duration: [X seconds]
   - Status: PASS

3. Compilation
   - Old package: PASS
   - Both packages together: PASS
   - Warnings: [NONE]
   - Status: PASS

4. Coexistence
   - Naming conflicts: [NONE]
   - Circular dependencies: [NONE]
   - Import analysis: PASS
   - Status: PASS

OVERALL: PASS - Old implementation completely untouched
```

## Standards

- Run all verification commands
- Document any issues found (should be none)
- Report results clearly
- If any check fails, investigate and fix before completing task

## Dependencies

- Task 060 (SDK configuration complete)

## Notes

**This is a Safety Task**: If this task fails, something went wrong in earlier tasks. The old implementation should be completely untouched.

**Why Coexistence**: Unit 5 needs both implementations to work during CLI migration. Commands will be migrated one at a time, switching from old to new SDK types.

**Success Criteria**: All verification checks pass. If any fail, previous tasks need correction before proceeding.

**Migration Document**: The `MIGRATION.md` file helps future developers understand the two implementations and the migration timeline.
