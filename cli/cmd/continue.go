package cmd

import (
	"fmt"
	"os"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// NewContinueCmd creates the continue command.
func NewContinueCmd() *cobra.Command {
	var branchName string
	var issueNumber int

	cmd := &cobra.Command{
		Use:   "continue",
		Short: "Continue existing project work",
		Long: `Continue work on an existing sow project and launch Claude Code with project context.

This command loads the existing project state and launches the orchestrator
with a custom prompt showing current progress and next steps.

The command can continue projects in different ways:

Default (no flags):
  - Continues the project in the current branch
  - Loads project state and shows current status

With --branch:
  - Checks out the specified branch first
  - Then continues the project in that branch

With --issue:
  - Finds the branch linked to the GitHub issue
  - Checks out that branch
  - Then continues the project

Examples:
  sow continue
  sow continue --branch feat/auth
  sow continue --issue 123`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runContinue(cmd, branchName, issueNumber)
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch to checkout before continuing")
	cmd.Flags().IntVar(&issueNumber, "issue", 0, "GitHub issue number whose branch to checkout")
	cmd.MarkFlagsMutuallyExclusive("branch", "issue")

	return cmd
}

func runContinue(cmd *cobra.Command, branchName string, issueNumber int) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Require sow to be initialized
	if !ctx.IsInitialized() {
		fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
		fmt.Fprintln(os.Stderr, "Run: sow init")
		return fmt.Errorf("not initialized")
	}

	// Select and checkout branch
	selectedBranch, err := selectBranch(ctx, branchName, issueNumber)
	if err != nil {
		return err
	}

	// Check project exists
	if !projectpkg.Exists(ctx) {
		return fmt.Errorf("no project found in branch '%s' - use 'sow new' to create one", selectedBranch)
	}

	// Load project
	project, err := projectpkg.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Generate continue prompt
	continuePrompt, err := generateContinuePrompt(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Launch Claude Code with the prompt
	return launchClaudeCode(cmd, ctx, continuePrompt)
}

// selectBranch determines which branch to use and checks it out if needed.
func selectBranch(ctx *sow.Context, branchName string, issueNumber int) (string, error) {
	// Handle branch selection
	if branchName != "" {
		// Checkout specified branch
		if err := ctx.Git().CheckoutBranch(branchName); err != nil {
			return "", fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
		}
		return branchName, nil
	}

	if issueNumber > 0 {
		// Find and checkout issue's linked branch
		return checkoutIssueBranch(ctx, issueNumber)
	}

	// Use current branch
	currentBranch, err := ctx.Git().CurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return currentBranch, nil
}

// checkoutIssueBranch finds the branch linked to an issue and checks it out.
func checkoutIssueBranch(ctx *sow.Context, issueNumber int) (string, error) {
	ghExec := sowexec.NewLocal("gh")
	gh := sow.NewGitHub(ghExec)

	// Get linked branches
	branches, err := gh.GetLinkedBranches(issueNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get linked branches for issue #%d: %w", issueNumber, err)
	}

	// Error if no linked branches
	if len(branches) == 0 {
		return "", fmt.Errorf("no linked branch found for issue #%d - create one with: sow new --issue %d", issueNumber, issueNumber)
	}

	// Error if multiple linked branches
	if len(branches) > 1 {
		fmt.Fprintf(os.Stderr, "Error: issue #%d has multiple linked branches:\n", issueNumber)
		for _, b := range branches {
			fmt.Fprintf(os.Stderr, "  - %s (%s)\n", b.Name, b.URL)
		}
		fmt.Fprintf(os.Stderr, "\nSpecify the branch explicitly with: sow continue --branch <branch-name>\n")
		return "", fmt.Errorf("multiple linked branches found")
	}

	// Checkout the single linked branch
	branchName := branches[0].Name
	if err := ctx.Git().CheckoutBranch(branchName); err != nil {
		return "", fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
	}

	return branchName, nil
}

// generateContinuePrompt creates the custom prompt for continuing projects.
// Composes base greeting + continue project context.
func generateContinuePrompt(ctx *sow.Context, project *projectpkg.Project) (string, error) {
	state := project.State()

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
		BranchName:         state.Project.Branch,
		ProjectName:        state.Project.Name,
		ProjectDescription: state.Project.Description,
		StatechartState:    state.Statechart.Current_state,
	}

	// Add issue number if present
	if state.Project.Github_issue != nil {
		issueNum := int(*state.Project.Github_issue)
		promptCtx.IssueNumber = &issueNum
	}

	// Add phase status
	promptCtx.DiscoveryEnabled = state.Phases.Discovery.Enabled
	promptCtx.DiscoveryStatus = state.Phases.Discovery.Status
	promptCtx.DesignEnabled = state.Phases.Design.Enabled
	promptCtx.DesignStatus = state.Phases.Design.Status
	promptCtx.ImplementationStatus = state.Phases.Implementation.Status
	promptCtx.ReviewStatus = state.Phases.Review.Status
	promptCtx.FinalizeStatus = state.Phases.Finalize.Status

	// Count tasks
	tasks := state.Phases.Implementation.Tasks
	promptCtx.TasksTotal = len(tasks)

	for i := range tasks {
		switch tasks[i].Status {
		case "completed":
			promptCtx.TasksCompleted++
		case "in_progress":
			promptCtx.TasksInProgress++
			if promptCtx.CurrentTaskID == "" {
				promptCtx.CurrentTaskID = tasks[i].Id
				promptCtx.CurrentTaskName = tasks[i].Name
				promptCtx.CurrentTaskStatus = tasks[i].Status
			}
		case "pending":
			promptCtx.TasksPending++
		case "abandoned":
			promptCtx.TasksAbandoned++
		}
	}

	// Generate state-specific guidance and next actions
	promptCtx.StateSpecificGuidance = generateStateGuidance(state)
	promptCtx.NextActions = generateNextActions(state)
	promptCtx.CurrentPhaseDescription = getCurrentPhaseDescription(state)
	promptCtx.NextActionSummary = generateNextActionSummary(state)

	// Render continue section
	continueSection, err := prompts.Render(prompts.PromptCommandContinue, promptCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render continue section: %w", err)
	}

	// Compose: base + continue context
	return base + "\n\n" + continueSection, nil
}

// generateStateGuidance provides context-specific guidance based on current state.
func generateStateGuidance(state *schemas.ProjectState) string {
	switch state.Statechart.Current_state {
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

// generateNextActions provides specific next actions based on current state.
func generateNextActions(state *schemas.ProjectState) string {
	switch state.Statechart.Current_state {
	case "DiscoveryDecision":
		return "Run the Discovery Worthiness Rubric and decide whether to enable or skip discovery. Use `sow agent project phase enable discovery --type <type>` or `sow agent project phase skip discovery`."
	case "DiscoveryActive":
		return "Conduct research, create discovery artifacts, and request human approval when ready. Use `sow agent project artifact add` and `sow agent project phase complete discovery` when done."
	case "DesignDecision":
		return "Determine if design work is needed. Enable design with `sow agent project phase enable design` or skip with `sow agent project phase skip design`."
	case "DesignActive":
		return "Create ADRs and design documents. Use `sow agent project artifact add` and `sow agent project phase complete design` when done."
	case "ImplementationPlanning":
		return "Create a task breakdown using `sow agent task add`. When all tasks are defined, mark them approved to transition to execution."
	case "ImplementationExecuting":
		if state.Phases.Implementation.Tasks_approved {
			return "Spawn implementer agents via Task tool to execute pending tasks. Update task status as work completes."
		}
		return "Await human approval of the task plan before execution."
	case "ReviewActive":
		return "Perform quality review of completed work. Create review report with `sow agent project review add-report`. Approve to finalize or fail to loop back to implementation."
	case "FinalizeDocumentation":
		return "Update documentation files. Record updates with `sow agent project finalize add-document`. Transition when complete."
	case "FinalizeChecks":
		return "Run tests, linters, and builds. Verify all checks pass. Transition when validated."
	case "FinalizeDelete":
		return "Delete project folder with `sow agent project delete`. Create PR with `sow agent project create-pr`."
	default:
		return "Continue with the current phase work."
	}
}

// getCurrentPhaseDescription returns a human-readable current phase description.
func getCurrentPhaseDescription(state *schemas.ProjectState) string {
	currentPhase, phaseStatus := determineCurrentPhase(state)
	return fmt.Sprintf("%s phase (%s)", currentPhase, phaseStatus)
}

// generateNextActionSummary provides a brief summary of what's next.
func generateNextActionSummary(state *schemas.ProjectState) string {
	switch state.Statechart.Current_state {
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
		if len(state.Phases.Implementation.Tasks) > 0 {
			completed := 0
			for _, t := range state.Phases.Implementation.Tasks {
				if t.Status == "completed" {
					completed++
				}
			}
			if completed == len(state.Phases.Implementation.Tasks) {
				return "All tasks complete! Moving to review."
			}
			return fmt.Sprintf("I'll continue implementing tasks (%d/%d done).", completed, len(state.Phases.Implementation.Tasks))
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
