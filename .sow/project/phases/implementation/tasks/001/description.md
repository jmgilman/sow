# Task: Create Implementation Task Breakdown

## Objective

Create a comprehensive task breakdown for implementing the `sow project status` CLI command.

## Feature Description

The `sow project status` command should provide a pretty-printed overview of the current project state. It's a read-only command that helps users quickly understand where their project stands.

### Expected Output Example

```
Project: add-authentication
Branch: feat/auth
Type: standard

State: ImplementationExecuting

Phases:
  implementation  [in_progress]  3/5 tasks completed
  review          [pending]
  finalize        [pending]

Tasks (implementation):
  [completed]   010  Implement JWT middleware
  [completed]   020  Add login endpoint
  [completed]   030  Add logout endpoint
  [in_progress] 040  Add session management
  [pending]     050  Add auth tests
```

### Requirements

1. **Project header** - Show project name, branch, type, current state
2. **Phase summary** - List all phases with status and task progress (X/Y completed)
3. **Task list** - Show tasks in current phase with status, ID, and name
4. **No arguments required** - Works in any directory with an active project
5. **Error handling** - Clear message if no project exists

### Technical Considerations

- Command should be added under `cli/cmd/project/`
- Follows existing patterns in the codebase (use `cmdutil`, `sow.Context`, etc.)
- Should use the existing state loading from `internal/sdks/project/state/`
- Output formatting should be clean and readable (consider using a table library or manual formatting)
- Tests should verify output format and error cases

### Scope Boundaries

**In scope:**
- Basic status display as described above
- Clean terminal output

**Out of scope (for now):**
- Colorized output
- JSON output format
- Verbose mode with additional details
- Progress bars or animations

## Deliverable

Create task description files in `.sow/project/context/tasks/` following the naming convention `{id}-{name}.md`. Each task should be self-contained with:
- Clear requirements
- Acceptance criteria
- Relevant inputs (file paths the implementer needs to reference)
