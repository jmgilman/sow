# Agent Struct and Standard Agent Definitions

## Context

This task is part of **Work Unit 001: Agent System Core**, implementing a lightweight agent definition system for sow.

The sow framework is building a multi-agent system where different agent roles (implementer, architect, reviewer, planner, researcher, decomposer) can be defined as simple data structures with embedded prompt templates. This enables the executor system (future work unit 002) to spawn any agent with appropriate context.

**Key insight from the design doc**: Agents are lightweight configuration (data), not behavior. They are simple structs representing roles, not interfaces with varying behavior. This makes them easily serializable and extensible.

## Requirements

### Agent Struct Definition

Create a new package at `cli/internal/agents/` with the core Agent struct in `agents.go`:

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

### Standard Agent Definitions

Define the following standard agents as package-level variables with pointer type (`*Agent`):

1. **Implementer**
   - Name: `"implementer"`
   - Description: `"Code implementation using Test-Driven Development"`
   - Capabilities: `"Must be able to read/write files, execute shell commands, search codebase"`
   - PromptPath: `"implementer.md"`

2. **Architect**
   - Name: `"architect"`
   - Description: `"System design and architecture decisions"`
   - Capabilities: `"Must be able to read/write files, search codebase"`
   - PromptPath: `"architect.md"`

3. **Reviewer**
   - Name: `"reviewer"`
   - Description: `"Code review and quality assessment"`
   - Capabilities: `"Must be able to read files, search codebase, execute shell commands"`
   - PromptPath: `"reviewer.md"`

4. **Planner**
   - Name: `"planner"`
   - Description: `"Research codebase and create comprehensive implementation task breakdown"`
   - Capabilities: `"Must be able to read files, search codebase, write task descriptions"`
   - PromptPath: `"planner.md"`

5. **Researcher**
   - Name: `"researcher"`
   - Description: `"Focused, impartial research with comprehensive source investigation and citation"`
   - Capabilities: `"Must be able to read files, search codebase, access web resources"`
   - PromptPath: `"researcher.md"`

6. **Decomposer**
   - Name: `"decomposer"`
   - Description: `"Specialized for decomposing complex features into project-sized, implementable work units"`
   - Capabilities: `"Must be able to read files, search codebase, write specifications"`
   - PromptPath: `"decomposer.md"`

### StandardAgents Function

Implement a function that returns all standard agents:

```go
func StandardAgents() []*Agent {
    return []*Agent{Implementer, Architect, Reviewer, Planner, Researcher, Decomposer}
}
```

## Acceptance Criteria

1. **Package created** at `cli/internal/agents/` with `agents.go` file
2. **Agent struct defined** with all four fields: Name, Description, Capabilities, PromptPath
3. **6 standard agents defined** as package-level variables with appropriate metadata
4. **StandardAgents() function** returns all 6 agents
5. **Unit tests verify**:
   - All standard agents have non-empty Name, Description, Capabilities, PromptPath
   - All agent Names are unique
   - StandardAgents() returns exactly 6 agents
   - StandardAgents() includes all expected agent names
6. **Code follows project conventions**:
   - Package comment explaining purpose
   - Godoc comments on exported types and functions
   - Table-driven tests

## Technical Details

### Package Structure

```
cli/internal/agents/
└── agents.go           # Agent struct + StandardAgents()
└── agents_test.go      # Unit tests
```

### Import Path

The package will be imported as:
```go
import "github.com/jmgilman/sow/cli/internal/agents"
```

### Naming Conventions

Follow the codebase conventions observed in other packages:
- Package-level variables for standard agents use PascalCase (e.g., `Implementer`, `Architect`)
- Function names use PascalCase for exported functions
- Struct field names use PascalCase

## Relevant Inputs

These files provide context for this task:

- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/implementer.md` - Existing implementer agent definition with YAML frontmatter, shows expected description and capabilities
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/planner.md` - Existing planner agent definition
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/reviewer.md` - Existing reviewer agent definition
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/researcher.md` - Existing researcher agent definition
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.claude/agents/decomposer.md` - Existing decomposer agent definition
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.sow/knowledge/designs/multi-agent-architecture.md` - Design doc with Agent struct specification (lines 138-189)
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/cli/internal/sdks/project/state/registry.go` - Example registry pattern from the codebase
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/cli/internal/sdks/project/state/registry_test.go` - Example test patterns for registry-style code

## Examples

### Expected Agent Definition Pattern

```go
var Implementer = &Agent{
    Name:         "implementer",
    Description:  "Code implementation using Test-Driven Development",
    Capabilities: "Must be able to read/write files, execute shell commands, search codebase",
    PromptPath:   "implementer.md",
}
```

### Expected Test Pattern (Table-Driven)

```go
func TestStandardAgents(t *testing.T) {
    agents := StandardAgents()

    if len(agents) != 6 {
        t.Errorf("StandardAgents() returned %d agents, want 6", len(agents))
    }

    // Check all expected agents are present
    expectedNames := []string{"implementer", "architect", "reviewer", "planner", "researcher", "decomposer"}
    for _, expected := range expectedNames {
        found := false
        for _, agent := range agents {
            if agent.Name == expected {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("StandardAgents() missing expected agent: %s", expected)
        }
    }
}
```

### Expected Validation Test Pattern

```go
func TestAgentFieldsNotEmpty(t *testing.T) {
    for _, agent := range StandardAgents() {
        t.Run(agent.Name, func(t *testing.T) {
            if agent.Name == "" {
                t.Error("Agent.Name is empty")
            }
            if agent.Description == "" {
                t.Error("Agent.Description is empty")
            }
            if agent.Capabilities == "" {
                t.Error("Agent.Capabilities is empty")
            }
            if agent.PromptPath == "" {
                t.Error("Agent.PromptPath is empty")
            }
        })
    }
}
```

## Dependencies

None - this is the foundational task for the agent system.

## Constraints

- **No executor logic** - This task only defines agent data structures, not execution behavior
- **No registry implementation** - Registry is task 020
- **No template loading** - Template embedding is task 030
- **Pointer receivers** - Standard agents should be pointer types (`*Agent`) for consistency with design doc and to allow nil checks
- **Keep it simple** - Agents are configuration, not behavior; avoid adding methods that belong in executor layer
