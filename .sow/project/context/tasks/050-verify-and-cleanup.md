# Verify Lint, Tests, and Clean Up Old Schemas

## Context

This task is the final step of the `libs/schemas` module migration project. After all schemas are migrated and consumers updated (Tasks 010-040), this task verifies everything works correctly and removes the old `cli/schemas/` directory.

This is a critical verification and cleanup task that ensures:
1. All code passes golangci-lint
2. All tests pass
3. The old schemas directory is removed
4. No broken references remain

## Requirements

### 1. Verify libs/schemas Module

Run full verification on the new module:

```bash
cd libs/schemas

# Verify code generation
go generate ./...

# Verify build
go build ./...

# Run all tests
go test ./... -v

# Run linter (if .golangci.yml exists in libs/schemas)
# Otherwise run from repo root
```

### 2. Verify CLI Module

Run full verification on the CLI module:

```bash
cd cli

# Verify module dependencies
go mod tidy

# Verify build
go build ./...

# Run all tests
go test ./...

# Run linter
golangci-lint run
```

### 3. Run golangci-lint from Repository Root

The repo root has a `.golangci.yml` that applies to all modules:

```bash
# From repository root
golangci-lint run ./libs/schemas/...
golangci-lint run ./cli/...
```

Expected: No new issues (the config uses `new: true` from master merge base).

### 4. Remove Old cli/schemas Directory

After verification passes, remove the old schemas:

```bash
rm -rf cli/schemas/
```

### 5. Update cli/go.mod

Remove any remaining self-references to cli/schemas if present.

### 6. Final Verification

Run one more round of verification after cleanup:

```bash
# From repository root
cd cli && go build ./... && go test ./...
cd ../libs/schemas && go build ./... && go test ./...
```

## Acceptance Criteria

1. [ ] `libs/schemas/` passes all tests
2. [ ] `cli/` passes all tests
3. [ ] `golangci-lint run` passes with no new issues in both modules
4. [ ] Old `cli/schemas/` directory is deleted
5. [ ] No remaining imports of `github.com/jmgilman/sow/cli/schemas` anywhere
6. [ ] `go mod tidy` is clean in both modules (no changes after running)
7. [ ] Code generation is idempotent (`go generate ./...` produces no diff)

### Verification Commands

```bash
# Check no old imports remain anywhere
grep -r '"github.com/jmgilman/sow/cli/schemas' --include="*.go" .
# Should return: no matches

# Check cli/schemas directory is gone
ls -la cli/schemas/
# Should return: No such file or directory

# Verify generation is idempotent
cd libs/schemas
go generate ./...
git diff --exit-code
# Should return: exit 0 (no changes)

# Full test suite
cd cli && go test ./...
cd ../libs/schemas && go test ./...
# All tests should pass
```

## Technical Details

### golangci-lint Configuration

The project uses golangci-lint v2 configuration at `.golangci.yml`. Key settings:
- `new: true` - Only reports new issues
- `new-from-merge-base: master` - Compares against master branch

For generated files, the config excludes revive's var-naming rule for `cue_types_gen.go` files.

### Test Count Expectations

Based on the existing code:
- `libs/schemas/project/` should have 53+ tests (from schemas_test.go)
- `cli/` should have all existing tests passing

### Edge Cases to Verify

1. **Embedded CUE files**: Ensure `CUESchemas` embed.FS works correctly
2. **Import aliases**: Some files may use aliases - verify they still work
3. **Integration tests**: Some tests are tagged with `//go:build integration`

### Cleanup Checklist

Before removing cli/schemas/:
- [ ] Verified no imports reference it
- [ ] All consumers use libs/schemas
- [ ] Tests pass in both modules
- [ ] Lint passes

## Relevant Inputs

- `libs/schemas/` - New module to verify
- `cli/` - CLI module to verify
- `.golangci.yml` - Linter configuration at repo root
- `cli/.golangci.yml` - CLI-specific linter configuration

## Examples

### Expected golangci-lint Output

```bash
$ golangci-lint run ./libs/schemas/...

Stats:
  Linters: 20 enabled
  Files: 10
  Issues: 0

# No issues - success!
```

### Expected Test Output

```bash
$ cd libs/schemas && go test ./...
ok      github.com/jmgilman/sow/libs/schemas    0.123s
ok      github.com/jmgilman/sow/libs/schemas/project    1.234s

$ cd ../cli && go test ./...
ok      github.com/jmgilman/sow/cli/cmd    0.456s
ok      github.com/jmgilman/sow/cli/cmd/agent    0.234s
# ... all packages pass
```

### Final Directory State

After completion:
```
libs/
└── schemas/
    ├── go.mod
    ├── go.sum
    ├── README.md
    ├── doc.go
    ├── embed.go
    ├── cue.mod/module.cue
    ├── config.cue
    ├── user_config.cue
    ├── refs_cache.cue
    ├── refs_committed.cue
    ├── refs_local.cue
    ├── knowledge_index.cue
    ├── cue_types_gen.go
    └── project/
        ├── project.cue
        ├── phase.cue
        ├── task.cue
        ├── artifact.cue
        ├── test_helper.cue
        ├── cue_types_gen.go
        └── schemas_test.go

cli/
├── go.mod          # Updated with libs/schemas dependency
├── internal/       # All imports updated to libs/schemas
├── cmd/            # All imports updated to libs/schemas
└── # NO cli/schemas/ directory
```

## Dependencies

- Task 010: Create libs/schemas Go module structure
- Task 020: Migrate project schemas to libs/schemas/project
- Task 030: Update consumer import paths
- Task 040: Add README.md and doc.go documentation

All must be completed before this final verification and cleanup.

## Constraints

- Do NOT remove cli/schemas/ until all verification passes
- Do NOT skip golangci-lint verification
- Do NOT force-push or skip CI checks
- If any verification fails, investigate and fix before cleanup
- The deletion of cli/schemas/ is permanent - ensure everything works first
