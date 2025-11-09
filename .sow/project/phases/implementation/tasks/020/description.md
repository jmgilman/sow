# Task 020: Rename Implementation to GitHubCLI

## Context

This task is part of a refactoring project to extract a GitHubClient interface from the existing GitHub CLI implementation. Task 010 created the interface definition. Now we need to rename the existing implementation to clarify it's the CLI-specific implementation, distinguishing it from the future API implementation.

**Project Goal**: Extract GitHubClient interface from existing GitHub CLI implementation to enable dual client support (CLI + API) while maintaining backward compatibility.

**Previous Task**: Task 010 defined the GitHubClient interface in `github_client.go`.

**This Task's Role**: Rename the existing GitHub struct to GitHubCLI, update all references, and make it implement the new interface. This clarifies that this is one specific implementation (CLI-based) rather than the generic client.

## Requirements

**File Operations:**
1. Rename `cli/internal/sow/github.go` to `cli/internal/sow/github_cli.go`
2. Rename `cli/internal/sow/github_test.go` to `cli/internal/sow/github_cli_test.go`

**Code Changes in github_cli.go:**
1. Rename struct `GitHub` to `GitHubCLI`
2. Rename constructor `NewGitHub()` to `NewGitHubCLI()`
3. Add deprecated `NewGitHub()` wrapper for backward compatibility
4. Update all method receivers from `(g *GitHub)` to `(g *GitHubCLI)`
5. Add interface compliance check: `var _ GitHubClient = (*GitHubCLI)(nil)`
6. Add CheckAvailability() method that delegates to existing Ensure()
7. Update struct godoc to clarify it's the CLI implementation

**Code Changes in github_cli_test.go:**
1. Change all `sow.GitHub` references to `sow.GitHubCLI`
2. Change all `sow.NewGitHub()` calls to `sow.NewGitHubCLI()`
3. Update test function names if they reference the struct name
4. No changes to test logic - tests should pass with just name changes

**Backward Compatibility Wrapper:**
```go
// NewGitHub creates a GitHub CLI client.
// Deprecated: Use NewGitHubCLI() for explicit CLI client, or NewGitHubClient() for auto-detection.
func NewGitHub(executor exec.Executor) *GitHubCLI {
    return NewGitHubCLI(executor)
}
```

**CheckAvailability Implementation:**
```go
// CheckAvailability implements GitHubClient.
// It delegates to Ensure() which checks both installation and authentication.
func (g *GitHubCLI) CheckAvailability() error {
    return g.Ensure()
}
```

## Acceptance Criteria

- [ ] File renamed: `github.go` -> `github_cli.go`
- [ ] File renamed: `github_test.go` -> `github_cli_test.go`
- [ ] Struct renamed: `GitHub` -> `GitHubCLI` in github_cli.go
- [ ] Constructor renamed: `NewGitHub()` -> `NewGitHubCLI()` in github_cli.go
- [ ] Deprecated `NewGitHub()` wrapper exists and calls `NewGitHubCLI()`
- [ ] All method receivers updated to `*GitHubCLI`
- [ ] CheckAvailability() method implemented delegating to Ensure()
- [ ] Interface compliance check added: `var _ GitHubClient = (*GitHubCLI)(nil)`
- [ ] Struct godoc updated to mention it's the CLI implementation
- [ ] Test file references updated to use `GitHubCLI` and `NewGitHubCLI()`
- [ ] All existing tests pass: `go test ./cli/internal/sow`
- [ ] Package compiles: `go build ./cli/internal/sow`
- [ ] No changes to test logic (only struct/function name changes)

## Technical Details

**Renaming Pattern:**

Every occurrence of these patterns needs updating:
- Struct declaration: `type GitHub struct` -> `type GitHubCLI struct`
- Constructor: `func NewGitHub(` -> `func NewGitHubCLI(`
- Method receivers: `func (g *GitHub)` -> `func (g *GitHubCLI)`
- Variable types in tests: `*sow.GitHub` -> `*sow.GitHubCLI`
- Constructor calls in tests: `sow.NewGitHub(` -> `sow.NewGitHubCLI(`

**Test File Updates:**

The test file `github_cli_test.go` needs these changes:
- Line 37: `github := sow.NewGitHub(mock)` -> `github := sow.NewGitHubCLI(mock)`
- Line 62: `github := sow.NewGitHub(mock)` -> `github := sow.NewGitHubCLI(mock)`
- Line 86: `github := sow.NewGitHub(mock)` -> `github := sow.NewGitHubCLI(mock)`

**Interface Compliance Check:**

Add this at the bottom of github_cli.go (following pattern from executor.go:166):
```go
// Compile-time check that GitHubCLI implements GitHubClient.
var _ GitHubClient = (*GitHubCLI)(nil)
```

This ensures at compile time that GitHubCLI correctly implements all interface methods.

**Updated Struct Documentation:**

```go
// GitHubCLI implements GitHubClient using the gh CLI tool.
//
// All operations require the GitHub CLI (gh) to be installed and authenticated.
// The client accepts an Executor interface, making it easy to mock in tests.
//
// For auto-detection between CLI and API clients, use NewGitHubClient() factory.
type GitHubCLI struct {
    gh exec.Executor
}
```

**CheckAvailability Implementation Details:**

The new CheckAvailability() method should be simple - it delegates to the existing Ensure() method:

```go
// CheckAvailability implements GitHubClient.
// For CLI client, this checks that gh is installed and authenticated.
func (g *GitHubCLI) CheckAvailability() error {
    return g.Ensure()
}
```

Keep the existing CheckInstalled(), CheckAuthenticated(), and Ensure() methods unchanged - they're still useful for CLI-specific error handling.

**Testing Approach:**

1. Write the code following TDD:
   - Run existing tests first to ensure they pass: `go test ./cli/internal/sow`
   - Make the rename changes
   - Run tests again - they should all pass with just name changes
   - Add compile-time interface check - compilation failure indicates missing interface methods

2. No new test logic needed - existing tests validate behavior
3. The interface compliance check validates implementation correctness

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/.sow/project/context/issue-90.md` - Complete project requirements (Step 2: lines 194-237)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github.go` - Current implementation to rename (entire file)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_test.go` - Tests to update (entire file)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_client.go` - Interface definition (created in Task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/exec/executor.go` - Interface compliance pattern (line 166)

## Examples

**Before (github.go:15-37):**
```go
type GitHub struct {
    gh exec.Executor
}

func NewGitHub(executor exec.Executor) *GitHub {
    return &GitHub{
        gh: executor,
    }
}
```

**After (github_cli.go:15-47):**
```go
type GitHubCLI struct {
    gh exec.Executor
}

func NewGitHubCLI(executor exec.Executor) *GitHubCLI {
    return &GitHubCLI{
        gh: executor,
    }
}

// NewGitHub creates a GitHub CLI client.
// Deprecated: Use NewGitHubCLI() for explicit CLI client, or NewGitHubClient() for auto-detection.
func NewGitHub(executor exec.Executor) *GitHubCLI {
    return NewGitHubCLI(executor)
}

// CheckAvailability implements GitHubClient.
func (g *GitHubCLI) CheckAvailability() error {
    return g.Ensure()
}
```

**Test Update Example:**

Before (github_test.go:37):
```go
github := sow.NewGitHub(mock)
```

After (github_cli_test.go:37):
```go
github := sow.NewGitHubCLI(mock)
```

## Dependencies

**Depends On:**
- Task 010 (Define GitHubClient Interface) - Must be completed first

**Reason**: The interface compliance check requires the GitHubClient interface to exist.

## Constraints

**DO NOT:**
- Change any method signatures (except receiver type name)
- Change test logic or assertions
- Remove the existing CheckInstalled(), CheckAuthenticated(), or Ensure() methods
- Break backward compatibility (NewGitHub must still exist)
- Modify any files outside cli/internal/sow/ in this task

**DO:**
- Keep all error types unchanged (ErrGHNotInstalled, etc.)
- Keep all existing methods with same behavior
- Maintain same test coverage
- Keep Issue and LinkedBranch types in this file
- Add deprecation notice to old constructor

**Compatibility:**
- Old code using `sow.NewGitHub()` continues to work via deprecated wrapper
- All method behavior remains identical
- Error handling unchanged
- No breaking changes

**Testing:**
- All existing tests must pass without modification to test logic
- Only struct/function name changes in tests
- Package must compile without errors
- Interface compliance must verify at compile time
