# Executor Registry

## Context

This task is part of building the Executor System for sow - a multi-agent orchestration framework. The ExecutorRegistry provides a central lookup mechanism for executor instances, similar to the existing AgentRegistry pattern.

**Why this is needed:**
- The orchestrator needs to look up executors by name (e.g., "claude-code", "cursor")
- User configuration binds agent roles to executor names
- Registry pattern matches the existing AgentRegistry for consistency
- Supports listing available executors for help text and validation

**Design reference:** The multi-agent architecture design document specifies ExecutorRegistry with Register, Get, List methods.

## Requirements

### 1. Create ExecutorRegistry

Create `cli/internal/agents/executor_registry.go`:

```go
// ExecutorRegistry provides lookup and listing of registered executors.
// It is the central registry for all executor implementations in the system.
//
// The registry is designed to be populated at initialization time with
// available executors. Thread safety is not required since registration
// happens only during initialization.
type ExecutorRegistry struct {
    executors map[string]Executor
}

// NewExecutorRegistry creates a new empty ExecutorRegistry.
// Unlike NewAgentRegistry, this does not pre-populate with defaults
// because executors may need configuration (yoloMode, model, etc.).
func NewExecutorRegistry() *ExecutorRegistry

// Register adds an executor to the registry.
// The executor is registered under its Name().
// Panics if an executor with the same name is already registered.
func (r *ExecutorRegistry) Register(executor Executor)

// Get returns an executor by name.
// Returns (executor, nil) if found, (nil, error) if not found.
func (r *ExecutorRegistry) Get(name string) (Executor, error)

// List returns the names of all registered executors.
// The order is not guaranteed.
func (r *ExecutorRegistry) List() []string
```

### 2. Follow Existing Registry Pattern

The ExecutorRegistry should follow the same pattern as AgentRegistry in `cli/internal/agents/registry.go`:
- Map-based storage
- Panic on duplicate registration
- Return error for unknown executor lookup
- List returns all registered items

### 3. Key Differences from AgentRegistry

- **No pre-population**: ExecutorRegistry starts empty because executors need configuration
- **List returns names**: Returns `[]string` instead of `[]Executor` for simpler usage in help text
- **Register uses Name()**: The executor's Name() method determines its key

## Acceptance Criteria

1. **ExecutorRegistry struct** with internal executors map

2. **NewExecutorRegistry()** creates empty registry
   - Returns initialized registry with empty map
   - Does NOT pre-populate with any executors

3. **Register()** adds executors to registry
   - Uses `executor.Name()` as the key
   - Panics if duplicate name registered (with clear error message)

4. **Get()** retrieves executors by name
   - Returns (executor, nil) when found
   - Returns (nil, error) when not found
   - Error message follows format: "unknown executor: {name}"

5. **List()** returns all executor names
   - Returns slice of strings (not Executor objects)
   - Returns empty slice if no executors registered
   - Order not guaranteed

6. **Unit tests cover**:
   - Creating empty registry
   - Registering single executor
   - Registering multiple executors
   - Getting registered executor
   - Getting unknown executor (error case)
   - Getting empty string (error case)
   - Listing no executors
   - Listing after registration
   - Duplicate registration panics
   - Error message format verification

## Technical Details

### Package Location

File: `cli/internal/agents/executor_registry.go`
Test: `cli/internal/agents/executor_registry_test.go`

### Import Dependencies

```go
import "fmt"
```

### Code Style

Follow existing AgentRegistry patterns:
- Godoc comments on all exported types and methods
- Example usage in godoc
- Table-driven tests
- Test helper for panic verification (see `TestAgentRegistry_RegisterDuplicatePanics`)

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/registry.go` - AgentRegistry implementation to follow as pattern
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/registry_test.go` - AgentRegistry tests to follow as pattern
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/executor.go` - Executor interface (from task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/executor_mock.go` - MockExecutor for tests (from task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/.sow/knowledge/designs/multi-agent-architecture.md` - Design document specifying ExecutorRegistry (lines 430-450)

## Examples

### Creating and Populating Registry

```go
// Create empty registry
registry := NewExecutorRegistry()

// Register executors (typically done at initialization)
registry.Register(NewClaudeExecutor(true, "sonnet"))
registry.Register(NewCursorExecutor(true))

// Look up by name
executor, err := registry.Get("claude-code")
if err != nil {
    return fmt.Errorf("failed to get executor: %w", err)
}

// List available executors for help text
names := registry.List()
fmt.Println("Available executors:", strings.Join(names, ", "))
```

### Test Examples

```go
func TestExecutorRegistry_Get(t *testing.T) {
    tests := []struct {
        name         string
        executorName string
        wantError    bool
    }{
        {
            name:         "registered executor",
            executorName: "test-executor",
            wantError:    false,
        },
        {
            name:         "unknown executor",
            executorName: "unknown",
            wantError:    true,
        },
        {
            name:         "empty string",
            executorName: "",
            wantError:    true,
        },
    }

    registry := NewExecutorRegistry()
    registry.Register(&MockExecutor{
        NameFunc: func() string { return "test-executor" },
    })

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            executor, err := registry.Get(tt.executorName)
            if (err != nil) != tt.wantError {
                t.Errorf("Get(%q) error = %v, wantError %v", tt.executorName, err, tt.wantError)
                return
            }
            if !tt.wantError && executor == nil {
                t.Errorf("Get(%q) returned nil executor", tt.executorName)
            }
        })
    }
}

func TestExecutorRegistry_RegisterDuplicatePanics(t *testing.T) {
    registry := NewExecutorRegistry()

    mock := &MockExecutor{
        NameFunc: func() string { return "test-executor" },
    }

    registry.Register(mock)

    defer func() {
        if r := recover(); r == nil {
            t.Error("expected panic on duplicate registration")
        } else {
            msg, ok := r.(string)
            if !ok {
                t.Error("expected panic message to be a string")
            }
            if !strings.Contains(msg, "test-executor") {
                t.Errorf("panic message should mention executor name: %s", msg)
            }
            if !strings.Contains(msg, "already registered") {
                t.Errorf("panic message should contain 'already registered': %s", msg)
            }
        }
    }()

    // Should panic
    registry.Register(mock)
}
```

## Dependencies

- **Task 010** (Executor Interface): Requires `Executor` interface and `MockExecutor`

## Constraints

- Do NOT pre-populate registry with default executors
- Do NOT add thread safety (registration is initialization-time only)
- Registry keys must come from `executor.Name()`, not separate parameter
- Follow exact same patterns as AgentRegistry for consistency
