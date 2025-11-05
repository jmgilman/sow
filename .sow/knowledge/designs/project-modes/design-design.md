# Design Project Type Design

**Author**: Architecture Team
**Date**: 2025-11-04
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
- **SDK-based implementation**: Built using Project SDK builder pattern

## Goals and Non-Goals

**Goals**:
- Track design inputs that inform decisions
- Plan design outputs as tasks before drafting
- Support iterative review workflow (needs_review status)
- Auto-approve artifacts when tasks completed
- Move approved documents to appropriate knowledge locations
- Prevent creating artifacts before planning tasks
- Use Project SDK for consistent, declarative configuration

**Non-Goals**:
- GitHub issue integration (design is pre-implementation planning)
- Multi-document ToC requirement (each document stands alone)
- Template enforcement (templates are guidance, not validation)
- Complex approval chains (single review cycle sufficient)

## Project SDK Configuration

The design project type uses the Project SDK builder pattern. All configuration is declarative and registered via the global registry.

### Package Structure

```
cli/internal/projects/design/
├── design.go          # SDK configuration using builder pattern
├── states.go          # State constants
├── events.go          # Event constants
├── guards.go          # Guard helper functions
├── prompts.go         # Prompt generator functions
├── metadata.go        # Embedded CUE metadata schemas
├── cue/
│   └── design_metadata.cue  # Phase metadata validation
└── templates/
    ├── orchestrator.md      # Orchestrator-level guidance
    ├── active.md            # Active state prompt template
    └── finalizing.md        # Finalizing state prompt template
```

### Registration

**File**: `cli/internal/projects/design/design.go`

```go
package design

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// init registers the design project type on package load.
func init() {
	state.Register("design", NewDesignProjectConfig())
}

// NewDesignProjectConfig creates the complete configuration for design project type.
func NewDesignProjectConfig() *project.ProjectTypeConfig {
	builder := project.NewProjectTypeConfigBuilder("design")
	builder = configurePhases(builder)
	builder = configureTransitions(builder)
	builder = configureEventDeterminers(builder)
	builder = configurePrompts(builder)
	builder = builder.WithInitializer(initializeDesignProject)
	return builder.Build()
}
```

### Phase Configuration

Design has a 2-phase structure:
1. **design** - Plan outputs, draft documents, iterate through reviews
2. **finalization** - Move documents, create PR, cleanup

```go
func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("design",
			project.WithStartState(sdkstate.State(Active)),
			project.WithEndState(sdkstate.State(Active)),
			project.WithOutputs("design", "adr", "architecture", "diagram", "spec"),  // Allowed artifact types
			project.WithTasks(),  // Phase supports tasks
			project.WithMetadataSchema(designMetadataSchema),
		).
		WithPhase("finalization",
			project.WithStartState(sdkstate.State(Finalizing)),
			project.WithEndState(sdkstate.State(Finalizing)),
			project.WithOutputs("pr"),
			project.WithMetadataSchema(finalizationMetadataSchema),
		)
}
```

**Key points**:
- `design` phase has ONE state (Active) - no intra-phase transitions
- Both phases support metadata validation via embedded CUE schemas
- Only `design` phase supports tasks (document planning)
- Artifact types constrained to document types

### Phase Initialization

```go
// initializeDesignProject creates all phases for a new design project.
// This is called during project creation to set up the phase structure.
func initializeDesignProject(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
	now := p.Created_at
	phaseNames := []string{"design", "finalization"}

	for _, phaseName := range phaseNames {
		// Get initial inputs for this phase (empty slice if none provided)
		inputs := []projschema.ArtifactState{}
		if initialInputs != nil {
			if phaseInputs, exists := initialInputs[phaseName]; exists {
				inputs = phaseInputs
			}
		}

		// Determine initial status based on phase
		status := "pending"
		enabled := false
		if phaseName == "design" {
			status = "active"  // Design starts immediately in active state
			enabled = true
		}

		p.Phases[phaseName] = projschema.PhaseState{
			Status:     status,
			Enabled:    enabled,
			Created_at: now,
			Inputs:     inputs,
			Outputs:    []projschema.ArtifactState{},
			Tasks:      []projschema.TaskState{},
			Metadata:   make(map[string]interface{}),
		}
	}

	return nil
}
```

## State Machine

### States

The design workflow progresses through three states across two phases:

**Active** (design phase, status="active")
- **Purpose**: Plan outputs, draft documents, iterate through reviews
- **Duration**: Majority of design time
- **Operations**: Create document tasks, draft documents, mark for review, approve
- **Complete condition**: All tasks completed or abandoned (at least one completed)

**Finalizing** (finalization phase, status="in_progress")
- **Purpose**: Move approved documents to targets, create PR, cleanup
- **Duration**: Short - automated finalization
- **Operations**: Execute finalization checklist
- **Complete condition**: All finalization tasks completed

**Completed** (terminal state)
- **Purpose**: Design finished
- **Duration**: Permanent
- **Operations**: None (project complete)

### State Transitions

```
Active (design.active)
  │
  │ EventCompleteDesign
  │ Guard: allDocumentsApproved(p)
  ▼
Finalizing (finalization.in_progress)
  │
  │ EventCompleteFinalization
  │ Guard: allFinalizationTasksComplete(p)
  ▼
Completed (terminal)
```

### Transition Configuration

```go
func configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		// Initial state
		SetInitialState(sdkstate.State(Active)).

		// Active → Finalizing
		AddTransition(
			sdkstate.State(Active),
			sdkstate.State(Finalizing),
			sdkstate.Event(EventCompleteDesign),
			project.WithGuard(func(p *state.Project) bool {
				return allDocumentsApproved(p)
			}),
			project.WithOnExit(func(p *state.Project) error {
				// Mark design phase as completed
				phase := p.Phases["design"]
				phase.Status = "completed"
				phase.Completed_at = time.Now()
				p.Phases["design"] = phase
				return nil
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Enable and activate finalization phase
				phase := p.Phases["finalization"]
				phase.Enabled = true
				phase.Status = "in_progress"
				phase.Started_at = time.Now()
				p.Phases["finalization"] = phase
				return nil
			}),
		).

		// Finalizing → Completed
		AddTransition(
			sdkstate.State(Finalizing),
			sdkstate.State(Completed),
			sdkstate.Event(EventCompleteFinalization),
			project.WithGuard(func(p *state.Project) bool {
				return allFinalizationTasksComplete(p)
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Mark finalization phase as completed
				phase := p.Phases["finalization"]
				phase.Status = "completed"
				phase.Completed_at = time.Now()
				p.Phases["finalization"] = phase
				return nil
			}),
		)
}
```

**Key patterns**:
- Guards are closures that capture project state via SDK binding
- OnEntry/OnExit actions update phase status and timestamps
- Only inter-phase transitions (no intra-phase transitions in design)

### Event Determiners

```go
func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(sdkstate.State(Active), func(p *state.Project) (sdkstate.Event, error) {
			// Active always advances to Finalizing
			return sdkstate.Event(EventCompleteDesign), nil
		}).
		OnAdvance(sdkstate.State(Finalizing), func(p *state.Project) (sdkstate.Event, error) {
			// Complete finalization phase
			return sdkstate.Event(EventCompleteFinalization), nil
		})
}
```

## States and Events

**File**: `cli/internal/projects/design/states.go`

```go
package design

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Design project states
const (
	// Active indicates active design phase
	Active = state.State("Active")

	// Finalizing indicates finalization in progress
	Finalizing = state.State("Finalizing")

	// Completed indicates design finished
	Completed = state.State("Completed")
)
```

**File**: `cli/internal/projects/design/events.go`

```go
package design

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Design project events trigger state transitions
const (
	// EventCompleteDesign transitions from Active to Finalizing
	// Fired when all documents are approved
	EventCompleteDesign = state.Event("complete_design")

	// EventCompleteFinalization transitions from Finalizing to Completed
	// Fired when all finalization tasks are completed
	EventCompleteFinalization = state.Event("complete_finalization")
)
```

## Guards

Guards are pure functions that check transition conditions. They operate on `*state.Project` and return boolean results.

**File**: `cli/internal/projects/design/guards.go`

```go
package design

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// allDocumentsApproved checks if all document tasks are completed or abandoned.
// Guards Active → Finalizing transition.
// Returns false if no tasks exist, or if at least one completed task doesn't exist.
func allDocumentsApproved(p *state.Project) bool {
	phase, exists := p.Phases["design"]
	if !exists {
		return false
	}

	tasks := phase.Tasks
	if len(tasks) == 0 {
		return false // Must have at least one task
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

	// Must have at least one completed task (can't complete with all abandoned)
	return hasCompleted
}

// allFinalizationTasksComplete checks if all finalization tasks are completed.
// Guards Finalizing → Completed transition.
func allFinalizationTasksComplete(p *state.Project) bool {
	phase, exists := p.Phases["finalization"]
	if !exists {
		return false
	}

	tasks := phase.Tasks
	if len(tasks) == 0 {
		return false // Must have finalization tasks
	}

	for _, task := range tasks {
		if task.Status != "completed" {
			return false
		}
	}
	return true
}

// Helper functions for counting (used in prompts and error messages)

func countUnresolvedTasks(p *state.Project) int {
	phase, exists := p.Phases["design"]
	if !exists {
		return 0
	}

	count := 0
	for _, task := range phase.Tasks {
		if task.Status != "completed" && task.Status != "abandoned" {
			count++
		}
	}
	return count
}

// validateTaskForCompletion checks if a task can be marked as completed.
// Returns error if artifact doesn't exist at task.metadata.artifact_path.
func validateTaskForCompletion(p *state.Project, taskID string) error {
	phase, exists := p.Phases["design"]
	if !exists {
		return fmt.Errorf("design phase not found")
	}

	// Find task
	var task *projschema.TaskState
	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			task = &phase.Tasks[i]
			break
		}
	}

	if task == nil {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Check metadata exists
	if task.Metadata == nil {
		return fmt.Errorf("task %s has no metadata - set artifact_path before completing", taskID)
	}

	// Check artifact_path exists
	artifactPathRaw, ok := task.Metadata["artifact_path"]
	if !ok {
		return fmt.Errorf("task %s has no artifact_path in metadata - link artifact to task before completing", taskID)
	}

	artifactPath, ok := artifactPathRaw.(string)
	if !ok {
		return fmt.Errorf("artifact_path must be a string")
	}

	// Check artifact exists
	var foundArtifact *projschema.ArtifactState
	for i := range phase.Outputs {
		if phase.Outputs[i].Path == artifactPath {
			foundArtifact = &phase.Outputs[i]
			break
		}
	}

	if foundArtifact == nil {
		return fmt.Errorf("artifact not found at %s - add artifact before completing task", artifactPath)
	}

	return nil
}

// autoApproveArtifact approves the artifact linked to a task when task is completed.
// This is called during task status update to "completed".
func autoApproveArtifact(p *state.Project, taskID string) error {
	phase, exists := p.Phases["design"]
	if !exists {
		return fmt.Errorf("design phase not found")
	}

	// Find task
	var task *projschema.TaskState
	for i := range phase.Tasks {
		if phase.Tasks[i].Id == taskID {
			task = &phase.Tasks[i]
			break
		}
	}

	if task == nil {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Get artifact path from metadata
	artifactPath, ok := task.Metadata["artifact_path"].(string)
	if !ok {
		return fmt.Errorf("task %s has invalid artifact_path", taskID)
	}

	// Find and approve artifact
	phaseData := p.Phases["design"]
	for i := range phaseData.Outputs {
		if phaseData.Outputs[i].Path == artifactPath {
			phaseData.Outputs[i].Approved = true
			p.Phases["design"] = phaseData
			return nil
		}
	}

	return fmt.Errorf("artifact not found at %s", artifactPath)
}
```

**Key patterns**:
- Guards operate on full `*state.Project`, accessing phases directly
- Pure functions with no side effects (except helpers that modify state)
- Used in transition configurations via closures
- Similar to standard project guards

## Prompt Generation

Prompts provide contextual guidance for each state. The SDK integrates prompts via the `WithPrompt()` builder method.

**File**: `cli/internal/projects/design/prompts.go`

```go
package design

import (
	"embed"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sdks/project/templates"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

//go:embed templates/*.md
var templatesFS embed.FS

func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		// Orchestrator-level prompt (how design projects work)
		WithOrchestratorPrompt(generateOrchestratorPrompt).

		// State-level prompts (what to do in each state)
		WithPrompt(sdkstate.State(Active), generateActivePrompt).
		WithPrompt(sdkstate.State(Finalizing), generateFinalizingPrompt)
}

// generateOrchestratorPrompt generates the orchestrator-level prompt for design projects.
func generateOrchestratorPrompt(p *state.Project) string {
	prompt, err := templates.Render(templatesFS, "templates/orchestrator.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering orchestrator prompt: %v", err)
	}
	return prompt
}

// generateActivePrompt generates the prompt for the Active state.
// Focus: Plan outputs, draft documents, iterate through reviews.
func generateActivePrompt(p *state.Project) string {
	var buf strings.Builder

	// Project header
	buf.WriteString(fmt.Sprintf("# Design: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Current state
	buf.WriteString("## Current State: Active Design\n\n")

	// Design phase info
	phase, exists := p.Phases["design"]
	if !exists {
		return "Error: design phase not found"
	}

	// Show inputs if any
	if len(phase.Inputs) > 0 {
		buf.WriteString("### Design Inputs\n\n")
		buf.WriteString("Sources informing this design:\n\n")
		for _, input := range phase.Inputs {
			buf.WriteString(fmt.Sprintf("- %s\n", input.Path))
			if input.Metadata != nil {
				if desc, ok := input.Metadata["description"].(string); ok {
					buf.WriteString(fmt.Sprintf("  %s\n", desc))
				}
			}
		}
		buf.WriteString("\n")
	}

	// Document tasks
	buf.WriteString("### Design Documents\n\n")

	if len(phase.Tasks) == 0 {
		buf.WriteString("No documents planned yet.\n\n")
		buf.WriteString("**Important**: Create at least one document task before adding artifacts.\n\n")
		buf.WriteString("**Next steps**: Plan document tasks\n\n")
	} else {
		// Count task statuses
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
		buf.WriteString(fmt.Sprintf("- Abandoned: %d\n\n", abandoned))

		// List documents
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
				if docType, ok := task.Metadata["document_type"].(string); ok {
					buf.WriteString(fmt.Sprintf("    Type: %s\n", docType))
				}
			}
		}
		buf.WriteString("\n")

		// Advancement readiness
		if allDocumentsApproved(p) {
			buf.WriteString("✓ All documents approved!\n\n")
			buf.WriteString("Ready to finalize. Run: `sow project advance`\n\n")
		} else {
			unresolvedCount := countUnresolvedTasks(p)
			buf.WriteString(fmt.Sprintf("**Next steps**: Continue design work (%d documents remaining)\n\n", unresolvedCount))
		}
	}

	// Render additional guidance from template
	guidance, err := templates.Render(templatesFS, "templates/active.md", p)
	if err != nil {
		return buf.String() + fmt.Sprintf("\nError rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// generateFinalizingPrompt generates the prompt for the Finalizing state.
// Focus: Move documents to targets, create PR, cleanup.
func generateFinalizingPrompt(p *state.Project) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Design: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n\n", p.Branch))

	buf.WriteString("## Current State: Finalizing\n\n")
	buf.WriteString("All documents approved. Finalizing design by moving artifacts, creating PR, and cleaning up.\n\n")

	// Finalization tasks
	phase, exists := p.Phases["finalization"]
	if !exists {
		return "Error: finalization phase not found"
	}

	buf.WriteString("### Finalization Tasks\n\n")
	for _, task := range phase.Tasks {
		status := "[ ]"
		if task.Status == "completed" {
			status = "[✓]"
		}
		buf.WriteString(fmt.Sprintf("%s %s\n", status, task.Name))
	}
	buf.WriteString("\n")

	// Advancement readiness
	if allFinalizationTasksComplete(p) {
		buf.WriteString("✓ All finalization tasks complete!\n\n")
		buf.WriteString("Ready to complete design. Run: `sow project advance`\n\n")
	}

	// Render additional guidance from template
	guidance, err := templates.Render(templatesFS, "templates/finalizing.md", p)
	if err != nil {
		return buf.String() + fmt.Sprintf("\nError rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}
```

**Key patterns**:
- Prompt generators are simple functions: `func(*state.Project) string`
- Registered via `WithPrompt(state, generator)` in configuration
- Can use embedded templates for complex prompts
- Provide contextual guidance based on current state

## Task Lifecycle

### Tasks Represent Documents to Create

Each task tracks a document's lifecycle from planning through approval:
- **Task name**: Document title/description
- **Task inputs** (optional): Input artifacts informing this specific document
- **Task outputs**: The drafted document artifact
- **Task status**: Current state of document creation
- **Task metadata**: Links to artifact, stores document type and target location

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
- `artifact_path`: Set when artifact added to task outputs (drafting)
- `template`: Optional, set when task created

### Task-to-Artifact Linking

**Via task outputs**: The SDK already supports task outputs, so artifacts are linked by adding them to `task.outputs`.

**Task completion auto-approves artifact**:

This logic is implemented in the transition actions or via helper functions called during task status updates. The SDK doesn't enforce this automatically, but the design project type implements it via guards and actions.

## Input Management

### Input Artifacts

Track sources that inform the design:
- Exploration findings
- Existing documentation
- External references (URLs, papers, etc.)
- Related code examples

**Storage**: `design.inputs` array (separate from `design.outputs`)

**Rationale**:
- Clear separation: inputs inform, outputs are created
- Traceable: know what informed each design decision
- Resumable: new orchestrator can read inputs for context

### Input vs Output Artifacts

**Inputs** (`phase.inputs`):
- Sources that inform design
- Do not require approval
- Not moved during finalization
- Can also be task-specific (task.inputs)

**Outputs** (`phase.outputs`):
- Documents being created
- Linked to tasks via task.outputs
- Require approval (auto-approved on task completion)
- Moved to target locations during finalization

## Metadata Schemas

**File**: `cli/internal/projects/design/metadata.go`

```go
package design

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/design_metadata.cue
var designMetadataSchema string

//go:embed cue/finalization_metadata.cue
var finalizationMetadataSchema string
```

**File**: `cli/internal/projects/design/cue/design_metadata.cue`

```cue
package design

// Metadata schema for design phase
{
	// No required metadata for design phase
	// Optional metadata can be added as needed
}
```

**File**: `cli/internal/projects/design/cue/finalization_metadata.cue`

```cue
package design

// Metadata schema for finalization phase
{
	// pr_url: Optional URL of created pull request
	pr_url?: string

	// project_deleted: Flag indicating .sow/project/ has been deleted
	project_deleted?: bool
}
```

## Artifact Management

### Artifact Guard: Task Existence

**Concept**: Cannot add artifacts before creating tasks

This is enforced at the orchestrator level or via validation. The SDK doesn't provide built-in guards for artifact addition, but the design workflow should validate this.

**Rationale**:
- Enforces planning before drafting
- Ensures every document has task tracking it
- Prevents orphaned artifacts

### Artifact Auto-Approval

**Triggered when task completed**:

The auto-approval logic is implemented via helper functions in guards.go (`autoApproveArtifact`) and called during task status updates or transition actions.

**Workflow**:
1. Task in `needs_review`
2. Orchestrator updates task status to `completed`
3. Validation checks artifact exists at `task.metadata.artifact_path` (or in task.outputs)
4. Helper function sets `artifact.approved = true` on the linked artifact
5. Task marked completed

## Orchestrator Workflow

### Active State Workflow

1. **Initialize design** (via `sow project new` on `design/*` branch)

2. **Register inputs** (optional):
   - Add to `design.inputs` via orchestrator operations
   - Track sources that inform design decisions

3. **Plan outputs as tasks**:
   - Create task for each document to create
   - Set metadata: `document_type`, `target_location`, optional `template`
   - Tasks start in `pending` status

4. **Draft documents**:
   - Update task status to `in_progress`
   - Orchestrator drafts document
   - Add document as artifact to `design.outputs`
   - Link artifact to task via `task.outputs` or `task.metadata.artifact_path`

5. **Mark ready for review**:
   - Update task status to `needs_review`
   - User reviews document

6. **Review iteration**:
   - If changes needed: update task status to `in_progress`
   - Orchestrator updates document
   - Mark `needs_review` again
   - Repeat until satisfied

7. **Approve document**:
   - Update task status to `completed`
   - Auto-approve linked artifact (via helper function)

8. **Repeat for remaining documents**

9. **Complete design**:
   - When all documents approved: `sow project advance`
   - Transitions to Finalizing

### Finalizing State Workflow

1. **Transition to Finalizing**: `sow project advance` from Active

2. **Create finalization tasks** (orchestrator):
   - For each completed task with `target_location`:
     - Create task to move artifact to target location
   - Create task to create PR with design artifacts
   - Create task to delete `.sow/project/` directory

3. **Execute tasks**:
   - Orchestrator completes each task sequentially
   - Updates task status to `completed`

4. **Complete finalization**: `sow project advance`

5. **Design finished**: State machine reaches Completed

## Testing Strategy

### Unit Tests

**Schema validation**:
- ProjectState validates correctly with type="design"
- Phase status constraints enforced
- State machine state constraints enforced

**Guards** (`guards_test.go`):
- `allDocumentsApproved` returns false with pending tasks
- `allDocumentsApproved` returns true when all completed/abandoned
- `allDocumentsApproved` requires at least one completed
- `allFinalizationTasksComplete` validates all tasks completed
- `validateTaskForCompletion` checks artifact exists
- `autoApproveArtifact` approves linked artifact

**Configuration building** (`design_test.go`):
- Builder creates valid ProjectTypeConfig
- All transitions configured correctly
- Guards bound via closures
- Prompts registered for all states

### Integration Tests

**Full workflow** (`integration_test.go`):
1. Create design project on `design/test` branch
2. Add 2 inputs to design phase
3. Create 2 document tasks with metadata
4. Draft and approve both documents
5. Advance to Finalizing
6. Complete finalization
7. Verify documents moved to target locations

**Edge cases**:
- Add artifact before creating tasks (orchestrator should prevent)
- Complete task without artifact_path (should fail validation)
- Complete task with non-existent artifact (should fail validation)
- Complete task successfully (should auto-approve artifact)
- `needs_review` → `in_progress` transition (should work)
- Complete with all tasks abandoned (should fail - need at least one completed)

### SDK Integration Tests

**State machine integration**:
- BuildMachine creates valid state machine
- Guards prevent invalid transitions
- Events fire correctly
- Prompts generate for all states

**Registry integration**:
- `state.Register("design", config)` succeeds
- `state.Get("design")` returns correct config
- Config validates project state correctly

## Migration Notes

No migration from old design mode - users restart active design sessions.

Design sessions are typically short (days to weeks), so breaking change acceptable.

## Success Criteria

1. ✅ Can create design project on `design/*` branch
2. ✅ Can register inputs to track design sources
3. ✅ Can create document tasks with metadata
4. ✅ Can link artifacts to tasks via task.outputs or metadata
5. ✅ Can transition tasks through pending → in_progress → needs_review → completed
6. ✅ Can go backward from needs_review → in_progress for revisions
7. ✅ Completing task auto-approves linked artifact
8. ✅ Cannot complete task without artifact existing (validation)
9. ✅ Can advance to Finalizing when all documents approved (at least one completed)
10. ✅ Finalization moves documents to target locations
11. ✅ Zero-context resumability works (project state on disk)
12. ✅ Prompts provide clear guidance at each state
13. ✅ SDK configuration is declarative and testable

## Future Enhancements

**Not in scope for initial implementation**:

- Template validation (verify document matches template structure)
- Automatic template loading based on `task.metadata.template`
- Multi-reviewer approval workflow
- Document versioning/iteration tracking
- Linking designs to implementations (track which projects implement this design)
- Design impact analysis (which systems affected by this design)

## References

- **Core Design**: [Modes to Project Types](core-design.md)
- **Project SDK**: `cli/internal/sdks/project/` (builder, config, machine)
- **Standard Project**: `cli/internal/projects/standard/` (reference implementation)
- **State Machine SDK**: `cli/internal/sdks/state/` (underlying state machine)
