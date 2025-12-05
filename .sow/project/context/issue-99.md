# Issue #99: Executor System

**URL**: https://github.com/jmgilman/sow/issues/99
**State**: OPEN

## Description

# Work Unit 002: Executor System

## Behavioral Goal

As a sow orchestrator, I need an executor system that can spawn AI CLI tools (Claude Code, Cursor) with agent prompts and manage sessions, so that workers can be invoked programmatically and resumed for bidirectional communication.

## Scope

### In Scope
- Executor interface (Spawn, Resume, SupportsResumption)
- ExecutorRegistry for managing executor instances
- ClaudeExecutor implementation (invokes `claude` CLI)
- CursorExecutor implementation (invokes `cursor-agent` CLI)
- Task schema update: add `session_id` field
- Task schema update: add `paused` status
- Unit tests with mocked subprocess execution

### Out of Scope
- CLI commands (work unit 003)
- User configuration (work unit 004) - executors use hardcoded defaults initially
- Windsurf executor (can be added later following same pattern)

## Existing Code Context

**Existing executor pattern** (`cli/internal/exec/executor.go:24-45`):
```go
type Executor interface {
    Command() string
    Exists() bool
    Run(args ...string) (stdout, stderr string, err error)
    RunContext(ctx context.Context, args ...string) (stdout, stderr string, err error)
}
```
The new agent Executor interface is different - it's specific to AI CLI invocation with session management.

**Task schema** (`cli/schemas/project/task.cue:27`):
```cue
status: "pending" | "in_progress" | "needs_review" | "completed" | "abandoned"
```
Needs `"paused"` added for blocked workflows.

**Agent package** (from work unit 001):
- `*Agent` struct with Name, PromptPath
- `LoadPrompt(path)` function

## Documentation Context

**Design doc** (`.sow/knowledge/designs/multi-agent-architecture.md`):
- Section "Executor Interface" (lines 294-330) specifies the interface
- Section "Executor Implementations" (lines 332-428) shows Claude and Cursor implementations
- Section "Session Management Protocol" (lines 815-935) describes session lifecycle
- Section "State-Based Subprocess Protocol" (lines 937-972) explains the communication model

## File Structure

Add to `cli/internal/agents/`:

```
cli/internal/agents/
├── ... (from work unit 001)
├── executor.go           # Executor interface
├── executor_test.go      # Interface tests
├── executor_claude.go    # ClaudeExecutor
├── executor_claude_test.go
├── executor_cursor.go    # CursorExecutor
├── executor_cursor_test.go
├── executor_registry.go  # ExecutorRegistry
└── executor_registry_test.go
```

Update schema:
```
cli/schemas/project/task.cue  # Add session_id, paused status
```

## Implementation Approach

### Executor Interface

```go
// Executor invokes agent CLIs and manages sessions
type Executor interface {
    // Name returns the executor identifier (e.g., "claude-code", "cursor")
    Name() string

    // Spawn invokes an agent with the given prompt and session ID
    // Blocks until the subprocess exits
    Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error

    // Resume continues an existing session with additional prompt
    // Blocks until the subprocess exits
    Resume(ctx context.Context, sessionID string, prompt string) error

    // SupportsResumption indicates if this executor can resume sessions
    SupportsResumption() bool
}
```

### ClaudeExecutor

```go
type ClaudeExecutor struct {
    yoloMode bool   // --dangerously-skip-permissions
    model    string // --model flag
}

func NewClaudeExecutor(yoloMode bool, model string) *ClaudeExecutor

func (e *ClaudeExecutor) Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
    // Load agent prompt template
    agentPrompt, err := LoadPrompt(agent.PromptPath)
    // Build full prompt: agentPrompt + task prompt
    // Build args: [--dangerously-skip-permissions] [--model X] [--session-id Y]
    // Execute: claude <args> with prompt on stdin
    // Block until exit
}

func (e *ClaudeExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
    // Execute: claude --resume <sessionID> with prompt on stdin
}
```

### CursorExecutor

```go
type CursorExecutor struct {
    yoloMode bool
}

func NewCursorExecutor(yoloMode bool) *CursorExecutor

func (e *CursorExecutor) Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
    // Execute: cursor-agent agent [--chat-id Y] with prompt on stdin
}

func (e *CursorExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
    // Execute: cursor-agent agent --resume <sessionID> with prompt on stdin
}
```

### ExecutorRegistry

```go
type ExecutorRegistry struct {
    executors map[string]Executor
}

func NewExecutorRegistry() *ExecutorRegistry
func (r *ExecutorRegistry) Register(name string, executor Executor)
func (r *ExecutorRegistry) Get(name string) (Executor, error)
func (r *ExecutorRegistry) List() []string
```

### Schema Update

In `cli/schemas/project/task.cue`:

```cue
#TaskState: {
    // ... existing fields ...

    // Add paused to status enum
    status: "pending" | "in_progress" | "needs_review" | "completed" | "abandoned" | "paused"

    // Add session_id for resumable sessions
    // Set when worker spawned, cleared when task reaches terminal state
    session_id?: string
}
```

## Dependencies

- **Work Unit 001** (Agent System Core): Requires `*Agent` struct and `LoadPrompt()` function

## Acceptance Criteria

1. **Executor interface** defined with Spawn, Resume, SupportsResumption, Name
2. **ClaudeExecutor** implements interface:
   - Spawns `claude` with correct flags
   - Supports `--session-id` for new sessions
   - Supports `--resume` for continuing sessions
   - Handles yoloMode and model settings
3. **CursorExecutor** implements interface:
   - Spawns `cursor-agent agent` with correct flags
   - Supports `--chat-id` for new sessions
   - Supports `--resume` for continuing sessions
4. **ExecutorRegistry** supports Register, Get, List
5. **Task schema updated**:
   - `session_id` optional field added
   - `paused` status added to enum
6. **Unit tests** cover:
   - Executor interface compliance (both implementations)
   - Correct CLI flag construction
   - Registry operations
   - Error handling (unknown executor, resume not supported)
7. **Subprocess execution mockable** for testing (don't actually invoke CLIs in tests)

## Testing Strategy

- Mock subprocess execution using interface or function injection
- Test CLI argument construction without actually running CLIs
- Test registry operations
- Verify schema changes don't break existing task loading
- Table-driven tests for different executor configurations
