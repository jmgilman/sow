# Task 010: Define GitHubClient Interface

## Context

This task is part of a refactoring project to extract a GitHubClient interface from the existing GitHub CLI implementation. The goal is to enable dual client support (CLI + API) for Claude Code web integration, where web VMs will use the GitHub API via token authentication instead of the `gh` CLI tool.

**Project Goal**: Extract GitHubClient interface from existing GitHub CLI implementation to enable dual client support (CLI + API) while maintaining backward compatibility.

**This Task's Role**: Define the interface contract that both CLI and API implementations will follow. This is the foundation that enables all subsequent refactoring work.

## Requirements

Create a new file `cli/internal/sow/github_client.go` that defines the GitHubClient interface.

**Interface Requirements:**
1. Define all GitHub operations currently supported by the existing GitHub struct
2. Add new methods required by the design: UpdatePullRequest, MarkPullRequestReady
3. Use CheckAvailability instead of CheckInstalled (more generic name for API client)
4. Enhanced CreatePullRequest signature to return PR number and support draft flag
5. Include comprehensive godoc comments for each method

**Method Signatures (from design):**
- `CheckAvailability() error` - Verifies GitHub access is available
- `ListIssues(label, state string) ([]Issue, error)` - Lists issues with filters
- `GetIssue(number int) (*Issue, error)` - Retrieves single issue
- `CreateIssue(title, body string, labels []string) (*Issue, error)` - Creates issue
- `GetLinkedBranches(number int) ([]LinkedBranch, error)` - Gets linked branches
- `CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)` - Creates linked branch
- `CreatePullRequest(title, body string, draft bool) (number int, url string, error)` - Creates PR (NEW signature)
- `UpdatePullRequest(number int, title, body string) error` - Updates PR (NEW)
- `MarkPullRequestReady(number int) error` - Marks draft PR ready (NEW)

**Documentation Requirements:**
- Interface-level godoc explaining the two implementations (CLI and API)
- Per-method godoc explaining behavior contract and parameters
- Note about using NewGitHubClient() factory for automatic environment detection
- Clear parameter documentation with examples

**Design Decisions:**
- Interface methods use concrete types (Issue, LinkedBranch) not pointers except where semantically meaningful
- CheckAvailability abstracts "is this client ready?" across implementations
- CreatePullRequest returns both number and URL for downstream operations
- Match existing public method patterns where possible

## Acceptance Criteria

- [ ] File `cli/internal/sow/github_client.go` exists
- [ ] GitHubClient interface defined with all 9 required methods
- [ ] CheckAvailability method defined (replaces CheckInstalled in interface)
- [ ] CreatePullRequest signature includes draft parameter and returns (int, string, error)
- [ ] UpdatePullRequest and MarkPullRequestReady methods defined
- [ ] Interface-level godoc comment exists explaining dual implementations
- [ ] Each method has godoc comment explaining behavior and parameters
- [ ] Package compiles without errors
- [ ] Interface uses existing Issue and LinkedBranch types from github.go

## Technical Details

**Package Declaration:**
```go
package sow
```

**Interface Structure Pattern to Follow:**

Look at `cli/internal/exec/executor.go:20-45` for the interface documentation pattern. The interface should have:
1. Interface-level documentation explaining purpose and implementations
2. Usage guidance pointing to factory function
3. Per-method documentation with parameter descriptions

**Type Reuse:**

The interface will reference existing types defined in `cli/internal/sow/github.go`:
- `Issue` struct (lines 76-85)
- `LinkedBranch` struct (lines 88-91)

These types remain in github.go for now and will be accessible since they're in the same package.

**Method Documentation Examples:**

```go
// CheckAvailability verifies that GitHub access is available.
// For CLI: checks gh is installed and authenticated.
// For API: validates token and connectivity.
CheckAvailability() error

// CreatePullRequest creates a PR (optionally as draft).
// Returns PR number, URL, and error.
//
// Parameters:
//   - title: PR title
//   - body: PR description (supports markdown)
//   - draft: If true, creates PR as draft
//
// Returns:
//   - number: PR number for subsequent operations
//   - url: Full GitHub URL to the PR
//   - error: Any error during creation
CreatePullRequest(title, body string, draft bool) (number int, url string, error)
```

**Testing Approach:**

This task defines the interface only - no implementation. The interface will be validated when GitHubCLI implementation is updated in subsequent tasks. However, you should:
1. Ensure package compiles with `go build ./cli/internal/sow`
2. Run existing tests to ensure no breakage: `go test ./cli/internal/sow`

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/.sow/project/context/issue-90.md` - Complete project requirements and design
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github.go` - Existing GitHub implementation to understand current method signatures
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/exec/executor.go` - Interface documentation pattern to follow (lines 20-45)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/cmd/project/wizard_state.go` - Existing local GitHubClient interface that partially matches (lines 35-45)

## Examples

**Interface Declaration Pattern (from executor.go:20-45):**

```go
// Executor defines the interface for executing external commands.
//
// This interface allows for easy mocking in tests while providing a consistent
// API for command execution across the codebase.
type Executor interface {
    // Command returns the command name this executor wraps.
    Command() string

    // Exists checks if the command exists in PATH.
    Exists() bool

    // ... more methods
}
```

**Expected Interface Structure:**

```go
package sow

// GitHubClient defines operations for GitHub issue, PR, and branch management.
//
// Two implementations exist:
// - GitHubCLI: Wraps `gh` CLI for local development
// - GitHubAPI: Uses REST/GraphQL APIs for web VMs (future implementation)
//
// Use NewGitHubClient() factory for automatic environment detection.
type GitHubClient interface {
    // CheckAvailability verifies that GitHub access is available.
    // For CLI: checks gh is installed and authenticated.
    // For API: validates token and connectivity.
    CheckAvailability() error

    // ListIssues returns issues matching label and state filters.
    // ... detailed doc
    ListIssues(label, state string) ([]Issue, error)

    // ... rest of methods with detailed docs
}
```

## Dependencies

None - this is the first task and is purely additive (new file creation).

## Constraints

**DO NOT:**
- Modify existing github.go file in this task
- Change any existing method signatures in GitHub struct
- Create implementations - interface definition only
- Break existing code - this is purely additive

**DO:**
- Keep interface in same package as existing GitHub struct (package sow)
- Reuse existing Issue and LinkedBranch types
- Follow existing documentation patterns from executor.go
- Match existing method names where they exist in GitHub struct
- Use clear, actionable documentation for implementers

**Compatibility:**
- All existing code continues to work unchanged
- No existing tests should break
- Interface is purely additive at this stage

**File Location:**
- Must be in `cli/internal/sow/` directory
- Must be named `github_client.go`
- Must use `package sow`
