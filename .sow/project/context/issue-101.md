# Issue #101: Agent CLI Commands

**URL**: https://github.com/jmgilman/sow/issues/101
**State**: OPEN

## Description

# Work Unit 003: Agent CLI Commands

## Behavioral Goal

As a sow orchestrator (AI agent), I need CLI commands to spawn and resume worker agents, so that I can delegate tasks to specialized agents and communicate with them through the session management protocol.

## Scope

### In Scope
- `sow agent spawn <agent-name> <task-id>` command
- `sow agent resume <task-id> <prompt>` command
- `sow agent list` command
- Session ID generation and persistence to task state
- Integration with agent and executor registries
- Integration tests

### Out of Scope
- User configuration (work unit 004) - commands use default executor (Claude) initially
- Executor selection from config (added after work unit 004)

## Existing Code Context

**Command structure** (`cli/cmd/root.go:74-85`):
```go
cmd.AddCommand(NewInitCmd())
cmd.AddCommand(project.NewProjectCmd())
// ...
```
New `agent` command follows same pattern.

**Subcommand package pattern** (`cli/cmd/refs/refs.go`):
```go
func NewRefsCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "refs", ...}
    cmd.AddCommand(newAddCmd())
    cmd.AddCommand(newListCmd())
    // ...
    return cmd
}
```

**Task state loading** (`cli/cmd/task.go:197-203`):
```go
proj, err := state.Load(ctx)
// Find task by ID in phase
```

**Context access** (`cli/internal/cmdutil/context.go`):
```go
ctx := cmdutil.GetContext(cmd.Context())
```

## Documentation Context

**Design doc** (`.sow/knowledge/designs/multi-agent-architecture.md`):
- Section "Spawn Agent" (lines 998-1055) shows command implementation
- Section "Resume Agent" (lines 1057-1105) shows resume implementation
- Section "Complete Usage Flow" (lines 1165-1274) shows end-to-end example

## File Structure

Create new command package:

```
cli/cmd/agent/
├── agent.go        # Parent command
├── spawn.go        # sow agent spawn
├── spawn_test.go
├── resume.go       # sow agent resume
├── resume_test.go
├── list.go         # sow agent list
└── list_test.go
```

Update root:
```
cli/cmd/root.go     # Add agent.NewAgentCmd()
```

## Implementation Approach

### Parent Command

```go
// cli/cmd/agent/agent.go
func NewAgentCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "agent",
        Short: "Manage AI agents",
        Long:  `Spawn, resume, and manage AI agent workers.`,
    }

    cmd.AddCommand(newSpawnCmd())
    cmd.AddCommand(newResumeCmd())
    cmd.AddCommand(newListCmd())

    return cmd
}
```

### Spawn Command

```go
// cli/cmd/agent/spawn.go
func newSpawnCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "spawn <agent-name> <task-id>",
        Short: "Spawn an agent to execute a task",
        Args:  cobra.ExactArgs(2),
        RunE:  runSpawn,
    }
    return cmd
}

func runSpawn(cmd *cobra.Command, args []string) error {
    agentName := args[0]
    taskID := args[1]

    ctx := cmdutil.GetContext(cmd.Context())

    // 1. Get agent from registry
    agent, err := agents.DefaultRegistry().Get(agentName)

    // 2. Get executor (default: Claude for now, config integration later)
    executor := agents.NewClaudeExecutor(false, "")

    // 3. Load task state
    proj, err := state.Load(ctx)
    task := findTask(proj, taskID)

    // 4. Generate session ID if not present
    if task.SessionID == "" {
        task.SessionID = uuid.New().String()
        saveTask(proj, task)  // Persist before spawn!
    }

    // 5. Build task prompt
    prompt := buildTaskPrompt(ctx, task)

    // 6. Spawn agent (blocks until exit)
    return executor.Spawn(cmd.Context(), agent, prompt, task.SessionID)
}
```

### Resume Command

```go
// cli/cmd/agent/resume.go
func newResumeCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "resume <task-id> <prompt>",
        Short: "Resume a paused agent session",
        Args:  cobra.ExactArgs(2),
        RunE:  runResume,
    }
    return cmd
}

func runResume(cmd *cobra.Command, args []string) error {
    taskID := args[0]
    prompt := args[1]

    ctx := cmdutil.GetContext(cmd.Context())

    // 1. Load task state
    proj, err := state.Load(ctx)
    task := findTask(proj, taskID)

    // 2. Verify session exists
    if task.SessionID == "" {
        return fmt.Errorf("no session found for task %s", taskID)
    }

    // 3. Get executor (matches original spawn)
    executor := agents.NewClaudeExecutor(false, "")

    // 4. Check resumption support
    if !executor.SupportsResumption() {
        return fmt.Errorf("executor does not support session resumption")
    }

    // 5. Resume (blocks until exit)
    return executor.Resume(cmd.Context(), task.SessionID, prompt)
}
```

### List Command

```go
// cli/cmd/agent/list.go
func newListCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List available agents",
        RunE:  runList,
    }
    return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
    agents := agents.DefaultRegistry().List()

    fmt.Println("Available agents:")
    for _, agent := range agents {
        fmt.Printf("  %-12s %s\n", agent.Name, agent.Description)
    }
    return nil
}
```

### Task Prompt Builder

```go
func buildTaskPrompt(ctx *sow.Context, task *state.Task) string {
    // Build prompt with task location
    return fmt.Sprintf(`Execute task %s.

Task location: .sow/project/phases/%s/tasks/%s/

Read state.yaml for task metadata, description.md for requirements,
and feedback/ for any corrections from previous iterations.
`, task.ID, task.Phase, task.ID)
}
```

## Dependencies

- **Work Unit 001** (Agent System Core): AgentRegistry, Agent struct
- **Work Unit 002** (Executor System): Executor interface, ClaudeExecutor

## Acceptance Criteria

1. **`sow agent spawn <agent> <task-id>`** works:
   - Looks up agent by name
   - Generates and persists session ID
   - Invokes executor.Spawn()
   - Blocks until subprocess exits
2. **`sow agent resume <task-id> <prompt>`** works:
   - Loads session ID from task state
   - Errors if no session exists
   - Invokes executor.Resume()
   - Blocks until subprocess exits
3. **`sow agent list`** displays all available agents with descriptions
4. **Session ID persisted** before subprocess spawn (critical for crash recovery)
5. **Integration tests** cover:
   - Spawn with valid agent and task
   - Spawn with unknown agent (error)
   - Resume with existing session
   - Resume without session (error)
   - List shows all standard agents
6. **Helpful error messages** for common failures

## Testing Strategy

- Mock executor to avoid actual CLI invocation
- Test session ID generation and persistence
- Test error handling (unknown agent, missing session, etc.)
- Integration tests with test project state
- Verify task state is saved before spawn (not after)
