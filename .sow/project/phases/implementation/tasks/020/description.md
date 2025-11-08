# Implement File Selection Handler Screen

## Context

This task implements the interactive file selection screen (`handleFileSelect`) for the project wizard. This screen allows users to select zero or more knowledge files from `.sow/knowledge/` to attach as context to their new project.

The wizard uses the `huh` library (charmbracelet/huh) for interactive terminal UIs. This task follows established patterns from other wizard handlers (e.g., `handleNameEntry`, `handlePromptEntry`, `handleIssueSelect`) and integrates the file selection step into the wizard flow.

The overall project goal is to enable knowledge file selection during project creation. This specific task focuses on:
1. Creating the UI screen with multi-select and filtering capabilities
2. Handling user interaction (selection, cancellation, empty results)
3. Storing selected files in wizard state
4. Proper state transitions

This builds on Task 010 which added the state constant and file discovery helper.

## Requirements

### 1. Implement handleFileSelect Method

In `cli/cmd/project/wizard_state.go`:
- Add a new method `handleFileSelect()` to the `Wizard` struct
- The method signature: `func (w *Wizard) handleFileSelect() error`
- Insert the method after `handleNameEntry` to maintain logical ordering

The handler must:
1. Discover knowledge files using `discoverKnowledgeFiles` helper
2. Handle edge cases (errors, empty directory) gracefully by skipping to next state
3. Build multi-select options from discovered files
4. Display interactive selection screen with filtering enabled
5. Handle user cancellation
6. Store selected files in `w.choices["knowledge_files"]`
7. Transition to `StatePromptEntry`

### 2. Wire Handler into State Machine

In `cli/cmd/project/wizard_state.go`, update the `handleState` method:
- Add a case for `StateFileSelect` that calls `w.handleFileSelect()`
- Insert it between the `StateNameEntry` and `StatePromptEntry` cases

### 3. Update State Transitions in Existing Handlers

Update the following handlers to transition to `StateFileSelect` instead of directly to `StatePromptEntry`:

**In `handleNameEntry`**:
- Change: `w.state = StatePromptEntry` → `w.state = StateFileSelect`

**In `createLinkedBranch`** (for GitHub issue flow):
- Change: `w.state = StatePromptEntry` → `w.state = StateFileSelect`

This ensures file selection happens in both flows (branch name and GitHub issue).

### 4. UI Requirements

The screen must:
- Use `huh.MultiSelect` for selection
- Enable filtering with `.Filterable(true)` for easy navigation
- Set a reasonable limit (10) for visible items with scrolling
- Display relative file paths from knowledge directory
- Show helpful title and description
- Allow selecting zero or more files (zero is valid - user can skip)
- Handle cancellation gracefully

## Acceptance Criteria

### Functional Requirements

1. **Handler exists and is wired correctly**:
   - `handleFileSelect` method is implemented on `Wizard` struct
   - Method is called from `handleState` switch statement
   - Handler follows existing patterns from other handlers

2. **File discovery integration**:
   - Calls `discoverKnowledgeFiles` with correct path
   - Path uses `w.ctx.MainRepoRoot()` to construct knowledge directory path
   - Handles discovery errors gracefully (logs and skips)
   - Handles empty directory gracefully (skips selection)

3. **Multi-select UI works**:
   - Screen displays with proper title and description
   - Files are shown as selectable options
   - Filtering is enabled (user can type to filter)
   - User can select zero or more files
   - Space bar toggles selection
   - Enter confirms selection
   - Ctrl+C cancels (transitions to StateCancelled)

4. **State management**:
   - Selected files stored in `w.choices["knowledge_files"]` as `[]string`
   - Empty selection (zero files) is valid and stored as empty slice
   - State transitions correctly to `StatePromptEntry`
   - Cancellation transitions to `StateCancelled`

5. **Edge cases handled**:
   - Non-existent knowledge directory: skip selection, proceed to prompt entry
   - Empty knowledge directory: skip selection, proceed to prompt entry
   - Permission errors: log error, skip selection, proceed to prompt entry
   - User selects 0 files: valid, store empty slice, proceed to prompt entry
   - Large file lists (100+ files): filtering makes navigation manageable

6. **Integration with existing flows**:
   - Branch name path: `StateNameEntry` → `StateFileSelect` → `StatePromptEntry`
   - GitHub issue path: `StateTypeSelect` → (createLinkedBranch) → `StateFileSelect` → `StatePromptEntry`

### Test Requirements (TDD Approach)

Following the project's TDD methodology, write tests FIRST, then implement:

1. **Integration test** in `wizard_integration_test.go`:
   - Test file selection with knowledge files present
   - Test file selection with empty knowledge directory (should skip)
   - Test file selection with non-existent directory (should skip)
   - Test that selected files are stored in choices
   - Test state transitions
   - Test cancellation handling

2. **Unit tests** (if needed for helper logic):
   - Test option building from file list
   - Test path handling

3. **Test patterns to follow**:
   - Use `SOW_TEST=1` environment variable (already set in TestMain)
   - Create temporary directories with test files
   - Use `t.TempDir()` for isolated test environments
   - Verify state transitions
   - Check `wizard.choices` for stored values

### Code Quality

- Follow existing handler patterns from `handleNameEntry`, `handlePromptEntry`, etc.
- Add comprehensive godoc comment for the handler
- Use `debugLog` for debugging information
- Match error handling style of other handlers
- Ensure no linting errors

## Technical Details

### Handler Implementation Pattern

Follow the pattern from existing handlers:

```go
// handleFileSelect allows selecting knowledge files to attach as context.
func (w *Wizard) handleFileSelect() error {
    debugLog("Wizard", "State=%s", w.state)

    // Discover knowledge files
    knowledgeDir := filepath.Join(w.ctx.MainRepoRoot(), ".sow", "knowledge")
    files, err := discoverKnowledgeFiles(knowledgeDir)
    if err != nil {
        // Log error but don't fail - just skip file selection
        debugLog("FileSelect", "Failed to discover files: %v", err)
        w.state = StatePromptEntry
        return nil
    }

    // If no files exist, skip selection
    if len(files) == 0 {
        debugLog("FileSelect", "No knowledge files found, skipping")
        w.state = StatePromptEntry
        return nil
    }

    // Build multi-select options
    options := make([]huh.Option[string], 0, len(files))
    for _, file := range files {
        // Use relative path for display
        options = append(options, huh.NewOption(file, file))
    }

    var selectedFiles []string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewMultiSelect[string]().
                Title("Select knowledge files to provide context (optional):").
                Description("Type to filter • Space to select • Enter to confirm").
                Options(options...).
                Value(&selectedFiles).
                Filterable(true). // Enable filtering for easy navigation
                Limit(10), // Limit visible items (user can scroll/filter)
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return fmt.Errorf("file selection error: %w", err)
    }

    // Store selected files (empty slice is valid - user can skip)
    w.choices["knowledge_files"] = selectedFiles

    w.state = StatePromptEntry
    return nil
}
```

### handleState Integration

Add the case in the switch statement:

```go
func (w *Wizard) handleState() error {
    switch w.state {
    case StateEntry:
        return w.handleEntry()
    // ... other cases
    case StateNameEntry:
        return w.handleNameEntry()
    case StateFileSelect:
        return w.handleFileSelect()
    case StatePromptEntry:
        return w.handlePromptEntry()
    // ... rest
    }
}
```

### Test Structure Example

```go
func TestWizardFileSelection(t *testing.T) {
    // Set up test repo with knowledge files
    ctx := setupTestContext(t)
    knowledgeDir := filepath.Join(ctx.MainRepoRoot(), ".sow", "knowledge")
    os.MkdirAll(filepath.Join(knowledgeDir, "designs"), 0755)
    os.WriteFile(filepath.Join(knowledgeDir, "designs", "test.md"), []byte("# Test"), 0644)

    wizard := NewWizard(nil, ctx, []string{})
    wizard.state = StateFileSelect

    err := wizard.handleFileSelect()
    assert.NoError(t, err)

    // Verify state transition
    assert.Equal(t, StatePromptEntry, wizard.state)

    // Verify knowledge files were discovered (in test mode, selection happens automatically)
    // The exact behavior depends on how SOW_TEST mode handles MultiSelect
}

func TestWizardFileSelection_EmptyDirectory(t *testing.T) {
    ctx := setupTestContext(t)
    knowledgeDir := filepath.Join(ctx.MainRepoRoot(), ".sow", "knowledge")
    os.MkdirAll(knowledgeDir, 0755) // Create empty directory

    wizard := NewWizard(nil, ctx, []string{})
    wizard.state = StateFileSelect

    err := wizard.handleFileSelect()
    assert.NoError(t, err)

    // Should skip to prompt entry
    assert.Equal(t, StatePromptEntry, wizard.state)
}

func TestWizardFileSelection_NonExistent(t *testing.T) {
    ctx := setupTestContext(t)
    // Don't create knowledge directory

    wizard := NewWizard(nil, ctx, []string{})
    wizard.state = StateFileSelect

    err := wizard.handleFileSelect()
    assert.NoError(t, err)

    // Should skip to prompt entry
    assert.Equal(t, StatePromptEntry, wizard.state)
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/cmd/project/wizard_state.go` - Contains existing handlers to follow as patterns
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/cmd/project/wizard_helpers.go` - Contains discoverKnowledgeFiles helper to use
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/cmd/project/wizard_integration_test.go` - Contains test patterns for integration tests
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/.sow/knowledge/designs/file-selector-wizard.md` - Design document with UI specifications

## Examples

### Multi-Select with Filtering

The `huh.MultiSelect` with `.Filterable(true)` provides built-in filtering:

```
┌─────────────────────────────────────────────────────┐
│ Select knowledge files to provide context (optional)│
│ Type to filter...                                   │
│                                                     │
│ [ ] designs/interactive-wizard-ux-flow.md          │
│ [x] designs/sdk-addbranch-api.md                   │
│ [ ] designs/cli-enhanced-advance.md                │
│ [x] adrs/001-project-state-machine.md              │
│ [ ] guides/testing-conventions.md                  │
│                                                     │
│ Type to filter • Space to select • Enter to confirm│
└─────────────────────────────────────────────────────┘
```

When user types "wizard":

```
┌─────────────────────────────────────────────────────┐
│ Select knowledge files to provide context (optional)│
│ > wizard_                                           │
│                                                     │
│ [ ] designs/interactive-wizard-ux-flow.md          │
│ [ ] designs/interactive-wizard-technical-impl.md   │
│                                                     │
│ Type to filter • Space to select • Enter to confirm│
└─────────────────────────────────────────────────────┘
```

### State Transition Updates

In `handleNameEntry`, change:

```go
// OLD:
w.choices["name"] = name
w.choices["branch"] = branchName
w.state = StatePromptEntry  // ← Change this
return nil

// NEW:
w.choices["name"] = name
w.choices["branch"] = branchName
w.state = StateFileSelect  // ← To this
return nil
```

In `createLinkedBranch`, change:

```go
// OLD:
w.choices["branch"] = createdBranch
w.choices["name"] = issue.Title
w.state = StatePromptEntry  // ← Change this
return nil

// NEW:
w.choices["branch"] = createdBranch
w.choices["name"] = issue.Title
w.state = StateFileSelect  // ← To this
return nil
```

## Dependencies

**Depends on Task 010**:
- Requires `StateFileSelect` constant to be defined
- Requires `discoverKnowledgeFiles` helper function to be implemented
- Requires state transition validation to be updated

## Constraints

### Performance

- File discovery must complete quickly (< 100ms for typical repos)
- UI must render instantly with filtered results
- Filtering should be responsive even with 500+ files

### User Experience

- Zero selection must be valid (user can skip by pressing Enter without selecting)
- Filtering must be intuitive (substring matching)
- Instructions must be clear (shown in description)
- Cancellation must work (Ctrl+C)

### Error Handling

- All errors during file discovery should be logged but not fail the wizard
- Users should never see cryptic errors - graceful degradation
- If file selection fails for any reason, proceed to prompt entry

### Compatibility

- Must work in all terminal environments supported by huh library
- Must work on macOS, Linux, Windows
- Must respect SOW_TEST=1 for automated testing

## Notes

- The `huh` library handles filtering internally when `.Filterable(true)` is set
- The `Limit(10)` parameter shows 10 items at a time with scrolling
- In test mode (SOW_TEST=1), interactive forms may behave differently
- The design specifies filtering as REQUIRED for v1, not an optional enhancement
- Selected files are stored as `[]string` in choices for use in Task 030
