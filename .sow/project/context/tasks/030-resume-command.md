# Resume Command Implementation

## Context

This task implements the `sow agent resume <task-id> <prompt>` command, which allows the orchestrator to continue a paused agent session with additional instructions or feedback.

The multi-agent architecture design specifies:
- Resume uses the session ID stored in task state from a previous spawn
- The resume command is used for iterative refinement (review feedback, corrections)
- Not all executors support resumption - must check before attempting
- The resume command blocks until the subprocess exits

The resume command enables bidirectional orchestrator-worker communication:
1. Orchestrator spawns worker -> Worker executes -> Worker pauses for feedback
2. Orchestrator reviews work -> Provides feedback via resume -> Worker continues
3. Cycle repeats until task complete

## Requirements

### Resume Command (`cli/cmd/agent/resume.go`)

1. **Command structure**:
   - Use: `"resume <task-id> <prompt>"`
   - Short: `"Resume a paused agent session with feedback"`
   - Long: Detailed explanation of resume behavior, session requirements, use cases
   - Args: `cobra.ExactArgs(2)` - requires task ID and prompt
   - Flags:
     - `--phase` (string): Optional phase override, defaults to smart resolution
   - RunE: `runResume`

2. **Implementation flow**:

   a. **Get sow context**: Use `cmdutil.GetContext(cmd.Context())`

   b. **Check initialization**: Verify sow is initialized with `ctx.IsInitialized()`

   c. **Load project state**: Use `state.Load(ctx)` to load project

   d. **Find task by ID**:
      - Resolve phase using existing `resolveTaskPhase()` helper
      - Search for task in phase by ID
      - Return helpful error if task not found

   e. **Verify session exists**:
      - Check if `task.Session_id` is non-empty
      - Return error if empty: `"no session found for task %s (spawn first with 'sow agent spawn')"`

   f. **Create executor and check resumption support**:
      - Use `agents.NewClaudeExecutor(false, "")` for default executor
      - Call `executor.SupportsResumption()`
      - Return error if not supported: `"executor does not support session resumption"`

   g. **Resume session**:
      - Call `executor.Resume(cmd.Context(), task.Session_id, prompt)`
      - This blocks until subprocess exits

   h. **Return**: Return any error from resume, or nil on success

3. **Error handling**:
   - Task not found: `"task %s not found in phase %s"`
   - No session: `"no session found for task %s (spawn first with 'sow agent spawn')"`
   - No resumption support: `"executor does not support session resumption"`
   - Not initialized: `"sow not initialized. Run 'sow init' first"`
   - No project: `"no active project found"`

### Register in Parent Command

Update `cli/cmd/agent/agent.go` to register the resume command:

```go
func NewAgentCmd() *cobra.Command {
    cmd := &cobra.Command{...}

    cmd.AddCommand(newListCmd())
    cmd.AddCommand(newSpawnCmd())
    cmd.AddCommand(newResumeCmd())  // Add this

    return cmd
}
```

## Acceptance Criteria

### Functional Requirements
- `sow agent resume 010 "Use RS256 algorithm"` resumes task 010's session with the prompt
- Returns error if task has no session ID (never spawned)
- Returns error if executor doesn't support resumption
- Returns error if task not found
- Command blocks until subprocess exits
- Works with `--phase` flag to override phase resolution

### Test Requirements (TDD approach)

Create `cli/cmd/agent/resume_test.go`:

1. **TestNewResumeCmd_Structure**: Verify command Use, Short, Long, Args
2. **TestNewResumeCmd_HasPhaseFlag**: Verify --phase flag exists
3. **TestRunResume_TaskNotFound**: Test error for non-existent task
4. **TestRunResume_NoSessionID**: Test error when task.Session_id is empty
5. **TestRunResume_ExecutorNoResumption**: Test error when executor doesn't support resume
6. **TestRunResume_CallsExecutorResume**: Verify executor.Resume called with session ID and prompt
7. **TestRunResume_PassesCorrectPrompt**: Verify user's prompt passed to executor

### Test Infrastructure

Reuse test infrastructure patterns from spawn_test.go:

```go
func TestRunResume_NoSessionID(t *testing.T) {
    deps := setupTestDeps(t)
    // Create task with empty Session_id
    deps.addTask("010", "") // empty session

    err := runResumeWithDeps(deps, []string{"010", "feedback"})

    if err == nil {
        t.Fatal("expected error for missing session")
    }
    if !strings.Contains(err.Error(), "no session found") {
        t.Errorf("expected 'no session found' error, got: %v", err)
    }
}

func TestRunResume_ExecutorNoResumption(t *testing.T) {
    deps := setupTestDeps(t)
    deps.addTask("010", "session-123")
    deps.executor.SupportsResumptionFunc = func() bool { return false }

    err := runResumeWithDeps(deps, []string{"010", "feedback"})

    if err == nil {
        t.Fatal("expected error for no resumption support")
    }
}
```

## Technical Details

### Package Imports

```go
import (
    "fmt"

    "github.com/jmgilman/sow/cli/internal/agents"
    "github.com/jmgilman/sow/cli/internal/cmdutil"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/spf13/cobra"
)
```

### File Structure
```
cli/cmd/agent/
    resume.go       # newResumeCmd(), runResume()
    resume_test.go  # Tests for resume command
```

### Executor Resume Interface

From `cli/internal/agents/executor.go`:
```go
type Executor interface {
    // ...
    Resume(ctx context.Context, sessionID string, prompt string) error
    SupportsResumption() bool
}
```

From `cli/internal/agents/executor_claude.go`:
```go
func (e *ClaudeExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
    args := []string{"--resume", sessionID}
    if err := e.runner.Run(ctx, "claude", args, strings.NewReader(prompt)); err != nil {
        return fmt.Errorf("claude resume failed: %w", err)
    }
    return nil
}

func (e *ClaudeExecutor) SupportsResumption() bool {
    return true
}
```

### Dependency Injection

Use the same pattern as spawn.go for executor injection:

```go
var newExecutor = func() agents.Executor {
    return agents.NewClaudeExecutor(false, "")
}
```

Or share the variable with spawn.go if in same package.

## Relevant Inputs

- `cli/cmd/agent/agent.go` - Parent command to update
- `cli/cmd/agent/spawn.go` - Similar command structure to follow (from task 020)
- `cli/cmd/task.go` - Pattern for loading project, finding tasks
- `cli/cmd/helpers.go` - resolveTaskPhase() helper
- `cli/internal/agents/executor.go` - Executor interface (Resume, SupportsResumption)
- `cli/internal/agents/executor_claude.go` - ClaudeExecutor.Resume implementation
- `cli/internal/agents/executor_mock.go` - MockExecutor for testing
- `cli/internal/sdks/project/state/loader.go` - state.Load() function
- `cli/schemas/project/cue_types_gen.go` - TaskState.Session_id field
- `.sow/knowledge/designs/multi-agent-architecture.md` - Design spec (lines 1057-1105)
- `.sow/project/context/issue-101.md` - Issue requirements

## Examples

### Command usage:
```bash
# Resume task 010 with feedback
sow agent resume 010 "Use RS256 algorithm for JWT signing"

# Resume with explicit phase
sow agent resume 010 "Add error handling for edge cases" --phase implementation

# Multi-word prompts work naturally
sow agent resume 010 "The tests are failing. Please check the mock setup and verify the assertions."
```

### Error messages:
```
Error: no session found for task 010 (spawn first with 'sow agent spawn')

Error: task 999 not found in phase implementation

Error: executor does not support session resumption
```

### Typical workflow:
```bash
# 1. Orchestrator spawns worker
sow agent spawn implementer 010

# 2. Worker executes, pauses for feedback (sets task status to "paused")
# 3. Orchestrator reviews, provides feedback
sow agent resume 010 "Good progress! Please add input validation for the email field"

# 4. Worker continues with feedback
# 5. Cycle repeats until task complete
```

## Dependencies

- Task 010 (Agent Command Package) must be completed first
- Task 020 (Spawn Command) should be completed first (shared patterns)
- Requires existing `cli/internal/agents/` package
- Requires existing `cli/internal/sdks/project/state/` package

## Constraints

- Do NOT implement user configuration selection - use ClaudeExecutor directly
- Do NOT modify the session ID during resume (only read it)
- Must check `SupportsResumption()` before calling `Resume()`
- Tests must mock the executor to avoid spawning real subprocesses
- The resume call must block (synchronous) - do not run in background
- Error message for missing session must guide user to spawn first
