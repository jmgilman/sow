package design

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/modes"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
)

const (
	// IndexPath is the path to the design index relative to .sow/.
	IndexPath = "design/index.yaml"
)

var (
	// indexManager is the generic index manager for design mode.
	indexManager = modes.NewIndexManager[schemas.DesignIndex]("design", IndexPath)
)

// LoadIndex loads the design index from disk.
// Returns ErrNoDesign if design directory doesn't exist.
func LoadIndex(ctx *sow.Context) (*schemas.DesignIndex, error) {
	index, err := indexManager.Load(ctx)
	if err != nil {
		// Map generic error to design-specific error
		return nil, ErrNoDesign
	}
	return index, nil
}

// SaveIndex saves the design index to disk.
func SaveIndex(ctx *sow.Context, index *schemas.DesignIndex) error {
	if err := indexManager.Save(ctx, index); err != nil {
		return fmt.Errorf("failed to save design index: %w", err)
	}
	return nil
}

// InitDesign creates a new design directory and index.
func InitDesign(ctx *sow.Context, topic, branch string) error {
	fs := ctx.FS()
	if fs == nil {
		return sow.ErrNotInitialized
	}

	// Check if design already exists
	exists, _ := fs.Exists("design")
	if exists {
		return ErrDesignExists
	}

	// Create design directory
	if err := fs.MkdirAll("design", 0755); err != nil {
		return fmt.Errorf("failed to create design directory: %w", err)
	}

	// Create initial index
	index := &schemas.DesignIndex{
		Design: struct {
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
		Inputs:  []schemas.DesignInput{},
		Outputs: []schemas.DesignOutput{},
	}

	if err := SaveIndex(ctx, index); err != nil {
		// Clean up on failure
		_ = fs.RemoveAll("design")
		return fmt.Errorf("failed to save initial index: %w", err)
	}

	return nil
}

// Exists checks if a design directory exists.
func Exists(ctx *sow.Context) bool {
	return indexManager.Exists(ctx)
}

// Delete removes the design directory and all its contents.
func Delete(ctx *sow.Context) error {
	if err := indexManager.Delete(ctx); err != nil {
		return ErrNoDesign
	}
	return nil
}

// GetFilePath returns the absolute path to a file in the design directory.
func GetFilePath(ctx *sow.Context, relativePath string) string {
	return indexManager.GetFilePath(ctx, relativePath)
}
