# Task Log

## 2025-11-07 - Iteration 1

### TDD: Writing Tests First

**Action**: Wrote comprehensive unit tests for project discovery utilities
**Files**:
- `cli/cmd/project/wizard_helpers_test.go`

**Tests Added**:
1. `TestFormatProjectProgress` - 5 test cases covering different project types and task count scenarios
2. `TestListProjects_EmptyWorktreesDirectory` - Verifies empty list returned for empty worktrees dir
3. `TestListProjects_MissingWorktreesDirectory` - Verifies no error when worktrees dir doesn't exist
4. `TestListProjects_DirectoryWithoutStateFile` - Verifies skipping directories without state files
5. `TestListProjects_SingleValidProject` - Verifies correct metadata extraction from a single project
6. `TestListProjects_MultipleProjectsSorted` - Verifies sorting by modification time (newest first)
7. `TestListProjects_ProjectWithTasks` - Verifies task counting across multiple phases

**Rationale**: Following TDD methodology - tests define the behavior before implementation exists.

### TDD: Implementation (GREEN Phase)

**Action**: Implemented project discovery utilities to make all tests pass
**Files**:
- `cli/cmd/project/wizard_helpers.go`
- `cli/cmd/project/wizard_helpers_test.go`

**Implementation Details**:

1. **ProjectInfo struct** - Holds project metadata for display:
   - Branch, Name, Type, Phase (all string fields)
   - TasksCompleted, TasksTotal (int fields for task counts)
   - ModTime (time.Time for sorting by recency)

2. **listProjects function** - Discovers projects by scanning worktrees:
   - Uses `ctx.MainRepoRoot()` to handle being run from worktree or main repo
   - Uses `filepath.Walk()` to recursively find all `state.yaml` files
   - Validates state files are in `.sow/project/` directories
   - Creates separate context for each worktree to load project state
   - Counts tasks across ALL phases (not just active phase)
   - Skips corrupted projects with warnings to stderr (graceful degradation)
   - Returns empty list for missing worktrees directory (not an error)
   - Sorts by modification time descending (most recent first)

3. **formatProjectProgress function** - Formats progress info for display:
   - Capitalizes first letter of project type
   - Includes task counts only when TasksTotal > 0
   - Format: "Standard: implementation, 3/5 tasks completed" or "Design: active"
   - Phase name preserved as-is from state machine (e.g., "ImplementationExecuting")

**Test Results**: All 7 test cases pass (plus all existing tests still pass)

**Key Design Decisions**:
- Used filepath.Walk instead of ReadDir to handle nested branch names (e.g., "feat/auth")
- Graceful error handling: corrupted projects print warnings but don't fail the whole operation
- Task counting across ALL phases matches user expectation of project-wide progress
- Sorting by state file mtime (not directory mtime) for accurate recency tracking
