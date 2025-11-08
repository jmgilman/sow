# File Selector for Project Wizard

**Status**: Draft
**Created**: 2025-11-08
**Author**: AI Orchestrator

## Executive Summary

This design adds a file selection screen to the project creation wizard that allows users to select knowledge files from `.sow/knowledge/` to attach as context to their new project. The feature includes **built-in filtering** (substring search) to enable efficient navigation even with hundreds of files. The feature integrates seamlessly into the existing wizard state machine flow and follows established patterns from the codebase.

## Problem Statement

When creating a new project through `sow project wizard`, users cannot currently provide additional context from existing knowledge base documents. The orchestrator only receives:
- GitHub issue details (if creating from an issue)
- User's initial prompt (optional)

However, `.sow/knowledge/` may contain relevant:
- Architecture design documents
- ADRs (Architecture Decision Records)
- Implementation guides
- Other design specifications

These files could provide valuable context to help the orchestrator make better decisions during project initialization and execution.

## Goals

1. **Enable knowledge file selection** during project creation
2. **Maintain existing UX patterns** from the wizard (huh library, state machine)
3. **Support flexible file selection** (single or multiple files)
4. **Provide filtering/search** for easy navigation of large file lists
5. **Gracefully handle edge cases** (empty knowledge directory, no selection)
6. **Integrate cleanly** with existing prompt generation system

## Non-Goals

- Selecting files from outside `.sow/knowledge/`
- Inline file preview or editing
- File management operations (create, delete, move)
- File content validation or parsing

## Design Overview

### User Flow

The file selector inserts between existing wizard states:

```
Current flow:
  StateNameEntry → StatePromptEntry → StateComplete

New flow:
  StateNameEntry → StateFileSelect → StatePromptEntry → StateComplete

For GitHub issue flow:
  StateIssueSelect → StateTypeSelect → StateFileSelect → StatePromptEntry → StateComplete
```

The file selection screen is **optional** - users can skip it and proceed without attaching any knowledge files.

### UI Design

The screen uses `huh.MultiSelect` with filtering enabled to allow selecting zero or more files:

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

**With filtering active:**

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

**Display rules**:
- Files shown with paths relative to `.sow/knowledge/`
- Sorted alphabetically for consistent presentation
- **Filtering enabled** - users can type to filter the list (substring matching)
- Supports large file lists (500+) - filtering makes navigation fast
- If `.sow/knowledge/` is empty or doesn't exist, skip screen entirely

**Filtering behavior**:
- Substring matching (case-insensitive)
- Matches anywhere in the file path
- Real-time filtering as user types
- Future enhancement: Could add fuzzy matching (like fzf) using custom filter function

### Technical Implementation

#### State Machine Changes

Add new state constant to `wizard_state.go`:

```go
const (
    // ... existing states
    StateNameEntry      WizardState = "name_entry"
    StateFileSelect     WizardState = "file_select"  // NEW
    StatePromptEntry    WizardState = "prompt_entry"
    // ... rest
)
```

Update state transition validation in `wizard_helpers.go`:

```go
func validateStateTransition(from, to WizardState) error {
    validTransitions := map[WizardState][]WizardState{
        // ... existing
        StateNameEntry:   {StateFileSelect, StateCancelled},  // CHANGED
        StateFileSelect:  {StatePromptEntry, StateCancelled}, // NEW
        StateTypeSelect:  {StateFileSelect, StateCancelled},  // CHANGED (for issue flow)
        // ... rest
    }
    // ... validation logic
}
```

#### Handler Implementation

Add handler method in `wizard_state.go`:

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

#### File Discovery Helper

Add utility function in `wizard_helpers.go`:

```go
// discoverKnowledgeFiles finds all files in .sow/knowledge/ directory.
// Returns relative paths from knowledge directory (e.g., "designs/api.md").
// Returns empty slice if directory doesn't exist (not an error).
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

#### Integration with Project Initialization

Modify `initializeProject` in `shared.go` to accept knowledge files:

```go
func initializeProject(
    ctx *sow.Context,
    branch string,
    description string,
    issue *sow.Issue,
    knowledgeFiles []string,  // NEW parameter
) (*state.Project, error) {
    // ... existing setup ...

    // Prepare initial inputs
    var initialInputs map[string][]projschema.ArtifactState

    // Add issue input if provided (existing logic)
    if issue != nil {
        // ... existing issue handling ...
        initialInputs = map[string][]projschema.ArtifactState{
            "implementation": {issueArtifact},
        }
    }

    // NEW: Add knowledge file inputs
    if len(knowledgeFiles) > 0 {
        if initialInputs == nil {
            initialInputs = make(map[string][]projschema.ArtifactState)
        }

        knowledgeArtifacts := make([]projschema.ArtifactState, 0, len(knowledgeFiles))
        for _, file := range knowledgeFiles {
            artifact := projschema.ArtifactState{
                Type:       "reference", // Knowledge files are references
                Path:       filepath.Join("../../knowledge", file), // Relative to project dir
                Approved:   true,
                Created_at: time.Now(),
                Metadata: map[string]interface{}{
                    "source":      "user_selected",
                    "description": "Knowledge file selected during project creation",
                },
            }
            knowledgeArtifacts = append(knowledgeArtifacts, artifact)
        }

        // Determine target phase (depends on project type)
        // For standard/exploration/design projects, add to first phase
        // Could also add to "common" if project type supports it
        targetPhase := determineKnowledgeInputPhase(ctx)

        if _, exists := initialInputs[targetPhase]; !exists {
            initialInputs[targetPhase] = []projschema.ArtifactState{}
        }
        initialInputs[targetPhase] = append(initialInputs[targetPhase], knowledgeArtifacts...)
    }

    // ... rest of existing logic ...
}

// determineKnowledgeInputPhase determines which phase should receive knowledge inputs.
// For most project types, this is the first phase. Could be made configurable.
func determineKnowledgeInputPhase(ctx *sow.Context) string {
    // For now, use "implementation" as default (first phase for standard projects)
    // This could be enhanced to detect project type and use appropriate phase
    return "implementation"
}
```

Update `finalizeCreation` in `wizard_state.go` to pass knowledge files:

```go
func (w *Wizard) finalizeCreation() error {
    // ... existing extraction ...

    // Extract knowledge files (NEW)
    var knowledgeFiles []string
    if files, ok := w.choices["knowledge_files"].([]string); ok {
        knowledgeFiles = files
    }

    // ... existing worktree setup ...

    // Initialize project WITH knowledge files
    project, err := initializeProject(worktreeCtx, branch, name, issue, knowledgeFiles)
    if err != nil {
        return fmt.Errorf("failed to initialize project: %w", err)
    }

    // ... rest of existing logic ...
}
```

### Edge Cases

| Case | Behavior |
|------|----------|
| `.sow/knowledge/` doesn't exist | Skip file selection, proceed to prompt entry |
| `.sow/knowledge/` is empty | Skip file selection, proceed to prompt entry |
| User selects 0 files | Valid - proceed without knowledge inputs |
| 100+ files in knowledge directory | All files shown, filtering enabled for easy navigation |
| 500+ files (large repository) | Performance may degrade slightly, but filtering makes this manageable |
| 1000+ files (very large repository) | Consider adding warning or pagination if performance issues arise |
| Permission errors reading directory | Log error, skip file selection, continue wizard |
| File is deleted between discovery and submission | Not an error - references will fail gracefully later |
| User filters but matches no files | Empty list shown, can clear filter or skip selection |

### Testing Strategy

#### Unit Tests

Add tests to `wizard_helpers_test.go`:

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
        {
            name: "handles non-existent directory",
            setup: func(t *testing.T) string {
                return filepath.Join(t.TempDir(), "nonexistent")
            },
            expected: []string{},
            wantErr:  false,
        },
        {
            name: "sorts alphabetically",
            setup: func(t *testing.T) string {
                dir := t.TempDir()
                knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
                os.MkdirAll(knowledgeDir, 0755)
                os.WriteFile(filepath.Join(knowledgeDir, "zebra.md"), []byte("Z"), 0644)
                os.WriteFile(filepath.Join(knowledgeDir, "apple.md"), []byte("A"), 0644)
                return knowledgeDir
            },
            expected: []string{"apple.md", "zebra.md"},
            wantErr:  false,
        },
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

#### Integration Tests

Add test to `wizard_integration_test.go`:

```go
func TestWizardFileSelection(t *testing.T) {
    // Set up test repo with knowledge files
    ctx := setupTestRepo(t)
    knowledgeDir := filepath.Join(ctx.MainRepoRoot(), ".sow", "knowledge")
    os.MkdirAll(filepath.Join(knowledgeDir, "designs"), 0755)
    os.WriteFile(filepath.Join(knowledgeDir, "designs", "test.md"), []byte("# Test"), 0644)

    // Run wizard with SOW_TEST=1 (auto-selects first option)
    os.Setenv("SOW_TEST", "1")
    defer os.Unsetenv("SOW_TEST")

    wizard := NewWizard(nil, ctx, []string{})
    wizard.choices["action"] = "create"
    wizard.choices["source"] = "branch"
    wizard.choices["type"] = "standard"
    wizard.choices["name"] = "test-project"
    wizard.state = StateFileSelect

    err := wizard.handleFileSelect()
    assert.NoError(t, err)

    // Verify knowledge files were discovered
    files, ok := wizard.choices["knowledge_files"].([]string)
    assert.True(t, ok)
    assert.Contains(t, files, "designs/test.md")

    // Verify state transition
    assert.Equal(t, StatePromptEntry, wizard.state)
}
```

## Implementation Plan

### Phase 1: Core File Selection (2-3 days)

1. Add `StateFileSelect` constant and update state machine
2. Implement `discoverKnowledgeFiles` helper
3. Implement `handleFileSelect` screen with filtering enabled
4. Add unit tests for file discovery
5. Update state transition validation

### Phase 2: Integration with Project Init (2-3 days)

1. Modify `initializeProject` to accept knowledge files
2. Implement artifact creation for knowledge inputs
3. Update `finalizeCreation` to pass knowledge files
4. Add integration tests for full flow
5. Test with various project types

### Phase 3: Polish and Edge Cases (1-2 days)

1. Test filtering with large file lists (100+, 500+ files)
2. Verify performance with realistic knowledge directories
3. Add helpful descriptions and instructions
4. Test edge cases (empty dir, permission errors, filter no matches)
5. Update documentation
6. Manual testing with real repositories

## Security Considerations

- **Path traversal**: Use `filepath.Clean` and validate all paths stay within `.sow/knowledge/`
- **Symlink attacks**: Use `filepath.EvalSymlinks` to resolve before validation
- **Permission errors**: Handle gracefully, don't expose full filesystem paths in errors
- **Arbitrary file inclusion**: Only allow files from `.sow/knowledge/`, never user-provided absolute paths

## Alternatives Considered

### Alternative 1: Inline file browser with preview

**Approach**: Use `huh` to build a file tree navigator with file content preview.

**Pros**:
- Users can see file contents before selecting
- Better for unfamiliar repositories

**Cons**:
- Significantly more complex UX
- Requires pagination for file contents
- Slower workflow for users who know what they want
- Not consistent with existing wizard patterns

**Decision**: Rejected - too complex for initial implementation

### Alternative 2: Post-creation input addition

**Approach**: Let users create project first, then use `sow input add` to attach knowledge files.

**Pros**:
- Simpler wizard flow
- Uses existing CLI commands
- More flexible timing

**Cons**:
- Extra step for common use case
- Knowledge files not available during initial orchestrator prompt
- Breaks "everything in one flow" UX goal

**Decision**: Rejected - less convenient for users

### Alternative 3: Auto-detect relevant files

**Approach**: Use heuristics (file names, content) to automatically suggest files.

**Pros**:
- Less user effort
- Could be smarter over time

**Cons**:
- Complex heuristics hard to get right
- Users still need override mechanism
- Magic behavior can be confusing
- Scope creep

**Decision**: Deferred - could add later as enhancement

## Open Questions

1. **Should we support searching/filtering files?**
   - Answer: **YES** - Filtering is included in v1 using huh's built-in `.Filterable(true)`. Substring matching for now, fuzzy search could be added later as enhancement.

2. **Should file selection be per-phase or project-level?**
   - Answer: Project-level for v1 (add to first phase). Phase-specific can come later.

3. **What's the max reasonable file count?**
   - Answer: **No hard limit** - with filtering enabled, users can navigate large file lists efficiently. Will monitor performance with 500+ files and add pagination/warning if needed.

4. **Should we validate file types (only .md)?**
   - Answer: No - allow any file type. Orchestrator can handle various formats.

5. **Should we use fuzzy search (like fzf) instead of substring matching?**
   - Answer: Not in v1. Substring matching is sufficient for most use cases and is built into huh. Fuzzy matching could be added later using custom filter function if users request it.

## Success Metrics

- Users can successfully attach knowledge files during project creation
- File selection screen loads in <100ms for typical repositories (<100 knowledge files)
- Filtering works smoothly with 500+ files (no noticeable lag)
- Users can find desired files in <5 seconds even with 200+ files (via filtering)
- Zero user-reported bugs related to file path handling
- No degradation in wizard completion rate
- Positive user feedback on feature usefulness and filtering UX

## References

- Existing wizard implementation: `cli/cmd/project/wizard*.go`
- State machine design: `.sow/knowledge/designs/project-modes/`
- Huh library docs: `https://github.com/charmbracelet/huh`
- Project state schema: `cli/schemas/project/`
