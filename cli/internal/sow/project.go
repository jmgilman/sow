package sow

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jmgilman/sow/cli/internal/statechart"
	"github.com/jmgilman/sow/cli/schemas"
)

// Project represents an active sow project with full state machine integration.
// All operations automatically persist state changes to disk.
type Project struct {
	sow     *Sow
	machine *statechart.Machine
}

// Machine returns the underlying state machine for advanced operations.
// Most callers should use the higher-level Project methods instead.
func (p *Project) Machine() *statechart.Machine {
	return p.machine
}

// State returns the project state for read-only access.
func (p *Project) State() *schemas.ProjectState {
	return p.machine.ProjectState()
}

// Name returns the project name.
func (p *Project) Name() string {
	return p.State().Project.Name
}

// Branch returns the project's git branch.
func (p *Project) Branch() string {
	return p.State().Project.Branch
}

// Description returns the project description.
func (p *Project) Description() string {
	return p.State().Project.Description
}

// save persists the current state to disk atomically.
// This is called automatically after all mutations.
func (p *Project) save() error {
	return p.machine.Save()
}

// EnablePhase enables a phase and transitions the state machine.
func (p *Project) EnablePhase(phaseName string, opts ...PhaseOption) error {
	// Apply options
	cfg := &phaseConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	state := p.State()
	now := time.Now()

	// Handle each phase
	switch phaseName {
	case "discovery":
		if cfg.discoveryType == "" {
			return fmt.Errorf("discovery type required (use WithDiscoveryType option)")
		}
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "pending"
		state.Phases.Discovery.Discovery_type = &cfg.discoveryType
		state.Phases.Discovery.Started_at = &now

		// Create discovery directory structure
		if err := p.createPhaseStructure("discovery"); err != nil {
			return err
		}

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventEnableDiscovery); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "design":
		state.Phases.Design.Enabled = true
		state.Phases.Design.Status = "pending"
		state.Phases.Design.Started_at = &now

		// Create design directory structure
		if err := p.createPhaseStructure("design"); err != nil {
			return err
		}

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventEnableDesign); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	default:
		return ErrInvalidPhase
	}

	// Auto-save
	return p.save()
}

// CompletePhase marks a phase as completed and transitions the state machine.
func (p *Project) CompletePhase(phaseName string) error {
	state := p.State()
	now := time.Now()

	switch phaseName {
	case "discovery":
		if !state.Phases.Discovery.Enabled {
			return ErrPhaseNotEnabled
		}
		state.Phases.Discovery.Status = "completed"
		state.Phases.Discovery.Completed_at = &now

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventCompleteDiscovery); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "design":
		if !state.Phases.Design.Enabled {
			return ErrPhaseNotEnabled
		}
		state.Phases.Design.Status = "completed"
		state.Phases.Design.Completed_at = &now

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventCompleteDesign); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "implementation":
		state.Phases.Implementation.Status = "completed"
		state.Phases.Implementation.Completed_at = &now

		// Fire state machine event (transitions to review)
		if err := p.machine.Fire(statechart.EventAllTasksComplete); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "review":
		state.Phases.Review.Status = "completed"
		state.Phases.Review.Completed_at = &now

		// Fire state machine event (handled by review pass/fail)
		// This is a placeholder - actual review completion happens via AddReviewReport

	case "finalize":
		state.Phases.Finalize.Status = "completed"
		state.Phases.Finalize.Completed_at = &now

		// Finalize has substates, handled by specialized methods

	default:
		return ErrInvalidPhase
	}

	// Auto-save
	return p.save()
}

// AddTask creates a new implementation task.
func (p *Project) AddTask(name string, opts ...TaskOption) (*Task, error) {
	state := p.State()

	// Generate task ID (gap-numbered: 010, 020, 030...)
	id := p.generateTaskID()

	// Apply options
	cfg := &taskConfig{
		status:   "pending",
		parallel: false,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Create task
	task := schemas.Task{
		Id:           id,
		Name:         name,
		Status:       cfg.status,
		Parallel:     cfg.parallel,
		Dependencies: cfg.dependencies,
	}

	// Add to state
	state.Phases.Implementation.Tasks = append(state.Phases.Implementation.Tasks, task)

	// Create task directory structure
	if err := p.createTaskStructure(id); err != nil {
		return nil, fmt.Errorf("failed to create task structure: %w", err)
	}

	// If this is the first task, fire event to transition from Planning to Executing
	if len(state.Phases.Implementation.Tasks) == 1 {
		if err := p.machine.Fire(statechart.EventTaskCreated); err != nil {
			return nil, fmt.Errorf("state transition failed: %w", err)
		}
	}

	// Auto-save
	if err := p.save(); err != nil {
		return nil, err
	}

	return &Task{
		project: p,
		id:      id,
	}, nil
}

// GetTask retrieves a task by ID.
func (p *Project) GetTask(id string) (*Task, error) {
	state := p.State()

	// Find task in state
	for _, t := range state.Phases.Implementation.Tasks {
		if t.Id == id {
			return &Task{
				project: p,
				id:      id,
			}, nil
		}
	}

	return nil, ErrNoTask
}

// ListTasks returns all tasks.
func (p *Project) ListTasks() []*Task {
	state := p.State()
	tasks := make([]*Task, 0, len(state.Phases.Implementation.Tasks))

	for _, t := range state.Phases.Implementation.Tasks {
		tasks = append(tasks, &Task{
			project: p,
			id:      t.Id,
		})
	}

	return tasks
}

// InferTaskID attempts to infer the task ID from context.
// Returns the task ID from the current directory or the active in_progress task.
func (p *Project) InferTaskID() (string, error) {
	// TODO: Implement directory-based inference
	// For now, check for single in_progress task

	state := p.State()
	var inProgressID string
	count := 0

	for _, t := range state.Phases.Implementation.Tasks {
		if t.Status == "in_progress" {
			inProgressID = t.Id
			count++
		}
	}

	if count == 0 {
		return "", fmt.Errorf("no in_progress task found")
	}

	if count > 1 {
		return "", fmt.Errorf("multiple in_progress tasks found, use --id flag")
	}

	return inProgressID, nil
}

// AddArtifact adds an artifact to a phase (discovery or design).
func (p *Project) AddArtifact(phaseName, path string, approved bool) error {
	state := p.State()
	now := time.Now()

	artifact := schemas.Artifact{
		Path:       path,
		Approved:   approved,
		Created_at: now,
	}

	switch phaseName {
	case "discovery":
		if !state.Phases.Discovery.Enabled {
			return ErrPhaseNotEnabled
		}
		state.Phases.Discovery.Artifacts = append(state.Phases.Discovery.Artifacts, artifact)

	case "design":
		if !state.Phases.Design.Enabled {
			return ErrPhaseNotEnabled
		}
		state.Phases.Design.Artifacts = append(state.Phases.Design.Artifacts, artifact)

	default:
		return ErrInvalidPhase
	}

	// Auto-save
	return p.save()
}

// ApproveArtifact marks an artifact as approved.
func (p *Project) ApproveArtifact(phaseName, path string) error {
	state := p.State()

	switch phaseName {
	case "discovery":
		if !state.Phases.Discovery.Enabled {
			return ErrPhaseNotEnabled
		}
		for i := range state.Phases.Discovery.Artifacts {
			if state.Phases.Discovery.Artifacts[i].Path == path {
				state.Phases.Discovery.Artifacts[i].Approved = true
				return p.save()
			}
		}

	case "design":
		if !state.Phases.Design.Enabled {
			return ErrPhaseNotEnabled
		}
		for i := range state.Phases.Design.Artifacts {
			if state.Phases.Design.Artifacts[i].Path == path {
				state.Phases.Design.Artifacts[i].Approved = true
				return p.save()
			}
		}

	default:
		return ErrInvalidPhase
	}

	return fmt.Errorf("artifact not found: %s", path)
}

// IncrementReviewIteration increments the review iteration counter.
func (p *Project) IncrementReviewIteration() error {
	state := p.State()
	state.Phases.Review.Iteration++
	return p.save()
}

// AddReviewReport adds a review report and transitions state based on assessment.
func (p *Project) AddReviewReport(path, assessment string) error {
	state := p.State()
	now := time.Now()

	// Validate assessment
	if assessment != "pass" && assessment != "fail" {
		return fmt.Errorf("invalid assessment: must be 'pass' or 'fail'")
	}

	// Generate report ID (001, 002, 003...)
	reportID := fmt.Sprintf("%03d", len(state.Phases.Review.Reports)+1)

	// Create report
	report := schemas.ReviewReport{
		Id:         reportID,
		Path:       path,
		Created_at: now,
		Assessment: assessment,
	}

	state.Phases.Review.Reports = append(state.Phases.Review.Reports, report)

	// Fire state machine event based on assessment
	if assessment == "pass" {
		if err := p.machine.Fire(statechart.EventReviewPass); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}
	} else {
		if err := p.machine.Fire(statechart.EventReviewFail); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}
	}

	// Auto-save
	return p.save()
}

// AddDocumentation records a documentation file update during finalize.
func (p *Project) AddDocumentation(path string) error {
	state := p.State()

	// Handle the any type for Documentation_updates
	if state.Phases.Finalize.Documentation_updates == nil {
		state.Phases.Finalize.Documentation_updates = []string{path}
	} else {
		// Type assert to []string
		if updates, ok := state.Phases.Finalize.Documentation_updates.([]string); ok {
			state.Phases.Finalize.Documentation_updates = append(updates, path)
		} else if updates, ok := state.Phases.Finalize.Documentation_updates.([]interface{}); ok {
			// Handle interface{} slice from YAML unmarshaling
			strUpdates := make([]string, 0, len(updates)+1)
			for _, u := range updates {
				if s, ok := u.(string); ok {
					strUpdates = append(strUpdates, s)
				}
			}
			strUpdates = append(strUpdates, path)
			state.Phases.Finalize.Documentation_updates = strUpdates
		} else {
			// If type is unexpected, reset to new slice
			state.Phases.Finalize.Documentation_updates = []string{path}
		}
	}

	return p.save()
}

// MoveArtifact records an artifact moved to knowledge during finalize.
func (p *Project) MoveArtifact(from, to string) error {
	state := p.State()

	move := map[string]string{
		"from": from,
		"to":   to,
	}

	// Handle the any type for Artifacts_moved
	if state.Phases.Finalize.Artifacts_moved == nil {
		state.Phases.Finalize.Artifacts_moved = []map[string]string{move}
	} else {
		// Type assert to []map[string]string
		if moves, ok := state.Phases.Finalize.Artifacts_moved.([]map[string]string); ok {
			state.Phases.Finalize.Artifacts_moved = append(moves, move)
		} else if moves, ok := state.Phases.Finalize.Artifacts_moved.([]interface{}); ok {
			// Handle interface{} slice from YAML unmarshaling
			mapMoves := make([]map[string]string, 0, len(moves)+1)
			for _, m := range moves {
				if mm, ok := m.(map[string]interface{}); ok {
					strMap := make(map[string]string)
					for k, v := range mm {
						if s, ok := v.(string); ok {
							strMap[k] = s
						}
					}
					mapMoves = append(mapMoves, strMap)
				}
			}
			mapMoves = append(mapMoves, move)
			state.Phases.Finalize.Artifacts_moved = mapMoves
		} else {
			// If type is unexpected, reset to new slice
			state.Phases.Finalize.Artifacts_moved = []map[string]string{move}
		}
	}

	return p.save()
}

// createPhaseStructure creates the directory structure for a phase.
func (p *Project) createPhaseStructure(phaseName string) error {
	phaseDir := filepath.Join(".sow/project/phases", phaseName)

	// Create phase directory
	if err := p.sow.fs.MkdirAll(phaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create phase directory: %w", err)
	}

	// Create log file
	logPath := filepath.Join(phaseDir, "log.md")
	logContent := []byte(fmt.Sprintf("# %s Phase Log\n\n", capitalize(phaseName)))
	if err := p.sow.writeFile(logPath, logContent, 0644); err != nil {
		return fmt.Errorf("failed to create phase log: %w", err)
	}

	// Phase-specific structures
	switch phaseName {
	case "discovery":
		// Create research directory
		researchDir := filepath.Join(phaseDir, "research")
		if err := p.sow.fs.MkdirAll(researchDir, 0755); err != nil {
			return err
		}

	case "design":
		// Create ADRs and design docs directories
		adrsDir := filepath.Join(phaseDir, "adrs")
		if err := p.sow.fs.MkdirAll(adrsDir, 0755); err != nil {
			return err
		}

		docsDir := filepath.Join(phaseDir, "design-docs")
		if err := p.sow.fs.MkdirAll(docsDir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// createTaskStructure creates the directory structure for a task.
func (p *Project) createTaskStructure(id string) error {
	taskDir := filepath.Join(".sow/project/phases/implementation/tasks", id)

	// Create task directory
	if err := p.sow.fs.MkdirAll(taskDir, 0755); err != nil {
		return err
	}

	// Create state.yaml
	taskState := schemas.TaskState{}
	taskState.Task.Id = id
	taskState.Task.Name = ""  // Will be set later
	taskState.Task.Phase = "implementation"
	taskState.Task.Status = "pending"
	taskState.Task.Created_at = time.Now()
	taskState.Task.Updated_at = time.Now()
	taskState.Task.Iteration = 1
	taskState.Task.Assigned_agent = ""
	taskState.Task.References = []string{}
	taskState.Task.Feedback = []schemas.Feedback{}
	taskState.Task.Files_modified = []string{}

	statePath := filepath.Join(taskDir, "state.yaml")
	if err := p.sow.writeYAML(statePath, &taskState); err != nil {
		return err
	}

	// Create description.md
	descPath := filepath.Join(taskDir, "description.md")
	descContent := []byte("# Task Description\n\nDescription will be added here.\n")
	if err := p.sow.writeFile(descPath, descContent, 0644); err != nil {
		return err
	}

	// Create log.md
	logPath := filepath.Join(taskDir, "log.md")
	logContent := []byte("# Task Log\n\n")
	if err := p.sow.writeFile(logPath, logContent, 0644); err != nil {
		return err
	}

	// Create feedback directory
	feedbackDir := filepath.Join(taskDir, "feedback")
	if err := p.sow.fs.MkdirAll(feedbackDir, 0755); err != nil {
		return err
	}

	return nil
}

// generateTaskID generates a gap-numbered task ID (010, 020, 030...).
func (p *Project) generateTaskID() string {
	state := p.State()

	// Find the highest existing ID
	maxID := 0
	for _, t := range state.Phases.Implementation.Tasks {
		var id int
		fmt.Sscanf(t.Id, "%d", &id)
		if id > maxID {
			maxID = id
		}
	}

	// Next ID is maxID + 10
	nextID := maxID + 10
	return fmt.Sprintf("%03d", nextID)
}
