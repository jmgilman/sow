package project

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sow"
)

// TestNormalizeName tests the normalizeName function with various inputs.
func TestNormalizeName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Basic transformations
		{"Web Based Agents", "web-based-agents"},
		{"API V2", "api-v2"},
		{"UPPERCASE", "uppercase"},
		{"  spaces  ", "spaces"},

		// Special character handling
		{"With!Invalid@Chars#", "withinvalidchars"},
		{"feature--name", "feature-name"},
		{"-leading-trailing-", "leading-trailing"},

		// Edge cases
		{"", ""},
		{"   ", ""},
		{"!!!@@@###", ""},
		{"123-numbers-456", "123-numbers-456"},
		{"under_scores", "under_scores"},

		// Multiple consecutive hyphens
		{"multiple---hyphens", "multiple-hyphens"},
		{"many----hyphens----here", "many-hyphens-here"},

		// Unicode characters (should be removed)
		{"café", "caf"},
		{"hello-世界", "hello"},
		{"emoji-🎉-test", "emoji-test"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeName(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeName(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestNormalizeName_EdgeCases tests edge cases more thoroughly.
func TestNormalizeName_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "     ",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "mixed valid and invalid",
			input:    "valid-@#$-name",
			expected: "valid-name",
		},
		{
			name:     "hyphens at start and end",
			input:    "---middle---",
			expected: "middle",
		},
		{
			name:     "consecutive hyphens throughout",
			input:    "a--b---c----d",
			expected: "a-b-c-d",
		},
		{
			name:     "very long name",
			input:    "this-is-a-very-long-project-name-that-should-still-work-correctly",
			expected: "this-is-a-very-long-project-name-that-should-still-work-correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeName(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeName(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestGetTypePrefix tests the getTypePrefix function.
func TestGetTypePrefix(t *testing.T) {
	testCases := []struct {
		projectType string
		expected    string
	}{
		// Valid types
		{"standard", "feat/"},
		{"exploration", "explore/"},
		{"design", "design/"},
		{"breakdown", "breakdown/"},

		// Invalid types (should return default)
		{"unknown", "feat/"},
		{"", "feat/"},
		{"invalid-type", "feat/"},
	}

	for _, tc := range testCases {
		t.Run(tc.projectType, func(t *testing.T) {
			result := getTypePrefix(tc.projectType)
			if result != tc.expected {
				t.Errorf("getTypePrefix(%q) = %q; want %q", tc.projectType, result, tc.expected)
			}
		})
	}
}

// TestGetTypeOptions tests the getTypeOptions function.
func TestGetTypeOptions(t *testing.T) {
	options := getTypeOptions()

	// Should have 5 options (4 types + cancel)
	if len(options) != 5 {
		t.Errorf("getTypeOptions() returned %d options; want 5", len(options))
	}

	// Note: We can't easily inspect huh.Option internals in tests,
	// so we just verify the count. The order is verified implicitly
	// by the order they're created in getTypeOptions().
	// This is a limitation of the huh library's API.
}

// TestPreviewBranchName tests the previewBranchName function.
func TestPreviewBranchName(t *testing.T) {
	testCases := []struct {
		projectType string
		name        string
		expected    string
	}{
		// Test each project type
		{"standard", "Add JWT Auth", "feat/add-jwt-auth"},
		{"exploration", "Research Caching", "explore/research-caching"},
		{"design", "API Architecture", "design/api-architecture"},
		{"breakdown", "Split Large Task", "breakdown/split-large-task"},

		// Test name normalization
		{"standard", "With!Special@Chars", "feat/withspecialchars"},
		{"standard", "UPPERCASE", "feat/uppercase"},
		{"standard", "  spaces  ", "feat/spaces"},

		// Test unknown type (should use default)
		{"unknown", "Test Project", "feat/test-project"},
	}

	for _, tc := range testCases {
		t.Run(tc.projectType+"/"+tc.name, func(t *testing.T) {
			result := previewBranchName(tc.projectType, tc.name)
			if result != tc.expected {
				t.Errorf("previewBranchName(%q, %q) = %q; want %q",
					tc.projectType, tc.name, result, tc.expected)
			}
		})
	}
}

// TestWithSpinner_PropagatesError tests that withSpinner propagates errors.
func TestWithSpinner_PropagatesError(t *testing.T) {
	expectedErr := errors.New("test error")

	err := withSpinner("Test operation", func() error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("withSpinner() = %v; want %v", err, expectedErr)
	}
}

// TestWithSpinner_ReturnsNilOnSuccess tests that withSpinner returns nil on success.
func TestWithSpinner_ReturnsNilOnSuccess(t *testing.T) {
	err := withSpinner("Test operation", func() error {
		return nil
	})

	if err != nil {
		t.Errorf("withSpinner() = %v; want nil", err)
	}
}

// TestProjectTypesMap verifies the projectTypes map configuration.
func TestProjectTypesMap(t *testing.T) {
	// Verify all four types exist
	expectedTypes := map[string]struct {
		prefix      string
		description string
	}{
		"standard": {
			prefix:      "feat/",
			description: "Feature work and bug fixes",
		},
		"exploration": {
			prefix:      "explore/",
			description: "Research and investigation",
		},
		"design": {
			prefix:      "design/",
			description: "Architecture and design documents",
		},
		"breakdown": {
			prefix:      "breakdown/",
			description: "Decompose work into tasks",
		},
	}

	if len(projectTypes) != len(expectedTypes) {
		t.Errorf("projectTypes has %d entries; want %d", len(projectTypes), len(expectedTypes))
	}

	for typeName, expected := range expectedTypes {
		config, exists := projectTypes[typeName]
		if !exists {
			t.Errorf("projectTypes missing type %q", typeName)
			continue
		}

		if config.Prefix != expected.prefix {
			t.Errorf("projectTypes[%q].Prefix = %q; want %q",
				typeName, config.Prefix, expected.prefix)
		}

		if config.Description != expected.description {
			t.Errorf("projectTypes[%q].Description = %q; want %q",
				typeName, config.Description, expected.description)
		}
	}
}

// TestIsValidBranchName_ValidNames tests valid branch names.
func TestIsValidBranchName_ValidNames(t *testing.T) {
	testCases := []string{
		"feat/auth",
		"explore/api-v2",
		"123-feature",
		"design/architecture",
		"breakdown/task-list",
		"feat/web-based-agents",
		"feature-name",
		"feat/under_scores_ok",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			err := isValidBranchName(tc)
			if err != nil {
				t.Errorf("isValidBranchName(%q) returned error: %v; want nil", tc, err)
			}
		})
	}
}

// TestIsValidBranchName_InvalidNames tests invalid branch names.
func TestIsValidBranchName_InvalidNames(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
	}{
		{"", "branch name cannot be empty"},
		{"/leading-slash", "branch name cannot start or end with /"},
		{"trailing-slash/", "branch name cannot start or end with /"},
		{"feat..auth", "branch name cannot contain double dots"},
		{"feat//auth", "branch name cannot contain consecutive slashes"},
		{"feat.lock", "branch name cannot end with .lock"},
		{"feat/branch.lock", "branch name cannot end with .lock"},
		{"feat~auth", "branch name contains invalid character: ~"},
		{"feat^auth", "branch name contains invalid character: ^"},
		{"feat:auth", "branch name contains invalid character: :"},
		{"feat?auth", "branch name contains invalid character: ?"},
		{"feat*auth", "branch name contains invalid character: *"},
		{"feat[auth", "branch name contains invalid character: ["},
		{"feat\\auth", "branch name contains invalid character: \\"},
		{"feat with spaces", "branch name contains invalid character:  "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := isValidBranchName(tc.name)
			if err == nil {
				t.Errorf("isValidBranchName(%q) returned nil; want error containing %q", tc.name, tc.expectedError)
				return
			}
			if err.Error() != tc.expectedError {
				t.Errorf("isValidBranchName(%q) error = %q; want %q", tc.name, err.Error(), tc.expectedError)
			}
		})
	}
}

// TestCheckBranchState_NoBranchNoWorktreeNoProject tests when branch doesn't exist.
func TestCheckBranchState_NoBranchNoWorktreeNoProject(t *testing.T) {
	ctx, _ := setupTestContext(t)

	state, err := checkBranchState(ctx, "feat/nonexistent")
	if err != nil {
		t.Fatalf("checkBranchState failed: %v", err)
	}

	if state.BranchExists {
		t.Error("BranchExists should be false when branch doesn't exist")
	}
	if state.WorktreeExists {
		t.Error("WorktreeExists should be false when branch doesn't exist")
	}
	if state.ProjectExists {
		t.Error("ProjectExists should be false when branch doesn't exist")
	}
}

// TestCheckBranchState_BranchExistsNoWorktree tests when branch exists but no worktree.
func TestCheckBranchState_BranchExistsNoWorktree(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Need to create an initial commit for git branch to work
	readmeFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.CommandContext(context.Background(), "git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create a branch using git CLI
	cmd = exec.CommandContext(context.Background(), "git", "branch", "feat/test-branch")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	state, err := checkBranchState(ctx, "feat/test-branch")
	if err != nil {
		t.Fatalf("checkBranchState failed: %v", err)
	}

	if !state.BranchExists {
		t.Error("BranchExists should be true when branch exists")
	}
	if state.WorktreeExists {
		t.Error("WorktreeExists should be false when worktree doesn't exist")
	}
	if state.ProjectExists {
		t.Error("ProjectExists should be false when worktree doesn't exist")
	}
}

// TestCheckBranchState_WorktreeExistsNoProject tests when worktree exists but no project.
func TestCheckBranchState_WorktreeExistsNoProject(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Need to create an initial commit for git branch to work
	readmeFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.CommandContext(context.Background(), "git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create branch and worktree
	branchName := "feat/test-worktree"
	worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)

	// Create branch
	cmd = exec.CommandContext(context.Background(), "git", "branch", branchName)
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	// Create worktree directory structure (but no project)
	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		t.Fatalf("failed to create worktree directory: %v", err)
	}

	state, err := checkBranchState(ctx, branchName)
	if err != nil {
		t.Fatalf("checkBranchState failed: %v", err)
	}

	if !state.BranchExists {
		t.Error("BranchExists should be true when branch exists")
	}
	if !state.WorktreeExists {
		t.Error("WorktreeExists should be true when worktree directory exists")
	}
	if state.ProjectExists {
		t.Error("ProjectExists should be false when project file doesn't exist")
	}
}

// TestCheckBranchState_FullStack tests when branch, worktree, and project all exist.
func TestCheckBranchState_FullStack(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Need to create an initial commit for git branch to work
	readmeFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.CommandContext(context.Background(), "git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create branch and worktree with project
	branchName := "feat/test-full"
	worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)

	// Create branch
	cmd = exec.CommandContext(context.Background(), "git", "branch", branchName)
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	// Create worktree directory structure
	projectDir := filepath.Join(worktreePath, ".sow", "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	// Create state.yaml file (project exists)
	stateFile := filepath.Join(projectDir, "state.yaml")
	if err := os.WriteFile(stateFile, []byte("name: test\n"), 0644); err != nil {
		t.Fatalf("failed to create state.yaml: %v", err)
	}

	state, err := checkBranchState(ctx, branchName)
	if err != nil {
		t.Fatalf("checkBranchState failed: %v", err)
	}

	if !state.BranchExists {
		t.Error("BranchExists should be true when branch exists")
	}
	if !state.WorktreeExists {
		t.Error("WorktreeExists should be true when worktree directory exists")
	}
	if !state.ProjectExists {
		t.Error("ProjectExists should be true when state.yaml exists")
	}
}

// TestFormatProjectProgress tests the formatProjectProgress function.
func TestFormatProjectProgress(t *testing.T) {
	testCases := []struct {
		name     string
		proj     ProjectInfo
		expected string
	}{
		{
			name: "project with tasks shows task counts",
			proj: ProjectInfo{
				Type:           "standard",
				Phase:          "implementation",
				TasksCompleted: 3,
				TasksTotal:     5,
			},
			expected: "Standard: implementation, 3/5 tasks completed",
		},
		{
			name: "project without tasks excludes task portion",
			proj: ProjectInfo{
				Type:       "design",
				Phase:      "active",
				TasksTotal: 0,
			},
			expected: "Design: active",
		},
		{
			name: "exploration type with tasks",
			proj: ProjectInfo{
				Type:           "exploration",
				Phase:          "gathering",
				TasksCompleted: 4,
				TasksTotal:     7,
			},
			expected: "Exploration: gathering, 4/7 tasks completed",
		},
		{
			name: "breakdown type with all tasks completed",
			proj: ProjectInfo{
				Type:           "breakdown",
				Phase:          "completed",
				TasksCompleted: 10,
				TasksTotal:     10,
			},
			expected: "Breakdown: completed, 10/10 tasks completed",
		},
		{
			name: "standard type with zero completed tasks",
			proj: ProjectInfo{
				Type:           "standard",
				Phase:          "planning",
				TasksCompleted: 0,
				TasksTotal:     8,
			},
			expected: "Standard: planning, 0/8 tasks completed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatProjectProgress(tc.proj)
			if result != tc.expected {
				t.Errorf("formatProjectProgress() = %q; want %q", result, tc.expected)
			}
		})
	}
}

// TestListProjects_EmptyWorktreesDirectory tests listProjects with no worktrees.
func TestListProjects_EmptyWorktreesDirectory(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create empty worktrees directory
	worktreesDir := filepath.Join(tmpDir, ".sow", "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		t.Fatalf("failed to create worktrees directory: %v", err)
	}

	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects() returned error: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("listProjects() returned %d projects; want 0", len(projects))
	}
}

// TestListProjects_MissingWorktreesDirectory tests listProjects when worktrees dir doesn't exist.
func TestListProjects_MissingWorktreesDirectory(t *testing.T) {
	ctx, _ := setupTestContext(t)

	// Don't create worktrees directory - it doesn't exist
	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects() returned error: %v; want nil", err)
	}

	if len(projects) != 0 {
		t.Errorf("listProjects() returned %d projects; want 0", len(projects))
	}
}

// TestListProjects_DirectoryWithoutStateFile tests that directories without state files are skipped.
func TestListProjects_DirectoryWithoutStateFile(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create worktrees directory with a subdirectory that has no state file
	worktreesDir := filepath.Join(tmpDir, ".sow", "worktrees")
	emptyDir := filepath.Join(worktreesDir, "feat", "empty-worktree")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("failed to create empty worktree directory: %v", err)
	}

	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects() returned error: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("listProjects() returned %d projects; want 0 (directory without state should be skipped)", len(projects))
	}
}

// TestListProjects_SingleValidProject tests discovery of a single valid project.
func TestListProjects_SingleValidProject(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create a valid project worktree
	branchName := "feat/test-project"
	worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)
	projectDir := filepath.Join(worktreePath, ".sow", "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	// Initialize git repo in the worktree directory
	setupTestRepo(t, worktreePath)

	// Create a minimal valid state.yaml
	stateContent := `name: test-project
type: standard
branch: feat/test-project
state: implementation
created_at: 2025-01-01T10:00:00Z
updated_at: 2025-01-01T10:00:00Z
phases: {}
statechart:
  current_state: ImplementationExecuting
  updated_at: 2025-01-01T10:00:00Z
`
	stateFile := filepath.Join(projectDir, "state.yaml")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("failed to create state.yaml: %v", err)
	}

	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects() returned error: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("listProjects() returned %d projects; want 1", len(projects))
	}

	proj := projects[0]
	if proj.Branch != branchName {
		t.Errorf("Branch = %q; want %q", proj.Branch, branchName)
	}
	if proj.Name != "test-project" {
		t.Errorf("Name = %q; want %q", proj.Name, "test-project")
	}
	if proj.Type != "standard" {
		t.Errorf("Type = %q; want %q", proj.Type, "standard")
	}
	if proj.Phase != "ImplementationExecuting" {
		t.Errorf("Phase = %q; want %q", proj.Phase, "ImplementationExecuting")
	}
	if proj.TasksTotal != 0 || proj.TasksCompleted != 0 {
		t.Errorf("Tasks = %d/%d; want 0/0 (no tasks in project)", proj.TasksCompleted, proj.TasksTotal)
	}
}

// TestListProjects_MultipleProjectsSorted tests that multiple projects are returned sorted by modification time.
func TestListProjects_MultipleProjectsSorted(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	worktreesDir := filepath.Join(tmpDir, ".sow", "worktrees")

	// Helper to create a project with specific modification time
	createProject := func(branchName string, modTime time.Time) {
		worktreePath := filepath.Join(worktreesDir, branchName)
		projectDir := filepath.Join(worktreePath, ".sow", "project")
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			t.Fatalf("failed to create project directory: %v", err)
		}

		// Initialize git repo in the worktree directory
		setupTestRepo(t, worktreePath)

		// Extract a valid project name from branch (feat/xxx -> xxx, no forward slashes)
		projectName := strings.ReplaceAll(branchName, "/", "-")

		stateContent := `name: ` + projectName + `
type: standard
branch: ` + branchName + `
state: implementation
created_at: 2025-01-01T10:00:00Z
updated_at: 2025-01-01T10:00:00Z
phases: {}
statechart:
  current_state: ImplementationExecuting
  updated_at: 2025-01-01T10:00:00Z
`
		stateFile := filepath.Join(projectDir, "state.yaml")
		if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
			t.Fatalf("failed to create state.yaml: %v", err)
		}

		// Set modification time
		if err := os.Chtimes(stateFile, modTime, modTime); err != nil {
			t.Fatalf("failed to set modification time: %v", err)
		}
	}

	// Create three projects with different modification times
	now := time.Now()
	createProject("feat/oldest", now.Add(-2*time.Hour))
	createProject("feat/middle", now.Add(-1*time.Hour))
	createProject("feat/newest", now)

	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects() returned error: %v", err)
	}

	if len(projects) != 3 {
		t.Fatalf("listProjects() returned %d projects; want 3", len(projects))
	}

	// Should be sorted newest first
	if projects[0].Branch != "feat/newest" {
		t.Errorf("projects[0].Branch = %q; want %q", projects[0].Branch, "feat/newest")
	}
	if projects[1].Branch != "feat/middle" {
		t.Errorf("projects[1].Branch = %q; want %q", projects[1].Branch, "feat/middle")
	}
	if projects[2].Branch != "feat/oldest" {
		t.Errorf("projects[2].Branch = %q; want %q", projects[2].Branch, "feat/oldest")
	}
}

// TestListProjects_ProjectWithTasks tests that task counts are calculated correctly.
func TestListProjects_ProjectWithTasks(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	branchName := "feat/with-tasks"
	worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)
	projectDir := filepath.Join(worktreePath, ".sow", "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	// Initialize git repo in the worktree directory
	setupTestRepo(t, worktreePath)

	// Create state with tasks across multiple phases
	stateContent := `name: project-with-tasks
type: standard
branch: feat/with-tasks
state: implementation
created_at: 2025-01-01T10:00:00Z
updated_at: 2025-01-01T10:00:00Z
statechart:
  current_state: ImplementationExecuting
  updated_at: 2025-01-01T10:00:00Z
phases:
  planning:
    status: completed
    enabled: true
    created_at: 2025-01-01T10:00:00Z
    inputs: []
    outputs: []
    tasks:
      - id: "010"
        name: "Planning Task 1"
        phase: planning
        status: completed
        created_at: 2025-01-01T10:00:00Z
        updated_at: 2025-01-01T11:00:00Z
        iteration: 1
        assigned_agent: implementer
        inputs: []
        outputs: []
      - id: "020"
        name: "Planning Task 2"
        phase: planning
        status: completed
        created_at: 2025-01-01T10:00:00Z
        updated_at: 2025-01-01T11:00:00Z
        iteration: 1
        assigned_agent: implementer
        inputs: []
        outputs: []
  implementation:
    status: in_progress
    enabled: true
    created_at: 2025-01-01T12:00:00Z
    inputs: []
    outputs: []
    tasks:
      - id: "030"
        name: "Implementation Task 1"
        phase: implementation
        status: completed
        created_at: 2025-01-01T12:00:00Z
        updated_at: 2025-01-01T13:00:00Z
        iteration: 1
        assigned_agent: implementer
        inputs: []
        outputs: []
      - id: "040"
        name: "Implementation Task 2"
        phase: implementation
        status: in_progress
        created_at: 2025-01-01T12:00:00Z
        updated_at: 2025-01-01T13:00:00Z
        iteration: 1
        assigned_agent: implementer
        inputs: []
        outputs: []
      - id: "050"
        name: "Implementation Task 3"
        phase: implementation
        status: pending
        created_at: 2025-01-01T12:00:00Z
        updated_at: 2025-01-01T13:00:00Z
        iteration: 1
        assigned_agent: implementer
        inputs: []
        outputs: []
`
	stateFile := filepath.Join(projectDir, "state.yaml")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("failed to create state.yaml: %v", err)
	}

	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects() returned error: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("listProjects() returned %d projects; want 1", len(projects))
	}

	proj := projects[0]
	// Should have 3 completed tasks (2 from planning, 1 from implementation) out of 5 total
	if proj.TasksCompleted != 3 {
		t.Errorf("TasksCompleted = %d; want 3", proj.TasksCompleted)
	}
	if proj.TasksTotal != 5 {
		t.Errorf("TasksTotal = %d; want 5", proj.TasksTotal)
	}
}

// TestIsProtectedBranch tests the isProtectedBranch helper function.
func TestIsProtectedBranch(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected bool
	}{
		{"main is protected", "main", true},
		{"master is protected", "master", true},
		{"feat branch is not protected", "feat/something", false},
		{"explore branch is not protected", "explore/test", false},
		{"empty string is not protected", "", false},
		{"main-like but different is not protected", "main-branch", false},
		{"master-like but different is not protected", "master-feature", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isProtectedBranch(tt.branch)
			if result != tt.expected {
				t.Errorf("isProtectedBranch(%q) = %v; want %v", tt.branch, result, tt.expected)
			}
		})
	}
}

// TestValidateProjectName tests the validateProjectName function.
func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		prefix    string
		wantErr   bool
		errSubstr string
	}{
		// Valid cases
		{
			name:    "valid simple name",
			input:   "my project",
			prefix:  "feat/",
			wantErr: false,
		},
		{
			name:    "valid name with numbers",
			input:   "Project 123",
			prefix:  "explore/",
			wantErr: false,
		},
		{
			name:    "valid name with special chars that get normalized",
			input:   "API@v2",
			prefix:  "feat/",
			wantErr: false,
		},

		// Empty/whitespace cases
		{
			name:      "empty string",
			input:     "",
			prefix:    "feat/",
			wantErr:   true,
			errSubstr: "cannot be empty",
		},
		{
			name:      "only whitespace",
			input:     "   ",
			prefix:    "feat/",
			wantErr:   true,
			errSubstr: "cannot be empty",
		},

		// Protected branch cases (after normalization)
		{
			name:      "normalizes to main",
			input:     "main",
			prefix:    "",
			wantErr:   true,
			errSubstr: "protected",
		},
		{
			name:      "normalizes to master",
			input:     "master",
			prefix:    "",
			wantErr:   true,
			errSubstr: "protected",
		},
		{
			name:      "normalizes to main with uppercase",
			input:     "MAIN",
			prefix:    "",
			wantErr:   true,
			errSubstr: "protected",
		},

		// Invalid git patterns (should be caught by isValidBranchName)
		{
			name:      "produces consecutive dots after normalization",
			input:     "name",
			prefix:    "feat/..",
			wantErr:   true,
			errSubstr: "double dots",
		},
		{
			name:      "produces leading slash",
			input:     "test",
			prefix:    "/feat/",
			wantErr:   true,
			errSubstr: "start or end with /",
		},

		// Various prefixes
		{
			name:    "with explore prefix",
			input:   "Research Topic",
			prefix:  "explore/",
			wantErr: false,
		},
		{
			name:    "with design prefix",
			input:   "Architecture",
			prefix:  "design/",
			wantErr: false,
		},
		{
			name:    "with breakdown prefix",
			input:   "Task List",
			prefix:  "breakdown/",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectName(tt.input, tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectName(%q, %q) error = %v; wantErr %v",
					tt.input, tt.prefix, err, tt.wantErr)
				return
			}
			if err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("validateProjectName(%q, %q) error = %q; want substring %q",
						tt.input, tt.prefix, err.Error(), tt.errSubstr)
				}
			}
		})
	}
}

// TestIsValidBranchName_ProtectedBranches tests that protected branches are rejected.
func TestIsValidBranchName_ProtectedBranches(t *testing.T) {
	tests := []struct {
		name   string
		branch string
	}{
		{"main is protected", "main"},
		{"master is protected", "master"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isValidBranchName(tt.branch)
			if err == nil {
				t.Errorf("isValidBranchName(%q) returned nil; want error for protected branch", tt.branch)
				return
			}
			if !strings.Contains(err.Error(), "protected") {
				t.Errorf("isValidBranchName(%q) error = %q; want error containing 'protected'",
					tt.branch, err.Error())
			}
		})
	}
}

// TestShouldCheckUncommittedChanges tests the conditional logic for uncommitted changes checking.
func TestShouldCheckUncommittedChanges(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create an initial commit so we can create branches
	readmeFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.CommandContext(context.Background(), "git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Get the current branch (should be master or main)
	currentBranch, err := ctx.Git().CurrentBranch()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}

	tests := []struct {
		name         string
		targetBranch string
		wantCheck    bool
		wantErr      bool
	}{
		{
			name:         "current equals target - should check",
			targetBranch: currentBranch,
			wantCheck:    true,
			wantErr:      false,
		},
		{
			name:         "current does not equal target - should not check",
			targetBranch: "feat/different-branch",
			wantCheck:    false,
			wantErr:      false,
		},
		{
			name:         "another different branch - should not check",
			targetBranch: "explore/something",
			wantCheck:    false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldCheck, err := shouldCheckUncommittedChanges(ctx, tt.targetBranch)
			if (err != nil) != tt.wantErr {
				t.Errorf("shouldCheckUncommittedChanges() error = %v; wantErr %v", err, tt.wantErr)
				return
			}
			if shouldCheck != tt.wantCheck {
				t.Errorf("shouldCheckUncommittedChanges(%q) = %v; want %v",
					tt.targetBranch, shouldCheck, tt.wantCheck)
			}
		})
	}
}

// TestPerformUncommittedChangesCheckIfNeeded tests the conditional uncommitted changes validation.
func TestPerformUncommittedChangesCheckIfNeeded(t *testing.T) {
	// Set environment variable to skip the actual uncommitted changes check
	// so we can test the conditional logic without needing real git state
	t.Setenv("SOW_SKIP_UNCOMMITTED_CHECK", "1")

	ctx, tmpDir := setupTestContext(t)

	// Create an initial commit
	readmeFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.CommandContext(context.Background(), "git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Get the current branch
	currentBranch, err := ctx.Git().CurrentBranch()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}

	tests := []struct {
		name         string
		targetBranch string
		wantErr      bool
	}{
		{
			name:         "different branch - no check needed",
			targetBranch: "feat/different",
			wantErr:      false,
		},
		{
			name:         "same branch - check runs but passes due to SOW_SKIP_UNCOMMITTED_CHECK",
			targetBranch: currentBranch,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := performUncommittedChangesCheckIfNeeded(ctx, tt.targetBranch)
			if (err != nil) != tt.wantErr {
				t.Errorf("performUncommittedChangesCheckIfNeeded(%q) error = %v; wantErr %v",
					tt.targetBranch, err, tt.wantErr)
			}
		})
	}
}

// TestPerformUncommittedChangesCheckIfNeeded_ErrorMessage tests the enhanced error message.
func TestPerformUncommittedChangesCheckIfNeeded_ErrorMessage(t *testing.T) {
	// Don't set SOW_SKIP_UNCOMMITTED_CHECK so the check runs
	ctx, tmpDir := setupTestContext(t)

	// Create an initial commit
	readmeFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.CommandContext(context.Background(), "git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Get the current branch
	currentBranch, err := ctx.Git().CurrentBranch()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}

	// Create uncommitted changes
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("uncommitted content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Stage the file to create uncommitted changes
	cmd = exec.CommandContext(context.Background(), "git", "add", "test.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to stage test file: %v", err)
	}

	// Now try to check on same branch - should fail with enhanced message
	err = performUncommittedChangesCheckIfNeeded(ctx, currentBranch)
	if err == nil {
		t.Fatal("performUncommittedChangesCheckIfNeeded() returned nil; want error")
	}

	// Verify the error message contains expected parts
	errMsg := err.Error()
	expectedParts := []string{
		"uncommitted changes",
		currentBranch,
		"git add",
		"git commit",
		"git stash",
	}

	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("error message missing expected part %q\nGot: %s", part, errMsg)
		}
	}
}

// TestCanCreateProject tests the canCreateProject validation function.
func TestCanCreateProject(t *testing.T) {
	tests := []struct {
		name       string
		state      *BranchState
		branchName string
		wantErr    bool
		errSubstr  string
	}{
		{
			name: "no branch, no worktree, no project - OK",
			state: &BranchState{
				BranchExists:   false,
				WorktreeExists: false,
				ProjectExists:  false,
			},
			branchName: "feat/new-project",
			wantErr:    false,
		},
		{
			name: "branch exists, no worktree, no project - OK",
			state: &BranchState{
				BranchExists:   true,
				WorktreeExists: false,
				ProjectExists:  false,
			},
			branchName: "feat/existing-branch",
			wantErr:    false,
		},
		{
			name: "project already exists - error",
			state: &BranchState{
				BranchExists:   true,
				WorktreeExists: true,
				ProjectExists:  true,
			},
			branchName: "feat/has-project",
			wantErr:    true,
			errSubstr:  "already has a project",
		},
		{
			name: "worktree exists but no project - error (inconsistent state)",
			state: &BranchState{
				BranchExists:   true,
				WorktreeExists: true,
				ProjectExists:  false,
			},
			branchName: "feat/orphan-worktree",
			wantErr:    true,
			errSubstr:  "worktree exists but project missing",
		},
		{
			name: "no branch but worktree exists - error (inconsistent state)",
			state: &BranchState{
				BranchExists:   false,
				WorktreeExists: true,
				ProjectExists:  false,
			},
			branchName: "feat/stale-worktree",
			wantErr:    true,
			errSubstr:  "worktree exists but project missing",
		},
		{
			name: "no branch but project exists somehow - error",
			state: &BranchState{
				BranchExists:   false,
				WorktreeExists: true,
				ProjectExists:  true,
			},
			branchName: "feat/weird-state",
			wantErr:    true,
			errSubstr:  "already has a project",
		},
		{
			name: "all false - OK (fresh start)",
			state: &BranchState{
				BranchExists:   false,
				WorktreeExists: false,
				ProjectExists:  false,
			},
			branchName: "feat/totally-new",
			wantErr:    false,
		},
		{
			name: "branch name in error message",
			state: &BranchState{
				BranchExists:   true,
				WorktreeExists: true,
				ProjectExists:  true,
			},
			branchName: "feat/test-name-123",
			wantErr:    true,
			errSubstr:  "feat/test-name-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := canCreateProject(tt.state, tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("canCreateProject() error = %v; wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("canCreateProject() error = %q; want substring %q",
						err.Error(), tt.errSubstr)
				}
			}
		})
	}
}

// TestValidateProjectExists tests the validateProjectExists function.
func TestValidateProjectExists(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, ctx *sow.Context, tmpDir string, branchName string)
		branchName string
		wantErr    bool
		errSubstr  string
	}{
		{
			name: "all exist - OK",
			setup: func(t *testing.T, _ *sow.Context, tmpDir string, branchName string) {
				// Create initial commit
				createInitialCommit(t, tmpDir)

				// Create branch
				cmd := exec.CommandContext(context.Background(), "git", "branch", branchName)
				cmd.Dir = tmpDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("failed to create branch: %v", err)
				}

				// Create worktree and project
				worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)
				projectDir := filepath.Join(worktreePath, ".sow", "project")
				if err := os.MkdirAll(projectDir, 0755); err != nil {
					t.Fatalf("failed to create project dir: %v", err)
				}

				stateFile := filepath.Join(projectDir, "state.yaml")
				if err := os.WriteFile(stateFile, []byte("name: test\n"), 0644); err != nil {
					t.Fatalf("failed to create state.yaml: %v", err)
				}
			},
			branchName: "feat/complete",
			wantErr:    false,
		},
		{
			name: "branch missing - error",
			setup: func(t *testing.T, _ *sow.Context, tmpDir string, _ string) {
				createInitialCommit(t, tmpDir)
				// Don't create branch
			},
			branchName: "feat/no-branch",
			wantErr:    true,
			errSubstr:  "does not exist",
		},
		{
			name: "worktree missing - error",
			setup: func(t *testing.T, _ *sow.Context, tmpDir string, branchName string) {
				createInitialCommit(t, tmpDir)

				// Create branch but no worktree
				cmd := exec.CommandContext(context.Background(), "git", "branch", branchName)
				cmd.Dir = tmpDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("failed to create branch: %v", err)
				}
			},
			branchName: "feat/no-worktree",
			wantErr:    true,
			errSubstr:  "worktree for branch",
		},
		{
			name: "project missing - error",
			setup: func(t *testing.T, _ *sow.Context, tmpDir string, branchName string) {
				createInitialCommit(t, tmpDir)

				// Create branch
				cmd := exec.CommandContext(context.Background(), "git", "branch", branchName)
				cmd.Dir = tmpDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("failed to create branch: %v", err)
				}

				// Create worktree but no project
				worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)
				if err := os.MkdirAll(worktreePath, 0755); err != nil {
					t.Fatalf("failed to create worktree: %v", err)
				}
			},
			branchName: "feat/no-project",
			wantErr:    true,
			errSubstr:  "project for branch",
		},
		{
			name: "error includes branch name",
			setup: func(t *testing.T, _ *sow.Context, tmpDir string, _ string) {
				createInitialCommit(t, tmpDir)
				// Don't create anything
			},
			branchName: "feat/my-special-branch",
			wantErr:    true,
			errSubstr:  "feat/my-special-branch",
		},
		{
			name: "nothing exists - error mentions branch first",
			setup: func(t *testing.T, _ *sow.Context, tmpDir string, _ string) {
				createInitialCommit(t, tmpDir)
			},
			branchName: "feat/nonexistent",
			wantErr:    true,
			errSubstr:  "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, tmpDir := setupTestContext(t)
			tt.setup(t, ctx, tmpDir, tt.branchName)

			err := validateProjectExists(ctx, tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectExists() error = %v; wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("validateProjectExists() error = %q; want substring %q",
						err.Error(), tt.errSubstr)
				}
			}
		})
	}
}

// TestListExistingProjects tests the listExistingProjects function.
func TestListExistingProjects(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, tmpDir string)
		wantBranches  []string
		wantErr       bool
	}{
		{
			name: "no projects",
			setup: func(t *testing.T, tmpDir string) {
				createInitialCommit(t, tmpDir)
			},
			wantBranches: []string{},
			wantErr:      false,
		},
		{
			name: "single project",
			setup: func(t *testing.T, tmpDir string) {
				createInitialCommit(t, tmpDir)
				createTestProject(t, tmpDir, "feat/project-one")
			},
			wantBranches: []string{"feat/project-one"},
			wantErr:      false,
		},
		{
			name: "multiple projects sorted alphabetically",
			setup: func(t *testing.T, tmpDir string) {
				createInitialCommit(t, tmpDir)
				createTestProject(t, tmpDir, "feat/zzz-last")
				createTestProject(t, tmpDir, "feat/aaa-first")
				createTestProject(t, tmpDir, "explore/middle")
			},
			wantBranches: []string{"explore/middle", "feat/aaa-first", "feat/zzz-last"},
			wantErr:      false,
		},
		{
			name: "branch without project excluded",
			setup: func(t *testing.T, tmpDir string) {
				createInitialCommit(t, tmpDir)
				createTestProject(t, tmpDir, "feat/has-project")

				// Create branch without project
				cmd := exec.CommandContext(context.Background(), "git", "branch", "feat/no-project")
				cmd.Dir = tmpDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("failed to create branch: %v", err)
				}
			},
			wantBranches: []string{"feat/has-project"},
			wantErr:      false,
		},
		{
			name: "worktree without project excluded",
			setup: func(t *testing.T, tmpDir string) {
				createInitialCommit(t, tmpDir)
				createTestProject(t, tmpDir, "feat/complete")

				// Create branch with worktree but no project
				branchName := "feat/orphan-worktree"
				cmd := exec.CommandContext(context.Background(), "git", "branch", branchName)
				cmd.Dir = tmpDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("failed to create branch: %v", err)
				}

				worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)
				if err := os.MkdirAll(worktreePath, 0755); err != nil {
					t.Fatalf("failed to create worktree: %v", err)
				}
			},
			wantBranches: []string{"feat/complete"},
			wantErr:      false,
		},
		{
			name: "mixed project types sorted together",
			setup: func(t *testing.T, tmpDir string) {
				createInitialCommit(t, tmpDir)
				createTestProject(t, tmpDir, "feat/standard")
				createTestProject(t, tmpDir, "explore/research")
				createTestProject(t, tmpDir, "design/architecture")
				createTestProject(t, tmpDir, "breakdown/tasks")
			},
			wantBranches: []string{"breakdown/tasks", "design/architecture", "explore/research", "feat/standard"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, tmpDir := setupTestContext(t)
			tt.setup(t, tmpDir)

			branches, err := listExistingProjects(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("listExistingProjects() error = %v; wantErr %v", err, tt.wantErr)
				return
			}

			if len(branches) != len(tt.wantBranches) {
				t.Errorf("listExistingProjects() returned %d branches; want %d\nGot: %v\nWant: %v",
					len(branches), len(tt.wantBranches), branches, tt.wantBranches)
				return
			}

			for i, branch := range branches {
				if branch != tt.wantBranches[i] {
					t.Errorf("listExistingProjects()[%d] = %q; want %q",
						i, branch, tt.wantBranches[i])
				}
			}
		})
	}
}

// Helper function to create an initial commit in a test repo.
func createInitialCommit(t *testing.T, tmpDir string) {
	t.Helper()

	readmeFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.CommandContext(context.Background(), "git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

// Helper function to create a complete test project (branch + worktree + project).
func createTestProject(t *testing.T, tmpDir string, branchName string) {
	t.Helper()

	// Create branch
	cmd := exec.CommandContext(context.Background(), "git", "branch", branchName)
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create branch %s: %v", branchName, err)
	}

	// Create worktree and project
	worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)
	projectDir := filepath.Join(worktreePath, ".sow", "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir for %s: %v", branchName, err)
	}

	stateFile := filepath.Join(projectDir, "state.yaml")
	if err := os.WriteFile(stateFile, []byte("name: test\n"), 0644); err != nil {
		t.Fatalf("failed to create state.yaml for %s: %v", branchName, err)
	}
}

// TestFormatError tests the formatError function with various inputs.
func TestFormatError(t *testing.T) {
	testCases := []struct {
		name       string
		problem    string
		howToFix   string
		nextSteps  string
		expected   string
	}{
		{
			name:      "all three parts provided",
			problem:   "Something went wrong",
			howToFix:  "Here's how to fix it",
			nextSteps: "Next steps to take",
			expected:  "Something went wrong\n\nHere's how to fix it\n\nNext steps to take",
		},
		{
			name:      "empty problem",
			problem:   "",
			howToFix:  "Fix instructions",
			nextSteps: "Next steps",
			expected:  "Fix instructions\n\nNext steps",
		},
		{
			name:      "empty how to fix",
			problem:   "Problem description",
			howToFix:  "",
			nextSteps: "Next steps",
			expected:  "Problem description\n\nNext steps",
		},
		{
			name:      "empty next steps",
			problem:   "Problem description",
			howToFix:  "Fix instructions",
			nextSteps: "",
			expected:  "Problem description\n\nFix instructions",
		},
		{
			name:      "all empty",
			problem:   "",
			howToFix:  "",
			nextSteps: "",
			expected:  "",
		},
		{
			name:      "only problem",
			problem:   "Just a problem",
			howToFix:  "",
			nextSteps: "",
			expected:  "Just a problem",
		},
		{
			name:      "multiline content",
			problem:   "Line 1\nLine 2",
			howToFix:  "Fix line 1\nFix line 2",
			nextSteps: "Step 1\nStep 2",
			expected:  "Line 1\nLine 2\n\nFix line 1\nFix line 2\n\nStep 1\nStep 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatError(tc.problem, tc.howToFix, tc.nextSteps)
			if result != tc.expected {
				t.Errorf("formatError() =\n%q\n\nwant:\n%q", result, tc.expected)
			}
		})
	}
}

// TestErrorProtectedBranch tests the errorProtectedBranch function.
func TestErrorProtectedBranch(t *testing.T) {
	testCases := []struct {
		name       string
		branchName string
		expected   string
	}{
		{
			name:       "main branch",
			branchName: "main",
			expected: `Cannot create project on protected branch 'main'

Projects must be created on feature branches.

Action: Choose a different project name`,
		},
		{
			name:       "master branch",
			branchName: "master",
			expected: `Cannot create project on protected branch 'master'

Projects must be created on feature branches.

Action: Choose a different project name`,
		},
		{
			name:       "custom protected branch name",
			branchName: "production",
			expected: `Cannot create project on protected branch 'production'

Projects must be created on feature branches.

Action: Choose a different project name`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := errorProtectedBranch(tc.branchName)
			if result != tc.expected {
				t.Errorf("errorProtectedBranch(%q) =\n%s\n\nwant:\n%s", tc.branchName, result, tc.expected)
			}
		})
	}
}

// TestErrorIssueAlreadyLinked tests the errorIssueAlreadyLinked function.
func TestErrorIssueAlreadyLinked(t *testing.T) {
	testCases := []struct {
		name         string
		issueNumber  int
		linkedBranch string
		expected     string
	}{
		{
			name:         "issue 123",
			issueNumber:  123,
			linkedBranch: "feat/add-jwt-auth",
			expected: `Issue #123 already has a linked branch: feat/add-jwt-auth

To continue working on this issue:
  Select "Continue existing project" from the main menu`,
		},
		{
			name:         "issue 1",
			issueNumber:  1,
			linkedBranch: "explore/research",
			expected: `Issue #1 already has a linked branch: explore/research

To continue working on this issue:
  Select "Continue existing project" from the main menu`,
		},
		{
			name:         "large issue number",
			issueNumber:  9999,
			linkedBranch: "feat/feature-name",
			expected: `Issue #9999 already has a linked branch: feat/feature-name

To continue working on this issue:
  Select "Continue existing project" from the main menu`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := errorIssueAlreadyLinked(tc.issueNumber, tc.linkedBranch)
			if result != tc.expected {
				t.Errorf("errorIssueAlreadyLinked(%d, %q) =\n%s\n\nwant:\n%s",
					tc.issueNumber, tc.linkedBranch, result, tc.expected)
			}
		})
	}
}

// TestErrorBranchHasProject tests the errorBranchHasProject function.
func TestErrorBranchHasProject(t *testing.T) {
	testCases := []struct {
		name        string
		branchName  string
		projectName string
		expected    string
	}{
		{
			name:        "explore branch",
			branchName:  "explore/web-agents",
			projectName: "web agents",
			expected: `Branch 'explore/web-agents' already has a project

To continue this project:
  Select "Continue existing project" from the main menu

To create a different project:
  Choose a different project name (currently: "web agents")`,
		},
		{
			name:        "feat branch",
			branchName:  "feat/auth",
			projectName: "auth",
			expected: `Branch 'feat/auth' already has a project

To continue this project:
  Select "Continue existing project" from the main menu

To create a different project:
  Choose a different project name (currently: "auth")`,
		},
		{
			name:        "long project name",
			branchName:  "feat/very-long-branch-name",
			projectName: "very long project name",
			expected: `Branch 'feat/very-long-branch-name' already has a project

To continue this project:
  Select "Continue existing project" from the main menu

To create a different project:
  Choose a different project name (currently: "very long project name")`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := errorBranchHasProject(tc.branchName, tc.projectName)
			if result != tc.expected {
				t.Errorf("errorBranchHasProject(%q, %q) =\n%s\n\nwant:\n%s",
					tc.branchName, tc.projectName, result, tc.expected)
			}
		})
	}
}

// TestErrorUncommittedChanges tests the errorUncommittedChanges function.
func TestErrorUncommittedChanges(t *testing.T) {
	testCases := []struct {
		name          string
		currentBranch string
		expected      string
	}{
		{
			name:          "feat branch",
			currentBranch: "feat/add-jwt-auth-123",
			expected: `Repository has uncommitted changes

You are currently on branch 'feat/add-jwt-auth-123'.
Creating a worktree requires switching to a different branch first.

To fix:
  Commit: git add . && git commit -m "message"
  Or stash: git stash`,
		},
		{
			name:          "main branch",
			currentBranch: "main",
			expected: `Repository has uncommitted changes

You are currently on branch 'main'.
Creating a worktree requires switching to a different branch first.

To fix:
  Commit: git add . && git commit -m "message"
  Or stash: git stash`,
		},
		{
			name:          "explore branch",
			currentBranch: "explore/research",
			expected: `Repository has uncommitted changes

You are currently on branch 'explore/research'.
Creating a worktree requires switching to a different branch first.

To fix:
  Commit: git add . && git commit -m "message"
  Or stash: git stash`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := errorUncommittedChanges(tc.currentBranch)
			if result != tc.expected {
				t.Errorf("errorUncommittedChanges(%q) =\n%s\n\nwant:\n%s",
					tc.currentBranch, result, tc.expected)
			}
		})
	}
}

// TestErrorInconsistentState tests the errorInconsistentState function.
func TestErrorInconsistentState(t *testing.T) {
	testCases := []struct {
		name         string
		branchName   string
		worktreePath string
		expected     string
	}{
		{
			name:         "feat branch",
			branchName:   "feat/xyz",
			worktreePath: ".sow/worktrees/feat/xyz",
			expected: `Worktree exists but project missing

Branch 'feat/xyz' has a worktree at .sow/worktrees/feat/xyz
but no .sow/project/ directory.

To fix:
  1. Remove worktree: git worktree remove feat/xyz
  2. Delete directory: rm -rf .sow/worktrees/feat/xyz
  3. Try creating project again`,
		},
		{
			name:         "explore branch",
			branchName:   "explore/research",
			worktreePath: ".sow/worktrees/explore/research",
			expected: `Worktree exists but project missing

Branch 'explore/research' has a worktree at .sow/worktrees/explore/research
but no .sow/project/ directory.

To fix:
  1. Remove worktree: git worktree remove explore/research
  2. Delete directory: rm -rf .sow/worktrees/explore/research
  3. Try creating project again`,
		},
		{
			name:         "absolute path",
			branchName:   "feat/test",
			worktreePath: "/home/user/repo/.sow/worktrees/feat/test",
			expected: `Worktree exists but project missing

Branch 'feat/test' has a worktree at /home/user/repo/.sow/worktrees/feat/test
but no .sow/project/ directory.

To fix:
  1. Remove worktree: git worktree remove feat/test
  2. Delete directory: rm -rf /home/user/repo/.sow/worktrees/feat/test
  3. Try creating project again`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := errorInconsistentState(tc.branchName, tc.worktreePath)
			if result != tc.expected {
				t.Errorf("errorInconsistentState(%q, %q) =\n%s\n\nwant:\n%s",
					tc.branchName, tc.worktreePath, result, tc.expected)
			}
		})
	}
}

// TestErrorGitHubCLIMissing tests the errorGitHubCLIMissing function.
func TestErrorGitHubCLIMissing(t *testing.T) {
	expected := `GitHub CLI not found

The 'gh' command is required for GitHub issue integration.

To install:
  macOS: brew install gh
  Linux: See https://cli.github.com/

Or select "From branch name" instead.`

	result := errorGitHubCLIMissing()
	if result != expected {
		t.Errorf("errorGitHubCLIMissing() =\n%s\n\nwant:\n%s", result, expected)
	}
}

// TestShowError tests that showError compiles and accepts correct input.
func TestShowError(t *testing.T) {
	// Set SOW_TEST=1 to skip interactive prompts
	t.Setenv("SOW_TEST", "1")

	testCases := []struct {
		name    string
		message string
	}{
		{
			name:    "simple message",
			message: "Test error message",
		},
		{
			name:    "multiline message",
			message: "Line 1\nLine 2\nLine 3",
		},
		{
			name:    "empty message",
			message: "",
		},
		{
			name:    "formatted error",
			message: formatError("Problem", "How to fix", "Next steps"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic and should return nil in test mode
			err := showError(tc.message)
			if err != nil {
				t.Errorf("showError(%q) returned error: %v", tc.message, err)
			}
		})
	}
}

// TestShowErrorWithOptions tests that showErrorWithOptions compiles and accepts correct input.
func TestShowErrorWithOptions(t *testing.T) {
	// Set SOW_TEST=1 to skip interactive prompts
	t.Setenv("SOW_TEST", "1")

	testCases := []struct {
		name    string
		message string
		options map[string]string
	}{
		{
			name:    "single option",
			message: "Error occurred",
			options: map[string]string{
				"retry": "Try again",
			},
		},
		{
			name:    "multiple options",
			message: "Choose an action",
			options: map[string]string{
				"retry":    "Try again",
				"continue": "Continue existing project",
				"cancel":   "Cancel",
			},
		},
		{
			name:    "empty options",
			message: "Error",
			options: map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			// In test mode, function should handle gracefully
			// We can't fully test the interactive behavior, but we can ensure it compiles
			// and doesn't panic with valid inputs
			_, _ = showErrorWithOptions(tc.message, tc.options)
		})
	}
}

// TestWrapValidationError tests the wrapValidationError function.
func TestWrapValidationError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		context  string
		wantNil  bool
		contains []string
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			context: "any context",
			wantNil: true,
		},
		{
			name:     "wraps error with context",
			err:      errors.New("branch name contains spaces"),
			context:  "project name validation",
			wantNil:  false,
			contains: []string{"branch name contains spaces", "project name validation"},
		},
		{
			name:     "preserves original error message",
			err:      errors.New("invalid character: ~"),
			context:  "branch validation",
			wantNil:  false,
			contains: []string{"invalid character: ~"},
		},
		{
			name:     "empty context still wraps",
			err:      errors.New("test error"),
			context:  "",
			wantNil:  false,
			contains: []string{"test error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := wrapValidationError(tc.err, tc.context)

			if tc.wantNil {
				if result != nil {
					t.Errorf("wrapValidationError() = %v; want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("wrapValidationError() = nil; want error")
			}

			errMsg := result.Error()
			for _, substr := range tc.contains {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("wrapValidationError() error = %q; want to contain %q",
						errMsg, substr)
				}
			}
		})
	}
}

// Mock GitHubClient for testing GitHub error handling functions.
type mockGitHubClient struct {
	checkInstalledErr     error
	checkAuthenticatedErr error
	linkedBranches        []sow.LinkedBranch
	linkedBranchesErr     error
}

func (m *mockGitHubClient) Ensure() error {
	return nil
}

func (m *mockGitHubClient) CheckInstalled() error {
	return m.checkInstalledErr
}

func (m *mockGitHubClient) CheckAuthenticated() error {
	return m.checkAuthenticatedErr
}

func (m *mockGitHubClient) ListIssues(_, _ string) ([]sow.Issue, error) {
	return nil, nil
}

func (m *mockGitHubClient) GetLinkedBranches(_ int) ([]sow.LinkedBranch, error) {
	return m.linkedBranches, m.linkedBranchesErr
}

func (m *mockGitHubClient) CreateLinkedBranch(_ int, _ string, _ bool) (string, error) {
	return "", nil
}

func (m *mockGitHubClient) GetIssue(_ int) (*sow.Issue, error) {
	return nil, nil
}

// TestCheckGitHubCLI tests the checkGitHubCLI function.
func TestCheckGitHubCLI(t *testing.T) {
	tests := []struct {
		name        string
		installErr  error
		authErr     error
		wantErr     bool
		errContains string
	}{
		{
			name:       "both checks pass",
			installErr: nil,
			authErr:    nil,
			wantErr:    false,
		},
		{
			name:        "not installed",
			installErr:  sow.ErrGHNotInstalled{},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "not authenticated",
			installErr:  nil,
			authErr:     sow.ErrGHNotAuthenticated{},
			wantErr:     true,
			errContains: "not authenticated",
		},
		{
			name:        "generic installation error",
			installErr:  errors.New("generic error"),
			wantErr:     true,
			errContains: "failed to check gh installation",
		},
		{
			name:        "generic authentication error",
			installErr:  nil,
			authErr:     errors.New("generic auth error"),
			wantErr:     true,
			errContains: "failed to check gh authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockGitHubClient{
				checkInstalledErr:     tt.installErr,
				checkAuthenticatedErr: tt.authErr,
			}

			err := checkGitHubCLI(mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkGitHubCLI() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message %q does not contain %q",
						err.Error(), tt.errContains)
				}
			}
		})
	}
}

// TestFormatGitHubError_RateLimit tests rate limit error formatting.
func TestFormatGitHubError_RateLimit(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		contains    []string
		notContains []string
	}{
		{
			name: "rate limit error",
			err: sow.ErrGHCommand{
				Command: "issue list",
				Stderr:  "API rate limit exceeded for user",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"rate limit", "Wait a few minutes"},
			notContains: []string{"network"},
		},
		{
			name: "case insensitive matching",
			err: sow.ErrGHCommand{
				Command: "test",
				Stderr:  "RATE LIMIT exceeded",
				Err:     errors.New("exit code 1"),
			},
			contains: []string{"rate limit"},
		},
	}
	runFormatGitHubErrorTests(t, tests)
}

// TestFormatGitHubError_Network tests network error formatting.
func TestFormatGitHubError_Network(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		contains    []string
		notContains []string
	}{
		{
			name: "network error - connection",
			err: sow.ErrGHCommand{
				Command: "issue view",
				Stderr:  "dial tcp: connection refused",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"Cannot reach GitHub", "internet connection"},
			notContains: []string{"rate limit"},
		},
		{
			name: "network error - timeout",
			err: sow.ErrGHCommand{
				Command: "issue view",
				Stderr:  "request timeout",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"Cannot reach GitHub"},
			notContains: []string{"rate limit"},
		},
		{
			name: "network error - unreachable",
			err: sow.ErrGHCommand{
				Command: "issue view",
				Stderr:  "host unreachable",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"Cannot reach GitHub"},
			notContains: []string{"rate limit"},
		},
	}
	runFormatGitHubErrorTests(t, tests)
}

// TestFormatGitHubError_NotFound tests not found error formatting.
func TestFormatGitHubError_NotFound(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		contains    []string
		notContains []string
	}{
		{
			name: "not found error",
			err: sow.ErrGHCommand{
				Command: "issue view 999",
				Stderr:  "issue not found",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"Resource not found", "issue number"},
			notContains: []string{"network"},
		},
		{
			name: "not found error - does not exist",
			err: sow.ErrGHCommand{
				Command: "repo view",
				Stderr:  "repository does not exist",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"Resource not found"},
			notContains: []string{"network"},
		},
	}
	runFormatGitHubErrorTests(t, tests)
}

// TestFormatGitHubError_Permission tests permission error formatting.
func TestFormatGitHubError_Permission(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		contains    []string
		notContains []string
	}{
		{
			name: "permission denied error",
			err: sow.ErrGHCommand{
				Command: "issue create",
				Stderr:  "permission denied",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"Permission denied", "repository access"},
			notContains: []string{"network"},
		},
		{
			name: "permission denied - forbidden",
			err: sow.ErrGHCommand{
				Command: "issue create",
				Stderr:  "403 Forbidden",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"Permission denied"},
			notContains: []string{"network"},
		},
		{
			name: "permission denied - not authorized",
			err: sow.ErrGHCommand{
				Command: "issue create",
				Stderr:  "not authorized to access",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"Permission denied"},
			notContains: []string{"network"},
		},
	}
	runFormatGitHubErrorTests(t, tests)
}

// TestFormatGitHubError_Other tests other error formatting cases.
func TestFormatGitHubError_Other(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		contains    []string
		notContains []string
	}{
		{
			name: "unknown error",
			err: sow.ErrGHCommand{
				Command: "some command",
				Stderr:  "some unexpected error",
				Err:     errors.New("exit code 1"),
			},
			contains:    []string{"GitHub command failed", "gh some command", "gh auth status"},
			notContains: []string{"network", "rate limit"},
		},
		{
			name:     "non-GitHub command error",
			err:      errors.New("some other error"),
			contains: []string{"GitHub operation failed"},
		},
	}
	runFormatGitHubErrorTests(t, tests)
}

// runFormatGitHubErrorTests is a helper that runs test cases for formatGitHubError.
func runFormatGitHubErrorTests(t *testing.T, tests []struct {
	name        string
	err         error
	contains    []string
	notContains []string
}) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatGitHubError(tt.err)

			for _, substr := range tt.contains {
				if !strings.Contains(strings.ToLower(result), strings.ToLower(substr)) {
					t.Errorf("formatGitHubError() result does not contain %q\nGot: %s", substr, result)
				}
			}

			for _, substr := range tt.notContains {
				if strings.Contains(strings.ToLower(result), strings.ToLower(substr)) {
					t.Errorf("formatGitHubError() result should not contain %q\nGot: %s", substr, result)
				}
			}
		})
	}
}

// TestCheckIssueLinkedBranch tests the checkIssueLinkedBranch function.
func TestCheckIssueLinkedBranch(t *testing.T) {
	tests := []struct {
		name          string
		branches      []sow.LinkedBranch
		branchesErr   error
		wantErr       bool
		errContains   string
		notErrContain string
	}{
		{
			name:     "no linked branches - OK",
			branches: []sow.LinkedBranch{},
			wantErr:  false,
		},
		{
			name: "one linked branch - error",
			branches: []sow.LinkedBranch{
				{Name: "feat/auth", URL: "https://github.com/test/repo/tree/feat/auth"},
			},
			wantErr:     true,
			errContains: "feat/auth",
		},
		{
			name: "multiple linked branches - error with first",
			branches: []sow.LinkedBranch{
				{Name: "feat/first-branch", URL: "https://github.com/test/repo/tree/feat/first-branch"},
				{Name: "feat/second-branch", URL: "https://github.com/test/repo/tree/feat/second-branch"},
			},
			wantErr:       true,
			errContains:   "feat/first-branch",
			notErrContain: "second-branch",
		},
		{
			name:        "GetLinkedBranches error",
			branchesErr: sow.ErrGHCommand{Command: "issue develop --list", Stderr: "some error"},
			wantErr:     true,
			errContains: "GitHub command failed",
		},
		{
			name:        "GetLinkedBranches network error",
			branchesErr: sow.ErrGHCommand{Command: "issue develop --list", Stderr: "network timeout"},
			wantErr:     true,
			errContains: "Cannot reach GitHub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockGitHubClient{
				linkedBranches:    tt.branches,
				linkedBranchesErr: tt.branchesErr,
			}

			err := checkIssueLinkedBranch(mock, 123)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkIssueLinkedBranch() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message %q does not contain %q",
						err.Error(), tt.errContains)
				}
			}

			if err != nil && tt.notErrContain != "" {
				if strings.Contains(err.Error(), tt.notErrContain) {
					t.Errorf("error message %q should not contain %q",
						err.Error(), tt.notErrContain)
				}
			}
		})
	}
}

// TestFilterIssuesBySowLabel tests the filterIssuesBySowLabel function.
func TestFilterIssuesBySowLabel(t *testing.T) {
	tests := []struct {
		name     string
		issues   []sow.Issue
		expected int
	}{
		{
			name:     "empty input",
			issues:   []sow.Issue{},
			expected: 0,
		},
		{
			name: "all issues have sow label",
			issues: []sow.Issue{
				{Number: 1, Title: "Issue 1", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "sow"}}},
				{Number: 2, Title: "Issue 2", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "sow"}}},
			},
			expected: 2,
		},
		{
			name: "no issues have sow label",
			issues: []sow.Issue{
				{Number: 1, Title: "Issue 1", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "bug"}}},
				{Number: 2, Title: "Issue 2", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "feature"}}},
			},
			expected: 0,
		},
		{
			name: "mixed labels",
			issues: []sow.Issue{
				{Number: 1, Title: "Issue 1", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "sow"}}},
				{Number: 2, Title: "Issue 2", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "bug"}}},
				{Number: 3, Title: "Issue 3", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "sow"}, {Name: "feature"}}},
			},
			expected: 2,
		},
		{
			name: "issues with no labels",
			issues: []sow.Issue{
				{Number: 1, Title: "Issue 1", Labels: []struct {
					Name string `json:"name"`
				}{}},
				{Number: 2, Title: "Issue 2", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "sow"}}},
			},
			expected: 1,
		},
		{
			name: "case sensitive matching",
			issues: []sow.Issue{
				{Number: 1, Title: "Issue 1", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "sow"}}},
				{Number: 2, Title: "Issue 2", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "SOW"}}},
				{Number: 3, Title: "Issue 3", Labels: []struct {
					Name string `json:"name"`
				}{{Name: "Sow"}}},
			},
			expected: 1, // Only exact match "sow"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterIssuesBySowLabel(tt.issues)

			if len(result) != tt.expected {
				t.Errorf("filterIssuesBySowLabel() returned %d issues; want %d", len(result), tt.expected)
			}

			// Verify all returned issues have the "sow" label
			for _, issue := range result {
				if !issue.HasLabel("sow") {
					t.Errorf("issue #%d in result does not have 'sow' label", issue.Number)
				}
			}
		})
	}
}

// TestErrorGitHubNotAuthenticated tests the errorGitHubNotAuthenticated function.
func TestErrorGitHubNotAuthenticated(t *testing.T) {
	result := errorGitHubNotAuthenticated()

	expectedParts := []string{
		"GitHub CLI not authenticated",
		"not logged in",
		"gh auth login",
		"From branch name",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("errorGitHubNotAuthenticated() result does not contain %q\nGot: %s", part, result)
		}
	}
}

// TestDiscoverKnowledgeFiles tests the discoverKnowledgeFiles function.
func TestDiscoverKnowledgeFiles(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string // Returns knowledge dir path
		expected []string
		wantErr  bool
	}{
		{
			name: "discovers markdown files in flat directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
				if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
					t.Fatalf("failed to create knowledge dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "README.md"), []byte("# Readme"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "overview.md"), []byte("# Overview"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return knowledgeDir
			},
			expected: []string{"README.md", "overview.md"},
			wantErr:  false,
		},
		{
			name: "discovers files in nested directories",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
				if err := os.MkdirAll(filepath.Join(knowledgeDir, "designs"), 0755); err != nil {
					t.Fatalf("failed to create designs dir: %v", err)
				}
				if err := os.MkdirAll(filepath.Join(knowledgeDir, "adrs"), 0755); err != nil {
					t.Fatalf("failed to create adrs dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "README.md"), []byte("# Readme"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "designs", "api.md"), []byte("# API"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "adrs", "001-decision.md"), []byte("# Decision"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return knowledgeDir
			},
			expected: []string{"README.md", "adrs/001-decision.md", "designs/api.md"},
			wantErr:  false,
		},
		{
			name: "handles non-existent directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
				// Don't create the directory
				return knowledgeDir
			},
			expected: []string{},
			wantErr:  false,
		},
		{
			name: "handles empty directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
				if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
					t.Fatalf("failed to create knowledge dir: %v", err)
				}
				// Don't create any files
				return knowledgeDir
			},
			expected: []string{},
			wantErr:  false,
		},
		{
			name: "sorts results alphabetically",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
				if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
					t.Fatalf("failed to create knowledge dir: %v", err)
				}
				// Create files in non-alphabetical order
				if err := os.WriteFile(filepath.Join(knowledgeDir, "zebra.md"), []byte("z"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "alpha.md"), []byte("a"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "beta.md"), []byte("b"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return knowledgeDir
			},
			expected: []string{"alpha.md", "beta.md", "zebra.md"},
			wantErr:  false,
		},
		{
			name: "returns relative paths from knowledge directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
				if err := os.MkdirAll(filepath.Join(knowledgeDir, "deep", "nested", "path"), 0755); err != nil {
					t.Fatalf("failed to create nested dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "deep", "nested", "path", "file.md"), []byte("# File"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return knowledgeDir
			},
			expected: []string{"deep/nested/path/file.md"},
			wantErr:  false,
		},
		{
			name: "discovers various file types",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
				if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
					t.Fatalf("failed to create knowledge dir: %v", err)
				}
				// Create files with different extensions
				if err := os.WriteFile(filepath.Join(knowledgeDir, "doc.md"), []byte("md"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "data.json"), []byte("{}"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "notes.txt"), []byte("notes"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return knowledgeDir
			},
			expected: []string{"data.json", "doc.md", "notes.txt"},
			wantErr:  false,
		},
		{
			name: "excludes directories from results",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
				if err := os.MkdirAll(filepath.Join(knowledgeDir, "subdir"), 0755); err != nil {
					t.Fatalf("failed to create subdir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(knowledgeDir, "file.md"), []byte("file"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return knowledgeDir
			},
			expected: []string{"file.md"},
			wantErr:  false,
		},
		{
			name: "handles deeply nested files",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				knowledgeDir := filepath.Join(dir, ".sow", "knowledge")
				deepPath := filepath.Join(knowledgeDir, "a", "b", "c", "d", "e")
				if err := os.MkdirAll(deepPath, 0755); err != nil {
					t.Fatalf("failed to create deep path: %v", err)
				}
				if err := os.WriteFile(filepath.Join(deepPath, "deep.md"), []byte("deep"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return knowledgeDir
			},
			expected: []string{"a/b/c/d/e/deep.md"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			got, err := discoverKnowledgeFiles(dir)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Errorf("got %d files, want %d\nGot: %v\nWant: %v", len(got), len(tt.expected), got, tt.expected)
				return
			}

			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("file[%d] = %q; want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

// TestValidateStateTransition_FileSelectTransitions tests state transitions involving StateFileSelect.
func TestValidateStateTransition_FileSelectTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    WizardState
		to      WizardState
		wantErr bool
	}{
		// Valid transitions to StateFileSelect
		{
			name:    "StateNameEntry to StateFileSelect is valid",
			from:    StateNameEntry,
			to:      StateFileSelect,
			wantErr: false,
		},
		{
			name:    "StateTypeSelect to StateFileSelect is valid (GitHub issue flow)",
			from:    StateTypeSelect,
			to:      StateFileSelect,
			wantErr: false,
		},
		// Valid transitions from StateFileSelect
		{
			name:    "StateFileSelect to StatePromptEntry is valid",
			from:    StateFileSelect,
			to:      StatePromptEntry,
			wantErr: false,
		},
		{
			name:    "StateFileSelect to StateCancelled is valid",
			from:    StateFileSelect,
			to:      StateCancelled,
			wantErr: false,
		},
		// Invalid transitions
		{
			name:    "StateFileSelect to StateComplete is invalid",
			from:    StateFileSelect,
			to:      StateComplete,
			wantErr: true,
		},
		{
			name:    "StateFileSelect to StateEntry is invalid",
			from:    StateFileSelect,
			to:      StateEntry,
			wantErr: true,
		},
		{
			name:    "StatePromptEntry to StateFileSelect is invalid (backwards)",
			from:    StatePromptEntry,
			to:      StateFileSelect,
			wantErr: true,
		},
		// Verify StateNameEntry no longer goes directly to StatePromptEntry
		{
			name:    "StateNameEntry to StatePromptEntry is invalid (should go through FileSelect)",
			from:    StateNameEntry,
			to:      StatePromptEntry,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStateTransition(tt.from, tt.to)

			if tt.wantErr && err == nil {
				t.Errorf("validateStateTransition(%s, %s) expected error, got nil", tt.from, tt.to)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateStateTransition(%s, %s) unexpected error: %v", tt.from, tt.to, err)
			}
		})
	}
}
