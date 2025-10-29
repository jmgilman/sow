package standard

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/internal/logging"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/projects"
	"gopkg.in/yaml.v3"
)

// StandardProject implements the standard 4-phase project lifecycle.
//
// Phase sequence:
//  1. Planning (required) - Gather context, confirm requirements, create task list
//  2. Implementation (required) - Code implementation with tasks
//  3. Review (required) - Code review with possible iteration
//  4. Finalize (required) - Documentation, checks, and cleanup
//
//nolint:revive // name is intentional to distinguish from custom project types
type StandardProject struct {
	state   *projects.StandardProjectState
	ctx     *sow.Context
	machine *statechart.Machine
	phases  map[string]domain.Phase
}

// New creates a new StandardProject.
func New(state *projects.StandardProjectState, ctx *sow.Context) *StandardProject {
	p := &StandardProject{
		state:  state,
		ctx:    ctx,
		phases: make(map[string]domain.Phase),
	}

	// Create phase instances (they need parent project for Save())
	p.phases["planning"] = NewPlanningPhase(&state.Phases.Planning, p, ctx)
	p.phases["implementation"] = NewImplementationPhase(&state.Phases.Implementation, p, ctx)
	p.phases["review"] = NewReviewPhase(&state.Phases.Review, p, ctx)
	p.phases["finalize"] = NewFinalizePhase(&state.Phases.Finalize, p, ctx)

	// Build state machine
	p.machine = p.buildStateMachine()

	return p
}

// Implements Project interface

// Name returns the project name.
func (p *StandardProject) Name() string {
	return p.state.Project.Name
}

// Branch returns the git branch for this project.
func (p *StandardProject) Branch() string {
	return p.state.Project.Branch
}

// Description returns the project description.
func (p *StandardProject) Description() string {
	return p.state.Project.Description
}

// Type returns the project type identifier.
func (p *StandardProject) Type() string {
	return "standard"
}

// CurrentPhase returns the currently active phase based on state machine state.
func (p *StandardProject) CurrentPhase() domain.Phase {
	currentState := p.machine.State()

	switch currentState {
	case PlanningActive:
		return p.phases["planning"]
	case ImplementationPlanning, ImplementationExecuting:
		return p.phases["implementation"]
	case ReviewActive:
		return p.phases["review"]
	case FinalizeDocumentation, FinalizeChecks, FinalizeDelete:
		return p.phases["finalize"]
	default:
		return nil
	}
}

// Phase retrieves a phase by name.
func (p *StandardProject) Phase(name string) (domain.Phase, error) {
	phase, ok := p.phases[name]
	if !ok {
		return nil, project.ErrPhaseNotFound
	}
	return phase, nil
}

// Machine returns the project's state machine.
func (p *StandardProject) Machine() *statechart.Machine {
	return p.machine
}

// InitialState returns the initial state for the project's state machine.
func (p *StandardProject) InitialState() statechart.State {
	return PlanningActive
}

// Save persists the project state to disk.
func (p *StandardProject) Save() error {
	if err := p.machine.Save(); err != nil {
		return fmt.Errorf("failed to save state machine: %w", err)
	}
	return nil
}

// Log records an action in the project log.
func (p *StandardProject) Log(action, result string, opts ...domain.LogOption) error {
	entry := &logging.LogEntry{
		Timestamp: time.Now(),
		AgentID:   "orchestrator",
		Action:    action,
		Result:    result,
	}

	// Apply options
	for _, opt := range opts {
		opt(entry)
	}

	// Use logging package to append to project log
	if err := logging.AppendLog(p.ctx.FS(), "project/log.md", entry); err != nil {
		return fmt.Errorf("failed to append log: %w", err)
	}
	return nil
}

// buildStateMachine constructs the state machine for this project using the builder pattern.
func (p *StandardProject) buildStateMachine() *statechart.Machine {
	// Get current state from the project
	currentState := statechart.State(p.state.Statechart.Current_state)

	// Convert to ProjectState for machine (schemas.ProjectState is an alias for projects.StandardProjectState)
	projectState := (*schemas.ProjectState)(p.state)

	// Create prompt generator with full context access
	promptGen := NewStandardPromptGenerator(p.ctx)

	// Create builder
	builder := statechart.NewBuilder(currentState, projectState, promptGen)

	// Configure all state transitions with the builder
	builder.
		// NoProject → PlanningActive (unconditional)
		AddTransition(
			statechart.NoProject,
			PlanningActive,
			EventProjectInit,
		).
		// PlanningActive → ImplementationPlanning (requires task list approved)
		AddTransition(
			PlanningActive,
			ImplementationPlanning,
			EventCompletePlanning,
			statechart.WithGuard(func() bool {
				return PlanningComplete(p.state.Phases.Planning)
			}),
		).
		// Allow project deletion from planning
		AddTransition(
			PlanningActive,
			statechart.NoProject,
			EventProjectDelete,
		).
		// ImplementationPlanning → ImplementationExecuting (task created)
		AddTransition(
			ImplementationPlanning,
			ImplementationExecuting,
			EventTaskCreated,
			statechart.WithGuard(func() bool {
				return HasAtLeastOneTask(projectState)
			}),
		).
		// ImplementationPlanning → ImplementationExecuting (tasks approved)
		AddTransition(
			ImplementationPlanning,
			ImplementationExecuting,
			EventTasksApproved,
			statechart.WithGuard(func() bool {
				return TasksApproved(projectState)
			}),
		).
		// Allow project deletion from implementation planning
		AddTransition(
			ImplementationPlanning,
			statechart.NoProject,
			EventProjectDelete,
		).
		// ImplementationExecuting → ReviewActive (all tasks done)
		AddTransition(
			ImplementationExecuting,
			ReviewActive,
			EventAllTasksComplete,
			statechart.WithGuard(func() bool {
				return AllTasksComplete(projectState)
			}),
		).
		// Allow project deletion from implementation executing
		AddTransition(
			ImplementationExecuting,
			statechart.NoProject,
			EventProjectDelete,
		).
		// ReviewActive → ImplementationPlanning (review failed - loop back)
		AddTransition(
			ReviewActive,
			ImplementationPlanning,
			EventReviewFail,
			statechart.WithGuard(func() bool {
				return LatestReviewApproved(projectState)
			}),
		).
		// ReviewActive → FinalizeDocumentation (review passed)
		AddTransition(
			ReviewActive,
			FinalizeDocumentation,
			EventReviewPass,
			statechart.WithGuard(func() bool {
				return LatestReviewApproved(projectState)
			}),
		).
		// Allow project deletion from review
		AddTransition(
			ReviewActive,
			statechart.NoProject,
			EventProjectDelete,
		).
		// FinalizeDocumentation → FinalizeChecks (documentation handled)
		AddTransition(
			FinalizeDocumentation,
			FinalizeChecks,
			EventDocumentationDone,
			statechart.WithGuard(func() bool {
				return DocumentationAssessed(projectState)
			}),
		).
		// Allow project deletion from finalize documentation
		AddTransition(
			FinalizeDocumentation,
			statechart.NoProject,
			EventProjectDelete,
		).
		// FinalizeChecks → FinalizeDelete (checks handled)
		AddTransition(
			FinalizeChecks,
			FinalizeDelete,
			EventChecksDone,
			statechart.WithGuard(func() bool {
				return ChecksAssessed(projectState)
			}),
		).
		// Allow project deletion from finalize checks
		AddTransition(
			FinalizeChecks,
			statechart.NoProject,
			EventProjectDelete,
		).
		// FinalizeDelete → NoProject (project deleted)
		AddTransition(
			FinalizeDelete,
			statechart.NoProject,
			EventProjectDelete,
			statechart.WithGuard(func() bool {
				return ProjectDeleted(projectState)
			}),
		)

	// Build and configure the machine
	machine := builder.Build()
	machine.SetFilesystem(p.ctx.FS())

	return machine
}

// Task management methods

// InferTaskID attempts to infer the task ID from context.
// Returns the task ID from the active in_progress task.
func (p *StandardProject) InferTaskID() (string, error) {
	var inProgressID string
	count := 0

	for _, t := range p.state.Phases.Implementation.Tasks {
		if t.Status == "in_progress" {
			inProgressID = t.Id
			count++
		}
	}

	if count == 0 {
		return "", project.ErrNoTask
	}

	if count > 1 {
		return "", fmt.Errorf("multiple in_progress tasks found, explicit ID required")
	}

	return inProgressID, nil
}

// GetTask retrieves a task by ID via the implementation phase.
func (p *StandardProject) GetTask(id string) (*domain.Task, error) {
	// Delegate to implementation phase
	implPhase, ok := p.phases["implementation"].(*ImplementationPhase)
	if !ok {
		return nil, fmt.Errorf("implementation phase not available")
	}

	return implPhase.GetTask(id)
}

// Filesystem helper methods (for Task and internal operations)

// ReadYAML reads and unmarshals a YAML file.
func (p *StandardProject) ReadYAML(path string, v interface{}) error {
	data, err := p.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return nil
}

// WriteYAML marshals a value to YAML and writes it atomically.
func (p *StandardProject) WriteYAML(path string, v interface{}) error {
	// Encode to a node first, then customize it to remove null values.
	var node yaml.Node
	if err := node.Encode(v); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	// Remove null values from the node tree.
	removeNullNodes(&node)

	// Marshal the cleaned node.
	data, err := yaml.Marshal(&node)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	fs := p.ctx.FS()

	// Write to temp file first
	tmpPath := path + ".tmp"
	if err := p.WriteFile(tmpPath, data); err != nil {
		return err
	}

	// Atomic rename
	if err := fs.Rename(tmpPath, path); err != nil {
		_ = fs.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// ReadFile reads a file's contents using the context's filesystem.
func (p *StandardProject) ReadFile(path string) ([]byte, error) {
	fs := p.ctx.FS()
	f, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer func() { _ = f.Close() }()

	var data []byte
	buf := make([]byte, 4096)
	for {
		n, readErr := f.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if readErr != nil {
			if readErr.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to read file: %w", readErr)
		}
	}

	return data, nil
}

// WriteFile writes a file using the context's filesystem.
func (p *StandardProject) WriteFile(path string, data []byte) error {
	fs := p.ctx.FS()
	if err := fs.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

// CreatePullRequest creates a pull request for the project using GitHub CLI.
// The provided body should contain the main PR description written by the orchestrator.
// This function adds issue references and footer automatically.
// The PR URL is stored in the project state.
func (p *StandardProject) CreatePullRequest(body string) (string, error) {
	state := p.machine.ProjectState()

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
	if state.Project.Github_issue != nil {
		issueNum := *state.Project.Github_issue
		if issueNum > 0 {
			fullBody += fmt.Sprintf("\n\nCloses #%d\n", issueNum)
		}
	}

	// Add footer
	fullBody += "\n---\n\n🤖 Generated with sow\n"

	// Create PR via GitHub CLI using context
	prURL, err := p.ctx.GitHub().CreatePullRequest(title, fullBody)
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}

	// Store PR URL in state metadata
	if state.Phases.Finalize.Metadata == nil {
		state.Phases.Finalize.Metadata = make(map[string]interface{})
	}
	state.Phases.Finalize.Metadata["pr_url"] = prURL

	// Save state
	if err := p.Save(); err != nil {
		return "", fmt.Errorf("failed to save PR URL: %w", err)
	}

	return prURL, nil
}

// removeNullNodes recursively removes null value nodes from a YAML node tree.
func removeNullNodes(node *yaml.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, child := range node.Content {
			removeNullNodes(child)
		}
	case yaml.MappingNode:
		filtered := make([]*yaml.Node, 0, len(node.Content))
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			if value.Kind == yaml.ScalarNode && value.Tag == "!!null" {
				continue
			}

			removeNullNodes(value)
			filtered = append(filtered, key, value)
		}
		node.Content = filtered
	}
}
