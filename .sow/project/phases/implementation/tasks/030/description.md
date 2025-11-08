# Task 030: Implement Name Entry with Real-Time Preview and Validation

## Context

This task implements the core name entry screen with real-time branch preview - the most critical UX feature of the branch name creation path. As users type their project name, they see exactly what git branch will be created, helping them understand the normalization process and avoid mistakes.

This is part of Work Unit 002 (Project Creation Workflow - Branch Name Path). The wizard foundation from Work Unit 001 provides the `normalizeName()` function, `previewBranchName()` helper, and validation patterns. This task combines those utilities with huh's `DescriptionFunc` feature to create a responsive, validated input experience.

**Project Goal**: Build an interactive wizard for creating new sow projects via branch name selection, including type selection, name entry with real-time preview, prompt entry with external editor support, and project initialization in git worktrees.

**Why This Task**: The name entry is where users define their project. The real-time preview is critical for helping users understand what branch name will be created, especially since the normalization process (lowercase, hyphenation, special character removal) can significantly change their input.

## Requirements

### Handler Implementation

Create the `handleNameEntry()` function in `cli/cmd/project/wizard_state.go` to replace the current stub implementation.

**Function Location**: Replace the stub at lines 145-150 in `wizard_state.go`

**Display Requirements**:
- Show context line: "Type: <type>" (e.g., "Type: Exploration")
- Input field with:
  - Title: "Enter project name:"
  - Placeholder: "e.g., Web Based Agents"
  - Validation on submit
- Real-time preview using `huh.NewNote()` with `DescriptionFunc`
  - Title: "Branch Preview"
  - Updates as user types (bound to input variable)
  - Format: `<prefix>/<normalized-name>` or `<prefix>/<project-name>` when empty

**Real-Time Preview Implementation**:
```go
huh.NewNote().
    Title("Branch Preview").
    DescriptionFunc(func() string {
        if name == "" {
            return fmt.Sprintf("%s/<project-name>", prefix)
        }
        normalized := normalizeName(name)
        return fmt.Sprintf("%s/%s", prefix, normalized)
    }, &name)  // Critical: bind to name variable
```

**Validation Rules** (applied on submit, not during typing):
1. **Not empty**: Trimmed name must not be empty or whitespace-only
   - Error: "project name cannot be empty"

2. **Not protected branch**: After adding prefix, must not be "main" or "master"
   - Error: "cannot use protected branch name"
   - Check using: `ctx.Git().IsProtectedBranch(branchName)`

3. **Valid git branch name**: Must be a valid git ref name
   - Error: custom error from git validation
   - Validation TBD: implement `isValidBranchName()` helper

4. **No existing project**: Branch must not already have a sow project
   - Check using new `checkBranchState()` helper
   - If project exists, show detailed error with guidance (see Error Handling)

**State Transitions**:
- Valid name entered → store name and branch, transition to `StatePromptEntry`
- Validation fails → show inline error, stay in `StateNameEntry` (user can retry)
- User presses Esc → transition back to `StateTypeSelect` (go back)
- User presses Ctrl+C → catch `huh.ErrUserAborted`, transition to `StateCancelled`

**Data Storage**:
- Store original name in `w.choices["name"]` (e.g., "Web Based Agents")
- Store full branch name in `w.choices["branch"]` (e.g., "explore/web-based-agents")

### Branch State Validation

Create the `checkBranchState()` helper function in `wizard_helpers.go`:

**Purpose**: Check if a branch exists, has a worktree, and has an existing project

**Return Type**:
```go
type BranchState struct {
    BranchExists   bool
    WorktreeExists bool
    ProjectExists  bool
}

func checkBranchState(ctx *sow.Context, branchName string) (*BranchState, error)
```

**Implementation Logic**:
1. Check if branch exists using `ctx.Git().Branches()`
2. Check if worktree exists using `sow.WorktreePath()` and `os.Stat()`
3. If worktree exists, check for `.sow/project/state.yaml` file

**Error Handling**:
- Return error if git operations fail
- Return BranchState with appropriate flags set

### Error Handling

**Inline Validation Errors**:
Use huh's inline validation - errors appear below the input field as user tries to submit.

**Existing Project Error**:
If `checkBranchState()` returns `ProjectExists = true`, use `showError()` to display:

```
Error: Branch '<branch>' already has a project

To continue this project:
  Select "Continue existing project" from the main menu

To create a different project:
  Choose a different project name
```

After showing error, return `nil` to stay in `StateNameEntry` state (user can try different name).

### Integration Points

**Upstream**: Called from `handleState()` when `w.state == StateNameEntry`, triggered by `handleTypeSelect()` after type selection

**Downstream**: Transitions to `StatePromptEntry` (task 040) which uses the branch name for context display

## Acceptance Criteria

### Functional Requirements

1. **Preview Updates in Real-Time**
   - As user types, preview shows `<prefix>/<normalized-name>`
   - When input is empty, preview shows `<prefix>/<project-name>` placeholder
   - Preview updates within 50ms of keystroke (no visible lag)
   - Normalization rules applied: lowercase, spaces→hyphens, special chars removed

2. **Validation Works**
   - Empty names rejected with "project name cannot be empty"
   - Protected branch names (feat/main, feat/master) rejected
   - Invalid git characters rejected
   - Errors display inline below input field
   - User can correct and retry without restarting wizard

3. **Branch State Checked**
   - If branch already has project, show detailed error with guidance
   - Error message includes "Continue existing project" instruction
   - User stays in name entry state to try different name

4. **Navigation Works**
   - Esc key goes back to type selection
   - Ctrl+C cancels wizard entirely
   - Enter submits (with validation)

5. **Data Stored Correctly**
   - Original name stored in `w.choices["name"]`
   - Full branch name (with prefix) stored in `w.choices["branch"]`

### Test Requirements (TDD Approach)

Write tests BEFORE implementing the handler:

**Unit Tests** (add to `wizard_helpers_test.go`):

```go
func TestCheckBranchState_NoBranchNoWorktreeNoProject(t *testing.T) {
    // Test when branch doesn't exist
    // Verify: all flags are false
}

func TestCheckBranchState_BranchExistsNoWorktree(t *testing.T) {
    // Create branch but no worktree
    // Verify: BranchExists=true, others false
}

func TestCheckBranchState_WorktreeExistsNoProject(t *testing.T) {
    // Create branch and worktree but no project
    // Verify: BranchExists=true, WorktreeExists=true, ProjectExists=false
}

func TestCheckBranchState_FullStack(t *testing.T) {
    // Create branch, worktree, and project
    // Verify: all flags are true
}

func TestIsValidBranchName_ValidNames(t *testing.T) {
    // Test valid branch names
    // Examples: "feat/auth", "explore/api-v2", "123-feature"
}

func TestIsValidBranchName_InvalidNames(t *testing.T) {
    // Test invalid branch names
    // Examples: "feat..auth", "feat.", "feat/", "feat with spaces"
}
```

**Integration Tests** (add to `wizard_state_test.go`):

```go
func TestHandleNameEntry_ValidName(t *testing.T) {
    // Mock form to submit "Web Based Agents"
    // Verify: choices["name"] = "Web Based Agents"
    // Verify: choices["branch"] = "feat/web-based-agents" (assuming standard type)
    // Verify: state transitions to StatePromptEntry
}

func TestHandleNameEntry_EmptyName(t *testing.T) {
    // Mock form to submit empty string
    // Verify: validation error returned
    // Verify: state stays in StateNameEntry
}

func TestHandleNameEntry_ProtectedBranch(t *testing.T) {
    // Mock form to submit "main"
    // Verify: validation error about protected branch
    // Verify: state stays in StateNameEntry
}

func TestHandleNameEntry_ExistingProject(t *testing.T) {
    // Create project on branch first
    // Mock form to submit same branch name
    // Verify: error shown via showError()
    // Verify: state stays in StateNameEntry
}

func TestHandleNameEntry_PreviewGeneration(t *testing.T) {
    // Test preview function with various inputs
    // Verify: empty input shows "<prefix>/<project-name>"
    // Verify: "Web Based Agents" shows "feat/web-based-agents"
    // Verify: special chars removed correctly
}
```

**Manual Testing**:
1. Select "Exploration" type, enter "Web Based Agents"
   - Verify preview shows "explore/web-based-agents" as you type
2. Try to enter empty name → see inline error
3. Try to enter "main" → see protected branch error
4. Enter valid name → proceeds to prompt entry
5. Press Esc during entry → returns to type selection
6. Create project on branch, try to create another → see guidance error

## Technical Details

### Implementation Pattern

```go
func (w *Wizard) handleNameEntry() error {
    var name string
    projectType := w.choices["type"].(string)
    prefix := getTypePrefix(projectType)

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Enter project name:").
                Placeholder("e.g., Web Based Agents").
                Value(&name).
                Validate(func(s string) error {
                    // Validation 1: Not empty
                    if strings.TrimSpace(s) == "" {
                        return fmt.Errorf("project name cannot be empty")
                    }

                    // Validation 2: Not protected branch
                    normalized := normalizeName(s)
                    branchName := fmt.Sprintf("%s/%s", prefix, normalized)

                    if w.ctx.Git().IsProtectedBranch(branchName) {
                        return fmt.Errorf("cannot use protected branch name")
                    }

                    // Validation 3: Valid git branch name
                    if err := isValidBranchName(branchName); err != nil {
                        return err
                    }

                    return nil
                }),

            // Real-time preview
            huh.NewNote().
                Title("Branch Preview").
                DescriptionFunc(func() string {
                    if name == "" {
                        return fmt.Sprintf("%s/<project-name>", prefix)
                    }
                    normalized := normalizeName(name)
                    return fmt.Sprintf("%s/%s", prefix, normalized)
                }, &name),  // CRITICAL: Bind to name variable
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateTypeSelect  // Go back
            return nil
        }
        return fmt.Errorf("name entry error: %w", err)
    }

    // Post-submit validation: check branch state
    normalized := normalizeName(name)
    branchName := fmt.Sprintf("%s/%s", prefix, normalized)

    state, err := checkBranchState(w.ctx, branchName)
    if err != nil {
        return fmt.Errorf("failed to check branch state: %w", err)
    }

    if state.ProjectExists {
        showError(fmt.Sprintf(
            "Error: Branch '%s' already has a project\n\n"+
            "To continue this project:\n"+
            "  Select \"Continue existing project\" from the main menu\n\n"+
            "To create a different project:\n"+
            "  Choose a different project name",
            branchName))
        return nil  // Stay in current state to retry
    }

    // Store both original name and full branch name
    w.choices["name"] = name
    w.choices["branch"] = branchName
    w.state = StatePromptEntry

    return nil
}
```

### Branch Name Validation Helper

Implement `isValidBranchName()` in `wizard_helpers.go`:

```go
// isValidBranchName checks if a string is a valid git branch name.
// Returns nil if valid, error describing the problem if invalid.
//
// Git branch name rules:
// - Cannot start or end with /
// - Cannot contain ..
// - Cannot contain consecutive slashes //
// - Cannot end with .lock
// - Cannot contain special characters: ~, ^, :, ?, *, [, \
// - Cannot contain whitespace
func isValidBranchName(name string) error {
    if name == "" {
        return fmt.Errorf("branch name cannot be empty")
    }

    // Check for invalid patterns
    if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
        return fmt.Errorf("branch name cannot start or end with /")
    }

    if strings.Contains(name, "..") {
        return fmt.Errorf("branch name cannot contain ..")
    }

    if strings.Contains(name, "//") {
        return fmt.Errorf("branch name cannot contain consecutive slashes")
    }

    if strings.HasSuffix(name, ".lock") {
        return fmt.Errorf("branch name cannot end with .lock")
    }

    // Check for invalid characters
    invalidChars := []string{"~", "^", ":", "?", "*", "[", "\\", " "}
    for _, char := range invalidChars {
        if strings.Contains(name, char) {
            return fmt.Errorf("branch name contains invalid character: %s", char)
        }
    }

    return nil
}
```

### Package and Imports

**wizard_state.go** - add imports:
```go
import (
    "strings"  // For TrimSpace in validation
)
```

**wizard_helpers.go** - add imports:
```go
import (
    "os"         // For os.Stat in checkBranchState
    "path/filepath"  // For path joining
)
```

### File Structure

```
cli/cmd/project/
├── wizard_state.go           # MODIFY: Replace handleNameEntry stub
├── wizard_helpers.go         # MODIFY: Add checkBranchState, isValidBranchName
├── wizard_state_test.go      # CREATE: Add integration tests
├── wizard_helpers_test.go    # MODIFY: Add unit tests for helpers
```

## Relevant Inputs

### Existing Code to Understand

- `cli/cmd/project/wizard_helpers.go:38-87` - `normalizeName()` function showing exact normalization rules
- `cli/cmd/project/wizard_helpers.go:129-139` - `previewBranchName()` helper showing preview format
- `cli/cmd/project/wizard_helpers.go:141-167` - `showError()` helper for displaying errors
- `cli/internal/sow/git.go:73-92` - `IsProtectedBranch()` method for validation
- `cli/internal/sow/git.go:118-136` - `Branches()` method for listing branches
- `cli/internal/sow/worktree.go:11-16` - `WorktreePath()` function for worktree location

### Design Documents

- `.sow/knowledge/designs/interactive-wizard-ux-flow.md:196-242` - Complete name entry screen specification with validation
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md:184-234` - Implementation example with preview
- `.sow/knowledge/designs/huh-library-verification.md:113-151` - Real-time preview capability with DescriptionFunc binding
- `.sow/project/context/issue-69.md:122-178` - Detailed requirements including preview, validation, and error handling

### Testing Patterns

- `cli/cmd/project/wizard_helpers_test.go:8-104` - `TestNormalizeName` showing comprehensive normalization tests
- `cli/cmd/project/shared_test.go:23-70` - Test setup utilities for git repo and context

## Examples

### Example: Real-Time Preview

```
[User typing: "W"]
Enter project name:
┌────────────────────────────────────────┐
│ W                                      │
└────────────────────────────────────────┘

Branch Preview: feat/w

[User typing: "Web B"]
Enter project name:
┌────────────────────────────────────────┐
│ Web B                                  │
└────────────────────────────────────────┘

Branch Preview: feat/web-b

[User typing: "Web Based Agents"]
Enter project name:
┌────────────────────────────────────────┐
│ Web Based Agents                       │
└────────────────────────────────────────┘

Branch Preview: feat/web-based-agents
```

### Example: Validation Errors

```
[User submits empty name]
Enter project name:
┌────────────────────────────────────────┐
│                                        │
└────────────────────────────────────────┘
⚠ project name cannot be empty

[User submits "main"]
Enter project name:
┌────────────────────────────────────────┐
│ main                                   │
└────────────────────────────────────────┘

Branch Preview: feat/main
⚠ cannot use protected branch name

[User corrects to "authentication"]
Enter project name:
┌────────────────────────────────────────┐
│ authentication                         │
└────────────────────────────────────────┘

Branch Preview: feat/authentication
[Validation passes, proceeds to prompt entry]
```

### Example: Existing Project Handling

```
[User tries to create project on branch that already has one]

╔══════════════════════════════════════════════════════════╗
║                        Error                             ║
╚══════════════════════════════════════════════════════════╝

Error: Branch 'feat/auth' already has a project

To continue this project:
  Select "Continue existing project" from the main menu

To create a different project:
  Choose a different project name

[Press Enter to acknowledge]

[Returns to name entry screen for retry]
```

## Dependencies

### Upstream Dependencies (Must Complete First)

- **Work Unit 001**: Wizard Foundation and State Machine ✅ COMPLETE
  - Provides: `normalizeName()` function
  - Provides: `previewBranchName()` helper
  - Provides: `showError()` helper
  - Provides: `WizardState` enum

- **Task 020**: Type selection handler (this work unit)
  - Provides: `w.choices["type"]` for determining prefix

### Downstream Dependencies (Will Use This Task)

- **Task 040**: Prompt entry handler
  - Reads: `w.choices["branch"]` for context display

- **Task 050**: Finalization handler
  - Reads: `w.choices["name"]` for project description
  - Reads: `w.choices["branch"]` for worktree creation

## Constraints

### Performance Requirements

- Preview must update within 50ms of keystroke
- `normalizeName()` must be fast (string ops only, no I/O)
- Trust huh's `DescriptionFunc` automatic caching

### UX Requirements

- Preview must always be visible while user is typing
- Validation errors must be inline (below input field)
- Error messages must be actionable (tell user what to do)
- User must be able to go back (Esc) or cancel (Ctrl+C)

### Git Validation Requirements

- Must check for protected branches (main, master)
- Must validate git ref name rules
- Must check for existing projects on branch
- Must use Context.Git() methods (not raw git commands)

### Testing Requirements

- Tests written BEFORE implementation (TDD)
- Unit tests for pure functions (checkBranchState, isValidBranchName)
- Integration tests for handler flow
- Manual tests for UX and preview responsiveness

### What NOT to Do

- ❌ Don't validate during typing (only on submit) - causes bad UX
- ❌ Don't modify normalizeName() - it's shared and tested
- ❌ Don't show error dialogs for validation - use inline errors
- ❌ Don't skip branch state check - prevents conflicts
- ❌ Don't block protected branches in preview - only in validation
- ❌ Don't use raw git commands - use Context.Git() methods

## Notes

### Critical Implementation Details

1. **DescriptionFunc Binding**: The `&name` binding in DescriptionFunc is what makes the preview reactive. Without it, the preview won't update as user types.

2. **Two-Phase Validation**: Inline validation (in Validate callback) runs on submit attempt. Post-submit validation (checkBranchState) runs after form succeeds. This ensures fast inline feedback while allowing complex checks.

3. **Error Recovery**: When validation fails, return `nil` to stay in current state. This allows user to correct their input without restarting the wizard.

4. **Data Storage**: Store both the original name and the normalized branch name. Downstream handlers need both.

### Performance Optimization

- `normalizeName()` is called on every keystroke for preview
- Keep it simple: string operations only, no I/O
- `DescriptionFunc` has automatic caching - trust it
- Don't add debouncing - huh handles that

### Git Branch Name Rules

Full git ref validation is complex. We implement a simplified subset:
- No invalid characters: ~^:?*[\ and space
- No .. or // sequences
- No leading/trailing /
- No .lock suffix

This catches the most common errors while keeping validation simple. Git will catch any edge cases during worktree creation.
