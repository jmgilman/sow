# Task 050: Integration Testing and Error Handling Polish

## Context

This final task ensures the complete GitHub issue integration workflow is robust, well-tested, and provides excellent user experience through comprehensive integration testing and error handling refinement. It validates that all components work together correctly and handles edge cases gracefully.

This task differs from the unit tests in Tasks 010-040 by focusing on end-to-end flows, error recovery paths, and real-world scenarios that span multiple wizard states. The goal is ensuring users have a smooth experience regardless of network conditions, GitHub state, or user actions.

## Requirements

### 1. Create Integration Test Suite

Create new file `cli/cmd/project/wizard_integration_test.go` with comprehensive end-to-end tests:

```go
package project

import (
    "testing"

    "github.com/jmgilman/sow/cli/internal/sow"
)

// TestCompleteGitHubIssueWorkflow tests the entire flow from source selection to finalization.
func TestCompleteGitHubIssueWorkflow(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    initGitRepo(t, tmpDir)
    ctx, err := sow.NewContext(tmpDir)
    if err != nil {
        t.Fatalf("failed to create context: %v", err)
    }

    mockIssues := []sow.Issue{
        {
            Number: 123,
            Title:  "Add JWT authentication",
            Body:   "Implement JWT-based auth",
            State:  "open",
            URL:    "https://github.com/test/repo/issues/123",
        },
    }

    wizard := &Wizard{
        state: StateCreateSource,
        ctx:   ctx,
        choices: make(map[string]interface{}),
        github: &mockGitHub{
            ensureErr:                nil,
            listIssuesResult:         mockIssues,
            getLinkedBranchesResult:  []sow.LinkedBranch{}, // No linked branches
            getIssueResult:          &mockIssues[0],
            createLinkedBranchResult: "feat/add-jwt-authentication-123",
        },
        cmd: nil, // Skip Claude launch
    }

    // Simulate: User selects "From GitHub issue"
    wizard.choices["source"] = "issue"
    wizard.state = StateIssueSelect

    // Step 1: handleIssueSelect - should fetch and store issues
    err = wizard.handleIssueSelect()
    if err != nil {
        t.Fatalf("handleIssueSelect failed: %v", err)
    }

    issues := wizard.choices["issues"].([]sow.Issue)
    if len(issues) != 1 {
        t.Fatalf("expected 1 issue, got %d", len(issues))
    }

    // Simulate: User selects first issue
    wizard.choices["selectedIssueNumber"] = 123

    // Step 2: showIssueSelectScreen - should validate and transition to type select
    err = wizard.showIssueSelectScreen()
    if err != nil {
        t.Fatalf("showIssueSelectScreen failed: %v", err)
    }

    if wizard.state != StateTypeSelect {
        t.Fatalf("expected state %s, got %s", StateTypeSelect, wizard.state)
    }

    issue := wizard.choices["issue"].(*sow.Issue)
    if issue.Number != 123 {
        t.Fatalf("expected issue 123, got %d", issue.Number)
    }

    // Simulate: User selects "standard" type
    wizard.choices["type"] = "standard"

    // Step 3: handleTypeSelect - should create branch and transition to prompt entry
    err = wizard.handleTypeSelect()
    if err != nil {
        t.Fatalf("handleTypeSelect failed: %v", err)
    }

    if wizard.state != StatePromptEntry {
        t.Fatalf("expected state %s, got %s", StatePromptEntry, wizard.state)
    }

    branchName := wizard.choices["branch"].(string)
    if branchName != "feat/add-jwt-authentication-123" {
        t.Fatalf("expected branch feat/add-jwt-authentication-123, got %s", branchName)
    }

    // Simulate: User enters prompt
    wizard.choices["prompt"] = "Focus on middleware implementation"
    wizard.state = StateComplete

    // Step 4: finalize - should create project with issue metadata
    err = wizard.finalize()
    if err != nil {
        t.Fatalf("finalize failed: %v", err)
    }

    // Verify project created with issue context
    worktreePath := sow.WorktreePath(tmpDir, branchName)
    issueFilePath := filepath.Join(worktreePath, ".sow", "project", "context", "issue-123.md")

    if _, err := os.Stat(issueFilePath); os.IsNotExist(err) {
        t.Error("issue context file not created")
    }

    t.Log("✓ Complete GitHub issue workflow succeeded")
}

// TestErrorRecoveryPaths tests various error scenarios and recovery.
func TestErrorRecoveryPaths(t *testing.T) {
    tests := []struct {
        name          string
        setupMock     func(*mockGitHub)
        expectedState WizardState
        description   string
    }{
        {
            name: "GitHub CLI not installed",
            setupMock: func(m *mockGitHub) {
                m.ensureErr = sow.ErrGHNotInstalled{}
            },
            expectedState: StateCreateSource,
            description:   "Should return to source selection",
        },
        {
            name: "GitHub CLI not authenticated",
            setupMock: func(m *mockGitHub) {
                m.ensureErr = sow.ErrGHNotAuthenticated{}
            },
            expectedState: StateCreateSource,
            description:   "Should return to source selection",
        },
        {
            name: "Empty issue list",
            setupMock: func(m *mockGitHub) {
                m.listIssuesResult = []sow.Issue{}
            },
            expectedState: StateCreateSource,
            description:   "Should return to source selection",
        },
        {
            name: "Network error fetching issues",
            setupMock: func(m *mockGitHub) {
                m.listIssuesErr = fmt.Errorf("network timeout")
            },
            expectedState: StateCreateSource,
            description:   "Should return to source selection",
        },
        {
            name: "Issue already has linked branch",
            setupMock: func(m *mockGitHub) {
                m.getLinkedBranchesResult = []sow.LinkedBranch{
                    {Name: "feat/existing", URL: "..."},
                }
            },
            expectedState: StateIssueSelect,
            description:   "Should return to issue selection",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockGitHub{}
            tt.setupMock(mock)

            wizard := &Wizard{
                state:   StateIssueSelect,
                ctx:     testContext(t),
                choices: make(map[string]interface{}),
                github:  mock,
            }

            // Run handler
            _ = wizard.handleIssueSelect()

            // Verify correct state transition
            if wizard.state != tt.expectedState {
                t.Errorf("%s: expected state %s, got %s",
                    tt.description, tt.expectedState, wizard.state)
            }
        })
    }
}

// TestBranchNameGeneration tests various issue titles produce valid branch names.
func TestBranchNameGeneration(t *testing.T) {
    tests := []struct {
        issueTitle  string
        issueNumber int
        projectType string
        expected    string
    }{
        {"Add JWT authentication", 123, "standard", "feat/add-jwt-authentication-123"},
        {"Refactor Database Schema", 456, "standard", "feat/refactor-database-schema-456"},
        {"Web Based Agents", 789, "exploration", "explore/web-based-agents-789"},
        {"Special!@#$% Characters!", 111, "design", "design/specialcharacters-111"},
        {"Multiple   Spaces", 222, "standard", "feat/multiple-spaces-222"},
        {"UPPERCASE TITLE", 333, "breakdown", "breakdown/uppercase-title-333"},
        {"CamelCaseTitle", 444, "standard", "feat/camelcasetitle-444"},
        {"Title-with-hyphens", 555, "standard", "feat/title-with-hyphens-555"},
        {"Title_with_underscores", 666, "standard", "feat/title_with_underscores-666"},
    }

    for _, tt := range tests {
        t.Run(tt.issueTitle, func(t *testing.T) {
            prefix := getTypePrefix(tt.projectType)
            slug := normalizeName(tt.issueTitle)
            result := fmt.Sprintf("%s%s-%d", prefix, slug, tt.issueNumber)

            if result != tt.expected {
                t.Errorf("expected %q, got %q", tt.expected, result)
            }

            // Verify it's a valid branch name
            if err := isValidBranchName(result); err != nil {
                t.Errorf("generated invalid branch name %q: %v", result, err)
            }
        })
    }
}
```

### 2. Add Manual Testing Checklist

Create `cli/cmd/project/TESTING.md` with manual test scenarios:

```markdown
# GitHub Issue Integration - Manual Testing Guide

## Prerequisites
- Repository with GitHub remote
- `gh` CLI installed and authenticated
- Issues in repository with 'sow' label

## Test Scenarios

### Scenario 1: Happy Path - Create from Issue
1. Create test issue in GitHub with 'sow' label
2. Run `sow project`
3. Select "Create new project"
4. Select "From GitHub issue"
5. Verify: Spinner shows "Fetching issues from GitHub..."
6. Verify: Issue appears in list with format "#123: Title"
7. Select the test issue
8. Select project type (e.g., "Standard")
9. Verify: Type selection shows "Issue: #123 - Title"
10. Verify: Spinner shows "Creating linked branch..."
11. Enter optional prompt (or skip)
12. Verify: Success message shows issue link
13. Verify: Claude launches in worktree
14. Verify in GitHub: Issue shows linked branch in UI

**Expected**: Project created successfully with issue context

### Scenario 2: Issue Already Linked
1. Create project from issue (Scenario 1)
2. Exit Claude, run `sow project` again
3. Select "Create new project" → "From GitHub issue"
4. Select the SAME issue
5. Verify: Error shows branch name and suggests "Continue existing project"
6. Press Enter
7. Verify: Returns to issue list
8. Select "Cancel" to exit

**Expected**: Clear error, can select different issue or cancel

### Scenario 3: GitHub CLI Not Installed
1. Temporarily rename `gh` CLI (or unset PATH)
2. Run `sow project`
3. Select "Create new project" → "From GitHub issue"
4. Verify: Error shows installation instructions
5. Verify: Error suggests "From branch name" alternative
6. Press Enter
7. Verify: Returns to source selection
8. Select "From branch name" path
9. Verify: Can create project without GitHub

**Expected**: Helpful error, fallback path works

### Scenario 4: GitHub CLI Not Authenticated
1. Run `gh auth logout`
2. Run `sow project`
3. Select "Create new project" → "From GitHub issue"
4. Verify: Error shows "gh auth login" instructions
5. Run `gh auth login` in another terminal
6. Run `sow project` again
7. Verify: Can now access issues

**Expected**: Clear auth instructions, can retry after fixing

### Scenario 5: No Issues with 'sow' Label
1. Ensure no issues in repo have 'sow' label
2. Run `sow project`
3. Select "Create new project" → "From GitHub issue"
4. Verify: Error explains no issues found
5. Verify: Error explains how to create and label issue
6. Press Enter
7. Verify: Returns to source selection

**Expected**: Helpful guidance on creating labeled issues

### Scenario 6: External Editor Integration
1. Set `EDITOR` environment variable (e.g., `export EDITOR=vim`)
2. Start creating project from issue
3. At prompt entry, press Ctrl+E
4. Verify: Editor opens with temp file
5. Write multi-line prompt:
   ```
   Focus on:
   - Middleware implementation
   - Integration tests
   - Error handling
   ```
6. Save and exit editor
7. Verify: Content appears in wizard
8. Press Enter
9. Verify: Project created with full prompt

**Expected**: External editor works, multi-line content preserved

### Scenario 7: Network Error During Fetch
1. Disconnect network (or use network simulator)
2. Run `sow project`
3. Select "Create new project" → "From GitHub issue"
4. Verify: Error shows network issue explanation
5. Reconnect network
6. Return to source selection, try again
7. Verify: Works after network restored

**Expected**: Graceful degradation, can retry

### Scenario 8: Unicode in Issue Title
1. Create issue with Unicode title: "Add 日本語 support"
2. Create project from this issue
3. Verify: Title displays correctly in wizard
4. Verify: Branch name is ASCII-safe: "feat/add-support-123"
5. Verify: Issue context file contains Unicode title

**Expected**: Unicode handled correctly, branch name sanitized

### Scenario 9: Very Long Issue Title
1. Create issue with 200+ character title
2. Create project from this issue
3. Verify: Title displays (may wrap in terminal)
4. Verify: Branch name is reasonable length
5. Verify: Project created successfully

**Expected**: Long titles handled gracefully

### Scenario 10: Branch Name Path Still Works
1. Run `sow project`
2. Select "Create new project" → "From branch name"
3. Select type, enter name, enter prompt
4. Verify: Works exactly as before (no regression)
5. Verify: No issue metadata in project state

**Expected**: Existing functionality unaffected
```

### 3. Polish Error Messages

Review and refine all error messages for clarity and helpfulness. Update messages in `wizard_state.go` to match this format:

**Error Message Template**:
```
[What went wrong]

[How to fix it / Why it happened]

[Next steps / Alternative paths]
```

**Example refinements**:

```go
// BEFORE:
"Failed to fetch issues: network timeout"

// AFTER:
"Failed to fetch issues from GitHub\n\n" +
    "This may be due to:\n" +
    "  - Network connectivity issues\n" +
    "  - GitHub API being temporarily unavailable\n" +
    "  - GitHub API rate limits\n\n" +
    "Please try again in a moment, or select 'From branch name' to continue without GitHub integration."
```

### 4. Add Helpful Debug Information

Add optional verbose mode for troubleshooting. In `handleIssueSelect()`:

```go
// Fetch issues with 'sow' label
var issues []sow.Issue
var fetchErr error

err := withSpinner("Fetching issues from GitHub...", func() error {
    issues, fetchErr = w.github.ListIssues("sow", "open")

    // Debug logging if SOW_DEBUG=1
    if os.Getenv("SOW_DEBUG") == "1" {
        fmt.Fprintf(os.Stderr, "[DEBUG] Fetched %d issues\n", len(issues))
        for _, issue := range issues {
            fmt.Fprintf(os.Stderr, "[DEBUG] Issue #%d: %s\n", issue.Number, issue.Title)
        }
    }

    return fetchErr
})
```

This helps users and developers troubleshoot issues without modifying code.

### 5. Validate All State Transitions

Create a state transition validator to catch logic errors:

```go
// validateStateTransition checks if a state transition is valid.
// This helps catch logic errors during development.
func (w *Wizard) validateStateTransition(from, to WizardState) error {
    // Define valid transitions
    validTransitions := map[WizardState][]WizardState{
        StateEntry: {StateCreateSource, StateProjectSelect, StateCancelled},
        StateCreateSource: {StateIssueSelect, StateTypeSelect, StateCancelled},
        StateIssueSelect: {StateTypeSelect, StateCreateSource, StateCancelled},
        StateTypeSelect: {StateNameEntry, StatePromptEntry, StateCancelled},
        StateNameEntry: {StatePromptEntry, StateTypeSelect, StateCancelled},
        StatePromptEntry: {StateComplete, StateCancelled},
        StateProjectSelect: {StateContinuePrompt, StateCancelled},
        StateContinuePrompt: {StateComplete, StateCancelled},
    }

    allowed, exists := validTransitions[from]
    if !exists {
        return fmt.Errorf("unknown source state: %s", from)
    }

    for _, validTo := range allowed {
        if validTo == to {
            return nil // Transition is valid
        }
    }

    return fmt.Errorf("invalid transition from %s to %s", from, to)
}

// Use in setState helper:
func (w *Wizard) setState(newState WizardState) error {
    if os.Getenv("SOW_DEBUG") == "1" {
        if err := w.validateStateTransition(w.state, newState); err != nil {
            fmt.Fprintf(os.Stderr, "[WARN] %v\n", err)
        }
    }
    w.state = newState
    return nil
}
```

## Acceptance Criteria

### Integration Test Coverage

1. **Complete workflow test**: End-to-end test from source selection to finalization passes
2. **Error recovery tests**: All error scenarios transition to correct states
3. **Branch name tests**: All special character and edge case titles produce valid branch names
4. **State transition tests**: No invalid state transitions occur
5. **Both paths tested**: GitHub issue path and branch name path both work

### Manual Testing

1. **All scenarios pass**: Each manual test scenario in TESTING.md succeeds
2. **Error messages helpful**: Users understand what went wrong and how to fix it
3. **No dead ends**: Every error state has a path forward (retry, fallback, or cancel)
4. **External editor works**: Ctrl+E opens editor for all common editors (vim, nano, code, etc.)
5. **GitHub integration visible**: Created branches visible in GitHub UI with issue link

### Code Quality

1. **No lint errors**: `go vet` and `golangci-lint` pass
2. **Test coverage**: >80% coverage for new code
3. **Debug mode works**: `SOW_DEBUG=1` provides useful troubleshooting info
4. **Documentation updated**: TESTING.md is comprehensive and accurate

## Technical Details

### Integration Test Strategy

Integration tests use real `Wizard` struct with mock GitHub client but real filesystem operations (via `t.TempDir()`). This validates:
- State machine logic
- File creation
- Worktree management
- Context storage

But avoids:
- Real GitHub API calls
- Real Claude launch
- Network dependencies

### State Transition Validation

The validator is only active in debug mode to avoid performance overhead in production. It catches common errors like:
- Skipping required states
- Looping back incorrectly
- Going directly to completion from wrong state

### Error Message Consistency

All error messages follow the template:
1. What happened (1-2 sentences)
2. Why / how to fix (bullet points or numbered steps)
3. Next steps (clear action items)

This ensures users always know:
- What went wrong
- How to fix it
- What to do next

### Debug Output Format

Debug output uses stderr (not stdout) to avoid interfering with normal output. Format:
```
[DEBUG] <component>: <message>
[WARN] <component>: <warning>
[ERROR] <component>: <error>
```

Examples:
```
[DEBUG] GitHub: Fetched 3 issues
[DEBUG] GitHub: Issue #123: Add JWT auth
[WARN] State transition: invalid transition from entry to prompt_entry
[ERROR] GitHub: Failed to create branch: network timeout
```

## Relevant Inputs

### Existing Test Files
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_state_test.go` - Existing unit tests
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_helpers_test.go` - Helper function tests

### Wizard Implementation
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_state.go` - All wizard state handlers
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_helpers.go` - Helper functions

### GitHub Client
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/internal/sow/github.go` - Mock-able GitHub operations

### Design Specifications
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/knowledge/designs/interactive-wizard-ux-flow.md`
  - Lines 418-535: All error message formats

- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/project/context/issue-70.md`
  - Lines 554-722: Testing requirements
  - Lines 724-862: Manual testing scenarios

## Examples

### Example: Integration Test Output
```
=== RUN   TestCompleteGitHubIssueWorkflow
    wizard_integration_test.go:45: ✓ Fetched 1 issue
    wizard_integration_test.go:53: ✓ Validated issue has no linked branch
    wizard_integration_test.go:62: ✓ Created branch feat/add-jwt-authentication-123
    wizard_integration_test.go:71: ✓ Initialized project with issue metadata
    wizard_integration_test.go:78: ✓ Issue context file created
    wizard_integration_test.go:85: ✓ Complete GitHub issue workflow succeeded
--- PASS: TestCompleteGitHubIssueWorkflow (0.15s)
```

### Example: Debug Mode Output
```bash
$ SOW_DEBUG=1 sow project
[DEBUG] Wizard: State=entry
[DEBUG] User selected: create
[DEBUG] Wizard: State=create_source
[DEBUG] User selected: issue
[DEBUG] Wizard: State=issue_select
[DEBUG] GitHub: Calling gh issue list --label sow --state open
[DEBUG] GitHub: Fetched 2 issues
[DEBUG] GitHub: Issue #123: Add JWT authentication
[DEBUG] GitHub: Issue #124: Refactor schema
[DEBUG] User selected: 123
[DEBUG] GitHub: Checking linked branches for issue 123
[DEBUG] GitHub: No linked branches found
[DEBUG] GitHub: Fetching full issue details for #123
[DEBUG] Wizard: State=type_select
```

### Example: Refined Error Message
```
╔══════════════════════════════════════════════════════════╗
║                         Error                            ║
╚══════════════════════════════════════════════════════════╝

Failed to fetch issues from GitHub

This may be due to:
  • Network connectivity issues
  • GitHub API being temporarily unavailable
  • GitHub API rate limits (try again in 1 hour)
  • Repository visibility settings

Troubleshooting:
  1. Check your network connection
  2. Visit https://status.github.com/ for GitHub status
  3. Try: gh auth status (to verify authentication)

You can:
  • Try again (may be transient)
  • Select 'From branch name' to continue without GitHub
  • Cancel and check your setup

[Press Enter to return to source selection]
```

## Dependencies

### Prerequisites
- Tasks 010-040 (all GitHub issue workflow tasks) - MUST be complete
- All unit tests passing
- Manual testing environment available

### Depends On
- Complete wizard implementation
- Mock GitHub client
- Test helpers

### Enables
- Confident deployment
- User documentation
- Future enhancements built on solid foundation

## Constraints

### Must Not
- **Skip manual testing**: Integration tests don't catch all UI/UX issues
- **Ignore edge cases**: Real users encounter unexpected scenarios
- **Over-engineer**: Keep debug mode simple and optional

### Must Do
- **Test all paths**: Both GitHub issue path and branch name path
- **Verify in GitHub**: Actually check that issue-branch links appear in UI
- **Document findings**: Update TESTING.md with any new scenarios discovered

### Performance
- **Integration tests fast**: Should run in <5 seconds total
- **Debug mode negligible**: <1% performance overhead when enabled
- **Manual tests reasonable**: Each scenario should take <2 minutes

## Notes

### Why Manual Testing Matters

Integration tests validate logic but can't catch:
- Terminal rendering issues
- Spinner animation problems
- External editor integration bugs
- Actual GitHub API behavior
- Network timeout handling

Manual testing complements automated tests.

### Debug Mode Best Practices

Debug mode should:
- Use stderr (not stdout)
- Be completely optional (default off)
- Provide actionable information
- Not spam the terminal
- Help diagnose real issues

### Continuous Integration

These tests should run in CI/CD:
- Unit tests: Always
- Integration tests: Always
- Manual tests: Before releases

Consider adding GitHub Actions workflow:
```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
    - run: go test ./cli/cmd/project/... -v
```

### User Feedback Loop

After implementation:
1. Internal testing by team
2. Documentation review
3. Alpha testing with select users
4. Beta release
5. Gather feedback
6. Refine error messages
7. Update TESTING.md

This ensures the feature meets real user needs.

### Future Test Additions

As bugs are discovered:
1. Add reproduction to integration tests
2. Fix the bug
3. Verify fix with test
4. Update TESTING.md if new scenario

This prevents regressions.
