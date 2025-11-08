package project

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
		{"feat..auth", "branch name cannot contain .."},
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

	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create a branch using git CLI
	cmd = exec.Command("git", "branch", "feat/test-branch")
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

	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create branch and worktree
	branchName := "feat/test-worktree"
	worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)

	// Create branch
	cmd = exec.Command("git", "branch", branchName)
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

	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create branch and worktree with project
	branchName := "feat/test-full"
	worktreePath := filepath.Join(tmpDir, ".sow", "worktrees", branchName)

	// Create branch
	cmd = exec.Command("git", "branch", branchName)
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
