package logformat

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormattingWriter(t *testing.T) {
	var output bytes.Buffer
	w := NewFormattingWriter(&output)

	// Write init event
	initEvent := `{"type":"system","subtype":"init","session_id":"test-123","model":"claude"}` + "\n"
	n, err := w.Write([]byte(initEvent))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(initEvent) {
		t.Errorf("Write returned %d, want %d", n, len(initEvent))
	}

	if !strings.Contains(output.String(), "SESSION STARTED") {
		t.Error("output should contain SESSION STARTED")
	}
	if !strings.Contains(output.String(), "test-123") {
		t.Error("output should contain session ID")
	}
}

func TestFormattingWriter_PartialLines(t *testing.T) {
	var output bytes.Buffer
	w := NewFormattingWriter(&output)

	// Write partial line
	part1 := `{"type":"system","subtype":"in`
	_, err := w.Write([]byte(part1))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Output should be empty (line not complete)
	if output.Len() > 0 {
		t.Error("partial line should not produce output")
	}

	// Complete the line
	part2 := `it","session_id":"abc"}` + "\n"
	_, err = w.Write([]byte(part2))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Now should have output
	if !strings.Contains(output.String(), "SESSION STARTED") {
		t.Error("completed line should produce output")
	}
}

func TestFormattingWriter_MultipleLines(t *testing.T) {
	var output bytes.Buffer
	w := NewFormattingWriter(&output)

	// Write multiple events at once
	events := `{"type":"system","subtype":"init","session_id":"test-123","model":"claude"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Thinking..."}]}}
{"type":"result","subtype":"success","duration_ms":1000,"num_turns":1}
`
	_, err := w.Write([]byte(events))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	result := output.String()

	checks := []string{
		"SESSION STARTED",
		"THINKING",
		"SESSION COMPLETE",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("output missing %q", check)
		}
	}
}

func TestFormattingWriter_Flush(t *testing.T) {
	var output bytes.Buffer
	w := NewFormattingWriter(&output)

	// Write incomplete line
	_, _ = w.Write([]byte(`{"type":"system","subtype":"init","session_id":"abc"}`))

	// Should be no output yet (no newline)
	if output.Len() > 0 {
		t.Error("should have no output before flush")
	}

	// Flush should process the remaining buffer
	err := w.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if !strings.Contains(output.String(), "SESSION STARTED") {
		t.Error("flush should process remaining buffer")
	}
}

func TestFormatFile(t *testing.T) {
	input := strings.NewReader(`{"type":"system","subtype":"init","session_id":"test","model":"claude"}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","content":"file1\nfile2"}]}}
{"type":"result","subtype":"success","duration_ms":500,"num_turns":1,"result":"Done"}
`)

	var output bytes.Buffer
	err := FormatFile(input, &output)
	if err != nil {
		t.Fatalf("FormatFile failed: %v", err)
	}

	result := output.String()

	checks := []string{
		"SESSION STARTED",
		"TOOL: Bash",
		"RESULT [OK]",
		"SESSION COMPLETE",
		"Done",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("output missing %q\nGot:\n%s", check, result)
		}
	}
}

func TestDualWriter(t *testing.T) {
	var raw bytes.Buffer
	var formatted bytes.Buffer

	w := NewDualWriter(&raw, &formatted)

	event := `{"type":"system","subtype":"init","session_id":"dual-test","model":"claude"}` + "\n"
	_, err := w.Write([]byte(event))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Raw should have the original JSON
	if !strings.Contains(raw.String(), `"type":"system"`) {
		t.Error("raw output should contain original JSON")
	}

	// Formatted should have human-readable output
	if !strings.Contains(formatted.String(), "SESSION STARTED") {
		t.Error("formatted output should contain SESSION STARTED")
	}
}
