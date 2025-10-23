package review

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
	data := &phasesSchema.ReviewPhase{
		Enabled:    true,
		Status:     "pending",
		Created_at: time.Now(),
		Iteration:  1,
		Reports:    []phasesSchema.ReviewReport{},
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

	if phase.EntryState() != statechart.ReviewActive {
		t.Errorf("Expected entry state to be ReviewActive, got %s", phase.EntryState())
	}
}

func TestMetadata(t *testing.T) {
	phase := New(nil, phases.ProjectInfo{})
	meta := phase.Metadata()

	if meta.Name != "review" {
		t.Errorf("Expected name to be 'review', got %s", meta.Name)
	}

	if len(meta.States) != 1 {
		t.Errorf("Expected 1 state, got %d", len(meta.States))
	}

	if meta.SupportsTasks {
		t.Error("Expected SupportsTasks to be false")
	}

	if meta.SupportsArtifacts {
		t.Error("Expected SupportsArtifacts to be false")
	}
}

func TestAddToMachine(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.ReviewActive)

	// Provide data that will make guard pass
	data := &phasesSchema.ReviewPhase{
		Reports: []phasesSchema.ReviewReport{
			{Path: "report-001.md", Assessment: "pass", Approved: true},
		},
	}

	phase := New(data, phases.ProjectInfo{})

	phase.AddToMachine(sm, statechart.FinalizeDocumentation)

	canFire, _ := sm.CanFire(statechart.EventReviewPass)
	if !canFire {
		t.Error("Expected EventReviewPass to be configured")
	}
}

func TestLatestReviewApprovedGuard_NoReports(t *testing.T) {
	data := &phasesSchema.ReviewPhase{
		Reports: []phasesSchema.ReviewReport{},
	}

	phase := New(data, phases.ProjectInfo{})

	if phase.latestReviewApprovedGuard(context.Background()) {
		t.Error("Expected guard to fail with no reports")
	}
}

func TestLatestReviewApprovedGuard_LatestApproved(t *testing.T) {
	data := &phasesSchema.ReviewPhase{
		Reports: []phasesSchema.ReviewReport{
			{Path: "report-001.md", Assessment: "pass", Approved: true},
		},
	}

	phase := New(data, phases.ProjectInfo{})

	if !phase.latestReviewApprovedGuard(context.Background()) {
		t.Error("Expected guard to pass with approved report")
	}
}

func TestLatestReviewApprovedGuard_LatestNotApproved(t *testing.T) {
	data := &phasesSchema.ReviewPhase{
		Reports: []phasesSchema.ReviewReport{
			{Path: "report-001.md", Assessment: "fail", Approved: false},
		},
	}

	phase := New(data, phases.ProjectInfo{})

	if phase.latestReviewApprovedGuard(context.Background()) {
		t.Error("Expected guard to fail with unapproved report")
	}
}

func TestLatestReviewFailedGuard_NoReports(t *testing.T) {
	data := &phasesSchema.ReviewPhase{
		Reports: []phasesSchema.ReviewReport{},
	}

	phase := New(data, phases.ProjectInfo{})

	if phase.LatestReviewFailedGuard(context.Background()) {
		t.Error("Expected guard to fail with no reports")
	}
}

func TestLatestReviewFailedGuard_LatestFailed(t *testing.T) {
	data := &phasesSchema.ReviewPhase{
		Reports: []phasesSchema.ReviewReport{
			{Path: "report-001.md", Assessment: "fail", Approved: false},
		},
	}

	phase := New(data, phases.ProjectInfo{})

	if !phase.LatestReviewFailedGuard(context.Background()) {
		t.Error("Expected guard to pass with failed report")
	}
}

func TestLatestReviewFailedGuard_LatestPassed(t *testing.T) {
	data := &phasesSchema.ReviewPhase{
		Reports: []phasesSchema.ReviewReport{
			{Path: "report-001.md", Assessment: "pass", Approved: true},
		},
	}

	phase := New(data, phases.ProjectInfo{})

	if phase.LatestReviewFailedGuard(context.Background()) {
		t.Error("Expected guard to fail with passing report")
	}
}

func TestPrepareTemplateData_FirstIteration(t *testing.T) {
	data := &phasesSchema.ReviewPhase{
		Iteration: 1,
		Reports:   []phasesSchema.ReviewReport{},
	}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)
	templateData := phase.prepareTemplateData()

	if templateData["ReviewIteration"] != int64(1) {
		t.Errorf("Expected ReviewIteration to be 1, got %v", templateData["ReviewIteration"])
	}

	if templateData["HasPreviousReview"] != false {
		t.Errorf("Expected HasPreviousReview to be false, got %v", templateData["HasPreviousReview"])
	}
}

func TestPrepareTemplateData_SecondIteration(t *testing.T) {
	data := &phasesSchema.ReviewPhase{
		Iteration: 2,
		Reports: []phasesSchema.ReviewReport{
			{Path: "report-001.md", Assessment: "fail", Approved: false},
		},
	}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)
	templateData := phase.prepareTemplateData()

	if templateData["ReviewIteration"] != int64(2) {
		t.Errorf("Expected ReviewIteration to be 2, got %v", templateData["ReviewIteration"])
	}

	if templateData["HasPreviousReview"] != true {
		t.Errorf("Expected HasPreviousReview to be true, got %v", templateData["HasPreviousReview"])
	}

	if templateData["PreviousAssessment"] != "fail" {
		t.Errorf("Expected PreviousAssessment to be 'fail', got %v", templateData["PreviousAssessment"])
	}
}

func TestRenderPrompt_Active(t *testing.T) {
	data := &phasesSchema.ReviewPhase{
		Iteration: 1,
		Reports:   []phasesSchema.ReviewReport{},
	}

	project := phases.ProjectInfo{
		Name:        "Test Project",
		Description: "Test Description",
		Branch:      "test-branch",
	}

	phase := New(data, project)
	prompt := phase.renderPrompt("active")

	if strings.Contains(prompt, "Error") {
		t.Errorf("Expected prompt to render without errors, got: %s", prompt)
	}

	if !strings.Contains(prompt, "Test Project") {
		t.Error("Expected prompt to contain project name")
	}
}

func TestFullTransitionFlow_Pass(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.ReviewActive)

	data := &phasesSchema.ReviewPhase{
		Iteration: 1,
		Reports: []phasesSchema.ReviewReport{
			{Path: "report-001.md", Assessment: "pass", Approved: true},
		},
	}

	phase := New(data, phases.ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, statechart.FinalizeDocumentation)

	// Transition to finalize
	if err := sm.Fire(statechart.EventReviewPass); err != nil {
		t.Fatalf("Failed to fire EventReviewPass: %v", err)
	}

	currentState, ok := sm.MustState().(statechart.State)
	if !ok {
		t.Fatal("Failed to cast state to statechart.State")
	}
	if currentState != statechart.FinalizeDocumentation {
		t.Errorf("Expected state to be FinalizeDocumentation, got %s", currentState)
	}
}

func TestFullTransitionFlow_PassBlocked(t *testing.T) {
	sm := stateless.NewStateMachine(statechart.ReviewActive)

	data := &phasesSchema.ReviewPhase{
		Iteration: 1,
		Reports: []phasesSchema.ReviewReport{
			{Path: "report-001.md", Assessment: "fail", Approved: false},
		},
	}

	phase := New(data, phases.ProjectInfo{Name: "Test"})
	phase.AddToMachine(sm, statechart.FinalizeDocumentation)

	// Try to transition to finalize (should be blocked)
	canFire, _ := sm.CanFire(statechart.EventReviewPass)
	if canFire {
		t.Error("Expected EventReviewPass to be blocked with unapproved report")
	}
}
