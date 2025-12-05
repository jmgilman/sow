package agents

import "fmt"

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
