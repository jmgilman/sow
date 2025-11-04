// Package project provides commands for managing project lifecycle.
package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/project"
	"github.com/spf13/cobra"
)

func newContinueCmd() *cobra.Command {
	var branchName string
	var noLaunch bool

	cmd := &cobra.Command{
		Use:   "continue",
		Short: "Continue an existing project",
		Long: `Continue work on an existing project.

Loads the project state, generates a continue prompt, and launches Claude Code.

Examples:
  sow project continue
  sow project continue --branch feat/auth
  sow project continue --no-launch

Claude Code Flags:
  Use -- to pass additional flags to the Claude Code CLI:
    sow project continue -- --model opus --verbose`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runContinue(cmd, branchName, noLaunch)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to continue project on")
	cmd.Flags().BoolVar(&noLaunch, "no-launch", false, "Don't launch Claude Code (for testing)")

	return cmd
}

func runContinue(cmd *cobra.Command, branchName string, noLaunch bool) error {
	// 1. Get main repo context
	mainCtx := cmdutil.GetContext(cmd.Context())

	// Require sow to be initialized
	if !mainCtx.IsInitialized() {
		fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
		fmt.Fprintln(os.Stderr, "Run: sow init")
		return fmt.Errorf("not initialized")
	}

	// Extract Claude Code flags (everything after --)
	var claudeFlags []string
	if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
		claudeFlags = cmd.Flags().Args()[dashIndex:]
	}

	// 2. Determine target branch
	var selectedBranch string
	var err error

	if branchName != "" {
		selectedBranch, err = handleBranchScenarioContinue(mainCtx, branchName)
		if err != nil {
			return err
		}
	} else {
		selectedBranch, err = handleCurrentBranchScenarioContinue(mainCtx)
		if err != nil {
			return err
		}
	}

	// 3. Generate worktree path
	worktreePath := sow.WorktreePath(mainCtx.RepoRoot(), selectedBranch)
	worktreesDir := filepath.Join(mainCtx.RepoRoot(), ".sow", "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// 4. Ensure worktree exists
	if err := sow.EnsureWorktree(mainCtx, worktreePath, selectedBranch); err != nil {
		return fmt.Errorf("failed to ensure worktree: %w", err)
	}

	// 5. Create worktree context
	worktreeCtx, err := sow.NewContext(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to create worktree context: %w", err)
	}

	// 6. Load existing project using SDK
	proj, err := state.Load(worktreeCtx)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Continuing project '%s' on branch %s\n", proj.Name, selectedBranch)

	// 7. Generate continue prompt
	prompt, err := generateContinuePrompt(worktreeCtx, proj)
	if err != nil {
		return fmt.Errorf("failed to generate continue prompt: %w", err)
	}

	// 8. Launch Claude Code from worktree directory (unless --no-launch)
	if noLaunch {
		return nil
	}
	return launchClaudeCode(cmd, worktreeCtx, prompt, claudeFlags)
}

// handleBranchScenarioContinue handles the --branch flag scenario for continue command.
// Returns: (branchName, error).
func handleBranchScenarioContinue(ctx *sow.Context, branchName string) (string, error) {
	git := ctx.Git()

	// Check if branch exists
	branches, err := git.Branches()
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %w", err)
	}

	branchExists := false
	for _, b := range branches {
		if b == branchName {
			branchExists = true
			break
		}
	}

	if !branchExists {
		return "", fmt.Errorf("branch %s does not exist - create it first with: sow project new --branch %s \"description\"", branchName, branchName)
	}

	// Checkout the branch
	if err := git.CheckoutBranch(branchName); err != nil {
		return "", fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
	}

	return branchName, nil
}

// handleCurrentBranchScenarioContinue handles the no-flags scenario (current branch) for continue command.
// Returns: (branchName, error).
func handleCurrentBranchScenarioContinue(ctx *sow.Context) (string, error) {
	git := ctx.Git()

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return currentBranch, nil
}

// generateContinuePrompt creates the custom prompt for continuing projects.
// Composes base greeting + continue project context.
func generateContinuePrompt(ctx *sow.Context, proj *state.Project) (string, error) {
	// Render base greeting (orchestrator introduction)
	baseCtx := &prompts.GreetContext{
		SowInitialized: ctx.IsInitialized(),
		HasProject:     false, // Don't include project info in base, we'll add it separately
	}

	base, err := prompts.Render(prompts.PromptGreetBase, baseCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render base greeting: %w", err)
	}

	// Build continue context
	promptCtx := &prompts.ContinueProjectContext{
		BranchName:         proj.Branch,
		ProjectName:        proj.Name,
		ProjectDescription: proj.Description,
		StatechartState:    proj.Statechart.Current_state,
	}

	// Add issue number if present in metadata
	// Note: github_issue linking not yet supported in new schema
	// Will be added in future when project-level metadata is supported

	// Add phase status
	if planning, ok := proj.Phases["planning"]; ok {
		promptCtx.PlanningStatus = planning.Status
	}
	if impl, ok := proj.Phases["implementation"]; ok {
		promptCtx.ImplementationStatus = impl.Status
	}
	if review, ok := proj.Phases["review"]; ok {
		promptCtx.ReviewStatus = review.Status
	}
	if finalize, ok := proj.Phases["finalize"]; ok {
		promptCtx.FinalizeStatus = finalize.Status
	}

	// Count tasks across all phases
	for _, phase := range proj.Phases {
		promptCtx.TasksTotal += len(phase.Tasks)
		for _, task := range phase.Tasks {
			switch task.Status {
			case "completed":
				promptCtx.TasksCompleted++
			case "in_progress":
				promptCtx.TasksInProgress++
				if promptCtx.CurrentTaskID == "" {
					promptCtx.CurrentTaskID = task.Id
					promptCtx.CurrentTaskName = task.Name
					promptCtx.CurrentTaskStatus = task.Status
				}
			case "pending":
				promptCtx.TasksPending++
			case "abandoned":
				promptCtx.TasksAbandoned++
			}
		}
	}

	// Generate state-specific guidance and next actions
	promptCtx.StateSpecificGuidance = generateStateGuidance(proj.Statechart.Current_state, proj.Phases)
	promptCtx.NextActions = generateNextActions(proj.Statechart.Current_state, proj.Phases)
	promptCtx.CurrentPhaseDescription = getCurrentPhaseDescription(proj.Statechart.Current_state)
	promptCtx.NextActionSummary = generateNextActionSummary(proj.Statechart.Current_state, proj.Phases)

	// Render continue section
	continueSection, err := prompts.Render(prompts.PromptCommandContinue, promptCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render continue section: %w", err)
	}

	// Compose: base + continue context
	return base + "\n\n" + continueSection, nil
}

// Helper functions for generating prompt context

func generateStateGuidance(currentState string, _ map[string]project.PhaseState) string {
	switch currentState {
	case "DiscoveryDecision":
		return "You are at the Discovery Decision state. Assess whether discovery work is needed for this project using the Discovery Worthiness Rubric."
	case "DiscoveryActive":
		return "You are in the Discovery phase. Research and investigate to build context for this work."
	case "DesignDecision":
		return "You are at the Design Decision state. Determine if design/architecture work is warranted before implementation."
	case "DesignActive":
		return "You are in the Design phase. Create architectural decisions and design documents as needed."
	case "ImplementationPlanning":
		return "You are in Implementation Planning. Break down the work into concrete tasks for execution."
	case "ImplementationExecuting":
		return "You are executing implementation tasks. Spawn implementer agents to complete each task."
	case "ReviewActive":
		return "You are in the Review phase. Perform quality validation of the completed work."
	case "FinalizeDocumentation":
		return "You are in Finalize (Documentation). Update relevant documentation files."
	case "FinalizeChecks":
		return "You are in Finalize (Checks). Run final validation (tests, linters, build)."
	case "FinalizeDelete":
		return "You are in Finalize (Delete). Clean up project structure and create PR."
	default:
		return "Continue working on the current project phase."
	}
}

func generateNextActions(currentState string, phases map[string]project.PhaseState) string {
	switch currentState {
	case "DiscoveryDecision":
		return "Run the Discovery Worthiness Rubric and decide whether to enable or skip discovery."
	case "DiscoveryActive":
		return "Conduct research, create discovery artifacts, and request human approval when ready."
	case "DesignDecision":
		return "Determine if design work is needed."
	case "DesignActive":
		return "Create ADRs and design documents."
	case "ImplementationPlanning":
		return "Create a task breakdown. When all tasks are defined, mark them approved to transition to execution."
	case "ImplementationExecuting":
		impl, ok := phases["implementation"]
		if ok && impl.Metadata != nil {
			if approved, ok := impl.Metadata["tasks_approved"].(bool); ok && approved {
				return "Spawn implementer agents to execute pending tasks. Update task status as work completes."
			}
		}
		return "Await human approval of the task plan before execution."
	case "ReviewActive":
		return "Perform quality review of completed work. Create review report."
	case "FinalizeDocumentation":
		return "Update documentation files. Record updates."
	case "FinalizeChecks":
		return "Run tests, linters, and builds. Verify all checks pass."
	case "FinalizeDelete":
		return "Delete project folder. Create PR."
	default:
		return "Continue with the current phase work."
	}
}

func getCurrentPhaseDescription(currentState string) string {
	switch currentState {
	case "DiscoveryDecision", "DiscoveryActive":
		return "Discovery phase"
	case "DesignDecision", "DesignActive":
		return "Design phase"
	case "ImplementationPlanning", "ImplementationExecuting":
		return "Implementation phase"
	case "ReviewActive":
		return "Review phase"
	case "FinalizeDocumentation", "FinalizeChecks", "FinalizeDelete":
		return "Finalize phase"
	default:
		return "Unknown phase"
	}
}

func generateNextActionSummary(currentState string, phases map[string]project.PhaseState) string {
	switch currentState {
	case "DiscoveryDecision":
		return "I'll assess if discovery is needed and proceed accordingly."
	case "DiscoveryActive":
		return "I'll continue discovery work."
	case "DesignDecision":
		return "I'll determine if design work is warranted."
	case "DesignActive":
		return "I'll continue design work."
	case "ImplementationPlanning":
		return "I'll work on the task breakdown."
	case "ImplementationExecuting":
		impl, ok := phases["implementation"]
		if ok {
			completed := 0
			total := len(impl.Tasks)
			for _, t := range impl.Tasks {
				if t.Status == "completed" {
					completed++
				}
			}
			if total > 0 && completed == total {
				return "All tasks complete! Moving to review."
			}
			if total > 0 {
				return fmt.Sprintf("I'll continue implementing tasks (%d/%d done).", completed, total)
			}
		}
		return "I'll await task approval to begin execution."
	case "ReviewActive":
		return "I'll perform quality review of the work."
	case "FinalizeDocumentation":
		return "I'll update the necessary documentation."
	case "FinalizeChecks":
		return "I'll run final validation checks."
	case "FinalizeDelete":
		return "I'll clean up and create the PR."
	default:
		return "I'll continue where we left off."
	}
}
