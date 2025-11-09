# Issue #90: GitHub Client Interface Extraction & Refactoring

**URL**: https://github.com/jmgilman/sow/issues/90
**State**: OPEN

## Description

# Work Unit 002: GitHub Client Interface Extraction & Refactoring

## Behavioral Goal

As a developer working with the sow CLI, I need a flexible GitHub client architecture that can work with both the `gh` CLI tool (for local development) and the GitHub REST/GraphQL API (for web VMs), so that all GitHub operations function identically across both environments without any code changes.

**Success Criteria:**
- GitHubClient interface exists and defines all required GitHub operations
- Existing `gh` CLI implementation renamed to GitHubCLI and implements the interface
- Factory function auto-detects environment (GITHUB_TOKEN vs gh CLI) and returns appropriate client
- All existing callsites continue to work without modification (backward compatibility)
- Enhanced method signatures support new operations (draft PRs, PR updates, marking PR ready)
- Interface compliance verified at compile time
- Mock implementation available for testing consumers

## Existing Code Context

This work unit refactors the existing GitHub client implementation to extract a common interface that will enable dual client support for Claude Code web integration. The current implementation at `cli/internal/sow/github.go` wraps the `gh` CLI tool using the executor pattern from `cli/internal/exec/`. The GitHub client is created on-demand in various commands (issue commands, project wizard) rather than stored in the sow.Context.

### Current Architecture

The existing GitHub client follows a clean architecture:

1. **Executor Pattern**: The client accepts an `exec.Executor` interface, making it easy to mock in tests. The executor abstraction handles command execution and already supports both real commands (LocalExecutor) and mocking (MockExecutor).

2. **Error Handling**: Custom error types with Unwrap() support for error chains:
   - `ErrGHNotInstalled` - gh CLI not found
   - `ErrGHNotAuthenticated` - gh not authenticated
   - `ErrGHCommand` - command execution failure with stderr details

3. **Instantiation Pattern**: GitHub clients are created directly where needed using `sow.NewGitHub(exec.NewLocal("gh"))`. There's no centralized context storage (the `context.GitHub()` method exists but creates a new empty instance).

4. **Existing Interface**: The project wizard already defines a local `GitHubClient` interface at `cli/cmd/project/wizard_state.go:37-45` that partially matches what we need. This demonstrates the need for a shared interface.

### Key Files to Understand

**Current Implementation:**
- `cli/internal/sow/github.go:1-466` - Current GitHub struct and all operations
- `cli/internal/sow/github_test.go:1-108` - Existing test patterns using MockExecutor
- `cli/internal/exec/executor.go:1-167` - Executor interface and LocalExecutor
- `cli/internal/exec/mock.go:1-76` - MockExecutor for testing

**Usage Patterns:**
- `cli/cmd/issue/check.go:32-33` - Creates GitHub client: `gh := sow.NewGitHub(exec.NewLocal("gh"))`
- `cli/cmd/issue/list.go:35` - Same pattern
- `cli/cmd/issue/show.go:34` - Same pattern
- `cli/cmd/project/wizard_state.go:35-45` - Defines local GitHubClient interface
- `cli/cmd/project/wizard_state.go:68` - Creates client in wizard constructor
- `cli/internal/sow/context.go:124-129` - Empty GitHub() method (creates new empty instance)

**Patterns to Follow:**
- `cli/internal/exec/executor.go:20-45` - Interface design with clear documentation
- `cli/internal/exec/executor.go:166` - Compile-time interface compliance: `var _ Executor = (*LocalExecutor)(nil)`
- `cli/internal/exec/mock.go:16-23` - Mock implementation pattern

### Key Files to Create

- `cli/internal/sow/github_client.go` - GitHubClient interface definition
- `cli/internal/sow/github_cli.go` - Renamed CLI implementation (from github.go)
- `cli/internal/sow/github_factory.go` - Auto-detection factory
- `cli/internal/sow/github_mock.go` - Mock implementation for testing consumers

### Key Files to Modify

- `cli/internal/sow/github_test.go` - Rename to github_cli_test.go, update references
- `cli/cmd/project/wizard_state.go:35-45` - Remove local interface, use shared one
- All callsites using `sow.NewGitHub()` - Update to use factory (optional, can be phased)

## Design Context

The Claude Code Web integration design (`.sow/knowledge/designs/claude-code-web-integration.md`, Section 4: GitHub Integration) specifies a dual GitHub client architecture. This work unit implements Phase 1 of that design: extracting the interface to enable both CLI and API implementations.

### Design Requirements

From the design document (lines 206-218), the interface must support:

```go
type GitHubClient interface {
    CheckAvailability() error  // Renamed from CheckInstalled
    ListIssues(label, state string) ([]Issue, error)
    GetIssue(number int) (*Issue, error)
    CreateIssue(title, body string, labels []string) (*Issue, error)
    GetLinkedBranches(number int) ([]LinkedBranch, error)
    CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)
    CreatePullRequest(title, body string, draft bool) (number int, url string, error)
    UpdatePullRequest(number int, title, body string) error  // NEW
    MarkPullRequestReady(number int) error  // NEW
}
```

### Discovery Analysis Insights

The discovery document (`.sow/project/discovery/claude-code-web-analysis.md`, Section 4: GitHub Integration) identifies the following gaps:

**Current vs Design Gaps** (lines 266-278):

| Method | Current | Design | Status |
|--------|---------|--------|--------|
| CheckInstalled | ✓ | CheckAvailability | Different names |
| CreatePullRequest | ✓ but incomplete | Enhanced | Signature mismatch (no draft, missing number return) |
| UpdatePullRequest | ✗ | ✓ | MISSING |
| MarkPullRequestReady | ✗ | ✓ | MISSING |

**Key Constraints** (lines 287-314):
1. Extract interface without breaking existing callsites
2. Rename implementation from `GitHub` to `GitHubCLI`
3. Add missing methods: UpdatePullRequest, MarkPullRequestReady
4. Enhance CreatePullRequest signature to return PR number and support draft flag
5. Rename CheckInstalled to CheckAvailability in interface (keep backward compat in impl)

### ADR Guidance

The ADR at `.sow/knowledge/adrs/github-client-dual-implementation.md` provides detailed implementation strategy. Key points:

**Refactoring Strategy** (recommend Option A from discovery doc):
1. Create GitHubClient interface first
2. Keep GitHub struct as-is initially
3. Add factory that returns current GitHub implementation
4. Gradually rename to GitHubCLI
5. Add interface compliance check

**Factory Auto-Detection Logic** (lines 97-110):
- Check for GITHUB_TOKEN environment variable
- If present: use API client (work unit 004, not this one)
- If absent: use CLI client (this work unit)
- Extract owner/repo from git remote for API client

## Implementation Approach

This refactoring follows the **strangler pattern**: introduce the new interface alongside existing code, migrate gradually, avoid breaking changes.

### Step 1: Define the Interface

Create `cli/internal/sow/github_client.go` with the interface definition:

```go
package sow

// GitHubClient defines operations for GitHub issue, PR, and branch management.
//
// Two implementations exist:
// - GitHubCLI: Wraps `gh` CLI for local development
// - GitHubAPI: Uses REST/GraphQL APIs for web VMs (work unit 004)
//
// Use NewGitHubClient() factory for automatic environment detection.
type GitHubClient interface {
    // CheckAvailability verifies that GitHub access is available.
    // For CLI: checks gh is installed and authenticated.
    // For API: validates token and connectivity.
    CheckAvailability() error

    // ListIssues returns issues matching label and state filters.
    ListIssues(label, state string) ([]Issue, error)

    // GetIssue retrieves a single issue by number.
    GetIssue(number int) (*Issue, error)

    // CreateIssue creates a new issue with title, body, and labels.
    CreateIssue(title, body string, labels []string) (*Issue, error)

    // GetLinkedBranches returns branches linked to an issue.
    GetLinkedBranches(number int) ([]LinkedBranch, error)

    // CreateLinkedBranch creates a branch linked to an issue.
    CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)

    // CreatePullRequest creates a PR (optionally as draft).
    // Returns PR number, URL, and error.
    CreatePullRequest(title, body string, draft bool) (number int, url string, error)

    // UpdatePullRequest updates PR title and body.
    UpdatePullRequest(number int, title, body string) error

    // MarkPullRequestReady converts a draft PR to ready for review.
    MarkPullRequestReady(number int) error
}
```

**Key Design Decisions:**
- Interface methods use concrete types (Issue, LinkedBranch) not pointers except where semantically meaningful (GetIssue returns *Issue to allow nil)
- CheckAvailability abstracts "is this client ready?" across implementations
- CreatePullRequest returns both number and URL for downstream operations
- Method signatures match existing patterns where possible

### Step 2: Rename Implementation

Rename `cli/internal/sow/github.go` to `cli/internal/sow/github_cli.go` and update:

```go
// GitHubCLI implements GitHubClient using the gh CLI tool.
type GitHubCLI struct {
    gh exec.Executor
}

// NewGitHubCLI creates a GitHub client backed by gh CLI.
func NewGitHubCLI(executor exec.Executor) *GitHubCLI {
    return &GitHubCLI{gh: executor}
}

// CheckAvailability implements GitHubClient.
func (g *GitHubCLI) CheckAvailability() error {
    return g.Ensure() // Reuse existing Ensure() logic
}

// Add UpdatePullRequest implementation
func (g *GitHubCLI) UpdatePullRequest(number int, title, body string) error {
    // gh pr edit <number> --title <title> --body <body>
}

// Add MarkPullRequestReady implementation
func (g *GitHubCLI) MarkPullRequestReady(number int) error {
    // gh pr ready <number>
}

// Update CreatePullRequest signature
func (g *GitHubCLI) CreatePullRequest(title, body string, draft bool) (int, string, error) {
    // Add --draft flag support
    // Parse PR number from URL
    // Return number, url, error
}

// Interface compliance check
var _ GitHubClient = (*GitHubCLI)(nil)
```

**Backward Compatibility:**
- Keep all existing methods (CheckInstalled, CheckAuthenticated, Ensure)
- CheckAvailability delegates to Ensure()
- Old callers using `*GitHub` will need to update to `*GitHubCLI` or use interface

### Step 3: Create Factory

Create `cli/internal/sow/github_factory.go` with auto-detection:

```go
package sow

import (
    "os"
    "github.com/jmgilman/sow/cli/internal/exec"
)

// NewGitHubClient creates a GitHub client with automatic environment detection.
//
// Detection logic:
// - If GITHUB_TOKEN env var is set: Use API client (work unit 004)
// - Otherwise: Use CLI client backed by gh command
//
// For web VMs, GITHUB_TOKEN will be set by Claude Code.
// For local dev, gh CLI is expected to be installed and authenticated.
func NewGitHubClient() (GitHubClient, error) {
    if token := os.Getenv("GITHUB_TOKEN"); token != "" {
        // Future: Return GitHubAPI client (work unit 004)
        return nil, errors.New("API client not yet implemented; unset GITHUB_TOKEN to use gh CLI")
    }

    // Default to CLI client
    return NewGitHubCLI(exec.NewLocal("gh")), nil
}
```

**Key Points:**
- Factory returns interface, not concrete type
- GITHUB_TOKEN presence triggers API mode (future work unit)
- For now, API mode returns error with helpful message
- CLI mode works immediately with no changes

### Step 4: Create Mock Implementation

Create `cli/internal/sow/github_mock.go` for testing consumers:

```go
package sow

// MockGitHub is a test mock for GitHubClient.
//
// Usage in tests:
//   mock := &sow.MockGitHub{
//       CheckAvailabilityFunc: func() error { return nil },
//       GetIssueFunc: func(n int) (*Issue, error) {
//           return &Issue{Number: n, Title: "Test"}, nil
//       },
//   }
type MockGitHub struct {
    CheckAvailabilityFunc     func() error
    ListIssuesFunc            func(label, state string) ([]Issue, error)
    GetIssueFunc              func(number int) (*Issue, error)
    // ... all interface methods
}

func (m *MockGitHub) CheckAvailability() error {
    if m.CheckAvailabilityFunc != nil {
        return m.CheckAvailabilityFunc()
    }
    return nil
}

// ... implement all interface methods

// Interface compliance check
var _ GitHubClient = (*MockGitHub)(nil)
```

### Step 5: Update Tests

Rename `cli/internal/sow/github_test.go` to `cli/internal/sow/github_cli_test.go` and update:

- Change `sow.GitHub` to `sow.GitHubCLI`
- Change `sow.NewGitHub()` to `sow.NewGitHubCLI()`
- Add tests for new methods (UpdatePullRequest, MarkPullRequestReady)
- Add factory tests in new `cli/internal/sow/github_factory_test.go`

### Step 6: Migration Strategy (Phased)

**Phase 1 (This Work Unit):**
- Create interface and rename implementation
- Existing callsites continue using `NewGitHub()` which still exists
- Deprecate `NewGitHub()` in godoc comments

**Phase 2 (Future):**
- Update callsites to use factory: `NewGitHubClient()`
- Remove deprecated `NewGitHub()` constructor

**Phase 3 (Work Unit 004):**
- Implement GitHubAPI
- Factory returns API client when GITHUB_TOKEN set

## Dependencies

### Depends On
- None (this is a foundation piece)

### Depended On By
- **Work Unit 004 (GitHub API Implementation)** - Requires the GitHubClient interface extracted in this work unit. The API client will implement this interface to provide GitHub operations via REST/GraphQL APIs instead of the gh CLI.

### Why Dependencies Exist
Work unit 004 cannot begin until the interface contract is established. The interface defines what operations both implementations must support. Without it, we risk API and CLI implementations diverging in behavior or capability.

## Acceptance Criteria

Objective, measurable completion criteria that reviewers will verify:

1. ✅ **Interface exists**: `cli/internal/sow/github_client.go` defines GitHubClient interface with all required methods
2. ✅ **Implementation renamed**: GitHub struct renamed to GitHubCLI in `cli/internal/sow/github_cli.go`
3. ✅ **Compile-time compliance**: `var _ GitHubClient = (*GitHubCLI)(nil)` present and compiles
4. ✅ **Factory created**: `cli/internal/sow/github_factory.go` with NewGitHubClient() auto-detection
5. ✅ **Mock available**: `cli/internal/sow/github_mock.go` implements GitHubClient for testing
6. ✅ **New methods implemented**: UpdatePullRequest and MarkPullRequestReady work with gh CLI
7. ✅ **Enhanced CreatePullRequest**: Supports draft flag, returns PR number and URL
8. ✅ **CheckAvailability added**: Interface method delegates to existing Ensure() logic
9. ✅ **All existing tests pass**: Renamed tests in github_cli_test.go pass without modification to test logic
10. ✅ **Factory tests exist**: Tests verify GITHUB_TOKEN detection and client type selection
11. ✅ **Backward compatibility maintained**: Old NewGitHub() still works (deprecated but functional)
12. ✅ **Documentation updated**: Interface and types have clear godoc comments

## Testing Strategy (TDD)

### Unit Tests

**Interface Compliance Tests** (`cli/internal/sow/github_client_test.go`):
```go
func TestGitHubCLI_ImplementsInterface(t *testing.T) {
    var _ GitHubClient = (*GitHubCLI)(nil)
}

func TestMockGitHub_ImplementsInterface(t *testing.T) {
    var _ GitHubClient = (*MockGitHub)(nil)
}
```

**Factory Tests** (`cli/internal/sow/github_factory_test.go`):
```go
func TestNewGitHubClient_WithGitHubToken_ReturnsError(t *testing.T) {
    os.Setenv("GITHUB_TOKEN", "test-token")
    defer os.Unsetenv("GITHUB_TOKEN")

    client, err := NewGitHubClient()
    // Should error because API client not yet implemented
    if err == nil {
        t.Error("expected error for unimplemented API client")
    }
}

func TestNewGitHubClient_WithoutToken_ReturnsCLIClient(t *testing.T) {
    os.Unsetenv("GITHUB_TOKEN")

    client, err := NewGitHubClient()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Verify type (requires type assertion in test)
    if _, ok := client.(*GitHubCLI); !ok {
        t.Errorf("expected *GitHubCLI, got %T", client)
    }
}
```

**New Method Tests** (`cli/internal/sow/github_cli_test.go`):
```go
func TestGitHubCLI_UpdatePullRequest(t *testing.T) {
    mock := &exec.MockExecutor{
        ExistsFunc: func() bool { return true },
        RunSilentFunc: func(...string) error { return nil },
        RunFunc: func(args ...string) (string, string, error) {
            // Verify command: gh pr edit 123 --title "New" --body "Body"
            return "", "", nil
        },
    }

    gh := NewGitHubCLI(mock)
    err := gh.UpdatePullRequest(123, "New Title", "New Body")
    // Assertions...
}

func TestGitHubCLI_MarkPullRequestReady(t *testing.T) {
    mock := &exec.MockExecutor{
        ExistsFunc: func() bool { return true },
        RunSilentFunc: func(...string) error { return nil },
        RunFunc: func(args ...string) (string, string, error) {
            // Verify command: gh pr ready 123
            return "", "", nil
        },
    }

    gh := NewGitHubCLI(mock)
    err := gh.MarkPullRequestReady(123)
    // Assertions...
}

func TestGitHubCLI_CreatePullRequest_Draft(t *testing.T) {
    mock := &exec.MockExecutor{
        ExistsFunc: func() bool { return true },
        RunSilentFunc: func(...string) error { return nil },
        RunFunc: func(args ...string) (string, string, error) {
            // Verify --draft flag present
            // Return PR URL
            return "https://github.com/owner/repo/pull/42\n", "", nil
        },
    }

    gh := NewGitHubCLI(mock)
    number, url, err := gh.CreatePullRequest("Title", "Body", true)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if number != 42 {
        t.Errorf("expected PR number 42, got %d", number)
    }
    // More assertions...
}
```

### Integration Tests

**Existing CLI Tests Continue to Pass:**
- All existing tests in github_cli_test.go (renamed from github_test.go)
- No changes to test logic, only struct/function names
- Verify gh CLI operations still work via MockExecutor

### Test Files Structure

- `cli/internal/sow/github_client_test.go` - Interface compliance tests
- `cli/internal/sow/github_cli_test.go` - CLI implementation tests (renamed from github_test.go)
- `cli/internal/sow/github_factory_test.go` - Factory auto-detection tests
- `cli/internal/sow/github_mock_test.go` - Mock implementation tests

## Implementation Notes

### Backward Compatibility

**Preservation Strategy:**
- Keep old `NewGitHub()` function as deprecated wrapper to `NewGitHubCLI()`
- All existing callsites work unchanged initially
- Deprecation warning in godoc guides users to new factory
- Migration can happen gradually in separate PRs

**Example Deprecation:**
```go
// NewGitHub creates a GitHub CLI client.
// Deprecated: Use NewGitHubCLI() for explicit CLI client, or NewGitHubClient() for auto-detection.
func NewGitHub(executor exec.Executor) *GitHubCLI {
    return NewGitHubCLI(executor)
}
```

### Error Handling Patterns

Follow existing patterns from current implementation:

**Custom Error Types:**
- Keep ErrGHNotInstalled, ErrGHNotAuthenticated, ErrGHCommand
- Add ErrGitHubAPI for API implementation (work unit 004)
- All errors support Unwrap() for error chains

**Error Messages:**
- Clear, actionable messages (e.g., "Run: gh auth login")
- Include command that failed and stderr output
- Consistent format across CLI and API implementations

### Refactoring Risks & Mitigation

**Medium Risk: Renaming GitHub to GitHubCLI**
- Affects all files that reference the type
- Mitigation: Comprehensive grep for `*sow.GitHub`, `sow.GitHub{`, etc.
- Keep deprecated NewGitHub() wrapper during transition
- Update in phases: rename file, update tests, update callers

**Low Risk: Interface extraction is additive**
- Doesn't break existing code
- Interface methods match existing public methods
- Only adds new abstraction layer

**Low Risk: New methods are additions**
- UpdatePullRequest, MarkPullRequestReady don't exist yet
- No existing callers to break
- CreatePullRequest signature change requires careful testing

### File Organization

**Before:**
```
cli/internal/sow/
├── github.go          # GitHub struct + methods
├── github_test.go     # Tests
└── context.go         # Context with GitHub() method
```

**After:**
```
cli/internal/sow/
├── github_client.go      # GitHubClient interface (NEW)
├── github_cli.go         # GitHubCLI implementation (renamed from github.go)
├── github_cli_test.go    # CLI tests (renamed from github_test.go)
├── github_factory.go     # NewGitHubClient() factory (NEW)
├── github_factory_test.go # Factory tests (NEW)
├── github_mock.go        # MockGitHub implementation (NEW)
└── context.go            # Context (unchanged for now)
```

### Future Extension Points

**For Work Unit 004 (API Implementation):**
- Interface is ready: `type GitHubAPI struct { ... }`
- Factory needs one line change: `return NewGitHubAPI(token, owner, repo), nil`
- All callsites using interface already work with both implementations

**For Context Integration:**
- Could add `context.SetGitHub(client GitHubClient)` method
- Factory could be called in context initialization
- Lower priority: current pattern of creating client where needed works fine

### gh CLI Command Reference

For implementing new methods, here are the gh commands:

**Update PR:**
```bash
gh pr edit <number> --title "New Title" --body "New Body"
```

**Mark PR Ready:**
```bash
gh pr ready <number>
```

**Create Draft PR:**
```bash
gh pr create --title "Title" --body "Body" --draft
```

**Get PR Number from URL:**
```
Input:  https://github.com/owner/repo/pull/42
Output: 42 (parse from URL path)
```

## Code Quality Standards

**Documentation:**
- All exported types, functions, and methods have godoc comments
- Interface methods documented with behavior contracts
- Example usage in package-level documentation
- Error types documented with when they're returned

**Testing:**
- All new methods have unit tests
- Mock executor used to avoid gh CLI dependency in tests
- Factory tests cover both environment scenarios
- Interface compliance verified at compile time

**Code Style:**
- Follow existing patterns in executor.go and github.go
- Use table-driven tests where appropriate
- Clear variable names (avoid abbreviations except common ones like gh, err)
- Consistent error handling (wrap with context, return early)

## Summary

This work unit establishes the foundation for dual GitHub client support by extracting a clean interface from the existing gh CLI implementation. The refactoring is designed to be low-risk, backward-compatible, and enable future work (API implementation) without requiring changes to consuming code.

**Key Deliverables:**
1. GitHubClient interface defining all GitHub operations
2. GitHubCLI implementation (renamed from GitHub)
3. Factory with environment auto-detection
4. Mock implementation for testing
5. Enhanced method signatures (draft PRs, PR updates)
6. Comprehensive test coverage

**Not in Scope:**
- GitHub API implementation (work unit 004)
- Updating all callsites to use factory (can be phased)
- Breaking changes to existing code
- New GitHub operations beyond interface requirements
