package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// LocalExecutor executes commands on the local system using os/exec.
type LocalExecutor struct {
	command string
}

// NewLocalExecutor creates a new LocalExecutor for the specified command.
//
// The command should be the base command name (e.g., "gh", "claude", "git").
// The executor will use exec.LookPath to find the command in PATH.
//
// Example:
//
//	gh := exec.NewLocalExecutor("gh")
//	claude := exec.NewLocalExecutor("claude")
func NewLocalExecutor(command string) *LocalExecutor {
	return &LocalExecutor{
		command: command,
	}
}

// Command returns the command name this executor wraps.
func (e *LocalExecutor) Command() string {
	return e.command
}

// Exists checks if the command exists in PATH.
//
// Returns true if the command can be found, false otherwise.
// This is useful for checking prerequisites before attempting to run commands.
//
// Example:
//
//	gh := exec.NewLocalExecutor("gh")
//	if !gh.Exists() {
//	    return fmt.Errorf("gh CLI not found")
//	}
func (e *LocalExecutor) Exists() bool {
	_, err := exec.LookPath(e.command)
	return err == nil
}

// Run executes the command with the given arguments.
//
// Returns stdout, stderr, and error. If the command fails, err will be non-nil
// and stderr will contain the error output.
//
// Example:
//
//	gh := exec.NewLocalExecutor("gh")
//	stdout, stderr, err := gh.Run("issue", "list", "--label", "bug")
//	if err != nil {
//	    fmt.Printf("Command failed: %s\n", stderr)
//	}
func (e *LocalExecutor) Run(args ...string) (stdout, stderr string, err error) {
	return e.RunContext(context.Background(), args...)
}

// RunContext executes the command with the given arguments and context.
//
// The context can be used for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	stdout, stderr, err := gh.RunContext(ctx, "pr", "create")
//
// Returns stdout, stderr, and error. If the command fails or is cancelled,
// err will be non-nil.
func (e *LocalExecutor) RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error) {
	cmd := exec.CommandContext(ctx, e.command, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()

	return stdoutBuf.String(), stderrBuf.String(), err
}

// RunSilent executes the command but only returns an error.
//
// This is useful when you don't care about the output, only success/failure.
// Stdout and stderr are discarded.
//
// Example:
//
//	gh := exec.NewLocalExecutor("gh")
//	if err := gh.RunSilent("auth", "status"); err != nil {
//	    return fmt.Errorf("not authenticated")
//	}
func (e *LocalExecutor) RunSilent(args ...string) error {
	_, _, err := e.Run(args...)
	return err
}

// RunSilentContext is like RunSilent but accepts a context for cancellation.
func (e *LocalExecutor) RunSilentContext(ctx context.Context, args ...string) error {
	_, _, err := e.RunContext(ctx, args...)
	return err
}

// MustExist panics if the command doesn't exist in PATH.
//
// This is useful in initialization code where the command is required.
//
// Example:
//
//	func init() {
//	    gh := exec.NewLocalExecutor("gh")
//	    gh.MustExist() // Panic if gh not installed
//	}
func (e *LocalExecutor) MustExist() {
	if !e.Exists() {
		panic(fmt.Sprintf("required command %q not found in PATH", e.command))
	}
}

// Compile-time check that LocalExecutor implements Executor.
var _ Executor = (*LocalExecutor)(nil)
