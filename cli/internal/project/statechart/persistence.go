package statechart

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
	"gopkg.in/yaml.v3"
)

const (
	// stateFilePath is the absolute path from repo root (for os operations).
	stateFilePath = ".sow/project/state.yaml"
	// stateFilePathChrooted is the relative path from .sow/ (for chrooted billy FS).
	stateFilePathChrooted = "project/state.yaml"
)

// LoadProjectState reads the project state from disk and returns the state and current state.
// This is used by the new composable phases architecture to load state before building the machine.
func LoadProjectState(fs sow.FS) (*schemas.ProjectState, State, error) {
	data, err := fs.ReadFile(stateFilePathChrooted)
	if err != nil {
		if os.IsNotExist(err) || err.Error() == "file does not exist" {
			return nil, NoProject, fmt.Errorf("no project state file found")
		}
		return nil, NoProject, fmt.Errorf("failed to read state file: %w", err)
	}

	var state schemas.ProjectState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, NoProject, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	currentState := State(state.Statechart.Current_state)
	return &state, currentState, nil
}

// LoadFS reads the state from disk using the provided filesystem.
// If no project exists, returns a machine in NoProject state.
// If fs is nil, falls back to using os.ReadFile directly.
//
// Deprecated: Use LoadProjectState() with types.DetectProjectType() and BuildStateMachine() instead.
// This function is kept for backward compatibility.
func LoadFS(fs sow.FS) (*Machine, error) {
	var data []byte
	var err error

	// Use filesystem if available, otherwise use os
	if fs != nil {
		data, err = fs.ReadFile(stateFilePathChrooted)
	} else {
		data, err = os.ReadFile(stateFilePath)
	}

	if err != nil {
		if os.IsNotExist(err) || err.Error() == "file does not exist" {
			// No project yet - start with NoProject state
			m := NewMachine(nil)
			m.fs = fs
			return m, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state schemas.ProjectState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Create machine starting from the stored state
	currentState := State(state.Statechart.Current_state)
	m := NewMachineAt(currentState, &state)
	m.fs = fs
	return m, nil
}

// NewProjectState creates an initialized project state with default values.
// This is a helper function for creating new projects.
func NewProjectState(name, description, branch string) *schemas.ProjectState {
	now := time.Now()

	state := &schemas.ProjectState{}

	// Statechart metadata
	// StandardProject starts in PlanningActive state
	state.Statechart.Current_state = string(PlanningActive)

	// Project metadata
	state.Project.Type = "standard"
	state.Project.Name = name
	state.Project.Branch = branch
	state.Project.Description = description
	state.Project.Created_at = now
	state.Project.Updated_at = now

	// Planning phase (required, always enabled)
	state.Phases.Planning.Enabled = true
	state.Phases.Planning.Status = "in_progress"
	state.Phases.Planning.Created_at = now
	state.Phases.Planning.Artifacts = []phases.Artifact{}

	// Implementation phase (required, enabled by default)
	state.Phases.Implementation.Enabled = true
	state.Phases.Implementation.Status = "pending"
	state.Phases.Implementation.Created_at = now
	state.Phases.Implementation.Tasks = []phases.Task{}
	state.Phases.Implementation.Metadata = map[string]interface{}{
		"tasks_approved": false,
	}

	// Review phase (required, enabled by default)
	state.Phases.Review.Enabled = true
	state.Phases.Review.Status = "pending"
	state.Phases.Review.Created_at = now
	state.Phases.Review.Artifacts = []phases.Artifact{}
	state.Phases.Review.Metadata = map[string]interface{}{
		"iteration": 1,
	}

	// Finalize phase (required, enabled by default)
	state.Phases.Finalize.Enabled = true
	state.Phases.Finalize.Status = "pending"
	state.Phases.Finalize.Created_at = now
	state.Phases.Finalize.Artifacts = []phases.Artifact{}
	state.Phases.Finalize.Metadata = map[string]interface{}{
		"project_deleted": false,
	}

	return state
}

// NewWithProject creates a new project state machine with initial project metadata.
// This is used when creating a new project.
//
// Deprecated: Use types.DetectProjectType() and BuildStateMachine() instead.
// This function is kept for backward compatibility.
func NewWithProject(name, description, branch string, fs sow.FS) (*Machine, error) {
	state := NewProjectState(name, description, branch)

	m := NewMachine(state)
	m.fs = fs
	return m, nil
}

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
