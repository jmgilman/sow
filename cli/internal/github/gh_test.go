package github

import (
	"testing"
)

func TestErrGHNotInstalled(t *testing.T) {
	err := ErrGHNotInstalled{}
	expected := "GitHub CLI (gh) not found. Install from: https://cli.github.com/"
	if err.Error() != expected {
		t.Errorf("ErrGHNotInstalled.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestErrGHNotAuthenticated(t *testing.T) {
	err := ErrGHNotAuthenticated{}
	expected := "GitHub CLI not authenticated. Run: gh auth login"
	if err.Error() != expected {
		t.Errorf("ErrGHNotAuthenticated.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestErrGHCommand(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrGHCommand
		expected string
	}{
		{
			name: "with stderr",
			err: ErrGHCommand{
				Command: "issue list",
				Stderr:  "permission denied",
			},
			expected: "gh issue list failed: permission denied",
		},
		{
			name: "with error",
			err: ErrGHCommand{
				Command: "issue view 123",
				Err:     ErrGHNotInstalled{},
			},
			expected: "gh issue view 123 failed: GitHub CLI (gh) not found. Install from: https://cli.github.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ErrGHCommand.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIssueHasLabel(t *testing.T) {
	issue := &Issue{
		Number: 123,
		Title:  "Test Issue",
		Labels: []struct {
			Name string `json:"name"`
		}{
			{Name: "sow"},
			{Name: "bug"},
		},
	}

	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "has label", label: "sow", want: true},
		{name: "has different label", label: "bug", want: true},
		{name: "does not have label", label: "feature", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := issue.HasLabel(tt.label); got != tt.want {
				t.Errorf("Issue.HasLabel(%q) = %v, want %v", tt.label, got, tt.want)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Add authentication system",
			expected: "add-authentication-system",
		},
		{
			name:     "with special characters",
			input:    "Fix: Bug in login (urgent!)",
			expected: "fix-bug-in-login-urgent",
		},
		{
			name:     "with underscores",
			input:    "Refactor_user_model",
			expected: "refactor-user-model",
		},
		{
			name:     "with multiple spaces",
			input:    "Add    multiple   spaces",
			expected: "add-multiple-spaces",
		},
		{
			name:     "already kebab-case",
			input:    "already-kebab-case",
			expected: "already-kebab-case",
		},
		{
			name:     "with trailing/leading spaces",
			input:    "  trim me  ",
			expected: "trim-me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toKebabCase(tt.input); got != tt.expected {
				t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// Note: CreatePullRequest requires actual gh CLI integration and is tested via E2E tests.
// The function is straightforward: it calls gh pr create and parses the URL from output.
// Mock testing would require dependency injection which adds complexity for minimal benefit.
