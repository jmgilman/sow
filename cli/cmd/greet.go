package cmd

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exec"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// NewGreetCmd creates the greet command.
func NewGreetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "greet",
		Short: "Generate orchestrator initialization prompt",
		Long: `Inspect repository state and output a contextual greeting prompt.

This command is designed to be called by the /sow-greet slash command
via the ! prefix. It outputs a fully-rendered prompt based on current
repository and project state.

The output is a complete markdown prompt that Claude processes to
provide a context-aware greeting and initialization.`,
		RunE: runGreet,
	}

	return cmd
}

func runGreet(cmd *cobra.Command, _ []string) error {
	sowCtx := cmdutil.GetContext(cmd.Context())

	// Detect context
	context := detectGreetContext(sowCtx)

	// Select appropriate template based on context
	var promptID prompts.PromptID
	switch {
	case !context.SowInitialized:
		promptID = prompts.PromptGreetStandard
	case !context.HasProject:
		promptID = prompts.PromptGreetOperator
	default:
		promptID = prompts.PromptGreetOrchestrator
	}

	// Render using central prompts package
	output, err := prompts.Render(promptID, context)
	if err != nil {
		return fmt.Errorf("failed to render greeting: %w", err)
	}

	// Write to stdout (becomes slash command content)
	if _, err := fmt.Fprint(cmd.OutOrStdout(), output); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
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
	ghExec := exec.NewLocal("gh")
	gh := sow.NewGitHub(ghExec)
	if err := gh.Ensure(); err == nil {
		ctx.GHAvailable = true
		if issues, err := gh.ListIssues("sow", "open"); err == nil {
			ctx.OpenIssues = len(issues)
		}
	}

	if !projectpkg.Exists(sowCtx) {
		return ctx
	}

	ctx.HasProject = true

	// Load project
	project, err := projectpkg.Load(sowCtx)
	if err != nil {
		// Log error but continue with hasProject=false
		return &prompts.GreetContext{SowInitialized: true}
	}

	state := project.State()

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
	if state.Phases.Discovery.Enabled && state.Phases.Discovery.Status != "completed" && state.Phases.Discovery.Status != "skipped" {
		return "discovery", state.Phases.Discovery.Status
	}
	if state.Phases.Design.Enabled && state.Phases.Design.Status != "completed" && state.Phases.Design.Status != "skipped" {
		return "design", state.Phases.Design.Status
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
