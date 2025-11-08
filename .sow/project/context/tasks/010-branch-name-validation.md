# Task 010: Branch Name Validation and Uncommitted Changes

## Context

This task is part of the interactive wizard project (Work Unit 005: Validation and Error Handling). The wizard allows users to create new sow projects from branch names or GitHub issues, and it needs comprehensive validation to catch errors early and provide clear, actionable feedback.

This task implements the core validation functions for branch names and uncommitted changes detection. These functions will be used throughout the wizard flows (Work Units 002-004) to validate user input before attempting to create worktrees or projects.

The validation follows a "defensive validation" philosophy: don't assume clean input, validate everything before operations that could fail or corrupt state. Better to catch errors early with clear messages than to fail mysteriously later.

**Key architectural decision**: The `normalizeName()` function is already implemented in `cli/cmd/project/wizard_helpers.go` and tested in `wizard_helpers_test.go`. This task focuses on NEW validation functions that don't exist yet.

## Requirements

### 1. Branch Name Validation Function

Create `isValidBranchName()` in `wizard_helpers.go` to validate complete branch names against git ref rules.

**Validation rules** (from design document lines 604-646):
- Not empty or whitespace-only
- Not a protected branch (main, master)
- Valid git ref name:
  - No spaces (anywhere in name)
  - No `..` (two consecutive dots)
  - No `//` (two consecutive slashes)
  - No special characters: `~`, `^`, `:`, `?`, `*`, `[`
  - No leading or trailing slashes
  - No component starting with `.` (except special refs)

**Function signature**:
```go
// isValidBranchName validates a complete branch name against git ref rules.
// Returns nil if valid, or error with clear description of what's wrong.
func isValidBranchName(name string) error
```

**Expected behavior**:
- Return `nil` for valid names: "feat/valid-name", "explore/test"
- Return error for empty/whitespace: "branch name cannot be empty"
- Return error for protected: "cannot use protected branch name"
- Return error with specific pattern: "branch name cannot contain spaces", "branch name cannot contain consecutive dots", etc.

### 2. Project Name Validation Function

Create `validateProjectName()` in `wizard_helpers.go` to validate user input during the name entry screen.

**Function signature**:
```go
// validateProjectName validates user input for project name entry.
// Called by huh input field validator during name entry screen.
// Returns nil if valid, or error with user-friendly message.
func validateProjectName(name string, prefix string) error
```

**Validation logic**:
1. Check if input is empty or whitespace → return error "project name cannot be empty"
2. Normalize the input using existing `normalizeName()` function
3. Build full branch name: `prefix + normalized`
4. Validate using `isValidBranchName()`
5. Return any validation error

**Why this function?**: The wizard uses huh library's inline validation. This function bridges user input validation with git branch rules.

### 3. Conditional Uncommitted Changes Check

Create `shouldCheckUncommittedChanges()` and `performUncommittedChangesCheckIfNeeded()` in `wizard_helpers.go`.

**Critical conditional logic** (from design lines 480-516):
- Git worktrees can't have the same branch checked out twice
- If `currentBranch == targetBranch`, sow must switch the main repo to master/main before creating worktree
- Switching branches with uncommitted changes fails
- Therefore: only check when `currentBranch == targetBranch`

**Function signatures**:
```go
// shouldCheckUncommittedChanges determines if validation is needed.
// Returns true only when current branch == target branch.
func shouldCheckUncommittedChanges(ctx *sow.Context, targetBranch string) (bool, error)

// performUncommittedChangesCheckIfNeeded runs validation conditionally.
// Uses existing sow.CheckUncommittedChanges() but adds enhanced error message.
func performUncommittedChangesCheckIfNeeded(ctx *sow.Context, targetBranch string) error
```

**Implementation notes**:
- Use `ctx.Git().CurrentBranch()` to get current branch
- Reuse existing `sow.CheckUncommittedChanges(ctx)` for actual validation
- Enhance error message to include the 3-part pattern (what/how/next)

**Enhanced error message** (from design lines 485-493):
```
Repository has uncommitted changes

You are currently on branch 'feat/add-jwt-auth-123'.
Creating a worktree requires switching to a different branch first.

To fix:
  Commit: git add . && git commit -m "message"
  Or stash: git stash
```

### 4. Protected Branch Check Helper

The existing code has `ctx.Git().IsProtectedBranch(branch)` but it's in the `sow` package. Create a helper function in `wizard_helpers.go` for convenience.

**Function signature**:
```go
// isProtectedBranch checks if a branch name is protected (main or master).
// Convenience wrapper around ctx.Git().IsProtectedBranch().
func isProtectedBranch(name string) bool
```

## Acceptance Criteria

### Functional Requirements

1. **Branch Validation Detects All Invalid Patterns**:
   - Empty string → error
   - Whitespace-only → error
   - "main" → error (protected)
   - "master" → error (protected)
   - "has spaces" → error (spaces)
   - "has..dots" → error (consecutive dots)
   - "has//slashes" → error (consecutive slashes)
   - "/leading-slash" → error (leading slash)
   - "trailing-slash/" → error (trailing slash)
   - "has~tilde" → error (special character)
   - "has^caret" → error (special character)
   - "has:colon" → error (special character)
   - "has?question" → error (special character)
   - "has*asterisk" → error (special character)
   - "has[bracket" → error (special character)

2. **Branch Validation Accepts Valid Names**:
   - "feat/valid-name" → valid ✓
   - "explore/test" → valid ✓
   - "feature-123" → valid ✓
   - "bug_fix" → valid ✓
   - "feat/epic/task" → valid ✓

3. **Project Name Validation Works**:
   - Empty input → error
   - Valid input → validates full branch name
   - Input that normalizes to protected branch → error

4. **Uncommitted Changes Check is Conditional**:
   - `shouldCheckUncommittedChanges()` returns true when current == target
   - `shouldCheckUncommittedChanges()` returns false when current != target
   - `performUncommittedChangesCheckIfNeeded()` only checks when needed
   - Error message includes current branch name and fix commands

### Test Requirements (TDD Approach)

**Write ALL tests FIRST, then implement functions to pass them.**

1. **Branch Validation Tests** (`wizard_helpers_test.go`):
   ```go
   func TestIsValidBranchName(t *testing.T)
   ```
   - Test all invalid patterns from design doc (lines 731-755)
   - Test all valid names
   - Test error messages are specific (not generic)
   - At least 20 test cases

2. **Project Name Validation Tests** (`wizard_helpers_test.go`):
   ```go
   func TestValidateProjectName(t *testing.T)
   ```
   - Test empty input
   - Test valid input with normalization
   - Test input that normalizes to protected branch
   - Test with different prefixes (feat/, explore/, etc.)
   - At least 10 test cases

3. **Uncommitted Changes Tests** (`wizard_helpers_test.go`):
   ```go
   func TestShouldCheckUncommittedChanges(t *testing.T)
   func TestPerformUncommittedChangesCheckIfNeeded(t *testing.T)
   ```
   - Mock `ctx.Git().CurrentBranch()`
   - Test current == target (should check)
   - Test current != target (should skip)
   - Test error message formatting
   - At least 8 test cases

4. **Protected Branch Tests** (`wizard_helpers_test.go`):
   ```go
   func TestIsProtectedBranch(t *testing.T)
   ```
   - Test "main" → true
   - Test "master" → true
   - Test "feat/something" → false

### Code Quality

- All functions have clear godoc comments
- Error messages follow 3-part pattern where applicable
- Tests use table-driven approach
- Tests are comprehensive and cover edge cases
- No external dependencies beyond existing packages

## Technical Details

### File Location

**Add functions to existing file**: `/cli/cmd/project/wizard_helpers.go`
**Add tests to existing file**: `/cli/cmd/project/wizard_helpers_test.go`

**Why not a new package?** The issue description suggests creating files in `cli/cmd/project/wizard/`, but the actual wizard implementation lives directly in `cli/cmd/project/`. Adding a subdirectory would create unnecessary complexity for a small set of validation functions.

### Imports Required

```go
import (
    "fmt"
    "strings"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

### Existing Functions to Reuse

**Already implemented** (do NOT reimplement):
- `normalizeName(name string) string` - in `wizard_helpers.go`
- `sow.CheckUncommittedChanges(ctx)` - in `cli/internal/sow/worktree.go`
- `ctx.Git().IsProtectedBranch(branch)` - in `cli/internal/sow/git.go`
- `ctx.Git().CurrentBranch()` - in `cli/internal/sow/git.go`

### Git Ref Validation Rules

From git documentation (`man git-check-ref-format`):

**Invalid patterns**:
- Spaces (anywhere)
- `..` (two consecutive dots) - used in refspecs
- `//` (two consecutive slashes) - creates empty components
- `~`, `^`, `:` - have special meaning in git
- `?`, `*`, `[` - shell glob characters
- Leading or trailing `/` - creates empty components
- Component starting with `.` - reserved

**Valid characters**:
- Alphanumeric: `a-z`, `A-Z`, `0-9`
- Separators: `/`, `-`, `_`
- Dots: `.` (but not `..` and not at start of component)

### Testing Strategy

**Test-Driven Development (TDD)**:

1. **Write test cases FIRST**:
   - Create test table with all cases from design doc
   - Include edge cases (empty, whitespace, special characters)
   - Include valid cases to prevent false positives

2. **Run tests (they will fail)**:
   ```bash
   cd cli/cmd/project
   go test -v -run TestIsValidBranchName
   ```

3. **Implement function to pass tests**:
   - Start simple (basic checks)
   - Add complexity incrementally
   - Re-run tests after each change

4. **Refactor if needed**:
   - Clean up code
   - Add comments
   - Ensure all tests still pass

**Example test structure**:
```go
func TestIsValidBranchName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errMsg  string // optional: check specific error message
    }{
        {
            name:    "valid branch with prefix",
            input:   "feat/valid-name",
            wantErr: false,
        },
        {
            name:    "protected branch main",
            input:   "main",
            wantErr: true,
            errMsg:  "protected branch",
        },
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := isValidBranchName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("isValidBranchName(%q) error = %v, wantErr %v",
                    tt.input, err, tt.wantErr)
            }
            if err != nil && tt.errMsg != "" {
                if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("error message %q does not contain %q",
                        err.Error(), tt.errMsg)
                }
            }
        })
    }
}
```

## Relevant Inputs

**Design documents** (contain exact validation rules and test cases):
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` (lines 380-535)
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` (lines 480-646, 706-755)

**Existing implementation to extend**:
- `cli/cmd/project/wizard_helpers.go` (existing normalizeName function and structure)
- `cli/cmd/project/wizard_helpers_test.go` (existing test patterns to follow)

**Existing validation code to reuse**:
- `cli/internal/sow/git.go` (IsProtectedBranch function, lines 73-92)
- `cli/internal/sow/worktree.go` (CheckUncommittedChanges function, lines 89-122)

**State machine integration**:
- `cli/cmd/project/wizard_state.go` (wizard state machine that will call these functions)

## Examples

### Example 1: Branch Name Validation Implementation

```go
// isValidBranchName validates a complete branch name against git ref rules.
// Returns nil if valid, or error with clear description of what's wrong.
//
// Validation rules:
//   - Not empty or whitespace-only
//   - Not a protected branch (main, master)
//   - No spaces
//   - No ".." (consecutive dots)
//   - No "//" (consecutive slashes)
//   - No special chars: ~, ^, :, ?, *, [
//   - No leading/trailing slashes
//
// Example:
//   err := isValidBranchName("feat/add-auth")  // nil (valid)
//   err := isValidBranchName("main")           // error (protected)
//   err := isValidBranchName("has spaces")     // error (spaces)
func isValidBranchName(name string) error {
    // Trim and check empty
    name = strings.TrimSpace(name)
    if name == "" {
        return fmt.Errorf("branch name cannot be empty")
    }

    // Check protected branches
    if isProtectedBranch(name) {
        return fmt.Errorf("cannot use protected branch name")
    }

    // Check for spaces
    if strings.Contains(name, " ") {
        return fmt.Errorf("branch name cannot contain spaces")
    }

    // Check for invalid patterns
    invalidPatterns := map[string]string{
        "..": "consecutive dots",
        "//": "consecutive slashes",
    }
    for pattern, desc := range invalidPatterns {
        if strings.Contains(name, pattern) {
            return fmt.Errorf("branch name cannot contain %s", desc)
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
    if strings.HasPrefix(name, "/") {
        return fmt.Errorf("branch name cannot start with /")
    }
    if strings.HasSuffix(name, "/") {
        return fmt.Errorf("branch name cannot end with /")
    }

    return nil
}
```

### Example 2: Project Name Validation

```go
// validateProjectName validates user input for project name entry.
// Called by huh input field validator during name entry screen.
//
// The function:
//   1. Checks for empty input
//   2. Normalizes the name using normalizeName()
//   3. Builds full branch name (prefix + normalized)
//   4. Validates using isValidBranchName()
//
// Example:
//   err := validateProjectName("Web Agents", "feat/")
//   // Normalizes to "web-agents", validates "feat/web-agents"
func validateProjectName(name string, prefix string) error {
    if strings.TrimSpace(name) == "" {
        return fmt.Errorf("project name cannot be empty")
    }

    normalized := normalizeName(name)
    branchName := prefix + normalized

    return isValidBranchName(branchName)
}
```

### Example 3: Conditional Uncommitted Changes Check

```go
// shouldCheckUncommittedChanges determines if validation is needed.
// Returns true only when current branch == target branch.
//
// Why conditional? Git worktrees can't have same branch checked out twice.
// If current == target, sow must switch main repo to master/main first.
// Switching with uncommitted changes fails, so we must check first.
func shouldCheckUncommittedChanges(ctx *sow.Context, targetBranch string) (bool, error) {
    currentBranch, err := ctx.Git().CurrentBranch()
    if err != nil {
        return false, fmt.Errorf("failed to get current branch: %w", err)
    }

    // Only check if we'll need to switch branches
    return currentBranch == targetBranch, nil
}

// performUncommittedChangesCheckIfNeeded runs validation conditionally.
// Uses existing sow.CheckUncommittedChanges() but adds enhanced error message.
func performUncommittedChangesCheckIfNeeded(ctx *sow.Context, targetBranch string) error {
    shouldCheck, err := shouldCheckUncommittedChanges(ctx, targetBranch)
    if err != nil {
        return err
    }

    if !shouldCheck {
        return nil // No check needed
    }

    // Use existing validation
    if err := sow.CheckUncommittedChanges(ctx); err != nil {
        // Enhance with user-friendly message
        return fmt.Errorf(
            "Repository has uncommitted changes\n\n"+
            "You are currently on branch '%s'.\n"+
            "Creating a worktree requires switching to a different branch first.\n\n"+
            "To fix:\n"+
            "  Commit: git add . && git commit -m \"message\"\n"+
            "  Or stash: git stash",
            targetBranch,
        )
    }

    return nil
}
```

### Example 4: Test Cases from Design Document

```go
func TestIsValidBranchName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        // Valid cases (from design lines 736-737)
        {"valid with prefix", "feat/valid-name", false},
        {"valid exploration", "explore/test", false},

        // Invalid cases (from design lines 738-744)
        {"protected main", "main", true},
        {"protected master", "master", true},
        {"has spaces", "has spaces", true},
        {"consecutive dots", "has..dots", true},
        {"consecutive slashes", "has//slashes", true},
        {"leading slash", "/leading-slash", true},
        {"trailing slash", "trailing-slash/", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := isValidBranchName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("isValidBranchName(%q) error = %v, wantErr %v",
                    tt.input, err, tt.wantErr)
            }
        })
    }
}
```

## Dependencies

### Required Before This Task

- Existing wizard infrastructure in `cli/cmd/project/wizard_*.go`
- Existing `normalizeName()` function in `wizard_helpers.go`
- Existing `sow.CheckUncommittedChanges()` in `cli/internal/sow/worktree.go`

### Provides For Other Tasks

This task provides core validation functions that will be used by:
- Task 020: Branch/worktree/project state checking (needs `isValidBranchName`)
- Task 030: Error display components (needs validation results)
- Work Units 002-004: Wizard flows (need all validation functions)

### External Dependencies

- Go standard library: `strings`, `fmt`
- Existing sow packages: `cli/internal/sow`
- No new external dependencies required

## Constraints

### Performance Requirements

- `isValidBranchName()` must be fast (< 1ms) - called on every user input
- `normalizeName()` already exists and is fast
- No filesystem or network operations

### Error Message Consistency

- Follow 3-part pattern where applicable (what/how/next)
- Be specific about what's wrong (don't just say "invalid")
- Provide examples in error messages when helpful

### What NOT to Do

- **Don't reimplement `normalizeName()`** - it already exists
- **Don't move wizard files to subdirectory** - they're already in the right place
- **Don't modify protected branch list** - main and master only
- **Don't add validation beyond git ref rules** - keep it simple
- **Don't create a separate validation package** - add to existing files

## Notes for Implementer

### TDD is Critical

The design document provides comprehensive test cases (lines 706-755). These are your requirements. Write tests first, implement second. This ensures:
1. Complete coverage of edge cases
2. Clear understanding of requirements
3. Confidence that your implementation is correct

### Error Messages Matter

Users will see these error messages when they make mistakes. Make them:
- **Clear**: Say exactly what's wrong
- **Actionable**: Tell them how to fix it
- **Friendly**: Don't be harsh or technical

### Existing Code is Your Friend

The wizard already has `normalizeName()` and comprehensive tests. Study these to understand:
- Code style and conventions
- Test structure and patterns
- Documentation standards

### Integration Points

These validation functions will be called from:
1. **Name entry screen**: `validateProjectName()` in huh validator
2. **State checking**: `isValidBranchName()` before worktree creation
3. **Pre-flight checks**: `performUncommittedChangesCheckIfNeeded()` before operations

Make sure your functions integrate seamlessly with the existing wizard state machine.
