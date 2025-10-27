package exploration

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
)

// AddFile adds a file to the exploration index.
func AddFile(ctx *sow.Context, path, description string, tags []string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Check if file already exists
	for _, f := range index.Files {
		if f.Path == path {
			return ErrFileExists
		}
	}

	// Add file
	file := schemas.ExplorationFile{
		Path:        path,
		Description: description,
		Tags:        tags,
		Created_at:  time.Now(),
	}
	index.Files = append(index.Files, file)

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// UpdateFile updates a file's metadata in the exploration index.
func UpdateFile(ctx *sow.Context, path, description string, tags []string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and update file
	found := false
	for i, f := range index.Files {
		if f.Path == path {
			if description != "" {
				index.Files[i].Description = description
			}
			if tags != nil {
				index.Files[i].Tags = tags
			}
			found = true
			break
		}
	}

	if !found {
		return ErrFileNotFound
	}

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// RemoveFile removes a file from the exploration index.
func RemoveFile(ctx *sow.Context, path string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and remove file
	found := false
	newFiles := make([]schemas.ExplorationFile, 0, len(index.Files))
	for _, f := range index.Files {
		if f.Path == path {
			found = true
			continue
		}
		newFiles = append(newFiles, f)
	}

	if !found {
		return ErrFileNotFound
	}

	index.Files = newFiles

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// GetFile retrieves a file's metadata from the exploration index.
func GetFile(ctx *sow.Context, path string) (*schemas.ExplorationFile, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	// Find file
	for _, f := range index.Files {
		if f.Path == path {
			return &f, nil
		}
	}

	return nil, ErrFileNotFound
}

// ListFiles returns all files in the exploration index.
func ListFiles(ctx *sow.Context) ([]schemas.ExplorationFile, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	return index.Files, nil
}

// UpdateStatus updates the exploration status.
func UpdateStatus(ctx *sow.Context, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"active":    true,
		"completed": true,
		"abandoned": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s (must be active, completed, or abandoned)", status)
	}

	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Update status
	index.Exploration.Status = status

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// AddTopic adds a topic to the exploration's parking lot.
func AddTopic(ctx *sow.Context, topic string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Check if topic already exists
	for _, t := range index.Topics {
		if t.Topic == topic {
			return fmt.Errorf("topic already exists: %s", topic)
		}
	}

	// Add topic
	newTopic := schemas.ExplorationTopic{
		Topic:    topic,
		Status:   "pending",
		Added_at: time.Now(),
	}
	index.Topics = append(index.Topics, newTopic)

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// UpdateTopicStatus updates a topic's status and optionally associates files with it.
func UpdateTopicStatus(ctx *sow.Context, topic, status string, relatedFiles []string) error {
	// Validate status
	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s (must be pending, in_progress, or completed)", status)
	}

	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and update topic
	found := false
	for i, t := range index.Topics {
		if t.Topic == topic {
			index.Topics[i].Status = status
			if status == "completed" {
				completedAt := time.Now()
				index.Topics[i].Completed_at = completedAt
			}
			if relatedFiles != nil {
				index.Topics[i].Related_files = relatedFiles
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("topic not found: %s", topic)
	}

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// ListTopics returns all topics in the exploration index.
func ListTopics(ctx *sow.Context) ([]schemas.ExplorationTopic, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	return index.Topics, nil
}

// GetPendingTopics returns all pending topics.
func GetPendingTopics(ctx *sow.Context) ([]schemas.ExplorationTopic, error) {
	topics, err := ListTopics(ctx)
	if err != nil {
		return nil, err
	}

	var pending []schemas.ExplorationTopic
	for _, t := range topics {
		if t.Status == "pending" {
			pending = append(pending, t)
		}
	}

	return pending, nil
}

// AddJournalEntry adds an entry to the exploration journal.
func AddJournalEntry(ctx *sow.Context, entryType, content string) error {
	// Validate entry type
	validTypes := map[string]bool{
		"decision":        true,
		"insight":         true,
		"question":        true,
		"topic_added":     true,
		"topic_completed": true,
		"note":            true,
	}
	if !validTypes[entryType] {
		return fmt.Errorf("invalid entry type: %s", entryType)
	}

	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Add journal entry
	entry := schemas.JournalEntry{
		Timestamp: time.Now(),
		Type:      entryType,
		Content:   content,
	}
	index.Journal = append(index.Journal, entry)

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// GetRecentJournal returns the most recent journal entries.
func GetRecentJournal(ctx *sow.Context, limit int) ([]schemas.JournalEntry, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	// Return all if limit is 0 or negative
	if limit <= 0 {
		return index.Journal, nil
	}

	// Return last N entries
	start := len(index.Journal) - limit
	if start < 0 {
		start = 0
	}

	return index.Journal[start:], nil
}
