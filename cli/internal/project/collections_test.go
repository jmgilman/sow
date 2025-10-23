package project

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
)

// setupTestRepo creates a temporary directory with an initialized git repository
// and returns a sow.Context for testing. The directory is automatically cleaned
// up when the test completes.
func setupTestRepo(t *testing.T) *sow.Context {
	t.Helper()

	// Create temp directory
	tmpDir := t.TempDir()

	cmdCtx := context.Background()

	// Initialize git repo
	cmd := exec.CommandContext(cmdCtx, "git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user (required for commits)
	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.email: %v", err)
	}

	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	// Disable GPG signing for tests
	cmd = exec.CommandContext(cmdCtx, "git", "config", "commit.gpgsign", "false")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to disable gpgsign: %v", err)
	}

	// Create initial commit
	cmd = exec.CommandContext(cmdCtx, "git", "commit", "--allow-empty", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Create .sow directory
	sowDir := filepath.Join(tmpDir, ".sow")
	if err := os.MkdirAll(sowDir, 0755); err != nil {
		t.Fatalf("Failed to create .sow directory: %v", err)
	}

	// Create sow context
	ctx, err := sow.NewContext(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}

	return ctx
}

// mockProject implements domain.Project for testing.
type mockProject struct {
	saveCount int
}

func (m *mockProject) Name() string                                        { return "test" }
func (m *mockProject) Branch() string                                      { return "test" }
func (m *mockProject) Description() string                                 { return "test" }
func (m *mockProject) Type() string                                        { return "test" }
func (m *mockProject) CurrentPhase() domain.Phase                          { return nil }
func (m *mockProject) Phase(_ string) (domain.Phase, error)               { return nil, nil }
func (m *mockProject) Machine() *statechart.Machine                        { return nil }
func (m *mockProject) InitialState() statechart.State                      { return statechart.NoProject }
func (m *mockProject) Save() error {
	m.saveCount++
	return nil
}
func (m *mockProject) Log(_, _ string, _ ...domain.LogOption) error       { return nil }
func (m *mockProject) InferTaskID() (string, error)                        { return "", nil }
func (m *mockProject) GetTask(_ string) (*domain.Task, error)              { return nil, nil }
func (m *mockProject) CreatePullRequest(_ string) (string, error)          { return "", nil }
func (m *mockProject) ReadYAML(_ string, _ interface{}) error              { return nil }
func (m *mockProject) WriteYAML(_ string, _ interface{}) error             { return nil }
func (m *mockProject) ReadFile(_ string) ([]byte, error)                   { return nil, nil }
func (m *mockProject) WriteFile(_ string, _ []byte) error                  { return nil }

// TestArtifactCollectionAdd verifies artifacts can be added with metadata.
func TestArtifactCollectionAdd(t *testing.T) {
	now := time.Now()
	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Artifacts:  []phasesSchema.Artifact{},
	}

	proj := &mockProject{}
	collection := NewArtifactCollection(phaseState, proj)

	// Add artifact with metadata
	metadata := map[string]interface{}{
		"type":       "review",
		"assessment": "pass",
	}
	err := collection.Add("test.md", domain.WithMetadata(metadata))
	if err != nil {
		t.Fatalf("Failed to add artifact: %v", err)
	}

	// Verify artifact was added
	if len(phaseState.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(phaseState.Artifacts))
	}

	artifact := phaseState.Artifacts[0]
	if artifact.Path != "test.md" {
		t.Errorf("Expected path 'test.md', got '%s'", artifact.Path)
	}

	if artifact.Approved {
		t.Error("Expected artifact to be unapproved initially")
	}

	if artifact.Metadata["type"] != "review" {
		t.Errorf("Expected type 'review', got %v", artifact.Metadata["type"])
	}

	// Verify Save was called
	if proj.saveCount != 1 {
		t.Errorf("Expected Save to be called once, called %d times", proj.saveCount)
	}
}

// TestArtifactCollectionApprove verifies artifacts can be approved.
func TestArtifactCollectionApprove(t *testing.T) {
	now := time.Now()
	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Artifacts: []phasesSchema.Artifact{
			{
				Path:       "test.md",
				Approved:   false,
				Created_at: now,
			},
		},
	}

	proj := &mockProject{}
	collection := NewArtifactCollection(phaseState, proj)

	// Approve artifact
	err := collection.Approve("test.md")
	if err != nil {
		t.Fatalf("Failed to approve artifact: %v", err)
	}

	// Verify artifact is approved
	if !phaseState.Artifacts[0].Approved {
		t.Error("Expected artifact to be approved")
	}

	// Verify Save was called
	if proj.saveCount != 1 {
		t.Errorf("Expected Save to be called once, called %d times", proj.saveCount)
	}
}

// TestArtifactCollectionList verifies artifacts can be listed.
func TestArtifactCollectionList(t *testing.T) {
	now := time.Now()
	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Artifacts: []phasesSchema.Artifact{
			{Path: "test1.md", Approved: false, Created_at: now},
			{Path: "test2.md", Approved: true, Created_at: now},
		},
	}

	proj := &mockProject{}
	collection := NewArtifactCollection(phaseState, proj)

	// List artifacts
	artifacts := collection.List()
	if len(artifacts) != 2 {
		t.Fatalf("Expected 2 artifacts, got %d", len(artifacts))
	}

	// Verify pointers point to actual state
	if artifacts[0].Path != "test1.md" {
		t.Errorf("Expected first artifact path 'test1.md', got '%s'", artifacts[0].Path)
	}
}

// TestArtifactCollectionAllApproved verifies AllApproved check.
func TestArtifactCollectionAllApproved(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		artifacts []phasesSchema.Artifact
		expected  bool
	}{
		{
			name:      "no artifacts",
			artifacts: []phasesSchema.Artifact{},
			expected:  true,
		},
		{
			name: "all approved",
			artifacts: []phasesSchema.Artifact{
				{Path: "test1.md", Approved: true, Created_at: now},
				{Path: "test2.md", Approved: true, Created_at: now},
			},
			expected: true,
		},
		{
			name: "some unapproved",
			artifacts: []phasesSchema.Artifact{
				{Path: "test1.md", Approved: true, Created_at: now},
				{Path: "test2.md", Approved: false, Created_at: now},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phaseState := &phasesSchema.Phase{
				Status:     "in_progress",
				Created_at: now,
				Enabled:    true,
				Artifacts:  tt.artifacts,
			}

			proj := &mockProject{}
			collection := NewArtifactCollection(phaseState, proj)

			if got := collection.AllApproved(); got != tt.expected {
				t.Errorf("AllApproved() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestTaskCollectionAdd verifies tasks can be added.
func TestTaskCollectionAdd(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()
	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Tasks:      []phasesSchema.Task{},
	}

	proj := &mockProject{}
	collection := NewTaskCollection(phaseState, proj, ctx)

	// Add task
	task, err := collection.Add("Test task",
		domain.WithDescription("Test description"),
		domain.WithAgent("implementer"),
	)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	if task == nil {
		t.Fatal("Expected non-nil task")
	}

	// Verify task was added to phase state
	if len(phaseState.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(phaseState.Tasks))
	}

	phaseTask := phaseState.Tasks[0]
	if phaseTask.Name != "Test task" {
		t.Errorf("Expected name 'Test task', got '%s'", phaseTask.Name)
	}

	if phaseTask.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", phaseTask.Status)
	}

	// Verify ID was generated (gap-numbered)
	if phaseTask.Id == "" {
		t.Error("Expected non-empty task ID")
	}

	// Verify Save was called
	if proj.saveCount != 1 {
		t.Errorf("Expected Save to be called once, called %d times", proj.saveCount)
	}
}

// TestTaskCollectionList verifies tasks can be listed.
func TestTaskCollectionList(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()
	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "pending", Parallel: false},
			{Id: "020", Name: "Task 2", Status: "completed", Parallel: false},
		},
	}

	proj := &mockProject{}
	collection := NewTaskCollection(phaseState, proj, ctx)

	// List tasks
	tasks := collection.List()
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}

	// Verify task IDs
	if tasks[0].ID != "010" {
		t.Errorf("Expected first task ID '010', got '%s'", tasks[0].ID)
	}
	if tasks[1].ID != "020" {
		t.Errorf("Expected second task ID '020', got '%s'", tasks[1].ID)
	}
}

// TestTaskCollectionApprove verifies task approval sets metadata.
func TestTaskCollectionApprove(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()
	var err error
	phaseState := &phasesSchema.Phase{
		Status:     "pending",
		Created_at: now,
		Enabled:    true,
		Tasks: []phasesSchema.Task{
			{Id: "010", Name: "Task 1", Status: "pending", Parallel: false},
		},
		Metadata: make(map[string]interface{}),
	}

	proj := &mockProject{}
	collection := NewTaskCollection(phaseState, proj, ctx)

	// Approve tasks
	err = collection.Approve()
	if err != nil {
		t.Fatalf("Failed to approve tasks: %v", err)
	}

	// Verify metadata was set
	if approved, ok := phaseState.Metadata["tasks_approved"].(bool); !ok || !approved {
		t.Error("Expected tasks_approved metadata to be true")
	}

	// Verify status changed to in_progress
	if phaseState.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", phaseState.Status)
	}

	// Verify Save was called
	if proj.saveCount != 1 {
		t.Errorf("Expected Save to be called once, called %d times", proj.saveCount)
	}
}
