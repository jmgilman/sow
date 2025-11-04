# Task 080: Remove Old Commands and Cleanup

# Task 080: Remove Old Commands and Cleanup

## Overview

Remove all deprecated commands that have been replaced by the unified command structure. Clean up any dead code and ensure the CLI is in a clean, maintainable state.

## Context

After implementing the unified command structure, old commands under `cli/cmd/agent/` are no longer used. These must be removed to avoid confusion and reduce maintenance burden.

## Requirements

### Commands/Files to Remove

**Old agent commands** (replaced by unified commands):

```
cli/cmd/agent/artifact.go         → Replaced by sow input/output
cli/cmd/agent/artifact_add.go     → Replaced by sow input/output add
cli/cmd/agent/artifact_approve.go → Replaced by sow input/output set --index N approved true
cli/cmd/agent/artifact_list.go    → Replaced by sow input/output list
```

**Check for other old command patterns**:
- Any commands referencing old `internal/project` package (should use SDK now)
- Any commands with old artifact/task/phase patterns
- Any dead code from old implementations

**Do NOT remove**:
- `cli/cmd/agent/log.go` - Still used for logging
- `cli/cmd/agent/agent.go` - Root agent command (may still have subcommands)
- Other agent commands that aren't replaced (like status, info, etc.)
- Worktree-related commands

## Approach

### Step 1: Verify All Tests Pass

Before removing code, ensure all integration tests from Task 070 are passing. This confirms the new commands fully replace old functionality.

```bash
# Run all integration tests
go test ./cli/testdata/script/...
```

### Step 2: Remove Old Command Files

Delete identified files:

```bash
rm cli/cmd/agent/artifact.go
rm cli/cmd/agent/artifact_add.go
rm cli/cmd/agent/artifact_approve.go
rm cli/cmd/agent/artifact_list.go
# ... other identified files
```

### Step 3: Update Agent Command Registration

Modify `cli/cmd/agent/agent.go` to remove references to deleted commands:

**Before**:
```go
cmd.AddCommand(newArtifactCmd())
// ... other commands
```

**After**:
```go
// Artifact commands removed - use sow input/output instead
// ... other commands
```

### Step 4: Search for References

Search codebase for any remaining references to removed commands:

```bash
# Search for old command references
git grep -n "artifactCmd"
git grep -n "artifact_add"
git grep -n "artifact_approve"
# etc.
```

Remove or update any found references.

### Step 5: Check Imports

Look for unused imports that were only needed for removed commands:

```go
// Old
import (
    "github.com/jmgilman/sow/cli/internal/project/domain"  // May be unused now
)

// New - remove if no longer used
```

### Step 6: Update Help Text

Ensure CLI help text doesn't reference removed commands:

```bash
sow --help
sow agent --help
```

Review output and update any stale documentation.

### Step 7: Final Test Run

Run all tests again to ensure nothing broke:

```bash
go test ./...
```

### Step 8: Verify Clean Build

```bash
go build ./cli
```

Should complete with no errors or warnings.

## Files to Remove

### Confirmed Removals

- `cli/cmd/agent/artifact.go`
- `cli/cmd/agent/artifact_add.go`
- `cli/cmd/agent/artifact_approve.go`
- `cli/cmd/agent/artifact_list.go`

### Investigate for Removal

Check these patterns and remove if they're old command implementations:

```bash
# Find potential old commands
find cli/cmd/agent -name "*.go" | grep -E "(task|phase|artifact)"
```

Review each file:
- If it implements old command patterns (pre-unified) → Remove
- If it's still actively used → Keep

### Do Not Remove

- `cli/cmd/agent/log.go` - Logging is still valid
- `cli/cmd/agent/status.go` - Status commands still valid
- `cli/cmd/agent/info.go` - Info commands still valid
- Any worktree-related commands
- Root command files (`cli/cmd/root.go`, `cli/cmd/agent/agent.go`)

## Files to Modify

### `cli/cmd/agent/agent.go`

Remove command registrations for deleted commands.

### `cli/cmd/root.go` (if needed)

Ensure root command doesn't reference removed commands.

## Acceptance Criteria

- [ ] All old agent artifact commands removed
- [ ] No references to removed commands in codebase
- [ ] CLI help text accurate (no stale references)
- [ ] All tests passing after cleanup
- [ ] Clean build with no errors
- [ ] No dead code remaining
- [ ] Git grep finds no references to old command patterns

## Testing Strategy

**Verification only** - No new tests needed.

1. Run all integration tests before cleanup
2. Remove old code
3. Run all tests again
4. Verify clean build
5. Manual CLI help text verification

## Dependencies

- Task 070 (Integration Testing) - All tests must pass before cleanup

## References

- **Old commands**: `cli/cmd/agent/artifact*.go`
- **Agent root**: `cli/cmd/agent/agent.go`
- **Command registration patterns**: Review existing command structure

## Notes

- This is a cleanup task - be thorough but careful
- Verify each file before deletion (review git history if unsure)
- Keep a backup or rely on git to restore if needed
- Focus on removing old command implementations, not shared utilities
- After this task, the CLI should have only unified commands
- Old integration tests were already deleted in Task 070

## Checklist for Removal

For each file/command being considered for removal:

- [ ] Is this command replaced by a unified command?
- [ ] Are there any remaining references in the codebase?
- [ ] Is there a test that would fail without this?
- [ ] Is this command documented anywhere that needs updating?
- [ ] Does removing this break the build?

If all checks pass → Safe to remove.

## Final Verification

After all removals:

```bash
# Verify no dead imports
go mod tidy

# Verify clean build
go build ./cli

# Verify all tests pass
go test ./...

# Verify no references to old patterns
git grep -i "sow agent artifact"
git grep -i "sow agent phase approve"
git grep -i "sow agent task add-reference"

# Should return no results (or only in documentation/comments explaining migration)
```
