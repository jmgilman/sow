# Cleanup Old Location and Final Verification

## Context

This is the final task in creating the `libs/exec` module. Tasks 010-040 created the new module and migrated all consumers. This task removes the old `cli/internal/exec/` directory and performs final verification to ensure everything works correctly.

## Requirements

### 1. Remove Old exec Directory

Delete the entire `cli/internal/exec/` directory:
- `cli/internal/exec/executor.go` (old interface and LocalExecutor)
- `cli/internal/exec/mock.go` (old hand-written MockExecutor)

### 2. Verify No References Remain

Search the codebase for any remaining references to `cli/internal/exec`:
```bash
grep -r "cli/internal/exec" --include="*.go"
```

This should return no results.

### 3. Run Full Build and Tests

Verify the entire project builds and tests pass:

```bash
# Build libs/exec module
cd libs/exec && go build ./...

# Test libs/exec module
cd libs/exec && go test -v ./...

# Build CLI (which depends on libs/exec)
cd cli && go build ./...

# Test CLI
cd cli && go test ./...
```

### 4. Run Linting

Verify all linting passes:
```bash
golangci-lint run
```

### 5. Verify Module Structure

Confirm the final structure matches the target:
```
libs/exec/
├── go.mod                 # github.com/jmgilman/sow/libs/exec
├── go.sum
├── README.md              # Per READMES.md standard
├── doc.go                 # Package documentation
├── executor.go            # Executor interface (port)
├── local.go               # LocalExecutor implementation (adapter)
├── local_test.go          # Tests for LocalExecutor
└── mocks/
    └── executor.go        # Generated mock via moq
```

### 6. Verify Acceptance Criteria from Issue

Cross-check against the original issue acceptance criteria:

1. [ ] New `libs/exec` Go module exists and compiles
2. [ ] Clean `Executor` interface designed
3. [ ] `LocalExecutor` implementation works correctly
4. [ ] Mock generated via `moq` in `mocks/` subpackage
5. [ ] All tests pass with proper behavioral coverage
6. [ ] `golangci-lint run` passes with no issues
7. [ ] README.md follows READMES.md standard
8. [ ] Package documentation in doc.go
9. [ ] All 8 consumer files updated to new import paths
10. [ ] Old `cli/internal/exec/` removed

## Acceptance Criteria

1. [ ] `cli/internal/exec/` directory is completely removed
2. [ ] No references to `cli/internal/exec` remain in the codebase
3. [ ] `cd libs/exec && go build ./...` succeeds
4. [ ] `cd libs/exec && go test -v ./...` passes
5. [ ] `cd libs/exec && go test -race ./...` passes
6. [ ] `cd cli && go build ./...` succeeds
7. [ ] `cd cli && go test ./...` passes
8. [ ] `golangci-lint run` passes for both modules
9. [ ] All original issue acceptance criteria are met

## Technical Details

### Files to Remove

```
cli/internal/exec/
├── executor.go  # DELETE
└── mock.go      # DELETE
```

### Verification Commands

```bash
# Check for lingering references
grep -r "cli/internal/exec" --include="*.go"
# Should return nothing

# Verify libs/exec module
cd libs/exec
go mod tidy
go build ./...
go test -v ./...
go test -race ./...

# Verify CLI module
cd cli
go mod tidy
go build ./...
go test ./...

# Lint both
golangci-lint run --config .golangci.yml
```

### Final Directory Structure

After cleanup, the exec-related files should be:

```
# New location (libs/exec/)
libs/exec/
├── go.mod
├── go.sum
├── README.md
├── doc.go
├── executor.go
├── local.go
├── local_test.go
└── mocks/
    └── executor.go

# Old location (should not exist)
cli/internal/exec/  # DELETED
```

## Relevant Inputs

- `cli/internal/exec/executor.go` - File to delete
- `cli/internal/exec/mock.go` - File to delete
- `libs/exec/` - New module location (from tasks 010-030)
- `.sow/project/context/issue-115.md` - Original issue with acceptance criteria

## Examples

### Verification Script

```bash
#!/bin/bash
set -e

echo "=== Checking for lingering references ==="
if grep -r "cli/internal/exec" --include="*.go" 2>/dev/null; then
    echo "ERROR: Found references to cli/internal/exec"
    exit 1
fi
echo "OK: No lingering references"

echo "=== Building libs/exec ==="
cd libs/exec
go mod tidy
go build ./...
echo "OK: libs/exec builds"

echo "=== Testing libs/exec ==="
go test -v ./...
go test -race ./...
echo "OK: libs/exec tests pass"

echo "=== Building CLI ==="
cd ../../cli
go build ./...
echo "OK: CLI builds"

echo "=== Testing CLI ==="
go test ./...
echo "OK: CLI tests pass"

echo "=== Linting ==="
cd ..
golangci-lint run
echo "OK: Linting passes"

echo "=== All checks passed ==="
```

## Dependencies

- All previous tasks (010-040) must be completed first
- All consumers must be migrated before removing old files

## Constraints

- **Remove files only after verification** - First confirm no references, then delete
- **Run all tests before considering complete** - Both unit and race tests
- **Do not leave broken builds** - If something fails, investigate and fix before proceeding
- Must pass `golangci-lint run` with the project's `.golangci.yml` configuration
