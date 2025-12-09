# Create libs/exec Module Structure and Executor Interface

## Context

This task is part of creating the `libs/exec` module - a standalone Go module providing a clean command execution abstraction using the ports and adapters pattern. The exec package is being moved from `cli/internal/exec/` to `libs/exec/` as part of a broader refactoring effort to decouple library packages from CLI constructs.

**This is NOT just a file move.** This is a design opportunity to create a clean, testable command execution API as a standalone module that can be reused outside the CLI context.

The exec package is a leaf package with no internal dependencies, making it an ideal candidate for extraction. Once complete, `libs/exec` will be consumed by `libs/git` (future) and various CLI command files.

## Requirements

### 1. Create the Go Module

Create a new Go module at `libs/exec/` with:

```
libs/exec/
├── go.mod           # Module definition
├── go.sum           # Dependencies (if any)
├── doc.go           # Package documentation
├── README.md        # Per READMES.md standard
└── executor.go      # Executor interface (port)
```

**go.mod** should define module path as `github.com/jmgilman/sow/libs/exec` with Go version matching the project (Go 1.25.3 based on libs/schemas).

### 2. Design the Executor Interface (Port)

Design a clean, minimal interface that captures what consumers actually need. The current interface has:

- `Command() string` - Get the wrapped command name
- `Exists() bool` - Check if command exists in PATH
- `Run(args ...string) (stdout, stderr string, err error)` - Execute command
- `RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error)` - Execute with context
- `RunSilent(args ...string) error` - Execute discarding output
- `RunSilentContext(ctx context.Context, args ...string) error` - Silent with context

**Design considerations:**

- The interface should follow the "accept interfaces, return concrete types" pattern (per STYLE.md)
- Context support is essential for cancellation/timeout
- The `Command()` and `Exists()` methods are convenience methods - evaluate if they belong on the interface or just the concrete type
- Keep the interface minimal - only what's actually needed by consumers
- Consider if separate stdout/stderr returns are needed vs combined output

**Recommended interface design:**

```go
// Executor defines the interface for executing shell commands.
type Executor interface {
    // Command returns the command name this executor wraps.
    Command() string

    // Exists checks if the command exists in PATH.
    Exists() bool

    // Run executes the command with the given arguments.
    // Returns stdout, stderr, and error.
    Run(args ...string) (stdout, stderr string, err error)

    // RunContext executes the command with the given arguments and context.
    // The context can be used for cancellation and timeouts.
    RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error)

    // RunSilent executes the command but only returns an error.
    // Stdout and stderr are discarded.
    RunSilent(args ...string) error

    // RunSilentContext is like RunSilent but accepts a context for cancellation.
    RunSilentContext(ctx context.Context, args ...string) error
}
```

### 3. Package Documentation (doc.go)

Create comprehensive package documentation following the existing pattern in `libs/schemas/doc.go`:

- Package overview explaining purpose
- Example usage
- Key types and their relationships
- Include `//go:generate` comment for mock generation (will be implemented in task 030)

### 4. README.md

Create README following READMES.md standard with:

- **Overview**: 1-3 sentences explaining what the package does
- **Quick Start**: Import and basic usage example
- **Usage**: Common tasks (creating executor, running commands, handling errors)
- **Testing**: How consumers should use mocks
- **Configuration**: Any options (if applicable)

## Acceptance Criteria

1. [ ] `libs/exec/go.mod` exists with correct module path `github.com/jmgilman/sow/libs/exec`
2. [ ] `libs/exec/executor.go` contains the `Executor` interface with proper documentation
3. [ ] `libs/exec/doc.go` contains package-level documentation with examples
4. [ ] `libs/exec/README.md` follows READMES.md standard
5. [ ] Module compiles: `cd libs/exec && go build ./...` succeeds
6. [ ] Linting passes: `golangci-lint run` shows no issues for the new files
7. [ ] Interface design is minimal but complete (all methods needed by consumers)
8. [ ] All exported identifiers have doc comments starting with the identifier name

**Test requirements** (verified via compilation, no unit tests needed for interface-only code):
- Interface compiles correctly
- Doc comments are valid

## Technical Details

### File Locations

- Module root: `libs/exec/`
- Interface definition: `libs/exec/executor.go`
- Package docs: `libs/exec/doc.go`
- README: `libs/exec/README.md`
- Module file: `libs/exec/go.mod`

### Naming Conventions (per STYLE.md)

- Interface: `Executor` (describes behavior, no I-prefix or -er suffix needed)
- Package: `exec` (short, lowercase, singular)
- Methods: Verb-based names (Run, Exists)

### Import Structure

The module should have minimal dependencies:
- `context` (stdlib) for context support
- No external dependencies at the interface level

### Doc Comment Format

```go
// Executor defines the interface for executing shell commands.
//
// The interface allows for easy mocking in tests while providing a consistent
// API for command execution across the codebase.
type Executor interface {
    // Command returns the command name this executor wraps.
    Command() string
    // ...
}
```

## Relevant Inputs

- `cli/internal/exec/executor.go` - Current implementation to extract interface from
- `cli/internal/exec/mock.go` - Current mock implementation (shows what methods are needed)
- `libs/schemas/go.mod` - Reference for module structure
- `libs/schemas/doc.go` - Reference for documentation style
- `libs/schemas/README.md` - Reference for README structure
- `.standards/STYLE.md` - Go style standards
- `.standards/READMES.md` - README standards
- `.sow/project/context/issue-115.md` - Full issue requirements

## Examples

### executor.go structure

```go
package exec

import "context"

// Executor defines the interface for executing shell commands.
//
// This interface allows for easy mocking in tests while providing a consistent
// API for command execution across the codebase.
type Executor interface {
    // Command returns the command name this executor wraps.
    Command() string

    // Exists checks if the command exists in PATH.
    Exists() bool

    // Run executes the command with the given arguments.
    // Returns stdout, stderr, and error.
    Run(args ...string) (stdout, stderr string, err error)

    // RunContext executes the command with the given arguments and context.
    // The context can be used for cancellation and timeouts.
    RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error)

    // RunSilent executes the command but only returns an error.
    // Stdout and stderr are discarded.
    RunSilent(args ...string) error

    // RunSilentContext is like RunSilent but accepts a context for cancellation.
    RunSilentContext(ctx context.Context, args ...string) error
}
```

### go.mod structure

```go
module github.com/jmgilman/sow/libs/exec

go 1.25.3
```

### README.md Quick Start example

```go
import "github.com/jmgilman/sow/libs/exec"

// Create an executor for the gh CLI
gh := exec.NewLocalExecutor("gh")

// Check if command exists
if !gh.Exists() {
    return fmt.Errorf("gh CLI not found")
}

// Run a command
stdout, stderr, err := gh.Run("issue", "list", "--label", "bug")
if err != nil {
    return fmt.Errorf("gh issue list failed: %s", stderr)
}
```

## Dependencies

None - this is the first task and creates the foundation.

## Constraints

- **No implementation code in executor.go** - Only the interface definition. Implementation goes in task 020.
- **Do not remove old files** - The old `cli/internal/exec/` files stay until task 050.
- **Do not update consumers** - Consumer migration is task 040.
- **Keep the interface minimal** - Resist adding methods that aren't currently used by consumers.
- Must pass `golangci-lint run` with the project's `.golangci.yml` configuration.
