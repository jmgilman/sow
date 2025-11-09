# Task 050: Create GitHub Mock Implementation

## Context

This task is part of a refactoring project to extract a GitHubClient interface from the existing GitHub CLI implementation. Previous tasks created the interface, implementations, and factory. Now we need a mock implementation for testing code that depends on GitHubClient.

**Project Goal**: Extract GitHubClient interface from existing GitHub CLI implementation to enable dual client support (CLI + API) while maintaining backward compatibility.

**Previous Tasks**:
- Task 010: Created GitHubClient interface
- Task 020: Renamed GitHub to GitHubCLI
- Task 030: Implemented new methods
- Task 040: Created factory function

**This Task's Role**: Create a mock implementation of GitHubClient that consumers can use in their tests. This is different from the MockExecutor pattern - this is a mock at the GitHubClient level, useful for testing higher-level code (like commands or wizards) that use GitHub operations.

**Why This Matters**: The project wizard and commands depend on GitHubClient. They need a way to test their logic without calling real GitHub operations or mocking executor internals.

## Requirements

Create a new file `cli/internal/sow/github_mock.go` with:

**1. MockGitHub Struct**
- Implements GitHubClient interface
- Has function fields for each interface method
- Each field is a func matching the interface method signature
- Fields are optional (nil-safe with sensible defaults)

**2. Method Implementations**
- Each method checks if corresponding func field is set
- If set: call the func and return its result
- If not set: return sensible default (empty slice, nil, success)
- Follow pattern from MockExecutor in exec/mock.go

**3. Documentation**
- Explain when to use MockGitHub vs MockExecutor
- Show usage examples in godoc
- Document default behavior when funcs are nil

**4. Interface Compliance**
- Add compile-time check: `var _ GitHubClient = (*MockGitHub)(nil)`

**Mock Structure:**
```go
type MockGitHub struct {
    CheckAvailabilityFunc     func() error
    ListIssuesFunc            func(label, state string) ([]Issue, error)
    GetIssueFunc              func(number int) (*Issue, error)
    CreateIssueFunc           func(title, body string, labels []string) (*Issue, error)
    GetLinkedBranchesFunc     func(number int) ([]LinkedBranch, error)
    CreateLinkedBranchFunc    func(issueNumber int, branchName string, checkout bool) (string, error)
    CreatePullRequestFunc     func(title, body string, draft bool) (number int, url string, error)
    UpdatePullRequestFunc     func(number int, title, body string) error
    MarkPullRequestReadyFunc  func(number int) error
}
```

## Acceptance Criteria

- [ ] File `cli/internal/sow/github_mock.go` created
- [ ] MockGitHub struct defined with all interface method func fields
- [ ] All 9 interface methods implemented
- [ ] Each method calls corresponding func field if set
- [ ] Each method returns sensible default if func field is nil
- [ ] Interface compliance check added: `var _ GitHubClient = (*MockGitHub)(nil)`
- [ ] Comprehensive godoc at struct level with usage examples
- [ ] Godoc explains when to use MockGitHub vs MockExecutor
- [ ] Unit tests written demonstrating mock usage (following TDD)
- [ ] Test: Mock with custom funcs returns custom values
- [ ] Test: Mock with nil funcs returns sensible defaults
- [ ] All tests pass: `go test ./cli/internal/sow`
- [ ] Package compiles: `go build ./cli/internal/sow`

## Technical Details

**Implementation Pattern (from exec/mock.go):**

```go
// MockGitHub is a mock implementation of GitHubClient for testing.
//
// Usage in tests:
//   mock := &sow.MockGitHub{
//       CheckAvailabilityFunc: func() error { return nil },
//       GetIssueFunc: func(n int) (*sow.Issue, error) {
//           return &sow.Issue{Number: n, Title: "Test"}, nil
//       },
//   }
//
// Methods with nil func fields return sensible defaults:
//   - CheckAvailability: nil (success)
//   - ListIssues: empty slice
//   - GetIssue: nil issue
//   - Error-returning methods: nil (success)
//
// Use MockGitHub when testing code that depends on GitHubClient interface.
// Use MockExecutor (from exec package) when testing GitHubCLI implementation.
type MockGitHub struct {
    CheckAvailabilityFunc     func() error
    ListIssuesFunc            func(label, state string) ([]Issue, error)
    GetIssueFunc              func(number int) (*Issue, error)
    CreateIssueFunc           func(title, body string, labels []string) (*Issue, error)
    GetLinkedBranchesFunc     func(number int) ([]LinkedBranch, error)
    CreateLinkedBranchFunc    func(issueNumber int, branchName string, checkout bool) (string, error)
    CreatePullRequestFunc     func(title, body string, draft bool) (number int, url string, error)
    UpdatePullRequestFunc     func(number int, title, body string) error
    MarkPullRequestReadyFunc  func(number int) error
}

// CheckAvailability calls the mock function if set, otherwise returns nil.
func (m *MockGitHub) CheckAvailability() error {
    if m.CheckAvailabilityFunc != nil {
        return m.CheckAvailabilityFunc()
    }
    return nil
}

// ListIssues calls the mock function if set, otherwise returns empty slice.
func (m *MockGitHub) ListIssues(label, state string) ([]Issue, error) {
    if m.ListIssuesFunc != nil {
        return m.ListIssuesFunc(label, state)
    }
    return []Issue{}, nil
}

// GetIssue calls the mock function if set, otherwise returns nil.
func (m *MockGitHub) GetIssue(number int) (*Issue, error) {
    if m.GetIssueFunc != nil {
        return m.GetIssueFunc(number)
    }
    return nil, nil
}

// ... continue pattern for all interface methods
```

**Default Return Values:**

Follow these patterns for nil func fields:
- `error` return: `return nil` (success)
- `[]Issue` return: `return []Issue{}, nil` (empty slice, no error)
- `[]LinkedBranch` return: `return []LinkedBranch{}, nil` (empty slice)
- `*Issue` return: `return nil, nil` (no issue, no error)
- `string` return: `return "", nil` (empty string, no error)
- `(int, string, error)` return: `return 0, "", nil` (zeros, no error)

These defaults let tests pass without explicit mocking of unused methods.

**Testing Approach (TDD):**

Create `cli/internal/sow/github_mock_test.go`:

```go
package sow_test

import (
    "errors"
    "testing"

    "github.com/jmgilman/sow/cli/internal/sow"
)

func TestMockGitHub_ImplementsInterface(t *testing.T) {
    var _ sow.GitHubClient = (*sow.MockGitHub)(nil)
}

func TestMockGitHub_WithCustomFunc_ReturnsCustomValue(t *testing.T) {
    mock := &sow.MockGitHub{
        GetIssueFunc: func(number int) (*sow.Issue, error) {
            return &sow.Issue{
                Number: number,
                Title:  "Custom Issue",
            }, nil
        },
    }

    issue, err := mock.GetIssue(123)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if issue == nil {
        t.Fatal("expected issue, got nil")
    }
    if issue.Number != 123 {
        t.Errorf("expected number 123, got %d", issue.Number)
    }
    if issue.Title != "Custom Issue" {
        t.Errorf("expected title 'Custom Issue', got %s", issue.Title)
    }
}

func TestMockGitHub_WithNilFunc_ReturnsDefault(t *testing.T) {
    mock := &sow.MockGitHub{}

    // Should return nil, nil (not panic)
    issue, err := mock.GetIssue(123)

    if err != nil {
        t.Errorf("expected nil error, got %v", err)
    }
    if issue != nil {
        t.Errorf("expected nil issue, got %+v", issue)
    }
}

func TestMockGitHub_ListIssues_WithNilFunc_ReturnsEmptySlice(t *testing.T) {
    mock := &sow.MockGitHub{}

    issues, err := mock.ListIssues("sow", "open")

    if err != nil {
        t.Errorf("expected nil error, got %v", err)
    }
    if issues == nil {
        t.Error("expected empty slice, got nil")
    }
    if len(issues) != 0 {
        t.Errorf("expected empty slice, got %d issues", len(issues))
    }
}

func TestMockGitHub_ErrorCase(t *testing.T) {
    expectedErr := errors.New("mock error")
    mock := &sow.MockGitHub{
        CheckAvailabilityFunc: func() error {
            return expectedErr
        },
    }

    err := mock.CheckAvailability()

    if err != expectedErr {
        t.Errorf("expected mock error, got %v", err)
    }
}

func TestMockGitHub_CreatePullRequest_CustomFunc(t *testing.T) {
    mock := &sow.MockGitHub{
        CreatePullRequestFunc: func(title, body string, draft bool) (int, string, error) {
            return 42, "https://github.com/owner/repo/pull/42", nil
        },
    }

    number, url, err := mock.CreatePullRequest("Title", "Body", true)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if number != 42 {
        t.Errorf("expected number 42, got %d", number)
    }
    if url != "https://github.com/owner/repo/pull/42" {
        t.Errorf("unexpected url: %s", url)
    }
}
```

**When to Use MockGitHub vs MockExecutor:**

Document this clearly in godoc:

- **MockGitHub**: Testing code that uses GitHubClient interface
  - Commands (issue list, issue check, etc.)
  - Wizard flows
  - Any code that calls GitHub operations

- **MockExecutor**: Testing GitHubCLI implementation itself
  - Testing github_cli.go methods
  - Verifying correct gh CLI commands are called
  - Testing error handling in CLI layer

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/.sow/project/context/issue-90.md` - Complete project requirements (Step 4: lines 276-309)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_client.go` - Interface to implement (created in Task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/exec/mock.go` - MockExecutor pattern to follow (entire file, especially lines 16-75)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/cmd/project/wizard_state.go` - Example consumer that will use MockGitHub (lines 35-45, 54, 68)

## Examples

**Using MockGitHub in Command Tests:**

```go
func TestIssueCheckCommand_Available(t *testing.T) {
    mock := &sow.MockGitHub{
        GetIssueFunc: func(number int) (*sow.Issue, error) {
            return &sow.Issue{
                Number: number,
                Title:  "Test Issue",
                State:  "open",
            }, nil
        },
        GetLinkedBranchesFunc: func(number int) ([]sow.LinkedBranch, error) {
            return []sow.LinkedBranch{}, nil // No branches = available
        },
    }

    // Use mock in command test
    // ...
}
```

**Using MockGitHub in Wizard Tests:**

```go
func TestWizard_IssueSelection(t *testing.T) {
    mock := &sow.MockGitHub{
        ListIssuesFunc: func(label, state string) ([]sow.Issue, error) {
            return []sow.Issue{
                {Number: 1, Title: "Issue 1"},
                {Number: 2, Title: "Issue 2"},
            }, nil
        },
    }

    wizard := &Wizard{
        github: mock,
        // ... other fields
    }

    // Test wizard logic
}
```

**Simulating Errors:**

```go
func TestHandleGitHubError(t *testing.T) {
    mock := &sow.MockGitHub{
        CheckAvailabilityFunc: func() error {
            return sow.ErrGHNotInstalled{}
        },
    }

    err := mock.CheckAvailability()
    // Verify error handling
}
```

**Minimal Mock (uses defaults):**

```go
func TestSimpleFlow(t *testing.T) {
    // Only mock what you need, rest returns defaults
    mock := &sow.MockGitHub{
        CheckAvailabilityFunc: func() error { return nil },
    }

    // All other methods return sensible defaults
    issues, _ := mock.ListIssues("sow", "open")  // Returns []
    issue, _ := mock.GetIssue(123)                // Returns nil
}
```

## Dependencies

**Depends On:**
- Task 010 (Define GitHubClient Interface) - Interface must exist to implement

**Depended On By:**
- Task 060 (Update Wizard Interface) - Could use MockGitHub in wizard tests

**Reason**: Mock implements the interface created in Task 010.

## Constraints

**DO NOT:**
- Add any real GitHub logic (it's a mock)
- Make network calls
- Depend on gh CLI or executor
- Add state tracking beyond function fields
- Make it complicated (follow simple MockExecutor pattern)

**DO:**
- Follow MockExecutor pattern exactly
- Keep it simple and focused
- Provide sensible defaults for nil funcs
- Write comprehensive tests demonstrating usage
- Follow TDD approach (write tests first)
- Document when to use this vs MockExecutor

**Testing Requirements:**
- Write tests FIRST (TDD methodology)
- Test interface compliance (compile-time check)
- Test custom func usage
- Test nil func defaults
- Test all method signatures match interface
- Test error simulation

**Code Quality:**
- Clear godoc with usage examples
- Simple, predictable behavior
- No surprises or magic
- Easy to understand and use
- Consistent with exec.MockExecutor pattern
