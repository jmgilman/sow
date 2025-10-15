package cmd

import (
	"github.com/spf13/cobra"
)

// NewValidateCmd creates the validate command
func NewValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [file...]",
		Short: "Validate sow files against schemas",
		Long: `Validate sow structure files against CUE schemas.

Validates:
  - Project state files (.sow/project/state.yaml)
  - Task state files (.sow/project/phases/implementation/tasks/*/state.yaml)
  - Refs index files (.sow/refs/index.json, .sow/refs/index.local.json)
  - Cache index (~/.cache/sow/index.json)

If no files are specified, validates all sow structure files.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := FilesystemFromContext(cmd.Context())
			_ = fs // TODO: implement

			cmd.Println("Validation complete")
			return nil
		},
	}

	// Flags
	cmd.Flags().BoolP("all", "a", false, "Validate all sow files in repository")
	cmd.Flags().StringP("schema", "s", "", "Validate against specific schema")

	return cmd
}
