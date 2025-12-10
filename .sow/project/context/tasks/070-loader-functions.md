# Task 070: Migrate Loader Functions with Backend Integration

## Context

This task is part of the `libs/project` module consolidation effort. It migrates the Load/Create functions from `cli/internal/sdks/project/state/loader.go` to `libs/project/state/` and updates them to use the Backend interface.

The loader functions are the primary entry points for working with project state:
- `Load` - Load existing project from a Backend
- `LoadFromFS` - Convenience wrapper for YAML files
- `Create` - Create new project and save via Backend
- `CreateOnFS` - Convenience wrapper for creating with YAML backend
- `Save` - Save project state via Backend

The key architectural change is replacing direct `sow.Context` usage with the `Backend` interface.

## Requirements

### 1. Migrate Load Function (state/loader.go)

Update Load to use Backend:

```go
// Load reads project state from a backend and returns an initialized Project.
//
// The Load pipeline:
//  1. Load raw ProjectState from backend
//  2. Validate structure with CUE
//  3. Create Project wrapper
//  4. Lookup and attach ProjectTypeConfig from registry
//  5. Build state machine initialized with current state
//  6. Validate metadata against embedded schemas
//
// Returns an error if any step fails. Error messages are actionable and
// indicate which step failed and why.
func Load(ctx context.Context, backend Backend) (*Project, error)
```

### 2. Add LoadFromFS Convenience Function

Provide a convenience wrapper for filesystem-based loading:

```go
// LoadFromFS loads a project from a YAML file on the filesystem.
// This is a convenience wrapper around Load with YAMLBackend.
//
// The fs parameter should be rooted at the .sow directory.
func LoadFromFS(ctx context.Context, fs core.FS) (*Project, error) {
    backend := NewYAMLBackend(fs)
    return Load(ctx, backend)
}
```

### 3. Migrate Create Function

Update Create to use Backend:

```go
// Create initializes a new project and saves it via the backend.
//
// The Create pipeline:
//  1. Validate branch provided
//  2. Detect project type from branch name
//  3. Lookup project type configuration from registry
//  4. Generate project name from description
//  5. Create minimal ProjectState with metadata fields
//  6. Let project type initialize phases and state
//  7. Build state machine with config's initial state
//  8. Set initial phase status to in_progress
//  9. Save to backend
//
// Parameters:
//   - ctx: context for cancellation
//   - backend: storage backend for persistence
//   - branch: branch name (used to detect project type)
//   - description: project description
//   - initialInputs: optional map of phase name to initial input artifacts
//
// Returns a fully initialized Project ready for use.
func Create(
    ctx context.Context,
    backend Backend,
    branch string,
    description string,
    initialInputs map[string][]project.ArtifactState,
) (*Project, error)
```

### 4. Add CreateOnFS Convenience Function

Provide a convenience wrapper for filesystem-based creation:

```go
// CreateOnFS creates a new project with a YAML file backend.
// This is a convenience wrapper around Create with YAMLBackend.
//
// The fs parameter should be rooted at the .sow directory.
func CreateOnFS(
    ctx context.Context,
    fs core.FS,
    branch string,
    description string,
    initialInputs map[string][]project.ArtifactState,
) (*Project, error) {
    backend := NewYAMLBackend(fs)
    return Create(ctx, backend, branch, description, initialInputs)
}
```

### 5. Migrate Save Function

Update Save to be a standalone function using Backend:

```go
// Save validates and writes project state to the backend.
//
// The Save pipeline:
//  1. Sync statechart state from machine (if present)
//  2. Update timestamps
//  3. Validate structure with CUE
//  4. Validate metadata with embedded schemas
//  5. Save to backend
//
// Returns an error if any step fails. Validation errors prevent writing,
// ensuring the state file always contains valid data.
func Save(ctx context.Context, p *Project) error
```

### 6. Helper Functions

Migrate helper functions used by Load/Create:

```go
// detectProjectType determines project type from branch name.
func detectProjectType(branchName string) string

// generateProjectName converts a description to a kebab-case project name.
func generateProjectName(description string) string

// markInitialPhaseInProgress sets the initial phase status to in_progress.
func markInitialPhaseInProgress(p *Project) error
```

### 7. Context Integration

All functions should accept `context.Context`:
- Pass context to Backend methods
- Support cancellation for long operations
- Use context for timeouts if needed

## Acceptance Criteria

1. [ ] `Load` function uses Backend interface
2. [ ] `LoadFromFS` provides convenience for YAML files
3. [ ] `Create` function uses Backend interface
4. [ ] `CreateOnFS` provides convenience for YAML files
5. [ ] `Save` function uses Backend interface
6. [ ] All functions accept `context.Context`
7. [ ] CUE validation is integrated (from Task 080)
8. [ ] Registry lookup is integrated (from Task 080)
9. [ ] Error messages are clear and actionable with proper wrapping (`%w`)
10. [ ] `golangci-lint run ./...` passes with no issues
11. [ ] `go test -race ./...` passes with no failures
12. [ ] Code adheres to `.standards/STYLE.md` (file organization, naming, error handling)
13. [ ] Tests adhere to `.standards/TESTING.md` (table-driven, testify assertions, behavioral coverage)

### Test Requirements (TDD)

Write comprehensive tests using MemoryBackend:

**state/loader_test.go:**

**Load tests:**
- Load succeeds with valid project state
- Load fails for missing project (ErrNotFound)
- Load fails for invalid YAML
- Load fails for unknown project type
- Load attaches correct config from registry
- Load builds state machine correctly
- Load validates metadata

**Create tests:**
- Create succeeds with valid inputs
- Create detects project type from branch
- Create generates valid project name
- Create initializes phases correctly
- Create sets initial phase in_progress
- Create saves to backend
- Create fails for empty branch
- Create fails for unknown project type

**Save tests:**
- Save syncs statechart from machine
- Save updates timestamps
- Save validates structure
- Save validates metadata
- Save persists to backend
- Save fails validation gracefully

**LoadFromFS and CreateOnFS tests:**
- Test with in-memory filesystem (billy.NewMemory)
- Verify they use YAMLBackend correctly

## Technical Details

### Import Dependencies

```go
import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/jmgilman/go/fs/core"
    "github.com/jmgilman/sow/libs/schemas/project"
)
```

### Registry Integration

The Load function needs to look up configs from the Registry (Task 080). Until Task 080 is complete, use a placeholder:

```go
// In loader.go
func Load(ctx context.Context, backend Backend) (*Project, error) {
    // ... load and validate structure ...

    // Lookup config from registry
    config, exists := GetConfig(projectState.Type)
    if !exists {
        return nil, fmt.Errorf("unknown project type: %s", projectState.Type)
    }

    // ... continue with config ...
}
```

### Error Handling Pattern

Use wrapped errors with context:

```go
func Load(ctx context.Context, backend Backend) (*Project, error) {
    projectState, err := backend.Load(ctx)
    if err != nil {
        return nil, fmt.Errorf("load project state: %w", err)
    }

    if err := validateStructure(projectState); err != nil {
        return nil, fmt.Errorf("validate structure: %w", err)
    }

    // ...
}
```

### Branch Detection Logic

Preserve existing branch detection:

```go
func detectProjectType(branchName string) string {
    switch {
    case strings.HasPrefix(branchName, "explore/"):
        return "exploration"
    case strings.HasPrefix(branchName, "design/"):
        return "design"
    case strings.HasPrefix(branchName, "breakdown/"):
        return "breakdown"
    default:
        return "standard"
    }
}
```

### Project Name Generation

Preserve existing name generation:

```go
func generateProjectName(description string) string {
    name := description
    if len(name) > 50 {
        name = name[:50]
    }

    // Convert to kebab-case
    result := ""
    for i, r := range name {
        if r == ' ' || r == '_' {
            if i > 0 && len(result) > 0 && result[len(result)-1] != '-' {
                result += "-"
            }
        } else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
            result += string(r)
        } else if r >= 'A' && r <= 'Z' {
            result += string(r + 32) // Convert to lowercase
        }
    }

    // Remove trailing hyphen
    if len(result) > 0 && result[len(result)-1] == '-' {
        result = result[:len(result)-1]
    }

    return result
}
```

## Relevant Inputs

- `cli/internal/sdks/project/state/loader.go` - Current Load/Create/Save implementation
- `state/backend.go` from Task 020 - Backend interface
- `state/backend_memory.go` from Task 030 - MemoryBackend for testing
- `state/project.go` from Task 040 - Project type
- `state/validate.go` from Task 080 - CUE validation
- `state/registry.go` from Task 080 - Config registry
- `.standards/STYLE.md` - Code style requirements
- `.standards/TESTING.md` - Testing requirements

## Examples

### Loading a Project

```go
// Load from filesystem
fs := billy.NewLocal(".sow")
proj, err := state.LoadFromFS(context.Background(), fs)
if err != nil {
    return fmt.Errorf("load project: %w", err)
}

fmt.Printf("Loaded project: %s (type: %s)\n", proj.Name, proj.Type)
fmt.Printf("Current state: %s\n", proj.Machine().State())
```

### Creating a Project

```go
// Create with memory backend (testing)
backend := state.NewMemoryBackend()
proj, err := state.Create(
    context.Background(),
    backend,
    "feat/auth-implementation",
    "Implement JWT authentication",
    nil, // no initial inputs
)
if err != nil {
    return fmt.Errorf("create project: %w", err)
}

// Or with filesystem
fs := billy.NewLocal(".sow")
proj, err := state.CreateOnFS(
    context.Background(),
    fs,
    "feat/auth-implementation",
    "Implement JWT authentication",
    nil,
)
```

### Saving a Project

```go
// After making changes to project
proj.Phases["planning"].Status = "completed"

if err := state.Save(context.Background(), proj); err != nil {
    return fmt.Errorf("save project: %w", err)
}
```

### Testing with MemoryBackend

```go
func TestProjectWorkflow(t *testing.T) {
    ctx := context.Background()
    backend := state.NewMemoryBackend()

    // Create project
    proj, err := state.Create(ctx, backend, "feat/test", "Test project", nil)
    require.NoError(t, err)
    assert.Equal(t, "test-project", proj.Name)

    // Verify persisted
    loaded, err := state.Load(ctx, backend)
    require.NoError(t, err)
    assert.Equal(t, proj.Name, loaded.Name)
}
```

## Dependencies

- Task 020: Backend interface
- Task 030: MemoryBackend (for testing)
- Task 040: State wrapper types (Project)
- Task 050: State machine (for Machine building)
- Task 060: Project config (for BuildMachine)
- Task 080: Registry and validation (partial - can stub registry initially)

## Constraints

- Do NOT import from `cli/internal/sow` - use Backend interface instead
- Do NOT modify existing project type registrations - they'll be updated in Task 090
- Preserve existing validation behavior (CUE structure + metadata)
- Preserve existing branch detection and name generation logic
- Support both synchronous operations (context can be used for cancellation)
- Keep convenience functions thin wrappers around core functions
