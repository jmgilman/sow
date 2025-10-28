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
- Implement six schema improvements enabling clean mode-to-project translation
- Remove all mode-specific code, directories, and CLI commands
- Support automatic project type detection via branch prefix
- Enable state machine workflows through phase-specific states

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

Extend base schemas (`schemas/phases/common.cue`) with six improvements:

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

Project type schemas define valid state transitions declaratively. CLI validates transitions when updating state.

**State transition guards** (example for exploration):
```yaml
# .sow/project/state.yaml
statechart:
  current_state: Gathering

# Valid transitions with guards:
# Gathering → Researching: min_topics_added && user_approved
# Researching → Summarizing: all_topics_completed
# Summarizing → Finalizing: summary_artifact_created && user_approved
# Finalizing → Completed: all_finalization_tasks_completed

# CLI checks guard before allowing transition:
sow project transition researching
# Error: Cannot transition to 'researching': minimum 3 topics required
```

**Orchestrator prompt loading**:
```bash
sow prompt project/exploration/researching  # Load state-specific prompt
```

Prompts stored in `.claude/prompts/project/<type>/<state>.md`. Orchestrator loads appropriate prompt when entering state.

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

### Guard Validation Architecture

**Responsibility**: Guards are validated by the state machine when phase completion events are fired.

**Flow**:
1. Orchestrator calls `sow agent complete` (existing command)
2. CLI loads Phase interface implementation (e.g., `ExplorationPhase`, `DesignPhase`)
3. Phase's `Complete()` method:
   - Performs phase-specific validation (optional pre-checks)
   - Fires event on state machine: `phase.project.Machine().Fire(EventCompleteGathering)`
4. State machine checks guard function before transition
5. If guard passes: transition succeeds, state updated, phase marked completed
6. If guard fails: event rejected, state unchanged, error returned to orchestrator

**Implementation pattern** (follows existing standard project pattern):

```go
// In cli/internal/project/exploration/gathering.go
func (p *GatheringPhase) Complete() error {
    // Optional: phase-specific pre-validation
    if len(p.state.Tasks) < 3 {
        return fmt.Errorf("minimum 3 topics required, found %d", len(p.state.Tasks))
    }

    // Fire event - state machine validates guard
    if err := p.project.Machine().Fire(statechart.EventCompleteGathering); err != nil {
        return fmt.Errorf("failed to complete gathering: %w", err)
    }

    // Update phase status on success
    p.state.Status = "completed"
    now := time.Now()
    p.state.Completed_at = &now

    return p.project.Save()
}

// In cli/internal/project/statechart/guards.go (new guards for exploration)
func (m *Machine) gatheringComplete() bool {
    // Guard inspects project state
    tasks := m.projectState.Phases.Exploration.Tasks
    return len(tasks) >= 3 // Minimum topic count requirement
}

// In cli/internal/project/statechart/machine.go (state machine configuration)
m.sm.Configure(ExplorationGathering).
    Permit(EventCompleteGathering, ExplorationResearching, m.gatheringComplete).
    OnEntry(m.onEntry(ExplorationGathering))
```

**Key architectural points**:
- Each project type defines custom events (e.g., `EventCompleteGathering`, `EventCompleteDrafting`)
- Guards are methods on `Machine` that inspect `m.projectState`
- State machine configuration in `configure()` links: events → guards → transitions
- Existing `sow agent complete` command works unchanged - it calls `Phase.Complete()`
- Guards enforce workflow invariants (minimum tasks, artifacts approved, etc.)

**Consistency with existing architecture**: This approach matches the current standard project implementation. See `cli/internal/project/standard/planning.go:109-136` and `cli/internal/project/statechart/guards.go` for reference patterns.

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

After the wizard creates a project, it launches Claude Code with a state-specific prompt following the same pattern as the existing `sow start` command.

**Wizard completion flow**:
1. Wizard creates `.sow/project/state.yaml` with initial state for project type
2. Wizard creates `.sow/project/log.md` with initialization entry
3. Wizard determines initial prompt ID based on project type:
   - `exploration` → `prompts.PromptExplorationGathering`
   - `design` → `prompts.PromptDesignDrafting`
   - `breakdown` → `prompts.PromptBreakdownDecomposing`
   - `standard` → `prompts.PromptPlanningActive`
4. Wizard loads project to get context for prompt rendering
5. Wizard renders state-specific prompt with project context
6. Wizard launches Claude Code: `exec.Command("claude", prompt).Run()`

**Implementation pattern**:

```go
// In cli/cmd/project.go (wizard completion)
func (w *wizard) launchOrchestrator() error {
    // Load the newly created project
    project, err := loader.Load(w.ctx)
    if err != nil {
        return fmt.Errorf("failed to load project: %w", err)
    }

    // Get initial state from project
    state := project.Machine().State()

    // Map state to prompt ID
    promptID := statechart.StateToPromptID(state)

    // Generate state-specific prompt with context
    promptCtx := &prompts.StatechartContext{
        State:        string(state),
        ProjectState: project.Machine().ProjectState(),
    }

    prompt, err := prompts.Render(promptID, promptCtx)
    if err != nil {
        return fmt.Errorf("failed to render prompt: %w", err)
    }

    // Launch Claude Code with prompt
    claudeCmd := exec.Command("claude", prompt)
    claudeCmd.Stdin = os.Stdin
    claudeCmd.Stdout = os.Stdout
    claudeCmd.Stderr = os.Stderr
    claudeCmd.Dir = w.ctx.RepoRoot()

    return claudeCmd.Run()
}
```

**Prompt infrastructure required**:

1. **New prompt constants** (`cli/internal/prompts/prompts.go`):
```go
const (
    // Exploration project prompts
    PromptExplorationGathering    PromptID = "statechart.exploration.gathering"
    PromptExplorationResearching  PromptID = "statechart.exploration.researching"
    PromptExplorationSummarizing  PromptID = "statechart.exploration.summarizing"
    PromptExplorationFinalizing   PromptID = "statechart.exploration.finalizing"

    // Design project prompts
    PromptDesignDrafting         PromptID = "statechart.design.drafting"
    PromptDesignReviewing        PromptID = "statechart.design.reviewing"
    PromptDesignApproved         PromptID = "statechart.design.approved"
    PromptDesignFinalizing       PromptID = "statechart.design.finalizing"

    // Breakdown project prompts
    PromptBreakdownDecomposing   PromptID = "statechart.breakdown.decomposing"
    PromptBreakdownDocumenting   PromptID = "statechart.breakdown.documenting"
    PromptBreakdownApproving     PromptID = "statechart.breakdown.approving"
    PromptBreakdownFinalizing    PromptID = "statechart.breakdown.finalizing"
)
```

2. **State → Prompt mapping** (update `statechart/prompts.go`):
```go
var statePrompts = map[State]prompts.PromptID{
    // Existing standard project states
    NoProject:               prompts.PromptNoProject,
    PlanningActive:          prompts.PromptPlanningActive,

    // Exploration states
    ExplorationGathering:    prompts.PromptExplorationGathering,
    ExplorationResearching:  prompts.PromptExplorationResearching,
    ExplorationSummarizing:  prompts.PromptExplorationSummarizing,
    ExplorationFinalizing:   prompts.PromptExplorationFinalizing,

    // Design states
    DesignDrafting:          prompts.PromptDesignDrafting,
    DesignReviewing:         prompts.PromptDesignReviewing,
    DesignApproved:          prompts.PromptDesignApproved,
    DesignFinalizing:        prompts.PromptDesignFinalizing,

    // Breakdown states
    BreakdownDecomposing:    prompts.PromptBreakdownDecomposing,
    BreakdownDocumenting:    prompts.PromptBreakdownDocumenting,
    BreakdownApproving:      prompts.PromptBreakdownApproving,
    BreakdownFinalizing:     prompts.PromptBreakdownFinalizing,
}

// StateToPromptID helper (new)
func StateToPromptID(state State) prompts.PromptID {
    if id, ok := statePrompts[state]; ok {
        return id
    }
    return prompts.PromptNoProject // Fallback
}
```

3. **Prompt template files** (must be created as part of implementation):
```
cli/internal/prompts/templates/statechart/
  exploration/
    gathering.tmpl     - Guide orchestrator in topic discovery
    researching.tmpl   - Guide orchestrator in investigation
    summarizing.tmpl   - Guide orchestrator in synthesis
    finalizing.tmpl    - Guide orchestrator in artifact finalization
  design/
    drafting.tmpl      - Guide orchestrator in document creation
    reviewing.tmpl     - Guide orchestrator in approval workflow
    approved.tmpl      - Guide orchestrator in preparation for finalization
    finalizing.tmpl    - Guide orchestrator in artifact movement and PR
  breakdown/
    decomposing.tmpl   - Guide orchestrator in work unit identification
    documenting.tmpl   - Guide orchestrator in specification writing
    approving.tmpl     - Guide orchestrator in approval workflow
    finalizing.tmpl    - Guide orchestrator in GitHub issue creation
```

**Key points**:
- Wizard doesn't need special handoff logic - reuses existing prompt generation pattern
- All state-specific orchestrator behavior is encoded in prompt templates
- Prompts are **part of this implementation** (not pre-existing) - Phase 4 work
- Each state gets focused prompt with specific guidance for that workflow stage
- Pattern matches existing `sow start` command architecture

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
