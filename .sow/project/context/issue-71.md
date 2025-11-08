# Issue #71: Project Continuation Workflow

**URL**: https://github.com/jmgilman/sow/issues/71
**State**: OPEN

## Description

# Work Unit 004: Project Continuation Workflow

## Behavioral Goal

**As a** sow user who wants to resume work on an existing project,
**I need** an interactive workflow to discover and select active projects across all worktrees,
**so that** I can quickly return to any project with full context about its current state and continue where I left off.

### Success Criteria for Reviewers

- User can discover all active projects regardless of current branch
- Project list displays meaningful progress information (type, phase, task completion)
- Selection process validates project still exists before launching Claude
- Optional continuation prompt allows providing new context or continuing from last state
- Claude launches in the correct worktree with appropriate 3-layer prompt

---

## Existing Code Context

### Explanatory Overview

This work unit implements the "Continue existing project" path from the interactive wizard. It leverages the existing worktree architecture where each project lives in isolation at `.sow/worktrees/<branch>/.sow/project/`.

The continuation workflow differs from creation in **one critical way**: there is NO uncommitted changes check. Since the worktree already exists and we're not switching branches in the main repo, uncommitted changes are irrelevant.

The implementation will extract and refactor logic from `cli/cmd/project/continue.go` (which will be deleted in favor of the wizard), reusing:
- Project discovery algorithm (scan worktrees, load state)
- Continuation prompt generation (3-layer: orchestrator + type + state)
- Claude launch logic (shared with creation paths)

### Reference File List

**Core continuation logic to extract/migrate:**
- `cli/cmd/project/continue.go:103-106` - Project loading via SDK
- `cli/cmd/project/continue.go:110-114` - Continuation prompt generation
- `cli/cmd/project/continue.go:167-196` - `generateContinuePrompt()` function (3-layer structure)

**Worktree infrastructure (reuse as-is):**
- `cli/internal/sow/worktree.go:11-16` - `WorktreePath()` function
- `cli/internal/sow/worktree.go:18-87` - `EnsureWorktree()` idempotent worktree creation

**Context management (reuse as-is):**
- `cli/internal/sow/context.go` - Context creation and repository detection

**Shared utilities from work unit 001:**
- State machine types and handlers (StateProjectSelect, StateContinuePrompt)
- Loading spinner component
- Claude launch utility function

---

## Existing Documentation Context

### Design Documents

From `.sow/knowledge/designs/interactive-wizard-ux-flow.md` (lines 299-376):
- **Path 2: Continue Existing Project** - Complete user flow from project list to launch
- Project list format: `<branch> - <project-name>` with progress `[<Type>: <phase>, x/y tasks completed]`
- Progress only includes task counts if the current phase has tasks
- Continuation prompt is optional (user can skip to continue from last state)

From `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` (lines 281-350):
- **Project discovery algorithm**: Scan `.sow/worktrees/`, read state files, handle corrupted/missing gracefully
- **Progress formatting**: Dynamic based on whether phase has tasks
- **Sorting**: Most recently modified first (helps users find recent work)

From `.sow/knowledge/designs/huh-library-verification.md`:
- **External editor**: Ctrl+E (not Ctrl+O as some early docs stated)
- Multi-line text area with CharLimit support
- Loading spinner available via `huh/spinner` package

### Project Discovery Specification

From technical design (lines 406-477), the `listProjects()` function must:
1. Scan `.sow/worktrees/` directory for subdirectories
2. For each directory, check for `.sow/project/state.yaml`
3. Load project state using SDK to extract:
   - Project name
   - Project type
   - Current phase/state
   - Task counts (if phase has tasks)
4. Skip directories with corrupted or missing state files (graceful degradation)
5. Sort by modification time (most recent first)
6. Return structured list for display

### Key Architectural Decisions

From `.sow/project/discovery/analysis.md`:
- **Current branch is irrelevant**: User can run from any branch, experience is identical
- **Worktree isolation**: Each project exists in its own worktree, no main repo interference
- **3-layer prompt system**: Base orchestrator + project type orchestrator + current state + optional user prompt
- **Registry pattern**: Project types configured via registry with type-specific prompts

---

## Implementation Scope

### What Needs to Be Built

#### 1. Project Discovery Function

**Location**: Shared utility in wizard package (or extracted to shared.go)

**Function signature**:
```go
type ProjectInfo struct {
    Branch         string
    Name           string
    Type           string
    Phase          string
    TasksCompleted int
    TasksTotal     int
    ModTime        time.Time
}

func listProjects(ctx *sow.Context) ([]ProjectInfo, error)
```

**Algorithm**:
1. Get worktrees directory: `<repoRoot>/.sow/worktrees/`
2. Read directory entries via `os.ReadDir()`
3. For each subdirectory:
   - Construct state file path: `<worktree>/.sow/project/state.yaml`
   - Check if state file exists via `os.Stat()`
   - If exists, load project state via SDK `state.LoadFromPath()`
   - If load fails, skip entry (corrupted/invalid)
   - Extract project metadata: name, type, phase
   - Count tasks if active phase has tasks:
     - Iterate phase.Tasks
     - Count where task.Status == "completed"
   - Store modification time from `os.Stat()` result
4. Sort results by ModTime descending (most recent first)
5. Return list

**Error handling**:
- If worktrees directory doesn't exist → return empty list (no projects)
- If individual state file corrupted → skip that project (log warning)
- If state loading fails → skip that project (graceful degradation)

#### 2. Progress Formatting Function

**Location**: Shared utility

**Function signature**:
```go
func formatProjectProgress(proj ProjectInfo) string
```

**Format rules**:
- If TasksTotal > 0: `"<Type>: <phase>, <completed>/<total> tasks completed"`
- If TasksTotal == 0: `"<Type>: <phase>"`

**Examples**:
- `"Standard: implementation, 3/5 tasks completed"`
- `"Design: active"` (design phase typically has no tasks)
- `"Exploration: gathering, 4/7 tasks completed"`

#### 3. State Machine Integration

**New states** (add to wizard state machine from work unit 001):
```go
const (
    StateProjectSelect  WizardState = "project_select"
    StateContinuePrompt WizardState = "continue_prompt"
)
```

**State transitions**:
- `StateEntry` → `StateProjectSelect` (when user selects "Continue existing project")
- `StateProjectSelect` → `StateContinuePrompt` (after valid project selected)
- `StateContinuePrompt` → `StateComplete` (trigger finalization)

**Cancel handling**:
- StateProjectSelect: User can select "Cancel" → `StateCancelled`
- StateContinuePrompt: User presses Esc → `StateCancelled`

#### 4. Project Selection Screen (StateProjectSelect)

**Implementation using huh library**:
```go
func (w *Wizard) handleProjectSelect() error {
    // Discover projects with spinner
    var projects []ProjectInfo
    var err error

    spinner.New().
        Type(spinner.Dot).
        Title("Discovering projects...").
        Action(func() {
            projects, err = listProjects(w.ctx)
        }).
        Run()

    if err != nil {
        return fmt.Errorf("failed to list projects: %w", err)
    }

    if len(projects) == 0 {
        fmt.Fprintln(os.Stderr, "No existing projects found")
        w.state = StateCancelled
        return nil
    }

    // Build selection options
    var selectedBranch string
    options := make([]huh.Option[string], 0, len(projects)+1)

    for _, proj := range projects {
        label := formatProjectOption(proj)
        options = append(options, huh.NewOption(label, proj.Branch))
    }
    options = append(options, huh.NewOption("Cancel", "cancel"))

    // Show selection form
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
        return err
    }

    if selectedBranch == "cancel" {
        w.state = StateCancelled
        return nil
    }

    // Validate project still exists
    proj, err := validateProjectExists(w.ctx, selectedBranch, projects)
    if err != nil {
        showError(fmt.Sprintf("Error: %s\n\nPress Enter to return to project list", err.Error()))
        return nil // Stay in StateProjectSelect to retry
    }

    w.choices["project"] = proj
    w.state = StateContinuePrompt
    return nil
}

func formatProjectOption(proj ProjectInfo) string {
    progress := formatProjectProgress(proj)
    return fmt.Sprintf("%s - %s\n    [%s]", proj.Branch, proj.Name, progress)
}
```

#### 5. Project Validation

**Function to verify project still exists**:
```go
func validateProjectExists(ctx *sow.Context, branch string, projects []ProjectInfo) (ProjectInfo, error) {
    // Find project in list
    for _, p := range projects {
        if p.Branch == branch {
            // Double-check state file still exists
            worktreePath := sow.WorktreePath(ctx.RepoRoot(), branch)
            statePath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")

            if _, err := os.Stat(statePath); err != nil {
                return ProjectInfo{}, fmt.Errorf("project no longer exists (state file missing)")
            }

            return p, nil
        }
    }

    return ProjectInfo{}, fmt.Errorf("project not found")
}
```

#### 6. Continuation Prompt Entry (StateContinuePrompt)

**Implementation**:
```go
func (w *Wizard) handleContinuePrompt() error {
    proj := w.choices["project"].(ProjectInfo)
    var prompt string

    // Build context display
    contextInfo := fmt.Sprintf(
        "Project: %s\nBranch: %s\nState: %s",
        proj.Name,
        proj.Branch,
        formatProjectProgress(proj),
    )

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewText().
                Title("What would you like to work on? (optional):").
                Description(contextInfo + "\n\nPress Ctrl+E to open $EDITOR for multi-line input").
                CharLimit(5000).
                Value(&prompt),
                // Editor enabled by default for Text fields
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return err
    }

    w.choices["prompt"] = prompt
    w.state = StateComplete
    return nil
}
```

#### 7. Finalization Flow

**Continuation-specific finalization** (in `w.finalize()` method):
```go
// In wizard finalize() method, handle continuation path:
if w.choices["action"] == "continue" {
    proj := w.choices["project"].(ProjectInfo)
    userPrompt := w.choices["prompt"].(string)

    // 1. Ensure worktree exists (idempotent)
    worktreePath := sow.WorktreePath(w.ctx.RepoRoot(), proj.Branch)
    if err := sow.EnsureWorktree(w.ctx, worktreePath, proj.Branch); err != nil {
        return fmt.Errorf("failed to ensure worktree: %w", err)
    }

    // 2. Create worktree context
    worktreeCtx, err := sow.NewContext(worktreePath)
    if err != nil {
        return fmt.Errorf("failed to create worktree context: %w", err)
    }

    // 3. Load fresh project state from worktree
    projectState, err := state.Load(worktreeCtx)
    if err != nil {
        return fmt.Errorf("failed to load project state: %w", err)
    }

    // 4. Generate 3-layer continuation prompt
    basePrompt, err := generateContinuePrompt(projectState)
    if err != nil {
        return fmt.Errorf("failed to generate continuation prompt: %w", err)
    }

    // 5. Append user prompt if provided
    fullPrompt := basePrompt
    if userPrompt != "" {
        fullPrompt += "\n\nUser request:\n" + userPrompt
    }

    // 6. Success message
    fmt.Fprintf(os.Stderr, "✓ Continuing project '%s' on branch %s\n", proj.Name, proj.Branch)

    // 7. Launch Claude in worktree
    return launchClaudeCode(worktreeCtx, fullPrompt, w.claudeFlags)
}
```

#### 8. Continuation Prompt Generation

**Extract from continue.go** and make shared:
```go
// generateContinuePrompt creates the 3-layer prompt for continuing projects.
// Layer 1: Base Orchestrator Introduction
// Layer 2: Project Type Orchestrator Prompt
// Layer 3: Current State Prompt
func generateContinuePrompt(proj *state.Project) (string, error) {
    var buf strings.Builder

    // Layer 1: Base Orchestrator Introduction
    baseOrch, err := templates.Render(prompts.FS, "templates/greet/orchestrator.md", nil)
    if err != nil {
        return "", fmt.Errorf("failed to render base orchestrator prompt: %w", err)
    }
    buf.WriteString(baseOrch)
    buf.WriteString("\n\n---\n\n")

    // Layer 2: Project Type Orchestrator Prompt
    projectTypePrompt := proj.Config().OrchestratorPrompt(proj)
    if projectTypePrompt != "" {
        buf.WriteString(projectTypePrompt)
        buf.WriteString("\n\n---\n\n")
    }

    // Layer 3: Current State Prompt
    currentState := proj.Machine().State()
    statePrompt := proj.Config().GetStatePrompt(currentState, proj)
    if statePrompt != "" {
        buf.WriteString(statePrompt)
        buf.WriteString("\n")
    }

    return buf.String(), nil
}
```

**Location**: Move to shared utility (work unit 001 may have already created this, coordinate)

---

## Acceptance Criteria

### Functional Requirements

1. **Project discovery works correctly**
   - All worktrees with valid `.sow/project/state.yaml` are found
   - Corrupted/missing state files are skipped gracefully (no crash)
   - Empty worktrees directory handled (shows "No projects found")

2. **Project list displays accurate information**
   - Branch name displayed correctly (preserving slashes)
   - Project name matches state.yaml
   - Project type shown correctly
   - Current phase/state shown correctly
   - Task progress shown ONLY when phase has tasks
   - Format: `"<Type>: <phase>, x/y tasks completed"` or `"<Type>: <phase>"`

3. **Sort order is correct**
   - Projects sorted by most recently modified first
   - Modification time based on state file mtime

4. **Project selection validation works**
   - Selected project verified to still exist before proceeding
   - If project deleted between discovery and selection: error shown, return to list
   - Cancel option works (returns to entry screen or exits)

5. **Continuation prompt entry functional**
   - Displays correct project context (name, branch, state)
   - Multi-line text area accepts input
   - Ctrl+E opens $EDITOR for external editing
   - Optional - user can leave empty (continue from last state)
   - Character limit enforced (5000 chars)

6. **State correctly loaded and prompt generated**
   - Fresh project state loaded from worktree (not cached)
   - 3-layer prompt structure correct:
     - Layer 1: Base orchestrator prompt
     - Layer 2: Project type orchestrator prompt
     - Layer 3: Current state prompt
   - User prompt appended if provided

7. **Claude launches successfully**
   - Launches in correct worktree directory
   - Receives correct continuation prompt
   - Claude flags passed through if provided

### Non-Functional Requirements

8. **No uncommitted changes check**
   - Unlike creation, continuation does NOT check uncommitted changes
   - This is intentional: worktree already exists, no branch switching needed

9. **Graceful error handling**
   - Missing worktrees directory: "No projects found" (not crash)
   - Corrupted state file: Skip project, log warning
   - Project deleted during selection: Error message, retry
   - Network/file system errors: Clear error messages

10. **Performance acceptable**
    - Project discovery completes in <2 seconds for 20 projects
    - Loading spinner shown for operations >500ms
    - No blocking UI during file I/O

---

## Technical Details

### Go Packages and Imports

```go
import (
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"

    "github.com/charmbracelet/huh"
    "github.com/charmbracelet/huh/spinner"
    "github.com/jmgilman/sow/cli/internal/prompts"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/jmgilman/sow/cli/internal/sdks/project/templates"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

### File Structure

```
cli/cmd/project/
├── wizard.go          # Main wizard state machine (work unit 001)
├── wizard_states.go   # State handlers (ADD: StateProjectSelect, StateContinuePrompt)
├── wizard_shared.go   # Shared utilities (ADD: project discovery, progress formatting)
└── continue.go        # DELETE after extracting logic
```

### Data Structures

**ProjectInfo** (used for project list):
```go
type ProjectInfo struct {
    Branch         string    // Git branch name (e.g., "feat/auth")
    Name           string    // Project name from state.yaml
    Type           string    // Project type (standard, exploration, design, breakdown)
    Phase          string    // Current phase/state
    TasksCompleted int       // Number of completed tasks (0 if no tasks)
    TasksTotal     int       // Total number of tasks (0 if no tasks)
    ModTime        time.Time // State file modification time
}
```

**Wizard choices** (stored in `w.choices` map):
- `"action"`: "continue"
- `"project"`: `ProjectInfo` struct
- `"prompt"`: `string` (user's continuation prompt, may be empty)

### Configuration

**Character limit**: 5000 characters for continuation prompt
**Loading spinner threshold**: 500ms (show spinner for operations that might take longer)
**Sorting**: Descending by ModTime (most recent first)

### Integration with Work Unit 001

This work unit depends on:
- State machine foundation (`WizardState`, `Wizard` struct, `handleState()` pattern)
- Claude launch utility (`launchClaudeCode()` function)
- Loading spinner usage patterns
- Error display functions (`showError()`)

This work unit extends:
- State machine with 2 new states: `StateProjectSelect`, `StateContinuePrompt`
- Wizard choices map with continuation-specific data

---

## Testing Requirements

### Unit Tests (TDD Approach)

**Note**: Write tests FIRST, then implement to pass them (TDD methodology)

1. **Test `listProjects()` function**
   - Empty worktrees directory → returns empty list
   - Valid projects → returns all with correct metadata
   - Corrupted state file → skips that project
   - Missing state file → skips that directory
   - Multiple projects → sorted by ModTime descending

2. **Test `formatProjectProgress()` function**
   - Phase with tasks → includes task counts
   - Phase without tasks → excludes task counts
   - Zero tasks completed → "0/5 tasks completed"
   - All tasks completed → "5/5 tasks completed"

3. **Test `validateProjectExists()` function**
   - Valid project → returns ProjectInfo
   - Missing state file → returns error
   - Project not in list → returns error

4. **Test `generateContinuePrompt()` function**
   - Contains all 3 layers
   - Layers separated by "---"
   - Handles missing project type prompt gracefully
   - Handles missing state prompt gracefully

### Integration Tests

5. **Test project discovery end-to-end**
   - Setup: Create test worktrees with state files
   - Execute: Call `listProjects()`
   - Verify: All projects found, metadata correct, sorted correctly
   - Teardown: Remove test worktrees

6. **Test state machine transitions**
   - StateEntry → StateProjectSelect (action="continue")
   - StateProjectSelect → StateContinuePrompt (valid selection)
   - StateProjectSelect → StateCancelled (cancel selected)
   - StateContinuePrompt → StateComplete (prompt submitted)
   - StateContinuePrompt → StateCancelled (Esc pressed)

7. **Test finalization flow**
   - Setup: Create test project in worktree
   - Execute: Complete continuation flow through finalization
   - Verify: Correct prompt generated, Claude would launch (mock)
   - Teardown: Remove test project

### Manual Testing Scenarios

8. **Happy path: Continue existing project**
   - Create project via creation workflow
   - Exit Claude
   - Run wizard, select "Continue existing project"
   - Verify project listed with correct progress
   - Select project, enter continuation prompt
   - Verify Claude launches with correct context

9. **Empty state: No projects exist**
   - Ensure no worktrees exist
   - Run wizard, select "Continue existing project"
   - Verify "No existing projects found" message
   - Verify returns to entry or exits gracefully

10. **Error recovery: Project deleted during selection**
    - Start wizard, list projects
    - Manually delete one project's state file
    - Select that project
    - Verify error message shown
    - Verify returns to project list

11. **External editor: Ctrl+E works**
    - Select continue, choose project
    - Press Ctrl+E at prompt entry
    - Verify $EDITOR opens
    - Write multi-line prompt, save, exit
    - Verify prompt captured correctly

12. **Skip prompt: Empty continuation**
    - Select continue, choose project
    - Leave prompt empty, press Enter
    - Verify Claude launches with base 3-layer prompt only

---

## Examples

### Example 1: Project List Display

```
Select a project to continue:

  ○ feat/auth - Add JWT authentication
    [Standard: implementation, 3/5 tasks completed]

  ○ design/cli-ux - CLI UX improvements
    [Design: active]

  ○ explore/web-agents - Web Based Agents
    [Exploration: gathering, 4/7 tasks completed]

  ○ breakdown/interactive-wizard - Break down wizard implementation
    [Breakdown: planning, 2/8 tasks completed]

  ○ Cancel

[Use arrow keys to navigate, Enter to select]
```

### Example 2: Continuation Prompt Entry

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

### Example 3: Generated Continuation Prompt Structure

```markdown
[LAYER 1: Base Orchestrator - from templates/greet/orchestrator.md]
You are the orchestrator agent...
(full base orchestrator prompt)

---

[LAYER 2: Project Type Orchestrator - from project type config]
You are coordinating a Standard project...
(project-type-specific orchestrator guidance)

---

[LAYER 3: Current State Prompt - from state machine]
You are in the implementation phase...
Current task: Implement token refresh logic
(state-specific guidance)

[APPENDED: User's continuation prompt if provided]
User request:
Let's focus on the token refresh logic and write
comprehensive integration tests for the auth flow
```

### Example 4: Code Structure

```go
// In wizard_states.go

func (w *Wizard) handleProjectSelect() error {
    // 1. Discover projects with loading indicator
    var projects []ProjectInfo
    var discoverErr error

    _ = spinner.New().
        Type(spinner.Dot).
        Title("Discovering projects...").
        Action(func() {
            projects, discoverErr = listProjects(w.ctx)
        }).
        Run()

    if discoverErr != nil {
        return fmt.Errorf("failed to discover projects: %w", discoverErr)
    }

    // 2. Handle empty list
    if len(projects) == 0 {
        fmt.Fprintln(os.Stderr, "\nNo existing projects found.\n")
        w.state = StateCancelled
        return nil
    }

    // 3. Build selection form
    var selectedBranch string
    options := make([]huh.Option[string], 0, len(projects)+1)

    for _, proj := range projects {
        // Format: "branch - name\n    [progress]"
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
        return err
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
        showError("Project not found")
        return nil // Stay in current state to retry
    }

    // Double-check state file still exists
    worktreePath := sow.WorktreePath(w.ctx.RepoRoot(), selectedBranch)
    statePath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")
    if _, err := os.Stat(statePath); err != nil {
        showError("Project no longer exists (state file missing)\n\nPress Enter to return to project list")
        return nil // Stay in current state to retry
    }

    // 7. Save selection and transition
    w.choices["project"] = *selectedProj
    w.state = StateContinuePrompt
    return nil
}
```

---

## Dependencies

### Required Prior Work

**Work Unit 001** (Wizard Foundation and State Machine):
- State machine types: `WizardState`, `Wizard` struct
- State handler pattern: `handleState()` switch
- Shared utilities: `launchClaudeCode()`, loading spinner patterns
- Error display: `showError()` function

### Existing Infrastructure to Reuse

- `cli/internal/sow/worktree.go`: `WorktreePath()`, `EnsureWorktree()`
- `cli/internal/sow/context.go`: Context creation and management
- `cli/internal/sdks/project/state/`: Project state SDK for loading state
- `cli/internal/prompts/`: Template rendering for orchestrator prompts

### Migration from Existing Code

**Extract from `cli/cmd/project/continue.go`**:
- Lines 167-196: `generateContinuePrompt()` function → Move to shared utilities
- Lines 103-106: Project loading pattern → Integrate into finalization
- Lines 110-114: Prompt generation pattern → Integrate into finalization

**After extraction, DELETE `continue.go`** (replaced by wizard)

---

## Constraints

### Performance Requirements

- Project discovery must complete in <2 seconds for up to 20 projects
- Loading spinner shown for any operation >500ms
- UI must remain responsive during file I/O

### Security Considerations

- Validate all file paths to prevent directory traversal
- Handle corrupted state files gracefully (no arbitrary code execution)
- Sanitize branch names before using in file paths

### Compatibility Requirements

- Must work with existing worktree structure
- Must load state files created by current project creation code
- Must work with all project types (standard, exploration, design, breakdown)

### UX Constraints

- No uncommitted changes check (unlike creation workflow)
- Project list must fit in terminal (use scrolling if >20 projects)
- Error messages must be actionable (tell user what to do)
- Cancel at any point should return cleanly (no orphaned state)

### Known Limitations

**huh library external editor bug** (GitHub issue #686):
- Ctrl+E may affect multiple fields in same group
- Mitigation: Keep text fields in separate groups (current design does this)
- Impact: Minimal, current design unaffected

**No remote project discovery**:
- Only discovers projects in local `.sow/worktrees/`
- Future work: Could add remote worktree discovery
- Out of scope for this work unit

---

## Relevant Inputs

This section lists all files that provide necessary context for implementing this work unit:

**Design documents**:
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` - Complete UX flow, Path 2 (lines 299-376), validation rules, error messages
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - State machine, huh integration, project discovery algorithm (lines 281-350, 406-477)
- `.sow/knowledge/designs/huh-library-verification.md` - Library capabilities, external editor Ctrl+E keybinding, loading spinner

**Existing code to migrate/reference**:
- `cli/cmd/project/continue.go` - Current continuation logic to extract (especially generateContinuePrompt)
- `cli/internal/sow/worktree.go` - WorktreePath(), EnsureWorktree() functions
- `cli/internal/sow/context.go` - Context creation and repository detection

**Discovery analysis**:
- `.sow/project/discovery/analysis.md` - Codebase overview, architectural patterns, project type configuration

**Work unit 001 deliverables** (dependencies):
- Wizard state machine foundation
- Shared utility functions
- Loading spinner patterns
- Error display utilities

---

## Notes for Implementers

### Critical Differences from Creation

**No uncommitted changes check**: This is the KEY difference from project creation. Continuation assumes the worktree already exists, so there's no risk of needing to switch branches in the main repo.

**Idempotent worktree creation**: `EnsureWorktree()` is idempotent - calling it on an existing worktree is safe and does nothing. Always call it to ensure worktree exists.

**Fresh state loading**: Always load project state fresh from the worktree, don't cache or reuse the ProjectInfo from discovery (it might be stale).

### Progress Display Nuances

Some phases have tasks (implementation, planning), others don't (active, gathering). Only show task counts when `TasksTotal > 0`.

Examples:
- Standard implementation: Has tasks → "Standard: implementation, 3/5 tasks completed"
- Design active: No tasks → "Design: active"
- Exploration gathering: Has tasks → "Exploration: gathering, 4/7 tasks completed"

### Error Handling Philosophy

**Fail gracefully in discovery**: If one project's state file is corrupted, skip it and log a warning. Don't fail the entire list operation.

**Fail loudly in selection**: If user selects a project that's gone missing, show clear error and return to list. This is user-facing so needs good messaging.

**Fail fatally in finalization**: If worktree creation or state loading fails during finalization, that's a real error - report it clearly.

### Sorting Rationale

Projects sorted by most recently modified help users quickly find their recent work. Use state file modification time (`os.Stat().ModTime()`) as the sort key.

### External Editor

The huh library uses **Ctrl+E** (not Ctrl+O) for external editor. This is enabled by default for Text fields. Set `CharLimit()` and optionally `EditorExtension(".md")`.

### Testing Strategy

Follow TDD: write tests first, implement to pass them. Focus on:
1. Discovery edge cases (empty, corrupted, missing)
2. Formatting variations (with/without tasks)
3. State machine transitions
4. Finalization flow integration

Manual testing critical for:
- Real terminal interaction
- External editor behavior
- Error message clarity
- Overall UX flow
