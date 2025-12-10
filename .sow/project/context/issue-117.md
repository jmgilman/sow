# Issue #117: libs/git module

**URL**: https://github.com/jmgilman/sow/issues/117
**State**: OPEN

## Description

# Work Unit 004: libs/git Module

Create a new `libs/git` Go module providing Git and GitHub operations using the ports and adapters pattern.

## Objective

Move Git/GitHub code from `cli/internal/sow/` to `libs/git/`, creating a standalone Go module with clean interfaces (ports) and implementations (adapters), decoupled from `sow.Context`, with full standards compliance and proper documentation.

**This is NOT just a file move.** The key changes are:
1. Decouple from `sow.Context` - functions accept explicit parameters
2. Design clean interfaces for Git and GitHub operations
3. Separate interface definitions from implementations

## Scope

### What's Moving
- `git.go` - Git struct and operations
- `github_client.go` - GitHubClient interface
- `github_cli.go` - GitHubCLI implementation (uses `gh` CLI)
- `github_factory.go` - NewGitHubClient factory
- `github_mock.go` - Mock for testing
- `worktree.go` - EnsureWorktree and related functions
- `types.go` - Issue, LinkedBranch types

### Target Structure
```
libs/git/
├── go.mod
├── go.sum
├── README.md                 # Per READMES.md standard
├── doc.go                    # Package documentation
│
├── git.go                    # Git struct and operations
├── git_test.go               # Tests for Git operations
├── client.go                 # GitHubClient interface (port)
├── client_cli.go             # GitHubCLI implementation (adapter)
├── client_cli_test.go        # Tests for CLI implementation
├── factory.go                # NewGitHubClient factory
├── worktree.go               # EnsureWorktree(git, repoRoot, path, branch)
├── worktree_test.go          # Tests for worktree operations
├── types.go                  # Issue, LinkedBranch, etc.
├── errors.go                 # ErrGHNotInstalled, etc.
│
└── mocks/
    └── client.go             # Generated mock via moq
```

## Standards Requirements

### Go Code Standards (STYLE.md)
- Accept interfaces, return concrete types
- Error handling with proper wrapping (`%w`)
- No global mutable state
- Define interfaces in consumer packages where possible
- Short receiver names (1-2 chars)
- Functions under 80 lines

### Testing Standards (TESTING.md)
- Behavioral test coverage for all Git operations
- Table-driven tests with `t.Run()`
- Use `testify/assert` and `testify/require`
- Mock generation via `moq` for `GitHubClient` interface
- No external dependencies in unit tests (mock `Executor`)

### README Standards (READMES.md)
- Overview: Git and GitHub operations for sow
- Quick Start: Create Git instance, basic operations
- Usage: Branch operations, GitHub integration, worktrees
- Testing: How to use mocks

### Linting
- Must pass `golangci-lint run` with project's `.golangci.yml`
- Proper error wrapping for external errors

## API Design Requirements

### Ports and Adapters Pattern

**Git Operations:**
```go
// Git provides git repository operations.
type Git struct {
    repoRoot string
    exec     exec.Executor
}

func NewGit(repoRoot string, exec exec.Executor) *Git

func (g *Git) CurrentBranch(ctx context.Context) (string, error)
func (g *Git) Checkout(ctx context.Context, branch string) error
func (g *Git) CreateBranch(ctx context.Context, branch string) error
// ... other operations
```

**GitHub Client (Port):**
```go
// GitHubClient defines the interface for GitHub API operations.
type GitHubClient interface {
    GetIssue(ctx context.Context, owner, repo string, number int) (*Issue, error)
    CreateIssue(ctx context.Context, owner, repo string, issue *Issue) (*Issue, error)
    ListLinkedBranches(ctx context.Context, owner, repo string, number int) ([]LinkedBranch, error)
    // ... other operations
}
```

**GitHub CLI Adapter:**
```go
// GitHubCLI implements GitHubClient using the gh CLI tool.
type GitHubCLI struct {
    exec exec.Executor
}

func NewGitHubCLI(exec exec.Executor) *GitHubCLI
```

### Decoupled from Context

**Before:**
```go
func EnsureWorktree(ctx *Context, path, branch string) error
```

**After:**
```go
func EnsureWorktree(git *Git, repoRoot, path, branch string) error
```

### Error Handling
- Define sentinel errors: `ErrGHNotInstalled`, `ErrNotGitRepository`, `ErrBranchExists`
- Wrap all errors with context

## Consumer Impact

~20+ files currently use Git/GitHub operations. After this work:
- All imports change to `github.com/jmgilman/sow/libs/git`
- Callers create `Git` and `GitHubClient` instances explicitly
- Worktree functions accept `*Git` instead of `*Context`

## Dependencies

- `libs/exec` - For command execution

## Acceptance Criteria

1. [ ] New `libs/git` Go module exists and compiles
2. [ ] Clean `GitHubClient` interface designed
3. [ ] `Git` struct decoupled from `sow.Context`
4. [ ] Worktree functions accept explicit parameters
5. [ ] Mock generated via `moq` in `mocks/` subpackage
6. [ ] Proper sentinel errors defined
7. [ ] All tests pass with proper behavioral coverage
8. [ ] `golangci-lint run` passes with no issues
9. [ ] README.md follows READMES.md standard
10. [ ] Package documentation in doc.go
11. [ ] All 20+ consumer files updated to new imports and signatures
12. [ ] Old git code removed from `cli/internal/sow/`

## Out of Scope

- Adding new Git operations
- GitHub API client (only `gh` CLI wrapper)
- Git credential management
- SSH key handling
