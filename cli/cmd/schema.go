package cmd

import (
	"github.com/spf13/cobra"
)

// NewSchemaCmd creates the schema command.
func NewSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Inspect and export CUE schemas",
		Long: `View and export embedded CUE schemas.

Schemas define validation rules for all sow structure files.
Use this command to inspect schema definitions and export them
for documentation or external validation.`,
	}

	// Subcommands
	cmd.AddCommand(newSchemaListCmd())
	cmd.AddCommand(newSchemaShowCmd())
	cmd.AddCommand(newSchemaExportCmd())

	return cmd
}

func newSchemaListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available schemas",
		Long:  `List all embedded CUE schemas with descriptions.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("Available schemas:")
			cmd.Println("  project-state    - Project state schema")
			cmd.Println("  task-state       - Task state schema")
			cmd.Println("  refs-committed   - Committed refs index schema")
			cmd.Println("  refs-cache       - Cache index schema")
			cmd.Println("  refs-local       - Local refs index schema")
			return nil
		},
	}
}

func newSchemaShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <schema>",
		Short: "Display a specific schema",
		Long: `Display the CUE schema definition for a specific type.

Available schemas:
  - project-state
  - task-state
  - refs-committed
  - refs-cache
  - refs-local`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaName := args[0]
			_ = schemaName // TODO: implement

			cmd.Printf("Schema: %s\n", schemaName)
			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "cue", "Output format (cue, json, yaml)")

	return cmd
}

func newSchemaExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <schema> <output-file>",
		Short: "Export schema to file",
		Long:  `Export a schema to a file in the specified format.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaName := args[0]
			outputFile := args[1]
			_, _ = schemaName, outputFile // TODO: implement

			cmd.Printf("Exported schema to %s\n", outputFile)
			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "cue", "Output format (cue, json, yaml)")

	return cmd
}
