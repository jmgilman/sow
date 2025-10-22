package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewNewCmd creates the new command.
func NewNewCmd() *cobra.Command {
	var issueNumber int

	cmd := &cobra.Command{
		Use:   "new [prompt]",
		Short: "Create a new project and launch orchestrator",
		Long: `Create a new sow project and launch Claude Code with project context.

This command initializes a new project structure and launches the orchestrator
with a custom prompt that provides project context and next steps.

The command can create projects from GitHub issues or from a custom prompt:

From GitHub issue (--issue):
  - Fetches the issue details
  - Validates the issue has the 'sow' label
  - Creates and checks out a linked branch
  - Initializes project with issue title as description

From custom prompt (positional argument):
  - Uses the prompt as the initial project context
  - Initializes project on current branch

Examples:
  sow new "Create a login endpoint on the API"
  sow new --issue 123
  sow new --issue 123 "Add JWT authentication"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNew(cmd, args, issueNumber)
		},
	}

	cmd.Flags().IntVar(&issueNumber, "issue", 0, "GitHub issue number to create project from")

	return cmd
}

func runNew(cmd *cobra.Command, args []string, issueNumber int) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Require sow to be initialized
	if !ctx.IsInitialized() {
		fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
		fmt.Fprintln(os.Stderr, "Run: sow init")
		return fmt.Errorf("not initialized")
	}

	// Check not on protected branch
	currentBranch, err := ctx.Git().CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if ctx.Git().IsProtectedBranch(currentBranch) && issueNumber == 0 {
		return fmt.Errorf("cannot create project on protected branch '%s' - create a feature branch first", currentBranch)
	}

	// Check no existing project
	if projectpkg.Exists(ctx) {
		return fmt.Errorf("project already exists in this branch - use 'sow continue' to resume")
	}

	// Parse initial prompt (optional positional argument)
	var initialPrompt string
	if len(args) > 0 {
		initialPrompt = args[0]
	}

	// Handle GitHub issue flow
	var issue *sow.Issue
	var branchName string
	if issueNumber > 0 {
		issue, branchName, err = handleIssueFlow(ctx, issueNumber)
		if err != nil {
			return err
		}

		// Update current branch after checkout
		currentBranch = branchName
	}

	// Initialize project structure
	if err := initializeProject(ctx, issue, initialPrompt); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	// Generate new project prompt
	newPrompt, err := generateNewProjectPrompt(ctx, issue, initialPrompt, currentBranch)
	if err != nil {
		return fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Launch Claude Code with the prompt
	return launchClaudeCode(cmd, ctx, newPrompt)
}

// handleIssueFlow handles the GitHub issue workflow:
// - Fetch issue
// - Validate has 'sow' label
// - Check not already claimed
// - Create and checkout linked branch.
func handleIssueFlow(_ *sow.Context, issueNumber int) (*sow.Issue, string, error) {
	ghExec := sowexec.NewLocal("gh")
	gh := sow.NewGitHub(ghExec)

	// Fetch issue
	issue, err := gh.GetIssue(issueNumber)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch issue #%d: %w", issueNumber, err)
	}

	// Validate has 'sow' label
	if !issue.HasLabel("sow") {
		return nil, "", fmt.Errorf("issue #%d does not have the 'sow' label - add it with: gh issue edit %d --add-label sow", issueNumber, issueNumber)
	}

	// Check not already claimed
	branches, err := gh.GetLinkedBranches(issueNumber)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check linked branches: %w", err)
	}

	if len(branches) > 0 {
		fmt.Fprintf(os.Stderr, "Error: issue #%d already has linked branch(es):\n", issueNumber)
		for _, b := range branches {
			fmt.Fprintf(os.Stderr, "  - %s (%s)\n", b.Name, b.URL)
		}
		return nil, "", fmt.Errorf("issue already claimed")
	}

	// Create and checkout linked branch
	branchName, err := gh.CreateLinkedBranch(issueNumber, "", true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create linked branch: %w", err)
	}

	return issue, branchName, nil
}

// initializeProject creates the project structure via statechart.
func initializeProject(ctx *sow.Context, issue *sow.Issue, initialPrompt string) error {
	// Determine project name and description
	var projectName, projectDescription string
	var githubIssueNum *int

	if issue != nil {
		// Use issue title as description, generate name from it
		projectDescription = issue.Title
		projectName = generateProjectName(issue.Title)
		githubIssueNum = &issue.Number
	} else {
		// Use initial prompt or generate generic name
		if initialPrompt != "" {
			projectDescription = initialPrompt
			projectName = generateProjectName(initialPrompt)
		} else {
			projectDescription = "New project"
			projectName = "new-project"
		}
	}

	// Create project using the internal project package
	project, err := projectpkg.Create(ctx, projectName, projectDescription)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Set github_issue field if provided
	if githubIssueNum != nil {
		state := project.State()
		state.Project.Github_issue = *githubIssueNum

		// Save the updated state via the machine
		if err := project.Machine().Save(); err != nil {
			return fmt.Errorf("failed to save github_issue link: %w", err)
		}
	}

	return nil
}

// generateNewProjectPrompt creates the custom prompt for new projects.
// Composes base greeting + new project context.
func generateNewProjectPrompt(ctx *sow.Context, issue *sow.Issue, initialPrompt, branchName string) (string, error) {
	// Render base greeting (orchestrator introduction)
	// Use empty GreetContext since we don't need the state-specific sections
	baseCtx := &prompts.GreetContext{
		SowInitialized: ctx.IsInitialized(),
		HasProject:     false, // Don't include project info in base
	}

	base, err := prompts.Render(prompts.PromptGreetBase, baseCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render base greeting: %w", err)
	}

	// Render new project context
	newProjectCtx := &prompts.NewProjectContext{
		RepoRoot:        ctx.RepoRoot(),
		BranchName:      branchName,
		InitialPrompt:   initialPrompt,
		StatechartState: "DiscoveryDecision",
	}

	if issue != nil {
		newProjectCtx.IssueNumber = &issue.Number
		newProjectCtx.IssueTitle = issue.Title
		newProjectCtx.IssueBody = issue.Body
	}

	newSection, err := prompts.Render(prompts.PromptCommandNew, newProjectCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render new project section: %w", err)
	}

	// Compose: base + new project context
	return base + "\n\n" + newSection, nil
}

// launchClaudeCode launches Claude Code with the given prompt.
func launchClaudeCode(cmd *cobra.Command, ctx *sow.Context, prompt string) error {
	claude := sowexec.NewLocal("claude")
	if !claude.Exists() {
		fmt.Fprintln(os.Stderr, "Error: Claude Code CLI not found")
		fmt.Fprintln(os.Stderr, "Install from: https://claude.com/download")
		return fmt.Errorf("claude not found")
	}

	claudeCmd := exec.CommandContext(cmd.Context(), claude.Command(), prompt)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr
	claudeCmd.Dir = ctx.RepoRoot()

	return claudeCmd.Run()
}

// generateProjectName converts a description to a kebab-case project name.
func generateProjectName(description string) string {
	// Reuse the toKebabCase logic from github.go
	// For now, use a simplified version
	name := description
	if len(name) > 50 {
		name = name[:50]
	}

	// Convert to kebab-case
	return toKebabCase(name)
}

// toKebabCase converts a string to kebab-case.
// (Duplicated from github.go for simplicity - could be extracted to a shared util package).
func toKebabCase(s string) string {
	// Simple implementation - just lowercase and replace spaces
	result := ""
	for i, r := range s {
		if r == ' ' || r == '_' {
			if i > 0 && result[len(result)-1] != '-' {
				result += "-"
			}
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else if r >= 'A' && r <= 'Z' {
			result += string(r + 32) // Convert to lowercase
		}
	}

	// Remove trailing hyphen
	if len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}

	return result
}
