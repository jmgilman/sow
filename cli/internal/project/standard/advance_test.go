package standard

import (
	"errors"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// Helper function to create a project in a specific state.
func setupTestProjectInState(t *testing.T, state string) *StandardProject {
	t.Helper()

	ctx := setupTestRepo(t)
	now := time.Now()

	projectState := &projects.StandardProjectState{}
	projectState.Statechart.Current_state = state
	projectState.Project.Type = "standard"
	projectState.Project.Name = "test"
	projectState.Project.Branch = "test"
	projectState.Project.Description = "test"
	projectState.Project.Created_at = now
	projectState.Project.Updated_at = now

	projectState.Phases.Planning.Status = "completed"
	projectState.Phases.Planning.Enabled = true
	projectState.Phases.Planning.Created_at = now

	projectState.Phases.Implementation.Status = "in_progress"
	projectState.Phases.Implementation.Enabled = true
	projectState.Phases.Implementation.Created_at = now

	projectState.Phases.Review.Status = "pending"
	projectState.Phases.Review.Enabled = true
	projectState.Phases.Review.Created_at = now

	projectState.Phases.Finalize.Status = "pending"
	projectState.Phases.Finalize.Enabled = true
	projectState.Phases.Finalize.Created_at = now

	proj := New(projectState, ctx)

	return proj
}

// PlanningPhase.Advance() tests

func TestPlanningPhaseAdvance_Success(t *testing.T) {
	proj := setupTestProjectInState(t, "PlanningActive")

	// Add and approve task list artifact (prerequisite for planning completion)
	planningPhase, ok := proj.phases["planning"].(*PlanningPhase)
	if !ok {
		t.Fatal("Failed to get planning phase")
	}
	taskListType := "task_list"
	err := planningPhase.AddArtifact("task-list.md", domain.WithType(&taskListType))
	if err != nil {
		t.Fatalf("Failed to add artifact: %v", err)
	}

	err = planningPhase.ApproveArtifact("task-list.md")
	if err != nil {
		t.Fatalf("Failed to approve artifact: %v", err)
	}

	// Act: Advance planning phase
	err = planningPhase.Advance()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify state transitioned to ImplementationPlanning
	if proj.Machine().State() != ImplementationPlanning {
		t.Errorf("Expected state %s, got %s", ImplementationPlanning, proj.Machine().State())
	}
}

func TestPlanningPhaseAdvance_GuardFailure(t *testing.T) {
	proj := setupTestProjectInState(t, "PlanningActive")
	planningPhase, ok := proj.phases["planning"].(*PlanningPhase)
	if !ok {
		t.Fatal("Failed to get planning phase")
	}

	// Don't approve task list - guard should fail

	// Act: Advance planning phase
	err := planningPhase.Advance()

	// Assert
	if err == nil {
		t.Error("Expected error when task list not approved")
	}

	if !errors.Is(err, project.ErrCannotAdvance) {
		t.Errorf("Expected ErrCannotAdvance, got: %v", err)
	}

	// Verify state unchanged
	if proj.Machine().State() != PlanningActive {
		t.Errorf("Expected state unchanged at %s, got %s", PlanningActive, proj.Machine().State())
	}
}

// ImplementationPhase.Advance() tests

func TestImplementationPhaseAdvance_FromPlanning(t *testing.T) {
	proj := setupTestProjectInState(t, "ImplementationPlanning")
	implPhase, ok := proj.phases["implementation"].(*ImplementationPhase)
	if !ok {
		t.Fatal("Failed to get implementation phase")
	}

	// Add at least one task and approve tasks (prerequisite)
	_, err := implPhase.AddTask("Test task", domain.WithDescription("Test description"))
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Set tasks_approved flag directly (guard checks the typed field)
	tasksApproved := true
	proj.state.Phases.Implementation.Tasks_approved = &tasksApproved

	// Act: Advance implementation phase
	err = implPhase.Advance()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify state transitioned to ImplementationExecuting
	if proj.Machine().State() != ImplementationExecuting {
		t.Errorf("Expected state %s, got %s", ImplementationExecuting, proj.Machine().State())
	}
}

func TestImplementationPhaseAdvance_FromExecuting(t *testing.T) {
	proj := setupTestProjectInState(t, "ImplementationExecuting")
	implPhase, ok := proj.phases["implementation"].(*ImplementationPhase)
	if !ok {
		t.Fatal("Failed to get implementation phase")
	}

	// Add a task and mark it as completed (prerequisite)
	task, err := implPhase.AddTask("Test task",
		domain.WithDescription("Test description"),
		domain.WithStatus("completed"))
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Verify task is marked completed
	if task.Status() != "completed" {
		t.Fatalf("Expected task status to be completed, got: %s", task.Status())
	}

	// Act: Advance implementation phase
	err = implPhase.Advance()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify state transitioned to ReviewActive
	if proj.Machine().State() != ReviewActive {
		t.Errorf("Expected state %s, got %s", ReviewActive, proj.Machine().State())
	}
}

func TestImplementationPhaseAdvance_UnexpectedState(t *testing.T) {
	// Put implementation phase in a state it shouldn't be in
	proj := setupTestProjectInState(t, "PlanningActive")
	implPhase, ok := proj.phases["implementation"].(*ImplementationPhase)
	if !ok {
		t.Fatal("Failed to get implementation phase")
	}

	// Act: Try to advance from wrong state
	err := implPhase.Advance()

	// Assert
	if err == nil {
		t.Error("Expected error when in unexpected state")
	}

	if !errors.Is(err, project.ErrUnexpectedState) {
		t.Errorf("Expected ErrUnexpectedState, got: %v", err)
	}
}

func TestImplementationPhaseAdvance_GuardFailure_NoTasks(t *testing.T) {
	proj := setupTestProjectInState(t, "ImplementationPlanning")
	implPhase, ok := proj.phases["implementation"].(*ImplementationPhase)
	if !ok {
		t.Fatal("Failed to get implementation phase")
	}

	// Don't add any tasks - guard should fail

	// Act: Advance implementation phase
	err := implPhase.Advance()

	// Assert
	if err == nil {
		t.Error("Expected error when no tasks exist")
	}

	if !errors.Is(err, project.ErrCannotAdvance) {
		t.Errorf("Expected ErrCannotAdvance, got: %v", err)
	}
}

func TestImplementationPhaseAdvance_GuardFailure_IncompleteTask(t *testing.T) {
	proj := setupTestProjectInState(t, "ImplementationExecuting")
	implPhase, ok := proj.phases["implementation"].(*ImplementationPhase)
	if !ok {
		t.Fatal("Failed to get implementation phase")
	}

	// Add a task but don't complete it
	_, err := implPhase.AddTask("Test task",
		domain.WithDescription("Test description"),
		domain.WithStatus("in_progress"))
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Act: Advance implementation phase
	err = implPhase.Advance()

	// Assert
	if err == nil {
		t.Error("Expected error when task not completed")
	}

	if !errors.Is(err, project.ErrCannotAdvance) {
		t.Errorf("Expected ErrCannotAdvance, got: %v", err)
	}
}

// ReviewPhase.Advance() tests

func TestReviewPhaseAdvance_Pass(t *testing.T) {
	proj := setupTestProjectInState(t, "ReviewActive")
	reviewPhase, ok := proj.phases["review"].(*ReviewPhase)
	if !ok {
		t.Fatal("Failed to get review phase")
	}

	// Add and approve review artifact with "pass" assessment
	reviewType := "review"
	passAssessment := "pass"
	err := reviewPhase.AddArtifact("review-001.md",
		domain.WithType(&reviewType),
		domain.WithAssessment(&passAssessment))
	if err != nil {
		t.Fatalf("Failed to add artifact: %v", err)
	}

	err = reviewPhase.ApproveArtifact("review-001.md")
	if err != nil {
		t.Fatalf("Failed to approve artifact: %v", err)
	}

	// Act: Advance review phase
	err = reviewPhase.Advance()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify state transitioned to FinalizeDocumentation
	if proj.Machine().State() != FinalizeDocumentation {
		t.Errorf("Expected state %s, got %s", FinalizeDocumentation, proj.Machine().State())
	}
}

func TestReviewPhaseAdvance_Fail(t *testing.T) {
	proj := setupTestProjectInState(t, "ReviewActive")
	reviewPhase, ok := proj.phases["review"].(*ReviewPhase)
	if !ok {
		t.Fatal("Failed to get review phase")
	}

	// Add and approve review artifact with "fail" assessment
	reviewType := "review"
	failAssessment := "fail"
	err := reviewPhase.AddArtifact("review-001.md",
		domain.WithType(&reviewType),
		domain.WithAssessment(&failAssessment))
	if err != nil {
		t.Fatalf("Failed to add artifact: %v", err)
	}

	err = reviewPhase.ApproveArtifact("review-001.md")
	if err != nil {
		t.Fatalf("Failed to approve artifact: %v", err)
	}

	// Act: Advance review phase
	err = reviewPhase.Advance()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify state transitioned to ImplementationPlanning (loop back)
	if proj.Machine().State() != ImplementationPlanning {
		t.Errorf("Expected state %s, got %s", ImplementationPlanning, proj.Machine().State())
	}
}

func TestReviewPhaseAdvance_NoApprovedArtifact(t *testing.T) {
	proj := setupTestProjectInState(t, "ReviewActive")
	reviewPhase, ok := proj.phases["review"].(*ReviewPhase)
	if !ok {
		t.Fatal("Failed to get review phase")
	}

	// Don't add or approve any review artifact

	// Act: Advance review phase
	err := reviewPhase.Advance()

	// Assert
	if err == nil {
		t.Error("Expected error when no approved review artifact exists")
	}

	if !errors.Is(err, project.ErrCannotAdvance) {
		t.Errorf("Expected ErrCannotAdvance, got: %v", err)
	}
}

func TestReviewPhaseAdvance_InvalidAssessment(t *testing.T) {
	proj := setupTestProjectInState(t, "ReviewActive")
	reviewPhase, ok := proj.phases["review"].(*ReviewPhase)
	if !ok {
		t.Fatal("Failed to get review phase")
	}

	// Add review artifact with invalid assessment
	reviewType := "review"
	invalidAssessment := "maybe"
	err := reviewPhase.AddArtifact("review-001.md",
		domain.WithType(&reviewType),
		domain.WithAssessment(&invalidAssessment))
	if err != nil {
		t.Fatalf("Failed to add artifact: %v", err)
	}

	err = reviewPhase.ApproveArtifact("review-001.md")
	if err != nil {
		t.Fatalf("Failed to approve artifact: %v", err)
	}

	// Act: Advance review phase
	err = reviewPhase.Advance()

	// Assert
	if err == nil {
		t.Error("Expected error with invalid assessment")
	}

	if !errors.Is(err, project.ErrUnexpectedState) {
		t.Errorf("Expected ErrUnexpectedState, got: %v", err)
	}
}

func TestReviewPhaseAdvance_MissingAssessment(t *testing.T) {
	proj := setupTestProjectInState(t, "ReviewActive")
	reviewPhase, ok := proj.phases["review"].(*ReviewPhase)
	if !ok {
		t.Fatal("Failed to get review phase")
	}

	// Add review artifact without assessment
	reviewType := "review"
	err := reviewPhase.AddArtifact("review-001.md",
		domain.WithType(&reviewType))
	// No assessment
	if err != nil {
		t.Fatalf("Failed to add artifact: %v", err)
	}

	err = reviewPhase.ApproveArtifact("review-001.md")
	if err != nil {
		t.Fatalf("Failed to approve artifact: %v", err)
	}

	// Act: Advance review phase
	err = reviewPhase.Advance()

	// Assert
	if err == nil {
		t.Error("Expected error when assessment is missing")
	}

	if !errors.Is(err, project.ErrCannotAdvance) {
		t.Errorf("Expected ErrCannotAdvance, got: %v", err)
	}
}

// FinalizePhase.Advance() tests

func TestFinalizePhaseAdvance_FromDocumentation(t *testing.T) {
	proj := setupTestProjectInState(t, "FinalizeDocumentation")
	finalizePhase, ok := proj.phases["finalize"].(*FinalizePhase)
	if !ok {
		t.Fatal("Failed to get finalize phase")
	}

	// Act: Advance finalize phase
	err := finalizePhase.Advance()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify state transitioned to FinalizeChecks
	if proj.Machine().State() != FinalizeChecks {
		t.Errorf("Expected state %s, got %s", FinalizeChecks, proj.Machine().State())
	}
}

func TestFinalizePhaseAdvance_FromChecks(t *testing.T) {
	proj := setupTestProjectInState(t, "FinalizeChecks")
	finalizePhase, ok := proj.phases["finalize"].(*FinalizePhase)
	if !ok {
		t.Fatal("Failed to get finalize phase")
	}

	// Act: Advance finalize phase
	err := finalizePhase.Advance()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify state transitioned to FinalizeDelete
	if proj.Machine().State() != FinalizeDelete {
		t.Errorf("Expected state %s, got %s", FinalizeDelete, proj.Machine().State())
	}
}

func TestFinalizePhaseAdvance_FromDelete(t *testing.T) {
	proj := setupTestProjectInState(t, "FinalizeDelete")
	finalizePhase, ok := proj.phases["finalize"].(*FinalizePhase)
	if !ok {
		t.Fatal("Failed to get finalize phase")
	}

	// Set project_deleted flag (prerequisite)
	projectDeleted := true
	proj.state.Phases.Finalize.Project_deleted = &projectDeleted

	// Act: Advance finalize phase
	err := finalizePhase.Advance()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify state transitioned to NoProject
	if proj.Machine().State() != "NoProject" {
		t.Errorf("Expected state NoProject, got %s", proj.Machine().State())
	}
}

func TestFinalizePhaseAdvance_UnexpectedState(t *testing.T) {
	// Put finalize phase in a state it shouldn't be in
	proj := setupTestProjectInState(t, "PlanningActive")
	finalizePhase, ok := proj.phases["finalize"].(*FinalizePhase)
	if !ok {
		t.Fatal("Failed to get finalize phase")
	}

	// Act: Try to advance from wrong state
	err := finalizePhase.Advance()

	// Assert
	if err == nil {
		t.Error("Expected error when in unexpected state")
	}

	if !errors.Is(err, project.ErrUnexpectedState) {
		t.Errorf("Expected ErrUnexpectedState, got: %v", err)
	}
}

func TestFinalizePhaseAdvance_GuardFailure_ProjectNotDeleted(t *testing.T) {
	proj := setupTestProjectInState(t, "FinalizeDelete")
	finalizePhase, ok := proj.phases["finalize"].(*FinalizePhase)
	if !ok {
		t.Fatal("Failed to get finalize phase")
	}

	// Don't set project_deleted flag - guard should fail

	// Act: Advance finalize phase
	err := finalizePhase.Advance()

	// Assert
	if err == nil {
		t.Error("Expected error when project not deleted")
	}

	if !errors.Is(err, project.ErrCannotAdvance) {
		t.Errorf("Expected ErrCannotAdvance, got: %v", err)
	}
}
