# Task 020: Issue Listing Screen with Spinner

## Context

This task implements the GitHub issue listing screen, which displays issues labeled with 'sow' in a selectable list. This is the second step in the GitHub issue workflow, executed after GitHub CLI validation (Task 010) succeeds.

The issue listing screen fetches issues from GitHub using the existing `github.ListIssues()` method, displays them in a huh Select prompt with format "#123: Issue Title", and handles edge cases like empty lists gracefully. A loading spinner provides feedback during the network fetch operation.

This task builds on the wizard foundation (Work Unit 001) and follows the patterns established in the design documents for async operations with spinners and dynamic option lists.

## Requirements

### 1. Implement Issue Fetching with Spinner

Update `handleIssueSelect()` in `cli/cmd/project/wizard_state.go` to fetch issues with a loading indicator:

```go
func (w *Wizard) handleIssueSelect() error {
    // GitHub CLI validation (from Task 010)
    if err := w.github.Ensure(); err != nil {
        return w.handleGitHubError(err)
    }

    // Fetch issues with 'sow' label using spinner
    var issues []sow.Issue
    var fetchErr error

    err := withSpinner("Fetching issues from GitHub...", func() error {
        issues, fetchErr = w.github.ListIssues("sow", "open")
        return fetchErr
    })

    if err != nil {
        errorMsg := fmt.Sprintf("Failed to fetch issues: %v\n\n" +
            "This may be a network issue or a GitHub API problem.\n" +
            "Please try again or select 'From branch name' instead.")
        _ = showError(errorMsg)
        w.state = StateCreateSource
        return nil
    }

    // Handle empty issue list
    if len(issues) == 0 {
        errorMsg := "No issues found with 'sow' label\n\n" +
            "To use GitHub issue integration:\n" +
            "  1. Create an issue in your repository\n" +
            "  2. Add the 'sow' label to the issue\n" +
            "  3. Try again\n\n" +
            "Or select 'From branch name' to continue without an issue."
        _ = showError(errorMsg)
        w.state = StateCreateSource
        return nil
    }

    // Store issues in choices for next step (Task 030)
    w.choices["issues"] = issues

    // Proceed to issue selection (next screen)
    return w.showIssueSelectScreen()
}
```

### 2. Implement Issue Selection Screen

Create new method `showIssueSelectScreen()` in `cli/cmd/project/wizard_state.go`:

```go
// showIssueSelectScreen displays the issue selection prompt.
// Issues are retrieved from w.choices["issues"] (set by handleIssueSelect).
func (w *Wizard) showIssueSelectScreen() error {
    issues, ok := w.choices["issues"].([]sow.Issue)
    if !ok {
        return fmt.Errorf("issues not found in choices")
    }

    var selectedIssueNumber int

    // Build select options
    options := make([]huh.Option[int], 0, len(issues)+1)
    for _, issue := range issues {
        label := fmt.Sprintf("#%d: %s", issue.Number, issue.Title)
        options = append(options, huh.NewOption(label, issue.Number))
    }

    // Add cancel option
    options = append(options, huh.NewOption("Cancel", -1))

    // Create select form
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[int]().
                Title("Select an issue (filtered by 'sow' label):").
                Options(options...).
                Value(&selectedIssueNumber),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return fmt.Errorf("issue selection error: %w", err)
    }

    // Handle cancel
    if selectedIssueNumber == -1 {
        w.state = StateCancelled
        return nil
    }

    // Store selected issue number for next step (Task 030)
    w.choices["selectedIssueNumber"] = selectedIssueNumber

    // Next: Validate issue doesn't have linked branch (Task 030)
    // For now, just mark complete
    w.state = StateComplete

    return nil
}
```

### 3. Update Imports

Add required imports to `cli/cmd/project/wizard_state.go`:

```go
import (
    // ... existing imports
    "errors"
    "fmt"

    "github.com/charmbracelet/huh"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

## Acceptance Criteria

### Functional Requirements

1. **Spinner displays**: When fetching issues, loading spinner appears with message "Fetching issues from GitHub..."
2. **Issues fetched**: Calls `github.ListIssues("sow", "open")` to get open issues with 'sow' label
3. **Network errors handled**: If fetch fails, shows error with suggestion to retry or use branch name path
4. **Empty list handled**: If no issues found, shows helpful message explaining how to create labeled issues
5. **Issue format**: Each issue displayed as "#123: Issue Title" with number before colon
6. **Cancel option**: List includes "Cancel" option that exits wizard gracefully
7. **Selection stored**: Selected issue number stored in `w.choices["selectedIssueNumber"]` for next task
8. **Graceful degradation**: All errors return user to source selection (not wizard cancellation)

### Test Requirements (TDD Approach)

Write tests before implementing. Add to `cli/cmd/project/wizard_state_test.go`:

#### Test 1: Successful Issue Fetch and Display
```go
func TestHandleIssueSelect_SuccessfulFetch(t *testing.T) {
    mockIssues := []sow.Issue{
        {Number: 123, Title: "Add JWT authentication", State: "open"},
        {Number: 124, Title: "Refactor schema", State: "open"},
    }

    wizard := &Wizard{
        state:  StateIssueSelect,
        ctx:    testContext(t),
        choices: make(map[string]interface{}),
        github: &mockGitHub{listIssuesResult: mockIssues},
    }

    // This will run handleIssueSelect which should store issues
    err := wizard.handleState()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Verify issues stored
    storedIssues, ok := wizard.choices["issues"].([]sow.Issue)
    if !ok {
        t.Fatal("issues not stored in choices")
    }

    if len(storedIssues) != 2 {
        t.Errorf("expected 2 issues, got %d", len(storedIssues))
    }
}
```

#### Test 2: Empty Issue List
```go
func TestHandleIssueSelect_EmptyList(t *testing.T) {
    wizard := &Wizard{
        state:  StateIssueSelect,
        ctx:    testContext(t),
        choices: make(map[string]interface{}),
        github: &mockGitHub{listIssuesResult: []sow.Issue{}}, // Empty list
    }

    err := wizard.handleIssueSelect()
    if err != nil {
        t.Errorf("expected nil error, got %v", err)
    }

    // Should return to source selection
    if wizard.state != StateCreateSource {
        t.Errorf("expected state %s, got %s", StateCreateSource, wizard.state)
    }
}
```

#### Test 3: Network Error During Fetch
```go
func TestHandleIssueSelect_FetchError(t *testing.T) {
    wizard := &Wizard{
        state:  StateIssueSelect,
        ctx:    testContext(t),
        choices: make(map[string]interface{}),
        github: &mockGitHub{
            listIssuesErr: fmt.Errorf("network timeout"),
        },
    }

    err := wizard.handleIssueSelect()
    if err != nil {
        t.Errorf("expected nil error (wizard continues), got %v", err)
    }

    // Should return to source selection
    if wizard.state != StateCreateSource {
        t.Errorf("expected state %s, got %s", StateCreateSource, wizard.state)
    }
}
```

#### Test 4: Issue Selection Options Format
```go
func TestShowIssueSelectScreen_OptionFormat(t *testing.T) {
    issues := []sow.Issue{
        {Number: 123, Title: "Add JWT authentication", State: "open"},
        {Number: 456, Title: "Refactor database schema", State: "open"},
    }

    wizard := &Wizard{
        state:  StateIssueSelect,
        ctx:    testContext(t),
        choices: map[string]interface{}{
            "issues": issues,
        },
        github: &mockGitHub{},
    }

    // Test that options are built correctly
    // This test would need to capture the form options
    // For now, verify method doesn't crash
    // Real verification would happen in integration test

    // Note: Full form testing requires either integration test
    // or extracting option building to separate testable function
}
```

#### Update Mock GitHub Client

Update the `mockGitHub` struct from Task 010:

```go
type mockGitHub struct {
    ensureErr        error
    listIssuesResult []sow.Issue  // NEW
    listIssuesErr    error         // NEW
    // ... other fields from previous tasks
}

func (m *mockGitHub) ListIssues(label, state string) ([]sow.Issue, error) {
    if m.listIssuesErr != nil {
        return nil, m.listIssuesErr
    }
    return m.listIssuesResult, nil
}
```

### Integration Test

Test the complete flow from validation through selection:

```go
func TestIssueWorkflow_ValidationToSelection(t *testing.T) {
    mockIssues := []sow.Issue{
        {Number: 123, Title: "Test Issue", State: "open", URL: "https://github.com/test/repo/issues/123"},
    }

    wizard := &Wizard{
        state:  StateIssueSelect,
        ctx:    testContext(t),
        choices: make(map[string]interface{}),
        github: &mockGitHub{
            ensureErr:        nil, // Validation succeeds
            listIssuesResult: mockIssues,
            listIssuesErr:    nil,
        },
    }

    // Run issue selection handler
    err := wizard.handleIssueSelect()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Verify issues stored for display
    storedIssues, ok := wizard.choices["issues"].([]sow.Issue)
    if !ok || len(storedIssues) != 1 {
        t.Errorf("expected 1 issue stored, got %d", len(storedIssues))
    }
}
```

### Non-Functional Requirements

- **Responsive**: Spinner provides feedback during potentially slow network operations
- **Helpful errors**: All error messages explain problem and suggest solutions
- **Consistent format**: Issue display format matches design specification
- **No data loss**: Fetched issues stored in choices map for reuse if needed

## Technical Details

### Issue Data Structure

From `cli/internal/sow/github.go`:

```go
type Issue struct {
    Number int    `json:"number"`
    Title  string `json:"title"`
    Body   string `json:"body"`
    State  string `json:"state"`
    URL    string `json:"url"`
    Labels []struct {
        Name string `json:"name"`
    } `json:"labels"`
}
```

### ListIssues Method Signature

```go
func (g *GitHub) ListIssues(label, state string) ([]Issue, error)
```

Parameters:
- `label`: Filter by label (use "sow")
- `state`: Filter by state ("open", "closed", or "all")

Returns up to 1000 issues matching criteria.

### Spinner Helper Function

Use existing `withSpinner()` helper from `cli/cmd/project/wizard_helpers.go`:

```go
func withSpinner(title string, action func() error) error {
    var err error

    _ = spinner.New().
        Title(title).
        Action(func() {
            err = action()
        }).
        Run()

    return err
}
```

The spinner:
- Displays animated loading indicator
- Shows provided title message
- Executes action in background
- Returns action's error result

### Select Options with Generic Type

The huh Select prompt uses generic types:

```go
huh.NewSelect[int]().
    Options(
        huh.NewOption("#123: Title", 123),  // Display label, value
        huh.NewOption("#124: Title", 124),
        huh.NewOption("Cancel", -1),
    ).
    Value(&selectedIssueNumber)  // Binds to int variable
```

Using `int` type for issue numbers allows direct storage and comparison.

### Error Recovery Pattern

All errors in this task follow the same pattern:
1. Display helpful error message
2. Explain how to fix (or alternative)
3. Return to `StateCreateSource` (not cancel)
4. Return `nil` (wizard continues, not fatal error)

This ensures users always have a path forward.

## Relevant Inputs

### GitHub Client Implementation
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/internal/sow/github.go`
  - Lines 76-101: `Issue` struct definition
  - Lines 140-173: `ListIssues()` method implementation
  - Lines 56-67: `ErrGHCommand` error type for command failures

### Wizard State Machine
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_state.go`
  - Lines 31-38: `Wizard` struct with choices map
  - Lines 66-88: `handleState()` dispatcher pattern
  - Lines 168-173: Current `handleIssueSelect()` stub to enhance

### Helper Functions
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_helpers.go`
  - Lines 190-201: `withSpinner()` function for async operations
  - Lines 145-171: `showError()` function for error display

### Design Specifications
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/knowledge/designs/interactive-wizard-ux-flow.md`
  - Lines 84-108: Issue selection screen design
  - Lines 360-421: Loading indicator patterns

- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/knowledge/designs/interactive-wizard-technical-implementation.md`
  - Lines 280-335: Select prompt example with dynamic options
  - Lines 860-881: Spinner integration patterns

- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/project/context/issue-70.md`
  - Lines 179-223: Issue listing screen specification
  - Lines 789-794: Loading spinner requirements

## Examples

### Example: Successful Issue Fetch
```
User: Selects "From GitHub issue"
Wizard: Validates GitHub CLI ✓
Display: [Spinner] "Fetching issues from GitHub..."
GitHub: Returns 3 issues with 'sow' label
Display: Select prompt with:
  ○ #123: Add JWT authentication
  ○ #124: Refactor database schema
  ○ #125: Implement rate limiting
  ○ Cancel
User: Selects #123
Wizard: Stores selectedIssueNumber = 123
Wizard: Proceeds to validation (Task 030)
```

### Example: Empty Issue List
```
User: Selects "From GitHub issue"
Wizard: Validates GitHub CLI ✓
Display: [Spinner] "Fetching issues from GitHub..."
GitHub: Returns empty list []
Display: Error message:
  "No issues found with 'sow' label

  To use GitHub issue integration:
    1. Create an issue in your repository
    2. Add the 'sow' label to the issue
    3. Try again

  Or select 'From branch name' to continue without an issue."
User: Presses Enter
Wizard: Returns to StateCreateSource
User: Can select "From branch name" instead
```

### Example: Network Timeout
```
User: Selects "From GitHub issue"
Wizard: Validates GitHub CLI ✓
Display: [Spinner] "Fetching issues from GitHub..."
GitHub: Network timeout after 30s
Display: Error message:
  "Failed to fetch issues: network timeout

  This may be a network issue or a GitHub API problem.
  Please try again or select 'From branch name' instead."
User: Presses Enter
Wizard: Returns to StateCreateSource
User: Can try again or select alternative
```

## Dependencies

### Prerequisites
- Task 010 (GitHub CLI validation) - MUST be complete
- Work Unit 001 (wizard foundation) - COMPLETE (assumed)
- `cli/internal/sow/github.go` - GitHub client (exists)

### Depends On
- `github.ListIssues()` method for fetching issues
- `withSpinner()` helper for loading indicator
- `showError()` helper for error display

### Enables
- Task 030 (issue validation) - Needs selected issue number
- Task 040 (type selection with issue context) - Needs issue data

## Constraints

### Must Not
- **Block indefinitely**: Network operations must timeout (handled by `gh` CLI)
- **Show too many issues**: If >50 issues, consider pagination (future enhancement)
- **Expose technical errors**: Wrap GitHub errors in user-friendly messages

### Must Do
- **Always show spinner**: Even if fetch is fast, provides feedback
- **Handle all edge cases**: Empty list, network errors, API errors
- **Store issue data**: Full issue objects needed by later steps
- **Maintain choices state**: Don't overwrite existing wizard choices

### Performance
- **Fast for small lists**: <1 second for repos with <20 issues
- **Acceptable for large lists**: <5 seconds for repos with 100+ issues
- **Timeout protection**: `gh` CLI has built-in timeout (~30s for network operations)

## Notes

### GitHub API Rate Limits

The `gh` CLI handles GitHub API authentication and rate limits transparently. Authenticated users get 5,000 requests/hour, which is more than sufficient for issue listing operations.

If a user hits rate limits, the error will be returned by `ListIssues()` and displayed via the error handler.

### Label Filter Implementation

The label filter is handled by the GitHub API via `gh issue list --label sow`. Only issues with the exact label "sow" are returned. This is case-sensitive.

### Issue Ordering

Issues are returned in default GitHub API order (most recently created first). This is usually the desired behavior. If different ordering is needed, it can be added in a future enhancement.

### Internationalization

Issue titles may contain Unicode characters. The huh library handles Unicode display correctly in terminals that support it.

### Testing Network Operations

Tests use mock GitHub client to avoid real network calls. This makes tests:
- Fast (no network delay)
- Reliable (no network failures)
- Deterministic (same results every time)

### Future Enhancements

Possible improvements for later iterations:
- Search/filter issues by keywords
- Display issue assignee, milestone, labels
- Pagination for repositories with many issues
- Cache issue list for session duration
- Sort by created date, updated date, or title
