package state

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/libs/schemas/project"
)

// Load reads project state from a backend and returns an initialized Project.
//
// The Load pipeline:
//  1. Load raw ProjectState from backend
//  2. Validate structure with CUE
//  3. Create Project wrapper
//  4. Lookup and attach ProjectTypeConfig from registry
//  5. Build state machine initialized with current state
//  6. Validate metadata against embedded schemas
//
// Returns an error if any step fails. Error messages are actionable and
// indicate which step failed and why.
func Load(ctx context.Context, backend Backend) (*Project, error) {
	// 1. Load raw ProjectState from backend
	projectState, err := backend.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load project state: %w", err)
	}

	// 2. Validate structure with CUE
	if err := validateStructure(projectState); err != nil {
		return nil, fmt.Errorf("validate structure: %w", err)
	}

	// 3. Create Project wrapper
	proj := NewProject(*projectState, backend)

	// 4. Lookup and attach ProjectTypeConfig from registry
	config, exists := GetConfig(projectState.Type)
	if !exists {
		return nil, fmt.Errorf("unknown project type: %s", projectState.Type)
	}
	proj.SetConfig(config)

	// 5. Build state machine initialized with current state
	initialState := projectState.Statechart.Current_state
	machine := config.BuildMachine(proj, initialState)
	proj.SetMachine(machine)

	// 6. Validate metadata against embedded schemas
	if err := config.Validate(proj); err != nil {
		return nil, fmt.Errorf("validate metadata: %w", err)
	}

	return proj, nil
}

// LoadFromFS loads a project from a YAML file on the filesystem.
// This is a convenience wrapper around Load with YAMLBackend.
//
// The fs parameter should be rooted at the .sow directory.
func LoadFromFS(ctx context.Context, fs core.FS) (*Project, error) {
	backend := NewYAMLBackend(fs)
	return Load(ctx, backend)
}

// Save validates and writes project state to the backend.
//
// The Save pipeline:
//  1. Sync statechart state from machine (if present)
//  2. Update timestamps
//  3. Validate structure with CUE
//  4. Validate metadata with embedded schemas
//  5. Save to backend
//
// Returns an error if any step fails. Validation errors prevent writing,
// ensuring the state file always contains valid data.
func Save(ctx context.Context, p *Project) error {
	// 1. Sync statechart state from machine
	if p.machine != nil {
		currentState := p.machine.MustState()
		if stateStr, ok := currentState.(string); ok {
			p.Statechart.Current_state = stateStr
		}
		p.Statechart.Updated_at = timeNow()
	}

	// 2. Update timestamp
	p.Updated_at = timeNow()

	// 3. Validate structure with CUE
	if err := validateStructure(&p.ProjectState); err != nil {
		return fmt.Errorf("validate structure: %w", err)
	}

	// 4. Validate metadata with embedded schemas
	if p.config != nil {
		if err := p.config.Validate(p); err != nil {
			return fmt.Errorf("validate metadata: %w", err)
		}
	}

	// 5. Save to backend
	if err := p.backend.Save(ctx, &p.ProjectState); err != nil {
		return fmt.Errorf("save to backend: %w", err)
	}

	return nil
}

// timeNow returns the current time. It is a variable to allow tests to override it.
var timeNow = time.Now

// CreateOpts contains options for creating a new project.
type CreateOpts struct {
	// Branch is the git branch for this project (required).
	Branch string

	// Description is a human-readable description (required).
	Description string

	// ProjectType overrides the type detected from branch name.
	// If empty, type is detected from branch prefix.
	ProjectType string

	// InitialInputs are artifacts to pass to the initializer.
	// Keys are phase names, values are artifact lists.
	InitialInputs map[string][]project.ArtifactState
}

// Create initializes a new project and saves it to the backend.
//
// The Create pipeline:
//  1. Detect or use explicit project type
//  2. Lookup ProjectTypeConfig from registry
//  3. Create Project wrapper with initial metadata
//  4. Initialize via config (creates phases, sets up structure)
//  5. Build state machine with config's initial state
//  6. Mark initial phase as in_progress if applicable
//  7. Validate and save to backend
//
// Returns the created project or an error if any step fails.
func Create(ctx context.Context, backend Backend, opts CreateOpts) (*Project, error) {
	// 1. Detect or use explicit project type
	projectType := opts.ProjectType
	if projectType == "" {
		projectType = detectProjectType(opts.Branch)
	}

	// 2. Lookup ProjectTypeConfig from registry
	config, exists := GetConfig(projectType)
	if !exists {
		return nil, fmt.Errorf("unknown project type: %s", projectType)
	}

	// 3. Create Project wrapper with initial metadata
	now := timeNow()
	projectState := project.ProjectState{
		Name:        generateProjectName(opts.Description),
		Type:        projectType,
		Branch:      opts.Branch,
		Description: opts.Description,
		Created_at:  now,
		Updated_at:  now,
		Phases:      make(map[string]project.PhaseState),
		Statechart: project.StatechartState{
			Current_state: config.InitialState(),
			Updated_at:    now,
		},
	}
	proj := NewProject(projectState, backend)
	proj.SetConfig(config)

	// 4. Initialize via config (creates phases, sets up structure)
	if err := config.Initialize(proj, opts.InitialInputs); err != nil {
		return nil, fmt.Errorf("initialize project: %w", err)
	}

	// 5. Build state machine with config's initial state
	initialState := config.InitialState()
	machine := config.BuildMachine(proj, initialState)
	proj.SetMachine(machine)

	// 6. Mark initial phase as in_progress if applicable
	if err := markInitialPhaseInProgress(proj, config, initialState); err != nil {
		return nil, fmt.Errorf("mark initial phase: %w", err)
	}

	// 7. Validate and save to backend
	if err := Save(ctx, proj); err != nil {
		return nil, err
	}

	return proj, nil
}

// CreateOnFS creates a new project and saves it to YAML on the filesystem.
// This is a convenience wrapper around Create with YAMLBackend.
//
// The fs parameter should be rooted at the .sow directory.
func CreateOnFS(ctx context.Context, fs core.FS, opts CreateOpts) (*Project, error) {
	backend := NewYAMLBackend(fs)
	return Create(ctx, backend, opts)
}

// markInitialPhaseInProgress marks the initial phase as in_progress.
// This is called during Create to set up the initial phase state.
func markInitialPhaseInProgress(p *Project, config ProjectTypeConfig, initialState string) error {
	// Find which phase the initial state belongs to
	phaseName := config.GetPhaseForState(initialState)
	if phaseName == "" {
		return nil // No phase for this state
	}

	// Only mark if this is the start state for the phase
	if !config.IsPhaseStartState(phaseName, initialState) {
		return nil
	}

	return MarkPhaseInProgress(p, phaseName)
}

// detectProjectType determines project type from branch name.
// Returns "exploration" for "explore/" prefix, "design" for "design/" prefix,
// "breakdown" for "breakdown/" prefix, and "standard" for all other cases.
func detectProjectType(branchName string) string {
	switch {
	case strings.HasPrefix(branchName, "explore/"):
		return "exploration"
	case strings.HasPrefix(branchName, "design/"):
		return "design"
	case strings.HasPrefix(branchName, "breakdown/"):
		return "breakdown"
	default:
		return "standard"
	}
}

// generateProjectName converts a description to a kebab-case project name.
// It truncates to 50 characters, converts to lowercase, replaces spaces and
// underscores with hyphens, removes special characters, and removes trailing hyphens.
func generateProjectName(description string) string {
	name := description
	if len(name) > 50 {
		name = name[:50]
	}

	// Convert to kebab-case
	result := ""
	for i, r := range name {
		if r == ' ' || r == '_' {
			if i > 0 && len(result) > 0 && result[len(result)-1] != '-' {
				result += "-"
			}
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else if r >= 'A' && r <= 'Z' {
			result += string(r + 32) // Convert to lowercase
		}
	}

	// Remove trailing hyphen
	if len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}

	return result
}
