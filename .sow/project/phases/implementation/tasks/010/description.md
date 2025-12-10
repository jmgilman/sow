# Task 010: Create libs/git Module Foundation and Types

## Context

This task is part of work unit 004: creating a new `libs/git` Go module. The overall goal is to move Git/GitHub code from `cli/internal/sow/` to `libs/git/`, creating a standalone Go module with clean interfaces (ports) and implementations (adapters), decoupled from `sow.Context`.

This is the foundational task that creates the module structure, defines shared types, and establishes error handling patterns. Subsequent tasks will build on this foundation.

**Critical Design Principle**: The new module must NOT depend on `sow.Context`. Functions accept explicit parameters (repo root paths, executors) instead of a Context object.

## Requirements

### 1. Create Module Structure

Create the following directory structure:
```
libs/git/
├── go.mod                # Module definition
├── go.sum                # Dependencies (after go mod tidy)
├── doc.go                # Package documentation
├── types.go              # Issue, LinkedBranch, Label types
├── errors.go             # Error type definitions
└── mocks/
    └── .gitkeep          # Placeholder (mocks generated in later task)
```

### 2. Create go.mod

Module path: `github.com/jmgilman/sow/libs/git`

Dependencies to include:
- `github.com/jmgilman/sow/libs/exec` - For command execution abstraction

Use Go version 1.25.3 (matching libs/exec).

### 3. Create doc.go

Package documentation following the pattern from `libs/exec/doc.go`:
- Explain the package purpose: Git and GitHub operations for sow
- Document the ports and adapters design pattern
- Provide usage examples for Git operations and GitHubClient
- Reference the mocks subpackage for testing
- Use complete godoc formatting

### 4. Create types.go

Move and adapt these types from `cli/internal/sow/github_cli.go`:

```go
// Issue represents a GitHub issue.
type Issue struct {
    Number int    `json:"number"`
    Title  string `json:"title"`
    Body   string `json:"body"`
    State  string `json:"state"`
    URL    string `json:"url"`
    Labels []Label `json:"labels"`
}

// Label represents a GitHub label.
type Label struct {
    Name string `json:"name"`
}

// HasLabel checks if an issue has a specific label.
func (i *Issue) HasLabel(label string) bool

// LinkedBranch represents a branch linked to an issue.
type LinkedBranch struct {
    Name string
    URL  string
}
```

Note: The Labels field is changed from inline `[]struct{ Name string }` to a proper named type `[]Label` for better usability.

### 5. Create errors.go

Define error types following the pattern from the existing code:

```go
// ErrGHNotInstalled is returned when the gh CLI is not found in PATH.
type ErrGHNotInstalled struct{}
func (e ErrGHNotInstalled) Error() string

// ErrGHNotAuthenticated is returned when gh CLI is installed but not authenticated.
type ErrGHNotAuthenticated struct{}
func (e ErrGHNotAuthenticated) Error() string

// ErrGHCommand is returned when a gh command fails.
type ErrGHCommand struct {
    Command string
    Stderr  string
    Err     error
}
func (e ErrGHCommand) Error() string
func (e ErrGHCommand) Unwrap() error

// ErrNotGitRepository is returned when the path is not a git repository.
type ErrNotGitRepository struct {
    Path string
}
func (e ErrNotGitRepository) Error() string

// ErrBranchExists is returned when attempting to create a branch that already exists.
type ErrBranchExists struct {
    Branch string
}
func (e ErrBranchExists) Error() string
```

Error messages should be clear and actionable. Use `%w` wrapping pattern in `Unwrap()`.

## Acceptance Criteria

1. **Module compiles**: `cd libs/git && go build ./...` succeeds
2. **Module initializes**: `go mod tidy` completes without errors
3. **Types are exported**: Issue, Label, LinkedBranch are accessible from other packages
4. **HasLabel method works**: Test with a simple example
5. **Error types implement error interface**: All error types have Error() method
6. **ErrGHCommand supports unwrapping**: Has Unwrap() method for error chain
7. **Documentation present**: doc.go provides clear package overview
8. **Linting passes**: `golangci-lint run` reports no issues

### Test Requirements (TDD)

Write tests in `types_test.go` for:
- `Issue.HasLabel()` - returns true when label exists
- `Issue.HasLabel()` - returns false when label doesn't exist
- `Issue.HasLabel()` - handles empty labels slice

Write tests in `errors_test.go` for:
- Each error type's Error() method returns expected message format
- `ErrGHCommand.Unwrap()` returns the wrapped error
- Error types can be used with `errors.Is()` and `errors.As()`

## Technical Details

### Module Declaration

```go
// go.mod
module github.com/jmgilman/sow/libs/git

go 1.25.3

require github.com/jmgilman/sow/libs/exec v0.0.0
```

Note: The version will be resolved by go mod tidy using replace directives or workspace mode.

### Import Structure

```go
// imports in this module
import (
    "github.com/jmgilman/sow/libs/exec"
)
```

### File Organization (per STYLE.md)

Each file should organize code in this order:
1. Package declaration and imports
2. Constants
3. Type declarations
4. Functions/methods

## Relevant Inputs

- `libs/exec/go.mod` - Reference for module structure and Go version
- `libs/exec/doc.go` - Reference for documentation pattern
- `libs/exec/executor.go` - Reference for interface design pattern
- `libs/config/repo.go` - Reference for core.FS pattern (contrast: git uses real paths)
- `libs/config/repo_test.go` - Reference for billy.MemoryFS testing (contrast: git uses t.TempDir())
- `cli/internal/sow/github_cli.go:84-99` - Source of Issue and LinkedBranch types
- `cli/internal/sow/github_cli.go:50-79` - Source of error types
- `.standards/STYLE.md` - Coding standards to follow
- `.standards/TESTING.md` - Testing standards to follow
- `.sow/project/context/issue-117.md` - Full requirements context

## Examples

### Using Issue.HasLabel

```go
issue := &git.Issue{
    Number: 123,
    Title:  "My Issue",
    Labels: []git.Label{{Name: "bug"}, {Name: "sow"}},
}

if issue.HasLabel("sow") {
    fmt.Println("This is a sow issue")
}
```

### Using Error Types

```go
func CheckGH(exec exec.Executor) error {
    if !exec.Exists() {
        return git.ErrGHNotInstalled{}
    }
    return nil
}

// Checking errors
err := CheckGH(executor)
var notInstalled git.ErrGHNotInstalled
if errors.As(err, &notInstalled) {
    fmt.Println("Please install gh: https://cli.github.com/")
}
```

## Dependencies

- No other tasks need to complete first (this is the foundation)
- libs/exec must exist (it already does in the repository)

## Constraints

- Do NOT import from `cli/internal/sow/` - this module must be standalone
- Do NOT add any git operations yet - those come in task 020
- Do NOT add GitHubClient interface yet - that comes in task 030
- Keep the scope minimal - only types and errors needed by other tasks
- Follow STYLE.md exactly - short receiver names, proper error wrapping, etc.

## Design Note: Filesystem Abstraction

Unlike `libs/config` which uses `github.com/jmgilman/go/fs/core.FS` for filesystem abstraction and `billy.MemoryFS` for testing, the `libs/git` module uses **real filesystem paths**. This is intentional because:

1. The go-git library (`github.com/jmgilman/go/git`) requires real paths
2. Git CLI commands need real directories
3. Worktree operations create actual files on disk

Tests in this module use `t.TempDir()` to create real temporary git repositories.
