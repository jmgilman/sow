# Task 050: Command Integration and Cleanup

## Context

This is the final task in the wizard foundation work unit. At this point, we have:
- ✅ Huh library installed (Task 010)
- ✅ Shared utilities extracted (Task 020)
- ✅ Wizard foundation and state machine built (Task 030)
- ✅ Helper functions implemented (Task 040)

Now we need to:
1. **Integrate the wizard** as the primary `sow project` command
2. **Remove old commands** (`new` and `continue` subcommands)
3. **Delete obsolete files** (`new.go` and `continue.go`)
4. **Verify nothing is broken**

**Critical Safety Check**: Before deleting `new.go` and `continue.go`, verify that all their functionality is either:
- Moved to `shared.go` (shared utilities)
- Replaced by wizard screens (even if stubs)
- No longer needed

This task completes the foundational work, making the wizard the official way to create and continue projects.

## Requirements

### Step 1: Update project.go to Use Wizard

Modify `cli/cmd/project/project.go`:

**Before**:
```go
func NewProjectCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "project",
        Short: "Manage projects",
        // ...
    }

    cmd.AddCommand(newNewCmd())
    cmd.AddCommand(newContinueCmd())
    cmd.AddCommand(newSetCmd())
    cmd.AddCommand(newDeleteCmd())

    return cmd
}
```

**After**:
```go
func NewProjectCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "project",
        Short: "Create or continue a project (interactive)",
        Long: `Interactive wizard for creating or continuing projects.

The wizard guides you through:
  - Creating new projects from GitHub issues or branch names
  - Continuing existing projects
  - Selecting project types and providing descriptions

Examples:
  sow project                    # Launch interactive wizard
  sow project -- --model opus    # Launch wizard, pass flags to Claude`,
        RunE: runWizard,  // Use wizard as the main command
        Args: cobra.NoArgs,
    }

    // Keep set and delete subcommands
    cmd.AddCommand(newSetCmd())
    cmd.AddCommand(newDeleteCmd())

    return cmd
}
```

**Key Changes**:
- Remove `newNewCmd()` and `newContinueCmd()` from subcommands
- Add `RunE: runWizard` to main command
- Update Short and Long descriptions
- Keep `set` and `delete` subcommands (they're still useful)

### Step 2: Update shared.go to Use Shared Functions

**IMPORTANT**: Before deleting `new.go` and `continue.go`, we need to verify they're using the shared functions we extracted.

In Task 020, we extracted functions to `shared.go` but left the originals unchanged. Now we update the originals to use the shared functions, test that everything works, and only THEN delete them.

**Modify new.go** (temporarily, will be deleted after verification):

Find the embedded logic for:
- Project initialization (lines 148-196)
- Prompt generation (lines 359-395)
- Claude launch (lines 397-418)

Replace with calls to shared functions:
```go
// Instead of inline project creation:
proj, err := initializeProject(worktreeCtx, selectedBranch, description, issue)

// Instead of inline prompt generation:
prompt, err := generateNewProjectPrompt(proj, description)

// launchClaudeCode is already using the shared function
```

**Modify continue.go** (temporarily, will be deleted after verification):

Find the embedded logic for:
- Prompt generation (lines 167-196)

Replace with call to shared function:
```go
// Instead of inline prompt generation:
prompt, err := generateContinuePrompt(proj)

// launchClaudeCode is already using the shared function
```

**Test that old commands still work**:
```bash
cd cli/
go build -o /tmp/sow ./cmd/sow
/tmp/sow project new --branch test/migration "Test migration"
/tmp/sow project continue --branch test/migration
```

If these work, the shared functions are correct and we can safely delete the files.

### Step 3: Delete Obsolete Files

After verifying the old commands work with shared functions:

**Delete**:
- `cli/cmd/project/new.go` (entire file)
- `cli/cmd/project/continue.go` (entire file)

**Verify no references remain**:
```bash
cd cli/
grep -r "newNewCmd\|newContinueCmd" .
# Should only find references in git history, not in code
```

### Step 4: Update Imports

After deletion, update any files that imported from `new.go` or `continue.go`:

**Check for imports**:
```bash
cd cli/
grep -r "project/new\|project/continue" .
```

If any imports exist, remove them or update to use `shared.go` functions.

### Step 5: Final Verification

**Build the CLI**:
```bash
cd cli/
go build -o /tmp/sow ./cmd/sow
```

**Test the wizard command**:
```bash
/tmp/sow project
# Should launch the wizard entry screen
```

**Test that old subcommands are gone**:
```bash
/tmp/sow project new
# Should show error: unknown command "new" for "sow project"

/tmp/sow project continue
# Should show error: unknown command "continue" for "sow project"
```

**Test that other subcommands still work**:
```bash
/tmp/sow project set --help
/tmp/sow project delete --help
```

**Run tests**:
```bash
cd cli/
go test ./cmd/project/...
# All tests should pass
```

## Acceptance Criteria

### Command Integration
- [ ] `sow project` launches the wizard (entry screen)
- [ ] `sow project --help` shows wizard documentation
- [ ] `sow project` accepts Claude Code flags after `--`
- [ ] `sow project new` returns error (subcommand removed)
- [ ] `sow project continue` returns error (subcommand removed)
- [ ] `sow project set` still works (subcommand retained)
- [ ] `sow project delete` still works (subcommand retained)

### File Cleanup
- [ ] `cli/cmd/project/new.go` deleted
- [ ] `cli/cmd/project/continue.go` deleted
- [ ] No references to `newNewCmd()` in codebase
- [ ] No references to `newContinueCmd()` in codebase
- [ ] No dead code or unused imports
- [ ] No compilation errors

### Functionality Preservation
- [ ] All functionality from `new.go` is preserved (in shared.go or wizard)
- [ ] All functionality from `continue.go` is preserved (in shared.go or wizard)
- [ ] No regression in existing features
- [ ] Test suite passes completely

### Documentation
- [ ] Command help text updated
- [ ] No references to old flags in documentation
- [ ] Examples show wizard usage

## Relevant Inputs

- `cli/cmd/project/project.go` - Command structure to modify
- `cli/cmd/project/new.go` - File to delete (after verification)
- `cli/cmd/project/continue.go` - File to delete (after verification)
- `cli/cmd/project/shared.go` - Shared functions that replace old logic
- `cli/cmd/project/wizard.go` - Wizard command to integrate
- `.sow/project/context/issue-68.md` - Migration strategy (section 7)

## Examples

### Example 1: Updated project.go

```go
package project

import (
    "github.com/spf13/cobra"
)

// NewProjectCmd creates the project command with wizard as primary interface.
func NewProjectCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "project",
        Short: "Create or continue a project (interactive)",
        Long: `Interactive wizard for creating or continuing projects.

The wizard guides you through:
  - Creating new projects from GitHub issues or branch names
  - Continuing existing projects
  - Selecting project types and providing descriptions

Examples:
  sow project                    # Launch interactive wizard
  sow project -- --model opus    # Launch wizard, pass flags to Claude`,
        RunE: runWizard,
        Args: cobra.NoArgs,
    }

    cmd.AddCommand(newSetCmd())
    cmd.AddCommand(newDeleteCmd())

    return cmd
}
```

### Example 2: Verifying Deletion

```bash
# Before deletion - verify shared functions work
cd cli/
go test ./cmd/project/ -run TestShared
# All shared function tests pass

# Build and test old commands use shared functions
go build -o /tmp/sow ./cmd/sow
/tmp/sow project new --branch test/verify "Test"
# Works correctly

# Now safe to delete
rm cmd/project/new.go
rm cmd/project/continue.go

# Verify no references
grep -r "newNewCmd\|newContinueCmd" ./cmd/project/
# No matches (except maybe in test files or comments)

# Build with wizard
go build -o /tmp/sow ./cmd/sow
/tmp/sow project
# Wizard launches
```

### Example 3: Testing the Migration

```bash
# Test wizard launches
$ sow project
╔══════════════════════════════════════════════════════════╗
║                     Sow Project Manager                  ║
╚══════════════════════════════════════════════════════════╝

What would you like to do?

  ○ Create new project
  ○ Continue existing project
  ○ Cancel

# Test old commands fail gracefully
$ sow project new
Error: unknown command "new" for "sow project"
Run 'sow project --help' for usage.

# Test kept commands still work
$ sow project set --help
Set project field values

Usage:
  sow project set [field] [value] [flags]
```

## Dependencies

- **Task 010**: Huh library (wizard uses it)
- **Task 020**: Shared utilities (replaces old logic)
- **Task 030**: Wizard foundation (becomes main command)
- **Task 040**: Helper functions (wizard uses them)

## Constraints

- **Safety first**: Verify old commands work with shared functions before deletion
- **No data loss**: Ensure all functionality is preserved
- **Clean migration**: No half-deleted code or broken references
- **Backward incompatible**: This is a breaking change - document it
- **Keep useful commands**: Don't delete `set` and `delete` subcommands

## Testing Requirements

### Pre-Deletion Tests

**Before deleting files**, run these tests:

```bash
# Test old commands work
cd cli/
go build -o /tmp/sow ./cmd/sow

# Test new command
/tmp/sow project new --branch test/pre-delete "Pre-delete test"
# Should work (creates project)

# Test continue command
/tmp/sow project continue --branch test/pre-delete
# Should work (continues project)

# Test shared functions are used
# Check that changes to shared.go affect behavior
```

### Post-Deletion Tests

**After deleting files**, run these tests:

```bash
# Verify compilation
cd cli/
go build -o /tmp/sow ./cmd/sow
# No errors

# Verify wizard works
/tmp/sow project
# Shows entry screen

# Verify old commands gone
/tmp/sow project new
# Error: unknown command "new"

/tmp/sow project continue
# Error: unknown command "continue"

# Verify other commands work
/tmp/sow project set --help
/tmp/sow project delete --help

# Run test suite
go test ./cmd/project/...
# All tests pass
```

### Integration Tests

**Manual Testing Checklist**:
- [ ] Run `sow project` from various directories
- [ ] Test wizard entry screen navigation
- [ ] Test cancellation (Esc, Cancel option)
- [ ] Test Claude Code flag passing (`sow project -- --verbose`)
- [ ] Verify help text is correct (`sow project --help`)
- [ ] Verify no compilation warnings

## Implementation Notes

### Migration Strategy (from issue section 7)

The three-phase migration approach:

**Phase 1: Extract without breaking** (Task 020)
- Create `shared.go` with extracted functions
- Keep `new.go` and `continue.go` unchanged
- Verify shared functions work in isolation

**Phase 2: Build wizard foundation** (Tasks 030-040)
- Add huh dependency
- Build wizard command structure
- Implement state machine and helpers
- Test wizard in isolation

**Phase 3: Replace commands** (This task)
- Update `new.go`/`continue.go` to use shared functions
- Test that old commands still work
- Update `project.go` to use wizard
- Delete `new.go` and `continue.go`
- Test that wizard is the only way

### Why This Order?

1. **Safety**: Old commands work throughout development
2. **Verification**: Can test shared functions independently
3. **Rollback**: Can revert wizard changes without losing functionality
4. **Confidence**: Each step is verified before next step

### Breaking Changes

This is a **breaking change** for users who:
- Use `sow project new --branch ...` in scripts
- Use `sow project continue --branch ...` in automation
- Rely on flag-based interface

**Mitigation**: The wizard is more powerful and guides users through the same workflows.

### What If Something Breaks?

If tests fail after deletion:
1. **Don't panic** - we have git history
2. **Check error messages** - usually a missing import or reference
3. **Verify shared.go** - ensure all functions are correct
4. **Run git diff** - see what actually changed
5. **Worst case**: `git restore new.go continue.go` and debug

### Future Work

After this task, subsequent work units will:
- Implement real handlers for wizard screens (currently stubs)
- Add validation and error handling
- Implement GitHub issue workflow
- Implement continue workflow
- Add finalization logic

But the **foundation** is complete and the **migration** is done.

## Success Indicators

After completing this task:
1. `sow project` launches the wizard
2. Old `new` and `continue` subcommands are gone
3. Files `new.go` and `continue.go` are deleted
4. No compilation errors or warnings
5. All tests pass
6. No dead code or broken references
7. Documentation is updated
8. Migration is complete and irreversible (by design)
9. Foundation work unit is 100% complete
10. Ready for subsequent work units to build on this foundation
