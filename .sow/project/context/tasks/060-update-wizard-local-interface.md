# Task 060: Update Wizard Local Interface Usage

## Context

This task is part of a refactoring project to extract a GitHubClient interface from the existing GitHub CLI implementation. All previous tasks created the shared infrastructure. Now we need to update the project wizard to use the shared interface instead of its local duplicate.

**Project Goal**: Extract GitHubClient interface from existing GitHub CLI implementation to enable dual client support (CLI + API) while maintaining backward compatibility.

**Previous Tasks**:
- Task 010: Created shared GitHubClient interface
- Task 020: Renamed GitHub to GitHubCLI
- Task 030: Implemented new methods
- Task 040: Created factory function
- Task 050: Created MockGitHub

**This Task's Role**: Remove the duplicate GitHubClient interface from wizard_state.go and use the shared interface from sow package. This demonstrates the benefit of the refactoring - one canonical interface used everywhere.

**Why This Matters**: The wizard currently defines its own local GitHubClient interface (lines 35-45). This is exactly the problem we're solving - duplicated interface definitions that can drift apart. By using the shared interface, we ensure consistency and enable future API client support.

## Requirements

Update `cli/cmd/project/wizard_state.go`:

**1. Remove Local Interface (lines 35-45)**
```go
// DELETE THIS:
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

**2. Update Wizard Struct (line 54)**
```go
// CHANGE FROM:
github GitHubClient // GitHub client for issue operations

// TO:
github sow.GitHubClient // GitHub client for issue operations
```

**3. Constructor Remains Unchanged (lines 59-70)**
The constructor already creates `sow.NewGitHub(ghExec)` which will work since NewGitHub still exists (deprecated wrapper from Task 020).

**4. Optional: Consider Using Factory**
The constructor could optionally be updated to use the factory:
```go
// CURRENT:
ghExec := sowexec.NewLocal("gh")
github: sow.NewGitHub(ghExec),

// OPTIONAL FUTURE:
client, err := sow.NewGitHubClient()
if err != nil {
    return nil, err  // Would need signature change
}
github: client,
```

However, this is NOT required for this task. The deprecated NewGitHub wrapper is sufficient. Factory usage can be a future enhancement.

## Acceptance Criteria

- [ ] Local GitHubClient interface removed from wizard_state.go (lines 35-45)
- [ ] Wizard struct field type updated to sow.GitHubClient (line 54)
- [ ] Wizard constructor still works (uses deprecated NewGitHub wrapper)
- [ ] No changes to wizard behavior or logic
- [ ] All wizard tests pass (if any exist)
- [ ] Package compiles: `go build ./cli/cmd/project`
- [ ] Integration test: wizard can still create projects from issues
- [ ] No other code depends on the local GitHubClient interface

## Technical Details

**File Changes:**

Only one file needs modification: `cli/cmd/project/wizard_state.go`

**Line-by-Line Changes:**

Lines 35-45: **DELETE**
```go
// GitHubClient defines the interface for GitHub operations used by the wizard.
// This interface allows for easy mocking in tests.
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

Line 54: **UPDATE**
```go
// BEFORE:
github GitHubClient // GitHub client for issue operations

// AFTER:
github sow.GitHubClient // GitHub client for issue operations
```

Lines 59-70: **NO CHANGE** (constructor still works)

**Why NewGitHub Still Works:**

Task 020 created a deprecated wrapper:
```go
// NewGitHub creates a GitHub CLI client.
// Deprecated: Use NewGitHubCLI() for explicit CLI client, or NewGitHubClient() for auto-detection.
func NewGitHub(executor exec.Executor) *GitHubCLI {
    return NewGitHubCLI(executor)
}
```

Since `*GitHubCLI` implements `GitHubClient`, the assignment works:
```go
var github sow.GitHubClient = sow.NewGitHub(ghExec)  // OK
```

**Interface Compatibility:**

The local interface had these methods:
- Ensure() ✗ (not in GitHubClient)
- CheckInstalled() ✗ (not in GitHubClient)
- CheckAuthenticated() ✗ (not in GitHubClient)
- ListIssues() ✓
- GetLinkedBranches() ✓
- CreateLinkedBranch() ✓
- GetIssue() ✓

The wizard code doesn't actually call Ensure(), CheckInstalled(), or CheckAuthenticated() directly - it only uses the methods that exist in GitHubClient. So the switch is safe.

**Testing Approach:**

1. **Manual verification**:
   - Compile the code: `go build ./cli/cmd/project`
   - Run wizard: `sow project new` (if safe in test environment)

2. **Check for test files**:
   - Look for `wizard_state_test.go` or similar
   - If tests exist, run them: `go test ./cli/cmd/project`
   - Tests should pass unchanged

3. **Search for local interface usage**:
   ```bash
   grep -r "type GitHubClient interface" cli/cmd/project/
   ```
   Should find no results after deletion.

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/.sow/project/context/issue-90.md` - Complete project requirements (mentions wizard at lines 53-54, 74)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/cmd/project/wizard_state.go` - File to update (lines 35-45 for local interface, line 54 for field type, lines 59-70 for constructor)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_client.go` - Shared interface to use (created in Task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_cli.go` - Implementation the wizard uses (created in Task 020)

## Examples

**Before (wizard_state.go):**
```go
package project

import (
    // ...
)

// GitHubClient defines the interface for GitHub operations used by the wizard.
// This interface allows for easy mocking in tests.
type GitHubClient interface {
    Ensure() error
    CheckInstalled() error
    CheckAuthenticated() error
    ListIssues(label, state string) ([]sow.Issue, error)
    GetLinkedBranches(number int) ([]sow.LinkedBranch, error)
    CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)
    GetIssue(number int) (*sow.Issue, error)
}

type Wizard struct {
    state       WizardState
    ctx         *sow.Context
    choices     map[string]interface{}
    claudeFlags []string
    cmd         *cobra.Command
    github      GitHubClient // GitHub client for issue operations
    testMode    bool
}
```

**After (wizard_state.go):**
```go
package project

import (
    // ...
)

// Local interface removed - using shared sow.GitHubClient

type Wizard struct {
    state       WizardState
    ctx         *sow.Context
    choices     map[string]interface{}
    claudeFlags []string
    cmd         *cobra.Command
    github      sow.GitHubClient // GitHub client for issue operations
    testMode    bool
}
```

**Wizard Usage (unchanged):**
```go
func (w *Wizard) someMethod() error {
    // These methods exist in sow.GitHubClient
    issues, err := w.github.ListIssues("sow", "open")
    if err != nil {
        return err
    }

    branches, err := w.github.GetLinkedBranches(issueNumber)
    // ...
}
```

**Future: Using MockGitHub in Tests**
```go
func TestWizard_SomeFlow(t *testing.T) {
    mock := &sow.MockGitHub{
        ListIssuesFunc: func(label, state string) ([]sow.Issue, error) {
            return []sow.Issue{{Number: 1, Title: "Test"}}, nil
        },
    }

    wizard := &Wizard{
        github: mock,
        // ... other fields
    }

    // Test wizard logic
}
```

## Dependencies

**Depends On:**
- Task 010 (Define GitHubClient Interface) - Shared interface must exist
- Task 020 (Rename Implementation to GitHubCLI) - Deprecated NewGitHub must exist

**Reason**: Wizard needs the shared interface and the backward-compatible constructor.

## Constraints

**DO NOT:**
- Change wizard behavior or logic
- Modify wizard constructor signature
- Break existing wizard tests
- Add complexity or refactoring beyond interface swap
- Update to use factory (optional future enhancement)

**DO:**
- Keep changes minimal and focused
- Only change interface definition location
- Preserve all existing functionality
- Maintain backward compatibility
- Test that wizard still works

**Testing Requirements:**
- Run existing tests (if any)
- Verify package compiles
- Consider manual testing if safe
- Check for any code depending on local interface

**Search for Dependencies:**

Before deleting the local interface, search for any code that might reference it:
```bash
# In cli/cmd/project/ directory:
grep -r "GitHubClient" .
```

If you find code outside wizard_state.go using the local interface, either:
1. Update it to use sow.GitHubClient
2. Or note it as a finding if it's complex

**Backward Compatibility:**
- Constructor still uses deprecated NewGitHub wrapper
- Future PR can update to use factory if desired
- No breaking changes to wizard API
- Wizard tests (if any) should pass unchanged
