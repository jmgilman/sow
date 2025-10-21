# CLI Architecture Refactor

**Status**: Planning
**Created**: 2025-01-21
**Goal**: Restructure internal packages for better maintainability and extensibility

---

## Motivation

The current architecture has the following issues:

1. **Poor cohesion**: `statechart` package is project-specific but lives separately, creating confusion
2. **Misleading abstractions**: `statechart.LoadFS()` loads entire project state, not just statechart
3. **God object risk**: `sow` package accumulating too much functionality - will become unmaintainable
4. **Discovery problems**: Hard to understand ownership and find relevant code
5. **Scalability concerns**: Adding new features (issues, workflows, analytics) will bloat existing packages

**Without intervention, we risk creating an unmaintainable monolith.**

---

## Core Design Principles

### 1. Context-Based Architecture

Introduce `sow.Context` as the universal dependency container:

```go
type Context struct {
    fs     SowFS           // Scoped to .sow/
    repo   *Git            // Local git operations
    github *GitHub         // Remote GitHub operations
}
```

All subsystems receive this context and use it to access what they need.

### 2. Aggregate Boundaries

Each major subsystem owns its domain completely:

- **Project**: Owns statechart, tasks, phase management
- **Refs**: Owns external references system
- **Issues** (future): Owns GitHub issue integration
- **Workflows** (future): Owns automation and hooks

### 3. Filesystem Abstraction

`SowFS` is an `fs.FS` chrooted to `.sow/` directory:

```go
type SowFS interface {
    fs.FS
    // Additional write operations as needed
}
```

All subsystems use this instead of raw billy filesystem.

### 4. Dependency Injection

Components depend on context, not each other:

```go
// Good
proj, err := project.Load(ctx)

// Bad (current)
proj, err := sow.GetProject() // sow knows about projects
```

### 5. Self-Contained Packages

Each package manages its own:
- State loading/saving
- Validation
- Business logic
- Error handling

---

## Target Architecture

```
cli/
├── cmd/                    # CLI commands (thin layer)
│   ├── root.go
│   ├── project/
│   ├── task/
│   └── refs/
│
├── internal/
│   ├── sow/                # Core context and primitives
│   │   ├── context.go      # Context type
│   │   ├── fs.go           # SowFS abstraction
│   │   ├── git.go          # Git operations
│   │   └── github.go       # GitHub client
│   │
│   ├── project/            # Project aggregate (self-contained)
│   │   ├── project.go      # Main project type and operations
│   │   ├── load.go         # Load(ctx), New(ctx, ...)
│   │   ├── state.go        # State management
│   │   ├── options.go      # Option pattern for operations
│   │   ├── statechart/     # State machine (internal to project)
│   │   │   ├── machine.go
│   │   │   ├── states.go
│   │   │   ├── events.go
│   │   │   └── guards.go
│   │   └── tasks/          # Task management (internal to project)
│   │       ├── task.go
│   │       ├── state.go
│   │       └── feedback.go
│   │
│   ├── refs/               # External references system
│   │   ├── manager.go      # NewManager(ctx)
│   │   ├── file.go
│   │   ├── git.go
│   │   ├── url.go
│   │   └── registry.go
│   │
│   ├── logging/            # Logging system
│   │   ├── entry.go
│   │   └── writer.go
│   │
│   ├── cmdutil/            # CLI utilities
│   │   └── context.go
│   │
│   └── (future)
│       ├── issues/         # GitHub issue integration
│       ├── workflows/      # Automation system
│       └── analytics/      # Usage analytics
│
└── schemas/                # CUE schemas and generated types
    ├── cue/
    └── cue_types_gen.go
```

---

## Key Type Definitions

### sow.Context

```go
// Context provides access to sow subsystems
type Context struct {
    fs     SowFS
    repo   *Git
    github *GitHub
}

// NewContext creates a context rooted at repository directory
func NewContext(repoRoot string) (*Context, error)

// Accessors
func (c *Context) FS() SowFS
func (c *Context) Git() *Git
func (c *Context) GitHub() *GitHub
```

### sow.SowFS

```go
// SowFS is a filesystem scoped to .sow/ directory
type SowFS interface {
    fs.FS
    // Write operations (TBD based on needs)
}
```

### project.Project

```go
// Project represents an active sow project
type Project struct {
    ctx     *sow.Context
    machine *statechart.Machine
}

// Load loads the active project from context
func Load(ctx *sow.Context) (*Project, error)

// New creates a new project
func New(ctx *sow.Context, name, desc string, opts ...Option) (*Project, error)

// Operations
func (p *Project) State() *schemas.ProjectState
func (p *Project) EnablePhase(name string, opts ...PhaseOption) error
func (p *Project) AddTask(name string, opts ...TaskOption) (*tasks.Task, error)
// ... etc
```

### refs.Manager

```go
// Manager handles external references
type Manager struct {
    ctx *sow.Context
}

// NewManager creates a refs manager
func NewManager(ctx *sow.Context) *Manager

// Operations
func (m *Manager) Add(url string, opts ...Option) error
func (m *Manager) List() ([]Ref, error)
func (m *Manager) Update(id string) error
// ... etc
```

---

## Migration Phases

### Phase 1: Create Context Foundation ✅ COMPLETED

**Goal**: Introduce new types without breaking existing code

**Tasks**:
- [x] Create `internal/sow/context.go`
  - [x] Define `Context` struct
  - [x] Implement `NewContext(repoRoot)`
  - [x] Add accessor methods
- [x] Create `internal/sow/fs.go`
  - [x] Define `SowFS` interface
  - [x] Implement chroot logic to `.sow/`
  - [x] Add write operations as needed
- [x] Create `internal/sow/git.go`
  - [x] Define `Git` type
  - [x] Implement `CurrentBranch()`, `IsRepo()`, etc.
  - [x] Extract from current `sow.go`
- [x] Create `internal/sow/github.go`
  - [x] Define `GitHub` type
  - [x] Move `internal/github/gh.go` functions here
  - [x] Make methods on `GitHub` type
- [x] Update `cmd/root.go` to create context in `PersistentPreRunE`
- [x] Store context in cobra command context (alongside old `Sow` for now)

**Validation**: ✅ Old code still works, new context available in commands, build succeeds

**Completed**: 2025-01-21

---

### Phase 2: Extract Project Package

**Goal**: Create self-contained `project` package with statechart and tasks

**Tasks**:
- [ ] Create `internal/project/` package structure
- [ ] Move `internal/statechart/` → `internal/project/statechart/`
  - [ ] Update imports
  - [ ] Make package internal to project (lowercase types if needed)
- [ ] Create `internal/project/tasks/` subpackage
  - [ ] Extract task-related code from current `sow/task.go`
  - [ ] Define `Task` type
  - [ ] Implement task operations
- [ ] Create `internal/project/project.go`
  - [ ] Define `Project` type with context
  - [ ] Implement `Load(ctx)` using context.FS()
  - [ ] Implement `New(ctx, name, desc, opts...)`
  - [ ] Move project operations from `sow/project.go`
- [ ] Create `internal/project/state.go`
  - [ ] State management operations
  - [ ] Save/load helpers
- [ ] Create `internal/project/options.go`
  - [ ] Option pattern for tasks, phases, etc.
  - [ ] `WithTaskID()`, `WithDiscoveryType()`, etc.
- [ ] **TEMPORARY**: Update `sow.Sow` to delegate to `project.Load(ctx)` for backward compatibility
  - [ ] Mark all delegation methods with `// DEPRECATED: Remove in Phase 4` comments
  - [ ] Document exactly which methods are temporary in Phase 4 section
- [ ] Update commands to use `project.Load(ctx)` directly

**Validation**: Project operations work through new package, old `sow` facade still works

**Note**: All backward compatibility shims added in this phase MUST be removed in Phase 4.

---

### Phase 3: Migrate Commands to Context

**Goal**: Update all commands to use context-based API

**Tasks**:
- [ ] Update `cmd/project/*.go` commands
  - [ ] Get context from cobra context
  - [ ] Use `project.Load(ctx)` instead of `sow.GetProject()`
  - [ ] Use `project.New(ctx, ...)` instead of `sow.CreateProject()`
- [ ] Update `cmd/task/*.go` commands
  - [ ] Use `project.Load(ctx)` then `proj.GetTask()`
  - [ ] Remove dependency on old `sow` methods
- [ ] Update `cmd/refs/*.go` commands
  - [ ] Create `refs.NewManager(ctx)`
  - [ ] Use manager methods instead of old functions
- [ ] Update other commands (`init`, `log`, `session-info`, etc.)
  - [ ] Use context where applicable
  - [ ] Simplify command logic

**Validation**: All commands work with new API, no old `sow.Sow` methods used

---

### Phase 4: Remove Old Sow API

**Goal**: Delete deprecated code and finalize migration

**IMPORTANT**: This phase removes ALL backward compatibility shims added in Phases 2 and 3.
We track exactly what needs to be removed to ensure nothing is left behind.

**Backward Compatibility Shims to Remove** (tracked from earlier phases):
- [ ] From Phase 2: Project delegation methods in `sow.Sow`
  - [ ] `GetProject()` → Use `project.Load(ctx)` directly
  - [ ] `CreateProject()` → Use `project.New(ctx, ...)` directly
  - [ ] `DeleteProject()` → Use `proj.Delete()` directly
  - [ ] Any project-related wrapper methods
- [ ] From Phase 2: Task delegation methods in `sow.Sow`
  - [ ] Any task-related wrapper methods
- [ ] From Phase 3: Command delegation methods
  - [ ] Any command wrapper methods we add

**Files to Delete**:
- [ ] `sow/project.go` (moved to `internal/project/`)
- [ ] `sow/task.go` (moved to `internal/project/tasks/`)
- [ ] Any other files marked for deletion in earlier phases

**Final Cleanup**:
- [ ] Update `sow/sow.go` to be minimal
  - [ ] Keep only `Init()`, `IsInitialized()`, `HasProject()`
  - [ ] Or consider removing entirely if context can handle it
- [ ] Clean up imports across codebase
- [ ] Verify no references to old API remain
- [ ] Update any internal documentation

**Validation**:
- ✅ Clean build, all tests pass
- ✅ No deprecated code remains
- ✅ No backward compatibility shims remain
- ✅ `git grep` for old method names returns no results

---

### Phase 5: Polish and Document

**Goal**: Finalize architecture and documentation

**Tasks**:
- [ ] Add package documentation
  - [ ] `sow/doc.go` explaining context pattern
  - [ ] `project/doc.go` explaining project aggregate
  - [ ] `refs/doc.go` explaining refs system
- [ ] Add examples in doc comments
- [ ] Update tests to use context-based API
- [ ] Add integration tests for common workflows
- [ ] Update `REFACTOR.md` with completion notes
- [ ] Consider creating architecture decision record (ADR)

**Validation**: Documentation clear, examples work, architecture sustainable

---

## Design Decisions

### Context Lifecycle

**Decision**: Create context once per CLI command invocation

**Rationale**:
- Commands are short-lived (seconds)
- No benefit to long-lived context
- Simpler to reason about

**Implementation**:
```go
// cmd/root.go PersistentPreRunE
ctx, err := sow.NewContext(repoRoot)
cmd.SetContext(cmdutil.WithSowContext(cmd.Context(), ctx))
```

---

### Lazy vs Eager Loading

**Decision**: Lazy load GitHub client, eager load FS and Git

**Rationale**:
- FS and Git needed by almost all operations
- GitHub only needed for issue/PR operations
- GitHub requires network and may fail/timeout

**Implementation**:
```go
func (c *Context) GitHub() *GitHub {
    if c.github == nil {
        c.github = newGitHub() // Lazy init
    }
    return c.github
}
```

---

### Error Handling for Missing .sow/

**Decision**: Components return specific errors, commands decide how to handle

**Rationale**:
- Different commands have different requirements
- `sow project status` requires `.sow/`, `sow init` doesn't
- Clear error types enable better UX

**Implementation**:
```go
// project/load.go
var ErrNotInitialized = errors.New("sow not initialized in repository")
var ErrNoProject = errors.New("no active project found")

func Load(ctx *sow.Context) (*Project, error) {
    if !ctx.FS().Initialized() {
        return nil, ErrNotInitialized
    }
    // ...
}
```

---

### Context Immutability

**Decision**: Context is immutable after creation

**Rationale**:
- Simpler reasoning about state
- No hidden mutations
- Thread-safe if needed in future

**Implementation**: No setter methods on Context, only getters

---

## Success Criteria

This refactor is successful when:

1. ✅ **Clear Ownership**: Obvious which package owns what functionality
2. ✅ **Easy Discovery**: New developers can find code quickly
3. ✅ **Extensible**: Can add new subsystems (issues, workflows) without touching existing code
4. ✅ **Testable**: Components testable in isolation with mock context
5. ✅ **Maintainable**: No files >500 lines, no god objects
6. ✅ **Documented**: Package docs explain architecture clearly

---

## Risks and Mitigations

### Risk: Breaking existing functionality

**Mitigation**:
- Migrate in phases
- Keep old API alongside new during transition
- Comprehensive testing at each phase

### Risk: Over-abstraction

**Mitigation**:
- Start simple, add complexity only when needed
- Prefer concrete types over interfaces
- Optimize for readability

### Risk: Performance regression

**Mitigation**:
- Context creation is cheap (no heavy I/O)
- Lazy loading for expensive operations
- Benchmark critical paths

---

## Current Status

**Phase**: Phase 1 Complete ✅
**Next Step**: Begin Phase 2 - Extract Project Package
**Last Updated**: 2025-01-21

**Phase 1 Achievements**:
- ✅ Context foundation with FS, Git, GitHub access
- ✅ Leverages existing `fs/billy` and `git` packages
- ✅ Interface-based `Executor` pattern for testability
- ✅ MockExecutor for unit testing without external dependencies
- ✅ All code compiles and tests pass
- ✅ Zero breaking changes - fully backward compatible

### Phase 1 Summary

Successfully created the context foundation with:
- `sow.Context` type providing unified access to FS, Git, and GitHub
- `SowFS` type alias for `core.FS` (scoped filesystem operations to .sow/ chroot)
- `Git` type wrapping `git.Repository` for repository operations (CurrentBranch, IsProtectedBranch, Branches, HasUncommittedChanges)
- `GitHub` type with methods for issue/PR operations (lazy-loaded, accepts `Executor` interface)
- `Executor` interface pattern for testable command execution
- Updated `cmd/root.go` to create and store context
- Added `cmdutil` functions for context storage/retrieval
- Backward compatibility maintained - old `sow.Sow` still works alongside new API

**Key Implementation Details**:
- `SowFS` uses `github.com/jmgilman/go/fs/billy` (billy-backed local filesystem)
- `Git` uses `github.com/jmgilman/go/git` (clean wrapper around go-git)
- `Executor` interface with `LocalExecutor` implementation and `MockExecutor` for tests
- `GitHub` accepts `Executor` interface, making it easily testable
- Both leverage existing, tested abstractions instead of custom implementations

**Files Created**:
- `internal/sow/context.go` (89 lines) - Context type
- `internal/sow/fs.go` (54 lines) - SowFS alias and NewSowFS
- `internal/sow/git.go` (134 lines) - Git wrapper
- `internal/sow/github.go` (380 lines) - GitHub CLI wrapper
- `internal/exec/executor.go` (166 lines) - Executor interface + LocalExecutor
- `internal/exec/mock.go` (77 lines) - MockExecutor for testing
- `internal/cmdutil/context.go` - Added SowContext helpers

**Tests Created**:
- `internal/sow/github_test.go` - Unit tests with mocked executor

**Build Status**: ✅ Compiles successfully
**Test Status**: ✅ All tests pass
**Breaking Changes**: None - fully backward compatible

---

## Notes

- Keep this document updated as we progress through phases
- Mark tasks complete with `[x]` as we finish them
- Add any discovered issues or decisions to relevant sections
- Consider creating ADR if architecture decisions become complex
