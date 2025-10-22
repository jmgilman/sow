package design

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
	data := &phasesSchema.DesignPhase{
		Enabled:    true,
		Status:     "pending",
		Created_at: time.Now(),
		Artifacts:  []phasesSchema.Artifact{},
	}

	project := ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(true, data, project)

	if phase == nil {
		t.Fatal("Expected non-nil phase")
	}

	if !phase.optional {
		t.Error("Expected phase to be optional")
	}
}

func TestEntryState(t *testing.T) {
	phase := New(true, nil, ProjectInfo{})

	if phase.EntryState() != phases.DesignDecision {
		t.Errorf("Expected entry state to be DesignDecision, got %s", phase.EntryState())
	}
}

func TestMetadata(t *testing.T) {
	phase := New(true, nil, ProjectInfo{})
	meta := phase.Metadata()

	if meta.Name != "design" {
		t.Errorf("Expected name to be 'design', got %s", meta.Name)
	}

	if len(meta.States) != 2 {
		t.Errorf("Expected 2 states, got %d", len(meta.States))
	}

	if !meta.SupportsArtifacts {
		t.Error("Expected SupportsArtifacts to be true")
	}

	if meta.SupportsTasks {
		t.Error("Expected SupportsTasks to be false")
	}
}

func TestAddToMachine_Optional(t *testing.T) {
	sm := stateless.NewStateMachine(phases.DesignDecision)
	phase := New(true, nil, ProjectInfo{})

	phase.AddToMachine(sm, phases.ImplementationPlanning)

	canFire, _ := sm.CanFire(phases.EventEnableDesign)
	if !canFire {
		t.Error("Expected EventEnableDesign to be permitted")
	}

	canFire, _ = sm.CanFire(phases.EventSkipDesign)
	if !canFire {
		t.Error("Expected EventSkipDesign to be permitted for optional phase")
	}
}

func TestAddToMachine_Required(t *testing.T) {
	sm := stateless.NewStateMachine(phases.DesignDecision)
	phase := New(false, nil, ProjectInfo{})

	phase.AddToMachine(sm, phases.ImplementationPlanning)

	canFire, _ := sm.CanFire(phases.EventSkipDesign)
	if canFire {
		t.Error("Expected EventSkipDesign to NOT be permitted for required phase")
	}
}

func TestArtifactsApprovedGuard_NoArtifacts(t *testing.T) {
	data := &phasesSchema.DesignPhase{
		Artifacts: []phasesSchema.Artifact{},
	}

	phase := New(true, data, ProjectInfo{})

	if !phase.artifactsApprovedGuard(context.Background()) {
		t.Error("Expected guard to pass with no artifacts")
	}
}

func TestArtifactsApprovedGuard_AllApproved(t *testing.T) {
	data := &phasesSchema.DesignPhase{
		Artifacts: []phasesSchema.Artifact{
			{Path: "adr-001.md", Approved: true},
			{Path: "design.md", Approved: true},
		},
	}

	phase := New(true, data, ProjectInfo{})

	if !phase.artifactsApprovedGuard(context.Background()) {
		t.Error("Expected guard to pass when all artifacts approved")
	}
}

func TestArtifactsApprovedGuard_SomeUnapproved(t *testing.T) {
	data := &phasesSchema.DesignPhase{
		Artifacts: []phasesSchema.Artifact{
			{Path: "adr-001.md", Approved: true},
			{Path: "design.md", Approved: false},
		},
	}

	phase := New(true, data, ProjectInfo{})

	if phase.artifactsApprovedGuard(context.Background()) {
		t.Error("Expected guard to fail when some artifacts unapproved")
	}
}

func TestPrepareTemplateData(t *testing.T) {
	architectUsed := true
	data := &phasesSchema.DesignPhase{
		Architect_used: &architectUsed,
		Artifacts: []phasesSchema.Artifact{
			{Path: "adr-001.md", Approved: true},
			{Path: "design.md", Approved: false},
		},
	}

	project := ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(true, data, project)
	templateData := phase.prepareTemplateData()

	if templateData["ArchitectUsed"] != true {
		t.Errorf("Expected ArchitectUsed to be true, got %v", templateData["ArchitectUsed"])
	}

	if templateData["ArtifactCount"] != 2 {
		t.Errorf("Expected ArtifactCount to be 2, got %v", templateData["ArtifactCount"])
	}

	if templateData["ApprovedCount"] != 1 {
		t.Errorf("Expected ApprovedCount to be 1, got %v", templateData["ApprovedCount"])
	}
}

func TestRenderPrompt_Decision(t *testing.T) {
	project := ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(true, &phasesSchema.DesignPhase{}, project)
	prompt := phase.renderPrompt("decision")

	if strings.Contains(prompt, "Error") {
		t.Errorf("Expected prompt to render without errors, got: %s", prompt)
	}

	if !strings.Contains(prompt, "Test Project") {
		t.Error("Expected prompt to contain project name")
	}
}

func TestFullTransitionFlow(t *testing.T) {
	sm := stateless.NewStateMachine(phases.DesignDecision)

	data := &phasesSchema.DesignPhase{
		Enabled:   true,
		Status:    "pending",
		Artifacts: []phasesSchema.Artifact{},
	}

	phase := New(true, data, ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, phases.ImplementationPlanning)

	// Enable design
	sm.Fire(phases.EventEnableDesign)

	currentState := sm.MustState().(phases.State)
	if currentState != phases.DesignActive {
		t.Errorf("Expected state to be DesignActive, got %s", currentState)
	}

	// Complete design
	sm.Fire(phases.EventCompleteDesign)

	currentState = sm.MustState().(phases.State)
	if currentState != phases.ImplementationPlanning {
		t.Errorf("Expected state to be ImplementationPlanning, got %s", currentState)
	}
}
