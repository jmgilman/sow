# Project Type Refactor - Implementation Plan

This document outlines the step-by-step plan to refactor the codebase to support multiple project types while keeping the CLI generic and type-agnostic.

## Goals

1. **CLI is type-agnostic** - Works through interfaces, no knowledge of specific project types
2. **Single generic Phase schema** - All phases use the same schema structure
3. **Helper collections** - Reusable logic for artifacts and tasks
4. **Clean abstractions** - Project and Phase interfaces separate concerns
5. **Extensible** - New project types require no CLI changes

## Core Architectural Principle

**Phases are generic data structures.** What makes a phase unique is:
1. **Guards** - Phase-specific transition logic
2. **Prompts** - Phase-specific instructions for the orchestrator
3. **Supported operations** - Which helpers are available (artifacts, tasks, both, neither)

Phase-specific data goes in the generic `metadata` field, not custom schema fields.

## Final Architecture

```
┌─────────────────────────────────────────────────┐
│                  CLI Commands                    │
│         (Type-agnostic, uses interfaces)        │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│           Project Interface                      │
│  - CurrentPhase() Phase                         │
│  - Phase(name) (Phase, error)                   │
│  - Save() error                                  │
└────────────────────┬────────────────────────────┘
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
┌──────────────────┐   ┌──────────────────┐
│ StandardProject  │   │ ExplorationProject│ (future)
└──────────────────┘   └──────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────┐
│              Phase Interface                     │
│  - AddArtifact(...) error                       │
│  - AddTask(...) (*Task, error)                  │
│  - Complete() error                             │
└─────────────────────┬───────────────────────────┘
                      │
       ┌──────────────┴──────────────┐
       ▼                             ▼
┌──────────────────┐         ┌──────────────────┐
│ ArtifactCollection│        │  TaskCollection  │
│  (helper logic)   │        │  (helper logic)  │
└──────────────────┘         └──────────────────┘
       │                             │
       └──────────────┬──────────────┘
                      ▼
              ┌───────────────┐
              │ phases.Phase  │
              │ (CUE schema)  │
              └───────────────┘
```

## Phase 1: Consolidate to Single Generic Phase Schema

### 1.1 Delete Phase-Specific Schemas

**Delete these files**:
- `schemas/phases/discovery.cue`
- `schemas/phases/design.cue`
- `schemas/phases/implementation.cue`
- `schemas/phases/review.cue`
- `schemas/phases/finalize.cue`

**Rationale**: Replace with single generic `#Phase` schema.

### 1.2 Create Generic Phase Schema

**File**: `schemas/phases/common.cue`

**Replace existing content with**:
```cue
package phases

import "time"

// Phase is the universal schema for all phases in all project types.
// What makes a phase unique is its guards, prompts, and which operations it supports.
#Phase: {
	// Common metadata
	status:  "pending" | "in_progress" | "completed" | "skipped"
	enabled: bool

	// Timestamps
	created_at:   time.Time
	started_at?:  time.Time @go(,optional=nillable)
	completed_at?: time.Time @go(,optional=nillable)

	// Generic collections (used by phases that need them)
	artifacts: [...#Artifact]  // Used by discovery, design, review
	tasks:     [...#Task]      // Used by implementation

	// Phase-specific data (discovery_type, iteration, etc.)
	metadata?: {[string]: _} @go(,optional=nillable)
}

// Artifact represents a phase artifact requiring human approval
#Artifact: {
	// Path relative to .sow/project/
	path: string

	// Human approval status
	approved: bool

	// When artifact was created
	created_at: time.Time

	// Phase-specific metadata (type, assessment, etc.)
	metadata?: {[string]: _} @go(,optional=nillable)
}

// Task represents an implementation task
#Task: {
	// Gap-numbered ID (010, 020, 030...)
	id: string & =~"^[0-9]{3,}$"

	// Task name
	name: string & !=""

	// Task status
	status: "pending" | "in_progress" | "completed" | "abandoned"

	// Can run in parallel with other tasks
	parallel: bool

	// Task IDs this task depends on
	dependencies?: [...string] @go(,optional=nillable)
}
```

**Rationale**:
- All phases share the same structure
- `artifacts[]` present but empty for phases that don't use them
- `tasks[]` present but empty for phases that don't use them
- Phase-specific fields go in `metadata`

### 1.3 Update StandardProjectState Schema

**File**: `schemas/projects/standard.cue`

**Change**:
```cue
// OLD (using specific phase types):
phases: {
    discovery: p.#DiscoveryPhase
    design: p.#DesignPhase
    implementation: p.#ImplementationPhase
    review: p.#ReviewPhase
    finalize: p.#FinalizePhase
}

// NEW (all use generic Phase):
phases: {
    discovery: p.#Phase
    design: p.#Phase
    implementation: p.#Phase
    review: p.#Phase
    finalize: p.#Phase
}
```

### 1.4 Update Re-exports

**File**: `schemas/project_state.cue`

**Remove old phase type exports**:
```cue
// DELETE these lines:
#DiscoveryPhase: phases.#DiscoveryPhase
#DesignPhase: phases.#DesignPhase
#ImplementationPhase: phases.#ImplementationPhase
#ReviewPhase: phases.#ReviewPhase
#FinalizePhase: phases.#FinalizePhase
#ReviewReport: phases.#ReviewReport

// KEEP only:
#Phase: phases.#Phase
#Artifact: phases.#Artifact
#Task: phases.#Task
```

### 1.5 Regenerate Go Types

**Commands**:
```bash
cd cli/schemas
go generate ./...
cd ..
go build ./...  # Verify compilation
```

**Expected result**:
- `phases.Phase` is now a concrete struct with `Artifacts []Artifact`, `Tasks []Task`, `Metadata map[string]any`
- All old phase types are gone

## Phase 2: Define Interfaces

### 2.1 Create Domain Interfaces

**File**: `internal/project/domain.go` (NEW)

```go
package project

import (
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// Project is the aggregate root for all project types.
// The CLI works exclusively through this interface.
type Project interface {
	// Identity
	Name() string
	Branch() string
	Description() string
	Type() string

	// Phase access
	CurrentPhase() Phase
	Phase(name string) (Phase, error)

	// State machine
	Machine() *statechart.Machine

	// Persistence
	Save() error

	// Logging
	Log(action, result string, opts ...LogOption) error
}

// Phase represents any phase in any project type.
// Operations not supported by a phase return ErrNotSupported.
type Phase interface {
	// Metadata
	Name() string
	Status() string
	Enabled() bool

	// Artifact operations (discovery, design, review)
	// Returns schema types directly - no wrapper needed
	AddArtifact(path string, opts ...ArtifactOption) error
	ApproveArtifact(path string) error
	ListArtifacts() []*phases.Artifact

	// Task operations (implementation only)
	// Returns concrete Task type - single implementation for all project types
	AddTask(name string, opts ...TaskOption) (*Task, error)
	GetTask(id string) (*Task, error)
	ListTasks() []*Task
	ApproveTasks() error

	// Generic field access (for metadata)
	Set(field string, value interface{}) error
	Get(field string) (interface{}, error)

	// Lifecycle
	Complete() error
	Skip() error
	Enable(opts ...PhaseOption) error
}

// Options for creating artifacts
type ArtifactOption func(*ArtifactConfig)

type ArtifactConfig struct {
	Metadata map[string]interface{}
}

func WithMetadata(metadata map[string]interface{}) ArtifactOption {
	return func(c *ArtifactConfig) {
		c.Metadata = metadata
	}
}

// Options for creating tasks
type TaskOption func(*TaskConfig)

type TaskConfig struct {
	Status       string
	Description  string
	Agent        string
	Dependencies []string
	// ... other fields
}

// Options for phase enable
type PhaseOption func(*PhaseConfig)

type PhaseConfig struct {
	// Phase-specific options (discovery type, etc.)
	Metadata map[string]interface{}
}

// Options for logging
type LogOption func(*LogEntry)

type LogEntry struct {
	// ... logging fields
}
```

### 2.2 Create Standard Errors

**File**: `internal/project/errors.go` (NEW)

```go
package project

import "errors"

var (
	// ErrNotSupported indicates an operation is not supported by this phase
	ErrNotSupported = errors.New("operation not supported by this phase")

	// ErrPhaseNotFound indicates a phase with the given name doesn't exist
	ErrPhaseNotFound = errors.New("phase not found")

	// ErrNoProject indicates no project exists
	ErrNoProject = errors.New("no project exists")

	// ErrProjectExists indicates a project already exists
	ErrProjectExists = errors.New("project already exists")

	// ErrInvalidPhase indicates an invalid phase name
	ErrInvalidPhase = errors.New("invalid phase")

	// ErrPhaseNotEnabled indicates the phase is not enabled
	ErrPhaseNotEnabled = errors.New("phase not enabled")

	// ErrNoTask indicates no task with that ID exists
	ErrNoTask = errors.New("task not found")
)
```

## Phase 3: Implement Helper Collections and Phases

### 3.1 Current vs New Structure

**Current structure** (to be refactored):
```
internal/project/
├── project.go          # Project operations with direct state access
├── task.go             # Task operations
├── helpers.go          # Phase detection utilities (DELETE)
├── log.go              # Logging helpers (KEEP)
├── state.go            # Load/Create functions (KEEP)
│
└── types/              # Composable phases architecture (DELETE)
    ├── types.go
    └── standard/
        ├── standard.go
        └── standard_test.go
```

**New structure**:
```
internal/project/
├── domain.go           # Interfaces: Project, Phase (NEW)
├── errors.go           # Standard errors (NEW)
├── registry.go         # DetectProjectType logic (NEW)
├── project.go          # Load/Create/Delete functions (REFACTORED)
├── task.go             # Task concrete type (REFACTORED)
├── log.go              # Logging helpers (UNCHANGED)
│
└── standard/           # StandardProject implementation
    ├── project.go      # StandardProject (implements Project)
    ├── artifacts.go    # ArtifactCollection helper (NEW)
    ├── tasks.go        # TaskCollection helper (NEW)
    ├── discovery.go    # DiscoveryPhase (thin wrapper)
    ├── design.go       # DesignPhase (thin wrapper)
    ├── implementation.go # ImplementationPhase (thin wrapper)
    ├── review.go       # ReviewPhase (thin wrapper)
    └── finalize.go     # FinalizePhase (thin wrapper)
```

### 3.2 Implement Artifact Collection Helper

**File**: `internal/project/standard/artifacts.go` (NEW)

```go
package standard

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// ArtifactCollection provides artifact operations on a generic Phase.
type ArtifactCollection struct {
	state   *phases.Phase
	project project.Project
}

// NewArtifactCollection creates a new artifact collection.
func NewArtifactCollection(state *phases.Phase, proj project.Project) *ArtifactCollection {
	return &ArtifactCollection{
		state:   state,
		project: proj,
	}
}

// Add adds a new artifact to the phase.
func (ac *ArtifactCollection) Add(path string, opts ...project.ArtifactOption) error {
	cfg := &project.ArtifactConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	artifact := phases.Artifact{
		Path:      path,
		Approved:  false,
		CreatedAt: time.Now(),
		Metadata:  cfg.Metadata,
	}

	ac.state.Artifacts = append(ac.state.Artifacts, artifact)
	return ac.project.Save()
}

// Approve marks an artifact as approved.
func (ac *ArtifactCollection) Approve(path string) error {
	for i := range ac.state.Artifacts {
		if ac.state.Artifacts[i].Path == path {
			ac.state.Artifacts[i].Approved = true
			return ac.project.Save()
		}
	}
	return fmt.Errorf("artifact not found: %s", path)
}

// List returns all artifacts in the phase.
func (ac *ArtifactCollection) List() []*phases.Artifact {
	result := make([]*phases.Artifact, len(ac.state.Artifacts))
	for i := range ac.state.Artifacts {
		result[i] = &ac.state.Artifacts[i]
	}
	return result
}

// AllApproved checks if all artifacts are approved.
func (ac *ArtifactCollection) AllApproved() bool {
	for _, a := range ac.state.Artifacts {
		if !a.Approved {
			return false
		}
	}
	return true
}
```

### 3.3 Implement Task Collection Helper

**File**: `internal/project/standard/tasks.go` (NEW)

```go
package standard

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// TaskCollection provides task operations on a generic Phase.
type TaskCollection struct {
	state   *phases.Phase
	project project.Project
	ctx     *sow.Context
}

// NewTaskCollection creates a new task collection.
func NewTaskCollection(state *phases.Phase, proj project.Project, ctx *sow.Context) *TaskCollection {
	return &TaskCollection{
		state:   state,
		project: proj,
		ctx:     ctx,
	}
}

// Add adds a new task to the phase.
func (tc *TaskCollection) Add(name string, opts ...project.TaskOption) (*project.Task, error) {
	cfg := &project.TaskConfig{
		Status: "pending",
		Agent:  "implementer",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Generate task ID
	id := tc.generateTaskID()

	// Validate dependencies
	for _, depID := range cfg.Dependencies {
		if !tc.taskExists(depID) {
			return nil, fmt.Errorf("dependency task not found: %s", depID)
		}
	}

	task := phases.Task{
		Id:           id,
		Name:         name,
		Status:       cfg.Status,
		Parallel:     false,
		Dependencies: cfg.Dependencies,
	}

	tc.state.Tasks = append(tc.state.Tasks, task)

	// Create task directory structure
	if err := tc.createTaskStructure(id, name, cfg); err != nil {
		return nil, err
	}

	if err := tc.project.Save(); err != nil {
		return nil, err
	}

	return &project.Task{Project: tc.project, ID: id}, nil
}

// Get retrieves a task by ID.
func (tc *TaskCollection) Get(id string) (*project.Task, error) {
	if !tc.taskExists(id) {
		return nil, project.ErrNoTask
	}
	return &project.Task{Project: tc.project, ID: id}, nil
}

// List returns all tasks.
func (tc *TaskCollection) List() []*project.Task {
	tasks := make([]*project.Task, 0, len(tc.state.Tasks))
	for _, t := range tc.state.Tasks {
		tasks = append(tasks, &project.Task{Project: tc.project, ID: t.Id})
	}
	return tasks
}

// Approve marks tasks as approved for execution.
func (tc *TaskCollection) Approve() error {
	if len(tc.state.Tasks) == 0 {
		return fmt.Errorf("cannot approve: no tasks exist")
	}

	// Set approval in metadata
	if tc.state.Metadata == nil {
		tc.state.Metadata = make(map[string]interface{})
	}
	tc.state.Metadata["tasks_approved"] = true

	tc.state.Status = "in_progress"
	if tc.state.StartedAt == nil {
		now := time.Now()
		tc.state.StartedAt = &now
	}

	return tc.project.Save()
}

// Helper methods
func (tc *TaskCollection) generateTaskID() string {
	maxID := 0
	for _, t := range tc.state.Tasks {
		var id int
		fmt.Sscanf(t.Id, "%d", &id)
		if id > maxID {
			maxID = id
		}
	}
	return fmt.Sprintf("%03d", maxID+10)
}

func (tc *TaskCollection) taskExists(id string) bool {
	for _, t := range tc.state.Tasks {
		if t.Id == id {
			return true
		}
	}
	return false
}

func (tc *TaskCollection) createTaskStructure(id, name string, cfg *project.TaskConfig) error {
	// Implementation similar to current project.go createTaskStructure
	// Uses tc.ctx for filesystem access
	// ...
	return nil
}
```

### 3.4 Implement StandardProject

**File**: `internal/project/standard/project.go`

```go
package standard

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/phases"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// StandardProject implements the standard 5-phase project lifecycle.
type StandardProject struct {
	state   *projects.StandardProjectState
	ctx     *sow.Context
	machine *statechart.Machine
	phases  map[string]project.Phase
}

// New creates a new StandardProject.
func New(state *projects.StandardProjectState, ctx *sow.Context) *StandardProject {
	p := &StandardProject{
		state:  state,
		ctx:    ctx,
		phases: make(map[string]project.Phase),
	}

	// Create phase instances (they need parent project for Save())
	p.phases["discovery"] = NewDiscoveryPhase(&state.Phases.Discovery, p, ctx)
	p.phases["design"] = NewDesignPhase(&state.Phases.Design, p, ctx)
	p.phases["implementation"] = NewImplementationPhase(&state.Phases.Implementation, p, ctx)
	p.phases["review"] = NewReviewPhase(&state.Phases.Review, p, ctx)
	p.phases["finalize"] = NewFinalizePhase(&state.Phases.Finalize, p, ctx)

	// Build state machine
	p.machine = p.buildStateMachine()
	p.machine.SetFilesystem(ctx.FS())

	return p
}

// Implements Project interface
func (p *StandardProject) Name() string {
	return p.state.Project.Name
}

func (p *StandardProject) Branch() string {
	return p.state.Project.Branch
}

func (p *StandardProject) Description() string {
	return p.state.Project.Description
}

func (p *StandardProject) Type() string {
	return "standard"
}

func (p *StandardProject) CurrentPhase() project.Phase {
	currentState := p.machine.State()

	switch currentState {
	case statechart.DiscoveryActive:
		return p.phases["discovery"]
	case statechart.DesignActive:
		return p.phases["design"]
	case statechart.ImplementationPlanning, statechart.ImplementationExecuting:
		return p.phases["implementation"]
	case statechart.ReviewActive:
		return p.phases["review"]
	case statechart.FinalizeDocumentation, statechart.FinalizeChecks, statechart.FinalizeDelete:
		return p.phases["finalize"]
	default:
		return nil
	}
}

func (p *StandardProject) Phase(name string) (project.Phase, error) {
	phase, ok := p.phases[name]
	if !ok {
		return nil, project.ErrPhaseNotFound
	}
	return phase, nil
}

func (p *StandardProject) Machine() *statechart.Machine {
	return p.machine
}

func (p *StandardProject) Save() error {
	return p.machine.Save()
}

func (p *StandardProject) Log(action, result string, opts ...project.LogOption) error {
	// Implementation...
	return nil
}

func (p *StandardProject) buildStateMachine() *statechart.Machine {
	// Existing BuildStateMachine logic from types/standard/standard.go
	// Creates state machine using composable phases architecture
	// ...
	return nil
}
```

### 3.5 Implement Thin Phase Wrappers

Each phase is now a thin wrapper that delegates to helpers.

**File**: `internal/project/standard/review.go` (EXAMPLE)

```go
package standard

import (
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/sow"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
)

// ReviewPhase implements the review phase for standard projects.
type ReviewPhase struct {
	state     *phasesSchema.Phase  // Generic schema!
	artifacts *ArtifactCollection
	project   *StandardProject
	ctx       *sow.Context
}

// NewReviewPhase creates a new review phase.
func NewReviewPhase(state *phasesSchema.Phase, proj *StandardProject, ctx *sow.Context) *ReviewPhase {
	return &ReviewPhase{
		state:     state,
		artifacts: NewArtifactCollection(state, proj),
		project:   proj,
		ctx:       ctx,
	}
}

// Implements Phase interface - delegate to helpers
func (p *ReviewPhase) Name() string {
	return "review"
}

func (p *ReviewPhase) Status() string {
	return p.state.Status
}

func (p *ReviewPhase) Enabled() bool {
	return p.state.Enabled
}

func (p *ReviewPhase) AddArtifact(path string, opts ...project.ArtifactOption) error {
	return p.artifacts.Add(path, opts...)
}

func (p *ReviewPhase) ApproveArtifact(path string) error {
	return p.artifacts.Approve(path)
}

func (p *ReviewPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

func (p *ReviewPhase) AddTask(name string, opts ...project.TaskOption) (*project.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *ReviewPhase) GetTask(id string) (*project.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *ReviewPhase) ListTasks() []*project.Task {
	return nil
}

func (p *ReviewPhase) ApproveTasks() error {
	return project.ErrNotSupported
}

func (p *ReviewPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

func (p *ReviewPhase) Get(field string) (interface{}, error) {
	if p.state.Metadata == nil {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	val, ok := p.state.Metadata[field]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	return val, nil
}

func (p *ReviewPhase) Complete() error {
	// Update status and timestamps
	p.state.Status = "completed"
	now := time.Now()
	p.state.CompletedAt = &now

	if err := p.project.Save(); err != nil {
		return err
	}

	// Fire state machine event
	return p.project.Machine().Fire(statechart.EventReviewPass)
}

func (p *ReviewPhase) Skip() error {
	return project.ErrNotSupported // Review is required
}

func (p *ReviewPhase) Enable(opts ...project.PhaseOption) error {
	return project.ErrNotSupported // Review is always enabled
}

// Phase-specific guard (used by state machine)
func (p *ReviewPhase) AllReviewsApproved() bool {
	// Check for artifacts with type=review that aren't approved
	for _, artifact := range p.state.Artifacts {
		if artifactType, ok := artifact.Metadata["type"].(string); ok {
			if artifactType == "review" && !artifact.Approved {
				return false
			}
		}
	}
	return true
}

// Helper for accessing metadata
func (p *ReviewPhase) Iteration() int {
	if p.state.Metadata == nil {
		return 1
	}
	if iter, ok := p.state.Metadata["iteration"].(int); ok {
		return iter
	}
	return 1
}
```

**Similar thin wrappers for**:
- `discovery.go` - Uses ArtifactCollection, no tasks
- `design.go` - Uses ArtifactCollection, no tasks
- `implementation.go` - Uses TaskCollection, no artifacts
- `finalize.go` - Uses neither (or custom logic)

### 3.6 Create Registry

**File**: `internal/project/registry.go` (NEW)

```go
package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project/standard"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// DetectAndCreate detects the project type from state and returns a Project interface.
func DetectAndCreate(state *projects.StandardProjectState, ctx *sow.Context) (Project, error) {
	// State migration: default empty type to "standard"
	if state.Project.Type == "" {
		state.Project.Type = "standard"
	}

	// For now, only StandardProject exists
	switch state.Project.Type {
	case "standard":
		return standard.New(state, ctx), nil
	default:
		return nil, fmt.Errorf("unknown project type: %s", state.Project.Type)
	}
}
```

### 3.7 Update Load Function

**File**: `internal/project/project.go` (UPDATE)

Change `Load()` to return `Project` interface:

```go
func Load(ctx *sow.Context) (Project, error) {
	// Check if project exists
	exists, err := ctx.FS().Exists("project/state.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}
	if !exists {
		return nil, ErrNoProject
	}

	// Load state from disk
	state, err := loadProjectState(ctx.FS())
	if err != nil {
		return nil, fmt.Errorf("failed to load project state: %w", err)
	}

	// Detect type and create appropriate Project implementation
	return DetectAndCreate(state, ctx)
}

func loadProjectState(fs sow.FS) (*projects.StandardProjectState, error) {
	// Load and unmarshal YAML
	// ...
	return state, nil
}
```

### 3.8 Refactor Task Type

**File**: `internal/project/task.go` (REFACTOR)

Update Task to work with `Project` interface:

```go
package project

import (
	"fmt"
	"path/filepath"

	"github.com/jmgilman/sow/cli/schemas"
)

// Task represents an implementation task.
// This concrete type works with any project type through the Project interface.
type Task struct {
	Project Project  // Interface, not concrete type
	ID      string
}

// Name returns the task name from project state.
func (t *Task) Name() string {
	// Access through phase interface
	phase := t.Project.CurrentPhase()
	tasks := phase.ListTasks()
	for _, task := range tasks {
		if task.ID == t.ID {
			return task.Name() // Assuming we add Name() method
		}
	}
	return ""
}

// Status returns the current task status.
func (t *Task) Status() string {
	// Similar implementation using phase interface
	// ...
	return ""
}

// State returns the task state from disk.
func (t *Task) State() (*schemas.TaskState, error) {
	statePath := filepath.Join("project/phases/implementation/tasks", t.ID, "state.yaml")
	// Read and unmarshal...
	return nil, nil
}

// SetStatus updates the task status.
func (t *Task) SetStatus(status string) error {
	// Update task in phase state
	// Call t.Project.Save()
	return nil
}

// ... other methods working through Project interface
```

## Phase 4: Update CLI Commands

### 4.1 Delete Phase-Specific Commands

**Delete these files**:
```bash
rm cmd/agent/review_add.go
rm cmd/agent/review_approve.go
rm cmd/agent/review_increment.go
rm cmd/agent/finalize_complete.go
rm cmd/agent/finalize_doc.go
rm cmd/agent/finalize_move.go
rm cmd/agent/enable.go
rm cmd/agent/skip.go
```

**Rationale**: Functionality is covered by generic commands.

### 4.2 Update artifact_add.go to Support Metadata

**File**: `cmd/agent/artifact_add.go`

**Add flags**:
```go
var metadataFlags []string

func init() {
	artifactAddCmd.Flags().StringArrayVar(&metadataFlags, "metadata", nil,
		"Metadata key=value pairs (can specify multiple times)")
}
```

**Parse metadata and call via interface**:
```go
func runArtifactAdd(cmd *cobra.Command, args []string) error {
	ctx := sow.FromContext(cmd.Context())

	// Parse metadata
	metadata := make(map[string]interface{})
	for _, pair := range metadataFlags {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid metadata format: %s (use key=value)", pair)
		}
		metadata[parts[0]] = parts[1]
	}

	// Load project
	proj, err := project.Load(ctx)
	if err != nil {
		return err
	}

	// Get current phase
	phase := proj.CurrentPhase()

	// Add artifact
	err = phase.AddArtifact(path, project.WithMetadata(metadata))
	if errors.Is(err, project.ErrNotSupported) {
		return fmt.Errorf("current phase (%s) does not support artifacts", phase.Name())
	}

	fmt.Printf("✓ Artifact added: %s\n", path)
	return err
}
```

### 4.3 Update set.go for Generic Field Setting

**File**: `cmd/agent/set.go`

```go
func runSet(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: sow agent set <field> <value>")
	}

	field := args[0]
	value := args[1]

	ctx := sow.FromContext(cmd.Context())
	proj, err := project.Load(ctx)
	if err != nil {
		return err
	}

	phase := proj.CurrentPhase()

	// Try to parse value as int, bool, or keep as string
	var parsedValue interface{} = value
	if intVal, err := strconv.Atoi(value); err == nil {
		parsedValue = intVal
	} else if boolVal, err := strconv.ParseBool(value); err == nil {
		parsedValue = boolVal
	}

	err = phase.Set(field, parsedValue)
	if errors.Is(err, project.ErrNotSupported) {
		return fmt.Errorf("field %s not supported by phase %s", field, phase.Name())
	}

	fmt.Printf("✓ Set %s = %v\n", field, parsedValue)
	return err
}
```

### 4.4 Update complete.go

**File**: `cmd/agent/complete.go`

```go
func runComplete(cmd *cobra.Command, args []string) error {
	ctx := sow.FromContext(cmd.Context())

	proj, err := project.Load(ctx)
	if err != nil {
		return err
	}

	phase := proj.CurrentPhase()

	if err := phase.Complete(); err != nil {
		return fmt.Errorf("failed to complete phase: %w", err)
	}

	fmt.Printf("✓ Phase %s completed\n", phase.Name())
	return nil
}
```

### 4.5 Update Other Generic Commands

Similar updates for:
- `artifact_approve.go` - Use `phase.ApproveArtifact()`
- `artifact_list.go` - Use `phase.ListArtifacts()`
- `task/*.go` - Use `phase.AddTask()`, `phase.GetTask()`, etc.
- `status.go` - Use `proj.CurrentPhase()`, handle `ErrNotSupported` gracefully
- `info.go` - Use phase interface methods

All commands work through interfaces, no knowledge of concrete types.

## Phase 5: Update Phase Prompts

### 5.1 Update Review Phase Prompt

**File**: Wherever review phase prompt lives (likely in phases package or templates)

**Update to use generic commands**:
```markdown
# Review Phase

You are in autonomous review mode. Your task:

1. Review all code changes since implementation
2. Create a review report at `.sow/project/phases/review/reports/001.md`
3. Add it as an artifact with review metadata:

   ```bash
   sow agent artifact add \
     --path project/phases/review/reports/001.md \
     --metadata type=review \
     --metadata assessment=pass
   ```

   Use `assessment=fail` if issues require changes.

4. Wait for human approval of the artifact
5. If approved, complete the phase:

   ```bash
   sow agent complete
   ```

The phase will only complete when all review artifacts are approved.
```

### 5.2 Similar Updates for Other Phases

Update prompts to:
- Use `sow agent artifact add --metadata` instead of phase-specific commands
- Use `sow agent set` for setting metadata fields
- Use `sow agent complete` for phase completion

## Phase 6: Update Tests

### 6.1 Update Schema Tests

Tests that reference old phase types need updates:
- Replace `state.Phases.Review.Reports` with `state.Phases.Review.Artifacts`
- Filter artifacts by `Metadata["type"] == "review"`
- Update assertions to check generic Phase structure

### 6.2 Update Integration Tests

Tests that call CLI commands:
- Replace `sow agent review add` with `sow agent artifact add --metadata type=review`
- Replace `sow agent review approve` with `sow agent artifact approve`
- Replace `sow agent review increment` with `sow agent set iteration <n>`

### 6.3 Add Interface Tests

Create tests in `internal/project/standard/` that verify:
- Phases correctly implement `Phase` interface
- Unsupported operations return `ErrNotSupported`
- Guards check metadata correctly
- Helper collections work properly

## Phase 7: Cleanup

### 7.1 Delete Legacy Code

**Delete**:
- `internal/project/types/` directory (entire directory)
- `internal/project/helpers.go` (phase detection no longer needed)
- Phase-specific logic in `internal/project/project.go` that was moved to `standard/`

### 7.2 Update Documentation

- Update README with new CLI command examples
- Document metadata conventions (review artifacts, iteration field, etc.)
- Add examples of adding new project types

## Implementation Order

Execute phases in order, with verification after each:

1. **Phase 1** (Schemas) - Delete phase-specific schemas, create generic Phase, run `go generate`
2. **Phase 2** (Interfaces) - Define Project and Phase interfaces
3. **Phase 3** (Implementation) - Helper collections, StandardProject, thin phase wrappers
4. **Phase 4** (CLI) - Update commands to use interfaces
5. **Phase 5** (Prompts) - Update templates
6. **Phase 6** (Tests) - Fix/update all tests
7. **Phase 7** (Cleanup) - Remove dead code

## Success Criteria

✅ All phases use single generic `phases.Phase` schema
✅ Helper collections eliminate duplication
✅ Phase implementations are ~50 lines each
✅ No CLI commands reference specific project types
✅ Review phase uses artifacts with metadata, not custom types
✅ Guards check metadata correctly
✅ All tests pass
✅ Can describe how to add a new project type without CLI changes

## Future: Adding Exploration Project Type

With this architecture, adding a new project type requires:

1. Define `ExplorationProjectState` schema with phases: `{exploration: p.#Phase, decomposition: p.#Phase, finalize: p.#Phase}`
2. Implement `internal/project/exploration/` package:
   - `project.go` - ExplorationProject (implements Project)
   - `exploration.go`, `decomposition.go`, `finalize.go` - Thin phase wrappers using same helper collections
3. Update `registry.go` to detect and instantiate exploration projects
4. Write exploration-specific guards and prompts

**No CLI changes needed** - all commands work through `Project` and `Phase` interfaces.
**No new helper code needed** - ArtifactCollection and TaskCollection work for any phase.

---

**Estimated Effort**: 3-5 days for complete implementation and testing.
