package sowfs

import (
	"fmt"
	"path/filepath"
)

// KnowledgeFS provides access to the .sow/knowledge/ directory.
//
// This domain handles repository-specific documentation including:
//   - overview.md - Repository overview and onboarding
//   - architecture/ - Architecture documentation
//   - adrs/ - Architecture Decision Records
//
// All paths are relative to .sow/knowledge/
type KnowledgeFS interface {
	// ReadFile reads a file from the knowledge directory.
	// Path is relative to .sow/knowledge/
	ReadFile(path string) ([]byte, error)

	// WriteFile writes a file to the knowledge directory.
	// Creates parent directories if needed.
	// Path is relative to .sow/knowledge/
	WriteFile(path string, data []byte) error

	// Exists checks if a file or directory exists in knowledge.
	// Path is relative to .sow/knowledge/
	Exists(path string) (bool, error)

	// ListADRs returns a list of all ADR file paths in adrs/.
	// Returns paths relative to .sow/knowledge/adrs/
	ListADRs() ([]string, error)

	// ReadADR reads an ADR by its filename.
	// Filename should be just the file name (e.g., "001-architecture.md")
	// Equivalent to ReadFile("adrs/" + filename)
	ReadADR(filename string) ([]byte, error)

	// WriteADR writes an ADR file.
	// Filename should be just the file name (e.g., "001-architecture.md")
	// Equivalent to WriteFile("adrs/" + filename, data)
	WriteADR(filename string, data []byte) error

	// MkdirAll creates a directory path in knowledge.
	// Path is relative to .sow/knowledge/
	MkdirAll(path string) error
}

// KnowledgeFSImpl is the concrete implementation of KnowledgeFS.
type KnowledgeFSImpl struct {
	// sowFS is the parent SowFS (for accessing chrooted .sow filesystem)
	sowFS *SowFSImpl
}

// Ensure KnowledgeFSImpl implements KnowledgeFS
var _ KnowledgeFS = (*KnowledgeFSImpl)(nil)

// NewKnowledgeFS creates a new KnowledgeFS instance.
// validator is not used for KnowledgeFS as it only handles plain markdown files.
func NewKnowledgeFS(sowFS *SowFSImpl, validator interface{}) *KnowledgeFSImpl {
	return &KnowledgeFSImpl{
		sowFS: sowFS,
	}
}

// ReadFile reads a file from knowledge directory
func (k *KnowledgeFSImpl) ReadFile(path string) ([]byte, error) {
	fullPath := filepath.Join("knowledge", path)
	return k.sowFS.fs.ReadFile(fullPath)
}

// WriteFile writes a file to knowledge directory
func (k *KnowledgeFSImpl) WriteFile(path string, data []byte) error {
	fullPath := filepath.Join("knowledge", path)

	// Create parent directories if needed
	dir := filepath.Dir(fullPath)
	if err := k.sowFS.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	return k.sowFS.fs.WriteFile(fullPath, data, 0644)
}

// Exists checks if a path exists in knowledge
func (k *KnowledgeFSImpl) Exists(path string) (bool, error) {
	fullPath := filepath.Join("knowledge", path)
	return k.sowFS.fs.Exists(fullPath)
}

// ListADRs lists all ADR files
func (k *KnowledgeFSImpl) ListADRs() ([]string, error) {
	adrPath := "knowledge/adrs"

	// Check if adrs directory exists
	exists, err := k.sowFS.fs.Exists(adrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check adrs directory: %w", err)
	}
	if !exists {
		// No adrs directory means no ADRs
		return []string{}, nil
	}

	// Read directory entries
	entries, err := k.sowFS.fs.ReadDir(adrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read adrs directory: %w", err)
	}

	// Filter for files only (no directories)
	var adrs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			adrs = append(adrs, entry.Name())
		}
	}

	return adrs, nil
}

// ReadADR reads an ADR by filename
func (k *KnowledgeFSImpl) ReadADR(filename string) ([]byte, error) {
	// Sanitize filename to prevent path traversal
	filename = filepath.Base(filename)
	adrPath := filepath.Join("adrs", filename)
	return k.ReadFile(adrPath)
}

// WriteADR writes an ADR file
func (k *KnowledgeFSImpl) WriteADR(filename string, data []byte) error {
	// Sanitize filename to prevent path traversal
	filename = filepath.Base(filename)
	adrPath := filepath.Join("adrs", filename)
	return k.WriteFile(adrPath, data)
}

// MkdirAll creates directory path in knowledge
func (k *KnowledgeFSImpl) MkdirAll(path string) error {
	fullPath := filepath.Join("knowledge", path)
	return k.sowFS.fs.MkdirAll(fullPath, 0755)
}
