package project

import (
	"fmt"
	"os"
	osExec "os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sdks/project/templates"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/libs/exec"
	"github.com/jmgilman/sow/libs/git"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
)

// initializeProject creates a new project in the given context.
// It creates the project directory structure, optionally writes issue context,
// registers knowledge file artifacts, and initializes project state.
//
// Parameters:
//   - ctx: The sow context (should be a worktree context)
//   - branch: The branch name for the project
//   - description: The project description
//   - issue: Optional GitHub issue to link (can be nil)
//   - knowledgeFiles: Optional list of knowledge file paths to register as artifacts (can be nil or empty)
//
// Returns the created project or an error.
func initializeProject(
	ctx *sow.Context,
	branch string,
	description string,
	issue *git.Issue,
	knowledgeFiles []string,
) (*state.Project, error) {
	// Get the worktree root path
	worktreePath := ctx.RepoRoot()

	// Ensure .sow/project directory exists in worktree
	projectDir := filepath.Join(worktreePath, ".sow", "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	// Always create context directory for project files
	contextDir := filepath.Join(projectDir, "context")
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", err)
	}

	// Recreate context now that .sow directory exists
	// This ensures ctx.FS() is properly initialized before calling state.Create
	ctx, err := sow.NewContext(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to recreate context after directory creation: %w", err)
	}

	// Prepare initial inputs if issue is provided
	var initialInputs map[string][]projschema.ArtifactState
	if issue != nil {
		// Write issue body to file
		issueFileName := fmt.Sprintf("issue-%d.md", issue.Number)
		issuePath := filepath.Join(contextDir, issueFileName)
		issueContent := fmt.Sprintf("# Issue #%d: %s\n\n**URL**: %s\n**State**: %s\n\n## Description\n\n%s\n",
			issue.Number, issue.Title, issue.URL, issue.State, issue.Body)

		if err := os.WriteFile(issuePath, []byte(issueContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write issue file: %w", err)
		}

		// Create github_issue artifact for implementation phase
		issueArtifact := projschema.ArtifactState{
			Type:       "github_issue",
			Path:       fmt.Sprintf("context/%s", issueFileName),
			Approved:   true, // Auto-approved
			Created_at: time.Now(),
			Metadata: map[string]interface{}{
				"issue_number": issue.Number,
				"issue_url":    issue.URL,
				"issue_title":  issue.Title,
			},
		}

		initialInputs = map[string][]projschema.ArtifactState{
			"implementation": {issueArtifact},
		}
	}

	// Add knowledge file inputs if provided
	if len(knowledgeFiles) > 0 {
		// Initialize initialInputs if not already done
		if initialInputs == nil {
			initialInputs = make(map[string][]projschema.ArtifactState)
		}

		// Create artifacts for each knowledge file
		knowledgeArtifacts := make([]projschema.ArtifactState, 0, len(knowledgeFiles))
		for _, file := range knowledgeFiles {
			artifact := projschema.ArtifactState{
				Type:       "reference",
				Path:       filepath.Join("../../knowledge", file),
				Approved:   true, // Auto-approved
				Created_at: time.Now(),
				Metadata: map[string]interface{}{
					"source":      "user_selected",
					"description": "Knowledge file selected during project creation",
				},
			}
			knowledgeArtifacts = append(knowledgeArtifacts, artifact)
		}

		// Determine target phase for knowledge files
		targetPhase := determineKnowledgeInputPhase("standard")

		// Add knowledge artifacts to target phase
		if _, exists := initialInputs[targetPhase]; !exists {
			initialInputs[targetPhase] = []projschema.ArtifactState{}
		}
		initialInputs[targetPhase] = append(initialInputs[targetPhase], knowledgeArtifacts...)
	}

	// Create project using SDK with initial inputs
	proj, err := state.Create(ctx, branch, description, initialInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return proj, nil
}

// determineKnowledgeInputPhase determines which phase should receive knowledge file inputs.
// For now, all project types use "implementation" as the default first phase.
// This can be enhanced in the future to support different phases based on project type.
func determineKnowledgeInputPhase(_ string) string {
	// For now, use "implementation" as default (first phase for standard projects)
	// This could be enhanced to detect project type and use appropriate phase
	return "implementation"
}

// generateNewProjectPrompt creates the custom prompt for new projects.
// Uses 3-layer structure: Base Orchestrator + Project Type Orchestrator + Initial State.
//
// Parameters:
//   - proj: The project state
//   - initialPrompt: Optional user's initial request (can be empty)
//
// Returns the combined prompt or an error.
func generateNewProjectPrompt(proj *state.Project, initialPrompt string) (string, error) {
	var buf strings.Builder

	// Layer 1: Base Orchestrator Introduction
	baseOrch, err := templates.Render(prompts.FS, "templates/greet/orchestrator.md", nil)
	if err != nil {
		return "", fmt.Errorf("failed to render base orchestrator prompt: %w", err)
	}
	buf.WriteString(baseOrch)
	buf.WriteString("\n\n---\n\n")

	// Layer 2: Project Type Orchestrator Prompt
	projectTypePrompt := proj.Config().OrchestratorPrompt(proj)
	if projectTypePrompt != "" {
		buf.WriteString(projectTypePrompt)
		buf.WriteString("\n\n---\n\n")
	}

	// Layer 3: Initial State Prompt
	initialState := proj.Machine().State()
	statePrompt := proj.Config().GetStatePrompt(initialState, proj)
	if statePrompt != "" {
		buf.WriteString(statePrompt)
		buf.WriteString("\n\n---\n\n")
	}

	// Add initial user prompt if provided
	if initialPrompt != "" {
		buf.WriteString("## User's Initial Request\n\n")
		buf.WriteString(initialPrompt)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// generateContinuePrompt creates the custom prompt for continuing projects.
// Uses 3-layer structure: Base Orchestrator + Project Type Orchestrator + Current State.
//
// Parameters:
//   - proj: The project state
//
// Returns the combined prompt or an error.
func generateContinuePrompt(proj *state.Project) (string, error) {
	var buf strings.Builder

	// Layer 1: Base Orchestrator Introduction
	baseOrch, err := templates.Render(prompts.FS, "templates/greet/orchestrator.md", nil)
	if err != nil {
		return "", fmt.Errorf("failed to render base orchestrator prompt: %w", err)
	}
	buf.WriteString(baseOrch)
	buf.WriteString("\n\n---\n\n")

	// Layer 2: Project Type Orchestrator Prompt
	projectTypePrompt := proj.Config().OrchestratorPrompt(proj)
	if projectTypePrompt != "" {
		buf.WriteString(projectTypePrompt)
		buf.WriteString("\n\n---\n\n")
	}

	// Layer 3: Current State Prompt
	currentState := proj.Machine().State()
	statePrompt := proj.Config().GetStatePrompt(currentState, proj)
	if statePrompt != "" {
		buf.WriteString(statePrompt)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// launchClaudeCode executes the Claude Code CLI with the given prompt and flags.
//
// Parameters:
//   - cmd: The cobra command (for context)
//   - ctx: The sow context
//   - prompt: The prompt to pass to Claude
//   - claudeFlags: Additional flags to pass to Claude CLI
//
// Returns an error if Claude is not found or execution fails.
func launchClaudeCode(
	cmd *cobra.Command,
	ctx *sow.Context,
	prompt string,
	claudeFlags []string,
) error {
	claude := exec.NewLocalExecutor("claude")
	if !claude.Exists() {
		fmt.Fprintln(os.Stderr, "Error: Claude Code CLI not found")
		fmt.Fprintln(os.Stderr, "Install from: https://claude.com/download")
		return fmt.Errorf("claude not found")
	}

	// Build command args: prompt first, then any additional flags
	args := []string{prompt}
	args = append(args, claudeFlags...)

	claudeCmd := osExec.CommandContext(cmd.Context(), claude.Command(), args...)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr
	claudeCmd.Dir = ctx.RepoRoot()

	return claudeCmd.Run()
}
