package loader

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/sow"
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

// TestLoad_StandardProject verifies standard projects load correctly.
func TestLoad_StandardProject(t *testing.T) {
	ctx := setupTestRepo(t)

	// Create a standard project
	proj, err := Create(ctx, "test-project", "Test description")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Verify it's a standard project
	if proj.Type() != "standard" {
		t.Errorf("Expected type 'standard', got %q", proj.Type())
	}

	// Load the project
	loaded, err := Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	// Verify loaded project is standard
	if loaded.Type() != "standard" {
		t.Errorf("Expected loaded type 'standard', got %q", loaded.Type())
	}

	if loaded.Name() != "test-project" {
		t.Errorf("Expected name 'test-project', got %q", loaded.Name())
	}
}

// TestLoad_UnimplementedType verifies error messages for unimplemented project types.
func TestLoad_UnimplementedType(t *testing.T) {
	tests := []struct {
		name        string
		projectType string
		expectedErr string
	}{
		{
			name:        "exploration type not implemented",
			projectType: "exploration",
			expectedErr: "exploration project type not yet implemented",
		},
		{
			name:        "design type not implemented",
			projectType: "design",
			expectedErr: "design project type not yet implemented",
		},
		{
			name:        "breakdown type not implemented",
			projectType: "breakdown",
			expectedErr: "breakdown project type not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupTestRepo(t)

			// Create a standard project first
			_, err := Create(ctx, "test-project", "Test description")
			if err != nil {
				t.Fatalf("Failed to create project: %v", err)
			}

			// Manually modify the state to set the project type
			// This simulates what would happen if we had created a non-standard project
			stateData, err := ctx.FS().ReadFile("project/state.yaml")
			if err != nil {
				t.Fatalf("Failed to read state: %v", err)
			}

			// Replace project type
			modifiedState := strings.Replace(string(stateData), "type: standard", "type: "+tt.projectType, 1)
			if err := ctx.FS().WriteFile("project/state.yaml", []byte(modifiedState), 0644); err != nil {
				t.Fatalf("Failed to write modified state: %v", err)
			}

			// Try to load - should fail with appropriate error
			_, err = Load(ctx)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestLoad_UnknownType verifies error for unknown project types.
func TestLoad_UnknownType(t *testing.T) {
	ctx := setupTestRepo(t)

	// Create a standard project first
	_, err := Create(ctx, "test-project", "Test description")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Manually set unknown project type
	stateData, err := ctx.FS().ReadFile("project/state.yaml")
	if err != nil {
		t.Fatalf("Failed to read state: %v", err)
	}

	modifiedState := strings.Replace(string(stateData), "type: standard", "type: unknown-type", 1)
	if err := ctx.FS().WriteFile("project/state.yaml", []byte(modifiedState), 0644); err != nil {
		t.Fatalf("Failed to write modified state: %v", err)
	}

	// Try to load - should fail
	_, err = Load(ctx)
	if err == nil {
		t.Fatal("Expected error for unknown type, got nil")
	}

	if !strings.Contains(err.Error(), "unknown project type") {
		t.Errorf("Expected error containing 'unknown project type', got %q", err.Error())
	}
}

// TestCreate_DetectsTypeFromBranch verifies Create detects type from branch name.
func TestCreate_DetectsTypeFromBranch(t *testing.T) {
	tests := []struct {
		name          string
		branchName    string
		expectedType  string
		expectError   bool
		errorContains string
	}{
		{
			name:         "standard branch creates standard project",
			branchName:   "feat/new-feature",
			expectedType: "standard",
			expectError:  false,
		},
		{
			name:         "main branch creates standard project",
			branchName:   "main",
			expectedType: "standard",
			expectError:  false,
		},
		{
			name:          "exploration branch not yet implemented",
			branchName:    "explore/auth",
			expectError:   true,
			errorContains: "exploration",
		},
		{
			name:          "design branch not yet implemented",
			branchName:    "design/api",
			expectError:   true,
			errorContains: "design",
		},
		{
			name:          "breakdown branch not yet implemented",
			branchName:    "breakdown/features",
			expectError:   true,
			errorContains: "breakdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupTestRepo(t)

			// Create the specified branch
			cmdCtx := context.Background()
			cmd := exec.CommandContext(cmdCtx, "git", "checkout", "-b", tt.branchName)
			cmd.Dir = ctx.RepoRoot()
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to create branch: %v", err)
			}

			// Try to create project
			proj, err := Create(ctx, "test-project", "Test description")

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			// No error expected
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if proj.Type() != tt.expectedType {
				t.Errorf("Expected type %q, got %q", tt.expectedType, proj.Type())
			}
		})
	}
}

// TestLoad_NoProject verifies error when no project exists.
func TestLoad_NoProject(t *testing.T) {
	ctx := setupTestRepo(t)

	_, err := Load(ctx)
	if err == nil {
		t.Fatal("Expected error when loading non-existent project, got nil")
	}

	if !errors.Is(err, project.ErrNoProject) {
		t.Errorf("Expected project.ErrNoProject, got %v", err)
	}
}

// TestCreate_ValidatesName verifies kebab-case validation.
func TestCreate_ValidatesName(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		expectError bool
	}{
		{
			name:        "valid kebab-case",
			projectName: "my-project",
			expectError: false,
		},
		{
			name:        "valid with numbers",
			projectName: "project-123",
			expectError: false,
		},
		{
			name:        "invalid uppercase",
			projectName: "My-Project",
			expectError: true,
		},
		{
			name:        "invalid spaces",
			projectName: "my project",
			expectError: true,
		},
		{
			name:        "invalid underscore",
			projectName: "my_project",
			expectError: true,
		},
		{
			name:        "invalid double hyphen",
			projectName: "my--project",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupTestRepo(t)

			_, err := Create(ctx, tt.projectName, "Test description")

			if tt.expectError && err == nil {
				t.Error("Expected error for invalid name, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestCreate_CreatesDirectoryStructure verifies directory creation.
func TestCreate_CreatesDirectoryStructure(t *testing.T) {
	ctx := setupTestRepo(t)

	_, err := Create(ctx, "test-project", "Test description")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Verify directories exist
	expectedDirs := []string{
		"project",
		"project/context",
		"project/phases/implementation",
		"project/phases/implementation/tasks",
		"project/phases/review",
		"project/phases/review/reports",
		"project/phases/finalize",
	}

	for _, dir := range expectedDirs {
		exists, err := ctx.FS().Exists(dir)
		if err != nil {
			t.Fatalf("Error checking directory %s: %v", dir, err)
		}
		if !exists {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}

	// Verify log files exist
	expectedFiles := []string{
		"project/log.md",
		"project/phases/implementation/log.md",
		"project/phases/review/log.md",
		"project/phases/finalize/log.md",
	}

	for _, file := range expectedFiles {
		exists, err := ctx.FS().Exists(file)
		if err != nil {
			t.Fatalf("Error checking file %s: %v", file, err)
		}
		if !exists {
			t.Errorf("Expected file %s to exist", file)
		}
	}
}

// TestCreate_InitializesState verifies initial state is correct.
func TestCreate_InitializesState(t *testing.T) {
	ctx := setupTestRepo(t)

	proj, err := Create(ctx, "test-project", "Test description")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Verify state
	if proj.Name() != "test-project" {
		t.Errorf("Expected name 'test-project', got %q", proj.Name())
	}

	if proj.Description() != "Test description" {
		t.Errorf("Expected description 'Test description', got %q", proj.Description())
	}

	if proj.Type() != "standard" {
		t.Errorf("Expected type 'standard', got %q", proj.Type())
	}

	// Verify state machine
	machine := proj.Machine()
	if machine == nil {
		t.Fatal("Expected non-nil machine")
	}

	// Verify initial state (should be PlanningActive for standard projects)
	state := machine.State()
	if state == "" {
		t.Error("Expected non-empty initial state")
	}
}

// TestDelete_RemovesProject verifies Delete removes the project directory.
func TestDelete_RemovesProject(t *testing.T) {
	ctx := setupTestRepo(t)

	// Create project
	_, err := Create(ctx, "test-project", "Test description")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Verify it exists
	if !Exists(ctx) {
		t.Fatal("Expected project to exist before deletion")
	}

	// Delete project
	err = Delete(ctx)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Verify it no longer exists
	if Exists(ctx) {
		t.Error("Expected project to not exist after deletion")
	}
}

// TestExists verifies Exists check.
func TestExists(t *testing.T) {
	ctx := setupTestRepo(t)

	// Should not exist initially
	if Exists(ctx) {
		t.Error("Expected project to not exist initially")
	}

	// Create project
	_, err := Create(ctx, "test-project", "Test description")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Should exist now
	if !Exists(ctx) {
		t.Error("Expected project to exist after creation")
	}
}
