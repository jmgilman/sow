package agents

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Executor invokes agent CLIs and manages sessions.
// Each implementation handles a specific CLI tool (Claude, Cursor, etc.).
// The interface supports spawning new agent sessions and resuming existing ones.
//
// Executor methods block until the subprocess exits. Session state is persisted
// externally (e.g., in task state.yaml files), not in Go memory. This design
// supports the spawn -> block -> exit -> read state workflow.
//
// Example usage:
//
//	func spawnWorker(ctx context.Context, executor Executor, agent *Agent, taskID string) error {
//	    sessionID := uuid.New().String()
//	    prompt := fmt.Sprintf("Execute task %s", taskID)
//	    return executor.Spawn(ctx, agent, prompt, sessionID)
//	}
type Executor interface {
	// Name returns the executor identifier (e.g., "claude-code", "cursor").
	// Used for registry lookup and configuration.
	Name() string

	// Spawn invokes an agent with the given prompt and session ID.
	// Blocks until the subprocess exits.
	// The sessionID should be persisted before calling Spawn.
	// Returns error if subprocess fails or context is cancelled.
	Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error

	// Resume continues an existing session with additional prompt.
	// Blocks until the subprocess exits.
	// Returns error if session not found, executor doesn't support resumption,
	// subprocess fails, or context is cancelled.
	Resume(ctx context.Context, sessionID string, prompt string) error

	// SupportsResumption indicates if this executor can resume sessions.
	// Some CLIs may not support session resumption.
	SupportsResumption() bool
}

// CommandRunner abstracts subprocess execution for testability.
// In production, this is backed by os/exec. In tests, it's mocked.
//
// This abstraction allows executor implementations to be tested without
// actually spawning CLI processes.
type CommandRunner interface {
	// Run executes a command with the given arguments and stdin.
	// Returns error if command fails or context is cancelled.
	Run(ctx context.Context, name string, args []string, stdin io.Reader) error
}

// DefaultCommandRunner implements CommandRunner using os/exec.
// It connects subprocess stdout/stderr to os.Stdout/os.Stderr for
// interactive CLI use.
type DefaultCommandRunner struct{}

// Run executes a command with the given arguments and stdin.
// The subprocess stdout and stderr are connected to the parent process's
// stdout and stderr for interactive CLI experience.
func (r *DefaultCommandRunner) Run(ctx context.Context, name string, args []string, stdin io.Reader) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %s failed: %w", name, err)
	}
	return nil
}

// Compile-time check that DefaultCommandRunner implements CommandRunner.
var _ CommandRunner = (*DefaultCommandRunner)(nil)
