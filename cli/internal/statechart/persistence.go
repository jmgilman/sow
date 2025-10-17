package statechart

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const stateFilePath = ".sow/project/state.yaml"

// Load reads the state from disk and creates a state machine.
// If no project exists, returns a machine in NoProject state.
func Load() (*Machine, error) {
	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No project yet - start with NoProject state
			return NewMachine(nil), nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state ProjectState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Create machine starting from the stored state
	m := NewMachineAt(state.Statechart.CurrentState, &state)
	return m, nil
}

// Save writes the current state to disk atomically.
func (m *Machine) Save() error {
	if m.projectState == nil {
		return nil // Nothing to save
	}

	// Update the statechart metadata with current state
	m.projectState.Statechart.CurrentState = m.State()

	// Marshal to YAML
	data, err := yaml.Marshal(m.projectState)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

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
