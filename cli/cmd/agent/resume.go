package agent

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/spf13/cobra"
)

// newResumeCmd creates the resume subcommand.
func newResumeCmd() *cobra.Command {
	var phase string

	cmd := &cobra.Command{
		Use:   "resume <task-id> <prompt>",
		Short: "Resume a paused agent session with feedback",
		Long: `Resume a paused agent session with additional instructions or feedback.

The resume command continues an existing agent session that was previously
spawned. It is used for iterative refinement when the orchestrator needs
to provide additional guidance or corrections to a worker agent.

Prerequisites:
  - The task must have been previously spawned with 'sow agent spawn'
  - A session ID must exist in the task state (set by spawn)
  - The executor must support session resumption

The resume command is typically used in this workflow:
  1. Orchestrator spawns worker -> Worker executes -> Worker pauses
  2. Orchestrator reviews work -> Provides feedback via resume -> Worker continues
  3. Cycle repeats until task complete

The command blocks until the agent subprocess exits.

Examples:
  # Resume task 010 with specific feedback
  sow agent resume 010 "Use RS256 algorithm for JWT signing"

  # Resume with explicit phase
  sow agent resume 010 "Add error handling for edge cases" --phase implementation

  # Multi-word feedback prompts work naturally
  sow agent resume 010 "The tests are failing. Please check the mock setup."`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(cmd, args, phase)
		},
	}

	cmd.Flags().StringVar(&phase, "phase", "", "Target phase (defaults to smart resolution)")

	return cmd
}

// runResume implements the resume command logic.
func runResume(cmd *cobra.Command, args []string, explicitPhase string) error {
	taskID := args[0]
	prompt := args[1]

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

	// Verify session exists
	sessionID := task.Session_id
	if sessionID == "" {
		return fmt.Errorf("no session found for task %s (spawn first with 'sow agent spawn')", taskID)
	}

	// Create executor and check resumption support
	executor := newExecutor()
	if !executor.SupportsResumption() {
		return fmt.Errorf("executor does not support session resumption")
	}

	// Resume session
	if err := executor.Resume(cmd.Context(), sessionID, prompt); err != nil {
		return fmt.Errorf("resume failed: %w", err)
	}

	return nil
}
