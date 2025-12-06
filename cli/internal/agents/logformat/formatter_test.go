package logformat

import (
	"strings"
	"testing"
)

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantType EventType
		wantErr bool
	}{
		{
			name:    "system init event",
			input:   `{"type":"system","subtype":"init","session_id":"abc-123","model":"claude-opus-4-5"}`,
			wantType: EventTypeSystem,
		},
		{
			name:    "assistant event",
			input:   `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}]}}`,
			wantType: EventTypeAssistant,
		},
		{
			name:    "user event",
			input:   `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","content":"output"}]}}`,
			wantType: EventTypeUser,
		},
		{
			name:    "result event",
			input:   `{"type":"result","subtype":"success","duration_ms":1000,"num_turns":5}`,
			wantType: EventTypeResult,
		},
		{
			name:    "invalid json",
			input:   `{invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseEvent([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if event.Type != tt.wantType {
				t.Errorf("got type %q, want %q", event.Type, tt.wantType)
			}
		})
	}
}

func TestFormatter_FormatInit(t *testing.T) {
	input := `{"type":"system","subtype":"init","session_id":"abc-123","model":"claude-opus-4-5-20251101","cwd":"/home/user/project","claude_code_version":"2.0.60","permissionMode":"bypassPermissions"}`

	event, err := ParseEvent([]byte(input))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	f := NewFormatter()
	output := f.Format(event)

	// Check key elements are present
	checks := []string{
		"SESSION STARTED",
		"Session ID: abc-123",
		"Model: claude-opus-4-5-20251101",
		"Working Dir: /home/user/project",
		"Claude Code: v2.0.60",
		"Permission Mode: bypassPermissions",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing %q\nGot:\n%s", check, output)
		}
	}
}

func TestFormatter_FormatToolCall(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		checks []string
	}{
		{
			name:  "Bash tool",
			input: `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"ls -la","description":"List files"}}]}}`,
			checks: []string{
				"TOOL: Bash",
				"Command: ls -la",
				"Description: List files",
			},
		},
		{
			name:  "Read tool",
			input: `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/path/to/file.go"}}]}}`,
			checks: []string{
				"TOOL: Read",
				"File: /path/to/file.go",
			},
		},
		{
			name:  "Write tool",
			input: `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Write","input":{"file_path":"/path/to/file.go","content":"line1\nline2\nline3"}}]}}`,
			checks: []string{
				"TOOL: Write",
				"File: /path/to/file.go",
				"Content: 3 lines",
			},
		},
		{
			name:  "Thinking text",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Let me analyze this problem..."}]}}`,
			checks: []string{
				"THINKING",
				"Let me analyze this problem...",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseEvent([]byte(tt.input))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			f := NewFormatter()
			output := f.Format(event)

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("output missing %q\nGot:\n%s", check, output)
				}
			}
		})
	}
}

func TestFormatter_FormatToolResult(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		checks []string
	}{
		{
			name:  "successful result",
			input: `{"type":"user","message":{"content":[{"type":"tool_result","content":"file contents here","is_error":false}]},"tool_use_result":{"stdout":"file contents here","stderr":""}}`,
			checks: []string{
				"RESULT [OK]",
				"file contents here",
			},
		},
		{
			name:  "error result",
			input: `{"type":"user","message":{"content":[{"type":"tool_result","content":"command not found","is_error":true}]},"tool_use_result":{"stdout":"","stderr":"command not found"}}`,
			checks: []string{
				"RESULT [ERROR]",
				"STDERR:",
				"command not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseEvent([]byte(tt.input))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			f := NewFormatter()
			output := f.Format(event)

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("output missing %q\nGot:\n%s", check, output)
				}
			}
		})
	}
}

func TestFormatter_FormatResult(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		checks []string
	}{
		{
			name:  "successful session",
			input: `{"type":"result","subtype":"success","is_error":false,"duration_ms":120000,"duration_api_ms":125000,"num_turns":10,"total_cost_usd":1.5,"result":"Task completed successfully"}`,
			checks: []string{
				"SESSION COMPLETE",
				"Duration: 2m 0s",
				"Turns: 10",
				"Cost: $1.50",
				"FINAL RESULT:",
				"Task completed successfully",
			},
		},
		{
			name:  "failed session",
			input: `{"type":"result","subtype":"error","is_error":true,"duration_ms":5000,"num_turns":1,"result":"Failed to complete"}`,
			checks: []string{
				"SESSION FAILED",
				"Duration: 5.0s",
				"Turns: 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseEvent([]byte(tt.input))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			f := NewFormatter()
			output := f.Format(event)

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("output missing %q\nGot:\n%s", check, output)
				}
			}
		})
	}
}

func TestFormatter_Truncation(t *testing.T) {
	// Create a very long content
	longContent := strings.Repeat("a", 2000)
	input := `{"type":"user","message":{"content":[{"type":"tool_result","content":"` + longContent + `","is_error":false}]},"tool_use_result":{"stdout":"` + longContent + `"}}`

	event, err := ParseEvent([]byte(input))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	f := NewFormatter()
	output := f.Format(event)

	// Should be truncated
	if !strings.Contains(output, "truncated") {
		t.Error("long content should be truncated")
	}

	// Should not contain the full 2000 chars
	if strings.Contains(output, longContent) {
		t.Error("output should not contain full long content")
	}
}

func TestFormatter_TodoWrite(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"TodoWrite","input":{"todos":[{"content":"First task","status":"completed","activeForm":"Doing first"},{"content":"Second task","status":"in_progress","activeForm":"Doing second"},{"content":"Third task","status":"pending","activeForm":"Doing third"}]}}]}}`

	event, err := ParseEvent([]byte(input))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	f := NewFormatter()
	output := f.Format(event)

	checks := []string{
		"TOOL: TodoWrite",
		"[x] First task",
		"[>] Second task",
		"[ ] Third task",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing %q\nGot:\n%s", check, output)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ms   int64
		want string
	}{
		{500, "500ms"},
		{1000, "1.0s"},
		{5500, "5.5s"},
		{60000, "1m 0s"},
		{125000, "2m 5s"},
		{3600000, "60m 0s"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.ms)
			if got != tt.want {
				t.Errorf("formatDuration(%d) = %q, want %q", tt.ms, got, tt.want)
			}
		})
	}
}
