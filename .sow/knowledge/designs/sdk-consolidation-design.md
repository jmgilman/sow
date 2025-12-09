# Project SDK Consolidation Design

**Status**: Draft
**Author**: Orchestrator
**Date**: 2024-12-09
**Branch**: `design/project-sdk-lib`

## Overview

This document describes the consolidation of the project and state SDKs into a single, cohesive Go module at `libs/project`. The goal is to clean up internal code that has accumulated technical debt through multiple refactors while **preserving the existing public API** so that all dependents (primarily `cli/internal/projects/*`) continue to work with only import path changes.

This work is part of a broader `libs/` refactoring effort documented in `.sow/knowledge/explorations/libs-refactoring-summary.md`.

## Problem Statement

### Current State

The project and state SDKs currently exist as two separate packages within `cli/internal/sdks/`:

```
cli/internal/sdks/
├── state/           # Generic state machine framework (~5,000 LOC)
│   ├── machine.go   # State machine wrapper
│   ├── builder.go   # Fluent API for building machines
│   ├── states.go    # Base State type
│   └── events.go    # Event type and helpers
│
└── project/         # Project-specific SDK (~5,800 LOC)
    ├── builder.go   # ProjectTypeConfigBuilder
    ├── config.go    # ProjectTypeConfig
    ├── machine.go   # BuildMachine() with closure binding
    ├── options.go   # PhaseOpt, TransitionOption builders
    ├── types.go     # GuardTemplate, Action types
    ├── branch.go    # BranchConfig, BranchPath
    └── state/       # Persistence layer (~4,800 LOC)
        ├── project.go      # Project wrapper type
        ├── phase.go        # Phase type
        ├── artifact.go     # Artifact type
        ├── task.go         # Task type
        ├── collections.go  # Collection types
        ├── loader.go       # Load()/Save() persistence
        └── registry.go     # Project type registry
```

### Issues Identified

1. **Architectural Debt**: Multiple refactors have left the code with awkward patterns:
   - Interface-based `ProjectTypeConfig` to avoid import cycles
   - Closure wrappers to bind guards to project instances
   - String-based discriminator values in branches (not type-safe)

2. **Unclear Boundaries**: The separation between `state/` (generic) and `project/` (specific) is muddied:
   - The `project/state/` subpackage creates confusing nesting
   - Some "project" concepts leak into the generic state machine

3. **Coupling Patterns**:
   - CLI must explicitly import all project types via blank imports
   - Registry uses global map pattern
   - Metadata validation via embedded strings

4. **Tight Persistence Coupling**: The `Project` type is directly coupled to YAML file storage via `sow.Context`, making it impossible to use alternative storage backends.

5. **Discoverability**: Code spread across nested packages makes it hard to understand the full API surface.

## Goals

### Primary Goals

1. Consolidate into a single `libs/project` module with clean internal structure while maintaining API compatibility
2. Introduce a storage backend abstraction to separate persistent state from in-memory state

### Success Criteria

- All project type implementations (`cli/internal/projects/*`) continue to work with only import path changes
- No changes to public API signatures
- Cleaner internal organization
- Single import path for consumers
- Storage backend abstraction enables future backend implementations without API changes

### Non-Goals

- Adding new features (beyond storage abstraction)
- Breaking API changes
- Plugin/dynamic loading system
- Performance optimization
- Implementing alternative storage backends (future work)

## Design Decisions

### Resolved Questions

1. **Module Path**: `libs/project` (not `libs/projectsdk`)
2. **State Subpackage**: Keep `state/` as a subpackage (complexity warrants separation)
3. **CUE Schemas**: Will move to `libs/schemas` as part of broader refactoring (separate effort)

## Proposed Structure

### New Module Location

```
libs/project/
├── go.mod                    # Separate Go module
├── go.sum
├── doc.go                    # Package documentation
│
├── builder.go                # ProjectTypeConfigBuilder (fluent API)
├── config.go                 # ProjectTypeConfig type
├── machine.go                # State machine wrapper and binding
├── options.go                # PhaseOpt, TransitionOption, BranchOption
├── types.go                  # Core types: State, Event, Guard, Action
├── branch.go                 # BranchConfig, BranchPath
├── registry.go               # Project type registry
├── runtime.go                # BuildMachine, FireWithPhaseUpdates
│
├── state/                    # Project state types (persistence layer)
│   ├── doc.go                # Subpackage documentation
│   ├── project.go            # Project wrapper type (in-memory representation)
│   ├── phase.go              # Phase type and helpers
│   ├── artifact.go           # Artifact type
│   ├── task.go               # Task type
│   ├── collections.go        # PhaseCollection, ArtifactCollection, TaskCollection
│   ├── convert.go            # CUE → wrapper conversion
│   ├── validate.go           # Structural validation
│   │
│   ├── backend.go            # Backend interface definition
│   ├── backend_yaml.go       # YAML file backend (default)
│   └── loader.go             # Load()/Save() using backend
│
└── internal/                 # Private implementation details
    └── stateless/            # State machine internals (if needed)
```

### Key Changes

1. **Flatten the Hierarchy**: Eliminate `project/state/` nesting by promoting persistence types to `state/` subpackage directly under module root.

2. **Merge State Machine**: Absorb the generic state machine code into the project SDK rather than maintaining it as a separate package. The state machine is only used for projects anyway.

3. **Unified Types File**: Consolidate `State`, `Event`, `Guard`, `Action`, and related types into a single `types.go`.

4. **Internal Package**: Move implementation details that don't need to be exported into `internal/`.

5. **Storage Backend Abstraction**: Introduce a `Backend` interface that decouples state persistence from the `Project` type.

## Storage Backend Abstraction

### Motivation

Currently, the `Project` type holds a `*sow.Context` and directly implements YAML file I/O in `Save()` and `Load()`. This creates several problems:

1. **Testing Difficulty**: Tests must set up filesystem contexts
2. **Tight Coupling**: Cannot use projects without a filesystem
3. **No Extensibility**: Cannot easily add database, remote, or in-memory backends
4. **Separation of Concerns**: Business logic mixed with I/O

### Design

Separate project state into two concerns:

1. **In-Memory State**: The `Project` struct with runtime fields (machine, config)
2. **Persistent State**: Handled by a `Backend` interface

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

### YAML File Backend (Default)

The default implementation preserves current behavior:

```go
// YAMLBackend implements Backend using YAML files on a core.FS filesystem.
type YAMLBackend struct {
    fs   core.FS
    path string  // Relative path within fs (default: "project/state.yaml")
}

// NewYAMLBackend creates a backend that stores state in YAML files.
func NewYAMLBackend(fs core.FS) *YAMLBackend {
    return &YAMLBackend{
        fs:   fs,
        path: "project/state.yaml",
    }
}

func (b *YAMLBackend) Load(ctx context.Context) (*project.ProjectState, error) {
    data, err := b.fs.ReadFile(b.path)
    if err != nil {
        return nil, fmt.Errorf("failed to read state: %w", err)
    }

    var state project.ProjectState
    if err := yaml.Unmarshal(data, &state); err != nil {
        return nil, fmt.Errorf("failed to unmarshal: %w", err)
    }

    return &state, nil
}

func (b *YAMLBackend) Save(ctx context.Context, state *project.ProjectState) error {
    data, err := yaml.Marshal(state)
    if err != nil {
        return fmt.Errorf("failed to marshal: %w", err)
    }

    // Atomic write: temp file + rename
    tmpPath := b.path + ".tmp"
    if err := b.fs.WriteFile(tmpPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }

    if err := b.fs.Rename(tmpPath, b.path); err != nil {
        _ = b.fs.Remove(tmpPath)
        return fmt.Errorf("failed to rename temp file: %w", err)
    }

    return nil
}

func (b *YAMLBackend) Exists(ctx context.Context) (bool, error) {
    _, err := b.fs.Stat(b.path)
    if err != nil {
        if os.IsNotExist(err) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

func (b *YAMLBackend) Delete(ctx context.Context) error {
    return b.fs.Remove(b.path)
}
```

### Updated Project Type

The `Project` type no longer holds storage details:

```go
// Project wraps the CUE-generated ProjectState with runtime behavior.
type Project struct {
    project.ProjectState  // Embedded for serialization

    // Runtime-only fields (not serialized)
    config  ProjectTypeConfig
    machine *Machine
    backend Backend  // Storage backend (nil for transient projects)
}

// Save persists the project state using the configured backend.
func (p *Project) Save(ctx context.Context) error {
    if p.backend == nil {
        return fmt.Errorf("no backend configured")
    }

    // Sync statechart state from machine
    if p.machine != nil {
        p.Statechart.Current_state = p.machine.State().String()
        p.Statechart.Updated_at = time.Now()
    }

    // Update timestamp
    p.Updated_at = time.Now()

    // Validate before saving
    if err := validateStructure(p.ProjectState); err != nil {
        return fmt.Errorf("CUE validation failed: %w", err)
    }
    if p.config != nil {
        if err := p.config.Validate(p); err != nil {
            return fmt.Errorf("metadata validation failed: %w", err)
        }
    }

    // Delegate to backend
    return p.backend.Save(ctx, &p.ProjectState)
}
```

### Updated Loader Functions

```go
// Load reads project state using the provided backend.
func Load(ctx context.Context, backend Backend) (*Project, error) {
    // 1. Load raw state from backend
    projectState, err := backend.Load(ctx)
    if err != nil {
        return nil, err
    }

    // 2. Validate structure with CUE
    if err := validateStructure(*projectState); err != nil {
        return nil, fmt.Errorf("CUE validation failed: %w", err)
    }

    // 3. Wrap in Project type
    proj := &Project{
        ProjectState: *projectState,
        backend:      backend,
    }

    // 4. Lookup and attach type config
    config, exists := Registry[proj.Type]
    if !exists {
        return nil, fmt.Errorf("unknown project type: %s", proj.Type)
    }
    proj.config = config

    // 5. Build state machine
    initialState := State(proj.Statechart.Current_state)
    proj.machine = config.BuildMachine(proj, initialState)

    // 6. Validate metadata
    if err := config.Validate(proj); err != nil {
        return nil, fmt.Errorf("metadata validation failed: %w", err)
    }

    return proj, nil
}

// LoadFromFS is a convenience function that creates a YAML backend.
// This preserves the current API for CLI usage.
func LoadFromFS(ctx context.Context, fs core.FS) (*Project, error) {
    backend := NewYAMLBackend(fs)
    return Load(ctx, backend)
}

// Create initializes a new project with the given backend.
func Create(ctx context.Context, backend Backend, branch, description string, initialInputs map[string][]project.ArtifactState) (*Project, error) {
    // ... (same logic as current, but uses backend for Save)
}

// CreateOnFS is a convenience function that creates a YAML backend.
func CreateOnFS(ctx context.Context, fs core.FS, branch, description string, initialInputs map[string][]project.ArtifactState) (*Project, error) {
    backend := NewYAMLBackend(fs)
    return Create(ctx, backend, branch, description, initialInputs)
}
```

### API Compatibility

The `*FromFS` convenience functions maintain backward compatibility:

```go
// Current usage (CLI)
project, err := state.Load(ctx)  // ctx has FS

// New usage (same behavior)
project, err := state.LoadFromFS(ctx, fs)

// New usage (explicit backend)
backend := state.NewYAMLBackend(fs)
project, err := state.Load(ctx, backend)
```

Consumers continue using the same patterns. The only change is passing `fs` explicitly instead of via `sow.Context`.

### Future Backend Examples

The abstraction enables future backends without API changes:

```go
// In-memory backend for testing
type MemoryBackend struct {
    state *project.ProjectState
}

// SQLite backend for local persistence
type SQLiteBackend struct {
    db   *sql.DB
    path string
}

// Remote API backend
type APIBackend struct {
    client *http.Client
    url    string
}
```

## API Preservation

### Public API Surface to Preserve

The following public API must remain unchanged (signatures and behavior):

#### Builder API

```go
// ProjectTypeConfigBuilder methods
func (b *ProjectTypeConfigBuilder) WithPhase(name string, opts ...PhaseOpt) *ProjectTypeConfigBuilder
func (b *ProjectTypeConfigBuilder) SetInitialState(state State) *ProjectTypeConfigBuilder
func (b *ProjectTypeConfigBuilder) AddTransition(from, to State, event Event, opts ...TransitionOption) *ProjectTypeConfigBuilder
func (b *ProjectTypeConfigBuilder) AddBranch(from State, opts ...BranchOption) *ProjectTypeConfigBuilder
func (b *ProjectTypeConfigBuilder) OnAdvance(state State, determiner AdvanceDeterminer) *ProjectTypeConfigBuilder
func (b *ProjectTypeConfigBuilder) WithPrompt(state State, generator PromptGenerator) *ProjectTypeConfigBuilder
func (b *ProjectTypeConfigBuilder) WithOrchestratorPrompt(generator OrchestratorPromptGenerator) *ProjectTypeConfigBuilder
func (b *ProjectTypeConfigBuilder) WithInitializer(fn ProjectInitializer) *ProjectTypeConfigBuilder
func (b *ProjectTypeConfigBuilder) Build() (*ProjectTypeConfig, error)
```

#### Option Functions

```go
// Phase options
func WithStartState(state State) PhaseOpt
func WithEndState(state State) PhaseOpt
func WithInputs(types ...string) PhaseOpt
func WithOutputs(types ...string) PhaseOpt
func WithTasks() PhaseOpt
func WithMetadataSchema(schema string) PhaseOpt

// Transition options
func WithGuard(description string, fn func(*state.Project) bool) TransitionOption
func WithOnEntry(fn func(*state.Project) error) TransitionOption
func WithOnExit(fn func(*state.Project) error) TransitionOption
func WithDescription(desc string) TransitionOption
func WithFailedPhase(name string) TransitionOption

// Branch options
func BranchOn(discriminator func(*state.Project) string) BranchOption
func When(value string, event Event, toState State, opts ...TransitionOption) BranchOption
```

#### State Types (state subpackage)

```go
// Core types
type Project struct { ... }
type Phase struct { ... }
type Artifact struct { ... }
type Task struct { ... }

// Collections
type PhaseCollection []Phase
type ArtifactCollection []Artifact
type TaskCollection []Task

// Backend interface
type Backend interface { ... }

// Loader functions
func Load(ctx context.Context, backend Backend) (*Project, error)
func LoadFromFS(ctx context.Context, fs core.FS) (*Project, error)  // Convenience
func (p *Project) Save(ctx context.Context) error

// Registry
func Register(name string, config *ProjectTypeConfig)
func Get(name string) (*ProjectTypeConfig, bool)
func List() []string
```

### Import Path Changes

Consumers will need to update imports:

```go
// Before
import (
    "github.com/jmgilman/sow/cli/internal/sdks/project"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/jmgilman/sow/cli/internal/sdks/state"  // if used directly
)

// After
import (
    "github.com/jmgilman/sow/libs/project"
    "github.com/jmgilman/sow/libs/project/state"
)
```

## Internal Cleanup Opportunities

While preserving the public API, the following internal improvements can be made:

### 1. Eliminate Interface Indirection

The current `ProjectTypeConfig` is an interface to avoid import cycles. With proper package structure, this can become a concrete struct:

```go
// Current (interface to avoid cycles)
type ProjectTypeConfig interface {
    Name() string
    BuildMachine(*state.Project) (*Machine, error)
    // ...
}

// After (concrete struct, cycles avoided by structure)
type ProjectTypeConfig struct {
    name         string
    phases       []PhaseConfig
    transitions  []TransitionConfig
    // ...
}
```

### 2. Simplify Guard Binding

The current closure-based guard binding can be simplified:

```go
// Current: Template with closure wrapper
type GuardTemplate struct {
    Description string
    Check       func(*state.Project) bool
}

// In BuildMachine(), wraps to func() bool via closure

// After: Direct binding in machine construction
// (internal simplification, same external behavior)
```

### 3. Consolidate Type Definitions

Currently spread across multiple files:
- `sdks/state/states.go` - State type
- `sdks/state/events.go` - Event type
- `sdks/project/types.go` - Guard, Action types

Consolidate into single `types.go`:

```go
package project

// State represents a state machine state
type State string

// Event represents a state transition trigger
type Event string

// Guard is a condition that must be true for a transition
type Guard struct {
    Description string
    Check       func(*state.Project) bool
}

// Action is executed during state transitions
type Action func(*state.Project) error

// NoProject is the initial state before any project exists
const NoProject State = "NoProject"
```

### 4. Clean Up Registry

Move from global map to more structured approach:

```go
// Current: Global map with init() registration
var registry = make(map[string]*ProjectTypeConfig)

func init() {
    Register("standard", NewStandardProjectConfig())
}

// After: Same public API, cleaner internal implementation
// with proper synchronization and error handling
```

## Migration Plan

### Phase 1: Create New Module

1. Create `libs/project/` directory
2. Initialize `go.mod` with module path `github.com/jmgilman/sow/libs/project`
3. Copy and reorganize code from `cli/internal/sdks/`
4. Implement `Backend` interface and `YAMLBackend`
5. Add `LoadFromFS` and `CreateOnFS` convenience functions
6. Ensure all tests pass

### Phase 2: Update Consumers

1. Update `cli/internal/projects/*` to use new imports
2. Update CLI commands to use new imports
3. Update calls from `Load(ctx)` to `LoadFromFS(ctx, ctx.FS())`
4. Run full test suite
5. Verify all project types work correctly

### Phase 3: Cleanup

1. Remove old `cli/internal/sdks/` packages
2. Update documentation
3. Update any references in CLAUDE.md or other docs

## Testing Strategy

### Compatibility Testing

1. **API Compatibility**: Compile all existing consumers against new module without code changes (only import changes)
2. **Behavior Compatibility**: Run existing integration tests for all project types
3. **Regression Suite**: Existing tests in `sdks/project/` and `sdks/state/` should pass

### Backend Testing

1. **Interface Compliance**: Test that `YAMLBackend` correctly implements `Backend`
2. **Round-Trip**: Load → modify → save → load should preserve all data
3. **Atomic Writes**: Verify temp file + rename pattern
4. **Error Cases**: Test missing files, invalid YAML, permission errors

### New Tests

1. **Module Isolation**: Test that the module can be imported independently
2. **Cross-Package**: Test interactions between `project` and `project/state`
3. **Memory Backend**: Add simple in-memory backend for unit tests

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Hidden API usage | Consumers break | Thorough grep for all import paths |
| Subtle behavior change | Runtime errors | Run full integration test suite |
| CUE schema coupling | Schema changes break | Coordinate with `libs/schemas` migration |
| Go module versioning | Import errors | Start at v0 to allow iteration |
| Backend abstraction overhead | Performance regression | Keep abstraction thin; benchmark if needed |

## Appendix: Current Dependencies

### Packages that import `sdks/project`

```
cli/internal/projects/breakdown
cli/internal/projects/design
cli/internal/projects/exploration
cli/internal/projects/standard
cli/cmd/project/advance.go
cli/cmd/project/wizard.go
cli/cmd/agent/spawn.go
```

### Packages that import `sdks/project/state`

```
cli/internal/projects/breakdown
cli/internal/projects/design
cli/internal/projects/exploration
cli/internal/projects/standard
cli/cmd/project/*.go (most commands)
cli/cmd/task/*.go
cli/cmd/input/*.go
cli/cmd/output/*.go
cli/cmd/agent/*.go
```

### Packages that import `sdks/state`

```
cli/internal/sdks/project (internal dependency only)
```

## Related Documents

- `.sow/knowledge/explorations/libs-refactoring-summary.md` - Broader libs/ refactoring plan
- Task 040: Complete architecture design (comprehensive)
