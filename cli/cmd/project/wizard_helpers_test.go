package project

import (
	"errors"
	"testing"
)

// TestNormalizeName tests the normalizeName function with various inputs.
func TestNormalizeName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Basic transformations
		{"Web Based Agents", "web-based-agents"},
		{"API V2", "api-v2"},
		{"UPPERCASE", "uppercase"},
		{"  spaces  ", "spaces"},

		// Special character handling
		{"With!Invalid@Chars#", "withinvalidchars"},
		{"feature--name", "feature-name"},
		{"-leading-trailing-", "leading-trailing"},

		// Edge cases
		{"", ""},
		{"   ", ""},
		{"!!!@@@###", ""},
		{"123-numbers-456", "123-numbers-456"},
		{"under_scores", "under_scores"},

		// Multiple consecutive hyphens
		{"multiple---hyphens", "multiple-hyphens"},
		{"many----hyphens----here", "many-hyphens-here"},

		// Unicode characters (should be removed)
		{"café", "caf"},
		{"hello-世界", "hello"},
		{"emoji-🎉-test", "emoji-test"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeName(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeName(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestNormalizeName_EdgeCases tests edge cases more thoroughly.
func TestNormalizeName_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "     ",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "mixed valid and invalid",
			input:    "valid-@#$-name",
			expected: "valid-name",
		},
		{
			name:     "hyphens at start and end",
			input:    "---middle---",
			expected: "middle",
		},
		{
			name:     "consecutive hyphens throughout",
			input:    "a--b---c----d",
			expected: "a-b-c-d",
		},
		{
			name:     "very long name",
			input:    "this-is-a-very-long-project-name-that-should-still-work-correctly",
			expected: "this-is-a-very-long-project-name-that-should-still-work-correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeName(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeName(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestGetTypePrefix tests the getTypePrefix function.
func TestGetTypePrefix(t *testing.T) {
	testCases := []struct {
		projectType string
		expected    string
	}{
		// Valid types
		{"standard", "feat/"},
		{"exploration", "explore/"},
		{"design", "design/"},
		{"breakdown", "breakdown/"},

		// Invalid types (should return default)
		{"unknown", "feat/"},
		{"", "feat/"},
		{"invalid-type", "feat/"},
	}

	for _, tc := range testCases {
		t.Run(tc.projectType, func(t *testing.T) {
			result := getTypePrefix(tc.projectType)
			if result != tc.expected {
				t.Errorf("getTypePrefix(%q) = %q; want %q", tc.projectType, result, tc.expected)
			}
		})
	}
}

// TestGetTypeOptions tests the getTypeOptions function.
func TestGetTypeOptions(t *testing.T) {
	options := getTypeOptions()

	// Should have 5 options (4 types + cancel)
	if len(options) != 5 {
		t.Errorf("getTypeOptions() returned %d options; want 5", len(options))
	}

	// Note: We can't easily inspect huh.Option internals in tests,
	// so we just verify the count. The order is verified implicitly
	// by the order they're created in getTypeOptions().
	// This is a limitation of the huh library's API.
}

// TestPreviewBranchName tests the previewBranchName function.
func TestPreviewBranchName(t *testing.T) {
	testCases := []struct {
		projectType string
		name        string
		expected    string
	}{
		// Test each project type
		{"standard", "Add JWT Auth", "feat/add-jwt-auth"},
		{"exploration", "Research Caching", "explore/research-caching"},
		{"design", "API Architecture", "design/api-architecture"},
		{"breakdown", "Split Large Task", "breakdown/split-large-task"},

		// Test name normalization
		{"standard", "With!Special@Chars", "feat/withspecialchars"},
		{"standard", "UPPERCASE", "feat/uppercase"},
		{"standard", "  spaces  ", "feat/spaces"},

		// Test unknown type (should use default)
		{"unknown", "Test Project", "feat/test-project"},
	}

	for _, tc := range testCases {
		t.Run(tc.projectType+"/"+tc.name, func(t *testing.T) {
			result := previewBranchName(tc.projectType, tc.name)
			if result != tc.expected {
				t.Errorf("previewBranchName(%q, %q) = %q; want %q",
					tc.projectType, tc.name, result, tc.expected)
			}
		})
	}
}

// TestWithSpinner_PropagatesError tests that withSpinner propagates errors.
func TestWithSpinner_PropagatesError(t *testing.T) {
	expectedErr := errors.New("test error")

	err := withSpinner("Test operation", func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("withSpinner() = %v; want %v", err, expectedErr)
	}
}

// TestWithSpinner_ReturnsNilOnSuccess tests that withSpinner returns nil on success.
func TestWithSpinner_ReturnsNilOnSuccess(t *testing.T) {
	err := withSpinner("Test operation", func() error {
		return nil
	})

	if err != nil {
		t.Errorf("withSpinner() = %v; want nil", err)
	}
}

// TestProjectTypesMap verifies the projectTypes map configuration.
func TestProjectTypesMap(t *testing.T) {
	// Verify all four types exist
	expectedTypes := map[string]struct {
		prefix      string
		description string
	}{
		"standard": {
			prefix:      "feat/",
			description: "Feature work and bug fixes",
		},
		"exploration": {
			prefix:      "explore/",
			description: "Research and investigation",
		},
		"design": {
			prefix:      "design/",
			description: "Architecture and design documents",
		},
		"breakdown": {
			prefix:      "breakdown/",
			description: "Decompose work into tasks",
		},
	}

	if len(projectTypes) != len(expectedTypes) {
		t.Errorf("projectTypes has %d entries; want %d", len(projectTypes), len(expectedTypes))
	}

	for typeName, expected := range expectedTypes {
		config, exists := projectTypes[typeName]
		if !exists {
			t.Errorf("projectTypes missing type %q", typeName)
			continue
		}

		if config.Prefix != expected.prefix {
			t.Errorf("projectTypes[%q].Prefix = %q; want %q",
				typeName, config.Prefix, expected.prefix)
		}

		if config.Description != expected.description {
			t.Errorf("projectTypes[%q].Description = %q; want %q",
				typeName, config.Description, expected.description)
		}
	}
}
