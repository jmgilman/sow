package design

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"gopkg.in/yaml.v3"
)

const (
	// IndexPath is the path to the design index relative to .sow/.
	IndexPath = "design/index.yaml"
)

// LoadIndex loads the design index from disk.
// Returns ErrNoDesign if design directory doesn't exist.
func LoadIndex(ctx *sow.Context) (*schemas.DesignIndex, error) {
	fs := ctx.FS()
	if fs == nil {
		return nil, sow.ErrNotInitialized
	}

	// Check if design directory exists
	exists, err := fs.Exists("design")
	if err != nil {
		return nil, fmt.Errorf("failed to check design directory: %w", err)
	}
	if !exists {
		return nil, ErrNoDesign
	}

	// Read index file
	data, err := fs.ReadFile(IndexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read design index: %w", err)
	}

	// Parse YAML
	var index schemas.DesignIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse design index: %w", err)
	}

	return &index, nil
}

// SaveIndex saves the design index to disk.
func SaveIndex(ctx *sow.Context, index *schemas.DesignIndex) error {
	fs := ctx.FS()
	if fs == nil {
		return sow.ErrNotInitialized
	}

	// Marshal to YAML
	data, err := yaml.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal design index: %w", err)
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
	fs := ctx.FS()
	if fs == nil {
		return false
	}
	exists, _ := fs.Exists("design")
	return exists
}

// Delete removes the design directory and all its contents.
func Delete(ctx *sow.Context) error {
	fs := ctx.FS()
	if fs == nil {
		return sow.ErrNotInitialized
	}

	exists, _ := fs.Exists("design")
	if !exists {
		return ErrNoDesign
	}

	if err := fs.RemoveAll("design"); err != nil {
		return fmt.Errorf("failed to remove design directory: %w", err)
	}

	return nil
}

// GetFilePath returns the absolute path to a file in the design directory.
func GetFilePath(ctx *sow.Context, relativePath string) string {
	return filepath.Join(ctx.RepoRoot(), ".sow", "design", relativePath)
}
