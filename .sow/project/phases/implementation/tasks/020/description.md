# Implement LocalExecutor with Tests

## Context

This task is part of creating the `libs/exec` module - a standalone Go module providing a clean command execution abstraction. Task 010 created the module structure and `Executor` interface. This task implements `LocalExecutor`, the concrete type that executes commands on the local system using `os/exec`.

The LocalExecutor is the primary adapter in the ports-and-adapters pattern, where `Executor` is the port (interface) and `LocalExecutor` is the adapter (implementation).

## Requirements

### 1. Implement LocalExecutor (local.go)

Create `libs/exec/local.go` with the `LocalExecutor` implementation:

```go
// LocalExecutor executes commands on the local system.
type LocalExecutor struct {
    command string
}
```

**Constructor:**
```go
// NewLocalExecutor creates a new LocalExecutor for the specified command.
//
// The command should be the base command name (e.g., "gh", "claude", "git").
// The executor will use exec.LookPath to find the command in PATH.
func NewLocalExecutor(command string) *LocalExecutor
```

**Methods to implement:**

1. `Command() string` - Return the command name
2. `Exists() bool` - Check via `exec.LookPath`
3. `Run(args ...string) (stdout, stderr string, err error)` - Delegate to RunContext with `context.Background()`
4. `RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error)` - Core implementation using `exec.CommandContext`
5. `RunSilent(args ...string) error` - Delegate to Run, discard output
6. `RunSilentContext(ctx context.Context, args ...string) error` - Delegate to RunContext, discard output

**Optional convenience method:**
```go
// MustExist panics if the command doesn't exist in PATH.
// This is useful in initialization code where the command is required.
func (e *LocalExecutor) MustExist()
```

### 2. Write Tests (local_test.go)

Create comprehensive behavioral tests in `libs/exec/local_test.go`. Follow TDD approach - write tests alongside implementation.

**Test behaviors to cover:**

1. **Command existence checking:**
   - Test `Exists()` returns true for known commands (e.g., "echo", "true")
   - Test `Exists()` returns false for non-existent commands

2. **Basic command execution:**
   - Test `Run()` captures stdout correctly
   - Test `Run()` captures stderr correctly
   - Test `Run()` returns error for failed commands
   - Test `Run()` returns correct exit status on failure

3. **Context support:**
   - Test `RunContext()` respects context cancellation
   - Test `RunContext()` respects context timeout

4. **Silent execution:**
   - Test `RunSilent()` returns nil on success
   - Test `RunSilent()` returns error on failure
   - Test `RunSilentContext()` respects context cancellation

5. **Edge cases:**
   - Test with empty args
   - Test with args containing spaces
   - Test `MustExist()` panics for non-existent command

**Test approach:**
- Use real commands that exist on all Unix systems: `echo`, `true`, `false`, `sh`
- Use table-driven tests with `t.Run()` for multiple scenarios
- Use `testify/assert` and `testify/require`
- No mocking needed - we're testing the concrete implementation
- Use `t.TempDir()` if any file operations are needed (unlikely)

### 3. Ensure Interface Compliance

Add compile-time interface check:
```go
// Compile-time check that LocalExecutor implements Executor.
var _ Executor = (*LocalExecutor)(nil)
```

## Acceptance Criteria

1. [ ] `libs/exec/local.go` contains `LocalExecutor` implementation
2. [ ] `libs/exec/local_test.go` contains comprehensive behavioral tests
3. [ ] All `Executor` interface methods are implemented
4. [ ] Compile-time interface check is present
5. [ ] Tests pass: `cd libs/exec && go test -v ./...`
6. [ ] Tests pass with race detector: `cd libs/exec && go test -race ./...`
7. [ ] Linting passes: `golangci-lint run` shows no issues
8. [ ] Error handling follows STYLE.md (proper wrapping with `%w`)
9. [ ] All exported identifiers have doc comments

**Test coverage requirements:**
- All public methods have at least one test
- Error paths are tested
- Context cancellation is tested
- Edge cases (empty args, spaces in args) are tested

## Technical Details

### File Locations

- Implementation: `libs/exec/local.go`
- Tests: `libs/exec/local_test.go`

### Implementation Pattern

```go
package exec

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
)

// LocalExecutor executes commands on the local system.
type LocalExecutor struct {
    command string
}

// NewLocalExecutor creates a new LocalExecutor for the specified command.
func NewLocalExecutor(command string) *LocalExecutor {
    return &LocalExecutor{
        command: command,
    }
}

// RunContext executes the command with context support.
func (e *LocalExecutor) RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error) {
    cmd := exec.CommandContext(ctx, e.command, args...)

    var stdoutBuf, stderrBuf bytes.Buffer
    cmd.Stdout = &stdoutBuf
    cmd.Stderr = &stderrBuf

    err = cmd.Run()

    return stdoutBuf.String(), stderrBuf.String(), err
}
```

### Test Pattern (following TESTING.md)

```go
package exec

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLocalExecutor_Exists(t *testing.T) {
    tests := []struct {
        name    string
        command string
        want    bool
    }{
        {name: "echo exists", command: "echo", want: true},
        {name: "true exists", command: "true", want: true},
        {name: "nonexistent command", command: "definitely-not-a-command-12345", want: false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            e := NewLocalExecutor(tt.command)
            got := e.Exists()
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestLocalExecutor_Run(t *testing.T) {
    t.Run("captures stdout", func(t *testing.T) {
        e := NewLocalExecutor("echo")
        stdout, stderr, err := e.Run("hello")

        require.NoError(t, err)
        assert.Equal(t, "hello\n", stdout)
        assert.Empty(t, stderr)
    })

    t.Run("captures stderr", func(t *testing.T) {
        e := NewLocalExecutor("sh")
        stdout, stderr, err := e.Run("-c", "echo error >&2")

        require.NoError(t, err)
        assert.Empty(t, stdout)
        assert.Equal(t, "error\n", stderr)
    })

    t.Run("returns error for failed command", func(t *testing.T) {
        e := NewLocalExecutor("false")
        _, _, err := e.Run()

        assert.Error(t, err)
    })
}

func TestLocalExecutor_RunContext_Cancellation(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
    defer cancel()

    e := NewLocalExecutor("sleep")
    _, _, err := e.RunContext(ctx, "10")

    assert.Error(t, err)
    assert.ErrorIs(t, err, context.DeadlineExceeded)
}
```

### Receiver Names (per STYLE.md)

Use short receiver name `e` for `LocalExecutor`:
```go
func (e *LocalExecutor) Run(...) ...
```

### Error Handling

Don't wrap errors from `os/exec` - they already contain sufficient context. Return them directly:
```go
err = cmd.Run()
return stdoutBuf.String(), stderrBuf.String(), err  // Don't wrap
```

## Relevant Inputs

- `cli/internal/exec/executor.go` - Current implementation to reference
- `libs/exec/executor.go` - Interface definition (from task 010)
- `.standards/STYLE.md` - Go style standards
- `.standards/TESTING.md` - Testing standards
- `.sow/project/context/issue-115.md` - Full issue requirements

## Examples

### local.go complete structure

```go
package exec

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
)

// LocalExecutor executes commands on the local system using os/exec.
type LocalExecutor struct {
    command string
}

// NewLocalExecutor creates a new LocalExecutor for the specified command.
//
// The command should be the base command name (e.g., "gh", "claude", "git").
// The executor will use exec.LookPath to find the command in PATH.
//
// Example:
//
//     gh := exec.NewLocalExecutor("gh")
//     claude := exec.NewLocalExecutor("claude")
func NewLocalExecutor(command string) *LocalExecutor {
    return &LocalExecutor{
        command: command,
    }
}

// Command returns the command name this executor wraps.
func (e *LocalExecutor) Command() string {
    return e.command
}

// Exists checks if the command exists in PATH.
func (e *LocalExecutor) Exists() bool {
    _, err := exec.LookPath(e.command)
    return err == nil
}

// Run executes the command with the given arguments.
func (e *LocalExecutor) Run(args ...string) (stdout, stderr string, err error) {
    return e.RunContext(context.Background(), args...)
}

// RunContext executes the command with context support.
func (e *LocalExecutor) RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error) {
    cmd := exec.CommandContext(ctx, e.command, args...)

    var stdoutBuf, stderrBuf bytes.Buffer
    cmd.Stdout = &stdoutBuf
    cmd.Stderr = &stderrBuf

    err = cmd.Run()

    return stdoutBuf.String(), stderrBuf.String(), err
}

// RunSilent executes the command discarding output.
func (e *LocalExecutor) RunSilent(args ...string) error {
    _, _, err := e.Run(args...)
    return err
}

// RunSilentContext executes the command with context, discarding output.
func (e *LocalExecutor) RunSilentContext(ctx context.Context, args ...string) error {
    _, _, err := e.RunContext(ctx, args...)
    return err
}

// MustExist panics if the command doesn't exist in PATH.
func (e *LocalExecutor) MustExist() {
    if !e.Exists() {
        panic(fmt.Sprintf("required command %q not found in PATH", e.command))
    }
}

// Compile-time check that LocalExecutor implements Executor.
var _ Executor = (*LocalExecutor)(nil)
```

## Dependencies

- Task 010 must be completed first (interface definition)

## Constraints

- **Do not remove old files** - The old `cli/internal/exec/` files stay until task 050
- **Do not update consumers** - Consumer migration is task 040
- **Follow TDD** - Write tests alongside implementation, not after
- Tests must work on macOS/Linux (use portable commands like `echo`, `true`, `false`, `sh`)
- Must pass `golangci-lint run` with the project's `.golangci.yml` configuration
- No hand-written mock.go - mock generation is task 030
