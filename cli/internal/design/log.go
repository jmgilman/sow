package design

import (
	"fmt"
	"os"
	"time"

	"github.com/jmgilman/sow/cli/internal/logging"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// Design-specific log actions.
const (
	ActionInputAdded        = "input_added"
	ActionInputRemoved      = "input_removed"
	ActionOutputAdded       = "output_added"
	ActionOutputRemoved     = "output_removed"
	ActionOutputTargetSet   = "output_target_set"
	ActionStatusChanged     = "status_changed"
)

// Log result values.
const (
	ResultSuccess = "success"
	ResultError   = "error"
)

// LogPath is the file path for the design log.
const LogPath = "design/log.md"

// logAction is a helper that creates and appends a log entry.
func logAction(ctx *sow.Context, action string, files []string, notes string) {
	entry := &logging.LogEntry{
		Timestamp: time.Now(),
		AgentID:   "orchestrator",
		Action:    action,
		Result:    ResultSuccess,
		Files:     files,
		Notes:     notes,
	}

	if err := logging.AppendLog(ctx.FS(), LogPath, entry); err != nil {
		// Log failures are non-fatal warnings
		fmt.Fprintf(os.Stderr, "Warning: failed to log action: %v\n", err)
	}
}

// LogInputAdded logs the addition of an input to the design.
func LogInputAdded(ctx *sow.Context, inputType, path, description string, tags []string) {
	notes := fmt.Sprintf("Added input [%s]: %s\nDescription: %s", inputType, path, description)
	if len(tags) > 0 {
		notes += fmt.Sprintf("\nTags: %v", tags)
	}
	logAction(ctx, ActionInputAdded, []string{path}, notes)
}

// LogInputRemoved logs the removal of an input from the design.
func LogInputRemoved(ctx *sow.Context, path string) {
	notes := fmt.Sprintf("Removed input: %s", path)
	logAction(ctx, ActionInputRemoved, []string{path}, notes)
}

// LogOutputAdded logs the addition of an output to the design.
func LogOutputAdded(ctx *sow.Context, path, description, targetLocation, docType string, tags []string) {
	notes := fmt.Sprintf("Added output [%s]: %s\nDescription: %s\nTarget: %s", docType, path, description, targetLocation)
	if len(tags) > 0 {
		notes += fmt.Sprintf("\nTags: %v", tags)
	}
	logAction(ctx, ActionOutputAdded, []string{path}, notes)
}

// LogOutputRemoved logs the removal of an output from the design.
func LogOutputRemoved(ctx *sow.Context, path string) {
	notes := fmt.Sprintf("Removed output: %s", path)
	logAction(ctx, ActionOutputRemoved, []string{path}, notes)
}

// LogOutputTargetSet logs updating an output's target location.
func LogOutputTargetSet(ctx *sow.Context, path, targetLocation string) {
	notes := fmt.Sprintf("Updated output target: %s\nNew target: %s", path, targetLocation)
	logAction(ctx, ActionOutputTargetSet, []string{path}, notes)
}

// LogStatusChanged logs a change to the design status.
func LogStatusChanged(ctx *sow.Context, oldStatus, newStatus string) {
	notes := fmt.Sprintf("Status changed: %s → %s", oldStatus, newStatus)
	logAction(ctx, ActionStatusChanged, nil, notes)
}
