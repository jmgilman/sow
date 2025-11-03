package state

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/project"
	"gopkg.in/yaml.v3"
)

// Load reads project state from disk and returns an initialized Project.
//
// The Load pipeline:
//  1. Read YAML file from .sow/project/state.yaml
//  2. Unmarshal into CUE-generated ProjectState
//  3. Validate structure with CUE
//  4. Convert to wrapper types (Project, Phase collections)
//  5. Lookup and attach ProjectTypeConfig from registry
//  6. Build state machine initialized with current state
//  7. Validate metadata against embedded schemas
//
// Returns an error if any step fails. Error messages are actionable and
// indicate which step failed and why.
func Load(ctx *sow.Context) (*Project, error) {
	// 1. Read YAML file using FS abstraction
	data, err := ctx.FS().ReadFile("project/state.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read state: %w", err)
	}

	// 2. Unmarshal into CUE-generated type
	var projectState project.ProjectState
	if err := yaml.Unmarshal(data, &projectState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	// 3. Validate structure with CUE
	if err := validateStructure(projectState); err != nil {
		return nil, fmt.Errorf("CUE validation failed: %w", err)
	}

	// 4. Convert to wrapper types
	proj := &Project{
		ProjectState: projectState, // Embed CUE type
		ctx:          ctx,          // Store context for Save()
	}

	// 5. Lookup and attach type config
	config, exists := Registry[proj.Type]
	if !exists {
		return nil, fmt.Errorf("unknown project type: %s", proj.Type)
	}
	proj.config = config

	// 6. Build state machine initialized with current state
	initialState := State(proj.Statechart.Current_state)
	proj.machine = config.BuildMachine(proj, initialState)

	// 7. Validate metadata against embedded schemas
	if err := config.Validate(proj); err != nil {
		return nil, fmt.Errorf("metadata validation failed: %w", err)
	}

	return proj, nil
}

// Save validates and writes project state to disk atomically.
//
// The Save pipeline:
//  1. Sync statechart state from machine (if present)
//  2. Update timestamps
//  3. Validate structure with CUE
//  4. Validate metadata with embedded schemas
//  5. Marshal to YAML
//  6. Atomic write (temp file + rename)
//
// Returns an error if any step fails. Validation errors prevent writing,
// ensuring the state file always contains valid data.
func (p *Project) Save() error {
	// 1. Sync statechart state from machine
	if p.machine != nil {
		p.Statechart.Current_state = p.machine.State().String()
		p.Statechart.Updated_at = time.Now()
	}

	// 2. Update timestamp
	p.Updated_at = time.Now()

	// 3. Validate structure with CUE
	// ProjectState is embedded, so we can validate p directly
	if err := validateStructure(p.ProjectState); err != nil {
		return fmt.Errorf("CUE validation failed: %w", err)
	}

	// 4. Validate metadata with embedded schemas
	if p.config != nil {
		if err := p.config.Validate(p); err != nil {
			return fmt.Errorf("metadata validation failed: %w", err)
		}
	}

	// 5. Marshal to YAML
	data, err := yaml.Marshal(p.ProjectState)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	// 6. Atomic write using FS abstraction (temp file + rename)
	fs := p.ctx.FS()
	path := "project/state.yaml"
	tmpPath := path + ".tmp"

	if err := fs.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := fs.Rename(tmpPath, path); err != nil {
		// Clean up temp file on rename failure
		_ = fs.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
