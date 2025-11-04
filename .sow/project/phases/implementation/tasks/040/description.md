# Task 040: Create Prompt Functions

# Task 040: Create Prompt Functions

## Overview

Extract and refactor prompt generation logic into simple functions matching the SDK's `PromptGenerator` signature: `func(*state.Project) string`. These functions build contextual prompts for each state by combining template rendering with dynamic components.

**Critical**: Use SDK types and eliminate dependencies on old `StandardPromptGenerator` struct and `statechart.PromptComponents`.

## Context

**Design Reference**: `.sow/knowledge/designs/project-sdk-implementation.md` (lines 762-777) shows WithPrompt() usage

**SDK Signature**: `type PromptGenerator func(*state.Project) string` (see `cli/internal/sdks/project/types.go`)

**Existing Reference**: `cli/internal/project/standard/prompts.go` contains the dynamic logic to preserve (git status, task summaries, iteration tracking, etc.)

**Architecture Change**: Old system used generator struct with `GeneratePrompt(state, projectState)` method. New SDK uses simple functions: `generatePrompt(project) string`.

## Requirements

### Prompt Functions File

Create `cli/internal/projects/standard/prompts.go` with:

1. **Template Registry Setup** (preserve from old implementation):
```go
package standard

import (
	"embed"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// StatePromptID uniquely identifies prompt templates
type StatePromptID string

const (
	PromptPlanningActive          StatePromptID = "planning_active"
	PromptImplementationPlanning  StatePromptID = "implementation_planning"
	PromptImplementationExecuting StatePromptID = "implementation_executing"
	PromptReviewActive            StatePromptID = "review_active"
	PromptFinalizeDocumentation   StatePromptID = "finalize_documentation"
	PromptFinalizeChecks          StatePromptID = "finalize_checks"
	PromptFinalizeDelete          StatePromptID = "finalize_delete"
)

//go:embed templates/*.md
var templatesFS embed.FS

var standardRegistry *prompts.Registry[StatePromptID]

func init() {
	standardRegistry = prompts.NewRegistry[StatePromptID]()

	// Map IDs to template files
	templates := map[StatePromptID]string{
		PromptPlanningActive:          "templates/planning_active.md",
		PromptImplementationPlanning:  "templates/implementation_planning.md",
		PromptImplementationExecuting: "templates/implementation_executing.md",
		PromptReviewActive:            "templates/review_active.md",
		PromptFinalizeDocumentation:   "templates/finalize_documentation.md",
		PromptFinalizeChecks:          "templates/finalize_checks.md",
		PromptFinalizeDelete:          "templates/finalize_delete.md",
	}

	for id, path := range templates {
		if err := standardRegistry.RegisterFromFS(templatesFS, id, path); err != nil {
			panic(fmt.Sprintf("failed to register template %s: %v", id, err))
		}
	}
}
```

2. **Seven Prompt Functions** (one per non-NoProject state):

Each function signature: `func(p *state.Project) string`

**Required Functions**:
- `generatePlanningPrompt(p *state.Project) string`
- `generateImplementationPlanningPrompt(p *state.Project) string`
- `generateImplementationExecutingPrompt(p *state.Project) string`
- `generateReviewPrompt(p *state.Project) string`
- `generateFinalizeDocumentationPrompt(p *state.Project) string`
- `generateFinalizeChecksPrompt(p *state.Project) string`
- `generateFinalizeDeletePrompt(p *state.Project) string`

3. **Dynamic Components to Preserve**:

Each function should build prompts with:
- **Project header**: name, branch, description
- **Git status**: uncommitted changes (via `p.Context().Git()`)
- **Task summaries**: count by status (completed, in progress, pending)
- **Iteration tracking**: review cycle number from metadata
- **Recent commits**: conditional on task completion
- **Artifact status**: show approval state

4. **Helper Functions** (preserve from old implementation):
- `findPreviousReviewArtifact(p *state.Project, iteration int) *state.Artifact`
- `extractReviewAssessment(artifact *state.Artifact) string`

### Adaptation Requirements

**Old Type → New Type Mappings**:
- `*schemas.ProjectState` → `*state.Project`
- Direct field access → Collection methods: `p.Phases.Get("planning")`
- `projectState.Phases.Planning.Artifacts` → `phase.Outputs` (outputs are artifacts)
- `projectState.Phases.Implementation.Tasks` → `phase.Tasks`
- Metadata access: `phase.Metadata["key"]` with type assertions

**Template Rendering**:
```go
templateCtx := &prompts.StatechartContext{
	State:        string(PlanningActive),
	ProjectState: p.ProjectState, // Embed the CUE-generated state
}
guidance, err := standardRegistry.Render(PromptPlanningActive, templateCtx)
```

**Example Function Structure**:
```go
func generatePlanningPrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Add git status (handle errors gracefully)
	if ctx := p.Context(); ctx != nil {
		git := ctx.Git()
		hasChanges, err := git.HasUncommittedChanges()
		if err == nil {
			buf.WriteString("## Git Status\n\n")
			if hasChanges {
				buf.WriteString("Has uncommitted changes.\n\n")
			} else {
				buf.WriteString("Working tree clean.\n\n")
			}
		}
	}

	// Render template
	templateCtx := &prompts.StatechartContext{
		State:        string(PlanningActive),
		ProjectState: p.ProjectState,
	}
	guidance, err := standardRegistry.Render(PromptPlanningActive, templateCtx)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	// Show artifacts if any
	phase, _ := p.Phases.Get("planning")
	if phase != nil && len(phase.Outputs) > 0 {
		buf.WriteString("\n## Planning Artifacts\n\n")
		for _, artifact := range phase.Outputs {
			status := "pending"
			if artifact.Approved {
				status = "approved"
			}
			buf.WriteString(fmt.Sprintf("- %s (%s)\n", artifact.Path, status))
		}
	}

	return buf.String()
}
```

## Acceptance Criteria

- [ ] File `cli/internal/projects/standard/prompts.go` created
- [ ] Template registry setup with `//go:embed` and `init()`
- [ ] 7 prompt functions with signature `func(p *state.Project) string`
- [ ] Each function builds prompts with dynamic components:
  - [ ] Project header (name, branch, description)
  - [ ] Git status (when available)
  - [ ] Task summaries (for implementation/review states)
  - [ ] Iteration tracking (for review state)
  - [ ] Artifact displays (for relevant states)
- [ ] Helper functions for review artifact lookup
- [ ] Template rendering using local registry
- [ ] SDK type usage (no `statechart.PromptComponents` references)
- [ ] Graceful error handling (prompts never panic, show errors as strings)
- [ ] File compiles: `go build ./cli/internal/projects/standard/...`
- [ ] Old package untouched

## Validation Commands

```bash
# Verify file exists
ls cli/internal/projects/standard/prompts.go

# Verify compilation
go build ./cli/internal/projects/standard/...

# Verify no old dependencies
! grep "StandardPromptGenerator" cli/internal/projects/standard/prompts.go
! grep "statechart.PromptComponents" cli/internal/projects/standard/prompts.go

# Verify SDK type usage
grep "state.Project" cli/internal/projects/standard/prompts.go

# Verify template embedding
grep "//go:embed templates/" cli/internal/projects/standard/prompts.go

# Verify old package untouched
git diff cli/internal/project/standard/
```

## Dependencies

- Task 010 (templates copied)
- Task 020 (states defined for constants)

## Standards

- Functions match SDK signature exactly: `func(*state.Project) string`
- Use descriptive function names: `generateXXXPrompt`
- Handle errors gracefully (return error messages as strings, don't panic)
- Preserve all existing dynamic logic from old implementation
- Use SDK collection pattern for phase access
- Comments explain what each section of the prompt provides

## Notes

**Simplification**: Old system had generator struct with state routing. New system uses direct function calls - much simpler.

**Template Preservation**: Templates (copied in Task 010) contain the bulk of prompt content. Functions add dynamic context around them.

**Context Access**: `p.Context()` provides access to git/GitHub operations. Check for nil before using.

**Phase Access Pattern**:
```go
phase, err := p.Phases.Get("implementation")
if err != nil {
	// Phase doesn't exist, handle gracefully
}
```

**These functions will be registered in Task 6** via `WithPrompt(state, function)`.
