# Project Modes Implementation Design

**Author**: Architecture Team
**Date**: 2025-10-28
**Status**: Proposed
**Related ADR**: [ADR-003: Consolidate Operating Modes into Project Types](../knowledge/adrs/003-consolidate-modes-to-projects.md)

## Overview

This design describes how to consolidate sow's three operating modes (exploration, design, breakdown) into specialized project types within the unified project system. The implementation removes standalone mode directories and commands, introduces discriminated union project types, extends core schemas with six targeted improvements, and enables state machine workflows through phase-specific state values.

The exploration research (January 2025) validated this approach through comprehensive analysis of schema mappings, feature preservation, and state machine design. This document focuses on the implementation strategy.

## Goals and Non-Goals

**Goals**:
- Consolidate three mode implementations into unified project type system
- Implement seven schema and interface improvements enabling clean mode-to-project translation
- Remove all mode-specific code, directories, and CLI commands
- Support automatic project type detection via branch prefix
- Enable state machine workflows through SDK builder pattern with PhaseOperationResult

**Non-Goals** (explicitly out of scope):
- Migration tooling for existing mode sessions (breaking change, users restart)
- Backward compatibility with old mode indexes
- Gradual deprecation period (clean break acceptable)
- New project types beyond exploration/design/breakdown (future work)

## Design

### Implementation Strategy

Replace standalone mode systems with project type discrimination. Each mode becomes a specialized `ProjectState` schema variant. Branch prefix determines which project type to instantiate. Shared schemas (Phase, Task, Artifact) extend to support all mode requirements.

**Key approach**: Discriminated union pattern using CUE schemas. Base `#ProjectState` type discriminates on `project.type` field. Each variant defines mode-specific constraints while inheriting common project structure.

### Component Breakdown

#### 1. Schema Extensions

Extend base schemas (`schemas/phases/common.cue`) and Phase interface with seven improvements:

**Change 1: Artifact approval field**
```cue
#Artifact: {
    path: string
    description?: string
    approved?: bool  // NEW: enables design approval workflow
    created_at: time.Time
    metadata?: {[string]: _}
}
```

**Change 2: Phase inputs field**
```cue
#Phase: {
    status: string  // MODIFIED: remove constraint (see Change 6)
    enabled: bool
    created_at: time.Time
    started_at?: time.Time
    completed_at?: time.Time

    inputs?: [...#Artifact]   // NEW: tracks sources informing phase
    artifacts: [...#Artifact]
    tasks: [...#Task]
    metadata?: {[string]: _}
}
```

**Change 3: Task refs field**
```cue
#Task: {
    id: string & =~"^[0-9]{3,}$"
    name: string & !=""
    status: "pending" | "in_progress" | "completed" | "abandoned"
    parallel: bool
    dependencies?: [...string]

    refs?: [...#Artifact]     // NEW: exploration topic files
    metadata?: {[string]: _}  // NEW: breakdown GitHub metadata (Change 4)
}
```

**Change 5: Logging for journaling**

Reuse existing `sow agent log` command with new `action=journal` type:
```bash
sow agent log --action journal --result note --notes "Exploration insight..."
```

No schema change required. Exploration journaling uses existing logging infrastructure.

**Change 6: Unconstrained phase status**

Remove `status: "pending" | "in_progress" | "completed" | "skipped"` constraint. Allow project types to define meaningful states:
```cue
// Before (constrained):
#Phase: {
    status: "pending" | "in_progress" | "completed" | "skipped"
}

// After (unconstrained, project types constrain):
#Phase: {
    status: string  // Project type schemas constrain valid values
}
```

**Change 7: Phase interface signatures (PhaseOperationResult pattern)**

Update Phase interface methods to return `PhaseOperationResult` for operations that may trigger state machine events:

```go
// In cli/internal/project/domain/interfaces.go
type Phase interface {
    // Metadata
    Name() string
    Status() string
    Enabled() bool

    // Operations that may trigger events return PhaseOperationResult
    ApproveArtifact(path string) (*PhaseOperationResult, error)
    ApproveTasks() (*PhaseOperationResult, error)
    Complete() (*PhaseOperationResult, error)
    Set(field string, value interface{}) (*PhaseOperationResult, error)

    // Simple operations return just error
    AddArtifact(path string, opts ...ArtifactOption) error
    AddTask(name string, opts ...TaskOption) (*Task, error)
    Get(field string) (interface{}, error)
    // ... other existing methods
}
```

This enables phases to declaratively trigger state transitions while keeping CLI generic. See "PhaseOperationResult Pattern for Event Triggering" section below for details.

#### 2. Project Type Schemas

Create discriminated union with four project types:

**Base discriminator** (`schemas/projects/projects.cue`):
```cue
#ProjectState:
    | #StandardProjectState
    | #ExplorationProjectState
    | #DesignProjectState
    | #BreakdownProjectState
```

**Exploration project type** (`schemas/projects/exploration.cue`):
```cue
#ExplorationProjectState: {
    project: {
        type: "exploration"  // Discriminator field
        name: string
        branch: string
        description: string
        created_at: time.Time
        updated_at: time.Time
    }

    statechart: {
        current_state: "Gathering" | "Researching" | "Summarizing" | "Finalizing" | "Completed"
    }

    phases: {
        exploration: #Phase & {
            status: "gathering" | "researching" | "summarizing" | "completed"
            enabled: true
            // tasks track research topics/areas dynamically added during exploration
        }

        finalization: #Phase & {
            status: "pending" | "in_progress" | "completed"
            enabled: true
            // tasks track: create summary, move artifacts, create PR, cleanup
        }
    }
}
```

**Design project type** (`schemas/projects/design.cue`):
```cue
#DesignProjectState: {
    project: {
        type: "design"
        name: string
        branch: string
        description: string
        created_at: time.Time
        updated_at: time.Time
    }

    statechart: {
        current_state: "Drafting" | "Reviewing" | "Approved" | "Finalizing" | "Completed"
    }

    phases: {
        design: #Phase & {
            status: "drafting" | "reviewing" | "approved" | "completed"
            enabled: true
            // inputs and artifacts fields inherited from #Phase and #Artifact
            // tasks track individual documents to create (one task per document)
            // task completion tied to artifact approval
        }

        finalization: #Phase & {
            status: "pending" | "in_progress" | "completed"
            enabled: true
            // tasks track: move artifacts to targets, create PR, cleanup
        }
    }
}
```

**Breakdown project type** (`schemas/projects/breakdown.cue`):
```cue
#BreakdownProjectState: {
    project: {
        type: "breakdown"
        name: string
        branch: string
        description: string
        created_at: time.Time
        updated_at: time.Time
    }

    statechart: {
        current_state: "Decomposing" | "Documenting" | "Approving" | "Finalizing" | "Completed"
    }

    phases: {
        breakdown: #Phase & {
            status: "decomposing" | "documenting" | "approving" | "completed"
            enabled: true
            // tasks represent work units that will become GitHub issues
            // task.metadata stores issue-specific data
        }

        finalization: #Phase & {
            status: "pending" | "in_progress" | "completed"
            enabled: true
            // tasks track: create GitHub issues, link to tracking doc, create PR, cleanup
        }
    }
}
```

**Standard project type**: Unchanged, remains default for feature/bugfix work.

#### 3. Branch Prefix Detection

CLI automatically detects project type from branch prefix:

**Detection logic** (`internal/projects/types.go`):
```go
func DetectProjectType(branchName string) string {
    switch {
    case strings.HasPrefix(branchName, "explore/"):
        return "exploration"
    case strings.HasPrefix(branchName, "design/"):
        return "design"
    case strings.HasPrefix(branchName, "breakdown/"):
        return "breakdown"
    default:
        return "standard"  // Default
    }
}
```

**Project type determination**:
- Branch prefix automatically determines project type
- Type affects schema loaded, initial state, and phase configuration
- No manual type selection needed (convention-based)

**Note**: Interactive wizard flow for project creation/resumption is detailed in separate design document: `interactive-project-launch-design.md`

#### 4. State Machine Implementation

Each project type constructs its state machine using the SDK builder pattern. State machines are built at runtime when projects are loaded.

**Builder Pattern Usage**:

Project types use `MachineBuilder` to construct state machines with transitions, guards, and prompt generation:

```go
// In cli/internal/project/exploration/project.go
func (p *ExplorationProject) buildStateMachine() *statechart.Machine {
    // Get current state from project
    currentState := statechart.State(p.state.Statechart.Current_state)
    projectState := (*schemas.ProjectState)(p.state)

    // Create prompt generator (implements PromptGenerator interface)
    promptGen := NewExplorationPromptGenerator(p.ctx)

    // Create builder with initial state, project state, and prompt generator
    builder := statechart.NewBuilder(currentState, projectState, promptGen)

    // Configure all state transitions with guards using options pattern
    builder.
        // Gathering → Researching (with guard)
        AddTransition(
            ExplorationGathering,
            ExplorationResearching,
            EventCompleteGathering,
            statechart.WithGuard(func() bool {
                return GatheringComplete(projectState)
            }),
        ).
        // Researching → Summarizing (with guard)
        AddTransition(
            ExplorationResearching,
            ExplorationSummarizing,
            EventCompleteResearching,
            statechart.WithGuard(func() bool {
                return statechart.AllTasksComplete(projectState.Phases.Exploration.Tasks)
            }),
        ).
        // Summarizing → Finalizing (with guard)
        AddTransition(
            ExplorationSummarizing,
            ExplorationFinalizing,
            EventBeginFinalizing,
            statechart.WithGuard(func() bool {
                return SummaryApproved(projectState)
            }),
        ).
        // Finalizing → Completed (unconditional, no guard)
        AddTransition(
            ExplorationFinalizing,
            ExplorationCompleted,
            EventProjectComplete,
        )

    // Build and return final machine
    machine := builder.Build()
    machine.SetFilesystem(p.ctx.FS())

    return machine
}
```

**Guard Function Examples** (exploration):

```go
// In cli/internal/project/exploration/guards.go

// Minimum topics required to begin research
func GatheringComplete(state *schemas.ProjectState) bool {
    return len(state.Phases.Exploration.Tasks) >= 3
}

// Summary artifact exists and is approved
func SummaryApproved(state *schemas.ProjectState) bool {
    for _, a := range state.Phases.Exploration.Artifacts {
        if strings.Contains(a.Path, "summary") && a.Approved {
            return true
        }
    }
    return false
}
```

**Prompt Generation on State Entry**:

When the state machine transitions to a new state, it automatically calls `PromptGenerator.GeneratePrompt()`:

```go
// In cli/internal/project/exploration/prompts.go
type ExplorationPromptGenerator struct {
    components *statechart.PromptComponents
}

func (g *ExplorationPromptGenerator) GeneratePrompt(
    state statechart.State,
    projectState *schemas.ProjectState,
) (string, error) {
    // Route to state-specific generator
    switch state {
    case ExplorationGathering:
        return g.generateGatheringPrompt(projectState)
    case ExplorationResearching:
        return g.generateResearchingPrompt(projectState)
    case ExplorationSummarizing:
        return g.generateSummarizingPrompt(projectState)
    default:
        return "", fmt.Errorf("unknown state: %s", state)
    }
}

func (g *ExplorationPromptGenerator) generateGatheringPrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    // Use reusable SDK components
    buf.WriteString(g.components.ProjectHeader(projectState))
    gitStatus, _ := g.components.GitStatus()
    buf.WriteString(gitStatus)

    // Add exploration-specific sections
    buf.WriteString("\n## Current Research Topics\n\n")
    for _, task := range projectState.Phases.Exploration.Tasks {
        buf.WriteString(fmt.Sprintf("- %s (%s)\n", task.Name, task.Status))
    }

    // Include static guidance via template
    guidance, _ := g.components.RenderTemplate(
        prompts.PromptExplorationGathering,
        ctx,
    )
    buf.WriteString(guidance)

    return buf.String(), nil
}
```

**State Transition Flow**:

1. Phase calls `Complete()` → returns `WithEvent(EventCompleteGathering)`
2. CLI fires event: `machine.Fire(EventCompleteGathering)`
3. State machine evaluates guard: `GatheringComplete(projectState)`
4. If guard passes:
   - Transition to `ExplorationResearching` state
   - Call `PromptGenerator.GeneratePrompt(ExplorationResearching, projectState)`
   - Print generated prompt to user
   - Update `state.yaml` with new state
5. If guard fails:
   - Reject transition
   - Return error to CLI
   - State unchanged

**Benefits**:
- **Type-safe**: Compile-time checking of states and events
- **Flexible**: Guards can inspect any part of project state
- **Reusable**: Common guards provided by SDK
- **Dynamic prompts**: Access to git, GitHub, filesystem during generation
- **Testable**: Each component tested independently

#### 5. Code Removal

Delete mode-specific implementations:

**Directories to remove**:
- `internal/exploration/`
- `internal/design/`
- `internal/breakdown/`
- `.sow/exploration/` support
- `.sow/design/` support (note: current design mode is last usage)
- `.sow/breakdown/` support

**CLI commands to remove**:
- `sow exploration *`
- `sow design *` (except during this design session)
- `sow breakdown *`

**Commands to update**:
- `sow project` - add type detection and wizard for project creation/resumption
- CLI internally handles type-specific state and task metadata (no user-facing changes)

### PhaseOperationResult Pattern for Event Triggering

**Purpose**: Enable phases to declaratively trigger state machine events while keeping the CLI generic across all project types.

**Pattern**: Phase operations that may trigger state transitions return `PhaseOperationResult` containing an optional event. The CLI fires the event if present.

**Type Definition** (`cli/internal/project/domain/phase_operation_result.go`):

```go
// PhaseOperationResult represents the outcome of a phase operation.
type PhaseOperationResult struct {
    Event statechart.Event  // Optional event to fire (empty string = no event)
}

// NoEvent returns a result with no event to fire.
func NoEvent() *PhaseOperationResult {
    return &PhaseOperationResult{}
}

// WithEvent returns a result that will fire the given event.
func WithEvent(event statechart.Event) *PhaseOperationResult {
    return &PhaseOperationResult{Event: event}
}
```

**Updated Phase Interface** (`cli/internal/project/domain/interfaces.go`):

```go
type Phase interface {
    // Metadata
    Name() string
    Status() string
    Enabled() bool

    // Operations that may trigger events return PhaseOperationResult
    ApproveArtifact(path string) (*PhaseOperationResult, error)
    ApproveTasks() (*PhaseOperationResult, error)
    Complete() (*PhaseOperationResult, error)
    Set(field string, value interface{}) (*PhaseOperationResult, error)

    // Simple operations return just error
    AddArtifact(path string, opts ...ArtifactOption) error
    AddTask(name string, opts ...TaskOption) (*Task, error)
    Get(field string) (interface{}, error)
    // ... other methods
}
```

**Usage Examples**:

```go
// Exploration: Completing gathering phase triggers transition
func (p *GatheringPhase) Complete() (*domain.PhaseOperationResult, error) {
    p.state.Status = "completed"
    if err := p.project.Save(); err != nil {
        return nil, err
    }
    return domain.WithEvent(EventCompleteGathering), nil
}

// Design: Approving individual artifact may trigger transition
func (p *DraftingPhase) ApproveArtifact(path string) (*domain.PhaseOperationResult, error) {
    if err := p.artifacts.Approve(path); err != nil {
        return nil, err
    }

    // Check if ALL artifacts now approved
    allApproved := true
    for _, a := range p.state.Artifacts {
        if !a.Approved {
            allApproved = false
            break
        }
    }

    if allApproved {
        // Trigger transition to reviewing
        return domain.WithEvent(EventBeginReviewing), nil
    }

    return domain.NoEvent(), nil
}

// Breakdown: Set operation may trigger event based on metadata
func (p *BreakdownPhase) Set(field string, value interface{}) (*domain.PhaseOperationResult, error) {
    if p.state.Metadata == nil {
        p.state.Metadata = make(map[string]interface{})
    }
    p.state.Metadata[field] = value

    if err := p.project.Save(); err != nil {
        return nil, err
    }

    // Project-specific: decomposition_complete flag triggers transition
    if field == "decomposition_complete" && value == true {
        return domain.WithEvent(EventBeginDocumenting), nil
    }

    return domain.NoEvent(), nil
}
```

**Benefits**:
- **CLI stays generic**: No project-type conditionals, works for all types
- **Declarative events**: Phases specify when transitions occur without directly manipulating state machine
- **Flexible**: Complex logic can be in phase methods, guards validate pre-conditions
- **Testable**: Phase implementations can be unit tested independently
- **Consistent**: All phase operations follow same pattern

### Guard Validation Architecture

**Responsibility**: Guards are validated by the state machine when phase operations return events via the PhaseOperationResult pattern.

**Flow**:
1. Orchestrator calls `sow agent complete` (existing command)
2. CLI loads Phase interface implementation (e.g., `ExplorationPhase`, `DesignPhase`)
3. Phase's `Complete()` method:
   - Performs phase-specific validation
   - Updates phase state and saves
   - Returns `PhaseOperationResult` with event to fire
4. CLI checks if result contains an event
5. If event present: CLI fires event via `machine.Fire(result.Event)`
6. State machine checks guard function before transition
7. If guard passes: transition succeeds, state machine updates state
8. If guard fails: event rejected, state unchanged, error returned to CLI and orchestrator

**Implementation pattern** (follows State Machine SDK architecture):

```go
// In cli/internal/project/exploration/guards.go
// Guards are pure functions owned by the project type
func GatheringComplete(state *schemas.ProjectState) bool {
    // Guard inspects project state
    tasks := state.Phases.Exploration.Tasks
    return len(tasks) >= 3 // Minimum topic count requirement
}

// In cli/internal/project/exploration/gathering.go
func (p *GatheringPhase) Complete() (*domain.PhaseOperationResult, error) {
    // Phase-specific validation
    if len(p.state.Tasks) < 3 {
        return nil, fmt.Errorf("minimum 3 topics required, found %d", len(p.state.Tasks))
    }

    // Update phase status
    p.state.Status = "completed"
    now := time.Now()
    p.state.Completed_at = &now

    // Save state before returning event
    if err := p.project.Save(); err != nil {
        return nil, err
    }

    // Return event declaratively (CLI will fire it)
    return domain.WithEvent(EventCompleteGathering), nil
}

// In cli/internal/project/exploration/project.go
// Project type constructs state machine using builder pattern
func (p *ExplorationProject) buildStateMachine() *statechart.Machine {
    // Get current state from project
    currentState := statechart.State(p.state.Statechart.Current_state)
    projectState := (*schemas.ProjectState)(p.state)

    // Create prompt generator with full context access
    promptGen := NewExplorationPromptGenerator(p.ctx)

    // Create builder
    builder := statechart.NewBuilder(currentState, projectState, promptGen)

    // Configure all state transitions with guards
    builder.
        AddTransition(
            ExplorationGathering,
            ExplorationResearching,
            EventCompleteGathering,
            statechart.WithGuard(func() bool {
                return GatheringComplete(projectState)
            }),
        ).
        AddTransition(
            ExplorationResearching,
            ExplorationSummarizing,
            EventCompleteResearching,
            statechart.WithGuard(func() bool {
                return AllTasksComplete(projectState.Phases.Exploration.Tasks)
            }),
        )
        // ... more transitions

    // Build final machine
    machine := builder.Build()
    machine.SetFilesystem(p.ctx.FS())

    return machine
}

// In cli/cmd/agent/complete.go (CLI handles event firing)
result, err := phase.Complete()
if err != nil {
    return fmt.Errorf("failed to complete phase: %w", err)
}

// Fire event if phase returned one
if result.Event != "" {
    machine := proj.Machine()
    if err := machine.Fire(result.Event); err != nil {
        return fmt.Errorf("failed to fire event %s: %w", result.Event, err)
    }
    // Prompt generation happens automatically on state entry
}
```

**Key architectural points**:
- **Project types own their workflows**: Each type defines its own states, events, guards in separate package
- **Guards are pure functions**: Accept `*schemas.ProjectState`, return bool, no side effects
- **Builder pattern with closures**: Guards passed as closures capturing project state
- **PhaseOperationResult pattern**: Phases return events declaratively, CLI fires them
- **CLI stays generic**: No project-type conditionals, works for all types
- **SDK provides infrastructure**: MachineBuilder, PromptGenerator interface, PromptComponents, common guards
- **Automatic prompt generation**: PromptGenerator.GeneratePrompt() called on state entry

**Consistency with existing architecture**: This approach follows the State Machine SDK refactoring (January 2025). Reference implementation: `cli/internal/project/standard/` (states.go, events.go, guards.go, prompts.go, project.go). See `cli/DESIGN.md` for complete SDK architecture documentation.

### Project Type Implementation Structure

Each project type follows a consistent file organization pattern, with all project-specific code in a dedicated package.

**File Structure** (example: exploration project type):

```
cli/internal/project/exploration/
├── project.go          # Main struct, implements Project interface
│                      # Contains buildStateMachine() using builder pattern
│
├── states.go          # State constants specific to exploration workflow
│                      # Example: ExplorationGathering, ExplorationResearching
│
├── events.go          # Event constants that trigger transitions
│                      # Example: EventCompleteGathering, EventBeginResearching
│
├── guards.go          # Pure functions checking transition conditions
│                      # Example: GatheringComplete(state), AllTopicsResearched(state)
│
├── prompts.go         # PromptGenerator implementation
│                      # Generates state-specific prompts with external data access
│
├── gathering.go       # GatheringPhase implementation (Phase interface)
├── researching.go     # ResearchingPhase implementation
├── summarizing.go     # SummarizingPhase implementation
└── finalization.go    # FinalizationPhase implementation
```

**Component Responsibilities**:

**1. project.go** - Main project struct:
```go
type ExplorationProject struct {
    state   *projects.ExplorationProjectState
    ctx     *sow.Context
    machine *statechart.Machine
    phases  map[string]domain.Phase
}

// Implements Project interface
func (p *ExplorationProject) Name() string { ... }
func (p *ExplorationProject) CurrentPhase() domain.Phase { ... }
func (p *ExplorationProject) Machine() *statechart.Machine { ... }

// Constructs state machine using builder
func (p *ExplorationProject) buildStateMachine() *statechart.Machine {
    promptGen := NewExplorationPromptGenerator(p.ctx)
    builder := statechart.NewBuilder(currentState, projectState, promptGen)
    // Configure all transitions...
    return builder.Build()
}
```

**2. states.go** - State constants:
```go
const (
    ExplorationGathering    = statechart.State("Gathering")
    ExplorationResearching  = statechart.State("Researching")
    ExplorationSummarizing  = statechart.State("Summarizing")
    ExplorationFinalizing   = statechart.State("Finalizing")
    ExplorationCompleted    = statechart.State("Completed")
)
```

**3. events.go** - Event constants:
```go
const (
    EventCompleteGathering   = statechart.Event("complete_gathering")
    EventBeginResearching    = statechart.Event("begin_researching")
    EventCompleteSummarizing = statechart.Event("complete_summarizing")
    EventBeginFinalizing     = statechart.Event("begin_finalizing")
)
```

**4. guards.go** - Guard functions:
```go
// Pure functions, no side effects
func GatheringComplete(state *schemas.ProjectState) bool {
    return len(state.Phases.Exploration.Tasks) >= 3
}

func AllTopicsResearched(state *schemas.ProjectState) bool {
    return statechart.TasksComplete(state.Phases.Exploration.Tasks)
}

func SummaryApproved(state *schemas.ProjectState) bool {
    for _, a := range state.Phases.Exploration.Artifacts {
        if a.Path == "summary.md" && a.Approved {
            return true
        }
    }
    return false
}
```

**5. prompts.go** - Prompt generator:
```go
type ExplorationPromptGenerator struct {
    components *statechart.PromptComponents
}

func (g *ExplorationPromptGenerator) GeneratePrompt(
    state statechart.State,
    projectState *schemas.ProjectState,
) (string, error) {
    switch state {
    case ExplorationGathering:
        return g.generateGatheringPrompt(projectState)
    case ExplorationResearching:
        return g.generateResearchingPrompt(projectState)
    // ... other states
    }
}

func (g *ExplorationPromptGenerator) generateGatheringPrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    // Use reusable components
    buf.WriteString(g.components.ProjectHeader(projectState))
    buf.WriteString(g.components.GitStatus())

    // Add exploration-specific context
    buf.WriteString("## Research Topics\n\n")
    for _, task := range projectState.Phases.Exploration.Tasks {
        buf.WriteString(fmt.Sprintf("- %s (%s)\n", task.Name, task.Status))
    }

    // Render template for guidance
    guidance, _ := g.components.RenderTemplate(
        prompts.PromptExplorationGathering,
        ctx,
    )
    buf.WriteString(guidance)

    return buf.String(), nil
}
```

**6. gathering.go (and other phase files)** - Phase implementations:
```go
type GatheringPhase struct {
    state     *phasesSchema.Phase
    artifacts *project.ArtifactCollection
    tasks     *project.TaskCollection
    project   *ExplorationProject
    ctx       *sow.Context
}

// Implements Phase interface
func (p *GatheringPhase) Complete() (*domain.PhaseOperationResult, error) {
    // Validation
    if len(p.state.Tasks) < 3 {
        return nil, fmt.Errorf("minimum 3 topics required")
    }

    // Update state
    p.state.Status = "completed"
    if err := p.project.Save(); err != nil {
        return nil, err
    }

    // Return event declaratively
    return domain.WithEvent(EventCompleteGathering), nil
}

// ... other Phase interface methods
```

**Benefits of this structure**:
- **Clear separation**: Each file has single responsibility
- **Easy navigation**: Find states, events, guards quickly
- **Testable**: Each component can be tested independently
- **Consistent**: All project types follow same pattern
- **Self-documenting**: File names indicate contents

**Loader Integration** (`cli/internal/project/loader/loader.go`):
```go
func Load(ctx *sow.Context) (domain.Project, error) {
    state, _, err := statechart.LoadProjectState(ctx.FS())
    projectType := state.Project.Type

    switch projectType {
    case "standard":
        return standard.New((*projects.StandardProjectState)(state), ctx), nil
    case "exploration":
        return exploration.New((*projects.ExplorationProjectState)(state), ctx), nil
    case "design":
        return design.New((*projects.DesignProjectState)(state), ctx), nil
    case "breakdown":
        return breakdown.New((*projects.BreakdownProjectState)(state), ctx), nil
    default:
        return nil, fmt.Errorf("unknown project type: %s", projectType)
    }
}
```

### State Machine Specifications

Each project type implements a state machine guiding workflow progression. The following sections define complete state machines with transitions, guards, and orchestrator behavior per state.

#### Exploration Project State Machine

**States**: `Gathering → Researching → Summarizing → Finalizing → Completed`

**Phase mapping**:
- States `Gathering`, `Researching`, `Summarizing` → `exploration` phase
- State `Finalizing` → `finalization` phase
- State `Completed` → terminal state (project complete)

**State definitions**:

**1. Gathering**
- **Purpose**: Identify research topics and areas to investigate
- **Phase status**: `exploration.status = "gathering"`
- **Orchestrator behavior**:
  - Prompt user for research topics
  - Create task for each topic added
  - Track topics in `exploration.tasks[]`
- **Valid transitions**: → Researching
- **Transition guard**: `len(exploration.tasks) >= 3 && user_confirms`
- **Exit criteria**: Minimum 3 research topics identified and user approves transition

**2. Researching**
- **Purpose**: Investigate each research topic, document findings
- **Phase status**: `exploration.status = "researching"`
- **Orchestrator behavior**:
  - Work through tasks sequentially or in parallel
  - Create artifacts for findings (files, notes, code examples)
  - Update `exploration.artifacts[]` as findings documented
  - Allow dynamic task addition (new topics discovered during research)
- **Valid transitions**: → Summarizing
- **Transition guard**: `all_tasks_completed && user_confirms`
- **Exit criteria**: All research tasks marked completed

**3. Summarizing**
- **Purpose**: Create comprehensive summary of exploration findings
- **Phase status**: `exploration.status = "summarizing"`
- **Orchestrator behavior**:
  - Generate summary document consolidating all findings
  - Create summary artifact (`.sow/design/` or specified location)
  - Add summary to `exploration.artifacts[]`
  - Present summary to user for approval
- **Valid transitions**: → Finalizing
- **Transition guard**: `summary_artifact_exists && artifact_approved && user_confirms`
- **Exit criteria**: Summary artifact created, approved, and user confirms finalization

**4. Finalizing**
- **Purpose**: Move artifacts to target locations, create PR, cleanup
- **Phase status**: `finalization.status = "in_progress"`
- **Phase transition**: `exploration.status = "completed"`, activate `finalization` phase
- **Orchestrator behavior**:
  - Create finalization tasks:
    - Move summary to `.sow/knowledge/explorations/`
    - Move other artifacts to appropriate locations
    - Create PR with exploration findings
    - Delete `.sow/project/` directory
  - Execute tasks sequentially
  - Update `finalization.tasks[]` as completed
- **Valid transitions**: → Completed
- **Transition guard**: `all_finalization_tasks_completed`
- **Exit criteria**: All finalization tasks completed successfully

**5. Completed**
- **Purpose**: Terminal state, exploration finished
- **Phase status**: `finalization.status = "completed"`
- **Orchestrator behavior**: None (project complete, orchestrator exits)
- **Valid transitions**: None (terminal state)

**Guard implementation notes**:
- `len(exploration.tasks) >= 3`: Enforce minimum topic count for meaningful exploration
- `all_tasks_completed`: Check `task.status == "completed"` for all tasks in phase
- `summary_artifact_exists`: Check `exploration.artifacts[]` contains summary with appropriate path
- `artifact_approved`: Check `artifact.approved == true` (user explicitly approved summary)
- `user_confirms`: Orchestrator asks explicit confirmation before state transition

#### Design Project State Machine

**States**: `Drafting → Reviewing → Approved → Finalizing → Completed`

**Phase mapping**:
- States `Drafting`, `Reviewing`, `Approved` → `design` phase
- State `Finalizing` → `finalization` phase
- State `Completed` → terminal state

**State definitions**:

**1. Drafting**
- **Purpose**: Create design documents (ADRs, design docs, Arc42 updates, diagrams)
- **Phase status**: `design.status = "drafting"`
- **Orchestrator behavior**:
  - Review `design.inputs[]` to understand what's being designed
  - Determine document types needed (use decision tree from design mode)
  - Create task for each document to produce
  - For each task:
    - Load appropriate template via `sow prompt design/<type>`
    - Generate document in `.sow/design/` workspace
    - Add as artifact to `design.artifacts[]` with `approved = false`
    - Mark task completed
- **Valid transitions**: → Reviewing
- **Transition guard**: `all_tasks_completed && all_artifacts_created && user_confirms`
- **Exit criteria**: All planned documents created as artifacts

**2. Reviewing**
- **Purpose**: User reviews generated documents, provides feedback
- **Phase status**: `design.status = "reviewing"`
- **Orchestrator behavior**:
  - Present each artifact for user review
  - For each artifact:
    - Display document content
    - Ask: "Approve this document? (y)es/(n)o/(e)dit"
    - If edit requested: make changes, re-present
    - If approved: set `artifact.approved = true`
  - Track approval status across all artifacts
- **Valid transitions**: → Approved
- **Transition guard**: `all_artifacts_approved && user_confirms`
- **Exit criteria**: All artifacts have `approved = true`

**3. Approved**
- **Purpose**: All documents approved, ready for finalization
- **Phase status**: `design.status = "approved"`
- **Orchestrator behavior**:
  - Confirm with user: "All documents approved. Ready to finalize?"
  - Wait for user confirmation
- **Valid transitions**: → Finalizing
- **Transition guard**: `user_confirms`
- **Exit criteria**: User confirms transition to finalization

**4. Finalizing**
- **Purpose**: Move artifacts to target locations, create PR, cleanup
- **Phase status**: `finalization.status = "in_progress"`
- **Phase transition**: `design.status = "completed"`, activate `finalization` phase
- **Orchestrator behavior**:
  - Create finalization tasks:
    - For each artifact: move to `artifact.metadata.target` location
    - Create directories if needed (e.g., `.sow/knowledge/adrs/`)
    - Create PR with design documents
    - Delete `.sow/design/` workspace
    - Delete `.sow/project/` directory
  - Execute tasks sequentially
  - Update `finalization.tasks[]` as completed
- **Valid transitions**: → Completed
- **Transition guard**: `all_finalization_tasks_completed`
- **Exit criteria**: All artifacts moved, PR created, workspace cleaned

**5. Completed**
- **Purpose**: Terminal state, design finished
- **Phase status**: `finalization.status = "completed"`
- **Orchestrator behavior**: None (project complete, orchestrator exits)
- **Valid transitions**: None (terminal state)

**Guard implementation notes**:
- `all_artifacts_created`: Check each task has corresponding artifact in `design.artifacts[]`
- `all_artifacts_approved`: Check `artifact.approved == true` for all artifacts
- `artifact.metadata.target`: Each artifact must specify target location for finalization
- Design projects may have `design.inputs[]` referencing exploration findings or existing docs

#### Breakdown Project State Machine

**States**: `Decomposing → Documenting → Approving → Finalizing → Completed`

**Phase mapping**:
- States `Decomposing`, `Documenting`, `Approving` → `breakdown` phase
- State `Finalizing` → `finalization` phase
- State `Completed` → terminal state

**State definitions**:

**1. Decomposing**
- **Purpose**: Break down feature/project into implementable work units
- **Phase status**: `breakdown.status = "decomposing"`
- **Orchestrator behavior**:
  - Analyze scope (from design docs, user description, or codebase context)
  - Identify logical work units
  - Create task for each work unit with:
    - Clear, descriptive name
    - Initial dependencies (if applicable)
    - Estimated complexity/scope
  - Track units in `breakdown.tasks[]`
  - Allow iterative refinement (add/remove/reorder tasks)
- **Valid transitions**: → Documenting
- **Transition guard**: `len(breakdown.tasks) >= 1 && user_confirms`
- **Exit criteria**: All work units identified and user approves moving to documentation

**2. Documenting**
- **Purpose**: Create detailed specification for each work unit
- **Phase status**: `breakdown.status = "documenting"`
- **Orchestrator behavior**:
  - For each task:
    - Create detailed description in `task.metadata.description`
    - Define acceptance criteria
    - Identify technical considerations
    - Note dependencies on other tasks
    - Estimate effort/complexity
    - Mark task as documented
  - Create task descriptions in `.sow/project/phases/breakdown/tasks/{id}/description.md`
  - Update `breakdown.tasks[]` with metadata
- **Valid transitions**: → Approving
- **Transition guard**: `all_tasks_documented && user_confirms`
- **Exit criteria**: All tasks have detailed specifications

**3. Approving**
- **Purpose**: User reviews and approves work unit specifications
- **Phase status**: `breakdown.status = "approving"`
- **Orchestrator behavior**:
  - Present each task specification for review
  - For each task:
    - Display task name, description, dependencies, acceptance criteria
    - Ask: "Approve this work unit? (y)es/(n)o/(e)dit"
    - If edit: revise specification, re-present
    - If approved: mark task approved in `task.metadata.approved = true`
  - Track approval across all tasks
- **Valid transitions**: → Finalizing
- **Transition guard**: `all_tasks_approved && user_confirms`
- **Exit criteria**: All work units approved by user

**4. Finalizing**
- **Purpose**: Create GitHub issues from work units, create PR, cleanup
- **Phase status**: `finalization.status = "in_progress"`
- **Phase transition**: `breakdown.status = "completed"`, activate `finalization` phase
- **Orchestrator behavior**:
  - Create finalization tasks:
    - For each approved task: create GitHub issue via `gh issue create`
    - Store issue URL and number in `task.metadata.github_issue_url`, `task.metadata.github_issue_number`
    - Create tracking document linking all issues (optional)
    - Create PR with breakdown documentation (optional)
    - Delete `.sow/project/` directory
  - Execute tasks sequentially
  - Update `finalization.tasks[]` as completed
- **Valid transitions**: → Completed
- **Transition guard**: `all_github_issues_created`
- **Exit criteria**: All GitHub issues created, metadata updated, cleanup complete

**5. Completed**
- **Purpose**: Terminal state, breakdown finished
- **Phase status**: `finalization.status = "completed"`
- **Orchestrator behavior**: None (project complete, orchestrator exits)
- **Valid transitions**: None (terminal state)

**Guard implementation notes**:
- `len(breakdown.tasks) >= 1`: Minimum one work unit required
- `all_tasks_documented`: Check each task has `description.md` file created
- `all_tasks_approved`: Check `task.metadata.approved == true` for all tasks
- `all_github_issues_created`: Check each task has `task.metadata.github_issue_number` populated
- Task dependencies tracked in `task.dependencies[]` (array of task IDs)

### Task Usage Across Project Types

All project types use tasks for tracking work, with different semantics per type:

**Exploration tasks**:
- Represent research topics or investigation areas
- Added dynamically during exploration workflow
- Example: "Investigate authentication patterns", "Research OAuth vs JWT", "Compare auth libraries"
- Task completion marks investigation area as done
- Flexible: new tasks can be added mid-exploration as new areas discovered

**Design tasks**:
- Represent individual design documents to create
- One task per planned document (ADR, design doc, Arc42 update, etc.)
- Task lifecycle: create task → write document → add as artifact → request approval → mark task/artifact approved
- Task completion tied to artifact approval
- Example: "Create ADR-015 OAuth decision", "Write OAuth integration design", "Update Arc42 section 5"

**Breakdown tasks**:
- Represent work units that will become GitHub issues
- Tasks map directly to future implementation work
- task.metadata stores GitHub-specific data (issue URL, number, etc.)
- Task completion in breakdown phase means "specification ready"
- Finalization phase creates actual GitHub issues from completed tasks

**Finalization tasks** (all project types):
- Track finalization workflow steps
- Typically include: move artifacts, create PR, cleanup workspace
- Sequential execution (not parallel)
- Auto-generated by orchestrator when entering finalization phase

This unified task model provides consistent tracking across all workflows while allowing type-specific semantics.

### Data Flow

**Project initialization flow**:
1. User runs `sow project` on branch `explore/topic-research`
2. CLI detects no existing project, detects branch prefix → `type = "exploration"`
3. CLI launches wizard: "Detected exploration project. Create new exploration project? (Y/n)"
4. User confirms
5. CLI loads `#ExplorationProjectState` schema
6. CLI creates `.sow/project/state.yaml` with discriminated type:
   ```yaml
   project:
     type: exploration
     name: topic-research
     branch: explore/topic-research
   statechart:
     current_state: Gathering
   phases:
     exploration:
       status: gathering
       enabled: true
       tasks: []
     finalization:
       status: pending
       enabled: true
       tasks: []
   ```
5. Orchestrator loads prompt: `sow prompt project/exploration/gathering`
6. Orchestrator guides user through gathering phase, adding research tasks dynamically

**State transition flow** (exploration example):
1. User completes summarizing phase (summary artifact created)
2. Orchestrator checks guard conditions: `summary_artifact_created = true`
3. Orchestrator proposes transition: "Summary complete. Ready to finalize?"
4. User approves
5. Orchestrator transitions exploration phase: `exploration.status = "completed"`
6. Orchestrator activates finalization phase: `finalization.status = "in_progress"`
7. CLI updates `state.yaml`: `current_state: Finalizing`
8. Orchestrator loads new prompt: `sow prompt project/exploration/finalizing`
9. Orchestrator creates finalization tasks:
   - Move summary artifact to target location
   - Create PR with exploration findings
   - Clean up `.sow/project/` directory
10. Orchestrator executes finalization tasks sequentially
11. On completion: `finalization.status = "completed"`, `current_state: Completed`

## Implementation Notes

### Orchestrator Initialization and Handoff

After the wizard creates a project, it launches Claude Code with a dynamically-generated prompt via the project's PromptGenerator. This follows the State Machine SDK pattern where prompts are generated at runtime with access to external systems.

**Wizard completion flow**:
1. Wizard creates `.sow/project/state.yaml` with initial state for project type
2. Wizard creates `.sow/project/log.md` with initialization entry
3. Wizard loads the newly created project via `loader.Load()`
4. Project constructs state machine using builder pattern (includes PromptGenerator)
5. State machine generates initial prompt via `PromptGenerator.GeneratePrompt()`
6. Wizard launches Claude Code with generated prompt: `exec.Command("claude", prompt).Run()`

**Implementation pattern**:

```go
// In cli/cmd/project.go (wizard completion)
func (w *wizard) launchOrchestrator() error {
    // Load the newly created project
    // This triggers: 1) schema loading, 2) type detection, 3) buildStateMachine()
    project, err := loader.Load(w.ctx)
    if err != nil {
        return fmt.Errorf("failed to load project: %w", err)
    }

    // Get state machine (already built with PromptGenerator)
    machine := project.Machine()

    // Generate state-specific prompt dynamically
    // This calls project type's PromptGenerator.GeneratePrompt()
    prompt, err := machine.GeneratePrompt()
    if err != nil {
        return fmt.Errorf("failed to generate prompt: %w", err)
    }

    // Launch Claude Code with generated prompt
    claudeCmd := exec.Command("claude", prompt)
    claudeCmd.Stdin = os.Stdin
    claudeCmd.Stdout = os.Stdout
    claudeCmd.Stderr = os.Stderr
    claudeCmd.Dir = w.ctx.RepoRoot()

    return claudeCmd.Run()
}
```

**PromptGenerator Implementation** (exploration example):

Each project type implements `PromptGenerator` interface to generate state-specific prompts dynamically:

```go
// In cli/internal/project/exploration/prompts.go
type ExplorationPromptGenerator struct {
    components *statechart.PromptComponents
    ctx        *sow.Context
}

func NewExplorationPromptGenerator(ctx *sow.Context) *ExplorationPromptGenerator {
    return &ExplorationPromptGenerator{
        components: statechart.NewPromptComponents(ctx),
        ctx:        ctx,
    }
}

func (g *ExplorationPromptGenerator) GeneratePrompt(
    state statechart.State,
    projectState *schemas.ProjectState,
) (string, error) {
    // Route to state-specific generator
    switch state {
    case ExplorationGathering:
        return g.generateGatheringPrompt(projectState)
    case ExplorationResearching:
        return g.generateResearchingPrompt(projectState)
    case ExplorationSummarizing:
        return g.generateSummarizingPrompt(projectState)
    case ExplorationFinalizing:
        return g.generateFinalizingPrompt(projectState)
    default:
        return "", fmt.Errorf("unknown exploration state: %s", state)
    }
}

// State-specific prompt generation with external data access
func (g *ExplorationPromptGenerator) generateGatheringPrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    // Reusable SDK components
    buf.WriteString(g.components.ProjectHeader(projectState))

    // External system access: git status
    gitStatus, err := g.components.GitStatus()
    if err == nil {
        buf.WriteString(gitStatus)
    }

    // External system access: recent commits
    commits, err := g.components.RecentCommits(5)
    if err == nil {
        buf.WriteString(commits)
    }

    // Exploration-specific context
    buf.WriteString("\n## Research Topics\n\n")
    if len(projectState.Phases.Exploration.Tasks) > 0 {
        for _, task := range projectState.Phases.Exploration.Tasks {
            buf.WriteString(fmt.Sprintf("- %s (%s)\n", task.Name, task.Status))
        }
    } else {
        buf.WriteString("No topics identified yet.\n")
    }

    // Static guidance via optional template
    guidance, err := g.components.RenderTemplate(
        prompts.PromptExplorationGathering,
        &prompts.StatechartContext{
            State:        string(ExplorationGathering),
            ProjectState: projectState,
        },
    )
    if err == nil {
        buf.WriteString("\n" + guidance)
    }

    return buf.String(), nil
}
```

**Prompt Infrastructure Requirements**:

1. **Template files** (optional, for static guidance only):
```
cli/internal/prompts/templates/statechart/
  exploration/
    gathering.tmpl     - Static guidance for topic discovery
    researching.tmpl   - Static guidance for investigation
  design/
    drafting.tmpl      - Static guidance for document creation
  breakdown/
    decomposing.tmpl   - Static guidance for work unit identification
```

2. **Template IDs** (optional, in `cli/internal/prompts/prompts.go`):
```go
const (
    PromptExplorationGathering    PromptID = "statechart.exploration.gathering"
    PromptExplorationResearching  PromptID = "statechart.exploration.researching"
    // ... etc for other states
)
```

**Key differences from template-only approach**:
- **No state→prompt mapping**: PromptGenerator.GeneratePrompt() handles routing internally
- **Dynamic generation**: Prompts composed at runtime, not template-only
- **External system access**: Can include git status, GitHub issues, filesystem data
- **Templates optional**: Templates only for static guidance, bulk of prompt is dynamic code
- **Type-specific**: Each project type implements PromptGenerator independently

**Phased rollout**:
1. **Phase 1: Schema extensions** - Implement six schema changes, update CUE validation
2. **Phase 2: Project type schemas** - Create exploration/design/breakdown schemas
3. **Phase 3: CLI detection** - Implement branch prefix detection and type initialization
4. **Phase 4: State machine + prompts** - Add state transition validation, create prompt templates, implement prompt loading
5. **Phase 5: Code removal** - Delete mode-specific code after verification
6. **Phase 6: Documentation** - Update `.claude/CLAUDE.md` and user docs

**Validation approach**:
- Create test branches (`explore/test`, `design/test`, `breakdown/test`)
- Initialize projects and verify correct schema loaded
- Test state transitions and guard validation
- Verify prompt loading for each state
- Confirm zero-context resumability

**Rollback strategy**: If critical issues discovered, revert commits in reverse phase order. Schema changes are backward compatible (all new fields optional except status unconstrain, which only affects new projects).

## Testing Approach

**Unit tests**:
- Schema validation for all project types
- Branch prefix detection logic
- State transition guard checking
- Discriminated union type resolution

**Integration tests**:
- Initialize exploration project, verify schema structure
- Initialize design project, verify input/output tracking
- Initialize breakdown project, verify task metadata
- State transition workflow for each project type
- Prompt loading for all type/state combinations

**Manual verification**:
- Create test branch for each mode (explore/test, design/test, breakdown/test)
- Walk through full workflow (gathering → researching → summarizing for exploration)
- Verify zero-context resumability (stop/restart at each state)
- Verify branch prefix detection correctly determines project type

## Alternatives Considered

### Option 1: Keep Modes as Separate Systems with Shared Library

**Description**: Extract common functionality (state tracking, logging, artifact management) into shared library. Modes remain separate but reduce duplication.

**Pros**:
- Non-breaking change (gradual migration possible)
- Maintains mode independence
- Reduces code duplication

**Cons**:
- Doesn't solve mental model fragmentation
- Still three separate directory structures
- No unified state machine workflow
- Shared library adds complexity (versioning, API stability)

**Why not chosen**: Doesn't address core problem of conceptual fragmentation. Users still learn three systems. State machine workflows remain impossible. Partial solution that adds architectural complexity.

### Option 2: Configuration-Driven Modes (Single Project Type)

**Description**: Single project type with mode behavior controlled by configuration flags. Add `mode: "exploration" | "design" | "breakdown"` field instead of discriminated types.

**Pros**:
- Simpler than discriminated union
- Single project schema to maintain
- Easier migration path

**Cons**:
- Mode-specific behavior via conditionals throughout codebase
- Less type safety (can't constrain fields per mode)
- State machine definitions less clear
- Adding new modes requires modifying core schema
- CUE schema validation less effective

**Why not chosen**: Discriminated union provides superior type safety and separation of concerns. Mode-specific constraints expressed naturally in schemas. Adding new project types doesn't affect existing schemas. Code organization clearer (project type handlers separate).

## References

- **ADR-003**: [Consolidate Operating Modes into Project Types](../knowledge/adrs/003-consolidate-modes-to-projects.md) - Architectural decision documenting why this change is being made
- **Exploration findings**: [Modes-to-Projects Consolidation](../knowledge/explorations/modes-to-projects-2025-01.md) - Comprehensive research validating approach, documenting schema mappings and feature preservation
- **CUE Language**: [CUE Discriminated Unions](https://cuelang.org/docs/tutorials/tour/types/disjunctions/) - Pattern used for project type discrimination

## Future Considerations

**Additional project types**: Framework supports adding new types without modifying existing schemas:
- Refactoring projects (`refactor/` prefix)
- Debugging projects (`debug/` prefix)
- Performance optimization projects (`perf/` prefix)

**Custom state machines**: Allow users to define custom state machines per project type in configuration. Current design hardcodes states in schemas—could be externalized.

**Cross-project type workflows**: Enable automated transitions between types. Example: Exploration completion triggers design project initialization with exploration outputs as inputs.

**Type-specific agents**: Each project type could have specialized agent definitions optimized for that workflow. Exploration agent focused on research, design agent on documentation quality, etc.
