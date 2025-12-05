# Claude Executor Implementation

## Context

This task is part of building the Executor System for sow - a multi-agent orchestration framework. ClaudeExecutor invokes the Claude Code CLI (`claude`) to spawn and resume agent sessions.

**Why this is needed:**
- Claude Code is the primary AI CLI tool for sow
- The executor must spawn Claude with the correct flags for session management
- Support for yoloMode (`--dangerously-skip-permissions`) and model selection
- Session resumption enables bidirectional communication with workers

**Design reference:** The multi-agent architecture design document specifies ClaudeExecutor with specific CLI flags and behavior.

## Requirements

### 1. Create ClaudeExecutor

Create `cli/internal/agents/executor_claude.go`:

```go
// ClaudeExecutor implements Executor for the Claude Code CLI.
// It spawns claude processes with appropriate flags for agent execution
// and supports session resumption for bidirectional communication.
type ClaudeExecutor struct {
    yoloMode bool   // When true, adds --dangerously-skip-permissions flag
    model    string // Model to use (e.g., "sonnet", "opus"), empty for default
    runner   CommandRunner
}

// NewClaudeExecutor creates a ClaudeExecutor with the given configuration.
// yoloMode: if true, skips permission prompts (--dangerously-skip-permissions)
// model: model name to use (empty string uses default)
func NewClaudeExecutor(yoloMode bool, model string) *ClaudeExecutor

// NewClaudeExecutorWithRunner creates a ClaudeExecutor with a custom CommandRunner.
// This is primarily for testing to inject mock command execution.
func NewClaudeExecutorWithRunner(yoloMode bool, model string, runner CommandRunner) *ClaudeExecutor
```

### 2. Implement Executor Interface

#### Name()
```go
func (e *ClaudeExecutor) Name() string {
    return "claude-code"
}
```

#### Spawn()
The Spawn method must:
1. Load the agent's prompt template using `LoadPrompt(agent.PromptPath)`
2. Combine agent prompt with task-specific prompt
3. Build CLI arguments:
   - `--dangerously-skip-permissions` if yoloMode is true
   - `--model <model>` if model is not empty
   - `--session-id <sessionID>` if sessionID is not empty
4. Execute `claude <args>` with combined prompt on stdin
5. Block until subprocess exits
6. Return error from subprocess (if any)

```go
func (e *ClaudeExecutor) Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
    // Load agent prompt template
    agentPrompt, err := LoadPrompt(agent.PromptPath)
    if err != nil {
        return fmt.Errorf("failed to load agent prompt: %w", err)
    }

    // Combine prompts
    fullPrompt := agentPrompt + "\n\n" + prompt

    // Build args
    args := []string{}
    if e.yoloMode {
        args = append(args, "--dangerously-skip-permissions")
    }
    if e.model != "" {
        args = append(args, "--model", e.model)
    }
    if sessionID != "" {
        args = append(args, "--session-id", sessionID)
    }

    // Execute
    return e.runner.Run(ctx, "claude", args, strings.NewReader(fullPrompt))
}
```

#### Resume()
The Resume method must:
1. Build CLI arguments: `--resume <sessionID>`
2. Execute `claude --resume <sessionID>` with prompt on stdin
3. Block until subprocess exits
4. Return error from subprocess (if any)

```go
func (e *ClaudeExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
    args := []string{"--resume", sessionID}
    return e.runner.Run(ctx, "claude", args, strings.NewReader(prompt))
}
```

#### SupportsResumption()
```go
func (e *ClaudeExecutor) SupportsResumption() bool {
    return true // Claude Code supports --resume
}
```

### 3. Default CommandRunner

When no runner is provided via `NewClaudeExecutorWithRunner`, use `DefaultCommandRunner`:

```go
func NewClaudeExecutor(yoloMode bool, model string) *ClaudeExecutor {
    return &ClaudeExecutor{
        yoloMode: yoloMode,
        model:    model,
        runner:   &DefaultCommandRunner{},
    }
}
```

## Acceptance Criteria

1. **ClaudeExecutor struct** with yoloMode, model, and runner fields

2. **Name()** returns `"claude-code"`

3. **Spawn()** correctly builds CLI command:
   - Loads agent prompt via LoadPrompt
   - Concatenates agent prompt + task prompt
   - Adds `--dangerously-skip-permissions` when yoloMode is true
   - Adds `--model <model>` when model is not empty
   - Adds `--session-id <sessionID>` when sessionID is not empty
   - Passes combined prompt via stdin
   - Returns LoadPrompt errors wrapped with context
   - Returns runner errors as-is

4. **Resume()** correctly builds CLI command:
   - Uses `--resume <sessionID>` flag
   - Passes prompt via stdin

5. **SupportsResumption()** returns true

6. **Constructor functions**:
   - `NewClaudeExecutor` uses DefaultCommandRunner
   - `NewClaudeExecutorWithRunner` accepts custom runner

7. **Interface compliance**:
   - Compile-time check: `var _ Executor = (*ClaudeExecutor)(nil)`

8. **Unit tests verify**:
   - Name() returns correct value
   - Spawn() builds correct args for various configurations:
     - yoloMode=false, model="", sessionID="" (minimal)
     - yoloMode=true (adds --dangerously-skip-permissions)
     - model="sonnet" (adds --model sonnet)
     - sessionID="abc-123" (adds --session-id abc-123)
     - All flags combined
   - Spawn() loads agent prompt and combines with task prompt
   - Spawn() returns error if LoadPrompt fails
   - Resume() builds correct args (--resume <sessionID>)
   - Resume() passes prompt via stdin
   - SupportsResumption() returns true

## Technical Details

### Package Location

File: `cli/internal/agents/executor_claude.go`
Test: `cli/internal/agents/executor_claude_test.go`

### Import Dependencies

```go
import (
    "context"
    "fmt"
    "strings"
)
```

### CLI Flag Reference

Claude Code CLI flags used:
- `--dangerously-skip-permissions` - Skip permission prompts (yolo mode)
- `--model <name>` - Model to use (e.g., "sonnet", "opus")
- `--session-id <uuid>` - Start new session with specific ID
- `--resume <uuid>` - Resume existing session

### Code Style

- Comprehensive godoc comments
- Error messages include context (e.g., "failed to load agent prompt")
- Table-driven tests for flag combinations

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/executor.go` - Executor interface and CommandRunner (from task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/executor_mock.go` - MockCommandRunner for tests (from task 010)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/agents.go` - Agent struct with PromptPath field
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/internal/agents/templates.go` - LoadPrompt function
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/.sow/knowledge/designs/multi-agent-architecture.md` - Design document specifying ClaudeExecutor (lines 335-386)

## Examples

### Production Usage

```go
// Create executor with yolo mode and specific model
executor := NewClaudeExecutor(true, "sonnet")

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
func TestClaudeExecutor_Spawn_BuildsCorrectArgs(t *testing.T) {
    tests := []struct {
        name      string
        yoloMode  bool
        model     string
        sessionID string
        wantArgs  []string
    }{
        {
            name:      "minimal",
            yoloMode:  false,
            model:     "",
            sessionID: "",
            wantArgs:  []string{},
        },
        {
            name:      "yolo mode",
            yoloMode:  true,
            model:     "",
            sessionID: "",
            wantArgs:  []string{"--dangerously-skip-permissions"},
        },
        {
            name:      "with model",
            yoloMode:  false,
            model:     "sonnet",
            sessionID: "",
            wantArgs:  []string{"--model", "sonnet"},
        },
        {
            name:      "with session ID",
            yoloMode:  false,
            model:     "",
            sessionID: "abc-123",
            wantArgs:  []string{"--session-id", "abc-123"},
        },
        {
            name:      "all flags",
            yoloMode:  true,
            model:     "opus",
            sessionID: "xyz-789",
            wantArgs:  []string{"--dangerously-skip-permissions", "--model", "opus", "--session-id", "xyz-789"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var capturedArgs []string
            runner := &MockCommandRunner{
                RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
                    capturedArgs = args
                    return nil
                },
            }

            executor := NewClaudeExecutorWithRunner(tt.yoloMode, tt.model, runner)
            err := executor.Spawn(context.Background(), Implementer, "test", tt.sessionID)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            if !reflect.DeepEqual(capturedArgs, tt.wantArgs) {
                t.Errorf("args = %v, want %v", capturedArgs, tt.wantArgs)
            }
        })
    }
}

func TestClaudeExecutor_Spawn_CombinesPrompts(t *testing.T) {
    var capturedStdin string
    runner := &MockCommandRunner{
        RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
            data, _ := io.ReadAll(stdin)
            capturedStdin = string(data)
            return nil
        },
    }

    executor := NewClaudeExecutorWithRunner(false, "", runner)
    err := executor.Spawn(context.Background(), Implementer, "Execute task 010", "")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Should contain both agent prompt and task prompt
    if !strings.Contains(capturedStdin, "Execute task 010") {
        t.Error("stdin should contain task prompt")
    }
    // Agent prompt is loaded from template, so just verify it's not empty
    if len(capturedStdin) < len("Execute task 010") + 10 {
        t.Error("stdin should contain agent prompt before task prompt")
    }
}

func TestClaudeExecutor_Resume_BuildsCorrectArgs(t *testing.T) {
    var capturedName string
    var capturedArgs []string
    runner := &MockCommandRunner{
        RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
            capturedName = name
            capturedArgs = args
            return nil
        },
    }

    executor := NewClaudeExecutorWithRunner(false, "", runner)
    err := executor.Resume(context.Background(), "session-123", "feedback")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if capturedName != "claude" {
        t.Errorf("command = %q, want %q", capturedName, "claude")
    }

    wantArgs := []string{"--resume", "session-123"}
    if !reflect.DeepEqual(capturedArgs, wantArgs) {
        t.Errorf("args = %v, want %v", capturedArgs, wantArgs)
    }
}
```

## Dependencies

- **Task 010** (Executor Interface): Requires `Executor` interface, `CommandRunner` interface, `MockCommandRunner`
- Existing agents package: Uses `Agent` struct and `LoadPrompt` function

## Constraints

- Do NOT actually invoke `claude` CLI in tests - use MockCommandRunner
- Do NOT add configuration for session flags (`--session-id`, `--resume`) - these are hardcoded
- yoloMode and model settings come from constructor, not per-call
- Session ID is per-call to support multiple concurrent sessions
