# Task 060: Update Consumer Files to Use libs/git

## Context

This task is part of work unit 004: creating a new `libs/git` Go module. All the core functionality has been implemented in tasks 010-050. This task updates all consumer files in the CLI to use the new `libs/git` module instead of `cli/internal/sow/`.

**Critical Changes**:
1. Import path changes from `github.com/jmgilman/sow/cli/internal/sow` to `github.com/jmgilman/sow/libs/git`
2. Some function signatures changed (e.g., `EnsureWorktree` now takes `*git.Git` instead of `*sow.Context`)
3. Type names remain the same (Issue, LinkedBranch, etc.) but are in the `git` package

## Requirements

### 1. Update cli/internal/sow/context.go

The Context struct needs to use the new `libs/git` types:

```go
import (
    "github.com/jmgilman/sow/libs/git"
)

type Context struct {
    fs       FS
    repo     *git.Git      // Changed from *Git
    github   *git.GitHubCLI // Changed from *GitHubCLI
    repoRoot string
    // ... worktree fields
}

// Git returns the Git repository wrapper.
func (c *Context) Git() *git.Git {
    return c.repo
}

// GitHub returns the GitHub client.
func (c *Context) GitHub() *git.GitHubCLI {
    // ... lazy initialization
}
```

The `NewContext` function should use `git.NewGit()` instead of the local implementation.

### 2. Update Worktree Callers

Files that call `EnsureWorktree` need signature updates:

**Before**:
```go
err := sow.EnsureWorktree(ctx, path, branch)
```

**After**:
```go
import "github.com/jmgilman/sow/libs/git"

err := git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), path, branch)
```

Files to update:
- `cli/cmd/project/wizard_helpers.go`

### 3. Update GitHub Client Usage

Files that use GitHub types need import updates:

**Before**:
```go
import "github.com/jmgilman/sow/cli/internal/sow"

var issues []sow.Issue
gh := sow.NewGitHubCLI(exec)
```

**After**:
```go
import "github.com/jmgilman/sow/libs/git"

var issues []git.Issue
gh := git.NewGitHubCLI(exec)
```

Files to update:
- `cli/cmd/issue/list.go`
- `cli/cmd/issue/show.go`
- `cli/cmd/issue/check.go`
- `cli/cmd/project/wizard_state.go`
- `cli/cmd/project/wizard_helpers.go`

### 4. Update Test Files

Test files need similar import and type updates:

Files to update:
- `cli/cmd/project/wizard_state_test.go`
- `cli/cmd/project/wizard_helpers_test.go`
- `cli/cmd/worktree_test.go`
- Any other test files using git/github types

### 5. Update MockGitHub References

The `MockGitHub` type has moved to `libs/git/mocks/GitHubClientMock`:

**Before**:
```go
mock := &sow.MockGitHub{...}
```

**After**:
```go
import "github.com/jmgilman/sow/libs/git/mocks"

mock := &mocks.GitHubClientMock{...}
```

Or use the manually-written mock if preferred for simpler defaults.

### 6. Consumer Files List

Based on grep analysis, these files need updates:

**cli/internal/sow/** (internal restructuring):
- `context.go` - Use libs/git types
- Remove: `git.go`, `github_*.go`, `worktree.go`, `types.go` (moved to libs/git)

**cli/cmd/issue/**:
- `list.go` - Import git, use git.Issue, git.NewGitHubCLI
- `show.go` - Import git, use git types
- `check.go` - Import git, use git types

**cli/cmd/project/**:
- `wizard_state.go` - Import git, use git.Issue
- `wizard_helpers.go` - Import git, use git.EnsureWorktree with new signature

**cli/cmd/**:
- `worktree.go` - Uses ctx.Git().Repository() - may work unchanged
- `worktree_test.go` - Update if needed

**Test files**:
- `cli/cmd/project/wizard_state_test.go`
- `cli/cmd/project/wizard_helpers_test.go`
- `cli/cmd/worktree_test.go`
- `cli/internal/sow/github_cli_test.go` - REMOVE (tests moved to libs/git)
- `cli/internal/sow/worktree_test.go` - REMOVE (tests moved to libs/git)

## Acceptance Criteria

1. **All imports updated**: No references to old git/github code in cli/internal/sow
2. **Context uses libs/git**: Context.Git() returns *git.Git
3. **Context uses libs/git**: Context.GitHub() returns *git.GitHubCLI
4. **EnsureWorktree calls updated**: New signature used everywhere
5. **GitHub types imported correctly**: git.Issue, git.LinkedBranch used
6. **Mock usage updated**: Either mocks.GitHubClientMock or manual mock
7. **All tests pass**: `go test ./cli/...` passes
8. **Build succeeds**: `go build ./cli/...` succeeds
9. **No unused imports**: All imports are used
10. **Old code removed**: git.go, github_*.go, worktree.go removed from cli/internal/sow

### Test Requirements

After updates, verify:
- `cd cli && go build ./...` succeeds
- `cd cli && go test ./...` passes
- No import cycles between cli and libs/git

## Technical Details

### Import Alias Pattern

When there's a naming conflict with the standard `context` package:

```go
import (
    "context"

    gitpkg "github.com/jmgilman/sow/libs/git"
)
```

Or use the natural package name when no conflict:
```go
import "github.com/jmgilman/sow/libs/git"

var issues []git.Issue
```

### Context Restructuring

The `cli/internal/sow/context.go` should:
1. Import `libs/git`
2. Store `*git.Git` and `*git.GitHubCLI` as fields
3. Initialize them using `git.NewGit()` and `git.NewGitHubCLI()`
4. Expose them via accessor methods

### Files to Remove from cli/internal/sow/

After migration, these files should be DELETED (their functionality is now in libs/git):
- `git.go` - Moved to libs/git/git.go
- `github_client.go` - Moved to libs/git/client.go
- `github_cli.go` - Moved to libs/git/client_cli.go
- `github_factory.go` - Moved to libs/git/factory.go
- `github_mock.go` - Replaced by libs/git/mocks/
- `worktree.go` - Moved to libs/git/worktree.go
- All their corresponding `*_test.go` files

### Files to Keep in cli/internal/sow/

These files should remain (with possible modifications):
- `context.go` - Updated to use libs/git
- `context_test.go` - Updated as needed
- `fs.go` - Unrelated to git
- `sow.go` - Unrelated to git
- `errors.go` - Unrelated to git (ErrNotInitialized, etc.)
- `config.go` - Unrelated to git
- `user_config.go` - Unrelated to git

## Relevant Inputs

- `cli/internal/sow/context.go` - Primary file to update
- `cli/cmd/issue/*.go` - Consumer files
- `cli/cmd/project/wizard_*.go` - Consumer files
- `cli/cmd/worktree.go` - Consumer file
- `libs/git/*.go` - New module (from tasks 010-050)
- `.standards/STYLE.md` - Coding standards to follow
- `.standards/TESTING.md` - Testing standards to follow

## Examples

### Updated context.go

```go
package sow

import (
    "github.com/jmgilman/sow/libs/exec"
    "github.com/jmgilman/sow/libs/git"
)

type Context struct {
    fs       FS
    repo     *git.Git
    github   *git.GitHubCLI
    repoRoot string
    // ...
}

func NewContext(repoRoot string) (*Context, error) {
    // ...
    gitRepo, err := git.NewGit(repoRoot)
    if err != nil {
        return nil, fmt.Errorf("failed to open git repository: %w", err)
    }

    return &Context{
        fs:       sowFS,
        repo:     gitRepo,
        repoRoot: repoRoot,
    }, nil
}

func (c *Context) Git() *git.Git {
    return c.repo
}

func (c *Context) GitHub() *git.GitHubCLI {
    if c.github == nil {
        ghExec := exec.NewLocalExecutor("gh")
        c.github = git.NewGitHubCLI(ghExec)
    }
    return c.github
}
```

### Updated issue/list.go

```go
package issue

import (
    "github.com/jmgilman/sow/libs/exec"
    "github.com/jmgilman/sow/libs/git"
)

func newListCmd() *cobra.Command {
    // ...
    RunE: func(cmd *cobra.Command, _ []string) error {
        ghExec := exec.NewLocalExecutor("gh")
        gh := git.NewGitHubCLI(ghExec)

        issues, err := gh.ListIssues("sow", state)
        // ...
    }
}

func printIssuesTable(cmd *cobra.Command, issues []git.Issue) {
    // ...
}
```

### Updated wizard_helpers.go

```go
package project

import (
    "github.com/jmgilman/sow/libs/git"
)

func createWorktree(ctx *sow.Context, branch string) (string, error) {
    path := git.WorktreePath(ctx.RepoRoot(), branch)

    // Check for uncommitted changes
    if err := git.CheckUncommittedChanges(ctx.Git()); err != nil {
        return "", err
    }

    // Create worktree with new signature
    if err := git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), path, branch); err != nil {
        return "", err
    }

    return path, nil
}
```

## Dependencies

- Tasks 010-050 must complete first - libs/git must be fully implemented
- The libs/git module must be importable from cli

## Constraints

- Do NOT break existing functionality
- Do NOT change the public API of Context (method signatures should remain compatible)
- Do NOT leave orphaned/unused code in cli/internal/sow
- All tests must pass after migration
- Follow existing code style and patterns
