package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// NewStartCmd creates the start command.
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch Claude Code in orchestrator mode",
		Long: `Start Claude Code with sow orchestrator.

The orchestrator will greet you and present options based on your
repository state. You can then choose to continue existing work or
start something new.

This command generates a context-aware greeting prompt and launches
Claude Code with it. The greeting automatically detects:
- Whether sow is initialized
- Whether an active project exists
- Current project phase and task status

Claude will greet you with context-aware information and let you
choose what to do next.

Examples:
  sow start    Launch orchestrator with greeting`,
		RunE: runStart,
	}

	return cmd
}

func runStart(cmd *cobra.Command, _ []string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Note: We don't require sow to be initialized - the greeting handles uninitialized state
	// and provides guidance to the user

	// Check claude CLI available
	claude := sowexec.NewLocal("claude")
	if !claude.Exists() {
		fmt.Fprintln(os.Stderr, "Error: Claude Code CLI not found")
		fmt.Fprintln(os.Stderr, "Install from: https://claude.com/download")
		return fmt.Errorf("claude not found")
	}

	// Generate greeting prompt
	greetingPrompt, err := generateGreetingPrompt(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate greeting: %w", err)
	}

	// Launch Claude Code with the greeting prompt
	claudeCmd := exec.CommandContext(cmd.Context(), claude.Command(), greetingPrompt)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr
	claudeCmd.Dir = ctx.RepoRoot()

	// Run and wait for completion
	return claudeCmd.Run()
}

// generateGreetingPrompt creates the context-aware greeting prompt.
func generateGreetingPrompt(sowCtx *sow.Context) (string, error) {
	// Detect context
	context := detectGreetContext(sowCtx)

	// Render base greeting (always included)
	base, err := prompts.Render(prompts.PromptGreetBase, context)
	if err != nil {
		return "", fmt.Errorf("failed to render base greeting: %w", err)
	}

	// Select state template based on context
	var statePromptID prompts.PromptID
	switch {
	case !context.SowInitialized:
		statePromptID = prompts.PromptGreetStateUninit
	case !context.HasProject:
		statePromptID = prompts.PromptGreetStateOperator
	default:
		statePromptID = prompts.PromptGreetStateOrch
	}

	// Render state section
	state, err := prompts.Render(statePromptID, context)
	if err != nil {
		return "", fmt.Errorf("failed to render state section: %w", err)
	}

	// Compose base + state
	return base + "\n\n" + state, nil
}

// detectGreetContext inspects the repository and builds greeting context.
func detectGreetContext(sowCtx *sow.Context) *prompts.GreetContext {
	ctx := &prompts.GreetContext{
		SowInitialized: sowCtx.IsInitialized(),
	}

	if !ctx.SowInitialized {
		return ctx
	}

	// Try to query GitHub for open sow issues
	ghExec := sowexec.NewLocal("gh")
	gh := sow.NewGitHub(ghExec)
	if err := gh.Ensure(); err == nil {
		ctx.GHAvailable = true
		if issues, err := gh.ListIssues("sow", "open"); err == nil {
			ctx.OpenIssues = len(issues)
		}
	}

	if !loader.Exists(sowCtx) {
		return ctx
	}

	ctx.HasProject = true

	// Load project
	project, err := loader.Load(sowCtx)
	if err != nil {
		// Log error but continue with hasProject=false
		return &prompts.GreetContext{SowInitialized: true}
	}

	state := project.Machine().ProjectState()

	// Build project context
	projCtx := &prompts.ProjectGreetContext{
		Name:        state.Project.Name,
		Branch:      state.Project.Branch,
		Description: state.Project.Description,
	}

	// Determine current phase
	currentPhase, phaseStatus := determineCurrentPhase(state)
	projCtx.CurrentPhase = currentPhase
	projCtx.PhaseStatus = phaseStatus

	// Count tasks
	tasks := state.Phases.Implementation.Tasks
	projCtx.TasksTotal = len(tasks)

	for i := range tasks {
		switch tasks[i].Status {
		case "completed":
			projCtx.TasksComplete++
		case "in_progress":
			projCtx.TasksInProgress++
			if projCtx.CurrentTask == nil {
				projCtx.CurrentTask = &prompts.TaskGreetContext{
					ID:   tasks[i].Id,
					Name: tasks[i].Name,
				}
			}
		case "pending":
			projCtx.TasksPending++
		case "abandoned":
			projCtx.TasksAbandoned++
		}
	}

	ctx.Project = projCtx

	return ctx
}

// determineCurrentPhase finds the active phase in the project.
func determineCurrentPhase(state *schemas.ProjectState) (string, string) {
	// Check phases in order
	if state.Phases.Planning.Status != "completed" && state.Phases.Planning.Status != "skipped" {
		return "planning", state.Phases.Planning.Status
	}
	if state.Phases.Implementation.Status != "completed" && state.Phases.Implementation.Status != "skipped" {
		return "implementation", state.Phases.Implementation.Status
	}
	if state.Phases.Review.Status != "completed" && state.Phases.Review.Status != "skipped" {
		return "review", state.Phases.Review.Status
	}
	if state.Phases.Finalize.Status != "completed" && state.Phases.Finalize.Status != "skipped" {
		return "finalize", state.Phases.Finalize.Status
	}

	return "unknown", "unknown"
}
