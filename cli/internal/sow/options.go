package sow

// PhaseConfig holds configuration for phase enablement.
// Exported for use by project package.
type PhaseConfig struct {
	discoveryType string
}

// DiscoveryType returns the discovery type.
func (c *PhaseConfig) DiscoveryType() string {
	return c.discoveryType
}

// PhaseOption is a functional option for phase operations.
type PhaseOption func(*PhaseConfig)

// WithDiscoveryType sets the discovery type (bug, feature, docs, refactor, general).
func WithDiscoveryType(discoveryType string) PhaseOption {
	return func(cfg *PhaseConfig) {
		cfg.discoveryType = discoveryType
	}
}

// TaskConfig holds configuration for task creation.
// Exported for use by project package.
type TaskConfig struct {
	id           string
	Status       string   // Exported for direct access
	Parallel     bool     // Exported for direct access
	dependencies []string
	Agent        string // Exported for direct access
	description  string
}

// ID returns the task ID.
func (c *TaskConfig) ID() string {
	return c.id
}

// Dependencies returns the task dependencies.
func (c *TaskConfig) Dependencies() []string {
	return c.dependencies
}

// Description returns the task description.
func (c *TaskConfig) Description() string {
	return c.description
}

// TaskOption is a functional option for task operations.
type TaskOption func(*TaskConfig)

// WithStatus sets the initial task status.
func WithStatus(status string) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.Status = status
	}
}

// WithParallel marks the task as executable in parallel.
func WithParallel(parallel bool) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.Parallel = parallel
	}
}

// WithDependencies sets task dependencies.
func WithDependencies(deps ...string) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.dependencies = deps
	}
}

// WithAgent sets the assigned agent.
func WithAgent(agent string) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.Agent = agent
	}
}

// WithDescription sets the task description.
func WithDescription(description string) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.description = description
	}
}

// WithID sets an explicit task ID (otherwise auto-generated).
func WithID(id string) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.id = id
	}
}
