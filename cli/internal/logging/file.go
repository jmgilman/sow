package logging

import (
	"errors"
	"fmt"
	"os"

	"github.com/jmgilman/go/fs/core"
)

// FileSystem is the interface required for file operations.
// Supports both variants: with and without perm parameter.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm ...os.FileMode) error
}

// SimpleFileSystem wraps a filesystem that doesn't take perm parameter.
type SimpleFileSystem struct {
	ReadFileFn  func(path string) ([]byte, error)
	WriteFileFn func(path string, data []byte) error
}

// ReadFile implements FileSystem.ReadFile by delegating to ReadFileFn.
func (s SimpleFileSystem) ReadFile(path string) ([]byte, error) {
	return s.ReadFileFn(path)
}

// WriteFile implements FileSystem.WriteFile by delegating to WriteFileFn.
func (s SimpleFileSystem) WriteFile(path string, data []byte, _ ...os.FileMode) error {
	return s.WriteFileFn(path, data)
}

// AppendLog appends a log entry to the specified log file.
// The log file is created if it doesn't exist.
func AppendLog(fs interface{}, logPath string, entry *LogEntry) error {
	// Validate entry
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("invalid log entry: %w", err)
	}

	// Format the entry
	formatted := entry.Format()

	// Wrap filesystem if needed to match interface
	var reader func(string) ([]byte, error)
	var writer func(string, []byte) error

	switch f := fs.(type) {
	case FileSystem:
		reader = f.ReadFile
		writer = func(path string, data []byte) error {
			return f.WriteFile(path, data)
		}
	case interface {
		ReadFile(string) ([]byte, error)
		WriteFile(string, []byte) error
	}:
		reader = f.ReadFile
		writer = f.WriteFile
	case interface {
		ReadFile(string) ([]byte, error)
		WriteFile(string, []byte, os.FileMode) error
	}:
		reader = f.ReadFile
		writer = func(path string, data []byte) error {
			return f.WriteFile(path, data, 0644)
		}
	default:
		return fmt.Errorf("unsupported filesystem type")
	}

	// Read existing content (if any)
	existing, err := reader(logPath)
	if err != nil {
		// If file doesn't exist, start with empty content
		if !isNotExist(err) {
			return fmt.Errorf("failed to read log file: %w", err)
		}
		existing = []byte{}
	}

	// Append new entry
	updated := append(existing, []byte(formatted)...)

	// Write back
	if err := writer(logPath, updated); err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	return nil
}

// isNotExist checks if an error is a "not exist" error.
func isNotExist(err error) bool {
	if err == nil {
		return false
	}
	// Check for core.ErrNotExist or os.ErrNotExist
	return errors.Is(err, core.ErrNotExist) || os.IsNotExist(err)
}
