# Core Design: Modes to Project Types

**Author**: Architecture Team
**Date**: 2025-10-31
**Status**: Proposed
**Related ADRs**: [ADR-001](../../knowledge/adrs/001-consolidate-modes-to-projects.md), [ADR-002](../../knowledge/adrs/002-wizard-based-project-initialization.md)

## Overview

This document specifies the foundational changes required to support transitioning sow's three operating modes (exploration, design, breakdown) into specialized project types within the unified project system.

**Key Insight**: Much of the required infrastructure already exists (~70% complete). This design focuses on the remaining foundational pieces that individual project type designs will depend on.

## What's Already Implemented

The following infrastructure is already in place:

- ✅ **PhaseOperationResult pattern**: Type, `NoEvent()`, `WithEvent()` helpers
- ✅ **Phase interface**: Methods return `(*PhaseOperationResult, error)` for event triggering
- ✅ **State machine SDK**: `MachineBuilder`, guards, transitions, prompt generation
- ✅ **Prompt infrastructure**: `PromptGenerator` interface, `PromptComponents` helpers
- ✅ **Discriminated union foundation**: `#ProjectState` discriminator exists (currently only StandardProjectState)
- ✅ **Phase.status unconstrained**: Already `string` type, allows project-specific values

See implementation in:
- `cli/internal/project/domain/` (interfaces, PhaseOperationResult)
- `cli/internal/project/statechart/` (state machine SDK)
- `cli/schemas/phases/common.cue` (base schemas)
- `cli/schemas/projects/standard.cue` (discriminated union pattern)

## Foundational Changes Required

### 1. Schema Extensions

Four schema changes enable all three project types to use the common Phase/Task/Artifact model:

#### Change 1: Make `Artifact.approved` Optional

**Current**:
```cue
#Artifact: {
    approved: bool  // REQUIRED
}
```

**Required**:
```cue
#Artifact: {
    approved?: bool @go(,optional=nillable)  // OPTIONAL
}
```

**Rationale**:
- Exploration artifacts don't require approval
- Design artifacts require approval
- Making it optional supports both use cases

#### Change 2: Add `Phase.inputs` Field

**Add to `#Phase`**:
```cue
#Phase: {
    // ... existing fields

    inputs?: [...#Artifact] @go(,optional=nillable)  // NEW
}
```

**Rationale**:
- Design and breakdown phases track input sources (exploration findings, docs, references)
- Reusing `#Artifact` type provides structured tracking with paths, descriptions, metadata
- Clear separation: `inputs` inform work, `artifacts` are outputs

**Usage**:
- Design: Track which explorations/docs inform design decisions
- Breakdown: Track which design docs are being broken down
- Exploration: Not used (no inputs field needed)

#### Change 3: Add `Task.refs` Field

**Add to `#Task`**:
```cue
#Task: {
    // ... existing fields

    refs?: [...#Artifact] @go(,optional=nillable)  // NEW
}
```

**Rationale**:
- Exploration tasks (topics) need to reference files created during research
- Provides richer metadata than simple string paths
- Enables tracking which artifacts relate to which topics

**Usage**:
- Exploration: Link research findings to topics
- Design/Breakdown: Not typically used

#### Change 4: Add `Task.metadata` Field

**Add to `#Task`**:
```cue
#Task: {
    // ... existing fields

    metadata?: {[string]: _} @go(,optional=nillable)  // NEW
}
```

**Rationale**:
- Breakdown tasks store GitHub issue metadata (URLs, numbers, labels)
- Matches existing `Phase.metadata` and `Artifact.metadata` patterns
- Provides extensibility for future task-specific data

**Usage**:
- Breakdown: Store `github_issue_url`, `github_issue_number`, etc.
- Other types: Can use for project-specific task data

**Note**: Task status remains constrained to `"pending" | "in_progress" | "needs_review" | "completed" | "abandoned"`. No project-type-specific task statuses. Use `metadata` for finer-grained tracking if needed.

### 2. Intra-Phase State Progression

#### Problem

Project types like exploration have multiple states within a single phase:
- `exploration` phase: Active → Summarizing states
- Both states belong to same phase, different workflow stages

Current architecture only supports:
- `sow agent complete` - transitions between **phases**
- No mechanism for transitions between **states within a phase**

#### Solution: `sow agent advance` Command

**New CLI command**:
```bash
sow agent advance
```

**Purpose**: Advance to next state within the current phase (intra-phase state transition)

**Comparison**:

| Command | Scope | Example |
|---------|-------|---------|
| `sow agent advance` | State within phase | exploration.Active → exploration.Summarizing |
| `sow agent complete` | Phase to phase | exploration phase → finalization phase |

**Implementation**:

1. **New Phase interface method**:
```go
type Phase interface {
    // ... existing methods

    // Advance to next state within this phase
    Advance() (*PhaseOperationResult, error)
}
```

2. **CLI command** (`cli/cmd/agent/advance.go`):
```go
func NewAdvanceCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "advance",
        Short: "Advance to next state within current phase",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmdutil.GetContext(cmd.Context())
            proj, err := loader.Load(ctx)
            if err != nil {
                return err
            }

            phase := proj.CurrentPhase()
            result, err := phase.Advance()
            if errors.Is(err, project.ErrNotSupported) {
                return fmt.Errorf("phase %s does not support state advancement", phase.Name())
            }
            if err != nil {
                return fmt.Errorf("failed to advance: %w", err)
            }

            // Fire event if returned
            if result.Event != "" {
                machine := proj.Machine()
                if err := machine.Fire(result.Event); err != nil {
                    return fmt.Errorf("failed to fire event: %w", err)
                }
                // Save after transition
                if err := proj.Save(); err != nil {
                    return fmt.Errorf("failed to save: %w", err)
                }
            }

            cmd.Println("✓ Advanced to next state")
            return nil
        },
    }
    return cmd
}
```

**Phase implementation pattern**:

Phases without internal states return `ErrNotSupported`:
```go
// Standard project phases
func (p *PlanningPhase) Advance() (*domain.PhaseOperationResult, error) {
    return nil, project.ErrNotSupported
}
```

Phases with internal states implement state-specific logic:
```go
// Exploration phase
func (p *ExplorationPhase) Advance() (*domain.PhaseOperationResult, error) {
    switch p.state.Status {
    case "active":
        // Validate guard conditions
        if !allTasksResolved(p.state.Tasks) {
            return nil, fmt.Errorf("cannot advance: unresolved tasks remain")
        }
        // Return event to trigger Active → Summarizing transition
        return domain.WithEvent(EventBeginSummarizing), nil

    case "summarizing":
        // Summarizing is final state in exploration phase
        // Use sow agent complete to move to finalization phase
        return nil, fmt.Errorf("already in final state, use 'sow agent complete' to finish phase")

    default:
        return nil, fmt.Errorf("unknown state: %s", p.state.Status)
    }
}
```

### 3. Project Type System

#### Discriminated Union Schema

**Update `#ProjectState` discriminator** (`cli/schemas/projects/projects.cue`):

```cue
package projects

// ProjectState is the root discriminated union for all project types.
// The discriminator field is project.type.
#ProjectState:
    | #StandardProjectState
    | #ExplorationProjectState
    | #DesignProjectState
    | #BreakdownProjectState
```

**Create type-specific schemas**:

Each project type defines its own schema file:
- `cli/schemas/projects/standard.cue` (already exists)
- `cli/schemas/projects/exploration.cue` (new)
- `cli/schemas/projects/design.cue` (new)
- `cli/schemas/projects/breakdown.cue` (new)

**Schema structure** (pattern for new types):

```cue
#ExplorationProjectState: {
    // Discriminator
    project: {
        type: "exploration"  // Fixed value
        name: string
        branch: string
        description: string
        created_at: time.Time
        updated_at: time.Time
    }

    // State machine position
    statechart: {
        current_state: "Active" | "Summarizing" | "Finalizing" | "Completed"
    }

    // Phase definitions with type-specific constraints
    phases: {
        exploration: p.#Phase & {
            status: "active" | "summarizing" | "completed"
            enabled: true
        }

        finalization: p.#Phase & {
            status: p.#GenericStatus
            enabled: true
        }
    }
}
```

**Key points**:
- `project.type` is the discriminator field
- `statechart.current_state` constrains valid states for this type
- Phase status values constrained per project type
- Phases compose base `#Phase` type with additional constraints

#### Branch Prefix Detection

**New function** (`cli/internal/project/types.go`):

```go
package project

import "strings"

// DetectProjectType determines project type from branch name convention.
// Returns "standard" for unknown prefixes (default).
func DetectProjectType(branchName string) string {
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

**Usage in loader** (`cli/internal/project/loader/loader.go`):

```go
func Load(ctx *sow.Context) (domain.Project, error) {
    // ... existing existence check

    // Load state from disk
    state, _, err := statechart.LoadProjectState(ctx.FS())
    if err != nil {
        return nil, fmt.Errorf("failed to load project state: %w", err)
    }

    // Route based on project type
    switch state.Project.Type {
    case "standard":
        return standard.New((*projects.StandardProjectState)(state), ctx), nil
    case "exploration":
        return exploration.New((*projects.ExplorationProjectState)(state), ctx), nil
    case "design":
        return design.New((*projects.DesignProjectState)(state), ctx), nil
    case "breakdown":
        return breakdown.New((*projects.BreakdownProjectState)(state), ctx), nil
    default:
        return nil, fmt.Errorf("unknown project type: %s", state.Project.Type)
    }
}

func Create(ctx *sow.Context, name, description string) (domain.Project, error) {
    // ... existing validation

    // Detect type from branch name
    branch, err := ctx.Git().CurrentBranch()
    if err != nil {
        return nil, fmt.Errorf("failed to get branch: %w", err)
    }

    projectType := project.DetectProjectType(branch)

    // Create type-specific initial state
    switch projectType {
    case "exploration":
        state := statechart.NewExplorationProjectState(name, description, branch)
        return exploration.New(state, ctx), nil
    case "design":
        state := statechart.NewDesignProjectState(name, description, branch)
        return design.New(state, ctx), nil
    case "breakdown":
        state := statechart.NewBreakdownProjectState(name, description, branch)
        return breakdown.New(state, ctx), nil
    default:
        state := statechart.NewProjectState(name, description, branch)
        return standard.New(state, ctx), nil
    }
}
```

### 4. Project Type Implementation Structure

Each project type follows consistent file organization in dedicated package:

```
cli/internal/project/{type}/
├── project.go          # Main struct, implements Project interface
│                       # Contains buildStateMachine() using builder pattern
│
├── states.go           # State constants (e.g., Active, Summarizing)
├── events.go           # Event constants (e.g., EventBeginSummarizing)
├── guards.go           # Pure guard functions checking transition conditions
├── prompts.go          # PromptGenerator implementation
│
└── {phase}.go          # Phase implementation files
    └── exploration.go  # Implements Phase interface for exploration phase
```

**Minimal project.go structure**:

```go
package exploration

import (
    "github.com/jmgilman/sow/cli/internal/project/domain"
    "github.com/jmgilman/sow/cli/internal/project/statechart"
    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/jmgilman/sow/cli/schemas/projects"
)

type ExplorationProject struct {
    state   *projects.ExplorationProjectState
    ctx     *sow.Context
    machine *statechart.Machine
    phases  map[string]domain.Phase
}

func New(state *projects.ExplorationProjectState, ctx *sow.Context) *ExplorationProject {
    p := &ExplorationProject{
        state: state,
        ctx:   ctx,
    }

    // Initialize phases
    p.phases = map[string]domain.Phase{
        "exploration":  newExplorationPhase(p),
        "finalization": newFinalizationPhase(p),
    }

    // Build state machine
    p.machine = p.buildStateMachine()

    return p
}

func (p *ExplorationProject) buildStateMachine() *statechart.Machine {
    currentState := statechart.State(p.state.Statechart.Current_state)
    projectState := (*schemas.ProjectState)(p.state)

    promptGen := NewExplorationPromptGenerator(p.ctx)
    builder := statechart.NewBuilder(currentState, projectState, promptGen)

    // Configure all state transitions with guards
    builder.
        AddTransition(
            ExplorationActive,
            ExplorationSummarizing,
            EventBeginSummarizing,
            statechart.WithGuard(func() bool {
                return AllTasksResolved(projectState.Phases.Exploration.Tasks)
            }),
        ).
        AddTransition(
            ExplorationSummarizing,
            ExplorationFinalizing,
            EventCompleteSummarizing,
            statechart.WithGuard(func() bool {
                return SummaryArtifactApproved(projectState.Phases.Exploration.Artifacts)
            }),
        )
        // ... more transitions

    machine := builder.Build()
    machine.SetFilesystem(p.ctx.FS())
    return machine
}

// Implement Project interface methods...
func (p *ExplorationProject) Name() string { return p.state.Project.Name }
func (p *ExplorationProject) Type() string { return "exploration" }
func (p *ExplorationProject) Machine() *statechart.Machine { return p.machine }
// ... etc
```

**Individual project type designs** will specify:
- Specific states and events
- Guard functions and conditions
- Phase implementations
- Prompt generation logic

### 5. Code Removal

After project types implemented, remove mode-specific code:

**Directories to delete**:
- `cli/internal/exploration/`
- `cli/internal/design/`
- `cli/internal/breakdown/`
- `.sow/exploration/` support
- `.sow/design/` support
- `.sow/breakdown/` support

**CLI commands to delete**:
- `cli/cmd/exploration/`
- `cli/cmd/design/`
- `cli/cmd/breakdown/`

**Schemas to delete**:
- `cli/schemas/exploration_index.cue`
- `cli/schemas/design_index.cue`
- `cli/schemas/breakdown_index.cue`

## Implementation Order

1. **Schema extensions** (minimal, foundational)
   - Make `Artifact.approved` optional
   - Add `Phase.inputs`, `Task.refs`, `Task.metadata`
   - Regenerate Go types

2. **Intra-phase progression** (new primitive)
   - Add `Phase.Advance()` to interface
   - Implement `sow agent advance` command
   - Update existing phases to return `ErrNotSupported`

3. **Project type system** (infrastructure)
   - Create discriminated union schema
   - Implement branch detection
   - Update loader with type routing

4. **Individual project types** (separate designs)
   - Exploration (see exploration-design.md)
   - Design (see design-design.md)
   - Breakdown (see breakdown-design.md)

5. **Code removal** (cleanup)
   - Delete mode-specific implementations
   - Update documentation

## Testing Strategy

**Schema validation**:
- CUE validation passes for all four project types
- Go type generation works correctly
- Optional fields handle nil values properly

**Advance command**:
- Phases without states return `ErrNotSupported`
- Guards prevent invalid transitions
- Events fire correctly on valid transitions

**Type detection**:
- Branch prefixes correctly map to types
- Unknown prefixes default to standard
- Loader instantiates correct project type

**Integration**:
- Create project on each branch prefix type
- Verify correct schema loaded
- Verify state machine constructed
- Verify prompts generated

## Dependencies

**External packages** (already in use):
- `github.com/qmuntal/stateless` - State machine implementation
- CUE - Schema validation and code generation

**Internal packages**:
- All infrastructure packages already exist
- No new dependencies required

## Migration

**Breaking change**: No backward compatibility with existing mode sessions.

**Migration path**:
1. Users complete or abandon active mode sessions
2. Deploy new version with project types
3. Start new work using unified `sow project` command

**Justification**: Framework in pre-public phase, breaking change acceptable to avoid complexity of dual support.

## Success Criteria

1. ✅ All four schema changes implemented and validated
2. ✅ `sow agent advance` command works for phases with internal states
3. ✅ Branch detection correctly routes to project types
4. ✅ Loader instantiates correct project type based on discriminator
5. ✅ Project type file structure established
6. ✅ All tests pass

## References

- **ADR-001**: [Consolidate Operating Modes into Project Types](../../knowledge/adrs/001-consolidate-modes-to-projects.md)
- **ADR-002**: [Interactive Wizard for Project Initialization](../../knowledge/adrs/002-wizard-based-project-initialization.md)
- **Exploration findings**: [Modes-to-Projects Consolidation](../../knowledge/explorations/modes-to-projects-2025-01.md)
- **State Machine SDK**: `cli/internal/project/statechart/` (existing implementation)
- **Standard Project**: `cli/internal/project/standard/` (reference implementation)
