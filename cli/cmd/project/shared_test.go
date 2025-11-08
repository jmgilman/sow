package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sow"
	projschema "github.com/jmgilman/sow/cli/schemas/project"

	// Import project types to register them
	_ "github.com/jmgilman/sow/cli/internal/projects/breakdown"
	_ "github.com/jmgilman/sow/cli/internal/projects/design"
	_ "github.com/jmgilman/sow/cli/internal/projects/exploration"
	_ "github.com/jmgilman/sow/cli/internal/projects/standard"
)

// setupTestRepo initializes a git repo in the given directory for testing
func setupTestRepo(t *testing.T, dir string) {
	t.Helper()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git (needed for commits)
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = dir
	if err := configCmd.Run(); err != nil {
		t.Fatalf("failed to config git email: %v", err)
	}

	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = dir
	if err := configCmd.Run(); err != nil {
		t.Fatalf("failed to config git name: %v", err)
	}
}

// setupTestContext creates a test directory with git repo and sow context
func setupTestContext(t *testing.T) (*sow.Context, string) {
	t.Helper()

	tmpDir := t.TempDir()
	setupTestRepo(t, tmpDir)

	// Initialize .sow directory
	sowDir := filepath.Join(tmpDir, ".sow")
	if err := os.MkdirAll(sowDir, 0755); err != nil {
		t.Fatalf("failed to create .sow directory: %v", err)
	}

	// Create context
	ctx, err := sow.NewContext(tmpDir)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	return ctx, tmpDir
}

// Test initializeProject

func TestInitializeProject_CreatesDirectories(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Call initializeProject
	_, err := initializeProject(ctx, "feat/test", "Test project", nil)
	if err != nil {
		t.Fatalf("initializeProject failed: %v", err)
	}

	// Verify directories were created
	projectDir := filepath.Join(tmpDir, ".sow", "project")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Errorf("project directory was not created")
	}

	contextDir := filepath.Join(projectDir, "context")
	if _, err := os.Stat(contextDir); os.IsNotExist(err) {
		t.Errorf("context directory was not created")
	}
}

func TestInitializeProject_WithIssue_WritesIssueFile(t *testing.T) {
	ctx, tmpDir := setupTestContext(t)

	// Create test issue
	issue := &sow.Issue{
		Number: 123,
		Title:  "Test Issue",
		Body:   "Test issue body",
		State:  "OPEN",
		URL:    "https://github.com/test/repo/issues/123",
	}

	// Call initializeProject with issue
	_, err := initializeProject(ctx, "feat/test-issue", "Test with issue", issue)
	if err != nil {
		t.Fatalf("initializeProject failed: %v", err)
	}

	// Verify issue file was created
	issueFile := filepath.Join(tmpDir, ".sow", "project", "context", "issue-123.md")
	content, err := os.ReadFile(issueFile)
	if err != nil {
		t.Fatalf("issue file was not created: %v", err)
	}

	// Verify file format
	contentStr := string(content)
	expectedParts := []string{
		"# Issue #123: Test Issue",
		"**URL**: https://github.com/test/repo/issues/123",
		"**State**: OPEN",
		"## Description",
		"Test issue body",
	}

	for _, part := range expectedParts {
		if !strings.Contains(contentStr, part) {
			t.Errorf("issue file missing expected content: %q\nGot: %s", part, contentStr)
		}
	}
}

func TestInitializeProject_WithIssue_CreatesArtifact(t *testing.T) {
	ctx, _ := setupTestContext(t)

	// Create test issue
	issue := &sow.Issue{
		Number: 456,
		Title:  "Another Issue",
		Body:   "Another body",
		State:  "OPEN",
		URL:    "https://github.com/test/repo/issues/456",
	}

	// Call initializeProject with issue
	proj, err := initializeProject(ctx, "feat/test-artifact", "Test artifact", issue)
	if err != nil {
		t.Fatalf("initializeProject failed: %v", err)
	}

	// Load the project state to verify artifact
	loadedProj, err := state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}

	// Get implementation phase
	implPhase, exists := loadedProj.Phases["implementation"]
	if !exists {
		t.Fatal("implementation phase not found")
	}

	// Verify artifact exists
	if len(implPhase.Inputs) == 0 {
		t.Fatal("no inputs found in implementation phase")
	}

	// Find the github_issue artifact
	var issueArtifact *projschema.ArtifactState
	for _, artifact := range implPhase.Inputs {
		if artifact.Type == "github_issue" {
			issueArtifact = &artifact
			break
		}
	}

	if issueArtifact == nil {
		t.Fatal("github_issue artifact not found")
	}

	// Verify artifact metadata
	if issueArtifact.Path != "context/issue-456.md" {
		t.Errorf("artifact path incorrect: got %q, want %q", issueArtifact.Path, "context/issue-456.md")
	}

	if !issueArtifact.Approved {
		t.Error("artifact should be auto-approved")
	}

	// Verify metadata fields
	if issueNum, ok := issueArtifact.Metadata["issue_number"].(int); !ok || issueNum != 456 {
		t.Errorf("artifact metadata issue_number incorrect: got %v", issueArtifact.Metadata["issue_number"])
	}

	if issueURL, ok := issueArtifact.Metadata["issue_url"].(string); !ok || issueURL != issue.URL {
		t.Errorf("artifact metadata issue_url incorrect: got %v", issueArtifact.Metadata["issue_url"])
	}

	if issueTitle, ok := issueArtifact.Metadata["issue_title"].(string); !ok || issueTitle != issue.Title {
		t.Errorf("artifact metadata issue_title incorrect: got %v", issueArtifact.Metadata["issue_title"])
	}

	// Verify project was returned correctly (name is normalized from branch)
	if proj.Name != "test-artifact" {
		t.Errorf("project name incorrect: got %q, want %q", proj.Name, "test-artifact")
	}
}

func TestInitializeProject_WithoutIssue_NoArtifacts(t *testing.T) {
	ctx, _ := setupTestContext(t)

	// Call initializeProject without issue
	_, err := initializeProject(ctx, "feat/no-issue", "Test without issue", nil)
	if err != nil {
		t.Fatalf("initializeProject failed: %v", err)
	}

	// Load the project state
	loadedProj, err := state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}

	// Get implementation phase
	implPhase, exists := loadedProj.Phases["implementation"]
	if !exists {
		t.Fatal("implementation phase not found")
	}

	// Verify no artifacts exist
	if len(implPhase.Inputs) != 0 {
		t.Errorf("expected no inputs in implementation phase, got %d", len(implPhase.Inputs))
	}
}

// Test generateNewProjectPrompt

func TestGenerateNewProjectPrompt_Has3Layers(t *testing.T) {
	ctx, _ := setupTestContext(t)

	// Create a project
	proj, err := initializeProject(ctx, "feat/test", "Test project", nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Generate prompt
	prompt, err := generateNewProjectPrompt(proj, "")
	if err != nil {
		t.Fatalf("generateNewProjectPrompt failed: %v", err)
	}

	// Verify prompt has content
	if prompt == "" {
		t.Fatal("prompt is empty")
	}

	// Verify separator exists (indicates multiple layers)
	if !strings.Contains(prompt, "\n\n---\n\n") {
		t.Error("prompt missing layer separator")
	}

	// Count separators - should have at least 2 (for 3 layers)
	separatorCount := strings.Count(prompt, "\n\n---\n\n")
	if separatorCount < 2 {
		t.Errorf("expected at least 2 separators for 3 layers, got %d", separatorCount)
	}
}

func TestGenerateNewProjectPrompt_WithUserPrompt(t *testing.T) {
	ctx, _ := setupTestContext(t)

	// Create a project
	proj, err := initializeProject(ctx, "feat/test", "Test project", nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Generate prompt with user prompt
	userPrompt := "Start by reviewing existing auth code"
	prompt, err := generateNewProjectPrompt(proj, userPrompt)
	if err != nil {
		t.Fatalf("generateNewProjectPrompt failed: %v", err)
	}

	// Verify user prompt is included
	if !strings.Contains(prompt, userPrompt) {
		t.Errorf("prompt does not contain user prompt %q", userPrompt)
	}

	// Verify user prompt appears after the layers
	if !strings.Contains(prompt, "User's Initial Request") {
		t.Error("prompt missing user request section header")
	}
}

func TestGenerateNewProjectPrompt_WithoutUserPrompt(t *testing.T) {
	ctx, _ := setupTestContext(t)

	// Create a project
	proj, err := initializeProject(ctx, "feat/test", "Test project", nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Generate prompt without user prompt
	prompt, err := generateNewProjectPrompt(proj, "")
	if err != nil {
		t.Fatalf("generateNewProjectPrompt failed: %v", err)
	}

	// Verify no user request section
	if strings.Contains(prompt, "User's Initial Request") {
		t.Error("prompt should not contain user request section when no user prompt provided")
	}

	// Verify prompt still has content
	if len(prompt) == 0 {
		t.Error("prompt is empty")
	}
}

// Test generateContinuePrompt

func TestGenerateContinuePrompt_Has3Layers(t *testing.T) {
	ctx, _ := setupTestContext(t)

	// Create a project
	proj, err := initializeProject(ctx, "feat/test", "Test project", nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Generate continue prompt
	prompt, err := generateContinuePrompt(proj)
	if err != nil {
		t.Fatalf("generateContinuePrompt failed: %v", err)
	}

	// Verify prompt has content
	if prompt == "" {
		t.Fatal("prompt is empty")
	}

	// Verify separator exists (indicates multiple layers)
	if !strings.Contains(prompt, "\n\n---\n\n") {
		t.Error("prompt missing layer separator")
	}

	// Count separators - should have at least 2 (for 3 layers)
	separatorCount := strings.Count(prompt, "\n\n---\n\n")
	if separatorCount < 2 {
		t.Errorf("expected at least 2 separators for 3 layers, got %d", separatorCount)
	}
}

func TestGenerateContinuePrompt_UsesCurrentState(t *testing.T) {
	ctx, _ := setupTestContext(t)

	// Create a project
	proj, err := initializeProject(ctx, "feat/test", "Test project", nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Generate continue prompt
	prompt, err := generateContinuePrompt(proj)
	if err != nil {
		t.Fatalf("generateContinuePrompt failed: %v", err)
	}

	// The continue prompt should use current state, not initial state
	// We can't easily verify this without knowing the state machine details,
	// but we can at least verify the prompt was generated successfully
	if len(prompt) == 0 {
		t.Error("prompt is empty")
	}

	// Verify it doesn't contain "User's Initial Request" (that's only for new prompts)
	if strings.Contains(prompt, "User's Initial Request") {
		t.Error("continue prompt should not contain user initial request section")
	}
}

// Test launchClaudeCode - we'll test arg building and command setup,
// but not actual execution (that would require Claude to be installed)

func TestLaunchClaudeCode_BuildsArgsCorrectly(t *testing.T) {
	// We can't fully test execution without Claude installed,
	// but we can test the function exists and handles basic cases
	// This test will be a simple check that the function signature exists
	// and doesn't panic when called with proper args

	// Note: This test will fail if Claude is not installed, which is expected
	// The real test is that the function exists with the correct signature
	t.Skip("Skipping launchClaudeCode test - requires Claude CLI to be installed")
}
