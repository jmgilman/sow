package prompts_test

import (
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
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

func TestRender_StatechartContext_PlanningActive(t *testing.T) {
	state := &schemas.ProjectState{}
	state.Project.Name = "test-project"
	state.Project.Description = "Test description"
	state.Project.Branch = "main"
	state.Phases.Planning.Status = "in_progress"
	state.Phases.Planning.Artifacts = []phases.Artifact{
		{Path: "task-list.md", Approved: true, Metadata: map[string]interface{}{"type": "task_list"}},
		{Path: "context.md", Approved: false},
	}

	ctx := &prompts.StatechartContext{
		State:        prompts.StatePlanningActive,
		ProjectState: state,
	}

	output, err := prompts.Render(prompts.PromptPlanningActive, ctx)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify planning phase content
	if !strings.Contains(output, "PLANNING PHASE") {
		t.Error("Expected output to mention 'PLANNING PHASE'")
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
	state.Phases.Implementation.Tasks = []phases.Task{
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
		State:        prompts.StatePlanningActive,
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
