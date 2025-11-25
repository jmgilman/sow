package agents

import (
	"strings"
	"testing"
)

// TestNewAgentRegistry verifies that NewAgentRegistry creates a registry pre-populated with standard agents.
func TestNewAgentRegistry(t *testing.T) {
	registry := NewAgentRegistry()

	if registry == nil {
		t.Fatal("NewAgentRegistry() returned nil")
	}

	// Should have all 6 standard agents
	agents := registry.List()
	if len(agents) != 6 {
		t.Errorf("NewAgentRegistry() contains %d agents, want 6", len(agents))
	}
}

// TestAgentRegistry_Get verifies that Get returns correct agents.
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
			name:      "architect exists",
			agentName: "architect",
			wantError: false,
		},
		{
			name:      "reviewer exists",
			agentName: "reviewer",
			wantError: false,
		},
		{
			name:      "planner exists",
			agentName: "planner",
			wantError: false,
		},
		{
			name:      "researcher exists",
			agentName: "researcher",
			wantError: false,
		},
		{
			name:      "decomposer exists",
			agentName: "decomposer",
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

// TestAgentRegistry_GetErrorMessage verifies the error message format.
func TestAgentRegistry_GetErrorMessage(t *testing.T) {
	registry := NewAgentRegistry()

	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}

	// Error should be lowercase and contain the agent name
	errMsg := err.Error()
	if !strings.Contains(errMsg, "unknown agent") {
		t.Errorf("error message = %q, want to contain 'unknown agent'", errMsg)
	}
	if !strings.Contains(errMsg, "nonexistent") {
		t.Errorf("error message = %q, want to contain 'nonexistent'", errMsg)
	}
}

// TestAgentRegistry_List verifies that List returns all registered agents.
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

// TestAgentRegistry_Register verifies that Register adds agents to the registry.
func TestAgentRegistry_Register(t *testing.T) {
	// Create a fresh registry without standard agents
	registry := &AgentRegistry{
		agents: make(map[string]*Agent),
	}

	customAgent := &Agent{
		Name:         "custom",
		Description:  "Custom agent",
		Capabilities: "Custom capabilities",
		PromptPath:   "custom.md",
	}

	// Register should not panic
	registry.Register(customAgent)

	// Verify agent was added
	agent, err := registry.Get("custom")
	if err != nil {
		t.Fatalf("Get(custom) error = %v", err)
	}
	if agent.Name != "custom" {
		t.Errorf("Get(custom).Name = %q, want %q", agent.Name, "custom")
	}
}

// TestAgentRegistry_RegisterDuplicatePanics verifies that registering a duplicate agent panics.
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
			if !strings.Contains(msg, "already registered") {
				t.Errorf("expected panic message to contain 'already registered', got: %s", msg)
			}
		}
	}()

	// Should panic - implementer already registered via NewAgentRegistry
	registry.Register(&Agent{Name: "implementer"})
}

// TestAgentRegistry_ListAfterRegister verifies that List includes newly registered agents.
func TestAgentRegistry_ListAfterRegister(t *testing.T) {
	registry := NewAgentRegistry()

	customAgent := &Agent{
		Name:         "custom",
		Description:  "Custom agent",
		Capabilities: "Custom capabilities",
		PromptPath:   "custom.md",
	}

	registry.Register(customAgent)

	agents := registry.List()

	// Should now have 7 agents
	if len(agents) != 7 {
		t.Errorf("List() returned %d agents after Register, want 7", len(agents))
	}

	// Verify custom agent is included
	found := false
	for _, agent := range agents {
		if agent.Name == "custom" {
			found = true
			break
		}
	}
	if !found {
		t.Error("List() missing custom agent after Register")
	}
}

// TestAgentRegistry_ListReturnsSlice verifies that List returns a slice.
func TestAgentRegistry_ListReturnsSlice(t *testing.T) {
	registry := NewAgentRegistry()
	agents := registry.List()

	// List should return a slice, not nil
	if agents == nil {
		t.Error("List() returned nil")
	}
}
