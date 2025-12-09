package exec

import "context"

// Executor defines the interface for executing shell commands.
//
// This interface allows for easy mocking in tests while providing a consistent
// API for command execution across the codebase. Implementations wrap a single
// command (e.g., "gh", "git", "kubectl") and execute it with provided arguments.
//
// The interface is designed to be minimal while covering common use cases:
//   - Command identification and existence checks
//   - Synchronous execution with output capture
//   - Context support for cancellation and timeouts
//   - Silent execution when output isn't needed
//
// Example:
//
//	executor := exec.NewLocal("gh")
//	if !executor.Exists() {
//	    return fmt.Errorf("gh CLI not found")
//	}
//	stdout, stderr, err := executor.Run("issue", "list", "--label", "bug")
type Executor interface {
	// Command returns the command name this executor wraps.
	Command() string

	// Exists checks if the command exists in PATH.
	Exists() bool

	// Run executes the command with the given arguments.
	// Returns stdout, stderr, and error.
	Run(args ...string) (stdout, stderr string, err error)

	// RunContext executes the command with the given arguments and context.
	// The context can be used for cancellation and timeouts.
	RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error)

	// RunSilent executes the command but only returns an error.
	// Stdout and stderr are discarded.
	RunSilent(args ...string) error

	// RunSilentContext is like RunSilent but accepts a context for cancellation.
	RunSilentContext(ctx context.Context, args ...string) error
}
