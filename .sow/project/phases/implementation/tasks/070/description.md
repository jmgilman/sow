# Task 070: Implement Prompt Generation

## Context

This task implements prompt generation for the exploration project type. Prompts provide contextual guidance to users and the orchestrator at each state of the workflow.

The exploration project type needs:
1. **Orchestrator prompt** - Explains how exploration projects work (shown at project start/resume)
2. **State-specific prompts** - Provide guidance for each state (Active, Summarizing, Finalizing)

Prompts can be generated programmatically (string building) or use embedded markdown templates. The standard project type uses templates for complex prompts.

## Requirements

### Create Prompt Generation File

Create `cli/internal/projects/exploration/prompts.go` with:

1. **Package declaration and imports**:
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
   ```

2. **configurePrompts function**:
   - Register all prompt generators with builder
   - Orchestrator prompt: `generateOrchestratorPrompt`
   - Active state: `generateActivePrompt`
   - Summarizing state: `generateSummarizingPrompt`
   - Finalizing state: `generateFinalizingPrompt`

3. **generateOrchestratorPrompt function**:
   - Render `templates/orchestrator.md` using templates.Render
   - Handle errors gracefully (return error message string)

4. **generateActivePrompt function**:
   - Build prompt programmatically using strings.Builder
   - Include:
     - Project header (name, branch, description)
     - Current state: "Active Research"
     - Research topics summary (count by status)
     - List all topics with ID, name, status
     - Advancement readiness check (call `allTasksResolved()`)
     - Render additional guidance from `templates/active.md`

5. **generateSummarizingPrompt function**:
   - Build prompt showing summarization phase
   - Include:
     - Project header
     - Current state: "Summarizing Findings"
     - Completed topics list
     - Abandoned topics list (if any)
     - Summary artifacts status (count, approval state)
     - Advancement readiness (call `allSummariesApproved()`)
     - Render guidance from `templates/summarizing.md`

6. **generateFinalizingPrompt function**:
   - Build prompt showing finalization status
   - Include:
     - Project header
     - Current state: "Finalizing"
     - Finalization tasks checklist (with status checkmarks)
     - Advancement readiness (call `allFinalizationTasksComplete()`)
     - Render guidance from `templates/finalizing.md`

### Create Template Directory and Templates

Create directory:
```
cli/internal/projects/exploration/templates/
```

Create minimal template files (detailed content optional for this task):

1. **templates/orchestrator.md**:
   - Explain exploration project workflow
   - Describe 2-phase structure
   - Guide orchestrator on coordinating work

2. **templates/active.md**:
   - Guidance for active research phase
   - How to add/manage research topics
   - When to advance to summarizing

3. **templates/summarizing.md**:
   - Guidance for summarizing phase
   - How to create summary artifacts
   - Approval process
   - Single vs multiple summary requirements

4. **templates/finalizing.md**:
   - Guidance for finalization
   - Artifact movement to `.sow/knowledge/explorations/`
   - PR creation
   - Cleanup steps

## Acceptance Criteria

- [ ] File `prompts.go` created with all required functions
- [ ] Directory `templates/` created
- [ ] All 4 template files created
- [ ] `configurePrompts` registers all prompt generators
- [ ] Orchestrator prompt uses template rendering
- [ ] State prompts build dynamic status information
- [ ] State prompts use helper functions (allTasksResolved, etc.)
- [ ] State prompts include advancement readiness indicators
- [ ] Error handling for template rendering
- [ ] Templates embedded using `//go:embed`
- [ ] Code follows Go formatting standards (gofmt)
- [ ] No compilation errors

## Technical Details

### Prompt Generator Signature

All prompt generators have signature:
```go
func(p *state.Project) string
```

They receive project state and return markdown-formatted guidance string.

### Template Rendering

Use the SDK template renderer:
```go
prompt, err := templates.Render(templatesFS, "templates/file.md", p)
if err != nil {
    return fmt.Sprintf("Error rendering template: %v", err)
}
```

Templates have access to project fields and helper functions like `phase`, `countTasksByStatus`, etc.

### String Building Pattern

For dynamic prompts, use strings.Builder for efficiency:
```go
var buf strings.Builder
buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
buf.WriteString("## Current State\n")
// ... more content
return buf.String()
```

### Status Symbols

Use visual indicators for better UX:
- ✓ for completed/approved
- [ ] for pending
- [✓] for completed checklist items
- Counts: "3 completed, 1 pending"

### Advancement Indicators

Check guard functions and provide clear guidance:
```go
if allTasksResolved(p) {
    buf.WriteString("✓ All research topics resolved!\n")
    buf.WriteString("Ready to create summary. Run: `sow project advance`\n")
} else {
    unresolvedCount := countUnresolvedTasks(p)
    buf.WriteString(fmt.Sprintf("**Next steps**: Continue research (%d topics remaining)\n", unresolvedCount))
}
```

### Template Access to Project

Templates use Go text/template syntax:
```markdown
# Project: {{.Name}}
Branch: {{.Branch}}

{{$exploration := phase . "exploration"}}
{{if hasApprovedOutput $exploration "summary"}}
✓ Summary approved
{{end}}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/prompts.go` - Reference prompt implementation
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/templates/orchestrator.md` - Reference orchestrator template
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/templates/renderer.go` - Template rendering utilities
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/knowledge/designs/project-modes/exploration-design.md` - Design specification (lines 489-743)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/project/context/issue-36.md` - Requirements

## Examples

### Standard Project Prompt Configuration (Reference)

From `cli/internal/projects/standard/prompts.go:252-264`:

```go
func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
    return builder.
        WithOrchestratorPrompt(generateOrchestratorPrompt).
        WithPrompt(sdkstate.State(ImplementationPlanning), generateImplementationPlanningPrompt).
        WithPrompt(sdkstate.State(ImplementationExecuting), generateImplementationExecutingPrompt).
        WithPrompt(sdkstate.State(ReviewActive), generateReviewPrompt).
        WithPrompt(sdkstate.State(FinalizeChecks), generateFinalizeChecksPrompt).
        WithPrompt(sdkstate.State(FinalizePRCreation), generateFinalizePRCreationPrompt).
        WithPrompt(sdkstate.State(FinalizeCleanup), generateFinalizeCleanupPrompt)
}
```

### Template-Based Prompt (Reference)

From `cli/internal/projects/standard/prompts.go:16-25`:

```go
func generateOrchestratorPrompt(p *state.Project) string {
    prompt, err := templates.Render(templatesFS, "templates/orchestrator.md", p)
    if err != nil {
        return fmt.Sprintf("Error rendering orchestrator prompt: %v", err)
    }
    return prompt
}
```

### Programmatic Prompt Building (Reference)

From `cli/internal/projects/standard/prompts.go:67-96`:

```go
func generateImplementationExecutingPrompt(p *state.Project) string {
    var buf strings.Builder

    buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
    buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
    if p.Description != "" {
        buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
    }
    buf.WriteString("\n")

    // Add task summary
    if implPhase, exists := p.Phases["implementation"]; exists && len(implPhase.Tasks) > 0 {
        buf.WriteString(taskSummary(implPhase.Tasks))
        buf.WriteString("\n")
    }

    // Render guidance template
    guidance, err := templates.Render(templatesFS, "templates/implementation_executing.md", p)
    if err != nil {
        return fmt.Sprintf("Error rendering template: %v", err)
    }
    buf.WriteString(guidance)

    return buf.String()
}
```

### Task Status Summary Example

```go
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
```

## Dependencies

- Task 010 (Package structure) - Provides package directory
- Task 020 (States and events) - State constants used in prompt registration
- Task 040 (Guards) - Guard and helper functions used in prompts
- Will be called by SDK when rendering state-specific guidance
- Templates referenced in Task 030 (Phase configuration)

## Constraints

- All prompts must return strings (never nil or panic)
- Error handling must be graceful (return error message as string)
- Prompts should be informative but concise
- Must use markdown formatting for readability
- Template rendering errors should not crash application
- State-specific prompts must accurately reflect current project state
- Must call guard functions to determine advancement readiness
- Template files must be in `templates/` subdirectory for embed to work
