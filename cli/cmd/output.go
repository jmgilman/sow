package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/libs/project/state"
	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/spf13/cobra"
)

// NewOutputCmd creates the output command.
func NewOutputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "Manage phase output artifacts",
		Long: `Manage phase output artifacts.

Output commands allow you to add, modify, remove, and list output artifacts
for project phases. Artifacts are managed using zero-based indices.`,
	}

	cmd.AddCommand(newOutputAddCmd())
	cmd.AddCommand(newOutputSetCmd())
	cmd.AddCommand(newOutputRemoveCmd())
	cmd.AddCommand(newOutputListCmd())

	return cmd
}

// newOutputAddCmd creates the output add subcommand.
func newOutputAddCmd() *cobra.Command {
	var phaseName, artifactType, path string
	var approved bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add output artifact to phase",
		Long: `Add an output artifact to a phase.

Creates a new output artifact with the specified type and path.
The artifact is appended to the phase's output list.

Examples:
  # Add task_list artifact to planning phase
  sow output add --type task_list --path planning/tasks.md --phase planning

  # Add with approval flag
  sow output add --type review --path review/report.md --approved --phase review

  # Use active phase (defaults to current state)
  sow output add --type task_list --path planning/breakdown.md`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runOutputAdd(cmd, phaseName, artifactType, path, approved)
		},
	}

	cmd.Flags().StringVarP(&phaseName, "phase", "p", "", "Target phase (defaults to active phase)")
	cmd.Flags().StringVar(&artifactType, "type", "", "Artifact type (required)")
	cmd.Flags().StringVar(&path, "path", "", "Artifact path relative to .sow/project/ (required)")
	cmd.Flags().BoolVar(&approved, "approved", false, "Mark artifact as approved")

	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

// newOutputSetCmd creates the output set subcommand.
func newOutputSetCmd() *cobra.Command {
	var phaseName string
	var index int

	cmd := &cobra.Command{
		Use:   "set <field-path> <value>",
		Short: "Set field on output artifact by index",
		Long: `Set a field on an output artifact using its index.

Supports both direct fields and metadata fields using dot notation:
  - Direct fields: type, path, approved
  - Metadata fields: metadata.* (any custom field)

Examples:
  # Set approved field
  sow output set --index 0 approved true --phase planning

  # Set metadata field
  sow output set --index 0 metadata.assessment pass --phase review

  # Set nested metadata
  sow output set --index 0 metadata.reviewer.name orchestrator --phase review`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOutputSet(cmd, args, phaseName, index)
		},
	}

	cmd.Flags().StringVarP(&phaseName, "phase", "p", "", "Target phase (defaults to active phase)")
	cmd.Flags().IntVar(&index, "index", -1, "Artifact index (required)")

	_ = cmd.MarkFlagRequired("index")

	return cmd
}

// newOutputRemoveCmd creates the output remove subcommand.
func newOutputRemoveCmd() *cobra.Command {
	var phaseName string
	var index int

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove output artifact by index",
		Long: `Remove an output artifact from a phase by its index.

The artifact at the specified index is removed from the phase's output list.
Indices of subsequent artifacts are shifted down.

Examples:
  # Remove first output
  sow output remove --index 0 --phase planning

  # Remove from active phase
  sow output remove --index 1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runOutputRemove(cmd, phaseName, index)
		},
	}

	cmd.Flags().StringVarP(&phaseName, "phase", "p", "", "Target phase (defaults to active phase)")
	cmd.Flags().IntVar(&index, "index", -1, "Artifact index (required)")

	_ = cmd.MarkFlagRequired("index")

	return cmd
}

// newOutputListCmd creates the output list subcommand.
func newOutputListCmd() *cobra.Command {
	var phaseName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List output artifacts for phase",
		Long: `List all output artifacts for a phase.

Displays artifacts with their indices and details including type, path,
approval status, and metadata.

Examples:
  # List outputs for planning phase
  sow output list --phase planning

  # List outputs for active phase
  sow output list`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runOutputList(cmd, phaseName)
		},
	}

	cmd.Flags().StringVarP(&phaseName, "phase", "p", "", "Target phase (defaults to active phase)")

	return cmd
}

// runOutputAdd implements the output add command logic.
func runOutputAdd(cmd *cobra.Command, phaseName, artifactType, path string, approved bool) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project
	proj, err := cmdutil.LoadProject(cmd.Context(), ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Determine target phase
	if phaseName == "" {
		phaseName = getActivePhase(proj)
		if phaseName == "" {
			return fmt.Errorf("could not determine active phase")
		}
	}

	// Get phase from map
	phaseState, exists := proj.Phases[phaseName]
	if !exists {
		return fmt.Errorf("phase not found: %s", phaseName)
	}

	// Create artifact
	artifact := project.ArtifactState{
		Type:       artifactType,
		Path:       path,
		Approved:   approved,
		Created_at: time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	// Add to outputs
	phaseState.Outputs = append(phaseState.Outputs, artifact)

	// Update phase back in map
	proj.Phases[phaseName] = phaseState

	// Save project state
	if err := proj.Save(cmd.Context()); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Added output artifact [%d] to phase %s\n", len(phaseState.Outputs)-1, phaseName)
	return nil
}

// runOutputSet implements the output set command logic.
func runOutputSet(cmd *cobra.Command, args []string, phaseName string, index int) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project
	proj, err := cmdutil.LoadProject(cmd.Context(), ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Determine target phase
	if phaseName == "" {
		phaseName = getActivePhase(proj)
		if phaseName == "" {
			return fmt.Errorf("could not determine active phase")
		}
	}

	// Get phase from map
	phaseState, exists := proj.Phases[phaseName]
	if !exists {
		return fmt.Errorf("phase not found: %s", phaseName)
	}

	// Get field path and value from args
	fieldPath := args[0]
	value := args[1]

	// Wrap artifacts in state.Artifact for field path mutation
	artifacts := make([]state.Artifact, len(phaseState.Outputs))
	for i, a := range phaseState.Outputs {
		artifacts[i] = state.Artifact{ArtifactState: a}
	}

	// Set field using artifact helper
	if err := cmdutil.SetArtifactField(&artifacts, index, fieldPath, value); err != nil {
		return fmt.Errorf("failed to set field: %w", err)
	}

	// Unwrap back to ArtifactState
	for i, a := range artifacts {
		phaseState.Outputs[i] = a.ArtifactState
	}

	// Update phase back in map
	proj.Phases[phaseName] = phaseState

	// Save project state
	if err := proj.Save(cmd.Context()); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Set %s on output artifact [%d] in phase %s\n", fieldPath, index, phaseName)
	return nil
}

// runOutputRemove implements the output remove command logic.
func runOutputRemove(cmd *cobra.Command, phaseName string, index int) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project
	proj, err := cmdutil.LoadProject(cmd.Context(), ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Determine target phase
	if phaseName == "" {
		phaseName = getActivePhase(proj)
		if phaseName == "" {
			return fmt.Errorf("could not determine active phase")
		}
	}

	// Get phase from map
	phaseState, exists := proj.Phases[phaseName]
	if !exists {
		return fmt.Errorf("phase not found: %s", phaseName)
	}

	// Validate index
	if err := cmdutil.IndexInRange(len(phaseState.Outputs), index); err != nil {
		return err
	}

	// Remove artifact by index
	phaseState.Outputs = append(phaseState.Outputs[:index], phaseState.Outputs[index+1:]...)

	// Update phase back in map
	proj.Phases[phaseName] = phaseState

	// Save project state
	if err := proj.Save(cmd.Context()); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Removed output artifact [%d] from phase %s\n", index, phaseName)
	return nil
}

// runOutputList implements the output list command logic.
func runOutputList(cmd *cobra.Command, phaseName string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return fmt.Errorf("sow not initialized. Run 'sow init' first")
	}

	// Load project
	proj, err := cmdutil.LoadProject(cmd.Context(), ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("no active project found")
		}
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Determine target phase
	if phaseName == "" {
		phaseName = getActivePhase(proj)
		if phaseName == "" {
			return fmt.Errorf("could not determine active phase")
		}
	}

	// Get phase from map
	phaseState, exists := proj.Phases[phaseName]
	if !exists {
		return fmt.Errorf("phase not found: %s", phaseName)
	}

	// Convert to state.Artifact for formatting
	artifacts := make([]state.Artifact, len(phaseState.Outputs))
	for i, a := range phaseState.Outputs {
		artifacts[i] = state.Artifact{ArtifactState: a}
	}

	// Format and display
	output := cmdutil.FormatArtifactList(artifacts)
	fmt.Println(output)

	return nil
}
