package discovery

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
	data := &phasesSchema.DiscoveryPhase{
		Enabled:    true,
		Status:     "pending",
		Created_at: time.Now(),
		Artifacts:  []phasesSchema.Artifact{},
	}

	project := phases.ProjectInfo{
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

	if phase.data != data {
		t.Error("Expected phase data to match")
	}

	if phase.project.Name != "Test Project" {
		t.Errorf("Expected project name to be 'Test Project', got %s", phase.project.Name)
	}
}

func TestEntryState(t *testing.T) {
	phase := New(true, nil, phases.ProjectInfo{})

	if phase.EntryState() != statechart.DiscoveryDecision {
		t.Errorf("Expected entry state to be DiscoveryDecision, got %s", phase.EntryState())
	}
}

func TestMetadata(t *testing.T) {
	phase := New(true, nil, phases.ProjectInfo{})
	meta := phase.Metadata()

	if meta.Name != "discovery" {
		t.Errorf("Expected name to be 'discovery', got %s", meta.Name)
	}

	if len(meta.States) != 2 {
		t.Errorf("Expected 2 states, got %d", len(meta.States))
	}

	expectedStates := map[statechart.State]bool{
		statechart.DiscoveryDecision: true,
		statechart.DiscoveryActive:   true,
	}

	for _, state := range meta.States {
		if !expectedStates[state] {
			t.Errorf("Unexpected state: %s", state)
		}
	}

	if meta.SupportsTasks {
		t.Error("Expected SupportsTasks to be false")
	}

	if !meta.SupportsArtifacts {
		t.Error("Expected SupportsArtifacts to be true")
	}

	if len(meta.CustomFields) != 1 {
		t.Errorf("Expected 1 custom field, got %d", len(meta.CustomFields))
	}

	if meta.CustomFields[0].Name != "discovery_type" {
		t.Errorf("Expected custom field name to be 'discovery_type', got %s", meta.CustomFields[0].Name)
	}

	if meta.CustomFields[0].Type != phases.StringField {
		t.Errorf("Expected custom field type to be StringField, got %s", meta.CustomFields[0].Type)
	}
}

func TestAddToMachine_Optional(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.DiscoveryDecision)
	phase := New(true, nil, phases.ProjectInfo{})

	phase.AddToMachine(sm, statechart.DesignDecision)

	// Verify DiscoveryDecision permits EventEnableDiscovery
	canFire, err := sm.CanFire(statechart.EventEnableDiscovery)
	if err != nil {
		t.Fatalf("Error checking EventEnableDiscovery: %v", err)
	}
	if !canFire {
		t.Error("Expected EventEnableDiscovery to be permitted")
	}

	// Verify DiscoveryDecision permits EventSkipDiscovery (optional)
	canFire, err = sm.CanFire(statechart.EventSkipDiscovery)
	if err != nil {
		t.Fatalf("Error checking EventSkipDiscovery: %v", err)
	}
	if !canFire {
		t.Error("Expected EventSkipDiscovery to be permitted for optional phase")
	}
}

func TestAddToMachine_Required(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.DiscoveryDecision)
	phase := New(false, nil, phases.ProjectInfo{}) // Not optional

	phase.AddToMachine(sm, statechart.DesignDecision)

	// Verify DiscoveryDecision permits EventEnableDiscovery
	canFire, err := sm.CanFire(statechart.EventEnableDiscovery)
	if err != nil {
		t.Fatalf("Error checking EventEnableDiscovery: %v", err)
	}
	if !canFire {
		t.Error("Expected EventEnableDiscovery to be permitted")
	}

	// Verify DiscoveryDecision does NOT permit EventSkipDiscovery (required)
	canFire, err = sm.CanFire(statechart.EventSkipDiscovery)
	if err != nil {
		t.Fatalf("Error checking EventSkipDiscovery: %v", err)
	}
	if canFire {
		t.Error("Expected EventSkipDiscovery to NOT be permitted for required phase")
	}
}

func TestAddToMachine_ActiveState(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.DiscoveryActive)

	data := &phasesSchema.DiscoveryPhase{
		Enabled:   true,
		Status:    "in_progress",
		Artifacts: []phasesSchema.Artifact{},
	}

	phase := New(true, data, phases.ProjectInfo{})
	phase.AddToMachine(sm, statechart.DesignDecision)

	// Verify DiscoveryActive permits EventCompleteDiscovery (with guard)
	canFire, err := sm.CanFire(statechart.EventCompleteDiscovery)
	if err != nil {
		t.Fatalf("Error checking EventCompleteDiscovery: %v", err)
	}

	// Should be true since no artifacts (guard passes)
	if !canFire {
		t.Error("Expected EventCompleteDiscovery to be permitted when no artifacts")
	}
}

func TestArtifactsApprovedGuard_NoArtifacts(t *testing.T) {
	data := &phasesSchema.DiscoveryPhase{
		Artifacts: []phasesSchema.Artifact{},
	}

	phase := New(true, data, phases.ProjectInfo{})

	if !phase.artifactsApprovedGuard(context.Background()) {
		t.Error("Expected guard to pass with no artifacts")
	}
}

func TestArtifactsApprovedGuard_AllApproved(t *testing.T) {
	data := &phasesSchema.DiscoveryPhase{
		Artifacts: []phasesSchema.Artifact{
			{Path: "artifact1.md", Approved: true},
			{Path: "artifact2.md", Approved: true},
		},
	}

	phase := New(true, data, phases.ProjectInfo{})

	if !phase.artifactsApprovedGuard(context.Background()) {
		t.Error("Expected guard to pass when all artifacts approved")
	}
}

func TestArtifactsApprovedGuard_SomeUnapproved(t *testing.T) {
	data := &phasesSchema.DiscoveryPhase{
		Artifacts: []phasesSchema.Artifact{
			{Path: "artifact1.md", Approved: true},
			{Path: "artifact2.md", Approved: false},
		},
	}

	phase := New(true, data, phases.ProjectInfo{})

	if phase.artifactsApprovedGuard(context.Background()) {
		t.Error("Expected guard to fail when some artifacts unapproved")
	}
}

func TestArtifactsApprovedGuard_NilData(t *testing.T) {
	phase := New(true, nil, phases.ProjectInfo{})

	if phase.artifactsApprovedGuard(context.Background()) {
		t.Error("Expected guard to fail when data is nil")
	}
}

func TestPrepareTemplateData(t *testing.T) {
	discoveryType := "bug"
	data := &phasesSchema.DiscoveryPhase{
		Discovery_type: &discoveryType,
		Artifacts: []phasesSchema.Artifact{
			{Path: "artifact1.md", Approved: true},
			{Path: "artifact2.md", Approved: false},
		},
	}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(true, data, project)
	templateData := phase.prepareTemplateData()

	// Check project info
	if templateData["ProjectName"] != "Test Project" {
		t.Errorf("Expected ProjectName to be 'Test Project', got %v", templateData["ProjectName"])
	}

	if templateData["ProjectDescription"] != "Test Description" {
		t.Errorf("Expected ProjectDescription to be 'Test Description', got %v", templateData["ProjectDescription"])
	}

	if templateData["ProjectBranch"] != "test-branch" {
		t.Errorf("Expected ProjectBranch to be 'test-branch', got %v", templateData["ProjectBranch"])
	}

	// Check discovery type
	if templateData["DiscoveryType"] != "bug" {
		t.Errorf("Expected DiscoveryType to be 'bug', got %v", templateData["DiscoveryType"])
	}

	// Check artifact counts
	if templateData["ArtifactCount"] != 2 {
		t.Errorf("Expected ArtifactCount to be 2, got %v", templateData["ArtifactCount"])
	}

	if templateData["ApprovedCount"] != 1 {
		t.Errorf("Expected ApprovedCount to be 1, got %v", templateData["ApprovedCount"])
	}
}

func TestRenderPrompt_Decision(t *testing.T) {
	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(true, &phasesSchema.DiscoveryPhase{}, project)
	prompt := phase.renderPrompt("decision")

	// Verify template rendered successfully
	if strings.Contains(prompt, "Error") {
		t.Errorf("Expected prompt to render without errors, got: %s", prompt)
	}

	// Verify project name appears in prompt
	if !strings.Contains(prompt, "Test Project") {
		t.Error("Expected prompt to contain project name")
	}

	// Verify prompt contains key content
	if !strings.Contains(prompt, "DISCOVERY") {
		t.Error("Expected prompt to mention DISCOVERY")
	}
}

func TestRenderPrompt_Active(t *testing.T) {
	discoveryType := "feature"
	data := &phasesSchema.DiscoveryPhase{
		Discovery_type: &discoveryType,
		Artifacts: []phasesSchema.Artifact{
			{Path: "research.md", Approved: true},
		},
	}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(true, data, project)
	prompt := phase.renderPrompt("active")

	// Verify template rendered successfully
	if strings.Contains(prompt, "Error") {
		t.Errorf("Expected prompt to render without errors, got: %s", prompt)
	}

	// Verify project name appears in prompt
	if !strings.Contains(prompt, "Test Project") {
		t.Error("Expected prompt to contain project name")
	}

	// Verify discovery type appears in prompt
	if !strings.Contains(prompt, "feature") {
		t.Error("Expected prompt to contain discovery type")
	}
}

func TestFullTransitionFlow_Optional(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.DiscoveryDecision)

	data := &phasesSchema.DiscoveryPhase{
		Enabled:   true,
		Status:    "pending",
		Artifacts: []phasesSchema.Artifact{},
	}

	phase := New(true, data, phases.ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, statechart.DesignDecision)

	// Test skip path
	err := sm.Fire(statechart.EventSkipDiscovery)
	if err != nil {
		t.Fatalf("Error firing EventSkipDiscovery: %v", err)
	}

	currentState, ok := sm.MustState().(statechart.State)
	if !ok {
		t.Fatal("Failed to cast state to statechart.State")
	}
	if currentState != statechart.DesignDecision {
		t.Errorf("Expected state to be DesignDecision after skip, got %s", currentState)
	}
}

func TestFullTransitionFlow_EnableAndComplete(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.DiscoveryDecision)

	data := &phasesSchema.DiscoveryPhase{
		Enabled:   true,
		Status:    "pending",
		Artifacts: []phasesSchema.Artifact{},
	}

	phase := New(true, data, phases.ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, statechart.DesignDecision)

	// Enable discovery
	err := sm.Fire(statechart.EventEnableDiscovery)
	if err != nil {
		t.Fatalf("Error firing EventEnableDiscovery: %v", err)
	}

	currentState, ok := sm.MustState().(statechart.State)
	if !ok {
		t.Fatal("Failed to cast state to statechart.State")
	}
	if currentState != statechart.DiscoveryActive {
		t.Errorf("Expected state to be DiscoveryActive after enable, got %s", currentState)
	}

	// Complete discovery (no artifacts, guard passes)
	err = sm.Fire(statechart.EventCompleteDiscovery)
	if err != nil {
		t.Fatalf("Error firing EventCompleteDiscovery: %v", err)
	}

	currentState, ok = sm.MustState().(statechart.State)
	if !ok {
		t.Fatal("Failed to cast state to statechart.State")
	}
	if currentState != statechart.DesignDecision {
		t.Errorf("Expected state to be DesignDecision after complete, got %s", currentState)
	}
}

func TestFullTransitionFlow_CompleteWithApprovedArtifacts(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.DiscoveryActive)

	data := &phasesSchema.DiscoveryPhase{
		Enabled: true,
		Status:  "in_progress",
		Artifacts: []phasesSchema.Artifact{
			{Path: "research.md", Approved: true},
		},
	}

	phase := New(true, data, phases.ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, statechart.DesignDecision)

	// Complete discovery (artifacts approved, guard passes)
	err := sm.Fire(statechart.EventCompleteDiscovery)
	if err != nil {
		t.Fatalf("Error firing EventCompleteDiscovery: %v", err)
	}

	currentState, ok := sm.MustState().(statechart.State)
	if !ok {
		t.Fatal("Failed to cast state to statechart.State")
	}
	if currentState != statechart.DesignDecision {
		t.Errorf("Expected state to be DesignDecision after complete, got %s", currentState)
	}
}

func TestFullTransitionFlow_CompleteBlockedByUnapprovedArtifacts(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.DiscoveryActive)

	data := &phasesSchema.DiscoveryPhase{
		Enabled: true,
		Status:  "in_progress",
		Artifacts: []phasesSchema.Artifact{
			{Path: "research.md", Approved: false},
		},
	}

	phase := New(true, data, phases.ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, statechart.DesignDecision)

	// Try to complete discovery (artifacts not approved, guard fails)
	canFire, err := sm.CanFire(statechart.EventCompleteDiscovery)
	if err != nil {
		t.Fatalf("Error checking EventCompleteDiscovery: %v", err)
	}

	if canFire {
		t.Error("Expected EventCompleteDiscovery to be blocked by unapproved artifacts")
	}
}
