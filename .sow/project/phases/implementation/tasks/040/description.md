# Task 040: Create GitHub Factory with Auto-Detection

## Context

This task is part of a refactoring project to extract a GitHubClient interface from the existing GitHub CLI implementation. Previous tasks created the interface, renamed the implementation, and added new methods. Now we need a factory function that automatically detects the environment and returns the appropriate client.

**Project Goal**: Extract GitHubClient interface from existing GitHub CLI implementation to enable dual client support (CLI + API) while maintaining backward compatibility.

**Previous Tasks**:
- Task 010: Created GitHubClient interface
- Task 020: Renamed GitHub to GitHubCLI
- Task 030: Implemented new methods

**This Task's Role**: Create a factory function that detects whether to use CLI or API client based on environment variables. For now, only the CLI client exists, but the factory establishes the pattern for future API client integration.

**Why This Matters**: Claude Code web VMs will set GITHUB_TOKEN environment variable, triggering API mode. Local development uses gh CLI. The factory makes this transparent to callers.

## Requirements

Create a new file `cli/internal/sow/github_factory.go` with:

**1. Factory Function: NewGitHubClient()**
- Returns (GitHubClient, error)
- Checks for GITHUB_TOKEN environment variable
- If GITHUB_TOKEN is set: return error (API client not yet implemented)
- If GITHUB_TOKEN not set: return GitHubCLI client
- Include helpful error messages

**2. Documentation**
- Explain auto-detection logic
- Document environment variable behavior
- Explain when to use factory vs specific constructors
- Note that API client is future work (work unit 004)

**Factory Signature:**
```go
func NewGitHubClient() (GitHubClient, error)
```

**Detection Logic:**
```
IF GITHUB_TOKEN env var is set:
    -> Return error: "API client not yet implemented; unset GITHUB_TOKEN to use gh CLI"
ELSE:
    -> Return NewGitHubCLI(exec.NewLocal("gh"))
```

**Error Message for API Mode:**
The error should be clear and actionable:
```go
return nil, fmt.Errorf("GitHub API client not yet implemented (work unit 004); unset GITHUB_TOKEN to use gh CLI")
```

## Acceptance Criteria

- [ ] File `cli/internal/sow/github_factory.go` created
- [ ] NewGitHubClient() function implemented with auto-detection logic
- [ ] Function checks GITHUB_TOKEN environment variable
- [ ] Returns error with helpful message if GITHUB_TOKEN is set
- [ ] Returns GitHubCLI client if GITHUB_TOKEN is not set
- [ ] Function has comprehensive godoc explaining behavior
- [ ] Unit tests written for both scenarios (following TDD)
- [ ] Test: GITHUB_TOKEN set returns error
- [ ] Test: GITHUB_TOKEN unset returns GitHubCLI client
- [ ] Test: Verify error message is helpful
- [ ] Test: Verify returned client type is *GitHubCLI
- [ ] All tests pass: `go test ./cli/internal/sow`
- [ ] Package compiles: `go build ./cli/internal/sow`

## Technical Details

**File Structure:**

```go
package sow

import (
    "fmt"
    "os"

    "github.com/jmgilman/sow/cli/internal/exec"
)

// NewGitHubClient creates a GitHub client with automatic environment detection.
//
// Detection logic:
// - If GITHUB_TOKEN env var is set: API client (work unit 004, not yet implemented)
// - Otherwise: CLI client backed by gh command
//
// For web VMs, GITHUB_TOKEN will be set by Claude Code.
// For local dev, gh CLI is expected to be installed and authenticated.
//
// Use this factory for automatic client selection. Use NewGitHubCLI() directly
// if you specifically need the CLI client regardless of environment.
//
// Returns error if GITHUB_TOKEN is set but API client not yet implemented.
func NewGitHubClient() (GitHubClient, error) {
    if token := os.Getenv("GITHUB_TOKEN"); token != "" {
        // Future: Return GitHubAPI client (work unit 004)
        return nil, fmt.Errorf("GitHub API client not yet implemented (work unit 004); unset GITHUB_TOKEN to use gh CLI")
    }

    // Default to CLI client
    return NewGitHubCLI(exec.NewLocal("gh")), nil
}
```

**Testing Approach (TDD):**

Create `cli/internal/sow/github_factory_test.go` with tests:

```go
package sow_test

import (
    "os"
    "testing"

    "github.com/jmgilman/sow/cli/internal/sow"
)

func TestNewGitHubClient_WithGitHubToken_ReturnsError(t *testing.T) {
    // Set GITHUB_TOKEN for this test
    os.Setenv("GITHUB_TOKEN", "test-token")
    defer os.Unsetenv("GITHUB_TOKEN")

    client, err := sow.NewGitHubClient()

    // Should error because API client not yet implemented
    if err == nil {
        t.Fatal("expected error for unimplemented API client, got nil")
    }

    if client != nil {
        t.Errorf("expected nil client, got %T", client)
    }

    // Verify error message is helpful
    expectedSubstr := "API client not yet implemented"
    if !strings.Contains(err.Error(), expectedSubstr) {
        t.Errorf("error message should mention API not implemented, got: %s", err.Error())
    }
}

func TestNewGitHubClient_WithoutToken_ReturnsCLIClient(t *testing.T) {
    // Ensure GITHUB_TOKEN is not set
    os.Unsetenv("GITHUB_TOKEN")

    client, err := sow.NewGitHubClient()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if client == nil {
        t.Fatal("expected non-nil client")
    }

    // Verify type is *GitHubCLI
    if _, ok := client.(*sow.GitHubCLI); !ok {
        t.Errorf("expected *sow.GitHubCLI, got %T", client)
    }
}

func TestNewGitHubClient_EmptyToken_ReturnsCLIClient(t *testing.T) {
    // Set empty token (should be treated as not set)
    os.Setenv("GITHUB_TOKEN", "")
    defer os.Unsetenv("GITHUB_TOKEN")

    client, err := sow.NewGitHubClient()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if client == nil {
        t.Fatal("expected non-nil client")
    }

    // Empty token should trigger CLI mode
    if _, ok := client.(*sow.GitHubCLI); !ok {
        t.Errorf("expected *sow.GitHubCLI for empty token, got %T", client)
    }
}
```

**Design Rationale:**

1. **Return interface, not concrete type**: Allows future API implementation without changing callers
2. **Check token presence, not validity**: Actual token validation happens in API client (future)
3. **Helpful error message**: Tells users how to work around missing API client
4. **Default to CLI**: Matches current behavior, no breaking changes

**Future Extension (Work Unit 004):**

When API client is implemented, the factory becomes:
```go
if token := os.Getenv("GITHUB_TOKEN"); token != "" {
    // Extract owner/repo from git remote
    owner, repo, err := extractRepoInfo()
    if err != nil {
        return nil, fmt.Errorf("failed to extract repo info: %w", err)
    }
    return NewGitHubAPI(token, owner, repo), nil
}
```

But that's future work - not part of this task.

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/.sow/project/context/issue-90.md` - Complete project requirements (Step 3: lines 239-274; Factory logic: lines 129-133)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_cli.go` - GitHubCLI constructor to call (will exist after Tasks 020-030)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_client.go` - Interface to return (created in Task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/exec/executor.go` - LocalExecutor to use (lines 47-66)

## Examples

**Factory Usage in Production Code:**

```go
// Automatic detection
client, err := sow.NewGitHubClient()
if err != nil {
    return err
}

issues, err := client.ListIssues("sow", "open")
```

**Explicit CLI Client:**

```go
// Force CLI client regardless of environment
client := sow.NewGitHubCLI(exec.NewLocal("gh"))
issues, err := client.ListIssues("sow", "open")
```

**Test with Mock:**

```go
// Tests can still use mock directly
mock := &sow.MockGitHub{
    ListIssuesFunc: func(label, state string) ([]sow.Issue, error) {
        return []sow.Issue{{Number: 1}}, nil
    },
}
issues, err := mock.ListIssues("sow", "open")
```

**Environment Behavior:**

```bash
# Local development (no token)
$ sow issue list
# Uses gh CLI automatically

# Web VM (token set by Claude Code)
$ export GITHUB_TOKEN="ghp_..."
$ sow issue list
# Error: API client not yet implemented
# (Will work after work unit 004)
```

## Dependencies

**Depends On:**
- Task 010 (Define GitHubClient Interface) - Must exist to return
- Task 020 (Rename Implementation to GitHubCLI) - Constructor must exist to call

**Depended On By:**
- Task 060 (Update Wizard Interface) - Wizard could optionally use factory

**Reason**: Factory needs interface definition and CLI constructor to work.

## Constraints

**DO NOT:**
- Implement API client in this task (that's work unit 004)
- Make factory do anything beyond environment detection
- Add complexity beyond GITHUB_TOKEN check
- Parse git remotes or validate tokens (API client's job)
- Change behavior of existing code

**DO:**
- Keep factory simple and focused
- Return clear, actionable error messages
- Write comprehensive tests for both paths
- Document detection logic clearly
- Follow TDD approach (write tests first)

**Testing Requirements:**
- Write tests FIRST (TDD methodology)
- Test with GITHUB_TOKEN set (should error)
- Test with GITHUB_TOKEN unset (should return CLI client)
- Test with empty GITHUB_TOKEN (edge case)
- Verify error messages are helpful
- Verify return types are correct

**Future Compatibility:**
- Factory signature won't change when API client is added
- Error return allows for API client integration
- Interface return allows transparent switching

**Error Message Requirements:**
- Must mention API client not implemented
- Should mention work unit 004 for context
- Must tell user how to work around (unset GITHUB_TOKEN)
- Should be one clear sentence
