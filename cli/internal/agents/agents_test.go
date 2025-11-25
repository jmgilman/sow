// Package agents provides agent definitions for the sow multi-agent system.
package agents

import "testing"

// TestAgentStructHasRequiredFields verifies that the Agent struct has the expected fields.
func TestAgentStructHasRequiredFields(t *testing.T) {
	// Verify struct can be instantiated with all required fields
	agent := &Agent{
		Name:         "test",
		Description:  "Test agent",
		Capabilities: "Can do tests",
		PromptPath:   "test.md",
	}

	if agent.Name != "test" {
		t.Errorf("Name field = %q, want %q", agent.Name, "test")
	}
	if agent.Description != "Test agent" {
		t.Errorf("Description field = %q, want %q", agent.Description, "Test agent")
	}
	if agent.Capabilities != "Can do tests" {
		t.Errorf("Capabilities field = %q, want %q", agent.Capabilities, "Can do tests")
	}
	if agent.PromptPath != "test.md" {
		t.Errorf("PromptPath field = %q, want %q", agent.PromptPath, "test.md")
	}
}

// TestStandardAgentsCount verifies that StandardAgents() returns exactly 6 agents.
func TestStandardAgentsCount(t *testing.T) {
	agents := StandardAgents()

	if len(agents) != 6 {
		t.Errorf("StandardAgents() returned %d agents, want 6", len(agents))
	}
}

// TestStandardAgentsContainsAllExpectedNames verifies that StandardAgents() includes all expected agent names.
func TestStandardAgentsContainsAllExpectedNames(t *testing.T) {
	agents := StandardAgents()

	expectedNames := []string{
		"implementer",
		"architect",
		"reviewer",
		"planner",
		"researcher",
		"decomposer",
	}

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

// TestStandardAgentNamesAreUnique verifies that all agent names in StandardAgents() are unique.
func TestStandardAgentNamesAreUnique(t *testing.T) {
	agents := StandardAgents()
	seen := make(map[string]bool)

	for _, agent := range agents {
		if seen[agent.Name] {
			t.Errorf("Duplicate agent name found: %s", agent.Name)
		}
		seen[agent.Name] = true
	}
}

// TestAgentFieldsNotEmpty verifies that all standard agents have non-empty fields.
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

// TestStandardAgentDefinitions verifies the specific definitions of each standard agent.
func TestStandardAgentDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		agent    *Agent
		wantDesc string
		wantCaps string
		wantPath string
	}{
		{
			name:     "Implementer",
			agent:    Implementer,
			wantDesc: "Code implementation using Test-Driven Development",
			wantCaps: "Must be able to read/write files, execute shell commands, search codebase",
			wantPath: "implementer.md",
		},
		{
			name:     "Architect",
			agent:    Architect,
			wantDesc: "System design and architecture decisions",
			wantCaps: "Must be able to read/write files, search codebase",
			wantPath: "architect.md",
		},
		{
			name:     "Reviewer",
			agent:    Reviewer,
			wantDesc: "Code review and quality assessment",
			wantCaps: "Must be able to read files, search codebase, execute shell commands",
			wantPath: "reviewer.md",
		},
		{
			name:     "Planner",
			agent:    Planner,
			wantDesc: "Research codebase and create comprehensive implementation task breakdown",
			wantCaps: "Must be able to read files, search codebase, write task descriptions",
			wantPath: "planner.md",
		},
		{
			name:     "Researcher",
			agent:    Researcher,
			wantDesc: "Focused, impartial research with comprehensive source investigation and citation",
			wantCaps: "Must be able to read files, search codebase, access web resources",
			wantPath: "researcher.md",
		},
		{
			name:     "Decomposer",
			agent:    Decomposer,
			wantDesc: "Specialized for decomposing complex features into project-sized, implementable work units",
			wantCaps: "Must be able to read files, search codebase, write specifications",
			wantPath: "decomposer.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.agent == nil {
				t.Fatalf("%s agent is nil", tt.name)
			}
			if tt.agent.Description != tt.wantDesc {
				t.Errorf("%s.Description = %q, want %q", tt.name, tt.agent.Description, tt.wantDesc)
			}
			if tt.agent.Capabilities != tt.wantCaps {
				t.Errorf("%s.Capabilities = %q, want %q", tt.name, tt.agent.Capabilities, tt.wantCaps)
			}
			if tt.agent.PromptPath != tt.wantPath {
				t.Errorf("%s.PromptPath = %q, want %q", tt.name, tt.agent.PromptPath, tt.wantPath)
			}
		})
	}
}

// TestStandardAgentsReturnsPointers verifies that StandardAgents returns pointers to agents.
func TestStandardAgentsReturnsPointers(t *testing.T) {
	agents := StandardAgents()

	for _, agent := range agents {
		if agent == nil {
			t.Error("StandardAgents() returned a nil agent")
		}
	}
}

// TestPackageLevelAgentsAreDefined verifies that all package-level agent variables are defined.
func TestPackageLevelAgentsAreDefined(t *testing.T) {
	tests := []struct {
		name  string
		agent *Agent
	}{
		{"Implementer", Implementer},
		{"Architect", Architect},
		{"Reviewer", Reviewer},
		{"Planner", Planner},
		{"Researcher", Researcher},
		{"Decomposer", Decomposer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.agent == nil {
				t.Errorf("%s is nil", tt.name)
			}
		})
	}
}
