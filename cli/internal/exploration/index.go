// Package exploration provides exploration mode functionality.
package exploration

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"gopkg.in/yaml.v3"
)

const (
	// IndexPath is the path to the exploration index relative to .sow/.
	IndexPath = "exploration/index.yaml"
)

// LoadIndex loads the exploration index from disk.
// Returns ErrNoExploration if exploration directory doesn't exist.
func LoadIndex(ctx *sow.Context) (*schemas.ExplorationIndex, error) {
	fs := ctx.FS()
	if fs == nil {
		return nil, sow.ErrNotInitialized
	}

	// Check if exploration directory exists
	exists, err := fs.Exists("exploration")
	if err != nil {
		return nil, fmt.Errorf("failed to check exploration directory: %w", err)
	}
	if !exists {
		return nil, ErrNoExploration
	}

	// Read index file
	data, err := fs.ReadFile(IndexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read exploration index: %w", err)
	}

	// Parse YAML
	var index schemas.ExplorationIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse exploration index: %w", err)
	}

	return &index, nil
}

// SaveIndex saves the exploration index to disk.
func SaveIndex(ctx *sow.Context, index *schemas.ExplorationIndex) error {
	fs := ctx.FS()
	if fs == nil {
		return sow.ErrNotInitialized
	}

	// Marshal to YAML
	data, err := yaml.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal exploration index: %w", err)
	}

	// Write atomically (write to temp file, then rename)
	tmpPath := IndexPath + ".tmp"
	if err := fs.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp index: %w", err)
	}

	if err := fs.Rename(tmpPath, IndexPath); err != nil {
		_ = fs.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp index: %w", err)
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
	fs := ctx.FS()
	if fs == nil {
		return false
	}
	exists, _ := fs.Exists("exploration")
	return exists
}

// Delete removes the exploration directory and all its contents.
func Delete(ctx *sow.Context) error {
	fs := ctx.FS()
	if fs == nil {
		return sow.ErrNotInitialized
	}

	exists, _ := fs.Exists("exploration")
	if !exists {
		return ErrNoExploration
	}

	if err := fs.RemoveAll("exploration"); err != nil {
		return fmt.Errorf("failed to remove exploration directory: %w", err)
	}

	return nil
}

// GetFilePath returns the absolute path to a file in the exploration directory.
func GetFilePath(ctx *sow.Context, relativePath string) string {
	return filepath.Join(ctx.RepoRoot(), ".sow", "exploration", relativePath)
}
