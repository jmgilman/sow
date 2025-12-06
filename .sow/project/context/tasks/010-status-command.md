# Task 010: Implement `sow project status` Command

## Context

Add a new CLI command that displays the current project state in a readable format. This is a read-only command that helps users quickly understand their project's progress.

## Goals

- Provide a quick overview of project state without opening files
- Show meaningful progress indicators (X/Y completed tasks)
- Follow existing CLI command patterns
- Handle the "no project" case gracefully

## Requirements

### Command Structure

1. Add `status.go` under `cli/cmd/project/`
2. Register command in `project.go` via `cmd.AddCommand(newStatusCmd())`
3. Command: `sow project status`
4. Short description: `Show current project status`

### Output Format

```
Project: {name}
Branch: {branch}
Type: {type}

State: {statechart.current_state}

Phases:
  {phase_name}  [{status}]  {X}/{Y} tasks completed
  ...

Tasks ({current_phase}):
  [{status}]  {id}  {name}
  ...
```

### Implementation Details

1. **Get sow context**: Use `cmdutil.RequireInitialized(cmd.Context())`
2. **Load project state**: Use `state.Load(ctx)` from `cli/internal/sdks/project/state`
3. **Determine current phase**: Use `proj.Statechart.Current_state` to infer current phase
4. **Calculate task counts**: Count tasks by status per phase
5. **Output to stdout**: Print formatted output (not stderr)

### Error Handling

- If `.sow` not initialized: Return `sow.ErrNotInitialized` (standard error)
- If no project exists (state.yaml missing): Return descriptive error "no active project"

## Acceptance Criteria

- [ ] `sow project status` displays formatted output when project exists
- [ ] Shows project header (name, branch, type, state)
- [ ] Shows phase list with status and task progress
- [ ] Shows task list for current/active phase
- [ ] Returns clear error when no project exists
- [ ] Follows existing command patterns (cobra, cmdutil)

## Technical Notes

### Determining Current Phase

The statechart state name typically includes the phase (e.g., `ImplementationPlanning`, `ReviewActive`). You may need to:
- Parse the state name to extract phase
- Or iterate phases and check which has `status: "in_progress"`

### Phase Display Order

Standard project phases are: `implementation`, `review`, `finalize`. Display in this order.

### Task Status Display

Map task status to display format:
- `pending` → `pending`
- `in_progress` → `in_progress`
- `completed` → `completed`
- `paused` → `paused`
- `needs_review` → `needs_review`
- `abandoned` → `abandoned`

## Relevant Inputs

- `cli/cmd/project/project.go` - Main project command, add subcommand here
- `cli/cmd/project/delete.go` - Example subcommand pattern to follow
- `cli/internal/sdks/project/state/loader.go` - `state.Load()` function
- `cli/internal/sdks/project/state/project.go` - Project type and methods
- `cli/internal/cmdutil/context.go` - `RequireInitialized()` helper
- `cli/schemas/project/project.cue` - Schema definitions
- `cli/schemas/project/phase.cue` - Phase schema
- `cli/schemas/project/task.cue` - Task schema
