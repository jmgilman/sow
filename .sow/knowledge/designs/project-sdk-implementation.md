# Project SDK Implementation Design

**Author**: Design Team
**Date**: 2025-11-02
**Status**: Proposed
**Size**: Comprehensive

---

## Executive Summary

This design introduces a unified Project SDK that enables defining project types through a fluent builder API while maintaining CUE schema validation. The SDK eliminates code duplication across project types, provides consistent state management patterns, and enables future extensibility for task and artifact state machines. This is a **major refactor** of the current project architecture, built in parallel packages to allow incremental migration from the existing `internal/project` implementation.

---

## Overview

The Project SDK provides a declarative API for defining project types with phases, state machines, prompts, and validation rules. Project types configure their complete behavior through a single fluent builder interface, and the SDK handles state machine wiring, phase transitions, CUE validation, and persistence.

**What is being built:**
- Unified Project SDK with fluent builder API (`internal/sdks/project`)
- Universal project state types wrapping CUE-generated schemas (`internal/sdks/project/state`)
- Improved state machine SDK forked from current statechart (`internal/sdks/state`)
- CUE schemas for project structure validation (`schemas/project`)
- Refactored project type implementations using the SDK (`internal/projects/standard`, `internal/projects/exploration`)

**Why it matters:**
- Current architecture has significant code duplication across project types (phase management, state machines, validation)
- Project types require manual wiring of phase transitions, guards, and validation
- No consistent pattern for metadata schemas across project types
- Difficult to extend with new project types or add task/artifact state machines
- **Enables unified command hierarchy** - Universal types make implementing the [Unified Command Hierarchy Design](../../knowledge/designs/command-hierarchy-design.md) straightforward since all project types share identical structure

---

## Goals and Non-Goals

### Goals

**Functional:**
- Single fluent builder API for complete project type configuration
- Universal Project/Phase/Artifact/Task types that work for all project types
- CUE schema validation for structure with runtime metadata validation
- Collection pattern for navigation (Get, Add, Remove) with direct field mutation
- Automatic phase status management based on state machine transitions
- Support for embedded per-phase metadata schemas (CUE)
- Clean separation between data types and orchestration logic
- Consolidated state management (all task state in project state file, not separate files)

**Non-Functional:**
- Load/Save operations in < 50ms for typical project state
- Zero-context resumability (all state on disk)
- Incremental migration path from existing implementation
- Clean cutover without gradual migration complexity

### Non-Goals

**Explicitly out of scope:**
- Task and artifact state machines (future enhancement, but architecture enables it)
- Backward compatibility with existing project state YAML (breaking changes acceptable in early development)
- Migration utilities for existing projects (fresh start with new SDK)
- Complex CUE features (unions, conditionals, loops) - keeping schemas simple
- External CUE validation tool (`cue vet`) for metadata (structure only)

### Success Metrics

- 70% reduction in code duplication across project types (measured by lines of code)
- New project types implementable in < 200 lines of configuration code
- Zero regressions in existing project type behavior after migration
- All existing tests pass with new implementation

---

## Background

### Current State

The current project architecture (`internal/project`) has significant issues:

**Code Duplication:**
- Each project type manually implements phase management
- State machine transitions hardcoded per project type
- Validation logic duplicated across project types
- Prompt generation patterns repeated

**Tight Coupling:**
- Phase types are project-type-specific structs (e.g., `ExplorationPhase` vs `StandardPhase`)
- Shared code difficult to extract due to type differences
- Circular dependencies between project and phase types

**Validation Inconsistency:**
- Some validation uses CUE schemas (`schemas/projects/`)
- Some validation uses manual checks in code
- No consistent pattern for metadata validation
- CUE schemas separate from project type implementations

**State Management Issues:**
- Task state split across separate files (`.sow/project/phases/implementation/tasks/<id>/state.yaml`)
- Requires multiple file I/O operations for task updates
- Atomic updates difficult (changes across multiple files)
- Single source of truth principle violated

**Extensibility Limitations:**
- Adding task/artifact state machines would require significant refactoring
- No clear pattern for project types to extend base functionality
- Metadata schemas not co-located with project types

### Motivation

**Problem being solved:**
1. **High implementation cost** - New project types require 500+ lines of boilerplate
2. **Maintenance burden** - Bug fixes must be replicated across project types
3. **Inconsistent patterns** - No standard way to define phases, state machines, validation
4. **Future blocked** - Cannot add task/artifact state machines without major refactor

**Impact of not solving:**
- Each new project type becomes harder to implement
- Bugs in shared patterns continue to be duplicated
- Team velocity decreases as codebase grows
- Task/artifact state machines remain unimplementable

### Requirements

**Functional Requirements:**
- FR1: Define complete project type through single builder API
- FR2: Support universal Project/Phase/Artifact/Task types across all project types
- FR3: Validate project structure using CUE schemas
- FR4: Validate phase metadata using embedded CUE schemas
- FR5: Automatic phase status transitions based on state machine
- FR6: Collection-based navigation with direct field mutation
- FR7: Load/Save with CUE validation and atomic writes

**Non-Functional Requirements:**
- NFR1: Incremental implementation (new SDK in parallel packages)
- NFR2: Type-safe operations for common fields (90% of access)
- NFR3: Runtime validation only for metadata fields (10% of access)
- NFR4: Self-documenting API (clear from builder usage)
- NFR5: Clean cutover when ready (no gradual migration complexity)

### Constraints

**Technical Constraints:**
- Must use CUE for consistency with other schemas (`breakdown_index.cue`, `design_index.cue`, etc.)
- Must support Go 1.21+ (current version)
- State machine must handle guard closures capturing project state

**Business Constraints:**
- Must preserve zero-context resumability
- Development timeline: 4-6 weeks for implementation + migration
- Breaking changes acceptable (sow in early development, no public usage)

---

## Design

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      CLI Commands Layer                         │
│  (sow advance, sow output set, sow task create, etc.)          │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                 Project State Operations                        │
│              (internal/sdks/project/state)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Project    │  │  Collections │  │   Load/Save  │         │
│  │ (+ methods)  │  │  (PhaseCol,  │  │   (YAML I/O) │         │
│  │              │  │  ArtifactCol,│  │              │         │
│  │              │  │   TaskCol)   │  │              │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└────────────────────────┬────────────────────────────────────────┘
                         │ embeds
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│              CUE-Generated Data Types                           │
│                 (schemas/project)                               │
│  ProjectState, PhaseState, ArtifactState, TaskState            │
│  (Pure data, no methods)                                        │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Project Type Registry                         │
│              (internal/sdks/project)                            │
│  ┌──────────────────────────────────────────────────┐          │
│  │         ProjectTypeConfig                        │          │
│  │  (phases, transitions, guards, prompts)         │          │
│  │                                                   │          │
│  │  BuildMachine(project) → binds guards via       │          │
│  │                          closures                │          │
│  │  Validate(project) → CUE + metadata checks      │          │
│  └──────────────────────────────────────────────────┘          │
└────────────────────────┬────────────────────────────────────────┘
                         │ uses
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                  State Machine SDK                              │
│                (internal/sdks/state)                            │
│  Builder, Machine, Guards, Events, States                      │
│  (Forked from internal/project/statechart)                     │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│              Project Type Implementations                       │
│                 (internal/projects/*)                           │
│  ┌─────────────────────┐  ┌─────────────────────┐             │
│  │  Standard Project   │  │ Exploration Project │             │
│  │  - Uses SDK builder │  │  - Uses SDK builder │             │
│  │  - Defines states   │  │  - Defines states   │             │
│  │  - Embeds metadata  │  │  - Embeds metadata  │             │
│  │    CUE schemas      │  │    CUE schemas      │             │
│  └─────────────────────┘  └─────────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

**Key architectural decisions:**
1. **Data-only child types** - Phase/Artifact/Task are pure data with zero business logic
2. **Project orchestrates everything** - All state machines and operations managed by Project
3. **CUE for structure, runtime for metadata** - Universal schemas in CUE, per-type metadata embedded
4. **Two-tier validation** - CUE validates structure, runtime validates metadata against embedded schemas
5. **Template binding pattern** - Guards defined as templates, bound to project instances via closures
6. **Parallel packages** - New SDK in `internal/sdks/*`, existing code untouched until cutover

### State Management

**Universal Data Types:**
All project types use the same underlying data structures:

```go
// Generated from CUE schemas
type ProjectState struct {
    Name       string
    Type       string
    Branch     string
    Phases     map[string]PhaseState
    Statechart StatechartState
}

type PhaseState struct {
    Status      string
    Enabled     bool
    Metadata    map[string]interface{}
    Inputs      []ArtifactState
    Outputs     []ArtifactState
    Tasks       []TaskState
}
```

> **Note**: These universal types directly enable the [Unified Command Hierarchy Design](../../knowledge/designs/command-hierarchy-design.md). Since all project types share identical structure (Phases with Inputs/Outputs/Tasks), CLI commands can work uniformly across project types without specialized logic per type.

**Wrapper Pattern:**
Runtime types embed generated types and add methods:

```go
// internal/sdks/project/state/project.go
type Project struct {
    schemas.ProjectState  // Embed CUE-generated type

    // Runtime-only fields (not serialized)
    config  *ProjectTypeConfig
    machine *state.Machine
}

func (p *Project) Advance() error { /* ... */ }
func (p *Project) Save() error { /* ... */ }
```

**Collection Pattern:**
Collections provide structural operations, direct field access for mutations:

```go
type PhaseCollection map[string]*Phase

func (pc PhaseCollection) Get(name string) (*Phase, error) {
    phase, exists := pc[name]
    if !exists {
        return nil, fmt.Errorf("phase not found: %s", name)
    }
    return phase, nil
}

// Usage in CLI commands:
phase, _ := project.Phases.Get("implementation")
phase.Status = "in_progress"  // Direct mutation
```

**Metadata Flexibility:**
Common fields are type-safe, custom fields in metadata map:

```go
// Type-safe common fields (90% of access)
artifact.Approved = true
artifact.Path = "/path/to/file"

// Runtime-validated metadata (10% of access)
if artifact.Metadata == nil {
    artifact.Metadata = make(map[string]interface{})
}
artifact.Metadata["assessment"] = "pass"
```

**Event Determination (OnAdvance):**
Generic `sow advance` command requires project-type-specific event determination logic:

```go
// Configure per-state event determination
.OnAdvance(ReviewActive, func(p *state.Project) (Event, error) {
    // Project-type-specific logic
    // Examine review artifact assessment → determine EventReviewPass or EventReviewFail
    phase, _ := p.Phases.Get("review")
    assessment := phase.Outputs[0].Metadata["assessment"]

    if assessment == "pass" {
        return EventReviewPass, nil
    }
    return EventReviewFail, nil
})
```

**Why needed:** Different states may have:
- **Single event** - Simple states with one possible next event (e.g., PlanningActive → EventCompletePlanning)
- **Conditional events** - Complex states with branching logic (e.g., ReviewActive → EventReviewPass or EventReviewFail based on assessment)

OnAdvance functions are stored in `ProjectTypeConfig` and called by generic `Project.Advance()` method.

---

## Component Breakdown

### 1. State Machine SDK (`internal/sdks/state`)

**Responsibility:** Generic state machine implementation

**Status:** Forked from `internal/project/statechart` with potential minor improvements

**Key Components:**
- `Builder` - Fluent API for constructing state machines
- `Machine` - Runtime state machine with state transitions
- `Guards` - Condition functions that must pass for transitions
- `Events` - Named triggers for state transitions
- `States` - Named states in the state machine

**Note:** This is largely a copy of the existing statechart SDK. Placed in new package to allow breaking changes if needed during implementation. Full details not included in this design doc.

### 2. CUE Schemas (`schemas/project`)

**Responsibility:** Define universal project structure for validation

**Files:**
- `project.cue` - ProjectState definition
- `phase.cue` - PhaseState definition
- `artifact.cue` - ArtifactState definition
- `task.cue` - TaskState definition
- `cue_types_gen.go` - Generated Go types

**Key Behavior:** Code generation via `cue exp gengotypes` produces Go types used as embedded base for wrapper types

### 3. Project State Types (`internal/sdks/project/state`)

**Responsibility:** Runtime project state with operations

**Components:**

**Project:**
- Wraps `ProjectState` with runtime fields (config, machine)
- Provides `Advance()`, `Save()`, `Load()` operations
- Helper methods for common guard patterns

**Phase/Artifact/Task:**
- Wrap generated CUE types
- Pure data (no business logic)
- Accessed via collections

**Collections:**
- `PhaseCollection` - map-based with Get()
- `ArtifactCollection` - slice-based with Get(index), Add()
- `TaskCollection` - slice-based with Get(id), Add()

**Loader:**
- `Load(ctx)` - Deserialize into CUE types, validate, convert to wrappers, attach config, build machine
- `Save(project)` - Convert to CUE types, validate, serialize, atomic write
- Conversion layer between CUE-generated types (for validation) and wrapper types (for runtime)

### 4. Project SDK (`internal/sdks/project`)

**Responsibility:** Builder API and type configuration

**Components:**

**ProjectTypeConfigBuilder:**
- Fluent API for defining project types
- Methods: `WithPhase()`, `SetInitialState()`, `AddTransition()`, `OnAdvance()`, `WithPrompt()`
- Returns `ProjectTypeConfig` on `Build()`

**ProjectTypeConfig:**
- Stores phase configs, transitions, event determiners, prompts
- `BuildMachine(project, initialState)` - Binds guards to project instance
- `GetEventDeterminer(state)` - Returns event determiner function for given state
- `Validate(project)` - Two-tier validation (structure + metadata)

**Registry:**
- Global registry of project types
- `Register(typeName, config)`
- Populated at app startup via `init()`

**Options:**
- `PhaseOpt` - Phase configuration options (`WithStartState`, `WithMetadataSchema`, etc.)
- `TransitionOption` - Transition options (`WithGuard`, `WithOnEntry`, `WithOnExit`)

### 5. Project Type Implementations (`internal/projects/*`)

**Responsibility:** Define specific project types using SDK

**Structure per project type:**
```
internal/projects/standard/
├── cue/                        # Metadata schemas (embedded)
│   ├── implementation_metadata.cue
│   ├── review_metadata.cue
│   └── finalize_metadata.cue
├── standard.go                 # NewStandardProjectConfig()
├── metadata.go                 # Embed CUE files
├── states.go                   # State constants
└── events.go                   # Event constants
```

**Key Behavior:**
- `NewStandardProjectConfig()` uses SDK builder
- Registers config with registry in `init()`
- Embeds metadata CUE schemas for validation
- Defines state/event constants
- Configures event determiners via `OnAdvance()` for each state (enables generic `sow advance` command)

---

## Data Models

### CUE Schema Definitions

#### schemas/project/project.cue

```cue
package project

import "time"

#ProjectState: {
    // Project identification
    name: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"
    type: string & =~"^[a-z0-9_]+$"
    branch: string & !=""
    description?: string

    // Timestamps
    created_at: time.Time
    updated_at: time.Time

    // Phases (map of phase name to phase state)
    phases: [string]: #PhaseState

    // State machine position
    statechart: #StatechartState
}

#StatechartState: {
    current_state: string & !=""
    updated_at: time.Time
}
```

#### schemas/project/phase.cue

```cue
package project

import "time"

#PhaseState: {
    // Status tracking
    status: string & !=""
    enabled: bool

    // Timestamps
    created_at: time.Time
    started_at?: time.Time
    completed_at?: time.Time

    // Metadata (validated by project type)
    metadata?: {...}

    // Collections
    inputs: [...#ArtifactState]
    outputs: [...#ArtifactState]
    tasks: [...#TaskState]
}
```

#### schemas/project/artifact.cue

```cue
package project

import "time"

#ArtifactState: {
    // Core fields
    type: string & !=""
    path: string & !=""
    approved: bool

    // Timestamps
    created_at: time.Time

    // Metadata (project-type specific)
    metadata?: {...}
}
```

#### schemas/project/task.cue

```cue
package project

import "time"

#TaskState: {
    // Identification
    id: string & =~"^[0-9]{3}$"
    name: string & !=""
    phase: string & !=""  // Which phase this task belongs to

    // Status
    status: "pending" | "in_progress" | "completed" | "abandoned"

    // Timestamps
    created_at: time.Time
    started_at?: time.Time
    updated_at: time.Time
    completed_at?: time.Time

    // Iteration and assignment
    iteration: int & >=1
    assigned_agent: string & !=""

    // Input artifacts (consumed by this task)
    // Examples: references, feedback
    inputs: [...#ArtifactState]

    // Output artifacts (produced by this task)
    // Examples: modified files
    outputs: [...#ArtifactState]

    // Metadata (project-type specific)
    metadata?: {...}
}
```

**Note**: Task state lives entirely in the project state file. Task folders (`.sow/project/phases/<phase>/tasks/<id>/`) still exist but contain only `description.md`, `log.md`, and `feedback/` directory - no `state.yaml` file.

### Example Metadata Schema

#### internal/projects/standard/cue/implementation_metadata.cue

```cue
package standard

// Metadata for implementation phase
{
    tasks_approved?: bool
    complexity?: "low" | "medium" | "high"
}
```

---

## APIs and Interfaces

### Project SDK Builder API

The complete builder API for defining project types:

```go
package project

// ===== Entry Point =====

func NewProjectTypeConfigBuilder(name string) *ProjectTypeConfigBuilder

// ===== Phase Configuration =====

func (b *ProjectTypeConfigBuilder) WithPhase(
    name string,
    opts ...PhaseOpt,
) *ProjectTypeConfigBuilder

// Phase Options (function type: func(*PhaseConfig))
func WithStartState(state State) PhaseOpt
func WithEndState(state State) PhaseOpt
func WithInputs(allowedTypes ...string) PhaseOpt      // Empty = allow all types
func WithOutputs(allowedTypes ...string) PhaseOpt     // Empty = allow all types
func WithTasks() PhaseOpt                             // Enable task support
func WithMetadataSchema(cueSchema string) PhaseOpt    // Embedded CUE schema

// ===== State Machine Configuration =====

func (b *ProjectTypeConfigBuilder) SetInitialState(state State) *ProjectTypeConfigBuilder

func (b *ProjectTypeConfigBuilder) AddTransition(
    from State,
    to State,
    event Event,
    opts ...TransitionOption,
) *ProjectTypeConfigBuilder

// Transition Options (function type: func(*TransitionConfig))
func WithGuard(guardFunc func(*Project) bool) TransitionOption
func WithOnEntry(action func(*Project) error) TransitionOption
func WithOnExit(action func(*Project) error) TransitionOption

// Event Determination (for generic Advance() command)
func (b *ProjectTypeConfigBuilder) OnAdvance(
    state State,
    determiner func(*Project) (Event, error),
) *ProjectTypeConfigBuilder

// ===== Prompt Configuration =====

func (b *ProjectTypeConfigBuilder) WithPrompt(
    state State,
    generator func(*Project) string,
) *ProjectTypeConfigBuilder

// ===== Build =====

func (b *ProjectTypeConfigBuilder) Build() *ProjectTypeConfig
```

### Project State Operations API

```go
package state

// ===== Project Operations =====

func Load(ctx context.Context) (*Project, error)
func (p *Project) Save() error
func (p *Project) Advance() error

// Helper methods for guards
func (p *Project) PhaseOutputApproved(phaseName, outputType string) bool
func (p *Project) PhaseMetadataBool(phaseName, key string) bool
func (p *Project) AllTasksComplete() bool

// ===== Collection Operations =====

// PhaseCollection
func (pc PhaseCollection) Get(name string) (*Phase, error)

// ArtifactCollection
func (ac ArtifactCollection) Get(index int) (*Artifact, error)
func (ac *ArtifactCollection) Add(artifact Artifact) error
func (ac *ArtifactCollection) Remove(index int) error

// TaskCollection
func (tc TaskCollection) Get(id string) (*Task, error)
func (tc *TaskCollection) Add(task Task) error
func (tc *TaskCollection) Remove(id string) error
```

### Complete Usage Example

**Standard project type definition:**

```go
package standard

import (
    "github.com/jmgilman/sow/cli/internal/sdks/project"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// Embedded metadata schemas
//go:embed cue/implementation_metadata.cue
var implementationMetadataSchema string

//go:embed cue/review_metadata.cue
var reviewMetadataSchema string

// Register on startup
func init() {
    project.Register("standard", NewStandardProjectConfig())
}

// Define complete project type configuration
func NewStandardProjectConfig() *project.ProjectTypeConfig {
    return project.NewProjectTypeConfigBuilder("standard").

        // Planning phase
        WithPhase("planning",
            project.WithStartState(PlanningActive),
            project.WithEndState(PlanningActive),
            project.WithOutputs("task_list"),
        ).

        // Implementation phase
        WithPhase("implementation",
            project.WithStartState(ImplementationPlanning),
            project.WithEndState(ImplementationExecuting),
            project.WithTasks(),
            project.WithMetadataSchema(implementationMetadataSchema),
        ).

        // Review phase
        WithPhase("review",
            project.WithStartState(ReviewActive),
            project.WithEndState(ReviewActive),
            project.WithOutputs("review"),
            project.WithMetadataSchema(reviewMetadataSchema),
        ).

        // State machine
        SetInitialState(NoProject).

        AddTransition(
            NoProject,
            PlanningActive,
            EventProjectInit,
        ).

        AddTransition(
            PlanningActive,
            ImplementationPlanning,
            EventCompletePlanning,
            project.WithGuard(func(p *state.Project) bool {
                return p.PhaseOutputApproved("planning", "task_list")
            }),
        ).

        AddTransition(
            ImplementationPlanning,
            ImplementationExecuting,
            EventTasksApproved,
            project.WithGuard(func(p *state.Project) bool {
                return p.PhaseMetadataBool("implementation", "tasks_approved")
            }),
        ).

        AddTransition(
            ImplementationExecuting,
            ReviewActive,
            EventAllTasksComplete,
            project.WithGuard(func(p *state.Project) bool {
                return p.AllTasksComplete()
            }),
        ).

        // Prompts
        WithPrompt(PlanningActive, func(p *state.Project) string {
            return "Planning phase active. Create and approve task list."
        }).

        WithPrompt(ImplementationPlanning, func(p *state.Project) string {
            return "Review and approve implementation tasks."
        }).

        WithPrompt(ImplementationExecuting, func(p *state.Project) string {
            return "Execute implementation tasks."
        }).

        WithPrompt(ReviewActive, func(p *state.Project) string {
            return "Review implementation and provide assessment."
        }).

        // Event Determination (for sow advance)
        OnAdvance(PlanningActive, func(p *state.Project) (Event, error) {
            // Simple case - single event from this state
            return EventCompletePlanning, nil
        }).

        OnAdvance(ImplementationPlanning, func(p *state.Project) (Event, error) {
            return EventTasksApproved, nil
        }).

        OnAdvance(ImplementationExecuting, func(p *state.Project) (Event, error) {
            return EventAllTasksComplete, nil
        }).

        OnAdvance(ReviewActive, func(p *state.Project) (Event, error) {
            // Complex case - conditional logic based on review assessment
            phase, err := p.Phases.Get("review")
            if err != nil {
                return "", err
            }

            // Find latest approved review artifact
            var latestReview *state.Artifact
            for i := len(phase.Outputs) - 1; i >= 0; i-- {
                artifact := &phase.Outputs[i]
                if artifact.Type == "review" && artifact.Approved {
                    latestReview = artifact
                    break
                }
            }

            if latestReview == nil {
                return "", fmt.Errorf("no approved review artifact found")
            }

            // Extract assessment from metadata
            assessment, ok := latestReview.Metadata["assessment"].(string)
            if !ok {
                return "", fmt.Errorf("review artifact missing assessment")
            }

            // Determine event based on assessment
            switch assessment {
            case "pass":
                return EventReviewPass, nil
            case "fail":
                return EventReviewFail, nil
            default:
                return "", fmt.Errorf("invalid assessment: %s", assessment)
            }
        }).

        Build()
}
```

---

## Load and Save Implementation

### Architecture: Conversion Layer

Load and Save use a conversion layer between CUE-generated types (for validation) and wrapper types (for runtime operations):

```
YAML File → CUE Types → Validate → Convert → Wrapper Types (runtime)
           ↑                                              ↓
           └─────────── Convert ← Validate ←─────────────┘
```

**Why conversion layer:**
- CUE validation works on exact schema types (generated from CUE)
- Wrapper types provide runtime API (collections, methods)
- Separation of concerns: validation vs operations

### Load() Implementation

**File:** `internal/sdks/project/state/loader.go`

```go
func Load(ctx context.Context) (*Project, error) {
    // 1. Read YAML file
    path := filepath.Join(ctx.WorkingDir(), ".sow/project/state.yaml")
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read state: %w", err)
    }

    // 2. Unmarshal into CUE-generated type
    var projectState schemas.ProjectState
    if err := yaml.Unmarshal(data, &projectState); err != nil {
        return nil, fmt.Errorf("failed to unmarshal: %w", err)
    }

    // 3. Validate structure with CUE
    if err := validateStructure(projectState); err != nil {
        return nil, fmt.Errorf("CUE validation failed: %w", err)
    }

    // 4. Convert to wrapper types
    project := &Project{
        Name:       projectState.Name,
        Type:       projectState.Type,
        Branch:     projectState.Branch,
        Phases:     convertPhases(projectState.Phases),
        Statechart: projectState.Statechart,
    }

    // 5. Lookup and attach type config
    config, exists := Registry[project.Type]
    if !exists {
        return nil, fmt.Errorf("unknown project type: %s", project.Type)
    }
    project.config = config

    // 6. Build state machine initialized with current state
    initialState := State(project.Statechart.CurrentState)
    project.machine = config.BuildMachine(project, initialState)

    // 7. Validate metadata against embedded schemas
    if err := config.Validate(project); err != nil {
        return nil, fmt.Errorf("metadata validation failed: %w", err)
    }

    return project, nil
}
```

### Save() Implementation

**File:** `internal/sdks/project/state/loader.go`

```go
func (p *Project) Save() error {
    // 1. Sync statechart state from machine
    if p.machine != nil {
        p.Statechart.CurrentState = p.machine.State().String()
        p.Statechart.UpdatedAt = time.Now()
    }

    // 2. Convert wrapper types to CUE-generated types
    projectState := schemas.ProjectState{
        Name:       p.Name,
        Type:       p.Type,
        Branch:     p.Branch,
        Phases:     convertPhasesToState(p.Phases),
        Statechart: p.Statechart,
    }

    // 3. Validate structure with CUE
    if err := validateStructure(projectState); err != nil {
        return fmt.Errorf("CUE validation failed: %w", err)
    }

    // 4. Validate metadata with embedded schemas
    if err := p.config.Validate(p); err != nil {
        return fmt.Errorf("metadata validation failed: %w", err)
    }

    // 5. Marshal to YAML
    data, err := yaml.Marshal(projectState)
    if err != nil {
        return fmt.Errorf("failed to marshal: %w", err)
    }

    // 6. Atomic write (temp file + rename)
    path := filepath.Join(p.workingDir, ".sow/project/state.yaml")
    tmpPath := path + ".tmp"

    if err := os.WriteFile(tmpPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }

    if err := os.Rename(tmpPath, path); err != nil {
        return fmt.Errorf("failed to rename temp file: %w", err)
    }

    return nil
}
```

### Conversion Functions

**File:** `internal/sdks/project/state/convert.go`

```go
func convertPhases(statePhases map[string]schemas.PhaseState) PhaseCollection {
    coll := make(PhaseCollection)
    for name, ps := range statePhases {
        coll[name] = &Phase{
            Status:      ps.Status,
            Enabled:     ps.Enabled,
            CreatedAt:   ps.CreatedAt,
            StartedAt:   ps.StartedAt,
            CompletedAt: ps.CompletedAt,
            Metadata:    ps.Metadata,
            Inputs:      convertArtifacts(ps.Inputs),
            Outputs:     convertArtifacts(ps.Outputs),
            Tasks:       convertTasks(ps.Tasks),
        }
    }
    return coll
}

func convertPhasesToState(phases PhaseCollection) map[string]schemas.PhaseState {
    statePhases := make(map[string]schemas.PhaseState)
    for name, p := range phases {
        statePhases[name] = schemas.PhaseState{
            Status:      p.Status,
            Enabled:     p.Enabled,
            CreatedAt:   p.CreatedAt,
            StartedAt:   p.StartedAt,
            CompletedAt: p.CompletedAt,
            Metadata:    p.Metadata,
            Inputs:      convertArtifactsToState(p.Inputs),
            Outputs:     convertArtifactsToState(p.Outputs),
            Tasks:       convertTasksToState(p.Tasks),
        }
    }
    return statePhases
}

func convertArtifacts(stateArtifacts []schemas.ArtifactState) ArtifactCollection {
    coll := make(ArtifactCollection, len(stateArtifacts))
    for i, sa := range stateArtifacts {
        coll[i] = Artifact{
            Type:      sa.Type,
            Path:      sa.Path,
            Approved:  sa.Approved,
            CreatedAt: sa.CreatedAt,
            Metadata:  sa.Metadata,
        }
    }
    return coll
}

func convertTasks(stateTasks []schemas.TaskState) TaskCollection {
    coll := make(TaskCollection, len(stateTasks))
    for i, st := range stateTasks {
        coll[i] = Task{
            ID:            st.ID,
            Name:          st.Name,
            Phase:         st.Phase,
            Status:        st.Status,
            CreatedAt:     st.CreatedAt,
            StartedAt:     st.StartedAt,
            UpdatedAt:     st.UpdatedAt,
            CompletedAt:   st.CompletedAt,
            Iteration:     st.Iteration,
            AssignedAgent: st.AssignedAgent,
            Inputs:        convertArtifacts(st.Inputs),
            Outputs:       convertArtifacts(st.Outputs),
            Metadata:      st.Metadata,
        }
    }
    return coll
}

// Reverse conversions follow same pattern
```

### CUE Validation

**File:** `internal/sdks/project/state/validate.go`

```go
func validateStructure(projectState schemas.ProjectState) error {
    // Use CUE to validate the generated type matches schema
    ctx := cuecontext.New()

    // Load embedded project schema
    schema := ctx.CompileString(embeddedProjectSchema)
    if schema.Err() != nil {
        return fmt.Errorf("invalid schema: %w", schema.Err())
    }

    // Encode project state to CUE value
    value := ctx.Encode(projectState)

    // Unify and validate
    result := schema.Unify(value)
    if err := result.Validate(cue.Concrete(true)); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    return nil
}
```

**Key points:**
- Conversion is simple field copying (one-time overhead per Load/Save)
- CUE validation happens on generated types (exact schema match)
- Wrapper types used exclusively at runtime (clean API)
- Collections serialize/deserialize as their underlying types (map/slice)

---

## Data Flow: CLI Command Lifecycle

### Simple Data Mutation Flow

**Command:** `sow output set --index 0 approved true`

**Execution flow:**

1. **Load Project**
   ```go
   project, err := state.Load(ctx)
   // - Reads .sow/project/state.yaml
   // - Deserializes into ProjectState (CUE-generated type)
   // - Validates structure with CUE schema
   // - Converts to wrapper types (Project with collections)
   // - Looks up type config from registry
   // - Builds state machine initialized with current state
   // - Validates metadata against embedded schemas
   ```

2. **Navigate Structure**
   ```go
   phase, err := project.Phases.Get("implementation")
   output, err := phase.Outputs.Get(0)
   ```

3. **Mutate Data**
   ```go
   output.Approved = true  // Direct field mutation
   ```

4. **Save**
   ```go
   err := project.Save()
   // - Syncs statechart.current_state from machine
   // - Converts wrapper types to CUE-generated types
   // - Validates structure with CUE schema
   // - Validates metadata against embedded schemas
   // - Serializes to YAML
   // - Atomic write (temp file + rename)
   ```

**Key points:**
- No state machine involved
- Collections provide navigation
- Direct field mutation (type-safe)
- Validation on save

### State Machine Flow

**Command:** `sow advance`

**Execution flow:**

1. **Load Project** (same as above)

2. **Determine Next Event**
   ```go
   // Get current state
   currentState := project.machine.State()

   // Get event determiner for this state (configured via OnAdvance)
   determiner := project.config.GetEventDeterminer(currentState)
   if determiner == nil {
       return fmt.Errorf("no event determiner for state: %s", currentState)
   }

   // Call project-type-specific determiner function
   event, err := determiner(project)
   // Returns: EventCompletePlanning (or error if conditions not met)
   ```

3. **Check Guard**
   ```go
   can, err := project.machine.CanFire(EventCompletePlanning)
   // Calls bound guard:
   //   func() bool {
   //     return guardFunc(project)  // project captured in closure
   //   }
   // Guard sees live project state
   ```

4. **Fire Event**
   ```go
   err := project.machine.Fire(EventCompletePlanning)
   // Executes:
   //   1. OnExit(PlanningActive)
   //      - Automatic: planning.Status = "completed"
   //   2. Transition to ImplementationPlanning
   //   3. OnEntry(ImplementationPlanning)
   //      - Automatic: implementation.Status = "in_progress"
   //      - Prompt generation
   ```

5. **Save** (same as above)

**Key points:**
- Machine built once during Load()
- Event determination is project-type-specific (configured via OnAdvance)
- Guards bound to project instance via closures (check conditions before transition)
- Phase status updates automatic
- Prompts generated on state entry

### Metadata Mutation Flow

**Command:** `sow phase metadata set implementation tasks_approved true`

**Execution flow:**

1. **Load Project** (same as above)

2. **Navigate and Mutate**
   ```go
   phase, err := project.Phases.Get("implementation")
   if phase.Metadata == nil {
       phase.Metadata = make(map[string]interface{})
   }
   phase.Metadata["tasks_approved"] = true
   ```

3. **Save**
   ```go
   err := project.Save()
   // During validation:
   //   1. CUE validates phase structure
   //   2. Looks up metadata schema for implementation phase
   //   3. Validates metadata against embedded CUE schema
   //   4. Returns error if validation fails
   ```

**Key points:**
- Metadata access explicit (not automatic routing)
- Type assertion on retrieval
- CUE validation on save

### Task Mutation Flow

**Command:** `sow task set --id 010 status completed`

**Execution flow:**

1. **Load Project** (same as above)

2. **Navigate and Mutate**
   ```go
   phase, err := project.Phases.Get("implementation")
   task, err := phase.Tasks.Get("010")
   task.Status = "completed"
   task.CompletedAt = time.Now()
   ```

3. **Save**
   ```go
   err := project.Save()
   // - Validates entire project state (including all tasks)
   // - Atomic write of single state file
   ```

**Key points:**
- All task state in project state file (not separate task state files)
- Single atomic Save() operation
- Task folders still exist for description.md, log.md, feedback/ (not for state)
- Consistent with unified command hierarchy (sow task commands operate on project state)

---

## Error Handling

### Load Errors

| Error | Handling |
|-------|----------|
| File not found | Return `ErrNoProject` (expected) |
| YAML parse error | Return error with line number |
| Unknown project type | Return error listing available types |
| CUE validation failure | Return error with CUE validation details |
| Metadata schema invalid | Return error with schema parse details |

### Save Errors

| Error | Handling |
|-------|----------|
| CUE validation failure | Fail save, return validation error (prevents corrupt state) |
| Metadata validation failure | Fail save, return schema validation error |
| Disk write error | Return error (state unchanged) |
| Atomic rename failure | Retry once, return error if fails |

### State Machine Errors

| Error | Handling |
|-------|----------|
| Guard fails | Return `ErrCannotAdvance` with reason |
| Invalid event | Return error listing valid events |
| OnEntry/OnExit action fails | Rollback transition, return action error |

**Design principle:** Fail fast on validation. Better to catch errors during Save() than persist invalid state.

---

## Cross-Cutting Concerns

### Validation Strategy

**Two-tier validation:**

1. **Structure Validation (CUE)**
   - Validates universal fields (name, type, status, etc.)
   - Validates collection structure (inputs, outputs, tasks)
   - Validates required fields present
   - Runs on: Load() and Save()

2. **Metadata Validation (Embedded CUE)**
   - Validates project-type-specific metadata
   - Schemas embedded in project type packages
   - Only validates phases with metadata schemas
   - Runs on: Save() only

**Validation timing:**
- **Load()**: CUE structure validation only (metadata may be from old version)
- **Save()**: Full validation (structure + metadata)

### Metadata Schema Management

**Embedding pattern:**

```go
package standard

import _ "embed"

//go:embed cue/implementation_metadata.cue
var implementationMetadataSchema string

func NewStandardProjectConfig() *project.ProjectTypeConfig {
    return project.NewProjectTypeConfigBuilder("standard").
        WithPhase("implementation",
            project.WithMetadataSchema(implementationMetadataSchema),
        ).
        // ...
        Build()
}
```

**Validation implementation:**

```go
func (ptc *ProjectTypeConfig) Validate(project *Project) error {
    // ... structure validation ...

    for phaseName, phaseConfig := range ptc.phaseConfigs {
        phase := project.Phases[phaseName]

        if phaseConfig.metadataSchema != "" {
            if err := validateMetadata(
                phase.Metadata,
                phaseConfig.metadataSchema,
            ); err != nil {
                return fmt.Errorf("phase %s: %w", phaseName, err)
            }
        } else if len(phase.Metadata) > 0 {
            return fmt.Errorf("phase %s does not support metadata", phaseName)
        }
    }

    return nil
}
```

---

## Implementation Plan

### Phase 1: Foundation (Week 1-2)

**Deliverables:**
- CUE schemas defined (`schemas/project/*.cue`)
- Code generation working (`cue_types_gen.go`)
- State machine SDK copied to `internal/sdks/state`
- Basic wrapper types (`internal/sdks/project/state`)

**Dependencies:** None

**Validation:** Generated types compile, unit tests pass

### Phase 2: SDK Builder (Week 2-3)

**Deliverables:**
- ProjectTypeConfigBuilder implementation
- Options pattern (PhaseOpt, TransitionOption)
- Registry and registration
- BuildMachine() with closure binding
- Validate() with two-tier validation

**Dependencies:** Phase 1 complete

**Validation:** Standard project type configurable via builder

### Phase 3: State Operations (Week 3-4)

**Deliverables:**
- Load() implementation (deserialize + attach config + build machine)
- Save() implementation (validate + serialize + atomic write)
- Collection implementations (Get, Add, Remove)
- Helper methods on Project

**Dependencies:** Phase 2 complete

**Validation:** Load/Save round-trip preserves state, collections work

### Phase 4: Project Types (Week 4-5)

**Deliverables:**
- Standard project using SDK (`internal/projects/standard`)
- Exploration project using SDK (`internal/projects/exploration`)
- Metadata schemas embedded
- State/event constants defined

**Dependencies:** Phase 3 complete

**Validation:** Both project types fully functional with new SDK

### Phase 5: CLI Migration (Week 5-6)

**Deliverables:**
- Update CLI commands to use `internal/sdks/project/state`
- Remove dependencies on `internal/project`
- Integration tests passing
- Documentation updated

**Dependencies:** Phase 4 complete

**Validation:** All existing tests pass, no regressions

### Phase 6: Cleanup (Week 6)

**Deliverables:**
- Delete old `internal/project` package
- Delete old `schemas/projects` package
- Final verification
- Release notes

**Dependencies:** Phase 5 complete

**Validation:** Clean build, no dead code

---

## Migration and Rollout Strategy

### Migration Approach

**Parallel Implementation:**
- New SDK built in `internal/sdks/*` packages
- Old implementation in `internal/project` untouched during development
- Breaking changes acceptable (sow in early development)

**Cutover Strategy:**
1. New project types implemented in `internal/projects/*`
2. CLI commands updated to import new packages
3. Old packages deleted after verification
4. Fresh start for active projects (recreate with new SDK)

**Rollback Plan:**
- If critical bugs found, temporarily revert CLI commands to old imports
- New SDK packages isolated, can be fixed without affecting old code
- No data migration utilities (acceptable in early development phase)

### Breaking Changes

**YAML Format:**
- YAML structure may change based on implementation needs
- CUE schema changes will naturally evolve the format
- No migration utilities provided (users recreate projects)

**Impact:**
- Active projects may need to be recreated
- Acceptable since sow has no public usage yet
- Clean slate approach simpler than migration complexity

**Mitigation:**
- Document YAML format changes in release notes
- Provide examples of new format
- Keep cutover timing flexible (when no critical active projects)

### Risk Mitigation

**Testing:**
- Comprehensive unit tests for SDK components
- Integration tests using actual project states
- Load tests for performance regression

**Phased Rollout:**
- Week 1-4: Development in parallel packages (low risk)
- Week 5: CLI migration (medium risk, extensive testing)
- Week 6: Cleanup (low risk)

**Monitoring:**
- Track Load/Save latency (target < 50ms)
- Monitor validation errors (catch schema issues)
- Track state machine transition failures

---

## Testing Strategy

### Unit Tests

**SDK Builder:**
- Phase configuration options apply correctly
- Transition options bind guards/actions
- Build() validates configuration (missing states, etc.)
- Invalid configurations rejected

**Collections:**
- Get() returns correct items
- Get() with invalid index/key returns error
- Add() appends to collections
- Remove() removes from collections

**Validation:**
- CUE validation catches structural errors
- Metadata validation catches schema violations
- Phases without metadata schemas reject metadata

### Integration Tests

**Load/Save:**
- Round-trip preserves all data
- Atomic write prevents corruption
- Validation errors prevent saves
- Invalid YAML returns clear errors

**State Machine:**
- Guards prevent invalid transitions
- Phase status updates automatic
- OnEntry/OnExit actions execute
- Prompts generated correctly

**Project Types:**
- Standard project full lifecycle
- Exploration project full lifecycle
- All guards work correctly
- All transitions allowed/blocked appropriately

### Performance Tests

**Load Performance:**
- Typical project state loads in < 50ms
- Large project states (100 tasks) load in < 100ms

**Save Performance:**
- Save with validation completes in < 50ms
- Validation overhead < 10ms

**State Machine:**
- Transition execution < 5ms
- Guard evaluation < 1ms

---

## Alternatives Considered

### Alternative 1: Pure Go Without CUE

**Description:** Remove CUE entirely, use go-playground/validator for all validation

**Pros:**
- Single language (Go only)
- Better IDE support
- Lower learning curve
- Simpler build process

**Cons:**
- Inconsistent with other schemas (breakdown_index, design_index, etc. all use CUE)
- Lose external validation capability
- Weaker metadata validation (tags vs CUE constraints)
- Manual type definitions instead of generation

**Why not chosen:** Consistency with existing schemas is critical. Every other index file uses CUE. Breaking this pattern for projects would create special-case architecture.

### Alternative 2: Direct Methods on CUE Types

**Description:** Add methods directly to generated CUE types instead of wrapping

**Pros:**
- No wrapper types needed
- Simpler type hierarchy
- Less embedding complexity

**Cons:**
- Business logic in schemas package
- Circular dependency risk (schemas imports project SDK)
- Generated code mixed with hand-written code
- Violates separation of concerns

**Why not chosen:** Schemas should be pure data definitions. Mixing generated and hand-written code in same package is fragile and confusing.

### Alternative 3: Separate Fields Instead of Metadata Map

**Description:** Define all phase fields as typed struct fields (no metadata map)

**Pros:**
- Compile-time type safety everywhere
- No type assertions needed
- Better IDE autocomplete

**Cons:**
- Field duplication across project types
- Cannot reuse code (each type has different struct)
- Adding custom fields requires changing universal schema
- Loses flexibility for future project types

**Why not chosen:** Code reuse and flexibility more important than compile-time safety for 10% of field access. Common fields (90% of access) remain type-safe.

---

## Open Questions

- [ ] Should metadata validation use go-playground/validator in addition to CUE for runtime checks?
- [ ] How do we handle schema evolution for metadata (old states with outdated metadata)?
- [ ] Should collection methods return pointers or values (currently returning pointers)?
- [ ] Do we need a way to define task metadata schemas similarly to phase metadata?
- [ ] Should Load() skip metadata validation to support old state files gracefully?

---

## References

- [ADR 004: Introduce Project SDK Architecture](../adrs/004-introduce-project-sdk-architecture.md) - Architecture decision record documenting this design
- [Unified Command Hierarchy Design](./command-hierarchy-design.md) - This SDK makes implementing the unified command hierarchy straightforward by providing universal Project/Phase/Artifact/Task types that work across all project types
- Current implementation: `cli/internal/project/`
- Current schemas: `cli/schemas/projects/`
- State machine implementation: `cli/internal/project/statechart/`

---

## Future Considerations

### Task and Artifact State Machines

This architecture enables future task/artifact state machines without breaking changes:

```go
func (p *Project) AdvanceTask(taskID string) error {
    taskMachine := p.config.BuildTaskMachine(taskID, p)
    event := determineNextTaskEvent(taskMachine)
    taskMachine.Fire(event)
    return p.Save()
}
```

Tasks remain pure data. Project orchestrates task machines. Guards can close over project state.

### Additional Enhancements

- JSON Schema generation from CUE for external tools
- Task metadata schema support (similar to phase metadata)
- Artifact metadata validation per-type
- Conditional phase features (metadata flags enabling/disabling tasks)
- Version migration system for metadata schemas
- GraphQL API for project state (using generated types)
