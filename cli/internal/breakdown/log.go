package breakdown

import (
	"fmt"
	"os"
	"time"

	"github.com/jmgilman/sow/cli/internal/logging"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// Breakdown-specific log actions.
const (
	ActionInputAdded          = "input_added"
	ActionInputRemoved        = "input_removed"
	ActionWorkUnitAdded       = "work_unit_added"
	ActionWorkUnitUpdated     = "work_unit_updated"
	ActionWorkUnitRemoved     = "work_unit_removed"
	ActionWorkUnitDocumentSet = "work_unit_document_set"
	ActionWorkUnitApproved    = "work_unit_approved"
	ActionWorkUnitPublished   = "work_unit_published"
	ActionStatusChanged       = "status_changed"
)

// Log result values.
const (
	ResultSuccess = "success"
	ResultError   = "error"
)

// LogPath is the file path for the breakdown log.
const LogPath = "breakdown/log.md"

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

// LogInputAdded logs the addition of an input to the breakdown.
func LogInputAdded(ctx *sow.Context, inputType, path, description string, tags []string) {
	notes := fmt.Sprintf("Added input [%s]: %s\nDescription: %s", inputType, path, description)
	if len(tags) > 0 {
		notes += fmt.Sprintf("\nTags: %v", tags)
	}
	logAction(ctx, ActionInputAdded, []string{path}, notes)
}

// LogInputRemoved logs the removal of an input from the breakdown.
func LogInputRemoved(ctx *sow.Context, path string) {
	notes := fmt.Sprintf("Removed input: %s", path)
	logAction(ctx, ActionInputRemoved, []string{path}, notes)
}

// LogWorkUnitAdded logs the addition of a work unit to the breakdown.
func LogWorkUnitAdded(ctx *sow.Context, id, title, description string, dependsOn []string) {
	notes := fmt.Sprintf("Added work unit [%s]: %s\nDescription: %s", id, title, description)
	if len(dependsOn) > 0 {
		notes += fmt.Sprintf("\nDepends on: %v", dependsOn)
	}
	logAction(ctx, ActionWorkUnitAdded, nil, notes)
}

// LogWorkUnitUpdated logs updates to a work unit's metadata.
func LogWorkUnitUpdated(ctx *sow.Context, id string, title, description *string, dependsOn []string) {
	notes := fmt.Sprintf("Updated work unit: %s", id)
	if title != nil {
		notes += fmt.Sprintf("\nNew title: %s", *title)
	}
	if description != nil {
		notes += fmt.Sprintf("\nNew description: %s", *description)
	}
	if dependsOn != nil {
		notes += fmt.Sprintf("\nNew dependencies: %v", dependsOn)
	}
	logAction(ctx, ActionWorkUnitUpdated, nil, notes)
}

// LogWorkUnitRemoved logs the removal of a work unit from the breakdown.
func LogWorkUnitRemoved(ctx *sow.Context, id string) {
	notes := fmt.Sprintf("Removed work unit: %s", id)
	logAction(ctx, ActionWorkUnitRemoved, nil, notes)
}

// LogWorkUnitDocumentSet logs setting a work unit's document path.
func LogWorkUnitDocumentSet(ctx *sow.Context, id, documentPath string) {
	notes := fmt.Sprintf("Set document path for work unit: %s\nDocument: %s", id, documentPath)
	logAction(ctx, ActionWorkUnitDocumentSet, []string{documentPath}, notes)
}

// LogWorkUnitApproved logs the approval of a work unit.
func LogWorkUnitApproved(ctx *sow.Context, id string) {
	notes := fmt.Sprintf("Approved work unit for publishing: %s", id)
	logAction(ctx, ActionWorkUnitApproved, nil, notes)
}

// LogWorkUnitPublished logs the publication of a work unit to GitHub.
func LogWorkUnitPublished(ctx *sow.Context, id, issueURL string, issueNumber int64) {
	notes := fmt.Sprintf("Published work unit: %s\nGitHub Issue: #%d\nURL: %s", id, issueNumber, issueURL)
	logAction(ctx, ActionWorkUnitPublished, nil, notes)
}

// LogStatusChanged logs a change to the breakdown status.
func LogStatusChanged(ctx *sow.Context, oldStatus, newStatus string) {
	notes := fmt.Sprintf("Status changed: %s → %s", oldStatus, newStatus)
	logAction(ctx, ActionStatusChanged, nil, notes)
}
