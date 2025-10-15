# CLI Library Reference Documentation

Comprehensive reference for the core libraries used in the sow CLI implementation.

---

## Quick Links

### Overview
- **[00_OVERVIEW.md](./00_OVERVIEW.md)** - Architecture overview and library integration patterns

### Library References
1. **[01_cue_library.md](./01_cue_library.md)** - CUE schema management, validation, encoding/decoding
2. **[02_fs_core_library.md](./02_fs_core_library.md)** - Filesystem abstraction interfaces
3. **[03_fs_billy_library.md](./03_fs_billy_library.md)** - Billy-backed filesystem implementation
4. **[04_git_library.md](./04_git_library.md)** - Git operations wrapper

---

## What's Inside

Each library reference document contains:

- **Overview** - Purpose and key concepts
- **Core Types & Interfaces** - Primary API surface
- **Usage Patterns** - Common usage examples
- **Application in sow CLI** - Concrete use cases with code
- **Error Handling** - Error types and handling strategies
- **Performance Considerations** - Optimization tips
- **Testing Strategies** - How to test code using the library

---

## Library Stack Summary

```
┌─────────────────────────────────────┐
│          sow CLI (Cobra)            │
└──────────┬──────────────────────────┘
           │
     ┌─────┴─────┬──────┬──────┐
     │           │      │      │
     ▼           ▼      ▼      ▼
┌────────┐ ┌────────┐ ┌────┐ ┌─────┐
│  cue   │ │fs/billy│ │core│ │ git │
└────────┘ └────────┘ └────┘ └─────┘
     │           │      │      │
     │           └──────┴──────┘
     │                  │
     ▼                  ▼
┌─────────────────────────────────────┐
│   cuelang.org/go    go-git/go-billy │
└─────────────────────────────────────┘
```

---

## Quick Reference

### CUE Library
- **Load schemas**: `loader.LoadFile(ctx, "schema.cue")`
- **Validate**: `cue.Validate(ctx, schema, data)`
- **Encode**: `cue.EncodeYAML(ctx, value)`
- **Decode**: `cue.Decode(ctx, value, &struct)`

### fs/core Library
- **Interface**: `core.FS` (composed of ReadFS, WriteFS, ManageFS, WalkFS, ChrootFS)
- **Read**: `fs.ReadFile(path)`
- **Write**: `fs.WriteFile(path, data, perm)`
- **Walk**: `fs.Walk(root, walkFn)`
- **Chroot**: `fs.Chroot(dir)` - Scoped filesystem

### fs/billy Library
- **Local FS**: `billy.NewLocal()`
- **Memory FS**: `billy.NewMemory()`
- **Unwrap**: `fs.Unwrap()` - Get billy.Filesystem for go-git
- **Implements**: Full `core.FS` interface

### git Library
- **Clone**: `git.Clone(ctx, url, opts...)`
- **Open**: `git.Open(path, opts...)`
- **Fetch**: `repo.Fetch(ctx, opts)`
- **Auth**: `git.BasicAuth(user, token)`, `git.SSHKeyFile(user, path, pass)`

---

## Common Patterns

### Pattern 1: Schema Validation
```go
// Embed schema
//go:embed schema.cue
var schemaCUE []byte

// Load and validate
schemaFS := billy.NewMemory()
schemaFS.WriteFile("schema.cue", schemaCUE, 0644)
loader := cue.NewLoader(schemaFS)
schema, _ := loader.LoadFile(ctx, "schema.cue")

dataLoader := cue.NewLoader(localFS)
data, _ := dataLoader.LoadFile(ctx, "data.yaml")

err := cue.Validate(ctx, schema, data)
```

### Pattern 2: Git + Filesystem
```go
fs := billy.NewLocal()

// Clone (requires billy.Filesystem)
repo, _ := git.Clone(ctx, url, git.WithFilesystem(fs.Unwrap()))

// Read files (uses core.FS)
data, _ := fs.ReadFile("config.yaml")

// Fetch updates
repo.Fetch(ctx, git.FetchOptions{})
```

### Pattern 3: Testing with Memory FS
```go
func TestCommand(t *testing.T) {
    fs := billy.NewMemory()
    fs.MkdirAll(".sow/project", 0755)
    fs.WriteFile(".sow/project/state.yaml", testData, 0644)

    cmd := NewCommand(fs)
    err := cmd.Execute()

    assert.NoError(t, err)
}
```

---

## Where Libraries Are Used

### CUE Library
- `internal/schema/` - Schema embedding, validation, encoding/decoding
- `internal/commands/schema.go` - Schema inspection commands
- `internal/commands/validate.go` - File validation

### fs/core Library
- `pkg/sowfs/` - Filesystem utilities and path management
- `internal/context/` - Context detection (task vs project)
- All commands - Filesystem abstraction layer

### fs/billy Library
- `cmd/sow/main.go` - Main filesystem instance
- `internal/schema/` - Memory filesystem for embedded schemas
- Tests - Memory filesystem for unit tests

### git Library
- `internal/refs/` - Refs management (clone, update, status)
- Commands: `refs add`, `refs update`, `refs status`

---

## Related Documentation

### Project Documentation
- [sow Architecture](../docs/ARCHITECTURE.md)
- [CLI Reference](../docs/CLI_REFERENCE.md)
- [ADR 001: Go CLI Architecture](../.sow/knowledge/adrs/001-go-cli-architecture.md)
- [ADR 002: Build and Distribution](../.sow/knowledge/adrs/002-build-and-distribution-strategy.md)

### External Documentation
- [CUE Language](https://cuelang.org/docs/)
- [go-git](https://github.com/go-git/go-git)
- [go-billy](https://github.com/go-git/go-billy)
- [Cobra CLI](https://github.com/spf13/cobra)

---

## Reading Order

**For Implementation**:
1. Start with [00_OVERVIEW.md](./00_OVERVIEW.md) - Understand how libraries work together
2. Read [02_fs_core_library.md](./02_fs_core_library.md) - Core abstraction
3. Read [03_fs_billy_library.md](./03_fs_billy_library.md) - Implementation details
4. Read [01_cue_library.md](./01_cue_library.md) - Schema validation
5. Read [04_git_library.md](./04_git_library.md) - Git operations

**For Quick Reference**:
- Jump to specific library reference based on what you're implementing
- Use code examples in "Application in sow CLI" sections

---

## Notes

- All code examples are based on actual library APIs from `/Users/josh/code/go`
- Examples show real-world usage patterns for the sow CLI
- Performance characteristics are measured values from the libraries
- Error handling follows platform error conventions

---

**Last Updated**: 2025-10-15
