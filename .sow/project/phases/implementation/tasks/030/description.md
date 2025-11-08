# Task 030: Error Display Components and Message Formatting

## Context

This task is part of Work Unit 005: Validation and Error Handling for the interactive wizard. The wizard needs to display clear, actionable error messages when validation fails. This task implements the error display infrastructure that transforms simple validation errors into user-friendly, formatted messages following the 3-part pattern.

**The 3-part error pattern** (from UX design lines 424-428):
1. **What went wrong**: Brief description of the problem
2. **How to fix**: Specific commands or actions to resolve it
3. **Next steps**: What the user should do now (retry, go back, exit)

This task creates the formatting functions and huh library integrations that will be used by all wizard flows (Work Units 002-004) to display validation errors consistently.

**Architecture note**: This module handles error formatting and display only. Error detection is handled by Tasks 010 and 020. This separation keeps concerns clean: validation modules detect problems, this module communicates them to users.

## Requirements

### 1. Error Formatting Function

Create `formatError()` in `wizard_helpers.go` to format error messages in the 3-part pattern:

```go
// formatError formats error messages in the consistent 3-part pattern:
//   1. What went wrong (title)
//   2. How to fix (problem and solution)
//   3. Next steps (what to do now)
//
// The function assembles the parts into a single formatted string suitable
// for display in huh components.
//
// Example:
//   msg := formatError(
//       "Cannot create project on protected branch 'main'",
//       "Projects must be created on feature branches.",
//       "Action: Choose a different project name",
//   )
func formatError(problem string, howToFix string, nextSteps string) string
```

**Implementation**:
- Combine the three parts with newlines for readability
- Use double newlines between sections for visual separation
- Don't add decorative borders (huh handles UI formatting)
- Return single string ready for huh display components

**Format structure**:
```
{problem}

{howToFix}

{nextSteps}
```

### 2. Simple Error Display

Create `showError()` in `wizard_helpers.go` to display errors with a simple "Press Enter to continue" acknowledgment:

```go
// showError displays a simple error message with acknowledgment.
// Uses huh.NewForm with a Note and Confirm for "Press Enter to continue".
//
// This is used for errors where the user just needs to acknowledge and
// try again (e.g., validation errors during name entry).
//
// Example:
//   if err := validateProjectName(name); err != nil {
//       showError(formatError("Invalid name", err.Error(), "Try again"))
//       return // retry
//   }
func showError(message string) error
```

**Implementation using huh library**:
```go
func showError(message string) error {
    confirm := false
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewNote().
                Title("Error").
                Description(message),
            huh.NewConfirm().
                Title("Press Enter to continue").
                Value(&confirm).
                Affirmative("Continue"),
        ),
    )
    return form.Run()
}
```

**Why Confirm instead of just Note?** Huh requires user interaction to dismiss. Using Confirm with a single "Continue" option provides a clean "Press Enter" experience.

### 3. Error Display with Options

Create `showErrorWithOptions()` in `wizard_helpers.go` to display errors with multiple recovery paths:

```go
// showErrorWithOptions displays an error with multiple action choices.
// Returns the user's selected choice.
//
// Used for errors where multiple recovery paths are available
// (e.g., "continue existing project" vs "change name").
//
// Example:
//   choice, err := showErrorWithOptions(
//       formatError(...),
//       map[string]string{
//           "retry": "Change project name",
//           "continue": "Continue existing project",
//           "cancel": "Cancel",
//       },
//   )
func showErrorWithOptions(message string, options map[string]string) (string, error)
```

**Implementation using huh library**:
```go
func showErrorWithOptions(message string, options map[string]string) (string, error) {
    var selected string

    // Convert map to huh options
    huhOptions := make([]huh.Option[string], 0, len(options))
    for key, label := range options {
        huhOptions = append(huhOptions, huh.NewOption(label, key))
    }

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewNote().
                Title("Error").
                Description(message),
            huh.NewSelect[string]().
                Title("What would you like to do?").
                Options(huhOptions...).
                Value(&selected),
        ),
    )

    if err := form.Run(); err != nil {
        return "", err
    }

    return selected, nil
}
```

### 4. Specific Error Messages

Implement helper functions in `wizard_helpers.go` for the 6 specific error scenarios from the design document (lines 421-535). Each function returns a formatted error message string:

**4.1 Protected Branch Error**:
```go
// errorProtectedBranch returns the formatted error message for attempting
// to create a project on a protected branch (main or master).
func errorProtectedBranch(branchName string) string
```

**Message** (exact text from design lines 436-442):
```
Cannot create project on protected branch '{branchName}'

Projects must be created on feature branches.

Action: Choose a different project name
```

**4.2 Issue Already Linked Error**:
```go
// errorIssueAlreadyLinked returns the formatted error message when a GitHub
// issue already has a linked branch.
func errorIssueAlreadyLinked(issueNumber int, linkedBranch string) string
```

**Message** (exact text from design lines 452-457):
```
Issue #{issueNumber} already has a linked branch: {linkedBranch}

To continue working on this issue:
  Select "Continue existing project" from the main menu
```

**4.3 Branch Already Has Project Error**:
```go
// errorBranchHasProject returns the formatted error message when attempting
// to create a project on a branch that already has one.
func errorBranchHasProject(branchName string, projectName string) string
```

**Message** (exact text from design lines 467-475):
```
Branch '{branchName}' already has a project

To continue this project:
  Select "Continue existing project" from the main menu

To create a different project:
  Choose a different project name (currently: "{projectName}")
```

**4.4 Uncommitted Changes Error**:
```go
// errorUncommittedChanges returns the formatted error message when the
// repository has uncommitted changes and worktree creation requires switching branches.
func errorUncommittedChanges(currentBranch string) string
```

**Message** (exact text from design lines 485-493):
```
Repository has uncommitted changes

You are currently on branch '{currentBranch}'.
Creating a worktree requires switching to a different branch first.

To fix:
  Commit: git add . && git commit -m "message"
  Or stash: git stash
```

**4.5 Inconsistent State Error**:
```go
// errorInconsistentState returns the formatted error message when a worktree
// exists but the project directory is missing.
func errorInconsistentState(branchName string, worktreePath string) string
```

**Message** (exact text from design lines 504-513):
```
Worktree exists but project missing

Branch '{branchName}' has a worktree at {worktreePath}
but no .sow/project/ directory.

To fix:
  1. Remove worktree: git worktree remove {branchName}
  2. Delete directory: rm -rf {worktreePath}
  3. Try creating project again
```

**4.6 GitHub CLI Missing Error**:
```go
// errorGitHubCLIMissing returns the formatted error message when the gh
// command is not installed.
func errorGitHubCLIMissing() string
```

**Message** (exact text from design lines 524-533):
```
GitHub CLI not found

The 'gh' command is required for GitHub issue integration.

To install:
  macOS: brew install gh
  Linux: See https://cli.github.com/

Or select "From branch name" instead.
```

### 5. Integration with Validation Errors

Create helper functions in `wizard_helpers.go` to wrap validation errors with formatted messages:

```go
// wrapValidationError wraps a validation error with user-friendly formatting.
// If err is nil, returns nil.
// Otherwise, wraps with formatError and returns displayable error.
func wrapValidationError(err error, context string) error
```

**Purpose**: Convert simple validation errors like "branch name contains spaces" into formatted messages with context and next steps.

## Acceptance Criteria

### Functional Requirements

1. **Error Formatting Works**:
   - `formatError()` combines three parts with proper spacing
   - Output is readable and matches design doc format
   - Handles empty strings gracefully

2. **Simple Error Display Works**:
   - `showError()` displays message in huh UI
   - User can dismiss with Enter key
   - Returns without error after acknowledgment

3. **Error with Options Works**:
   - `showErrorWithOptions()` displays message and choices
   - Returns selected option key
   - Handles all option map variations

4. **Specific Error Messages Match Design**:
   - All 6 error message functions return EXACT text from design doc
   - Variable substitution works correctly (branch names, issue numbers)
   - Messages are word-for-word matches (no paraphrasing)

5. **Validation Error Wrapping Works**:
   - Wraps simple errors with context
   - Returns nil when input is nil
   - Adds appropriate next steps

### Test Requirements (TDD Approach)

**Write ALL tests FIRST, then implement functions to pass them.**

1. **Error Formatting Tests** (`wizard_helpers_test.go`):
   ```go
   func TestFormatError(t *testing.T)
   ```
   - Test with all three parts provided
   - Test with empty parts
   - Test spacing and newlines
   - At least 5 test cases

2. **Specific Error Message Tests** (`wizard_helpers_test.go`):
   ```go
   func TestErrorProtectedBranch(t *testing.T)
   func TestErrorIssueAlreadyLinked(t *testing.T)
   func TestErrorBranchHasProject(t *testing.T)
   func TestErrorUncommittedChanges(t *testing.T)
   func TestErrorInconsistentState(t *testing.T)
   func TestErrorGitHubCLIMissing(t *testing.T)
   ```
   - Test each error message matches design doc EXACTLY
   - Test variable substitution (branch names, numbers)
   - Use string comparison to verify exact match
   - At least 15 test cases total (2-3 per error message)

3. **Error Display Tests** (`wizard_helpers_test.go`):
   ```go
   func TestShowError(t *testing.T)
   func TestShowErrorWithOptions(t *testing.T)
   ```
   - These are harder to test due to terminal UI
   - Focus on ensuring they compile and accept correct inputs
   - Consider manual testing for UI behavior
   - At least 3 test cases

4. **Validation Wrapping Tests** (`wizard_helpers_test.go`):
   ```go
   func TestWrapValidationError(t *testing.T)
   ```
   - Test wrapping real validation errors
   - Test nil input → nil output
   - Test context is added
   - At least 4 test cases

### Code Quality

- All error messages match design doc word-for-word
- Functions have clear godoc comments
- Variable names are descriptive
- Tests verify exact text (no fuzzy matching)

## Technical Details

### File Location

**Add functions to**: `/cli/cmd/project/wizard_helpers.go`
**Add tests to**: `/cli/cmd/project/wizard_helpers_test.go`

### Imports Required

```go
import (
    "fmt"
    "strings"
    "github.com/charmbracelet/huh"
)
```

### Huh Library Integration

**Note component** (for messages):
```go
huh.NewNote().
    Title("Error").
    Description(message)
```

**Confirm component** (for acknowledgment):
```go
var confirm bool
huh.NewConfirm().
    Title("Press Enter to continue").
    Value(&confirm).
    Affirmative("Continue")
```

**Select component** (for options):
```go
var selected string
huh.NewSelect[string]().
    Title("What would you like to do?").
    Options(
        huh.NewOption("Label", "value"),
        // ...
    ).
    Value(&selected)
```

**Form assembly**:
```go
form := huh.NewForm(
    huh.NewGroup(
        // Components...
    ),
)
err := form.Run()
```

### Text Formatting Standards

**From design document**:
- Use double newlines between sections
- Indent lists with 2 spaces
- Use colons after section headers ("To fix:", "Action:")
- Quote branch names with single quotes: 'feat/xyz'
- Use Issue numbers with hash: #123
- Keep line length reasonable (< 80 chars where possible)

### Error Message Immutability

**CRITICAL**: The error messages from the design document (lines 421-535) must be implemented word-for-word. These messages were carefully crafted during the design phase and reviewed for:
- Clarity
- Actionability
- Tone
- Completeness

Do NOT paraphrase or "improve" them. Implement exactly as specified.

## Relevant Inputs

**Design documents** (contain exact error messages):
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` (lines 421-535: ALL error messages)
- `.sow/knowledge/designs/huh-library-verification.md` (huh component patterns)

**Existing implementation**:
- `cli/cmd/project/wizard_helpers.go` (file to extend)
- `cli/cmd/project/wizard_helpers_test.go` (test file to extend)
- `cli/cmd/project/wizard_state.go` (will call these functions)

**Related tasks**:
- Task 010: Branch name validation (produces errors to format)
- Task 020: State validation (produces errors to format)
- Task 040: GitHub error handling (produces errors to format)

## Examples

### Example 1: Format Error Implementation

```go
// formatError formats error messages in the consistent 3-part pattern.
func formatError(problem string, howToFix string, nextSteps string) string {
    var parts []string

    if problem != "" {
        parts = append(parts, problem)
    }
    if howToFix != "" {
        parts = append(parts, howToFix)
    }
    if nextSteps != "" {
        parts = append(parts, nextSteps)
    }

    return strings.Join(parts, "\n\n")
}
```

### Example 2: Protected Branch Error

```go
// errorProtectedBranch returns the formatted error message for attempting
// to create a project on a protected branch.
func errorProtectedBranch(branchName string) string {
    return formatError(
        fmt.Sprintf("Cannot create project on protected branch '%s'", branchName),
        "Projects must be created on feature branches.",
        "Action: Choose a different project name",
    )
}
```

### Example 3: Show Error Implementation

```go
// showError displays a simple error message with acknowledgment.
func showError(message string) error {
    confirm := false
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewNote().
                Title("Error").
                Description(message),
            huh.NewConfirm().
                Title("Press Enter to continue").
                Value(&confirm).
                Affirmative("Continue"),
        ),
    )
    return form.Run()
}
```

### Example 4: Error with Options Implementation

```go
// showErrorWithOptions displays an error with multiple action choices.
func showErrorWithOptions(message string, options map[string]string) (string, error) {
    var selected string

    // Convert map to huh options
    var huhOptions []huh.Option[string]
    for key, label := range options {
        huhOptions = append(huhOptions, huh.NewOption(label, key))
    }

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewNote().
                Title("Error").
                Description(message),
            huh.NewSelect[string]().
                Title("What would you like to do?").
                Options(huhOptions...).
                Value(&selected),
        ),
    )

    if err := form.Run(); err != nil {
        return "", err
    }

    return selected, nil
}
```

### Example 5: Test Exact Message Match

```go
func TestErrorProtectedBranch(t *testing.T) {
    branchName := "main"
    got := errorProtectedBranch(branchName)

    // Expected message from design doc (lines 436-442)
    want := `Cannot create project on protected branch 'main'

Projects must be created on feature branches.

Action: Choose a different project name`

    if got != want {
        t.Errorf("errorProtectedBranch() =\n%s\n\nwant:\n%s", got, want)
    }
}
```

### Example 6: Uncommitted Changes Error

```go
// errorUncommittedChanges returns the formatted error message when the
// repository has uncommitted changes.
func errorUncommittedChanges(currentBranch string) string {
    return formatError(
        "Repository has uncommitted changes",
        fmt.Sprintf(
            "You are currently on branch '%s'.\n"+
            "Creating a worktree requires switching to a different branch first.",
            currentBranch,
        ),
        "To fix:\n"+
        "  Commit: git add . && git commit -m \"message\"\n"+
        "  Or stash: git stash",
    )
}
```

## Dependencies

### Required Before This Task

- **Task 010**: Branch validation (produces validation errors)
- **Task 020**: State validation (produces state errors)
- Existing wizard infrastructure in `cli/cmd/project/wizard_*.go`

### Provides For Other Tasks

- **Task 040**: GitHub error handling (uses formatError and display functions)
- **Work Units 002-004**: Wizard flows (use all error display functions)

### External Dependencies

- charmbracelet/huh library (already in use)
- Go standard library: `fmt`, `strings`

## Constraints

### Message Immutability

**CRITICAL**: Error messages from design doc must be implemented EXACTLY as specified. This is not negotiable. The messages were:
1. Carefully crafted for clarity
2. Reviewed by multiple stakeholders
3. User-tested for comprehension
4. Approved in the design phase

Changing them requires going back through the design review process.

### Performance Requirements

- Error formatting should be instant (< 1ms)
- Error display waits for user input (no timeout)
- No caching needed (errors are rare)

### Accessibility

- Messages are plain text (screen reader friendly)
- Use clear, simple language
- Avoid jargon and technical terms where possible
- Provide concrete commands users can copy-paste

### What NOT to Do

- **Don't paraphrase error messages** - Use exact text from design
- **Don't add emoji or decoration** - Keep it professional
- **Don't make assumptions** - If design doesn't specify, ask
- **Don't add custom error messages** - Only the 6 specified ones
- **Don't modify huh library behavior** - Use it as designed

## Notes for Implementer

### Word-for-Word Implementation

When implementing the 6 specific error messages, copy the text DIRECTLY from the design document (lines 421-535). Don't retype or paraphrase. Copy-paste to ensure exactness.

**Verification strategy**:
1. Copy message from design doc
2. Remove markdown formatting (backticks, boxes)
3. Add variable substitution (fmt.Sprintf)
4. Write test that compares exact string
5. Run test to verify match

### Testing Error Display

The `showError()` and `showErrorWithOptions()` functions use terminal UI, which is hard to test automatically. Focus your tests on:
- Function compiles and accepts correct inputs
- Returns correct types
- Doesn't panic with valid inputs

For actual UI testing, use manual verification:
```bash
# Run wizard and trigger each error scenario
cd cli
go run . project wizard
# Try each error condition
```

### Integration with Validation

These error display functions will be called from the wizard state machine. Example usage pattern:

```go
// In wizard state handler
if err := isValidBranchName(name); err != nil {
    // Wrap validation error with formatted message
    msg := errorProtectedBranch(name)
    showError(msg)
    return // Stay in current state, let user retry
}
```

### Future Extensibility

While this task implements 6 specific error messages, the infrastructure (formatError, showError, showErrorWithOptions) is designed to be reusable. Future work units can add new error messages using the same patterns.

Keep the specific error functions (errorProtectedBranch, etc.) separate from the generic infrastructure (formatError). This makes it easy to add new messages later without changing the core functionality.
