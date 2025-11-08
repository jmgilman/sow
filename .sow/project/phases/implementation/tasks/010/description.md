# Add File Selection State and Discovery Helper

## Context

This task is part of implementing a file selector feature for the project wizard that allows users to select knowledge files from `.sow/knowledge/` during project creation. The feature enhances the wizard by enabling users to provide additional context from their knowledge base when creating new projects.

The project wizard uses a state machine pattern to guide users through different screens. Each state represents a different screen in the wizard flow. This task adds a new state (`StateFileSelect`) and the helper function to discover knowledge files.

The overall project goal is to:
1. Add a new wizard state for file selection between name entry and prompt entry
2. Implement file discovery from `.sow/knowledge/`
3. Create the UI screen with filtering support
4. Integrate selected files into project initialization
5. Ensure comprehensive testing

This task focuses on the foundational pieces: adding the state constant and implementing the file discovery logic.

## Requirements

You must implement the following changes in the wizard codebase:

### 1. Add StateFileSelect Constant

In `cli/cmd/project/wizard_state.go`:
- Add a new state constant `StateFileSelect` with value `"file_select"`
- Insert it between `StateNameEntry` and `StatePromptEntry` in the const block
- Maintain alphabetical ordering within logical groups

The state represents the file selection screen that appears after users enter a project name and before they enter an initial prompt.

### 2. Implement discoverKnowledgeFiles Helper

In `cli/cmd/project/wizard_helpers.go`:
- Create a new function `discoverKnowledgeFiles(knowledgeDir string) ([]string, error)`
- The function must walk the knowledge directory tree and return all file paths
- Return relative paths from the knowledge directory (e.g., `"designs/api.md"` not absolute paths)
- Sort results alphabetically for consistent presentation
- Handle edge cases gracefully:
  - If directory doesn't exist: return empty slice, NOT an error
  - If directory is empty: return empty slice
  - If permission errors occur: return error
  - Skip directories, only return files

### 3. Update State Transition Validation

In `cli/cmd/project/wizard_helpers.go`, update the `validateStateTransition` function:
- Add `StateFileSelect` to the `validTransitions` map
- `StateNameEntry` should allow transitions to: `StateFileSelect`, `StateCancelled`
- `StateFileSelect` should allow transitions to: `StatePromptEntry`, `StateCancelled`
- `StateTypeSelect` should allow transitions to: `StateFileSelect`, `StateCancelled` (for GitHub issue flow)

The validation ensures the state machine can only transition through valid paths, preventing logic errors.

## Acceptance Criteria

### Functional Requirements

1. **State constant exists**:
   - `StateFileSelect` constant is defined in `wizard_state.go`
   - Value is `"file_select"`
   - Placed logically in the constant list

2. **File discovery works correctly**:
   - `discoverKnowledgeFiles` returns all files in the directory tree
   - Paths are relative to knowledge directory
   - Results are sorted alphabetically
   - Subdirectories are traversed
   - Only files are returned (directories excluded)

3. **Edge cases handled**:
   - Non-existent directory returns empty slice (not error)
   - Empty directory returns empty slice
   - Permission errors return descriptive error
   - Deeply nested files are discovered

4. **State transitions validated**:
   - Transition from `StateNameEntry` to `StateFileSelect` is allowed
   - Transition from `StateFileSelect` to `StatePromptEntry` is allowed
   - Transition from `StateTypeSelect` to `StateFileSelect` is allowed (for GitHub issue flow)
   - Invalid transitions are rejected

### Test Requirements (TDD Approach)

Following the project's TDD methodology, write tests FIRST, then implement:

1. **Unit tests for discoverKnowledgeFiles** in `wizard_helpers_test.go`:
   - Test discovering markdown files in flat directory
   - Test discovering files in nested directories (subdirectories)
   - Test handling non-existent directory (should return empty slice)
   - Test handling empty directory (should return empty slice)
   - Test alphabetical sorting
   - Test relative path construction
   - Test various file types (.md, .txt, .json, etc.)
   - Test permission errors (if feasible)

2. **Unit tests for validateStateTransition** in `wizard_helpers_test.go`:
   - Test valid transitions involving StateFileSelect
   - Test invalid transitions are rejected
   - Test transition from StateNameEntry to StateFileSelect
   - Test transition from StateFileSelect to StatePromptEntry
   - Test transition from StateTypeSelect to StateFileSelect

3. **Test coverage**:
   - All code paths in `discoverKnowledgeFiles` must be covered
   - All new transitions in `validateStateTransition` must be tested
   - Edge cases must have explicit test cases

### Code Quality

- Follow existing code patterns in wizard files
- Add comprehensive godoc comments
- Use existing error handling patterns
- Match code style of surrounding functions
- Ensure no linting errors

## Technical Details

### File Discovery Implementation Pattern

Use `filepath.Walk` to traverse the directory tree, following the pattern used elsewhere in the codebase:

```go
func discoverKnowledgeFiles(knowledgeDir string) ([]string, error) {
    // Check if directory exists
    if _, err := os.Stat(knowledgeDir); err != nil {
        if os.IsNotExist(err) {
            return []string{}, nil // Not an error - just empty
        }
        return nil, fmt.Errorf("failed to stat knowledge directory: %w", err)
    }

    var files []string

    err := filepath.Walk(knowledgeDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip directories
        if info.IsDir() {
            return nil
        }

        // Get relative path from knowledge directory
        relPath, err := filepath.Rel(knowledgeDir, path)
        if err != nil {
            return err
        }

        files = append(files, relPath)
        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to walk knowledge directory: %w", err)
    }

    // Sort alphabetically for consistent ordering
    sort.Strings(files)

    return files, nil
}
```

### State Transition Validation Update

In the `validateStateTransition` function, update the map as follows:

```go
validTransitions := map[WizardState][]WizardState{
    // ... existing entries
    StateNameEntry:   {StateFileSelect, StateCancelled},  // CHANGED: was StatePromptEntry
    StateFileSelect:  {StatePromptEntry, StateCancelled}, // NEW
    StateTypeSelect:  {StateFileSelect, StateCancelled},  // CHANGED: was StateNameEntry (for issue flow)
    // ... rest
}
```

### Test Structure Example

Use the existing test patterns from `wizard_helpers_test.go`:

```go
func TestDiscoverKnowledgeFiles(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(t *testing.T) string // Returns knowledge dir path
        expected []string
        wantErr  bool
    }{
        {
            name: "discovers markdown files",
            setup: func(t *testing.T) string {
                dir := t.TempDir()
                knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
                os.MkdirAll(filepath.Join(knowledgeDir, "designs"), 0755)
                os.WriteFile(filepath.Join(knowledgeDir, "designs", "api.md"), []byte("# API"), 0644)
                os.WriteFile(filepath.Join(knowledgeDir, "README.md"), []byte("# Readme"), 0644)
                return knowledgeDir
            },
            expected: []string{"README.md", "designs/api.md"},
            wantErr:  false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            dir := tt.setup(t)
            got, err := discoverKnowledgeFiles(dir)

            if tt.wantErr && err == nil {
                t.Error("expected error, got nil")
            }
            if !tt.wantErr && err != nil {
                t.Errorf("unexpected error: %v", err)
            }

            if !reflect.DeepEqual(got, tt.expected) {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/cmd/project/wizard_state.go` - Contains state constants and wizard structure, shows pattern for adding new states
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/cmd/project/wizard_helpers.go` - Contains helper functions and validateStateTransition, shows patterns for validation and error handling
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/cmd/project/wizard_helpers_test.go` - Contains test patterns and examples to follow
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/.sow/knowledge/designs/file-selector-wizard.md` - Complete design document with technical specifications

## Examples

### State Constant Addition

```go
const (
    StateEntry          WizardState = "entry"
    StateCreateSource   WizardState = "create_source"
    StateIssueSelect    WizardState = "issue_select"
    StateTypeSelect     WizardState = "type_select"
    StateNameEntry      WizardState = "name_entry"
    StateFileSelect     WizardState = "file_select"  // NEW - insert here
    StatePromptEntry    WizardState = "prompt_entry"
    // ... rest
)
```

### File Discovery Usage

The function will be called like this in the handler (next task):

```go
knowledgeDir := filepath.Join(w.ctx.MainRepoRoot(), ".sow", "knowledge")
files, err := discoverKnowledgeFiles(knowledgeDir)
if err != nil {
    // Log error but don't fail
    debugLog("FileSelect", "Failed to discover files: %v", err)
    w.state = StatePromptEntry
    return nil
}

if len(files) == 0 {
    // Skip selection if no files
    w.state = StatePromptEntry
    return nil
}
```

## Dependencies

None - this task is foundational and has no dependencies on other tasks.

## Constraints

### Performance

- File discovery should complete in < 100ms for typical repositories (< 100 files)
- For repositories with 500+ knowledge files, performance may degrade but must remain functional
- Walking the directory tree is acceptable; no need for optimization in v1

### Security

- Use `filepath.Clean` to sanitize paths
- Never expose absolute filesystem paths in errors
- Validate all paths stay within `.sow/knowledge/`

### Compatibility

- Must work on macOS, Linux, and Windows
- Use `filepath.Join` for cross-platform path handling
- Test with various directory structures

### Error Handling

- Non-existent directory is NOT an error (graceful degradation)
- Permission errors should return descriptive errors
- All errors should be wrapped with context using `fmt.Errorf("...: %w", err)`

## Notes

- The SOW_TEST=1 environment variable is used to skip interactive prompts in tests
- The SOW_DEBUG=1 environment variable enables debug logging via `debugLog` helper
- Gap numbering (010, 020, 030) is used for task IDs to allow insertion
- This task is Task 010 because it's the foundation for subsequent work
