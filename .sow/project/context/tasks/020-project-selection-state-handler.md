# Task 020: Project Selection State Handler

## Context

This task implements the `handleProjectSelect` state handler for the wizard, which displays discovered projects and allows users to select one to continue. This is the first screen in the project continuation workflow.

This is part of work unit 004 (Project Continuation Workflow). The wizard state machine foundation was implemented in work unit 001, and the project discovery utilities were implemented in task 010.

The state handler uses the huh library for terminal UI and integrates with the discovery utilities to present an interactive project selection list with progress information.

## Requirements

Implement the `handleProjectSelect` method in `cli/cmd/project/wizard_state.go`.

### Method Signature

```go
func (w *Wizard) handleProjectSelect() error
```

### State Handler Behavior

The handler is called when `w.state == StateProjectSelect` and should:

1. **Discover projects** with a loading spinner
2. **Handle empty list** gracefully (show message and cancel)
3. **Build selection form** with project options and cancel
4. **Show interactive selection** using huh.Select
5. **Validate selected project** still exists
6. **Save selection** to `w.choices["project"]` as ProjectInfo
7. **Transition to next state** (StateContinuePrompt)

### Detailed Implementation Steps

**Step 1: Discover Projects**

Use the `withSpinner` helper and `listProjects` utility:

```go
var projects []ProjectInfo
err := withSpinner("Discovering projects...", func() error {
    var discoverErr error
    projects, discoverErr = listProjects(w.ctx)
    return discoverErr
})

if err != nil {
    return fmt.Errorf("failed to discover projects: %w", err)
}
```

**Step 2: Handle Empty List**

If no projects found, show message and cancel:

```go
if len(projects) == 0 {
    fmt.Fprintln(os.Stderr, "\nNo existing projects found.\n")
    w.state = StateCancelled
    return nil
}
```

**Step 3: Build Selection Options**

Create huh options with formatted project display:

```go
var selectedBranch string
options := make([]huh.Option[string], 0, len(projects)+1)

for _, proj := range projects {
    // Format: "branch - name\n    [progress]"
    progress := formatProjectProgress(proj)
    label := fmt.Sprintf("%s - %s\n    [%s]", proj.Branch, proj.Name, progress)
    options = append(options, huh.NewOption(label, proj.Branch))
}

// Add cancel option at the end
options = append(options, huh.NewOption("Cancel", "cancel"))
```

**Step 4: Show Selection Form**

Use huh.Select for interactive selection:

```go
form := huh.NewForm(
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("Select a project to continue:").
            Options(options...).
            Value(&selectedBranch),
    ),
)

if err := form.Run(); err != nil {
    if errors.Is(err, huh.ErrUserAborted) {
        w.state = StateCancelled
        return nil
    }
    return fmt.Errorf("project selection error: %w", err)
}
```

**Step 5: Handle Cancellation**

If user selects "Cancel" option:

```go
if selectedBranch == "cancel" {
    w.state = StateCancelled
    return nil
}
```

**Step 6: Validate Project Still Exists**

Double-check the selected project's state file exists (race condition protection):

```go
// Find selected project in the list
var selectedProj *ProjectInfo
for i := range projects {
    if projects[i].Branch == selectedBranch {
        selectedProj = &projects[i]
        break
    }
}

if selectedProj == nil {
    // Shouldn't happen unless there's a bug
    return fmt.Errorf("internal error: selected project not found in list")
}

// Double-check state file still exists (race condition check)
worktreePath := sow.WorktreePath(w.ctx.MainRepoRoot(), selectedBranch)
statePath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")

if _, err := os.Stat(statePath); err != nil {
    // Project was deleted between discovery and selection
    _ = showError("Project no longer exists (state file missing)\n\nPress Enter to try again")
    // Stay in current state to retry
    return nil
}
```

**Step 7: Save Selection and Transition**

Store the ProjectInfo and move to continuation prompt:

```go
w.choices["project"] = *selectedProj
w.state = StateContinuePrompt
return nil
```

### Integration with Existing Wizard

The `handleProjectSelect` handler should be called from the existing `handleState` switch in `wizard_state.go`:

```go
func (w *Wizard) handleState() error {
    switch w.state {
    case StateEntry:
        return w.handleEntry()
    // ... other cases ...
    case StateProjectSelect:
        return w.handleProjectSelect()
    case StateContinuePrompt:
        return w.handleContinuePrompt()
    default:
        return fmt.Errorf("unknown state: %s", w.state)
    }
}
```

Note: The StateProjectSelect and StateContinuePrompt constants are already defined in wizard_state.go from work unit 001.

## Acceptance Criteria

### Functional Requirements

1. **Discovers projects correctly**
   - Calls listProjects with wizard context
   - Shows loading spinner during discovery
   - Handles discovery errors with clear message

2. **Handles empty list gracefully**
   - Shows "No existing projects found" message
   - Transitions to StateCancelled (not an error)
   - Returns nil (user exit, not failure)

3. **Displays projects correctly**
   - Each project shows: `branch - name\n    [progress]`
   - Progress formatted via formatProjectProgress
   - Cancel option appears at the end
   - Projects in discovery order (most recent first)

4. **Selection works properly**
   - User can navigate with arrow keys
   - User can select with Enter
   - User can press Esc to abort (transitions to StateCancelled)
   - Selection stored in w.choices["project"] as ProjectInfo

5. **Validation catches race conditions**
   - Verifies selected project still exists
   - Shows error and stays in state if project deleted
   - User can retry selection after error

6. **State transitions correctly**
   - Valid selection → StateContinuePrompt
   - Cancel selected → StateCancelled
   - Esc pressed → StateCancelled
   - Empty list → StateCancelled
   - Project deleted → stays in StateProjectSelect (retry)

### Test Requirements (TDD Approach)

Write tests FIRST in `cli/cmd/project/wizard_state_test.go`, then implement to pass them:

**Unit tests for handleProjectSelect:**
- Empty project list → shows message, transitions to StateCancelled
- Single project → displays correctly, can be selected
- Multiple projects → all displayed, selection works
- Cancel selected → transitions to StateCancelled
- Project validation failure → shows error, stays in state
- Discovery error → returns error with context

**Integration tests:**
- Full flow from StateEntry → StateProjectSelect → StateContinuePrompt
- User can navigate and select projects
- Error recovery works (project deleted between discovery and selection)

Note: Integration tests may need to mock huh.Form.Run() to simulate user input.

## Technical Details

### Go Packages and Imports

```go
import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/charmbracelet/huh"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

### File Location

Add method to: `cli/cmd/project/wizard_state.go`

This file contains the Wizard type and all state handler methods (handleEntry, handleCreateSource, etc.).

### Error Handling Pattern

Follow the existing wizard error handling pattern:

- User abort (Esc) → transition to StateCancelled, return nil
- Cancel option → transition to StateCancelled, return nil
- Recoverable error (project deleted) → show error, stay in state, return nil
- Fatal error (discovery failure, system error) → return error

### UI Formatting

The project list should display as:

```
Select a project to continue:

  ○ feat/auth - Add JWT authentication
    [Standard: implementation, 3/5 tasks completed]

  ○ design/cli-ux - CLI UX improvements
    [Design: active]

  ○ Cancel

[Use arrow keys to navigate, Enter to select]
```

The two-line format (branch-name on first line, progress on second line indented) improves readability.

### MainRepoRoot vs RepoRoot

**CRITICAL:** Use `w.ctx.MainRepoRoot()` when constructing worktree paths, not `w.ctx.RepoRoot()`:

```go
// CORRECT: Works whether wizard is run from main repo or a worktree
worktreePath := sow.WorktreePath(w.ctx.MainRepoRoot(), selectedBranch)

// WRONG: Fails if wizard is run from within a worktree
// worktreePath := sow.WorktreePath(w.ctx.RepoRoot(), selectedBranch)
```

This ensures the wizard works correctly even if invoked from within a worktree.

## Relevant Inputs

**Design documents:**
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` - Lines 299-337 contain the exact screen layout and behavior specification for project selection
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - State handler patterns and error handling

**Existing state handlers to reference:**
- `cli/cmd/project/wizard_state.go` - handleEntry, handleCreateSource, handleTypeSelect show the pattern
- `cli/cmd/project/wizard_state.go` - Lines 68-88 show the handleState dispatcher pattern

**Task 010 deliverables (dependencies):**
- `cli/cmd/project/wizard_helpers.go` - listProjects() function for discovery
- `cli/cmd/project/wizard_helpers.go` - formatProjectProgress() function for display
- `cli/cmd/project/wizard_helpers.go` - ProjectInfo struct definition

**Helper utilities:**
- `cli/cmd/project/wizard_helpers.go` - withSpinner() for loading indicators
- `cli/cmd/project/wizard_helpers.go` - showError() for error display

**Context and worktree utilities:**
- `cli/internal/sow/context.go` - MainRepoRoot() method
- `cli/internal/sow/worktree.go` - WorktreePath() function

**Huh library patterns:**
- `.sow/knowledge/designs/huh-library-verification.md` - Select widget usage, error handling with ErrUserAborted

## Examples

### Example 1: Complete Handler Implementation

```go
func (w *Wizard) handleProjectSelect() error {
    // 1. Discover projects with spinner
    var projects []ProjectInfo
    err := withSpinner("Discovering projects...", func() error {
        var discoverErr error
        projects, discoverErr = listProjects(w.ctx)
        return discoverErr
    })

    if err != nil {
        return fmt.Errorf("failed to discover projects: %w", err)
    }

    // 2. Handle empty list
    if len(projects) == 0 {
        fmt.Fprintln(os.Stderr, "\nNo existing projects found.\n")
        w.state = StateCancelled
        return nil
    }

    // 3. Build selection options
    var selectedBranch string
    options := make([]huh.Option[string], 0, len(projects)+1)

    for _, proj := range projects {
        progress := formatProjectProgress(proj)
        label := fmt.Sprintf("%s - %s\n    [%s]", proj.Branch, proj.Name, progress)
        options = append(options, huh.NewOption(label, proj.Branch))
    }
    options = append(options, huh.NewOption("Cancel", "cancel"))

    // 4. Show selection
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Select a project to continue:").
                Options(options...).
                Value(&selectedBranch),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return fmt.Errorf("project selection error: %w", err)
    }

    // 5. Handle cancellation
    if selectedBranch == "cancel" {
        w.state = StateCancelled
        return nil
    }

    // 6. Validate project still exists
    var selectedProj *ProjectInfo
    for i := range projects {
        if projects[i].Branch == selectedBranch {
            selectedProj = &projects[i]
            break
        }
    }

    if selectedProj == nil {
        return fmt.Errorf("internal error: selected project not found in list")
    }

    // Double-check state file still exists
    worktreePath := sow.WorktreePath(w.ctx.MainRepoRoot(), selectedBranch)
    statePath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")
    if _, err := os.Stat(statePath); err != nil {
        _ = showError("Project no longer exists (state file missing)\n\nPress Enter to try again")
        return nil // Stay in current state to retry
    }

    // 7. Save selection and transition
    w.choices["project"] = *selectedProj
    w.state = StateContinuePrompt
    return nil
}
```

### Example 2: Test Structure

```go
func TestHandleProjectSelect(t *testing.T) {
    t.Run("empty project list shows message and cancels", func(t *testing.T) {
        // Setup: Create wizard with empty worktrees
        // Execute: Call handleProjectSelect
        // Assert: State is StateCancelled, message shown
    })

    t.Run("valid project selection transitions to continue prompt", func(t *testing.T) {
        // Setup: Create wizard with test projects
        // Mock: User selects first project
        // Execute: Call handleProjectSelect
        // Assert: State is StateContinuePrompt, choices["project"] is set
    })

    t.Run("cancel option transitions to cancelled state", func(t *testing.T) {
        // Setup: Create wizard with test projects
        // Mock: User selects "Cancel"
        // Execute: Call handleProjectSelect
        // Assert: State is StateCancelled
    })

    t.Run("project deleted between discovery and selection", func(t *testing.T) {
        // Setup: Create wizard with test project
        // Execute: Discover projects, delete state file, select project
        // Assert: Error shown, state remains StateProjectSelect (retry)
    })
}
```

## Dependencies

**Required from task 010:**
- `listProjects()` function implemented
- `formatProjectProgress()` function implemented
- `ProjectInfo` struct defined

**Required from work unit 001:**
- Wizard struct with state, choices, ctx fields
- StateProjectSelect and StateContinuePrompt constants defined
- handleState dispatcher exists
- showError() and withSpinner() helpers exist

**Existing infrastructure:**
- `cli/internal/sow/context.go` - Context type with MainRepoRoot()
- `cli/internal/sow/worktree.go` - WorktreePath() function

## Constraints

### UX Requirements

- Loading spinner must show during project discovery (operations >500ms)
- Error messages must be actionable (tell user what happened)
- Project list must be scrollable if >10 projects (huh handles this automatically)
- Cancel option must always be available

### Error Recovery

- Recoverable errors (project deleted) → show error, stay in state, let user retry
- Fatal errors (discovery failure) → return error, bubble up to wizard runner
- User abort (Esc) → clean exit via StateCancelled

### State Machine Integration

- Handler should ONLY modify w.state and w.choices
- Should NOT modify any project state or files
- Should NOT call other handlers directly
- All state transitions explicit and documented

### Race Condition Handling

- Project might be deleted between discovery and selection
- Validation step catches this and lets user retry
- No retry limit needed (user can always cancel)
