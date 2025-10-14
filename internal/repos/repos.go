package repos

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Index represents the repository index structure
type Index struct {
	Repositories []Repository `json:"repositories"`
}

// Repository represents a single repository entry
type Repository struct {
	Name      string  `json:"name"`
	Path      string  `json:"path"`
	Source    string  `json:"source"`
	Purpose   string  `json:"purpose"`
	Type      string  `json:"type"` // "clone" or "symlink"
	Branch    *string `json:"branch"`
	UpdatedAt string  `json:"updated_at"`
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

// AddRepository adds a new repository to the index
func (idx *Index) AddRepository(repo Repository) {
	idx.Repositories = append(idx.Repositories, repo)
}

// RemoveRepository removes a repository from the index by name
// Returns true if the repository was found and removed, false otherwise
func (idx *Index) RemoveRepository(name string) bool {
	for i, repo := range idx.Repositories {
		if repo.Name == name {
			idx.Repositories = append(idx.Repositories[:i], idx.Repositories[i+1:]...)
			return true
		}
	}
	return false
}

// FindRepository finds a repository by name
func (idx *Index) FindRepository(name string) (*Repository, bool) {
	for i := range idx.Repositories {
		if idx.Repositories[i].Name == name {
			return &idx.Repositories[i], true
		}
	}
	return nil, false
}

// ExtractNameFromSource extracts a name from a git URL or path
func ExtractNameFromSource(source string) string {
	// Remove trailing slash
	source = strings.TrimSuffix(source, "/")

	// Remove .git suffix
	source = strings.TrimSuffix(source, ".git")

	// Get the last part of the path
	parts := strings.Split(source, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "repo"
}

// IsGitSource checks if a string is a git URL (vs local path)
func IsGitSource(source string) bool {
	return strings.HasPrefix(source, "https://") ||
		strings.HasPrefix(source, "http://") ||
		strings.HasPrefix(source, "git@")
}

// CreateEmptyIndex creates an empty index file
func CreateEmptyIndex(path string) error {
	idx := &Index{Repositories: []Repository{}}
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
