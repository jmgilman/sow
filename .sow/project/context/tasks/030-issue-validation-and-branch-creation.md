# Task 030: Issue Validation and Branch Creation

## Context

This task implements two critical pieces of the GitHub issue workflow: validating that the selected issue doesn't already have a linked branch, and creating a new linked branch if validation passes. This bridges the gap between issue selection (Task 020) and type selection (Task 040).

The validation prevents users from accidentally creating duplicate projects for the same issue, while the branch creation establishes the GitHub link between the issue and the branch using `gh issue develop`. This link is visible in the GitHub UI and enables better tracking of work.

This task reuses the existing `GetLinkedBranches()` and `CreateLinkedBranch()` methods from the GitHub client, focusing on integrating them into the wizard flow with appropriate error handling.

## Requirements

### 1. Implement Linked Branch Validation

After user selects an issue in `showIssueSelectScreen()` (from Task 020), validate the issue before proceeding. Update the method in `cli/cmd/project/wizard_state.go`:

```go
func (w *Wizard) showIssueSelectScreen() error {
    // ... existing code for issue selection ...

    // Handle cancel
    if selectedIssueNumber == -1 {
        w.state = StateCancelled
        return nil
    }

    // NEW: Validate issue doesn't have linked branch
    linkedBranches, err := w.github.GetLinkedBranches(selectedIssueNumber)
    if err != nil {
        errorMsg := fmt.Sprintf("Failed to check linked branches: %v\n\n" +
            "Please try again or select 'From branch name' instead.")
        _ = showError(errorMsg)
        w.state = StateCreateSource
        return nil
    }

    if len(linkedBranches) > 0 {
        return w.handleAlreadyLinkedError(selectedIssueNumber, linkedBranches[0])
    }

    // NEW: Fetch full issue details
    issue, err := w.github.GetIssue(selectedIssueNumber)
    if err != nil {
        errorMsg := fmt.Sprintf("Failed to get issue details: %v\n\n" +
            "Please try again.")
        _ = showError(errorMsg)
        // Stay in current state to allow retry
        return nil
    }

    // Store issue in choices for next steps
    w.choices["issue"] = issue

    // Proceed to type selection
    w.state = StateTypeSelect

    return nil
}
```

### 2. Implement Already Linked Error Handler

Create new method `handleAlreadyLinkedError()` in `cli/cmd/project/wizard_state.go`:

```go
// handleAlreadyLinkedError displays error when issue has existing linked branch.
// Returns nil to keep wizard running (user can select different issue).
func (w *Wizard) handleAlreadyLinkedError(issueNumber int, branch sow.LinkedBranch) error {
    errorMsg := fmt.Sprintf(
        "Issue #%d already has a linked branch: %s\n\n"+
            "To continue working on this issue:\n"+
            "  Select \"Continue existing project\" from the main menu\n\n"+
            "To work on a different issue:\n"+
            "  Select a different issue from the list",
        issueNumber,
        branch.Name,
    )

    _ = showError(errorMsg)

    // Return to issue select to let user choose different issue
    // Keep issues list in choices so we don't need to fetch again
    return w.showIssueSelectScreen()
}
```

### 3. Enhance Type Selection for Issue Path

Modify `handleTypeSelect()` in `cli/cmd/project/wizard_state.go` to show issue context and route correctly:

```go
func (w *Wizard) handleTypeSelect() error {
    var selectedType string

    // Check if we have issue context
    var contextNote *huh.Note
    issue, hasIssue := w.choices["issue"].(*sow.Issue)
    if hasIssue {
        contextNote = huh.NewNote().
            Description(fmt.Sprintf("Issue: #%d - %s", issue.Number, issue.Title))
    }

    // Build form groups
    groups := []*huh.Group{}

    // Add context note if present
    if contextNote != nil {
        groups = append(groups, huh.NewGroup(contextNote))
    }

    // Add type selection
    groups = append(groups, huh.NewGroup(
        huh.NewSelect[string]().
            Title("What type of project?").
            Options(getTypeOptions()...).
            Value(&selectedType),
    ))

    form := huh.NewForm(groups...)

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return fmt.Errorf("type selection error: %w", err)
    }

    if selectedType == "cancel" {
        w.state = StateCancelled
        return nil
    }

    w.choices["type"] = selectedType

    // Route based on context
    if hasIssue {
        // GitHub issue path: create branch then go to prompt entry
        return w.createLinkedBranch()
    } else {
        // Branch name path: go to name entry
        w.state = StateNameEntry
        return nil
    }
}
```

### 4. Implement Branch Creation

Create new method `createLinkedBranch()` in `cli/cmd/project/wizard_state.go`:

```go
// createLinkedBranch generates a branch name from the issue and creates a linked branch.
func (w *Wizard) createLinkedBranch() error {
    issue, ok := w.choices["issue"].(*sow.Issue)
    if !ok {
        return fmt.Errorf("issue not found in choices")
    }

    projectType, ok := w.choices["type"].(string)
    if !ok {
        return fmt.Errorf("type not found in choices")
    }

    // Generate branch name: <prefix><issue-slug>-<number>
    prefix := getTypePrefix(projectType)
    issueSlug := normalizeName(issue.Title)
    branchName := fmt.Sprintf("%s%s-%d", prefix, issueSlug, issue.Number)

    // Create linked branch via gh issue develop with spinner
    var createdBranch string
    err := withSpinner("Creating linked branch...", func() error {
        var err error
        // Pass checkout=false because we use worktrees, not traditional checkout
        createdBranch, err = w.github.CreateLinkedBranch(issue.Number, branchName, false)
        return err
    })

    if err != nil {
        errorMsg := fmt.Sprintf("Failed to create linked branch: %v\n\n"+
            "This may be a GitHub API issue. Please try again.",
            err)
        _ = showError(errorMsg)
        // Stay in current state to allow retry
        return nil
    }

    // Store branch name and use issue title as project name
    w.choices["branch"] = createdBranch
    w.choices["name"] = issue.Title

    // Proceed to prompt entry
    w.state = StatePromptEntry

    return nil
}
```

### 5. Update Imports

Ensure all necessary imports in `cli/cmd/project/wizard_state.go`:

```go
import (
    "errors"
    "fmt"
    "os"

    "github.com/charmbracelet/huh"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

## Acceptance Criteria

### Functional Requirements

1. **Linked branch check**: After issue selection, calls `github.GetLinkedBranches(issueNumber)`
2. **Already linked error**: If linked branches exist, shows error with branch name and suggestions
3. **Retry on error**: Error returns user to issue list (not source selection) to choose different issue
4. **Full issue fetch**: Calls `github.GetIssue(number)` to get complete issue details including title
5. **Issue context display**: Type selection screen shows "Issue: #123 - Title" when coming from issue path
6. **Branch name generation**: Creates branch as `<prefix><slug>-<number>` (e.g., "feat/add-jwt-auth-123")
7. **Linked branch creation**: Calls `github.CreateLinkedBranch(issue, branch, false)` with checkout=false
8. **Spinner during creation**: Shows loading spinner with "Creating linked branch..."
9. **State storage**: Stores branch name and issue title in choices for finalization
10. **Correct routing**: Issue path skips name entry, goes directly to prompt entry after branch creation

### Test Requirements (TDD Approach)

Write tests before implementing. Add to `cli/cmd/project/wizard_state_test.go`:

#### Test 1: Issue With Linked Branch
```go
func TestShowIssueSelectScreen_IssueAlreadyLinked(t *testing.T) {
    wizard := &Wizard{
        state: StateIssueSelect,
        ctx:   testContext(t),
        choices: map[string]interface{}{
            "issues": []sow.Issue{
                {Number: 123, Title: "Test Issue", State: "open"},
            },
        },
        github: &mockGitHub{
            getLinkedBranchesResult: []sow.LinkedBranch{
                {Name: "feat/existing-branch", URL: "https://..."},
            },
        },
    }

    // Simulate user selecting issue 123
    wizard.choices["selectedIssueNumber"] = 123

    err := wizard.showIssueSelectScreen()

    // Should not return error (wizard continues)
    if err != nil {
        t.Errorf("expected nil error, got %v", err)
    }

    // Should loop back to issue selection, NOT cancel or change state
    // (In practice, showError() is called and user sees list again)
}
```

#### Test 2: Issue Without Linked Branch
```go
func TestShowIssueSelectScreen_NoLinkedBranch(t *testing.T) {
    wizard := &Wizard{
        state: StateIssueSelect,
        ctx:   testContext(t),
        choices: map[string]interface{}{
            "issues": []sow.Issue{
                {Number: 123, Title: "Test Issue", State: "open"},
            },
        },
        github: &mockGitHub{
            getLinkedBranchesResult: []sow.LinkedBranch{}, // No linked branches
            getIssueResult: &sow.Issue{
                Number: 123,
                Title:  "Test Issue",
                Body:   "Issue description",
                State:  "open",
                URL:    "https://github.com/test/repo/issues/123",
            },
        },
    }

    wizard.choices["selectedIssueNumber"] = 123

    err := wizard.showIssueSelectScreen()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Should transition to type selection
    if wizard.state != StateTypeSelect {
        t.Errorf("expected state %s, got %s", StateTypeSelect, wizard.state)
    }

    // Should store issue
    issue, ok := wizard.choices["issue"].(*sow.Issue)
    if !ok {
        t.Fatal("issue not stored in choices")
    }

    if issue.Number != 123 {
        t.Errorf("expected issue 123, got %d", issue.Number)
    }
}
```

#### Test 3: Branch Name Generation
```go
func TestCreateLinkedBranch_BranchNameGeneration(t *testing.T) {
    tests := []struct {
        issueTitle  string
        issueNumber int
        projectType string
        expected    string
    }{
        {"Add JWT authentication", 123, "standard", "feat/add-jwt-authentication-123"},
        {"Refactor Database Schema", 456, "standard", "feat/refactor-database-schema-456"},
        {"Web Based Agents", 789, "exploration", "explore/web-based-agents-789"},
        {"Special!@#$ Chars", 111, "design", "design/specialchars-111"},
    }

    for _, tt := range tests {
        t.Run(tt.issueTitle, func(t *testing.T) {
            wizard := &Wizard{
                state: StateTypeSelect,
                ctx:   testContext(t),
                choices: map[string]interface{}{
                    "issue": &sow.Issue{
                        Number: tt.issueNumber,
                        Title:  tt.issueTitle,
                    },
                    "type": tt.projectType,
                },
                github: &mockGitHub{
                    createLinkedBranchResult: tt.expected, // Mock returns expected name
                },
            }

            err := wizard.createLinkedBranch()
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            // Verify branch name stored
            branchName, ok := wizard.choices["branch"].(string)
            if !ok {
                t.Fatal("branch name not stored")
            }

            if branchName != tt.expected {
                t.Errorf("expected branch %q, got %q", tt.expected, branchName)
            }
        })
    }
}
```

#### Test 4: Type Selection Routing
```go
func TestHandleTypeSelect_RoutingWithIssue(t *testing.T) {
    wizard := &Wizard{
        state: StateTypeSelect,
        ctx:   testContext(t),
        choices: map[string]interface{}{
            "issue": &sow.Issue{Number: 123, Title: "Test"},
        },
        github: &mockGitHub{
            createLinkedBranchResult: "feat/test-123",
        },
    }

    // Simulate user selecting "standard" type
    wizard.choices["type"] = "standard"

    err := wizard.handleTypeSelect()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Should create branch and go to prompt entry (not name entry)
    if wizard.state != StatePromptEntry {
        t.Errorf("expected state %s, got %s", StatePromptEntry, wizard.state)
    }
}

func TestHandleTypeSelect_RoutingWithoutIssue(t *testing.T) {
    wizard := &Wizard{
        state: StateTypeSelect,
        ctx:   testContext(t),
        choices: make(map[string]interface{}),
    }

    wizard.choices["type"] = "standard"

    err := wizard.handleTypeSelect()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Should go to name entry (branch name path)
    if wizard.state != StateNameEntry {
        t.Errorf("expected state %s, got %s", StateNameEntry, wizard.state)
    }
}
```

#### Update Mock GitHub Client

Add new fields to `mockGitHub`:

```go
type mockGitHub struct {
    ensureErr                error
    listIssuesResult         []sow.Issue
    listIssuesErr            error
    getLinkedBranchesResult  []sow.LinkedBranch  // NEW
    getLinkedBranchesErr     error               // NEW
    getIssueResult           *sow.Issue          // NEW
    getIssueErr              error               // NEW
    createLinkedBranchResult string              // NEW
    createLinkedBranchErr    error               // NEW
}

func (m *mockGitHub) GetLinkedBranches(number int) ([]sow.LinkedBranch, error) {
    if m.getLinkedBranchesErr != nil {
        return nil, m.getLinkedBranchesErr
    }
    return m.getLinkedBranchesResult, nil
}

func (m *mockGitHub) GetIssue(number int) (*sow.Issue, error) {
    if m.getIssueErr != nil {
        return nil, m.getIssueErr
    }
    return m.getIssueResult, nil
}

func (m *mockGitHub) CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error) {
    if m.createLinkedBranchErr != nil {
        return "", m.createLinkedBranchErr
    }
    // Return provided branch name or mock result
    if branchName != "" {
        return branchName, nil
    }
    return m.createLinkedBranchResult, nil
}
```

### Non-Functional Requirements

- **Prevent duplicates**: Never allow creating second project for same issue
- **Clear errors**: Error messages explain why issue can't be used and what to do instead
- **Idempotent validation**: Multiple checks for same issue should give same result
- **Preserve context**: Issue data stored in choices for use in later screens

## Technical Details

### Linked Branch Detection

The `GetLinkedBranches()` method uses `gh issue develop --list <number>`:

```bash
$ gh issue develop --list 123
BRANCH              URL
feat/add-auth-123   https://github.com/owner/repo/tree/feat/add-auth-123
```

Returns empty array if no branches linked.

### Issue Develop Command

The `CreateLinkedBranch()` method uses `gh issue develop <number> --name <branch>`:

```bash
$ gh issue develop 123 --name feat/add-jwt-auth-123
Created branch feat/add-jwt-auth-123 based on main
```

This single command:
1. Creates the branch
2. Links it to the issue
3. Updates GitHub's issue-branch tracking

**Critical**: Pass `checkout=false` because sow uses worktrees, not traditional checkout.

### Branch Name Normalization

Branch names follow the pattern: `<prefix><normalized-title>-<number>`

Normalization (using existing `normalizeName()` function):
1. Convert to lowercase
2. Replace spaces with hyphens
3. Remove invalid characters (keep only a-z, 0-9, -, _)
4. Collapse multiple hyphens
5. Remove leading/trailing hyphens

Examples:
- "Add JWT authentication" → "add-jwt-authentication"
- "Fix Bug #42" → "fix-bug-42"
- "Web-Based Agents" → "web-based-agents"

### Type Selection Form Enhancement

When issue context exists, the form includes two groups:
1. Context note (read-only display)
2. Type selection (interactive)

This is achieved by conditionally building the groups array before creating the form.

### Error Recovery Strategies

Different errors have different recovery paths:

| Error Type | Recovery Action |
|------------|-----------------|
| Already linked | Return to issue list (choose different issue) |
| API failure | Stay in current state (allow retry) |
| Network timeout | Stay in current state (allow retry) |
| Branch creation fails | Stay in current state (allow retry) |

## Relevant Inputs

### GitHub Client Implementation
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/internal/sow/github.go`
  - Lines 87-91: `LinkedBranch` struct definition
  - Lines 201-254: `GetLinkedBranches()` implementation
  - Lines 175-199: `GetIssue()` implementation
  - Lines 256-323: `CreateLinkedBranch()` implementation
  - Lines 438-465: `toKebabCase()` helper (similar to our normalization)

### Wizard State Machine
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_state.go`
  - Lines 175-205: `handleTypeSelect()` method to enhance
  - Lines 168-173: Current `handleIssueSelect()` stub

### Helper Functions
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_helpers.go`
  - Lines 42-91: `normalizeName()` function for branch name generation
  - Lines 93-106: `getTypePrefix()` function for branch prefixes
  - Lines 190-201: `withSpinner()` for loading indicators

### Design Specifications
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/knowledge/designs/interactive-wizard-ux-flow.md`
  - Lines 109-141: Type selection with issue context
  - Lines 445-458: Already linked error message format

- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/project/context/issue-70.md`
  - Lines 224-262: Issue validation logic specification
  - Lines 309-338: Branch creation specification
  - Lines 744-759: Branch naming convention

## Examples

### Example: Issue Already Linked
```
User: Selects issue #123
Wizard: Calls GetLinkedBranches(123)
GitHub: Returns [feat/add-auth-123]
Display: Error message:
  "Issue #123 already has a linked branch: feat/add-auth-123

  To continue working on this issue:
    Select "Continue existing project" from the main menu

  To work on a different issue:
    Select a different issue from the list"
User: Presses Enter
Wizard: Returns to issue selection screen
User: Can select different issue
```

### Example: Successful Validation and Branch Creation
```
User: Selects issue #124 "Refactor database schema"
Wizard: Calls GetLinkedBranches(124)
GitHub: Returns [] (empty)
Wizard: Calls GetIssue(124)
GitHub: Returns full issue details
Wizard: Stores issue in choices
Wizard: Transitions to StateTypeSelect
Display: Type selection with context:
  "Issue: #124 - Refactor database schema

  What type of project?
  ○ Feature work and bug fixes
  ○ Research and investigation
  ..."
User: Selects "Feature work and bug fixes"
Wizard: Generates branch: "feat/refactor-database-schema-124"
Display: [Spinner] "Creating linked branch..."
Wizard: Calls CreateLinkedBranch(124, "feat/refactor-database-schema-124", false)
GitHub: Creates branch and links to issue ✓
Wizard: Stores branch name and issue title
Wizard: Transitions to StatePromptEntry
```

### Example: Branch Creation Failure
```
User: Selects issue, chooses type
Wizard: Generates branch name
Display: [Spinner] "Creating linked branch..."
GitHub: API rate limit exceeded
Display: Error message:
  "Failed to create linked branch: rate limit exceeded

  This may be a GitHub API issue. Please try again."
User: Presses Enter
Wizard: Stays in StateTypeSelect
User: Can retry or cancel
```

## Dependencies

### Prerequisites
- Task 010 (GitHub CLI validation) - MUST be complete
- Task 020 (issue listing) - MUST be complete
- Work Unit 001 (wizard foundation) - COMPLETE (assumed)
- Work Unit 002 (branch name path) - COMPLETE (assumed)

### Depends On
- `github.GetLinkedBranches()` for validation
- `github.GetIssue()` for full issue details
- `github.CreateLinkedBranch()` for branch creation
- `normalizeName()` for branch name generation
- `withSpinner()` for loading indicator

### Enables
- Task 040 (prompt entry with issue context) - Needs branch name and issue data
- Task 050 (finalization with issue metadata) - Needs complete issue information

## Constraints

### Must Not
- **Create duplicate projects**: Never bypass linked branch check
- **Checkout branches**: Always pass `checkout=false` to CreateLinkedBranch
- **Modify existing branches**: Only create new branches, never modify existing
- **Hard-code prefixes**: Use `getTypePrefix()` for all branch prefix logic

### Must Do
- **Validate before create**: Always check linked branches before attempting creation
- **Store full issue**: Keep complete issue object for finalization
- **Show context**: Display issue information in all subsequent screens
- **Handle all errors**: Network, API, and validation errors gracefully

### Performance
- **Fast validation**: `GetLinkedBranches` is fast (~200ms)
- **Acceptable creation**: Branch creation typically <1 second
- **Spinner required**: Always show spinner for network operations

## Notes

### Worktree vs Checkout

Traditional git workflow: `git checkout feat/xyz` switches working directory

Worktree workflow: `git worktree add .sow/worktrees/feat/xyz feat/xyz` creates separate directory

**Why checkout=false?**
- We create worktree separately using `sow.EnsureWorktree()`
- Checking out would fail if branch already checked out in worktree
- Worktrees allow multiple branches active simultaneously

### Issue Number in Branch Name

Including the issue number ensures unique branch names even if multiple issues have similar titles:

- Issue #123 "Add Auth" → `feat/add-auth-123`
- Issue #456 "Add Auth" → `feat/add-auth-456`

The number also makes it easy to find the issue from the branch name.

### GitHub UI Integration

After `gh issue develop` creates the link:
- Issue page shows linked branch in sidebar
- Branch page shows linked issue
- Closing PR can automatically close issue

This provides better traceability for work.

### Retry Logic

For transient errors (network timeout, API rate limit), we stay in the current state rather than returning to source selection. This allows users to retry immediately rather than navigating back through the wizard.

For permanent errors (issue already linked), we return to issue selection so users can choose a different issue.

### Future Enhancements

Possible improvements:
- Allow unlinking existing branch and creating new one
- Support for draft issues (currently only open/closed)
- Custom branch name override (currently auto-generated)
- Display existing linked branch details before error
