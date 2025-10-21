package sow

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/internal/github"
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
		state.Phases.Discovery.Discovery_type = cfg.discoveryType
		state.Phases.Discovery.Started_at = now.Format(time.RFC3339)

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
		state.Phases.Design.Started_at = now.Format(time.RFC3339)

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
		state.Phases.Discovery.Completed_at = now.Format(time.RFC3339)

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventCompleteDiscovery); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "design":
		if !state.Phases.Design.Enabled {
			return ErrPhaseNotEnabled
		}
		state.Phases.Design.Status = "completed"
		state.Phases.Design.Completed_at = now.Format(time.RFC3339)

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventCompleteDesign); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "implementation":
		state.Phases.Implementation.Status = "completed"
		state.Phases.Implementation.Completed_at = now.Format(time.RFC3339)

		// Fire state machine event (transitions to review)
		if err := p.machine.Fire(statechart.EventAllTasksComplete); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "review":
		state.Phases.Review.Status = "completed"
		state.Phases.Review.Completed_at = now.Format(time.RFC3339)

		// Fire state machine event (handled by review pass/fail)
		// This is a placeholder - actual review completion happens via AddReviewReport

	case "finalize":
		state.Phases.Finalize.Status = "completed"
		state.Phases.Finalize.Completed_at = now.Format(time.RFC3339)

		// Finalize has substates, handled by specialized methods

	default:
		return ErrInvalidPhase
	}

	// Auto-save
	return p.save()
}

// SkipPhase marks an optional phase as skipped and transitions the state machine.
func (p *Project) SkipPhase(phaseName string) error {
	state := p.State()

	switch phaseName {
	case "discovery":
		state.Phases.Discovery.Enabled = false
		state.Phases.Discovery.Status = "skipped"

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventSkipDiscovery); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "design":
		state.Phases.Design.Enabled = false
		state.Phases.Design.Status = "skipped"

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventSkipDesign); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	default:
		return fmt.Errorf("only discovery and design phases can be skipped")
	}

	// Auto-save
	return p.save()
}

// CompleteFinalizeSubphase completes a finalize subphase (documentation or checks).
func (p *Project) CompleteFinalizeSubphase(subphase string) error {
	var event statechart.Event

	switch subphase {
	case "documentation":
		event = statechart.EventDocumentationDone
	case "checks":
		event = statechart.EventChecksDone
	default:
		return fmt.Errorf("invalid finalize subphase: must be 'documentation' or 'checks'")
	}

	// Fire state machine event
	if err := p.machine.Fire(event); err != nil {
		return fmt.Errorf("state transition failed: %w", err)
	}

	// Auto-save
	return p.save()
}

// AddTask creates a new implementation task.
func (p *Project) AddTask(name string, opts ...TaskOption) (*Task, error) {
	state := p.State()

	// Apply options
	cfg := &taskConfig{
		status:   "pending",
		parallel: false,
		agent:    "implementer", // Default agent
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Determine task ID (explicit or auto-generated)
	id := cfg.id
	if id == "" {
		id = p.generateTaskID()
	} else {
		// Validate explicit ID format
		if !isValidTaskID(id) {
			return nil, fmt.Errorf("invalid task ID: must be 3 digits (e.g., 010, 020, 015)")
		}
		// Check for duplicates
		for _, t := range state.Phases.Implementation.Tasks {
			if t.Id == id {
				return nil, fmt.Errorf("task ID already exists: %s", id)
			}
		}
	}

	// Validate dependencies exist
	for _, depID := range cfg.dependencies {
		found := false
		for _, t := range state.Phases.Implementation.Tasks {
			if t.Id == depID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("dependency task not found: %s", depID)
		}
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

	// Create task directory structure and files
	if err := p.createTaskStructure(id, name, cfg); err != nil {
		return nil, fmt.Errorf("failed to create task structure: %w", err)
	}

	// NOTE: State transition removed - tasks now require human approval
	// before transitioning to ImplementationExecuting. Use ApproveTasks() method.

	// Auto-save
	if err := p.save(); err != nil {
		return nil, err
	}

	return &Task{
		project: p,
		id:      id,
	}, nil
}

// ApproveTasks approves the task plan and transitions to execution mode.
// This must be called after all tasks are created to begin autonomous execution.
func (p *Project) ApproveTasks() error {
	state := p.State()

	// Validate at least one task exists
	if len(state.Phases.Implementation.Tasks) == 0 {
		return fmt.Errorf("cannot approve: no tasks have been created")
	}

	// Set approval flag
	state.Phases.Implementation.Tasks_approved = true

	// Fire transition event
	if err := p.machine.Fire(statechart.EventTasksApproved); err != nil {
		return fmt.Errorf("state transition failed: %w", err)
	}

	return p.save()
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

	// Create report (not approved by default - requires human approval)
	report := schemas.ReviewReport{
		Id:         reportID,
		Path:       path,
		Created_at: now,
		Assessment: assessment,
		Approved:   false,
	}

	state.Phases.Review.Reports = append(state.Phases.Review.Reports, report)

	// NOTE: State transition removed - reviews now require human approval
	// before transitioning. Use ApproveReview() method.

	// Auto-save
	return p.save()
}

// ApproveReview approves a review report and triggers the appropriate state transition.
// The transition depends on the report's assessment (pass or fail).
func (p *Project) ApproveReview(reportID string) error {
	state := p.State()

	// Find the report
	var report *schemas.ReviewReport
	for i := range state.Phases.Review.Reports {
		if state.Phases.Review.Reports[i].Id == reportID {
			report = &state.Phases.Review.Reports[i]
			break
		}
	}

	if report == nil {
		return fmt.Errorf("review report not found: %s", reportID)
	}

	// Mark as approved
	report.Approved = true

	// Fire appropriate transition based on assessment
	var event statechart.Event
	if report.Assessment == "pass" {
		event = statechart.EventReviewPass
	} else {
		event = statechart.EventReviewFail
	}

	if err := p.machine.Fire(event); err != nil {
		return fmt.Errorf("state transition failed: %w", err)
	}

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

// CreatePullRequest creates a pull request for the project using GitHub CLI.
// The provided body should contain the main PR description written by the orchestrator.
// This function adds issue references and footer automatically.
// The PR URL is stored in the project state.
func (p *Project) CreatePullRequest(body string) (string, error) {
	state := p.State()

	// Generate PR title from project
	title := fmt.Sprintf("%s: %s", strings.Title(p.Name()), p.Description())

	// Wrap body with issue reference and footer
	fullBody := body

	// Add issue reference if linked (before footer)
	if state.Project.Github_issue != nil {
		var issueNum int
		if num, ok := state.Project.Github_issue.(int); ok {
			issueNum = num
		} else if num, ok := state.Project.Github_issue.(float64); ok {
			issueNum = int(num)
		}

		if issueNum > 0 {
			fullBody += fmt.Sprintf("\n\nCloses #%d\n", issueNum)
		}
	}

	// Add footer
	fullBody += "\n---\n\n🤖 Generated with sow\n"

	// Create PR via GitHub CLI
	prURL, err := github.CreatePullRequest(title, fullBody)
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}

	// Store PR URL in state
	state.Phases.Finalize.Pr_url = prURL

	// Save state
	if err := p.save(); err != nil {
		return "", fmt.Errorf("failed to save PR URL: %w", err)
	}

	return prURL, nil
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
	// Capitalize first letter of phase name
	phaseTitle := strings.ToUpper(phaseName[:1]) + phaseName[1:]
	logContent := []byte(fmt.Sprintf("# %s Phase Log\n\n", phaseTitle))
	if err := p.sow.writeFile(logPath, logContent); err != nil {
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
func (p *Project) createTaskStructure(id, name string, cfg *taskConfig) error {
	taskDir := filepath.Join(".sow/project/phases/implementation/tasks", id)

	// Create task directory
	if err := p.sow.fs.MkdirAll(taskDir, 0755); err != nil {
		return err
	}

	// Create state.yaml with actual values
	taskState := schemas.TaskState{}
	taskState.Task.Id = id
	taskState.Task.Name = name
	taskState.Task.Phase = "implementation"
	taskState.Task.Status = cfg.status
	taskState.Task.Created_at = time.Now()
	taskState.Task.Updated_at = time.Now()
	taskState.Task.Iteration = 1
	taskState.Task.Assigned_agent = cfg.agent
	taskState.Task.References = []string{}
	taskState.Task.Feedback = []schemas.Feedback{}
	taskState.Task.Files_modified = []string{}

	statePath := filepath.Join(taskDir, "state.yaml")
	if err := p.sow.writeYAML(statePath, &taskState); err != nil {
		return err
	}

	// Create description.md with actual description
	descPath := filepath.Join(taskDir, "description.md")
	descContent := fmt.Sprintf("# Task %s: %s\n\n%s\n", id, name, cfg.description)
	if err := p.sow.writeFile(descPath, []byte(descContent)); err != nil {
		return err
	}

	// Create log.md
	logPath := filepath.Join(taskDir, "log.md")
	logContent := fmt.Sprintf("# Task %s Log\n\nWorker actions will be logged here.\n", id)
	if err := p.sow.writeFile(logPath, []byte(logContent)); err != nil {
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
		_, _ = fmt.Sscanf(t.Id, "%d", &id)
		if id > maxID {
			maxID = id
		}
	}

	// Next ID is maxID + 10
	nextID := maxID + 10
	return fmt.Sprintf("%03d", nextID)
}

// isValidTaskID validates that a task ID is a valid 3-digit number.
// Gap-numbered IDs (010, 020, 030) are auto-generated, but users can
// specify intermediate IDs (015, 025) for insertion between tasks.
func isValidTaskID(id string) bool {
	// Must be exactly 3 digits
	if len(id) != 3 {
		return false
	}

	// Must be numeric and greater than 0
	var num int
	if _, err := fmt.Sscanf(id, "%d", &num); err != nil {
		return false
	}

	return num > 0
}

// AppendLog appends a log entry to the project log file.
func (p *Project) AppendLog(entry string) error {
	logPath := ".sow/project/log.md"

	// Read existing content
	existing, err := p.sow.readFile(logPath)
	if err != nil {
		return fmt.Errorf("failed to read project log: %w", err)
	}

	// Append new entry
	updated := append(existing, []byte(entry)...)

	// Write back
	if err := p.sow.writeFile(logPath, updated); err != nil {
		return fmt.Errorf("failed to write project log: %w", err)
	}

	return nil
}
