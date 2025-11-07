# Task 010: CLI Command Infrastructure and Flag Validation

## Context

This task is part of the CLI Enhanced Advance Command work unit, which adds explicit event selection, discovery, and dry-run validation to the `sow advance` command. This is the first task in the implementation, establishing the foundation for all four operation modes.

The current `sow advance` command only supports auto-determination (calling `DetermineEvent()` and firing that event). We need to add infrastructure to support:
- Optional `[event]` positional argument for explicit event selection
- `--list` flag for discovering available transitions
- `--dry-run` flag for validating transitions without executing

This task focuses solely on command signature changes and flag validation logic. Implementation of the actual modes will follow in subsequent tasks.

## Requirements

### Command Signature Changes

Modify `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/cmd/advance.go`:

1. **Accept optional event argument**:
   - Change line 41: `Args: cobra.NoArgs` to `Args: cobra.MaximumNArgs(1)`
   - This allows `sow advance` (no args) OR `sow advance [event]` (one arg)

2. **Add flags**:
   - Add `--list` boolean flag: Shows available transitions without executing
   - Add `--dry-run` boolean flag: Validates transition without executing
   - Define flags after the `cmd` definition (around line 85)

### Flag Validation Logic

Implement mutual exclusivity rules in the RunE function:

1. **`--list` cannot combine with `[event]`**:
   - Error: "cannot specify event argument with --list flag"
   - Reason: List mode shows all options, event argument selects one specific option

2. **`--dry-run` requires `[event]`**:
   - Error: "--dry-run requires an event argument"
   - Reason: Dry-run validates a specific transition, needs event to validate

3. **Both flags cannot be used together**:
   - Error: "cannot use --list and --dry-run together"
   - Reason: Contradictory operations (list all vs validate one)

### Error Message Format

All validation errors should:
- Clearly state what's wrong
- Be actionable (user knows how to fix)
- Return early (before loading project state)

## Acceptance Criteria

### Functional Tests (TDD)

Write tests BEFORE implementation in `cli/cmd/advance_test.go`:

1. **TestAdvanceCommandSignature**:
   - Verify command accepts 0 or 1 arguments
   - Verify command rejects 2+ arguments
   - Verify both flags are defined and boolean type

2. **TestAdvanceFlagValidation**:
   - `--list` with event argument → error
   - `--dry-run` without event argument → error
   - `--list` and `--dry-run` together → error
   - Valid combinations pass validation:
     - No flags, no args (auto mode)
     - No flags, one arg (explicit event mode)
     - `--list`, no args (discovery mode)
     - `--dry-run`, one arg (dry-run mode)

### Implementation Verification

After tests are written and failing:

1. Modify command Args field
2. Add flag definitions
3. Implement validation logic in RunE
4. All tests pass

### Code Quality

- Flag validation happens BEFORE loading project state (fast failure)
- Error messages are clear and actionable
- No breaking changes to existing auto-determination mode

## Technical Details

### File Structure

All changes in `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/cmd/advance.go`:

```go
// Around line 41 - Change Args validation
Args: cobra.MaximumNArgs(1),  // Changed from cobra.NoArgs

// Around line 85 - Add flag definitions
cmd.Flags().Bool("list", false, "List available transitions without executing")
cmd.Flags().Bool("dry-run", false, "Validate transition without executing")
```

### Flag Validation Pattern

In the RunE function (before loading project):

```go
RunE: func(cmd *cobra.Command, args []string) error {
    // Get flags
    listFlag, _ := cmd.Flags().GetBool("list")
    dryRunFlag, _ := cmd.Flags().GetBool("dry-run")

    // Get event argument if provided
    var event string
    if len(args) > 0 {
        event = args[0]
    }

    // Validate flag combinations
    if listFlag && event != "" {
        return fmt.Errorf("cannot specify event argument with --list flag")
    }

    if dryRunFlag && event == "" {
        return fmt.Errorf("--dry-run requires an event argument")
    }

    if listFlag && dryRunFlag {
        return fmt.Errorf("cannot use --list and --dry-run together")
    }

    // Now proceed to load project and execute mode...
    ctx := cmdutil.GetContext(cmd.Context())
    project, err := state.Load(ctx)
    // ... rest of implementation
}
```

### Testing Setup

Create `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/cmd/advance_test.go`:

```go
package cmd

import (
    "testing"
    "github.com/spf13/cobra"
)

func TestAdvanceCommandSignature(t *testing.T) {
    cmd := NewAdvanceCmd()

    // Test accepts 0 arguments
    err := cmd.Args(cmd, []string{})
    if err != nil {
        t.Errorf("should accept 0 arguments: %v", err)
    }

    // Test accepts 1 argument
    err = cmd.Args(cmd, []string{"event_name"})
    if err != nil {
        t.Errorf("should accept 1 argument: %v", err)
    }

    // Test rejects 2 arguments
    err = cmd.Args(cmd, []string{"event1", "event2"})
    if err == nil {
        t.Error("should reject 2 arguments")
    }

    // Test flags are defined
    if cmd.Flags().Lookup("list") == nil {
        t.Error("--list flag not defined")
    }
    if cmd.Flags().Lookup("dry-run") == nil {
        t.Error("--dry-run flag not defined")
    }
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/cli/cmd/advance.go` - Current advance command implementation to modify
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/project/context/issue-78.md` - Complete specification with command signature details (Section 5)
- `/Users/josh/code/sow/.sow/worktrees/78-cli-enhanced-advance-command-and-standard-project-refactoring/.sow/knowledge/designs/cli-enhanced-advance.md` - CLI design document with flag specifications (Section 5)

## Examples

### Before (Current)

```bash
sow advance              # Only supported mode
sow advance finalize     # ERROR: command accepts no arguments
sow advance --list       # ERROR: unknown flag
```

### After (This Task)

```bash
sow advance              # Still works (auto mode)
sow advance finalize     # Accepted (event specified)
sow advance --list       # Accepted (discovery mode)
sow advance --dry-run finalize  # Accepted (dry-run mode)

# Validation errors
sow advance --list finalize     # ERROR: cannot specify event with --list
sow advance --dry-run           # ERROR: --dry-run requires event
sow advance --list --dry-run    # ERROR: cannot use both flags
```

## Dependencies

None - this is the first task and establishes the foundation.

## Constraints

### Backward Compatibility

- `sow advance` (no arguments, no flags) must continue to work exactly as before
- No changes to auto-determination behavior in this task
- Only adding new capabilities, not modifying existing ones

### Performance

- Flag validation must be fast (no I/O operations)
- Validation happens before loading project state (fail fast)

### Code Style

- Follow existing cobra command patterns in the codebase
- Use consistent error message formatting
- Keep validation logic simple and readable

## Implementation Notes

### TDD Approach

1. Write all tests first (they will fail)
2. Run tests to confirm they fail for the right reasons
3. Implement minimal code to pass tests
4. Refactor if needed
5. Ensure all tests pass

### Test Organization

Tests in `advance_test.go` should be grouped by concern:
- Command signature tests
- Flag validation tests
- (Later tasks will add mode execution tests)

### Edge Cases

- Empty string event argument (treat as missing argument)
- Flag values other than true/false (cobra handles this)
- Multiple invocations of same flag (cobra handles this)

### Next Steps

After this task completes:
- Task 020 will implement auto-determination mode (enhanced)
- Task 030 will implement list mode
- Task 040 will implement dry-run mode
- Task 050 will implement explicit event mode
- Task 060+ will refactor standard project
