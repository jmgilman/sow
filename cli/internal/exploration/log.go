package exploration

import (
	"fmt"
	"os"
	"time"

	"github.com/jmgilman/sow/cli/internal/logging"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// Exploration-specific log actions.
const (
	ActionFileAdded       = "file_added"
	ActionFileRemoved     = "file_removed"
	ActionFileUpdated     = "file_updated"
	ActionTopicAdded      = "topic_added"
	ActionTopicCompleted  = "topic_completed"
	ActionTopicUpdated    = "topic_updated"
	ActionJournalEntry    = "journal_entry"
	ActionStatusChanged   = "status_changed"
)

// Log result values.
const (
	ResultSuccess = "success"
	ResultError   = "error"
)

// LogPath is the file path for the exploration log.
const LogPath = "exploration/log.md"

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

// LogFileAdded logs the addition of a file to the exploration.
func LogFileAdded(ctx *sow.Context, path, description string, tags []string) {
	notes := fmt.Sprintf("Added file: %s\nDescription: %s", path, description)
	if len(tags) > 0 {
		notes += fmt.Sprintf("\nTags: %v", tags)
	}
	logAction(ctx, ActionFileAdded, []string{path}, notes)
}

// LogFileUpdated logs the update of a file's metadata.
func LogFileUpdated(ctx *sow.Context, path, description string, tags []string) {
	notes := fmt.Sprintf("Updated file: %s", path)
	if description != "" {
		notes += fmt.Sprintf("\nNew description: %s", description)
	}
	if tags != nil {
		notes += fmt.Sprintf("\nNew tags: %v", tags)
	}
	logAction(ctx, ActionFileUpdated, []string{path}, notes)
}

// LogFileRemoved logs the removal of a file from the exploration.
func LogFileRemoved(ctx *sow.Context, path string) {
	notes := fmt.Sprintf("Removed file: %s", path)
	logAction(ctx, ActionFileRemoved, []string{path}, notes)
}

// LogTopicAdded logs the addition of a topic to the parking lot.
func LogTopicAdded(ctx *sow.Context, topic string) {
	notes := fmt.Sprintf("Added topic to parking lot: %s", topic)
	logAction(ctx, ActionTopicAdded, nil, notes)
}

// LogTopicUpdated logs a change to a topic's status.
func LogTopicUpdated(ctx *sow.Context, topic, status string, relatedFiles []string) {
	notes := fmt.Sprintf("Updated topic: %s\nNew status: %s", topic, status)
	if len(relatedFiles) > 0 {
		notes += fmt.Sprintf("\nRelated files: %v", relatedFiles)
	}
	logAction(ctx, ActionTopicUpdated, relatedFiles, notes)
}

// LogTopicCompleted logs the completion of a topic.
func LogTopicCompleted(ctx *sow.Context, topic string, relatedFiles []string) {
	notes := fmt.Sprintf("Completed topic: %s", topic)
	if len(relatedFiles) > 0 {
		notes += fmt.Sprintf("\nRelated files: %v", relatedFiles)
	}
	logAction(ctx, ActionTopicCompleted, relatedFiles, notes)
}

// LogJournalEntry logs the addition of a journal entry.
func LogJournalEntry(ctx *sow.Context, entryType, content string) {
	notes := fmt.Sprintf("Journal entry [%s]: %s", entryType, content)
	logAction(ctx, ActionJournalEntry, nil, notes)
}

// LogStatusChanged logs a change to the exploration status.
func LogStatusChanged(ctx *sow.Context, oldStatus, newStatus string) {
	notes := fmt.Sprintf("Status changed: %s → %s", oldStatus, newStatus)
	logAction(ctx, ActionStatusChanged, nil, notes)
}
