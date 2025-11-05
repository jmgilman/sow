# Exploration Project Type Design

**Author**: Architecture Team
**Date**: 2025-11-04
**Status**: Proposed
**Depends On**: [Core Design](core-design.md)

## Overview

This document specifies the exploration project type - a workflow for research, investigation, and knowledge gathering. Exploration projects help users systematically investigate topics, document findings, and synthesize results into comprehensive summaries.

**Key Characteristics**:
- **Flexible topic discovery**: Add research topics dynamically as exploration progresses
- **Task-based lifecycle**: Each topic is a task with independent status tracking
- **Synthesis-focused**: Culminates in approved summary artifact(s)
- **No GitHub integration**: Pure research workflow, no issue linking
- **SDK-based implementation**: Built using Project SDK builder pattern

## Goals and Non-Goals

**Goals**:
- Enable systematic research and investigation workflows
- Support dynamic topic discovery (add topics anytime during active research)
- Track research progress per topic independently
- Produce approved summary artifact(s) documenting findings
- Maintain simple, flexible workflow (avoid over-structuring)
- Use Project SDK for consistent, declarative configuration

**Non-Goals**:
- GitHub issue integration (exploration is pre-implementation research)
- Rigid topic approval workflow (topics can be added/abandoned freely)
- Deliverable tracking beyond summary (exploration is knowledge gathering, not delivery)
- Complex artifact approval chains (only summaries require approval)

## Project SDK Configuration

The exploration project type uses the Project SDK builder pattern introduced in the standard project type. All configuration is declarative and registered via the global registry.

### Package Structure

```
cli/internal/projects/exploration/
├── exploration.go      # SDK configuration using builder pattern
├── states.go          # State constants
├── events.go          # Event constants
├── guards.go          # Guard helper functions
├── prompts.go         # Prompt generator functions
├── metadata.go        # Embedded CUE metadata schemas
├── cue/
│   └── exploration_metadata.cue  # Phase metadata validation
└── templates/
    ├── orchestrator.md            # Orchestrator-level guidance
    ├── active.md                  # Active state prompt template
    ├── summarizing.md             # Summarizing state prompt template
    └── finalizing.md              # Finalizing state prompt template
```

### Registration

**File**: `cli/internal/projects/exploration/exploration.go`

```go
package exploration

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// init registers the exploration project type on package load.
func init() {
	state.Register("exploration", NewExplorationProjectConfig())
}

// NewExplorationProjectConfig creates the complete configuration for exploration project type.
func NewExplorationProjectConfig() *project.ProjectTypeConfig {
	builder := project.NewProjectTypeConfigBuilder("exploration")
	builder = configurePhases(builder)
	builder = configureTransitions(builder)
	builder = configureEventDeterminers(builder)
	builder = configurePrompts(builder)
	builder = builder.WithInitializer(initializeExplorationProject)
	return builder.Build()
}
```

### Phase Configuration

Exploration has a 2-phase structure:
1. **exploration** - Active research and summarizing
2. **finalization** - Move artifacts, create PR, cleanup

```go
func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("exploration",
			project.WithStartState(sdkstate.State(Active)),
			project.WithEndState(sdkstate.State(Summarizing)),
			project.WithOutputs("summary", "findings"),  // Allowed artifact types
			project.WithTasks(),  // Phase supports tasks
			project.WithMetadataSchema(explorationMetadataSchema),
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
- `exploration` phase has TWO states within it (Active, Summarizing)
- Both phases support metadata validation via embedded CUE schemas
- Only `exploration` phase supports tasks (research topics)
- Artifact types constrained to prevent invalid outputs

### Phase Initialization

```go
// initializeExplorationProject creates all phases for a new exploration project.
// This is called during project creation to set up the phase structure.
func initializeExplorationProject(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
	now := p.Created_at
	phaseNames := []string{"exploration", "finalization"}

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
		if phaseName == "exploration" {
			status = "active"  // Exploration starts immediately in active state
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

The exploration workflow progresses through four states across two phases:

**Active** (exploration phase, status="active")
- **Purpose**: Active research - identify topics, investigate, document findings
- **Duration**: Most of exploration time spent here
- **Operations**: Add/complete tasks, create research findings
- **Advance condition**: All tasks resolved (completed or abandoned)

**Summarizing** (exploration phase, status="summarizing")
- **Purpose**: Synthesize findings into comprehensive summary document(s)
- **Duration**: Short - create and approve summaries
- **Operations**: Create summaries, approve artifacts
- **Complete condition**: All summary artifacts approved

**Finalizing** (finalization phase, status="in_progress")
- **Purpose**: Move artifacts to permanent location, create PR, cleanup
- **Duration**: Short - automated finalization
- **Operations**: Execute finalization checklist
- **Complete condition**: All finalization tasks completed

**Completed** (terminal state)
- **Purpose**: Exploration finished
- **Duration**: Permanent
- **Operations**: None (project complete)

### State Transitions

```
Active (exploration.active)
  │
  │ EventBeginSummarizing
  │ Guard: allTasksResolved(p)
  ▼
Summarizing (exploration.summarizing)
  │
  │ EventCompleteSummarizing
  │ Guard: allSummariesApproved(p)
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

		// Active → Summarizing (intra-phase transition)
		AddTransition(
			sdkstate.State(Active),
			sdkstate.State(Summarizing),
			sdkstate.Event(EventBeginSummarizing),
			project.WithGuard(func(p *state.Project) bool {
				return allTasksResolved(p)
			}),
			project.WithOnEntry(func(p *state.Project) error {
				// Update exploration phase status to "summarizing"
				phase := p.Phases["exploration"]
				phase.Status = "summarizing"
				p.Phases["exploration"] = phase
				return nil
			}),
		).

		// Summarizing → Finalizing (inter-phase transition)
		AddTransition(
			sdkstate.State(Summarizing),
			sdkstate.State(Finalizing),
			sdkstate.Event(EventCompleteSummarizing),
			project.WithGuard(func(p *state.Project) bool {
				return allSummariesApproved(p)
			}),
			project.WithOnExit(func(p *state.Project) error {
				// Mark exploration phase as completed
				phase := p.Phases["exploration"]
				phase.Status = "completed"
				phase.Completed_at = time.Now()
				p.Phases["exploration"] = phase
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
- Intra-phase transition (Active → Summarizing) vs inter-phase (Summarizing → Finalizing)

### Event Determiners

Event determiners tell the state machine which event to fire when `sow project advance` is called.

```go
func configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		OnAdvance(sdkstate.State(Active), func(p *state.Project) (sdkstate.Event, error) {
			// Simple: Active always advances to Summarizing
			return sdkstate.Event(EventBeginSummarizing), nil
		}).
		OnAdvance(sdkstate.State(Summarizing), func(p *state.Project) (sdkstate.Event, error) {
			// Complete exploration phase, transition to finalization
			return sdkstate.Event(EventCompleteSummarizing), nil
		}).
		OnAdvance(sdkstate.State(Finalizing), func(p *state.Project) (sdkstate.Event, error) {
			// Complete finalization phase
			return sdkstate.Event(EventCompleteFinalization), nil
		})
}
```

## States and Events

**File**: `cli/internal/projects/exploration/states.go`

```go
package exploration

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Exploration project states
const (
	// Active indicates active research phase
	Active = state.State("Active")

	// Summarizing indicates synthesis/summarizing phase
	Summarizing = state.State("Summarizing")

	// Finalizing indicates finalization in progress
	Finalizing = state.State("Finalizing")

	// Completed indicates exploration finished
	Completed = state.State("Completed")
)
```

**File**: `cli/internal/projects/exploration/events.go`

```go
package exploration

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Exploration project events trigger state transitions
const (
	// EventBeginSummarizing transitions from Active to Summarizing
	// Fired when all research topics are resolved
	EventBeginSummarizing = state.Event("begin_summarizing")

	// EventCompleteSummarizing transitions from Summarizing to Finalizing
	// Fired when all summary artifacts are approved
	EventCompleteSummarizing = state.Event("complete_summarizing")

	// EventCompleteFinalization transitions from Finalizing to Completed
	// Fired when all finalization tasks are completed
	EventCompleteFinalization = state.Event("complete_finalization")
)
```

## Guards

Guards are pure functions that check transition conditions. They operate on `*state.Project` and return boolean results.

**File**: `cli/internal/projects/exploration/guards.go`

```go
package exploration

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// allTasksResolved checks if all research topics in exploration phase are completed or abandoned.
// Guards Active → Summarizing transition.
// Returns false if no tasks exist or if any task is still pending/in_progress.
func allTasksResolved(p *state.Project) bool {
	phase, exists := p.Phases["exploration"]
	if !exists {
		return false
	}

	tasks := phase.Tasks
	if len(tasks) == 0 {
		return false // Must have at least one research topic
	}

	for _, task := range tasks {
		if task.Status != "completed" && task.Status != "abandoned" {
			return false
		}
	}
	return true
}

// allSummariesApproved checks if at least one summary artifact exists and all summaries are approved.
// Guards Summarizing → Finalizing transition.
//
// Summary artifacts are identified by having the approved field set (even if false).
// Research findings typically don't have the approved field set at all.
func allSummariesApproved(p *state.Project) bool {
	phase, exists := p.Phases["exploration"]
	if !exists {
		return false
	}

	summaries := []projschema.ArtifactState{}

	// Collect all summary artifacts (those with approved field set)
	for _, artifact := range phase.Outputs {
		// Summaries explicitly require approval, so approved field is set
		// Research findings added during Active typically don't set approved field
		if artifact.Type == "summary" {
			summaries = append(summaries, artifact)
		}
	}

	// Must have at least one summary
	if len(summaries) == 0 {
		return false
	}

	// All summaries must be approved
	for _, summary := range summaries {
		if !summary.Approved {
			return false
		}
	}

	return true
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
	phase, exists := p.Phases["exploration"]
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

func countUnapprovedSummaries(p *state.Project) int {
	phase, exists := p.Phases["exploration"]
	if !exists {
		return 0
	}

	count := 0
	for _, artifact := range phase.Outputs {
		if artifact.Type == "summary" && !artifact.Approved {
			count++
		}
	}
	return count
}
```

**Key patterns**:
- Guards operate on full `*state.Project`, accessing phases directly
- Pure functions with no side effects
- Used in transition configurations via closures
- Similar to standard project guards (cli/internal/projects/standard/guards.go)

## Prompt Generation

Prompts provide contextual guidance for each state. The SDK integrates prompts via the `WithPrompt()` builder method.

**File**: `cli/internal/projects/exploration/prompts.go`

```go
package exploration

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
		// Orchestrator-level prompt (how exploration projects work)
		WithOrchestratorPrompt(generateOrchestratorPrompt).

		// State-level prompts (what to do in each state)
		WithPrompt(sdkstate.State(Active), generateActivePrompt).
		WithPrompt(sdkstate.State(Summarizing), generateSummarizingPrompt).
		WithPrompt(sdkstate.State(Finalizing), generateFinalizingPrompt)
}

// generateOrchestratorPrompt generates the orchestrator-level prompt for exploration projects.
// This explains how the exploration project type works and how to coordinate work through phases.
func generateOrchestratorPrompt(p *state.Project) string {
	prompt, err := templates.Render(templatesFS, "templates/orchestrator.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering orchestrator prompt: %v", err)
	}
	return prompt
}

// generateActivePrompt generates the prompt for the Active state.
// Focus: Identify research topics, investigate, document findings.
func generateActivePrompt(p *state.Project) string {
	var buf strings.Builder

	// Project header
	buf.WriteString(fmt.Sprintf("# Exploration: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Current state
	buf.WriteString("## Current State: Active Research\n\n")

	// Research topics summary
	phase, exists := p.Phases["exploration"]
	if !exists {
		return "Error: exploration phase not found"
	}

	if len(phase.Tasks) == 0 {
		buf.WriteString("No research topics identified yet.\n\n")
		buf.WriteString("**Next steps**: Create research topics to investigate.\n\n")
	} else {
		// Count task statuses
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

		buf.WriteString(fmt.Sprintf("### Research Topics (%d total)\n\n", len(phase.Tasks)))
		buf.WriteString(fmt.Sprintf("- Pending: %d\n", pending))
		buf.WriteString(fmt.Sprintf("- In Progress: %d\n", inProgress))
		buf.WriteString(fmt.Sprintf("- Completed: %d\n", completed))
		buf.WriteString(fmt.Sprintf("- Abandoned: %d\n\n", abandoned))

		// List topics
		for _, task := range phase.Tasks {
			buf.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", task.Id, task.Name, task.Status))
		}
		buf.WriteString("\n")

		// Advancement readiness
		if allTasksResolved(p) {
			buf.WriteString("✓ All research topics resolved!\n\n")
			buf.WriteString("Ready to create summary. Run: `sow project advance`\n\n")
		} else {
			unresolvedCount := countUnresolvedTasks(p)
			buf.WriteString(fmt.Sprintf("**Next steps**: Continue research (%d topics remaining)\n\n", unresolvedCount))
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

// generateSummarizingPrompt generates the prompt for the Summarizing state.
// Focus: Synthesize findings into comprehensive summary document(s).
func generateSummarizingPrompt(p *state.Project) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Exploration: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n\n", p.Branch))

	buf.WriteString("## Current State: Summarizing Findings\n\n")
	buf.WriteString("All research topics are resolved. Create comprehensive summary document(s) synthesizing findings.\n\n")

	// Research completed summary
	phase, exists := p.Phases["exploration"]
	if !exists {
		return "Error: exploration phase not found"
	}

	completed := []projschema.TaskState{}
	abandoned := []projschema.TaskState{}

	for _, task := range phase.Tasks {
		if task.Status == "completed" {
			completed = append(completed, task)
		} else if task.Status == "abandoned" {
			abandoned = append(abandoned, task)
		}
	}

	buf.WriteString(fmt.Sprintf("### Completed Topics: %d\n\n", len(completed)))
	for _, task := range completed {
		buf.WriteString(fmt.Sprintf("- %s\n", task.Name))
	}
	buf.WriteString("\n")

	if len(abandoned) > 0 {
		buf.WriteString(fmt.Sprintf("### Abandoned Topics: %d\n\n", len(abandoned)))
		for _, task := range abandoned {
			buf.WriteString(fmt.Sprintf("- %s\n", task.Name))
		}
		buf.WriteString("\n")
	}

	// Summary artifacts status
	summaries := []projschema.ArtifactState{}
	for _, artifact := range phase.Outputs {
		if artifact.Type == "summary" {
			summaries = append(summaries, artifact)
		}
	}

	buf.WriteString("### Summary Artifacts\n\n")
	if len(summaries) == 0 {
		buf.WriteString("No summaries created yet.\n\n")
		buf.WriteString("**Next steps**: Create summary document(s) synthesizing findings.\n\n")
	} else {
		approvedCount := 0
		for _, s := range summaries {
			if s.Approved {
				approvedCount++
			}
		}

		buf.WriteString(fmt.Sprintf("Total: %d | Approved: %d\n\n", len(summaries), approvedCount))

		for _, s := range summaries {
			status := "Pending approval"
			if s.Approved {
				status = "✓ Approved"
			}
			buf.WriteString(fmt.Sprintf("- %s (%s)\n", s.Path, status))
		}
		buf.WriteString("\n")

		// Advancement readiness
		if allSummariesApproved(p) {
			buf.WriteString("✓ All summaries approved!\n\n")
			buf.WriteString("Ready to finalize. Run: `sow project advance`\n\n")
		} else {
			unapprovedCount := countUnapprovedSummaries(p)
			buf.WriteString(fmt.Sprintf("**Next steps**: Review and approve %d summary document(s)\n\n", unapprovedCount))
		}
	}

	// Render additional guidance from template
	guidance, err := templates.Render(templatesFS, "templates/summarizing.md", p)
	if err != nil {
		return buf.String() + fmt.Sprintf("\nError rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// generateFinalizingPrompt generates the prompt for the Finalizing state.
// Focus: Move artifacts to permanent location, create PR, cleanup.
func generateFinalizingPrompt(p *state.Project) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Exploration: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n\n", p.Branch))

	buf.WriteString("## Current State: Finalizing\n\n")
	buf.WriteString("Summary approved. Finalizing exploration by moving artifacts, creating PR, and cleaning up.\n\n")

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
		buf.WriteString("Ready to complete exploration. Run: `sow project advance`\n\n")
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
- Can use embedded templates for complex prompts (like standard project)
- Provide contextual guidance based on current state

## Task Lifecycle

### Tasks Represent Research Topics

Each task is a research topic/area to investigate:
- **Task name**: Research topic (e.g., "OAuth 2.0 flows", "JWT token structure")
- **Task description** (via input artifact): Scope of investigation, questions to answer
- **Task status**: Current state of research on this topic
- **Task outputs** (optional): Artifacts created during this topic's research

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

**Note**: `needs_review` status is not used in exploration (no formal review cycle for individual topics).

### Task Operations by State

#### Active State

**Operations allowed**:
- ✅ Create new tasks (research topics)
- ✅ Update task status
- ✅ Add task outputs (research findings)
- ✅ Abandon tasks if needed

**Rationale**: Active research is flexible - discover new topics anytime

#### Summarizing State

**Operations restricted**:
- ❌ Cannot create new tasks (research must be complete)
- ❌ Cannot update task status (tasks are read-only)

**Rationale**:
- Summarizing means research is complete
- Adding tasks while summarizing indicates premature transition
- No backward transitions supported (keep architecture simple)

**Edge case**: User realizes more research needed while summarizing
- **Solution**: This indicates advancing to Summarizing was premature
- **Workaround**: Abandon current exploration, restart on new branch
- **Future**: Could add backward transition if this becomes common problem

## Artifact Management

### Artifact Types

Exploration uses two artifact categories:

**1. Research Findings** (during Active state)
- Files created during research (notes, diagrams, code samples)
- Added to `exploration` phase outputs
- **Do not require approval** (`approved = false` or not explicitly approved)
- Type: `"findings"` (or other descriptive types)

**2. Summary Artifacts** (during Summarizing state)
- One or more documents synthesizing findings
- Created by orchestrator (e.g., `summary.md`, `detailed-findings.md`)
- **All summaries require approval** (`approved = true`)
- Type: `"summary"`
- **Multiple summaries allowed** to preserve valuable detail

### Summary Artifact Requirements

**Location**: `.sow/project/phases/exploration/` (managed by orchestrator)

**Typical structure** (orchestrator can create multiple):
- `summary.md` - High-level overview and key findings (REQUIRED if multiple docs)
- `detailed-findings.md` - Comprehensive details per topic (preserves depth)
- `recommendations.md` - Actionable next steps
- `references.md` - External resources and citations

**Rationale for multiple summaries**:
- Forcing everything into single document often loses valuable detail
- Different audiences need different levels of detail
- Separating concerns (findings vs recommendations) improves clarity
- Preserves comprehensive research while maintaining navigability

**Structure requirement**:
- **If multiple summary artifacts**: `summary.md` **must** exist and serve as overview/ToC
- **If single summary**: Can use any filename (no ToC requirement)
- Validated during finalization phase

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

### Finalization Target

**Single summary**:
```
.sow/knowledge/explorations/quick-spike-2025-11.md
```

**Multiple summaries** (requires summary.md as ToC):
```
.sow/knowledge/explorations/auth-approaches-2025-11/
├── summary.md                  # REQUIRED: Overview and ToC
├── detailed-findings.md        # Comprehensive research
└── recommendations.md          # Actionable next steps
```

## Metadata Schemas

**File**: `cli/internal/projects/exploration/metadata.go`

```go
package exploration

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/exploration_metadata.cue
var explorationMetadataSchema string

//go:embed cue/finalization_metadata.cue
var finalizationMetadataSchema string
```

**File**: `cli/internal/projects/exploration/cue/exploration_metadata.cue`

```cue
package exploration

// Metadata schema for exploration phase
{
	// No required metadata for exploration phase
	// Optional metadata can be added as needed
}
```

**File**: `cli/internal/projects/exploration/cue/finalization_metadata.cue`

```cue
package exploration

// Metadata schema for finalization phase
{
	// pr_url: Optional URL of created pull request
	pr_url?: string

	// project_deleted: Flag indicating .sow/project/ has been deleted
	project_deleted?: bool
}
```

## Orchestrator Workflow

### Active State Workflow

1. **Initialize exploration** (via `sow project new` on `explore/*` branch)
2. **Identify research topics**:
   - Orchestrator helps user brainstorm topics
   - Creates task for each topic via SDK task creation
   - Topics start in `pending` status
3. **Research topics**:
   - Update status: task transitions to `in_progress`
   - Create findings artifacts: added to exploration phase outputs
   - Optionally link findings to tasks via task outputs
4. **Complete/abandon topics**:
   - Complete: task transitions to `completed`
   - Abandon: task transitions to `abandoned`
5. **Discover new topics** (dynamic):
   - Can add topics anytime during Active state
   - No limit on topic count
6. **When all resolved**:
   - Guard passes: `allTasksResolved(p) = true`
   - Orchestrator prompts user to advance
   - User/orchestrator runs: `sow project advance`

### Summarizing State Workflow

1. **Transition to Summarizing**: `sow project advance` from Active
2. **Create summary documents**:
   - Orchestrator synthesizes findings from completed topics
   - Creates one or more summary documents as phase outputs
   - Multiple documents preserve valuable detail
3. **Approve summaries**:
   - Set `approved = true` on each summary artifact
   - **All summaries must be approved** before proceeding
4. **Validate summary structure** (if multiple docs):
   - Verify `summary.md` exists as overview/ToC
   - If missing: orchestrator must create it before approval
5. **Complete exploration**: `sow project advance`

### Finalizing State Workflow

1. **Transition to Finalizing**: `sow project advance` from Summarizing
2. **Automatic summary structure validation**:
   - Count approved summary artifacts in exploration phase
   - **If multiple summaries**: Verify `summary.md` exists
     - If missing: **Fail with error** - must create summary.md
     - If exists: Proceed with finalization
   - **If single summary**: No ToC requirement
3. **Create finalization tasks** (orchestrator):
   - Determine target structure (file vs folder)
   - Create descriptive folder if needed (e.g., `auth-approaches-2025-11`)
   - Move all summary artifacts to `.sow/knowledge/explorations/`
   - Create PR with link to exploration location
   - Delete `.sow/project/` directory
4. **Execute tasks**:
   - Orchestrator completes each task sequentially
   - Updates task status to `completed`
5. **Complete finalization**: `sow project advance`
6. **Exploration finished**: State machine reaches Completed

## Testing Strategy

### Unit Tests

**Schema validation**:
- ProjectState validates correctly with type="exploration"
- Phase status constraints enforced
- State machine state constraints enforced

**Guards** (`guards_test.go`):
- `allTasksResolved` returns false with pending tasks
- `allTasksResolved` returns true when all completed/abandoned
- `allSummariesApproved` requires at least one summary, all approved
- `allFinalizationTasksComplete` validates all tasks completed

**Configuration building** (`exploration_test.go`):
- Builder creates valid ProjectTypeConfig
- All transitions configured correctly
- Guards bound via closures
- Prompts registered for all states

### Integration Tests

**Full workflow - single summary** (`integration_test.go`):
1. Create exploration project on `explore/test` branch
2. Add 3 research topics
3. Complete all topics
4. Advance to Summarizing
5. Create and approve single summary
6. Advance to Finalizing
7. Complete finalization
8. Verify single file moved to `.sow/knowledge/explorations/`

**Full workflow - multiple summaries**:
1. Create exploration project
2. Add and complete research topics
3. Advance to Summarizing
4. Create multiple summaries (including summary.md as ToC)
5. Approve all summaries
6. Advance to Finalizing (validates structure)
7. Complete finalization
8. Verify folder created with all summaries

**Edge cases**:
- Advance with unresolved tasks (should fail guard)
- Advance without summary approval (should fail guard)
- Multiple summaries without summary.md (should fail during finalization)

### SDK Integration Tests

**State machine integration**:
- BuildMachine creates valid state machine
- Guards prevent invalid transitions
- Events fire correctly
- Prompts generate for all states

**Registry integration**:
- `state.Register("exploration", config)` succeeds
- `state.Get("exploration")` returns correct config
- Config validates project state correctly

## Migration Notes

No migration from old exploration mode - users restart active explorations.

Exploration sessions are typically short (days, not weeks), so breaking change acceptable.

## Success Criteria

1. ✅ Can create exploration project on `explore/*` branch
2. ✅ Can add/update research topics dynamically during Active state
3. ✅ Cannot add topics in Summarizing state (enforced by SDK)
4. ✅ Can advance to Summarizing when all tasks resolved
5. ✅ Can create and approve multiple summary artifacts
6. ✅ Can advance to Finalizing when all summaries approved
7. ✅ Multiple summaries without `summary.md` fail validation with clear error
8. ✅ Single summary can be finalized without ToC requirement
9. ✅ Finalization moves artifacts to appropriate structure (file or folder)
10. ✅ Finalization creates PR with link to exploration location
11. ✅ Zero-context resumability works (project state on disk)
12. ✅ Prompts provide clear guidance at each state
13. ✅ All guards prevent invalid transitions
14. ✅ SDK configuration is declarative and testable

## Future Enhancements

**Not in scope for initial implementation** (can revisit if needed):

- Backward transitions (Summarizing → Active if more research needed)
- Task templates for common exploration types
- Multi-user collaboration on explorations
- Exploration templates (security, architecture, technology evaluation)
- Linking multiple explorations (series of related investigations)

## References

- **Core Design**: [Modes to Project Types](core-design.md)
- **Project SDK**: `cli/internal/sdks/project/` (builder, config, machine)
- **Standard Project**: `cli/internal/projects/standard/` (reference implementation)
- **State Machine SDK**: `cli/internal/sdks/state/` (underlying state machine)
