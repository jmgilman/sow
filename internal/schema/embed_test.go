package schema

import (
	"testing"
)

// TestEmbeddedSchemasNotEmpty verifies that all schemas are embedded and not empty
func TestEmbeddedSchemasNotEmpty(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{"project-state schema", projectStateCUE},
		{"task-state schema", taskStateCUE},
		{"sink-index schema", sinkIndexCUE},
		{"repo-index schema", repoIndexCUE},
		{"sow-version schema", sowVersionCUE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.schema == "" {
				t.Errorf("%s is empty - go:embed directive may have failed", tt.name)
			}

			// Verify it contains expected CUE package declaration
			if len(tt.schema) < 10 {
				t.Errorf("%s is too short (%d bytes) - may not be valid CUE", tt.name, len(tt.schema))
			}
		})
	}
}

// TestEmbeddedSchemasContainPackage verifies schemas contain CUE package declaration
func TestEmbeddedSchemasContainPackage(t *testing.T) {
	schemas := map[string]string{
		"project-state": projectStateCUE,
		"task-state":    taskStateCUE,
		"sink-index":    sinkIndexCUE,
		"repo-index":    repoIndexCUE,
		"sow-version":   sowVersionCUE,
	}

	for name, schema := range schemas {
		t.Run(name, func(t *testing.T) {
			// Each CUE schema should start with package declaration
			if len(schema) < 7 {
				t.Errorf("schema %s is too short", name)
				return
			}

			// Check if it contains "package" keyword (loose assertion)
			contains := false
			if len(schema) > 0 {
				for i := 0; i < len(schema)-7; i++ {
					if schema[i:i+7] == "package" {
						contains = true
						break
					}
				}
			}

			if !contains {
				t.Errorf("schema %s does not contain 'package' keyword", name)
			}
		})
	}
}

// TestGetSchema verifies the GetSchema function returns correct schemas
func TestGetSchema(t *testing.T) {
	tests := []struct {
		name       string
		schemaType string
		wantEmpty  bool
	}{
		{"project-state exists", "project-state", false},
		{"task-state exists", "task-state", false},
		{"sink-index exists", "sink-index", false},
		{"repo-index exists", "repo-index", false},
		{"sow-version exists", "sow-version", false},
		{"invalid type returns empty", "invalid-type", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSchema(tt.schemaType)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("GetSchema(%s) should return empty string, got %d bytes", tt.schemaType, len(result))
				}
			} else {
				if result == "" {
					t.Errorf("GetSchema(%s) returned empty string", tt.schemaType)
				}
				if len(result) < 10 {
					t.Errorf("GetSchema(%s) returned too short result: %d bytes", tt.schemaType, len(result))
				}
			}
		})
	}
}

// TestListSchemas verifies all schema types are listed
func TestListSchemas(t *testing.T) {
	schemas := ListSchemas()

	expectedCount := 5
	if len(schemas) != expectedCount {
		t.Errorf("ListSchemas() returned %d schemas, expected %d", len(schemas), expectedCount)
	}

	// Verify expected schemas are in the list
	expected := map[string]bool{
		"project-state": false,
		"task-state":    false,
		"sink-index":    false,
		"repo-index":    false,
		"sow-version":   false,
	}

	for _, schema := range schemas {
		if _, ok := expected[schema]; ok {
			expected[schema] = true
		} else {
			t.Errorf("ListSchemas() returned unexpected schema: %s", schema)
		}
	}

	for schema, found := range expected {
		if !found {
			t.Errorf("ListSchemas() missing expected schema: %s", schema)
		}
	}
}
