package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/your-org/sow/internal/schema"
)

// NewSchemaCmd creates the schema command
func NewSchemaCmd() *cobra.Command {
	var schemaName string
	var exportPath string

	cmd := &cobra.Command{
		Use:   "schema",
		Short: "View embedded CUE schemas",
		Long: `View embedded CUE schemas that define .sow/ file formats.

By default, lists all available schemas. Use --name to show a specific schema,
and --export to save a schema to a file.`,
		Example: `  # List all available schemas
  sow schema

  # Show a specific schema
  sow schema --name project-state

  # Export a schema to file
  sow schema --name project-state --export state.cue`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSchema(cmd, schemaName, exportPath)
		},
	}

	cmd.Flags().StringVar(&schemaName, "name", "", "Schema name to show (e.g., project-state, task-state)")
	cmd.Flags().StringVar(&exportPath, "export", "", "Export schema to file (requires --name)")

	return cmd
}

func runSchema(cmd *cobra.Command, schemaName string, exportPath string) error {
	// Validate: --export requires --name
	if exportPath != "" && schemaName == "" {
		return fmt.Errorf("--export requires --name to be specified")
	}

	// If no schema name specified, list all schemas
	if schemaName == "" {
		return listSchemas(cmd)
	}

	// Get the requested schema
	schemaContent := schema.GetSchema(schemaName)
	if schemaContent == "" {
		availableSchemas := schema.ListSchemas()
		return fmt.Errorf("schema %q not found. Available schemas: %s", schemaName, strings.Join(availableSchemas, ", "))
	}

	// If export path specified, write to file
	if exportPath != "" {
		return exportSchema(cmd, schemaName, schemaContent, exportPath)
	}

	// Otherwise, display the schema
	return showSchema(cmd, schemaName, schemaContent)
}

func listSchemas(cmd *cobra.Command) error {
	schemas := schema.ListSchemas()

	fmt.Fprintf(cmd.OutOrStdout(), "Available schemas:\n\n")

	for _, schemaName := range schemas {
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", schemaName)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nUse 'sow schema --name <schema-name>' to view a specific schema\n")

	return nil
}

func showSchema(cmd *cobra.Command, schemaName string, content string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Schema: %s\n\n", schemaName)
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", content)
	return nil
}

func exportSchema(cmd *cobra.Command, schemaName string, content string, exportPath string) error {
	// Create parent directories if needed
	dir := filepath.Dir(exportPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write schema to file
	if err := os.WriteFile(exportPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write schema to %s: %w", exportPath, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Schema %q exported to %s\n", schemaName, exportPath)

	return nil
}
