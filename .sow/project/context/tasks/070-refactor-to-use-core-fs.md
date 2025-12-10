# Refactor Config Functions to Use core.FS

## Context

This is a rework task from review feedback. The original implementation incorrectly:

1. Defined a local `FS` interface in `repo.go` instead of using `core.FS` from `github.com/jmgilman/go/fs/core`
2. Used `os.ReadFile` directly in `user.go` instead of accepting a `core.FS` parameter

The requirement was to decouple config loading from `sow.Context` by accepting explicit `core.FS` dependencies - not by defining our own interface or using `os` directly.

## Requirements

### 1. Update repo.go

**Remove** the local FS interface:
```go
// DELETE THIS
type FS interface {
    ReadFile(name string) ([]byte, error)
}
```

**Import** `core.FS`:
```go
import (
    "github.com/jmgilman/go/fs/core"
    // ...
)
```

**Update** `LoadRepoConfig` signature:
```go
// Before
func LoadRepoConfig(fsys FS) (*schemas.Config, error)

// After
func LoadRepoConfig(fsys core.FS) (*schemas.Config, error)
```

### 2. Update user.go

**Add** `core.FS` parameter to user config loading functions:

```go
// Before
func LoadUserConfig() (*schemas.UserConfig, error)
func LoadUserConfigFromPath(path string) (*schemas.UserConfig, error)

// After - accept FS for file operations
func LoadUserConfig(fsys core.FS) (*schemas.UserConfig, error)
func LoadUserConfigFromPath(fsys core.FS, path string) (*schemas.UserConfig, error)
```

**Replace** `os.ReadFile` with `fsys.ReadFile`:
```go
// Before
data, err := os.ReadFile(path)

// After
data, err := fsys.ReadFile(path)
```

**Note**: `GetUserConfigPath()` can still use `os` for environment variables and home directory lookup since those are about path resolution, not file reading.

### 3. Update tests - Use billy.MemoryFS

**IMPORTANT**: Use `billy.NewMemory()` from `github.com/jmgilman/go/fs/billy` for testing. Do NOT write mocks.

Update `repo_test.go`:
```go
import (
    "github.com/jmgilman/go/fs/billy"
    // ...
)

func TestLoadRepoConfig(t *testing.T) {
    // Create in-memory filesystem
    memfs := billy.NewMemory()

    // Write test file
    memfs.WriteFile("config.yaml", []byte("artifacts:\n  adrs: custom"), 0644)

    // Test loading
    cfg, err := LoadRepoConfig(memfs)
    // ...
}
```

Update `user_test.go`:
```go
import (
    "github.com/jmgilman/go/fs/billy"
    // ...
)

func TestLoadUserConfigFromPath(t *testing.T) {
    memfs := billy.NewMemory()

    // Create directory structure
    memfs.MkdirAll("home/.config/sow", 0755)
    memfs.WriteFile("home/.config/sow/config.yaml", []byte("..."), 0644)

    cfg, err := LoadUserConfigFromPath(memfs, "home/.config/sow/config.yaml")
    // ...
}
```

**Remove** the `mockFS` struct from `repo_test.go` - use `billy.NewMemory()` instead.

### 4. Update consumer files

Update all 8 consumer files in CLI to pass a `core.FS` instance:
- `cli/cmd/config/validate.go`
- `cli/cmd/config/show.go`
- `cli/cmd/config/reset.go`
- `cli/cmd/config/path.go`
- `cli/cmd/config/init.go`
- `cli/cmd/config/edit.go`
- `cli/cmd/agent/spawn.go`
- `cli/cmd/agent/resume.go`

The CLI already has access to `core.FS` via `sow.Context`. Consumers should use that:

```go
// In CLI commands that have access to sow.Context
func runCommand(cmd *cobra.Command, args []string) error {
    ctx := cmdutil.GetContext(cmd)
    userCfg, err := config.LoadUserConfig(ctx.FS)
    // ...
}
```

For commands that don't have a context (like config commands that work pre-init), create a local FS:
```go
import "github.com/jmgilman/go/fs/billy"

func runValidate(cmd *cobra.Command, _ []string) error {
    fsys := billy.NewLocal()
    // ...
}
```

### 5. Update go.mod

Add dependency on `github.com/jmgilman/go/fs/core` and `github.com/jmgilman/go/fs/billy` in `libs/config/go.mod`:

```
require (
    github.com/jmgilman/go/fs/core v0.0.0
    github.com/jmgilman/go/fs/billy v0.0.0
)

replace (
    github.com/jmgilman/go/fs/core => /Users/josh/code/go/fs/core
    github.com/jmgilman/go/fs/billy => /Users/josh/code/go/fs/billy
)
```

## Acceptance Criteria

1. [ ] `libs/config/repo.go` uses `core.FS` type (not local interface)
2. [ ] `libs/config/user.go` accepts `core.FS` parameter (not using `os` directly for file reading)
3. [ ] No local `FS` interface defined in libs/config
4. [ ] No `mockFS` struct in tests - uses `billy.NewMemory()` instead
5. [ ] All tests updated and passing
6. [ ] All consumer files updated to pass `core.FS`
7. [ ] `go build ./...` succeeds for both libs/config and cli
8. [ ] `go test ./...` passes
9. [ ] `golangci-lint run` passes

## Relevant Inputs

- `libs/config/repo.go` - Current implementation with local FS interface
- `libs/config/user.go` - Current implementation using os.ReadFile
- `libs/config/repo_test.go` - Tests to update (remove mockFS, use billy.NewMemory)
- `libs/config/user_test.go` - Tests to update
- `cli/cmd/config/validate.go` - Consumer to update
- `cli/cmd/agent/spawn.go` - Consumer to update
- `/Users/josh/code/go/fs/core/interfaces.go` - core.FS interface definition
- `/Users/josh/code/go/fs/billy/billy.go` - billy.MemoryFS and billy.LocalFS implementations

## Technical Notes

### core.FS Interface

The `core.FS` interface from `github.com/jmgilman/go/fs/core` provides:
```go
type FS interface {
    fs.FS // stdlib compatibility
    ReadFS
    WriteFS
    ManageFS
    WalkFS
    ChrootFS
    Type() FSType
}
```

Key methods from `ReadFS`:
```go
type ReadFS interface {
    Open(name string) (fs.File, error)
    Stat(name string) (fs.FileInfo, error)
    ReadDir(name string) ([]fs.DirEntry, error)
    ReadFile(name string) ([]byte, error)
    Exists(name string) (bool, error)
}
```

### Creating core.FS Instances

```go
import "github.com/jmgilman/go/fs/billy"

// For real filesystem access (rooted at /)
fsys := billy.NewLocal()

// For testing (in-memory)
memfs := billy.NewMemory()
memfs.WriteFile("config.yaml", []byte("..."), 0644)
```

### billy.MemoryFS for Testing

The `billy.MemoryFS` implements `core.FS` and provides an in-memory filesystem:
- `NewMemory()` creates a new empty in-memory FS
- `WriteFile(name, data, perm)` writes a file
- `MkdirAll(path, perm)` creates directories
- Files persist for the lifetime of the MemoryFS instance

## Constraints

- Do NOT change the API signatures beyond adding the `core.FS` parameter
- Do NOT change business logic - only the filesystem access mechanism
- Do NOT write custom mock implementations - use `billy.NewMemory()` for testing
- Maintain backward compatibility where possible by having consumers construct the FS
