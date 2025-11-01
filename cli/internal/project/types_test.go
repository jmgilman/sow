package project

import (
	"testing"
)

// TestDetectProjectType verifies branch prefix detection logic.
func TestDetectProjectType(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "exploration prefix",
			branch:   "explore/auth",
			expected: "exploration",
		},
		{
			name:     "exploration prefix with numbers",
			branch:   "explore/123-feature",
			expected: "exploration",
		},
		{
			name:     "design prefix",
			branch:   "design/api",
			expected: "design",
		},
		{
			name:     "design prefix with numbers",
			branch:   "design/456-system",
			expected: "design",
		},
		{
			name:     "breakdown prefix",
			branch:   "breakdown/features",
			expected: "breakdown",
		},
		{
			name:     "breakdown prefix with numbers",
			branch:   "breakdown/789-epic",
			expected: "breakdown",
		},
		{
			name:     "feature branch defaults to standard",
			branch:   "feat/new-thing",
			expected: "standard",
		},
		{
			name:     "fix branch defaults to standard",
			branch:   "fix/bug-123",
			expected: "standard",
		},
		{
			name:     "main branch defaults to standard",
			branch:   "main",
			expected: "standard",
		},
		{
			name:     "master branch defaults to standard",
			branch:   "master",
			expected: "standard",
		},
		{
			name:     "numbered branch defaults to standard",
			branch:   "123-feature-name",
			expected: "standard",
		},
		{
			name:     "empty branch defaults to standard",
			branch:   "",
			expected: "standard",
		},
		{
			name:     "branch with explore in middle defaults to standard",
			branch:   "feature/explore/something",
			expected: "standard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectProjectType(tt.branch)
			if got != tt.expected {
				t.Errorf("DetectProjectType(%q) = %q, want %q", tt.branch, got, tt.expected)
			}
		})
	}
}
