package sow

// phaseConfig holds configuration for phase enablement.
type phaseConfig struct {
	discoveryType string
}

// PhaseOption is a functional option for phase operations.
type PhaseOption func(*phaseConfig)

// WithDiscoveryType sets the discovery type (bug, feature, docs, refactor, general).
func WithDiscoveryType(discoveryType string) PhaseOption {
	return func(cfg *phaseConfig) {
		cfg.discoveryType = discoveryType
	}
}

// taskConfig holds configuration for task creation.
type taskConfig struct {
	status       string
	parallel     bool
	dependencies []string
	agent        string
	description  string
}

// TaskOption is a functional option for task operations.
type TaskOption func(*taskConfig)

// WithStatus sets the initial task status.
func WithStatus(status string) TaskOption {
	return func(cfg *taskConfig) {
		cfg.status = status
	}
}

// WithParallel marks the task as executable in parallel.
func WithParallel(parallel bool) TaskOption {
	return func(cfg *taskConfig) {
		cfg.parallel = parallel
	}
}

// WithDependencies sets task dependencies.
func WithDependencies(deps ...string) TaskOption {
	return func(cfg *taskConfig) {
		cfg.dependencies = deps
	}
}

// WithAgent sets the assigned agent.
func WithAgent(agent string) TaskOption {
	return func(cfg *taskConfig) {
		cfg.agent = agent
	}
}

// WithDescription sets the task description.
func WithDescription(description string) TaskOption {
	return func(cfg *taskConfig) {
		cfg.description = description
	}
}
