package agent

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jmgilman/sow/cli/internal/agents"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/libs/config"
	"github.com/jmgilman/sow/libs/schemas"
	"github.com/spf13/cobra"
)

// loadExecutorRegistry is a package-level variable that creates an ExecutorRegistry.
// This allows tests to inject a mock registry.
// Parameters:
//   - userConfig: user configuration with executor definitions
//   - outputDir: where agent output logs are saved
var loadExecutorRegistry = func(userConfig *schemas.UserConfig, outputDir string) (*agents.ExecutorRegistry, error) {
	return agents.LoadExecutorRegistry(userConfig, outputDir)
}

// newSpawnCmd creates the spawn subcommand.
func newSpawnCmd() *cobra.Command {
	var phase string
	var agentName string
	var customPrompt string

	cmd := &cobra.Command{
		Use:   "spawn [task-id]",
		Short: "Spawn an agent to execute work",
		Long: `Spawn an agent to execute work.

The spawn command has two modes:

TASK MODE (with task-id):
  Spawns an agent to execute a specific task. The agent is determined by the
  task's assigned_agent field unless overridden with --agent.

  sow agent spawn 010                     # Use task's assigned agent
  sow agent spawn 010 --agent reviewer    # Override with different agent
  sow agent spawn 010 --prompt "Focus on error handling"

TASKLESS MODE (with --agent only):
  Spawns an agent directly without a task. Useful for orchestrator to spawn
  planning or research agents before tasks exist.

  sow agent spawn --agent planner --prompt "Create implementation plan"
  sow agent spawn --agent researcher --prompt "Research auth libraries"

Session IDs are persisted before spawning to support crash recovery. For task
mode, session IDs are stored in the task state. For taskless mode, session IDs
are stored in the project's agent_sessions map.

Examples:
  # Task mode: spawn agent for task 010
  sow agent spawn 010

  # Task mode with custom prompt
  sow agent spawn 010 --prompt "Focus on performance"

  # Task mode with agent override
  sow agent spawn 010 --agent reviewer

  # Taskless mode: spawn planner directly
  sow agent spawn --agent planner --prompt "Plan the auth feature"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSpawn(cmd, args, phase, agentName, customPrompt)
		},
	}

	cmd.Flags().StringVar(&phase, "phase", "", "Target phase (defaults to smart resolution)")
	cmd.Flags().StringVar(&agentName, "agent", "", "Agent name (required when no task-id, optional override when task-id provided)")
	cmd.Flags().StringVar(&customPrompt, "prompt", "", "Additional prompt context to append")

	return cmd
}

// runSpawn implements the spawn command logic.
func runSpawn(cmd *cobra.Command, args []string, explicitPhase, agentFlag, customPrompt string) error {
	// Get sow context
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Determine mode based on arguments
	hasTaskID := len(args) > 0
	hasAgentFlag := agentFlag != ""

	// Validation: need at least one of task-id or --agent
	if !hasTaskID && !hasAgentFlag {
		return fmt.Errorf("must provide either <task-id> or --agent flag")
	}

	// Load project state (required for both modes)
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

	if hasTaskID {
		return runSpawnWithTask(cmd, proj, args[0], explicitPhase, agentFlag, customPrompt, executorRegistry, bindings)
	}
	return runSpawnTaskless(cmd, proj, agentFlag, customPrompt, executorRegistry, bindings)
}

// runSpawnWithTask handles spawning an agent for a specific task.
func runSpawnWithTask(cmd *cobra.Command, proj *state.Project, taskID, explicitPhase, agentOverride, customPrompt string, executorRegistry *agents.ExecutorRegistry, bindings *struct {
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

	// Determine agent: use override if provided, otherwise use task's assigned_agent
	agentName := task.Assigned_agent
	if agentOverride != "" {
		agentName = agentOverride
	}

	// Look up agent by name
	agentRegistry := agents.NewAgentRegistry()
	agent, err := agentRegistry.Get(agentName)
	if err != nil {
		return buildAgentNotFoundError(agentName, taskID, agentRegistry)
	}

	// Look up executor for this agent
	executor, err := executorRegistry.GetAgentExecutor(agentName, bindings)
	if err != nil {
		return fmt.Errorf("failed to get executor for agent %s: %w", agentName, err)
	}

	// Handle session ID
	sessionID := task.Session_id
	if sessionID == "" {
		// Generate new session ID
		sessionID = uuid.New().String()
		task.Session_id = sessionID

		// Update phase with modified task
		phaseState.Tasks[taskIndex] = *task
		proj.Phases[phaseName] = phaseState

		// CRITICAL: Save before spawning (crash recovery)
		if err := proj.Save(); err != nil {
			return fmt.Errorf("failed to save session ID: %w", err)
		}
	}

	// Build task prompt with optional custom prompt
	prompt := buildTaskPrompt(taskID, phaseName)
	if customPrompt != "" {
		prompt += "\n\n" + customPrompt
	}

	// Validate and spawn
	if err := executor.ValidateAvailability(); err != nil {
		return fmt.Errorf("executor not available: %w", err)
	}
	if err := executor.Spawn(cmd.Context(), agent, prompt, sessionID); err != nil {
		return fmt.Errorf("spawn failed: %w", err)
	}

	return nil
}

// runSpawnTaskless handles spawning an agent without a task.
func runSpawnTaskless(cmd *cobra.Command, proj *state.Project, agentName, customPrompt string, executorRegistry *agents.ExecutorRegistry, bindings *struct {
	Orchestrator *string `json:"orchestrator,omitempty"`
	Implementer  *string `json:"implementer,omitempty"`
	Architect    *string `json:"architect,omitempty"`
	Reviewer     *string `json:"reviewer,omitempty"`
	Planner      *string `json:"planner,omitempty"`
	Researcher   *string `json:"researcher,omitempty"`
	Decomposer   *string `json:"decomposer,omitempty"`
}) error {
	// Look up agent by name
	agentRegistry := agents.NewAgentRegistry()
	agent, err := agentRegistry.Get(agentName)
	if err != nil {
		return buildAgentNotFoundError(agentName, "", agentRegistry)
	}

	// Look up executor for this agent
	executor, err := executorRegistry.GetAgentExecutor(agentName, bindings)
	if err != nil {
		return fmt.Errorf("failed to get executor for agent %s: %w", agentName, err)
	}

	// Initialize agent_sessions map if nil
	if proj.Agent_sessions == nil {
		proj.Agent_sessions = make(map[string]string)
	}

	// Handle session ID: use existing or generate new
	sessionID := proj.Agent_sessions[agentName]
	if sessionID == "" {
		sessionID = uuid.New().String()
		proj.Agent_sessions[agentName] = sessionID

		// CRITICAL: Save before spawning (crash recovery)
		if err := proj.Save(); err != nil {
			return fmt.Errorf("failed to save session ID: %w", err)
		}
	}

	// Build prompt: just custom prompt for taskless spawn
	// The executor will prepend the agent template
	prompt := customPrompt
	if prompt == "" {
		prompt = "You have been spawned. Follow your agent instructions."
	}

	// Validate and spawn
	if err := executor.ValidateAvailability(); err != nil {
		return fmt.Errorf("executor not available: %w", err)
	}
	if err := executor.Spawn(cmd.Context(), agent, prompt, sessionID); err != nil {
		return fmt.Errorf("spawn failed: %w", err)
	}

	return nil
}

// buildAgentNotFoundError creates a helpful error message when an agent is not found.
func buildAgentNotFoundError(agentName, taskID string, registry *agents.AgentRegistry) error {
	availableAgents := registry.List()
	names := make([]string, len(availableAgents))
	for i, a := range availableAgents {
		names[i] = a.Name
	}
	sort.Strings(names)

	if taskID != "" {
		return fmt.Errorf("task %s has unknown assigned agent: %s\nAvailable agents: %s", taskID, agentName, strings.Join(names, ", "))
	}
	return fmt.Errorf("unknown agent: %s\nAvailable agents: %s", agentName, strings.Join(names, ", "))
}

// buildTaskPrompt creates the prompt to send to the spawned agent.
func buildTaskPrompt(taskID, phaseName string) string {
	return fmt.Sprintf(`Execute task %s.

Task location: .sow/project/phases/%s/tasks/%s/

Read state.yaml for task metadata, description.md for requirements,
and feedback/ for any corrections from previous iterations.
`, taskID, phaseName, taskID)
}

// resolveTaskPhase determines which phase to use for task operations.
// This is a local copy of the helper from cmd/helpers.go to avoid import cycles.
// It follows this priority:
//  1. Explicit --phase flag value (if provided)
//  2. Smart default based on project state
//  3. Error if no phase supports tasks
//
// Returns the resolved phase name or an error with helpful guidance.
func resolveTaskPhase(project *state.Project, explicitPhase string) (string, error) {
	// Get the project type config to query task support
	config := project.Config()

	// Case 1: Explicit phase provided via --phase flag
	if explicitPhase != "" {
		// Validate that the phase supports tasks
		if !config.PhaseSupportsTasks(explicitPhase) {
			supportedPhases := config.GetTaskSupportingPhases()
			if len(supportedPhases) == 0 {
				return "", fmt.Errorf("phase %s does not support tasks (no phases support tasks in this project type)", explicitPhase)
			}
			return "", fmt.Errorf("phase %s does not support tasks (supported phases: %s)",
				explicitPhase, strings.Join(supportedPhases, ", "))
		}
		return explicitPhase, nil
	}

	// Case 2: Smart default based on current project state
	currentState := state.State(project.Statechart.Current_state)
	defaultPhase := config.GetDefaultTaskPhase(currentState)

	if defaultPhase == "" {
		// No smart default found - provide helpful error
		supportedPhases := config.GetTaskSupportingPhases()
		if len(supportedPhases) == 0 {
			return "", fmt.Errorf("no phases in this project type support tasks")
		}
		return "", fmt.Errorf("could not determine default task phase for state %s\nSupported phases: %s\nSpecify phase with --phase flag",
			currentState, strings.Join(supportedPhases, ", "))
	}

	return defaultPhase, nil
}
