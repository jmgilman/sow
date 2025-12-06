# Spawn Command Implementation

## Context

This task implements the `sow agent spawn <agent-name> <task-id>` command, which is the primary mechanism for the orchestrator to delegate work to specialized worker agents.

The multi-agent architecture design specifies:
- The orchestrator (an AI agent running in Claude Code) spawns workers via CLI
- Session IDs must be persisted BEFORE spawning (critical for crash recovery)
- The spawn command blocks until the subprocess exits
- Task state contains the session ID for later resumption

The spawn command flow:
1. Look up agent by name from registry
2. Find task by ID in project state
3. Generate session ID if not present, persist to task state
4. Build task prompt with location information
5. Invoke executor.Spawn() which blocks until subprocess exits

## Requirements

### Spawn Command (`cli/cmd/agent/spawn.go`)

1. **Command structure**:
   - Use: `"spawn <agent-name> <task-id>"`
   - Short: `"Spawn an agent to execute a task"`
   - Long: Detailed explanation of spawn behavior, session management, blocking
   - Args: `cobra.ExactArgs(2)` - requires agent name and task ID
   - Flags:
     - `--phase` (string): Optional phase override, defaults to smart resolution
   - RunE: `runSpawn`

2. **Implementation flow**:

   a. **Get sow context**: Use `cmdutil.GetContext(cmd.Context())`

   b. **Check initialization**: Verify sow is initialized with `ctx.IsInitialized()`

   c. **Load project state**: Use `state.Load(ctx)` to load project

   d. **Look up agent**:
      - Create registry: `agents.NewAgentRegistry()`
      - Get agent: `registry.Get(agentName)`
      - Return helpful error if agent not found

   e. **Find task by ID**:
      - Resolve phase using existing `resolveTaskPhase()` helper from `helpers.go`
      - Search for task in phase by ID
      - Return helpful error if task not found

   f. **Handle session ID**:
      - Check if `task.Session_id` is empty
      - If empty, generate new UUID: `uuid.New().String()`
      - Update task state with session ID
      - **CRITICAL**: Save project state BEFORE spawning (crash recovery)

   g. **Build task prompt**:
      - Create a helper function `buildTaskPrompt(taskID, phaseName string) string`
      - Prompt format:
        ```
        Execute task {taskID}.

        Task location: .sow/project/phases/{phase}/tasks/{taskID}/

        Read state.yaml for task metadata, description.md for requirements,
        and feedback/ for any corrections from previous iterations.
        ```

   h. **Create executor and spawn**:
      - Use `agents.NewClaudeExecutor(false, "")` for default executor
      - Call `executor.Spawn(cmd.Context(), agent, prompt, task.Session_id)`
      - This blocks until subprocess exits

   i. **Return**: Return any error from spawn, or nil on success

3. **Error handling**:
   - Unknown agent: `"unknown agent: %s (available: implementer, architect, reviewer, planner, researcher, decomposer)"`
   - Task not found: `"task %s not found in phase %s"`
   - Not initialized: `"sow not initialized. Run 'sow init' first"`
   - No project: `"no active project found"`

### Helper Function

Add `buildTaskPrompt` function (can be in `spawn.go` or `helpers.go`):

```go
func buildTaskPrompt(taskID, phaseName string) string {
    return fmt.Sprintf(`Execute task %s.

Task location: .sow/project/phases/%s/tasks/%s/

Read state.yaml for task metadata, description.md for requirements,
and feedback/ for any corrections from previous iterations.
`, taskID, phaseName, taskID)
}
```

## Acceptance Criteria

### Functional Requirements
- `sow agent spawn implementer 010` spawns an implementer agent for task 010
- Session ID is generated and persisted to task state
- Session ID is persisted BEFORE subprocess spawn (verify via state file timing)
- Command blocks until subprocess exits
- Unknown agent returns helpful error listing available agents
- Missing task returns error with task ID and phase
- Works with `--phase` flag to override phase resolution

### Test Requirements (TDD approach)

Create `cli/cmd/agent/spawn_test.go`:

1. **TestNewSpawnCmd_Structure**: Verify command Use, Short, Long, Args
2. **TestNewSpawnCmd_HasPhaseFlag**: Verify --phase flag exists
3. **TestRunSpawn_UnknownAgent**: Test error for non-existent agent
4. **TestRunSpawn_TaskNotFound**: Test error for non-existent task
5. **TestRunSpawn_GeneratesSessionID**: Verify session ID generated when empty
6. **TestRunSpawn_PreservesExistingSessionID**: Verify existing session ID not overwritten
7. **TestRunSpawn_PersistsSessionBeforeSpawn**: Verify state saved before executor.Spawn called
8. **TestRunSpawn_CallsExecutorSpawn**: Verify executor.Spawn called with correct args
9. **TestRunSpawn_BuildsCorrectPrompt**: Verify prompt includes task location
10. **TestBuildTaskPrompt_Format**: Test prompt builder function

### Test Infrastructure

Tests need mock infrastructure to avoid spawning real subprocesses:

```go
// Use MockExecutor from cli/internal/agents/executor_mock.go
type testDeps struct {
    registry *agents.AgentRegistry
    executor *agents.MockExecutor
    project  *state.Project
}

func setupTestDeps(t *testing.T) *testDeps {
    // Create temp directory with test project state
    // Return mock executor to capture spawn calls
}
```

For testing session persistence timing, use a mock executor that records call order:

```go
var callOrder []string
mock := &agents.MockExecutor{
    SpawnFunc: func(...) error {
        callOrder = append(callOrder, "spawn")
        return nil
    },
}
// After runSpawn, verify "save" appears before "spawn" in callOrder
```

## Technical Details

### Package Imports

```go
import (
    "fmt"
    "strings"

    "github.com/google/uuid"
    "github.com/jmgilman/sow/cli/internal/agents"
    "github.com/jmgilman/sow/cli/internal/cmdutil"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/spf13/cobra"
)
```

### File Structure
```
cli/cmd/agent/
    spawn.go       # newSpawnCmd(), runSpawn(), buildTaskPrompt()
    spawn_test.go  # Tests for spawn command
```

### Task State Update

The task state has `Session_id` field (note the CUE-generated snake_case):

```go
// From cli/schemas/project/cue_types_gen.go
type TaskState struct {
    // ...
    Session_id string `json:"session_id,omitempty"`
    // ...
}
```

Update the task and save:
```go
// Find task index and update
phaseState.Tasks[taskIndex].Session_id = uuid.New().String()
proj.Phases[phaseName] = phaseState

// Save before spawning (critical!)
if err := proj.Save(); err != nil {
    return fmt.Errorf("failed to save session ID: %w", err)
}
```

### Dependency Injection for Testing

To test without spawning real processes, the implementation should allow injecting a mock executor. Options:

1. **Package-level variable** (simple):
   ```go
   var newExecutor = func() agents.Executor {
       return agents.NewClaudeExecutor(false, "")
   }

   // In tests:
   agent.newExecutor = func() agents.Executor { return mockExecutor }
   ```

2. **Function parameter** (cleaner but requires refactoring):
   ```go
   func runSpawnWithExecutor(cmd *cobra.Command, args []string, exec agents.Executor) error
   ```

Choose option 1 for consistency with existing code patterns.

## Relevant Inputs

- `cli/cmd/agent/agent.go` - Parent command (from task 010)
- `cli/cmd/task.go` - Pattern for loading project, finding tasks by ID
- `cli/cmd/helpers.go` - resolveTaskPhase() helper function
- `cli/internal/agents/executor.go` - Executor interface
- `cli/internal/agents/executor_claude.go` - ClaudeExecutor implementation
- `cli/internal/agents/executor_mock.go` - MockExecutor for testing
- `cli/internal/agents/registry.go` - AgentRegistry
- `cli/internal/sdks/project/state/loader.go` - state.Load() function
- `cli/internal/sdks/project/state/project.go` - Project wrapper type
- `cli/schemas/project/cue_types_gen.go` - TaskState with Session_id field
- `cli/internal/cmdutil/context.go` - GetContext() and RequireInitialized()
- `.sow/knowledge/designs/multi-agent-architecture.md` - Design specification (lines 998-1055)
- `.sow/project/context/issue-101.md` - Issue requirements

## Examples

### Command usage:
```bash
# Spawn implementer for task 010 (default phase)
sow agent spawn implementer 010

# Spawn with explicit phase
sow agent spawn implementer 010 --phase implementation
```

### Error messages:
```
Error: unknown agent: badagent
Available agents: architect, decomposer, implementer, planner, researcher, reviewer

Error: task 999 not found in phase implementation

Error: no active project found
```

### Task state after spawn:
```yaml
tasks:
  - id: "010"
    name: "Implement feature"
    session_id: "550e8400-e29b-41d4-a716-446655440000"  # Generated
    ...
```

## Dependencies

- Task 010 (Agent Command Package) must be completed first
- Requires existing `cli/internal/agents/` package (ClaudeExecutor)
- Requires existing `cli/internal/sdks/project/state/` package (state.Load)
- Requires `github.com/google/uuid` package (already in go.mod)

## Constraints

- Do NOT implement user configuration selection - use ClaudeExecutor directly
- Session ID MUST be persisted before subprocess spawn (critical for crash recovery)
- Use existing `resolveTaskPhase()` helper, don't duplicate logic
- Follow existing error message patterns from task.go
- Tests must mock the executor to avoid spawning real subprocesses
- The spawn call must block (synchronous) - do not run in background
