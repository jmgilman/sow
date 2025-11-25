# Agent Registry Implementation

## Context

This task is part of **Work Unit 001: Agent System Core**, implementing a lightweight agent definition system for sow.

The AgentRegistry provides a centralized lookup mechanism for agent definitions. It allows the executor system (future work unit 002) to look up agents by name and list all available agents. The registry follows the same patterns established in the codebase for similar registries (like `refs.Registry` and `state.Registry`).

**Task 010 must be completed first** - this task depends on the Agent struct and StandardAgents() function being available.

## Requirements

### AgentRegistry Struct

Create `registry.go` in the `cli/internal/agents/` package:

```go
// AgentRegistry provides lookup and listing of registered agents.
// It is the central registry for all agent definitions in the system.
type AgentRegistry struct {
    agents map[string]*Agent
}
```

### Constructor

```go
// NewAgentRegistry creates a new AgentRegistry pre-populated with all standard agents.
// This is the recommended way to create a registry for production use.
func NewAgentRegistry() *AgentRegistry
```

The constructor should:
1. Create a new registry with an initialized map
2. Register all standard agents from `StandardAgents()`
3. Return the populated registry

### Register Method

```go
// Register adds an agent to the registry.
// Panics if an agent with the same name is already registered.
// This follows the same pattern as other registries in the codebase.
func (r *AgentRegistry) Register(agent *Agent)
```

The Register method should:
1. Check if an agent with the same name already exists
2. If duplicate, panic with a clear error message: `"agent already registered: <name>"`
3. Otherwise, add the agent to the internal map

### Get Method

```go
// Get returns an agent by name.
// Returns (agent, nil) if found, (nil, error) if not found.
func (r *AgentRegistry) Get(name string) (*Agent, error)
```

The Get method should:
1. Look up the agent by name in the internal map
2. If found, return the agent and nil error
3. If not found, return nil and an error: `"unknown agent: <name>"`

### List Method

```go
// List returns all registered agents.
// The order of returned agents is not guaranteed.
func (r *AgentRegistry) List() []*Agent
```

The List method should:
1. Return a slice containing all registered agents
2. Order is not guaranteed (this is consistent with the codebase patterns)

## Acceptance Criteria

1. **AgentRegistry struct defined** with agents map field
2. **NewAgentRegistry()** returns registry pre-populated with standard agents
3. **Register()** adds agents and panics on duplicates
4. **Get()** returns correct agent or error for unknown agent
5. **List()** returns all registered agents
6. **Unit tests cover**:
   - NewAgentRegistry() pre-populates with standard agents
   - Get() returns correct agent by name
   - Get() returns error for unknown agent
   - Get() returns error for empty string
   - List() returns all registered agents
   - Register() adds agent to registry
   - Register() panics on duplicate name
   - List() after Register() includes new agent

## Technical Details

### Package Structure Update

```
cli/internal/agents/
├── agents.go           # Agent struct + StandardAgents() (from task 010)
├── agents_test.go      # Tests for agents.go (from task 010)
├── registry.go         # AgentRegistry implementation (this task)
└── registry_test.go    # Tests for registry (this task)
```

### Error Message Format

Follow the codebase convention for error messages:
- Use lowercase error messages
- Include the problematic value in the message
- Example: `fmt.Errorf("unknown agent: %s", name)`

### Panic Message Format

Follow the codebase convention for panic messages (see `cli/internal/refs/registry.go`):
- Use `fmt.Sprintf` with clear message
- Example: `panic(fmt.Sprintf("agent already registered: %s", name))`

## Relevant Inputs

These files provide context for this task:

- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/cli/internal/refs/registry.go` - Reference implementation showing registry pattern with Register/GetType/AllTypes
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/cli/internal/refs/registry_test.go` - Test patterns for registry operations
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/cli/internal/sdks/project/state/registry.go` - Alternative registry pattern with global map and Register/Get functions
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/cli/internal/sdks/project/state/registry_test.go` - Test patterns for simpler registry
- `/Users/josh/code/sow/.sow/worktrees/feat/agent-system-core-97/.sow/knowledge/designs/multi-agent-architecture.md` - Design doc with AgentRegistry specification (lines 199-221)

## Examples

### Expected NewAgentRegistry Usage

```go
func main() {
    registry := agents.NewAgentRegistry()

    // Get agent by name
    agent, err := registry.Get("implementer")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Agent: %s - %s\n", agent.Name, agent.Description)

    // List all agents
    for _, a := range registry.List() {
        fmt.Printf("- %s\n", a.Name)
    }
}
```

### Expected Test Pattern for Get

```go
func TestAgentRegistry_Get(t *testing.T) {
    tests := []struct {
        name      string
        agentName string
        wantError bool
    }{
        {
            name:      "implementer exists",
            agentName: "implementer",
            wantError: false,
        },
        {
            name:      "unknown agent",
            agentName: "unknown",
            wantError: true,
        },
        {
            name:      "empty string",
            agentName: "",
            wantError: true,
        },
    }

    registry := NewAgentRegistry()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            agent, err := registry.Get(tt.agentName)
            if (err != nil) != tt.wantError {
                t.Errorf("Get(%q) error = %v, wantError %v", tt.agentName, err, tt.wantError)
                return
            }
            if !tt.wantError && agent == nil {
                t.Errorf("Get(%q) returned nil agent", tt.agentName)
            }
            if !tt.wantError && agent.Name != tt.agentName {
                t.Errorf("Get(%q).Name = %q, want %q", tt.agentName, agent.Name, tt.agentName)
            }
        })
    }
}
```

### Expected Test Pattern for Register Panic

```go
func TestAgentRegistry_RegisterDuplicatePanics(t *testing.T) {
    registry := NewAgentRegistry()

    defer func() {
        if r := recover(); r == nil {
            t.Error("expected panic on duplicate registration")
        } else {
            msg, ok := r.(string)
            if !ok {
                t.Error("expected panic message to be a string")
            }
            if !strings.Contains(msg, "implementer") {
                t.Errorf("expected panic message to mention 'implementer', got: %s", msg)
            }
        }
    }()

    // Should panic - implementer already registered via NewAgentRegistry
    registry.Register(&Agent{Name: "implementer"})
}
```

### Expected Test Pattern for List

```go
func TestAgentRegistry_List(t *testing.T) {
    registry := NewAgentRegistry()
    agents := registry.List()

    // Should have all 6 standard agents
    if len(agents) != 6 {
        t.Errorf("List() returned %d agents, want 6", len(agents))
    }

    // Verify all standard agents present
    expectedNames := map[string]bool{
        "implementer": false,
        "architect":   false,
        "reviewer":    false,
        "planner":     false,
        "researcher":  false,
        "decomposer":  false,
    }

    for _, agent := range agents {
        if _, ok := expectedNames[agent.Name]; ok {
            expectedNames[agent.Name] = true
        }
    }

    for name, found := range expectedNames {
        if !found {
            t.Errorf("List() missing expected agent: %s", name)
        }
    }
}
```

## Dependencies

- **Task 010** - Agent struct and StandardAgents() function must be implemented first

## Constraints

- **No thread safety required** - The registry pattern in this codebase uses simple maps; thread safety is only added where demonstrated (like `refs.Registry` with mutex). For the agent registry, agents are registered at initialization time and then only read, so no mutex is needed.
- **Panic on duplicate** - Follow the established pattern of panicking on duplicate registration (this catches configuration errors early)
- **No custom agents yet** - User-extensible custom agents are deferred to v2
- **Instance-based registry** - Use instance-based registry (with `NewAgentRegistry()`) rather than global package-level map, to allow for testing isolation
