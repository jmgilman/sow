// Package exploration provides exploration mode functionality.
package exploration

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/modes"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
)

const (
	// IndexPath is the path to the exploration index relative to .sow/.
	IndexPath = "exploration/index.yaml"
)

var (
	// indexManager is the generic index manager for exploration mode.
	indexManager = modes.NewIndexManager[schemas.ExplorationIndex]("exploration", IndexPath)
)

// LoadIndex loads the exploration index from disk.
// Returns ErrNoExploration if exploration directory doesn't exist.
func LoadIndex(ctx *sow.Context) (*schemas.ExplorationIndex, error) {
	index, err := indexManager.Load(ctx)
	if err != nil {
		// Map generic error to exploration-specific error
		return nil, ErrNoExploration
	}
	return index, nil
}

// SaveIndex saves the exploration index to disk.
func SaveIndex(ctx *sow.Context, index *schemas.ExplorationIndex) error {
	if err := indexManager.Save(ctx, index); err != nil {
		return fmt.Errorf("failed to save exploration index: %w", err)
	}
	return nil
}

// InitExploration creates a new exploration directory and index.
func InitExploration(ctx *sow.Context, topic, branch string) error {
	fs := ctx.FS()
	if fs == nil {
		return sow.ErrNotInitialized
	}

	// Check if exploration already exists
	exists, _ := fs.Exists("exploration")
	if exists {
		return ErrExplorationExists
	}

	// Create exploration directory
	if err := fs.MkdirAll("exploration", 0755); err != nil {
		return fmt.Errorf("failed to create exploration directory: %w", err)
	}

	// Create initial index
	index := &schemas.ExplorationIndex{
		Exploration: struct {
			Topic      string    `json:"topic"`
			Branch     string    `json:"branch"`
			Created_at time.Time `json:"created_at"`
			Status     string    `json:"status"`
		}{
			Topic:      topic,
			Branch:     branch,
			Created_at: time.Now(),
			Status:     "active",
		},
		Files: []schemas.ExplorationFile{},
	}

	if err := SaveIndex(ctx, index); err != nil {
		// Clean up on failure
		_ = fs.RemoveAll("exploration")
		return fmt.Errorf("failed to save initial index: %w", err)
	}

	return nil
}

// Exists checks if an exploration directory exists.
func Exists(ctx *sow.Context) bool {
	return indexManager.Exists(ctx)
}

// Delete removes the exploration directory and all its contents.
func Delete(ctx *sow.Context) error {
	if err := indexManager.Delete(ctx); err != nil {
		return ErrNoExploration
	}
	return nil
}

// GetFilePath returns the absolute path to a file in the exploration directory.
func GetFilePath(ctx *sow.Context, relativePath string) string {
	return indexManager.GetFilePath(ctx, relativePath)
}
