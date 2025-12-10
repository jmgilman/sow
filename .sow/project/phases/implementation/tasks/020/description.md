# Task 020: Implement Backend Interface and YAMLBackend

## Context

This task is part of the `libs/project` module consolidation effort. The key architectural change in this consolidation is introducing a `Backend` interface that abstracts storage operations, decoupling project state from filesystem specifics.

Currently, the `Project` type in `cli/internal/sdks/project/state/` directly uses `sow.Context` for filesystem access, tightly coupling it to the CLI. The new `Backend` interface allows:
- Pluggable storage backends (YAML files, databases, in-memory for testing)
- Decoupling from CLI-specific types
- Cleaner testability

The design document at `.sow/knowledge/designs/sdk-consolidation-design.md` specifies the Backend interface and YAMLBackend implementation in detail.

## Requirements

### 1. Create Backend Interface (state/backend.go)

Define the storage backend interface following the design document:

```go
// Backend defines the interface for project state persistence.
// Implementations handle reading and writing project state to various
// storage systems (files, databases, remote APIs, etc.)
type Backend interface {
    // Load reads project state from storage.
    // Returns the raw ProjectState (CUE-generated type).
    // The caller is responsible for wrapping this in a Project with runtime fields.
    Load(ctx context.Context) (*project.ProjectState, error)

    // Save writes project state to storage.
    // Takes the raw ProjectState (CUE-generated type).
    // Implementation should handle atomic writes where possible.
    Save(ctx context.Context, state *project.ProjectState) error

    // Exists checks if a project exists in storage.
    Exists(ctx context.Context) (bool, error)

    // Delete removes project state from storage.
    Delete(ctx context.Context) error
}
```

### 2. Implement YAMLBackend (state/backend_yaml.go)

Create the default YAML file backend following the design document:

```go
// YAMLBackend implements Backend using YAML files on a core.FS filesystem.
type YAMLBackend struct {
    fs   core.FS
    path string  // Relative path within fs (default: "project/state.yaml")
}

// NewYAMLBackend creates a backend that stores state in YAML files.
// The fs parameter should be rooted at the .sow directory.
// Uses the default path "project/state.yaml".
func NewYAMLBackend(fs core.FS) *YAMLBackend

// NewYAMLBackendWithPath creates a backend with a custom file path.
// This is useful for testing or non-standard configurations.
func NewYAMLBackendWithPath(fs core.FS, path string) *YAMLBackend

// Load reads project state from the YAML file.
func (b *YAMLBackend) Load(ctx context.Context) (*project.ProjectState, error)

// Save writes project state to the YAML file atomically.
// Uses temp file + rename pattern to ensure atomic writes.
func (b *YAMLBackend) Save(ctx context.Context, state *project.ProjectState) error

// Exists checks if the project state file exists.
func (b *YAMLBackend) Exists(ctx context.Context) (bool, error)

// Delete removes the project state file.
func (b *YAMLBackend) Delete(ctx context.Context) error
```

### 3. Implementation Requirements

**Load operation:**
- Read file contents using `fs.ReadFile(path)`
- Unmarshal YAML into `project.ProjectState`
- Return appropriate errors for missing file, invalid YAML, etc.
- Wrap errors with context using `fmt.Errorf("...: %w", err)`

**Save operation:**
- Marshal `ProjectState` to YAML
- Use atomic write pattern: write to temp file, then rename
- Temp file path: `path + ".tmp"`
- Clean up temp file on rename failure
- Use file mode 0644

**Exists operation:**
- Use `fs.Stat(path)` to check existence
- Return `(false, nil)` for `fs.ErrNotExist`
- Return other errors as-is

**Delete operation:**
- Use `fs.Remove(path)` to delete the file
- Handle non-existent file gracefully (consider if this should be an error or silent)

### 4. Error Handling

Define backend-specific errors in `state/errors.go`:

```go
var (
    // ErrNotFound indicates the project state does not exist in storage.
    ErrNotFound = errors.New("project state not found")

    // ErrInvalidState indicates the stored state is invalid or corrupted.
    ErrInvalidState = errors.New("invalid project state")
)
```

The YAMLBackend.Load should:
- Return `ErrNotFound` wrapped with context when file doesn't exist
- Return `ErrInvalidState` wrapped with context for YAML parse errors

## Acceptance Criteria

1. [ ] `state/backend.go` defines the Backend interface with all methods documented
2. [ ] `state/backend_yaml.go` implements YAMLBackend with all methods
3. [ ] `state/errors.go` defines ErrNotFound and ErrInvalidState errors
4. [ ] NewYAMLBackend and NewYAMLBackendWithPath constructors work correctly
5. [ ] Load correctly reads and parses YAML files
6. [ ] Save uses atomic write pattern (temp file + rename)
7. [ ] Exists returns correct results for existing and non-existing files
8. [ ] Delete removes files correctly
9. [ ] All error cases return properly wrapped errors with `%w`
10. [ ] `golangci-lint run ./...` passes with no issues
11. [ ] `go test -race ./...` passes with no failures
12. [ ] Code adheres to `.standards/STYLE.md` (file organization, naming, error handling)
13. [ ] Tests adhere to `.standards/TESTING.md` (table-driven, testify assertions, behavioral coverage)

### Test Requirements (TDD)

Write comprehensive tests in `state/backend_yaml_test.go`:

**Load tests:**
- Load existing valid project state
- Load returns ErrNotFound for missing file
- Load returns ErrInvalidState for invalid YAML
- Load correctly parses all ProjectState fields

**Save tests:**
- Save creates new file with correct content
- Save overwrites existing file
- Save uses atomic write (verify temp file cleanup)
- Save preserves all ProjectState fields correctly

**Exists tests:**
- Exists returns true for existing file
- Exists returns false for missing file
- Exists returns error for permission issues (if testable)

**Delete tests:**
- Delete removes existing file
- Delete handles non-existent file appropriately

Use `github.com/jmgilman/go/fs/billy` with `billy.NewMemory()` for in-memory filesystem in tests.

## Technical Details

### Import Dependencies

```go
import (
    "context"
    "errors"
    "fmt"
    "io/fs"

    "github.com/jmgilman/go/fs/core"
    "github.com/jmgilman/sow/libs/schemas/project"
    "gopkg.in/yaml.v3"
)
```

### Atomic Write Pattern

The Save operation must use atomic writes to prevent data corruption if the process is interrupted:

```go
func (b *YAMLBackend) Save(ctx context.Context, state *project.ProjectState) error {
    data, err := yaml.Marshal(state)
    if err != nil {
        return fmt.Errorf("marshal project state: %w", err)
    }

    tmpPath := b.path + ".tmp"
    if err := b.fs.WriteFile(tmpPath, data, 0644); err != nil {
        return fmt.Errorf("write temp file: %w", err)
    }

    if err := b.fs.Rename(tmpPath, b.path); err != nil {
        _ = b.fs.Remove(tmpPath) // Clean up on failure
        return fmt.Errorf("rename temp file: %w", err)
    }

    return nil
}
```

### File Path Default

The default path "project/state.yaml" matches the current behavior where project state lives at `.sow/project/state.yaml`.

## Relevant Inputs

- `.sow/knowledge/designs/sdk-consolidation-design.md` - Backend interface design (lines 145-253)
- `cli/internal/sdks/project/state/loader.go` - Current Save implementation (lines 81-126)
- `libs/schemas/project/cue_types_gen.go` - ProjectState type definition
- `libs/config/repo.go` - Example of FS-based loading with error handling
- `.standards/STYLE.md` - Code style requirements
- `.standards/TESTING.md` - Testing requirements

## Examples

### Creating and Using YAMLBackend

```go
// Production use with real filesystem
fs := billy.NewLocal(".sow")
backend := state.NewYAMLBackend(fs)

// Check if project exists
exists, err := backend.Exists(ctx)
if err != nil {
    return fmt.Errorf("check project: %w", err)
}

if exists {
    // Load existing project
    projectState, err := backend.Load(ctx)
    if err != nil {
        return fmt.Errorf("load project: %w", err)
    }
    // Use projectState...
}
```

### Test Example

```go
func TestYAMLBackend_Load(t *testing.T) {
    tests := []struct {
        name      string
        setup     func(fs core.FS)
        wantErr   error
        wantState *project.ProjectState
    }{
        {
            name: "loads valid project",
            setup: func(fs core.FS) {
                data := `name: test-project
type: standard
branch: feat/test`
                fs.WriteFile("project/state.yaml", []byte(data), 0644)
            },
            wantState: &project.ProjectState{
                Name:   "test-project",
                Type:   "standard",
                Branch: "feat/test",
            },
        },
        {
            name: "returns error for missing file",
            setup: func(fs core.FS) {
                // No file created
            },
            wantErr: state.ErrNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fs := billy.NewMemory()
            fs.MkdirAll("project", 0755)
            if tt.setup != nil {
                tt.setup(fs)
            }

            backend := state.NewYAMLBackend(fs)
            got, err := backend.Load(context.Background())

            if tt.wantErr != nil {
                require.Error(t, err)
                assert.True(t, errors.Is(err, tt.wantErr))
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.wantState.Name, got.Name)
            assert.Equal(t, tt.wantState.Type, got.Type)
        })
    }
}
```

## Dependencies

- Task 010: Module foundation must exist with go.mod

## Constraints

- Do NOT import anything from `cli/internal/sow` - the backend must be independent
- Do NOT add CUE validation to the backend - that happens in the loader (Task 070)
- Use `core.FS` interface, not concrete filesystem types
- Do NOT implement the MemoryBackend here - that's Task 030
- Follow YAML field naming from `libs/schemas/project/cue_types_gen.go` (uses json tags but yaml.v3 respects them)
