//go:generate go run go.uber.org/mock/mockgen@latest -source=executor.go -destination=mock_executor.go -package=exec

// Package exec provides a clean command execution abstraction using the
// ports and adapters pattern.
//
// This package defines the [Executor] interface (port) for executing external
// commands. The interface enables dependency injection and easy mocking in tests,
// decoupling business logic from the actual command execution mechanism.
//
// # Interface Design
//
// The [Executor] interface is the core abstraction:
//
//	type Executor interface {
//	    Command() string
//	    Exists() bool
//	    Run(args ...string) (stdout, stderr string, err error)
//	    RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error)
//	    RunSilent(args ...string) error
//	    RunSilentContext(ctx context.Context, args ...string) error
//	}
//
// Each Executor wraps a single command (e.g., "gh", "git", "kubectl") and
// provides methods to check existence and execute with various options.
//
// # Usage
//
// Create an executor and run commands:
//
//	gh := exec.NewLocal("gh")
//	if !gh.Exists() {
//	    return fmt.Errorf("gh CLI not found")
//	}
//	stdout, stderr, err := gh.Run("issue", "list", "--label", "bug")
//
// Use context for timeouts and cancellation:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	stdout, stderr, err := gh.RunContext(ctx, "pr", "create")
//
// # Testing
//
// Use the generated mock for testing:
//
//	mock := exec.NewMockExecutor(ctrl)
//	mock.EXPECT().Exists().Return(true)
//	mock.EXPECT().Run("issue", "list").Return(`[{"number": 1}]`, "", nil)
//
// Or create a custom mock implementation:
//
//	type testExecutor struct {
//	    runFunc func(args ...string) (string, string, error)
//	}
//	func (t *testExecutor) Run(args ...string) (string, string, error) {
//	    return t.runFunc(args...)
//	}
//
// # Implementations
//
// This package provides:
//   - [LocalExecutor]: Executes commands on the local system using os/exec
//   - Generated mock via go:generate for testing
//
// See README.md for more examples.
package exec
