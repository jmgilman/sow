package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/schemas/project"
	"github.com/spf13/cobra"
)

// newTaskInputCmd creates the task input command.
func newTaskInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Short: "Manage task input artifacts",
		Long: `Manage task input artifacts.

Task input commands allow you to add, modify, remove, and list input artifacts
for tasks. Artifacts are managed using zero-based indices.`,
	}

	cmd.AddCommand(newTaskInputAddCmd())
	cmd.AddCommand(newTaskInputSetCmd())
	cmd.AddCommand(newTaskInputRemoveCmd())
	cmd.AddCommand(newTaskInputListCmd())

	return cmd
}

// newTaskInputAddCmd creates the task input add subcommand.
func newTaskInputAddCmd() *cobra.Command {
	var taskID, artifactType, path string
	var approved bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add input artifact to task",
		Long: `Add an input artifact to a task.

Creates a new input artifact with the specified type and path.
The artifact is appended to the task's input list.

Examples:
  # Add reference artifact
  sow task input add --id 010 --type reference --path sinks/style-guide.md

  # Add feedback artifact
  sow task input add --id 010 --type feedback --path feedback/001.md

  # Add with approval flag
  sow task input add --id 010 --type reference --path docs/arch.md --approved`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTaskInputAdd(cmd, taskID, artifactType, path, approved)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	cmd.Flags().StringVar(&artifactType, "type", "", "Artifact type (required)")
	cmd.Flags().StringVar(&path, "path", "", "Artifact path relative to task directory (required)")
	cmd.Flags().BoolVar(&approved, "approved", false, "Mark artifact as approved")

	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

// newTaskInputSetCmd creates the task input set subcommand.
func newTaskInputSetCmd() *cobra.Command {
	var taskID string
	var index int

	cmd := &cobra.Command{
		Use:   "set <field-path> <value>",
		Short: "Set field on task input artifact by index",
		Long: `Set a field on a task input artifact using its index.

Supports both direct fields and metadata fields using dot notation:
  - Direct fields: type, path, approved
  - Metadata fields: metadata.* (any custom field)

Examples:
  # Set approved field
  sow task input set --id 010 --index 0 approved true

  # Set metadata field (e.g., feedback status)
  sow task input set --id 010 --index 2 metadata.status addressed

  # Set nested metadata
  sow task input set --id 010 --index 0 metadata.reviewer.name orchestrator`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskInputSet(cmd, args, taskID, index)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	cmd.Flags().IntVar(&index, "index", -1, "Artifact index (required)")

	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("index")

	return cmd
}

// newTaskInputRemoveCmd creates the task input remove subcommand.
func newTaskInputRemoveCmd() *cobra.Command {
	var taskID string
	var index int

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove task input artifact by index",
		Long: `Remove an input artifact from a task by its index.

The artifact at the specified index is removed from the task's input list.
Indices of subsequent artifacts are shifted down.

Examples:
  # Remove first input
  sow task input remove --id 010 --index 0

  # Remove specific feedback item
  sow task input remove --id 010 --index 2`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTaskInputRemove(cmd, taskID, index)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	cmd.Flags().IntVar(&index, "index", -1, "Artifact index (required)")

	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("index")

	return cmd
}

// newTaskInputListCmd creates the task input list subcommand.
func newTaskInputListCmd() *cobra.Command {
	var taskID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List input artifacts for task",
		Long: `List all input artifacts for a task.

Displays artifacts with their indices and details including type, path,
approval status, and metadata.

Example:
  sow task input list --id 010`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTaskInputList(cmd, taskID)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// runTaskInputAdd implements the task input add command logic.
func runTaskInputAdd(cmd *cobra.Command, taskID, artifactType, path string, approved bool) error {
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

	// Create artifact
	artifact := project.ArtifactState{
		Type:       artifactType,
		Path:       path,
		Approved:   approved,
		Created_at: time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	// Add to task inputs
	phaseState.Tasks[taskIndex].Inputs = append(phaseState.Tasks[taskIndex].Inputs, artifact)
	phaseState.Tasks[taskIndex].Updated_at = time.Now()

	proj.Phases["implementation"] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Added input artifact [%d] to task [%s]\n", len(phaseState.Tasks[taskIndex].Inputs)-1, taskID)
	return nil
}

// runTaskInputSet implements the task input set command logic.
func runTaskInputSet(cmd *cobra.Command, args []string, taskID string, index int) error {
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

	// Get field path and value from args
	fieldPath := args[0]
	value := args[1]

	// Wrap artifacts in state.Artifact for field path mutation
	artifacts := make([]state.Artifact, len(phaseState.Tasks[taskIndex].Inputs))
	for i, a := range phaseState.Tasks[taskIndex].Inputs {
		artifacts[i] = state.Artifact{ArtifactState: a}
	}

	// Set field using artifact helper
	if err := cmdutil.SetArtifactField(&artifacts, index, fieldPath, value); err != nil {
		return fmt.Errorf("failed to set field: %w", err)
	}

	// Unwrap back to ArtifactState
	for i, a := range artifacts {
		phaseState.Tasks[taskIndex].Inputs[i] = a.ArtifactState
	}

	phaseState.Tasks[taskIndex].Updated_at = time.Now()

	proj.Phases["implementation"] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Set %s on input artifact [%d] in task [%s]\n", fieldPath, index, taskID)
	return nil
}

// runTaskInputRemove implements the task input remove command logic.
func runTaskInputRemove(cmd *cobra.Command, taskID string, index int) error {
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

	// Validate index
	if err := cmdutil.IndexInRange(len(phaseState.Tasks[taskIndex].Inputs), index); err != nil {
		return err
	}

	// Remove artifact by index
	phaseState.Tasks[taskIndex].Inputs = append(
		phaseState.Tasks[taskIndex].Inputs[:index],
		phaseState.Tasks[taskIndex].Inputs[index+1:]...,
	)
	phaseState.Tasks[taskIndex].Updated_at = time.Now()

	proj.Phases["implementation"] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Removed input artifact [%d] from task [%s]\n", index, taskID)
	return nil
}

// runTaskInputList implements the task input list command logic.
func runTaskInputList(cmd *cobra.Command, taskID string) error {
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

	// Convert to state.Artifact for formatting
	artifacts := make([]state.Artifact, len(phaseState.Tasks[taskIndex].Inputs))
	for i, a := range phaseState.Tasks[taskIndex].Inputs {
		artifacts[i] = state.Artifact{ArtifactState: a}
	}

	// Format and display
	output := cmdutil.FormatArtifactList(artifacts)
	fmt.Println(output)

	return nil
}
