# fs/billy Library Reference

**Package**: `github.com/jmgilman/go/fs/billy`

**Purpose**: go-billy-backed filesystem implementation of `core.FS` interface, enabling go-git compatibility.

---

## Overview

This package wraps go-billy's `osfs` (local) and `memfs` (in-memory) implementations, providing a thin adapter layer that:
- Implements `core.FS` interface
- Maintains access to underlying `billy.Filesystem` for go-git integration
- Provides thread-safe filesystem operations
- Works seamlessly with the git library

**Key Concept**: The `Unwrap()` method provides access to the underlying billy.Filesystem, which is required for go-git operations.

---

## Core Types

### LocalFS - Local Filesystem

```go
type LocalFS struct {
    bfs billy.Filesystem
}

func NewLocal(opts ...Option) *LocalFS
```

**Usage**:
```go
fs := billy.NewLocal()

// Use with core.FS interface
data, err := fs.ReadFile("config.json")

// Unwrap for go-git integration
billyFS := fs.Unwrap()
// Use billyFS with go-git...
```

### MemoryFS - In-Memory Filesystem

```go
type MemoryFS struct {
    bfs billy.Filesystem
}

func NewMemory(opts ...Option) *MemoryFS
```

**Usage**:
```go
fs := billy.NewMemory()
err := fs.WriteFile("temp.txt", []byte("data"), 0644)
```

---

## Key Methods

### Unwrap - Access Underlying Billy Filesystem

```go
func (lfs *LocalFS) Unwrap() billy.Filesystem
func (mfs *MemoryFS) Unwrap() billy.Filesystem
```

**Purpose**: Returns the underlying `billy.Filesystem` for go-git integration.

**Usage**:
```go
fs := billy.NewLocal()
billyFS := fs.Unwrap()

// Use with go-git
repo, err := git.Clone(ctx, url, git.WithFilesystem(billyFS))
```

### All core.FS Methods

Both `LocalFS` and `MemoryFS` implement the full `core.FS` interface:

**ReadFS**:
- `Open(name string) (fs.File, error)`
- `Stat(name string) (fs.FileInfo, error)`
- `ReadDir(name string) ([]fs.DirEntry, error)`
- `ReadFile(name string) ([]byte, error)`
- `Exists(name string) (bool, error)`

**WriteFS**:
- `Create(name string) (core.File, error)`
- `OpenFile(name string, flag int, perm fs.FileMode) (core.File, error)`
- `WriteFile(name string, data []byte, perm fs.FileMode) error`
- `Mkdir(name string, perm fs.FileMode) error`
- `MkdirAll(path string, perm fs.FileMode) error`

**ManageFS**:
- `Remove(name string) error`
- `RemoveAll(path string) error`
- `Rename(oldpath, newpath string) error`

**WalkFS**:
- `Walk(root string, walkFn fs.WalkDirFunc) error`

**ChrootFS**:
- `Chroot(dir string) (core.FS, error)`

---

## Thread Safety

- **FS instances** (`LocalFS`, `MemoryFS`) are safe for concurrent use by multiple goroutines
- **File handles** are NOT safe for concurrent use (standard Go file semantics)

---

## Application in sow CLI

### Use Case 1: Main Filesystem Abstraction

**File**: `internal/config/filesystem.go`

```go
package config

import (
    "github.com/jmgilman/go/fs/billy"
    "github.com/jmgilman/go/fs/core"
)

var (
    // Global filesystem instance
    // In production: local filesystem
    // In tests: can be swapped with memory filesystem
    Filesystem core.FS
)

func init() {
    // Initialize with local filesystem
    Filesystem = billy.NewLocal()
}

// For testing
func SetFilesystem(fs core.FS) {
    Filesystem = fs
}
```

**File**: `cmd/sow/main.go`

```go
package main

import (
    "github.com/spf13/cobra"
    "sow/internal/config"
    "sow/internal/commands"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "sow",
        Short: "AI-powered system of work",
    }

    // Pass filesystem to all commands
    fs := config.Filesystem

    rootCmd.AddCommand(commands.NewInitCmd(fs))
    rootCmd.AddCommand(commands.NewValidateCmd(fs))
    rootCmd.AddCommand(commands.NewLogCmd(fs))
    // ... other commands

    rootCmd.Execute()
}
```

### Use Case 2: Git Integration (Refs Management)

**File**: `internal/refs/manager.go`

```go
package refs

import (
    "context"
    "github.com/jmgilman/go/fs/billy"
    "github.com/jmgilman/go/git"
)

type Manager struct {
    cacheDir string
    fs       *billy.LocalFS
}

func NewManager(cacheDir string) *Manager {
    return &Manager{
        cacheDir: cacheDir,
        fs:       billy.NewLocal(),
    }
}

func (m *Manager) CloneRef(ctx context.Context, url, branch string) error {
    // Unwrap billy filesystem for go-git
    billyFS := m.fs.Unwrap()

    // Clone using git library (which requires billy.Filesystem)
    repo, err := git.Clone(ctx, url,
        git.WithFilesystem(billyFS),
        git.WithBranch(branch),
        git.WithDepth(1))
    if err != nil {
        return err
    }

    // Can also use fs for regular file operations
    exists, _ := m.fs.Exists(".sow/refs/index.json")

    return nil
}

func (m *Manager) UpdateRef(ctx context.Context, refID string) error {
    // Open existing repo
    billyFS := m.fs.Unwrap()

    repo, err := git.Open(refID, git.WithFilesystem(billyFS))
    if err != nil {
        return err
    }

    // Fetch updates
    return repo.Fetch(ctx, git.FetchOptions{})
}
```

### Use Case 3: Schema Embedding with Memory Filesystem

**File**: `internal/schema/embed.go`

```go
package schema

import (
    _ "embed"
    "context"
    "github.com/jmgilman/go/cue"
    "github.com/jmgilman/go/fs/billy"
)

//go:embed ../../schemas/cue/project-state.cue
var projectStateCUE []byte

//go:embed ../../schemas/cue/task-state.cue
var taskStateCUE []byte

var (
    schemaFS      *billy.MemoryFS
    loader        *cue.Loader
    projectSchema cue.Value
    taskSchema    cue.Value
)

func init() {
    // Create in-memory filesystem for embedded schemas
    schemaFS = billy.NewMemory()

    // Write embedded schemas to memory
    schemaFS.WriteFile("project-state.cue", projectStateCUE, 0644)
    schemaFS.WriteFile("task-state.cue", taskStateCUE, 0644)

    // Create CUE loader with memory filesystem
    loader = cue.NewLoader(schemaFS)

    // Load schemas
    ctx := context.Background()
    projectSchema, _ = loader.LoadFile(ctx, "project-state.cue")
    taskSchema, _ = loader.LoadFile(ctx, "task-state.cue")
}

// Schemas are now available in memory, no disk I/O needed
```

### Use Case 4: Testing with Memory Filesystem

**File**: `internal/commands/init_test.go`

```go
package commands

import (
    "testing"
    "github.com/jmgilman/go/fs/billy"
    "github.com/stretchr/testify/assert"
)

func TestInitCommand(t *testing.T) {
    // Use memory filesystem for fast testing
    fs := billy.NewMemory()

    // Create git directory to simulate repo
    fs.MkdirAll(".git", 0755)

    // Run init command
    cmd := NewInitCmd(fs)
    err := cmd.Execute()

    // Assert results
    assert.NoError(t, err)

    // Verify directory structure created
    exists, _ := fs.Exists(".sow/knowledge")
    assert.True(t, exists)

    exists, _ = fs.Exists(".sow/refs")
    assert.True(t, exists)
}

func TestInitCommandAlreadyInitialized(t *testing.T) {
    fs := billy.NewMemory()

    // Pre-create .sow directory
    fs.MkdirAll(".sow", 0755)

    // Should error on second init
    cmd := NewInitCmd(fs)
    err := cmd.Execute()

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already initialized")
}
```

### Use Case 5: Chroot for Scoped Operations

**File**: `internal/logging/writer.go`

```go
package logging

import (
    "fmt"
    "os"
    "github.com/jmgilman/go/fs/core"
)

func WriteTaskLog(fs core.FS, taskDir, message string) error {
    // Chroot to task directory for scoped operations
    taskFS, err := fs.Chroot(taskDir)
    if err != nil {
        return err
    }

    // All paths now relative to task directory
    // No need to build full paths, prevents directory traversal

    // Read current log
    var existingLog []byte
    exists, _ := taskFS.Exists("log.md")
    if exists {
        existingLog, _ = taskFS.ReadFile("log.md")
    }

    // Append new entry
    newLog := append(existingLog, []byte(message+"\n")...)

    // Write back (scoped to task directory)
    return taskFS.WriteFile("log.md", newLog, 0644)
}
```

---

## Filesystem Root

Both `LocalFS` and `MemoryFS` are rooted at `/` by default.

**LocalFS**: Maps to OS filesystem root
```go
fs := billy.NewLocal()
data, _ := fs.ReadFile("/Users/josh/code/sow/README.md")  // Absolute path
data, _ := fs.ReadFile("relative/path/file.txt")          // Relative to /
```

**MemoryFS**: Virtual root in memory
```go
fs := billy.NewMemory()
fs.WriteFile("test.txt", []byte("data"), 0644)  // Stored at / in memory
```

---

## Path Normalization

The library automatically normalizes paths:
- Converts backslashes to forward slashes (Windows compatibility)
- Cleans paths using `filepath.Clean`
- Billy handles security (prevents directory traversal)

```go
// These are equivalent after normalization:
fs.ReadFile("path/to/file.txt")
fs.ReadFile("path\\to\\file.txt")  // Windows-style
fs.ReadFile("path//to/../to/./file.txt")  // Redundant separators
```

---

## RemoveAll Implementation

Note: Billy doesn't have native `RemoveAll`, so this library implements it via recursive removal:

```go
func (lfs *LocalFS) RemoveAll(path string) error {
    // Returns nil if path doesn't exist (idempotent)
    // Recursively removes directories and contents
    // Returns first error encountered
}
```

---

## Key Benefits for sow CLI

1. **Go-Git Compatibility**: Seamless integration via `Unwrap()`
2. **Testing**: Fast in-memory filesystem for unit tests
3. **Type Safety**: Compile-time core.FS interface checks
4. **Performance**: No overhead beyond thin wrapper
5. **Simplicity**: Single implementation for all file operations

---

## Common Patterns

### Pattern 1: Dual-Mode (Local + Git)

```go
func processRef(ctx context.Context, fs *billy.LocalFS, refPath string) error {
    // Use core.FS interface for regular file operations
    data, err := fs.ReadFile(filepath.Join(refPath, "config.yaml"))
    if err != nil {
        return err
    }

    // Unwrap for git operations
    billyFS := fs.Unwrap()
    repo, err := git.Open(refPath, git.WithFilesystem(billyFS))
    if err != nil {
        return err
    }

    return repo.Fetch(ctx, git.FetchOptions{})
}
```

### Pattern 2: Test Fixture Setup

```go
func setupTestFS() *billy.MemoryFS {
    fs := billy.NewMemory()

    // Create test structure
    fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)

    // Populate test data
    fs.WriteFile(".sow/project/state.yaml", testStateYAML, 0644)
    fs.WriteFile(".sow/project/phases/implementation/tasks/010/state.yaml", testTaskYAML, 0644)

    return fs
}
```

---

## Related Documentation

- go-billy: https://github.com/go-git/go-billy
- fs/core interfaces: `github.com/jmgilman/go/fs/core`
- git library integration: `github.com/jmgilman/go/git`
