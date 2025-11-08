# Issue #70: GitHub Issue Integration Workflow

**URL**: https://github.com/jmgilman/sow/issues/70
**State**: OPEN

## Description

# Work Unit 003: GitHub Issue Integration Workflow

**Size**: 2-3 day project-sized work unit
**Dependencies**: Work Unit 001 (foundation), Work Unit 002 (reuses finalization logic)
**GitHub Issue**: [Will be created from this specification]

---

## 1. Behavioral Goal (User Story)

**As a** developer using `sow` to manage projects,
**I need** to create new projects directly from GitHub issues labeled with 'sow',
**So that** I can seamlessly transition from issue triage to implementation with automatic branch naming, issue linking, and context preservation.

### Success Criteria for Reviewers

- User can select "From GitHub issue" and see a list of issues filtered by the 'sow' label
- Issue selection validates that the issue doesn't already have a linked branch
- Branch is created with correct naming convention and automatically linked to the issue via `gh issue develop`
- Project is initialized with issue context (number, title, URL) stored in project state
- Claude launches with issue context included in the prompt
- Error handling gracefully degrades when `gh` CLI is missing or unauthenticated

---

## 2. Existing Code Context

### Explanatory Context

This work unit extends the wizard's creation flow to support GitHub issue integration. It leverages the existing `GitHub` client (in `cli/internal/sow/github.go`) which provides robust wrappers around the `gh` CLI for listing issues, checking linked branches, and creating branch-issue links.

The implementation builds on Work Unit 001's state machine foundation, adding two new states (`StateCreateSource` for source selection and `StateIssueSelect` for issue listing/selection), then reusing existing states from Work Unit 002 for type selection, prompt entry, and finalization.

The critical integration point is the `gh issue develop` command, which both creates the branch AND links it to the issue in a single operation. This is more elegant than creating a branch separately and manually linking it.

### Reference List: Key Files

**Existing GitHub Client** (fully implemented, ready to use):
- `cli/internal/sow/github.go:1-465` - Complete GitHub client implementation
  - Lines 105-136: `CheckInstalled()`, `CheckAuthenticated()`, `Ensure()` - gh CLI validation
  - Lines 147-173: `ListIssues(label, state)` - Fetch issues with label filter
  - Lines 201-254: `GetLinkedBranches(number)` - Check if issue has linked branch
  - Lines 256-323: `CreateLinkedBranch(issueNumber, branchName, checkout)` - Create linked branch via `gh issue develop`

**Foundation from Work Unit 001** (assumed complete):
- `cli/cmd/project/wizard_state.go` - State machine with constants for all wizard states
- `cli/cmd/project/wizard_helpers.go` - Name normalization, type configuration, error display utilities
- `cli/cmd/project/wizard.go` - Wizard struct with state tracking and choices map

**Finalization from Work Unit 002** (assumed complete, will reuse):
- `cli/cmd/project/shared.go` - Extracted finalization logic including:
  - Conditional uncommitted changes check
  - Worktree creation via `EnsureWorktree()`
  - Project initialization
  - 3-layer prompt generation
  - Claude launch

**Worktree Management** (existing, no changes):
- `cli/internal/sow/worktree.go:1-200` - Worktree creation and management

---

## 3. Existing Documentation Context

### UX Flow Design (`.sow/knowledge/designs/interactive-wizard-ux-flow.md`)

**Lines 60-174: Path 1A - GitHub Issue Workflow**

This section defines the complete user journey for creating projects from GitHub issues:

1. **Screen 1.1** (lines 64-77): Source selection - user chooses "From GitHub issue" vs "From branch name"
2. **Screen 1A.1** (lines 84-108): Issue selection screen - display issues filtered by 'sow' label, validate selected issue doesn't have linked branch
3. **Screen 1A.2** (lines 109-141): Type selection with issue context displayed
4. **Screen 1A.3** (lines 143-173): Initial prompt entry with issue context

**Lines 418-535: Error Messages**

Defines specific error formats relevant to this work unit:
- Lines 445-458: "Issue Already Linked" error - what to show when issue has linked branch
- Lines 517-535: "GitHub CLI Missing" error - helpful instructions for installing `gh`

**Lines 539-568: Journey 1 Example**

Complete walkthrough showing user creating project from issue #124, including all screens and state transitions.

### Technical Implementation (`.sow/knowledge/designs/interactive-wizard-technical-implementation.md`)

**Lines 143-276: Implementation Examples**

Provides concrete code patterns for:
- Select prompts for issue listing (similar to lines 280-335: project list example)
- Text areas with external editor (lines 239-276)
- Dynamic forms (lines 251-349)

**Lines 356-479: State Detection Logic**

Patterns for checking branch state, listing existing entities, and validating selections.

**Lines 860-881: Loading Indicators**

Spinner integration for async GitHub operations (fetching issues).

### Library Verification (`.sow/knowledge/designs/huh-library-verification.md`)

**Critical Finding - Lines 183-248: External Editor Keybinding**

Documents that the correct keybinding is **Ctrl+E** (not Ctrl+O as originally drafted). All screens in this work unit must use Ctrl+E.

**Lines 360-421: Loading Indicators**

Confirms `github.com/charmbracelet/huh/spinner` package supports loading spinners for async operations like fetching issues.

---

## 4. Implementation Scope

### What Needs to Be Built

#### 4.1 Source Selection Screen (NEW)

**State**: `StateCreateSource`

After user selects "Create new project" from entry screen, show:
- Three options: "From GitHub issue" / "From branch name" / "Cancel"
- Routes to `StateIssueSelect` for GitHub path
- Routes to `StateTypeSelect` for branch name path (Work Unit 002)
- Routes to `StateCancelled` for cancel

Implementation:
```go
func (w *Wizard) handleCreateSource() error {
    var source string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("How do you want to create this project?").
                Options(
                    huh.NewOption("From GitHub issue", "issue"),
                    huh.NewOption("From branch name", "branch"),
                    huh.NewOption("Cancel", "cancel"),
                ).
                Value(&source),
        ),
    )

    // Handle selection, transition to appropriate state
}
```

#### 4.2 GitHub CLI Availability Check (NEW)

Before showing issue list, verify `gh` CLI is available:
- Call `github.CheckInstalled()` - returns `ErrGHNotInstalled` if missing
- Call `github.CheckAuthenticated()` - returns `ErrGHNotAuthenticated` if not logged in
- If either check fails: show helpful error with install instructions and offer fallback to branch name path

Error handling:
```go
gh := sow.NewGitHub(exec.NewLocal("gh"))
if err := gh.Ensure(); err != nil {
    // Show error based on type
    if _, ok := err.(sow.ErrGHNotInstalled); ok {
        // Show installation instructions, offer fallback
    } else if _, ok := err.(sow.ErrGHNotAuthenticated); ok {
        // Show auth instructions
    }
    return nil // Allow retry
}
```

#### 4.3 Issue Listing Screen (NEW)

**State**: `StateIssueSelect`

Fetch and display issues with 'sow' label:
- Use spinner while fetching: `spinner.New().Title("Fetching issues from GitHub...").Action(func() { ... }).Run()`
- Call `github.ListIssues("sow", "open")` to fetch issues
- Handle empty list gracefully (show message, return to source selection)
- Display as Select with format: "#123: Issue Title"
- Include "Cancel" option

Implementation:
```go
// Fetch with spinner
var issues []sow.Issue
var fetchErr error
err := spinner.New().
    Type(spinner.Line).
    Title("Fetching issues from GitHub...").
    Action(func() {
        issues, fetchErr = github.ListIssues("sow", "open")
    }).
    Run()

if fetchErr != nil {
    showError(fmt.Sprintf("Failed to fetch issues: %v", fetchErr))
    return nil
}

if len(issues) == 0 {
    showError("No issues found with 'sow' label")
    w.state = StateCreateSource // Return to source selection
    return nil
}

// Build select options
options := make([]huh.Option[int], 0, len(issues)+1)
for _, issue := range issues {
    label := fmt.Sprintf("#%d: %s", issue.Number, issue.Title)
    options = append(options, huh.NewOption(label, issue.Number))
}
options = append(options, huh.NewOption("Cancel", -1))

// Show select form...
```

#### 4.4 Issue Selection and Validation (NEW)

When user selects an issue:
1. Call `github.GetLinkedBranches(issueNumber)` to check for linked branches
2. If linked branches exist (len > 0): show error, return to issue list
3. If no linked branches: store issue info in choices map, proceed to type selection

Validation logic:
```go
// After user selects issue
linkedBranches, err := github.GetLinkedBranches(selectedIssueNumber)
if err != nil {
    showError(fmt.Sprintf("Failed to check linked branches: %v", err))
    return nil // Stay in current state to retry
}

if len(linkedBranches) > 0 {
    branch := linkedBranches[0]
    errorMsg := fmt.Sprintf(
        "Issue #%d already has a linked branch: %s\n\n"+
        "To continue working on this issue:\n"+
        "  Select \"Continue existing project\" from the main menu",
        selectedIssueNumber, branch.Name,
    )
    showError(errorMsg)
    return nil // Stay in StateIssueSelect to allow selecting different issue
}

// Issue is available - fetch full issue details and store
issue, err := github.GetIssue(selectedIssueNumber)
if err != nil {
    showError(fmt.Sprintf("Failed to get issue details: %v", err))
    return nil
}

w.choices["issue"] = issue
w.state = StateTypeSelect
```

#### 4.5 Type Selection with Issue Context (REUSE + ENHANCE)

**State**: `StateTypeSelect` (shared with Work Unit 002)

Reuse type selection screen from Work Unit 002, but display issue context:
- Show header: "Issue: #123 - Issue Title"
- Same four project type options
- Transition to branch creation (NOT name entry like Work Unit 002)

Enhancement:
```go
func (w *Wizard) handleTypeSelect() error {
    var projectType string

    // Check if we have issue context
    var contextLine string
    if issue, ok := w.choices["issue"].(*sow.Issue); ok {
        contextLine = fmt.Sprintf("Issue: #%d - %s\n", issue.Number, issue.Title)
    }

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewNote().
                Description(contextLine),
            huh.NewSelect[string]().
                Title("What type of project?").
                Options(getTypeOptions()...).
                Value(&projectType),
        ),
    )

    // After selection, transition based on context
    w.choices["type"] = projectType
    if _, hasIssue := w.choices["issue"]; hasIssue {
        // GitHub issue path: proceed to branch creation
        w.state = StatePromptEntry // Skip name entry, branch created from issue
    } else {
        // Branch name path: proceed to name entry
        w.state = StateNameEntry
    }
}
```

#### 4.6 Branch Creation (NEW)

After type selection, generate branch name and create linked branch:
- Generate branch name: `<prefix><issue-slug>-<number>`
  - Example: issue "Add JWT authentication" + Standard type → `feat/add-jwt-authentication-123`
  - Use issue title to generate slug (normalize similar to branch name normalization)
- Call `github.CreateLinkedBranch(issueNumber, branchName, false)` (checkout=false, we use worktrees)
- Store generated branch name in choices map
- Transition to prompt entry

Implementation:
```go
// After type selection for issue path
issue := w.choices["issue"].(*sow.Issue)
projectType := w.choices["type"].(string)
prefix := getTypePrefix(projectType)

// Generate branch name from issue
issueSlug := normalizeName(issue.Title)
branchName := fmt.Sprintf("%s%s-%d", prefix, issueSlug, issue.Number)

// Create linked branch via gh issue develop
createdBranch, err := github.CreateLinkedBranch(issue.Number, branchName, false)
if err != nil {
    showError(fmt.Sprintf("Failed to create linked branch: %v", err))
    return nil // Stay in current state to retry
}

w.choices["branch"] = createdBranch
w.choices["name"] = issue.Title // Use issue title as project name
w.state = StatePromptEntry
```

#### 4.7 Prompt Entry with Issue Context (REUSE + ENHANCE)

**State**: `StatePromptEntry` (shared with Work Unit 002)

Reuse prompt entry screen from Work Unit 002, but display issue context:
- Show context: "Issue: #123 - Title" and "Branch: <branch-name>"
- Multi-line text area with Ctrl+E for external editor
- Optional prompt (user can skip)
- Transition to completion/finalization

Enhancement:
```go
func (w *Wizard) handlePromptEntry() error {
    var prompt string

    // Build context display
    var contextLines []string
    if issue, ok := w.choices["issue"].(*sow.Issue); ok {
        contextLines = append(contextLines,
            fmt.Sprintf("Issue: #%d - %s", issue.Number, issue.Title))
    }
    if branchName, ok := w.choices["branch"].(string); ok {
        contextLines = append(contextLines,
            fmt.Sprintf("Branch: %s", branchName))
    } else if name, ok := w.choices["name"].(string); ok {
        // Branch name path - show computed branch
        projectType := w.choices["type"].(string)
        prefix := getTypePrefix(projectType)
        normalized := normalizeName(name)
        contextLines = append(contextLines,
            fmt.Sprintf("Branch: %s%s", prefix, normalized))
    }

    contextDisplay := strings.Join(contextLines, "\n")

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewText().
                Title("Enter your task or question for Claude (optional):").
                Description(contextDisplay + "\nPress Ctrl+E to open $EDITOR for multi-line input").
                CharLimit(10000).
                Value(&prompt).
                EditorExtension(".md"),
        ),
    )

    // ... handle submission
}
```

#### 4.8 Finalization with Issue Context (REUSE + ENHANCE)

**State**: `StateComplete`

Reuse finalization logic from Work Unit 002, with enhancement to store issue metadata:
- All existing finalization steps (uncommitted changes check, worktree creation, etc.)
- Store issue context in project state: issue number, title, URL
- Include issue context in generated prompt
- Launch Claude

Enhancement to shared finalization:
```go
func (w *Wizard) finalize() error {
    // Get branch name (either from issue creation or user entry)
    branchName := w.choices["branch"].(string)

    // Conditional uncommitted changes check (existing logic)
    if err := performUncommittedChangesCheckIfNeeded(w.ctx, branchName); err != nil {
        return err
    }

    // Ensure worktree exists (existing logic)
    if err := sow.EnsureWorktree(w.ctx, branchName); err != nil {
        return err
    }

    // Initialize project
    projectType := w.choices["type"].(string)
    projectName := w.choices["name"].(string)

    projectConfig := buildProjectConfig(projectType)

    // Add issue context if present
    if issue, ok := w.choices["issue"].(*sow.Issue); ok {
        projectConfig.IssueNumber = issue.Number
        projectConfig.IssueTitle = issue.Title
        projectConfig.IssueURL = issue.URL
    }

    if err := initializeProject(w.ctx, branchName, projectName, projectConfig); err != nil {
        return err
    }

    // Generate 3-layer prompt with issue context
    prompt := buildPrompt(projectType, projectName, w.choices["prompt"].(string), issue)

    // Launch Claude in worktree
    return launchClaude(w.ctx, branchName, prompt)
}
```

### Integration with Foundation (Work Unit 001)

**New States to Add** to `cli/cmd/project/wizard_state.go`:
```go
const (
    // ... existing states from WU001
    StateCreateSource   WizardState = "create_source"   // NEW
    StateIssueSelect    WizardState = "issue_select"    // NEW
    // ... other states
)
```

**Reused States** from Work Unit 002:
- `StateTypeSelect` - enhanced to show issue context
- `StatePromptEntry` - enhanced to show issue context
- `StateComplete` - enhanced to store issue metadata

### Integration with Work Unit 002

**Shared Code**:
- `StateTypeSelect` handler must check for issue vs. branch context and route appropriately
- `StatePromptEntry` handler displays issue context when present
- Finalization logic reused entirely, with issue metadata stored when present

**Branch of Execution**:
- After entry screen → source selection (NEW)
- Source selection → issue path (NEW) OR branch name path (WU002)
- Both paths converge at finalization with different context

---

## 5. Acceptance Criteria

These are objective, measurable criteria that reviewers will verify:

### AC1: Source Selection Works
- [ ] After selecting "Create new project", user sees source selection screen
- [ ] Three options displayed: "From GitHub issue", "From branch name", "Cancel"
- [ ] Selecting "From GitHub issue" transitions to GitHub CLI check
- [ ] Selecting "From branch name" transitions to type selection (WU002 flow)
- [ ] Selecting "Cancel" exits wizard cleanly

### AC2: GitHub CLI Validation
- [ ] When `gh` CLI not installed, error message shows with installation instructions
- [ ] Error offers fallback: "Or select 'From branch name' instead"
- [ ] When `gh` not authenticated, error shows with `gh auth login` instructions
- [ ] After error, user can return to source selection to choose different path
- [ ] When `gh` is installed and authenticated, wizard proceeds to issue listing

### AC3: Issue Listing Displays Correctly
- [ ] Loading spinner appears with message "Fetching issues from GitHub..."
- [ ] After fetch, issues with 'sow' label are displayed
- [ ] Issue format: "#123: Issue Title" (number before colon, title after)
- [ ] Cancel option included in list
- [ ] Empty list shows helpful message and returns to source selection
- [ ] Selecting cancel returns to source selection

### AC4: Linked Branch Detection
- [ ] When user selects issue with linked branch, error is displayed
- [ ] Error message includes: issue number, linked branch name, suggestion to use "Continue existing project"
- [ ] After error, user remains on issue list (can select different issue)
- [ ] When user selects issue without linked branch, validation passes
- [ ] Wizard proceeds to type selection after validation passes

### AC5: Type Selection with Issue Context
- [ ] Type selection screen displays issue context header
- [ ] Header format: "Issue: #123 - Issue Title"
- [ ] Four project type options displayed with descriptions
- [ ] User can select any type
- [ ] After selection, wizard proceeds to branch creation (skips name entry)

### AC6: Branch Creation via gh issue develop
- [ ] Branch name generated correctly: `<prefix><issue-slug>-<number>`
- [ ] Example: "Add JWT auth" + Standard → `feat/add-jwt-auth-123`
- [ ] `gh issue develop <number> --name <branch>` command executed
- [ ] Branch created successfully (can verify with `git branch --list`)
- [ ] Branch is linked to issue (can verify with `gh issue develop --list <number>`)
- [ ] On error, user sees error message and can retry

### AC7: Prompt Entry with Context
- [ ] Prompt entry screen shows issue context: "Issue: #123 - Title"
- [ ] Screen shows branch context: "Branch: feat/..."
- [ ] Multi-line text area accepts input
- [ ] User can press Ctrl+E to open $EDITOR
- [ ] External editor opens, content saved when editor exits
- [ ] User can skip prompt entry (leave empty)
- [ ] After submission, wizard proceeds to finalization

### AC8: Project Initialization with Issue Metadata
- [ ] Worktree created at `.sow/worktrees/<branch-name>/`
- [ ] Project state file created at `.sow/worktrees/<branch-name>/.sow/project/state.yaml`
- [ ] Project state contains issue number
- [ ] Project state contains issue title
- [ ] Project state contains issue URL
- [ ] Project type correctly set in state

### AC9: Claude Launch with Issue Context
- [ ] Claude launches in worktree directory
- [ ] Generated prompt includes issue context
- [ ] Prompt includes: issue number, title, user's optional prompt
- [ ] Claude can see issue context when started
- [ ] Success message displayed: "✓ Initialized project '<name>' on branch <branch>"
- [ ] Success message: "✓ Launching Claude in worktree..."

### AC10: Error Recovery
- [ ] All errors show helpful messages with next steps
- [ ] GitHub API errors handled gracefully
- [ ] Network errors handled gracefully
- [ ] User can navigate back after errors
- [ ] No crashes or panics in error scenarios

---

## 6. Testing Requirements

### 6.1 Unit Tests

**Test: Branch Name Generation from Issue**
```go
func TestGenerateBranchNameFromIssue(t *testing.T) {
    tests := []struct {
        issueTitle   string
        issueNumber  int
        projectType  string
        expectedName string
    }{
        {"Add JWT authentication", 123, "standard", "feat/add-jwt-authentication-123"},
        {"Refactor Database Schema", 456, "standard", "feat/refactor-database-schema-456"},
        {"Web Based Agents", 789, "exploration", "explore/web-based-agents-789"},
        {"Special!@#$%Chars", 111, "design", "design/specialchars-111"},
    }

    for _, tt := range tests {
        t.Run(tt.issueTitle, func(t *testing.T) {
            result := generateBranchName(tt.issueTitle, tt.issueNumber, tt.projectType)
            if result != tt.expectedName {
                t.Errorf("got %q, want %q", result, tt.expectedName)
            }
        })
    }
}
```

**Test: GitHub CLI Error Handling**
```go
func TestGitHubCLINotInstalled(t *testing.T) {
    mockGH := &MockGitHub{
        ensureErr: sow.ErrGHNotInstalled{},
    }

    // Test that wizard shows appropriate error and offers fallback
    // ...
}

func TestGitHubCLINotAuthenticated(t *testing.T) {
    mockGH := &MockGitHub{
        ensureErr: sow.ErrGHNotAuthenticated{},
    }

    // Test that wizard shows auth instructions
    // ...
}
```

### 6.2 Integration Tests

**Test: Issue Listing with Mock GitHub Client**
```go
func TestIssueListingFlow(t *testing.T) {
    mockIssues := []sow.Issue{
        {Number: 123, Title: "Add JWT auth", State: "open"},
        {Number: 124, Title: "Refactor schema", State: "open"},
    }

    mockGH := &MockGitHub{
        listIssues: mockIssues,
    }

    wizard := NewWizard(mockGH)
    wizard.state = StateIssueSelect

    // Simulate user selecting first issue
    // Verify state transitions correctly
    // ...
}
```

**Test: Linked Branch Detection**
```go
func TestLinkedBranchDetection(t *testing.T) {
    mockGH := &MockGitHub{
        linkedBranches: []sow.LinkedBranch{
            {Name: "feat/existing-branch", URL: "..."},
        },
    }

    // Test that wizard shows error when issue has linked branch
    // Test that user stays on issue list after error
    // ...
}
```

**Test: End-to-End Issue to Project Creation**
```go
func TestCompleteIssueWorkflow(t *testing.T) {
    mockGH := &MockGitHub{
        listIssues: []sow.Issue{
            {Number: 123, Title: "Test Issue", State: "open"},
        },
        linkedBranches: nil, // No linked branches
    }

    // Simulate complete workflow:
    // 1. Select "From GitHub issue"
    // 2. Select issue #123
    // 3. Validate no linked branch
    // 4. Select Standard type
    // 5. Create linked branch
    // 6. Enter prompt
    // 7. Verify project created with issue metadata
    // ...
}
```

### 6.3 Manual Testing Scenarios

**Scenario 1: Happy Path - Create Project from Issue**
1. Run `sow project`
2. Select "Create new project"
3. Select "From GitHub issue"
4. Verify spinner shows "Fetching issues from GitHub..."
5. Select an issue from the list (without linked branch)
6. Select project type (e.g., Standard)
7. Verify branch created with correct name
8. Enter initial prompt (or skip)
9. Verify project created successfully
10. Verify Claude launches with issue context

**Scenario 2: Issue Already Has Linked Branch**
1. Run `sow project`
2. Go through flow to issue selection
3. Select issue that already has linked branch
4. Verify error message shows: issue number, branch name, suggestion to continue
5. Verify user stays on issue list
6. Select different issue without linked branch
7. Verify flow continues normally

**Scenario 3: GitHub CLI Not Installed**
1. Ensure `gh` is not in PATH (or temporarily rename)
2. Run `sow project`
3. Select "From GitHub issue"
4. Verify error shows installation instructions
5. Verify error offers fallback to "From branch name"
6. Select fallback option
7. Verify wizard routes to branch name path

**Scenario 4: External Editor Integration**
1. Set `$EDITOR` to preferred editor (e.g., `vim`, `nano`, `code`)
2. Go through issue creation flow to prompt entry
3. Press Ctrl+E
4. Verify editor opens with temp file
5. Write multi-line prompt in editor
6. Save and exit editor
7. Verify content appears in wizard
8. Submit and verify project created with prompt

**Scenario 5: Empty Issue List**
1. Run in repository with no issues labeled 'sow' (or mock)
2. Select "From GitHub issue"
3. Verify spinner appears
4. Verify message: "No issues found with 'sow' label"
5. Verify returns to source selection
6. Can select "From branch name" as alternative

**Scenario 6: Network Error During Issue Fetch**
1. Simulate network error (disconnect WiFi or mock)
2. Select "From GitHub issue"
3. Verify error shows helpful message
4. Verify user can retry or cancel
5. Reconnect and retry
6. Verify flow continues normally

---

## 7. Technical Notes

### 7.1 gh CLI Integration

The `gh issue develop` command is the cornerstone of this integration:
```bash
gh issue develop <issue-number> --name <branch-name>
```

This single command:
- Creates the branch
- Links it to the issue
- Adds the link to GitHub's issue-branch tracking

**Critical**: We pass `checkout=false` to `CreateLinkedBranch()` because we use worktrees, not traditional checkout.

### 7.2 Branch Naming Convention

Branch names follow the pattern:
```
<prefix><normalized-issue-title>-<issue-number>
```

Examples:
- Issue "Add JWT authentication" (#123) + Standard → `feat/add-jwt-authentication-123`
- Issue "Refactor Database Schema" (#456) + Standard → `feat/refactor-database-schema-456`
- Issue "Web Based Agents" (#789) + Exploration → `explore/web-based-agents-789`

Normalization rules (same as Work Unit 002):
1. Convert to lowercase
2. Replace spaces with hyphens
3. Remove invalid characters (keep only `a-z`, `0-9`, `-`, `_`)
4. Collapse multiple consecutive hyphens
5. Remove leading/trailing hyphens

### 7.3 Issue Metadata Storage

Project state should store:
```yaml
name: "Add JWT authentication"
type: standard
branch: feat/add-jwt-authentication-123
issue:
  number: 123
  title: "Add JWT authentication"
  url: "https://github.com/owner/repo/issues/123"
created_at: "2025-11-06T10:30:00Z"
# ... other fields
```

This allows future features to reference the original issue.

### 7.4 Error Handling Philosophy

All GitHub-related errors should:
1. Show what went wrong
2. Suggest how to fix it
3. Offer a path forward (retry, fallback, cancel)

Never leave the user stuck - always provide options.

### 7.5 Loading Indicators

Use spinners for all async operations:
- Fetching issues from GitHub (can take 1-3 seconds)
- Checking linked branches (usually fast but should still show progress)
- Creating linked branch (can take 1-2 seconds)

This provides feedback that the system is working, not frozen.

---

## 8. Implementation Checklist

### Phase 1: Source Selection
- [ ] Add `StateCreateSource` constant to wizard_state.go
- [ ] Implement `handleCreateSource()` handler
- [ ] Add routing from entry screen to source selection
- [ ] Test source selection transitions correctly

### Phase 2: GitHub CLI Integration
- [ ] Create GitHub client instance in wizard initialization
- [ ] Implement `gh` CLI availability check
- [ ] Create error display for missing/unauthenticated `gh`
- [ ] Test error handling with mock GitHub client

### Phase 3: Issue Listing
- [ ] Add `StateIssueSelect` constant to wizard_state.go
- [ ] Implement issue fetch with spinner
- [ ] Implement `handleIssueSelect()` handler
- [ ] Handle empty issue list case
- [ ] Test with mock GitHub client returning various scenarios

### Phase 4: Issue Validation
- [ ] Implement linked branch checking
- [ ] Create error display for already-linked issues
- [ ] Test validation with mock returning linked branches
- [ ] Test validation with mock returning no linked branches

### Phase 5: Type Selection Enhancement
- [ ] Modify `handleTypeSelect()` to detect issue context
- [ ] Add issue context display (header)
- [ ] Add routing logic: issue path → prompt entry, branch path → name entry
- [ ] Test both routing scenarios

### Phase 6: Branch Creation
- [ ] Implement branch name generation from issue
- [ ] Add unit tests for branch name generation
- [ ] Call `CreateLinkedBranch()` after type selection
- [ ] Handle branch creation errors
- [ ] Store branch name and issue metadata in choices map

### Phase 7: Prompt Entry Enhancement
- [ ] Modify `handlePromptEntry()` to show issue context
- [ ] Test with issue context present
- [ ] Test with branch name context (WU002 scenario)
- [ ] Verify external editor (Ctrl+E) works

### Phase 8: Finalization Enhancement
- [ ] Modify project initialization to accept issue metadata
- [ ] Store issue number, title, URL in project state
- [ ] Include issue context in generated prompt
- [ ] Test project state contains issue metadata

### Phase 9: Error Handling Polish
- [ ] Implement all error messages from design doc
- [ ] Test error recovery flows
- [ ] Verify user can navigate back after errors
- [ ] Test graceful degradation when GitHub unavailable

### Phase 10: Integration Testing
- [ ] Test complete workflow end-to-end
- [ ] Test with real GitHub API (manual)
- [ ] Test all error scenarios
- [ ] Verify Claude launches with correct context

---

## 9. Open Questions / Decisions Needed

### Q1: Should we support issue search/filtering beyond label?
**Current Scope**: Only fetch issues with 'sow' label
**Possible Enhancement**: Allow user to search by keyword, filter by milestone, assignee, etc.
**Recommendation**: Keep current scope, add as future enhancement if needed

### Q2: What to do if issue body is very long?
**Context**: Issue body could be included in Claude prompt
**Options**:
  A. Include full body in prompt (could be very long)
  B. Truncate to first N characters
  C. Don't include body, only title and number
**Recommendation**: Option C for initial implementation - only title/number. Can add body in future.

### Q3: Should we validate issue state (open vs closed)?
**Current Scope**: Fetch only open issues (`--state open`)
**Question**: Should we prevent creating projects from closed issues?
**Recommendation**: Current scope is correct - only show open issues in list

### Q4: How to handle multiple linked branches?
**Context**: An issue could theoretically have multiple linked branches
**Current Logic**: If any linked branches exist, show error
**Alternative**: Allow user to select which linked branch to continue
**Recommendation**: Keep current scope - if linked branches exist, suggest using "Continue existing project"

---

## 10. Success Metrics

After implementation, this work unit succeeds if:

1. **Functional**: All 10 acceptance criteria pass
2. **Tested**: All unit tests, integration tests pass; manual scenarios completed
3. **User-Friendly**: Errors are helpful, flow is intuitive, no dead ends
4. **Integrated**: Works seamlessly with WU001 foundation and WU002 branch name path
5. **Documented**: Code is well-commented, complex logic explained

---

## 11. Related Work Units

- **Work Unit 001**: Foundation (state machine, helpers) - PREREQUISITE
- **Work Unit 002**: Branch name path (type selection, prompt entry, finalization) - PREREQUISITE for shared code
- **Work Unit 004**: Project continuation - Will reference issue context if present
- **Work Unit 005**: Validation utilities - May enhance issue validation

---

## 12. References

### Design Documents
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` (lines 60-174: GitHub issue flow)
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` (implementation patterns)
- `.sow/knowledge/designs/huh-library-verification.md` (Ctrl+E keybinding, spinner usage)

### Existing Code
- `cli/internal/sow/github.go` (complete GitHub client implementation)
- `cli/internal/sow/worktree.go` (worktree management)
- `cli/cmd/project/wizard_state.go` (state machine from WU001)
- `cli/cmd/project/shared.go` (finalization logic from WU002)

### External Documentation
- GitHub CLI: https://cli.github.com/
- `gh issue develop` command: https://cli.github.com/manual/gh_issue_develop
- huh library: https://github.com/charmbracelet/huh
- huh spinner: https://github.com/charmbracelet/huh/tree/main/spinner

---

**End of Specification**
