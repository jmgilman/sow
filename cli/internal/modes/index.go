package modes

import (
	"fmt"
	"path/filepath"

	"github.com/jmgilman/sow/cli/internal/sow"
	"gopkg.in/yaml.v3"
)

// IndexManager provides generic CRUD operations for mode index files.
// T is the schema type for the index (e.g., schemas.ExplorationIndex, schemas.DesignIndex).
type IndexManager[T any] struct {
	directoryName string
	indexPath     string
}

// NewIndexManager creates a new generic index manager.
func NewIndexManager[T any](directoryName, indexPath string) *IndexManager[T] {
	return &IndexManager[T]{
		directoryName: directoryName,
		indexPath:     indexPath,
	}
}

// Load loads the index from disk.
// Returns an error if the directory doesn't exist.
func (m *IndexManager[T]) Load(ctx *sow.Context) (*T, error) {
	fs := ctx.FS()
	if fs == nil {
		return nil, sow.ErrNotInitialized
	}

	// Check if directory exists
	exists, err := fs.Exists(m.directoryName)
	if err != nil {
		return nil, fmt.Errorf("failed to check %s directory: %w", m.directoryName, err)
	}
	if !exists {
		return nil, fmt.Errorf("%s does not exist", m.directoryName)
	}

	// Read index file
	data, err := fs.ReadFile(m.indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	// Parse YAML
	var index T
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return &index, nil
}

// Save saves the index to disk.
// Uses atomic write (temp file + rename) for safety.
func (m *IndexManager[T]) Save(ctx *sow.Context, index *T) error {
	fs := ctx.FS()
	if fs == nil {
		return sow.ErrNotInitialized
	}

	// Marshal to YAML
	data, err := yaml.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	// Write atomically (write to temp file, then rename)
	tmpPath := m.indexPath + ".tmp"
	if err := fs.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp index: %w", err)
	}

	if err := fs.Rename(tmpPath, m.indexPath); err != nil {
		_ = fs.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp index: %w", err)
	}

	return nil
}

// Exists checks if the mode directory exists.
func (m *IndexManager[T]) Exists(ctx *sow.Context) bool {
	fs := ctx.FS()
	if fs == nil {
		return false
	}
	exists, _ := fs.Exists(m.directoryName)
	return exists
}

// Delete removes the directory and all its contents.
func (m *IndexManager[T]) Delete(ctx *sow.Context) error {
	fs := ctx.FS()
	if fs == nil {
		return sow.ErrNotInitialized
	}

	exists, _ := fs.Exists(m.directoryName)
	if !exists {
		return fmt.Errorf("%s does not exist", m.directoryName)
	}

	if err := fs.RemoveAll(m.directoryName); err != nil {
		return fmt.Errorf("failed to remove %s directory: %w", m.directoryName, err)
	}

	return nil
}

// GetFilePath returns the absolute path to a file in the mode directory.
func (m *IndexManager[T]) GetFilePath(ctx *sow.Context, relativePath string) string {
	return filepath.Join(ctx.RepoRoot(), ".sow", m.directoryName, relativePath)
}
