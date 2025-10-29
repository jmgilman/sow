package prompts_test

import (
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/prompts"
)

func TestRender_GreetContext_Uninitialized(t *testing.T) {
	ctx := &prompts.GreetContext{
		SowInitialized: false,
		HasProject:     false,
	}

	// Test base greeting
	base, err := prompts.Render(prompts.PromptGreetBase, ctx)
	if err != nil {
		t.Fatalf("Render base failed: %v", err)
	}

	if !strings.Contains(base, "sow Orchestrator") {
		t.Error("Expected base to mention 'sow Orchestrator'")
	}

	if !strings.Contains(base, "personal development assistant") {
		t.Error("Expected base to mention 'personal development assistant'")
	}

	// Test uninitialized state
	state, err := prompts.Render(prompts.PromptGreetStateUninit, ctx)
	if err != nil {
		t.Fatalf("Render state failed: %v", err)
	}

	if !strings.Contains(state, "Repository Not Initialized") {
		t.Error("Expected state to mention 'Repository Not Initialized'")
	}

	if !strings.Contains(state, "sow init") {
		t.Error("Expected state to mention 'sow init'")
	}
}

func TestRender_GreetContext_Operator(t *testing.T) {
	ctx := &prompts.GreetContext{
		SowInitialized: true,
		HasProject:     false,
		OpenIssues:     5,
		GHAvailable:    true,
	}

	// Test operator state
	state, err := prompts.Render(prompts.PromptGreetStateOperator, ctx)
	if err != nil {
		t.Fatalf("Render state failed: %v", err)
	}

	// Verify output contains expected content
	if !strings.Contains(state, "Sow initialized") {
		t.Error("Expected state to mention 'Sow initialized'")
	}

	if !strings.Contains(state, "No active project") {
		t.Error("Expected state to mention 'No active project'")
	}

	// Should mention open issues
	if !strings.Contains(state, "5 open") {
		t.Error("Expected state to mention '5 open' issues")
	}

	if !strings.Contains(state, "sow-labeled issues") {
		t.Error("Expected state to mention 'sow-labeled issues'")
	}
}

func TestRender_GreetContext_Orchestrator(t *testing.T) {
	ctx := &prompts.GreetContext{
		SowInitialized: true,
		HasProject:     true,
		Project: &prompts.ProjectGreetContext{
			Name:            "test-feature",
			Branch:          "feat/test",
			Description:     "Test feature implementation",
			CurrentPhase:    "implementation",
			PhaseStatus:     "in_progress",
			TasksTotal:      5,
			TasksComplete:   2,
			TasksInProgress: 1,
			TasksPending:    2,
			CurrentTask: &prompts.TaskGreetContext{
				ID:   "020",
				Name: "Implement core logic",
			},
		},
		OpenIssues:  3,
		GHAvailable: true,
	}

	// Test orchestrator state
	state, err := prompts.Render(prompts.PromptGreetStateOrch, ctx)
	if err != nil {
		t.Fatalf("Render state failed: %v", err)
	}

	// Verify project information is present
	if !strings.Contains(state, "test-feature") {
		t.Error("Expected state to mention project name 'test-feature'")
	}

	if !strings.Contains(state, "feat/test") {
		t.Error("Expected state to mention branch 'feat/test'")
	}

	if !strings.Contains(state, "implementation") {
		t.Error("Expected state to mention current phase 'implementation'")
	}

	if !strings.Contains(state, "020") {
		t.Error("Expected state to mention current task ID '020'")
	}

	if !strings.Contains(state, "Project Management Mode") {
		t.Error("Expected state to mention 'Project Management Mode'")
	}

	if !strings.Contains(state, "2/5") {
		t.Error("Expected state to show task completion '2/5'")
	}
}

// Note: Statechart-specific prompt tests have been moved to cli/internal/project/standard/prompts_test.go
// as they now use the standard project's own registry.

func TestRender_UnknownPromptID(t *testing.T) {
	ctx := &prompts.GreetContext{
		SowInitialized: true,
	}

	_, err := prompts.Render("unknown.prompt", ctx)
	if err == nil {
		t.Fatal("Expected error for unknown prompt ID")
	}

	if !strings.Contains(err.Error(), "unknown prompt ID") {
		t.Errorf("Expected 'unknown prompt ID' error, got: %v", err)
	}
}

func TestGreetContext_ToMap(t *testing.T) {
	ctx := &prompts.GreetContext{
		SowInitialized: true,
		HasProject:     true,
		OpenIssues:     3,
		GHAvailable:    true,
		Project: &prompts.ProjectGreetContext{
			Name:   "test",
			Branch: "main",
		},
	}

	data := ctx.ToMap()

	if data["SowInitialized"] != true {
		t.Error("Expected SowInitialized to be true")
	}

	if data["HasProject"] != true {
		t.Error("Expected HasProject to be true")
	}

	if data["OpenIssues"] != 3 {
		t.Error("Expected OpenIssues to be 3")
	}

	projectData, ok := data["Project"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected Project to be a map")
	}

	if projectData["Name"] != "test" {
		t.Error("Expected Project.Name to be 'test'")
	}
}

// Note: TestStatechartContext_ToMap has been moved to cli/internal/project/standard/prompts_test.go
// as StatechartContext is now used only by project-specific prompt generators.
