package project

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
	"gopkg.in/yaml.v3"
)

// TaskCollection provides task operations on a generic Phase.
// This helper eliminates duplication across phases that support tasks.
type TaskCollection struct {
	state   *phases.Phase
	project domain.Project
	ctx     *sow.Context
}

// NewTaskCollection creates a new task collection.
func NewTaskCollection(state *phases.Phase, proj domain.Project, ctx *sow.Context) *TaskCollection {
	return &TaskCollection{
		state:   state,
		project: proj,
		ctx:     ctx,
	}
}

// Add adds a new task to the phase.
func (tc *TaskCollection) Add(name string, opts ...domain.TaskOption) (*domain.Task, error) {
	cfg := &domain.TaskConfig{
		Status: "pending",
		Agent:  "implementer",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Generate task ID
	id := tc.generateTaskID()

	// Validate dependencies
	for _, depID := range cfg.Dependencies {
		if !tc.taskExists(depID) {
			return nil, fmt.Errorf("dependency task not found: %s", depID)
		}
	}

	// Create task in schema
	task := phases.Task{
		Id:           id,
		Name:         name,
		Status:       cfg.Status,
		Parallel:     cfg.Parallel,
		Dependencies: cfg.Dependencies,
	}

	tc.state.Tasks = append(tc.state.Tasks, task)

	// Create task directory structure
	if err := tc.createTaskStructure(id, name, cfg); err != nil {
		return nil, err
	}

	if err := tc.project.Save(); err != nil {
		return nil, err
	}

	return &domain.Task{Project: tc.project, ID: id}, nil
}

// Get retrieves a task by ID.
func (tc *TaskCollection) Get(id string) (*domain.Task, error) {
	if !tc.taskExists(id) {
		return nil, ErrNoTask
	}
	return &domain.Task{Project: tc.project, ID: id}, nil
}

// List returns all tasks.
func (tc *TaskCollection) List() []*domain.Task {
	tasks := make([]*domain.Task, 0, len(tc.state.Tasks))
	for _, t := range tc.state.Tasks {
		tasks = append(tasks, &domain.Task{Project: tc.project, ID: t.Id})
	}
	return tasks
}

// Approve marks tasks as approved for execution.
func (tc *TaskCollection) Approve() error {
	if len(tc.state.Tasks) == 0 {
		return fmt.Errorf("cannot approve: no tasks exist")
	}

	// Set approval in metadata
	if tc.state.Metadata == nil {
		tc.state.Metadata = make(map[string]interface{})
	}
	tc.state.Metadata["tasks_approved"] = true

	tc.state.Status = "in_progress"
	if tc.state.Started_at == nil {
		now := time.Now()
		tc.state.Started_at = &now
	}

	return tc.project.Save()
}

// Helper methods

func (tc *TaskCollection) generateTaskID() string {
	maxID := 0
	for _, t := range tc.state.Tasks {
		var id int
		fmt.Sscanf(t.Id, "%d", &id)
		if id > maxID {
			maxID = id
		}
	}
	return fmt.Sprintf("%03d", maxID+10)
}

func (tc *TaskCollection) taskExists(id string) bool {
	for _, t := range tc.state.Tasks {
		if t.Id == id {
			return true
		}
	}
	return false
}

func (tc *TaskCollection) createTaskStructure(id, name string, cfg *domain.TaskConfig) error {
	taskDir := filepath.Join("project/phases/implementation/tasks", id)
	fs := tc.ctx.FS()

	// Create task directory
	if err := fs.MkdirAll(taskDir, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Create state.yaml
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

	// Marshal to YAML node first
	var node yaml.Node
	if err := node.Encode(&taskState); err != nil {
		return fmt.Errorf("failed to encode task state: %w", err)
	}

	// Remove null values from the node tree
	removeNullNodes(&node)

	// Marshal the cleaned node
	stateData, err := yaml.Marshal(&node)
	if err != nil {
		return fmt.Errorf("failed to marshal task state: %w", err)
	}
	if err := fs.WriteFile(statePath, stateData, 0644); err != nil {
		return err
	}

	// Create description.md
	descPath := filepath.Join(taskDir, "description.md")
	descContent := fmt.Sprintf("# Task %s: %s\n\n%s\n", id, name, cfg.Description)
	if err := fs.WriteFile(descPath, []byte(descContent), 0644); err != nil {
		return err
	}

	// Create log.md
	logPath := filepath.Join(taskDir, "log.md")
	logContent := fmt.Sprintf("# Task %s Log\n\nWorker actions will be logged here.\n", id)
	if err := fs.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		return err
	}

	// Create feedback directory
	feedbackDir := filepath.Join(taskDir, "feedback")
	if err := fs.MkdirAll(feedbackDir, 0755); err != nil {
		return fmt.Errorf("failed to create feedback directory: %w", err)
	}

	return nil
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

			// Skip null values
			if value.Kind == yaml.ScalarNode && value.Tag == "!!null" {
				continue
			}

			// Recurse into non-null values
			removeNullNodes(value)

			// Keep the key-value pair
			filtered = append(filtered, key, value)
		}
		node.Content = filtered
	}
}
