# sow CLI Library Integration Overview

This directory contains reference documentation for the core libraries used in the sow CLI implementation.

---

## Library Stack

The sow CLI is built on four primary libraries from `/Users/josh/code/go`:

1. **fs/core** - Filesystem abstraction interfaces
2. **fs/billy** - Billy-backed filesystem implementation
3. **cue** - CUE schema management and validation
4. **git** - Git operations wrapper

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                      sow CLI                            │
│                  (Cobra Commands)                       │
└─────────────────┬───────────────────────────────────────┘
                  │
        ┌─────────┴─────────┬──────────────┬─────────────┐
        │                   │              │             │
        ▼                   ▼              ▼             ▼
┌───────────────┐  ┌────────────┐  ┌──────────┐  ┌──────────┐
│  cue library  │  │ fs/billy   │  │ fs/core  │  │   git    │
│               │  │            │  │          │  │          │
│ - Loader      │  │ - LocalFS  │  │ - FS     │  │- Clone   │
│ - Validator   │  │ - MemoryFS │  │ - ReadFS │  │- Fetch   │
│ - Encoder     │  │ - Unwrap() │  │ - WriteFS│  │- Push    │
│ - Decoder     │  │            │  │ - WalkFS │  │- Commit  │
└───────┬───────┘  └─────┬──────┘  └─────┬────┘  └────┬─────┘
        │                │               │            │
        │  Uses          │  Implements   │  Uses      │
        └────────────────┼───────────────┘            │
                         │                            │
                         ▼                            ▼
                  ┌──────────────┐          ┌─────────────────┐
                  │ billy.Filesystem        │  go-git         │
                  │ (go-billy)   │          │ (underlying)    │
                  └──────────────┘          └─────────────────┘
```

---

## How Libraries Work Together

### Example 1: Schema Validation

```go
// 1. Create memory filesystem with embedded schemas (fs/billy)
schemaFS := billy.NewMemory()
schemaFS.WriteFile("project.cue", embeddedProjectCUE, 0644)

// 2. Create CUE loader with filesystem (cue)
loader := cue.NewLoader(schemaFS)
schema, _ := loader.LoadFile(ctx, "project.cue")

// 3. Load data from local filesystem (fs/billy)
localFS := billy.NewLocal()
dataLoader := cue.NewLoader(localFS)
data, _ := dataLoader.LoadFile(ctx, ".sow/project/state.yaml")

// 4. Validate (cue)
err := cue.Validate(ctx, schema, data)
```

**Libraries Used**: `fs/billy` → `cue`

### Example 2: Refs Management with Git

```go
// 1. Create local filesystem (fs/billy)
fs := billy.NewLocal()

// 2. Clone repository (git + billy)
billyFS := fs.Unwrap()  // Get billy.Filesystem for go-git
repo, err := git.Clone(ctx, url,
    git.WithFilesystem(billyFS),
    git.WithAuth(auth))

// 3. Read files using core.FS interface (fs/core + fs/billy)
data, err := fs.ReadFile("config.yaml")

// 4. Fetch updates (git)
repo.Fetch(ctx, git.FetchOptions{Auth: auth})
```

**Libraries Used**: `fs/billy` → `git` → `fs/core`

### Example 3: Complete Workflow - Init Command

```go
func InitCommand(fs core.FS) error {
    // 1. Check if already initialized (fs/core)
    exists, _ := fs.Exists(".sow")
    if exists {
        return errors.New("already initialized")
    }

    // 2. Create directory structure (fs/core via fs/billy)
    fs.MkdirAll(".sow/knowledge", 0755)
    fs.MkdirAll(".sow/refs", 0755)

    // 3. Load embedded schema (cue + fs/billy)
    schemaFS := billy.NewMemory()
    schemaFS.WriteFile("version.cue", versionCUE, 0644)

    loader := cue.NewLoader(schemaFS)
    schema, _ := loader.LoadFile(ctx, "version.cue")

    // 4. Create version file (cue)
    versionData, _ := cue.EncodeYAML(ctx, schema)

    // 5. Write to disk (fs/core via fs/billy)
    fs.WriteFile(".sow/.version", versionData, 0644)

    return nil
}
```

**Libraries Used**: `fs/core` → `fs/billy` → `cue`

---

## Library Interaction Patterns

### Pattern 1: Filesystem Abstraction

**Goal**: Write filesystem-agnostic code that works in production and tests

```go
// Function accepts core.FS interface
func processFile(fs core.FS, path string) error {
    data, err := fs.ReadFile(path)
    if err != nil {
        return err
    }
    // Process data...
    return fs.WriteFile(path+".processed", result, 0644)
}

// Production: use local filesystem
func main() {
    fs := billy.NewLocal()
    processFile(fs, "input.txt")
}

// Testing: use memory filesystem
func TestProcessFile(t *testing.T) {
    fs := billy.NewMemory()
    fs.WriteFile("input.txt", []byte("test"), 0644)

    processFile(fs, "input.txt")

    result, _ := fs.ReadFile("input.txt.processed")
    assert.Equal(t, expected, result)
}
```

### Pattern 2: Schema-Driven Validation

**Goal**: Validate all files against embedded CUE schemas

```go
// 1. Embed schemas at compile time
//go:embed schemas/project.cue
var projectSchemaCUE []byte

// 2. Load into memory filesystem
func init() {
    schemaFS := billy.NewMemory()
    schemaFS.WriteFile("project.cue", projectSchemaCUE, 0644)

    loader := cue.NewLoader(schemaFS)
    projectSchema, _ = loader.LoadFile(ctx, "project.cue")
}

// 3. Validate files from any filesystem
func ValidateProject(fs core.ReadFS, path string) error {
    loader := cue.NewLoader(fs)
    data, _ := loader.LoadFile(ctx, path)

    return cue.Validate(ctx, projectSchema, data)
}
```

### Pattern 3: Git + Filesystem Integration

**Goal**: Clone repos and interact with files seamlessly

```go
func cloneAndProcess(ctx context.Context, url string) error {
    // 1. Create filesystem
    fs := billy.NewLocal()

    // 2. Clone using git (requires billy.Filesystem)
    billyFS := fs.Unwrap()
    repo, err := git.Clone(ctx, url, git.WithFilesystem(billyFS))
    if err != nil {
        return err
    }

    // 3. Use core.FS interface for file operations
    data, err := fs.ReadFile("config.yaml")

    // 4. Process with CUE
    loader := cue.NewLoader(fs)
    value, _ := loader.LoadFile(ctx, "schema.cue")

    // 5. Update repo
    repo.Fetch(ctx, git.FetchOptions{})

    return nil
}
```

---

## Key Design Decisions

### 1. Interface-Based Design (fs/core)

**Why**: Enables testing, multiple implementations, and future extensibility

```go
// Functions accept interfaces, not concrete types
func ValidateStructure(fs core.FS) error { ... }

// Can use LocalFS or MemoryFS
ValidateStructure(billy.NewLocal())
ValidateStructure(billy.NewMemory())
```

### 2. Billy as Bridge to go-git

**Why**: go-git requires billy.Filesystem, but we want core.FS interface

```go
// fs/billy provides both:
fs := billy.NewLocal()

// 1. Implements core.FS interface
data, _ := fs.ReadFile("file.txt")  // core.FS method

// 2. Unwraps to billy.Filesystem for go-git
billyFS := fs.Unwrap()
git.Clone(ctx, url, git.WithFilesystem(billyFS))
```

### 3. CUE for Schema Validation

**Why**: Strong typing, constraint validation, excellent error messages

```go
// Schema defines constraints
schema := `
name: string & =~"^[a-z-]+$"
port: int & >0 & <65536
`

// Validation provides detailed errors
err := cue.Validate(ctx, schema, data)
// Error: "port: invalid value 70000 (out of bound <65536)"
```

### 4. Embedded Schemas with go:embed

**Why**: Single binary distribution, no external schema files needed

```go
//go:embed schemas/project.cue
var projectSchemaCUE []byte

// Load into memory filesystem
schemaFS := billy.NewMemory()
schemaFS.WriteFile("project.cue", projectSchemaCUE, 0644)

// Binary now contains schemas, no runtime file I/O needed
```

---

## CLI Command Implementation Pattern

Standard pattern for implementing CLI commands:

```go
package commands

import (
    "github.com/spf13/cobra"
    "github.com/jmgilman/go/fs/core"
)

// Command factory accepts filesystem
func NewMyCommand(fs core.FS) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mycommand",
        Short: "Description",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runMyCommand(fs, args)
        },
    }

    // Add flags
    cmd.Flags().String("option", "", "Help text")

    return cmd
}

// Implementation uses filesystem abstraction
func runMyCommand(fs core.FS, args []string) error {
    // 1. Read files (fs/core)
    data, err := fs.ReadFile(".sow/project/state.yaml")
    if err != nil {
        return err
    }

    // 2. Validate (cue)
    if err := validateState(fs, data); err != nil {
        return err
    }

    // 3. Process and write (fs/core)
    return fs.WriteFile("output.txt", result, 0644)
}
```

---

## Testing Strategy

### Unit Tests: Memory Filesystem

```go
func TestCommand(t *testing.T) {
    // Use memory filesystem for fast, isolated tests
    fs := billy.NewMemory()

    // Set up test data
    fs.MkdirAll(".sow/project", 0755)
    fs.WriteFile(".sow/project/state.yaml", testData, 0644)

    // Run command
    cmd := NewMyCommand(fs)
    err := cmd.Execute()

    // Assert results
    assert.NoError(t, err)

    result, _ := fs.ReadFile("output.txt")
    assert.Equal(t, expected, result)
}
```

### Integration Tests: Local Filesystem

```go
func TestIntegration(t *testing.T) {
    // Use temp directory with local filesystem
    tmpDir := t.TempDir()

    fs := billy.NewLocal()
    chrootFS, _ := fs.Chroot(tmpDir)

    // Run real operations
    // ...
}
```

---

## Performance Characteristics

| Operation | Library | Performance | Notes |
|-----------|---------|-------------|-------|
| Schema loading | cue | ~100ms | One-time at init |
| Validation | cue | <100ms | Fast for typical files |
| File read/write | fs/billy | Native | No overhead |
| Git clone | git | Network-bound | Use shallow clones |
| Git fetch | git | Network-bound | Fast for small changes |

---

## Error Handling

All libraries use platform error types from `github.com/jmgilman/go/errors`:

```go
import "github.com/jmgilman/go/errors"

// CUE errors
if err != nil {
    if errors.Is(err, cue.CodeCUEValidationFailed) {
        // Handle validation error
        // err.Context() contains field paths
    }
}

// Git errors
if err != nil {
    if errors.Is(err, git.ErrAuthenticationFailed) {
        // Handle auth error
    }
}

// Filesystem errors (standard library)
if err != nil {
    if os.IsNotExist(err) {
        // Handle not found
    }
}
```

---

## Next Steps

1. Read individual library references:
   - [01_cue_library.md](./01_cue_library.md)
   - [02_fs_core_library.md](./02_fs_core_library.md)
   - [03_fs_billy_library.md](./03_fs_billy_library.md)
   - [04_git_library.md](./04_git_library.md)

2. Review ADRs:
   - [.sow/knowledge/adrs/001-go-cli-architecture.md](../.sow/knowledge/adrs/001-go-cli-architecture.md)
   - [.sow/knowledge/adrs/002-build-and-distribution-strategy.md](../.sow/knowledge/adrs/002-build-and-distribution-strategy.md)

3. Begin CLI implementation following the architecture defined in ADR 001
