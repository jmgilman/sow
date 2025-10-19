package task

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/statechart"
	"github.com/jmgilman/sow/cli/internal/task"
	"github.com/spf13/cobra"
)

// NewAddCmd creates the task add command.
func NewAddCmd(accessor SowFSAccessor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new task to the implementation phase",
		Long: `Add a new task to the current project's implementation phase.

Creates a new task with gap-numbered ID (010, 020, 030...) and initializes
the task state, description file, and log file. Tasks are atomic units of
work assigned to specific agent types.

The task ID is auto-generated unless specified with --id. Gap numbering
allows insertion of tasks between existing ones if needed.

Requirements:
  - Must be in a sow repository with an active project
  - Project implementation phase must be enabled
  - Task ID must be unique if specified

Example:
  sow task add "Add authentication" --description "Implement JWT auth" --agent implementer
  sow task add "Review auth code" --description "Review JWT implementation" --agent reviewer --dependencies 010`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(cmd, args, accessor)
		},
	}

	// Flags
	cmd.Flags().StringP("description", "d", "", "Task description (required)")
	cmd.Flags().StringP("agent", "a", "implementer", "Agent type to execute this task")
	cmd.Flags().BoolP("parallel", "p", false, "Whether task can run in parallel with others")
	cmd.Flags().StringSliceP("dependencies", "D", nil, "Task IDs this task depends on (comma-separated)")
	cmd.Flags().String("id", "", "Task ID (auto-generated if not specified)")

	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string, accessor SowFSAccessor) error {
	name := args[0]
	description, _ := cmd.Flags().GetString("description")
	agent, _ := cmd.Flags().GetString("agent")
	parallel, _ := cmd.Flags().GetBool("parallel")
	dependencies, _ := cmd.Flags().GetStringSlice("dependencies")
	idFlag, _ := cmd.Flags().GetString("id")

	// Get SowFS from context
	sowFS := accessor(cmd.Context())
	if sowFS == nil {
		return fmt.Errorf("not in a sow repository - run 'sow init' first")
	}

	// Get project (must exist)
	projectFS, err := sowFS.Project()
	if err != nil {
		return fmt.Errorf("no active project - run 'sow project init' first")
	}

	// === STATECHART INTEGRATION START ===

	// Load machine
	machine, err := statechart.Load()
	if err != nil {
		return fmt.Errorf("failed to load statechart: %w", err)
	}

	state := machine.ProjectState()

	// Validate implementation phase enabled
	if !state.Phases.Implementation.Enabled {
		return fmt.Errorf("implementation phase not enabled")
	}

	// Validate we're in planning or executing state
	currentState := machine.State()
	if currentState != statechart.ImplementationPlanning && currentState != statechart.ImplementationExecuting {
		return fmt.Errorf("cannot add tasks in current state: %s", currentState)
	}

	// Generate or validate task ID
	taskID := idFlag
	if taskID == "" {
		taskID = task.GenerateNextTaskID(state.Phases.Implementation.Tasks)
	} else {
		if err := task.ValidateTaskID(taskID); err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}
	}

	// Trim dependencies
	for i, dep := range dependencies {
		dependencies[i] = strings.TrimSpace(dep)
	}

	// Add task to state IN MEMORY
	if err := task.AddTaskToProjectState(state, taskID, name, parallel, dependencies); err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}

	// Fire event if we're in ImplementationPlanning (triggers transition to ImplementationExecuting)
	// This handles both the initial case (first task) and loop-back from review (additional tasks)
	if currentState == statechart.ImplementationPlanning {
		if err := machine.Fire(statechart.EventTaskCreated); err != nil {
			return fmt.Errorf("failed to transition to executing: %w", err)
		}
	}

	// Save state atomically
	if err := machine.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// === STATECHART INTEGRATION END ===

	// NOW create task files (idempotent on retry)
	taskState := task.NewTaskState(taskID, name, agent)

	taskFS, err := projectFS.TaskUnchecked(taskID)
	if err != nil {
		return fmt.Errorf("failed to create task filesystem: %w", err)
	}

	if err := taskFS.WriteState(taskState); err != nil {
		return fmt.Errorf("task in state but files failed: %w", err)
	}

	descContent := fmt.Sprintf("# Task %s: %s\n\n%s\n", taskID, name, description)
	if err := taskFS.WriteDescription(descContent); err != nil {
		return fmt.Errorf("task in state but files failed: %w", err)
	}

	logHeader := fmt.Sprintf("# Task %s Log\n\nWorker actions will be logged here.\n", taskID)
	if err := taskFS.AppendLog(logHeader); err != nil {
		return fmt.Errorf("task in state but files failed: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Created task %s: %s\n", taskID, name)
	cmd.Printf("\nTask Details:\n")
	cmd.Printf("  ID:           %s\n", taskID)
	cmd.Printf("  Name:         %s\n", name)
	cmd.Printf("  Status:       pending\n")
	cmd.Printf("  Agent:        %s\n", agent)
	cmd.Printf("  Parallel:     %v\n", parallel)
	if len(dependencies) > 0 {
		cmd.Printf("  Dependencies: %s\n", strings.Join(dependencies, ", "))
	} else {
		cmd.Printf("  Dependencies: none\n")
	}

	return nil
}
