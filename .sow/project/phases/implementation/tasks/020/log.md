# Task Log

## 2025-12-09 - Iteration 1

### Context Loaded
- Read task description.md for requirements
- Read cli/internal/sow/git.go as reference implementation
- Read cli/internal/sow/worktree_test.go for testing patterns
- Read libs/git/types.go, errors.go for existing types
- Read STYLE.md and TESTING.md for standards

### Tests Written (TDD - Red Phase)
Created `libs/git/git_test.go` with comprehensive tests:

**NewGit tests:**
- `TestNewGit_OpensValidRepository` - Opens valid git repo successfully
- `TestNewGit_ReturnsErrNotGitRepository` - Returns ErrNotGitRepository for non-git dirs
- `TestGit_Repository_ReturnsNonNil` - Repository() returns non-nil after open
- `TestGit_RepoRoot_ReturnsPath` - RepoRoot() returns path passed to NewGit

**CurrentBranch tests:**
- `TestGit_CurrentBranch_ReturnsCorrectBranch` - Returns "master" on default init
- `TestGit_CurrentBranch_ReturnsFeatureBranch` - Returns correct name on feature branch

**IsProtectedBranch tests:**
- `TestGit_IsProtectedBranch_ReturnsTrueForMain` - main is protected
- `TestGit_IsProtectedBranch_ReturnsTrueForMaster` - master is protected
- `TestGit_IsProtectedBranch_ReturnsFalseForOtherBranches` - develop, feature/x, etc. not protected

**HasUncommittedChanges tests:**
- `TestGit_HasUncommittedChanges_ReturnsFalseForCleanRepo` - Clean repo returns false
- `TestGit_HasUncommittedChanges_ReturnsTrueForModifiedFiles` - Modified tracked files return true
- `TestGit_HasUncommittedChanges_ReturnsTrueForStagedChanges` - Staged changes return true
- `TestGit_HasUncommittedChanges_ReturnsFalseForOnlyUntrackedFiles` - Untracked files do NOT count as uncommitted changes (key difference from source impl)

**Branches tests:**
- `TestGit_Branches_ReturnsLocalBranches` - Returns list of local branches
- `TestGit_Branches_DoesNotIncludeRemoteBranches` - Does not include remote tracking branches

**CheckoutBranch tests:**
- `TestGit_CheckoutBranch_SuccessfullyCheckoutExistingBranch` - Checks out existing branch
- `TestGit_CheckoutBranch_ReturnsErrorForNonExistentBranch` - Returns error for non-existent branch

### Implementation (Green Phase)
Created `libs/git/git.go` with:
- `Git` struct wrapping `*git.Repository` and `repoRoot` path
- `NewGit(repoRoot string)` - Opens repo, returns ErrNotGitRepository on failure
- `Repository()` - Returns underlying git.Repository
- `RepoRoot()` - Returns repo root path
- `CurrentBranch()` - Returns current branch name
- `IsProtectedBranch(branch string)` - Returns true for "main" or "master"
- `HasUncommittedChanges()` - Returns true for modified/staged/deleted files, **false for untracked**
- `Branches()` - Returns list of local branch names (filters out remotes)
- `CheckoutBranch(branchName string)` - Checks out specified branch

### Key Design Decisions
1. **Untracked files not counted as uncommitted changes** - This differs from the source implementation which used `status.IsClean()`. The task description explicitly requires "Returns false for only untracked files", so I iterate through each file status to exclude untracked files.

2. **ErrNotGitRepository returned on open failure** - Wraps any git.Open error with the typed error for caller convenience.

3. **Uses github.com/jmgilman/go/git wrapper** - Leverages the wrapper's CurrentBranch() and ListBranches() methods where possible, falls back to Underlying() for worktree operations.

### Dependencies Added
Updated `libs/git/go.mod`:
- `github.com/go-git/go-git/v5 v5.16.3`
- `github.com/jmgilman/go/git v0.4.0`

### Tests Run
- All 30 tests pass (including existing tests from task 010)
- Race detection passed: `go test -race ./...`

### Files Modified
- `libs/git/git.go` (new)
- `libs/git/git_test.go` (new)
- `libs/git/go.mod` (updated with dependencies)
- `libs/git/go.sum` (auto-generated)
