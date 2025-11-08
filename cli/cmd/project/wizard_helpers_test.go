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
				Type:      "design",
				Phase:     "active",
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
