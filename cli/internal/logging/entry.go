// Package logging provides structured log entry formatting and validation.
//
// This package handles the creation of markdown-formatted log entries
// for sow project and task logs.
package logging

import (
	"fmt"
	"strings"
	"time"
)

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

// LogEntry represents a structured log entry.
type LogEntry struct {
	// Timestamp is when the action occurred (ISO 8601 format)
	Timestamp time.Time

	// AgentID identifies the agent (e.g., "implementer-1", "orchestrator")
	AgentID string

	// Action is the type of action from ValidActions
	Action string

	// Result is the outcome: success, error, or partial
	Result string

	// Files is an optional list of files affected by this action
	Files []string

	// Notes is optional free-form description
	Notes string
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
	if err := ValidateAction(e.Action); err != nil {
		return err
	}
	if err := ValidateResult(e.Result); err != nil {
		return err
	}
	if e.AgentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	return nil
}

// ValidateAction checks if the given action is in the valid actions list.
func ValidateAction(action string) error {
	for _, valid := range ValidActions {
		if action == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid action %q: must be one of %v", action, ValidActions)
}

// ValidateResult checks if the given result is in the valid results list.
func ValidateResult(result string) error {
	for _, valid := range ValidResults {
		if result == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid result %q: must be one of %v", result, ValidResults)
}

// BuildAgentID constructs an agent ID from role and iteration.
//
// For workers with iterations, format is: {role}-{iteration}
// For orchestrator, just use "orchestrator"
//
// Examples:
//   - BuildAgentID("implementer", 1) -> "implementer-1"
//   - BuildAgentID("architect", 3) -> "architect-3"
//   - BuildAgentID("orchestrator", 0) -> "orchestrator"
func BuildAgentID(role string, iteration int) string {
	if role == "orchestrator" || iteration == 0 {
		return role
	}
	return fmt.Sprintf("%s-%d", role, iteration)
}
