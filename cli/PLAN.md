# CLI Architecture Refactoring Plan

**Status**: Approved
**Created**: 2025-10-19
**Purpose**: Refactor CLI to use unified domain abstraction layer

---

## Goal
Create a unified domain abstraction layer (`sow` package) that encapsulates filesystem operations and state machine logic, making CLI commands a thin presentation layer.

## Architecture Overview

### New Package Structure
```
cli/internal/
├── sow/                    # NEW - unified domain abstraction
│   ├── sow.go             # Sow type (main entrypoint)
│   ├── project.go         # Project type with phase/task orchestration
│   ├── task.go            # Task type with auto-save operations
│   ├── phase.go           # Phase helpers and types
│   ├── errors.go          # Domain-specific errors
│   ├── options.go         # Functional options pattern
│   └── sow_test.go        # Unit tests (in-memory FS)
├── statechart/            # KEEP - state machine (used internally by sow)
├── schemas/               # KEEP - CUE types
└── sowfs/                 # EVALUATE - may delete or keep minimal helpers
```

### Core Type Hierarchy
```
Sow (entrypoint)
  ├── fs.FS (billy abstraction)
  └── methods: Init(), CreateProject(), GetProject(), DeleteProject()

Project
  ├── sow *Sow (parent reference)
  ├── state *schemas.ProjectState
  ├── machine *statechart.Machine (accessible via Machine())
  └── methods: EnablePhase(), CompletePhase(), AddTask(), GetTask(), Save()

Task
  ├── project *Project (parent reference)
  ├── id string
  └── methods: SetStatus(), IncrementIteration(), AddFile(), AddFeedback()
    └── all methods auto-save via project.save()
```

## Key Design Decisions

1. **Auto-save**: Every mutation immediately persists to disk atomically
2. **State machine accessible**: `project.Machine()` for commands like session-info
3. **Concrete types**: No interfaces (mockable via in-memory FS)
4. **Big bang migration**: Single PR with all changes
5. **Encapsulation**: Commands never directly touch fs.FS or state machine

## Migration Steps

### Phase 1: Create Core Abstraction (sow package)

**File: cli/internal/sow/sow.go**
```go
type Sow struct {
    fs fs.FS
}

func New(fs fs.FS) *Sow
func (s *Sow) Init() error
func (s *Sow) HasProject() bool
func (s *Sow) GetProject() (*Project, error)
func (s *Sow) CreateProject(name, desc string) (*Project, error)
func (s *Sow) DeleteProject() error
```

**File: cli/internal/sow/project.go**
```go
type Project struct {
    sow     *Sow
    state   *schemas.ProjectState
    machine *statechart.Machine
}

// Access
func (p *Project) Machine() *statechart.Machine
func (p *Project) State() *schemas.ProjectState
func (p *Project) Name() string
func (p *Project) Branch() string

// Phase operations
func (p *Project) EnablePhase(name string, opts ...PhaseOption) error
func (p *Project) CompletePhase(name string) error
func (p *Project) GetPhase(name string) *Phase

// Task operations
func (p *Project) AddTask(name string, opts ...TaskOption) (*Task, error)
func (p *Project) GetTask(id string) (*Task, error)
func (p *Project) ListTasks() []*Task
func (p *Project) InferTaskID() (string, error)

// Artifact operations
func (p *Project) AddArtifact(phase, path string, approved bool) error
func (p *Project) ApproveArtifact(phase, path string) error

// Review operations
func (p *Project) IncrementReviewIteration() error
func (p *Project) AddReviewReport(path, assessment string) error

// Internal
func (p *Project) save() error  // atomic write
```

**File: cli/internal/sow/task.go**
```go
type Task struct {
    project *Project
    id      string
}

// Access (read-only, no auto-save)
func (t *Task) ID() string
func (t *Task) Name() string
func (t *Task) Status() string
func (t *Task) State() *schemas.TaskState

// Mutations (all auto-save)
func (t *Task) SetStatus(status string) error
func (t *Task) IncrementIteration() error
func (t *Task) SetAgent(agent string) error
func (t *Task) AddReference(path string) error
func (t *Task) AddFile(path string) error
func (t *Task) AddFeedback(msg string) (string, error)
func (t *Task) MarkFeedbackAddressed(feedbackID string) error
```

**File: cli/internal/sow/errors.go**
```go
var (
    ErrNoProject = errors.New("no active project")
    ErrProjectExists = errors.New("project already exists")
    ErrInvalidTransition = errors.New("invalid state transition")
    // etc.
)
```

### Phase 2: Update Root Command Context

**File: cli/cmd/root.go**
```go
// Change PersistentPreRunE to create Sow instance instead of SowFS
func PersistentPreRunE(cmd *cobra.Command, _ []string) error {
    cwd, _ := os.Getwd()
    baseFS := billy.NewLocal()
    fs, _ := baseFS.Chroot(cwd)

    sowInstance := sow.New(fs)  // NEW
    ctx := context.WithValue(cmd.Context(), "sow", sowInstance)
    cmd.SetContext(ctx)
    return nil
}

// Add helper
func sowFromContext(cmd *cobra.Command) *sow.Sow {
    return cmd.Context().Value("sow").(*sow.Sow)
}
```

### Phase 3: Refactor All Commands

**Before (project/init.go):**
```go
sfs := SowFSFromContext(cmd.Context())
// manual state creation
// manual file operations
machine, _ := statechart.Load()
machine.Fire(EventProjectInit)
machine.Save()
```

**After (project/init.go):**
```go
sow := sowFromContext(cmd.Context())
project, err := sow.CreateProject(name, description)
// Done! Auto-saved, state machine fired, prompt output
```

Commands to refactor (31 files):
- cmd/init.go
- cmd/validate.go
- cmd/log.go
- cmd/session_info.go
- project/*.go (14 files)
- task/*.go (12 files)

Each command simplified to:
1. Get sow from context
2. Call domain method
3. Handle errors
4. Print user feedback

### Phase 4: Update Tests

**Existing txtar tests:**
- Keep all existing tests (they test CLI behavior end-to-end)
- They should pass unchanged (same CLI interface)

**New unit tests (sow_test.go):**
```go
func TestProjectCreation(t *testing.T) {
    fs := memfs.New()  // in-memory FS
    sow := New(fs)

    project, err := sow.CreateProject("test", "description")
    require.NoError(t, err)

    // Verify state machine fired
    assert.Equal(t, statechart.DiscoveryDecision, project.Machine().State())

    // Verify file created
    exists, _ := fs.Stat(".sow/project/state.yaml")
    assert.NotNil(t, exists)
}

func TestTaskAutoSave(t *testing.T) {
    fs := memfs.New()
    sow := setupProject(fs)  // helper
    project, _ := sow.GetProject()

    task, _ := project.AddTask("test task")
    err := task.SetStatus("completed")
    require.NoError(t, err)

    // Reload from disk to verify auto-save
    project2, _ := sow.GetProject()
    task2, _ := project2.GetTask(task.ID())
    assert.Equal(t, "completed", task2.Status())
}
```

### Phase 5: Delete Old Abstractions

Remove or consolidate:
- `cli/internal/sowfs` package (replaced by sow package)
- `cli/internal/project` package (replaced by sow.Project)
- `cli/internal/task` package (replaced by sow.Task)
- `cli/internal/taskutil` package (logic moved to sow.Project)

Keep:
- `cli/internal/statechart` (used by sow package)
- `cli/internal/schemas` (shared types)
- `cli/internal/refs` (separate concern, migrate later)
- `cli/internal/logging` (separate concern, migrate later)

### Phase 6: Validation

1. Run all txtar tests: `go test ./cli/... -v`
2. Run new unit tests: `go test ./cli/internal/sow/...`
3. Manual testing of common workflows
4. Lint: `golangci-lint run ./cli/...`

## Benefits of New Architecture

**Before:**
- Commands: 50-100 lines (state management, fs operations, machine transitions)
- Fragmented logic across sowfs, project, task, statechart
- Hard to test business logic without CLI

**After:**
- Commands: 10-20 lines (thin presentation layer)
- All logic in sow package (single responsibility)
- Easy to test with in-memory FS
- Clear domain model matches documentation

## Example Transformation

**Old command pattern (project phase complete):**
```go
func runPhaseComplete(cmd *cobra.Command, args []string) error {
    sfs := SowFSFromContext(cmd.Context())
    machine, err := statechart.Load()
    // 30+ lines of validation, state updates, machine transitions
    return machine.Save()
}
```

**New command pattern:**
```go
func runPhaseComplete(cmd *cobra.Command, args []string) error {
    sow := sowFromContext(cmd.Context())
    project, err := sow.GetProject()
    if err != nil { return err }

    if err := project.CompletePhase(phaseName); err != nil {
        return err
    }

    fmt.Fprintf(os.Stderr, "✓ Completed %s phase\n", phaseName)
    return nil
}
```

## Out of Scope (Future Work)

- Refs system migration (keep current implementation)
- Logging system migration (keep current implementation)
- Validation command (keep current implementation)

These can be migrated later using the same pattern.

---

## Execution Order

1. Create `cli/internal/sow` package with all core types
2. Add comprehensive unit tests using in-memory FS
3. Update `cmd/root.go` to provide Sow in context
4. Refactor all 31 command files to use new API
5. Delete old packages (sowfs, project, task, taskutil)
6. Run full test suite
7. Manual validation of workflows
