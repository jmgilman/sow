package statechart

import (
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
		return nil, err
	}

	var state ProjectState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, err
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
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(stateFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Atomic write: write to temp file, then rename
	tmpFile := stateFilePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, stateFilePath)
}
