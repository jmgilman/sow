package agent

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jmgilman/sow/cli/internal/agents"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/spf13/cobra"
)

// newExecutor is a package-level variable that creates an Executor.
// This allows tests to inject a mock executor.
var newExecutor = func() agents.Executor {
	return agents.NewClaudeExecutor(false, "")
}

// newSpawnCmd creates the spawn subcommand.
func newSpawnCmd() *cobra.Command {
	var phase string

	cmd := &cobra.Command{
		Use:   "spawn <task-id>",
		Short: "Spawn an agent to execute a task",
		Long: `Spawn an agent to execute a task.

The spawn command is used by the orchestrator to delegate work to specialized
worker agents. The agent type is determined by the task's assigned_agent field.

It performs the following steps:

  1. Finds the task by ID in the project state
  2. Looks up the agent from the task's assigned_agent field
  3. Generates a session ID if not present (for crash recovery)
  4. Persists the session ID to task state BEFORE spawning
  5. Invokes the agent subprocess with the task prompt
  6. Blocks until the subprocess exits

The command blocks until the spawned agent completes its work. This is
intentional - the orchestrator waits for workers to finish before continuing.

Session IDs are persisted before spawning to support crash recovery. If the
orchestrator crashes, it can resume the session by reading the persisted ID.

Examples:
  # Spawn agent for task 010 (uses task's assigned_agent)
  sow agent spawn 010

  # Spawn with explicit phase
  sow agent spawn 010 --phase implementation`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSpawn(cmd, args, phase)
		},
	}

	cmd.Flags().StringVar(&phase, "phase", "", "Target phase (defaults to smart resolution)")

	return cmd
}

// runSpawn implements the spawn command logic.
func runSpawn(cmd *cobra.Command, args []string, explicitPhase string) error {
	taskID := args[0]

	// Get sow context
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project state
	proj, err := state.Load(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

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

	// Get agent name from task's assigned_agent field
	// (CUE schema ensures this is never empty)
	agentName := task.Assigned_agent

	// Look up agent by name
	registry := agents.NewAgentRegistry()
	agent, err := registry.Get(agentName)
	if err != nil {
		// Build list of available agents for helpful error message
		availableAgents := registry.List()
		names := make([]string, len(availableAgents))
		for i, a := range availableAgents {
			names[i] = a.Name
		}
		sort.Strings(names)
		return fmt.Errorf("task %s has unknown assigned agent: %s\nAvailable agents: %s", taskID, agentName, strings.Join(names, ", "))
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

	// Build task prompt
	prompt := buildTaskPrompt(taskID, phaseName)

	// Create executor and spawn
	executor := newExecutor()
	if err := executor.Spawn(cmd.Context(), agent, prompt, sessionID); err != nil {
		return fmt.Errorf("spawn failed: %w", err)
	}

	return nil
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
