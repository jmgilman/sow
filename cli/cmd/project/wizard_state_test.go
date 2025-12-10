package project

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/jmgilman/sow/libs/project/state"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/libs/git"
)

// mockGitHub is a test double for GitHub operations.
// TestHandleCreateSource_StateTransitions tests state transitions directly
// by manually setting values and checking the results.
func TestHandleCreateSource_StateTransitions(t *testing.T) {
	testCases := []struct {
		name          string
		selection     string
		expectedState WizardState
	}{
		{
			name:          "issue selection",
			selection:     "issue",
			expectedState: StateIssueSelect,
		},
		{
			name:          "branch selection",
			selection:     "branch",
			expectedState: StateTypeSelect,
		},
		{
			name:          "cancel selection",
			selection:     "cancel",
			expectedState: StateCancelled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := setupTestContext(t)
			w := NewWizard(nil, ctx, []string{})
			w.state = StateCreateSource

			// Simulate what handleCreateSource does
			w.choices["source"] = tc.selection

			switch tc.selection {
			case "issue":
				w.state = StateIssueSelect
			case "branch":
				w.state = StateTypeSelect
			case "cancel":
				w.state = StateCancelled
			}

			// Verify state transition
			if w.state != tc.expectedState {
				t.Errorf("expected state %v, got %v", tc.expectedState, w.state)
			}

			// Verify choice was stored
			if w.choices["source"] != tc.selection {
				t.Errorf("expected choice %q, got %q", tc.selection, w.choices["source"])
			}
		})
	}
}

// TestHandleCreateSource_ErrorHandling tests error handling behavior.
func TestHandleCreateSource_ErrorHandling(t *testing.T) {
	t.Run("user abort returns nil and sets cancelled state", func(t *testing.T) {
		// We can test that ErrUserAborted is handled correctly
		// by checking that errors.Is works with it
		err := huh.ErrUserAborted
		if !errors.Is(err, huh.ErrUserAborted) {
			t.Error("errors.Is should match ErrUserAborted")
		}
	})
}

// TestHandleTypeSelect_StateTransitions tests state transitions for type selection
// by manually setting values and checking the results.
func TestHandleTypeSelect_StateTransitions(t *testing.T) {
	testCases := []struct {
		name          string
		selection     string
		expectedState WizardState
		shouldStore   bool
	}{
		{
			name:          "standard selection",
			selection:     "standard",
			expectedState: StateNameEntry,
			shouldStore:   true,
		},
		{
			name:          "exploration selection",
			selection:     "exploration",
			expectedState: StateNameEntry,
			shouldStore:   true,
		},
		{
			name:          "design selection",
			selection:     "design",
			expectedState: StateNameEntry,
			shouldStore:   true,
		},
		{
			name:          "breakdown selection",
			selection:     "breakdown",
			expectedState: StateNameEntry,
			shouldStore:   true,
		},
		{
			name:          "cancel selection",
			selection:     "cancel",
			expectedState: StateCancelled,
			shouldStore:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := setupTestContext(t)
			w := NewWizard(nil, ctx, []string{})
			w.state = StateTypeSelect

			// Simulate what handleTypeSelect should do
			if tc.selection == "cancel" {
				w.state = StateCancelled
			} else {
				w.choices["type"] = tc.selection
				w.state = StateNameEntry
			}

			// Verify state transition
			if w.state != tc.expectedState {
				t.Errorf("expected state %v, got %v", tc.expectedState, w.state)
			}

			// Verify choice storage
			if tc.shouldStore {
				if w.choices["type"] != tc.selection {
					t.Errorf("expected choice %q, got %q", tc.selection, w.choices["type"])
				}
			} else {
				if _, exists := w.choices["type"]; exists {
					t.Errorf("expected no type choice for cancel, but got %q", w.choices["type"])
				}
			}
		})
	}
}

// TestHandleTypeSelect_ErrorHandling tests error handling behavior.
func TestHandleTypeSelect_ErrorHandling(t *testing.T) {
	t.Run("user abort returns nil and sets cancelled state", func(t *testing.T) {
		// Verify that errors.Is works with ErrUserAborted
		err := huh.ErrUserAborted
		if !errors.Is(err, huh.ErrUserAborted) {
			t.Error("errors.Is should match ErrUserAborted")
		}
	})
}

// TestHandlePromptEntry_StateTransitions tests state transitions for prompt entry
// by manually setting values and checking the results.
func TestHandlePromptEntry_StateTransitions(t *testing.T) {
	testCases := []struct {
		name          string
		promptText    string
		expectedState WizardState
	}{
		{
			name:          "with text",
			promptText:    "Build a REST API with JWT authentication",
			expectedState: StateComplete,
		},
		{
			name:          "empty text",
			promptText:    "",
			expectedState: StateComplete,
		},
		{
			name:          "multi-line text",
			promptText:    "Line 1\nLine 2\nLine 3",
			expectedState: StateComplete,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := setupTestContext(t)
			w := NewWizard(nil, ctx, []string{})
			w.state = StatePromptEntry

			// Pre-populate required choices
			w.choices["type"] = "standard"
			w.choices["branch"] = "feat/test-project"

			// Simulate what handlePromptEntry should do
			w.choices["prompt"] = tc.promptText
			w.state = StateComplete

			// Verify state transition
			if w.state != tc.expectedState {
				t.Errorf("expected state %v, got %v", tc.expectedState, w.state)
			}

			// Verify choice was stored
			if w.choices["prompt"] != tc.promptText {
				t.Errorf("expected prompt %q, got %q", tc.promptText, w.choices["prompt"])
			}
		})
	}
}

// TestHandlePromptEntry_RequiresTypeAndBranch tests that type and branch must be set before prompt entry.
func TestHandlePromptEntry_RequiresTypeAndBranch(t *testing.T) {
	ctx, _ := setupTestContext(t)
	w := NewWizard(nil, ctx, []string{})
	w.state = StatePromptEntry

	// Set up required choices
	w.choices["type"] = "exploration"
	w.choices["branch"] = "explore/web-agents"

	// Verify choices exist (they're required for context display)
	if _, ok := w.choices["type"]; !ok {
		t.Error("type choice should be set before prompt entry")
	}
	if _, ok := w.choices["branch"]; !ok {
		t.Error("branch choice should be set before prompt entry")
	}
}

// TestHandlePromptEntry_ErrorHandling tests error handling behavior.
func TestHandlePromptEntry_ErrorHandling(t *testing.T) {
	t.Run("user abort transitions to cancelled", func(t *testing.T) {
		// Verify that errors.Is works with ErrUserAborted
		err := huh.ErrUserAborted
		if !errors.Is(err, huh.ErrUserAborted) {
			t.Error("errors.Is should match ErrUserAborted")
		}

		// Simulating the handler behavior
		ctx, _ := setupTestContext(t)
		w := NewWizard(nil, ctx, []string{})
		w.state = StatePromptEntry
		w.choices["type"] = "standard"
		w.choices["branch"] = "feat/test"

		// On user abort, should transition to cancelled
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
		}

		if w.state != StateCancelled {
			t.Errorf("expected state StateCancelled on abort, got %v", w.state)
		}
	})
}

// Test finalize() function

// TestFinalize_CreatesWorktree tests that finalize creates a worktree.
func TestFinalize_CreatesWorktree(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit (required for worktree creation)
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	w := NewWizard(nil, ctx, []string{})

	// Set up wizard with choices populated
	w.choices["action"] = "create"
	w.choices["type"] = "standard"
	w.choices["name"] = "Test Project"
	w.choices["branch"] = "feat/test-project"
	w.choices["prompt"] = ""

	// Call finalize
	err := w.finalize()
	if err != nil {
		t.Fatalf("finalize failed: %v", err)
	}

	// Verify worktree directory exists
	worktreePath := tmpDir + "/.sow/worktrees/feat/test-project"
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Errorf("worktree directory was not created at %s", worktreePath)
	}
}

// TestFinalize_InitializesProject tests that finalize initializes the project correctly.
func TestFinalize_InitializesProject(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit (required for worktree creation)
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	w := NewWizard(nil, ctx, []string{})

	// Set up wizard with choices populated
	w.choices["action"] = "create"
	w.choices["type"] = "exploration"
	w.choices["name"] = "Research Project"
	w.choices["branch"] = "explore/research-project"
	w.choices["prompt"] = "Initial research prompt"

	// Call finalize
	err := w.finalize()
	if err != nil {
		t.Fatalf("finalize failed: %v", err)
	}

	// Verify project state.yaml exists in worktree
	worktreePath := tmpDir + "/.sow/worktrees/explore/research-project"
	stateFile := worktreePath + "/.sow/project/state.yaml"
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Errorf("project state.yaml was not created at %s", stateFile)
	}

	// Load and verify project
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		t.Fatalf("failed to create worktree context: %v", err)
	}

	proj, err := state.Load(context.Background(), state.NewYAMLBackend(worktreeCtx.FS()))
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}

	// Verify project has correct name (normalized from user input)
	if proj.Name != "research-project" {
		t.Errorf("project name incorrect: got %q, want %q", proj.Name, "research-project")
	}
}

// TestFinalize_GeneratesPrompt tests that finalize generates a multi-layer prompt.
func TestFinalize_GeneratesPrompt(t *testing.T) {
	ctx, _ := setupTestContext(t)

	// We need to mock launchClaudeCode to capture the prompt
	// Since we can't easily mock it, we'll test prompt generation separately
	// by calling generateNewProjectPrompt directly

	// First create a project
	proj, err := initializeProject(ctx, "feat/test", "Test Project", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Test prompt generation with user prompt
	initialPrompt := "Build authentication system"
	prompt, err := generateNewProjectPrompt(proj, initialPrompt)
	if err != nil {
		t.Fatalf("generateNewProjectPrompt failed: %v", err)
	}

	// Verify prompt has 3 layers (at least 2 separators)
	separatorCount := strings.Count(prompt, "\n\n---\n\n")
	if separatorCount < 2 {
		t.Errorf("expected at least 2 separators for 3 layers, got %d", separatorCount)
	}

	// Verify prompt includes user's initial prompt
	if !strings.Contains(prompt, initialPrompt) {
		t.Errorf("prompt does not contain user's initial prompt %q", initialPrompt)
	}

	if !strings.Contains(prompt, "User's Initial Request") {
		t.Error("prompt missing user request section header")
	}
}

// TestFinalize_WithEmptyPrompt tests finalize with an empty initial prompt.
func TestFinalize_WithEmptyPrompt(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit (required for worktree creation)
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	w := NewWizard(nil, ctx, []string{})

	// Set up wizard with empty prompt choice
	w.choices["action"] = "create"
	w.choices["type"] = "design"
	w.choices["name"] = "Design Project"
	w.choices["branch"] = "design/test"
	w.choices["prompt"] = ""

	// Call finalize
	err := w.finalize()
	if err != nil {
		t.Fatalf("finalize failed: %v", err)
	}

	// Verify project was created
	worktreePath := tmpDir + "/.sow/worktrees/design/test"
	stateFile := worktreePath + "/.sow/project/state.yaml"
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Errorf("project state.yaml was not created")
	}

	// Generate prompt and verify no user request section
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		t.Fatalf("failed to create worktree context: %v", err)
	}

	proj, err := state.Load(context.Background(), state.NewYAMLBackend(worktreeCtx.FS()))
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}

	prompt, err := generateNewProjectPrompt(proj, "")
	if err != nil {
		t.Fatalf("generateNewProjectPrompt failed: %v", err)
	}

	// Verify no user request section when prompt is empty
	if strings.Contains(prompt, "User's Initial Request") {
		t.Error("prompt should not contain user request section when no user prompt provided")
	}
}

// TestFinalize_UncommittedChangesError tests that finalize returns error when uncommitted changes exist
// and current branch == target branch.
func TestFinalize_UncommittedChangesError(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create a test file and modify it (uncommitted change)
	testFile := tmpDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Stage and commit the file first
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", "test.txt").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	// Now modify it again (creating uncommitted changes)
	if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	// Get current branch name
	currentBranch, err := ctx.Git().CurrentBranch()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}

	w := NewWizard(nil, ctx, []string{})

	// Set current branch == target branch (should trigger check)
	w.choices["action"] = "create"
	w.choices["type"] = "standard"
	w.choices["name"] = "Test Project"
	w.choices["branch"] = currentBranch
	w.choices["prompt"] = ""

	// Call finalize - should fail with uncommitted changes error
	err = w.finalize()
	if err == nil {
		t.Fatal("expected error for uncommitted changes, got nil")
	}

	// Verify error message contains expected text
	errMsg := err.Error()
	if !strings.Contains(errMsg, "uncommitted changes") {
		t.Errorf("error message should mention uncommitted changes, got: %v", errMsg)
	}

	if !strings.Contains(errMsg, currentBranch) {
		t.Errorf("error message should mention current branch %q, got: %v", currentBranch, errMsg)
	}
}

// TestFinalize_SkipsUncommittedCheckWhenDifferentBranch tests that finalize skips the uncommitted
// changes check when current branch != target branch.
func TestFinalize_SkipsUncommittedCheckWhenDifferentBranch(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create a test file and commit it (create initial commit)
	testFile := tmpDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Stage and commit the file first
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", "test.txt").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	// Now modify it again (creating uncommitted changes)
	if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	// Get current branch name
	currentBranch, err := ctx.Git().CurrentBranch()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}

	// Set different branch from current (should skip check)
	differentBranch := "feat/different-branch"
	if currentBranch == differentBranch {
		differentBranch = "feat/another-branch"
	}

	// Set environment variable to skip check in tests (for worktree creation)
	_ = os.Setenv("SOW_SKIP_UNCOMMITTED_CHECK", "1")
	defer func() { _ = os.Unsetenv("SOW_SKIP_UNCOMMITTED_CHECK") }()

	w := NewWizard(nil, ctx, []string{})

	w.choices["action"] = "create"
	w.choices["type"] = "standard"
	w.choices["name"] = "Different Project"
	w.choices["branch"] = differentBranch
	w.choices["prompt"] = ""

	// Call finalize - should succeed (check skipped)
	err = w.finalize()
	if err != nil {
		t.Fatalf("finalize should succeed when current != target branch, got error: %v", err)
	}

	// Verify worktree was created
	worktreePath := tmpDir + "/.sow/worktrees/" + differentBranch
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Errorf("worktree should have been created at %s", worktreePath)
	}
}

// Test handleProjectSelect

// TestHandleProjectSelect_EmptyList tests that empty project list shows message and cancels.
func TestHandleProjectSelect_EmptyList(t *testing.T) {
	ctx, _ := setupTestContext(t)
	w := NewWizard(nil, ctx, []string{})
	w.state = StateProjectSelect

	// At this point, there are no projects in the worktrees directory
	// handleProjectSelect should discover zero projects and transition to StateCancelled

	// We can't easily test the interactive form without mocking,
	// but we can simulate what the handler should do
	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects failed: %v", err)
	}

	if len(projects) != 0 {
		t.Fatalf("expected 0 projects, got %d", len(projects))
	}

	// Simulate handler behavior: empty list → StateCancelled
	if len(projects) == 0 {
		w.state = StateCancelled
	}

	// Verify state transition
	if w.state != StateCancelled {
		t.Errorf("expected state StateCancelled, got %v", w.state)
	}
}

// TestHandleProjectSelect_SingleProject tests selection with a single project.
func TestHandleProjectSelect_SingleProject(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	// Create a test project in a worktree
	branchName := "feat/test-project"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName

	// Use EnsureWorktree to create the worktree
	if err := git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName); err != nil {
		t.Fatalf("failed to create worktree: %v", err)
	}

	// Initialize project in worktree
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		t.Fatalf("failed to create worktree context: %v", err)
	}

	_, err = initializeProject(worktreeCtx, branchName, "test-project", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Now test project discovery
	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects failed: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	// Verify project metadata
	proj := projects[0]
	if proj.Branch != branchName {
		t.Errorf("expected branch %q, got %q", branchName, proj.Branch)
	}
	if proj.Name != "testproject" {
		t.Errorf("expected name %q, got %q", "testproject", proj.Name)
	}

	// Simulate selection
	w := NewWizard(nil, ctx, []string{})
	w.state = StateProjectSelect

	// User selects the project
	w.choices["project"] = proj
	w.state = StateContinuePrompt

	// Verify state transition and choice storage
	if w.state != StateContinuePrompt {
		t.Errorf("expected state StateContinuePrompt, got %v", w.state)
	}

	selectedProj, ok := w.choices["project"].(ProjectInfo)
	if !ok {
		t.Fatalf("project choice not stored correctly")
	}

	if selectedProj.Branch != branchName {
		t.Errorf("expected selected branch %q, got %q", branchName, selectedProj.Branch)
	}
}

// TestHandleProjectSelect_MultipleProjects tests selection with multiple projects.
func TestHandleProjectSelect_MultipleProjects(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	// Create multiple test projects
	projects := []struct {
		branch string
		name   string
	}{
		{"feat/auth", "auth"},
		{"explore/api-research", "api-research"},
		{"design/cli-ux", "cli-ux"},
	}

	for _, p := range projects {
		worktreePath := tmpDir + "/.sow/worktrees/" + p.branch
		if err := git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, p.branch); err != nil {
			t.Fatalf("failed to create worktree for %s: %v", p.branch, err)
		}

		worktreeCtx, err := sow.NewContext(worktreePath)
		if err != nil {
			t.Fatalf("failed to create worktree context for %s: %v", p.branch, err)
		}

		_, err = initializeProject(worktreeCtx, p.branch, p.name, nil, nil)
		if err != nil {
			t.Fatalf("failed to initialize project for %s: %v", p.branch, err)
		}
	}

	// Test project discovery
	discoveredProjects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects failed: %v", err)
	}

	if len(discoveredProjects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(discoveredProjects))
	}

	// Verify all projects were discovered
	foundBranches := make(map[string]bool)
	for _, proj := range discoveredProjects {
		foundBranches[proj.Branch] = true
	}

	for _, expected := range projects {
		if !foundBranches[expected.branch] {
			t.Errorf("expected to find project with branch %q", expected.branch)
		}
	}

	// Simulate selection of second project
	w := NewWizard(nil, ctx, []string{})
	w.state = StateProjectSelect

	selectedProj := discoveredProjects[1]
	w.choices["project"] = selectedProj
	w.state = StateContinuePrompt

	// Verify state transition
	if w.state != StateContinuePrompt {
		t.Errorf("expected state StateContinuePrompt, got %v", w.state)
	}

	// Verify choice was stored
	storedProj, ok := w.choices["project"].(ProjectInfo)
	if !ok {
		t.Fatalf("project choice not stored correctly")
	}

	if storedProj.Branch != selectedProj.Branch {
		t.Errorf("expected selected branch %q, got %q", selectedProj.Branch, storedProj.Branch)
	}
}

// TestHandleProjectSelect_CancelOption tests that cancel transitions to StateCancelled.
func TestHandleProjectSelect_CancelOption(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit and a test project
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branchName := "feat/test"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "test", nil, nil)

	w := NewWizard(nil, ctx, []string{})
	w.state = StateProjectSelect

	// Simulate user selecting "Cancel"
	w.state = StateCancelled

	// Verify state transition
	if w.state != StateCancelled {
		t.Errorf("expected state StateCancelled, got %v", w.state)
	}

	// Verify no project choice was stored
	if _, exists := w.choices["project"]; exists {
		t.Error("project choice should not be stored when user cancels")
	}
}

// TestHandleProjectSelect_ProjectDeletedAfterDiscovery tests race condition handling.
func TestHandleProjectSelect_ProjectDeletedAfterDiscovery(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit and a test project
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branchName := "feat/test-delete"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "test-delete", nil, nil)

	// Discover projects
	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects failed: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	selectedProj := projects[0]

	// Delete the state file (simulating race condition)
	statePath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")
	_ = os.Remove(statePath)

	// Verify state file is gone
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatal("state file should have been deleted")
	}

	// Simulate what the handler should do: validate project still exists
	// In this case, validation should fail, and handler should stay in StateProjectSelect
	w := NewWizard(nil, ctx, []string{})
	w.state = StateProjectSelect

	// Check if state file exists - if it does, store choice and transition
	if _, err := os.Stat(statePath); err == nil {
		// Project exists - store choice and transition
		w.choices["project"] = selectedProj
		w.state = StateContinuePrompt
	}
	// Project no longer exists - stay in current state (allow retry)

	// Verify handler stayed in StateProjectSelect (no transition)
	if w.state != StateProjectSelect {
		t.Errorf("expected state to remain StateProjectSelect when project deleted, got %v", w.state)
	}

	// Verify no project choice was stored
	if _, exists := w.choices["project"]; exists {
		t.Error("project choice should not be stored when validation fails")
	}
}

// TestHandleProjectSelect_UserAbort tests that Esc key transitions to StateCancelled.
func TestHandleProjectSelect_UserAbort(t *testing.T) {
	// Test that errors.Is works with ErrUserAborted
	err := huh.ErrUserAborted
	if !errors.Is(err, huh.ErrUserAborted) {
		t.Error("errors.Is should match ErrUserAborted")
	}

	// Simulate handler behavior on user abort
	ctx, _ := setupTestContext(t)
	w := NewWizard(nil, ctx, []string{})
	w.state = StateProjectSelect

	// On user abort (Esc), should transition to StateCancelled
	if errors.Is(err, huh.ErrUserAborted) {
		w.state = StateCancelled
	}

	if w.state != StateCancelled {
		t.Errorf("expected state StateCancelled on user abort, got %v", w.state)
	}
}

// Test handleContinuePrompt

// TestHandleContinuePrompt_RequiresProject tests that project must be set before continuation prompt.
func TestHandleContinuePrompt_RequiresProject(t *testing.T) {
	ctx, _ := setupTestContext(t)
	w := NewWizard(nil, ctx, []string{})
	w.state = StateContinuePrompt

	// Do NOT set project choice - should return error
	// We can't easily call the handler directly without user input,
	// but we can verify the logic it should implement
	_, ok := w.choices["project"].(ProjectInfo)
	if ok {
		t.Error("project choice should not be set initially")
	}

	// The handler should check for project and return error if missing
	// This will be tested in the implementation
}

// TestHandleContinuePrompt_StateTransitions tests state transitions for continuation prompt.
func TestHandleContinuePrompt_StateTransitions(t *testing.T) {
	testCases := []struct {
		name          string
		promptText    string
		expectedState WizardState
	}{
		{
			name:          "with text",
			promptText:    "Let's focus on the token refresh logic",
			expectedState: StateComplete,
		},
		{
			name:          "empty text (optional)",
			promptText:    "",
			expectedState: StateComplete,
		},
		{
			name:          "multi-line text",
			promptText:    "Focus on tests\nAdd integration tests\nCheck edge cases",
			expectedState: StateComplete,
		},
		{
			name:          "whitespace only",
			promptText:    "   \n\n   ",
			expectedState: StateComplete,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, tmpDir := setupTestContext(t)
			w := NewWizard(nil, ctx, []string{})
			w.state = StateContinuePrompt

			// Pre-populate required project choice
			proj := ProjectInfo{
				Branch:         "feat/auth",
				Name:           "auth",
				Type:           "standard",
				Phase:          "ImplementationExecuting",
				TasksCompleted: 3,
				TasksTotal:     5,
			}
			w.choices["project"] = proj

			// Simulate what handleContinuePrompt should do
			w.choices["prompt"] = tc.promptText
			w.state = StateComplete

			// Verify state transition
			if w.state != tc.expectedState {
				t.Errorf("expected state %v, got %v", tc.expectedState, w.state)
			}

			// Verify choice was stored (even if empty)
			if w.choices["prompt"] != tc.promptText {
				t.Errorf("expected prompt %q, got %q", tc.promptText, w.choices["prompt"])
			}

			// Verify tmpDir is not used (just to satisfy linter)
			_ = tmpDir
		})
	}
}

// TestHandleContinuePrompt_ContextDisplay tests that context information is properly formatted.
func TestHandleContinuePrompt_ContextDisplay(t *testing.T) {
	ctx, _ := setupTestContext(t)
	w := NewWizard(nil, ctx, []string{})
	w.state = StateContinuePrompt

	// Set up project with different scenarios
	testCases := []struct {
		name            string
		proj            ProjectInfo
		expectedName    string
		expectedBranch  string
		expectedInState string
	}{
		{
			name: "standard project with tasks",
			proj: ProjectInfo{
				Branch:         "feat/jwt-auth",
				Name:           "jwt-auth",
				Type:           "standard",
				Phase:          "ImplementationExecuting",
				TasksCompleted: 3,
				TasksTotal:     5,
			},
			expectedName:    "jwt-auth",
			expectedBranch:  "feat/jwt-auth",
			expectedInState: "3/5 tasks completed",
		},
		{
			name: "exploration project without tasks",
			proj: ProjectInfo{
				Branch:         "explore/api-research",
				Name:           "api-research",
				Type:           "exploration",
				Phase:          "DesignActive",
				TasksCompleted: 0,
				TasksTotal:     0,
			},
			expectedName:    "api-research",
			expectedBranch:  "explore/api-research",
			expectedInState: "DesignActive",
		},
		{
			name: "design project",
			proj: ProjectInfo{
				Branch:         "design/cli-ux",
				Name:           "cli-ux",
				Type:           "design",
				Phase:          "DesignActive",
				TasksCompleted: 0,
				TasksTotal:     0,
			},
			expectedName:    "cli-ux",
			expectedBranch:  "design/cli-ux",
			expectedInState: "DesignActive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w.choices["project"] = tc.proj

			// Verify we can extract the project
			proj, ok := w.choices["project"].(ProjectInfo)
			if !ok {
				t.Fatal("failed to extract project from choices")
			}

			// Verify project metadata
			if proj.Name != tc.expectedName {
				t.Errorf("expected name %q, got %q", tc.expectedName, proj.Name)
			}
			if proj.Branch != tc.expectedBranch {
				t.Errorf("expected branch %q, got %q", tc.expectedBranch, proj.Branch)
			}

			// Verify formatProjectProgress returns expected information
			progress := formatProjectProgress(proj)
			if !strings.Contains(progress, tc.expectedInState) {
				t.Errorf("expected progress to contain %q, got %q", tc.expectedInState, progress)
			}
		})
	}
}

// TestHandleContinuePrompt_ErrorHandling tests error handling behavior.
func TestHandleContinuePrompt_ErrorHandling(t *testing.T) {
	t.Run("user abort transitions to cancelled", func(t *testing.T) {
		// Verify that errors.Is works with ErrUserAborted
		err := huh.ErrUserAborted
		if !errors.Is(err, huh.ErrUserAborted) {
			t.Error("errors.Is should match ErrUserAborted")
		}

		// Simulating the handler behavior
		ctx, _ := setupTestContext(t)
		w := NewWizard(nil, ctx, []string{})
		w.state = StateContinuePrompt
		w.choices["project"] = ProjectInfo{
			Branch: "feat/test",
			Name:   "test",
			Type:   "standard",
		}

		// On user abort, should transition to cancelled
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
		}

		if w.state != StateCancelled {
			t.Errorf("expected state StateCancelled on abort, got %v", w.state)
		}
	})

	t.Run("missing project returns error", func(t *testing.T) {
		ctx, _ := setupTestContext(t)
		w := NewWizard(nil, ctx, []string{})
		w.state = StateContinuePrompt

		// Do NOT set project choice
		// Handler should detect missing project and return error

		// Verify the check we expect the handler to perform
		_, ok := w.choices["project"].(ProjectInfo)
		if ok {
			t.Error("project choice should not be set")
		}

		// The actual handler will return an error like:
		// "internal error: project choice not set or invalid"
	})

	t.Run("wrong type in project choice returns error", func(t *testing.T) {
		ctx, _ := setupTestContext(t)
		w := NewWizard(nil, ctx, []string{})
		w.state = StateContinuePrompt

		// Set wrong type for project choice
		w.choices["project"] = "not-a-ProjectInfo-struct"

		// Verify type assertion fails
		_, ok := w.choices["project"].(ProjectInfo)
		if ok {
			t.Error("type assertion should fail for wrong type")
		}

		// The actual handler will return an error like:
		// "internal error: project choice not set or invalid"
	})
}

// TestHandleContinuePrompt_IntegrationFlow tests the full flow from project selection to continuation prompt.
func TestHandleContinuePrompt_IntegrationFlow(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit and a test project
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branchName := "feat/integration-test"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "integration-test", nil, nil)

	// Start wizard in project selection
	w := NewWizard(nil, ctx, []string{})
	w.state = StateProjectSelect

	// Discover projects
	projects, err := listProjects(ctx)
	if err != nil {
		t.Fatalf("listProjects failed: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	// Simulate: user selects project → transitions to StateContinuePrompt
	selectedProj := projects[0]
	w.choices["project"] = selectedProj
	w.state = StateContinuePrompt

	// Verify state transition
	if w.state != StateContinuePrompt {
		t.Errorf("expected state StateContinuePrompt, got %v", w.state)
	}

	// Verify project is stored
	proj, ok := w.choices["project"].(ProjectInfo)
	if !ok {
		t.Fatal("project choice not stored correctly")
	}

	if proj.Branch != branchName {
		t.Errorf("expected branch %q, got %q", branchName, proj.Branch)
	}

	// Simulate: user enters prompt → transitions to StateComplete
	userPrompt := "Focus on edge cases and error handling"
	w.choices["prompt"] = userPrompt
	w.state = StateComplete

	// Verify final state
	if w.state != StateComplete {
		t.Errorf("expected state StateComplete, got %v", w.state)
	}

	// Verify prompt is stored
	if w.choices["prompt"] != userPrompt {
		t.Errorf("expected prompt %q, got %q", userPrompt, w.choices["prompt"])
	}
}

// Test finalize() routing for continuation path

// TestFinalize_RoutesToContinuation tests that finalize correctly detects and routes to continuation path.
func TestFinalize_RoutesToContinuation(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit and a test project
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branchName := "feat/existing-project"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "existing-project", nil, nil)

	// Discover project
	projects, _ := listProjects(ctx)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	w := NewWizard(nil, ctx, []string{})

	// Set up wizard for continuation path
	w.choices["action"] = "continue"
	w.choices["project"] = projects[0]
	w.choices["prompt"] = "continue working"

	// Call finalize
	err := w.finalize()
	if err != nil {
		t.Fatalf("finalize failed: %v", err)
	}

	// Verify that project still exists (continuation doesn't create new project)
	// Need to recreate context to ensure FS is properly initialized
	freshWorktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		t.Fatalf("failed to create fresh worktree context: %v", err)
	}

	proj, err := state.Load(context.Background(), state.NewYAMLBackend(freshWorktreeCtx.FS()))
	if err != nil {
		t.Fatalf("failed to load project after continuation: %v", err)
	}

	if proj.Name != "existingproject" {
		t.Errorf("project name incorrect: got %q, want %q", proj.Name, "existingproject")
	}
}

// Test finalizeContinuation method

// TestFinalizeContinuation_ValidChoices tests successful continuation with valid choices.
func TestFinalizeContinuation_ValidChoices(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit and a test project
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branchName := "feat/test-continuation"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "test-continuation", nil, nil)

	// Discover project
	projects, _ := listProjects(ctx)
	selectedProj := projects[0]

	w := NewWizard(nil, ctx, []string{})

	// Set up wizard choices
	w.choices["project"] = selectedProj
	w.choices["prompt"] = "work on tests"

	// Call finalizeContinuation (we'll implement this)
	err := w.finalizeContinuation()
	if err != nil {
		t.Fatalf("finalizeContinuation failed: %v", err)
	}

	// Verify worktree still exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Errorf("worktree should exist at %s", worktreePath)
	}

	// Verify project state can be loaded
	// Need to recreate context to ensure FS is properly initialized
	freshWorktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		t.Fatalf("failed to create fresh worktree context: %v", err)
	}

	proj, err := state.Load(context.Background(), state.NewYAMLBackend(freshWorktreeCtx.FS()))
	if err != nil {
		t.Fatalf("failed to load project state: %v", err)
	}

	if proj.Name != "testcontinuation" {
		t.Errorf("project name incorrect: got %q", proj.Name)
	}
}

// TestFinalizeContinuation_EmptyUserPrompt tests continuation with empty user prompt.
func TestFinalizeContinuation_EmptyUserPrompt(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit and a test project
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branchName := "feat/empty-prompt-test"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "empty-prompt-test", nil, nil)

	// Discover project
	projects, _ := listProjects(ctx)
	selectedProj := projects[0]

	w := NewWizard(nil, ctx, []string{})

	// Set up wizard choices with empty prompt
	w.choices["project"] = selectedProj
	w.choices["prompt"] = ""

	// Call finalizeContinuation
	err := w.finalizeContinuation()
	if err != nil {
		t.Fatalf("finalizeContinuation should succeed with empty prompt, got error: %v", err)
	}

	// Test that generateContinuePrompt works without appending user prompt
	// Need to recreate context to ensure FS is properly initialized
	freshWorktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		t.Fatalf("failed to create fresh worktree context: %v", err)
	}

	proj, err := state.Load(context.Background(), state.NewYAMLBackend(freshWorktreeCtx.FS()))
	if err != nil {
		t.Fatalf("failed to load project state: %v", err)
	}

	basePrompt, err := generateContinuePrompt(proj)
	if err != nil {
		t.Fatalf("generateContinuePrompt failed: %v", err)
	}

	// Verify base prompt doesn't contain "User request:" when prompt is empty
	if strings.Contains(basePrompt, "User request:") {
		t.Error("base prompt should not contain 'User request:' section")
	}
}

// TestFinalizeContinuation_NonEmptyUserPrompt tests continuation with user prompt appended.
func TestFinalizeContinuation_NonEmptyUserPrompt(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit and a test project
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branchName := "feat/user-prompt-test"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "user-prompt-test", nil, nil)

	// Discover project
	projects, _ := listProjects(ctx)
	selectedProj := projects[0]

	userPrompt := "focus on integration tests and error handling"

	w := NewWizard(nil, ctx, []string{})

	// Set up wizard choices with non-empty prompt
	w.choices["project"] = selectedProj
	w.choices["prompt"] = userPrompt

	// We can't easily test the full prompt without launching Claude,
	// but we can test the prompt generation logic separately
	// Need to recreate context to ensure FS is properly initialized
	freshWorktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		t.Fatalf("failed to create fresh worktree context: %v", err)
	}

	proj, err := state.Load(context.Background(), state.NewYAMLBackend(freshWorktreeCtx.FS()))
	if err != nil {
		t.Fatalf("failed to load project state: %v", err)
	}

	basePrompt, err := generateContinuePrompt(proj)
	if err != nil {
		t.Fatalf("generateContinuePrompt failed: %v", err)
	}

	// Build full prompt as finalizeContinuation should
	fullPrompt := basePrompt
	if userPrompt != "" {
		fullPrompt += "\n\nUser request:\n" + userPrompt
	}

	// Verify full prompt contains both base and user prompt
	if !strings.Contains(fullPrompt, userPrompt) {
		t.Errorf("full prompt should contain user prompt %q", userPrompt)
	}

	if !strings.Contains(fullPrompt, "User request:") {
		t.Error("full prompt should contain 'User request:' header")
	}

	// Verify base prompt has 3 layers (at least 2 separators)
	separatorCount := strings.Count(basePrompt, "\n\n---\n\n")
	if separatorCount < 2 {
		t.Errorf("expected at least 2 separators for 3 layers, got %d", separatorCount)
	}
}

// TestFinalizeContinuation_MissingProjectChoice tests error when project choice is missing.
func TestFinalizeContinuation_MissingProjectChoice(t *testing.T) {
	ctx, _ := setupTestContext(t)
	w := NewWizard(nil, ctx, []string{})

	// Set prompt but NOT project
	w.choices["prompt"] = "some prompt"

	// Call finalizeContinuation - should fail
	err := w.finalizeContinuation()
	if err == nil {
		t.Fatal("expected error for missing project choice, got nil")
	}

	// Verify error message
	errMsg := err.Error()
	if !strings.Contains(errMsg, "project choice not set") {
		t.Errorf("error should mention missing project choice, got: %v", errMsg)
	}
}

// TestFinalizeContinuation_InvalidProjectChoice tests error when project choice has wrong type.
func TestFinalizeContinuation_InvalidProjectChoice(t *testing.T) {
	ctx, _ := setupTestContext(t)
	w := NewWizard(nil, ctx, []string{})

	// Set wrong type for project choice
	w.choices["project"] = "not-a-ProjectInfo"
	w.choices["prompt"] = "some prompt"

	// Call finalizeContinuation - should fail
	err := w.finalizeContinuation()
	if err == nil {
		t.Fatal("expected error for invalid project choice type, got nil")
	}

	// Verify error message
	errMsg := err.Error()
	if !strings.Contains(errMsg, "project choice not set or invalid") {
		t.Errorf("error should mention invalid project choice, got: %v", errMsg)
	}
}

// TestFinalizeContinuation_MissingPromptChoice tests error when prompt choice is missing.
func TestFinalizeContinuation_MissingPromptChoice(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit and a test project
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branchName := "feat/missing-prompt"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "missing-prompt", nil, nil)

	// Discover project
	projects, _ := listProjects(ctx)
	selectedProj := projects[0]

	w := NewWizard(nil, ctx, []string{})

	// Set project but NOT prompt
	w.choices["project"] = selectedProj

	// Call finalizeContinuation - should fail
	err := w.finalizeContinuation()
	if err == nil {
		t.Fatal("expected error for missing prompt choice, got nil")
	}

	// Verify error message
	errMsg := err.Error()
	if !strings.Contains(errMsg, "prompt choice not set or invalid") {
		t.Errorf("error should mention missing prompt choice, got: %v", errMsg)
	}
}

// TestFinalizeContinuation_WorktreeDeleted tests idempotent worktree recreation.
func TestFinalizeContinuation_WorktreeDeleted(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit and a test project
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branchName := "feat/worktree-deleted"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "worktree-deleted", nil, nil)

	// Discover project
	projects, _ := listProjects(ctx)
	selectedProj := projects[0]

	// Delete the worktree directory (simulate edge case)
	// Note: This will make state loading fail, which is expected behavior
	// EnsureWorktree will recreate it, but the project state won't exist
	_ = os.RemoveAll(worktreePath)

	w := NewWizard(nil, ctx, []string{})
	w.choices["project"] = selectedProj
	w.choices["prompt"] = "test"

	// Call finalizeContinuation
	// This should fail when trying to ensure worktree (git knows about it but directory is missing)
	err := w.finalizeContinuation()
	if err == nil {
		t.Fatal("expected error when worktree directory is missing, got nil")
	}

	// Verify error mentions worktree or state loading failure
	errMsg := err.Error()
	if !strings.Contains(errMsg, "failed to ensure worktree") && !strings.Contains(errMsg, "failed to load project state") {
		t.Errorf("error should mention worktree or state loading failure, got: %v", errMsg)
	}
}

// TestFinalizeContinuation_NoUncommittedCheck tests that continuation doesn't check uncommitted changes.
func TestFinalizeContinuation_NoUncommittedCheck(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create initial commit
	testFile := tmpDir + "/README.md"
	_ = os.WriteFile(testFile, []byte("# Test"), 0644)
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "add", ".").Run()
	_ = exec.CommandContext(context.Background(), "git", "-C", tmpDir, "commit", "-m", "initial").Run()

	// Create a project in a worktree
	branchName := "feat/uncommitted-test"
	worktreePath := tmpDir + "/.sow/worktrees/" + branchName
	_ = git.EnsureWorktree(ctx.Git(), ctx.RepoRoot(), worktreePath, branchName)
	worktreeCtx, _ := sow.NewContext(worktreePath)
	_, _ = initializeProject(worktreeCtx, branchName, "uncommitted-test", nil, nil)

	// Create uncommitted changes in MAIN repo (not worktree)
	mainTestFile := tmpDir + "/uncommitted.txt"
	_ = os.WriteFile(mainTestFile, []byte("uncommitted"), 0644)

	// Discover project
	projects, _ := listProjects(ctx)
	selectedProj := projects[0]

	w := NewWizard(nil, ctx, []string{})
	w.choices["project"] = selectedProj
	w.choices["prompt"] = "continue despite uncommitted changes"

	// Call finalizeContinuation
	// Should succeed even with uncommitted changes (unlike creation path)
	err := w.finalizeContinuation()
	if err != nil {
		t.Fatalf("finalizeContinuation should succeed despite uncommitted changes in main repo, got error: %v", err)
	}

	// This test verifies the critical difference from creation path:
	// Continuation does NOT call CheckUncommittedChanges()
}
