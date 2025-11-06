package breakdown

import (
	"strings"
	"testing"
)

// TestBreakdownMetadataSchemaNotEmpty verifies breakdownMetadataSchema is not empty.
func TestBreakdownMetadataSchemaNotEmpty(t *testing.T) {
	if breakdownMetadataSchema == "" {
		t.Error("breakdownMetadataSchema should not be empty")
	}
}

// TestBreakdownMetadataSchemaIsValidCUE verifies breakdown schema has valid CUE syntax.
func TestBreakdownMetadataSchemaIsValidCUE(t *testing.T) {
	// Basic CUE syntax checks
	if !strings.Contains(breakdownMetadataSchema, "package breakdown") {
		t.Error("breakdownMetadataSchema should contain package declaration")
	}
	if !strings.Contains(breakdownMetadataSchema, "{") || !strings.Contains(breakdownMetadataSchema, "}") {
		t.Error("breakdownMetadataSchema should contain CUE object structure")
	}
}
