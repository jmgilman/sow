package implementation

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/phases"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/qmuntal/stateless"
)

func TestNew(t *testing.T) {
	data := &phasesSchema.ImplementationPhase{
		Enabled:        true,
		Status:         "pending",
		Created_at:     time.Now(),
		Tasks:          []phasesSchema.Task{},
		Tasks_approved: false,
	}

	project := ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)

	if phase == nil {
		t.Fatal("Expected non-nil phase")
	}
}

func TestEntryState(t *testing.T) {
	phase := New(nil, ProjectInfo{})

	if phase.EntryState() != phases.ImplementationPlanning {
		t.Errorf("Expected entry state to be ImplementationPlanning, got %s", phase.EntryState())
	}
}

func TestMetadata(t *testing.T) {
	phase := New(nil, ProjectInfo{})
	meta := phase.Metadata()

	if meta.Name != "implementation" {
		t.Errorf("Expected name to be 'implementation', got %s", meta.Name)
	}

	if len(meta.States) != 2 {
		t.Errorf("Expected 2 states, got %d", len(meta.States))
	}

	if !meta.SupportsTasks {
		t.Error("Expected SupportsTasks to be true")
	}

	if meta.SupportsArtifacts {
		t.Error("Expected SupportsArtifacts to be false")
	}
}

func TestAddToMachine(t *testing.T) {
	sm := stateless.NewStateMachine(phases.ImplementationPlanning)

	// Provide data that will make guards pass
	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "pending"},
		},
		Tasks_approved: true,
	}

	phase := New(data, ProjectInfo{})

	phase.AddToMachine(sm, phases.ReviewActive)

	canFire, _ := sm.CanFire(phases.EventTaskCreated)
	if !canFire {
		t.Error("Expected EventTaskCreated to be configured")
	}

	canFire, _ = sm.CanFire(phases.EventTasksApproved)
	if !canFire {
		t.Error("Expected EventTasksApproved to be configured")
	}
}

func TestHasAtLeastOneTaskGuard_NoTasks(t *testing.T) {
	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{},
	}

	phase := New(data, ProjectInfo{})

	if phase.hasAtLeastOneTaskGuard(context.Background()) {
		t.Error("Expected guard to fail with no tasks")
	}
}

func TestHasAtLeastOneTaskGuard_WithTasks(t *testing.T) {
	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "pending"},
		},
	}

	phase := New(data, ProjectInfo{})

	if !phase.hasAtLeastOneTaskGuard(context.Background()) {
		t.Error("Expected guard to pass with tasks")
	}
}

func TestTasksApprovedGuard_NotApproved(t *testing.T) {
	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "pending"},
		},
		Tasks_approved: false,
	}

	phase := New(data, ProjectInfo{})

	if phase.tasksApprovedGuard(context.Background()) {
		t.Error("Expected guard to fail when tasks not approved")
	}
}

func TestTasksApprovedGuard_Approved(t *testing.T) {
	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "pending"},
		},
		Tasks_approved: true,
	}

	phase := New(data, ProjectInfo{})

	if !phase.tasksApprovedGuard(context.Background()) {
		t.Error("Expected guard to pass when tasks approved")
	}
}

func TestAllTasksCompleteGuard_NoTasks(t *testing.T) {
	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{},
	}

	phase := New(data, ProjectInfo{})

	if phase.allTasksCompleteGuard(context.Background()) {
		t.Error("Expected guard to fail with no tasks")
	}
}

func TestAllTasksCompleteGuard_SomeIncomplete(t *testing.T) {
	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "completed"},
			{Id: "020", Name: "Task 2", Status: "in_progress"},
		},
	}

	phase := New(data, ProjectInfo{})

	if phase.allTasksCompleteGuard(context.Background()) {
		t.Error("Expected guard to fail when some tasks incomplete")
	}
}

func TestAllTasksCompleteGuard_AllComplete(t *testing.T) {
	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "completed"},
			{Id: "020", Name: "Task 2", Status: "completed"},
		},
	}

	phase := New(data, ProjectInfo{})

	if !phase.allTasksCompleteGuard(context.Background()) {
		t.Error("Expected guard to pass when all tasks complete")
	}
}

func TestAllTasksCompleteGuard_WithAbandoned(t *testing.T) {
	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "completed"},
			{Id: "020", Name: "Task 2", Status: "abandoned"},
		},
	}

	phase := New(data, ProjectInfo{})

	if !phase.allTasksCompleteGuard(context.Background()) {
		t.Error("Expected guard to pass with completed/abandoned tasks")
	}
}

func TestPrepareTemplateData(t *testing.T) {
	plannerUsed := true
	data := &phasesSchema.ImplementationPhase{
		Planner_used: &plannerUsed,
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "completed"},
			{Id: "020", Name: "Task 2", Status: "in_progress"},
			{Id: "030", Name: "Task 3", Status: "pending"},
		},
		Tasks_approved: true,
	}

	project := ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)
	templateData := phase.prepareTemplateData()

	if templateData["PlannerUsed"] != true {
		t.Errorf("Expected PlannerUsed to be true, got %v", templateData["PlannerUsed"])
	}

	if templateData["TaskTotal"] != 3 {
		t.Errorf("Expected TaskTotal to be 3, got %v", templateData["TaskTotal"])
	}

	if templateData["TaskCompleted"] != 1 {
		t.Errorf("Expected TaskCompleted to be 1, got %v", templateData["TaskCompleted"])
	}

	if templateData["TaskInProgress"] != 1 {
		t.Errorf("Expected TaskInProgress to be 1, got %v", templateData["TaskInProgress"])
	}

	if templateData["TaskPending"] != 1 {
		t.Errorf("Expected TaskPending to be 1, got %v", templateData["TaskPending"])
	}

	if templateData["TasksApproved"] != true {
		t.Errorf("Expected TasksApproved to be true, got %v", templateData["TasksApproved"])
	}
}

func TestRenderPrompt_Planning(t *testing.T) {
	project := ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(&phasesSchema.ImplementationPhase{}, project)
	prompt := phase.renderPrompt("planning")

	if strings.Contains(prompt, "Error") {
		t.Errorf("Expected prompt to render without errors, got: %s", prompt)
	}

	if !strings.Contains(prompt, "Test Project") {
		t.Error("Expected prompt to contain project name")
	}
}

func TestFullTransitionFlow_TaskCreated(t *testing.T) {
	sm := stateless.NewStateMachine(phases.ImplementationPlanning)

	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "pending"},
		},
		Tasks_approved: false,
	}

	phase := New(data, ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, phases.ReviewActive)

	// Transition via task created
	sm.Fire(phases.EventTaskCreated)

	currentState := sm.MustState().(phases.State)
	if currentState != phases.ImplementationExecuting {
		t.Errorf("Expected state to be ImplementationExecuting, got %s", currentState)
	}
}

func TestFullTransitionFlow_TasksApproved(t *testing.T) {
	sm := stateless.NewStateMachine(phases.ImplementationPlanning)

	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "pending"},
		},
		Tasks_approved: true,
	}

	phase := New(data, ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, phases.ReviewActive)

	// Transition via tasks approved
	sm.Fire(phases.EventTasksApproved)

	currentState := sm.MustState().(phases.State)
	if currentState != phases.ImplementationExecuting {
		t.Errorf("Expected state to be ImplementationExecuting, got %s", currentState)
	}
}

func TestFullTransitionFlow_AllTasksComplete(t *testing.T) {
	sm := stateless.NewStateMachine(phases.ImplementationExecuting)

	data := &phasesSchema.ImplementationPhase{
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "completed"},
			{Id: "020", Name: "Task 2", Status: "completed"},
		},
	}

	phase := New(data, ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, phases.ReviewActive)

	// Transition to next phase
	sm.Fire(phases.EventAllTasksComplete)

	currentState := sm.MustState().(phases.State)
	if currentState != phases.ReviewActive {
		t.Errorf("Expected state to be ReviewActive, got %s", currentState)
	}
}
