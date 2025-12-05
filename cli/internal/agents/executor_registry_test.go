package agents

import (
	"strings"
	"testing"
)

// TestNewExecutorRegistry verifies that NewExecutorRegistry creates an empty registry.
func TestNewExecutorRegistry(t *testing.T) {
	registry := NewExecutorRegistry()

	if registry == nil {
		t.Fatal("NewExecutorRegistry() returned nil")
	}

	// Should be empty (unlike AgentRegistry, no pre-population)
	executors := registry.List()
	if len(executors) != 0 {
		t.Errorf("NewExecutorRegistry() should be empty, got %d executors", len(executors))
	}
}

// TestExecutorRegistry_Register verifies that Register adds executors to the registry.
func TestExecutorRegistry_Register(t *testing.T) {
	registry := NewExecutorRegistry()

	mock := &MockExecutor{
		NameFunc: func() string { return "test-executor" },
	}

	// Register should not panic
	registry.Register(mock)

	// Verify executor was added
	executor, err := registry.Get("test-executor")
	if err != nil {
		t.Fatalf("Get(test-executor) error = %v", err)
	}
	if executor == nil {
		t.Fatal("Get(test-executor) returned nil executor")
	}
	if executor.Name() != "test-executor" {
		t.Errorf("Get(test-executor).Name() = %q, want %q", executor.Name(), "test-executor")
	}
}

// TestExecutorRegistry_RegisterMultiple verifies that multiple executors can be registered.
func TestExecutorRegistry_RegisterMultiple(t *testing.T) {
	registry := NewExecutorRegistry()

	mock1 := &MockExecutor{
		NameFunc: func() string { return "executor-1" },
	}
	mock2 := &MockExecutor{
		NameFunc: func() string { return "executor-2" },
	}
	mock3 := &MockExecutor{
		NameFunc: func() string { return "executor-3" },
	}

	registry.Register(mock1)
	registry.Register(mock2)
	registry.Register(mock3)

	// Should have all 3 executors
	executors := registry.List()
	if len(executors) != 3 {
		t.Errorf("List() returned %d executors, want 3", len(executors))
	}

	// Verify each can be retrieved
	for _, name := range []string{"executor-1", "executor-2", "executor-3"} {
		executor, err := registry.Get(name)
		if err != nil {
			t.Errorf("Get(%q) error = %v", name, err)
		}
		if executor == nil {
			t.Errorf("Get(%q) returned nil", name)
		}
	}
}

// TestExecutorRegistry_Get verifies that Get returns correct executors.
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

// TestExecutorRegistry_GetErrorMessage verifies the error message format.
func TestExecutorRegistry_GetErrorMessage(t *testing.T) {
	registry := NewExecutorRegistry()

	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent executor")
	}

	// Error should be lowercase and contain the executor name
	errMsg := err.Error()
	if !strings.Contains(errMsg, "unknown executor") {
		t.Errorf("error message = %q, want to contain 'unknown executor'", errMsg)
	}
	if !strings.Contains(errMsg, "nonexistent") {
		t.Errorf("error message = %q, want to contain 'nonexistent'", errMsg)
	}
}

// TestExecutorRegistry_ListEmpty verifies that List returns empty slice for empty registry.
func TestExecutorRegistry_ListEmpty(t *testing.T) {
	registry := NewExecutorRegistry()
	executors := registry.List()

	// List should return an empty slice, not nil
	if executors == nil {
		t.Error("List() returned nil, want empty slice")
	}
	if len(executors) != 0 {
		t.Errorf("List() returned %d executors, want 0", len(executors))
	}
}

// TestExecutorRegistry_List verifies that List returns all registered executor names.
func TestExecutorRegistry_List(t *testing.T) {
	registry := NewExecutorRegistry()

	mock1 := &MockExecutor{
		NameFunc: func() string { return "claude-code" },
	}
	mock2 := &MockExecutor{
		NameFunc: func() string { return "cursor" },
	}

	registry.Register(mock1)
	registry.Register(mock2)

	names := registry.List()

	// Should have 2 names
	if len(names) != 2 {
		t.Errorf("List() returned %d names, want 2", len(names))
	}

	// Verify both names are present (order not guaranteed)
	expectedNames := map[string]bool{
		"claude-code": false,
		"cursor":      false,
	}

	for _, name := range names {
		if _, ok := expectedNames[name]; ok {
			expectedNames[name] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("List() missing expected name: %s", name)
		}
	}
}

// TestExecutorRegistry_ListAfterRegister verifies that List includes newly registered executors.
func TestExecutorRegistry_ListAfterRegister(t *testing.T) {
	registry := NewExecutorRegistry()

	// Initially empty
	if len(registry.List()) != 0 {
		t.Errorf("List() should be empty initially")
	}

	// Register one
	registry.Register(&MockExecutor{
		NameFunc: func() string { return "first" },
	})
	if len(registry.List()) != 1 {
		t.Errorf("List() should have 1 executor after first Register")
	}

	// Register another
	registry.Register(&MockExecutor{
		NameFunc: func() string { return "second" },
	})
	if len(registry.List()) != 2 {
		t.Errorf("List() should have 2 executors after second Register")
	}
}

// TestExecutorRegistry_RegisterDuplicatePanics verifies that registering a duplicate executor panics.
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

	// Should panic - same name already registered
	registry.Register(mock)
}

// TestExecutorRegistry_RegisterUsesNameMethod verifies that Register uses executor.Name() as key.
func TestExecutorRegistry_RegisterUsesNameMethod(t *testing.T) {
	registry := NewExecutorRegistry()

	// Create mock with a specific Name() return value
	mock := &MockExecutor{
		NameFunc: func() string { return "my-executor-name" },
	}

	registry.Register(mock)

	// Should be able to retrieve by the Name() value
	executor, err := registry.Get("my-executor-name")
	if err != nil {
		t.Fatalf("Get(my-executor-name) error = %v", err)
	}
	if executor.Name() != "my-executor-name" {
		t.Errorf("Get(my-executor-name).Name() = %q, want %q", executor.Name(), "my-executor-name")
	}
}

// TestExecutorRegistry_ReturnsCorrectExecutor verifies that Get returns the exact registered executor.
func TestExecutorRegistry_ReturnsCorrectExecutor(t *testing.T) {
	registry := NewExecutorRegistry()

	mock1 := &MockExecutor{
		NameFunc:               func() string { return "executor-a" },
		SupportsResumptionFunc: func() bool { return true },
	}
	mock2 := &MockExecutor{
		NameFunc:               func() string { return "executor-b" },
		SupportsResumptionFunc: func() bool { return false },
	}

	registry.Register(mock1)
	registry.Register(mock2)

	// Get executor-a and verify it's the right one
	execA, err := registry.Get("executor-a")
	if err != nil {
		t.Fatalf("Get(executor-a) error = %v", err)
	}
	if !execA.SupportsResumption() {
		t.Error("Got wrong executor - executor-a should support resumption")
	}

	// Get executor-b and verify it's the right one
	execB, err := registry.Get("executor-b")
	if err != nil {
		t.Fatalf("Get(executor-b) error = %v", err)
	}
	if execB.SupportsResumption() {
		t.Error("Got wrong executor - executor-b should not support resumption")
	}
}
