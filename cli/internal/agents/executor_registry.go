package agents

import (
	"fmt"

	"github.com/jmgilman/sow/cli/schemas"
)

// ExecutorRegistry provides lookup and listing of registered executors.
// It is the central registry for all executor implementations in the system.
//
// The registry is designed to be populated at initialization time with
// available executors. Thread safety is not required since registration
// happens only during initialization.
//
// Example:
//
//	registry := agents.NewExecutorRegistry()
//	registry.Register(NewClaudeExecutor(true, "sonnet"))
//	executor, err := registry.Get("claude-code")
//	if err != nil {
//	    log.Fatal(err)
//	}
type ExecutorRegistry struct {
	executors map[string]Executor
}

// NewExecutorRegistry creates a new empty ExecutorRegistry.
// Unlike NewAgentRegistry, this does not pre-populate with defaults
// because executors may need configuration (yoloMode, model, etc.).
//
// Example:
//
//	registry := agents.NewExecutorRegistry()
//	registry.Register(executor)
func NewExecutorRegistry() *ExecutorRegistry {
	return &ExecutorRegistry{
		executors: make(map[string]Executor),
	}
}

// Register adds an executor to the registry.
// The executor is registered under its Name().
// Panics if an executor with the same name is already registered.
//
// Example:
//
//	registry := NewExecutorRegistry()
//	registry.Register(executor)
func (r *ExecutorRegistry) Register(executor Executor) {
	name := executor.Name()
	if _, exists := r.executors[name]; exists {
		panic(fmt.Sprintf("executor already registered: %s", name))
	}

	r.executors[name] = executor
}

// Get returns an executor by name.
// Returns (executor, nil) if found, (nil, error) if not found.
//
// Example:
//
//	executor, err := registry.Get("claude-code")
//	if err != nil {
//	    return fmt.Errorf("failed to get executor: %w", err)
//	}
func (r *ExecutorRegistry) Get(name string) (Executor, error) {
	executor, ok := r.executors[name]
	if !ok {
		return nil, fmt.Errorf("unknown executor: %s", name)
	}

	return executor, nil
}

// List returns the names of all registered executors.
// The order is not guaranteed.
//
// Example:
//
//	names := registry.List()
//	fmt.Println("Available executors:", strings.Join(names, ", "))
func (r *ExecutorRegistry) List() []string {
	names := make([]string, 0, len(r.executors))
	for name := range r.executors {
		names = append(names, name)
	}

	return names
}

// RegisterNamed adds an executor to the registry under a specific name.
// This allows registering the same executor type under different names
// (e.g., "claude-opus" and "claude-sonnet" both using ClaudeExecutor).
// Panics if an executor with the same name is already registered.
//
// Example:
//
//	registry := NewExecutorRegistry()
//	registry.RegisterNamed("claude-opus", NewClaudeExecutor(true, "opus", "", nil))
func (r *ExecutorRegistry) RegisterNamed(name string, executor Executor) {
	if _, exists := r.executors[name]; exists {
		panic(fmt.Sprintf("executor already registered: %s", name))
	}

	r.executors[name] = executor
}

// DefaultExecutorName is the default executor name used when no config exists.
const DefaultExecutorName = "claude-code"

// LoadExecutorRegistry creates an ExecutorRegistry populated from user configuration.
// It creates executor instances based on the config and registers them under their
// configured names.
//
// If userConfig is nil or has no executors defined, a default registry is created
// with a single "claude-code" executor using safe defaults.
//
// Parameters:
//   - userConfig: user configuration with executor definitions
//   - outputDir: directory for agent output logs (can be empty to disable logging)
//
// Example:
//
//	registry, err := LoadExecutorRegistry(userConfig, ".sow/project/agent-outputs")
//	if err != nil {
//	    return err
//	}
//	executor, err := registry.Get("claude-code")
func LoadExecutorRegistry(userConfig *schemas.UserConfig, outputDir string) (*ExecutorRegistry, error) {
	registry := NewExecutorRegistry()

	// If no config or no executors defined, create default
	if userConfig == nil || userConfig.Agents == nil || len(userConfig.Agents.Executors) == 0 {
		// Register default claude-code executor
		registry.RegisterNamed(DefaultExecutorName, NewClaudeExecutor(false, "", outputDir, nil))
		return registry, nil
	}

	// Create executors from config
	for name, execConfig := range userConfig.Agents.Executors {
		var yoloMode bool
		var model string
		var customArgs []string

		if execConfig.Settings != nil {
			if execConfig.Settings.Yolo_mode != nil {
				yoloMode = *execConfig.Settings.Yolo_mode
			}
			if execConfig.Settings.Model != nil {
				model = *execConfig.Settings.Model
			}
		}
		customArgs = execConfig.Custom_args

		var executor Executor
		switch execConfig.Type {
		case "claude":
			executor = NewClaudeExecutor(yoloMode, model, outputDir, customArgs)
		case "cursor":
			executor = NewCursorExecutor(yoloMode, outputDir, customArgs)
		default:
			return nil, fmt.Errorf("unknown executor type %q for executor %q", execConfig.Type, name)
		}

		registry.RegisterNamed(name, executor)
	}

	return registry, nil
}

// GetAgentExecutor looks up the executor for an agent based on bindings.
// It first finds the executor name from bindings, then looks up the executor.
//
// Parameters:
//   - agentName: the agent role (e.g., "implementer", "reviewer")
//   - bindings: the bindings configuration from user config (can be nil)
//
// Returns the executor for the agent, or error if not found.
func (r *ExecutorRegistry) GetAgentExecutor(agentName string, bindings *struct {
	Orchestrator *string `json:"orchestrator,omitempty"`
	Implementer  *string `json:"implementer,omitempty"`
	Architect    *string `json:"architect,omitempty"`
	Reviewer     *string `json:"reviewer,omitempty"`
	Planner      *string `json:"planner,omitempty"`
	Researcher   *string `json:"researcher,omitempty"`
	Decomposer   *string `json:"decomposer,omitempty"`
}) (Executor, error) {
	// Determine executor name from bindings
	executorName := DefaultExecutorName
	if bindings != nil {
		switch agentName {
		case "orchestrator":
			if bindings.Orchestrator != nil {
				executorName = *bindings.Orchestrator
			}
		case "implementer":
			if bindings.Implementer != nil {
				executorName = *bindings.Implementer
			}
		case "architect":
			if bindings.Architect != nil {
				executorName = *bindings.Architect
			}
		case "reviewer":
			if bindings.Reviewer != nil {
				executorName = *bindings.Reviewer
			}
		case "planner":
			if bindings.Planner != nil {
				executorName = *bindings.Planner
			}
		case "researcher":
			if bindings.Researcher != nil {
				executorName = *bindings.Researcher
			}
		case "decomposer":
			if bindings.Decomposer != nil {
				executorName = *bindings.Decomposer
			}
		}
	}

	return r.Get(executorName)
}
