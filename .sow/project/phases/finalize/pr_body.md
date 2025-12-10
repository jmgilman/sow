## Summary

Create a new `libs/exec` Go module providing a clean command execution abstraction using the ports and adapters pattern. This extracts `cli/internal/exec/` into a standalone, reusable module with a well-designed interface, generated mocks via moq, and full standards compliance.

This is part of the broader libs/ refactoring effort to decouple library packages from CLI constructs, enabling reuse outside the CLI context.

## Changes

### New `libs/exec` Module

- **Executor interface** (`executor.go`): 6 methods covering common use cases
  - `Command()` - returns wrapped command name
  - `Exists()` - checks if command exists in PATH
  - `Run(args...)` - execute with stdout/stderr capture
  - `RunContext(ctx, args...)` - execute with context support
  - `RunSilent(args...)` - execute discarding output
  - `RunSilentContext(ctx, args...)` - silent with context

- **LocalExecutor** (`local.go`): Concrete implementation using `os/exec`
  - `NewLocalExecutor(command)` constructor
  - `MustExist()` convenience method

- **Generated mock** (`mocks/executor.go`): Via moq with `//go:generate` directive

- **Documentation**: README.md (READMES.md standard), doc.go (package docs)

### Consumer Migration

Updated 8 files to use `github.com/jmgilman/sow/libs/exec`:
- `cli/internal/sow/github_cli.go`
- `cli/internal/sow/github_cli_test.go`
- `cli/internal/sow/github_factory.go`
- `cli/cmd/project/wizard_state.go`
- `cli/cmd/project/shared.go`
- `cli/cmd/issue/show.go`
- `cli/cmd/issue/list.go`
- `cli/cmd/issue/check.go`

### Cleanup

- Removed `cli/internal/exec/` directory

## Testing

- 21 behavioral test cases in `libs/exec/local_test.go`
- Table-driven tests with `t.Run()`
- Uses testify/assert and testify/require
- Tests context cancellation, timeouts, error paths, edge cases
- All CLI tests pass after migration
- Race detector passes: `go test -race ./...`
- Linting passes: `golangci-lint run ./...` (0 issues)

## Notes

- Interface design wraps a single command at construction (`NewLocalExecutor("gh")`) rather than per-call, better matching actual usage patterns
- `cli/go.mod` uses `replace` directive for local development

Closes #115

Generated with [sow](https://github.com/jmgilman/sow)
