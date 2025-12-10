# Task 030: Implement GitHubClient Interface and CLI Adapter

## Context

This task is part of work unit 004: creating a new `libs/git` Go module. The previous tasks created the module foundation (010) and Git operations (020). This task implements the GitHubClient interface (port) and the GitHubCLI adapter that wraps the `gh` CLI tool.

**Ports and Adapters Pattern**:
- `GitHubClient` is the **port** (interface) defining GitHub operations
- `GitHubCLI` is the **adapter** (implementation) using the `gh` CLI
- This design allows future implementations (e.g., REST API client) without changing consumers

## Requirements

### 1. Create client.go (Port/Interface)

Define the GitHubClient interface:

```go
// GitHubClient defines operations for GitHub issue, PR, and branch management.
//
// This interface enables multiple client implementations:
//   - GitHubCLI: Wraps the gh CLI tool for local development
//   - Future: GitHubAPI for web VMs or CI/CD environments
//
// Use NewGitHubClient() factory for automatic environment detection.
type GitHubClient interface {
    // CheckAvailability verifies that GitHub access is available and ready.
    // Returns ErrGHNotInstalled or ErrGHNotAuthenticated on failure.
    CheckAvailability() error

    // ListIssues returns issues matching the specified label and state.
    // state can be "open", "closed", or "all". Returns up to 1000 issues.
    ListIssues(label, state string) ([]Issue, error)

    // GetIssue retrieves a single issue by number.
    GetIssue(number int) (*Issue, error)

    // CreateIssue creates a new GitHub issue.
    CreateIssue(title, body string, labels []string) (*Issue, error)

    // GetLinkedBranches returns branches linked to an issue.
    GetLinkedBranches(number int) ([]LinkedBranch, error)

    // CreateLinkedBranch creates a branch linked to an issue.
    // branchName can be empty for auto-generated name.
    CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)

    // CreatePullRequest creates a PR, optionally as draft.
    // Returns PR number, URL, and error.
    CreatePullRequest(title, body string, draft bool) (number int, url string, err error)

    // UpdatePullRequest updates an existing PR's title and body.
    UpdatePullRequest(number int, title, body string) error

    // MarkPullRequestReady converts a draft PR to ready for review.
    MarkPullRequestReady(number int) error
}
```

### 2. Create client_cli.go (Adapter Implementation)

Implement the GitHubCLI adapter:

```go
// GitHubCLI implements GitHubClient using the gh CLI tool.
//
// All operations require the gh CLI to be installed and authenticated.
// The client accepts an Executor interface for testability.
type GitHubCLI struct {
    exec exec.Executor
}

// NewGitHubCLI creates a new GitHub CLI client with the given executor.
//
// Example:
//     ghExec := exec.NewLocalExecutor("gh")
//     github := git.NewGitHubCLI(ghExec)
func NewGitHubCLI(executor exec.Executor) *GitHubCLI
```

Implement all GitHubClient methods following the existing implementation in `cli/internal/sow/github_cli.go`. Key behaviors:

- `CheckAvailability()`: Check gh exists and is authenticated
- `ListIssues()`: Use `gh issue list --label X --state X --json ... --limit 1000`
- `GetIssue()`: Use `gh issue view N --json ...`
- `CreateIssue()`: Use `gh issue create --title X --body X --label X...`
- `GetLinkedBranches()`: Use `gh issue develop --list N`
- `CreateLinkedBranch()`: Use `gh issue develop N [--name X] [--checkout]`
- `CreatePullRequest()`: Use `gh pr create --title X --body X [--draft]`
- `UpdatePullRequest()`: Use `gh pr edit N --title X --body X`
- `MarkPullRequestReady()`: Use `gh pr ready N`

### 3. Create factory.go

Implement the factory function:

```go
// NewGitHubClient creates a GitHub client with automatic environment detection.
//
// Currently returns GitHubCLI. Future: could return GitHubAPI if GITHUB_TOKEN is set.
func NewGitHubClient() (GitHubClient, error)
```

### 4. Helper Functions

Include the `toKebabCase` helper function for branch name generation (used in CreateLinkedBranch fallback).

### 5. Add go:generate Directive

In client.go, add:
```go
//go:generate go run github.com/matryer/moq@latest -out mocks/client.go -pkg mocks . GitHubClient
```

## Acceptance Criteria

1. **GitHubClient interface defined**: All methods documented and exported
2. **GitHubCLI implements interface**: Compile-time check passes
3. **CheckAvailability works**: Returns appropriate errors for missing/unauthenticated gh
4. **ListIssues parses JSON**: Correctly deserializes gh output
5. **GetIssue retrieves single issue**: Includes all fields
6. **CreateIssue returns issue**: Parses issue number from URL
7. **GetLinkedBranches parses output**: Handles empty result and multiple branches
8. **CreateLinkedBranch creates branches**: With and without custom name
9. **PR operations work**: Create, update, and mark ready
10. **Factory creates client**: Returns working GitHubClient
11. **All tests pass**: Comprehensive mock-based testing

### Test Requirements (TDD)

Write tests in `client_cli_test.go`:

**CheckAvailability tests:**
- Returns nil when gh is installed and authenticated
- Returns ErrGHNotInstalled when gh not found
- Returns ErrGHNotAuthenticated when auth fails

**ListIssues tests:**
- Parses JSON response correctly
- Handles empty result
- Passes correct arguments to gh

**GetIssue tests:**
- Parses single issue JSON
- Returns error for non-existent issue

**CreateIssue tests:**
- Parses URL to extract issue number
- Handles labels correctly

**GetLinkedBranches tests:**
- Parses tab-separated output
- Returns empty slice when no branches linked
- Handles "no linked branches" error gracefully

**CreateLinkedBranch tests:**
- Passes correct arguments for custom name
- Passes correct arguments for auto-generated name
- Respects checkout flag

**CreatePullRequest tests:**
- Includes --draft flag when draft=true
- Omits --draft flag when draft=false
- Parses PR number from URL

**UpdatePullRequest tests:**
- Passes correct arguments
- Returns error on failure

**MarkPullRequestReady tests:**
- Passes correct arguments
- Returns error on failure

Use `libs/exec/mocks.ExecutorMock` for all tests. Follow the testing pattern from `cli/internal/sow/github_cli_test.go`.

## Technical Details

### Import Structure

```go
import (
    "encoding/json"
    "fmt"
    "strings"

    "github.com/jmgilman/sow/libs/exec"
)
```

### JSON Parsing

Use standard library `encoding/json` for parsing gh CLI JSON output.

### Error Wrapping

All gh command failures should return `ErrGHCommand` with:
- `Command`: The gh subcommand that failed (e.g., "issue list")
- `Stderr`: The stderr output from gh
- `Err`: The underlying error

### Compile-Time Interface Check

Add at the end of client_cli.go:
```go
var _ GitHubClient = (*GitHubCLI)(nil)
```

## Relevant Inputs

- `cli/internal/sow/github_client.go` - Source interface definition
- `cli/internal/sow/github_cli.go` - Source implementation (CRITICAL - read this!)
- `cli/internal/sow/github_cli_test.go` - Test patterns to follow
- `cli/internal/sow/github_factory.go` - Source factory implementation
- `libs/exec/executor.go` - Executor interface
- `libs/exec/mocks/executor.go` - Mock for testing
- `libs/git/types.go` - Issue, LinkedBranch types (from task 010)
- `libs/git/errors.go` - Error types (from task 010)
- `.standards/STYLE.md` - Coding standards
- `.standards/TESTING.md` - Testing standards

## Examples

### Using GitHubClient

```go
import (
    "github.com/jmgilman/sow/libs/exec"
    "github.com/jmgilman/sow/libs/git"
)

// Create client with factory
client, err := git.NewGitHubClient()
if err != nil {
    return err
}

// Or create explicitly with executor
ghExec := exec.NewLocalExecutor("gh")
client := git.NewGitHubCLI(ghExec)

// Check availability first
if err := client.CheckAvailability(); err != nil {
    var notInstalled git.ErrGHNotInstalled
    if errors.As(err, &notInstalled) {
        return fmt.Errorf("please install gh: https://cli.github.com/")
    }
    return err
}

// Use the client
issues, err := client.ListIssues("sow", "open")
```

### Testing with Mocks

```go
import (
    "testing"
    "github.com/jmgilman/sow/libs/exec/mocks"
    "github.com/jmgilman/sow/libs/git"
)

func TestListIssues(t *testing.T) {
    mock := &mocks.ExecutorMock{
        ExistsFunc: func() bool { return true },
        RunSilentFunc: func(args ...string) error { return nil },
        RunFunc: func(args ...string) (string, string, error) {
            return `[{"number":1,"title":"Test"}]`, "", nil
        },
    }

    client := git.NewGitHubCLI(mock)
    issues, err := client.ListIssues("sow", "open")

    require.NoError(t, err)
    assert.Len(t, issues, 1)
    assert.Equal(t, "Test", issues[0].Title)
}
```

## Dependencies

- Task 010 (module foundation) must complete first - provides types.go and errors.go
- `libs/exec` module must exist (it does)

## Constraints

- Do NOT import from `cli/internal/sow/` - this module must be standalone
- Do NOT use real gh CLI in tests - use mocks exclusively
- Do NOT implement GitHubAPI yet - only GitHubCLI
- The mock generation will happen in task 050
- Tests must not make network calls
- Follow existing implementation behavior exactly (for compatibility)
