# Issue #118: libs/project module

**URL**: https://github.com/jmgilman/sow/issues/118
**State**: OPEN

## Description

# Work Unit 005: libs/project Module

Create a new `libs/project` Go module consolidating the project and state SDKs with a storage backend abstraction.

## Objective

Consolidate `cli/internal/sdks/state/` and `cli/internal/sdks/project/` into a single `libs/project/` module, implementing the design specified in `.sow/knowledge/designs/sdk-consolidation-design.md`.

**This is NOT just a file move.** This work unit implements significant architectural improvements:
1. Introduce `Backend` interface for storage abstraction
2. Decouple from `sow.Context` - use `core.FS` or `Backend` directly
3. Consolidate two SDK packages into one cohesive module
4. Clean up internal technical debt while preserving public API

## Design Document

**CRITICAL**: This work unit MUST adhere to the design document at:
`.sow/knowledge/designs/sdk-consolidation-design.md`

Key design decisions from that document:
- Storage backend abstraction (`Backend` interface)
- `LoadFromFS` / `CreateOnFS` convenience functions for backward compatibility
- Flatten hierarchy: eliminate `project/state/` nesting
- Merge generic state machine into project SDK
- Consolidate type definitions into single `types.go`

## Scope

### What's Moving
- `cli/internal/sdks/state/` - Generic state machine framework (~5,000 LOC)
- `cli/internal/sdks/project/` - Project-specific SDK (~5,800 LOC)
- `cli/internal/sdks/project/state/` - Persistence layer (~4,800 LOC)

### Target Structure (from design doc)
```
libs/project/
├── go.mod
├── go.sum
├── README.md                    # Per READMES.md standard
├── doc.go                       # Package documentation
│
├── builder.go                   # ProjectTypeConfigBuilder (fluent API)
├── config.go                    # ProjectTypeConfig type
├── machine.go                   # State machine wrapper and binding
├── options.go                   # PhaseOpt, TransitionOption, BranchOption
├── types.go                     # Core types: State, Event, Guard, Action
├── branch.go                    # BranchConfig, BranchPath
├── registry.go                  # Project type registry
├── runtime.go                   # BuildMachine, FireWithPhaseUpdates
│
├── state/                       # Project state types (persistence layer)
│   ├── doc.go                   # Subpackage documentation
│   ├── project.go               # Project wrapper type
│   ├── phase.go                 # Phase type and helpers
│   ├── artifact.go              # Artifact type
│   ├── task.go                  # Task type
│   ├── collections.go           # PhaseCollection, ArtifactCollection, TaskCollection
│   ├── convert.go               # CUE → wrapper conversion
│   ├── validate.go              # Structural validation
│   │
│   ├── backend.go               # Backend interface definition
│   ├── backend_yaml.go          # YAML file backend (default)
│   └── loader.go                # Load()/Save() using backend
│
└── internal/                    # Private implementation details
    └── stateless/               # State machine internals (if needed)
```

## Standards Requirements

### Go Code Standards (STYLE.md)
- Accept interfaces, return concrete types
- Error handling with proper wrapping (`%w`)
- No global mutable state (registry uses proper synchronization)
- Functional options pattern for builders
- Functions under 80 lines
- Proper struct field ordering

### Testing Standards (TESTING.md)
- Behavioral test coverage for all operations
- Table-driven tests with `t.Run()`
- Use `testify/assert` and `testify/require`
- In-memory backend for unit tests
- No external dependencies in unit tests

### README Standards (READMES.md)
- Overview: Project SDK for sow state management
- Quick Start: Load project, basic operations
- Usage: Creating project types, state transitions, persistence
- Architecture: Backend abstraction, registry pattern

### Linting
- Must pass `golangci-lint run` with project's `.golangci.yml`
- Proper error wrapping for external errors

## API Design Requirements

### Storage Backend Abstraction (from design doc)

**Backend Interface:**
```go
type Backend interface {
    Load(ctx context.Context) (*project.ProjectState, error)
    Save(ctx context.Context, state *project.ProjectState) error
    Exists(ctx context.Context) (bool, error)
    Delete(ctx context.Context) error
}
```

**YAML Backend (default implementation):**
```go
type YAMLBackend struct {
    fs   core.FS
    path string
}

func NewYAMLBackend(fs core.FS) *YAMLBackend
```

**Updated Loader Functions:**
```go
// New API with explicit backend
func Load(ctx context.Context, backend Backend) (*Project, error)

// Convenience function for backward compatibility
func LoadFromFS(ctx context.Context, fs core.FS) (*Project, error)
```

### API Preservation

The following public API must remain unchanged (only import paths change):
- `ProjectTypeConfigBuilder` methods
- Phase option functions (`WithStartState`, `WithEndState`, etc.)
- Transition option functions (`WithGuard`, `WithOnEntry`, etc.)
- Branch option functions (`BranchOn`, `When`)
- State types (`Project`, `Phase`, `Artifact`, `Task`)
- Collection types (`PhaseCollection`, `ArtifactCollection`, `TaskCollection`)
- Registry functions (`Register`, `Get`, `List`)

### Internal Cleanup Opportunities (from design doc)

While preserving public API:
1. Eliminate interface indirection in `ProjectTypeConfig` (can be concrete struct)
2. Simplify guard binding (remove closure wrappers where possible)
3. Consolidate type definitions into single `types.go`
4. Clean up registry with proper synchronization

## Consumer Impact

~82 files currently import from SDK packages. After this work:
- Imports change to `github.com/jmgilman/sow/libs/project` and `github.com/jmgilman/sow/libs/project/state`
- `Load(ctx)` calls change to `LoadFromFS(ctx, fs)`
- All project type implementations (`cli/internal/projects/*`) need import updates

## Dependencies

- `libs/schemas/project` - For CUE-generated types (`ProjectState`, `PhaseState`, etc.)

## Acceptance Criteria

1. [ ] New `libs/project` Go module exists and compiles
2. [ ] `Backend` interface implemented per design doc
3. [ ] `YAMLBackend` preserves current file-based behavior
4. [ ] `LoadFromFS` / `CreateOnFS` convenience functions work
5. [ ] All public API preserved (only import paths change)
6. [ ] Internal cleanup completed (concrete types, simplified binding)
7. [ ] In-memory backend available for testing
8. [ ] All tests pass with proper behavioral coverage
9. [ ] `golangci-lint run` passes with no issues
10. [ ] README.md follows READMES.md standard
11. [ ] Package documentation in doc.go
12. [ ] All 82 consumer files updated to new imports
13. [ ] Old `cli/internal/sdks/` removed
14. [ ] No regression in any project type functionality

## Reference Documents

- **Design Document**: `.sow/knowledge/designs/sdk-consolidation-design.md`
- **Exploration Summary**: `.sow/knowledge/explorations/libs-refactoring-summary.md`

## Out of Scope

- Adding new project features
- Implementing alternative backends (database, remote API)
- Plugin/dynamic loading system
- Performance optimization
- Breaking API changes
