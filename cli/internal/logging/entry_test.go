package logging

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogEntry_Format tests the formatted output.
func TestLogEntry_Format(t *testing.T) {
	timestamp := time.Date(2025, 10, 16, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name  string
		entry LogEntry
		want  []string // Strings that should appear in output
	}{
		{
			name: "minimal entry",
			entry: LogEntry{
				Timestamp: timestamp,
				AgentID:   "implementer-1",
				Action:    "created_file",
				Result:    "success",
			},
			want: []string{
				"---",
				"timestamp: 2025-10-16 14:30:00",
				"agent: implementer-1",
				"action: created_file",
				"result: success",
				"---",
			},
		},
		{
			name: "entry with files",
			entry: LogEntry{
				Timestamp: timestamp,
				AgentID:   "implementer-2",
				Action:    "modified_file",
				Result:    "success",
				Files:     []string{"src/main.go", "tests/main_test.go"},
			},
			want: []string{
				"---",
				"timestamp: 2025-10-16 14:30:00",
				"agent: implementer-2",
				"action: modified_file",
				"result: success",
				"files:",
				"  - src/main.go",
				"  - tests/main_test.go",
				"---",
			},
		},
		{
			name: "entry with notes",
			entry: LogEntry{
				Timestamp: timestamp,
				AgentID:   "orchestrator",
				Action:    "started_task",
				Result:    "success",
				Notes:     "Starting implementation of JWT authentication",
			},
			want: []string{
				"---",
				"timestamp: 2025-10-16 14:30:00",
				"agent: orchestrator",
				"action: started_task",
				"result: success",
				"---",
				"Starting implementation of JWT authentication",
			},
		},
		{
			name: "full entry",
			entry: LogEntry{
				Timestamp: timestamp,
				AgentID:   "architect-1",
				Action:    "research",
				Result:    "partial",
				Files:     []string{"docs/design.md"},
				Notes:     "Researched authentication patterns",
			},
			want: []string{
				"---",
				"timestamp: 2025-10-16 14:30:00",
				"agent: architect-1",
				"action: research",
				"result: partial",
				"files:",
				"  - docs/design.md",
				"---",
				"Researched authentication patterns",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.Format()

			// Check that all expected strings are present
			for _, want := range tt.want {
				assert.Contains(t, got, want, "formatted output should contain: %s", want)
			}

			// Check that output starts and ends correctly
			assert.True(t, strings.HasPrefix(got, "---\n"), "should start with front matter")
			assert.True(t, strings.HasSuffix(got, "\n"), "should end with newline")
		})
	}
}

// TestLogEntry_Validate tests validation.
func TestLogEntry_Validate(t *testing.T) {
	validEntry := LogEntry{
		Timestamp: time.Now(),
		AgentID:   "implementer-1",
		Action:    "created_file",
		Result:    "success",
	}

	tests := []struct {
		name    string
		entry   LogEntry
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid entry",
			entry:   validEntry,
			wantErr: false,
		},
		{
			name: "invalid action",
			entry: LogEntry{
				Timestamp: time.Now(),
				AgentID:   "implementer-1",
				Action:    "invalid_action",
				Result:    "success",
			},
			wantErr: true,
			errMsg:  "invalid action",
		},
		{
			name: "invalid result",
			entry: LogEntry{
				Timestamp: time.Now(),
				AgentID:   "implementer-1",
				Action:    "created_file",
				Result:    "invalid_result",
			},
			wantErr: true,
			errMsg:  "invalid result",
		},
		{
			name: "missing agent ID",
			entry: LogEntry{
				Timestamp: time.Now(),
				AgentID:   "",
				Action:    "created_file",
				Result:    "success",
			},
			wantErr: true,
			errMsg:  "agent ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateAction tests action validation.
func TestValidateAction(t *testing.T) {
	tests := []struct {
		action  string
		wantErr bool
	}{
		{"started_task", false},
		{"created_file", false},
		{"modified_file", false},
		{"deleted_file", false},
		{"implementation_attempt", false},
		{"test_run", false},
		{"refactor", false},
		{"debugging", false},
		{"research", false},
		{"completed_task", false},
		{"paused_task", false},
		{"invalid_action", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			err := ValidateAction(tt.action)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid action")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateResult tests result validation.
func TestValidateResult(t *testing.T) {
	tests := []struct {
		result  string
		wantErr bool
	}{
		{"success", false},
		{"error", false},
		{"partial", false},
		{"invalid_result", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.result, func(t *testing.T) {
			err := ValidateResult(tt.result)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid result")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestBuildAgentID tests agent ID construction.
func TestBuildAgentID(t *testing.T) {
	tests := []struct {
		name      string
		role      string
		iteration int
		want      string
	}{
		{
			name:      "implementer with iteration",
			role:      "implementer",
			iteration: 1,
			want:      "implementer-1",
		},
		{
			name:      "architect with iteration",
			role:      "architect",
			iteration: 3,
			want:      "architect-3",
		},
		{
			name:      "orchestrator no iteration",
			role:      "orchestrator",
			iteration: 0,
			want:      "orchestrator",
		},
		{
			name:      "orchestrator with iteration ignored",
			role:      "orchestrator",
			iteration: 5,
			want:      "orchestrator",
		},
		{
			name:      "reviewer with iteration",
			role:      "reviewer",
			iteration: 2,
			want:      "reviewer-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildAgentID(tt.role, tt.iteration)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestLogEntry_Format_Roundtrip tests that format produces parseable output.
func TestLogEntry_Format_Roundtrip(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Date(2025, 10, 16, 14, 30, 0, 0, time.UTC),
		AgentID:   "implementer-1",
		Action:    "created_file",
		Result:    "success",
		Files:     []string{"src/auth.go", "tests/auth_test.go"},
		Notes:     "Created authentication module with RS256",
	}

	formatted := entry.Format()

	// Verify it contains key components
	assert.Contains(t, formatted, "---")
	assert.Contains(t, formatted, "timestamp: 2025-10-16 14:30:00")
	assert.Contains(t, formatted, "agent: implementer-1")
	assert.Contains(t, formatted, "action: created_file")
	assert.Contains(t, formatted, "result: success")
	assert.Contains(t, formatted, "files:")
	assert.Contains(t, formatted, "  - src/auth.go")
	assert.Contains(t, formatted, "  - tests/auth_test.go")
	assert.Contains(t, formatted, "Created authentication module with RS256")

	// Verify structure
	lines := strings.Split(formatted, "\n")
	assert.Equal(t, "---", lines[0], "should start with ---")
	assert.True(t, strings.HasPrefix(lines[1], "timestamp:"), "second line should be timestamp")
}
