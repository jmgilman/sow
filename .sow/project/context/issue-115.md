# Issue #115: libs/exec module

**URL**: https://github.com/jmgilman/sow/issues/115
**State**: OPEN

## Description

# Work Unit 002: libs/exec Module

Create a new `libs/exec` Go module providing a clean command execution abstraction using the ports and adapters pattern.

## Objective

Move `cli/internal/exec/` to `libs/exec/`, creating a standalone Go module with a well-designed interface (port) and implementation (adapter), full standards compliance, and proper documentation.

**This is NOT just a file move.** This is an opportunity to design a clean, testable command execution API.

## Scope

### What's Moving
- `Executor` interface
- `LocalExecutor` implementation
- `MockExecutor` for testing

### Target Structure
```
libs/exec/
├── go.mod
├── go.sum
├── README.md                 # Per READMES.md standard
├── doc.go                    # Package documentation
│
├── executor.go               # Executor interface (port)
├── local.go                  # LocalExecutor implementation (adapter)
├── local_test.go             # Tests for LocalExecutor
│
└── mocks/
    └── executor.go           # Generated mock via moq
```

## Standards Requirements

### Go Code Standards (STYLE.md)
- Accept interfaces, return concrete types
- Error handling with proper wrapping (`%w`)
- No global mutable state
- Proper naming: `Executor` interface, `LocalExecutor` concrete type
- Short receiver names (1-2 chars)
- Functions under 80 lines

### Testing Standards (TESTING.md)
- Behavioral test coverage (test what it does, not how)
- Table-driven tests with `t.Run()`
- Use `testify/assert` and `testify/require`
- Mock generation via `moq` for the `Executor` interface
- No external dependencies in unit tests (use temp directories for any file operations)

### README Standards (READMES.md)
- Overview: Command execution abstraction for sow
- Quick Start: Import and create executor
- Usage: Running commands, capturing output, handling errors
- Testing: How to use mocks

### Linting
- Must pass `golangci-lint run` with project's `.golangci.yml`
- Proper error wrapping for external errors

## API Design Requirements

### Ports and Adapters Pattern

**Port (Interface):**
```go
// Executor defines the interface for executing shell commands.
type Executor interface {
    // Run executes a command and returns combined output.
    Run(ctx context.Context, name string, args ...string) ([]byte, error)

    // RunInDir executes a command in a specific directory.
    RunInDir(ctx context.Context, dir, name string, args ...string) ([]byte, error)
}
```

**Adapter (Implementation):**
```go
// LocalExecutor executes commands on the local system.
type LocalExecutor struct {
    // Optional configuration
}

func NewLocalExecutor() *LocalExecutor
```

### Design Principles
- Interface should be minimal - only what's actually needed
- Implementation details (buffering, environment, etc.) hidden
- Easy to mock for testing consumers
- Context support for cancellation/timeout

## Consumer Impact

~8 files currently import from `cli/internal/exec/`. After this work:
- All imports change to `github.com/jmgilman/sow/libs/exec`
- Consumers of `libs/git` will use this indirectly

## Dependencies

None - this is a leaf package with no internal dependencies.

## Acceptance Criteria

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

## Out of Scope

- Adding new execution modes (streaming, pipes, etc.)
- Platform-specific implementations
- Shell interpretation (commands run directly, not via shell)
