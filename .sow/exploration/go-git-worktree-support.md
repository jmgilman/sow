# Go-git Worktree Support Analysis

## Summary

**go-git v5 does not support git worktree operations** (add/list/remove). Both the upstream go-git library and the jmgilman/go/git wrapper lack native support for creating linked worktrees. Sow must shell out to `git` CLI for worktree management operations.

## Detailed Findings

### Upstream go-git (github.com/go-git/go-git/v5)

**What exists:**
- `Repository.Worktree()` - Returns the single main worktree for a repository
- `Worktree` type - Represents a working tree with operations like Checkout, Add, Commit, Status
- `PlainOpenWithOptions` with `EnableDotGitCommonDir: true` - Can **read** existing linked worktrees created externally

**What does NOT exist:**
- No `Repository.CreateWorktree()` or equivalent for creating linked worktrees
- No `Repository.ListWorktrees()` for enumerating worktrees
- No `Worktree.Remove()` for removing linked worktrees
- No support for `git worktree add`, `git worktree list`, `git worktree remove`

**Status:**
- Feature request exists (Issue #396, labeled "enhancement", opened Jan 2024)
- Marked as "not planned" by maintainers
- A third-party fork (github.com/cooper/go-git) claims to add linked worktree support, but details unclear

**Workaround discussed:**
- Issue #285 proposed using separate `filesystem.Storage` with `memory.IndexStorage` per goroutine
- This allows concurrent checkouts but NOT true parallel worktrees
- All checkouts still share the same working directory

### jmgilman/go/git Wrapper

**Current version:** v0.3.1 (used by sow)

**Worktree-related API:**
```go
// CreateWorktree creates a new worktree at the specified path
func (r *Repository) CreateWorktree(path string, opts WorktreeOptions) (*Worktree, error)

// ListWorktrees returns all worktrees associated with this repository
func (r *Repository) ListWorktrees() ([]*Worktree, error)

// Worktree type with methods
func (w *Worktree) Checkout(ref string) error
func (w *Worktree) Path() string
func (w *Worktree) Remove() error
func (w *Worktree) Underlying() *gogit.Worktree
```

**Actual behavior:**

From the function documentation:

> "Note: go-git v5 does not support linked worktrees (git worktree add). This method returns the main worktree and checks it out to the specified reference. For true parallel worktrees, consider using separate clones."

**Translation:**
- `CreateWorktree()` does NOT create separate worktrees
- It just returns the main worktree and checks out a different branch in place
- `ListWorktrees()` exists but only returns the main worktree (verified: "This method returns only the main worktree")
- `Remove()` likely doesn't remove linked worktrees (since they aren't created)
- These are convenience wrappers with misleading names, not true worktree support

### Reading Existing Worktrees

While go-git cannot **create** worktrees, it CAN work with worktrees created by the git CLI.

**How:**
1. Git CLI creates worktree: `git worktree add .sow/worktrees/feat/auth feat/auth`
2. This creates `.sow/worktrees/feat/auth/.git` (file, not directory)
3. The `.git` file contains: `gitdir: /path/to/main/.git/worktrees/feat/auth`
4. go-git can open this:
   ```go
   repo, err := git.PlainOpenWithOptions(
       ".sow/worktrees/feat/auth",
       &git.PlainOpenOptions{EnableDotGitCommonDir: true},
   )
   ```
5. Repo is fully functional for reading/writing git operations

**Key insight:** Once worktree exists, go-git works fine. Problem is only creation/listing/removal.

## Implementation Strategy

### Git Worktree Operations (Wrapper handles CLI)

The `github.com/jmgilman/go/git` wrapper will be enhanced to shell out to git CLI for:

**1. Creating worktrees:**
```bash
git worktree add <path> <branch>
# or for new branch:
git worktree add -b <branch> <path> <start-point>
```

**2. Listing worktrees:**
```bash
git worktree list --porcelain
# Output format:
# worktree /path/to/main/repo
# HEAD abc123...
# branch refs/heads/main
#
# worktree /path/to/worktree
# HEAD def456...
# branch refs/heads/feat/auth
```

**3. Removing worktrees:**
```bash
git worktree remove <path>
# or force:
git worktree remove --force <path>
```

**4. Pruning orphaned worktrees:**
```bash
git worktree prune
```

### Git Operations Within Worktrees (go-git works)

Once worktree exists, sow can use jmgilman/go/git wrapper for:
- Opening repository: `git.Open(worktreePath)` with `EnableDotGitCommonDir`
- All normal git operations: status, commit, push, pull, etc.
- Branch operations: create, checkout, delete
- Reading git state

### Proposed Implementation

**Enhancement to:** `github.com/jmgilman/go/git`

The wrapper will be updated to actually create/list/remove worktrees:

```go
package git

import (
    "os/exec"
    "strings"
)

// CreateWorktree creates a new worktree for the given branch
// (Updated to actually shell out to git CLI)
func (r *Repository) CreateWorktree(path string, opts WorktreeOptions) (*Worktree, error) {
    // Implementation uses git CLI internally
    // Details: check if branch exists, run appropriate git worktree add command
    // Returns Worktree wrapper around created worktree
}

// ListWorktrees returns all worktrees (updated to actually list)
func (r *Repository) ListWorktrees() ([]*Worktree, error) {
    // git worktree list --porcelain
    // Parse output and return Worktree wrappers
}

// Remove removes a worktree (updated implementation)
func (w *Worktree) Remove() error {
    // git worktree remove <path>
}

// Additional method for pruning
func (r *Repository) PruneWorktrees() error {
    // git worktree prune
}
```

### Usage in Sow Commands

From sow's perspective, the wrapper handles all the complexity:

```go
func runExplore(cmd *cobra.Command, branchName, initialPrompt string) error {
    mainCtx := cmdutil.GetContext(cmd.Context())

    // Determine target branch
    branch := determineBranch(mainCtx, branchName)

    // Create worktree using wrapper API (wrapper shells out internally)
    worktreePath := filepath.Join(mainCtx.RepoRoot(), ".sow/worktrees", branch)

    worktree, err := mainCtx.Git().Repository().CreateWorktree(
        worktreePath,
        git.WorktreeOptions{Branch: plumbing.NewBranchReferenceName(branch)},
    )
    if err != nil {
        return fmt.Errorf("failed to create worktree: %w", err)
    }

    // Create context for worktree
    worktreeCtx, err := sow.NewContext(worktree.Path())
    if err != nil {
        return err
    }

    // Continue with mode operations using worktreeCtx
    // ...
}
```

## Why Shelling Out is Acceptable

**Precedent:** Sow already shells out for:
- `claude` CLI (sow start)
- `gh` CLI (GitHub operations)

**Benefits:**
- git CLI is universally available (required for git repos)
- Battle-tested implementation
- Handles edge cases (permissions, locking, etc.)
- Forward-compatible with future git worktree features

**Downsides:**
- Cross-platform command execution (but git is cross-platform)
- Parsing CLI output (but --porcelain provides stable format)
- Performance overhead (negligible for worktree operations)

**Conclusion:** Shelling out is the pragmatic choice given go-git limitations.

## Alternative: Third-Party Packages

### go-git-cmd-wrapper (github.com/ldez/go-git-cmd-wrapper)

A wrapper around git CLI commands:
```go
import "github.com/ldez/go-git-cmd-wrapper/worktree"

output, err := git.Worktree(worktree.Add, worktree.Branch("new"), worktree.Path("path"))
```

**Pros:**
- Higher-level API than raw exec
- Handles argument escaping

**Cons:**
- Additional dependency
- Still shells out underneath
- May not handle all edge cases we need
- Just a thin wrapper, not much value over exec.Command

**Recommendation:** Direct `exec.Command` is simpler and gives more control.

### cooper/go-git Fork

Claims to add linked worktree support.

**Cons:**
- Unmaintained fork (last commit unclear)
- Diverges from upstream go-git
- Unknown quality/completeness
- Would replace entire go-git dependency

**Recommendation:** Not worth the risk. Stick with upstream go-git + CLI.

## Testing Considerations

### Unit Tests

Mock git CLI operations:
```go
// Use interface for testability
type WorktreeManager interface {
    Create(branch, path string) error
    List() ([]Info, error)
    Remove(path string) error
}

// Real implementation shells out
type CLIWorktreeManager struct { ... }

// Test implementation uses fake
type FakeWorktreeManager struct {
    worktrees map[string]Info
}
```

### Integration Tests

Test against real git CLI:
- Requires git available in PATH
- Create temporary test repos
- Verify worktree creation/listing/removal
- Clean up in teardown

## Open Questions

1. **Error handling:** How to handle git CLI errors gracefully?
   - Parse stderr for specific errors (branch exists, path in use, etc.)
   - Provide helpful error messages to user

2. **Concurrent operations:** Can multiple `git worktree add` commands run safely?
   - Git handles locking at filesystem level
   - Should be safe, but test concurrent creation

3. **Path handling:** How to handle worktree paths with spaces or special characters?
   - Use proper quoting in exec.Command
   - Test with problematic paths

4. **Git version compatibility:** What's minimum git version required?
   - `git worktree` introduced in Git 2.5 (2015)
   - Reasonable to require Git 2.5+
   - Check version on first use? Or document requirement?

5. **Worktree path validation:** Should we validate paths before creating?
   - Check parent directory exists
   - Check path doesn't already exist
   - Check sufficient disk space?

## Recommendation

**Immediate action:** Enhance `github.com/jmgilman/go/git` wrapper to properly implement worktree operations by shelling out to git CLI.

**Sow integration:** Use the enhanced wrapper API directly - sow doesn't need to worry about CLI details.

**Future:** Monitor go-git for native worktree support. If/when added, wrapper can migrate internal implementation without changing its public API or affecting sow.

**References:**
- go-git Issue #396: https://github.com/go-git/go-git/issues/396
- go-git Issue #285 (concurrent worktrees): https://github.com/go-git/go-git/issues/285
- git-worktree documentation: https://git-scm.com/docs/git-worktree
