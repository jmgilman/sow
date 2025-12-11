package schemas

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validMinimalManifest creates a minimal valid RefManifest for testing.
func validMinimalManifest() *RefManifest {
	return &RefManifest{
		Schema_version: "1.0.0",
		Ref: RefIdentification{
			Title: "Go Team Standards",
			Link:  "go-standards",
		},
		Content: RefContent{
			Description: "Team Go coding conventions.",
			Classifications: []RefClassification{
				{Type: "guidelines"},
			},
			Tags: []string{"golang"},
		},
	}
}

// validFullManifest creates a full valid RefManifest with all optional sections.
func validFullManifest() *RefManifest {
	return &RefManifest{
		Schema_version: "1.0.0",
		Ref: RefIdentification{
			Title: "Go Team Standards",
			Link:  "go-standards",
		},
		Content: RefContent{
			Description: "Team Go coding conventions and best practices.",
			Summary:     "Complete reference for Go development.",
			Classifications: []RefClassification{
				{Type: "guidelines", Description: "Coding standards"},
				{Type: "code-examples"},
			},
			Tags: []string{"golang", "conventions", "testing"},
		},
		Provenance: RefProvenance{
			Authors: []string{"Platform Team"},
			Source:  "https://github.com/myorg/team-docs",
			License: "MIT",
		},
		Packaging: RefPackaging{
			Exclude: []string{"*.draft.md", ".DS_Store"},
		},
		Hints: RefHints{
			Suggested_queries: []string{"error handling patterns"},
			Primary_files:     []string{"README.md"},
		},
		Metadata: map[string]any{
			"team": "platform",
		},
	}
}

// ============================================================
// Valid Manifest Tests
// ============================================================

func TestValidateRefManifest_ValidCases(t *testing.T) {
	tests := []struct {
		name     string
		manifest *RefManifest
	}{
		{
			name:     "minimal valid manifest",
			manifest: validMinimalManifest(),
		},
		{
			name:     "full valid manifest",
			manifest: validFullManifest(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRefManifest(tt.manifest)
			assert.NoError(t, err)
		})
	}
}

func TestValidateRefManifest_AllClassificationTypes(t *testing.T) {
	types := []string{
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

	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			manifest := validMinimalManifest()
			manifest.Content.Classifications = []RefClassification{
				{Type: ClassificationType(typ)},
			}
			err := ValidateRefManifest(manifest)
			assert.NoError(t, err, "classification type %q should be valid", typ)
		})
	}
}

func TestValidateRefManifest_ValidSchemaVersions(t *testing.T) {
	versions := []string{
		"1.0.0",
		"0.1.0",
		"2.10.100",
		"0.0.1",
	}

	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			manifest := validMinimalManifest()
			manifest.Schema_version = version
			err := ValidateRefManifest(manifest)
			assert.NoError(t, err, "schema_version %q should be valid", version)
		})
	}
}

func TestValidateRefManifest_ValidLinkFormats(t *testing.T) {
	links := []string{
		"go-standards",
		"my-ref",
		"api-patterns",
		"a1",
		"test123",
		"my-very-long-link-name",
	}

	for _, link := range links {
		t.Run(link, func(t *testing.T) {
			manifest := validMinimalManifest()
			manifest.Ref.Link = link
			err := ValidateRefManifest(manifest)
			assert.NoError(t, err, "link %q should be valid", link)
		})
	}
}

// ============================================================
// Invalid Manifest Tests - Missing Required Fields
// ============================================================

func TestValidateRefManifest_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		manifest *RefManifest
	}{
		{
			name: "missing schema_version",
			manifest: &RefManifest{
				Schema_version: "",
				Ref: RefIdentification{
					Title: "Test",
					Link:  "test-ref",
				},
				Content: RefContent{
					Description:     "Test description.",
					Classifications: []RefClassification{{Type: "reference"}},
					Tags:            []string{"test"},
				},
			},
		},
		{
			name: "missing ref.title",
			manifest: &RefManifest{
				Schema_version: "1.0.0",
				Ref: RefIdentification{
					Title: "",
					Link:  "test-ref",
				},
				Content: RefContent{
					Description:     "Test description.",
					Classifications: []RefClassification{{Type: "reference"}},
					Tags:            []string{"test"},
				},
			},
		},
		{
			name: "missing ref.link",
			manifest: &RefManifest{
				Schema_version: "1.0.0",
				Ref: RefIdentification{
					Title: "Test",
					Link:  "",
				},
				Content: RefContent{
					Description:     "Test description.",
					Classifications: []RefClassification{{Type: "reference"}},
					Tags:            []string{"test"},
				},
			},
		},
		{
			name: "missing content.description",
			manifest: &RefManifest{
				Schema_version: "1.0.0",
				Ref: RefIdentification{
					Title: "Test",
					Link:  "test-ref",
				},
				Content: RefContent{
					Description:     "",
					Classifications: []RefClassification{{Type: "reference"}},
					Tags:            []string{"test"},
				},
			},
		},
		{
			name: "empty classifications array",
			manifest: &RefManifest{
				Schema_version: "1.0.0",
				Ref: RefIdentification{
					Title: "Test",
					Link:  "test-ref",
				},
				Content: RefContent{
					Description:     "Test description.",
					Classifications: []RefClassification{},
					Tags:            []string{"test"},
				},
			},
		},
		{
			name: "empty tags array",
			manifest: &RefManifest{
				Schema_version: "1.0.0",
				Ref: RefIdentification{
					Title: "Test",
					Link:  "test-ref",
				},
				Content: RefContent{
					Description:     "Test description.",
					Classifications: []RefClassification{{Type: "reference"}},
					Tags:            []string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRefManifest(tt.manifest)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrRefManifestValidation), "expected ErrRefManifestValidation")
		})
	}
}

// ============================================================
// Invalid Manifest Tests - Format Violations
// ============================================================

func TestValidateRefManifest_InvalidSchemaVersionFormats(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"missing patch", "1.0"},
		{"missing minor and patch", "1"},
		{"with v prefix", "v1.0.0"},
		{"with prerelease", "1.0.0-beta"},
		{"with extra dots", "1.0.0.0"},
		{"with letters", "1.0.a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validMinimalManifest()
			manifest.Schema_version = tt.version
			err := ValidateRefManifest(manifest)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrRefManifestValidation))
		})
	}
}

func TestValidateRefManifest_InvalidLinkFormats(t *testing.T) {
	tests := []struct {
		name string
		link string
	}{
		{"uppercase", "Go-Standards"},
		{"starts with hyphen", "-go-standards"},
		{"ends with hyphen", "go-standards-"},
		{"spaces", "go standards"},
		{"special chars underscore", "go_standards"},
		{"special chars dot", "go.standards"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validMinimalManifest()
			manifest.Ref.Link = tt.link
			err := ValidateRefManifest(manifest)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrRefManifestValidation))
		})
	}
}

func TestValidateRefManifest_InvalidClassificationType(t *testing.T) {
	tests := []struct {
		name            string
		classificationType string
	}{
		{"typo guidlines", "guidlines"},
		{"unknown type", "custom-type"},
		{"camelCase", "codeExamples"},
		{"uppercase", "GUIDELINES"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validMinimalManifest()
			manifest.Content.Classifications = []RefClassification{
				{Type: ClassificationType(tt.classificationType)},
			}
			err := ValidateRefManifest(manifest)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrRefManifestValidation))
		})
	}
}

// ============================================================
// Error Message Quality Tests
// ============================================================

func TestValidateRefManifest_ErrorContainsContext(t *testing.T) {
	// Test that error message is wrapped with ErrRefManifestValidation
	manifest := validMinimalManifest()
	manifest.Schema_version = "invalid"

	err := ValidateRefManifest(manifest)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrRefManifestValidation),
		"error should wrap ErrRefManifestValidation")
}

func TestValidateRefManifest_NilManifest(t *testing.T) {
	err := ValidateRefManifest(nil)
	require.Error(t, err)
}

// ============================================================
// Edge Cases
// ============================================================

func TestValidateRefManifest_MultipleClassifications(t *testing.T) {
	manifest := validMinimalManifest()
	manifest.Content.Classifications = []RefClassification{
		{Type: "guidelines"},
		{Type: "code-examples"},
		{Type: "architecture"},
	}
	err := ValidateRefManifest(manifest)
	assert.NoError(t, err)
}

func TestValidateRefManifest_MultipleTags(t *testing.T) {
	manifest := validMinimalManifest()
	manifest.Content.Tags = []string{"golang", "conventions", "testing", "best-practices"}
	err := ValidateRefManifest(manifest)
	assert.NoError(t, err)
}

func TestValidateRefManifest_ClassificationWithDescription(t *testing.T) {
	manifest := validMinimalManifest()
	manifest.Content.Classifications = []RefClassification{
		{Type: "guidelines", Description: "Coding standards and conventions"},
	}
	err := ValidateRefManifest(manifest)
	assert.NoError(t, err)
}

func TestValidateRefManifest_ProvenanceWithAllFields(t *testing.T) {
	manifest := validMinimalManifest()
	manifest.Provenance = RefProvenance{
		Authors: []string{"Platform Team", "DevOps Team"},
		Created: "2024-01-15T10:00:00Z",
		Updated: "2025-01-30T15:30:00Z",
		Source:  "https://github.com/myorg/team-docs",
		License: "MIT",
	}
	err := ValidateRefManifest(manifest)
	assert.NoError(t, err)
}

func TestValidateRefManifest_EmptyProvenance(t *testing.T) {
	manifest := validMinimalManifest()
	manifest.Provenance = RefProvenance{}
	err := ValidateRefManifest(manifest)
	assert.NoError(t, err)
}

func TestValidateRefManifest_PackagingWithExclude(t *testing.T) {
	manifest := validMinimalManifest()
	manifest.Packaging = RefPackaging{
		Exclude: []string{"*.draft.md", ".DS_Store", "tmp/"},
	}
	err := ValidateRefManifest(manifest)
	assert.NoError(t, err)
}

func TestValidateRefManifest_HintsWithAllFields(t *testing.T) {
	manifest := validMinimalManifest()
	manifest.Hints = RefHints{
		Suggested_queries: []string{
			"How should I structure error handling?",
			"What are the testing conventions?",
		},
		Primary_files: []string{
			"README.md",
			"standards/golang.md",
		},
	}
	err := ValidateRefManifest(manifest)
	assert.NoError(t, err)
}

func TestValidateRefManifest_NestedMetadata(t *testing.T) {
	manifest := validMinimalManifest()
	manifest.Metadata = map[string]any{
		"team":     "platform",
		"priority": 1,
		"tags":     []any{"internal", "approved"},
		"nested": map[string]any{
			"level1": map[string]any{
				"level2": "deep value",
			},
		},
	}
	err := ValidateRefManifest(manifest)
	assert.NoError(t, err)
}
