package state

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/libs/schemas/project"
	"gopkg.in/yaml.v3"
)

// defaultPath is the default location for project state files.
const defaultPath = "project/state.yaml"

// YAMLBackend implements Backend using YAML files on a core.FS filesystem.
type YAMLBackend struct {
	fs   core.FS
	path string // Relative path within fs (default: "project/state.yaml")
}

// NewYAMLBackend creates a backend that stores state in YAML files.
// The fs parameter should be rooted at the .sow directory.
// Uses the default path "project/state.yaml".
func NewYAMLBackend(fs core.FS) *YAMLBackend {
	return &YAMLBackend{
		fs:   fs,
		path: defaultPath,
	}
}

// NewYAMLBackendWithPath creates a backend with a custom file path.
// This is useful for testing or non-standard configurations.
func NewYAMLBackendWithPath(fs core.FS, path string) *YAMLBackend {
	return &YAMLBackend{
		fs:   fs,
		path: path,
	}
}

// Load reads project state from the YAML file.
func (b *YAMLBackend) Load(_ context.Context) (*project.ProjectState, error) {
	data, err := b.fs.ReadFile(b.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("load project state: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("load project state: %w", err)
	}

	var state project.ProjectState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal project state: %w", ErrInvalidState)
	}

	return &state, nil
}

// Save writes project state to the YAML file atomically.
// Uses temp file + rename pattern to ensure atomic writes.
func (b *YAMLBackend) Save(_ context.Context, state *project.ProjectState) error {
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal project state: %w", err)
	}

	tmpPath := b.path + ".tmp"
	if err := b.fs.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := b.fs.Rename(tmpPath, b.path); err != nil {
		_ = b.fs.Remove(tmpPath) // Clean up on failure
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// Exists checks if the project state file exists.
func (b *YAMLBackend) Exists(_ context.Context) (bool, error) {
	_, err := b.fs.Stat(b.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("check project state exists: %w", err)
	}
	return true, nil
}

// Delete removes the project state file.
func (b *YAMLBackend) Delete(_ context.Context) error {
	if err := b.fs.Remove(b.path); err != nil {
		return fmt.Errorf("delete project state: %w", err)
	}
	return nil
}

// Compile-time interface check.
var _ Backend = (*YAMLBackend)(nil)
