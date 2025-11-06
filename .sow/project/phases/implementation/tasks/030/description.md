# Task 030: Prompt Templates and Generators

## Context

This task implements the prompt generation system for the design project type. Prompts provide contextual guidance to the orchestrator at each state, explaining the current situation and suggesting next steps. They are crucial for zero-context resumability - allowing any orchestrator to understand project status by reading state and prompts.

The design project requires prompts for three contexts:
1. **Orchestrator-level prompt** - Explains how design projects work overall
2. **Active state prompt** - Guides document planning, drafting, and review
3. **Finalizing state prompt** - Guides finalization tasks and completion

Prompts combine dynamic state information (task counts, statuses) with static guidance from embedded markdown templates. This creates rich, informative prompts that adapt to project state while maintaining consistent messaging.

This task builds on Tasks 010 and 020 by using states, events, and guard functions to generate intelligent, state-aware prompts.

## Requirements

### Template Files

Create markdown template files in `templates/` directory:

1. **templates/orchestrator.md**
   - High-level explanation of design project workflow
   - When to use design projects
   - How phases progress (Active → Finalizing → Completed)
   - Task-based document tracking approach
   - Review workflow and auto-approval
   - Key commands and operations

2. **templates/active.md**
   - Guidance for Active state operations
   - How to plan document tasks
   - How to draft documents
   - How to use review workflow (needs_review status)
   - How to approve documents (auto-approval on completion)
   - When ready to advance to finalization

3. **templates/finalizing.md**
   - Guidance for Finalizing state operations
   - What finalization involves (move docs, create PR, cleanup)
   - How to complete finalization tasks
   - When ready to advance to Completed

### Prompt Generator Functions

Implement prompt generators in `prompts.go`:

1. **configurePrompts(builder *project.ProjectTypeConfigBuilder)**
   - Registers all prompt generators with builder
   - Chains calls to WithOrchestratorPrompt and WithPrompt
   - Returns builder for chaining

2. **generateOrchestratorPrompt(p *state.Project) string**
   - Renders orchestrator template with project context
   - Provides workflow overview
   - Returns formatted prompt string

3. **generateActivePrompt(p *state.Project) string**
   - Builds dynamic prompt showing:
     - Project name, branch, description
     - Current state: "Active Design"
     - Design inputs (if any)
     - Document task list with statuses
     - Task counts (pending, in_progress, needs_review, completed, abandoned)
     - Advancement readiness indicator
     - Next steps guidance
   - Appends static guidance from active.md template
   - Returns formatted prompt string

4. **generateFinalizingPrompt(p *state.Project) string**
   - Builds dynamic prompt showing:
     - Project name and branch
     - Current state: "Finalizing"
     - Finalization task list with completion status
     - Advancement readiness indicator
   - Appends static guidance from finalizing.md template
   - Returns formatted prompt string

### Prompt Content Requirements

**Active state prompt must show**:
- Task status icons: `[ ]` pending, `[~]` in_progress, `[?]` needs_review, `[✓]` completed, `[✗]` abandoned
- For each task: ID, name, status, artifact path, document type
- Total task count and breakdown by status
- Input artifacts with descriptions (if any)
- Clear indication when ready to advance
- Unresolved task count when not ready

**Finalizing state prompt must show**:
- Finalization task list with checkboxes
- Clear indication when all tasks complete
- Guidance on final advancement

## Acceptance Criteria

### Functional Requirements

- [ ] `templates/` directory created with three .md files
- [ ] `prompts.go` implements all four required functions
- [ ] Orchestrator prompt explains design workflow clearly
- [ ] Active prompt shows complete task status dynamically
- [ ] Finalizing prompt shows finalization progress
- [ ] All prompts use consistent formatting and style
- [ ] Prompts handle edge cases (no tasks, missing phases)
- [ ] Template rendering errors are handled gracefully

### Test Requirements (TDD)

Write tests in `prompts_test.go`:

**configurePrompts tests**:
- [ ] Returns non-nil builder for chaining
- [ ] Registers orchestrator prompt generator
- [ ] Registers Active state prompt generator
- [ ] Registers Finalizing state prompt generator

**generateOrchestratorPrompt tests**:
- [ ] Returns non-empty string
- [ ] Contains key workflow concepts
- [ ] Handles nil project gracefully

**generateActivePrompt tests**:
- [ ] Shows project name and branch
- [ ] Shows "Active Design" state
- [ ] Lists all tasks with correct status icons
- [ ] Shows task counts accurately
- [ ] Shows advancement readiness when guard passes
- [ ] Shows unresolved count when guard fails
- [ ] Handles empty task list
- [ ] Handles missing design phase
- [ ] Shows input artifacts when present
- [ ] Includes static guidance from template

**generateFinalizingPrompt tests**:
- [ ] Shows project name and branch
- [ ] Shows "Finalizing" state
- [ ] Lists finalization tasks with checkboxes
- [ ] Shows advancement readiness when guard passes
- [ ] Handles missing finalization phase
- [ ] Includes static guidance from template

### Code Quality

- [ ] All functions documented with clear descriptions
- [ ] Template files use consistent markdown formatting
- [ ] Prompt generators handle nil/missing data safely
- [ ] Error handling includes context in error messages
- [ ] Tests cover all edge cases
- [ ] Templates are informative and actionable

## Technical Details

### Package Structure

```go
package design

import (
	"embed"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sdks/project/templates"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

//go:embed templates/*.md
var templatesFS embed.FS
```

### Function Signatures

```go
func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder
func generateOrchestratorPrompt(p *state.Project) string
func generateActivePrompt(p *state.Project) string
func generateFinalizingPrompt(p *state.Project) string
```

### Template Rendering

Use the templates utility for rendering:
```go
prompt, err := templates.Render(templatesFS, "templates/orchestrator.md", p)
if err != nil {
    return fmt.Sprintf("Error rendering orchestrator prompt: %v", err)
}
```

### String Building Pattern

Build dynamic prompts using strings.Builder:
```go
var buf strings.Builder
buf.WriteString(fmt.Sprintf("# Design: %s\n", p.Name))
buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
// ... more content
return buf.String()
```

### Task Status Icons

Use consistent icons throughout:
- `[ ]` - pending
- `[~]` - in_progress
- `[?]` - needs_review
- `[✓]` - completed
- `[✗]` - abandoned

### Guard Integration

Check advancement readiness using guards:
```go
if allDocumentsApproved(p) {
    buf.WriteString("✓ All documents approved!\n\n")
    buf.WriteString("Ready to finalize. Run: `sow project advance`\n\n")
} else {
    unresolvedCount := countUnresolvedTasks(p)
    buf.WriteString(fmt.Sprintf("**Next steps**: Continue design work (%d documents remaining)\n\n", unresolvedCount))
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/.sow/knowledge/designs/project-modes/design-design.md` - Prompt specifications (lines 498-703)
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/prompts.go` - Reference prompt implementation
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/templates/active.md` - Reference template style
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/templates/orchestrator.md` - Reference orchestrator template
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/sdks/project/templates/templates.go` - Template rendering utilities

## Examples

### Active Prompt Output Example

```markdown
# Design: authentication-design
Branch: design/auth

## Current State: Active Design

### Design Inputs

Sources informing this design:

- context/exploration-findings.md
  Exploration findings on authentication libraries

### Design Documents

Total: 3 documents
- Pending: 1
- In Progress: 1
- Needs Review: 0
- Completed: 1
- Abandoned: 0

[ ] 010 - Authentication Overview (pending)
[~] 020 - JWT Implementation Details (in_progress)
    Artifact: project/jwt-design.md
    Type: design
[✓] 030 - Session Management (completed)
    Artifact: project/session-design.md
    Type: design

**Next steps**: Continue design work (2 documents remaining)

---

## Guidance: Active Design
...static template content...
```

### Finalizing Prompt Output Example

```markdown
# Design: authentication-design
Branch: design/auth

## Current State: Finalizing

All documents approved. Finalizing design by moving artifacts, creating PR, and cleaning up.

### Finalization Tasks

[✓] Move approved documents to targets
[ ] Create PR with design artifacts
[ ] Delete .sow/project/ directory

---

## Guidance: Finalizing
...static template content...
```

## Dependencies

- Task 010: Core Structure and Constants - Provides state constants
- Task 020: Guard Functions and Helpers - Provides guards and helpers used in prompts

## Constraints

### Error Handling Philosophy

Prompts should never fail completely:
- If template rendering fails, return error message + dynamic content
- If phase missing, show error but continue with other sections
- Always return a string (never empty or nil)

### Formatting Consistency

- Use markdown for all prompts
- Use consistent heading levels (# for title, ## for sections, ### for subsections)
- Use bullet points for lists
- Use code blocks for commands
- Use bold for emphasis on key actions

### Performance

Prompts are generated on-demand:
- Should be fast (< 10ms typically)
- Avoid complex computations
- Cache is handled by caller if needed

### User Experience

Prompts should be:
- Scannable (clear structure, icons for quick status check)
- Actionable (clear next steps, specific commands)
- Contextual (show relevant info for current state)
- Educational (explain what each state means)
