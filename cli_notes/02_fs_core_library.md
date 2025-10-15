# fs/core Library Reference

**Package**: `github.com/jmgilman/go/fs/core`

**Purpose**: Foundational interfaces for multi-provider filesystem abstraction compatible with Go's `io/fs` stdlib.

---

## Overview

The core package provides contracts that filesystem providers must implement, enabling applications to write filesystem-agnostic code that works with:
- Local filesystems (via billy)
- In-memory filesystems (for testing)
- Cloud storage (future: S3)

**Design Philosophy**:
- **Zero dependencies**: Only uses Go standard library
- **Interface composition**: Small focused interfaces compose into larger contracts
- **Stdlib compatibility**: Extends `fs.FS` and `fs.File` rather than replacing them
- **Optional capabilities**: Use type assertions for provider-specific features

---

## Core Interface Hierarchy

### Main FS Interface

```go
type FS interface {
    fs.FS  // Embeds stdlib fs.FS (provides Open returning fs.File)
    ReadFS
    WriteFS
    ManageFS
    WalkFS
    ChrootFS
}
```

All filesystem providers **MUST** implement this interface, which is composed of five sub-interfaces.

---

## Sub-Interfaces

### 1. ReadFS - Read Operations

All backends MUST support this interface.

```go
type ReadFS interface {
    // Open opens file for reading (returns fs.File)
    Open(name string) (fs.File, error)

    // Stat returns file metadata
    Stat(name string) (fs.FileInfo, error)

    // ReadDir reads directory and returns entries sorted by filename
    ReadDir(name string) ([]fs.DirEntry, error)

    // ReadFile reads entire file
    ReadFile(name string) ([]byte, error)

    // Exists reports whether file or directory exists
    Exists(name string) (bool, error)
}
```

**Usage**:
```go
data, err := fs.ReadFile(".sow/project/state.yaml")
if err != nil {
    return err
}
```

### 2. WriteFS - Write Operations

```go
type WriteFS interface {
    // Create creates or truncates file for writing
    Create(name string) (File, error)

    // OpenFile opens file with flags and permissions
    OpenFile(name string, flag int, perm fs.FileMode) (File, error)

    // WriteFile writes data to file
    WriteFile(name string, data []byte, perm fs.FileMode) error

    // Mkdir creates directory (fails if parent doesn't exist)
    Mkdir(name string, perm fs.FileMode) error

    // MkdirAll creates directory with parents
    MkdirAll(path string, perm fs.FileMode) error
}
```

**Usage**:
```go
err := fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
err = fs.WriteFile(".sow/project/state.yaml", data, 0644)
```

### 3. ManageFS - File Management

```go
type ManageFS interface {
    // Remove removes file or empty directory
    Remove(name string) error

    // RemoveAll removes path and children (returns nil if not exists)
    RemoveAll(path string) error

    // Rename renames (moves) oldpath to newpath
    Rename(oldpath, newpath string) error
}
```

**Usage**:
```go
// Cleanup project folder
err := fs.RemoveAll(".sow/project")
```

### 4. WalkFS - Directory Traversal

```go
type WalkFS interface {
    // Walk walks file tree, calling walkFn for each file/directory
    Walk(root string, walkFn fs.WalkDirFunc) error
}
```

**Usage**:
```go
err := fs.Walk(".sow", func(path string, d fs.DirEntry, err error) error {
    if err != nil {
        return err
    }
    fmt.Println(path)
    return nil
})
```

### 5. ChrootFS - Scoped Filesystem Views

```go
type ChrootFS interface {
    // Chroot returns filesystem scoped to directory
    // All operations relative to dir, cannot access paths outside
    Chroot(dir string) (FS, error)
}
```

**Usage**:
```go
// Scope filesystem to task directory
taskFS, err := fs.Chroot(".sow/project/phases/implementation/tasks/010")

// Now all operations are relative to task directory
data, _ := taskFS.ReadFile("state.yaml")  // Actually .sow/project/.../010/state.yaml
```

---

## File Interface

```go
type File interface {
    fs.File     // Embeds: Read, Close, Stat
    io.Writer   // Write([]byte) (int, error)
    Name() string
}
```

Optional capabilities (use type assertions):
- `io.Seeker` - Seek operations
- `io.ReaderAt` - ReadAt operations
- `io.WriterAt` - WriteAt operations
- `Truncater` - Truncate operations
- `Syncer` - Sync to disk

**Usage**:
```go
file, err := fs.Create("output.txt")
if err != nil {
    return err
}
defer file.Close()

// Write
_, err = file.Write([]byte("data"))

// Seek if supported
if seeker, ok := file.(io.Seeker); ok {
    seeker.Seek(0, io.SeekStart)
}
```

---

## Optional Interfaces (Provider-Specific)

### MetadataFS - Metadata Operations

Typically only local and memory filesystems. Cloud storage providers may not support.

```go
type MetadataFS interface {
    // Lstat returns file info without following symlinks
    Lstat(name string) (fs.FileInfo, error)

    // Chmod changes file permissions
    Chmod(name string, mode fs.FileMode) error

    // Chtimes changes access and modification times
    Chtimes(name string, atime, mtime time.Time) error
}
```

**Usage**:
```go
if mfs, ok := filesystem.(core.MetadataFS); ok {
    mfs.Chmod("file.txt", 0600)
}
```

### SymlinkFS - Symbolic Link Operations

Typically only local filesystems. Cloud storage does not support symlinks.

```go
type SymlinkFS interface {
    // Symlink creates symbolic link
    Symlink(oldname, newname string) error

    // Readlink returns destination of symbolic link
    Readlink(name string) (string, error)
}
```

**Usage**:
```go
if sfs, ok := filesystem.(core.SymlinkFS); ok {
    sfs.Symlink("/target/path", ".sow/refs/my-link")
}
```

### TempFS - Temporary File Operations

Typically only local and memory filesystems.

```go
type TempFS interface {
    // TempFile creates temporary file
    TempFile(dir, pattern string) (File, error)

    // TempDir creates temporary directory
    TempDir(dir, pattern string) (string, error)
}
```

**Usage**:
```go
if tfs, ok := filesystem.(core.TempFS); ok {
    tmpFile, _ := tfs.TempFile("", "sow-*.tmp")
    defer tmpFile.Close()
}
```

---

## Application in sow CLI

### Use Case 1: Filesystem Abstraction Layer

**File**: `pkg/sowfs/paths.go`

```go
package sowfs

import (
    "path/filepath"
    "github.com/jmgilman/go/fs/core"
)

// Paths provides standard .sow/ path constants
type Paths struct {
    RepoRoot   string
    SowDir     string
    ProjectDir string
    Knowledge  string
    Refs       string
}

func NewPaths(repoRoot string) *Paths {
    sowDir := filepath.Join(repoRoot, ".sow")
    return &Paths{
        RepoRoot:   repoRoot,
        SowDir:     sowDir,
        ProjectDir: filepath.Join(sowDir, "project"),
        Knowledge:  filepath.Join(sowDir, "knowledge"),
        Refs:       filepath.Join(sowDir, "refs"),
    }
}

// Helper functions using core.FS interface
func EnsureProjectStructure(fs core.FS, paths *Paths) error {
    // Create directories using MkdirAll
    dirs := []string{
        paths.Knowledge,
        paths.Refs,
        paths.ProjectDir,
    }

    for _, dir := range dirs {
        if err := fs.MkdirAll(dir, 0755); err != nil {
            return err
        }
    }

    return nil
}

func ProjectExists(fs core.ReadFS, paths *Paths) (bool, error) {
    return fs.Exists(filepath.Join(paths.ProjectDir, "state.yaml"))
}
```

### Use Case 2: Atomic File Writes

**File**: `pkg/sowfs/writer.go`

```go
package sowfs

import (
    "io/fs"
    "path/filepath"
    "github.com/jmgilman/go/fs/core"
)

// AtomicWrite writes file atomically (write to temp, then rename)
func AtomicWrite(filesystem core.FS, path string, data []byte, perm fs.FileMode) error {
    dir := filepath.Dir(path)
    base := filepath.Base(path)

    // Check if filesystem supports temp operations
    tfs, ok := filesystem.(core.TempFS)
    if !ok {
        // Fallback: direct write (not atomic but works on all providers)
        return filesystem.WriteFile(path, data, perm)
    }

    // Create temp file in same directory
    tmpFile, err := tfs.TempFile(dir, base+".tmp.*")
    if err != nil {
        return err
    }
    tmpPath := tmpFile.Name()

    // Write data
    _, err = tmpFile.Write(data)
    closeErr := tmpFile.Close()

    if err != nil {
        filesystem.Remove(tmpPath)
        return err
    }
    if closeErr != nil {
        filesystem.Remove(tmpPath)
        return closeErr
    }

    // Atomic rename
    return filesystem.Rename(tmpPath, path)
}
```

### Use Case 3: Context Detection

**File**: `internal/context/detector.go`

```go
package context

import (
    "path/filepath"
    "strings"
    "github.com/jmgilman/go/fs/core"
)

type Context struct {
    Type       string // "task" or "project"
    TaskDir    string
    ProjectDir string
    RepoRoot   string
}

func Detect(fs core.ReadFS, cwd string) (*Context, error) {
    // Walk up directory tree looking for .sow/
    repoRoot, err := findRepoRoot(fs, cwd)
    if err != nil {
        return nil, err
    }

    projectDir := filepath.Join(repoRoot, ".sow", "project")

    // Check if we're in a task directory
    if isTaskDir(cwd, projectDir) {
        return &Context{
            Type:       "task",
            TaskDir:    cwd,
            ProjectDir: projectDir,
            RepoRoot:   repoRoot,
        }, nil
    }

    return &Context{
        Type:       "project",
        ProjectDir: projectDir,
        RepoRoot:   repoRoot,
    }, nil
}

func findRepoRoot(fs core.ReadFS, dir string) (string, error) {
    for {
        sowDir := filepath.Join(dir, ".sow")
        exists, err := fs.Exists(sowDir)
        if err != nil {
            return "", err
        }
        if exists {
            return dir, nil
        }

        parent := filepath.Dir(dir)
        if parent == dir {
            return "", ErrNotInRepo
        }
        dir = parent
    }
}

func isTaskDir(cwd, projectDir string) bool {
    rel, err := filepath.Rel(projectDir, cwd)
    if err != nil {
        return false
    }

    parts := strings.Split(rel, string(filepath.Separator))
    return len(parts) == 4 && parts[0] == "phases" && parts[2] == "tasks"
}
```

### Use Case 4: Validation Command

**File**: `internal/commands/validate.go`

```go
package commands

import (
    "fmt"
    "path/filepath"
    "github.com/jmgilman/go/fs/core"
    "github.com/spf13/cobra"
)

func NewValidateCmd(fs core.FS) *cobra.Command {
    return &cobra.Command{
        Use:   "validate",
        Short: "Validate sow structure integrity",
        RunE: func(cmd *cobra.Command, args []string) error {
            return validateStructure(fs)
        },
    }
}

func validateStructure(fs core.FS) error {
    checks := []struct{
        name string
        fn   func() error
    }{
        {"Directory structure", func() error {
            return checkDirStructure(fs)
        }},
        {"Project state (if exists)", func() error {
            return checkProjectState(fs)
        }},
        {"Refs indexes", func() error {
            return checkRefsIndexes(fs)
        }},
    }

    for _, check := range checks {
        fmt.Printf("Checking %s... ", check.name)
        if err := check.fn(); err != nil {
            fmt.Println("FAIL")
            return err
        }
        fmt.Println("OK")
    }

    return nil
}

func checkDirStructure(fs core.ReadFS) error {
    requiredDirs := []string{
        ".sow",
        ".sow/knowledge",
        ".sow/refs",
    }

    for _, dir := range requiredDirs {
        exists, err := fs.Exists(dir)
        if err != nil {
            return err
        }
        if !exists {
            return fmt.Errorf("missing directory: %s", dir)
        }
    }

    return nil
}
```

---

## Stdlib Compatibility

Since `core.FS` embeds `fs.FS`, it's compatible with stdlib functions:

```go
import "io/fs"

// Works with stdlib functions
err := fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
    fmt.Println(path)
    return nil
})

// Read with stdlib
data, err := fs.ReadFile(filesystem, "config.json")
```

---

## Key Benefits for sow CLI

1. **Provider Independence**: Write once, run on any filesystem (local, memory, future cloud)
2. **Testing**: Use memory filesystem for fast unit tests
3. **Type Safety**: Compile-time interface checks
4. **Stdlib Compatible**: Works with existing Go filesystem functions
5. **Optional Features**: Check capabilities at runtime via type assertions

---

## Testing Strategy

```go
func TestWithMemoryFS(t *testing.T) {
    // Use billy.NewMemory() for in-memory testing
    fs := billy.NewMemory()

    // Set up test data
    fs.MkdirAll(".sow/project", 0755)
    fs.WriteFile(".sow/project/state.yaml", []byte("test"), 0644)

    // Test your functions
    exists, err := ProjectExists(fs, paths)
    assert.True(t, exists)
}
```

---

## Related Documentation

- Go stdlib `io/fs`: https://pkg.go.dev/io/fs
- Implementation: `github.com/jmgilman/go/fs/billy`
