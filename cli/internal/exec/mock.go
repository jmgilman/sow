package exec

import "context"

// MockExecutor is a mock implementation of Executor for testing.
//
// Usage in tests:
//
//	mock := &exec.MockExecutor{
//	    ExistsFunc: func() bool { return true },
//	    RunFunc: func(args ...string) (string, string, error) {
//	        return `{"number": 123}`, "", nil
//	    },
//	}
//	github := sow.NewGitHub(mock)
type MockExecutor struct {
	CommandFunc           func() string
	ExistsFunc            func() bool
	RunFunc               func(args ...string) (stdout, stderr string, err error)
	RunContextFunc        func(ctx context.Context, args ...string) (stdout, stderr string, err error)
	RunSilentFunc         func(args ...string) error
	RunSilentContextFunc  func(ctx context.Context, args ...string) error
}

// Command calls the mock function if set, otherwise returns "mock-command".
func (m *MockExecutor) Command() string {
	if m.CommandFunc != nil {
		return m.CommandFunc()
	}
	return "mock-command"
}

// Exists calls the mock function if set, otherwise returns true.
func (m *MockExecutor) Exists() bool {
	if m.ExistsFunc != nil {
		return m.ExistsFunc()
	}
	return true
}

// Run calls the mock function if set, otherwise returns empty strings and nil error.
func (m *MockExecutor) Run(args ...string) (stdout, stderr string, err error) {
	if m.RunFunc != nil {
		return m.RunFunc(args...)
	}
	return "", "", nil
}

// RunContext calls the mock function if set, otherwise delegates to Run.
func (m *MockExecutor) RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error) {
	if m.RunContextFunc != nil {
		return m.RunContextFunc(ctx, args...)
	}
	return m.Run(args...)
}

// RunSilent calls the mock function if set, otherwise returns nil.
func (m *MockExecutor) RunSilent(args ...string) error {
	if m.RunSilentFunc != nil {
		return m.RunSilentFunc(args...)
	}
	_, _, err := m.Run(args...)
	return err
}

// RunSilentContext calls the mock function if set, otherwise delegates to RunSilent.
func (m *MockExecutor) RunSilentContext(ctx context.Context, args ...string) error {
	if m.RunSilentContextFunc != nil {
		return m.RunSilentContextFunc(ctx, args...)
	}
	return m.RunSilent(args...)
}

// Compile-time check that MockExecutor implements Executor.
var _ Executor = (*MockExecutor)(nil)
