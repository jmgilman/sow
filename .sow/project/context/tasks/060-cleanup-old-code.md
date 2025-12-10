# Cleanup Old Config Code from CLI

## Context

This is the final cleanup task. After all consumers have been migrated to use `libs/config`, the old configuration code in `cli/internal/sow/` is no longer needed and should be removed.

This ensures:
1. No dead code remains in the codebase
2. No confusion about which config package to use
3. Clean separation between CLI-specific code and reusable libraries

## Requirements

### Files to Delete

Remove these files from `cli/internal/sow/`:

1. **cli/internal/sow/config.go** - Repo config loading (now in libs/config/repo.go)
2. **cli/internal/sow/user_config.go** - User config loading (now in libs/config/user.go)

### Verify No Remaining References

Before deletion, verify no other files reference the old symbols:
- `LoadConfig`
- `LoadUserConfig`
- `LoadUserConfigFromPath`
- `GetUserConfigPath`
- `ValidateUserConfig`
- `GetADRsPath`
- `GetDesignDocsPath`
- `GetExplorationsPath`
- `ValidExecutorTypes`
- `DefaultExecutorName`
- `DefaultADRsPath`
- `DefaultDesignDocsPath`
- `DefaultExplorationsPath`

### Update cli/internal/sow/ Package

After removing the files, ensure the `sow` package still compiles:
- Check if any remaining files in the package have broken imports
- The package should still contain context.go, fs.go, sow.go, errors.go, etc.

## Acceptance Criteria

1. [ ] `cli/internal/sow/config.go` is deleted
2. [ ] `cli/internal/sow/user_config.go` is deleted
3. [ ] No dangling references to removed symbols
4. [ ] `cli/internal/sow/` package still compiles
5. [ ] All CLI tests pass
6. [ ] `go build ./...` succeeds from cli directory
7. [ ] `go test ./...` passes from cli directory
8. [ ] `golangci-lint run` passes

### Verification Steps

Before deletion:
```bash
cd cli

# Search for any remaining references to old symbols
grep -r "sow\.LoadConfig\|sow\.LoadUserConfig\|sow\.GetUserConfigPath" --include="*.go" .
grep -r "sow\.ValidateUserConfig\|sow\.GetADRsPath\|sow\.GetDesignDocsPath" --include="*.go" .

# Should return no results (or only the files being deleted)
```

After deletion:
```bash
cd cli
go build ./...
go test ./...
golangci-lint run
```

## Technical Details

### What Stays in cli/internal/sow/

These files should remain:
- `context.go` - The Context type that bundles FS, Git, GitHub
- `fs.go` - NewFS() helper
- `sow.go` - Init(), DetectContext()
- `errors.go` - CLI-specific errors (ErrNoProject, ErrNotInitialized)
- `options.go` - CLI-specific options if any

### What Moves to libs/config/

These have been moved (in previous tasks):
- Config loading functions
- User config loading functions
- Path helper functions
- Validation functions
- Default constants

### Git Considerations

When deleting files, use git to track the deletion:
```bash
git rm cli/internal/sow/config.go
git rm cli/internal/sow/user_config.go
```

This ensures the deletion is properly tracked in version control.

## Relevant Inputs

- `cli/internal/sow/config.go` - File to delete
- `cli/internal/sow/user_config.go` - File to delete
- `cli/internal/sow/context.go` - File that should remain (for reference)
- `libs/config/` - New location of config functionality

## Examples

### Verification Example

```bash
# Check for any remaining imports of config symbols from internal/sow
$ grep -r "sow\.\(Load\|Get\|Validate\).*Config" --include="*.go" cli/cmd/

# After Task 050, this should return nothing
# If it returns results, those files need to be updated first
```

### Package Structure After Cleanup

```
cli/internal/sow/
├── context.go      # Context type - STAYS
├── fs.go           # NewFS() - STAYS
├── sow.go          # Init(), DetectContext() - STAYS
├── errors.go       # ErrNoProject, etc. - STAYS
└── options.go      # CLI options - STAYS (if exists)

# DELETED:
# config.go         - Moved to libs/config/repo.go
# user_config.go    - Moved to libs/config/user.go
```

## Dependencies

- Task 050 (update consumers) must be completed first
- All consumers must be successfully migrated before deletion

## Constraints

- Do NOT delete any files that are still referenced
- Do NOT modify any remaining files unless necessary for compilation
- Verify build and tests pass before considering task complete
- Use `git rm` for proper version control tracking
