// Package domain defines the core interfaces and types for the project system.
package domain

import (
	"github.com/jmgilman/sow/cli/internal/logging"
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
	Log(action, result string, opts ...logging.LogOption) error

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
	ApproveArtifact(path string) (*PhaseOperationResult, error)
	ListArtifacts() []*phases.Artifact

	// Task operations (implementation only)
	// Returns concrete Task type - single implementation for all project types
	AddTask(name string, opts ...TaskOption) (*Task, error)
	GetTask(id string) (*Task, error)
	ListTasks() []*Task
	ApproveTasks() (*PhaseOperationResult, error)

	// Generic field access (for metadata)
	Set(field string, value interface{}) (*PhaseOperationResult, error)
	Get(field string) (interface{}, error)

	// Lifecycle
	Complete() (*PhaseOperationResult, error)
	Skip() error
	Enable(opts ...PhaseOption) error

	// Advance to next state within this phase
	// Returns ErrNotSupported if phase has no internal states
	Advance() (*PhaseOperationResult, error)
}

// ArtifactOption configures artifact creation.
type ArtifactOption func(*ArtifactConfig)

// ArtifactConfig holds configuration for creating artifacts.
type ArtifactConfig struct {
	Type       *string
	Assessment *string
	Metadata   map[string]interface{}
}

// WithType sets the artifact type.
func WithType(artifactType *string) ArtifactOption {
	return func(c *ArtifactConfig) {
		c.Type = artifactType
	}
}

// WithAssessment sets the artifact assessment.
func WithAssessment(assessment *string) ArtifactOption {
	return func(c *ArtifactConfig) {
		c.Assessment = assessment
	}
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

// LogOption is an alias for logging.LogOption for backward compatibility.
type LogOption = logging.LogOption

// LogEntry is an alias for logging.LogEntry for backward compatibility.
type LogEntry = logging.LogEntry

// Re-exported logging helper functions for backward compatibility.
var (
	// WithFiles adds files to the log entry.
	WithFiles = logging.WithFiles
	// WithNotes adds notes to the log entry.
	WithNotes = logging.WithNotes
)
