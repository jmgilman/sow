package breakdown

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/modes"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
)

const (
	// IndexPath is the path to the breakdown index relative to .sow/.
	IndexPath = "breakdown/index.yaml"
)

var (
	// indexManager is the generic index manager for breakdown mode.
	indexManager = modes.NewIndexManager[schemas.BreakdownIndex]("breakdown", IndexPath)
)

// LoadIndex loads the breakdown index from disk.
// Returns ErrNoBreakdown if breakdown directory doesn't exist.
func LoadIndex(ctx *sow.Context) (*schemas.BreakdownIndex, error) {
	index, err := indexManager.Load(ctx)
	if err != nil {
		// Map generic error to breakdown-specific error
		return nil, ErrNoBreakdown
	}
	return index, nil
}

// SaveIndex saves the breakdown index to disk.
func SaveIndex(ctx *sow.Context, index *schemas.BreakdownIndex) error {
	if err := indexManager.Save(ctx, index); err != nil {
		return fmt.Errorf("failed to save breakdown index: %w", err)
	}
	return nil
}

// InitBreakdown creates a new breakdown directory and index.
func InitBreakdown(ctx *sow.Context, topic, branch string) error {
	fs := ctx.FS()
	if fs == nil {
		return sow.ErrNotInitialized
	}

	// Check if breakdown already exists
	exists, _ := fs.Exists("breakdown")
	if exists {
		return ErrBreakdownExists
	}

	// Create breakdown directory
	if err := fs.MkdirAll("breakdown", 0755); err != nil {
		return fmt.Errorf("failed to create breakdown directory: %w", err)
	}

	// Create units subdirectory
	if err := fs.MkdirAll("breakdown/units", 0755); err != nil {
		return fmt.Errorf("failed to create units directory: %w", err)
	}

	// Create initial index
	index := &schemas.BreakdownIndex{
		Breakdown: struct {
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
		Inputs:     []schemas.BreakdownInput{},
		Work_units: []schemas.BreakdownWorkUnit{},
	}

	if err := SaveIndex(ctx, index); err != nil {
		// Clean up on failure
		_ = fs.RemoveAll("breakdown")
		return fmt.Errorf("failed to save initial index: %w", err)
	}

	return nil
}

// Exists checks if a breakdown directory exists.
func Exists(ctx *sow.Context) bool {
	return indexManager.Exists(ctx)
}

// Delete removes the breakdown directory and all its contents.
func Delete(ctx *sow.Context) error {
	if err := indexManager.Delete(ctx); err != nil {
		return ErrNoBreakdown
	}
	return nil
}

// GetFilePath returns the absolute path to a file in the breakdown directory.
func GetFilePath(ctx *sow.Context, relativePath string) string {
	return indexManager.GetFilePath(ctx, relativePath)
}
