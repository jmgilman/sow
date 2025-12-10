package agent

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jmgilman/sow/cli/internal/agents"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/libs/config"
	"github.com/spf13/cobra"
)

// newResumeCmd creates the resume subcommand.
func newResumeCmd() *cobra.Command {
	var phase string
	var agentName string

	cmd := &cobra.Command{
		Use:   "resume [task-id] <prompt>",
		Short: "Resume a paused agent session with feedback",
		Long: `Resume a paused agent session with additional instructions or feedback.

The resume command has two modes:

TASK MODE (with task-id):
  Resumes an agent session that was previously spawned for a task.

  sow agent resume 010 "Use RS256 algorithm for JWT signing"
  sow agent resume 010 "Add error handling" --phase implementation

TASKLESS MODE (with --agent):
  Resumes an agent session that was spawned without a task.

  sow agent resume --agent planner "Focus on auth module first"
  sow agent resume --agent researcher "Look at OAuth 2.0 specifically"

Prerequisites:
  - The session must have been previously created with 'sow agent spawn'
  - The executor must support session resumption

The resume command is typically used in this workflow:
  1. Orchestrator spawns worker -> Worker executes -> Worker pauses
  2. Orchestrator reviews work -> Provides feedback via resume -> Worker continues
  3. Cycle repeats until task complete

Examples:
  # Task mode: resume task 010 with feedback
  sow agent resume 010 "Fix the failing tests"

  # Taskless mode: resume planner with additional context
  sow agent resume --agent planner "Also consider caching requirements"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(cmd, args, phase, agentName)
		},
	}

	cmd.Flags().StringVar(&phase, "phase", "", "Target phase (defaults to smart resolution)")
	cmd.Flags().StringVar(&agentName, "agent", "", "Agent name (for taskless session resumption)")

	return cmd
}

// runResume implements the resume command logic.
func runResume(cmd *cobra.Command, args []string, explicitPhase, agentFlag string) error {
	// Get sow context
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Determine mode based on --agent flag
	hasAgentFlag := agentFlag != ""

	// Parse arguments based on mode
	var taskID, prompt string
	if hasAgentFlag {
		// Taskless mode: only prompt is provided
		if len(args) != 1 {
			return fmt.Errorf("taskless resume requires exactly one argument: <prompt>\nUsage: sow agent resume --agent <name> <prompt>")
		}
		prompt = args[0]
	} else {
		// Task mode: task-id and prompt required
		if len(args) != 2 {
			return fmt.Errorf("task resume requires two arguments: <task-id> <prompt>\nUsage: sow agent resume <task-id> <prompt>")
		}
		taskID = args[0]
		prompt = args[1]
	}

	// Load project state
	proj, err := state.Load(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Load user config for executor settings
	userConfig, err := config.LoadUserConfig(ctx.FS())
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Compute output directory for agent logs
	outputDir := filepath.Join(ctx.RepoRoot(), ".sow", "project", "agent-outputs")

	// Create executor registry from user config
	executorRegistry, err := loadExecutorRegistry(userConfig, outputDir)
	if err != nil {
		return fmt.Errorf("failed to load executor registry: %w", err)
	}

	// Get bindings for executor lookup
	var bindings *struct {
		Orchestrator *string `json:"orchestrator,omitempty"`
		Implementer  *string `json:"implementer,omitempty"`
		Architect    *string `json:"architect,omitempty"`
		Reviewer     *string `json:"reviewer,omitempty"`
		Planner      *string `json:"planner,omitempty"`
		Researcher   *string `json:"researcher,omitempty"`
		Decomposer   *string `json:"decomposer,omitempty"`
	}
	if userConfig != nil && userConfig.Agents != nil {
		bindings = userConfig.Agents.Bindings
	}

	if hasAgentFlag {
		return runResumeTaskless(cmd, proj, agentFlag, prompt, executorRegistry, bindings)
	}
	return runResumeWithTask(cmd, proj, taskID, prompt, explicitPhase, executorRegistry, bindings)
}

// runResumeWithTask handles resuming a session for a specific task.
func runResumeWithTask(cmd *cobra.Command, proj *state.Project, taskID, prompt, explicitPhase string, executorRegistry *agents.ExecutorRegistry, bindings *struct {
	Orchestrator *string `json:"orchestrator,omitempty"`
	Implementer  *string `json:"implementer,omitempty"`
	Architect    *string `json:"architect,omitempty"`
	Reviewer     *string `json:"reviewer,omitempty"`
	Planner      *string `json:"planner,omitempty"`
	Researcher   *string `json:"researcher,omitempty"`
	Decomposer   *string `json:"decomposer,omitempty"`
}) error {
	// Resolve which phase to use
	phaseName, err := resolveTaskPhase(proj, explicitPhase)
	if err != nil {
		return err
	}

	// Get phase
	phaseState, exists := proj.Phases[phaseName]
	if !exists {
		return fmt.Errorf("phase not found: %s", phaseName)
	}

	// Find task by ID
	taskIndex := -1
	for i, t := range phaseState.Tasks {
		if t.Id == taskID {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return fmt.Errorf("task %s not found in phase %s", taskID, phaseName)
	}

	task := &phaseState.Tasks[taskIndex]

	// Verify session exists
	sessionID := task.Session_id
	if sessionID == "" {
		return fmt.Errorf("no session found for task %s (spawn first with 'sow agent spawn')", taskID)
	}

	// Determine agent name from task's assigned_agent
	agentName := task.Assigned_agent

	// Look up executor for this agent
	executor, err := executorRegistry.GetAgentExecutor(agentName, bindings)
	if err != nil {
		return fmt.Errorf("failed to get executor for agent %s: %w", agentName, err)
	}

	// Check resumption support
	if !executor.SupportsResumption() {
		return fmt.Errorf("executor does not support session resumption")
	}

	// Resume session
	if err := executor.Resume(cmd.Context(), sessionID, prompt); err != nil {
		return fmt.Errorf("resume failed: %w", err)
	}

	return nil
}

// runResumeTaskless handles resuming a session for a taskless agent.
func runResumeTaskless(cmd *cobra.Command, proj *state.Project, agentName, prompt string, executorRegistry *agents.ExecutorRegistry, bindings *struct {
	Orchestrator *string `json:"orchestrator,omitempty"`
	Implementer  *string `json:"implementer,omitempty"`
	Architect    *string `json:"architect,omitempty"`
	Reviewer     *string `json:"reviewer,omitempty"`
	Planner      *string `json:"planner,omitempty"`
	Researcher   *string `json:"researcher,omitempty"`
	Decomposer   *string `json:"decomposer,omitempty"`
}) error {
	// Look up session ID from project's agent_sessions
	if proj.Agent_sessions == nil {
		return fmt.Errorf("no session found for agent %s (spawn first with 'sow agent spawn --agent %s')", agentName, agentName)
	}

	sessionID, exists := proj.Agent_sessions[agentName]
	if !exists || sessionID == "" {
		return fmt.Errorf("no session found for agent %s (spawn first with 'sow agent spawn --agent %s')", agentName, agentName)
	}

	// Look up executor for this agent
	executor, err := executorRegistry.GetAgentExecutor(agentName, bindings)
	if err != nil {
		return fmt.Errorf("failed to get executor for agent %s: %w", agentName, err)
	}

	// Check resumption support
	if !executor.SupportsResumption() {
		return fmt.Errorf("executor does not support session resumption")
	}

	// Resume session
	if err := executor.Resume(cmd.Context(), sessionID, prompt); err != nil {
		return fmt.Errorf("resume failed: %w", err)
	}

	return nil
}
