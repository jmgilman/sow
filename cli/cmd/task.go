package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/project"
	"github.com/spf13/cobra"
)

// NewTaskCmd creates the task command.
func NewTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage project tasks",
		Long: `Manage project tasks.

Task commands allow you to create, modify, and list tasks within project phases.
Tasks are the atomic units of work, each with inputs, outputs, and status tracking.`,
	}

	cmd.AddCommand(newTaskAddCmd())
	cmd.AddCommand(newTaskSetCmd())
	cmd.AddCommand(newTaskAbandonCmd())
	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskInputCmd())
	cmd.AddCommand(newTaskOutputCmd())

	return cmd
}

// newTaskAddCmd creates the task add subcommand.
func newTaskAddCmd() *cobra.Command {
	var agent, description string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new task",
		Long: `Add a new task to the implementation phase.

Creates a new task with a gap-numbered ID (010, 020, 030...) and initializes
the task directory structure:
  - description.md (from --description or placeholder)
  - log.md (empty template)
  - feedback/ (empty directory)

Examples:
  # Add task with agent and description
  sow task add "Implement JWT signing" --agent implementer --description "Create RS256 signing"

  # Add task with defaults
  sow task add "Write tests" --agent implementer`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskAdd(cmd, args[0], agent, description)
		},
	}

	cmd.Flags().StringVar(&agent, "agent", "", "Agent type to assign (required)")
	cmd.Flags().StringVar(&description, "description", "", "Task description")

	_ = cmd.MarkFlagRequired("agent")

	return cmd
}

// newTaskSetCmd creates the task set subcommand.
func newTaskSetCmd() *cobra.Command {
	var taskID string

	cmd := &cobra.Command{
		Use:   "set <field-path> <value>",
		Short: "Set task field value",
		Long: `Set a task field value using dot notation.

Supports:
  - Direct fields: status, iteration, assigned_agent, name
  - Metadata fields: metadata.* (any custom field)

Examples:
  # Set status
  sow task set --id 010 status completed

  # Set iteration
  sow task set --id 010 iteration 2

  # Set metadata field
  sow task set --id 010 metadata.complexity high`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskSet(cmd, args, taskID)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// newTaskAbandonCmd creates the task abandon subcommand.
func newTaskAbandonCmd() *cobra.Command {
	var taskID string

	cmd := &cobra.Command{
		Use:   "abandon",
		Short: "Abandon a task",
		Long: `Mark a task as abandoned.

Sets the task status to "abandoned" and sets the completed_at timestamp.

Example:
  sow task abandon --id 010`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTaskAbandon(cmd, taskID)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// newTaskListCmd creates the task list subcommand.
func newTaskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Long: `List all tasks in the implementation phase.

Displays tasks with their IDs, names, and current status.

Example:
  sow task list`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTaskList(cmd)
		},
	}

	return cmd
}

// runTaskAdd implements the task add command logic.
func runTaskAdd(cmd *cobra.Command, name, agent, description string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project
	proj, err := state.Load(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Get implementation phase
	phaseState, exists := proj.Phases["implementation"]
	if !exists {
		return fmt.Errorf("implementation phase not found")
	}

	// Generate next task ID (gap-numbered)
	taskID := generateNextTaskID(phaseState.Tasks)

	// Create task
	now := time.Now()
	task := project.TaskState{
		Id:             taskID,
		Name:           name,
		Phase:          "implementation",
		Status:         "pending",
		Iteration:      1,
		Assigned_agent: agent,
		Created_at:     now,
		Updated_at:     now,
		Inputs:         []project.ArtifactState{},
		Outputs:        []project.ArtifactState{},
		Metadata:       make(map[string]interface{}),
	}

	// Add to phase tasks
	phaseState.Tasks = append(phaseState.Tasks, task)
	proj.Phases["implementation"] = phaseState

	// Create task directory
	if err := createTaskDirectory(ctx, taskID, description); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Added task [%s] %s\n", taskID, name)
	return nil
}

// runTaskSet implements the task set command logic.
func runTaskSet(cmd *cobra.Command, args []string, taskID string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project
	proj, err := state.Load(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Get implementation phase
	phaseState, exists := proj.Phases["implementation"]
	if !exists {
		return fmt.Errorf("implementation phase not found")
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
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Wrap in Task type for field path mutation
	task := &state.Task{
		TaskState: phaseState.Tasks[taskIndex],
	}

	// Set field using field path parser
	fieldPath := args[0]
	value := args[1]

	if err := cmdutil.SetField(task, fieldPath, value); err != nil {
		return fmt.Errorf("failed to set field: %w", err)
	}

	// Update task's updated_at timestamp
	task.Updated_at = time.Now()

	// Update back in phase
	phaseState.Tasks[taskIndex] = task.TaskState
	proj.Phases["implementation"] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Set %s on task [%s]\n", fieldPath, taskID)
	return nil
}

// runTaskAbandon implements the task abandon command logic.
func runTaskAbandon(cmd *cobra.Command, taskID string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project
	proj, err := state.Load(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Get implementation phase
	phaseState, exists := proj.Phases["implementation"]
	if !exists {
		return fmt.Errorf("implementation phase not found")
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
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Update task status
	now := time.Now()
	phaseState.Tasks[taskIndex].Status = "abandoned"
	phaseState.Tasks[taskIndex].Completed_at = now
	phaseState.Tasks[taskIndex].Updated_at = now

	proj.Phases["implementation"] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Abandoned task [%s]\n", taskID)
	return nil
}

// runTaskList implements the task list command logic.
func runTaskList(cmd *cobra.Command) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project
	proj, err := state.Load(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Get implementation phase
	phaseState, exists := proj.Phases["implementation"]
	if !exists {
		return fmt.Errorf("implementation phase not found")
	}

	// Display tasks
	if len(phaseState.Tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	fmt.Println("Tasks:")
	for _, task := range phaseState.Tasks {
		fmt.Printf("[%s] %s (%s)\n", task.Id, task.Name, task.Status)
	}

	return nil
}

// generateNextTaskID calculates the next gap-numbered task ID.
// Returns "010" for the first task, then "020", "030", etc.
func generateNextTaskID(tasks []project.TaskState) string {
	if len(tasks) == 0 {
		return "010"
	}

	// Find highest ID
	maxID := 0
	for _, task := range tasks {
		id, err := strconv.Atoi(task.Id)
		if err == nil && id > maxID {
			maxID = id
		}
	}

	// Return next gap-numbered ID
	nextID := maxID + 10
	return fmt.Sprintf("%03d", nextID)
}

// createTaskDirectory creates the task directory structure.
func createTaskDirectory(ctx *sow.Context, taskID, description string) error {
	taskDir := filepath.Join(ctx.RepoRoot(), ".sow/project/phases/implementation/tasks", taskID)

	// Create directories
	if err := os.MkdirAll(filepath.Join(taskDir, "feedback"), 0755); err != nil {
		return err
	}

	// Create description.md
	descPath := filepath.Join(taskDir, "description.md")
	descContent := description
	if descContent == "" {
		descContent = "# Task Description\n\nTODO: Add task description\n"
	} else {
		descContent = descContent + "\n"
	}
	if err := os.WriteFile(descPath, []byte(descContent), 0644); err != nil {
		return err
	}

	// Create empty log.md
	logPath := filepath.Join(taskDir, "log.md")
	logContent := "# Task Log\n\nWorker actions will be logged here.\n"
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		return err
	}

	return nil
}
