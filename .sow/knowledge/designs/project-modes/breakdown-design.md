# Breakdown Project Type Design

**Author**: Architecture Team
**Date**: 2025-11-04
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
- **SDK-based implementation**: Built using Project SDK builder pattern

## Goals and Non-Goals

**Goals**:
- Decompose design docs or large features into implementable work units
- Specify each work unit with sufficient detail
- Review all work units before publishing to GitHub
- Publish work units as GitHub issues in dependency order
- Track issue metadata (number, URL) for reference
- Support work unit dependencies
- Use Project SDK for consistent, declarative configuration

**Non-Goals**:
- Implementation tracking (that's standard project's job)
- Finalization/PR creation (no artifacts to commit)
- Issue lifecycle management (once published, GitHub owns lifecycle)
- Multi-repository breakdown (single repo only)

## Project SDK Configuration

The breakdown project type uses the Project SDK builder pattern. All configuration is declarative and registered via the global registry.

### Package Structure

```
cli/internal/projects/breakdown/
├── breakdown.go       # SDK configuration using builder pattern
├── states.go          # State constants
├── events.go          # Event constants
├── guards.go          # Guard helper functions
├── prompts.go         # Prompt generator functions
├── metadata.go        # Embedded CUE metadata schemas
├── publishing.go      # GitHub issue publishing logic
├── cue/
│   └── breakdown_metadata.cue  # Phase metadata validation
└── templates/
    ├── orchestrator.md         # Orchestrator-level guidance
    ├── active.md               # Active state prompt template
    └── publishing.md           # Publishing state prompt template
```

### Registration

**File**: `cli/internal/projects/breakdown/breakdown.go`

```go
package breakdown

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// init registers the breakdown project type on package load.
func init() {
	state.Register("breakdown", NewBreakdownProjectConfig())
}

// NewBreakdownProjectConfig creates the complete configuration for breakdown project type.
func NewBreakdownProjectConfig() *project.ProjectTypeConfig {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	builder = configurePhases(builder)
	builder = configureTransitions(builder)
	builder = configureEventDeterminers(builder)
	builder = configurePrompts(builder)
	builder = builder.WithInitializer(initializeBreakdownProject)
	return builder.Build()
}
```

### Phase Configuration

Breakdown has a single-phase structure:
1. **breakdown** - Decompose work, specify units, review, publish to GitHub

**No finalization phase** - Deliverables are GitHub issues, not committed files.

```go
func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("breakdown",
			project.WithStartState(sdkstate.State(Active)),
			project.WithEndState(sdkstate.State(Publishing)),
			project.WithOutputs("work_unit_spec"),  // Allowed artifact types
			project.WithTasks(),  // Phase supports tasks
			project.WithMetadataSchema(breakdownMetadataSchema),
		)
}
```

**Key points**:
- `breakdown` phase has TWO states within it (Active, Publishing)
- No finalization phase - project completes after publishing
- Only `breakdown` phase supports tasks (work unit specifications)
- Artifact types constrained to work unit specs

### Phase Initialization

```go
// initializeBreakdownProject creates the breakdown phase for a new breakdown project.
// This is called during project creation to set up the phase structure.
func initializeBreakdownProject(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
	now := p.Created_at

	// Get initial inputs for breakdown phase (empty slice if none provided)
	inputs := []projschema.ArtifactState{}
	if initialInputs != nil {
		if phaseInputs, exists := initialInputs["breakdown"]; exists {
			inputs = phaseInputs
		}
	}

	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",  // Breakdown starts immediately in active state
		Enabled:    true,
		Created_at: now,
		Inputs:     inputs,
		Outputs:    []projschema.ArtifactState{},
		Tasks:      []projschema.TaskState{},
		Metadata:   make(map[string]interface{}),
	}

	return nil
}
```

## State Machine

### States

The breakdown workflow progresses through three states within a single phase:

**Active** (breakdown phase, status="active")
- **Purpose**: Decompose work, specify units, review all work units
- **Duration**: Most of breakdown time
- **Operations**: Create work unit tasks, write specifications, iterate on reviews
- **Advance condition**: All tasks resolved and dependencies validated

**Publishing** (breakdown phase, status="publishing")
- **Purpose**: Create GitHub issues for approved work units in dependency order
- **Duration**: Short - automated issue creation
- **Operations**: Create GitHub issues, handle dependencies, track metadata
- **Complete condition**: All work units published

**Completed** (terminal state)
- **Purpose**: All issues published, project deleted
- **Duration**: Permanent
- **Operations**: None (project complete)

### State Transitions

```
Active (breakdown.active)
  │
  │ EventBeginPublishing
  │ Guard: allWorkUnitsApproved(p) && dependenciesValid(p)
  ▼
Publishing (breakdown.publishing)
  │
  │ EventCompleteBreakdown
  │ Guard: allWorkUnitsPublished(p)
  │ Action: Delete .sow/project/
  ▼
Completed (terminal)
```

### Transition Configuration

```go
func configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		// Initial state
		SetInitialState(sdkstate.State(Active)).

		// Active → Publishing (intra-phase transition)
		AddTransition(
			sdkstate.State(Active),
			sdkstate.State(Publishing),
			sdkstate.Event(EventBeginPublishing),
			project.WithGuard(func(p *state.Project) bool {
				// Both conditions must pass
				return allWorkUnitsApproved(p) && dependenciesValid(p)
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Update breakdown phase status to "publishing"
				phase := p.Phases["breakdown"]
				phase.Status = "publishing"
				p.Phases["breakdown"] = phase
				return nil
			}),
		).

		// Publishing → Completed
		AddTransition(
			sdkstate.State(Publishing),
			sdkstate.State(Completed),
			sdkstate.Event(EventCompleteBreakdown),
			project.WithGuard(func(p *state.Project) bool {
				return allWorkUnitsPublished(p)
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Mark breakdown phase as completed
				phase := p.Phases["breakdown"]
				phase.Status = "completed"
				phase.Completed_at = time.Now()
				p.Phases["breakdown"] = phase

				// Cleanup: delete project directory
				// This is done in OnEntry to ensure it happens even if save fails
				// Errors are logged but don't prevent transition
				return nil
			}),
		)
}
```

**Key patterns**:
- Guards are closures that capture project state via SDK binding
- OnEntry actions update phase status and timestamps
- Intra-phase transition (Active → Publishing) and terminal transition (Publishing → Completed)
- Cleanup happens in transition action, not separate phase

### Event Determiners

```go
func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(sdkstate.State(Active), func(p *state.Project) (sdkstate.Event, error) {
			// Active always advances to Publishing
			return sdkstate.Event(EventBeginPublishing), nil
		}).
		OnAdvance(sdkstate.State(Publishing), func(p *state.Project) (sdkstate.Event, error) {
			// Complete breakdown
			return sdkstate.Event(EventCompleteBreakdown), nil
		})
}
```

## States and Events

**File**: `cli/internal/projects/breakdown/states.go`

```go
package breakdown

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Breakdown project states
const (
	// Active indicates active breakdown phase
	Active = state.State("Active")

	// Publishing indicates GitHub issue creation in progress
	Publishing = state.State("Publishing")

	// Completed indicates breakdown finished
	Completed = state.State("Completed")
)
```

**File**: `cli/internal/projects/breakdown/events.go`

```go
package breakdown

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Breakdown project events trigger state transitions
const (
	// EventBeginPublishing transitions from Active to Publishing
	// Fired when all work units are approved and dependencies are valid
	EventBeginPublishing = state.Event("begin_publishing")

	// EventCompleteBreakdown transitions from Publishing to Completed
	// Fired when all work units are published to GitHub
	EventCompleteBreakdown = state.Event("complete_breakdown")
)
```

## Guards

Guards are pure functions that check transition conditions. They operate on `*state.Project` and return boolean results.

**File**: `cli/internal/projects/breakdown/guards.go`

```go
package breakdown

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// allWorkUnitsApproved checks if all work unit tasks are completed or abandoned.
// Guards Active → Publishing transition (combined with dependenciesValid).
// Returns false if no tasks exist, or if at least one completed task doesn't exist.
func allWorkUnitsApproved(p *state.Project) bool {
	phase, exists := p.Phases["breakdown"]
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

// dependenciesValid checks that task dependencies form a valid DAG (no cycles, all references valid).
// Guards Active → Publishing transition (combined with allWorkUnitsApproved).
func dependenciesValid(p *state.Project) bool {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return false
	}

	// Build adjacency list and task ID set
	graph := make(map[string][]string)
	taskIDs := make(map[string]bool)

	for _, task := range phase.Tasks {
		if task.Status == "completed" {
			taskIDs[task.Id] = true
			// Get dependencies from metadata
			if task.Metadata != nil {
				if depsRaw, ok := task.Metadata["dependencies"]; ok {
					if deps, ok := depsRaw.([]interface{}); ok {
						depStrings := make([]string, 0, len(deps))
						for _, d := range deps {
							if depStr, ok := d.(string); ok {
								depStrings = append(depStrings, depStr)
							}
						}
						if len(depStrings) > 0 {
							graph[task.Id] = depStrings
						}
					}
				}
			}
		}
	}

	// Check all dependencies point to valid tasks
	for taskID, deps := range graph {
		for _, depID := range deps {
			if !taskIDs[depID] {
				return false // Invalid dependency reference
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
				return true // Cycle detected
			}
		}

		recStack[taskID] = false
		return false
	}

	for taskID := range graph {
		if !visited[taskID] {
			if hasCycle(taskID) {
				return false // Cyclic dependency
			}
		}
	}

	return true
}

// allWorkUnitsPublished checks if all completed work units have been published to GitHub.
// Guards Publishing → Completed transition.
func allWorkUnitsPublished(p *state.Project) bool {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return false
	}

	hasCompleted := false

	for _, task := range phase.Tasks {
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

// Helper functions for counting (used in prompts and error messages)

func countUnresolvedTasks(p *state.Project) int {
	phase, exists := p.Phases["breakdown"]
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

func countUnpublishedTasks(p *state.Project) int {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return 0
	}

	count := 0
	for _, task := range phase.Tasks {
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

// validateTaskForCompletion checks if a task can be marked as completed.
// Returns error if artifact doesn't exist at task.metadata.artifact_path.
func validateTaskForCompletion(p *state.Project, taskID string) error {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return fmt.Errorf("breakdown phase not found")
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
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return fmt.Errorf("breakdown phase not found")
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
	phaseData := p.Phases["breakdown"]
	for i := range phaseData.Outputs {
		if phaseData.Outputs[i].Path == artifactPath {
			phaseData.Outputs[i].Approved = true
			p.Phases["breakdown"] = phaseData
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
- Dependency validation integrated into guard (not separate phase operation)

## Prompt Generation

Prompts provide contextual guidance for each state. The SDK integrates prompts via the `WithPrompt()` builder method.

**File**: `cli/internal/projects/breakdown/prompts.go`

```go
package breakdown

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
		// Orchestrator-level prompt (how breakdown projects work)
		WithOrchestratorPrompt(generateOrchestratorPrompt).

		// State-level prompts (what to do in each state)
		WithPrompt(sdkstate.State(Active), generateActivePrompt).
		WithPrompt(sdkstate.State(Publishing), generatePublishingPrompt)
}

// generateOrchestratorPrompt generates the orchestrator-level prompt for breakdown projects.
func generateOrchestratorPrompt(p *state.Project) string {
	prompt, err := templates.Render(templatesFS, "templates/orchestrator.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering orchestrator prompt: %v", err)
	}
	return prompt
}

// generateActivePrompt generates the prompt for the Active state.
// Focus: Decompose work, specify units, review all work units.
func generateActivePrompt(p *state.Project) string {
	var buf strings.Builder

	// Project header
	buf.WriteString(fmt.Sprintf("# Breakdown: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Current state
	buf.WriteString("## Current State: Active Breakdown\n\n")

	// Breakdown phase info
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return "Error: breakdown phase not found"
	}

	// Show inputs if any
	if len(phase.Inputs) > 0 {
		buf.WriteString("### Being Broken Down\n\n")
		buf.WriteString("Sources being decomposed:\n\n")
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

	// Work units
	buf.WriteString("### Work Units\n\n")

	if len(phase.Tasks) == 0 {
		buf.WriteString("No work units identified yet.\n\n")
		buf.WriteString("**Important**: Create at least one work unit task before adding specifications.\n\n")
		buf.WriteString("**Next steps**: Identify work units to create\n\n")
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

		buf.WriteString(fmt.Sprintf("Total: %d work units\n", len(phase.Tasks)))
		buf.WriteString(fmt.Sprintf("- Pending: %d\n", pending))
		buf.WriteString(fmt.Sprintf("- In Progress: %d\n", inProgress))
		buf.WriteString(fmt.Sprintf("- Needs Review: %d\n", needsReview))
		buf.WriteString(fmt.Sprintf("- Completed: %d\n", completed))
		buf.WriteString(fmt.Sprintf("- Abandoned: %d\n\n", abandoned))

		// List work units
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
			if task.Metadata != nil {
				if depsRaw, ok := task.Metadata["dependencies"]; ok {
					if deps, ok := depsRaw.([]interface{}); ok && len(deps) > 0 {
						depStrs := make([]string, 0, len(deps))
						for _, d := range deps {
							if depStr, ok := d.(string); ok {
								depStrs = append(depStrs, depStr)
							}
						}
						if len(depStrs) > 0 {
							buf.WriteString(fmt.Sprintf("    Depends on: %v\n", depStrs))
						}
					}
				}

				if artifactPath, ok := task.Metadata["artifact_path"].(string); ok {
					buf.WriteString(fmt.Sprintf("    Spec: %s\n", artifactPath))
				}
			}
		}
		buf.WriteString("\n")

		// Advancement readiness
		if allWorkUnitsApproved(p) && dependenciesValid(p) {
			buf.WriteString("✓ All work units approved and dependencies validated!\n\n")
			buf.WriteString("Ready to publish GitHub issues. Run: `sow project advance`\n\n")
		} else {
			if !allWorkUnitsApproved(p) {
				unresolvedCount := countUnresolvedTasks(p)
				buf.WriteString(fmt.Sprintf("**Next steps**: Continue breakdown work (%d work units remaining)\n\n", unresolvedCount))
			} else {
				buf.WriteString("**Next steps**: Dependency validation failed - check for cycles or invalid references\n\n")
			}
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

// generatePublishingPrompt generates the prompt for the Publishing state.
// Focus: Create GitHub issues for work units in dependency order.
func generatePublishingPrompt(p *state.Project) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Breakdown: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n\n", p.Branch))

	buf.WriteString("## Current State: Publishing\n\n")
	buf.WriteString("All work units approved. Creating GitHub issues in dependency order.\n\n")

	// Breakdown phase info
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return "Error: breakdown phase not found"
	}

	// Publishing status
	buf.WriteString("### Publishing Status\n\n")

	completed := []projschema.TaskState{}
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

	// Advancement readiness
	if allWorkUnitsPublished(p) {
		buf.WriteString("✓ All work units published!\n\n")
		buf.WriteString("Breakdown complete. Run: `sow project advance`\n\n")
	} else {
		buf.WriteString(fmt.Sprintf("**Next steps**: Publish remaining %d work units\n\n", unpublished))
	}

	// Render additional guidance from template
	guidance, err := templates.Render(templatesFS, "templates/publishing.md", p)
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

### Tasks Represent Work Units

Each task represents a work unit that will become a GitHub issue:
- **Task name**: Work unit title (becomes issue title)
- **Task inputs** (optional): Input artifacts informing this specific work unit
- **Task outputs**: The work unit specification artifact
- **Task status**: Current state of work unit specification
- **Task metadata**: Artifact path, dependencies, publishing status, GitHub issue info

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
  dependencies: ["001", "002"]        # Task IDs this depends on
  published: false                    # Set to true after issue created
  github_issue_number: 123            # Set during publishing
  github_issue_url: "https://github.com/org/repo/issues/123"  # Set during publishing
  work_unit_type: "feature"           # Optional: feature, bug, refactor, spike
```

**When metadata is set**:
- `artifact_path`: Set when artifact added to task outputs (during specification)
- `dependencies`: Set when task created or updated
- `published`: Initially false, set to true during Publishing state
- `github_issue_number`: Set during Publishing state after issue created
- `github_issue_url`: Set during Publishing state after issue created
- `work_unit_type`: Optional, set when task created

### Task Dependencies

**Purpose**: Determine GitHub issue creation order

**Storage**: `task.metadata.dependencies` as array of task IDs

**Validation**: Checked during Active → Publishing transition
- No cyclic dependencies
- All dependency IDs reference valid completed tasks

**Publishing order**:
- Tasks with no dependencies published first
- Tasks with dependencies published only after their dependencies
- Uses topological sort to determine safe ordering

## Work Unit Specifications

### Specification Format

Each work unit has a specification artifact stored in `breakdown.outputs`:

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

## Metadata Schemas

**File**: `cli/internal/projects/breakdown/metadata.go`

```go
package breakdown

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/breakdown_metadata.cue
var breakdownMetadataSchema string
```

**File**: `cli/internal/projects/breakdown/cue/breakdown_metadata.cue`

```cue
package breakdown

// Metadata schema for breakdown phase
{
	// No required metadata for breakdown phase
	// Optional metadata can be added as needed
}
```

## Publishing Workflow

### Publishing Logic

Publishing is implemented via helper functions in the orchestrator or a dedicated publishing module.

**File**: `cli/internal/projects/breakdown/publishing.go`

```go
package breakdown

import (
	"fmt"
	"sort"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// PublishingOrder determines the order in which work units should be published to GitHub.
// Returns task IDs in topological order (dependencies first).
func PublishingOrder(p *state.Project) ([]string, error) {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return nil, fmt.Errorf("breakdown phase not found")
	}

	// Build graph of completed tasks
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	taskIDs := []string{}

	for _, task := range phase.Tasks {
		if task.Status == "completed" {
			taskIDs = append(taskIDs, task.Id)
			inDegree[task.Id] = 0
			graph[task.Id] = []string{}

			// Get dependencies from metadata
			if task.Metadata != nil {
				if depsRaw, ok := task.Metadata["dependencies"]; ok {
					if deps, ok := depsRaw.([]interface{}); ok {
						for _, d := range deps {
							if depStr, ok := d.(string); ok {
								graph[task.Id] = append(graph[task.Id], depStr)
								inDegree[task.Id]++
							}
						}
					}
				}
			}
		}
	}

	// Topological sort using Kahn's algorithm
	queue := []string{}
	for _, id := range taskIDs {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	result := []string{}
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Process neighbors
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Verify all tasks processed (no cycles)
	if len(result) != len(taskIDs) {
		return nil, fmt.Errorf("cyclic dependency detected")
	}

	return result, nil
}

// PublishWorkUnit creates a GitHub issue for a single work unit.
// This is called by the orchestrator during the Publishing state.
func PublishWorkUnit(p *state.Project, taskID string, githubClient GitHubClient) error {
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return fmt.Errorf("breakdown phase not found")
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

	// Check if already published
	if task.Metadata != nil {
		if published, ok := task.Metadata["published"].(bool); ok && published {
			return fmt.Errorf("task %s already published", taskID)
		}
	}

	// Get artifact path from metadata
	if task.Metadata == nil || task.Metadata["artifact_path"] == nil {
		return fmt.Errorf("task %s has no artifact_path", taskID)
	}

	artifactPath := task.Metadata["artifact_path"].(string)

	// Find and read specification artifact
	var spec *projschema.ArtifactState
	for i := range phase.Outputs {
		if phase.Outputs[i].Path == artifactPath {
			spec = &phase.Outputs[i]
			break
		}
	}

	if spec == nil {
		return fmt.Errorf("artifact not found at %s", artifactPath)
	}

	// Create GitHub issue
	labels := []string{"sow"}
	if task.Metadata["work_unit_type"] != nil {
		if wuType, ok := task.Metadata["work_unit_type"].(string); ok {
			labels = append(labels, wuType)
		}
	}

	issue, err := githubClient.CreateIssue(task.Name, spec.Path, labels)
	if err != nil {
		return fmt.Errorf("failed to create GitHub issue: %w", err)
	}

	// Update task metadata
	phaseData := p.Phases["breakdown"]
	for i := range phaseData.Tasks {
		if phaseData.Tasks[i].Id == taskID {
			if phaseData.Tasks[i].Metadata == nil {
				phaseData.Tasks[i].Metadata = make(map[string]interface{})
			}
			phaseData.Tasks[i].Metadata["published"] = true
			phaseData.Tasks[i].Metadata["github_issue_number"] = issue.Number
			phaseData.Tasks[i].Metadata["github_issue_url"] = issue.URL
			break
		}
	}
	p.Phases["breakdown"] = phaseData

	return nil
}

// GitHubClient interface for creating issues
type GitHubClient interface {
	CreateIssue(title string, specPath string, labels []string) (*GitHubIssue, error)
}

// GitHubIssue represents a created GitHub issue
type GitHubIssue struct {
	Number int
	URL    string
}
```

**Key patterns**:
- Publishing logic separate from SDK configuration
- Helper functions operate on `*state.Project`
- Orchestrator calls these during Publishing state
- Resumable: check `published` flag before creating issue

## Orchestrator Workflow

### Active State Workflow

1. **Initialize breakdown** (via `sow project new` on `breakdown/*` branch)

2. **Register inputs** (what's being broken down):
   - Add to `breakdown.inputs` via orchestrator operations
   - Track design docs or features being decomposed

3. **Identify work units as tasks**:
   - Create task for each work unit
   - Set metadata: optional `dependencies`, `work_unit_type`
   - Tasks start in `pending` status

4. **Specify work units**:
   - Update task status to `in_progress`
   - Orchestrator writes specification
   - Add specification as artifact to `breakdown.outputs`
   - Link artifact to task via `task.metadata.artifact_path`

5. **Mark ready for review**:
   - Update task status to `needs_review`
   - User reviews specification

6. **Review iteration**:
   - If changes needed: update task status to `in_progress`
   - Orchestrator updates specification
   - Mark `needs_review` again
   - Repeat until satisfied

7. **Approve specification**:
   - Update task status to `completed`
   - Auto-approve linked artifact (via helper function)

8. **Repeat for remaining work units**

9. **Advance to Publishing**:
   - When all work units approved: `sow project advance`
   - Guards validate dependencies (no cycles, valid references)
   - Transitions to Publishing

### Publishing State Workflow

1. **Transition to Publishing**: `sow project advance` from Active

2. **Determine publishing order**:
   - Call `PublishingOrder(project)` to get topological sort
   - Returns task IDs in safe order (dependencies first)

3. **Publish work units** (orchestrator automates this):
   - For each work unit in order:
     - Check if already published (skip if `metadata.published == true`)
     - Call `PublishWorkUnit(project, taskID, githubClient)`
     - Reads specification from artifact
     - Creates GitHub issue
     - Stores issue metadata in task
     - Marks `published = true`
     - Saves state (resumability)

4. **All units published**:
   - When all completed tasks published: `sow project advance`
   - Transitions to Completed
   - Project directory deleted via transition action

5. **Breakdown finished**: State machine reaches Completed

## Testing Strategy

### Unit Tests

**Schema validation**:
- ProjectState validates correctly with type="breakdown"
- Phase status constraints enforced
- State machine state constraints enforced

**Guards** (`guards_test.go`):
- `allWorkUnitsApproved` returns false with pending tasks
- `allWorkUnitsApproved` returns true when all completed/abandoned
- `allWorkUnitsApproved` requires at least one completed
- `dependenciesValid` returns false with cyclic dependencies
- `dependenciesValid` returns false with invalid task references
- `dependenciesValid` returns true for valid DAG
- `allWorkUnitsPublished` returns false if any completed task not published
- `allWorkUnitsPublished` validates all completed tasks published

**Configuration building** (`breakdown_test.go`):
- Builder creates valid ProjectTypeConfig
- All transitions configured correctly
- Guards bound via closures
- Prompts registered for all states

**Publishing logic** (`publishing_test.go`):
- `PublishingOrder` returns correct topological order
- `PublishingOrder` detects cyclic dependencies
- `PublishWorkUnit` creates GitHub issue correctly
- `PublishWorkUnit` updates task metadata correctly
- `PublishWorkUnit` skips already-published tasks

### Integration Tests

**Full workflow** (`integration_test.go`):
1. Create breakdown project on `breakdown/test` branch
2. Add 1 input to breakdown phase
3. Create 3 work unit tasks with dependencies
4. Specify and approve all work units
5. Advance to Publishing
6. Publish all work units (mock GitHub API)
7. Advance to Completed
8. Verify project directory deleted

**Edge cases**:
- Add artifact before creating tasks (orchestrator should prevent)
- Complete task without artifact_path (should fail validation)
- Complete task with non-existent artifact (should fail validation)
- Advance with unresolved tasks (should fail guard)
- Advance with cyclic dependencies (should fail guard)
- Advance with invalid dependency reference (should fail guard)
- Advance without unpublished tasks (should fail guard)
- `needs_review` → `in_progress` transition (should work)

### SDK Integration Tests

**State machine integration**:
- BuildMachine creates valid state machine
- Guards prevent invalid transitions
- Events fire correctly
- Prompts generate for all states

**Registry integration**:
- `state.Register("breakdown", config)` succeeds
- `state.Get("breakdown")` returns correct config
- Config validates project state correctly

## Migration Notes

No migration from old breakdown mode - users restart active breakdown sessions.

Breakdown sessions are typically short (days to weeks), so breaking change acceptable.

## Success Criteria

1. ✅ Can create breakdown project on `breakdown/*` branch
2. ✅ Can register inputs to track what's being broken down
3. ✅ Can create work unit tasks with dependencies
4. ✅ Can link specifications to tasks via task.metadata or task.outputs
5. ✅ Can transition tasks through pending → in_progress → needs_review → completed
6. ✅ Can go backward from needs_review → in_progress for revisions
7. ✅ Completing task auto-approves specification artifact
8. ✅ Cannot complete task without artifact existing (validation)
9. ✅ Cannot advance to Publishing with unresolved tasks
10. ✅ Cannot advance to Publishing with cyclic dependencies
11. ✅ Cannot advance to Publishing with invalid dependency references
12. ✅ Publishing creates GitHub issues in dependency order
13. ✅ Publishing tracks issue metadata in tasks
14. ✅ Publishing is resumable (can interrupt and retry)
15. ✅ Cannot advance from Publishing with unpublished work units
16. ✅ Completing breakdown deletes project directory
17. ✅ Zero-context resumability works (project state on disk)
18. ✅ Prompts provide clear guidance at each state
19. ✅ SDK configuration is declarative and testable

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

## References

- **Core Design**: [Modes to Project Types](core-design.md)
- **Project SDK**: `cli/internal/sdks/project/` (builder, config, machine)
- **Standard Project**: `cli/internal/projects/standard/` (reference implementation)
- **State Machine SDK**: `cli/internal/sdks/state/` (underlying state machine)
