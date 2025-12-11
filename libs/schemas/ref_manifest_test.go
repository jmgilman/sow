package schemas

import (
	"encoding/json"
	"fmt"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

// Helper function to load and validate data against a schema definition.
func validateRefManifestSchema(t *testing.T, schemaPath string, data map[string]any) error {
	t.Helper()

	ctx := cuecontext.New()

	// Load the schemas package
	instances := load.Instances([]string{"."}, &load.Config{
		Dir: ".",
	})
	if len(instances) == 0 {
		t.Fatal("failed to load schema instances")
	}
	if instances[0].Err != nil {
		t.Fatalf("failed to load schema: %v", instances[0].Err)
	}

	schema := ctx.BuildInstance(instances[0])
	if schema.Err() != nil {
		t.Fatalf("failed to build schema: %v", schema.Err())
	}

	// Lookup the specific schema
	schemaType := schema.LookupPath(cue.ParsePath(schemaPath))
	if schemaType.Err() != nil {
		t.Fatalf("failed to lookup %s: %v", schemaPath, schemaType.Err())
	}

	// Convert test data to CUE value
	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	dataValue := ctx.CompileBytes(dataJSON)
	if dataValue.Err() != nil {
		t.Fatalf("failed to compile test data: %v", dataValue.Err())
	}

	// Validate
	unified := schemaType.Unify(dataValue)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}

// ============================================================
// RefManifest Schema Loading Tests
// ============================================================

func TestRefManifestSchemaExists(t *testing.T) {
	ctx := cuecontext.New()

	instances := load.Instances([]string{"."}, &load.Config{
		Dir: ".",
	})
	if len(instances) == 0 {
		t.Fatal("failed to load schema instances")
	}
	if instances[0].Err != nil {
		t.Fatalf("failed to load schema: %v", instances[0].Err)
	}

	schema := ctx.BuildInstance(instances[0])
	if schema.Err() != nil {
		t.Fatalf("failed to build schema: %v", schema.Err())
	}

	// Verify #RefManifest can be looked up
	refManifest := schema.LookupPath(cue.ParsePath("#RefManifest"))
	if refManifest.Err() != nil {
		t.Fatalf("failed to lookup #RefManifest: %v", refManifest.Err())
	}
}

func TestRefManifestSchemaTypesExist(t *testing.T) {
	ctx := cuecontext.New()

	instances := load.Instances([]string{"."}, &load.Config{
		Dir: ".",
	})
	if len(instances) == 0 {
		t.Fatal("failed to load schema instances")
	}
	if instances[0].Err != nil {
		t.Fatalf("failed to load schema: %v", instances[0].Err)
	}

	schema := ctx.BuildInstance(instances[0])
	if schema.Err() != nil {
		t.Fatalf("failed to build schema: %v", schema.Err())
	}

	// Verify all expected types can be looked up
	types := []string{
		"#RefManifest",
		"#RefIdentification",
		"#RefContent",
		"#RefClassification",
		"#ClassificationType",
		"#RefProvenance",
		"#RefPackaging",
		"#RefHints",
	}

	for _, typeName := range types {
		t.Run(typeName, func(t *testing.T) {
			schemaType := schema.LookupPath(cue.ParsePath(typeName))
			if schemaType.Err() != nil {
				t.Errorf("failed to lookup %s: %v", typeName, schemaType.Err())
			}
		})
	}
}

// ============================================================
// RefManifest Validation Tests - Valid Cases
// ============================================================

func TestValidRefManifest_Minimal(t *testing.T) {
	data := map[string]any{
		"schema_version": "1.0.0",
		"ref": map[string]any{
			"title": "Go Team Standards",
			"link":  "go-standards",
		},
		"content": map[string]any{
			"description": "Team Go coding conventions.",
			"classifications": []any{
				map[string]any{
					"type": "guidelines",
				},
			},
			"tags": []any{"golang"},
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err != nil {
		t.Errorf("valid minimal manifest should pass validation: %v", err)
	}
}

func TestValidRefManifest_Full(t *testing.T) {
	data := map[string]any{
		"schema_version": "1.0.0",
		"ref": map[string]any{
			"title": "Go Team Standards",
			"link":  "go-standards",
		},
		"content": map[string]any{
			"description": "Team Go coding conventions and best practices.",
			"summary":     "Complete reference for Go development.",
			"classifications": []any{
				map[string]any{
					"type":        "guidelines",
					"description": "Go coding standards",
				},
				map[string]any{
					"type": "code-examples",
				},
			},
			"tags": []any{"golang", "conventions", "testing"},
		},
		"provenance": map[string]any{
			"authors":  []any{"Platform Team"},
			"created":  "2024-01-15T10:00:00Z",
			"updated":  "2025-01-30T15:30:00Z",
			"source":   "https://github.com/myorg/team-docs",
			"license":  "MIT",
		},
		"packaging": map[string]any{
			"exclude": []any{"*.draft.md", ".DS_Store"},
		},
		"hints": map[string]any{
			"suggested_queries": []any{"How should I structure error handling?"},
			"primary_files":     []any{"README.md", "standards/golang.md"},
		},
		"metadata": map[string]any{
			"team":     "platform",
			"audience": "all-engineers",
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err != nil {
		t.Errorf("valid full manifest should pass validation: %v", err)
	}
}

// ============================================================
// schema_version Tests
// ============================================================

func TestValidSchemaVersion(t *testing.T) {
	validVersions := []string{
		"1.0.0",
		"0.1.0",
		"2.10.100",
		"0.0.1",
	}

	for _, version := range validVersions {
		t.Run(version, func(t *testing.T) {
			data := map[string]any{
				"schema_version": version,
				"ref": map[string]any{
					"title": "Test",
					"link":  "test-ref",
				},
				"content": map[string]any{
					"description":     "Test description.",
					"classifications": []any{map[string]any{"type": "reference"}},
					"tags":            []any{"test"},
				},
			}

			err := validateRefManifestSchema(t, "#RefManifest", data)
			if err != nil {
				t.Errorf("schema_version '%s' should pass validation: %v", version, err)
			}
		})
	}
}

func TestInvalidSchemaVersion_NotSemver(t *testing.T) {
	data := map[string]any{
		"schema_version": "1.0",
		"ref": map[string]any{
			"title": "Test",
			"link":  "test-ref",
		},
		"content": map[string]any{
			"description":     "Test description.",
			"classifications": []any{map[string]any{"type": "reference"}},
			"tags":            []any{"test"},
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err == nil {
		t.Error("schema_version '1.0' (not semver) should fail validation")
	}
}

func TestInvalidSchemaVersion_Empty(t *testing.T) {
	data := map[string]any{
		"schema_version": "",
		"ref": map[string]any{
			"title": "Test",
			"link":  "test-ref",
		},
		"content": map[string]any{
			"description":     "Test description.",
			"classifications": []any{map[string]any{"type": "reference"}},
			"tags":            []any{"test"},
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err == nil {
		t.Error("empty schema_version should fail validation")
	}
}

// ============================================================
// RefIdentification Tests
// ============================================================

func TestValidRefIdentification(t *testing.T) {
	data := map[string]any{
		"title": "Go Standards",
		"link":  "go-standards",
	}

	err := validateRefManifestSchema(t, "#RefIdentification", data)
	if err != nil {
		t.Errorf("valid ref identification should pass: %v", err)
	}
}

func TestValidLink_KebabCase(t *testing.T) {
	validLinks := []string{
		"go-standards",
		"my-ref",
		"api-patterns",
		"a1",
		"test123",
		"my-very-long-link-name",
	}

	for _, link := range validLinks {
		t.Run(link, func(t *testing.T) {
			data := map[string]any{
				"title": "Test",
				"link":  link,
			}

			err := validateRefManifestSchema(t, "#RefIdentification", data)
			if err != nil {
				t.Errorf("link '%s' should pass validation: %v", link, err)
			}
		})
	}
}

func TestInvalidLink_Uppercase(t *testing.T) {
	data := map[string]any{
		"title": "Test",
		"link":  "Go-Standards",
	}

	err := validateRefManifestSchema(t, "#RefIdentification", data)
	if err == nil {
		t.Error("link with uppercase should fail validation")
	}
}

func TestInvalidLink_StartsWithHyphen(t *testing.T) {
	data := map[string]any{
		"title": "Test",
		"link":  "-go-standards",
	}

	err := validateRefManifestSchema(t, "#RefIdentification", data)
	if err == nil {
		t.Error("link starting with hyphen should fail validation")
	}
}

func TestInvalidLink_EndsWithHyphen(t *testing.T) {
	data := map[string]any{
		"title": "Test",
		"link":  "go-standards-",
	}

	err := validateRefManifestSchema(t, "#RefIdentification", data)
	if err == nil {
		t.Error("link ending with hyphen should fail validation")
	}
}

func TestInvalidLink_Empty(t *testing.T) {
	data := map[string]any{
		"title": "Test",
		"link":  "",
	}

	err := validateRefManifestSchema(t, "#RefIdentification", data)
	if err == nil {
		t.Error("empty link should fail validation")
	}
}

func TestInvalidTitle_Empty(t *testing.T) {
	data := map[string]any{
		"title": "",
		"link":  "test",
	}

	err := validateRefManifestSchema(t, "#RefIdentification", data)
	if err == nil {
		t.Error("empty title should fail validation")
	}
}

// ============================================================
// RefContent Tests
// ============================================================

func TestValidRefContent(t *testing.T) {
	data := map[string]any{
		"description": "A test description.",
		"classifications": []any{
			map[string]any{"type": "guidelines"},
		},
		"tags": []any{"test"},
	}

	err := validateRefManifestSchema(t, "#RefContent", data)
	if err != nil {
		t.Errorf("valid ref content should pass: %v", err)
	}
}

func TestValidRefContent_WithSummary(t *testing.T) {
	data := map[string]any{
		"description": "A test description.",
		"summary":     "A longer summary of the content.",
		"classifications": []any{
			map[string]any{"type": "guidelines"},
		},
		"tags": []any{"test"},
	}

	err := validateRefManifestSchema(t, "#RefContent", data)
	if err != nil {
		t.Errorf("ref content with summary should pass: %v", err)
	}
}

func TestInvalidRefContent_EmptyDescription(t *testing.T) {
	data := map[string]any{
		"description": "",
		"classifications": []any{
			map[string]any{"type": "guidelines"},
		},
		"tags": []any{"test"},
	}

	err := validateRefManifestSchema(t, "#RefContent", data)
	if err == nil {
		t.Error("empty description should fail validation")
	}
}

func TestInvalidRefContent_EmptyClassifications(t *testing.T) {
	data := map[string]any{
		"description":     "Test description.",
		"classifications": []any{},
		"tags":            []any{"test"},
	}

	err := validateRefManifestSchema(t, "#RefContent", data)
	if err == nil {
		t.Error("empty classifications array should fail validation")
	}
}

func TestInvalidRefContent_EmptyTags(t *testing.T) {
	data := map[string]any{
		"description": "Test description.",
		"classifications": []any{
			map[string]any{"type": "guidelines"},
		},
		"tags": []any{},
	}

	err := validateRefManifestSchema(t, "#RefContent", data)
	if err == nil {
		t.Error("empty tags array should fail validation")
	}
}

// ============================================================
// ClassificationType Tests
// ============================================================

func TestValidClassificationTypes(t *testing.T) {
	validTypes := []string{
		"tutorial",
		"api-reference",
		"guidelines",
		"architecture",
		"runbook",
		"specification",
		"reference",
		"code-examples",
		"code-templates",
		"code-library",
		"uncategorized",
	}

	for _, classType := range validTypes {
		t.Run(classType, func(t *testing.T) {
			data := map[string]any{
				"description": "Test.",
				"classifications": []any{
					map[string]any{"type": classType},
				},
				"tags": []any{"test"},
			}

			err := validateRefManifestSchema(t, "#RefContent", data)
			if err != nil {
				t.Errorf("classification type '%s' should pass validation: %v", classType, err)
			}
		})
	}
}

func TestInvalidClassificationType_Typo(t *testing.T) {
	data := map[string]any{
		"description": "Test.",
		"classifications": []any{
			map[string]any{"type": "guidlines"}, // Typo
		},
		"tags": []any{"test"},
	}

	err := validateRefManifestSchema(t, "#RefContent", data)
	if err == nil {
		t.Error("misspelled classification type should fail validation")
	}
}

func TestInvalidClassificationType_Unknown(t *testing.T) {
	data := map[string]any{
		"description": "Test.",
		"classifications": []any{
			map[string]any{"type": "custom-type"},
		},
		"tags": []any{"test"},
	}

	err := validateRefManifestSchema(t, "#RefContent", data)
	if err == nil {
		t.Error("unknown classification type should fail validation")
	}
}

// ============================================================
// RefProvenance Tests
// ============================================================

func TestValidRefProvenance_Minimal(t *testing.T) {
	// All fields are optional, so empty struct should pass
	data := map[string]any{}

	err := validateRefManifestSchema(t, "#RefProvenance", data)
	if err != nil {
		t.Errorf("empty provenance should pass: %v", err)
	}
}

func TestValidRefProvenance_Full(t *testing.T) {
	data := map[string]any{
		"authors":  []any{"Platform Team", "DevOps Team"},
		"created":  "2024-01-15T10:00:00Z",
		"updated":  "2025-01-30T15:30:00Z",
		"source":   "https://github.com/myorg/team-docs",
		"license":  "MIT",
	}

	err := validateRefManifestSchema(t, "#RefProvenance", data)
	if err != nil {
		t.Errorf("full provenance should pass: %v", err)
	}
}

// ============================================================
// RefPackaging Tests
// ============================================================

func TestValidRefPackaging_Minimal(t *testing.T) {
	data := map[string]any{}

	err := validateRefManifestSchema(t, "#RefPackaging", data)
	if err != nil {
		t.Errorf("empty packaging should pass: %v", err)
	}
}

func TestValidRefPackaging_WithExclude(t *testing.T) {
	data := map[string]any{
		"exclude": []any{"*.draft.md", ".DS_Store", "tmp/"},
	}

	err := validateRefManifestSchema(t, "#RefPackaging", data)
	if err != nil {
		t.Errorf("packaging with exclude should pass: %v", err)
	}
}

// ============================================================
// RefHints Tests
// ============================================================

func TestValidRefHints_Minimal(t *testing.T) {
	data := map[string]any{}

	err := validateRefManifestSchema(t, "#RefHints", data)
	if err != nil {
		t.Errorf("empty hints should pass: %v", err)
	}
}

func TestValidRefHints_Full(t *testing.T) {
	data := map[string]any{
		"suggested_queries": []any{
			"How should I structure error handling?",
			"What are the testing conventions?",
		},
		"primary_files": []any{
			"README.md",
			"standards/golang.md",
		},
	}

	err := validateRefManifestSchema(t, "#RefHints", data)
	if err != nil {
		t.Errorf("full hints should pass: %v", err)
	}
}

// ============================================================
// RefManifest Missing Required Fields Tests
// ============================================================

func TestInvalidRefManifest_MissingSchemaVersion(t *testing.T) {
	data := map[string]any{
		"ref": map[string]any{
			"title": "Test",
			"link":  "test-ref",
		},
		"content": map[string]any{
			"description":     "Test description.",
			"classifications": []any{map[string]any{"type": "reference"}},
			"tags":            []any{"test"},
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err == nil {
		t.Error("missing schema_version should fail validation")
	}
}

func TestInvalidRefManifest_MissingRef(t *testing.T) {
	data := map[string]any{
		"schema_version": "1.0.0",
		"content": map[string]any{
			"description":     "Test description.",
			"classifications": []any{map[string]any{"type": "reference"}},
			"tags":            []any{"test"},
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err == nil {
		t.Error("missing ref should fail validation")
	}
}

func TestInvalidRefManifest_MissingContent(t *testing.T) {
	data := map[string]any{
		"schema_version": "1.0.0",
		"ref": map[string]any{
			"title": "Test",
			"link":  "test-ref",
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err == nil {
		t.Error("missing content should fail validation")
	}
}

func TestInvalidRefManifest_MissingRefTitle(t *testing.T) {
	data := map[string]any{
		"schema_version": "1.0.0",
		"ref": map[string]any{
			"link": "test-ref",
		},
		"content": map[string]any{
			"description":     "Test description.",
			"classifications": []any{map[string]any{"type": "reference"}},
			"tags":            []any{"test"},
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err == nil {
		t.Error("missing ref.title should fail validation")
	}
}

func TestInvalidRefManifest_MissingRefLink(t *testing.T) {
	data := map[string]any{
		"schema_version": "1.0.0",
		"ref": map[string]any{
			"title": "Test",
		},
		"content": map[string]any{
			"description":     "Test description.",
			"classifications": []any{map[string]any{"type": "reference"}},
			"tags":            []any{"test"},
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err == nil {
		t.Error("missing ref.link should fail validation")
	}
}

// ============================================================
// Metadata Tests (freeform)
// ============================================================

func TestValidMetadata_NestedStructure(t *testing.T) {
	data := map[string]any{
		"schema_version": "1.0.0",
		"ref": map[string]any{
			"title": "Test",
			"link":  "test-ref",
		},
		"content": map[string]any{
			"description":     "Test.",
			"classifications": []any{map[string]any{"type": "reference"}},
			"tags":            []any{"test"},
		},
		"metadata": map[string]any{
			"team":      "platform",
			"priority":  1,
			"tags":      []any{"internal", "approved"},
			"nested": map[string]any{
				"level1": map[string]any{
					"level2": "deep value",
				},
			},
		},
	}

	err := validateRefManifestSchema(t, "#RefManifest", data)
	if err != nil {
		t.Errorf("nested metadata should pass: %v", err)
	}
}
