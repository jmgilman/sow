# Task 010: Project Discovery Utilities

## Context

This task implements the core discovery and formatting utilities for the project continuation workflow. These utilities enable the wizard to discover all active projects across worktrees and format their progress information for user-friendly display.

This is part of work unit 004 (Project Continuation Workflow), which adds interactive project continuation to the sow wizard. The wizard foundation and state machine were implemented in work unit 001.

The project continuation path differs from creation in one critical way: there is NO uncommitted changes check, since the worktree already exists and we're not switching branches in the main repo.

## Requirements

Implement three utility functions in `cli/cmd/project/wizard_helpers.go`:

### 1. ProjectInfo Data Structure

Create a struct to hold project metadata for display:

```go
type ProjectInfo struct {
    Branch         string    // Git branch name (e.g., "feat/auth")
    Name           string    // Project name from state.yaml
    Type           string    // Project type (standard, exploration, design, breakdown)
    Phase          string    // Current phase/state from state machine
    TasksCompleted int       // Number of completed tasks (0 if phase has no tasks)
    TasksTotal     int       // Total number of tasks (0 if phase has no tasks)
    ModTime        time.Time // State file modification time for sorting
}
```

### 2. listProjects Function

Implement project discovery by scanning the worktrees directory:

**Function signature:**
```go
func listProjects(ctx *sow.Context) ([]ProjectInfo, error)
```

**Algorithm:**
1. Construct worktrees directory path: `<mainRepoRoot>/.sow/worktrees/`
   - IMPORTANT: Use `ctx.MainRepoRoot()` NOT `ctx.RepoRoot()` to handle being run from within a worktree
2. Read directory entries via `os.ReadDir(worktreesDir)`
3. If directory doesn't exist, return empty slice (not an error)
4. For each subdirectory:
   - Branch name is the subdirectory name (preserves forward slashes)
   - Construct state file path: `<worktree>/.sow/project/state.yaml`
   - Check if state file exists via `os.Stat()`
   - If doesn't exist, skip (not a project)
   - Load project state via `state.Load()` with a worktree context
   - If load fails, skip gracefully (corrupted project - log to stderr)
   - Extract metadata:
     - Name: `proj.Name`
     - Type: `proj.Type`
     - Phase: `proj.Machine().State().String()`
     - Tasks: Iterate ALL phases, count tasks where `task.Status == "completed"` and total tasks
   - Store modification time from `os.Stat()` result
5. Sort results by `ModTime` descending (most recent first)
6. Return sorted list

**Error handling:**
- Missing worktrees directory → return empty list, nil error
- Individual state file missing → skip that directory
- State loading failure → print warning to stderr, skip that project
- Other file system errors → return error

**Task counting logic:**
- Must count tasks across ALL phases (not just active phase)
- This matches the project-wide task completion view users expect
- Example: If implementation has 5 tasks (3 complete) and review has 2 tasks (0 complete), totals are 3/7

### 3. formatProjectProgress Function

Format progress information for display in project selection:

**Function signature:**
```go
func formatProjectProgress(proj ProjectInfo) string
```

**Format rules:**
- If `TasksTotal > 0`: return `"<Type>: <phase>, <completed>/<total> tasks completed"`
- If `TasksTotal == 0`: return `"<Type>: <phase>"`

**Examples:**
- Standard project with tasks: `"Standard: implementation, 3/5 tasks completed"`
- Design project without tasks: `"Design: active"`
- Exploration with tasks: `"Exploration: gathering, 4/7 tasks completed"`

**Implementation notes:**
- Type should be title-cased (use `strings.Title()` on first letter)
- Phase should be lowercase as-is from state machine
- Task counts use simple fraction format: `x/y tasks completed`

## Acceptance Criteria

### Functional Requirements

1. **ProjectInfo struct correctly defined**
   - All fields present with correct types
   - Includes time.Time for ModTime sorting

2. **listProjects discovers all valid projects**
   - Finds all subdirectories in `.sow/worktrees/`
   - Correctly loads project state from each worktree
   - Skips directories without `.sow/project/state.yaml`
   - Skips corrupted projects with warning (no crash)
   - Returns empty list for missing worktrees directory (no error)

3. **Task counting is accurate**
   - Counts tasks across ALL phases, not just active phase
   - Correctly identifies completed vs total tasks
   - Handles phases with no tasks (counts as 0/0)

4. **Sorting works correctly**
   - Projects sorted by modification time, most recent first
   - Uses state.yaml file modification time

5. **Progress formatting follows specification**
   - Includes task counts only when TasksTotal > 0
   - Format matches exactly: `"<Type>: <phase>, x/y tasks completed"` or `"<Type>: <phase>"`
   - Type is capitalized, phase is lowercase

6. **Error handling is graceful**
   - Missing worktrees directory: returns empty list, no error
   - Corrupted state file: warning to stderr, skip project
   - File system errors: clear error message returned

### Test Requirements (TDD Approach)

Write tests FIRST in `cli/cmd/project/wizard_helpers_test.go`, then implement to pass them:

**Unit tests for listProjects:**
- Empty worktrees directory → returns empty list, no error
- Missing worktrees directory → returns empty list, no error
- Single valid project → returns list with 1 entry, correct metadata
- Multiple projects → returns all, sorted by ModTime descending
- Corrupted state file → skips that project, warns, returns others
- Directory without state file → skips that directory
- Project with tasks in multiple phases → counts all tasks correctly
- Project with no tasks → returns 0/0 for task counts

**Unit tests for formatProjectProgress:**
- Project with tasks → includes ", x/y tasks completed"
- Project without tasks → excludes task portion
- Correct capitalization of type name
- Phase name preserved as-is from state machine

**Integration test:**
- Create test worktrees with state files
- Verify listProjects finds all and returns correct data
- Verify sorting by modification time works

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

    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

### File Location

Add to existing file: `cli/cmd/project/wizard_helpers.go`

This file already exists and contains shared utility functions like `normalizeName`, `showError`, `withSpinner`, etc.

### Context Handling for Worktrees

**CRITICAL:** The wizard might be run from within a worktree OR from the main repo. The listProjects function must handle both cases:

```go
// CORRECT: Use MainRepoRoot() to get the main repo path regardless of where we're running
worktreesDir := filepath.Join(ctx.MainRepoRoot(), ".sow", "worktrees")

// WRONG: Using RepoRoot() would fail if run from within a worktree
// worktreesDir := filepath.Join(ctx.RepoRoot(), ".sow", "worktrees")
```

When loading individual project states, create a new context for each worktree:

```go
worktreePath := filepath.Join(worktreesDir, branchName)
worktreeCtx, err := sow.NewContext(worktreePath)
if err != nil {
    // Log warning and skip
    fmt.Fprintf(os.Stderr, "Warning: failed to create context for %s: %v\n", branchName, err)
    continue
}

proj, err := state.Load(worktreeCtx)
// ...
```

### Task Counting Implementation

Count tasks across ALL phases (not just active):

```go
var tasksCompleted, tasksTotal int
for _, phase := range proj.Phases {
    for _, task := range phase.Tasks {
        tasksTotal++
        if task.Status == "completed" {
            tasksCompleted++
        }
    }
}
```

### Sorting Implementation

Use `sort.Slice` with descending ModTime comparison:

```go
sort.Slice(projects, func(i, j int) bool {
    return projects[i].ModTime.After(projects[j].ModTime)
})
```

## Relevant Inputs

**Design documents:**
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - Lines 406-477 contain the exact listProjects algorithm specification
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` - Lines 299-337 show the UX flow and progress format specification

**Existing helper functions:**
- `cli/cmd/project/wizard_helpers.go` - Pattern for shared utilities, existing helper functions to reference

**SDK state loading:**
- `cli/internal/sdks/project/state/loader.go` - `Load()` function for loading project state
- `cli/internal/sdks/project/state/project.go` - Project structure and methods like `Machine()`

**Context handling:**
- `cli/internal/sow/context.go` - Context creation, MainRepoRoot() vs RepoRoot() distinction
- `cli/internal/sow/worktree.go` - WorktreePath() function for constructing worktree paths

**Schema definitions:**
- `cli/schemas/project/cue_types_gen.go` - ProjectState, PhaseState, TaskState structure definitions

## Examples

### Example 1: listProjects Usage

```go
// In wizard state handler
projects, err := listProjects(w.ctx)
if err != nil {
    return fmt.Errorf("failed to discover projects: %w", err)
}

if len(projects) == 0 {
    fmt.Fprintln(os.Stderr, "No existing projects found")
    w.state = StateCancelled
    return nil
}

// Display projects...
```

### Example 2: formatProjectProgress Output

```go
proj1 := ProjectInfo{
    Type:           "standard",
    Phase:          "implementation",
    TasksCompleted: 3,
    TasksTotal:     5,
}
// formatProjectProgress(proj1) → "Standard: implementation, 3/5 tasks completed"

proj2 := ProjectInfo{
    Type:      "design",
    Phase:     "active",
    TasksTotal: 0, // No tasks
}
// formatProjectProgress(proj2) → "Design: active"
```

### Example 3: Test Structure

```go
func TestListProjects(t *testing.T) {
    t.Run("empty worktrees directory returns empty list", func(t *testing.T) {
        // Setup: Create temp dir with empty worktrees/
        // Execute: Call listProjects
        // Assert: len(projects) == 0, err == nil
    })

    t.Run("multiple projects sorted by modification time", func(t *testing.T) {
        // Setup: Create 3 test projects with different mtimes
        // Execute: Call listProjects
        // Assert: Returns 3 projects in correct order (newest first)
    })

    // Additional test cases...
}

func TestFormatProjectProgress(t *testing.T) {
    t.Run("project with tasks includes task counts", func(t *testing.T) {
        proj := ProjectInfo{
            Type:           "standard",
            Phase:          "implementation",
            TasksCompleted: 3,
            TasksTotal:     5,
        }
        result := formatProjectProgress(proj)
        expected := "Standard: implementation, 3/5 tasks completed"
        if result != expected {
            t.Errorf("got %q, want %q", result, expected)
        }
    })

    // Additional test cases...
}
```

## Dependencies

**Required from work unit 001:**
- Wizard state machine foundation exists in `cli/cmd/project/wizard_state.go`
- Helper functions file exists at `cli/cmd/project/wizard_helpers.go`
- Test file exists at `cli/cmd/project/wizard_helpers_test.go`

**Existing infrastructure:**
- `cli/internal/sow/context.go` - Context type with MainRepoRoot() method
- `cli/internal/sow/worktree.go` - WorktreePath() function
- `cli/internal/sdks/project/state/` - State loading SDK

## Constraints

### Performance Requirements

- Project discovery must complete in <2 seconds for up to 20 projects
- This is a read-only operation, no state modification
- File I/O should be efficient (no redundant reads)

### Error Handling Philosophy

- **Fail gracefully during discovery**: If one project's state is corrupted, skip it with a warning but continue discovering others
- **Return empty list for missing directory**: This is not an error condition, just means no projects exist yet
- **Clear error messages**: If a real error occurs (permissions, etc.), return descriptive error

### Data Integrity

- Never modify state files during discovery (read-only operation)
- Handle concurrent access gracefully (another process might be modifying projects)
- Use file modification time from stat, not custom tracking

### Testing Strategy

- Follow TDD: Write tests first, implement to pass them
- Use temporary directories for test fixtures
- Clean up test worktrees in defer blocks
- Test both happy path and error cases
- Use table-driven tests for formatting variations
