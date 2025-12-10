# Task 020: Implement Git Operations Struct

## Context

This task is part of work unit 004: creating a new `libs/git` Go module. The previous task (010) created the module foundation with types and errors. This task implements the `Git` struct that provides git repository operations.

The `Git` struct wraps the `github.com/jmgilman/go/git` package with sow-specific conveniences. It is decoupled from `sow.Context` - it accepts a repo root path and uses the go-git library directly.

**Key Design Decisions**:

1. The Git struct does NOT use `libs/exec.Executor` because it leverages the go-git library for most operations. The `exec.Executor` is only used by the GitHub CLI adapter (task 030).

2. **Why Real Filesystem Paths (not core.FS)**: Unlike `libs/config` which uses `github.com/jmgilman/go/fs/core.FS` for filesystem abstraction, git operations **must** use real filesystem paths because the go-git library (`github.com/jmgilman/go/git`) opens repositories using `git.Open(repoRoot string)` - a path-based API that requires real directories containing `.git/`. This is an intentional design decision since git repositories are fundamentally tied to the real filesystem.

## Requirements

### 1. Create git.go

Implement the `Git` struct that wraps git operations:

```go
// Git provides git repository operations.
//
// This type wraps github.com/jmgilman/go/git.Repository with sow-specific
// conveniences and protected branch checking.
type Git struct {
    repo     *git.Repository
    repoRoot string
}

// NewGit creates a new Git instance for the repository.
//
// The repoRoot should be the absolute path to the git repository root
// (the directory containing .git/).
//
// Returns ErrNotGitRepository if the directory is not a git repository.
func NewGit(repoRoot string) (*Git, error)

// Repository returns the underlying git.Repository for advanced operations.
func (g *Git) Repository() *git.Repository

// RepoRoot returns the absolute path to the repository root.
func (g *Git) RepoRoot() string

// CurrentBranch returns the name of the current git branch.
// Returns an empty string if HEAD is in detached state.
func (g *Git) CurrentBranch() (string, error)

// IsProtectedBranch checks if the given branch name is protected (main/master).
func (g *Git) IsProtectedBranch(branch string) bool

// HasUncommittedChanges checks if the repository has uncommitted changes.
// Returns true if there are modified, staged, or deleted files.
// Untracked files are NOT considered uncommitted changes.
func (g *Git) HasUncommittedChanges() (bool, error)

// Branches returns a list of all local branch names.
func (g *Git) Branches() ([]string, error)

// CheckoutBranch checks out the specified branch.
func (g *Git) CheckoutBranch(branchName string) error
```

### 2. Implementation Details

The implementation should:
- Use `github.com/jmgilman/go/git.Open()` to open the repository
- Return `ErrNotGitRepository` (from errors.go) when the path is not a git repo
- Use the underlying go-git library (`g.repo.Underlying()`) for operations not exposed by the wrapper
- Filter branches to only return local branches (not remote tracking branches)
- Be consistent with the existing implementation in `cli/internal/sow/git.go`

### 3. Error Handling

- `NewGit`: Return `ErrNotGitRepository{Path: repoRoot}` if not a git repo
- `CurrentBranch`: Wrap go-git errors with context
- `HasUncommittedChanges`: Wrap worktree access errors
- `CheckoutBranch`: Wrap checkout errors with branch name context

## Acceptance Criteria

1. **Git struct compiles**: All methods compile without errors
2. **NewGit opens repositories**: Successfully opens valid git repos
3. **NewGit rejects non-repos**: Returns ErrNotGitRepository for non-git directories
4. **CurrentBranch returns branch**: Returns correct branch name for HEAD
5. **IsProtectedBranch detects main/master**: Returns true for "main" and "master"
6. **HasUncommittedChanges detects changes**: Returns true for modified files
7. **HasUncommittedChanges ignores untracked**: Returns false for only untracked files
8. **Branches lists local only**: Does not include remote tracking branches
9. **CheckoutBranch switches branches**: Successfully checks out existing branches
10. **All tests pass**: Unit tests cover all behaviors

### Test Requirements (TDD)

Write tests in `git_test.go`:

**NewGit tests:**
- Opens a valid git repository successfully
- Returns ErrNotGitRepository for non-git directory
- Repository() returns non-nil after successful open
- RepoRoot() returns the path passed to NewGit

**CurrentBranch tests:**
- Returns correct branch name (e.g., "main", "feature/test")
- Returns empty string for detached HEAD (if applicable)

**IsProtectedBranch tests:**
- Returns true for "main"
- Returns true for "master"
- Returns false for other branches (e.g., "develop", "feature/x")

**HasUncommittedChanges tests:**
- Returns false for clean repository
- Returns true for modified tracked files
- Returns true for staged changes
- Returns false for only untracked files (important!)

**Branches tests:**
- Returns list of local branches
- Does not include remote tracking branches
- Returns empty slice for repo with only detached HEAD

**CheckoutBranch tests:**
- Successfully checks out existing branch
- Returns error for non-existent branch

**Testing Note**: These tests require **real git repositories in temp directories** using `t.TempDir()` and go-git. Unlike `libs/config` which uses `billy.MemoryFS` for in-memory testing, git operations cannot use in-memory filesystems because go-git opens repos by path string. Follow the pattern from `cli/internal/sow/worktree_test.go`.

## Technical Details

### Dependencies

Add to go.mod:
```go
require (
    github.com/go-git/go-git/v5 v5.x.x
    github.com/jmgilman/go/git v0.0.0
)
```

### Import Structure

```go
import (
    "fmt"

    gogit "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/jmgilman/go/git"
)
```

### Protected Branches

Protected branches are hardcoded as "main" and "master". This matches the existing behavior and is simple. No configuration needed.

### Receiver Names

Use `g` as the receiver name (short, following STYLE.md).

## Relevant Inputs

- `cli/internal/sow/git.go` - Source implementation to adapt (read this carefully!)
- `cli/internal/sow/worktree_test.go` - Testing patterns for git operations
- `libs/git/types.go` - Types from task 010 (once complete)
- `libs/git/errors.go` - Error types from task 010 (once complete)
- `libs/config/repo.go` - Reference for core.FS pattern (contrast: git uses real paths)
- `libs/config/repo_test.go` - Reference for billy.MemoryFS testing (contrast: git uses t.TempDir())
- `.standards/STYLE.md` - Coding standards
- `.standards/TESTING.md` - Testing standards

## Examples

### Creating a Git Instance

```go
import "github.com/jmgilman/sow/libs/git"

g, err := git.NewGit("/path/to/repo")
if err != nil {
    var notRepo git.ErrNotGitRepository
    if errors.As(err, &notRepo) {
        fmt.Printf("%s is not a git repository\n", notRepo.Path)
    }
    return err
}
```

### Checking Branch Status

```go
branch, err := g.CurrentBranch()
if err != nil {
    return err
}

if g.IsProtectedBranch(branch) {
    return fmt.Errorf("cannot work on protected branch: %s", branch)
}
```

### Testing Pattern

```go
func TestGit_CurrentBranch(t *testing.T) {
    // Create temp directory
    tempDir := t.TempDir()

    // Initialize git repo with go-git
    repo, err := gogit.PlainInit(tempDir, false)
    require.NoError(t, err)

    // Create initial commit (required for branch operations)
    wt, err := repo.Worktree()
    require.NoError(t, err)

    testFile := filepath.Join(tempDir, "test.txt")
    err = os.WriteFile(testFile, []byte("test"), 0644)
    require.NoError(t, err)

    _, err = wt.Add("test.txt")
    require.NoError(t, err)

    _, err = wt.Commit("initial", &gogit.CommitOptions{
        Author: &object.Signature{
            Name:  "Test",
            Email: "test@example.com",
            When:  time.Now(),
        },
    })
    require.NoError(t, err)

    // Now test our Git wrapper
    g, err := git.NewGit(tempDir)
    require.NoError(t, err)

    branch, err := g.CurrentBranch()
    require.NoError(t, err)
    assert.Equal(t, "master", branch) // go-git defaults to master
}
```

## Dependencies

- Task 010 (module foundation) must complete first - provides types.go and errors.go
- `github.com/jmgilman/go/git` package must be available

## Constraints

- Do NOT import from `cli/internal/sow/` - this module must be standalone
- Do NOT use `exec.Executor` for git operations - use go-git library directly
- Do NOT add worktree operations yet - those come in task 040
- Do NOT add GitHub operations yet - those come in task 030
- The Git struct should be usable without any GitHub functionality
- Tests must use real git repositories (no mocking git internals)
