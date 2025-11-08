# Task 020: Branch, Worktree, and Project State Validation

## Context

This task is part of Work Unit 005: Validation and Error Handling for the interactive wizard. The wizard needs to detect conflicts before attempting to create projects or worktrees, such as:
- Branch already has a project
- Worktree exists but project is missing (inconsistent state)
- Branch/worktree already exists from previous operations

The state validation module provides functions to check the current state of a branch across three dimensions:
1. **Branch exists** in git repository
2. **Worktree exists** at `.sow/worktrees/{branch}/`
3. **Project exists** at `.sow/worktrees/{branch}/.sow/project/state.yaml`

These checks prevent state corruption and provide clear error messages when conflicts are detected.

**Architecture note**: This module provides state detection only. Error message formatting is handled by Task 030 (Error Display Components). This separation keeps concerns clean: this module answers "what is the state?", Task 030 answers "how do we tell the user?".

## Requirements

### 1. Branch State Struct

Create a `BranchState` struct in `wizard_helpers.go` to represent the current state of a branch:

```go
// BranchState represents the current state of a branch across git, worktrees, and projects.
type BranchState struct {
    BranchExists  bool  // Branch exists in git repository
    WorktreeExists bool  // Worktree directory exists at .sow/worktrees/{branch}/
    ProjectExists bool  // Project state.yaml exists in worktree
}
```

### 2. State Checking Function

Create `checkBranchState()` in `wizard_helpers.go` to examine the current state:

```go
// checkBranchState examines branch, worktree, and project state for a given branch name.
// Used before creation to detect conflicts.
//
// Returns:
//   - BranchState with all three boolean flags set
//   - error if filesystem or git operations fail
func checkBranchState(ctx *sow.Context, branchName string) (*BranchState, error)
```

**Implementation requirements**:

**Check 1: Branch Exists**
- Use `ctx.Git().Repository().Branches()` to list all branches
- Search for exact match with `branchName`
- Set `state.BranchExists = true` if found

**Check 2: Worktree Exists**
- Use `sow.WorktreePath(ctx.RepoRoot(), branchName)` to get worktree path
- Use `os.Stat()` to check if directory exists
- Set `state.WorktreeExists = true` if directory exists

**Check 3: Project Exists**
- Only check if `state.WorktreeExists == true`
- Build path: `filepath.Join(worktreePath, ".sow", "project", "state.yaml")`
- Use `os.Stat()` to check if file exists
- Set `state.ProjectExists = true` if file exists

**Error handling**:
- Return error if git operations fail (can't list branches)
- Ignore "not found" errors from `os.Stat()` (those are valid states)
- Return error for other filesystem errors

### 3. Project Creation Validation

Create `canCreateProject()` in `wizard_helpers.go` to validate that project creation is allowed:

```go
// canCreateProject validates that project creation is allowed on this branch.
// Returns error if:
//   - Branch already has a project (state.ProjectExists == true)
//   - Inconsistent state: worktree exists but project missing
//
// Returns nil if creation is allowed.
func canCreateProject(state *BranchState, branchName string) error
```

**Validation logic**:

1. **Check if project already exists**:
   - If `state.ProjectExists == true`:
   - Return error: `fmt.Errorf("branch '%s' already has a project", branchName)`

2. **Check for inconsistent state**:
   - If `state.WorktreeExists == true` AND `state.ProjectExists == false`:
   - Return error: `fmt.Errorf("worktree exists but project missing for branch '%s'", branchName)`

3. **Allow creation**:
   - If neither condition is true, return `nil`

**Note**: The actual error message formatting with recovery instructions is handled by Task 030. This function returns simple error messages that can be enhanced later.

### 4. Project Continuation Validation

Create `validateProjectExists()` in `wizard_helpers.go` for the "continue existing project" flow:

```go
// validateProjectExists checks that a project at given branch exists.
// Used when continuing projects (Work Unit 004).
//
// Returns error if:
//   - Branch doesn't exist
//   - Worktree doesn't exist
//   - Project doesn't exist in worktree
func validateProjectExists(ctx *sow.Context, branchName string) error
```

**Validation logic**:

1. Call `checkBranchState(ctx, branchName)` to get current state
2. Check `state.BranchExists`:
   - If false: return error `fmt.Errorf("branch '%s' does not exist", branchName)`
3. Check `state.WorktreeExists`:
   - If false: return error `fmt.Errorf("worktree for branch '%s' does not exist", branchName)`
4. Check `state.ProjectExists`:
   - If false: return error `fmt.Errorf("project for branch '%s' does not exist", branchName)`
5. If all three exist: return `nil`

### 5. List Existing Projects Helper

Create `listExistingProjects()` in `wizard_helpers.go` to find all existing projects:

```go
// listExistingProjects finds all branches with existing projects.
// Used by "continue existing project" screen to show available options.
//
// Returns:
//   - Slice of branch names that have projects
//   - error if filesystem or git operations fail
func listExistingProjects(ctx *sow.Context) ([]string, error)
```

**Implementation**:

1. Get all branches: `ctx.Git().Repository().Branches()`
2. For each branch:
   - Call `checkBranchState(ctx, branch)`
   - If `state.ProjectExists == true`, add to results
3. Sort results alphabetically
4. Return sorted slice

**Performance note**: This function may be called multiple times. Consider caching if performance becomes an issue, but start with simple implementation first.

## Acceptance Criteria

### Functional Requirements

1. **State Detection is Accurate**:
   - Correctly detects when branch exists in git
   - Correctly detects when worktree directory exists
   - Correctly detects when project state.yaml exists
   - All three checks work independently

2. **State Combinations are Handled**:
   - Branch exists + worktree exists + project exists → `canCreateProject()` returns error
   - Branch exists + worktree exists + no project → `canCreateProject()` returns error (inconsistent)
   - Branch exists + no worktree + no project → `canCreateProject()` returns nil (OK)
   - No branch + no worktree + no project → `canCreateProject()` returns nil (OK)

3. **Project Continuation Validation Works**:
   - Missing branch → error
   - Missing worktree → error
   - Missing project → error
   - All three exist → nil (OK)

4. **Project Listing Works**:
   - Finds all branches with projects
   - Results are sorted alphabetically
   - Handles empty result (no projects exist)

### Test Requirements (TDD Approach)

**Write ALL tests FIRST, then implement functions to pass them.**

1. **State Checking Tests** (`wizard_helpers_test.go`):
   ```go
   func TestCheckBranchState(t *testing.T)
   ```
   - Create test git repository with test helper
   - Test branch exists (create branch, check state)
   - Test worktree exists (create directory at worktree path)
   - Test project exists (create state.yaml in worktree)
   - Test all combinations of the three flags
   - At least 10 test cases

2. **Creation Validation Tests** (`wizard_helpers_test.go`):
   ```go
   func TestCanCreateProject(t *testing.T)
   ```
   - Test with all state combinations
   - Test error messages are clear
   - Test allows creation when appropriate
   - At least 8 test cases

3. **Continuation Validation Tests** (`wizard_helpers_test.go`):
   ```go
   func TestValidateProjectExists(t *testing.T)
   ```
   - Test missing branch
   - Test missing worktree
   - Test missing project
   - Test all exist (success case)
   - At least 6 test cases

4. **Project Listing Tests** (`wizard_helpers_test.go`):
   ```go
   func TestListExistingProjects(t *testing.T)
   ```
   - Test empty repository (no projects)
   - Test single project
   - Test multiple projects
   - Test sorting order
   - At least 5 test cases

### Code Quality

- All functions have clear godoc comments
- Error messages include branch name for context
- Tests use the existing test helpers from `wizard_helpers_test.go`
- Tests clean up after themselves (remove test files/directories)

## Technical Details

### File Location

**Add functions to**: `/cli/cmd/project/wizard_helpers.go`
**Add tests to**: `/cli/cmd/project/wizard_helpers_test.go`

### Imports Required

```go
import (
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

### Existing Functions to Reuse

- `sow.WorktreePath(repoRoot, branch)` - Returns worktree directory path
- `ctx.Git().Repository().Branches()` - Lists all branches
- `ctx.RepoRoot()` - Returns repository root path
- `os.Stat()` - Check if file/directory exists
- `filepath.Join()` - Build file paths

### Testing with Test Repository

The existing `wizard_helpers_test.go` already has test helpers. Study these patterns:

```go
func TestSomething(t *testing.T) {
    // Create test git repo
    tempDir := t.TempDir()

    // Initialize git repo
    // (see existing tests for helper functions)

    // Create test branches, worktrees, projects

    // Run your test

    // TempDir() auto-cleans on test completion
}
```

### State Yaml Location

Project state is always at:
```
.sow/worktrees/{branch}/.sow/project/state.yaml
```

**Example**:
- Branch: `feat/add-auth`
- Worktree: `.sow/worktrees/feat/add-auth/`
- State: `.sow/worktrees/feat/add-auth/.sow/project/state.yaml`

### Branch Slashes in Paths

Branches like `feat/add-auth` create nested directories:
```
.sow/worktrees/feat/add-auth/
```

The `sow.WorktreePath()` function handles this correctly by using `filepath.Join()`, which preserves slashes as directory separators.

## Relevant Inputs

**Design documents**:
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` (lines 460-515: error scenarios for state conflicts)
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` (lines 356-402: state detection patterns)

**Existing implementation to study**:
- `cli/cmd/project/wizard_helpers.go` (existing patterns and test helpers)
- `cli/cmd/project/wizard_helpers_test.go` (test structure and git test helpers)

**Existing functions to reuse**:
- `cli/internal/sow/worktree.go` (WorktreePath function, lines 11-16)
- `cli/internal/sow/git.go` (git repository operations)
- `cli/internal/sow/context.go` (Context struct and methods)

## Examples

### Example 1: Check Branch State Implementation

```go
// checkBranchState examines branch, worktree, and project state for a given branch name.
func checkBranchState(ctx *sow.Context, branchName string) (*BranchState, error) {
    state := &BranchState{}

    // Check if branch exists in git
    branches, err := ctx.Git().Repository().Branches()
    if err != nil {
        return nil, fmt.Errorf("failed to list branches: %w", err)
    }
    for _, b := range branches {
        if b == branchName {
            state.BranchExists = true
            break
        }
    }

    // Check if worktree directory exists
    worktreePath := sow.WorktreePath(ctx.RepoRoot(), branchName)
    if _, err := os.Stat(worktreePath); err == nil {
        state.WorktreeExists = true

        // If worktree exists, check if project exists
        projectPath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")
        if _, err := os.Stat(projectPath); err == nil {
            state.ProjectExists = true
        }
    }

    return state, nil
}
```

### Example 2: Validate Project Creation

```go
// canCreateProject validates that project creation is allowed on this branch.
func canCreateProject(state *BranchState, branchName string) error {
    // Check if project already exists
    if state.ProjectExists {
        return fmt.Errorf("branch '%s' already has a project", branchName)
    }

    // Check for inconsistent state (worktree without project)
    if state.WorktreeExists && !state.ProjectExists {
        return fmt.Errorf("worktree exists but project missing for branch '%s'", branchName)
    }

    // Creation is allowed
    return nil
}
```

### Example 3: List Existing Projects

```go
// listExistingProjects finds all branches with existing projects.
func listExistingProjects(ctx *sow.Context) ([]string, error) {
    branches, err := ctx.Git().Repository().Branches()
    if err != nil {
        return nil, fmt.Errorf("failed to list branches: %w", err)
    }

    var projects []string
    for _, branch := range branches {
        state, err := checkBranchState(ctx, branch)
        if err != nil {
            return nil, err
        }
        if state.ProjectExists {
            projects = append(projects, branch)
        }
    }

    // Sort alphabetically for consistent UI
    sort.Strings(projects)

    return projects, nil
}
```

### Example 4: Test Case Structure

```go
func TestCheckBranchState(t *testing.T) {
    tests := []struct {
        name           string
        setup          func(t *testing.T, repoPath string) // Setup test state
        branchName     string
        wantBranch     bool
        wantWorktree   bool
        wantProject    bool
    }{
        {
            name: "nothing exists",
            setup: func(t *testing.T, repoPath string) {
                // No setup needed
            },
            branchName:   "feat/test",
            wantBranch:   false,
            wantWorktree: false,
            wantProject:  false,
        },
        {
            name: "branch exists only",
            setup: func(t *testing.T, repoPath string) {
                // Create branch (use test helper)
            },
            branchName:   "feat/test",
            wantBranch:   true,
            wantWorktree: false,
            wantProject:  false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create test repo
            tempDir := t.TempDir()

            // Setup test state
            tt.setup(t, tempDir)

            // Create context (see existing tests for pattern)
            ctx := createTestContext(t, tempDir)

            // Run test
            state, err := checkBranchState(ctx, tt.branchName)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            // Verify results
            if state.BranchExists != tt.wantBranch {
                t.Errorf("BranchExists = %v, want %v", state.BranchExists, tt.wantBranch)
            }
            if state.WorktreeExists != tt.wantWorktree {
                t.Errorf("WorktreeExists = %v, want %v", state.WorktreeExists, tt.wantWorktree)
            }
            if state.ProjectExists != tt.wantProject {
                t.Errorf("ProjectExists = %v, want %v", state.ProjectExists, tt.wantProject)
            }
        })
    }
}
```

## Dependencies

### Required Before This Task

- **Task 010**: Branch name validation (provides `isValidBranchName` used indirectly)
- Existing git repository operations in `cli/internal/sow/git.go`
- Existing worktree path utilities in `cli/internal/sow/worktree.go`

### Provides For Other Tasks

- **Task 030**: Error display components (needs state information)
- **Work Unit 002**: Branch name flow (uses `canCreateProject`)
- **Work Unit 003**: GitHub issue flow (uses `canCreateProject`)
- **Work Unit 004**: Continue flow (uses `validateProjectExists` and `listExistingProjects`)

### External Dependencies

- Go standard library: `fmt`, `os`, `path/filepath`, `sort`
- Existing sow packages: `cli/internal/sow`

## Constraints

### Performance Requirements

- `checkBranchState()` should complete in < 100ms (filesystem + git ops)
- `listExistingProjects()` may be slow if many branches exist, but that's acceptable
- No caching initially (keep it simple)

### State Consistency

- State checks must be atomic (check all three in one function call)
- Don't mix state checking with error formatting (separation of concerns)
- State struct should be immutable after creation

### What NOT to Do

- **Don't format detailed error messages here** - That's Task 030's job
- **Don't modify git repository state** - This is read-only validation
- **Don't create worktrees or projects** - This is detection only
- **Don't add retry logic** - Return errors immediately
- **Don't cache state** - Always check fresh (state can change between checks)

## Notes for Implementer

### Separation of Concerns

This module detects state. Task 030 formats errors. Keep them separate:

**This task**:
```go
if state.ProjectExists {
    return fmt.Errorf("branch '%s' already has a project", branchName)
}
```

**Task 030 will enhance**:
```go
// Wrap the error with recovery instructions
return formatError(
    "Branch Already Has Project",
    err.Error(),
    "To continue this project:\n  Select 'Continue existing project' from main menu",
)
```

### Test Helpers Are Your Friend

The existing `wizard_helpers_test.go` has helpers for:
- Creating test git repositories
- Setting up test branches
- Creating test worktrees
- Cleaning up after tests

Study these and reuse them. Don't reinvent test infrastructure.

### State Detection is Critical

The wizard relies on accurate state detection to prevent:
- Creating projects twice (data loss)
- Worktree conflicts (git errors)
- Inconsistent state (corruption)

Make your state checking robust. Handle edge cases:
- Partial worktree creation (directory exists but incomplete)
- Missing state.yaml (worktree exists but no project)
- Stale worktrees (branch deleted but directory remains)

### Error Messages are Simple Here

Don't try to create elaborate error messages in this task. Simple descriptions like:
- "branch 'feat/xyz' already has a project"
- "worktree exists but project missing for branch 'feat/xyz'"

Are perfect. Task 030 will make them user-friendly.
