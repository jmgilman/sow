package prompts_test

import (
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/schemas"
)

func TestRender_GreetContext_Standard(t *testing.T) {
	ctx := &prompts.GreetContext{
		SowInitialized: false,
		HasProject:     false,
	}

	output, err := prompts.Render(prompts.PromptGreetStandard, ctx)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify standard Claude Code mode
	if !strings.Contains(output, "Standard Claude Code") {
		t.Error("Expected output to mention 'Standard Claude Code'")
	}
}

func TestRender_GreetContext_Operator(t *testing.T) {
	ctx := &prompts.GreetContext{
		SowInitialized: true,
		HasProject:     false,
		OpenIssues:     5,
		GHAvailable:    true,
	}

	output, err := prompts.Render(prompts.PromptGreetOperator, ctx)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify output contains expected content
	if !strings.Contains(output, "sow orchestrator") {
		t.Error("Expected output to mention 'sow orchestrator'")
	}

	if !strings.Contains(output, "operator mode") {
		t.Error("Expected output to mention 'operator mode'")
	}

	// Should mention open issues
	if !strings.Contains(output, "5 open issues") {
		t.Error("Expected output to mention '5 open issues'")
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

	output, err := prompts.Render(prompts.PromptGreetOrchestrator, ctx)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify orchestrator mode
	if !strings.Contains(output, "orchestrator mode") {
		t.Error("Expected output to mention 'orchestrator mode'")
	}

	// Verify project information is present
	if !strings.Contains(output, "test-feature") {
		t.Error("Expected output to mention project name 'test-feature'")
	}

	if !strings.Contains(output, "feat/test") {
		t.Error("Expected output to mention branch 'feat/test'")
	}

	if !strings.Contains(output, "implementation") {
		t.Error("Expected output to mention current phase 'implementation'")
	}

	if !strings.Contains(output, "020") {
		t.Error("Expected output to mention current task ID '020'")
	}
}

func TestRender_StatechartContext_NoProject(t *testing.T) {
	ctx := &prompts.StatechartContext{
		State:        prompts.StateNoProject,
		ProjectState: nil,
	}

	output, err := prompts.Render(prompts.PromptNoProject, ctx)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify no project message
	if !strings.Contains(output, "NO ACTIVE PROJECT") {
		t.Error("Expected output to indicate no active project")
	}
}

func TestRender_StatechartContext_DesignActive(t *testing.T) {
	state := &schemas.ProjectState{}
	state.Project.Name = "test-project"
	state.Project.Description = "Test description"
	state.Project.Branch = "main"
	state.Phases.Design.Status = "in_progress"
	state.Phases.Design.Artifacts = []schemas.Artifact{
		{Path: "docs/design.md", Approved: true},
		{Path: "docs/adr-001.md", Approved: false},
	}

	ctx := &prompts.StatechartContext{
		State:        prompts.StateDesignActive,
		ProjectState: state,
	}

	output, err := prompts.Render(prompts.PromptDesignActive, ctx)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify design phase content
	if !strings.Contains(output, "DESIGN PHASE") {
		t.Error("Expected output to mention 'DESIGN PHASE'")
	}

	if !strings.Contains(output, "test-project") {
		t.Error("Expected output to mention project name")
	}

	// Should show artifact counts
	if !strings.Contains(output, "2 total") {
		t.Error("Expected output to show 2 total artifacts")
	}

	if !strings.Contains(output, "1 approved") {
		t.Error("Expected output to show 1 approved artifact")
	}
}

func TestRender_StatechartContext_ImplementationExecuting(t *testing.T) {
	state := &schemas.ProjectState{}
	state.Project.Name = "impl-project"
	state.Project.Description = "Implementation test"
	state.Project.Branch = "feat/impl"
	state.Phases.Implementation.Status = "in_progress"
	state.Phases.Implementation.Tasks = []schemas.Task{
		{Id: "010", Name: "Task 1", Status: "completed"},
		{Id: "020", Name: "Task 2", Status: "in_progress"},
		{Id: "030", Name: "Task 3", Status: "pending"},
	}

	ctx := &prompts.StatechartContext{
		State:        prompts.StateImplementationExecuting,
		ProjectState: state,
	}

	output, err := prompts.Render(prompts.PromptImplementationExecuting, ctx)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify implementation content
	if !strings.Contains(output, "IMPLEMENTATION EXECUTING") {
		t.Error("Expected output to mention 'IMPLEMENTATION EXECUTING'")
	}

	if !strings.Contains(output, "impl-project") {
		t.Error("Expected output to mention project name")
	}

	// Should show task breakdown
	if !strings.Contains(output, "Total: 3") {
		t.Error("Expected output to show 3 total tasks")
	}

	if !strings.Contains(output, "Completed: 1") {
		t.Error("Expected output to show 1 completed task")
	}

	if !strings.Contains(output, "In Progress: 1") {
		t.Error("Expected output to show 1 in-progress task")
	}

	if !strings.Contains(output, "Pending: 1") {
		t.Error("Expected output to show 1 pending task")
	}
}

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

func TestStatechartContext_ToMap(t *testing.T) {
	state := &schemas.ProjectState{}
	state.Project.Name = "test-project"
	state.Project.Description = "Description"
	state.Project.Branch = "main"

	ctx := &prompts.StatechartContext{
		State:        prompts.StateDesignActive,
		ProjectState: state,
	}

	data := ctx.ToMap()

	if data["ProjectName"] != "test-project" {
		t.Error("Expected ProjectName to be 'test-project'")
	}

	if data["ProjectDescription"] != "Description" {
		t.Error("Expected ProjectDescription to be 'Description'")
	}

	if data["ProjectBranch"] != "main" {
		t.Error("Expected ProjectBranch to be 'main'")
	}
}
