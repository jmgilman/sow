# Task 030: Continuation Prompt State Handler

## Context

This task implements the `handleContinuePrompt` state handler, which allows users to optionally provide additional context about what they want to work on when continuing a project. This is the second and final screen in the project continuation workflow before finalization.

This is part of work unit 004 (Project Continuation Workflow). The wizard state machine foundation was implemented in work unit 001, and the project selection screen was implemented in task 020.

The continuation prompt is optional - users can leave it empty to continue from the project's last state, or provide specific instructions to guide Claude's work.

## Requirements

Implement the `handleContinuePrompt` method in `cli/cmd/project/wizard_state.go`.

### Method Signature

```go
func (w *Wizard) handleContinuePrompt() error
```

### State Handler Behavior

The handler is called when `w.state == StateContinuePrompt` and should:

1. **Extract selected project** from `w.choices["project"]`
2. **Build context display** showing project name, branch, and current state
3. **Show prompt entry form** using huh.Text with multi-line and external editor support
4. **Save user's prompt** to `w.choices["prompt"]` (may be empty string)
5. **Transition to complete state** (StateComplete) to trigger finalization

### Detailed Implementation Steps

**Step 1: Extract Selected Project**

Retrieve the ProjectInfo from choices:

```go
proj, ok := w.choices["project"].(ProjectInfo)
if !ok {
    return fmt.Errorf("internal error: project choice not set or invalid")
}
```

**Step 2: Build Context Display**

Format project information for the Description field:

```go
progress := formatProjectProgress(proj)
contextInfo := fmt.Sprintf(
    "Project: %s\nBranch: %s\nState: %s",
    proj.Name,
    proj.Branch,
    progress,
)
```

**Step 3: Show Prompt Entry Form**

Use huh.Text for multi-line input with external editor support:

```go
var prompt string

form := huh.NewForm(
    huh.NewGroup(
        huh.NewText().
            Title("What would you like to work on? (optional):").
            Description(contextInfo + "\n\nPress Ctrl+E to open $EDITOR for multi-line input").
            CharLimit(5000).
            Value(&prompt).
            EditorExtension(".md"),
    ),
)

if err := form.Run(); err != nil {
    if errors.Is(err, huh.ErrUserAborted) {
        w.state = StateCancelled
        return nil
    }
    return fmt.Errorf("continuation prompt error: %w", err)
}
```

**Step 4: Save Prompt and Transition**

Store the prompt (even if empty) and move to completion:

```go
w.choices["prompt"] = prompt
w.state = StateComplete
return nil
```

### Integration with Existing Wizard

The `handleContinuePrompt` handler should be called from the existing `handleState` switch in `wizard_state.go`.

The StateComplete transition triggers the `finalize()` method, which will handle the continuation-specific finalization logic (implemented in task 040).

## Acceptance Criteria

### Functional Requirements

1. **Extracts project correctly**
   - Retrieves ProjectInfo from w.choices["project"]
   - Returns error if not set or wrong type
   - No type assertion panics

2. **Displays context information**
   - Shows project name
   - Shows branch name
   - Shows current state with progress (formatted via formatProjectProgress)
   - Instructions about Ctrl+E for external editor

3. **Prompt entry works correctly**
   - Multi-line text area accepts input
   - Character limit enforced (5000 chars)
   - Ctrl+E opens $EDITOR with .md extension
   - Optional - user can leave empty
   - Empty prompt is valid (not an error)

4. **Handles user actions**
   - Submit (Enter) → saves prompt, transitions to StateComplete
   - Abort (Esc) → transitions to StateCancelled
   - External editor (Ctrl+E) → opens editor, captures result

5. **State transitions correctly**
   - Prompt submitted → StateComplete
   - User aborts → StateCancelled
   - Always saves prompt (even if empty) before transitioning

6. **Stores prompt correctly**
   - w.choices["prompt"] set to string value
   - Empty string if user left blank
   - Whitespace preserved as entered

### Test Requirements (TDD Approach)

Write tests FIRST in `cli/cmd/project/wizard_state_test.go`, then implement to pass them:

**Unit tests for handleContinuePrompt:**
- Valid project in choices → form displays correctly
- Missing project in choices → returns error
- User submits non-empty prompt → stored in choices, transitions to StateComplete
- User submits empty prompt → empty string stored, transitions to StateComplete
- User aborts (Esc) → transitions to StateCancelled
- Prompt over character limit → truncated or error (depends on huh behavior)

**Integration tests:**
- Full flow: StateProjectSelect → StateContinuePrompt → StateComplete
- Context information displays correctly
- Prompt is passed to finalization

Note: External editor testing may need to be manual or mocked.

## Technical Details

### Go Packages and Imports

```go
import (
    "errors"
    "fmt"

    "github.com/charmbracelet/huh"
)
```

### File Location

Add method to: `cli/cmd/project/wizard_state.go`

This file contains the Wizard type and all state handler methods.

### Character Limit

The continuation prompt has a 5000 character limit. This is:
- Large enough for substantial context (multiple paragraphs)
- Small enough to avoid overwhelming Claude
- Consistent with similar prompt entry screens

The limit is enforced by huh.Text's CharLimit option.

### External Editor Support

The huh.Text field supports external editor via Ctrl+E:
- EditorExtension(".md") sets the temp file extension
- Uses $EDITOR environment variable
- Falls back to $VISUAL if $EDITOR not set
- Falls back to basic editors (vi, nano) if neither set

This is handled automatically by huh - no custom implementation needed.

### Prompt Optionality

The continuation prompt is explicitly optional. Users might want to:
- Continue from exactly where they left off (empty prompt)
- Provide new direction ("focus on tests")
- Add context ("user reported bug in refresh logic")

All of these are valid use cases. An empty prompt is not an error.

### UI Layout

The form should display as:

```
╔══════════════════════════════════════════════════════════╗
║                   Continue Project                       ║
╚══════════════════════════════════════════════════════════╝

Project: Add JWT authentication
Branch: feat/auth
State: Standard: implementation, 3/5 tasks completed

What would you like to work on? (optional):
Press Ctrl+E to open $EDITOR for multi-line input

┌────────────────────────────────────────────────────────┐
│ Let's focus on the token refresh logic and write      │
│ comprehensive integration tests for the auth flow     │
└────────────────────────────────────────────────────────┘

[Enter to continue, Ctrl+E for editor, Esc to cancel]
```

The Description field shows both the context information and the editor hint.

## Relevant Inputs

**Design documents:**
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` - Lines 338-376 contain the exact screen layout and behavior for continuation prompt
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - Continuation prompt implementation details

**Existing state handlers to reference:**
- `cli/cmd/project/wizard_state.go` - handlePromptEntry (lines 293-333) shows nearly identical pattern for new project prompts
- `cli/cmd/project/wizard_state.go` - Other handlers show error handling and state transition patterns

**Task 010 deliverables (dependencies):**
- `cli/cmd/project/wizard_helpers.go` - formatProjectProgress() function
- `cli/cmd/project/wizard_helpers.go` - ProjectInfo struct definition

**Task 020 deliverables (dependencies):**
- `cli/cmd/project/wizard_state.go` - handleProjectSelect sets w.choices["project"]

**Huh library capabilities:**
- `.sow/knowledge/designs/huh-library-verification.md` - Text field with CharLimit, EditorExtension, Ctrl+E external editor

## Examples

### Example 1: Complete Handler Implementation

```go
func (w *Wizard) handleContinuePrompt() error {
    // 1. Extract selected project
    proj, ok := w.choices["project"].(ProjectInfo)
    if !ok {
        return fmt.Errorf("internal error: project choice not set or invalid")
    }

    // 2. Build context display
    progress := formatProjectProgress(proj)
    contextInfo := fmt.Sprintf(
        "Project: %s\nBranch: %s\nState: %s",
        proj.Name,
        proj.Branch,
        progress,
    )

    // 3. Show prompt entry form
    var prompt string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewText().
                Title("What would you like to work on? (optional):").
                Description(contextInfo + "\n\nPress Ctrl+E to open $EDITOR for multi-line input").
                CharLimit(5000).
                Value(&prompt).
                EditorExtension(".md"),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return fmt.Errorf("continuation prompt error: %w", err)
    }

    // 4. Save prompt and transition
    w.choices["prompt"] = prompt
    w.state = StateComplete
    return nil
}
```

### Example 2: Test Structure

```go
func TestHandleContinuePrompt(t *testing.T) {
    t.Run("valid project displays context correctly", func(t *testing.T) {
        // Setup: Create wizard with project in choices
        // Mock: Form submission (may need to mock huh.Form)
        // Execute: Call handleContinuePrompt
        // Assert: Context info includes name, branch, state
    })

    t.Run("non-empty prompt saved and transitions to complete", func(t *testing.T) {
        // Setup: Create wizard with project in choices
        // Mock: User enters "focus on tests"
        // Execute: Call handleContinuePrompt
        // Assert: choices["prompt"] == "focus on tests", state == StateComplete
    })

    t.Run("empty prompt is valid", func(t *testing.T) {
        // Setup: Create wizard with project in choices
        // Mock: User submits without entering anything
        // Execute: Call handleContinuePrompt
        // Assert: choices["prompt"] == "", state == StateComplete
    })

    t.Run("user abort transitions to cancelled", func(t *testing.T) {
        // Setup: Create wizard with project in choices
        // Mock: User presses Esc
        // Execute: Call handleContinuePrompt
        // Assert: state == StateCancelled
    })

    t.Run("missing project returns error", func(t *testing.T) {
        // Setup: Create wizard WITHOUT project in choices
        // Execute: Call handleContinuePrompt
        // Assert: Returns error with "project choice not set"
    })
}
```

### Example 3: Usage in Finalization

The prompt saved here is used in finalization (task 040):

```go
// In finalize() method
if w.choices["action"] == "continue" {
    proj := w.choices["project"].(ProjectInfo)
    userPrompt := w.choices["prompt"].(string)  // From this handler

    // Generate base 3-layer prompt
    basePrompt, err := generateContinuePrompt(projectState)
    // ...

    // Append user prompt if provided
    fullPrompt := basePrompt
    if userPrompt != "" {
        fullPrompt += "\n\nUser request:\n" + userPrompt
    }

    // Launch Claude with fullPrompt
    // ...
}
```

## Dependencies

**Required from task 010:**
- `formatProjectProgress()` function implemented
- `ProjectInfo` struct defined

**Required from task 020:**
- `handleProjectSelect` sets w.choices["project"] as ProjectInfo

**Required from work unit 001:**
- Wizard struct with state, choices fields
- StateContinuePrompt and StateComplete constants defined
- handleState dispatcher exists

**Will be used by task 040:**
- This handler sets w.choices["prompt"] which finalization uses

## Constraints

### UX Requirements

- Prompt is explicitly optional (empty is valid)
- Character limit must be enforced (5000 chars)
- External editor must be supported (Ctrl+E)
- Context information must be visible while entering prompt
- Cancel must always be available (Esc)

### Input Validation

- No validation needed beyond character limit
- Empty prompt is valid
- Whitespace-only prompt is valid (user's choice)
- No sanitization needed (will be passed to Claude as-is)

### State Machine Integration

- Handler should ONLY modify w.state and w.choices
- Should NOT load or modify project state
- Should NOT interact with worktrees or files
- All state transitions explicit

### Error Handling

- Missing project in choices → fatal error (internal bug)
- User abort → clean exit via StateCancelled
- Form errors → return with context
- No retry logic needed (user can abort and restart)

### Editor Compatibility

- Must work with various $EDITOR values (vim, emacs, nano, code, etc.)
- Falls back gracefully if no editor configured
- Temp file cleanup handled by huh library
- .md extension helps editors enable markdown highlighting
