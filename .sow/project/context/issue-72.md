# Issue #72: Validation and Error Handling

**URL**: https://github.com/jmgilman/sow/issues/72
**State**: OPEN

## Description

# Work Unit 005: Validation and Error Handling

## Behavioral Goal (User Story)

As a user of the interactive wizard, I need comprehensive validation and clear error messages at every step so that I can understand what went wrong, why it matters, and exactly how to fix it, enabling me to successfully create or continue projects without confusion or frustration.

**Success Criteria for Reviewers:**
- Every validation function returns clear, actionable error messages
- All error scenarios from the design document are handled with appropriate messages
- Users can recover from every error state through back navigation, retry, or alternative paths
- Error messages follow the consistent 3-part pattern: what went wrong, how to fix, next steps
- All test cases pass, including edge cases for protected branches, uncommitted changes, GitHub errors, and state inconsistencies

## Existing Code Context

### Explanation

This work unit builds the comprehensive validation and error handling layer for the interactive wizard. The wizard (Work Unit 001) provides the UI framework and error display utilities (like `showError` and `showErrorWithOptions`), and this work unit creates all the validation logic and error messages that the wizard will use.

The existing codebase already has some validation primitives that we'll reuse and extend:
- `cli/internal/sow/git.go` contains `IsProtectedBranch()` which checks for "main" and "master"
- `cli/internal/sow/worktree.go` has `CheckUncommittedChanges()` which validates the repository state
- `cli/internal/sow/github.go` provides error types for GitHub operations (`ErrGHNotInstalled`, `ErrGHNotAuthenticated`, `ErrGHCommand`)

We'll extend these patterns to create a comprehensive validation suite that:
1. Validates branch names using git ref rules
2. Checks branch/worktree/project state combinations
3. Detects uncommitted changes only when needed (conditional check)
4. Handles all GitHub integration errors with clear recovery paths
5. Validates issue state (linked branches, missing labels)
6. Detects and reports project state corruption

Work Units 002-004 (the specific wizard flows) will use these validation functions throughout their implementations. This creates a clean separation: validation logic lives here, workflow logic lives in 002-004.

### Reference List

**Existing validation utilities to extend:**
- `cli/internal/sow/git.go:73-92` - IsProtectedBranch function (checks main/master)
- `cli/internal/sow/worktree.go:89-122` - CheckUncommittedChanges function (validates repository state)
- `cli/internal/sow/github.go:39-71` - GitHub error types (ErrGHNotInstalled, ErrGHNotAuthenticated, ErrGHCommand)

**Existing patterns to follow:**
- `cli/internal/sow/github.go:438-465` - toKebabCase helper function (pattern for name normalization)
- `cli/internal/sow/github.go:104-136` - Ensure() pattern (check installation then authentication)

**Integration points:**
- Work Unit 001 (foundation) provides error display utilities
- Work Units 002-004 (flows) will call these validation functions

## Existing Documentation Context

### UX Flow Design (Error Messages Section)

The UX design document (`.sow/knowledge/designs/interactive-wizard-ux-flow.md`, lines 421-535) specifies the **exact error messages** that must be implemented word-for-word. This is critical because these messages were carefully crafted to follow the 3-part pattern and provide users with actionable guidance.

**All error messages to implement verbatim (lines 421-535):**

1. **Protected Branch Error** (lines 429-443): User tried to create project on "main" or "master"
2. **Issue Already Linked** (lines 445-458): GitHub issue has an existing linked branch
3. **Branch Already Has Project** (lines 460-476): Branch has .sow/project/ directory already
4. **Uncommitted Changes** (lines 478-495): Repository has uncommitted changes and worktree creation would require switching branches
5. **Inconsistent State** (lines 497-515): Worktree exists but project is missing
6. **GitHub CLI Missing** (lines 517-535): `gh` command not found

### Technical Implementation Design (Validation Rules)

The technical design document (`.sow/knowledge/designs/interactive-wizard-technical-implementation.md`, lines 480-646) provides:

**Validation Rules** (lines 380-535):
- Branch name normalization algorithm (lines 570-599)
- Branch name validation rules (lines 604-646)
- Conditional uncommitted changes check logic (lines 480-516)
- State detection patterns (lines 356-402)

**Test Cases** (lines 706-755):
- Name normalization test cases with expected outputs
- Branch validation test cases (valid and invalid examples)

### Library Verification (huh Integration)

The huh library verification document (`.sow/knowledge/designs/huh-library-verification.md`) confirms that the huh library supports inline validation errors, which we'll use for field-level validation feedback.

**Critical finding:** External editor uses Ctrl+E (not Ctrl+O as originally specified).

## Implementation Scope

### 1. Branch Name Validation Module

**File:** `cli/cmd/project/wizard/validation.go`

Implement the following validation functions:

```go
// normalizeName converts user input to valid git branch name component
// Rules (from design lines 570-599):
// 1. Convert to lowercase
// 2. Replace spaces with hyphens
// 3. Remove invalid characters (keep only a-z, 0-9, -, _)
// 4. Collapse multiple consecutive hyphens
// 5. Remove leading/trailing hyphens
func normalizeName(name string) string

// isValidBranchName validates a complete branch name against git ref rules
// Rules (from design lines 604-646):
// - Not empty or whitespace
// - Not protected branch (main, master)
// - Valid git ref name:
//   - No spaces, .., //, ~, ^, :, ?, *, [
//   - No leading/trailing slashes
//   - No consecutive slashes
func isValidBranchName(name string) error

// validateProjectName validates user input for project name entry
// Called by name entry screen (Work Unit 002)
func validateProjectName(name string, prefix string) error
```

### 2. Branch State Validation Module

**File:** `cli/cmd/project/wizard/state.go`

Implement state checking functions:

```go
// BranchState represents the current state of a branch
type BranchState struct {
    BranchExists  bool
    WorktreeExists bool
    ProjectExists bool
}

// checkBranchState examines branch, worktree, and project state
// Used before creation to detect conflicts
func checkBranchState(ctx *sow.Context, branchName string) (*BranchState, error)

// canCreateProject validates that project creation is allowed
// Returns clear error if branch already has project
func canCreateProject(state *BranchState) error

// validateProjectExists checks that a project at given branch still exists
// Used when continuing projects (Work Unit 004)
func validateProjectExists(ctx *sow.Context, branchName string) error
```

### 3. Uncommitted Changes Validation

**File:** `cli/cmd/project/wizard/validation.go` (same file as branch validation)

Implement conditional validation:

```go
// shouldCheckUncommittedChanges determines if validation is needed
// Only returns true when current branch == target branch
// Logic from design lines 480-516
func shouldCheckUncommittedChanges(ctx *sow.Context, targetBranch string) (bool, error)

// performUncommittedChangesCheckIfNeeded runs validation conditionally
// Uses existing sow.CheckUncommittedChanges() from worktree.go
// Returns enhanced error message with fix instructions
func performUncommittedChangesCheckIfNeeded(ctx *sow.Context, targetBranch string) error
```

### 4. GitHub Integration Error Handling

**File:** `cli/cmd/project/wizard/github_errors.go`

Wrap existing GitHub error types with user-friendly messages:

```go
// checkGitHubCLI validates gh is installed and authenticated
// Returns user-friendly error with installation/auth instructions
func checkGitHubCLI(gh *sow.GitHub) error

// formatGitHubError converts GitHub errors to user-friendly messages
// Handles:
// - ErrGHNotInstalled: Show installation instructions
// - ErrGHNotAuthenticated: Show gh auth login instructions
// - ErrGHCommand: Parse stderr for specific error cases
//   - Network errors: "check connection, retry"
//   - Rate limit: "wait or authenticate for higher limit"
func formatGitHubError(err error) string
```

### 5. Issue Validation

**File:** `cli/cmd/project/wizard/github_errors.go` (same as GitHub errors)

Implement issue-specific validation:

```go
// checkIssueLinkedBranch validates issue doesn't have existing linked branch
// Returns error with branch name and suggestion to use "continue" path
func checkIssueLinkedBranch(gh *sow.GitHub, issueNumber int) error

// filterIssuesBySowLabel ensures only issues with 'sow' label are shown
// This is a utility for issue listing (Work Unit 003)
func filterIssuesBySowLabel(issues []sow.Issue) []sow.Issue
```

### 6. Error Display Components

**File:** `cli/cmd/project/wizard/errors.go`

Implement error display utilities (used by Work Unit 001):

```go
// showError displays a simple error with acknowledgment
// Uses huh.NewConfirm for "Press Enter to continue" pattern
func showError(message string) error

// showErrorWithOptions displays error with multiple action choices
// Returns the user's choice
// Used for errors with multiple recovery paths
func showErrorWithOptions(message string, options map[string]string) (string, error)

// formatError formats error message in 3-part pattern:
// 1. What went wrong
// 2. How to fix
// 3. Next steps
func formatError(title, problem, howToFix, nextSteps string) string
```

### 7. All Error Messages from Design Document

Implement these exact error messages from lines 421-535:

**Protected Branch Error:**
```
Cannot create project on protected branch 'main'

Projects must be created on feature branches.

Action: Choose a different project name

[Press Enter to retry]
```

**Issue Already Linked:**
```
Issue #123 already has a linked branch: feat/add-jwt-auth

To continue working on this issue:
  Select "Continue existing project" from the main menu

[Press Enter to return to issue list]
```

**Branch Already Has Project:**
```
Branch 'explore/web-agents' already has a project

To continue this project:
  Select "Continue existing project" from the main menu

To create a different project:
  Choose a different project name (currently: "web agents")

[Press Enter to retry name entry]
```

**Uncommitted Changes:**
```
Repository has uncommitted changes

You are currently on branch 'feat/add-jwt-auth-123'.
Creating a worktree requires switching to a different branch first.

To fix:
  Commit: git add . && git commit -m "message"
  Or stash: git stash

[Press Enter to exit wizard]
```

**Inconsistent State:**
```
Worktree exists but project missing

Branch 'feat/xyz' has a worktree at .sow/worktrees/feat/xyz
but no .sow/project/ directory.

To fix:
  1. Remove worktree: git worktree remove feat/xyz
  2. Delete directory: rm -rf .sow/worktrees/feat/xyz
  3. Try creating project again

[Press Enter to return to project list]
```

**GitHub CLI Missing:**
```
GitHub CLI not found

The 'gh' command is required for GitHub issue integration.

To install:
  macOS: brew install gh
  Linux: See https://cli.github.com/

Or select "From branch name" instead.

[Press Enter to return to source selection]
```

## Acceptance Criteria

### Functional Criteria

1. **Branch Name Normalization Works:**
   - "Web Based Agents" → "web-based-agents"
   - "API V2" → "api-v2"
   - "feature--name" → "feature-name"
   - "-leading-trailing-" → "leading-trailing"
   - "With!Invalid@Chars#" → "withinvalidchars"
   - All test cases from design doc lines 706-727 pass

2. **Branch Validation Works:**
   - "feat/valid-name" passes ✓
   - "explore/test" passes ✓
   - "main" fails with protected branch error ✗
   - "master" fails with protected branch error ✗
   - "has spaces" fails ✗
   - "has..dots" fails ✗
   - "has//slashes" fails ✗
   - "/leading-slash" fails ✗
   - "trailing-slash/" fails ✗
   - All test cases from design doc lines 731-755 pass

3. **State Checking Detects All Combinations:**
   - Branch exists + worktree exists + project exists → error
   - Branch exists + worktree exists + no project → inconsistent state error
   - Branch exists + no worktree + no project → OK (can create)
   - No branch + no worktree + no project → OK (can create)

4. **Uncommitted Changes Validation:**
   - Only runs when current branch == target branch
   - Ignores untracked files (?)
   - Detects modified files (M)
   - Detects deleted files (D)
   - Detects staged changes
   - Returns error with fix commands

5. **GitHub Error Handling:**
   - Detects `gh` not installed → installation instructions
   - Detects not authenticated → `gh auth login` instructions
   - Detects network errors → retry suggestion
   - Detects rate limit → wait instructions
   - All errors include recovery path

6. **Issue Validation:**
   - Detects linked branch → error with branch name
   - Filters out issues without 'sow' label
   - Handles missing/deleted issues gracefully

7. **Error Messages Match Design:**
   - All 6 error messages from design doc implemented word-for-word
   - All follow 3-part pattern (what, how, next)
   - All include appropriate recovery actions

### Testing Requirements (TDD Approach)

**Unit Tests** - Write these FIRST, then implement to pass them:

1. **Name Normalization Tests** (`validation_test.go`):
   ```go
   func TestNormalizeName(t *testing.T)
   ```
   - Test all cases from design lines 706-727
   - Test empty string, whitespace-only
   - Test already-normalized names
   - Test extreme cases (all invalid chars, all hyphens)

2. **Branch Validation Tests** (`validation_test.go`):
   ```go
   func TestIsValidBranchName(t *testing.T)
   func TestValidateProjectName(t *testing.T)
   ```
   - Test all cases from design lines 731-755
   - Test each invalid pattern individually
   - Test error messages are clear

3. **State Checking Tests** (`state_test.go`):
   ```go
   func TestCheckBranchState(t *testing.T)
   func TestCanCreateProject(t *testing.T)
   ```
   - Mock filesystem and git operations
   - Test all state combinations
   - Test error detection and messages

4. **Uncommitted Changes Tests** (`validation_test.go`):
   ```go
   func TestShouldCheckUncommittedChanges(t *testing.T)
   func TestPerformUncommittedChangesCheckIfNeeded(t *testing.T)
   ```
   - Test current == target (should check)
   - Test current != target (should skip)
   - Test error message formatting

5. **GitHub Error Tests** (`github_errors_test.go`):
   ```go
   func TestCheckGitHubCLI(t *testing.T)
   func TestFormatGitHubError(t *testing.T)
   func TestCheckIssueLinkedBranch(t *testing.T)
   ```
   - Mock GitHub client
   - Test each error type
   - Verify message formatting

6. **Error Display Tests** (`errors_test.go`):
   ```go
   func TestFormatError(t *testing.T)
   ```
   - Test 3-part pattern formatting
   - Test with all design doc error messages

### Manual Testing Scenarios

After implementation, manually verify:

1. **Trigger each error:**
   - Try creating project on "main" branch
   - Try creating project on branch with existing project
   - Try creating project with uncommitted changes (when current == target)
   - Try GitHub flow without `gh` installed
   - Try GitHub flow when not authenticated
   - Try selecting issue that has linked branch
   - Create worktree manually, delete .sow/project/, try to continue

2. **Verify error messages:**
   - Each error shows correct message from design doc
   - Messages are readable and actionable
   - Recovery options work correctly

3. **Verify recovery flows:**
   - Back navigation works from error states
   - Retry allows user to fix and continue
   - Alternative paths (e.g., branch name when gh missing) work

## Technical Details

### Language and Framework

- **Language:** Go
- **Packages:**
  - `cli/cmd/project/wizard` - New package for wizard validation
  - Imports: `cli/internal/sow`, `strings`, `fmt`, `os`, `path/filepath`
  - huh library for error display components
  - Existing git wrapper and GitHub client

### File Structure

```
cli/cmd/project/wizard/
├── validation.go         # Branch name validation, uncommitted changes
├── validation_test.go    # Tests for validation functions
├── state.go             # Branch/worktree/project state checking
├── state_test.go        # Tests for state checking
├── github_errors.go     # GitHub-specific error handling
├── github_errors_test.go # Tests for GitHub errors
├── errors.go            # Error display utilities
└── errors_test.go       # Tests for error display
```

### Code Patterns

**Validation Function Pattern:**
```go
// Validation functions return error with user-friendly message
func validateSomething(input string) error {
    if invalid(input) {
        return fmt.Errorf("clear description of what's wrong")
    }
    return nil
}
```

**State Checking Pattern:**
```go
// State checking returns struct with boolean flags
type SomeState struct {
    ConditionA bool
    ConditionB bool
}

func checkState(ctx *sow.Context) (*SomeState, error) {
    state := &SomeState{}
    // Check conditions...
    return state, nil
}
```

**Error Display Pattern:**
```go
// Error display uses huh library components
func showError(message string) error {
    return huh.NewForm(
        huh.NewGroup(
            huh.NewNote().Title("Error").Description(message),
            huh.NewConfirm().Title("Press Enter to continue"),
        ),
    ).Run()
}
```

### Git Ref Validation Rules

From git documentation and design doc:

**Invalid patterns:**
- Spaces (anywhere)
- `..` (two consecutive dots)
- `//` (two consecutive slashes)
- `~`, `^`, `:`, `?`, `*`, `[` (special characters)
- Leading or trailing `/`
- Component starts with `.` (except for special refs)

**Valid characters:**
- Alphanumeric: `a-z`, `A-Z`, `0-9`
- Separators: `/`, `-`, `_`
- Dots: `.` (but not `..` and not at start of component)

### Integration with Existing Code

**Reuse these existing functions:**
- `sow.Context.Git().IsProtectedBranch(branch)` - Check for main/master
- `sow.CheckUncommittedChanges(ctx)` - Detect uncommitted changes
- `sow.GitHub.CheckInstalled()` - Verify gh CLI exists
- `sow.GitHub.CheckAuthenticated()` - Verify gh is logged in
- `sow.GitHub.GetLinkedBranches(issueNumber)` - Check for linked branches

**Extend with user-friendly wrappers:**
- Wrap low-level checks with error messages from design doc
- Add conditional logic (e.g., uncommitted changes only when needed)
- Format errors in 3-part pattern

## Relevant Inputs

### Design Documents

- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` - **Critical: Lines 380-535 contain all validation rules and exact error messages**
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - **Critical: Lines 480-646 contain validation logic, lines 706-755 contain test cases**
- `.sow/knowledge/designs/huh-library-verification.md` - Library capabilities for error display

### Existing Validation Code

- `cli/internal/sow/git.go` - IsProtectedBranch function (lines 73-92)
- `cli/internal/sow/worktree.go` - CheckUncommittedChanges function (lines 89-122), EnsureWorktree logic
- `cli/internal/sow/github.go` - Error types (lines 39-71), toKebabCase helper (lines 438-465)

### Discovery Document

- `.sow/project/discovery/analysis.md` - Overview of existing patterns and architecture

## Examples

### Example 1: Name Normalization

```go
// Input: "Web Based Agents"
// Output: "web-based-agents"

// Implementation (from design lines 570-599):
func normalizeName(name string) string {
    // 1. Trim and convert to lowercase
    name = strings.ToLower(strings.TrimSpace(name))

    // 2. Replace spaces with hyphens
    name = strings.ReplaceAll(name, " ", "-")

    // 3. Remove invalid characters
    var result strings.Builder
    for _, r := range name {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
            result.WriteRune(r)
        }
    }
    name = result.String()

    // 4. Collapse consecutive hyphens
    for strings.Contains(name, "--") {
        name = strings.ReplaceAll(name, "--", "-")
    }

    // 5. Remove leading/trailing hyphens
    name = strings.Trim(name, "-")

    return name
}
```

### Example 2: Branch Validation with Clear Error

```go
func isValidBranchName(name string) error {
    if name == "" {
        return fmt.Errorf("branch name cannot be empty")
    }

    if name == "main" || name == "master" {
        return fmt.Errorf("cannot use protected branch name")
    }

    // Check for invalid patterns
    invalidPatterns := map[string]string{
        "..": "consecutive dots",
        "//": "consecutive slashes",
        " ": "spaces",
    }

    for pattern, description := range invalidPatterns {
        if strings.Contains(name, pattern) {
            return fmt.Errorf("branch name cannot contain %s", description)
        }
    }

    // Check for invalid characters
    invalidChars := "~^:?*["
    for _, char := range invalidChars {
        if strings.ContainsRune(name, char) {
            return fmt.Errorf("branch name contains invalid character: %c", char)
        }
    }

    // Check for leading/trailing slashes
    if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
        return fmt.Errorf("branch name cannot start or end with /")
    }

    return nil
}
```

### Example 3: State Checking

```go
func checkBranchState(ctx *sow.Context, branchName string) (*BranchState, error) {
    state := &BranchState{}

    // Check if branch exists
    repo := ctx.Git().Repository()
    branches, err := repo.Branches()
    if err != nil {
        return nil, err
    }
    for _, b := range branches {
        if b == branchName {
            state.BranchExists = true
            break
        }
    }

    // Check if worktree exists
    worktreePath := sow.WorktreePath(ctx.RepoRoot(), branchName)
    if _, err := os.Stat(worktreePath); err == nil {
        state.WorktreeExists = true

        // Check if project exists in worktree
        projectPath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")
        if _, err := os.Stat(projectPath); err == nil {
            state.ProjectExists = true
        }
    }

    return state, nil
}

func canCreateProject(state *BranchState) error {
    if state.ProjectExists {
        return fmt.Errorf("project already exists on this branch")
    }
    if state.WorktreeExists && !state.ProjectExists {
        return fmt.Errorf("worktree exists but project is missing - inconsistent state")
    }
    return nil
}
```

### Example 4: Error Display with Recovery Options

```go
func handleIssueLinkedError(branchName string) error {
    message := formatError(
        "Issue Already Linked",
        fmt.Sprintf("This issue already has a linked branch: %s", branchName),
        "To continue working on this issue:\n  Select \"Continue existing project\" from the main menu",
        "",
    )

    return showError(message)
}
```

## Dependencies

### Required Before This Work Unit

- **Work Unit 001** (Wizard Foundation): Provides error display utilities (`showError`, `showErrorWithOptions`) that this work unit implements

### Provides For Other Work Units

- **Work Unit 002** (Branch Name Flow): Uses `validateProjectName`, `normalizeName`, `checkBranchState`, `performUncommittedChangesCheckIfNeeded`
- **Work Unit 003** (GitHub Issue Flow): Uses `checkGitHubCLI`, `checkIssueLinkedBranch`, `filterIssuesBySowLabel`, `formatGitHubError`
- **Work Unit 004** (Continue Flow): Uses `validateProjectExists`, `checkBranchState`

### External Dependencies

- Go standard library: `strings`, `fmt`, `os`, `path/filepath`
- huh library: For error display components
- Existing sow packages: `cli/internal/sow`, `cli/internal/exec`

## Constraints

### Error Message Immutability

The error messages from the design document (lines 421-535) **must be implemented word-for-word**. These messages were carefully crafted during the design phase and reviewed for clarity, actionability, and tone. Do not paraphrase or "improve" them.

### Performance Requirements

- Name normalization must be fast (called on every keystroke for real-time preview)
- State checking should complete in < 100ms for good UX
- Cache GitHub CLI installation check (don't shell out repeatedly)

### Security Considerations

- Branch name validation must prevent git ref injection attacks
- Path traversal prevention in worktree path checking
- Don't expose sensitive error details (e.g., full file paths in production errors)

### Backward Compatibility

This is new code for the interactive wizard. No backward compatibility concerns with existing CLI flags.

### What NOT to Do

- **Don't modify existing validation in other packages** - Extend, don't change
- **Don't create separate "write tests" task** - Tests are part of TDD implementation
- **Don't add validation that's not in the design** - Scope creep prevention
- **Don't implement the wizard UI** - That's Work Unit 001
- **Don't implement workflow logic** - That's Work Units 002-004

## Notes for Implementer

### Critical UX Pattern

Every error message follows this 3-part structure:
1. **What went wrong**: Brief, clear statement of the problem
2. **How to fix**: Specific commands or actions to resolve it
3. **Next steps**: What the user should do now (retry, go back, etc.)

This pattern is already embedded in the error messages from the design doc. Your job is to implement them exactly as specified.

### Defensive Validation Philosophy

The design emphasizes "defensive validation" - don't assume clean input. Validate everything before operations that could fail or corrupt state. Better to catch errors early with clear messages than to fail mysteriously later.

### Recovery is Key

Every error should have a clear path forward:
- **Retry**: User can fix the issue and try again (e.g., change branch name)
- **Alternative path**: Different way to accomplish the goal (e.g., branch name instead of GitHub issue)
- **Exit cleanly**: User can cancel without corruption

### Test-Driven Development

Write tests FIRST for each validation function:
1. Write test cases (including all from design doc)
2. Run tests (they will fail)
3. Implement validation function
4. Run tests until they pass
5. Refactor if needed

This ensures complete test coverage and validates your understanding of requirements before coding.

### Size Estimate

This is a **2-3 day project-sized work unit** because it includes:
- 6 validation modules (normalization, validation, state, GitHub, errors, display)
- Complete test suite with ~50+ test cases
- Error message formatting and display utilities
- Integration with existing codebase patterns

Take time to get the error messages right - they're critical to user experience.
