# Breakdown Project Type Design

**Author**: Architecture Team
**Date**: 2025-10-31
**Status**: Proposed
**Depends On**: [Core Design](core-design.md)

## Overview

This document specifies the breakdown project type - a workflow for decomposing large features or design docs into implementable work units that become GitHub issues. Breakdown projects help users systematically break down complex work, specify each unit, and publish them as trackable GitHub issues.

**Key Characteristics**:
- **Input tracking**: Register design docs or features being broken down
- **Task-based work unit tracking**: Each task represents a work unit that becomes a GitHub issue
- **Dependency management**: Tasks can depend on other tasks (determines issue creation order)
- **Review workflow**: Draft spec → needs_review → completed (with auto-approval)
- **Publishing checkpoint**: Separate state for creating GitHub issues after full review
- **No finalization phase**: Deliverables are GitHub issues, not committed files

## Goals and Non-Goals

**Goals**:
- Decompose design docs or large features into implementable work units
- Specify each work unit with sufficient detail
- Review all work units before publishing to GitHub
- Publish work units as GitHub issues in dependency order
- Track issue metadata (number, URL) for reference
- Support work unit dependencies

**Non-Goals**:
- Implementation tracking (that's standard project's job)
- Finalization/PR creation (no artifacts to commit)
- Issue lifecycle management (once published, GitHub owns lifecycle)
- Multi-repository breakdown (single repo only)

## Project Schema

### CUE Schema Definition

**File**: `cli/schemas/projects/breakdown.cue`

```cue
package projects

import (
    "time"
    p "github.com/jmgilman/sow/cli/schemas/phases"
)

// BreakdownProjectState defines schema for breakdown project type.
//
// Breakdown follows a work decomposition workflow:
// Active (decompose, specify, review) → Publishing (create issues) → Completed
#BreakdownProjectState: {
    // Discriminator
    project: {
        type: "breakdown"  // Fixed discriminator value

        // Kebab-case project identifier
        name: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"

        // Git branch (typically breakdown/* prefix)
        branch: string & !=""

        // What is being broken down
        description: string

        // Timestamps
        created_at: time.Time
        updated_at: time.Time
    }

    // State machine position
    statechart: {
        current_state: "Active" | "Publishing" | "Completed"
    }

    // Single-phase structure (no finalization)
    phases: {
        // Phase 1: Breakdown (decompose, specify, review, publish)
        breakdown: p.#Phase & {
            // Custom status values for breakdown workflow
            status: "active" | "publishing" | "completed"
            enabled: true

            // Inputs track what's being broken down
            // (design docs, feature specs, etc.)
            inputs?: [...p.#Artifact]

            // Tasks represent work units that become GitHub issues
            // Task.metadata stores:
            //   - artifact_path: path to work unit spec
            //   - published: whether GitHub issue created
            //   - github_issue_number: issue number (set after publishing)
            //   - github_issue_url: issue URL (set after publishing)
            //   - work_unit_type: optional type (feature, bug, refactor, etc.)

            // Artifacts are work unit specifications
            // Auto-approved when task marked completed
        }
    }
}
```

### Go Type

**File**: `cli/schemas/projects/breakdown.go`

```go
package projects

import (
    "time"
    "github.com/jmgilman/sow/cli/schemas/phases"
)

// BreakdownProjectState is the Go representation of #BreakdownProjectState.
type BreakdownProjectState struct {
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
        Breakdown phases.Phase `json:"breakdown"`
    } `json:"phases"`
}
```

## State Machine

### States

**Active** (breakdown phase)
- **Purpose**: Decompose work, specify units, review all work units
- **Phase**: breakdown
- **Phase status**: `"active"`
- **Duration**: Most of breakdown time
- **Orchestrator focus**: Create work unit tasks, write specifications, iterate on reviews

**Publishing** (breakdown phase)
- **Purpose**: Create GitHub issues for approved work units in dependency order
- **Phase**: breakdown
- **Phase status**: `"publishing"`
- **Duration**: Short - automated issue creation
- **Orchestrator focus**: Create GitHub issues, handle dependencies, track metadata

**Completed** (breakdown phase)
- **Purpose**: Terminal state, all issues published
- **Phase**: breakdown
- **Phase status**: `"completed"`
- **Duration**: Permanent
- **Orchestrator focus**: None (project complete)

### State Transitions

```
Active
  │
  │ sow agent advance
  │ guard: AllWorkUnitsApproved
  ▼
Publishing
  │
  │ sow agent complete
  │ guard: AllWorkUnitsPublished
  │ (deletes .sow/project/)
  ▼
Completed
```

### Transition Details

#### Active → Publishing

**Trigger**: `sow agent advance`

**Event**: `EventBeginPublishing`

**Guard**: `AllWorkUnitsApproved(tasks)`
```go
// All tasks must be completed or abandoned
// At least one task must be completed
func AllWorkUnitsApproved(tasks []phases.Task) bool {
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

**Phase status change**: `breakdown.status = "active"` → `"publishing"`

**Orchestrator behavior change**:
- Before: Specify and review work units
- After: Create GitHub issues for completed work units

#### Publishing → Completed

**Trigger**: `sow agent complete`

**Event**: `EventCompleteBreakdown`

**Guard**: `AllWorkUnitsPublished(tasks)`
```go
// All completed tasks must have published = true in metadata
func AllWorkUnitsPublished(tasks []phases.Task) bool {
    for _, task := range tasks {
        // Only check completed tasks
        if task.Status == "completed" {
            if task.Metadata == nil {
                return false
            }

            publishedRaw, ok := task.Metadata["published"]
            if !ok {
                return false
            }

            published, ok := publishedRaw.(bool)
            if !ok || !published {
                return false
            }
        }
    }

    // Must have at least one completed task
    hasCompleted := false
    for _, task := range tasks {
        if task.Status == "completed" {
            hasCompleted = true
            break
        }
    }

    return hasCompleted
}
```

**Phase status change**: `breakdown.status = "completed"`

**Cleanup**: Delete `.sow/project/` directory

**State machine**: Terminal state reached

## Task Lifecycle

### Tasks Represent Work Units

Each task represents a work unit that will become a GitHub issue:
- **Task name**: Work unit title (becomes issue title)
- **Task description**: Brief summary (becomes issue body template)
- **Task status**: Current state of work unit specification
- **Task dependencies**: Other tasks this work unit depends on
- **Task metadata**: Artifact path, publishing status, GitHub issue info

### Task Status Flow

```
pending
  │
  │ Start specifying
  ▼
in_progress
  │
  │ Spec drafted, ready for review
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
  └─→ abandoned (work unit not needed)
```

**Standard statuses used**:
- `pending`: Work unit identified, not yet specified
- `in_progress`: Actively writing specification
- `needs_review`: Spec ready for human review
- `completed`: Spec approved, ready for publishing
- `abandoned`: Work unit not needed

### Task Metadata Structure

```yaml
metadata:
  artifact_path: "project/work-units/jwt-generation.md"
  published: false                    # Set to true after issue created
  github_issue_number: 123            # Set during publishing
  github_issue_url: "https://github.com/org/repo/issues/123"  # Set during publishing
  work_unit_type: "feature"           # Optional: feature, bug, refactor, spike
```

**When metadata is set**:
- `artifact_path`: Set when artifact added (during specification)
- `published`: Initially false, set to true during Publishing state
- `github_issue_number`: Set during Publishing state after issue created
- `github_issue_url`: Set during Publishing state after issue created
- `work_unit_type`: Optional, set when task created

### Task Dependencies

**Specifying dependencies**:
```bash
# Create dependent task
sow agent task create "Implement JWT validation" --dependencies "001"

# Or update existing task
sow agent task update 003 --dependencies "001,002"
```

**Publishing order**:
- Tasks with no dependencies published first
- Tasks with dependencies published only after their dependencies
- Uses topological sort to determine safe ordering
- Cyclic dependencies rejected during validation

### Task-to-Artifact Linking

**Same pattern as design project**:

1. Task in `needs_review` status
2. Human reviews spec, satisfied
3. Orchestrator runs: `sow agent task update <id> --status completed`
4. Breakdown phase detects completion, reads `task.metadata.artifact_path`
5. Finds artifact at that path, sets `approved = true`
6. Saves state

**Validation**: When transitioning to `completed`, artifact must exist at `metadata.artifact_path`

## Work Unit Specifications

### Specification Format

Each work unit has a specification artifact:

**Location**: `project/work-units/<work-unit-id>.md`

**Content**:
```markdown
# [Work Unit Title]

## Overview

[Brief description of what this work unit accomplishes]

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Technical Approach

[How this should be implemented]

## Dependencies

[List of other work units this depends on, if any]

## Testing Requirements

[What tests need to be added]

## Estimated Complexity

[T-shirt size: S/M/L/XL or points]
```

**Purpose**:
- Provides implementer with clear requirements
- Becomes basis for GitHub issue body
- Reviewed and approved before publishing
- Stored as artifact for traceability

### Artifact Operations

**Cannot add artifacts before creating tasks**:
```go
func (p *BreakdownPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
    if len(p.state.Tasks) == 0 {
        return fmt.Errorf(
            "cannot add artifacts before creating work unit tasks - "+
            "create at least one task first",
        )
    }
    return p.artifacts.Add(path, opts...)
}
```

**Same guard as design**: Enforces planning before drafting

## Publishing Workflow

### Publishing State Operations

**Goal**: Create GitHub issue for each completed work unit in dependency order

**Steps**:

1. **Validate dependencies**:
   - Check for cyclic dependencies
   - Ensure all dependencies point to valid tasks

2. **Topological sort**:
   - Determine safe publishing order
   - Tasks with no dependencies first
   - Then tasks whose dependencies are published

3. **For each task in order**:
   - Read work unit specification from artifact
   - Create GitHub issue:
     - Title: task name
     - Body: rendered specification
     - Labels: based on work_unit_type if present
   - Store issue metadata:
     - `metadata.github_issue_number`
     - `metadata.github_issue_url`
     - `metadata.published = true`
   - Save state after each issue created (resumability)

4. **All tasks published**:
   - Can complete Publishing state

### GitHub Issue Creation

**Issue title**: Task name

**Issue body template**:
```markdown
# Overview

[From spec]

## Acceptance Criteria

[From spec]

## Technical Approach

[From spec]

## Dependencies

[List of GitHub issue links for dependencies]

## Testing Requirements

[From spec]

---

*This issue was created from breakdown project: [project name]*
*Spec: [link to spec artifact in branch]*
```

**Issue labels**:
- `sow` (always)
- `work_unit_type` if specified (feature, bug, refactor, etc.)

**Dependencies**:
- If task has dependencies, include links to dependency issues in body
- GitHub doesn't have formal dependency tracking, so this is documentation only

### Error Handling

**If issue creation fails**:
- Log error with task ID and reason
- Do not mark task as published
- Allow retry by running `sow agent complete` again
- Already-published tasks skipped (check `metadata.published`)

**Resumability**:
- Publishing can be interrupted and resumed
- Each task tracks whether it's been published
- Retry is idempotent (won't create duplicate issues)

## Phase Implementation

### Breakdown Phase

**File**: `cli/internal/project/breakdown/breakdown.go`

```go
package breakdown

import (
    "fmt"
    "github.com/jmgilman/sow/cli/internal/project"
    "github.com/jmgilman/sow/cli/internal/project/domain"
    "github.com/jmgilman/sow/cli/schemas/phases"
)

type BreakdownPhase struct {
    state     *phases.Phase
    artifacts *project.ArtifactCollection
    tasks     *project.TaskCollection
    project   *BreakdownProject
}

func newBreakdownPhase(proj *BreakdownProject) *BreakdownPhase {
    return &BreakdownPhase{
        state:     &proj.state.Phases.Breakdown,
        artifacts: project.NewArtifactCollection(&proj.state.Phases.Breakdown, proj.ctx),
        tasks:     project.NewTaskCollection(&proj.state.Phases.Breakdown, proj, proj.ctx),
        project:   proj,
    }
}

// Advance handles Active → Publishing transition
func (p *BreakdownPhase) Advance() (*domain.PhaseOperationResult, error) {
    switch p.state.Status {
    case "active":
        // Validate all work units approved
        if !allWorkUnitsApproved(p.state.Tasks) {
            unresolvedCount := countUnresolved(p.state.Tasks)
            return nil, fmt.Errorf(
                "cannot advance to publishing: %d work units not yet completed or abandoned",
                unresolvedCount,
            )
        }

        // Validate dependencies
        if err := p.validateDependencies(); err != nil {
            return nil, fmt.Errorf("dependency validation failed: %w", err)
        }

        // Update phase status
        p.state.Status = "publishing"
        if err := p.project.Save(); err != nil {
            return nil, fmt.Errorf("failed to save state: %w", err)
        }

        // Return event to trigger state transition
        return domain.WithEvent(EventBeginPublishing), nil

    case "publishing":
        // Already in publishing state
        return nil, fmt.Errorf(
            "already in Publishing state - use 'sow agent complete' to finish breakdown",
        )

    default:
        return nil, fmt.Errorf("unknown breakdown status: %s", p.state.Status)
    }
}

// Complete handles Publishing → Completed transition
func (p *BreakdownPhase) Complete() (*domain.PhaseOperationResult, error) {
    // Can only complete from Publishing state
    if p.state.Status != "publishing" {
        return nil, fmt.Errorf(
            "cannot complete breakdown from %s state - advance to Publishing first",
            p.state.Status,
        )
    }

    // Validate all work units published (guard will also check)
    if !allWorkUnitsPublished(p.state.Tasks) {
        unpublishedCount := countUnpublished(p.state.Tasks)
        return nil, fmt.Errorf(
            "%d work unit(s) not yet published - publish all issues before completing",
            unpublishedCount,
        )
    }

    // Update phase status
    p.state.Status = "completed"
    if err := p.project.Save(); err != nil {
        return nil, fmt.Errorf("failed to save state: %w", err)
    }

    // Cleanup: delete project directory
    if err := p.project.ctx.FS().RemoveAll("project"); err != nil {
        // Log but don't fail - project is complete
        p.project.ctx.Logger().Warn("failed to cleanup project directory: %v", err)
    }

    // Return event to trigger state transition
    return domain.WithEvent(EventCompleteBreakdown), nil
}

// AddArtifact enforces task existence guard
func (p *BreakdownPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
    if len(p.state.Tasks) == 0 {
        return fmt.Errorf(
            "cannot add artifacts before creating work unit tasks - "+
            "create at least one task first",
        )
    }
    return p.artifacts.Add(path, opts...)
}

// UpdateTaskStatus handles artifact auto-approval on completion
func (p *BreakdownPhase) UpdateTaskStatus(id string, status string) error {
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

// validateDependencies checks for cycles and invalid references
func (p *BreakdownPhase) validateDependencies() error {
    // Build adjacency list
    graph := make(map[string][]string)
    taskIDs := make(map[string]bool)

    for _, task := range p.state.Tasks {
        if task.Status == "completed" {
            taskIDs[task.Id] = true
            if task.Dependencies != nil {
                graph[task.Id] = *task.Dependencies
            }
        }
    }

    // Check all dependencies point to valid tasks
    for taskID, deps := range graph {
        for _, depID := range deps {
            if !taskIDs[depID] {
                return fmt.Errorf(
                    "task %s depends on invalid task %s",
                    taskID, depID,
                )
            }
        }
    }

    // Check for cycles using DFS
    visited := make(map[string]bool)
    recStack := make(map[string]bool)

    var hasCycle func(string) bool
    hasCycle = func(taskID string) bool {
        visited[taskID] = true
        recStack[taskID] = true

        for _, depID := range graph[taskID] {
            if !visited[depID] {
                if hasCycle(depID) {
                    return true
                }
            } else if recStack[depID] {
                return true
            }
        }

        recStack[taskID] = false
        return false
    }

    for taskID := range graph {
        if !visited[taskID] {
            if hasCycle(taskID) {
                return fmt.Errorf("cyclic dependency detected in work unit tasks")
            }
        }
    }

    return nil
}

// AddInput adds an input artifact to phase.inputs
func (p *BreakdownPhase) AddInput(path string, opts ...domain.ArtifactOption) error {
    // Same implementation as design project
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

// PublishWorkUnit creates GitHub issue for a work unit
func (p *BreakdownPhase) PublishWorkUnit(taskID string) error {
    // Only allowed in Publishing state
    if p.state.Status != "publishing" {
        return fmt.Errorf("can only publish work units in Publishing state")
    }

    task, err := p.tasks.Get(taskID)
    if err != nil {
        return err
    }

    // Check if already published
    if task.Metadata != nil {
        if published, ok := task.Metadata["published"].(bool); ok && published {
            return fmt.Errorf("task %s already published", taskID)
        }
    }

    // Read work unit spec from artifact
    if task.Metadata == nil || task.Metadata["artifact_path"] == nil {
        return fmt.Errorf("task %s has no artifact_path", taskID)
    }

    artifactPath := task.Metadata["artifact_path"].(string)
    specContent, err := p.project.ctx.FS().ReadFile(artifactPath)
    if err != nil {
        return fmt.Errorf("failed to read work unit spec: %w", err)
    }

    // Create GitHub issue
    issue, err := p.project.ctx.GitHub().CreateIssue(
        task.Name,
        string(specContent),
        []string{"sow"},  // Labels
    )
    if err != nil {
        return fmt.Errorf("failed to create GitHub issue: %w", err)
    }

    // Update task metadata
    if task.Metadata == nil {
        task.Metadata = make(map[string]interface{})
    }
    task.Metadata["published"] = true
    task.Metadata["github_issue_number"] = issue.Number
    task.Metadata["github_issue_url"] = issue.URL

    return p.project.Save()
}

// ... other Phase interface methods
```

## Guards

**File**: `cli/internal/project/breakdown/guards.go`

```go
package breakdown

import "github.com/jmgilman/sow/cli/schemas/phases"

// AllWorkUnitsApproved checks if all work units are completed or abandoned.
// Guards Active → Publishing transition.
// Requires at least one completed task.
func AllWorkUnitsApproved(tasks []phases.Task) bool {
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

// AllWorkUnitsPublished checks if all completed work units have been published.
// Guards Publishing → Completed transition.
func AllWorkUnitsPublished(tasks []phases.Task) bool {
    hasCompleted := false

    for _, task := range tasks {
        // Only check completed tasks
        if task.Status == "completed" {
            hasCompleted = true

            if task.Metadata == nil {
                return false
            }

            publishedRaw, ok := task.Metadata["published"]
            if !ok {
                return false
            }

            published, ok := publishedRaw.(bool)
            if !ok || !published {
                return false
            }
        }
    }

    return hasCompleted
}

// Helper functions
func allWorkUnitsApproved(tasks []phases.Task) bool {
    return AllWorkUnitsApproved(tasks)
}

func allWorkUnitsPublished(tasks []phases.Task) bool {
    return AllWorkUnitsPublished(tasks)
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

func countUnpublished(tasks []phases.Task) int {
    count := 0
    for _, task := range tasks {
        if task.Status == "completed" {
            if task.Metadata == nil {
                count++
                continue
            }
            if published, ok := task.Metadata["published"].(bool); !ok || !published {
                count++
            }
        }
    }
    return count
}
```

## States and Events

**File**: `cli/internal/project/breakdown/states.go`

```go
package breakdown

import "github.com/jmgilman/sow/cli/internal/project/statechart"

// State constants for breakdown workflow
const (
    BreakdownActive     = statechart.State("Active")
    BreakdownPublishing = statechart.State("Publishing")
    BreakdownCompleted  = statechart.State("Completed")
)
```

**File**: `cli/internal/project/breakdown/events.go`

```go
package breakdown

import "github.com/jmgilman/sow/cli/internal/project/statechart"

// Event constants for breakdown transitions
const (
    // Active → Publishing
    EventBeginPublishing = statechart.Event("begin_publishing")

    // Publishing → Completed
    EventCompleteBreakdown = statechart.Event("complete_breakdown")
)
```

## Prompt Generation

**File**: `cli/internal/project/breakdown/prompts.go`

```go
package breakdown

import (
    "fmt"
    "strings"

    "github.com/jmgilman/sow/cli/internal/project/statechart"
    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/jmgilman/sow/cli/schemas"
)

type BreakdownPromptGenerator struct {
    components *statechart.PromptComponents
    ctx        *sow.Context
}

func NewBreakdownPromptGenerator(ctx *sow.Context) *BreakdownPromptGenerator {
    return &BreakdownPromptGenerator{
        components: statechart.NewPromptComponents(ctx),
        ctx:        ctx,
    }
}

func (g *BreakdownPromptGenerator) GeneratePrompt(
    state statechart.State,
    projectState *schemas.ProjectState,
) (string, error) {
    switch state {
    case BreakdownActive:
        return g.generateActivePrompt(projectState)
    case BreakdownPublishing:
        return g.generatePublishingPrompt(projectState)
    default:
        return "", fmt.Errorf("unknown breakdown state: %s", state)
    }
}

func (g *BreakdownPromptGenerator) generateActivePrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    // Project header
    buf.WriteString(g.components.ProjectHeader(projectState))
    buf.WriteString("\n")

    // Current state
    buf.WriteString("## Current State: Active Breakdown\n\n")
    buf.WriteString("You are in the Active state of breakdown. ")
    buf.WriteString("Your focus is decomposing work, specifying each unit, ")
    buf.WriteString("and reviewing all specifications before publishing.\n\n")

    phase := projectState.Phases.Breakdown

    // Inputs
    if phase.Inputs != nil && len(phase.Inputs) > 0 {
        buf.WriteString("## Being Broken Down\n\n")
        buf.WriteString("Sources being decomposed:\n\n")
        for _, input := range phase.Inputs {
            buf.WriteString(fmt.Sprintf("- %s\n", input.Path))
        }
        buf.WriteString("\n")
    }

    // Work units
    buf.WriteString("## Work Units\n\n")

    if len(phase.Tasks) == 0 {
        buf.WriteString("No work units identified yet. Start by creating work unit tasks:\n\n")
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent task create \"Work unit name\" \\\n")
        buf.WriteString("  --dependencies \"001,002\"  # Optional\n")
        buf.WriteString("```\n\n")
        buf.WriteString("**Important**: Create at least one task before adding specifications.\n\n")
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

        buf.WriteString(fmt.Sprintf("Total: %d work units\n", len(phase.Tasks)))
        buf.WriteString(fmt.Sprintf("- Pending: %d\n", pending))
        buf.WriteString(fmt.Sprintf("- In Progress: %d\n", inProgress))
        buf.WriteString(fmt.Sprintf("- Needs Review: %d\n", needsReview))
        buf.WriteString(fmt.Sprintf("- Completed: %d\n", completed))
        buf.WriteString(fmt.Sprintf("- Abandoned: %d\n", abandoned))
        buf.WriteString("\n")

        // List work units
        buf.WriteString("### Work Units:\n\n")
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

            // Show dependencies
            if task.Dependencies != nil && len(*task.Dependencies) > 0 {
                buf.WriteString(fmt.Sprintf("    Depends on: %v\n", *task.Dependencies))
            }

            // Show spec if linked
            if task.Metadata != nil {
                if artifactPath, ok := task.Metadata["artifact_path"].(string); ok {
                    buf.WriteString(fmt.Sprintf("    Spec: %s\n", artifactPath))
                }
            }
        }
        buf.WriteString("\n")
    }

    // Next steps
    buf.WriteString("## Next Steps\n\n")
    if allWorkUnitsApproved(phase.Tasks) && len(phase.Tasks) > 0 {
        buf.WriteString("✓ All work units approved!\n\n")
        buf.WriteString("Ready to publish GitHub issues. Run:\n\n")
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent advance\n")
        buf.WriteString("```\n\n")
    } else {
        buf.WriteString("Continue breakdown work:\n")
        buf.WriteString("- Identify work units: `sow agent task create \"Unit name\"`\n")
        buf.WriteString("- Add inputs: `sow agent artifact add <path> --input`\n")
        buf.WriteString("- Write specs: Create spec, then `sow agent artifact add <path>`\n")
        buf.WriteString("- Link to task: `sow agent task update <id> --metadata '{\"artifact_path\":\"...\"}'`\n")
        buf.WriteString("- Mark for review: `sow agent task update <id> --status needs_review`\n")
        buf.WriteString("- Approve: `sow agent task update <id> --status completed`\n\n")

        unresolvedCount := countUnresolved(phase.Tasks)
        if unresolvedCount > 0 {
            buf.WriteString(fmt.Sprintf("(%d work units remaining)\n\n", unresolvedCount))
        }
    }

    return buf.String(), nil
}

func (g *BreakdownPromptGenerator) generatePublishingPrompt(
    projectState *schemas.ProjectState,
) (string, error) {
    var buf strings.Builder

    buf.WriteString(g.components.ProjectHeader(projectState))
    buf.WriteString("\n")

    buf.WriteString("## Current State: Publishing\n\n")
    buf.WriteString("All work units approved. Creating GitHub issues for each work unit ")
    buf.WriteString("in dependency order.\n\n")

    phase := projectState.Phases.Breakdown

    // Publishing status
    buf.WriteString("## Publishing Status\n\n")

    completed := []phases.Task{}
    for _, task := range phase.Tasks {
        if task.Status == "completed" {
            completed = append(completed, task)
        }
    }

    published := 0
    unpublished := 0
    for _, task := range completed {
        if task.Metadata != nil {
            if pub, ok := task.Metadata["published"].(bool); ok && pub {
                published++
            } else {
                unpublished++
            }
        } else {
            unpublished++
        }
    }

    buf.WriteString(fmt.Sprintf("Total work units: %d\n", len(completed)))
    buf.WriteString(fmt.Sprintf("Published: %d\n", published))
    buf.WriteString(fmt.Sprintf("Unpublished: %d\n\n", unpublished))

    // List publishing status
    buf.WriteString("### Work Units:\n\n")
    for _, task := range completed {
        published := false
        var issueURL string

        if task.Metadata != nil {
            if pub, ok := task.Metadata["published"].(bool); ok && pub {
                published = true
            }
            if url, ok := task.Metadata["github_issue_url"].(string); ok {
                issueURL = url
            }
        }

        status := "[ ] Pending"
        if published {
            status = fmt.Sprintf("[✓] Published: %s", issueURL)
        }

        buf.WriteString(fmt.Sprintf("%s %s - %s\n", status, task.Id, task.Name))
    }
    buf.WriteString("\n")

    // Next steps
    buf.WriteString("## Next Steps\n\n")

    if allWorkUnitsPublished(phase.Tasks) {
        buf.WriteString("✓ All work units published!\n\n")
        buf.WriteString("Breakdown complete. Run:\n\n")
        buf.WriteString("```bash\n")
        buf.WriteString("sow agent complete\n")
        buf.WriteString("```\n\n")
    } else {
        buf.WriteString("Publish remaining work units:\n\n")
        buf.WriteString("For each unpublished work unit, the orchestrator should:\n")
        buf.WriteString("1. Verify dependencies are published (if any)\n")
        buf.WriteString("2. Create GitHub issue from specification\n")
        buf.WriteString("3. Update task metadata with issue number and URL\n")
        buf.WriteString("4. Mark as published\n\n")
        buf.WriteString(fmt.Sprintf("(%d work units remaining)\n\n", unpublished))
    }

    return buf.String(), nil
}
```

## Orchestrator Workflow

### Active State Workflow

1. **Initialize breakdown** (via wizard on `breakdown/*` branch)

2. **Register inputs** (what's being broken down):
   ```bash
   sow agent artifact add ../../knowledge/designs/auth-architecture.md --input
   ```

3. **Identify work units as tasks**:
   ```bash
   sow agent task create "Implement JWT token generation"
   sow agent task create "Implement JWT token validation" --dependencies "001"
   sow agent task create "Add JWT middleware to API routes" --dependencies "002"
   ```

4. **Specify first work unit**:
   ```bash
   # Mark task in progress
   sow agent task update 001 --status in_progress

   # Orchestrator writes specification...
   # (to project/work-units/001-jwt-generation.md)

   # Add as artifact
   sow agent artifact add project/work-units/001-jwt-generation.md

   # Link to task
   sow agent task update 001 --metadata '{"artifact_path":"project/work-units/001-jwt-generation.md"}'

   # Mark ready for review
   sow agent task update 001 --status needs_review
   ```

5. **Review iteration**:
   - User reviews specification
   - If changes needed:
     ```bash
     sow agent task update 001 --status in_progress
     # Orchestrator updates spec...
     sow agent task update 001 --status needs_review
     ```
   - Repeat until satisfied

6. **Approve specification**:
   ```bash
   sow agent task update 001 --status completed
   # Artifact automatically approved
   ```

7. **Repeat for remaining work units**

8. **Advance to Publishing**:
   ```bash
   sow agent advance  # Active → Publishing
   # Validates all work units approved
   # Validates no cyclic dependencies
   ```

### Publishing State Workflow

1. **Transition to Publishing**: `sow agent advance` from Active

2. **Publish work units** (orchestrator automates this):
   - Determine publishing order via topological sort
   - For each work unit in order:
     - Read specification from artifact
     - Create GitHub issue:
       - Title: task name
       - Body: specification content
       - Labels: `sow`, work_unit_type if set
     - Store issue metadata in task
     - Mark `published = true`
     - Save state (resumability)

3. **All units published**:
   ```bash
   sow agent complete  # Publishing → Completed
   # Validates all completed tasks published
   # Deletes .sow/project/ directory
   ```

4. **Breakdown finished**: State machine reaches Completed

## Testing Strategy

### Unit Tests

**Schema validation**:
- BreakdownProjectState validates correctly
- Phase status constraints enforced
- State machine state constraints enforced

**Guards**:
- `AllWorkUnitsApproved` returns false with pending tasks
- `AllWorkUnitsApproved` returns true when all completed/abandoned
- `AllWorkUnitsApproved` requires at least one completed
- `AllWorkUnitsPublished` returns false if any completed task not published
- `AllWorkUnitsPublished` validates all completed tasks published

**Dependency validation**:
- `validateDependencies` detects cyclic dependencies
- `validateDependencies` detects invalid task references
- `validateDependencies` passes for valid DAG

**Phase operations**:
- `AddArtifact()` fails if no tasks exist
- `UpdateTaskStatus()` to completed validates artifact exists
- `UpdateTaskStatus()` to completed auto-approves artifact
- `Advance()` validates all work units approved
- `Advance()` validates dependencies
- `Complete()` validates all work units published

### Integration Tests

**Full workflow**:
1. Create breakdown project on `breakdown/test` branch
2. Add 1 input
3. Create 3 work unit tasks with dependencies
4. Specify and approve all work units
5. Advance to Publishing
6. Publish all work units (mock GitHub API)
7. Complete breakdown
8. Verify project directory deleted

**Edge cases**:
- Add artifact before creating tasks (should fail)
- Complete task without artifact_path (should fail)
- Complete task with non-existent artifact (should fail)
- Advance with unresolved tasks (should fail)
- Advance with cyclic dependencies (should fail)
- Advance with invalid dependency reference (should fail)
- Complete Publishing with unpublished tasks (should fail)
- needs_review → in_progress transition (should work)

### Manual Verification

- Real breakdown workflow on test design doc
- Verify prompts provide useful guidance
- Verify dependency ordering works correctly
- Verify GitHub issue creation works
- Verify publishing resumability (interrupt and resume)

## Migration Notes

No migration from old breakdown mode - users restart active breakdown sessions.

Breakdown sessions are typically short (days to weeks), so breaking change acceptable.

## Success Criteria

1. ✅ Can create breakdown project on `breakdown/*` branch
2. ✅ Can register inputs to track what's being broken down
3. ✅ Cannot add artifacts before creating tasks
4. ✅ Can create work unit tasks with dependencies
5. ✅ Can link specifications to tasks via metadata
6. ✅ Can transition tasks through pending → in_progress → needs_review → completed
7. ✅ Can go backward from needs_review → in_progress for revisions
8. ✅ Completing task auto-approves specification artifact
9. ✅ Cannot complete task without artifact existing
10. ✅ Cannot advance to Publishing with unresolved tasks
11. ✅ Cannot advance to Publishing with cyclic dependencies
12. ✅ Publishing creates GitHub issues in dependency order
13. ✅ Publishing tracks issue metadata in tasks
14. ✅ Publishing is resumable (can interrupt and retry)
15. ✅ Cannot complete Publishing with unpublished work units
16. ✅ Completing breakdown deletes project directory (no finalization phase)
17. ✅ Zero-context resumability works
18. ✅ Prompts provide clear guidance at each state

## Future Enhancements

**Not in scope for initial implementation**:

- Multi-repository breakdown (publish issues to different repos)
- GitHub project board integration (add issues to board automatically)
- Story point estimation workflow
- Work unit templates (feature, bug, spike templates)
- Linking breakdown to parent epic/initiative
- Effort estimation aggregation
- Issue assignment during publishing
- Milestone integration
