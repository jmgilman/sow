# Task 040: GitHub Integration Error Handling

## Context

This task is part of Work Unit 005: Validation and Error Handling for the interactive wizard. The wizard includes a "From GitHub issue" flow that allows users to create projects linked to GitHub issues. This flow requires the `gh` CLI to be installed and authenticated, and it must handle various GitHub-specific errors gracefully.

The existing `cli/internal/sow/github.go` provides error types (`ErrGHNotInstalled`, `ErrGHNotAuthenticated`, `ErrGHCommand`) but returns technical error messages. This task wraps those errors with user-friendly messages that follow the 3-part pattern (what/how/next) and provide clear recovery instructions.

**Key architectural decision**: This task does NOT modify the existing GitHub client code in `cli/internal/sow/github.go`. Instead, it creates wrapper functions in the wizard package that catch and format GitHub errors for user display.

## Requirements

### 1. GitHub CLI Validation Function

Create `checkGitHubCLI()` in `wizard_helpers.go` to validate `gh` installation and authentication:

```go
// checkGitHubCLI validates that gh CLI is installed and authenticated.
// Returns nil if both checks pass, or user-friendly error if not.
//
// This is called before any GitHub operation to provide clear error
// messages instead of cryptic command failures.
func checkGitHubCLI(github GitHubClient) error
```

**Implementation logic**:

1. Call `github.CheckInstalled()` from existing GitHub client
2. If error is `ErrGHNotInstalled`:
   - Return formatted error using `errorGitHubCLIMissing()` from Task 030
3. Call `github.CheckAuthenticated()`
4. If error is `ErrGHNotAuthenticated`:
   - Return formatted error using new `errorGitHubNotAuthenticated()` function
5. If both pass: return `nil`

**Note**: The `github.Ensure()` method already does these checks, but we need granular control to format each error type differently.

### 2. GitHub Authentication Error Message

Create `errorGitHubNotAuthenticated()` in `wizard_helpers.go`:

```go
// errorGitHubNotAuthenticated returns the formatted error message when
// gh CLI is installed but not authenticated.
func errorGitHubNotAuthenticated() string
```

**Message format**:
```
GitHub CLI not authenticated

The 'gh' command is installed but you're not logged in.

To authenticate:
  Run: gh auth login
  Follow the prompts to log in

Or select "From branch name" instead.
```

**Note**: This is NOT in the design document's error list (lines 421-535), so we create it following the same pattern.

### 3. GitHub Command Error Formatter

Create `formatGitHubError()` in `wizard_helpers.go` to convert GitHub command errors into user-friendly messages:

```go
// formatGitHubError converts GitHub command errors to user-friendly messages.
// Handles specific error cases by parsing stderr output.
//
// Error types handled:
//   - Network errors: "check connection, retry"
//   - Rate limit: "wait or authenticate for higher limit"
//   - Issue not found: "check issue number"
//   - Permission denied: "check repository access"
//   - Unknown errors: show command that failed
//
// Returns formatted error string ready for display.
func formatGitHubError(err error) string
```

**Implementation logic**:

1. Check if error is `ErrGHCommand` using type assertion or `errors.As()`
2. If not, return generic error message
3. Parse `Stderr` field for specific patterns:

**Pattern 1: Network errors**
- Search for: "network", "connection", "timeout", "unreachable"
- Return: "GitHub API is unreachable. Check your internet connection and try again."

**Pattern 2: Rate limit**
- Search for: "rate limit", "API rate limit exceeded"
- Return: "GitHub API rate limit exceeded. Wait a few minutes or authenticate for higher limits."

**Pattern 3: Not found**
- Search for: "not found", "could not resolve", "does not exist"
- Return: "Resource not found. Check the issue number or repository access."

**Pattern 4: Permission denied**
- Search for: "permission denied", "forbidden", "not authorized"
- Return: "Permission denied. Check your GitHub repository access."

**Pattern 5: Unknown**
- Return: Format with command name from error
- Message: "GitHub command failed: gh {command}. Check gh CLI is working correctly."

### 4. Issue Linked Branch Validation

Create `checkIssueLinkedBranch()` in `wizard_helpers.go` to validate that an issue doesn't already have a linked branch:

```go
// checkIssueLinkedBranch validates that a GitHub issue doesn't have an
// existing linked branch. Used before creating a new project from an issue.
//
// Returns:
//   - nil if no linked branches (OK to create)
//   - formatted error if linked branch exists
func checkIssueLinkedBranch(github GitHubClient, issueNumber int) error
```

**Implementation logic**:

1. Call `github.GetLinkedBranches(issueNumber)`
2. If error: return `formatGitHubError(err)`
3. If result is empty slice: return `nil` (OK to create)
4. If result has branches:
   - Get first branch name: `branches[0].Name`
   - Return formatted error using `errorIssueAlreadyLinked(issueNumber, branchName)` from Task 030

**Note**: We only care about the first linked branch for error display. If multiple branches exist (rare), show the first one.

### 5. Issue Label Filtering

Create `filterIssuesBySowLabel()` in `wizard_helpers.go` to filter issues by the 'sow' label:

```go
// filterIssuesBySowLabel filters issues to only include those with 'sow' label.
// Used by issue selection screen to show only sow-related issues.
//
// Returns:
//   - Slice of issues that have the 'sow' label
func filterIssuesBySowLabel(issues []sow.Issue) []sow.Issue
```

**Implementation**:

1. Create empty result slice
2. Iterate through input issues
3. For each issue, call `issue.HasLabel("sow")`
4. If true, append to result
5. Return filtered slice

**Note**: The existing `sow.Issue` struct has a `HasLabel(label string) bool` method (in `cli/internal/sow/github.go` lines 93-101).

### 6. Integration Helper Function

Create `ensureGitHubAvailable()` in `wizard_helpers.go` as a convenience wrapper:

```go
// ensureGitHubAvailable checks that GitHub CLI is available and working.
// Returns nil if OK, or displays error and returns error.
//
// This is a convenience function that combines checkGitHubCLI with error display.
// Use this at the start of GitHub-dependent flows.
func ensureGitHubAvailable(github GitHubClient) error
```

**Implementation**:

1. Call `checkGitHubCLI(github)`
2. If error: call `showError(err.Error())` from Task 030
3. Return the error (caller can decide whether to continue or cancel)

## Acceptance Criteria

### Functional Requirements

1. **GitHub CLI Validation Works**:
   - Detects when `gh` is not installed
   - Detects when `gh` is not authenticated
   - Returns nil when both checks pass
   - Returns formatted error messages

2. **Error Message Formatting Works**:
   - Network errors → connection message
   - Rate limit errors → wait message
   - Not found errors → check resource message
   - Permission errors → access message
   - Unknown errors → show command that failed

3. **Issue Linked Branch Check Works**:
   - Returns nil when no linked branches exist
   - Returns error when linked branch exists
   - Includes branch name in error message
   - Handles GetLinkedBranches errors

4. **Issue Label Filtering Works**:
   - Filters issues to only those with 'sow' label
   - Returns empty slice when no matches
   - Preserves issue order
   - Doesn't modify original slice

5. **Integration Helper Works**:
   - Combines validation and error display
   - Returns error for caller to handle
   - Displays user-friendly messages

### Test Requirements (TDD Approach)

**Write ALL tests FIRST, then implement functions to pass them.**

1. **GitHub CLI Check Tests** (`wizard_helpers_test.go`):
   ```go
   func TestCheckGitHubCLI(t *testing.T)
   ```
   - Mock GitHubClient interface
   - Test not installed scenario
   - Test not authenticated scenario
   - Test both checks pass
   - At least 4 test cases

2. **GitHub Error Formatting Tests** (`wizard_helpers_test.go`):
   ```go
   func TestFormatGitHubError(t *testing.T)
   ```
   - Test each error pattern (network, rate limit, etc.)
   - Test with non-ErrGHCommand errors
   - Test with empty stderr
   - At least 8 test cases

3. **Issue Linked Branch Tests** (`wizard_helpers_test.go`):
   ```go
   func TestCheckIssueLinkedBranch(t *testing.T)
   ```
   - Mock GitHubClient interface
   - Test no linked branches (success)
   - Test one linked branch (error)
   - Test multiple linked branches (error with first)
   - Test GetLinkedBranches error
   - At least 5 test cases

4. **Issue Filtering Tests** (`wizard_helpers_test.go`):
   ```go
   func TestFilterIssuesBySowLabel(t *testing.T)
   ```
   - Test with all issues having 'sow' label
   - Test with no issues having 'sow' label
   - Test with mixed labels
   - Test with empty input
   - At least 5 test cases

### Code Quality

- All functions have clear godoc comments
- Error messages are user-friendly and actionable
- Tests use mock interfaces (no real GitHub API calls)
- Pattern matching is case-insensitive for robustness

## Technical Details

### File Location

**Add functions to**: `/cli/cmd/project/wizard_helpers.go`
**Add tests to**: `/cli/cmd/project/wizard_helpers_test.go`

### Imports Required

```go
import (
    "errors"
    "fmt"
    "strings"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

### Existing Types to Reuse

**GitHub error types** (from `cli/internal/sow/github.go`):
```go
type ErrGHNotInstalled struct{}
type ErrGHNotAuthenticated struct{}
type ErrGHCommand struct {
    Command string
    Stderr  string
    Err     error
}
```

**GitHub client interface** (already defined in `wizard_state.go`):
```go
type GitHubClient interface {
    Ensure() error
    CheckInstalled() error
    CheckAuthenticated() error
    ListIssues(label, state string) ([]sow.Issue, error)
    GetLinkedBranches(number int) ([]sow.LinkedBranch, error)
    CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)
    GetIssue(number int) (*sow.Issue, error)
}
```

**Issue and LinkedBranch types** (from `cli/internal/sow/github.go`):
```go
type Issue struct {
    Number int
    Title  string
    Body   string
    State  string
    URL    string
    Labels []struct {
        Name string
    }
}

func (i *Issue) HasLabel(label string) bool

type LinkedBranch struct {
    Name string
    URL  string
}
```

### Error Detection Patterns

**For `formatGitHubError()`, search stderr using case-insensitive matching**:

```go
stderrLower := strings.ToLower(stderr)

if strings.Contains(stderrLower, "network") ||
   strings.Contains(stderrLower, "connection") ||
   strings.Contains(stderrLower, "timeout") ||
   strings.Contains(stderrLower, "unreachable") {
    // Network error
}
```

**Priority order** (check in this order):
1. Rate limit (most specific)
2. Network errors
3. Not found
4. Permission denied
5. Unknown (fallback)

### Testing with Mocks

**Mock GitHubClient for tests**:

```go
type mockGitHubClient struct {
    checkInstalledErr     error
    checkAuthenticatedErr error
    linkedBranches        []sow.LinkedBranch
    linkedBranchesErr     error
}

func (m *mockGitHubClient) CheckInstalled() error {
    return m.checkInstalledErr
}

func (m *mockGitHubClient) CheckAuthenticated() error {
    return m.checkAuthenticatedErr
}

func (m *mockGitHubClient) GetLinkedBranches(number int) ([]sow.LinkedBranch, error) {
    return m.linkedBranches, m.linkedBranchesErr
}

// ... implement other methods
```

## Relevant Inputs

**Existing GitHub error types**:
- `cli/internal/sow/github.go` (lines 39-71: error types)
- `cli/internal/sow/github.go` (lines 87-101: LinkedBranch and Issue.HasLabel)
- `cli/internal/sow/github.go` (lines 105-136: CheckInstalled, CheckAuthenticated, Ensure)
- `cli/internal/sow/github.go` (lines 201-254: GetLinkedBranches implementation)

**Design documents**:
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` (lines 517-535: GitHub CLI missing error)
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` (lines 445-458: Issue already linked error)

**Related task outputs**:
- Task 030: Error display components (errorGitHubCLIMissing, errorIssueAlreadyLinked, formatError, showError)

**Wizard integration**:
- `cli/cmd/project/wizard_state.go` (GitHubClient interface, will use these functions)

## Examples

### Example 1: Check GitHub CLI Implementation

```go
// checkGitHubCLI validates that gh CLI is installed and authenticated.
func checkGitHubCLI(github GitHubClient) error {
    // Check installation
    if err := github.CheckInstalled(); err != nil {
        var notInstalled sow.ErrGHNotInstalled
        if errors.As(err, &notInstalled) {
            return errors.New(errorGitHubCLIMissing())
        }
        return fmt.Errorf("failed to check gh installation: %w", err)
    }

    // Check authentication
    if err := github.CheckAuthenticated(); err != nil {
        var notAuthenticated sow.ErrGHNotAuthenticated
        if errors.As(err, &notAuthenticated) {
            return errors.New(errorGitHubNotAuthenticated())
        }
        return fmt.Errorf("failed to check gh authentication: %w", err)
    }

    return nil
}
```

### Example 2: Format GitHub Error Implementation

```go
// formatGitHubError converts GitHub command errors to user-friendly messages.
func formatGitHubError(err error) string {
    var ghErr sow.ErrGHCommand
    if !errors.As(err, &ghErr) {
        // Not a GitHub command error, return generic message
        return fmt.Sprintf("GitHub operation failed: %v", err)
    }

    stderr := strings.ToLower(ghErr.Stderr)

    // Check for rate limit (most specific)
    if strings.Contains(stderr, "rate limit") {
        return "GitHub API rate limit exceeded.\n\n" +
            "To fix:\n" +
            "  Wait a few minutes and try again\n" +
            "  Or run: gh auth login (for higher limits)"
    }

    // Check for network errors
    if strings.Contains(stderr, "network") ||
       strings.Contains(stderr, "connection") ||
       strings.Contains(stderr, "timeout") ||
       strings.Contains(stderr, "unreachable") {
        return "Cannot reach GitHub.\n\n" +
            "Check your internet connection and try again."
    }

    // Check for not found
    if strings.Contains(stderr, "not found") ||
       strings.Contains(stderr, "does not exist") {
        return "Resource not found.\n\n" +
            "Check the issue number or repository access."
    }

    // Check for permission denied
    if strings.Contains(stderr, "permission denied") ||
       strings.Contains(stderr, "forbidden") ||
       strings.Contains(stderr, "not authorized") {
        return "Permission denied.\n\n" +
            "Check your GitHub repository access."
    }

    // Unknown error - show command
    return fmt.Sprintf("GitHub command failed: gh %s\n\n"+
        "Check that gh CLI is working correctly:\n"+
        "  Run: gh auth status",
        ghErr.Command)
}
```

### Example 3: Check Issue Linked Branch Implementation

```go
// checkIssueLinkedBranch validates that a GitHub issue doesn't have an
// existing linked branch.
func checkIssueLinkedBranch(github GitHubClient, issueNumber int) error {
    branches, err := github.GetLinkedBranches(issueNumber)
    if err != nil {
        // Format the GitHub error for user display
        return errors.New(formatGitHubError(err))
    }

    // If no linked branches, OK to create
    if len(branches) == 0 {
        return nil
    }

    // Issue already has linked branch - show error
    branchName := branches[0].Name
    return errors.New(errorIssueAlreadyLinked(issueNumber, branchName))
}
```

### Example 4: Filter Issues by Label Implementation

```go
// filterIssuesBySowLabel filters issues to only include those with 'sow' label.
func filterIssuesBySowLabel(issues []sow.Issue) []sow.Issue {
    var filtered []sow.Issue

    for _, issue := range issues {
        if issue.HasLabel("sow") {
            filtered = append(filtered, issue)
        }
    }

    return filtered
}
```

### Example 5: Test with Mock Client

```go
func TestCheckGitHubCLI(t *testing.T) {
    tests := []struct {
        name          string
        installErr    error
        authErr       error
        wantErr       bool
        errContains   string
    }{
        {
            name:        "both checks pass",
            installErr:  nil,
            authErr:     nil,
            wantErr:     false,
        },
        {
            name:        "not installed",
            installErr:  sow.ErrGHNotInstalled{},
            wantErr:     true,
            errContains: "not found",
        },
        {
            name:        "not authenticated",
            installErr:  nil,
            authErr:     sow.ErrGHNotAuthenticated{},
            wantErr:     true,
            errContains: "not authenticated",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockGitHubClient{
                checkInstalledErr:     tt.installErr,
                checkAuthenticatedErr: tt.authErr,
            }

            err := checkGitHubCLI(mock)

            if (err != nil) != tt.wantErr {
                t.Errorf("checkGitHubCLI() error = %v, wantErr %v", err, tt.wantErr)
            }

            if err != nil && tt.errContains != "" {
                if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("error message %q does not contain %q",
                        err.Error(), tt.errContains)
                }
            }
        })
    }
}
```

### Example 6: Not Authenticated Error Message

```go
// errorGitHubNotAuthenticated returns the formatted error message when
// gh CLI is installed but not authenticated.
func errorGitHubNotAuthenticated() string {
    return formatError(
        "GitHub CLI not authenticated",
        "The 'gh' command is installed but you're not logged in.",
        "To authenticate:\n"+
        "  Run: gh auth login\n"+
        "  Follow the prompts to log in\n\n"+
        "Or select \"From branch name\" instead.",
    )
}
```

## Dependencies

### Required Before This Task

- **Task 030**: Error display components (provides formatError, showError, errorGitHubCLIMissing, errorIssueAlreadyLinked)
- Existing GitHub client in `cli/internal/sow/github.go`
- GitHubClient interface in `wizard_state.go`

### Provides For Other Tasks

- **Work Unit 003**: GitHub issue flow (uses all these functions)
- Integration tests for GitHub-dependent wizard flows

### External Dependencies

- Existing sow packages: `cli/internal/sow`
- Go standard library: `errors`, `fmt`, `strings`

## Constraints

### No Modifications to Existing Code

**DO NOT modify** `cli/internal/sow/github.go`. The existing error types and GitHub client work correctly. This task creates WRAPPER functions in the wizard package.

**Why?** Separation of concerns:
- `cli/internal/sow/github.go`: Low-level GitHub operations
- `cli/cmd/project/wizard_helpers.go`: User-facing error formatting

### Error Message Consistency

All GitHub error messages should follow the 3-part pattern:
1. What went wrong
2. How to fix
3. Next steps (or alternative path)

### Performance Requirements

- Error checking should be fast (< 100ms for validation)
- No retry logic here (let caller decide)
- No caching of GitHub state (always fresh check)

### What NOT to Do

- **Don't call GitHub API without checking CLI first** - Always validate installation/auth
- **Don't expose technical details in errors** - Make them user-friendly
- **Don't modify existing GitHub error types** - Wrap them instead
- **Don't add new GitHub operations** - This is error handling only
- **Don't implement retry logic** - That's the wizard's job

## Notes for Implementer

### GitHub CLI Behavior

The `gh` CLI has specific behavior:
- `CheckInstalled()`: Returns `ErrGHNotInstalled` if command not found
- `CheckAuthenticated()`: Returns `ErrGHNotAuthenticated` if not logged in
- `GetLinkedBranches()`: Returns empty slice (not error) when no branches linked
- Command errors include stderr output for diagnosis

Study the existing implementation in `cli/internal/sow/github.go` to understand these behaviors.

### Error Pattern Priority

When parsing `ErrGHCommand.Stderr`, check patterns in order of specificity:
1. Rate limit (very specific message)
2. Network errors (multiple possible messages)
3. Not found (common for bad issue numbers)
4. Permission denied (access issues)
5. Unknown (catch-all)

This ordering ensures more specific errors are caught before generic ones.

### Mock Interface for Testing

The `GitHubClient` interface is defined in `wizard_state.go`. Your tests should:
1. Create mock implementation of this interface
2. Set return values for each test case
3. Pass mock to functions under test
4. Verify correct error handling

**Don't call real GitHub API in tests.** Use mocks exclusively.

### Integration with Wizard Flows

These functions will be called from Work Unit 003 (GitHub issue flow):

```go
// In issue selection flow
if err := ensureGitHubAvailable(w.github); err != nil {
    // Error already displayed, return to main menu
    w.state = StateEntry
    return nil
}

// Before creating project from issue
if err := checkIssueLinkedBranch(w.github, issueNumber); err != nil {
    showError(err.Error())
    // Let user try different issue
    return nil
}
```

Design your functions to integrate cleanly with this pattern.
