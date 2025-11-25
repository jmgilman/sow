package agents

import "fmt"

// AgentRegistry provides lookup and listing of registered agents.
// It is the central registry for all agent definitions in the system.
//
// The registry is designed to be populated at initialization time with
// standard agents and any custom agents. Thread safety is not required
// since registration happens only during initialization.
type AgentRegistry struct {
	agents map[string]*Agent
}

// NewAgentRegistry creates a new AgentRegistry pre-populated with all standard agents.
// This is the recommended way to create a registry for production use.
//
// Example:
//
//	registry := agents.NewAgentRegistry()
//	agent, err := registry.Get("implementer")
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewAgentRegistry() *AgentRegistry {
	r := &AgentRegistry{
		agents: make(map[string]*Agent),
	}

	// Register all standard agents
	for _, agent := range StandardAgents() {
		r.Register(agent)
	}

	return r
}

// Register adds an agent to the registry.
// Panics if an agent with the same name is already registered.
// This follows the same pattern as other registries in the codebase.
//
// Example:
//
//	registry := NewAgentRegistry()
//	registry.Register(&Agent{Name: "custom", Description: "Custom agent"})
func (r *AgentRegistry) Register(agent *Agent) {
	if _, exists := r.agents[agent.Name]; exists {
		panic(fmt.Sprintf("agent already registered: %s", agent.Name))
	}

	r.agents[agent.Name] = agent
}

// Get returns an agent by name.
// Returns (agent, nil) if found, (nil, error) if not found.
//
// Example:
//
//	agent, err := registry.Get("implementer")
//	if err != nil {
//	    return fmt.Errorf("failed to get agent: %w", err)
//	}
func (r *AgentRegistry) Get(name string) (*Agent, error) {
	agent, ok := r.agents[name]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %s", name)
	}

	return agent, nil
}

// List returns all registered agents.
// The order of returned agents is not guaranteed.
//
// Example:
//
//	for _, agent := range registry.List() {
//	    fmt.Printf("- %s: %s\n", agent.Name, agent.Description)
//	}
func (r *AgentRegistry) List() []*Agent {
	agents := make([]*Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	return agents
}
