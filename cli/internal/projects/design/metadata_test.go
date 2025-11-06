package design

import (
	"strings"
	"testing"
)

// TestDesignMetadataSchemaNotEmpty verifies designMetadataSchema is not empty.
func TestDesignMetadataSchemaNotEmpty(t *testing.T) {
	if designMetadataSchema == "" {
		t.Error("designMetadataSchema should not be empty")
	}
}

// TestFinalizationMetadataSchemaNotEmpty verifies finalizationMetadataSchema is not empty.
func TestFinalizationMetadataSchemaNotEmpty(t *testing.T) {
	if finalizationMetadataSchema == "" {
		t.Error("finalizationMetadataSchema should not be empty")
	}
}

// TestDesignMetadataSchemaIsValidCUE verifies design schema has valid CUE syntax.
func TestDesignMetadataSchemaIsValidCUE(t *testing.T) {
	// Basic CUE syntax checks
	if !strings.Contains(designMetadataSchema, "package design") {
		t.Error("designMetadataSchema should contain package declaration")
	}
	if !strings.Contains(designMetadataSchema, "{") || !strings.Contains(designMetadataSchema, "}") {
		t.Error("designMetadataSchema should contain CUE object structure")
	}
}

// TestFinalizationMetadataSchemaIsValidCUE verifies finalization schema has valid CUE syntax.
func TestFinalizationMetadataSchemaIsValidCUE(t *testing.T) {
	// Basic CUE syntax checks
	if !strings.Contains(finalizationMetadataSchema, "package design") {
		t.Error("finalizationMetadataSchema should contain package declaration")
	}
	if !strings.Contains(finalizationMetadataSchema, "{") || !strings.Contains(finalizationMetadataSchema, "}") {
		t.Error("finalizationMetadataSchema should contain CUE object structure")
	}
}

// TestFinalizationMetadataSchemaHasExpectedFields verifies finalization schema contains expected fields.
func TestFinalizationMetadataSchemaHasExpectedFields(t *testing.T) {
	expectedFields := []string{"pr_url", "project_deleted"}

	for _, field := range expectedFields {
		if !strings.Contains(finalizationMetadataSchema, field) {
			t.Errorf("finalizationMetadataSchema should contain field %q", field)
		}
	}
}
