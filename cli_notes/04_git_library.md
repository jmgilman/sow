# Git Library Reference

**Package**: `github.com/jmgilman/go/git`

**Purpose**: Clean, type-safe wrapper around go-git for Git repository operations.

---

## Overview

This library provides a simplified interface to go-git, focusing on:
- Worker Service caching workflows (clone → update → worktree creation)
- Bootstrap CLI and discovery needs
- Thin wrappers over go-git (not reimplementing Git)
- Billy filesystem for all I/O operations
- Escape hatches for advanced use cases

**Key Principles**:
1. **Thin wrappers** - Not reimplementing Git, just simplifying go-git
2. **Billy filesystem** - All I/O uses billy abstraction
3. **Escape hatches** - `Underlying()` methods provide access to raw go-git
4. **RemoteOperations interface** - Enables testing by mocking network operations

---

## Core Types

### Repository

Main type wrapping go-git repository.

```go
type Repository struct {
    // Internal fields
}

// Factory functions
func Init(path string, opts ...RepositoryOption) (*Repository, error)
func Clone(ctx context.Context, url string, opts ...RepositoryOption) (*Repository, error)
func Open(path string, opts ...RepositoryOption) (*Repository, error)

// Escape hatch
func (r *Repository) Underlying() *gogit.Repository
func (r *Repository) Filesystem() billy.Filesystem
```

### Other Value Types

```go
type Commit struct {
    Hash      string
    Author    string
    Email     string
    Message   string
    Timestamp time.Time
}

type Branch struct {
    Name      string
    Hash      string
    IsRemote  bool
}

type Tag struct {
    Name    string
    Hash    string
    Message string
}

type Remote struct {
    Name string
    URLs []string
}

type Worktree struct {
    path string
    wt   *gogit.Worktree
}
```

---

## Repository Options

Configuration options for repository operations:

```go
type RepositoryOption func(*repositoryConfig)

// Filesystem
func WithFilesystem(fs billy.Filesystem) RepositoryOption

// Authentication
func WithAuth(auth Auth) RepositoryOption

// Clone options
func WithBranch(branch string) RepositoryOption
func WithDepth(depth int) RepositoryOption
func WithSingleBranch() RepositoryOption

// Advanced
func WithRemoteOperations(ops RemoteOperations) RepositoryOption
```

---

## Authentication

Helper functions for creating authentication:

```go
type Auth interface {
    // go-git auth method
}

// SSH key from file
func SSHKeyFile(user, pemFile, password string) (Auth, error)

// SSH key from memory
func SSHKeyAuth(user string, pemBytes []byte, password string) (Auth, error)

// Basic auth (HTTPS)
func BasicAuth(username, password string) Auth

// No authentication (public repos)
func EmptyAuth() Auth
```

**Usage**:
```go
// SSH key
auth, err := git.SSHKeyFile("git", "/home/user/.ssh/id_rsa", "")

// Basic auth (GitHub token)
auth := git.BasicAuth("username", "ghp_token123")

// Use with operations
repo, err := git.Clone(ctx, "https://github.com/org/repo", git.WithAuth(auth))
```

---

## Core Operations

### Clone Repository

```go
func Clone(ctx context.Context, url string, opts ...RepositoryOption) (*Repository, error)
```

**Usage**:
```go
ctx := context.Background()

// Basic clone
repo, err := git.Clone(ctx, "https://github.com/org/repo")

// Shallow clone (faster, less disk)
repo, err := git.Clone(ctx, url,
    git.WithDepth(1),
    git.WithSingleBranch())

// With authentication
auth := git.BasicAuth("user", "token")
repo, err := git.Clone(ctx, url, git.WithAuth(auth))

// Custom filesystem location
fs := billy.NewLocal()
chrootFS, _ := fs.Chroot("/tmp/repos")
repo, err := git.Clone(ctx, url, git.WithFilesystem(chrootFS.Unwrap()))
```

### Open Existing Repository

```go
func Open(path string, opts ...RepositoryOption) (*Repository, error)
```

**Usage**:
```go
repo, err := git.Open("/path/to/repo")
if err != nil {
    return err
}

// With custom filesystem
fs := billy.NewLocal()
repo, err := git.Open("/path/to/repo", git.WithFilesystem(fs.Unwrap()))
```

### Initialize New Repository

```go
func Init(path string, opts ...RepositoryOption) (*Repository, error)
```

**Usage**:
```go
repo, err := git.Init("/path/to/new/repo")

// Create initial commit
hash, err := repo.CreateCommit(git.CommitOptions{
    Author:     "User",
    Email:      "user@example.com",
    Message:    "Initial commit",
    AllowEmpty: true,
})
```

### Fetch Updates

```go
type FetchOptions struct {
    RemoteName string
    Auth       Auth
    Force      bool
}

func (r *Repository) Fetch(ctx context.Context, opts FetchOptions) error
```

**Usage**:
```go
err := repo.Fetch(ctx, git.FetchOptions{
    RemoteName: "origin",
    Auth:       auth,
})
```

### Pull Changes

```go
type PullOptions struct {
    RemoteName string
    Branch     string
    Auth       Auth
}

func (r *Repository) Pull(ctx context.Context, opts PullOptions) error
```

**Usage**:
```go
err := repo.Pull(ctx, git.PullOptions{
    RemoteName: "origin",
    Branch:     "main",
    Auth:       auth,
})
```

### Push Changes

```go
type PushOptions struct {
    RemoteName string
    RefSpecs   []string
    Auth       Auth
    Force      bool
}

func (r *Repository) Push(ctx context.Context, opts PushOptions) error
```

**Usage**:
```go
err := repo.Push(ctx, git.PushOptions{
    RemoteName: "origin",
    RefSpecs:   []string{"refs/heads/main"},
    Auth:       auth,
})
```

---

## Commit Operations

### Get Commit

```go
func (r *Repository) GetCommit(hash string) (Commit, error)
func (r *Repository) GetHEAD() (Commit, error)
```

### Create Commit

```go
type CommitOptions struct {
    Author     string
    Email      string
    Message    string
    AllowEmpty bool
}

func (r *Repository) CreateCommit(opts CommitOptions) (string, error)
```

**Usage**:
```go
// Modify files first using billy filesystem
fs := repo.Filesystem()
file, _ := fs.Create("new-file.txt")
file.Write([]byte("content"))
file.Close()

// Stage changes
wt, _ := repo.Underlying().Worktree()
wt.Add("new-file.txt")

// Create commit
hash, err := repo.CreateCommit(git.CommitOptions{
    Author:  "Bot",
    Email:   "bot@example.com",
    Message: "Update file",
})
```

### Walk Commits

```go
func (r *Repository) WalkCommits(from, to string) iter.Seq2[Commit, error]
```

**Usage**:
```go
for commit, err := range repo.WalkCommits("v1.0.0", "v2.0.0") {
    if err != nil {
        return err
    }
    fmt.Printf("%s: %s\n", commit.Hash[:7], commit.Message)
}
```

---

## Branch Operations

### List Branches

```go
func (r *Repository) ListBranches() ([]Branch, error)
func (r *Repository) ListRemoteBranches() ([]Branch, error)
```

### Get/Create/Delete Branch

```go
func (r *Repository) GetBranch(name string) (Branch, error)
func (r *Repository) GetCurrentBranch() (Branch, error)
func (r *Repository) CreateBranch(name, startPoint string) error
func (r *Repository) DeleteBranch(name string) error
func (r *Repository) CheckoutBranch(name string) error
```

---

## Tag Operations

```go
func (r *Repository) ListTags() ([]Tag, error)
func (r *Repository) GetTag(name string) (Tag, error)
func (r *Repository) CreateTag(name, hash, message string) error
func (r *Repository) DeleteTag(name string) error
```

---

## Worktree Operations

**Note**: go-git v5 does not support linked worktrees (`git worktree add`). The `CreateWorktree` method returns the main worktree checked out to the specified reference.

```go
type WorktreeOptions struct {
    Hash   plumbing.Hash
    Branch string
}

func (r *Repository) CreateWorktree(path string, opts WorktreeOptions) (*Worktree, error)
func (wt *Worktree) Path() string
func (wt *Worktree) Remove() error
func (wt *Worktree) Underlying() *gogit.Worktree
```

**Usage**:
```go
// Create worktree at specific commit
wt, err := repo.CreateWorktree("/tmp/worktree", git.WorktreeOptions{
    Hash: plumbing.NewHash("abc123..."),
})
defer wt.Remove()

// Work with files in worktree
fmt.Println("Worktree at:", wt.Path())
```

---

## Remote Operations

```go
func (r *Repository) ListRemotes() ([]Remote, error)
func (r *Repository) GetRemote(name string) (Remote, error)
func (r *Repository) AddRemote(name, url string) error
func (r *Repository) RemoveRemote(name string) error
```

---

## Application in sow CLI

### Use Case 1: Clone External Refs

**File**: `internal/refs/clone.go`

```go
package refs

import (
    "context"
    "path/filepath"
    "github.com/jmgilman/go/fs/billy"
    "github.com/jmgilman/go/git"
)

type RefCloner struct {
    cacheDir string
    fs       *billy.LocalFS
}

func NewRefCloner(cacheDir string) *RefCloner {
    return &RefCloner{
        cacheDir: cacheDir,
        fs:       billy.NewLocal(),
    }
}

func (rc *RefCloner) Clone(ctx context.Context, url, branch string, auth git.Auth) (string, error) {
    // Determine cache path
    cachePath := filepath.Join(rc.cacheDir, "repos", hashURL(url))

    // Check if already exists
    exists, _ := rc.fs.Exists(cachePath)
    if exists {
        return cachePath, nil
    }

    // Create parent directory
    rc.fs.MkdirAll(filepath.Dir(cachePath), 0755)

    // Chroot filesystem to cache path
    cacheFS, err := rc.fs.Chroot(cachePath)
    if err != nil {
        return "", err
    }

    // Clone repository
    _, err = git.Clone(ctx, url,
        git.WithFilesystem(cacheFS.Unwrap()),
        git.WithBranch(branch),
        git.WithDepth(1),
        git.WithSingleBranch(),
        git.WithAuth(auth))

    if err != nil {
        return "", err
    }

    return cachePath, nil
}
```

### Use Case 2: Update Cached Refs

**File**: `internal/refs/update.go`

```go
package refs

import (
    "context"
    "github.com/jmgilman/go/git"
)

func (rc *RefCloner) Update(ctx context.Context, cachePath string, auth git.Auth) error {
    // Open existing repository
    repo, err := git.Open(cachePath)
    if err != nil {
        return err
    }

    // Fetch updates from remote
    return repo.Fetch(ctx, git.FetchOptions{
        RemoteName: "origin",
        Auth:       auth,
    })
}

func (rc *RefCloner) GetStatus(ctx context.Context, cachePath string) (string, int, error) {
    repo, err := git.Open(cachePath)
    if err != nil {
        return "", 0, err
    }

    // Get current HEAD
    head, err := repo.GetHEAD()
    if err != nil {
        return "", 0, err
    }

    // Fetch to get remote status (without merging)
    repo.Fetch(ctx, git.FetchOptions{RemoteName: "origin"})

    // Get remote branch
    branches, _ := repo.ListRemoteBranches()
    if len(branches) == 0 {
        return "current", 0, nil
    }

    // Compare commits (simplified - would walk commits to count)
    // For now just check if HEADs match
    if head.Hash == branches[0].Hash {
        return "current", 0, nil
    }

    return "behind", 1, nil  // Simplified: would count actual commits
}
```

### Use Case 3: Get Remote Repository Info

**File**: `internal/refs/inspect.go`

```go
package refs

import (
    "context"
    "github.com/jmgilman/go/git"
)

type RefInfo struct {
    URL         string
    Branch      string
    CommitHash  string
    CommitMsg   string
    LastUpdated time.Time
}

func (rc *RefCloner) Inspect(ctx context.Context, cachePath string) (*RefInfo, error) {
    repo, err := git.Open(cachePath)
    if err != nil {
        return nil, err
    }

    // Get current commit
    head, err := repo.GetHEAD()
    if err != nil {
        return nil, err
    }

    // Get remote URL
    remotes, err := repo.ListRemotes()
    if err != nil || len(remotes) == 0 {
        return nil, err
    }

    // Get current branch
    branch, err := repo.GetCurrentBranch()
    if err != nil {
        return nil, err
    }

    return &RefInfo{
        URL:         remotes[0].URLs[0],
        Branch:      branch.Name,
        CommitHash:  head.Hash,
        CommitMsg:   head.Message,
        LastUpdated: head.Timestamp,
    }, nil
}
```

### Use Case 4: List Commits Between Tags

**File**: `internal/refs/changelog.go`

```go
package refs

import (
    "fmt"
    "github.com/jmgilman/go/git"
)

func (rc *RefCloner) GetChangelog(cachePath, fromTag, toTag string) ([]string, error) {
    repo, err := git.Open(cachePath)
    if err != nil {
        return nil, err
    }

    // Get tag hashes
    from, err := repo.GetTag(fromTag)
    if err != nil {
        return nil, err
    }

    to, err := repo.GetTag(toTag)
    if err != nil {
        return nil, err
    }

    // Walk commits between tags
    var changes []string
    for commit, err := range repo.WalkCommits(from.Hash, to.Hash) {
        if err != nil {
            return nil, err
        }

        changes = append(changes, fmt.Sprintf("%s: %s (%s)",
            commit.Hash[:7],
            commit.Message,
            commit.Author))
    }

    return changes, nil
}
```

### Use Case 5: Authentication Management

**File**: `internal/refs/auth.go`

```go
package refs

import (
    "os"
    "path/filepath"
    "github.com/jmgilman/go/git"
)

func GetAuthForURL(url string) (git.Auth, error) {
    // Check for SSH URLs
    if strings.HasPrefix(url, "git@") {
        // Use SSH key
        home, _ := os.UserHomeDir()
        keyPath := filepath.Join(home, ".ssh", "id_rsa")

        return git.SSHKeyFile("git", keyPath, "")
    }

    // Check for HTTPS URLs
    if strings.HasPrefix(url, "https://") {
        // Check for GitHub token in environment
        token := os.Getenv("GITHUB_TOKEN")
        if token != "" {
            return git.BasicAuth("", token), nil
        }

        // Fall back to no auth (public repos)
        return git.EmptyAuth(), nil
    }

    return git.EmptyAuth(), nil
}
```

---

## Error Handling

The library wraps go-git errors with platform error types. Common errors:

- `ErrNotFound` - Repository, reference, tag, or branch not found
- `ErrAlreadyExists` - Branch, tag, or remote already exists
- `ErrAuthenticationFailed` - Authentication failure
- `ErrNetwork` - Network or connectivity issues
- `ErrInvalidInput` - Invalid reference or bad parameters
- `ErrConflict` - Dirty worktree or merge conflicts

**Usage**:
```go
repo, err := git.Clone(ctx, url)
if err != nil {
    if errors.Is(err, git.ErrAuthenticationFailed) {
        // Handle auth error
    }
    return err
}
```

---

## Context and Cancellation

Network operations (Clone, Fetch, Push, Pull) accept `context.Context` for timeout control:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

repo, err := git.Clone(ctx, url)
// Clone will be cancelled after 30 seconds
```

Local operations (branch, tag, commit) do not require context as they operate on local filesystem only.

---

## Testing

Use the testutil sub-package for testing:

```go
import "github.com/jmgilman/go/git/testutil"

// Create in-memory repository
repo, fs, err := testutil.NewMemoryRepo()

// Create test commits
hash, err := testutil.CreateTestCommit(repo, "Test commit")

// Create test files
err := testutil.CreateTestFile(fs, "test.txt", "content")
```

Mock network operations:
```go
mockOps := &mockRemoteOps{...}
repo, err := git.Clone(ctx, url, git.WithRemoteOperations(mockOps))
```

---

## Key Benefits for sow CLI

1. **Simplified API**: Clean interface over go-git complexity
2. **Billy Integration**: Seamless filesystem operations
3. **Type Safety**: Compile-time checks for options
4. **Testability**: Mock network operations for tests
5. **Escape Hatches**: Full go-git access when needed

---

## Related Documentation

- go-git: https://github.com/go-git/go-git
- go-billy: https://github.com/go-git/go-billy
- Platform errors: `github.com/jmgilman/go/errors`
