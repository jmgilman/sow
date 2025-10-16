# SowFS Package

**Single source of truth for .sow directory structure and operations.**

This package provides domain-specific filesystem abstractions that enforce the `.sow` directory layout and provide type-safe, validated access to all state files.

## Package Structure (833 lines)

```
internal/sowfs/
├── doc.go         - Package documentation
├── sowfs.go       - Main SowFS interface + constructors
├── knowledge.go   - KnowledgeFS interface + impl
├── refs.go        - RefsFS interface + impl
├── project.go     - ProjectFS interface + impl
└── task.go        - TaskFS interface + impl
```

## Interface Hierarchy

```
SowFS (main entry point)
├── Knowledge() → KnowledgeFS
├── Refs() → RefsFS
└── Project() → ProjectFS
                └── Task(id) → TaskFS
```

## Domain Interfaces

### SowFS - Main Entry Point

```go
type SowFS interface {
    Knowledge() *KnowledgeFS
    Refs() *RefsFS
    Project() (*ProjectFS, error)  // Error if no project
    RepoRoot() string
    Close() error
}
```

**Constructors** (return concrete types):
- `NewSowFS() (*SowFSImpl, error)` - From current directory
- `NewSowFSFromPath(path string) (*SowFSImpl, error)` - From specific path
- `NewSowFSWithFS(fs core.FS, repoRoot string) (*SowFSImpl, error)` - With custom fs (testing)

**Validates**:
- ✅ In git repository (walks up dirs looking for `.git`)
- ✅ `.sow` directory exists at repo root
- ✅ Chroots filesystem to `.sow/`

---

### KnowledgeFS - Repository Documentation

```go
type KnowledgeFS interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte) error
    Exists(path string) (bool, error)
    ListADRs() ([]string, error)
    ReadADR(filename string) ([]byte, error)
    WriteADR(filename string, data []byte) error
    MkdirAll(path string) error
}
```

**Manages**:
- `.sow/knowledge/` directory
- ADRs in `.sow/knowledge/adrs/`
- Architecture docs
- All paths relative to `.sow/knowledge/`

---

### RefsFS - External References

```go
type RefsFS interface {
    CommittedIndex() (*schemas.RefsCommittedIndex, error)
    LocalIndex() (*schemas.RefsLocalIndex, error)
    WriteCommittedIndex(*schemas.RefsCommittedIndex) error
    WriteLocalIndex(*schemas.RefsLocalIndex) error

    SymlinkExists(name string) (bool, error)
    CreateSymlink(target, name string) error
    RemoveSymlink(name string) error
    ListSymlinks() ([]string, error)
    ReadSymlink(name string) (string, error)
}
```

**Manages**:
- `.sow/refs/index.json` (committed, validated)
- `.sow/refs/index.local.json` (local-only, validated)
- Symlinks in `.sow/refs/`
- All index files validated against CUE schemas

---

### ProjectFS - Active Project

```go
type ProjectFS interface {
    State() (*schemas.ProjectState, error)
    WriteState(*schemas.ProjectState) error

    AppendLog(entry string) error
    ReadLog() (string, error)

    Task(taskID string) (*TaskFSImpl, error)
    Tasks() ([]*TaskFSImpl, error)

    Exists() (bool, error)

    ReadContext(path string) ([]byte, error)
    WriteContext(path string, data []byte) error
    ListContextFiles() ([]string, error)
}
```

**Manages**:
- `.sow/project/state.yaml` (validated)
- `.sow/project/log.md`
- `.sow/project/context/`
- Task access via `.sow/project/phases/implementation/tasks/`

---

### TaskFS - Individual Task

```go
type TaskFS interface {
    ID() string

    State() (*schemas.TaskState, error)
    WriteState(*schemas.TaskState) error

    AppendLog(entry string) error
    ReadLog() (string, error)

    ReadDescription() (string, error)
    WriteDescription(content string) error

    ListFeedback() ([]string, error)
    ReadFeedback(filename string) (string, error)
    WriteFeedback(filename string, content string) error

    Path() string
}
```

**Manages**:
- `.sow/project/phases/implementation/tasks/{id}/state.yaml` (validated)
- `.sow/project/phases/implementation/tasks/{id}/log.md`
- `.sow/project/phases/implementation/tasks/{id}/description.md`
- `.sow/project/phases/implementation/tasks/{id}/feedback/`

## Standard Errors

```go
var (
    ErrNotInGitRepo      = errors.New("not in git repository")
    ErrSowNotInitialized = errors.New(".sow directory not found")
    ErrProjectNotFound   = errors.New("no active project found")
    ErrTaskNotFound      = errors.New("task not found")
    ErrInvalidTaskID     = errors.New("invalid task ID format")
)
```

## Usage Pattern

```go
// In commands
func runCommand(cmd *cobra.Command) error {
    // Create SowFS
    sowFS, err := sowfs.NewSowFS()
    if err != nil {
        return err
    }
    defer sowFS.Close()

    // Access domain
    refsFS := sowFS.Refs()

    // Read (validated, typed)
    index, err := refsFS.CommittedIndex()

    // Modify
    index.Refs = append(index.Refs, newRef)

    // Write (validated before writing)
    return refsFS.WriteCommittedIndex(index)
}
```

## Testing Pattern

```go
func TestSomething(t *testing.T) {
    // Create in-memory filesystem
    fs := billy.NewMemory()
    fs.MkdirAll(".sow/refs", 0755)
    fs.WriteFile(".sow/refs/index.json", testData, 0644)

    // Create SowFS with test filesystem
    sowFS, err := sowfs.NewSowFSWithFS(fs, "/test/repo")
    require.NoError(t, err)

    // Test operations
    refsFS := sowFS.Refs()
    index, err := refsFS.CommittedIndex()

    assert.NoError(t, err)
    assert.Len(t, index.Refs, 0)
}
```

## Implementation Status

**Interfaces**: ✅ Complete (833 lines)
- All domain interfaces defined
- Comprehensive documentation
- Method signatures finalized
- Error types defined

**Implementation**: ⏳ Next step
- Concrete implementations needed
- Schema validation integration needed
- YAML/JSON marshaling needed
- Filesystem operations needed

## Key Design Decisions

1. **Constructors return concrete types** - Following Go idioms
2. **Domain-specific interfaces** - Clear separation of concerns
3. **Chrooted filesystem** - All paths relative to `.sow/`
4. **Validated I/O** - Schema validation on read/write
5. **Typed structs** - Return `schemas.*` types, not raw bytes
6. **Error wrapping** - Standard errors for common cases

## Next Steps

1. Implement `NewSowFS()` constructors
2. Implement domain-specific `Impl` types
3. Integrate CUE validation (load schemas, validate on read/write)
4. Add comprehensive tests with memory filesystem
5. Update commands to use SowFS instead of raw filesystem

## Benefits

✅ **Single source of truth** - All `.sow` structure knowledge centralized
✅ **Easy refactoring** - Change structure? Update sowfs package only
✅ **Type safety** - Validated schemas, typed structs
✅ **Testable** - Interface-based, works with memory filesystem
✅ **Discoverable** - Domain methods clearly organized
✅ **Encapsulation** - Implementation details hidden from commands
