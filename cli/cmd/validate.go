package cmd

import (
	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/spf13/cobra"
)

// NewValidateCmd creates the validate command
func NewValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate entire .sow directory structure",
		Long: `Validate all files in the .sow directory structure against their CUE schemas.

This command performs comprehensive validation of:
  - .sow/refs/index.json (committed refs index)
  - .sow/refs/index.local.json (local refs index, if exists)
  - .sow/project/state.yaml (project state, if exists)
  - .sow/project/phases/implementation/tasks/*/state.yaml (all tasks, if exist)

All validation errors are collected and reported together (not fail-fast).
Optional components are only validated if they exist.

Exit codes:
  0 - All validations passed
  1 - Validation errors found
  2 - Failed to initialize (not in sow repository)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Attempt to load .sow structure (validation happens during construction)
			sowFS, err := sowfs.NewSowFS()
			if err != nil {
				return err
			}
			defer sowFS.Close()

			// If we got here, validation passed
			cmd.Println("✓ All validations passed")
			return nil
		},
	}

	return cmd
}
