package project

import (
	"fmt"
	"strings"
	"time"
)

// LogEntry represents a structured log entry.
type LogEntry struct {
	Timestamp time.Time
	AgentID   string
	Action    string
	Result    string
	Files     []string
	Notes     string
}

// ValidActions defines the allowed action types for log entries.
var ValidActions = []string{
	"started_task",
	"created_file",
	"modified_file",
	"deleted_file",
	"implementation_attempt",
	"test_run",
	"refactor",
	"debugging",
	"research",
	"completed_task",
	"paused_task",
}

// ValidResults defines the allowed result types for log entries.
var ValidResults = []string{
	"success",
	"error",
	"partial",
}

// Format renders the log entry as structured markdown.
//
// The format is:
//
//	---
//	timestamp: YYYY-MM-DD HH:MM:SS
//	agent: {agentID}
//	action: {action}
//	result: {result}
//	files:
//	  - path/to/file1
//	  - path/to/file2
//	---
//
//	{notes}
func (e *LogEntry) Format() string {
	var b strings.Builder

	// Front matter
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("timestamp: %s\n", e.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("agent: %s\n", e.AgentID))
	b.WriteString(fmt.Sprintf("action: %s\n", e.Action))
	b.WriteString(fmt.Sprintf("result: %s\n", e.Result))

	// Optional files list
	if len(e.Files) > 0 {
		b.WriteString("files:\n")
		for _, file := range e.Files {
			b.WriteString(fmt.Sprintf("  - %s\n", file))
		}
	}

	b.WriteString("---\n")

	// Optional notes section
	if e.Notes != "" {
		b.WriteString("\n")
		b.WriteString(e.Notes)
		b.WriteString("\n")
	}

	return b.String()
}

// Validate checks if the log entry has valid action and result values.
func (e *LogEntry) Validate() error {
	if err := validateAction(e.Action); err != nil {
		return err
	}
	if err := validateResult(e.Result); err != nil {
		return err
	}
	if e.AgentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	return nil
}

// validateAction checks if the given action is in the valid actions list.
func validateAction(action string) error {
	for _, valid := range ValidActions {
		if action == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid action %q: must be one of %v", action, ValidActions)
}

// validateResult checks if the given result is in the valid results list.
func validateResult(result string) error {
	for _, valid := range ValidResults {
		if result == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid result %q: must be one of %v", result, ValidResults)
}

// LogOption configures a log entry.
type LogOption func(*LogEntry)

// WithFiles adds files to the log entry.
func WithFiles(files ...string) LogOption {
	return func(e *LogEntry) {
		e.Files = append(e.Files, files...)
	}
}

// WithNotes adds notes to the log entry.
func WithNotes(notes string) LogOption {
	return func(e *LogEntry) {
		e.Notes = notes
	}
}
