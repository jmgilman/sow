# Task 010: GitHub CLI Validation and Error Handling

## Context

This task implements GitHub CLI validation for the interactive project wizard. When users select "From GitHub issue" as their project source, the wizard must verify that the `gh` CLI is installed and authenticated before attempting to fetch issues.

This is the first step in the GitHub issue integration workflow (Work Unit 003), which extends the wizard to support creating projects from GitHub issues labeled with 'sow'. The validation logic ensures clear, helpful errors with fallback options when GitHub integration is unavailable.

The GitHub client itself already exists and is fully implemented at `cli/internal/sow/github.go` with all necessary methods (`CheckInstalled()`, `CheckAuthenticated()`, `Ensure()`). This task focuses on integrating those checks into the wizard flow with appropriate error handling and user guidance.

## Requirements

### 1. GitHub Client Integration in Wizard

Modify `cli/cmd/project/wizard_state.go` to include GitHub client as part of wizard state:

```go
type Wizard struct {
    state       WizardState
    ctx         *sow.Context
    choices     map[string]interface{}
    claudeFlags []string
    cmd         *cobra.Command
    github      *sow.GitHub  // NEW: GitHub client for issue operations
}
```

Update `NewWizard()` in `cli/cmd/project/wizard_state.go` to initialize the GitHub client:

```go
func NewWizard(cmd *cobra.Command, ctx *sow.Context, claudeFlags []string) *Wizard {
    ghExec := sowexec.NewLocal("gh")

    return &Wizard{
        state:       StateEntry,
        ctx:         ctx,
        choices:     make(map[string]interface{}),
        claudeFlags: claudeFlags,
        cmd:         cmd,
        github:      sow.NewGitHub(ghExec),  // NEW
    }
}
```

### 2. Modify handleIssueSelect to Validate GitHub CLI

Update the stub `handleIssueSelect()` method in `cli/cmd/project/wizard_state.go` to perform validation before proceeding:

```go
func (w *Wizard) handleIssueSelect() error {
    // Validate GitHub CLI is available and authenticated
    if err := w.github.Ensure(); err != nil {
        return w.handleGitHubError(err)
    }

    // If validation passes, continue to issue fetching (Task 020)
    // For now, stub until Task 020 implements issue listing
    fmt.Println("GitHub CLI validated successfully")
    w.state = StateComplete
    return nil
}
```

### 3. Implement GitHub Error Handler

Create a new method `handleGitHubError()` in `cli/cmd/project/wizard_state.go` that displays context-appropriate errors and offers fallback options:

```go
// handleGitHubError displays GitHub-related errors and offers fallback paths.
// Returns nil to keep wizard running (user can choose fallback).
func (w *Wizard) handleGitHubError(err error) error {
    var errorMsg string
    var fallbackMsg string

    // Determine error type using type assertion
    if _, ok := err.(sow.ErrGHNotInstalled); ok {
        errorMsg = "GitHub CLI not found\n\n" +
            "The 'gh' command is required for GitHub issue integration.\n\n" +
            "To install:\n" +
            "  macOS: brew install gh\n" +
            "  Linux: See https://cli.github.com/"
        fallbackMsg = "Or select 'From branch name' instead."

    } else if _, ok := err.(sow.ErrGHNotAuthenticated); ok {
        errorMsg = "GitHub CLI not authenticated\n\n" +
            "Run the following command to authenticate:\n" +
            "  gh auth login\n\n" +
            "Then try creating your project again."
        fallbackMsg = "Or select 'From branch name' instead."

    } else {
        // Generic GitHub error
        errorMsg = fmt.Sprintf("GitHub CLI error: %v", err)
        fallbackMsg = "Select 'From branch name' to continue without GitHub integration."
    }

    // Show error with fallback option
    fullMessage := errorMsg + "\n\n" + fallbackMsg
    _ = showError(fullMessage)

    // Return to source selection so user can choose "From branch name"
    w.state = StateCreateSource
    return nil
}
```

### 4. Import Requirements

Add necessary imports to `cli/cmd/project/wizard_state.go`:

```go
import (
    // ... existing imports
    sowexec "github.com/jmgilman/sow/cli/internal/exec"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

## Acceptance Criteria

### Functional Requirements

1. **GitHub client initialized**: Wizard struct contains `*sow.GitHub` client initialized with local executor
2. **Validation on issue path**: When user selects "From GitHub issue", `gh` CLI validation runs before proceeding
3. **Not installed error**: When `gh` is not in PATH, shows installation instructions with platform-specific guidance
4. **Not authenticated error**: When `gh` is not authenticated, shows `gh auth login` instructions
5. **Fallback offered**: All errors include suggestion to use "From branch name" path instead
6. **Returns to source selection**: After error, wizard state transitions to `StateCreateSource` (not cancelled)
7. **Success path**: When `gh` is installed and authenticated, validation passes silently and wizard continues

### Test Requirements (TDD Approach)

Write tests **before** implementing the functionality. Create tests in `cli/cmd/project/wizard_state_test.go`:

#### Test 1: GitHub Client Initialization
```go
func TestNewWizard_InitializesGitHubClient(t *testing.T) {
    cmd := &cobra.Command{}
    ctx := testContext(t)

    wizard := NewWizard(cmd, ctx, nil)

    if wizard.github == nil {
        t.Error("expected GitHub client to be initialized, got nil")
    }
}
```

#### Test 2: GitHub CLI Not Installed
```go
func TestHandleIssueSelect_GitHubNotInstalled(t *testing.T) {
    // Create wizard with mock GitHub client that returns ErrGHNotInstalled
    wizard := &Wizard{
        state:  StateIssueSelect,
        ctx:    testContext(t),
        choices: make(map[string]interface{}),
        github: &mockGitHub{ensureErr: sow.ErrGHNotInstalled{}},
    }

    err := wizard.handleIssueSelect()

    // Should not return error (wizard continues)
    if err != nil {
        t.Errorf("expected nil error, got %v", err)
    }

    // Should transition back to source selection
    if wizard.state != StateCreateSource {
        t.Errorf("expected state %s, got %s", StateCreateSource, wizard.state)
    }
}
```

#### Test 3: GitHub CLI Not Authenticated
```go
func TestHandleIssueSelect_GitHubNotAuthenticated(t *testing.T) {
    wizard := &Wizard{
        state:  StateIssueSelect,
        ctx:    testContext(t),
        choices: make(map[string]interface{}),
        github: &mockGitHub{ensureErr: sow.ErrGHNotAuthenticated{}},
    }

    err := wizard.handleIssueSelect()

    if err != nil {
        t.Errorf("expected nil error, got %v", err)
    }

    if wizard.state != StateCreateSource {
        t.Errorf("expected state %s, got %s", StateCreateSource, wizard.state)
    }
}
```

#### Test 4: GitHub CLI Validation Success
```go
func TestHandleIssueSelect_ValidationSuccess(t *testing.T) {
    wizard := &Wizard{
        state:  StateIssueSelect,
        ctx:    testContext(t),
        choices: make(map[string]interface{}),
        github: &mockGitHub{ensureErr: nil}, // No error = success
    }

    err := wizard.handleIssueSelect()

    if err != nil {
        t.Errorf("expected nil error, got %v", err)
    }

    // Should continue to completion (stub behavior for now)
    if wizard.state != StateComplete {
        t.Errorf("expected state %s, got %s", StateComplete, wizard.state)
    }
}
```

#### Mock GitHub Client
```go
// mockGitHub is a test double for GitHub operations
type mockGitHub struct {
    ensureErr error
}

func (m *mockGitHub) Ensure() error {
    return m.ensureErr
}

func (m *mockGitHub) ListIssues(label, state string) ([]sow.Issue, error) {
    return nil, nil // Stub for future tasks
}

func (m *mockGitHub) GetLinkedBranches(number int) ([]sow.LinkedBranch, error) {
    return nil, nil // Stub for future tasks
}

func (m *mockGitHub) CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error) {
    return "", nil // Stub for future tasks
}

func (m *mockGitHub) GetIssue(number int) (*sow.Issue, error) {
    return nil, nil // Stub for future tasks
}
```

### Non-Functional Requirements

- **No breaking changes**: Existing wizard flows (branch name path) continue to work unchanged
- **Clear messaging**: Error messages explain the problem, how to fix it, and offer alternatives
- **User-friendly**: Users are never stuck - always offered a path forward
- **Testable**: Mock GitHub client enables testing without actual `gh` CLI

## Technical Details

### Error Types from GitHub Client

The `cli/internal/sow/github.go` client defines these error types:

- `ErrGHNotInstalled`: Returned when `gh` command not found in PATH
- `ErrGHNotAuthenticated`: Returned when `gh` not authenticated (exit code 1 from `gh auth status`)
- `ErrGHCommand`: Generic command failure with stderr output

### Type Assertions for Error Handling

Use type assertions (not type switches) to identify error types:

```go
if _, ok := err.(sow.ErrGHNotInstalled); ok {
    // Handle not installed
}
```

This approach works because the error types implement the `error` interface directly (not as pointers).

### Wizard State Transitions

```
StateIssueSelect (entry)
  ↓
github.Ensure() called
  ↓
┌─────────────────────┐
│  Success?           │
├─────────────────────┤
│ Yes → Continue      │ (StateComplete for now, later StateIssueSelect continues to listing)
│ No  → Show Error    │ → Return to StateCreateSource
└─────────────────────┘
```

### Why Not Cancel?

Setting `state = StateCancelled` would exit the wizard entirely. Setting `state = StateCreateSource` returns the user to source selection, where they can choose "From branch name" as an alternative. This provides a better user experience.

## Relevant Inputs

### GitHub Client Implementation
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/internal/sow/github.go` - Complete GitHub client with all needed methods
  - Lines 42-53: `ErrGHNotInstalled` and `ErrGHNotAuthenticated` type definitions
  - Lines 105-136: `CheckInstalled()`, `CheckAuthenticated()`, `Ensure()` implementations

### Wizard Foundation
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_state.go` - Current wizard implementation
  - Lines 14-38: `Wizard` struct and state constants
  - Lines 40-49: `NewWizard()` constructor
  - Lines 168-173: Current `handleIssueSelect()` stub to replace

### Helper Functions
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_helpers.go` - Shared helper functions
  - Lines 145-171: `showError()` function for displaying errors

### Executor Pattern
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/internal/exec/executor.go` - Executor interface for command execution
  - Lines 20-45: `Executor` interface definition
  - Lines 47-66: `LocalExecutor` implementation

### Design Specifications
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/knowledge/designs/interactive-wizard-ux-flow.md` - UX flow design
  - Lines 517-535: GitHub CLI missing error message format

- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/project/context/issue-70.md` - GitHub issue specification
  - Lines 160-177: GitHub CLI availability check requirements

## Examples

### Example: Successful Validation Flow
```
User: Selects "From GitHub issue"
Wizard: Transitions to StateIssueSelect
Wizard: Calls github.Ensure()
GitHub: Checks gh installed ✓
GitHub: Checks gh authenticated ✓
GitHub: Returns nil
Wizard: Continues to issue fetching (Task 020)
```

### Example: Not Installed Error Flow
```
User: Selects "From GitHub issue"
Wizard: Transitions to StateIssueSelect
Wizard: Calls github.Ensure()
GitHub: Checks gh installed ✗
GitHub: Returns ErrGHNotInstalled{}
Wizard: Calls handleGitHubError()
Display: Shows error with install instructions
Display: Suggests "From branch name" fallback
Wizard: Transitions to StateCreateSource
User: Can select "From branch name" instead
```

### Example: Not Authenticated Error Flow
```
User: Selects "From GitHub issue"
Wizard: Transitions to StateIssueSelect
Wizard: Calls github.Ensure()
GitHub: Checks gh installed ✓
GitHub: Checks gh authenticated ✗
GitHub: Returns ErrGHNotAuthenticated{}
Wizard: Calls handleGitHubError()
Display: Shows "gh auth login" instructions
Display: Suggests "From branch name" fallback
Wizard: Transitions to StateCreateSource
User: Can run "gh auth login" and try again
```

## Dependencies

### Prerequisites
- Work Unit 001 (wizard foundation) - COMPLETE (assumed)
- Work Unit 002 (branch name path) - COMPLETE (assumed)

### Depends On
- `cli/internal/sow/github.go` - GitHub client implementation (exists)
- `cli/internal/exec/executor.go` - Executor interface (exists)
- `cli/cmd/project/wizard_helpers.go` - `showError()` helper (exists)

### Enables
- Task 020 (issue listing) - This validation must pass before fetching issues
- Task 030 (issue selection) - Depends on validated GitHub access

## Constraints

### Must Not
- **Break existing flows**: Branch name path must continue working if GitHub is unavailable
- **Block users**: Always offer a fallback path (never force GitHub CLI)
- **Expose internals**: Error messages should be user-friendly, not technical stack traces

### Must Do
- **Validate before operations**: Never attempt GitHub operations without validation
- **Provide clear guidance**: Installation/auth instructions must be actionable
- **Maintain idempotency**: Multiple calls to validation should behave identically

### Performance
- **Fast failure**: `gh --version` check is fast (~50ms), shouldn't add noticeable delay
- **No retries**: Don't auto-retry failed validation - let user fix and re-run wizard

## Notes

### Testing Without gh CLI

Tests use mock GitHub client to avoid requiring actual `gh` installation:

```go
mockGH := &mockGitHub{
    ensureErr: sow.ErrGHNotInstalled{},
}
wizard.github = mockGH
```

This allows testing all error paths without installing/uninstalling `gh`.

### Error Message Iteration

The error message wording can be refined based on user feedback. The key is ensuring:
1. Users understand what went wrong
2. Users know how to fix it
3. Users have an alternative path

### Platform-Specific Install Instructions

The install message mentions both macOS (`brew install gh`) and Linux (`https://cli.github.com/`). Consider adding Windows support in future iterations.

### Future Enhancement: Retry Option

Currently, errors return user to source selection. A future enhancement could offer "Retry" option after user has run `gh auth login`, avoiding the need to navigate back through the wizard.
