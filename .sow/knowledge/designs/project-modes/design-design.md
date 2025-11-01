# Design Project Type Design

**Author**: Architecture Team
**Date**: 2025-10-31
**Status**: Proposed
**Depends On**: [Core Design](core-design.md)

## Overview

This document specifies the design project type - a workflow for creating architecture and design documentation. Design projects help users track design artifacts (ADRs, design docs, architecture docs, diagrams) from planning through approval and finalization.

**Key Characteristics**:
- **Input tracking**: Register sources that inform design (explorations, references, existing docs)
- **Task-based document tracking**: Each task represents a document to create
- **Review workflow**: Draft → needs_review → completed (with auto-approval)
- **Metadata-driven**: Tasks store document type, target location, artifact path
- **No GitHub integration**: Pure design workflow, no issue linking

## Goals and Non-Goals

**Goals**:
- Track design inputs that inform decisions
- Plan design outputs as tasks before drafting
- Support iterative review workflow (needs_review status)
- Auto-approve artifacts when tasks completed
- Move approved documents to appropriate knowledge locations
- Prevent creating artifacts before planning tasks

**Non-Goals**:
- GitHub issue integration (design is pre-implementation planning)
- Multi-document ToC requirement (each document stands alone)
- Template enforcement (templates are guidance, not validation)
- Complex approval chains (single review cycle sufficient)

## Project Schema

### CUE Schema Definition

**File**: `cli/schemas/projects/design.cue`

```cue
package projects

import (
    "time"
    p "github.com/jmgilman/sow/cli/schemas/phases"
)

// DesignProjectState defines schema for design project type.
//
// Design follows a document creation workflow:
// Active (plan, draft, review documents) → Finalizing → Completed
#DesignProjectState: {
    // Discriminator
    project: {
        type: "design"  // Fixed discriminator value

        // Kebab-case project identifier
        name: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"

        // Git branch (typically design/* prefix)
        branch: string & !=""

        // Design focus/scope
        description: string

        // Timestamps
        created_at: time.Time
        updated_at: time.Time
    }

    // State machine position
    statechart: {
        current_state: "Active" | "Finalizing" | "Completed"
    }

    // 2-phase structure
    phases: {
        // Phase 1: Design (plan and create documents)
        design: p.#Phase & {
            // Custom status values for design workflow
            status: "active" | "completed"
            enabled: true

            // Inputs track sources informing design
            // (explorations, references, existing docs)
            inputs?: [...p.#Artifact]

            // Tasks represent documents to create
            // Task.metadata stores:
            //   - artifact_path: path to drafted document
            //   - document_type: "adr", "design", "architecture", etc.
            //   - target_location: where to move during finalization
            //   - template: optional template identifier

            // Artifacts are drafted documents
            // Auto-approved when task marked completed
        }

        // Phase 2: Finalization (move docs, create PR, cleanup)
        finalization: p.#Phase & {
            status: p.#GenericStatus
            enabled: true

            // Tasks created by orchestrator:
            // - Move approved documents to target locations
            // - Create PR with design artifacts
            // - Delete .sow/project/
        }
    }
}
```

### Go Type

**File**: `cli/schemas/projects/design.go`

```go
package projects

import (
    "time"
    "github.com/jmgilman/sow/cli/schemas/phases"
)

// DesignProjectState is the Go representation of #DesignProjectState.
type DesignProjectState struct {
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
        Design       phases.Phase `json:"design"`
        Finalization phases.Phase `json:"finalization"`
    } `json:"phases"`
}
```

## State Machine

### States

**Active** (design phase)
- **Purpose**: Plan outputs, draft documents, iterate through reviews
- **Phase**: design
- **Phase status**: `"active"`
- **Duration**: Majority of design time
- **Orchestrator focus**: Create document tasks, draft documents, iterate on feedback

**Finalizing** (finalization phase)
- **Purpose**: Move approved documents to targets, create PR, cleanup
- **Phase**: finalization
- **Phase status**: `"in_progress"`
- **Duration**: Short - automated finalization tasks
- **Orchestrator focus**: Execute finalization checklist

**Completed** (finalization phase)
- **Purpose**: Terminal state, design finished
- **Phase**: finalization
- **Phase status**: `"completed"`
- **Duration**: Permanent
- **Orchestrator focus**: None (project complete)

### State Transitions

```
Active
  │
  │ sow agent complete
  │ guard: AllDocumentsApproved
  ▼
Finalizing
  │
  │ sow agent complete
  │ guard: AllFinalizationTasksComplete
  ▼
Completed
```

### Transition Details

#### Active → Finalizing

**Trigger**: `sow agent complete`

**Event**: `EventCompleteDesign`

**Guard**: `AllDocumentsApproved(tasks)`
```go
// All tasks must be completed or abandoned
// At least one task must be completed (can't complete with all abandoned)
func AllDocumentsApproved(tasks []phases.Task) bool {
    if len(tasks) == 0 {
        return false  // Must have at least one task
    }

    hasCompleted := false
    for _, task := range tasks {
        // Check for unresolved tasks
        if task.Status != "completed" && task.Status != "abandoned" {
            return false
        }
        // Track if we have at least one completed
        if task.Status == "completed" {
            hasCompleted = true
        }
    }

    return hasCompleted
}
```

**Phase transition**: design → finalization

**Phase status changes**:
- `design.status = "completed"`
- `finalization.status = "in_progress"`

**Orchestrator behavior change**:
- Before: Draft/review documents
- After: Execute finalization tasks

#### Finalizing → Completed

**Trigger**: `sow agent complete`

**Event**: `EventCompleteFinalization`

**Guard**: `AllFinalizationTasksComplete(tasks)`
```go
// All finalization tasks must be completed
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
```

**Phase status change**: `finalization.status = "completed"`

**State machine**: Terminal state reached

## Task Lifecycle

### Tasks Represent Documents to Create

Each task tracks a document's lifecycle from planning through approval:
- **Task name**: Document title/description
- **Task description**: Document purpose, scope
- **Task status**: Current state of document creation
- **Task metadata**: Links to artifact, stores document metadata

### Task Status Flow

```
pending
  │
  │ Start drafting
  ▼
in_progress
  │
  │ Document drafted, ready for review
  ▼
needs_review
  │
  ├─→ in_progress (changes requested)
  │
  └─→ completed (approved) + artifact auto-approved
```

**Or abandon if not needed**:
```
pending/in_progress/needs_review
  │
  └─→ abandoned (document not needed)
```

**Standard statuses used** (no custom design statuses):
- `pending`: Document planned, not yet drafted
- `in_progress`: Actively drafting document
- `needs_review`: Document ready for human review
- `completed`: Document approved, artifact auto-approved
- `abandoned`: Document not needed

### Task Metadata Structure

```yaml
metadata:
  artifact_path: "project/auth-design.md"        # Path to drafted document
  document_type: "design"                        # Type: adr, design, architecture, diagram, spec
  target_location: ".sow/knowledge/designs/auth-design.md"  # Where to move during finalization
  template: "design-doc"                         # Optional: template identifier
```

**When metadata is set**:
- `document_type` and `target_location`: Set when task created (planning)
- `artifact_path`: Set when artifact added (drafting)
- `template`: Optional, set when task created

### Task-to-Artifact Linking

**Task completion auto-approves artifact**:

1. Task in `needs_review` status
2. Human reviews document, satisfied
3. Orchestrator runs: `sow agent task update <id> --status completed`
4. Design phase detects completion, reads `task.metadata.artifact_path`
5. Finds artifact at that path, sets `approved = true`
6. Saves state

**Validation**: When transitioning to `completed`, artifact must exist at `metadata.artifact_path`

### Task Operations by State

#### Active State

**Can perform**:
- ✅ Create tasks (`sow agent task create`)
- ✅ Update task status
- ✅ Update task metadata
- ✅ Add artifacts (only if at least one task exists)
- ✅ Delete tasks if needed

**Guards**:
- Cannot add artifacts until at least one task exists

**Rationale**: Tasks represent planned outputs - must plan before drafting

#### Finalizing State

**Cannot perform**:
- ❌ Create/update tasks (read-only)
- ❌ Add artifacts (read-only)

**Rationale**: Finalization only moves existing approved documents

## Input Management

### Input Artifacts

Track sources that inform the design:
- Exploration findings
- Existing documentation
- External references (URLs, papers, etc.)
- Related code examples

**Adding inputs**:
```bash
# Add exploration as input
sow agent artifact add ../../knowledge/explorations/auth-research.md --input

# Add external reference
sow agent artifact add https://oauth.net/2/ --input --description "OAuth 2.0 spec"
```

**Storage**: `design.inputs` array (separate from `design.artifacts`)

**Rationale**:
- Clear separation: inputs inform, artifacts are outputs
- Traceable: know what informed each design decision
- Resumable: new orchestrator can read inputs for context

### Input vs Output Artifacts

**Inputs** (`phase.inputs`):
- Sources that inform design
- Added with `--input` flag
- Do not require approval
- Not moved during finalization

**Outputs** (`phase.artifacts`):
- Documents being created
- Linked to tasks via metadata
- Require approval (auto-approved on task completion)
- Moved to target locations during finalization

## Artifact Management

### Artifact Guard: Task Existence

**Cannot add artifacts before creating tasks**:

```go
func (p *DesignPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
    if len(p.state.Tasks) == 0 {
        return fmt.Errorf(
            "cannot add artifacts before creating document tasks - "+
            "create at least one task to track document lifecycle first",
        )
    }
    return p.artifacts.Add(path, opts...)
}
```

**Rationale**:
- Enforces planning before drafting
- Ensures every document has task tracking it
- Prevents orphaned artifacts

### Artifact Auto-Approval

**Triggered when task completed**:

```go
func (p *DesignPhase) UpdateTaskStatus(id string, status string) error {
    task := p.GetTask(id)

    // Validate artifact exists if moving to completed
    if status == "completed" {
        if task.Metadata == nil || task.Metadata["artifact_path"] == nil {
            return fmt.Errorf("task %s has no artifact_path - cannot mark completed", id)
        }

        artifactPath := task.Metadata["artifact_path"].(string)
        artifact := p.findArtifact(artifactPath)

        if artifact == nil {
            return fmt.Errorf(
                "artifact not found at %s - add artifact before completing task",
                artifactPath,
            )
        }

        // Auto-approve artifact
        approved := true
        artifact.Approved = &approved
    }

    // Update task status
    task.Status = status
    return p.project.Save()
}
```

**Workflow**:
1. Task in `needs_review`
2. `sow agent task update <id> --status completed`
3. Phase validates artifact exists at `task.metadata.artifact_path`
4. Phase sets `artifact.approved = true`
5. Task marked completed

## Phase Implementations

### Design Phase

**File**: `cli/internal/project/design/design.go`

```go
package design

import (
    "fmt"
    "github.com/jmgilman/sow/cli/internal/project"
    "github.com/jmgilman/sow/cli/internal/project/domain"
    "github.com/jmgilman/sow/cli/schemas/phases"
)

type DesignPhase struct {
    state     *phases.Phase
    artifacts *project.ArtifactCollection
    tasks     *project.TaskCollection
    project   *DesignProject
}

func newDesignPhase(proj *DesignProject) *DesignPhase {
    return &DesignPhase{
        state:     &proj.state.Phases.Design,
        artifacts: project.NewArtifactCollection(&proj.state.Phases.Design, proj.ctx),
        tasks:     project.NewTaskCollection(&proj.state.Phases.Design, proj, proj.ctx),
        project:   proj,
    }
}

// Advance not supported - design has no internal states
func (p *DesignPhase) Advance() (*domain.PhaseOperationResult, error) {
    return nil, project.ErrNotSupported
}

// Complete handles phase completion (Active → Finalization)
func (p *DesignPhase) Complete() (*domain.PhaseOperationResult, error) {
    // Validate all documents approved (guard will also check)
    if !allDocumentsApproved(p.state.Tasks) {
        unresolvedCount := countUnresolved(p.state.Tasks)
        return nil, fmt.Errorf(
            "cannot complete design: %d tasks are not yet completed or abandoned",
            unresolvedCount,
        )
    }

    // Update phase status
    p.state.Status = "completed"
    if err := p.project.Save(); err != nil {
        return nil, fmt.Errorf("failed to save state: %w", err)
    }

    // Return event to trigger phase transition
    return domain.WithEvent(EventCompleteDesign), nil
}

// AddArtifact enforces task existence guard
func (p *DesignPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
    if len(p.state.Tasks) == 0 {
        return fmt.Errorf(
            "cannot add artifacts before creating document tasks - "+
            "create at least one task to track document lifecycle first",
        )
    }
    return p.artifacts.Add(path, opts...)
}

// UpdateTaskStatus handles artifact auto-approval on completion
func (p *DesignPhase) UpdateTaskStatus(id string, status string) error {
    task, err := p.tasks.Get(id)
    if err != nil {
        return err
    }

    // Validate artifact exists if moving to completed
    if status == "completed" {
        if task.Metadata == nil {
            return fmt.Errorf(
                "task %s has no metadata - set artifact_path before completing",
                id,
            )
        }

        artifactPathRaw, ok := task.Metadata["artifact_path"]
        if !ok {
            return fmt.Errorf(
                "task %s has no artifact_path in metadata - "+
                "link artifact to task before completing",
                id,
            )
        }

        artifactPath, ok := artifactPathRaw.(string)
        if !ok {
            return fmt.Errorf("artifact_path must be a string")
        }

        // Find artifact
        var targetArtifact *phases.Artifact
        for i := range p.state.Artifacts {
            if p.state.Artifacts[i].Path == artifactPath {
                targetArtifact = &p.state.Artifacts[i]
                break
            }
        }

        if targetArtifact == nil {
            return fmt.Errorf(
                "artifact not found at %s - add artifact before completing task",
                artifactPath,
            )
        }

        // Auto-approve artifact
        approved := true
        targetArtifact.Approved = &approved
    }

    // Update task status via collection
    return p.tasks.UpdateStatus(id, status)
}

// AddInput adds an input artifact to phase.inputs
func (p *DesignPhase) AddInput(path string, opts ...domain.ArtifactOption) error {
    config := &domain.ArtifactConfig{}
    for _, opt := range opts {
        opt(config)
    }

    now := time.Now()
    input := phases.Artifact{
        Path:       path,
        Created_at: now,
        Metadata:   config.Metadata,
    }

    if config.Type != nil {
        input.Type = config.Type
    }

    if p.state.Inputs == nil {
        p.state.Inputs = []phases.Artifact{}
    }

    p.state.Inputs = append(p.state.Inputs, input)
    return p.project.Save()
}

// ... other Phase interface methods
```

### Finalization Phase

**File**: `cli/internal/project/design/finalization.go`

Standard finalization phase (similar to exploration).

```go
package design

import (
    "fmt"
    "github.com/jmgilman/sow/cli/internal/project"
    "github.com/jmgilman/sow/cli/internal/project/domain"
    "github.com/jmgilman/sow/cli/schemas/phases"
)

type FinalizationPhase struct {
    state   *phases.Phase
    tasks   *project.TaskCollection
    project *DesignProject
}

func newFinalizationPhase(proj *DesignProject) *FinalizationPhase {
    return &FinalizationPhase{
        state:   &proj.state.Phases.Finalization,
        tasks:   project.NewTaskCollection(&proj.state.Phases.Finalization, proj, proj.ctx),
        project: proj,
    }
}

// Advance not supported - finalization has no internal states
func (p *FinalizationPhase) Advance() (*domain.PhaseOperationResult, error) {
    return nil, project.ErrNotSupported
}

// Complete when all finalization tasks done
func (p *FinalizationPhase) Complete() (*domain.PhaseOperationResult, error) {
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

// ... other Phase interface methods
```

## Guards

**File**: `cli/internal/project/design/guards.go`

```go
package design

import "github.com/jmgilman/sow/cli/schemas/phases"

// AllDocumentsApproved checks if all documents are completed or abandoned.
// Guards Active → Finalizing transition.
// Requires at least one completed task.
func AllDocumentsApproved(tasks []phases.Task) bool {
    if len(tasks) == 0 {
        return false  // Must have at least one task
    }

    hasCompleted := false
    for _, task := range tasks {
        // Check for unresolved tasks
        if task.Status != "completed" && task.Status != "abandoned" {
            return false
        }
        // Track if we have at least one completed
        if task.Status == "completed" {
            hasCompleted = true
        }
    }

    return hasCompleted
}

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
func allDocumentsApproved(tasks []phases.Task) bool {
    return AllDocumentsApproved(tasks)
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
```

## States and Events

**File**: `cli/internal/project/design/states.go`

```go
package design

import "github.com/jmgilman/sow/cli/internal/project/statechart"

// State constants for design workflow
const (
    DesignActive     = statechart.State("Active")
    DesignFinalizing = statechart.State("Finalizing")
    DesignCompleted  = statechart.State("Completed")
)
```

**File**: `cli/internal/project/design/events.go`

```go
package design

import "github.com/jmgilman/sow/cli/internal/project/statechart"

// Event constants for design transitions
const (
    // Active → Finalizing
    EventCompleteDesign = statechart.Event("complete_design")

    // Finalizing → Completed
    EventCompleteFinalization = statechart.Event("complete_finalization")
)
```

## Prompt Generation

**File**: `cli/internal/project/design/prompts.go`

```go
package design

import (
    "fmt"
    "strings"

    "github.com/jmgilman/sow/cli/internal/project/statechart"
    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/jmgilman/sow/cli/schemas"
)

type DesignPromptGenerator struct {
    components *statechart.PromptComponents
    ctx        *sow.Context
}

func NewDesignPromptGenerator(ctx *sow.Context) *DesignPromptGenerator {
    return &DesignPromptGenerator{
        components: statechart.NewPromptComponents(ctx),
        ctx:        ctx,
    }
}

func (g *DesignPromptGenerator) GeneratePrompt(
    state statechart.State,
    projectState *schemas.ProjectState,
) (string, error) {
    switch state {
    case DesignActive:
        return g.generateActivePrompt(projectState)
    case DesignFinalizing:
        return g.generateFinalizingPrompt(projectState)
    default:
        return "", fmt.Errorf("unknown design state: %s", state)
    }
}

func (g *DesignPromptGenerator) generateActivePrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    // Project header
    buf.WriteString(g.components.ProjectHeader(projectState))
    buf.WriteString("\n")

    // Current state
    buf.WriteString("## Current State: Active Design\n\n")
    buf.WriteString("You are in the Active state of design. ")
    buf.WriteString("Your focus is planning design outputs, drafting documents, ")
    buf.WriteString("and iterating through reviews.\n\n")

    phase := projectState.Phases.Design

    // Inputs
    if phase.Inputs != nil && len(phase.Inputs) > 0 {
        buf.WriteString("## Design Inputs\n\n")
        buf.WriteString("Sources informing this design:\n\n")
        for _, input := range phase.Inputs {
            buf.WriteString(fmt.Sprintf("- %s\n", input.Path))
        }
        buf.WriteString("\n")
    }

    // Document tasks
    buf.WriteString("## Design Documents\n\n")

    if len(phase.Tasks) == 0 {
        buf.WriteString("No documents planned yet. Start by creating document tasks:\n\n")
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent task create \"Document name\" \\\n")
        buf.WriteString("  --metadata '{\"document_type\":\"design\",\"target_location\":\"...\"}'\n")
        buf.WriteString("```\n\n")
        buf.WriteString("**Important**: Create at least one task before adding artifacts.\n\n")
    } else {
        pending := 0
        inProgress := 0
        needsReview := 0
        completed := 0
        abandoned := 0

        for _, task := range phase.Tasks {
            switch task.Status {
            case "pending":
                pending++
            case "in_progress":
                inProgress++
            case "needs_review":
                needsReview++
            case "completed":
                completed++
            case "abandoned":
                abandoned++
            }
        }

        buf.WriteString(fmt.Sprintf("Total: %d documents\n", len(phase.Tasks)))
        buf.WriteString(fmt.Sprintf("- Pending: %d\n", pending))
        buf.WriteString(fmt.Sprintf("- In Progress: %d\n", inProgress))
        buf.WriteString(fmt.Sprintf("- Needs Review: %d\n", needsReview))
        buf.WriteString(fmt.Sprintf("- Completed: %d\n", completed))
        buf.WriteString(fmt.Sprintf("- Abandoned: %d\n", abandoned))
        buf.WriteString("\n")

        // List documents
        buf.WriteString("### Documents:\n\n")
        for _, task := range phase.Tasks {
            statusIcon := "[ ]"
            switch task.Status {
            case "completed":
                statusIcon = "[✓]"
            case "abandoned":
                statusIcon = "[✗]"
            case "needs_review":
                statusIcon = "[?]"
            case "in_progress":
                statusIcon = "[~]"
            }

            buf.WriteString(fmt.Sprintf("%s %s - %s (%s)\n", statusIcon, task.Id, task.Name, task.Status))

            // Show artifact if linked
            if task.Metadata != nil {
                if artifactPath, ok := task.Metadata["artifact_path"].(string); ok {
                    buf.WriteString(fmt.Sprintf("    Artifact: %s\n", artifactPath))
                }
            }
        }
        buf.WriteString("\n")
    }

    // Next steps
    buf.WriteString("## Next Steps\n\n")
    if allDocumentsApproved(phase.Tasks) && len(phase.Tasks) > 0 {
        buf.WriteString("✓ All documents approved!\n\n")
        buf.WriteString("Ready to finalize. Run:\n\n")
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent complete\n")
        buf.WriteString("```\n\n")
    } else {
        buf.WriteString("Continue design work:\n")
        buf.WriteString("- Plan documents: `sow agent task create \"Document name\"`\n")
        buf.WriteString("- Add inputs: `sow agent artifact add <path> --input`\n")
        buf.WriteString("- Draft documents: Create document, then `sow agent artifact add <path>`\n")
        buf.WriteString("- Link to task: `sow agent task update <id> --metadata '{\"artifact_path\":\"...\"}'`\n")
        buf.WriteString("- Mark for review: `sow agent task update <id> --status needs_review`\n")
        buf.WriteString("- Approve: `sow agent task update <id> --status completed`\n\n")

        unresolvedCount := countUnresolved(phase.Tasks)
        if unresolvedCount > 0 {
            buf.WriteString(fmt.Sprintf("(%d documents remaining)\n\n", unresolvedCount))
        }
    }

    return buf.String(), nil
}

func (g *DesignPromptGenerator) generateFinalizingPrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    buf.WriteString(g.components.ProjectHeader(projectState))
    buf.WriteString("\n")

    buf.WriteString("## Current State: Finalizing\n\n")
    buf.WriteString("All documents approved. Finalizing design by moving artifacts, ")
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

1. **Initialize design** (via wizard on `design/*` branch)

2. **Register inputs** (optional):
   ```bash
   sow agent artifact add ../../knowledge/explorations/auth-research.md --input
   sow agent artifact add https://oauth.net/2/ --input
   ```

3. **Plan outputs as tasks**:
   ```bash
   sow agent task create "Authentication design doc" \
     --metadata '{"document_type":"design","target_location":".sow/knowledge/designs/auth-design.md"}'

   sow agent task create "ADR: OAuth vs JWT" \
     --metadata '{"document_type":"adr","target_location":".sow/knowledge/adrs/003-auth-approach.md"}'
   ```

4. **Draft first document**:
   ```bash
   # Mark task in progress
   sow agent task update 001 --status in_progress

   # Orchestrator drafts document...

   # Add as artifact
   sow agent artifact add project/auth-design.md

   # Link to task
   sow agent task update 001 --metadata '{"artifact_path":"project/auth-design.md",...}'
   ```

5. **Mark ready for review**:
   ```bash
   sow agent task update 001 --status needs_review
   ```

6. **Review iteration**:
   - User reviews document
   - If changes needed:
     ```bash
     sow agent task update 001 --status in_progress
     # Orchestrator updates document...
     sow agent task update 001 --status needs_review
     ```
   - Repeat until satisfied

7. **Approve document**:
   ```bash
   sow agent task update 001 --status completed
   # Artifact automatically approved
   ```

8. **Repeat for remaining documents**

9. **Complete design**:
   ```bash
   sow agent complete  # Active → Finalizing
   ```

### Finalizing State Workflow

1. **Transition to Finalizing**: `sow agent complete` from Active

2. **Create finalization tasks** (automatic):
   - For each completed task with `target_location`:
     - Move artifact to target location
   - Create PR with links to design artifacts
   - Delete `.sow/project/` directory

3. **Execute tasks**:
   - Orchestrator completes each task sequentially
   - Updates task status to `completed`

4. **Complete finalization**: `sow agent complete`

5. **Design finished**: State machine reaches Completed

## CLI Commands

### Add Input

**New command** (or extend existing artifact add):

```bash
# Option 1: New dedicated command
sow agent input add <path> [--description "..."] [--type "..."]

# Option 2: Flag on existing command (preferred)
sow agent artifact add <path> --input [--description "..."]
```

**Implementation**: Calls `DesignPhase.AddInput()` which adds to `phase.inputs`

### Link Artifact to Task

**Via metadata update**:

```bash
sow agent task update 001 --metadata '{"artifact_path":"project/auth-design.md"}'
```

**Or combined with other metadata**:

```bash
sow agent task update 001 \
  --metadata '{
    "artifact_path":"project/auth-design.md",
    "document_type":"design",
    "target_location":".sow/knowledge/designs/auth-design.md"
  }'
```

## Testing Strategy

### Unit Tests

**Schema validation**:
- DesignProjectState validates correctly
- Phase status constraints enforced
- State machine state constraints enforced

**Guards**:
- `AllDocumentsApproved` returns false with pending tasks
- `AllDocumentsApproved` returns true when all completed/abandoned
- `AllDocumentsApproved` requires at least one completed
- `AllFinalizationTasksComplete` validates all tasks completed

**Phase operations**:
- `AddArtifact()` fails if no tasks exist
- `UpdateTaskStatus()` to completed validates artifact exists
- `UpdateTaskStatus()` to completed auto-approves artifact
- `Complete()` validates all documents approved

### Integration Tests

**Full workflow**:
1. Create design project on `design/test` branch
2. Add 2 inputs
3. Create 2 document tasks
4. Draft and approve both documents
5. Complete design
6. Complete finalization
7. Verify documents moved to target locations

**Edge cases**:
- Add artifact before creating tasks (should fail)
- Complete task without artifact_path (should fail)
- Complete task with non-existent artifact (should fail)
- Complete task successfully (should auto-approve artifact)
- needs_review → in_progress transition (should work)
- Complete with all tasks abandoned (should fail - need at least one completed)

### Manual Verification

- Real design workflow on test topic
- Verify prompts provide useful guidance
- Verify state transitions feel natural
- Verify artifact auto-approval works
- Verify finalization moves documents correctly

## Migration Notes

No migration from old design mode - users restart active design sessions.

Design sessions are typically short (days to weeks), so breaking change acceptable.

## Success Criteria

1. ✅ Can create design project on `design/*` branch
2. ✅ Can register inputs to track design sources
3. ✅ Cannot add artifacts before creating tasks
4. ✅ Can create document tasks with metadata
5. ✅ Can link artifacts to tasks via metadata
6. ✅ Can transition tasks through pending → in_progress → needs_review → completed
7. ✅ Can go backward from needs_review → in_progress for revisions
8. ✅ Completing task auto-approves artifact at metadata.artifact_path
9. ✅ Cannot complete task without artifact existing
10. ✅ Can complete design when all documents approved (at least one completed)
11. ✅ Finalization moves documents to target locations
12. ✅ Zero-context resumability works
13. ✅ Prompts provide clear guidance at each state

## Future Enhancements

**Not in scope for initial implementation**:

- Template validation (verify document matches template structure)
- Automatic template loading based on `task.metadata.template`
- Multi-reviewer approval workflow
- Document versioning/iteration tracking
- Linking designs to implementations (track which projects implement this design)
- Design impact analysis (which systems affected by this design)
