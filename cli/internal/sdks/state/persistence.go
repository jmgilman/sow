package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// stateFilePath is the absolute path from repo root (for os operations).
	stateFilePath = ".sow/project/state.yaml"
	// stateFilePathChrooted is the relative path from .sow/ (for chrooted billy FS).
	stateFilePathChrooted = "project/state.yaml"
)

// Save writes the current state to disk atomically.
func (m *Machine) Save() error {
	if m.projectState == nil {
		return nil // Nothing to save
	}

	// Update timestamps
	m.projectState.Project.Updated_at = time.Now()

	// Update the statechart metadata with current state
	m.projectState.Statechart.Current_state = string(m.State())

	// Marshal to YAML with proper null handling
	// Note: yaml.v3 writes nil pointers as "null", but CUE validation rejects this.
	// We use a custom encoder to properly omit nil fields.
	encoder := yaml.NewEncoder(nil)
	encoder.SetIndent(2)

	// Encode to a node first, then customize it to remove null values
	var node yaml.Node
	if err := node.Encode(m.projectState); err != nil {
		return fmt.Errorf("failed to encode state: %w", err)
	}

	// Remove null values from the node tree
	removeNullNodes(&node)

	// Marshal the cleaned node
	data, err := yaml.Marshal(&node)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Use filesystem if available, otherwise use os
	if m.fs != nil {
		return m.saveFS(data)
	}
	return m.saveOS(data)
}

// removeNullNodes recursively removes null value nodes from a YAML node tree.
// This ensures that optional fields with nil pointers are omitted rather than written as "null".
func removeNullNodes(node *yaml.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		// For documents and sequences, recurse into content
		for _, child := range node.Content {
			removeNullNodes(child)
		}
	case yaml.MappingNode:
		// For mappings, filter out key-value pairs where value is null
		filtered := make([]*yaml.Node, 0, len(node.Content))
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			// Skip this pair if value is null
			if value.Kind == yaml.ScalarNode && value.Tag == "!!null" {
				continue
			}

			// Recursively clean the value
			removeNullNodes(value)

			// Keep this key-value pair
			filtered = append(filtered, key, value)
		}
		node.Content = filtered
	}
}

// saveFS saves using sow.FS filesystem (assumes fs is already chrooted to .sow/).
func (m *Machine) saveFS(data []byte) error {
	// Use chrooted path
	path := stateFilePathChrooted

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := m.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpFile := path + ".tmp"
	if err := m.fs.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	if err := m.fs.Rename(tmpFile, path); err != nil {
		_ = m.fs.Remove(tmpFile) // Clean up
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// saveOS saves using os package.
func (m *Machine) saveOS(data []byte) error {
	// Ensure directory exists
	dir := filepath.Dir(stateFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpFile := stateFilePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	if err := os.Rename(tmpFile, stateFilePath); err != nil {
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}
