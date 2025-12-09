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
// Use the generated mock from the mocks subpackage for testing:
//
//	import "github.com/jmgilman/sow/libs/exec/mocks"
//
//	mock := &mocks.ExecutorMock{
//	    ExistsFunc: func() bool { return true },
//	    RunFunc: func(args ...string) (string, string, error) {
//	        return `[{"number": 1}]`, "", nil
//	    },
//	}
//
// # Implementations
//
// This package provides:
//   - [LocalExecutor]: Executes commands on the local system using os/exec
//   - Generated mock via moq in the mocks subpackage for testing
//
// See README.md for more examples.
package exec
