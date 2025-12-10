package project

import (
	"fmt"
	"os"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <field-path> <value>",
		Short: "Set project field value",
		Long: `Set a project field using dot notation.

Supports both direct fields and metadata fields:
  - Direct fields: description, name, branch
  - Metadata fields: metadata.key or metadata.nested.key

Examples:
  sow project set description "Updated description"
  sow project set metadata.custom_field custom_value
  sow project set metadata.priority high
  sow project set metadata.score 42`,
		Args: cobra.ExactArgs(2),
		RunE: runSet,
	}
}

func runSet(cmd *cobra.Command, args []string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Load project
	proj, err := cmdutil.LoadProject(cmd.Context(), ctx)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	fieldPath := args[0]
	value := args[1]

	// Use field path parser from Task 010
	// SetField works on the embedded ProjectState
	if err := cmdutil.SetField(&proj.ProjectState, fieldPath, value); err != nil {
		return fmt.Errorf("failed to set field: %w", err)
	}

	// Save project
	if err := proj.Save(cmd.Context()); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Set %s = %s\n", fieldPath, value)
	return nil
}
