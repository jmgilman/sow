package cmd

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sow"
	"fmt"

	"github.com/spf13/cobra"
)

// NewValidateCmd creates the validate command.
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmdutil.GetContext(cmd.Context())

			// Check if sow is initialized
			if !ctx.IsInitialized() {
				return sow.ErrNotInitialized
			}

			repoRoot := ctx.RepoRoot()

			// Validate entire structure
			result, err := sow.Validate(repoRoot)
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			// Check for validation errors
			if result.HasErrors() {
				return fmt.Errorf("validation failed\n%s", result.Error())
			}

			// If we got here, validation passed
			cmd.Println("✓ All validations passed")
			return nil
		},
	}

	return cmd
}
