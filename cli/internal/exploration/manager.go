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
