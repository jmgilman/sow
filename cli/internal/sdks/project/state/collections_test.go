package state

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/schemas/project"
)

// PhaseCollection Tests

func TestPhaseCollection_Get_Success(t *testing.T) {
	// Arrange
	phases := PhaseCollection{
		"implementation": &Phase{
			PhaseState: project.PhaseState{
				Status:  "in_progress",
				Enabled: true,
			},
		},
		"review": &Phase{
			PhaseState: project.PhaseState{
				Status:  "pending",
				Enabled: true,
			},
		},
	}

	// Act
	phase, err := phases.Get("implementation")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if phase == nil {
		t.Fatal("Expected phase to be non-nil")
	}
	if phase.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got: %s", phase.Status)
	}
}

func TestPhaseCollection_Get_NotFound(t *testing.T) {
	// Arrange
	phases := PhaseCollection{
		"implementation": &Phase{
			PhaseState: project.PhaseState{
				Status:  "in_progress",
				Enabled: true,
			},
		},
	}

	// Act
	phase, err := phases.Get("nonexistent")

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if phase != nil {
		t.Error("Expected phase to be nil")
	}
	expectedMsg := "phase not found: nonexistent"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: '%s'", expectedMsg, err.Error())
	}
}

func TestPhaseCollection_Iterate(t *testing.T) {
	// Arrange
	phases := PhaseCollection{
		"planning":       &Phase{PhaseState: project.PhaseState{Status: "completed"}},
		"implementation": &Phase{PhaseState: project.PhaseState{Status: "in_progress"}},
		"review":         &Phase{PhaseState: project.PhaseState{Status: "pending"}},
	}

	// Act
	count := 0
	for name, phase := range phases {
		count++
		// Verify we can access both key and value
		if name == "" {
			t.Error("Phase name should not be empty")
		}
		if phase == nil {
			t.Error("Phase should not be nil")
		}
	}

	// Assert
	if count != 3 {
		t.Errorf("Expected 3 phases, got: %d", count)
	}
}

// ArtifactCollection Tests

func TestArtifactCollection_Get_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	artifacts := ArtifactCollection{
		{ArtifactState: project.ArtifactState{Type: "design_doc", Path: "design.md", Created_at: now}},
		{ArtifactState: project.ArtifactState{Type: "task_list", Path: "tasks.md", Created_at: now}},
	}

	// Act
	artifact, err := artifacts.Get(0)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if artifact == nil {
		t.Fatal("Expected artifact to be non-nil")
	}
	if artifact.Type != "design_doc" {
		t.Errorf("Expected type 'design_doc', got: %s", artifact.Type)
	}
}

func TestArtifactCollection_Get_OutOfRange(t *testing.T) {
	// Arrange
	now := time.Now()
	artifacts := ArtifactCollection{
		{ArtifactState: project.ArtifactState{Type: "design_doc", Path: "design.md", Created_at: now}},
	}

	// Act
	artifact, err := artifacts.Get(5)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if artifact != nil {
		t.Error("Expected artifact to be nil")
	}
	expectedMsg := "index out of range: 5 (length: 1)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: '%s'", expectedMsg, err.Error())
	}
}

func TestArtifactCollection_Get_NegativeIndex(t *testing.T) {
	// Arrange
	now := time.Now()
	artifacts := ArtifactCollection{
		{ArtifactState: project.ArtifactState{Type: "design_doc", Path: "design.md", Created_at: now}},
	}

	// Act
	artifact, err := artifacts.Get(-1)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if artifact != nil {
		t.Error("Expected artifact to be nil")
	}
	expectedMsg := "index out of range: -1 (length: 1)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: '%s'", expectedMsg, err.Error())
	}
}

func TestArtifactCollection_Add(t *testing.T) {
	// Arrange
	now := time.Now()
	artifacts := ArtifactCollection{
		{ArtifactState: project.ArtifactState{Type: "design_doc", Path: "design.md", Created_at: now}},
	}
	newArtifact := Artifact{ArtifactState: project.ArtifactState{Type: "task_list", Path: "tasks.md", Created_at: now}}

	// Act
	err := artifacts.Add(newArtifact)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(artifacts) != 2 {
		t.Errorf("Expected length 2, got: %d", len(artifacts))
	}
	if artifacts[1].Type != "task_list" {
		t.Errorf("Expected type 'task_list', got: %s", artifacts[1].Type)
	}
}

func TestArtifactCollection_Remove_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	artifacts := ArtifactCollection{
		{ArtifactState: project.ArtifactState{Type: "design_doc", Path: "design.md", Created_at: now}},
		{ArtifactState: project.ArtifactState{Type: "task_list", Path: "tasks.md", Created_at: now}},
		{ArtifactState: project.ArtifactState{Type: "review", Path: "review.md", Created_at: now}},
	}

	// Act
	err := artifacts.Remove(1)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(artifacts) != 2 {
		t.Errorf("Expected length 2, got: %d", len(artifacts))
	}
	// Verify correct item was removed and remaining items shifted
	if artifacts[0].Type != "design_doc" {
		t.Errorf("Expected first item to be 'design_doc', got: %s", artifacts[0].Type)
	}
	if artifacts[1].Type != "review" {
		t.Errorf("Expected second item to be 'review', got: %s", artifacts[1].Type)
	}
}

func TestArtifactCollection_Remove_OutOfRange(t *testing.T) {
	// Arrange
	now := time.Now()
	artifacts := ArtifactCollection{
		{ArtifactState: project.ArtifactState{Type: "design_doc", Path: "design.md", Created_at: now}},
	}

	// Act
	err := artifacts.Remove(5)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	expectedMsg := "index out of range: 5 (length: 1)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: '%s'", expectedMsg, err.Error())
	}
	// Verify collection unchanged
	if len(artifacts) != 1 {
		t.Errorf("Expected length to remain 1, got: %d", len(artifacts))
	}
}

// TaskCollection Tests

func TestTaskCollection_Get_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	tasks := TaskCollection{
		{TaskState: project.TaskState{Id: "010", Name: "Implement auth", Phase: "implementation", Status: "in_progress", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now}},
		{TaskState: project.TaskState{Id: "020", Name: "Write tests", Phase: "implementation", Status: "pending", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now}},
	}

	// Act
	task, err := tasks.Get("010")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if task == nil {
		t.Fatal("Expected task to be non-nil")
	}
	if task.Name != "Implement auth" {
		t.Errorf("Expected name 'Implement auth', got: %s", task.Name)
	}
}

func TestTaskCollection_Get_NotFound(t *testing.T) {
	// Arrange
	now := time.Now()
	tasks := TaskCollection{
		{TaskState: project.TaskState{Id: "010", Name: "Implement auth", Phase: "implementation", Status: "in_progress", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now}},
	}

	// Act
	task, err := tasks.Get("999")

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if task != nil {
		t.Error("Expected task to be nil")
	}
	expectedMsg := "task not found: 999"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: '%s'", expectedMsg, err.Error())
	}
}

func TestTaskCollection_Add(t *testing.T) {
	// Arrange
	now := time.Now()
	tasks := TaskCollection{
		{TaskState: project.TaskState{Id: "010", Name: "Implement auth", Phase: "implementation", Status: "in_progress", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now}},
	}
	newTask := Task{TaskState: project.TaskState{Id: "020", Name: "Write tests", Phase: "implementation", Status: "pending", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now}}

	// Act
	err := tasks.Add(newTask)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("Expected length 2, got: %d", len(tasks))
	}
	if tasks[1].Name != "Write tests" {
		t.Errorf("Expected name 'Write tests', got: %s", tasks[1].Name)
	}
}

func TestTaskCollection_Remove_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	tasks := TaskCollection{
		{TaskState: project.TaskState{Id: "010", Name: "Implement auth", Phase: "implementation", Status: "in_progress", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now}},
		{TaskState: project.TaskState{Id: "020", Name: "Write tests", Phase: "implementation", Status: "pending", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now}},
		{TaskState: project.TaskState{Id: "030", Name: "Deploy", Phase: "implementation", Status: "pending", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now}},
	}

	// Act
	err := tasks.Remove("020")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("Expected length 2, got: %d", len(tasks))
	}
	// Verify correct item was removed and remaining items shifted
	if tasks[0].Id != "010" {
		t.Errorf("Expected first task ID '010', got: %s", tasks[0].Id)
	}
	if tasks[1].Id != "030" {
		t.Errorf("Expected second task ID '030', got: %s", tasks[1].Id)
	}
}

func TestTaskCollection_Remove_NotFound(t *testing.T) {
	// Arrange
	now := time.Now()
	tasks := TaskCollection{
		{TaskState: project.TaskState{Id: "010", Name: "Implement auth", Phase: "implementation", Status: "in_progress", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now}},
	}

	// Act
	err := tasks.Remove("999")

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	expectedMsg := "task not found: 999"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: '%s'", expectedMsg, err.Error())
	}
	// Verify collection unchanged
	if len(tasks) != 1 {
		t.Errorf("Expected length to remain 1, got: %d", len(tasks))
	}
}
