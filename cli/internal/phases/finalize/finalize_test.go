package finalize

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/phases"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/qmuntal/stateless"
)

func TestNew(t *testing.T) {
	data := &phasesSchema.FinalizePhase{
		Enabled:         true,
		Status:          "pending",
		Created_at:      time.Now(),
		Project_deleted: false,
	}

	project := phases.ProjectInfo{
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
	phase := New(nil, phases.ProjectInfo{})

	if phase.EntryState() != statechart.FinalizeDocumentation {
		t.Errorf("Expected entry state to be FinalizeDocumentation, got %s", phase.EntryState())
	}
}

func TestMetadata(t *testing.T) {
	phase := New(nil, phases.ProjectInfo{})
	meta := phase.Metadata()

	if meta.Name != "finalize" {
		t.Errorf("Expected name to be 'finalize', got %s", meta.Name)
	}

	if len(meta.States) != 3 {
		t.Errorf("Expected 3 states, got %d", len(meta.States))
	}

	if meta.SupportsTasks {
		t.Error("Expected SupportsTasks to be false")
	}

	if meta.SupportsArtifacts {
		t.Error("Expected SupportsArtifacts to be false")
	}
}

func TestAddToMachine(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.FinalizeDocumentation)
	phase := New(nil, phases.ProjectInfo{})

	phase.AddToMachine(sm, statechart.NoProject)

	canFire, _ := sm.CanFire(statechart.EventDocumentationDone)
	if !canFire {
		t.Error("Expected EventDocumentationDone to be configured")
	}
}

func TestDocumentationAssessedGuard(t *testing.T) {
	data := &phasesSchema.FinalizePhase{}

	phase := New(data, phases.ProjectInfo{})

	// Should always return true
	if !phase.documentationAssessedGuard(context.Background()) {
		t.Error("Expected guard to always pass")
	}
}

func TestChecksAssessedGuard(t *testing.T) {
	data := &phasesSchema.FinalizePhase{}

	phase := New(data, phases.ProjectInfo{})

	// Should always return true
	if !phase.checksAssessedGuard(context.Background()) {
		t.Error("Expected guard to always pass")
	}
}

func TestProjectDeletedGuard_NotDeleted(t *testing.T) {
	data := &phasesSchema.FinalizePhase{
		Project_deleted: false,
	}

	phase := New(data, phases.ProjectInfo{})

	if phase.projectDeletedGuard(context.Background()) {
		t.Error("Expected guard to fail when project not deleted")
	}
}

func TestProjectDeletedGuard_Deleted(t *testing.T) {
	data := &phasesSchema.FinalizePhase{
		Project_deleted: true,
	}

	phase := New(data, phases.ProjectInfo{})

	if !phase.projectDeletedGuard(context.Background()) {
		t.Error("Expected guard to pass when project deleted")
	}
}

func TestPrepareTemplateData_NoUpdates(t *testing.T) {
	data := &phasesSchema.FinalizePhase{
		Documentation_updates: []string{},
		Artifacts_moved:       []struct {
			From string `json:"from"`
			To   string `json:"to"`
		}{},
		Project_deleted: false,
	}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)
	templateData := phase.prepareTemplateData()

	if templateData["HasDocumentationUpdates"] != false {
		t.Errorf("Expected HasDocumentationUpdates to be false, got %v", templateData["HasDocumentationUpdates"])
	}

	if templateData["HasArtifactsMoved"] != false {
		t.Errorf("Expected HasArtifactsMoved to be false, got %v", templateData["HasArtifactsMoved"])
	}

	if templateData["HasPRUrl"] != false {
		t.Errorf("Expected HasPRUrl to be false, got %v", templateData["HasPRUrl"])
	}
}

func TestPrepareTemplateData_WithUpdates(t *testing.T) {
	prURL := "https://github.com/user/repo/pull/123"
	data := &phasesSchema.FinalizePhase{
		Documentation_updates: []string{"README.md", "CHANGELOG.md"},
		Artifacts_moved: []struct {
			From string `json:"from"`
			To   string `json:"to"`
		}{
			{From: "phases/design/adr-001.md", To: ".sow/knowledge/adrs/adr-001.md"},
		},
		Pr_url:          &prURL,
		Project_deleted: true,
	}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)
	templateData := phase.prepareTemplateData()

	if templateData["HasDocumentationUpdates"] != true {
		t.Errorf("Expected HasDocumentationUpdates to be true, got %v", templateData["HasDocumentationUpdates"])
	}

	if templateData["HasArtifactsMoved"] != true {
		t.Errorf("Expected HasArtifactsMoved to be true, got %v", templateData["HasArtifactsMoved"])
	}

	if templateData["HasPRUrl"] != true {
		t.Errorf("Expected HasPRUrl to be true, got %v", templateData["HasPRUrl"])
	}

	if templateData["PRUrl"] != prURL {
		t.Errorf("Expected PRUrl to be %s, got %v", prURL, templateData["PRUrl"])
	}

	if templateData["ProjectDeleted"] != true {
		t.Errorf("Expected ProjectDeleted to be true, got %v", templateData["ProjectDeleted"])
	}
}

func TestRenderPrompt_Documentation(t *testing.T) {
	data := &phasesSchema.FinalizePhase{}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)
	prompt := phase.renderPrompt("documentation")

	if strings.Contains(prompt, "Error") {
		t.Errorf("Expected prompt to render without errors, got: %s", prompt)
	}

	if !strings.Contains(prompt, "Test Project") {
		t.Error("Expected prompt to contain project name")
	}
}

func TestRenderPrompt_Checks(t *testing.T) {
	data := &phasesSchema.FinalizePhase{}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)
	prompt := phase.renderPrompt("checks")

	if strings.Contains(prompt, "Error") {
		t.Errorf("Expected prompt to render without errors, got: %s", prompt)
	}
}

func TestRenderPrompt_Delete(t *testing.T) {
	data := &phasesSchema.FinalizePhase{}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)
	prompt := phase.renderPrompt("delete")

	if strings.Contains(prompt, "Error") {
		t.Errorf("Expected prompt to render without errors, got: %s", prompt)
	}
}

func TestFullTransitionFlow_ThreeStages(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.FinalizeDocumentation)

	data := &phasesSchema.FinalizePhase{
		Project_deleted: false,
	}

	phase := New(data, phases.ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, statechart.NoProject)

	// Stage 1: Documentation → Checks
	if err := sm.Fire(statechart.EventDocumentationDone); err != nil {
		t.Fatalf("Failed to fire EventDocumentationDone: %v", err)
	}
	currentState, ok := sm.MustState().(statechart.State)
	if !ok {
		t.Fatal("Failed to cast state to statechart.State")
	}
	if currentState != statechart.FinalizeChecks {
		t.Errorf("Expected state to be FinalizeChecks, got %s", currentState)
	}

	// Stage 2: Checks → Delete
	if err := sm.Fire(statechart.EventChecksDone); err != nil {
		t.Fatalf("Failed to fire EventChecksDone: %v", err)
	}
	currentState, ok = sm.MustState().(statechart.State)
	if !ok {
		t.Fatal("Failed to cast state to statechart.State")
	}
	if currentState != statechart.FinalizeDelete {
		t.Errorf("Expected state to be FinalizeDelete, got %s", currentState)
	}

	// Stage 3: Delete → NoProject (blocked by guard)
	canFire, _ := sm.CanFire(statechart.EventProjectDelete)
	if canFire {
		t.Error("Expected EventProjectDelete to be blocked when project not deleted")
	}

	// Set project_deleted flag
	data.Project_deleted = true

	// Now transition should work
	if err := sm.Fire(statechart.EventProjectDelete); err != nil {
		t.Fatalf("Failed to fire EventProjectDelete: %v", err)
	}
	currentState, ok = sm.MustState().(statechart.State)
	if !ok {
		t.Fatal("Failed to cast state to statechart.State")
	}
	if currentState != statechart.NoProject {
		t.Errorf("Expected state to be NoProject, got %s", currentState)
	}
}

func TestFullTransitionFlow_DeleteBlocked(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.FinalizeDelete)

	data := &phasesSchema.FinalizePhase{
		Project_deleted: false,
	}

	phase := New(data, phases.ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, statechart.NoProject)

	// Try to delete (should be blocked)
	canFire, _ := sm.CanFire(statechart.EventProjectDelete)
	if canFire {
		t.Error("Expected EventProjectDelete to be blocked when project_deleted is false")
	}
}
