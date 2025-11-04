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

// NewInputCmd creates the input command.
func NewInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Short: "Manage phase input artifacts",
		Long: `Manage phase input artifacts.

Input commands allow you to add, modify, remove, and list input artifacts
for project phases. Artifacts are managed using zero-based indices.`,
	}

	cmd.AddCommand(newInputAddCmd())
	cmd.AddCommand(newInputSetCmd())
	cmd.AddCommand(newInputRemoveCmd())
	cmd.AddCommand(newInputListCmd())

	return cmd
}

// newInputAddCmd creates the input add subcommand.
func newInputAddCmd() *cobra.Command {
	var phaseName, artifactType, path string
	var approved bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add input artifact to phase",
		Long: `Add an input artifact to a phase.

Creates a new input artifact with the specified type and path.
The artifact is appended to the phase's input list.

Examples:
  # Add context artifact to planning phase
  sow input add --type context --path context/research.md --phase planning

  # Add with approval flag
  sow input add --type context --path context/doc.md --approved --phase planning

  # Use active phase (defaults to current state)
  sow input add --type context --path context/design.md`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInputAdd(cmd, phaseName, artifactType, path, approved)
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

// newInputSetCmd creates the input set subcommand.
func newInputSetCmd() *cobra.Command {
	var phaseName string
	var index int

	cmd := &cobra.Command{
		Use:   "set <field-path> <value>",
		Short: "Set field on input artifact by index",
		Long: `Set a field on an input artifact using its index.

Supports both direct fields and metadata fields using dot notation:
  - Direct fields: type, path, approved
  - Metadata fields: metadata.* (any custom field)

Examples:
  # Set approved field
  sow input set --index 0 approved true --phase planning

  # Set metadata field
  sow input set --index 0 metadata.source external --phase planning

  # Set nested metadata
  sow input set --index 0 metadata.reviewer.name orchestrator --phase planning`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInputSet(cmd, args, phaseName, index)
		},
	}

	cmd.Flags().StringVarP(&phaseName, "phase", "p", "", "Target phase (defaults to active phase)")
	cmd.Flags().IntVar(&index, "index", -1, "Artifact index (required)")

	_ = cmd.MarkFlagRequired("index")

	return cmd
}

// newInputRemoveCmd creates the input remove subcommand.
func newInputRemoveCmd() *cobra.Command {
	var phaseName string
	var index int

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove input artifact by index",
		Long: `Remove an input artifact from a phase by its index.

The artifact at the specified index is removed from the phase's input list.
Indices of subsequent artifacts are shifted down.

Examples:
  # Remove first input
  sow input remove --index 0 --phase planning

  # Remove from active phase
  sow input remove --index 1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInputRemove(cmd, phaseName, index)
		},
	}

	cmd.Flags().StringVarP(&phaseName, "phase", "p", "", "Target phase (defaults to active phase)")
	cmd.Flags().IntVar(&index, "index", -1, "Artifact index (required)")

	_ = cmd.MarkFlagRequired("index")

	return cmd
}

// newInputListCmd creates the input list subcommand.
func newInputListCmd() *cobra.Command {
	var phaseName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List input artifacts for phase",
		Long: `List all input artifacts for a phase.

Displays artifacts with their indices and details including type, path,
approval status, and metadata.

Examples:
  # List inputs for planning phase
  sow input list --phase planning

  # List inputs for active phase
  sow input list`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInputList(cmd, phaseName)
		},
	}

	cmd.Flags().StringVarP(&phaseName, "phase", "p", "", "Target phase (defaults to active phase)")

	return cmd
}

// runInputAdd implements the input add command logic.
func runInputAdd(cmd *cobra.Command, phaseName, artifactType, path string, approved bool) error {
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

	// Add to inputs
	phaseState.Inputs = append(phaseState.Inputs, artifact)

	// Update phase back in map
	proj.Phases[phaseName] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Added input artifact [%d] to phase %s\n", len(phaseState.Inputs)-1, phaseName)
	return nil
}

// runInputSet implements the input set command logic.
func runInputSet(cmd *cobra.Command, args []string, phaseName string, index int) error {
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
	artifacts := make([]state.Artifact, len(phaseState.Inputs))
	for i, a := range phaseState.Inputs {
		artifacts[i] = state.Artifact{ArtifactState: a}
	}

	// Set field using artifact helper
	if err := cmdutil.SetArtifactField(&artifacts, index, fieldPath, value); err != nil {
		return fmt.Errorf("failed to set field: %w", err)
	}

	// Unwrap back to ArtifactState
	for i, a := range artifacts {
		phaseState.Inputs[i] = a.ArtifactState
	}

	// Update phase back in map
	proj.Phases[phaseName] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Set %s on input artifact [%d] in phase %s\n", fieldPath, index, phaseName)
	return nil
}

// runInputRemove implements the input remove command logic.
func runInputRemove(cmd *cobra.Command, phaseName string, index int) error {
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
	if err := cmdutil.IndexInRange(len(phaseState.Inputs), index); err != nil {
		return err
	}

	// Remove artifact by index
	phaseState.Inputs = append(phaseState.Inputs[:index], phaseState.Inputs[index+1:]...)

	// Update phase back in map
	proj.Phases[phaseName] = phaseState

	// Save project state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("Removed input artifact [%d] from phase %s\n", index, phaseName)
	return nil
}

// runInputList implements the input list command logic.
func runInputList(cmd *cobra.Command, phaseName string) error {
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
	artifacts := make([]state.Artifact, len(phaseState.Inputs))
	for i, a := range phaseState.Inputs {
		artifacts[i] = state.Artifact{ArtifactState: a}
	}

	// Format and display
	output := cmdutil.FormatArtifactList(artifacts)
	fmt.Println(output)

	return nil
}
