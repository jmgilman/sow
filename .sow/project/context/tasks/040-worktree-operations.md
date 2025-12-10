# Task 040: Implement Worktree Operations

## Context

This task is part of work unit 004: creating a new `libs/git` Go module. The previous tasks created the module foundation (010), Git operations (020), and GitHubClient (030). This task implements worktree-related operations.

**Critical Design Change**: The existing `EnsureWorktree(ctx *Context, path, branch string)` must be refactored to `EnsureWorktree(git *Git, repoRoot, path, branch string)` - accepting explicit parameters instead of a Context object.

## Requirements

### 1. Create worktree.go

Implement worktree operations as standalone functions:

```go
// WorktreePath returns the path where a worktree for the given branch should be.
// Preserves forward slashes in branch names to maintain git's branch namespacing.
// Example: branch "feat/auth" → "<repoRoot>/.sow/worktrees/feat/auth/"
func WorktreePath(repoRoot, branch string) string

// EnsureWorktree creates a git worktree at the specified path for the given branch.
// If the worktree already exists, returns nil (idempotent operation).
// Creates the branch if it doesn't exist.
//
// Parameters:
//   - g: Git instance for the main repository
//   - repoRoot: Absolute path to the main repository root
//   - path: Target path for the worktree
//   - branch: Branch name to checkout in the worktree
func EnsureWorktree(g *Git, repoRoot, path, branch string) error

// CheckUncommittedChanges verifies the repository has no uncommitted changes.
// Returns an error if uncommitted changes exist.
// Untracked files are allowed (they don't block worktree creation).
//
// Can be skipped in test environments by setting SOW_SKIP_UNCOMMITTED_CHECK=1.
func CheckUncommittedChanges(g *Git) error
```

### 2. Implementation Details

**WorktreePath**:
- Use `filepath.Join` to construct the path
- The branch name becomes part of the path (slashes preserved)
- Pattern: `{repoRoot}/.sow/worktrees/{branch}`

**EnsureWorktree**:
1. Check if worktree path already exists (idempotent)
2. Create parent directories for worktree path
3. Get current branch in main repo
4. If current branch == target branch, checkout a different branch first (git worktree limitation)
5. Check if target branch exists (`git show-ref --verify`)
6. If branch doesn't exist, create it from current HEAD
7. Create worktree using `git worktree add`
8. Use subprocess calls (`os/exec`) for git operations (more reliable than go-git for worktrees)

**CheckUncommittedChanges**:
1. Check for SOW_SKIP_UNCOMMITTED_CHECK environment variable
2. Get worktree status from go-git
3. Iterate through status, looking for:
   - Staged changes (Staging != ' ' && Staging != '?')
   - Modified or deleted files in worktree (Worktree == 'M' || Worktree == 'D')
4. Untracked files (status == '?') should NOT trigger an error

### 3. Important Design Decisions

**Why Real Filesystem Paths (not core.FS)**:

Unlike `libs/config` which uses `github.com/jmgilman/go/fs/core.FS` for filesystem abstraction, git and worktree operations **must** use real filesystem paths because:

1. **go-git library** - Opens repositories using real paths (`git.Open(repoRoot string)`)
2. **Git CLI commands** - Subprocess calls require real directories (`cmd.Dir = repoRoot`)
3. **Worktree creation** - Creates actual directories and files on disk
4. **Cross-process coordination** - The git CLI and go-git must see the same filesystem

This is an intentional design decision, not a limitation to work around. Git repositories are fundamentally tied to the real filesystem.

**Subprocess Usage**:

The worktree operations use `os/exec` directly (not `libs/exec.Executor`) because:
- Git worktree operations are complex and go-git support is limited
- The git CLI is reliable and always available where git repos exist
- This matches the existing implementation behavior

```go
import "os/exec"

cmd := exec.CommandContext(ctx, "git", "worktree", "add", path, branch)
cmd.Dir = repoRoot
output, err := cmd.CombinedOutput()
```

## Acceptance Criteria

1. **WorktreePath generates correct paths**: Handles simple and slash-containing branches
2. **EnsureWorktree is idempotent**: Returns nil if worktree exists
3. **EnsureWorktree creates worktree**: Successfully creates worktree at path
4. **EnsureWorktree creates branch**: Creates branch if it doesn't exist
5. **EnsureWorktree handles same-branch case**: Switches away before creating worktree
6. **CheckUncommittedChanges detects changes**: Returns error for modified files
7. **CheckUncommittedChanges allows untracked**: No error for untracked-only
8. **CheckUncommittedChanges respects env var**: Skips check when SOW_SKIP_UNCOMMITTED_CHECK=1
9. **All tests pass**: Comprehensive test coverage

### Test Requirements (TDD)

Write tests in `worktree_test.go`:

**WorktreePath tests:**
- Simple branch name: "main" → "{root}/.sow/worktrees/main"
- Slashed branch: "feat/auth" → "{root}/.sow/worktrees/feat/auth"
- Multiple slashes: "feature/epic/task" → "{root}/.sow/worktrees/feature/epic/task"

**EnsureWorktree tests:**
- Creates worktree when path doesn't exist
- Returns nil when worktree already exists (idempotent)
- Creates branch if it doesn't exist
- (Note: Testing branch-switching edge case is complex - may need integration test)

**CheckUncommittedChanges tests:**
- Returns nil for clean repository
- Returns nil for repository with only untracked files
- Returns error for modified tracked files
- Returns error for staged changes
- Returns nil when SOW_SKIP_UNCOMMITTED_CHECK=1 (even with changes)

**Testing Note**: These tests require **real git repositories in temp directories** using `t.TempDir()`. Unlike `libs/config` which uses `billy.MemoryFS` for in-memory testing, git operations cannot use in-memory filesystems because:
- go-git opens repos by path string
- git CLI commands need real directories
- Worktrees create actual files on disk

Follow test patterns from `cli/internal/sow/worktree_test.go`.

## Technical Details

### Import Structure

```go
import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)
```

### Using os/exec with Context

```go
func EnsureWorktree(g *Git, repoRoot, path, branch string) error {
    // Use background context for git commands
    ctx := context.Background()

    // Example: check if branch exists
    checkCmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
    checkCmd.Dir = repoRoot
    branchExists := checkCmd.Run() == nil

    // Example: create worktree
    addCmd := exec.CommandContext(ctx, "git", "worktree", "add", path, branch)
    addCmd.Dir = repoRoot
    if output, err := addCmd.CombinedOutput(); err != nil {
        return fmt.Errorf("failed to add worktree: %s", output)
    }
    return nil
}
```

### Status Codes in go-git

```go
// go-git/v5/plumbing/format/index/index.go
// Worktree status codes:
// ' ' = unmodified
// 'M' = modified
// 'D' = deleted
// '?' = untracked
// etc.
```

### Environment Variable Check

```go
if os.Getenv("SOW_SKIP_UNCOMMITTED_CHECK") == "1" {
    return nil
}
```

## Relevant Inputs

- `cli/internal/sow/worktree.go` - Source implementation (CRITICAL - read this!)
- `cli/internal/sow/worktree_test.go` - Test patterns to follow
- `libs/git/git.go` - Git struct from task 020
- `libs/git/errors.go` - Error types from task 010
- `libs/config/repo.go` - Reference for core.FS pattern (contrast: worktree uses real paths)
- `libs/config/repo_test.go` - Reference for billy.MemoryFS testing (contrast: worktree uses t.TempDir())
- `.standards/STYLE.md` - Coding standards
- `.standards/TESTING.md` - Testing standards

## Examples

### Using WorktreePath

```go
import "github.com/jmgilman/sow/libs/git"

// Get worktree path for a branch
repoRoot := "/home/user/myrepo"
branch := "feat/new-feature"
wtPath := git.WorktreePath(repoRoot, branch)
// wtPath = "/home/user/myrepo/.sow/worktrees/feat/new-feature"
```

### Using EnsureWorktree

```go
import "github.com/jmgilman/sow/libs/git"

g, err := git.NewGit(repoRoot)
if err != nil {
    return err
}

// Check for uncommitted changes first
if err := git.CheckUncommittedChanges(g); err != nil {
    return fmt.Errorf("cannot create worktree: %w", err)
}

// Create the worktree
wtPath := git.WorktreePath(repoRoot, "feat/auth")
if err := git.EnsureWorktree(g, repoRoot, wtPath, "feat/auth"); err != nil {
    return fmt.Errorf("failed to create worktree: %w", err)
}
```

### Testing Pattern

```go
func TestEnsureWorktree_CreatesWhenMissing(t *testing.T) {
    tempDir := t.TempDir()

    // Initialize git repo
    repo, err := gogit.PlainInit(tempDir, false)
    require.NoError(t, err)

    // Create initial commit
    wt, err := repo.Worktree()
    require.NoError(t, err)

    testFile := filepath.Join(tempDir, "test.txt")
    err = os.WriteFile(testFile, []byte("test"), 0644)
    require.NoError(t, err)

    _, err = wt.Add("test.txt")
    require.NoError(t, err)

    _, err = wt.Commit("initial", &gogit.CommitOptions{
        Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
    })
    require.NoError(t, err)

    // Create test branch
    headRef, err := repo.Head()
    require.NoError(t, err)
    branchRef := plumbing.NewBranchReferenceName("test-branch")
    err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
    require.NoError(t, err)

    // Test EnsureWorktree
    g, err := git.NewGit(tempDir)
    require.NoError(t, err)

    worktreePath := filepath.Join(tempDir, ".sow", "worktrees", "test-branch")
    err = git.EnsureWorktree(g, tempDir, worktreePath, "test-branch")
    require.NoError(t, err)

    // Verify worktree exists
    _, err = os.Stat(worktreePath)
    assert.NoError(t, err)
}
```

## Dependencies

- Task 010 (module foundation) must complete first
- Task 020 (Git operations) must complete first - provides Git struct

## Constraints

- Do NOT import from `cli/internal/sow/` - this module must be standalone
- Do NOT use libs/exec for git commands - use os/exec directly
- Worktree operations MUST use the git CLI (go-git worktree support is limited)
- Tests must create real git repositories (no mocking git internals)
- Match existing behavior exactly for compatibility
