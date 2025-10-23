package domain

import (
	"time"

	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// Project is the aggregate root for all project types.
// The CLI works exclusively through this interface.
type Project interface {
	// Identity
	Name() string
	Branch() string
	Description() string
	Type() string

	// Phase access
	CurrentPhase() Phase
	Phase(name string) (Phase, error)

	// State machine
	Machine() *statechart.Machine
	InitialState() statechart.State // Returns the state this project type starts in

	// Persistence
	Save() error

	// Logging
	Log(action, result string, opts ...LogOption) error

	// Task management
	InferTaskID() (string, error)
	GetTask(id string) (*Task, error)

	// Pull request
	CreatePullRequest(body string) (string, error)

	// Filesystem helpers (for Task and internal operations)
	ReadYAML(path string, v interface{}) error
	WriteYAML(path string, v interface{}) error
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
}

// Phase represents any phase in any project type.
// Operations not supported by a phase return ErrNotSupported.
type Phase interface {
	// Metadata
	Name() string
	Status() string
	Enabled() bool

	// Artifact operations (discovery, design, review)
	// Returns schema types directly - no wrapper needed
	AddArtifact(path string, opts ...ArtifactOption) error
	ApproveArtifact(path string) error
	ListArtifacts() []*phases.Artifact

	// Task operations (implementation only)
	// Returns concrete Task type - single implementation for all project types
	AddTask(name string, opts ...TaskOption) (*Task, error)
	GetTask(id string) (*Task, error)
	ListTasks() []*Task
	ApproveTasks() error

	// Generic field access (for metadata)
	Set(field string, value interface{}) error
	Get(field string) (interface{}, error)

	// Lifecycle
	Complete() error
	Skip() error
	Enable(opts ...PhaseOption) error
}

// ArtifactOption configures artifact creation.
type ArtifactOption func(*ArtifactConfig)

// ArtifactConfig holds configuration for creating artifacts.
type ArtifactConfig struct {
	Metadata map[string]interface{}
}

// WithMetadata adds metadata to an artifact.
func WithMetadata(metadata map[string]interface{}) ArtifactOption {
	return func(c *ArtifactConfig) {
		c.Metadata = metadata
	}
}

// TaskOption configures task creation.
type TaskOption func(*TaskConfig)

// TaskConfig holds configuration for creating tasks.
type TaskConfig struct {
	Status       string
	Description  string
	Agent        string
	Dependencies []string
	Parallel     bool
}

// WithStatus sets the initial task status.
func WithStatus(status string) TaskOption {
	return func(c *TaskConfig) {
		c.Status = status
	}
}

// WithDescription sets the task description.
func WithDescription(description string) TaskOption {
	return func(c *TaskConfig) {
		c.Description = description
	}
}

// WithAgent sets the agent responsible for the task.
func WithAgent(agent string) TaskOption {
	return func(c *TaskConfig) {
		c.Agent = agent
	}
}

// WithDependencies sets task dependencies.
func WithDependencies(deps []string) TaskOption {
	return func(c *TaskConfig) {
		c.Dependencies = deps
	}
}

// WithParallel marks the task as parallelizable.
func WithParallel(parallel bool) TaskOption {
	return func(c *TaskConfig) {
		c.Parallel = parallel
	}
}

// PhaseOption configures phase operations.
type PhaseOption func(*PhaseConfig)

// PhaseConfig holds configuration for phase operations.
type PhaseConfig struct {
	// Phase-specific options (discovery type, etc.)
	Metadata map[string]interface{}
}

// WithPhaseMetadata adds metadata for phase operations.
func WithPhaseMetadata(metadata map[string]interface{}) PhaseOption {
	return func(c *PhaseConfig) {
		c.Metadata = metadata
	}
}

// LogOption configures logging.
type LogOption func(*LogEntry)

// LogEntry represents a structured log entry.
type LogEntry struct {
	Timestamp time.Time
	AgentID   string
	Action    string
	Result    string
	Files     []string
	Notes     string
}

// WithFiles adds files to the log entry.
func WithFiles(files ...string) LogOption {
	return func(e *LogEntry) {
		e.Files = append(e.Files, files...)
	}
}

// WithNotes adds notes to the log entry.
func WithNotes(notes string) LogOption {
	return func(e *LogEntry) {
		e.Notes = notes
	}
}
