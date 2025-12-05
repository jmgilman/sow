package agents

import (
	"context"
	"io"
)

// MockExecutor is a mock implementation of Executor for testing.
// Uses function fields for configurable behavior.
//
// Usage in tests:
//
//	mock := &MockExecutor{
//	    NameFunc: func() string { return "test-executor" },
//	    SpawnFunc: func(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
//	        // verify arguments, return error, etc.
//	        return nil
//	    },
//	    SupportsResumptionFunc: func() bool { return true },
//	}
//
// When function fields are nil, methods return sensible defaults:
//   - Name() returns ""
//   - Spawn() returns nil
//   - Resume() returns nil
//   - SupportsResumption() returns false
type MockExecutor struct {
	NameFunc               func() string
	SpawnFunc              func(ctx context.Context, agent *Agent, prompt string, sessionID string) error
	ResumeFunc             func(ctx context.Context, sessionID string, prompt string) error
	SupportsResumptionFunc func() bool
}

// Name calls the mock function if set, otherwise returns empty string.
func (m *MockExecutor) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return ""
}

// Spawn calls the mock function if set, otherwise returns nil.
func (m *MockExecutor) Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
	if m.SpawnFunc != nil {
		return m.SpawnFunc(ctx, agent, prompt, sessionID)
	}
	return nil
}

// Resume calls the mock function if set, otherwise returns nil.
func (m *MockExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
	if m.ResumeFunc != nil {
		return m.ResumeFunc(ctx, sessionID, prompt)
	}
	return nil
}

// SupportsResumption calls the mock function if set, otherwise returns false.
func (m *MockExecutor) SupportsResumption() bool {
	if m.SupportsResumptionFunc != nil {
		return m.SupportsResumptionFunc()
	}
	return false
}

// Compile-time check that MockExecutor implements Executor.
var _ Executor = (*MockExecutor)(nil)

// MockCommandRunner is a mock implementation of CommandRunner for testing.
// Uses function field for configurable behavior and captures call parameters
// for verification in tests.
//
// Usage in tests:
//
//	runner := &MockCommandRunner{
//	    RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
//	        if name != "claude" {
//	            t.Errorf("wrong command: got %s, want claude", name)
//	        }
//	        return nil
//	    },
//	}
//
// After calling Run, you can verify the captured parameters:
//
//	runner.Run(ctx, "cmd", []string{"arg1"}, strings.NewReader("input"))
//	if runner.LastName != "cmd" {
//	    t.Errorf("LastName = %q, want %q", runner.LastName, "cmd")
//	}
//	if runner.LastStdin != "input" {
//	    t.Errorf("LastStdin = %q, want %q", runner.LastStdin, "input")
//	}
type MockCommandRunner struct {
	RunFunc func(ctx context.Context, name string, args []string, stdin io.Reader) error

	// Captured parameters from the last Run call for verification.
	LastName  string
	LastArgs  []string
	LastStdin string
}

// Run captures call parameters and calls the mock function if set.
// If RunFunc is nil, returns nil.
// stdin is read into LastStdin; if stdin is nil, LastStdin is set to "".
func (m *MockCommandRunner) Run(ctx context.Context, name string, args []string, stdin io.Reader) error {
	// Capture call parameters.
	m.LastName = name
	m.LastArgs = args
	if stdin != nil {
		data, _ := io.ReadAll(stdin)
		m.LastStdin = string(data)
	} else {
		m.LastStdin = ""
	}

	if m.RunFunc != nil {
		return m.RunFunc(ctx, name, args, stdin)
	}
	return nil
}

// Compile-time check that MockCommandRunner implements CommandRunner.
var _ CommandRunner = (*MockCommandRunner)(nil)
