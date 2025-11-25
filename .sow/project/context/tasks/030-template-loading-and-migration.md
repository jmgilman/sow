# Template Loading and Migration

## Context

This task is part of **Work Unit 001: Agent System Core**, implementing a lightweight agent definition system for sow.

Agent prompts are embedded in the sow binary using Go's `embed` package. This allows agent templates to be bundled with the CLI and versioned alongside the code. The templates are loaded at runtime via the `LoadPrompt()` function, which reads from the embedded filesystem.

This task also involves migrating the existing agent prompts from `.claude/agents/` to the new embedded templates location. The existing prompts use YAML frontmatter format that should be preserved.

**Tasks 010 and 020 should be completed first** - this task depends on the Agent struct being available to verify PromptPath references.

## Requirements

### Embedded Templates Setup

Create `templates.go` in the `cli/internal/agents/` package:

```go
import (
    "embed"
    "fmt"
    "io/fs"
)

//go:embed templates/*
var templatesFS embed.FS
```

### LoadPrompt Function

```go
// LoadPrompt loads an agent's prompt template from the embedded filesystem.
// The promptPath is relative to the templates/ directory.
//
// Example:
//   content, err := LoadPrompt("implementer.md")
func LoadPrompt(promptPath string) (string, error) {
    data, err := fs.ReadFile(templatesFS, "templates/"+promptPath)
    if err != nil {
        return "", fmt.Errorf("failed to load prompt %s: %w", promptPath, err)
    }
    return string(data), nil
}
```

### Template Files to Create

Create the following files in `cli/internal/agents/templates/`:

1. **implementer.md** - Migrate from `.claude/agents/implementer.md`
2. **planner.md** - Migrate from `.claude/agents/planner.md`
3. **reviewer.md** - Migrate from `.claude/agents/reviewer.md`
4. **researcher.md** - Migrate from `.claude/agents/researcher.md`
5. **decomposer.md** - Migrate from `.claude/agents/decomposer.md`
6. **architect.md** - NEW template (referenced in design doc but not yet existing)

### Template Content Requirements

**For migrated templates (implementer, planner, reviewer, researcher, decomposer)**:
- Copy the content from `.claude/agents/*.md` files
- Remove the YAML frontmatter (the `---` delimited header with name, description, tools, model)
- Keep only the markdown body content

**For architect.md (new template)**:
Create a new template following the same pattern as other agent templates:

```markdown
You are a software architect agent. Your instructions are provided dynamically via the sow prompt system.

## Initialization

Run this command immediately to load your base instructions:

```bash
sow prompt guidance/architect/base
```

The base prompt will guide you through:
1. Reading project context and requirements
2. Understanding the existing architecture
3. Making design decisions with proper documentation
4. Creating or updating architecture documentation

## Context Location

Your task context is located at:

```
.sow/project/phases/{phase}/tasks/{task-id}/
├── state.yaml        # Task metadata, iteration, references
├── description.md    # Requirements and acceptance criteria
├── log.md            # Your action log (append here)
└── feedback/         # Corrections from previous iterations (if any)
    └── {id}.md
```

Start by reading state.yaml to get your task ID and iteration number.
```

## Acceptance Criteria

1. **templates.go created** with `go:embed` directive and `LoadPrompt()` function
2. **templates/ directory created** under `cli/internal/agents/`
3. **6 template files created**:
   - `templates/implementer.md`
   - `templates/planner.md`
   - `templates/reviewer.md`
   - `templates/researcher.md`
   - `templates/decomposer.md`
   - `templates/architect.md`
4. **Template content migrated** from `.claude/agents/` (without YAML frontmatter)
5. **LoadPrompt() works** for all standard agent templates
6. **Unit tests verify**:
   - LoadPrompt() successfully loads each template
   - LoadPrompt() returns error for missing template
   - Loaded content is non-empty
   - All standard agent PromptPaths can be loaded
7. **Original files preserved** - `.claude/agents/*.md` files remain unchanged (coexistence during transition)

## Technical Details

### Package Structure Update

```
cli/internal/agents/
├── agents.go           # Agent struct + StandardAgents() (from task 010)
├── agents_test.go      # Tests for agents.go (from task 010)
├── registry.go         # AgentRegistry (from task 020)
├── registry_test.go    # Tests for registry (from task 020)
├── templates.go        # Template loading with go:embed (this task)
├── templates_test.go   # Tests for template loading (this task)
└── templates/          # Embedded template files (this task)
    ├── implementer.md
    ├── planner.md
    ├── reviewer.md
    ├── researcher.md
    ├── decomposer.md
    └── architect.md
```

### Embed Pattern

Follow the pattern established in `cli/internal/prompts/prompts.go`:

```go
//go:embed templates
var FS embed.FS
```

But for the agents package, use a more specific name:

```go
//go:embed templates/*
var templatesFS embed.FS
```

### Error Handling

Use `fmt.Errorf` with `%w` for error wrapping:
```go
return "", fmt.Errorf("failed to load prompt %s: %w", promptPath, err)
```

## Relevant Inputs

These files provide context for this task:

- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/cli/internal/prompts/prompts.go` - Reference implementation of embed pattern (lines 20-32)
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/implementer.md` - Source for implementer template migration
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/planner.md` - Source for planner template migration
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/reviewer.md` - Source for reviewer template migration
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/researcher.md` - Source for researcher template migration
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/decomposer.md` - Source for decomposer template migration
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.sow/knowledge/designs/multi-agent-architecture.md` - Design doc with template structure (lines 225-293)

## Examples

### Expected LoadPrompt Usage

```go
func main() {
    // Load implementer prompt
    content, err := agents.LoadPrompt("implementer.md")
    if err != nil {
        log.Fatalf("Failed to load prompt: %v", err)
    }
    fmt.Println(content)
}
```

### Expected Test Pattern

```go
func TestLoadPrompt(t *testing.T) {
    tests := []struct {
        name       string
        promptPath string
        wantError  bool
    }{
        {
            name:       "implementer exists",
            promptPath: "implementer.md",
            wantError:  false,
        },
        {
            name:       "architect exists",
            promptPath: "architect.md",
            wantError:  false,
        },
        {
            name:       "missing template",
            promptPath: "nonexistent.md",
            wantError:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            content, err := LoadPrompt(tt.promptPath)
            if (err != nil) != tt.wantError {
                t.Errorf("LoadPrompt(%q) error = %v, wantError %v", tt.promptPath, err, tt.wantError)
                return
            }
            if !tt.wantError && content == "" {
                t.Errorf("LoadPrompt(%q) returned empty content", tt.promptPath)
            }
        })
    }
}
```

### Expected Integration Test with StandardAgents

```go
func TestAllStandardAgentPromptsCanBeLoaded(t *testing.T) {
    for _, agent := range StandardAgents() {
        t.Run(agent.Name, func(t *testing.T) {
            content, err := LoadPrompt(agent.PromptPath)
            if err != nil {
                t.Errorf("LoadPrompt(%q) for agent %q failed: %v", agent.PromptPath, agent.Name, err)
            }
            if content == "" {
                t.Errorf("LoadPrompt(%q) for agent %q returned empty content", agent.PromptPath, agent.Name)
            }
        })
    }
}
```

### YAML Frontmatter Removal Example

**Before (from `.claude/agents/implementer.md`)**:
```markdown
---
name: implementer
description: Code implementation using Test-Driven Development
tools: Read, Write, Edit, Grep, Glob, Bash
model: inherit
---

You are a software implementer agent...
```

**After (in `cli/internal/agents/templates/implementer.md`)**:
```markdown
You are a software implementer agent...
```

## Dependencies

- **Task 010** - Agent struct and StandardAgents() must be implemented for integration test
- **Task 020** - Not strictly required, but AgentRegistry is useful for comprehensive testing

## Constraints

- **Preserve original files** - The `.claude/agents/*.md` files must NOT be modified or deleted (they may still be used during transition period)
- **Remove frontmatter only** - When migrating, only remove the YAML frontmatter; keep all other content exactly as-is
- **No template rendering** - This task only loads raw template content; rendering/interpolation is handled elsewhere (by the executor system in future work units)
- **UTF-8 content** - Templates are expected to be UTF-8 encoded text files
- **Error wrapping** - Use `%w` verb in `fmt.Errorf` for proper error chain propagation
