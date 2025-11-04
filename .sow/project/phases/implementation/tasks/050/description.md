# Task 050: Implement Task Commands (TDD)

# Task 050: Implement Task Commands (TDD)

## Overview

Implement task management commands and task-level input/output commands. Task state lives in the project state file (not separate files), but task directories are still created for `description.md`, `log.md`, and `feedback/`.

## Context

Tasks are stored in phase-level `tasks` collections, accessed by ID (not index). Task IDs are gap-numbered (010, 020, 030, ...) to allow insertion between tasks. The SDK provides collection methods: `Get(id)`, `Add(task)`, `Remove(id)`.

## Design References

- **Command spec**: `.sow/knowledge/designs/command-hierarchy-design.md` lines 388-446, 533-598, 740-875
- **Task schema**: `cli/schemas/project/task.cue`
- **SDK task type**: `cli/internal/sdks/project/state/task.go`
- **Task collection**: `cli/internal/sdks/project/state/collections.go` lines 50-83

## Requirements

### Commands to Implement

#### Task Management

```bash
sow task add <name> [--agent <agent>] [--description "..."]
sow task set --id <task-id> <field-path> <value>
sow task abandon --id <task-id>
sow task list
```

#### Task Input/Output

```bash
sow task input add --id <id> --type <type> [--path <path>] [...]
sow task input set --id <id> --index <n> <field-path> <value>
sow task input remove --id <id> --index <n>
sow task input list --id <id>

sow task output add --id <id> --type <type> [--path <path>] [...]
sow task output set --id <id> --index <n> <field-path> <value>
sow task output remove --id <id> --index <n>
sow task output list --id <id>
```

### Task Fields

**Identity**:
- `id` - Gap-numbered (010, 020, 030, ...)
- `name` - Task name
- `phase` - Phase name (usually "implementation")

**Status**:
- `status` - "pending" | "in_progress" | "completed" | "abandoned"
- `iteration` - Iteration number (starts at 1)
- `assigned_agent` - Agent type (e.g., "implementer")

**Timestamps**:
- `created_at`, `started_at`, `updated_at`, `completed_at`

**Collections**:
- `inputs` - Task input artifacts (references, feedback)
- `outputs` - Task output artifacts (modified files)

**Metadata**:
- `metadata.*` - Custom fields

### Task Input Types

- `reference` - Reference documents (style guides, design docs, code examples)
- `feedback` - Human feedback items

### Task Output Types

- `modified` - Files modified by the task

See `.sow/knowledge/designs/command-hierarchy-design.md` lines 927-946 for complete definitions.

### Task Directory Structure

On `task add`, create:

```
.sow/project/phases/implementation/tasks/{id}/
├── description.md   # Task description (from --description or placeholder)
├── log.md          # Task action log (empty initially)
└── feedback/       # Feedback directory (empty initially)
```

Note: No `state.yaml` in task directory - task state lives in project state.

## TDD Approach

### Step 1: Write Integration Tests First

Create three test files:

**`cli/testdata/script/unified_commands/tasks/task_operations.txtar`**:

```txtar
# Test: Task Operations
# Coverage: add, set, abandon, list

exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test
exec sow init
exec sow project new --branch feat/test --no-launch "Test tasks"

# Advance to implementation phase
exec sow advance  # PlanningActive -> ImplementationPlanning

# Test: Add task
exec sow task add "Implement feature" --agent implementer --description "Implement the feature"
exists .sow/project/phases/implementation/tasks/010/description.md
exists .sow/project/phases/implementation/tasks/010/log.md
exists .sow/project/phases/implementation/tasks/010/feedback
exec cat .sow/project/state.yaml
stdout 'id: "010"'
stdout 'name: Implement feature'
stdout 'status: pending'
stdout 'iteration: 1'
stdout 'assigned_agent: implementer'

# Test: Add second task (gap-numbered)
exec sow task add "Write tests" --agent implementer
exec cat .sow/project/state.yaml
stdout 'id: "020"'

# Test: List tasks
exec sow task list
stdout '\[010\] Implement feature \(pending\)'
stdout '\[020\] Write tests \(pending\)'

# Test: Set task status
exec sow task set --id 010 status in_progress
exec cat .sow/project/state.yaml
stdout 'status: in_progress'

# Test: Set task iteration
exec sow task set --id 010 iteration 2
exec cat .sow/project/state.yaml
stdout 'iteration: 2'

# Test: Set task metadata
exec sow task set --id 010 metadata.complexity high
exec cat .sow/project/state.yaml
stdout 'complexity: high'

# Test: Abandon task
exec sow task abandon --id 020
exec cat .sow/project/state.yaml
stdout 'status: abandoned'
```

**`cli/testdata/script/unified_commands/tasks/task_inputs_outputs.txtar`**:

```txtar
# Test: Task Input/Output Operations
# Coverage: add, set, remove, list for task artifacts

exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test
exec sow init
exec sow project new --branch feat/test --no-launch "Test task artifacts"
exec sow advance

# Add task
exec sow task add "Test task" --agent implementer

# Test: Add task input (reference)
exec mkdir -p .sow/sinks
exec sh -c 'echo "Style guide" > .sow/sinks/style-guide.md'
exec sow task input add --id 010 --type reference --path ../../../sinks/style-guide.md
exec cat .sow/project/state.yaml
stdout 'type: reference'
stdout 'path: ../../../sinks/style-guide.md'

# Test: Add task input (feedback)
exec mkdir -p .sow/project/phases/implementation/tasks/010/feedback
exec sh -c 'echo "Fix this" > .sow/project/phases/implementation/tasks/010/feedback/001.md'
exec sow task input add --id 010 --type feedback --path feedback/001.md
exec cat .sow/project/state.yaml
stdout 'type: feedback'

# Test: List task inputs
exec sow task input list --id 010
stdout '\[0\] reference: ../../../sinks/style-guide.md'
stdout '\[1\] feedback: feedback/001.md'

# Test: Set feedback metadata (status)
exec sow task input set --id 010 --index 1 metadata.status addressed
exec cat .sow/project/state.yaml
stdout 'status: addressed'

# Test: Add task output (modified file)
exec sow task output add --id 010 --type modified --path src/feature.go
exec cat .sow/project/state.yaml
stdout 'type: modified'
stdout 'path: src/feature.go'

# Test: List task outputs
exec sow task output list --id 010
stdout '\[0\] modified: src/feature.go'

# Test: Remove task input
exec sow task input remove --id 010 --index 0
exec sow task input list --id 010
! stdout 'reference'
stdout '\[0\] feedback: feedback/001.md'
```

**`cli/testdata/script/unified_commands/tasks/task_lifecycle.txtar`**:

Complete task lifecycle test.

### Step 2: Implement Commands

Create `cli/cmd/task.go`, `cli/cmd/task_input.go`, `cli/cmd/task_output.go`.

### Step 3: Run Integration Tests

Verify all tests pass.

## Implementation Details

### File Structure

```
cli/cmd/
├── task.go          # sow task add/set/abandon/list
├── task_input.go    # sow task input add/set/remove/list
└── task_output.go   # sow task output add/set/remove/list
```

### Task Add Implementation

```go
func runTaskAdd(cmd *cobra.Command, args []string, agent, description string) error {
    ctx := cmdutil.GetContext(cmd.Context())

    // Load project
    project, err := state.Load(ctx)
    if err != nil {
        return err
    }

    // Get implementation phase
    phase, err := project.Phases.Get("implementation")
    if err != nil {
        return err
    }

    // Generate task ID (gap-numbered)
    taskID := generateNextTaskID(phase.Tasks)

    // Create task
    task := state.Task{
        TaskState: schemas.TaskState{
            Id:            taskID,
            Name:          args[0],
            Phase:         "implementation",
            Status:        "pending",
            Iteration:     1,
            AssignedAgent: agent,
            CreatedAt:     time.Now(),
            UpdatedAt:     time.Now(),
            Inputs:        []schemas.ArtifactState{},
            Outputs:       []schemas.ArtifactState{},
            Metadata:      make(map[string]interface{}),
        },
    }

    // Add to phase
    if err := phase.Tasks.Add(task); err != nil {
        return err
    }

    // Create task directory
    if err := createTaskDirectory(ctx, taskID, description); err != nil {
        return err
    }

    // Save
    return project.Save()
}

func generateNextTaskID(tasks state.TaskCollection) string {
    if len(tasks) == 0 {
        return "010"
    }

    // Find highest ID
    maxID := 0
    for _, task := range tasks {
        id, _ := strconv.Atoi(task.Id)
        if id > maxID {
            maxID = id
        }
    }

    // Return next gap-numbered ID
    nextID := maxID + 10
    return fmt.Sprintf("%03d", nextID)
}

func createTaskDirectory(ctx *sow.Context, taskID, description string) error {
    taskDir := filepath.Join(ctx.RepoRoot(), ".sow/project/phases/implementation/tasks", taskID)

    // Create directories
    if err := os.MkdirAll(filepath.Join(taskDir, "feedback"), 0755); err != nil {
        return err
    }

    // Create description.md
    descPath := filepath.Join(taskDir, "description.md")
    descContent := description
    if descContent == "" {
        descContent = "# Task Description\n\nTODO: Add task description"
    }
    if err := os.WriteFile(descPath, []byte(descContent), 0644); err != nil {
        return err
    }

    // Create empty log.md
    logPath := filepath.Join(taskDir, "log.md")
    if err := os.WriteFile(logPath, []byte("# Task Log\n\n"), 0644); err != nil {
        return err
    }

    return nil
}
```

### Task Input/Output Implementation

These follow the same pattern as phase input/output (Task 040), but:
1. Navigate to task first: `phase.Tasks.Get(taskID)`
2. Access task inputs/outputs: `task.Inputs`, `task.Outputs`
3. Use index-based operations on artifact collections

### Task Set Implementation

Uses field path parser for:
- Direct fields: `status`, `iteration`, `assigned_agent`
- Metadata fields: `metadata.*`

### Task List Format

```
[010] Implement feature (pending)
[020] Write tests (in_progress)
[030] Add documentation (completed)
```

## Files to Create

### `cli/cmd/task.go`

Task management commands.

### `cli/cmd/task_input.go`

Task input operations.

### `cli/cmd/task_output.go`

Task output operations.

### Integration Test Files

- `cli/testdata/script/unified_commands/tasks/task_operations.txtar`
- `cli/testdata/script/unified_commands/tasks/task_inputs_outputs.txtar`
- `cli/testdata/script/unified_commands/tasks/task_lifecycle.txtar`

## Acceptance Criteria

- [ ] Integration tests written first
- [ ] Task add generates gap-numbered IDs
- [ ] Task add creates directory structure
- [ ] Task set modifies status, iteration, assigned_agent
- [ ] Task set modifies metadata via dot notation
- [ ] Task abandon sets status to abandoned
- [ ] Task list displays all tasks with IDs and status
- [ ] Task input/output commands work (add, set, remove, list)
- [ ] Task state lives in project state (not separate files)
- [ ] Integration tests pass

## Testing Strategy

**Integration tests only** - No unit tests for command logic.

Test scenarios:
1. Add task (verify directory created, state updated)
2. Add multiple tasks (verify gap numbering)
3. Set task status
4. Set task iteration
5. Set task metadata
6. Abandon task
7. List tasks
8. Add/set/remove/list task inputs
9. Add/set/remove/list task outputs
10. Feedback workflow (add feedback, mark addressed)

## Dependencies

- Task 010 (Field Path Parsing) - For set operations
- Task 020 (Project Commands) - For project state
- Task 040 (Input/Output Commands) - Similar pattern for task artifacts

## References

- **Task schema**: `cli/schemas/project/task.cue`
- **SDK task type**: `cli/internal/sdks/project/state/task.go`
- **Task collection**: `cli/internal/sdks/project/state/collections.go`
- **Command spec**: `.sow/knowledge/designs/command-hierarchy-design.md`

## Notes

- Task IDs are gap-numbered (010, 020, 030) for future insertion
- Task state consolidated in project state (no separate state files)
- Task directory still needed for description, log, feedback
- Tasks typically belong to implementation phase, but design allows tasks in any phase
- Feedback inputs use `metadata.status` ("pending" | "addressed") to track resolution
