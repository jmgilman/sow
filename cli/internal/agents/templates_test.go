package agents

import (
	"strings"
	"testing"
)

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
			name:       "reviewer exists",
			promptPath: "reviewer.md",
			wantError:  false,
		},
		{
			name:       "planner exists",
			promptPath: "planner.md",
			wantError:  false,
		},
		{
			name:       "researcher exists",
			promptPath: "researcher.md",
			wantError:  false,
		},
		{
			name:       "decomposer exists",
			promptPath: "decomposer.md",
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

func TestLoadPromptContentNotEmpty(t *testing.T) {
	templates := []string{
		"implementer.md",
		"architect.md",
		"reviewer.md",
		"planner.md",
		"researcher.md",
		"decomposer.md",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := LoadPrompt(tmpl)
			if err != nil {
				t.Errorf("LoadPrompt(%q) failed: %v", tmpl, err)
				return
			}
			if len(content) < 50 {
				t.Errorf("LoadPrompt(%q) returned suspiciously short content: %d bytes", tmpl, len(content))
			}
		})
	}
}

func TestLoadPromptErrorWrapping(t *testing.T) {
	_, err := LoadPrompt("nonexistent.md")
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
	if !strings.Contains(err.Error(), "nonexistent.md") {
		t.Errorf("error should contain the prompt path, got: %v", err)
	}
	if !strings.Contains(err.Error(), "failed to load prompt") {
		t.Errorf("error should contain 'failed to load prompt', got: %v", err)
	}
}

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

func TestLoadPromptNoYAMLFrontmatter(t *testing.T) {
	// Migrated templates should NOT have YAML frontmatter
	templates := []string{
		"implementer.md",
		"planner.md",
		"reviewer.md",
		"researcher.md",
		"decomposer.md",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := LoadPrompt(tmpl)
			if err != nil {
				t.Errorf("LoadPrompt(%q) failed: %v", tmpl, err)
				return
			}
			// Check that content does NOT start with YAML frontmatter
			if strings.HasPrefix(content, "---\n") {
				t.Errorf("LoadPrompt(%q) contains YAML frontmatter which should have been removed", tmpl)
			}
		})
	}
}

func TestLoadPromptArchitectTemplate(t *testing.T) {
	content, err := LoadPrompt("architect.md")
	if err != nil {
		t.Fatalf("LoadPrompt(architect.md) failed: %v", err)
	}

	// Verify architect template has expected self-initialization pattern
	checks := []string{
		"architect",
		"sow prompt",
		"guidance",
		"Initialization",
	}

	for _, check := range checks {
		if !strings.Contains(strings.ToLower(content), strings.ToLower(check)) {
			t.Errorf("architect.md should contain %q", check)
		}
	}
}
