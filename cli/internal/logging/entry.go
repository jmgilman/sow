// Package logging provides shared logging functionality for both projects and modes.
package logging

import (
	"fmt"
	"time"
)

// LogOption configures logging.
type LogOption func(*LogEntry)

// LogEntry represents a structured log entry.
type LogEntry struct {
	Timestamp time.Time
	AgentID   string
	Action    string
	Result    string
	Files     []string
	Notes     string
}

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

// Validate checks if the log entry has valid values.
func (e *LogEntry) Validate() error {
	if e.AgentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if e.Action == "" {
		return fmt.Errorf("action is required")
	}
	if e.Result == "" {
		return fmt.Errorf("result is required")
	}
	return nil
}

// Format renders the log entry as structured markdown.
func (e *LogEntry) Format() string {
	var b []byte

	// Front matter
	b = append(b, "---\n"...)
	b = append(b, fmt.Sprintf("timestamp: %s\n", e.Timestamp.Format("2006-01-02 15:04:05"))...)
	b = append(b, fmt.Sprintf("agent: %s\n", e.AgentID)...)
	b = append(b, fmt.Sprintf("action: %s\n", e.Action)...)
	b = append(b, fmt.Sprintf("result: %s\n", e.Result)...)

	// Optional files list
	if len(e.Files) > 0 {
		b = append(b, "files:\n"...)
		for _, file := range e.Files {
			b = append(b, fmt.Sprintf("  - %s\n", file)...)
		}
	}

	b = append(b, "---\n"...)

	// Optional notes section
	if e.Notes != "" {
		b = append(b, "\n"...)
		b = append(b, e.Notes...)
		b = append(b, "\n"...)
	}

	return string(b)
}
