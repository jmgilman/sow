package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-org/sow/internal/schema"
)

// TestNewSchemaCmd verifies schema command can be created
func TestNewSchemaCmd(t *testing.T) {
	cmd := NewSchemaCmd()

	if cmd == nil {
		t.Fatal("NewSchemaCmd() returned nil")
	}

	if cmd.Use != "schema" {
		t.Errorf("Schema command Use = %q, want %q", cmd.Use, "schema")
	}

	if cmd.Run == nil && cmd.RunE == nil {
		t.Error("Schema command has no Run or RunE function")
	}
}

// TestSchemaCmdHasFlags verifies required flags exist
func TestSchemaCmdHasFlags(t *testing.T) {
	cmd := NewSchemaCmd()

	flags := []string{"name", "export"}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Schema command missing --%s flag", flagName)
		}
	}
}

// TestSchemaCmdHasHelpText verifies help text is present
func TestSchemaCmdHasHelpText(t *testing.T) {
	cmd := NewSchemaCmd()

	if cmd.Short == "" {
		t.Error("Schema command has no Short description")
	}

	if cmd.Long == "" {
		t.Error("Schema command has no Long description")
	}

	if cmd.Example == "" {
		t.Error("Schema command has no Example text")
	}
}

// TestSchemaCmdListsSchemas verifies default behavior lists schemas
func TestSchemaCmdListsSchemas(t *testing.T) {
	cmd := NewSchemaCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run without args (should list all schemas)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	output := buf.String()

	// Verify output contains all schema names
	schemas := schema.ListSchemas()
	for _, schemaName := range schemas {
		if !strings.Contains(output, schemaName) {
			t.Errorf("Output missing schema %q: %s", schemaName, output)
		}
	}
}

// TestSchemaCmdShowsSpecificSchema verifies --name flag shows schema
func TestSchemaCmdShowsSpecificSchema(t *testing.T) {
	cmd := NewSchemaCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run with --name flag
	cmd.SetArgs([]string{"--name", "project-state"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("schema command with --name failed: %v", err)
	}

	output := buf.String()

	// Verify output contains schema content
	expectedSchema := schema.GetSchema("project-state")
	if !strings.Contains(output, expectedSchema) && expectedSchema != "" {
		t.Error("Output does not contain schema content")
	}

	// At minimum, should show something about the schema
	if output == "" {
		t.Error("No output when showing specific schema")
	}
}

// TestSchemaCmdExportsSchema verifies --export flag writes to file
func TestSchemaCmdExportsSchema(t *testing.T) {
	tmpDir := t.TempDir()
	exportFile := filepath.Join(tmpDir, "test-schema.cue")

	cmd := NewSchemaCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run with --name and --export flags
	cmd.SetArgs([]string{"--name", "project-state", "--export", exportFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("schema command with --export failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Verify file contains schema content
	content, err := os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	expectedSchema := schema.GetSchema("project-state")
	if string(content) != expectedSchema {
		t.Errorf("Export file content does not match schema.\nGot: %s\nWant: %s", string(content), expectedSchema)
	}

	// Verify confirmation message
	output := buf.String()
	if !strings.Contains(output, "exported") && !strings.Contains(output, "written") && !strings.Contains(output, "saved") {
		t.Errorf("Output missing export confirmation: %s", output)
	}
}

// TestSchemaCmdExportRequiresName verifies --export requires --name
func TestSchemaCmdExportRequiresName(t *testing.T) {
	tmpDir := t.TempDir()
	exportFile := filepath.Join(tmpDir, "test-schema.cue")

	cmd := NewSchemaCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run with --export but no --name
	cmd.SetArgs([]string{"--export", exportFile})

	err := cmd.Execute()
	// Should error or warn
	if err == nil {
		output := buf.String()
		if !strings.Contains(strings.ToLower(output), "require") && !strings.Contains(strings.ToLower(output), "must") {
			t.Error("Command should require --name when using --export")
		}
	}
}

// TestSchemaCmdInvalidSchemaName verifies error on invalid schema
func TestSchemaCmdInvalidSchemaName(t *testing.T) {
	cmd := NewSchemaCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run with invalid schema name
	cmd.SetArgs([]string{"--name", "invalid-schema-name"})

	err := cmd.Execute()
	// Should error or show helpful message
	if err == nil {
		output := buf.String()
		if !strings.Contains(strings.ToLower(output), "not found") && !strings.Contains(strings.ToLower(output), "invalid") && !strings.Contains(strings.ToLower(output), "unknown") {
			t.Error("Command should handle invalid schema name gracefully")
		}
	}
}

// TestSchemaCmdListOutputFormat verifies list output is user-friendly
func TestSchemaCmdListOutputFormat(t *testing.T) {
	cmd := NewSchemaCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	output := buf.String()

	// Verify output has some formatting (not just raw schema names)
	if !strings.Contains(output, "schema") || len(output) < 50 {
		t.Error("List output appears to lack user-friendly formatting")
	}
}

// TestSchemaCmdShowOutputFormat verifies show output is formatted
func TestSchemaCmdShowOutputFormat(t *testing.T) {
	cmd := NewSchemaCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"--name", "sow-version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	output := buf.String()

	// Should show the schema content
	if len(output) < 10 {
		t.Error("Show output appears empty or minimal")
	}

	// Should contain actual schema content
	expectedContent := schema.GetSchema("sow-version")
	if expectedContent != "" && !strings.Contains(output, expectedContent) {
		t.Error("Show output does not contain schema content")
	}
}

// TestSchemaCmdAllSchemas verifies all schemas can be retrieved
func TestSchemaCmdAllSchemas(t *testing.T) {
	schemas := schema.ListSchemas()

	for _, schemaName := range schemas {
		cmd := NewSchemaCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		cmd.SetArgs([]string{"--name", schemaName})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Failed to show schema %q: %v", schemaName, err)
		}

		output := buf.String()
		if output == "" {
			t.Errorf("No output for schema %q", schemaName)
		}
	}
}

// TestSchemaCmdExportCreatesDirectory verifies export creates parent dirs
func TestSchemaCmdExportCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	exportFile := filepath.Join(tmpDir, "subdir", "nested", "schema.cue")

	cmd := NewSchemaCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"--name", "task-state", "--export", exportFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("schema command with nested export path failed: %v", err)
	}

	// Verify file exists (parent directories were created)
	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Error("Export file was not created with nested path")
	}
}
