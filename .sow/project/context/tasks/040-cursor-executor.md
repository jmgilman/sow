# Cursor Executor Implementation

## Context

This task is part of building the Executor System for sow - a multi-agent orchestration framework. CursorExecutor invokes the Cursor Agent CLI (`cursor-agent`) to spawn and resume agent sessions.

**Why this is needed:**
- Cursor is an alternative AI CLI tool that users may prefer
- Enables user configuration to select Cursor for specific agent roles
- Follows same session management pattern as ClaudeExecutor
- Different CLI flags than Claude Code require separate implementation

**Design reference:** The multi-agent architecture design document specifies CursorExecutor with specific CLI flags and behavior.

## Requirements

### 1. Create CursorExecutor

Create `cli/internal/agents/executor_cursor.go`:

```go
// CursorExecutor implements Executor for the Cursor Agent CLI.
// It spawns cursor-agent processes with appropriate flags for agent execution
// and supports session resumption for bidirectional communication.
type CursorExecutor struct {
    yoloMode bool // When true, adds yolo mode flag (if cursor supports it)
    runner   CommandRunner
}

// NewCursorExecutor creates a CursorExecutor with the given configuration.
// yoloMode: if true, enables automatic mode (exact flag TBD based on cursor-agent CLI)
func NewCursorExecutor(yoloMode bool) *CursorExecutor

// NewCursorExecutorWithRunner creates a CursorExecutor with a custom CommandRunner.
// This is primarily for testing to inject mock command execution.
func NewCursorExecutorWithRunner(yoloMode bool, runner CommandRunner) *CursorExecutor
```

### 2. Implement Executor Interface

#### Name()
```go
func (e *CursorExecutor) Name() string {
    return "cursor"
}
```

#### Spawn()
The Spawn method must:
1. Load the agent's prompt template using `LoadPrompt(agent.PromptPath)`
2. Combine agent prompt with task-specific prompt
3. Build CLI arguments:
   - Base command: `cursor-agent agent`
   - `--chat-id <sessionID>` if sessionID is not empty
4. Execute `cursor-agent agent <args>` with combined prompt on stdin
5. Block until subprocess exits
6. Return error from subprocess (if any)

```go
func (e *CursorExecutor) Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
    // Load agent prompt template
    agentPrompt, err := LoadPrompt(agent.PromptPath)
    if err != nil {
        return fmt.Errorf("failed to load agent prompt: %w", err)
    }

    // Combine prompts
    fullPrompt := agentPrompt + "\n\n" + prompt

    // Build args - note cursor-agent uses subcommand "agent"
    args := []string{"agent"}
    if sessionID != "" {
        args = append(args, "--chat-id", sessionID)
    }

    // Execute
    return e.runner.Run(ctx, "cursor-agent", args, strings.NewReader(fullPrompt))
}
```

#### Resume()
The Resume method must:
1. Build CLI arguments: `agent --resume <sessionID>`
2. Execute `cursor-agent agent --resume <sessionID>` with prompt on stdin
3. Block until subprocess exits
4. Return error from subprocess (if any)

```go
func (e *CursorExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
    args := []string{"agent", "--resume", sessionID}
    return e.runner.Run(ctx, "cursor-agent", args, strings.NewReader(prompt))
}
```

#### SupportsResumption()
```go
func (e *CursorExecutor) SupportsResumption() bool {
    return true // Cursor supports --resume
}
```

### 3. Default CommandRunner

When no runner is provided via `NewCursorExecutorWithRunner`, use `DefaultCommandRunner`:

```go
func NewCursorExecutor(yoloMode bool) *CursorExecutor {
    return &CursorExecutor{
        yoloMode: yoloMode,
        runner:   &DefaultCommandRunner{},
    }
}
```

### 4. YoloMode Note

The `yoloMode` field is included for future compatibility but is not currently mapped to any CLI flag. When Cursor Agent CLI supports an equivalent permission-skip flag, it can be added. For now, the field is stored but not used in command construction.

## Acceptance Criteria

1. **CursorExecutor struct** with yoloMode and runner fields

2. **Name()** returns `"cursor"`

3. **Spawn()** correctly builds CLI command:
   - Loads agent prompt via LoadPrompt
   - Concatenates agent prompt + task prompt
   - Uses `cursor-agent` as command
   - Includes `agent` as first argument (subcommand)
   - Adds `--chat-id <sessionID>` when sessionID is not empty
   - Passes combined prompt via stdin
   - Returns LoadPrompt errors wrapped with context
   - Returns runner errors as-is

4. **Resume()** correctly builds CLI command:
   - Uses `cursor-agent agent --resume <sessionID>` format
   - Passes prompt via stdin

5. **SupportsResumption()** returns true

6. **Constructor functions**:
   - `NewCursorExecutor` uses DefaultCommandRunner
   - `NewCursorExecutorWithRunner` accepts custom runner

7. **Interface compliance**:
   - Compile-time check: `var _ Executor = (*CursorExecutor)(nil)`

8. **Unit tests verify**:
   - Name() returns correct value
   - Spawn() uses correct command (`cursor-agent`)
   - Spawn() includes `agent` subcommand as first arg
   - Spawn() builds correct args for various configurations:
     - sessionID="" (no --chat-id flag)
     - sessionID="abc-123" (adds --chat-id abc-123)
   - Spawn() loads agent prompt and combines with task prompt
   - Spawn() returns error if LoadPrompt fails
   - Resume() builds correct args (`agent --resume <sessionID>`)
   - Resume() passes prompt via stdin
   - SupportsResumption() returns true

## Technical Details

### Package Location

File: `cli/internal/agents/executor_cursor.go`
Test: `cli/internal/agents/executor_cursor_test.go`

### Import Dependencies

```go
import (
    "context"
    "fmt"
    "strings"
)
```

### CLI Flag Reference

Cursor Agent CLI flags used:
- `agent` - Subcommand for agent mode
- `--chat-id <uuid>` - Start new session with specific ID
- `--resume <uuid>` - Resume existing session

**Key difference from Claude:**
- Cursor uses `cursor-agent agent` (command + subcommand)
- Claude uses just `claude`
- Cursor uses `--chat-id` for new sessions
- Claude uses `--session-id` for new sessions

### Code Style

- Comprehensive godoc comments
- Error messages include context
- Table-driven tests
- Follow same structure as ClaudeExecutor for consistency

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/executor.go` - Executor interface and CommandRunner (from task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/executor_mock.go` - MockCommandRunner for tests (from task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/executor_claude.go` - ClaudeExecutor for pattern reference (from task 030)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/agents.go` - Agent struct with PromptPath field
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/templates.go` - LoadPrompt function
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/.sow/knowledge/designs/multi-agent-architecture.md` - Design document specifying CursorExecutor (lines 389-428)

## Examples

### Production Usage

```go
// Create executor with yolo mode (future flag support)
executor := NewCursorExecutor(true)

// Spawn new session
sessionID := uuid.New().String()
err := executor.Spawn(ctx, agents.Implementer, "Execute task 010", sessionID)
if err != nil {
    return fmt.Errorf("failed to spawn: %w", err)
}

// Resume session with feedback
err = executor.Resume(ctx, sessionID, "Fix the security issue on line 42")
if err != nil {
    return fmt.Errorf("failed to resume: %w", err)
}
```

### Test Examples

```go
func TestCursorExecutor_Spawn_BuildsCorrectArgs(t *testing.T) {
    tests := []struct {
        name      string
        sessionID string
        wantArgs  []string
    }{
        {
            name:      "no session ID",
            sessionID: "",
            wantArgs:  []string{"agent"},
        },
        {
            name:      "with session ID",
            sessionID: "abc-123",
            wantArgs:  []string{"agent", "--chat-id", "abc-123"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var capturedName string
            var capturedArgs []string
            runner := &MockCommandRunner{
                RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
                    capturedName = name
                    capturedArgs = args
                    return nil
                },
            }

            executor := NewCursorExecutorWithRunner(false, runner)
            err := executor.Spawn(context.Background(), Implementer, "test", tt.sessionID)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            if capturedName != "cursor-agent" {
                t.Errorf("command = %q, want %q", capturedName, "cursor-agent")
            }

            if !reflect.DeepEqual(capturedArgs, tt.wantArgs) {
                t.Errorf("args = %v, want %v", capturedArgs, tt.wantArgs)
            }
        })
    }
}

func TestCursorExecutor_Resume_BuildsCorrectArgs(t *testing.T) {
    var capturedName string
    var capturedArgs []string
    runner := &MockCommandRunner{
        RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
            capturedName = name
            capturedArgs = args
            return nil
        },
    }

    executor := NewCursorExecutorWithRunner(false, runner)
    err := executor.Resume(context.Background(), "session-123", "feedback")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if capturedName != "cursor-agent" {
        t.Errorf("command = %q, want %q", capturedName, "cursor-agent")
    }

    wantArgs := []string{"agent", "--resume", "session-123"}
    if !reflect.DeepEqual(capturedArgs, wantArgs) {
        t.Errorf("args = %v, want %v", capturedArgs, wantArgs)
    }
}

func TestCursorExecutor_Spawn_CombinesPrompts(t *testing.T) {
    var capturedStdin string
    runner := &MockCommandRunner{
        RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
            data, _ := io.ReadAll(stdin)
            capturedStdin = string(data)
            return nil
        },
    }

    executor := NewCursorExecutorWithRunner(false, runner)
    err := executor.Spawn(context.Background(), Implementer, "Execute task 010", "")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Should contain both agent prompt and task prompt
    if !strings.Contains(capturedStdin, "Execute task 010") {
        t.Error("stdin should contain task prompt")
    }
}
```

## Dependencies

- **Task 010** (Executor Interface): Requires `Executor` interface, `CommandRunner` interface, `MockCommandRunner`
- **Task 030** (ClaudeExecutor): Follow same patterns for consistency
- Existing agents package: Uses `Agent` struct and `LoadPrompt` function

## Constraints

- Do NOT actually invoke `cursor-agent` CLI in tests - use MockCommandRunner
- Do NOT add configuration for session flags (`--chat-id`, `--resume`) - these are hardcoded
- yoloMode is stored but not currently used (future compatibility)
- Session ID is per-call to support multiple concurrent sessions
- Command is `cursor-agent` (not `cursor`)
- Must include `agent` as first argument (subcommand)
