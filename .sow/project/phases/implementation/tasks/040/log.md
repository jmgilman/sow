# Task Log

## 2025-12-09 - Consumer Migration Complete

### Summary
Successfully migrated all 8 consumer files from `cli/internal/exec` to `github.com/jmgilman/sow/libs/exec`.

### Changes Made

1. **cli/go.mod**
   - Added `github.com/jmgilman/sow/libs/exec v0.0.0` to require block
   - Added `replace github.com/jmgilman/sow/libs/exec => ../libs/exec` directive

2. **cli/internal/sow/github_cli.go**
   - Changed import from `github.com/jmgilman/sow/cli/internal/exec` to `github.com/jmgilman/sow/libs/exec`
   - No constructor changes needed (uses interface type `exec.Executor`)

3. **cli/internal/sow/github_cli_test.go**
   - Changed import to `github.com/jmgilman/sow/libs/exec/mocks`
   - Changed all `&exec.MockExecutor{}` to `&mocks.ExecutorMock{}`

4. **cli/internal/sow/github_factory.go**
   - Changed import from `github.com/jmgilman/sow/cli/internal/exec` to `github.com/jmgilman/sow/libs/exec`
   - Changed `exec.NewLocal("gh")` to `exec.NewLocalExecutor("gh")`

5. **cli/cmd/project/wizard_state.go**
   - Removed aliased import `sowexec "github.com/jmgilman/sow/cli/internal/exec"`
   - Added `github.com/jmgilman/sow/libs/exec`
   - Changed `sowexec.NewLocal("gh")` to `exec.NewLocalExecutor("gh")`
   - Reorganized import groups per STYLE.md

6. **cli/cmd/project/shared.go**
   - Changed `os/exec` to `osExec "os/exec"` (alias needed to avoid conflict)
   - Removed `sowexec` alias and added `github.com/jmgilman/sow/libs/exec`
   - Changed `sowexec.NewLocal("claude")` to `exec.NewLocalExecutor("claude")`
   - Changed `exec.CommandContext` to `osExec.CommandContext`

7. **cli/cmd/issue/show.go**
   - Changed import from `github.com/jmgilman/sow/cli/internal/exec` to `github.com/jmgilman/sow/libs/exec`
   - Changed `exec.NewLocal("gh")` to `exec.NewLocalExecutor("gh")`
   - Changed `sow.NewGitHub` to `sow.NewGitHubCLI` (uses non-deprecated constructor)

8. **cli/cmd/issue/list.go**
   - Same changes as show.go

9. **cli/cmd/issue/check.go**
   - Same changes as show.go

### Verification
- `go build ./...` - Success (no errors)
- `go test ./...` - All tests pass (24 packages)
- `golangci-lint run` - 0 issues

### Notes
- Import groups reorganized per STYLE.md: stdlib, external, internal
- Used `NewGitHubCLI` instead of deprecated `NewGitHub` in issue commands
