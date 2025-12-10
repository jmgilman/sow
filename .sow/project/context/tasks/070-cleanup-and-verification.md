# Task 070: Cleanup Old Code and Final Verification

## Context

This task is the final task in work unit 004: creating the `libs/git` Go module. All functionality has been implemented in libs/git (tasks 010-050) and consumers have been updated (task 060). This task removes the old code from cli/internal/sow/ and performs final verification.

**Goal**: Clean slate - the old git/github code is completely removed and all tests/builds pass.

## Requirements

### 1. Remove Migrated Files from cli/internal/sow/

Delete these files (they've been moved to libs/git):

```
cli/internal/sow/
├── git.go              # DELETE - moved to libs/git/git.go
├── github_client.go    # DELETE - moved to libs/git/client.go
├── github_cli.go       # DELETE - moved to libs/git/client_cli.go
├── github_factory.go   # DELETE - moved to libs/git/factory.go
├── github_mock.go      # DELETE - replaced by libs/git/mocks/
├── worktree.go         # DELETE - moved to libs/git/worktree.go
└── Test files:
    ├── github_cli_test.go     # DELETE
    ├── github_factory_test.go # DELETE
    ├── github_mock_test.go    # DELETE
    └── worktree_test.go       # DELETE
```

### 2. Verify Remaining Files in cli/internal/sow/

These files should remain (possibly with updated imports):
```
cli/internal/sow/
├── context.go          # KEEP - updated to use libs/git
├── context_test.go     # KEEP - updated as needed
├── fs.go               # KEEP - unchanged
├── sow.go              # KEEP - unchanged
├── errors.go           # KEEP - unchanged (ErrNotInitialized, etc.)
├── config.go           # KEEP - unchanged
├── user_config.go      # KEEP - unchanged
├── user_config_test.go # KEEP - unchanged
└── options.go          # KEEP - unchanged (if exists)
```

### 3. Run Full Test Suite

Execute comprehensive verification:

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run linting
golangci-lint run

# Verify no import cycles
go list -json ./... | jq -r '.ImportPath' | sort
```

### 4. Verify libs/git Independence

Confirm libs/git has no dependencies on cli/internal/:
```bash
cd libs/git
go list -deps ./... | grep -E "cli/internal" && echo "ERROR: libs/git depends on cli!" || echo "OK: libs/git is independent"
```

### 5. Verify Consumer Updates

Check that no files still import the old paths:
```bash
grep -r "cli/internal/sow/.*git\|cli/internal/sow/.*github\|cli/internal/sow/.*worktree" --include="*.go" .
```

This should return no results.

### 6. Update go.work (if using workspace)

If the project uses a go.work file, ensure libs/git is included:
```go
// go.work
use (
    ./cli
    ./libs/exec
    ./libs/git
    // ...
)
```

## Acceptance Criteria

1. **Old files deleted**: All git/github files removed from cli/internal/sow
2. **Remaining files valid**: context.go, fs.go, sow.go, errors.go, config.go compile
3. **No orphan imports**: No import errors in remaining files
4. **Full build passes**: `go build ./...` from repo root succeeds
5. **All tests pass**: `go test ./...` passes (including race detector)
6. **Linting passes**: `golangci-lint run` has no issues
7. **libs/git independent**: No cli/internal dependencies in libs/git
8. **No old import paths**: No grep matches for old git/github imports
9. **Module compiles standalone**: `cd libs/git && go build ./...` works
10. **Consumer tests work**: cli tests using git functionality pass

### Verification Checklist

Run these commands and verify all pass:

```bash
# From repository root
cd /path/to/sow

# Build everything
go build ./...

# Test everything
go test ./...

# Test with race detector
go test -race ./...

# Lint
golangci-lint run

# Verify libs/git is independent
cd libs/git
go list -deps ./... | grep "cli/internal" && exit 1 || echo "libs/git is independent"
cd ..

# Verify no old imports remain
grep -r "internal/sow.*Git\|internal/sow.*GitHub\|internal/sow.*Worktree" --include="*.go" cli/ && echo "ERROR: Old imports found" || echo "OK: No old imports"

# Verify deleted files don't exist
for f in git.go github_client.go github_cli.go github_factory.go github_mock.go worktree.go; do
    test -f cli/internal/sow/$f && echo "ERROR: $f still exists" || echo "OK: $f deleted"
done
```

## Technical Details

### Files to Delete

| File | Reason |
|------|--------|
| cli/internal/sow/git.go | Moved to libs/git/git.go |
| cli/internal/sow/github_client.go | Moved to libs/git/client.go |
| cli/internal/sow/github_cli.go | Moved to libs/git/client_cli.go |
| cli/internal/sow/github_factory.go | Moved to libs/git/factory.go |
| cli/internal/sow/github_mock.go | Replaced by libs/git/mocks/ |
| cli/internal/sow/worktree.go | Moved to libs/git/worktree.go |
| cli/internal/sow/github_cli_test.go | Moved to libs/git/client_cli_test.go |
| cli/internal/sow/github_factory_test.go | No longer needed |
| cli/internal/sow/github_mock_test.go | No longer needed |
| cli/internal/sow/worktree_test.go | Moved to libs/git/worktree_test.go |

### Import Graph After Migration

```
libs/exec     (standalone, no sow deps)
     ↓
libs/git      (depends on libs/exec only)
     ↓
cli/internal/sow  (depends on libs/git, libs/exec)
     ↓
cli/cmd/*     (depends on cli/internal/sow, libs/git)
```

### Error Recovery

If tests fail after cleanup:
1. Check for missing imports in context.go
2. Verify all consumer files were updated in task 060
3. Check for type mismatches (git.Issue vs sow.Issue)
4. Run `go mod tidy` in both cli and libs/git directories

## Relevant Inputs

- Task 060 completion - consumers must be updated first
- `cli/internal/sow/` - files to verify and potentially clean up
- `libs/git/` - new module (from tasks 010-050)
- `.golangci.yml` - linting configuration
- `go.work` - workspace configuration (if exists)
- `.standards/STYLE.md` - Coding standards to verify compliance
- `.standards/TESTING.md` - Testing standards to verify compliance

## Examples

### File Deletion Commands

```bash
# From repository root
rm cli/internal/sow/git.go
rm cli/internal/sow/github_client.go
rm cli/internal/sow/github_cli.go
rm cli/internal/sow/github_factory.go
rm cli/internal/sow/github_mock.go
rm cli/internal/sow/worktree.go

# Test files
rm cli/internal/sow/github_cli_test.go
rm cli/internal/sow/github_factory_test.go
rm cli/internal/sow/github_mock_test.go
rm cli/internal/sow/worktree_test.go
```

### Full Verification Script

```bash
#!/bin/bash
set -e

echo "=== Building all packages ==="
go build ./...

echo "=== Running tests ==="
go test ./...

echo "=== Running tests with race detector ==="
go test -race ./...

echo "=== Running linter ==="
golangci-lint run

echo "=== Checking libs/git independence ==="
cd libs/git
if go list -deps ./... | grep -q "cli/internal"; then
    echo "ERROR: libs/git has dependency on cli/internal!"
    exit 1
fi
cd ../..

echo "=== Checking for old imports ==="
if grep -r "internal/sow.*[Gg]it\|internal/sow.*[Gg]ithub\|internal/sow.*[Ww]orktree" --include="*.go" cli/; then
    echo "ERROR: Old imports still exist!"
    exit 1
fi

echo "=== All checks passed! ==="
```

## Dependencies

- Task 060 must complete first - all consumers must be updated before deleting old code
- All previous tasks (010-050) must be complete - libs/git must be fully functional

## Constraints

- Do NOT delete files until consumers are updated (task 060)
- Do NOT delete context.go, fs.go, sow.go, errors.go - they're still needed
- If tests fail, fix consumer code rather than keeping old code
- Maintain backward compatibility for any external code using cli/internal/sow
- All CI checks must pass before considering the task complete
