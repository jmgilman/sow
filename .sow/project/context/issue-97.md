# Issue #97: Agent System Core

**URL**: https://github.com/jmgilman/sow/issues/97
**State**: OPEN

## Description

# Work Unit 001: Agent System Core

## Behavioral Goal

As a sow developer, I need a lightweight agent definition system so that agent roles (implementer, architect, reviewer, etc.) can be defined as simple data structures with embedded prompt templates, enabling the executor system to spawn any agent with appropriate context.

## Scope

### In Scope
- Agent struct definition with name, description, capabilities, promptPath
- Agent registry for looking up agents by name
- Embedded prompt templates using `go:embed`
- Migration of existing agent prompts from `.claude/agents/` to embedded templates
- `StandardAgents()` function returning all built-in agents
- Unit tests for registry and template loading

### Out of Scope
- Executor interface (work unit 002)
- CLI commands (work unit 003)
- User-extensible custom agents (deferred to v2)

## Existing Code Context

This work unit creates a new package. Key integration points:

**Prompt embedding pattern** (`cli/internal/prompts/prompts.go:20-32`):
```go
//go:embed templates
var FS embed.FS
```
Follow this pattern for agent template embedding.

**Existing agent definitions** (`.claude/agents/*.md`):
- `implementer.md` - TDD-based code implementation
- `planner.md` - Task planning and breakdown
- `researcher.md` - Research and exploration
- `reviewer.md` - Code review
- `decomposer.md` - Feature decomposition

These use YAML frontmatter format that should be preserved in embedded templates.

## Documentation Context

**Design doc** (`.sow/knowledge/designs/multi-agent-architecture.md`):
- Section "Agent Definition" (lines 138-189) specifies the Agent struct
- Section "Embedded Prompt Templates" (lines 225-293) specifies template structure
- Section "Agent Registry" (lines 199-221) specifies registry interface

## File Structure

Create new package at `cli/internal/agents/`:

```
cli/internal/agents/
├── agents.go           # Agent struct + StandardAgents()
├── agents_test.go      # Unit tests for agent definitions
├── registry.go         # AgentRegistry implementation
├── registry_test.go    # Unit tests for registry
├── templates.go        # Template loading with go:embed
├── templates_test.go   # Unit tests for template loading
└── templates/
    ├── implementer.md
    ├── planner.md
    ├── researcher.md
    ├── reviewer.md
    ├── decomposer.md
    └── architect.md    # New - referenced in design doc
```

## Implementation Approach

### Agent Struct

```go
// Agent represents a role in the sow system
type Agent struct {
    // Name is the agent identifier (e.g., "implementer", "architect")
    Name string

    // Description explains what this agent does
    Description string

    // Capabilities describes what the agent must be able to do (prose)
    Capabilities string

    // PromptPath is the path to the embedded prompt template
    // Relative to templates/ directory
    PromptPath string
}
```

### Standard Agents

Define all agents in `agents.go`:

```go
var (
    Implementer = &Agent{
        Name:         "implementer",
        Description:  "Code implementation using Test-Driven Development",
        Capabilities: "Must be able to read/write files, execute shell commands, search codebase",
        PromptPath:   "implementer.md",
    }
    // ... other agents
)

func StandardAgents() []*Agent {
    return []*Agent{Implementer, Architect, Reviewer, Planner, Researcher, Decomposer}
}
```

### Registry

```go
type AgentRegistry struct {
    agents map[string]*Agent
}

func NewAgentRegistry() *AgentRegistry
func (r *AgentRegistry) Register(agent *Agent)
func (r *AgentRegistry) Get(name string) (*Agent, error)
func (r *AgentRegistry) List() []*Agent
```

### Template Loading

```go
//go:embed templates/*
var templatesFS embed.FS

func LoadPrompt(promptPath string) (string, error) {
    data, err := fs.ReadFile(templatesFS, "templates/"+promptPath)
    if err != nil {
        return "", fmt.Errorf("failed to load prompt %s: %w", promptPath, err)
    }
    return string(data), nil
}
```

## Dependencies

None - this is the foundational work unit.

## Acceptance Criteria

1. **Agent struct defined** with all fields from design doc
2. **6+ standard agents** defined (implementer, architect, reviewer, planner, researcher, decomposer)
3. **AgentRegistry** supports Register, Get, List operations
4. **Template loading** works via `LoadPrompt(path)`
5. **Templates migrated** from `.claude/agents/` to embedded templates
6. **Unit tests** cover:
   - All standard agents are registered
   - Get returns correct agent by name
   - Get returns error for unknown agent
   - List returns all agents
   - LoadPrompt loads template content
   - LoadPrompt returns error for missing template
7. **No breaking changes** to existing `.claude/agents/` files (can coexist during transition)

## Testing Strategy

- Unit tests for AgentRegistry operations
- Unit tests for template loading
- Table-driven tests for all standard agents
- Error case coverage (unknown agent, missing template)
