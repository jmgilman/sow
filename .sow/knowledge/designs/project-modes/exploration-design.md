# Exploration Project Type Design

**Author**: Architecture Team
**Date**: 2025-10-31
**Status**: Proposed
**Depends On**: [Core Design](core-design.md)

## Overview

This document specifies the exploration project type - a workflow for research, investigation, and knowledge gathering. Exploration projects help users systematically investigate topics, document findings, and synthesize results into comprehensive summaries.

**Key Characteristics**:
- **Flexible topic discovery**: Add research topics dynamically as exploration progresses
- **Task-based lifecycle**: Each topic is a task with independent status tracking
- **Synthesis-focused**: Culminates in approved summary artifact
- **No GitHub integration**: Pure research workflow, no issue linking

## Goals and Non-Goals

**Goals**:
- Enable systematic research and investigation workflows
- Support dynamic topic discovery (add topics anytime during active research)
- Track research progress per topic independently
- Produce approved summary artifact documenting findings
- Maintain simple, flexible workflow (avoid over-structuring)

**Non-Goals**:
- GitHub issue integration (exploration is pre-implementation research)
- Rigid topic approval workflow (topics can be added/abandoned freely)
- Deliverable tracking beyond summary (exploration is knowledge gathering, not delivery)
- Complex artifact approval chains (only summary requires approval)

## Project Schema

### CUE Schema Definition

**File**: `cli/schemas/projects/exploration.cue`

```cue
package projects

import (
    "time"
    p "github.com/jmgilman/sow/cli/schemas/phases"
)

// ExplorationProjectState defines schema for exploration project type.
//
// Exploration follows a research workflow:
// Active (research topics) → Summarizing (synthesize findings) → Finalizing → Completed
#ExplorationProjectState: {
    // Discriminator
    project: {
        type: "exploration"  // Fixed discriminator value

        // Kebab-case project identifier
        name: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"

        // Git branch (typically explore/* prefix)
        branch: string & !=""

        // Research focus/question
        description: string

        // Timestamps
        created_at: time.Time
        updated_at: time.Time
    }

    // State machine position
    statechart: {
        current_state: "Active" | "Summarizing" | "Finalizing" | "Completed"
    }

    // 2-phase structure
    phases: {
        // Phase 1: Exploration (research and synthesis)
        exploration: p.#Phase & {
            // Custom status values for exploration workflow
            status: "active" | "summarizing" | "completed"
            enabled: true

            // Tasks represent research topics
            // Status tracks progress: pending, in_progress, completed, abandoned

            // Artifacts track research findings files
            // Summary artifact created during Summarizing state
        }

        // Phase 2: Finalization (move artifacts, create PR, cleanup)
        finalization: p.#Phase & {
            status: p.#GenericStatus
            enabled: true

            // Tasks created by orchestrator:
            // - Move summary to .sow/knowledge/explorations/
            // - Create PR with findings
            // - Delete .sow/project/
        }
    }
}
```

### Go Type

**File**: `cli/schemas/projects/exploration.go`

Hand-written Go type (CUE code generation doesn't handle unification patterns):

```go
package projects

import (
    "time"
    "github.com/jmgilman/sow/cli/schemas/phases"
)

// ExplorationProjectState is the Go representation of #ExplorationProjectState.
type ExplorationProjectState struct {
    Statechart struct {
        Current_state string `json:"current_state"`
    } `json:"statechart"`

    Project struct {
        Type        string    `json:"type"`
        Name        string    `json:"name"`
        Branch      string    `json:"branch"`
        Description string    `json:"description"`
        Created_at  time.Time `json:"created_at"`
        Updated_at  time.Time `json:"updated_at"`
    } `json:"project"`

    Phases struct {
        Exploration  phases.Phase `json:"exploration"`
        Finalization phases.Phase `json:"finalization"`
    } `json:"phases"`
}
```

## State Machine

### States

**Active** (exploration phase)
- **Purpose**: Active research - identify topics, investigate, document findings
- **Phase**: exploration
- **Phase status**: `"active"`
- **Duration**: Most of exploration time spent here
- **Orchestrator focus**: Help user identify research topics, work through investigations

**Summarizing** (exploration phase)
- **Purpose**: Synthesize findings into comprehensive summary document
- **Phase**: exploration
- **Phase status**: `"summarizing"`
- **Duration**: Short - create and approve summary
- **Orchestrator focus**: Generate summary from completed research, get user approval

**Finalizing** (finalization phase)
- **Purpose**: Move artifacts to permanent location, create PR, cleanup
- **Phase**: finalization
- **Phase status**: `"in_progress"`
- **Duration**: Short - automated finalization tasks
- **Orchestrator focus**: Execute finalization checklist

**Completed** (finalization phase)
- **Purpose**: Terminal state, exploration finished
- **Phase**: finalization
- **Phase status**: `"completed"`
- **Duration**: Permanent
- **Orchestrator focus**: None (project complete)

### State Transitions

```
Active
  │
  │ sow agent advance
  │ guard: AllTasksResolved
  ▼
Summarizing
  │
  │ sow agent complete
  │ guard: SummaryApproved
  ▼
Finalizing
  │
  │ sow agent complete
  │ guard: AllFinalizationTasksComplete
  ▼
Completed
```

### Transition Details

#### Active → Summarizing

**Trigger**: `sow agent advance`

**Event**: `EventBeginSummarizing`

**Guard**: `AllTasksResolved(tasks)`
```go
// All tasks must be either completed or abandoned
func AllTasksResolved(tasks []phases.Task) bool {
    for _, task := range tasks {
        if task.Status != "completed" && task.Status != "abandoned" {
            return false
        }
    }
    return len(tasks) > 0  // At least one task exists
}
```

**Phase status change**: `exploration.status = "active"` → `"summarizing"`

**Orchestrator behavior change**:
- Before: Add/update research topics, investigate, create findings
- After: Create summary document synthesizing all findings

#### Summarizing → Finalizing

**Trigger**: `sow agent complete`

**Event**: `EventCompleteSummarizing`

**Guard**: `SummaryApproved(artifacts)`
```go
// Summary artifact must exist and be approved
func SummaryApproved(artifacts []phases.Artifact) bool {
    for _, a := range artifacts {
        if strings.Contains(a.Path, "summary") &&
           a.Approved != nil && *a.Approved {
            return true
        }
    }
    return false
}
```

**Phase transition**: exploration → finalization

**Phase status changes**:
- `exploration.status = "completed"`
- `finalization.status = "in_progress"`

**Orchestrator behavior change**:
- Before: Create/approve summary
- After: Execute finalization tasks

#### Finalizing → Completed

**Trigger**: `sow agent complete`

**Event**: `EventCompleteFinalization`

**Guard**: `AllFinalizationTasksComplete(tasks)`
```go
// All finalization tasks must be completed
func AllFinalizationTasksComplete(tasks []phases.Task) bool {
    for _, task := range tasks {
        if task.Status != "completed" {
            return false
        }
    }
    return len(tasks) > 0
}
```

**Phase status change**: `finalization.status = "completed"`

**State machine**: Terminal state reached

## Task Lifecycle

### Tasks Represent Research Topics

Each task is a research topic/area to investigate:
- **Task name**: Research topic (e.g., "OAuth 2.0 flows", "JWT token structure")
- **Task description**: Scope of investigation, questions to answer
- **Task status**: Current state of research on this topic
- **Task refs**: Artifacts created during this topic's research (optional)

### Task Status Flow

```
pending
  │
  │ Start researching
  ▼
in_progress
  │
  ├─→ completed (research finished)
  │
  └─→ abandoned (topic not relevant)
```

**Standard statuses used** (no custom exploration statuses):
- `pending`: Topic identified, not yet researched
- `in_progress`: Actively investigating this topic
- `completed`: Research on this topic finished
- `abandoned`: Topic deemed not relevant or duplicate

### Task Operations by State

#### Active State

**Can perform**:
- ✅ Add new tasks (`sow agent task create`)
- ✅ Update task status (`sow agent task update <id> --status <status>`)
- ✅ Add task refs/metadata (`sow agent task update <id> --refs <path>`)
- ✅ Delete tasks if needed

**Rationale**: Active research is flexible - discover new topics anytime

#### Summarizing State

**Cannot perform**:
- ❌ Add new tasks (enforced by phase)
- ❌ Update task status (read-only)

**Rationale**:
- Summarizing means research is complete
- Adding tasks while summarizing indicates premature transition
- No backward transitions supported (keep architecture simple)

**Edge case**: User realizes more research needed while summarizing
- **Solution**: This indicates advancing to Summarizing was premature
- **Workaround**: Abandon current session, restart exploration
- **Future**: Could add backward transition if this becomes common problem

### Example Task Flow

**Scenario**: Researching authentication approaches

```bash
# Active state - identify topics
sow agent task create "OAuth 2.0 flows"
sow agent task create "JWT structure and validation"
sow agent task create "Session-based auth comparison"

# Start researching first topic
sow agent task update 001 --status in_progress

# During research, discover new topic
sow agent task create "Refresh token rotation strategies"

# Complete topics
sow agent task update 001 --status completed
sow agent task update 002 --status completed

# Abandon irrelevant topic
sow agent task update 003 --status abandoned

# Complete final topic
sow agent task update 004 --status completed

# All tasks resolved, ready to summarize
sow agent advance  # Active → Summarizing
```

## Artifact Management

### Artifact Types

Exploration uses two artifact categories:

**1. Research Findings** (during Active state)
- Files created during research (notes, diagrams, code samples)
- Added via `sow agent artifact add <path>`
- **Do not require approval** (`approved` field not set or false)
- Can be referenced by tasks via `task.refs`

**2. Summary Artifacts** (during Summarizing state)
- One or more documents synthesizing findings
- Created by orchestrator in `.sow/project/` (e.g., `summary.md`, `findings-detail.md`, `recommendations.md`)
- **All summaries require approval** (`approved = true` for each)
- **Multiple summaries allowed** to preserve valuable detail that would be lost in single document
- Finalized to `.sow/knowledge/explorations/<descriptive-folder>/` directory

### Artifact Operations by State

#### Active State

```bash
# Add research findings
sow agent artifact add project/findings/oauth-notes.md
sow agent artifact add project/findings/jwt-diagram.png

# Link artifacts to tasks (optional)
sow agent task update 001 --refs project/findings/oauth-notes.md
```

**Approval**: Not required for research findings

#### Summarizing State

```bash
# Orchestrator creates one or more summary documents
# Examples:
# - project/summary.md (high-level overview)
# - project/detailed-findings.md (comprehensive details)
# - project/recommendations.md (actionable recommendations)

# Add each summary as artifact
sow agent artifact add project/summary.md
sow agent artifact add project/detailed-findings.md
sow agent artifact add project/recommendations.md

# User reviews each summary
# Approve each when satisfied
sow agent artifact approve project/summary.md
sow agent artifact approve project/detailed-findings.md
sow agent artifact approve project/recommendations.md

# Guard now passes (all summaries approved), can complete phase
sow agent complete  # Summarizing → Finalizing
```

**Approval**: Required for **all** summary artifacts created in Summarizing state

### Summary Artifact Requirements

**Location**: `project/` directory (various filenames)

**Typical structure** (orchestrator can create multiple):
- `summary.md` - High-level overview and key findings
- `detailed-findings.md` - Comprehensive details per topic (preserves depth)
- `recommendations.md` - Actionable next steps
- `references.md` - External resources and citations
- Or any combination that preserves valuable information

**Rationale for multiple summaries**:
- Forcing everything into single document often loses valuable detail
- Different audiences need different levels of detail
- Separating concerns (findings vs recommendations) improves clarity
- Preserves comprehensive research while maintaining navigability

**Example summary.md content** (when multiple summaries exist):
```markdown
# [Topic] Exploration Summary

**Date**: [Date]
**Branch**: [Branch]

## Overview

[1-2 paragraph overview of the exploration scope and key outcomes]

## Research Questions

[Original exploration description/questions]

## High-Level Findings

[Synthesized insights across all topics - high-level only, 3-5 key points]

## Document Guide

This exploration produced multiple detailed documents:

- **[Detailed Findings](detailed-findings.md)** - Comprehensive research per topic with full context
- **[Recommendations](recommendations.md)** - Complete recommendation analysis with implementation guidance
- **[References](references.md)** - External resources and citations (if applicable)

## Quick Takeaways

[3-5 bullet points that capture the most important insights]

## Next Steps

[If applicable: recommended follow-up actions or explorations]
```

**Purpose of summary.md as ToC**:
- Provides single entry point for understanding exploration
- Enables agents to grasp key findings without reading all detailed documents
- Links to additional documents for those who need full context
- Ensures explorations remain discoverable and understandable over time

**Approval process**:
1. Orchestrator generates one or more summary documents from completed tasks
2. Orchestrator adds each as artifact: `sow agent artifact add project/<filename>.md`
3. Orchestrator presents each document to user for review
4. User requests edits → orchestrator updates summaries
5. User satisfied with all summaries → orchestrator runs: `sow agent artifact approve project/<filename>.md` for each
6. All summaries approved enables phase completion

**Finalization target**:
- All summary artifacts moved to: `.sow/knowledge/explorations/<descriptive-folder-name>/`
- Descriptive folder name based on exploration topic (e.g., `auth-approaches-2025-10`)
- Preserves document structure and relationships

**Required structure for multiple summaries** (validated during finalization):
- If more than one summary document exists, `summary.md` **must** exist and serve as overview/ToC
- `summary.md` provides high-level overview and links to other documents
- Allows agents to understand exploration without reading all detailed documents
- If only one summary exists, can be placed in folder or as single file
- **Orchestrator responsibility**: Create summary.md during Summarizing state when creating multiple documents

## Phase Implementations

### Exploration Phase

**File**: `cli/internal/project/exploration/exploration.go`

```go
package exploration

import (
    "fmt"
    "github.com/jmgilman/sow/cli/internal/project"
    "github.com/jmgilman/sow/cli/internal/project/domain"
    "github.com/jmgilman/sow/cli/schemas/phases"
    "github.com/jmgilman/sow/cli/schemas/projects"
)

type ExplorationPhase struct {
    state     *phases.Phase
    artifacts *project.ArtifactCollection
    tasks     *project.TaskCollection
    project   *ExplorationProject
}

func newExplorationPhase(proj *ExplorationProject) *ExplorationPhase {
    return &ExplorationPhase{
        state:     &proj.state.Phases.Exploration,
        artifacts: project.NewArtifactCollection(&proj.state.Phases.Exploration, proj.ctx),
        tasks:     project.NewTaskCollection(&proj.state.Phases.Exploration, proj, proj.ctx),
        project:   proj,
    }
}

// Advance handles intra-phase state progression
func (p *ExplorationPhase) Advance() (*domain.PhaseOperationResult, error) {
    switch p.state.Status {
    case "active":
        // Validate all tasks resolved
        if !allTasksResolved(p.state.Tasks) {
            unresolvedCount := countUnresolved(p.state.Tasks)
            return nil, fmt.Errorf(
                "cannot advance: %d tasks are not yet completed or abandoned",
                unresolvedCount,
            )
        }

        // Update phase status
        p.state.Status = "summarizing"
        if err := p.project.Save(); err != nil {
            return nil, fmt.Errorf("failed to save state: %w", err)
        }

        // Return event to trigger state transition
        return domain.WithEvent(EventBeginSummarizing), nil

    case "summarizing":
        // Already in final state within this phase
        return nil, fmt.Errorf(
            "already in Summarizing state - use 'sow agent complete' to finish exploration phase",
        )

    default:
        return nil, fmt.Errorf("unknown exploration status: %s", p.state.Status)
    }
}

// Complete handles phase completion (Summarizing → Finalization)
func (p *ExplorationPhase) Complete() (*domain.PhaseOperationResult, error) {
    // Can only complete from Summarizing state
    if p.state.Status != "summarizing" {
        return nil, fmt.Errorf(
            "cannot complete exploration from %s state - advance to Summarizing first",
            p.state.Status,
        )
    }

    // Validate all summary artifacts approved (guard will also check)
    if !summaryApproved(p.state.Artifacts) {
        unapprovedCount := countUnapprovedSummaries(p.state.Artifacts)
        return nil, fmt.Errorf(
            "%d summary artifact(s) not yet approved - approve all summaries before completing",
            unapprovedCount,
        )
    }

    // Update phase status
    p.state.Status = "completed"
    if err := p.project.Save(); err != nil {
        return nil, fmt.Errorf("failed to save state: %w", err)
    }

    // Return event to trigger phase transition
    return domain.WithEvent(EventCompleteSummarizing), nil
}

// AddTask creates a new research topic
func (p *ExplorationPhase) AddTask(name string, opts ...domain.TaskOption) (*domain.Task, error) {
    // Prevent task creation in Summarizing state
    if p.state.Status == "summarizing" {
        return nil, fmt.Errorf(
            "cannot add tasks in Summarizing state - research must be complete before summarizing",
        )
    }

    return p.tasks.Add(name, opts...)
}

// Artifact operations delegate to collection
func (p *ExplorationPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
    return p.artifacts.Add(path, opts...)
}

func (p *ExplorationPhase) ApproveArtifact(path string) (*domain.PhaseOperationResult, error) {
    if err := p.artifacts.Approve(path); err != nil {
        return nil, err
    }

    // Approving artifacts doesn't trigger transitions in exploration
    // (only summary approval matters, and that's checked by Complete guard)
    return domain.NoEvent(), nil
}

// ... other Phase interface methods
```

### Finalization Phase

**File**: `cli/internal/project/exploration/finalization.go`

Finalization phase with special handling for multi-document explorations.

```go
package exploration

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/jmgilman/sow/cli/internal/project"
    "github.com/jmgilman/sow/cli/internal/project/domain"
    "github.com/jmgilman/sow/cli/schemas/phases"
)

type FinalizationPhase struct {
    state   *phases.Phase
    tasks   *project.TaskCollection
    project *ExplorationProject
}

func newFinalizationPhase(proj *ExplorationProject) *FinalizationPhase {
    return &FinalizationPhase{
        state:   &proj.state.Phases.Finalization,
        tasks:   project.NewTaskCollection(&proj.state.Phases.Finalization, proj, proj.ctx),
        project: proj,
    }
}

// Complete validates summary structure before allowing finalization completion
func (p *FinalizationPhase) Complete() (*domain.PhaseOperationResult, error) {
    // Validate summary structure before checking task completion
    if err := p.validateSummaryStructure(); err != nil {
        return nil, err
    }

    // Validate all tasks completed
    for _, task := range p.state.Tasks {
        if task.Status != "completed" {
            return nil, fmt.Errorf("all finalization tasks must be completed")
        }
    }

    p.state.Status = "completed"
    if err := p.project.Save(); err != nil {
        return nil, err
    }

    return domain.WithEvent(EventCompleteFinalization), nil
}

// validateSummaryStructure ensures multi-document explorations have summary.md as ToC
func (p *FinalizationPhase) validateSummaryStructure() error {
    explorationPhase := &p.project.state.Phases.Exploration

    // Collect summary artifacts (have approved field set)
    summaries := []phases.Artifact{}
    for _, a := range explorationPhase.Artifacts {
        if a.Approved != nil {
            summaries = append(summaries, a)
        }
    }

    // Single summary: no ToC requirement
    if len(summaries) <= 1 {
        return nil
    }

    // Multiple summaries: verify summary.md exists
    hasSummaryMd := false
    for _, s := range summaries {
        filename := filepath.Base(s.Path)
        if filename == "summary.md" {
            hasSummaryMd = true
            break
        }
    }

    if !hasSummaryMd {
        // List the existing summaries for context
        summaryNames := make([]string, 0, len(summaries))
        for _, s := range summaries {
            summaryNames = append(summaryNames, filepath.Base(s.Path))
        }

        return fmt.Errorf(
            "multiple summary documents detected (%v) but summary.md is missing - "+
            "when creating multiple summaries, summary.md must serve as an overview and table of contents. "+
            "Please create summary.md with:\n"+
            "- High-level overview of the exploration\n"+
            "- Links to all other summary documents\n"+
            "- Quick takeaways (3-5 key insights)\n"+
            "Then add and approve it before completing the exploration.",
            summaryNames,
        )
    }

    return nil
}

// Advance not supported - finalization has no internal states
func (p *FinalizationPhase) Advance() (*domain.PhaseOperationResult, error) {
    return nil, project.ErrNotSupported
}

// ... other Phase interface methods
```

**Key implementation details**:
- `validateSummaryStructure()` called during `Complete()`
- Counts summary artifacts (artifacts with `approved` field set)
- If > 1 summary and no `summary.md`: **fails with descriptive error**
- Error message lists existing summaries and explains what's needed
- Orchestrator must create summary.md before completing exploration

## Guards

**File**: `cli/internal/project/exploration/guards.go`

```go
package exploration

import "github.com/jmgilman/sow/cli/schemas/phases"

// AllTasksResolved checks if all research topics are completed or abandoned.
// Guards Active → Summarizing transition.
func AllTasksResolved(tasks []phases.Task) bool {
    if len(tasks) == 0 {
        return false  // Must have at least one task
    }

    for _, task := range tasks {
        if task.Status != "completed" && task.Status != "abandoned" {
            return false
        }
    }
    return true
}

// SummaryApproved checks if at least one summary artifact exists and all summaries are approved.
// Guards Summarizing → Finalizing transition.
//
// A summary is identified by having been created during Summarizing state.
// In practice, this means any artifact added after transitioning to Summarizing.
// All such artifacts must be approved before transitioning to Finalizing.
func SummaryApproved(artifacts []phases.Artifact) bool {
    summaries := []phases.Artifact{}

    // Collect all summary artifacts
    // Convention: summaries are artifacts created during Summarizing state
    // They don't have to have "summary" in the name
    for _, a := range artifacts {
        // For now, identify summaries by checking if approved field is set
        // Research findings typically don't set approved field
        // Summary artifacts explicitly require approval
        if a.Approved != nil {
            summaries = append(summaries, a)
        }
    }

    // Must have at least one summary
    if len(summaries) == 0 {
        return false
    }

    // All summaries must be approved
    for _, s := range summaries {
        if !*s.Approved {
            return false
        }
    }

    return true
}

// Note: This implementation assumes that only summary artifacts have the approved
// field set (even if false). Research findings added during Active state should
// not set the approved field at all. This provides a clean distinction between
// research findings and summaries.

// AllFinalizationTasksComplete checks if all finalization tasks are done.
// Guards Finalizing → Completed transition.
func AllFinalizationTasksComplete(tasks []phases.Task) bool {
    if len(tasks) == 0 {
        return false  // Must have finalization tasks
    }

    for _, task := range tasks {
        if task.Status != "completed" {
            return false
        }
    }
    return true
}

// Helper functions
func allTasksResolved(tasks []phases.Task) bool {
    return AllTasksResolved(tasks)
}

func summaryApproved(artifacts []phases.Artifact) bool {
    return SummaryApproved(artifacts)
}

func countUnresolved(tasks []phases.Task) int {
    count := 0
    for _, task := range tasks {
        if task.Status != "completed" && task.Status != "abandoned" {
            count++
        }
    }
    return count
}

func countUnapprovedSummaries(artifacts []phases.Artifact) int {
    count := 0
    for _, a := range artifacts {
        // Summaries have approved field set
        if a.Approved != nil && !*a.Approved {
            count++
        }
    }
    return count
}
```

## States and Events

**File**: `cli/internal/project/exploration/states.go`

```go
package exploration

import "github.com/jmgilman/sow/cli/internal/project/statechart"

// State constants for exploration workflow
const (
    ExplorationActive      = statechart.State("Active")
    ExplorationSummarizing = statechart.State("Summarizing")
    ExplorationFinalizing  = statechart.State("Finalizing")
    ExplorationCompleted   = statechart.State("Completed")
)
```

**File**: `cli/internal/project/exploration/events.go`

```go
package exploration

import "github.com/jmgilman/sow/cli/internal/project/statechart"

// Event constants for exploration transitions
const (
    // Active → Summarizing
    EventBeginSummarizing = statechart.Event("begin_summarizing")

    // Summarizing → Finalizing
    EventCompleteSummarizing = statechart.Event("complete_summarizing")

    // Finalizing → Completed
    EventCompleteFinalization = statechart.Event("complete_finalization")
)
```

## Prompt Generation

**File**: `cli/internal/project/exploration/prompts.go`

```go
package exploration

import (
    "fmt"
    "strings"

    "github.com/jmgilman/sow/cli/internal/project/statechart"
    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/jmgilman/sow/cli/schemas"
)

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
    switch state {
    case ExplorationActive:
        return g.generateActivePrompt(projectState)
    case ExplorationSummarizing:
        return g.generateSummarizingPrompt(projectState)
    case ExplorationFinalizing:
        return g.generateFinalizingPrompt(projectState)
    default:
        return "", fmt.Errorf("unknown exploration state: %s", state)
    }
}

func (g *ExplorationPromptGenerator) generateActivePrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    // Project header
    buf.WriteString(g.components.ProjectHeader(projectState))
    buf.WriteString("\n")

    // Current state
    buf.WriteString("## Current State: Active Research\n\n")
    buf.WriteString("You are in the Active state of exploration. ")
    buf.WriteString("Your focus is identifying research topics, investigating them, ")
    buf.WriteString("and documenting findings.\n\n")

    // Research topics
    buf.WriteString("## Research Topics\n\n")
    phase := projectState.Phases.Exploration

    if len(phase.Tasks) == 0 {
        buf.WriteString("No topics identified yet. Start by creating research topics:\n\n")
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent task create \"Topic name\"\n")
        buf.WriteString("```\n\n")
    } else {
        pending := 0
        inProgress := 0
        completed := 0
        abandoned := 0

        for _, task := range phase.Tasks {
            switch task.Status {
            case "pending":
                pending++
            case "in_progress":
                inProgress++
            case "completed":
                completed++
            case "abandoned":
                abandoned++
            }
        }

        buf.WriteString(fmt.Sprintf("Total: %d topics\n", len(phase.Tasks)))
        buf.WriteString(fmt.Sprintf("- Pending: %d\n", pending))
        buf.WriteString(fmt.Sprintf("- In Progress: %d\n", inProgress))
        buf.WriteString(fmt.Sprintf("- Completed: %d\n", completed))
        buf.WriteString(fmt.Sprintf("- Abandoned: %d\n", abandoned))
        buf.WriteString("\n")

        // List topics
        buf.WriteString("### Topics:\n\n")
        for _, task := range phase.Tasks {
            buf.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", task.Id, task.Name, task.Status))
        }
        buf.WriteString("\n")
    }

    // Next steps
    buf.WriteString("## Next Steps\n\n")
    if allTasksResolved(phase.Tasks) && len(phase.Tasks) > 0 {
        buf.WriteString("✓ All research topics resolved!\n\n")
        buf.WriteString("Ready to create summary. Run:\n\n")
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent advance\n")
        buf.WriteString("```\n\n")
    } else {
        buf.WriteString("Continue research:\n")
        buf.WriteString("- Add topics: `sow agent task create \"Topic\"`\n")
        buf.WriteString("- Update status: `sow agent task update <id> --status <status>`\n")
        buf.WriteString("- Add findings: `sow agent artifact add <path>`\n\n")

        unresolvedCount := countUnresolved(phase.Tasks)
        if unresolvedCount > 0 {
            buf.WriteString(fmt.Sprintf("(%d topics remaining)\n\n", unresolvedCount))
        }
    }

    return buf.String(), nil
}

func (g *ExplorationPromptGenerator) generateSummarizingPrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    buf.WriteString(g.components.ProjectHeader(projectState))
    buf.WriteString("\n")

    buf.WriteString("## Current State: Summarizing Findings\n\n")
    buf.WriteString("All research topics are resolved. Your focus now is creating ")
    buf.WriteString("a comprehensive summary document synthesizing findings across all topics.\n\n")

    // Research completed
    phase := projectState.Phases.Exploration
    buf.WriteString("## Research Completed\n\n")

    completed := []phases.Task{}
    abandoned := []phases.Task{}

    for _, task := range phase.Tasks {
        if task.Status == "completed" {
            completed = append(completed, task)
        } else if task.Status == "abandoned" {
            abandoned = append(abandoned, task)
        }
    }

    buf.WriteString(fmt.Sprintf("**Completed**: %d topics\n", len(completed)))
    for _, task := range completed {
        buf.WriteString(fmt.Sprintf("- %s\n", task.Name))
    }
    buf.WriteString("\n")

    if len(abandoned) > 0 {
        buf.WriteString(fmt.Sprintf("**Abandoned**: %d topics\n", len(abandoned)))
        for _, task := range abandoned {
            buf.WriteString(fmt.Sprintf("- %s\n", task.Name))
        }
        buf.WriteString("\n")
    }

    // Summary status
    buf.WriteString("## Summary Artifacts\n\n")

    // Collect summaries (artifacts with approved field set)
    summaries := []phases.Artifact{}
    for _, artifact := range phase.Artifacts {
        if artifact.Approved != nil {
            summaries = append(summaries, artifact)
        }
    }

    if len(summaries) == 0 {
        buf.WriteString("No summaries created yet.\n\n")
    } else {
        approvedCount := 0
        for _, s := range summaries {
            if *s.Approved {
                approvedCount++
            }
        }

        buf.WriteString(fmt.Sprintf("Total: %d summary document(s)\n", len(summaries)))
        buf.WriteString(fmt.Sprintf("Approved: %d\n\n", approvedCount))

        for _, s := range summaries {
            status := "Pending approval"
            if *s.Approved {
                status = "✓ Approved"
            }
            buf.WriteString(fmt.Sprintf("- %s (%s)\n", s.Path, status))
        }
        buf.WriteString("\n")
    }

    // Next steps
    buf.WriteString("## Next Steps\n\n")

    allApproved := summaryApproved(phase.Artifacts)

    if allApproved {
        buf.WriteString("✓ All summaries approved!\n\n")
        buf.WriteString("Ready to finalize. Run:\n\n")
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent complete\n")
        buf.WriteString("```\n\n")
    } else if len(summaries) > 0 {
        unapproved := countUnapprovedSummaries(phase.Artifacts)
        buf.WriteString(fmt.Sprintf("Review and approve remaining %d summary document(s):\n\n", unapproved))
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent artifact approve <summary-path>\n")
        buf.WriteString("```\n\n")
    } else {
        buf.WriteString("Create summary document(s):\n")
        buf.WriteString("1. Write summaries (e.g., `project/summary.md`, `project/detailed-findings.md`)\n")
        buf.WriteString("2. Add each as artifact: `sow agent artifact add project/<filename>.md`\n")
        buf.WriteString("3. Review and approve each: `sow agent artifact approve project/<filename>.md`\n\n")
        buf.WriteString("Tips:\n")
        buf.WriteString("- Multiple summary documents help preserve valuable detail\n")
        buf.WriteString("- When creating multiple summaries, `summary.md` must serve as overview/ToC\n")
        buf.WriteString("- `summary.md` should link to all other documents and provide high-level takeaways\n\n")
    }

    return buf.String(), nil
}

func (g *ExplorationPromptGenerator) generateFinalizingPrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    buf.WriteString(g.components.ProjectHeader(projectState))
    buf.WriteString("\n")

    buf.WriteString("## Current State: Finalizing\n\n")
    buf.WriteString("Summary approved. Finalizing exploration by moving artifacts, ")
    buf.WriteString("creating PR, and cleaning up.\n\n")

    // Finalization tasks
    phase := projectState.Phases.Finalization
    buf.WriteString("## Finalization Tasks\n\n")

    for _, task := range phase.Tasks {
        status := "[ ]"
        if task.Status == "completed" {
            status = "[✓]"
        }
        buf.WriteString(fmt.Sprintf("%s %s\n", status, task.Name))
    }
    buf.WriteString("\n")

    allComplete := true
    for _, task := range phase.Tasks {
        if task.Status != "completed" {
            allComplete = false
            break
        }
    }

    if allComplete && len(phase.Tasks) > 0 {
        buf.WriteString("All finalization tasks complete! Run:\n\n")
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent complete\n")
        buf.WriteString("```\n\n")
    }

    return buf.String(), nil
}
```

## Orchestrator Workflow

### Active State Workflow

1. **Initialize exploration** (via wizard on `explore/*` branch)
2. **Identify research topics**:
   - Orchestrator helps user brainstorm topics
   - Creates task for each topic: `sow agent task create "Topic"`
   - Topics start in `pending` status
3. **Research topics**:
   - Update status: `sow agent task update <id> --status in_progress`
   - Create findings artifacts: `sow agent artifact add <path>`
   - Optionally link findings to tasks: `sow agent task update <id> --refs <path>`
4. **Complete/abandon topics**:
   - Complete: `sow agent task update <id> --status completed`
   - Abandon: `sow agent task update <id> --status abandoned`
5. **Discover new topics** (dynamic):
   - Can add topics anytime: `sow agent task create "New topic"`
   - No limit on topic count
6. **When all resolved**:
   - Guard passes: `AllTasksResolved = true`
   - Orchestrator prompts user to advance
   - User/orchestrator runs: `sow agent advance`

### Summarizing State Workflow

1. **Transition to Summarizing**: `sow agent advance` from Active
2. **Create summary documents**:
   - Orchestrator synthesizes findings from completed topics
   - Creates one or more summary documents:
     - `project/summary.md` - High-level overview
     - `project/detailed-findings.md` - Comprehensive details (optional)
     - `project/recommendations.md` - Actionable next steps (optional)
   - Multiple documents preserve valuable detail that would be lost in single summary
3. **Add summary artifacts**:
   - `sow agent artifact add project/summary.md`
   - `sow agent artifact add project/detailed-findings.md`
   - `sow agent artifact add project/recommendations.md`
4. **Review and iterate**:
   - Orchestrator presents each summary to user
   - User requests edits → orchestrator updates summaries
   - Repeat until user satisfied with all documents
5. **Approve summaries**:
   - `sow agent artifact approve project/summary.md`
   - `sow agent artifact approve project/detailed-findings.md`
   - `sow agent artifact approve project/recommendations.md`
   - **All summaries must be approved** before proceeding
6. **Complete exploration**: `sow agent complete`

### Finalizing State Workflow

1. **Transition to Finalizing**: `sow agent complete` from Summarizing

2. **Validate summary structure** (automatic):
   - Count approved summary artifacts
   - **If multiple summaries**: Verify `summary.md` exists
     - If missing: **Fail with error** - orchestrator must create summary.md as ToC
     - If exists: Proceed with finalization
   - **If single summary**: Can use any filename (no ToC requirement)

3. **Create finalization tasks** (automatic):
   - Determine target structure:
     - **Single summary**: Can move to `.sow/knowledge/explorations/<name>.md` OR create folder
     - **Multiple summaries**: Must create folder `.sow/knowledge/explorations/<descriptive-name>/`
   - Create descriptive folder if needed (e.g., `auth-approaches-2025-10`)
   - Move all summary artifacts to target location (preserving structure)
   - Optionally move key research findings
   - Create PR with link to exploration location
   - Delete `.sow/project/` directory

4. **Execute tasks**:
   - Orchestrator completes each task sequentially
   - Updates task status to `completed`

5. **Complete finalization**: `sow agent complete`

6. **Exploration finished**: State machine reaches Completed

**Finalized structure examples**:

Single summary:
```
.sow/knowledge/explorations/quick-spike-2025-10.md
```

Multiple summaries (requires summary.md as ToC):
```
.sow/knowledge/explorations/auth-approaches-2025-10/
├── summary.md                  # REQUIRED: Overview and ToC linking to other docs
├── detailed-findings.md        # Comprehensive research
└── recommendations.md          # Actionable next steps
```

## Testing Strategy

### Unit Tests

**Schema validation**:
- ExplorationProjectState validates correctly
- Phase status constraints enforced
- State machine state constraints enforced

**Guards**:
- `AllTasksResolved` returns false with pending tasks
- `AllTasksResolved` returns true when all completed/abandoned
- `SummaryApproved` requires at least one summary, all approved
- `AllFinalizationTasksComplete` validates all tasks completed

**Summary structure validation**:
- Multiple summaries require `summary.md` as ToC
- Validation occurs during finalization `Complete()` call
- Fails with descriptive error if missing (orchestrator must create)

**Phase operations**:
- `Advance()` from Active requires all tasks resolved
- `Advance()` from Summarizing returns error
- `Complete()` from Active returns error
- `Complete()` from Summarizing requires summary approved
- `AddTask()` works in Active, fails in Summarizing

### Integration Tests

**Full workflow - single summary**:
1. Create exploration project on `explore/test` branch
2. Add 3 research topics
3. Complete all topics
4. Advance to Summarizing
5. Create and approve single summary
6. Complete exploration
7. Complete finalization
8. Verify single file moved to `.sow/knowledge/explorations/`

**Full workflow - multiple summaries**:
1. Create exploration project on `explore/test` branch
2. Add 3 research topics
3. Complete all topics
4. Advance to Summarizing
5. Create and approve multiple summaries (including summary.md as ToC)
6. Complete exploration (triggers finalization)
7. Verify validation passes
8. Complete finalization
9. Verify folder created with all summaries

**Edge case - missing summary.md**:
1. Create multiple summaries without summary.md
2. Attempt to complete finalization
3. Verify error message lists existing summaries and explains requirement
4. Create and approve summary.md
5. Retry finalization - should succeed

**Edge cases**:
- Advance with unresolved tasks (should fail)
- Complete without summary approval (should fail)
- Add task in Summarizing state (should fail)
- Skip from Active to Finalizing (should be prevented by state machine)
- Multiple summaries without summary.md (should fail finalization with clear error)

### Manual Verification

- Real exploration workflow on test topic
- Verify prompts provide useful guidance
- Verify state transitions feel natural
- Verify can't get stuck in any state
- Verify finalization cleanup works correctly

## Migration Notes

No migration from old exploration mode - users restart active explorations.

Exploration sessions are typically short (days, not weeks), so breaking change acceptable.

## Success Criteria

1. ✅ Can create exploration project on `explore/*` branch
2. ✅ Can add/update research topics dynamically during Active state
3. ✅ Cannot add topics in Summarizing state
4. ✅ Can advance to Summarizing when all tasks resolved
5. ✅ Can create and approve multiple summary artifacts
6. ✅ Can complete exploration when all summaries approved
7. ✅ Multiple summaries without `summary.md` fail finalization with clear error
8. ✅ Single summary can be finalized without ToC requirement
9. ✅ Finalization moves artifacts to appropriate structure (file or folder)
10. ✅ Finalization creates PR with link to exploration location
11. ✅ Zero-context resumability works (can stop/restart at any state)
12. ✅ Prompts provide clear guidance at each state
13. ✅ All guards prevent invalid transitions

## Future Enhancements

**Not in scope for initial implementation** (can revisit if needed):

- Backward transitions (Summarizing → Active if more research needed)
- Task templates for common exploration types
- Artifact approval workflow for research findings (currently only summary)
- Multi-user collaboration on explorations
- Exploration templates (security, architecture, technology evaluation)
- Linking multiple explorations (series of related investigations)
