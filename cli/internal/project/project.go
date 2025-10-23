// Package project provides the Project type and its operations.
//
// Project is the aggregate root for managing sow projects with full state machine integration.
// It uses sow.Context for filesystem, git, and GitHub access.
//
// Load/Create functions are provided at package level to construct Project instances:
//   - Load(ctx) - Loads existing project from disk
//   - Create(ctx, name, description) - Creates new project
//   - CreateFromIssue(ctx, issueNumber, branchName) - Creates project from GitHub issue
//   - Delete(ctx) - Deletes the active project
//   - Exists(ctx) - Checks if a project exists
package project

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
	"gopkg.in/yaml.v3"
)

// Project represents an active sow project with full state machine integration.
// All operations automatically persist state changes to disk.
//
// Project is an aggregate root that owns its statechart and tasks.
// It uses sow.Context for filesystem, git, and GitHub access.
type Project struct {
	ctx     *sow.Context
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
	if err := p.machine.Save(); err != nil {
		return fmt.Errorf("failed to save project state: %w", err)
	}
	return nil
}

// EnablePhase enables a phase and transitions the state machine.
func (p *Project) EnablePhase(phaseName string, opts ...sow.PhaseOption) error {
	// Apply options
	cfg := &sow.PhaseConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	state := p.State()
	now := time.Now()

	// Handle each phase
	switch phaseName {
	case "discovery":
		if cfg.DiscoveryType() == "" {
			return fmt.Errorf("discovery type required (use WithDiscoveryType option)")
		}
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "in_progress"
		discoveryType := cfg.DiscoveryType()
		state.Phases.Discovery.Discovery_type = &discoveryType
		startedAt := now
		state.Phases.Discovery.Started_at = &startedAt

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
		state.Phases.Design.Status = "in_progress"
		designStartedAt := now
		state.Phases.Design.Started_at = &designStartedAt

		// Create design directory structure
		if err := p.createPhaseStructure("design"); err != nil {
			return err
		}

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventEnableDesign); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	default:
		return sow.ErrInvalidPhase
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
			return sow.ErrPhaseNotEnabled
		}
		state.Phases.Discovery.Status = "completed"
		discoveryCompletedAt := now
		state.Phases.Discovery.Completed_at = &discoveryCompletedAt

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventCompleteDiscovery); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "design":
		if !state.Phases.Design.Enabled {
			return sow.ErrPhaseNotEnabled
		}
		state.Phases.Design.Status = "completed"
		designCompletedAt := now
		state.Phases.Design.Completed_at = &designCompletedAt

		// Fire state machine event
		if err := p.machine.Fire(statechart.EventCompleteDesign); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "implementation":
		state.Phases.Implementation.Status = "completed"
		implCompletedAt := now
		state.Phases.Implementation.Completed_at = &implCompletedAt

		// Fire state machine event (transitions to review)
		if err := p.machine.Fire(statechart.EventAllTasksComplete); err != nil {
			return fmt.Errorf("state transition failed: %w", err)
		}

	case "review":
		state.Phases.Review.Status = "completed"
		reviewCompletedAt := now
		state.Phases.Review.Completed_at = &reviewCompletedAt

		// Fire state machine event (handled by review pass/fail)
		// This is a placeholder - actual review completion happens via AddReviewReport

	case "finalize":
		state.Phases.Finalize.Status = "completed"
		finalizeCompletedAt := now
		state.Phases.Finalize.Completed_at = &finalizeCompletedAt

		// Finalize has substates, handled by specialized methods

	default:
		return sow.ErrInvalidPhase
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
func (p *Project) AddTask(name string, opts ...sow.TaskOption) (*Task, error) {
	state := p.State()

	// Apply options
	cfg := &sow.TaskConfig{
		Status:   "pending",
		Parallel: false,
		Agent:    "implementer", // Default agent
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Determine task ID (explicit or auto-generated)
	id := cfg.ID()
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
	for _, depID := range cfg.Dependencies() {
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
	task := phases.Task{
		Id:           id,
		Name:         name,
		Status:       cfg.Status,
		Parallel:     cfg.Parallel,
		Dependencies: cfg.Dependencies(),
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

	// Set approval flag and update phase status
	state.Phases.Implementation.Tasks_approved = true
	state.Phases.Implementation.Status = "in_progress"

	// Set started_at if not already set
	if state.Phases.Implementation.Started_at == nil {
		now := time.Now()
		state.Phases.Implementation.Started_at = &now
	}

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

	return nil, sow.ErrNoTask
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

	artifact := phases.Artifact{
		Path:       path,
		Approved:   approved,
		Created_at: now,
	}

	switch phaseName {
	case "discovery":
		if !state.Phases.Discovery.Enabled {
			return sow.ErrPhaseNotEnabled
		}
		state.Phases.Discovery.Artifacts = append(state.Phases.Discovery.Artifacts, artifact)

	case "design":
		if !state.Phases.Design.Enabled {
			return sow.ErrPhaseNotEnabled
		}
		state.Phases.Design.Artifacts = append(state.Phases.Design.Artifacts, artifact)

	default:
		return sow.ErrInvalidPhase
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
			return sow.ErrPhaseNotEnabled
		}
		for i := range state.Phases.Discovery.Artifacts {
			if state.Phases.Discovery.Artifacts[i].Path == path {
				state.Phases.Discovery.Artifacts[i].Approved = true
				return p.save()
			}
		}

	case "design":
		if !state.Phases.Design.Enabled {
			return sow.ErrPhaseNotEnabled
		}
		for i := range state.Phases.Design.Artifacts {
			if state.Phases.Design.Artifacts[i].Path == path {
				state.Phases.Design.Artifacts[i].Approved = true
				return p.save()
			}
		}

	default:
		return sow.ErrInvalidPhase
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
	report := phases.ReviewReport{
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
	var report *phases.ReviewReport
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
	state.Phases.Finalize.Documentation_updates = append(state.Phases.Finalize.Documentation_updates, path)
	return p.save()
}

// MoveArtifact records an artifact moved to knowledge during finalize.
func (p *Project) MoveArtifact(from, to string) error {
	state := p.State()
	move := struct {
		From string `json:"from" yaml:"from"`
		To   string `json:"to" yaml:"to"`
	}{From: from, To: to}
	state.Phases.Finalize.Artifacts_moved = append(state.Phases.Finalize.Artifacts_moved, move)
	return p.save()
}

// CreatePullRequest creates a pull request for the project using GitHub CLI.
// The provided body should contain the main PR description written by the orchestrator.
// This function adds issue references and footer automatically.
// The PR URL is stored in the project state.
func (p *Project) CreatePullRequest(body string) (string, error) {
	state := p.State()

	// Generate PR title from project
	// Capitalize first letter of project name
	name := p.Name()
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	title := fmt.Sprintf("%s: %s", name, p.Description())

	// Wrap body with issue reference and footer
	fullBody := body

	// Add issue reference if linked (before footer)
	if state.Project.Github_issue != nil && *state.Project.Github_issue > 0 {
		fullBody += fmt.Sprintf("\n\nCloses #%d\n", *state.Project.Github_issue)
	}

	// Add footer
	fullBody += "\n---\n\n🤖 Generated with sow\n"

	// Create PR via GitHub CLI using context
	prURL, err := p.ctx.GitHub().CreatePullRequest(title, fullBody)
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}

	// Store PR URL in state
	state.Phases.Finalize.Pr_url = &prURL

	// Save state
	if err := p.save(); err != nil {
		return "", fmt.Errorf("failed to save PR URL: %w", err)
	}

	return prURL, nil
}

// Log creates and appends a structured log entry to the project log.
// Automatically uses "orchestrator" as agent ID.
//
// Example:
//
//	p.Log("created_file", "success",
//	      WithFiles("design.md"),
//	      WithNotes("Created initial design document"))
func (p *Project) Log(action, result string, opts ...LogOption) error {
	entry := &LogEntry{
		Timestamp: time.Now(),
		AgentID:   "orchestrator",
		Action:    action,
		Result:    result,
	}

	// Apply options
	for _, opt := range opts {
		opt(entry)
	}

	// Validate
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("invalid log entry: %w", err)
	}

	// Format and append
	formatted := entry.Format()
	return p.appendLog(formatted)
}

// appendLog appends a raw log entry to the project log file.
func (p *Project) appendLog(entry string) error {
	logPath := "project/log.md"

	// Read existing content
	existing, err := p.readFile(logPath)
	if err != nil {
		return fmt.Errorf("failed to read project log: %w", err)
	}

	// Append new entry
	updated := append(existing, []byte(entry)...)

	// Write back
	if err := p.writeFile(logPath, updated); err != nil {
		return fmt.Errorf("failed to write project log: %w", err)
	}

	return nil
}

// createPhaseStructure creates the directory structure for a phase.
func (p *Project) createPhaseStructure(phaseName string) error {
	phaseDir := filepath.Join("project/phases", phaseName)
	fs := p.ctx.FS()

	// Create phase directory
	if err := fs.MkdirAll(phaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create phase directory: %w", err)
	}

	// Create log file
	logPath := filepath.Join(phaseDir, "log.md")
	// Capitalize first letter of phase name
	phaseTitle := strings.ToUpper(phaseName[:1]) + phaseName[1:]
	logContent := []byte(fmt.Sprintf("# %s Phase Log\n\n", phaseTitle))
	if err := p.writeFile(logPath, logContent); err != nil {
		return fmt.Errorf("failed to create phase log: %w", err)
	}

	// Phase-specific structures
	switch phaseName {
	case "discovery":
		// Create research directory
		researchDir := filepath.Join(phaseDir, "research")
		if err := fs.MkdirAll(researchDir, 0755); err != nil {
			return fmt.Errorf("failed to create research directory: %w", err)
		}

	case "design":
		// Create ADRs and design docs directories
		adrsDir := filepath.Join(phaseDir, "adrs")
		if err := fs.MkdirAll(adrsDir, 0755); err != nil {
			return fmt.Errorf("failed to create adrs directory: %w", err)
		}

		docsDir := filepath.Join(phaseDir, "design-docs")
		if err := fs.MkdirAll(docsDir, 0755); err != nil {
			return fmt.Errorf("failed to create design-docs directory: %w", err)
		}
	}

	return nil
}

// createTaskStructure creates the directory structure for a task.
func (p *Project) createTaskStructure(id, name string, cfg *sow.TaskConfig) error {
	taskDir := filepath.Join("project/phases/implementation/tasks", id)
	fs := p.ctx.FS()

	// Create task directory
	if err := fs.MkdirAll(taskDir, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Create state.yaml with actual values
	taskState := schemas.TaskState{}
	taskState.Task.Id = id
	taskState.Task.Name = name
	taskState.Task.Phase = "implementation"
	taskState.Task.Status = cfg.Status
	taskState.Task.Created_at = time.Now()
	taskState.Task.Updated_at = time.Now()
	taskState.Task.Iteration = 1
	taskState.Task.Assigned_agent = cfg.Agent
	taskState.Task.References = []string{}
	taskState.Task.Feedback = []schemas.Feedback{}
	taskState.Task.Files_modified = []string{}

	statePath := filepath.Join(taskDir, "state.yaml")
	if err := p.writeYAML(statePath, &taskState); err != nil {
		return err
	}

	// Create description.md with actual description
	descPath := filepath.Join(taskDir, "description.md")
	descContent := fmt.Sprintf("# Task %s: %s\n\n%s\n", id, name, cfg.Description())
	if err := p.writeFile(descPath, []byte(descContent)); err != nil {
		return err
	}

	// Create log.md
	logPath := filepath.Join(taskDir, "log.md")
	logContent := fmt.Sprintf("# Task %s Log\n\nWorker actions will be logged here.\n", id)
	if err := p.writeFile(logPath, []byte(logContent)); err != nil {
		return err
	}

	// Create feedback directory
	feedbackDir := filepath.Join(taskDir, "feedback")
	if err := fs.MkdirAll(feedbackDir, 0755); err != nil {
		return fmt.Errorf("failed to create feedback directory: %w", err)
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

// Helper methods for filesystem operations

// readYAML reads and unmarshals a YAML file.
func (p *Project) readYAML(path string, v interface{}) error {
	data, err := p.readFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return nil
}

// writeFile writes a file using the context's filesystem.
func (p *Project) writeFile(path string, data []byte) error {
	fs := p.ctx.FS()
	f, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// readFile reads a file's contents using the context's filesystem.
func (p *Project) readFile(path string) ([]byte, error) {
	fs := p.ctx.FS()
	f, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer func() { _ = f.Close() }()

	var data []byte
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
	}

	return data, nil
}

// writeYAML marshals a value to YAML and writes it atomically.
func (p *Project) writeYAML(path string, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	fs := p.ctx.FS()

	// Write to temp file first
	tmpPath := path + ".tmp"
	if err := p.writeFile(tmpPath, data); err != nil {
		return err
	}

	// Atomic rename
	if err := fs.Rename(tmpPath, path); err != nil {
		_ = fs.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
