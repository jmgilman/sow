package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/spf13/cobra"
)

// newTaskOutputCmd creates the task output command.
func newTaskOutputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "Manage task output artifacts",
		Long: `Manage task output artifacts.

Task output commands allow you to add, modify, remove, and list output artifacts
for tasks. Artifacts are managed using zero-based indices.`,
	}

	cmd.AddCommand(newTaskOutputAddCmd())
	cmd.AddCommand(newTaskOutputSetCmd())
	cmd.AddCommand(newTaskOutputRemoveCmd())
	cmd.AddCommand(newTaskOutputListCmd())

	return cmd
}

// newTaskOutputAddCmd creates the task output add subcommand.
func newTaskOutputAddCmd() *cobra.Command {
	var taskID, artifactType, path, phase string
	var approved bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add output artifact to task",
		Long: `Add an output artifact to a task.

Creates a new output artifact with the specified type and path.
The artifact is appended to the task's output list.

Examples:
  # Add modified file
  sow task output add --id 010 --type modified --path src/auth/jwt.ts

  # Add with approval flag
  sow task output add --id 010 --type modified --path src/auth/verify.ts --approved`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTaskOutputAdd(cmd, taskID, artifactType, path, phase, approved)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	cmd.Flags().StringVar(&artifactType, "type", "", "Artifact type (required)")
	cmd.Flags().StringVar(&path, "path", "", "Artifact path (required)")
	cmd.Flags().BoolVar(&approved, "approved", false, "Mark artifact as approved")

	cmd.Flags().StringVar(&phase, "phase", "", "Target phase (defaults to current phase)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

// newTaskOutputSetCmd creates the task output set subcommand.
func newTaskOutputSetCmd() *cobra.Command {
	var taskID, phase string
	var index int

	cmd := &cobra.Command{
		Use:   "set <field-path> <value>",
		Short: "Set field on task output artifact by index",
		Long: `Set a field on a task output artifact using its index.

Supports both direct fields and metadata fields using dot notation:
  - Direct fields: type, path, approved
  - Metadata fields: metadata.* (any custom field)

Examples:
  # Set approved field
  sow task output set --id 010 --index 0 approved true

  # Set metadata field
  sow task output set --id 010 --index 0 metadata.reviewed true

  # Set nested metadata
  sow task output set --id 010 --index 0 metadata.reviewer.name orchestrator`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskOutputSet(cmd, args, taskID, phase, index)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	cmd.Flags().IntVar(&index, "index", -1, "Artifact index (required)")

	cmd.Flags().StringVar(&phase, "phase", "", "Target phase (defaults to current phase)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("index")

	return cmd
}

// newTaskOutputRemoveCmd creates the task output remove subcommand.
func newTaskOutputRemoveCmd() *cobra.Command {
	var taskID, phase string
	var index int

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove task output artifact by index",
		Long: `Remove an output artifact from a task by its index.

The artifact at the specified index is removed from the task's output list.
Indices of subsequent artifacts are shifted down.

Examples:
  # Remove first output
  sow task output remove --id 010 --index 0

  # Remove specific file
  sow task output remove --id 010 --index 2`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTaskOutputRemove(cmd, taskID, phase, index)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	cmd.Flags().IntVar(&index, "index", -1, "Artifact index (required)")

	cmd.Flags().StringVar(&phase, "phase", "", "Target phase (defaults to current phase)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("index")

	return cmd
}

// newTaskOutputListCmd creates the task output list subcommand.
func newTaskOutputListCmd() *cobra.Command {
	var taskID, phase string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List output artifacts for task",
		Long: `List all output artifacts for a task.

Displays artifacts with their indices and details including type, path,
approval status, and metadata.

Example:
  sow task output list --id 010`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTaskOutputList(cmd, taskID, phase)
		},
	}

	cmd.Flags().StringVar(&taskID, "id", "", "Task ID (required)")
	cmd.Flags().StringVar(&phase, "phase", "", "Target phase (defaults to current phase)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// runTaskOutputAdd implements the task output add command logic.
func runTaskOutputAdd(cmd *cobra.Command, taskID, artifactType, path, explicitPhase string, approved bool) error {
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

	// Add to task outputs
	phaseState.Tasks[taskIndex].Outputs = append(phaseState.Tasks[taskIndex].Outputs, artifact)
	phaseState.Tasks[taskIndex].Updated_at = time.Now()

	proj.Phases[phaseName] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Added output artifact [%d] to task [%s]\n", len(phaseState.Tasks[taskIndex].Outputs)-1, taskID)
	return nil
}

// runTaskOutputSet implements the task output set command logic.
func runTaskOutputSet(cmd *cobra.Command, args []string, taskID, explicitPhase string, index int) error {
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
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Get field path and value from args
	fieldPath := args[0]
	value := args[1]

	// Wrap artifacts in state.Artifact for field path mutation
	artifacts := make([]state.Artifact, len(phaseState.Tasks[taskIndex].Outputs))
	for i, a := range phaseState.Tasks[taskIndex].Outputs {
		artifacts[i] = state.Artifact{ArtifactState: a}
	}

	// Set field using artifact helper
	if err := cmdutil.SetArtifactField(&artifacts, index, fieldPath, value); err != nil {
		return fmt.Errorf("failed to set field: %w", err)
	}

	// Unwrap back to ArtifactState
	for i, a := range artifacts {
		phaseState.Tasks[taskIndex].Outputs[i] = a.ArtifactState
	}

	phaseState.Tasks[taskIndex].Updated_at = time.Now()

	proj.Phases[phaseName] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Set %s on output artifact [%d] in task [%s]\n", fieldPath, index, taskID)
	return nil
}

// runTaskOutputRemove implements the task output remove command logic.
func runTaskOutputRemove(cmd *cobra.Command, taskID, explicitPhase string, index int) error {
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
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Validate index
	if err := cmdutil.IndexInRange(len(phaseState.Tasks[taskIndex].Outputs), index); err != nil {
		return err
	}

	// Remove artifact by index
	phaseState.Tasks[taskIndex].Outputs = append(
		phaseState.Tasks[taskIndex].Outputs[:index],
		phaseState.Tasks[taskIndex].Outputs[index+1:]...,
	)
	phaseState.Tasks[taskIndex].Updated_at = time.Now()

	proj.Phases[phaseName] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Removed output artifact [%d] from task [%s]\n", index, taskID)
	return nil
}

// runTaskOutputList implements the task output list command logic.
func runTaskOutputList(cmd *cobra.Command, taskID, explicitPhase string) error {
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
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Convert to state.Artifact for formatting
	artifacts := make([]state.Artifact, len(phaseState.Tasks[taskIndex].Outputs))
	for i, a := range phaseState.Tasks[taskIndex].Outputs {
		artifacts[i] = state.Artifact{ArtifactState: a}
	}

	// Format and display
	output := cmdutil.FormatArtifactList(artifacts)
	fmt.Println(output)

	return nil
}
