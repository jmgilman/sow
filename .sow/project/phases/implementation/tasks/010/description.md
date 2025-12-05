# Executor Interface and Core Types

## Context

This task is part of building the Executor System for sow - a multi-agent orchestration framework. The executor system enables the sow orchestrator to spawn AI CLI tools (Claude Code, Cursor) with agent prompts and manage resumable sessions for bidirectional communication.

**Why this is needed:**
- The orchestrator needs to programmatically invoke worker agents (implementer, architect, reviewer, etc.)
- Workers run as subprocesses of AI CLI tools (Claude Code, Cursor)
- Sessions must be resumable to support paused workflows and review iterations
- The interface must be CLI-agnostic to support multiple executor implementations

**Design reference:** The multi-agent architecture design document specifies the Executor interface and session management protocol. This task creates the foundational interface that ClaudeExecutor and CursorExecutor will implement.

## Requirements

### 1. Define the Executor Interface

Create `cli/internal/agents/executor.go` with the `Executor` interface:

```go
// Executor invokes agent CLIs and manages sessions.
// Each implementation handles a specific CLI tool (Claude, Cursor, etc.).
// The interface supports spawning new agent sessions and resuming existing ones.
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
```

### 2. Create CommandRunner Interface for Testability

The executor implementations need to spawn subprocesses. Create a `CommandRunner` interface to enable mocking in tests:

```go
// CommandRunner abstracts subprocess execution for testability.
// In production, this is backed by os/exec. In tests, it's mocked.
type CommandRunner interface {
    // Run executes a command with the given arguments and stdin.
    // Returns error if command fails or context is cancelled.
    Run(ctx context.Context, name string, args []string, stdin io.Reader) error
}

// DefaultCommandRunner implements CommandRunner using os/exec.
type DefaultCommandRunner struct{}

func (r *DefaultCommandRunner) Run(ctx context.Context, name string, args []string, stdin io.Reader) error {
    cmd := exec.CommandContext(ctx, name, args...)
    cmd.Stdin = stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

### 3. Create MockExecutor for Testing

Create `cli/internal/agents/executor_mock.go` with a mock implementation following the codebase pattern:

```go
// MockExecutor is a mock implementation of Executor for testing.
// Uses function fields for configurable behavior.
type MockExecutor struct {
    NameFunc              func() string
    SpawnFunc             func(ctx context.Context, agent *Agent, prompt string, sessionID string) error
    ResumeFunc            func(ctx context.Context, sessionID string, prompt string) error
    SupportsResumptionFunc func() bool
}
```

### 4. Create MockCommandRunner for Testing

```go
// MockCommandRunner is a mock implementation of CommandRunner for testing.
type MockCommandRunner struct {
    RunFunc func(ctx context.Context, name string, args []string, stdin io.Reader) error
    // Captures the last call for verification
    LastName  string
    LastArgs  []string
    LastStdin string
}
```

## Acceptance Criteria

1. **Executor interface is defined** in `cli/internal/agents/executor.go`
   - Has Name(), Spawn(), Resume(), SupportsResumption() methods
   - Proper godoc comments explaining each method
   - Context parameter for cancellation support

2. **CommandRunner interface is defined** for subprocess abstraction
   - Single Run() method with context, command name, args, and stdin
   - DefaultCommandRunner implementation using os/exec
   - Connects stdout/stderr to os.Stdout/os.Stderr for interactive CLI use

3. **MockExecutor is provided** in `cli/internal/agents/executor_mock.go`
   - Function fields for each interface method
   - Default implementations return sensible values
   - Compile-time interface check `var _ Executor = (*MockExecutor)(nil)`
   - Follows pattern from `cli/internal/exec/mock.go`

4. **MockCommandRunner is provided** for testing executor implementations
   - Captures call parameters for verification in tests
   - Function field for configurable behavior

5. **Unit tests verify**:
   - MockExecutor implements Executor interface
   - MockCommandRunner implements CommandRunner interface
   - DefaultCommandRunner struct is usable (compile-time check)
   - Mock function fields are called when set
   - Default mock behavior works when function fields are nil

## Technical Details

### Package Location

All code goes in `cli/internal/agents/` alongside existing agent code:
- `executor.go` - Executor interface, CommandRunner interface, DefaultCommandRunner
- `executor_mock.go` - MockExecutor, MockCommandRunner
- `executor_test.go` - Tests for interface compliance and mock behavior

### Import Dependencies

```go
import (
    "context"
    "io"
    "os"
    "os/exec"
)
```

### Code Style

Follow existing patterns in the codebase:
- Comprehensive godoc comments (see `cli/internal/agents/agents.go` for style)
- Compile-time interface checks (see `cli/internal/exec/mock.go`)
- Table-driven tests with descriptive test names
- Error messages: lowercase, include relevant context

### Interface Design Rationale

**Why methods block directly (no Session object returned):**
- CLI is stateless - each invocation is independent
- Session state lives in task `state.yaml`, not Go memory
- Simpler model: spawn -> block -> exit -> read state

**Why stdin not stdout/stderr for prompts:**
- Prompts are sent via stdin
- stdout/stderr go to terminal for interactive CLI experience
- Subprocess communicates results via state.yaml file, not stdout

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/agents.go` - Agent struct definition that Spawn receives
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/templates.go` - LoadPrompt function used by executors
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/exec/mock.go` - Mock pattern to follow (function fields with defaults)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/exec/executor.go` - Similar interface pattern (different purpose - general command execution)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/.sow/knowledge/designs/multi-agent-architecture.md` - Design document specifying the Executor interface (lines 294-330)

## Examples

### Using the Interface (Production)

```go
func spawnWorker(ctx context.Context, executor Executor, agent *Agent, taskID string) error {
    sessionID := uuid.New().String()
    prompt := fmt.Sprintf("Execute task %s", taskID)
    return executor.Spawn(ctx, agent, prompt, sessionID)
}

func resumeWorker(ctx context.Context, executor Executor, sessionID string, feedback string) error {
    if !executor.SupportsResumption() {
        return fmt.Errorf("executor %s does not support resumption", executor.Name())
    }
    return executor.Resume(ctx, sessionID, feedback)
}
```

### Using MockExecutor in Tests

```go
func TestSpawnWorker(t *testing.T) {
    var capturedAgent *Agent
    var capturedPrompt string

    mock := &MockExecutor{
        NameFunc: func() string { return "test-executor" },
        SpawnFunc: func(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
            capturedAgent = agent
            capturedPrompt = prompt
            return nil
        },
        SupportsResumptionFunc: func() bool { return true },
    }

    err := spawnWorker(context.Background(), mock, Implementer, "010")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if capturedAgent.Name != "implementer" {
        t.Errorf("wrong agent: got %s, want implementer", capturedAgent.Name)
    }
}
```

### Using MockCommandRunner in Tests

```go
func TestClaudeExecutorSpawn(t *testing.T) {
    runner := &MockCommandRunner{
        RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
            // Verify command and args
            if name != "claude" {
                t.Errorf("wrong command: got %s, want claude", name)
            }
            return nil
        },
    }

    executor := NewClaudeExecutorWithRunner(false, "", runner)
    err := executor.Spawn(context.Background(), Implementer, "test prompt", "session-123")
    // Assert...
}
```

## Dependencies

- No dependencies on other tasks - this is the foundational task
- Uses only standard library imports plus existing `agents` package types

## Constraints

- Do NOT import the existing `exec.Executor` interface - this is a different abstraction for agent CLIs
- Do NOT actually invoke CLIs in tests - use the mock implementations
- Keep interface minimal - only what's needed for spawn/resume pattern
- Interface must support context cancellation for proper timeout handling
