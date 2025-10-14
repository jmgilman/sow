package sinks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Index represents the sink index structure
type Index struct {
	Sinks []Sink `json:"sinks"`
}

// Sink represents a single sink entry
type Sink struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
	WhenToUse   string   `json:"when_to_use"`
	Version     string   `json:"version"`
	Source      string   `json:"source"`
	UpdatedAt   string   `json:"updated_at"`
}

// Load reads the index from a JSON file
func (idx *Index) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	if err := json.Unmarshal(data, idx); err != nil {
		return fmt.Errorf("failed to parse index JSON: %w", err)
	}

	return nil
}

// Save writes the index to a JSON file
func (idx *Index) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

// AddSink adds a new sink to the index
func (idx *Index) AddSink(sink Sink) {
	idx.Sinks = append(idx.Sinks, sink)
}

// RemoveSink removes a sink from the index by name
// Returns true if the sink was found and removed, false otherwise
func (idx *Index) RemoveSink(name string) bool {
	for i, sink := range idx.Sinks {
		if sink.Name == name {
			idx.Sinks = append(idx.Sinks[:i], idx.Sinks[i+1:]...)
			return true
		}
	}
	return false
}

// FindSink finds a sink by name
func (idx *Index) FindSink(name string) (*Sink, bool) {
	for i := range idx.Sinks {
		if idx.Sinks[i].Name == name {
			return &idx.Sinks[i], true
		}
	}
	return nil, false
}

// ExtractNameFromURL extracts a name from a git URL or path
func ExtractNameFromURL(url string) string {
	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Get the last part of the path
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "sink"
}

// IsGitURL checks if a string is a git URL (vs local path)
func IsGitURL(url string) bool {
	return strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "git@")
}

// CreateEmptyIndex creates an empty index file
func CreateEmptyIndex(path string) error {
	idx := &Index{Sinks: []Sink{}}
	return idx.Save(path)
}

// LoadOrCreate loads an index file, or creates an empty one if it doesn't exist
func LoadOrCreate(path string) (*Index, error) {
	idx := &Index{}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create empty index
		if err := CreateEmptyIndex(path); err != nil {
			return nil, err
		}
		return idx, nil
	}

	// Load existing index
	if err := idx.Load(path); err != nil {
		return nil, err
	}

	return idx, nil
}
