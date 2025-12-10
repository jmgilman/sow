# Task Log

Worker actions will be logged here.

## 2024-12-09: Migrate CLI to libs/git - COMPLETED

### Summary
Successfully migrated all CLI consumers from `cli/internal/sow` Git/GitHub types to the new centralized `libs/git` module.

### Files Modified

**Core Files:**
- `cli/internal/sow/context.go` - Updated imports and types to use `libs/git.Git` and `libs/git.GitHubClient`
- `cli/go.mod` - Added `github.com/jmgilman/sow/libs/git` dependency with replace directive

**Issue Commands:**
- `cli/cmd/issue/list.go` - Updated to use `git.NewGitHubClient()` and `git.Issue`
- `cli/cmd/issue/show.go` - Updated to use `git.NewGitHubClient()` and `git.Issue`
- `cli/cmd/issue/check.go` - Updated to use `git.NewGitHubClient()`, `git.Issue`, `git.LinkedBranch`

**Project Wizard:**
- `cli/cmd/project/wizard_state.go` - Major updates: `git.Issue`, `git.LinkedBranch`, `git.GitHubClient`, `git.EnsureWorktree`
- `cli/cmd/project/wizard_helpers.go` - Updated to use `git.ErrGHCommand`, `git.WorktreePath`, `git.CheckUncommittedChanges`
- `cli/cmd/project/shared.go` - Updated `initializeProject` signature to use `*git.Issue`

**Test Files:**
- `cli/cmd/project/shared_test.go` - Updated test cases to use `git.Issue`
- `cli/cmd/project/wizard_helpers_test.go` - Updated to use `mocks.GitHubClientMock`, `git.Label`
- `cli/cmd/project/wizard_integration_test.go` - Updated to use `mocks.GitHubClientMock`
- `cli/cmd/project/wizard_state_test.go` - Updated `git.EnsureWorktree` calls to use new 4-parameter signature

### Files Removed
Deleted deprecated Git/GitHub files from `cli/internal/sow/`:
- `git.go`
- `github_cli.go`, `github_cli_test.go`
- `github_client.go`
- `github_factory.go`, `github_factory_test.go`
- `github_mock.go`, `github_mock_test.go`
- `worktree.go`, `worktree_test.go`

### Key Changes
1. **Type Migration:**
   - `sow.Issue` → `git.Issue`
   - `sow.LinkedBranch` → `git.LinkedBranch`
   - `sow.Label` → `git.Label`
   - `sow.GitHubClient` → `git.GitHubClient`
   - `sow.ErrGHCommand` → `git.ErrGHCommand`
   - `sow.ErrGitCommand` → `git.ErrGitCommand`

2. **Function Migration:**
   - `sow.WorktreePath()` → `git.WorktreePath()`
   - `sow.EnsureWorktree(ctx, path, branch)` → `git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), path, branch)`
   - `sow.CheckUncommittedChanges(ctx)` → `git.CheckUncommittedChanges(ctx.Git())`

3. **Mock Migration:**
   - `sow.MockGitHub` → `mocks.GitHubClientMock` (from `libs/git/mocks` package)

### Verification
- Build passes: `go build ./...`
- All tests pass: `go test ./...` (24 packages tested)
