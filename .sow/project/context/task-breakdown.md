# Task Breakdown: Unified CLI Command Structure

## Overview

Implement universal CLI commands (set, input/output, advance) that work with the SDK state layer. The SDK foundation is complete - we need CLI commands that leverage universal Project/Phase/Task/Artifact types.

**Testing Strategy**: Integration tests only (txtar format) for CLI commands, unit tests only for library code.

**Key Design References**:
- `.sow/knowledge/designs/command-hierarchy-design.md` - Complete command specifications
- `.sow/knowledge/designs/project-sdk-implementation.md` - SDK architecture and patterns
- Existing SDK: `cli/internal/sdks/project/` - Reference for Load/Save patterns

---

## Task 1: Implement Field Path Parsing Utility (TDD)

**Description**: Create field path parser that handles dot notation and routes metadata fields automatically.

**Scope**:
- Field path parser (`fieldpath.go`) - parses `metadata.foo.bar` notation
- Artifact operation helpers (`artifacts.go`) - shared logic for artifact commands
- Unit tests (this is library code, not CLI)

**Key Behaviors**:
- `approved` → direct field access
- `metadata.assessment` → routes to metadata map
- `metadata.foo.bar` → nested map traversal
- Type conversion (string → bool, int, etc.)
- Validation of known vs unknown fields

**Files to create**:
- `cli/internal/cmdutil/fieldpath.go`
- `cli/internal/cmdutil/fieldpath_test.go`
- `cli/internal/cmdutil/artifacts.go`
- `cli/internal/cmdutil/artifacts_test.go`

**Acceptance Criteria**:
- [ ] Field path parser handles dot notation
- [ ] Automatic routing: `metadata.*` → metadata map
- [ ] Known fields handled directly (approved, type, path, status, etc.)
- [ ] Type conversion for bool, int, string
- [ ] Clear errors for invalid paths
- [ ] All unit tests passing

**Agent**: implementer

**Dependencies**: None

---

## Task 2: Replace Project Command with New/Continue Subcommands (TDD)

**Description**: Delete existing `sow project` and replace with `sow project new` / `sow project continue` subcommands using SDK.

**Context**:
- Current `cli/cmd/project.go` has unified behavior (creates or continues based on state)
- Must split into two explicit subcommands
- Preserve all existing logic (worktree management, issue linking, etc.)
- Use SDK `state.Load(ctx)` and `project.Save()` instead of old internal/project package

**Commands to implement**:
- `sow project new --branch <branch> [--issue <number>] "<description>"`
- `sow project continue [--branch <branch>]`
- `sow project set <field-path> <value>`
- `sow project delete`

**Key Behaviors**:
- `new`: Creates project using `state.Create()`, same worktree/issue logic as before
- `continue`: Loads existing project using `state.Load()`, launches Claude Code
- `set`: Uses field path parser for scalar and metadata fields
- All operations use Load/Save cycle

**Integration test to write first**:
- `cli/testdata/script/unified_commands/project/project_lifecycle.txtar`
- Test: new, continue, set description, set metadata, delete

**Files to delete**:
- `cli/cmd/project.go`

**Files to create**:
- `cli/cmd/project/project.go` - Root command with subcommands
- `cli/cmd/project/new.go`
- `cli/cmd/project/continue.go`
- `cli/cmd/project/set.go`
- `cli/cmd/project/delete.go`
- `cli/testdata/script/unified_commands/project/project_lifecycle.txtar`

**Acceptance Criteria**:
- [ ] Integration test written first
- [ ] Old project.go deleted
- [ ] New/continue preserve all existing functionality
- [ ] Set uses field path parser
- [ ] All use SDK Load/Save
- [ ] Integration test passes

**Agent**: implementer

**Dependencies**: Task 1

---

## Task 3: Implement Phase Commands (TDD)

**Description**: Implement phase-level operations with metadata via dot notation.

**Context**:
- See `.sow/knowledge/designs/command-hierarchy-design.md` lines 512-532 for spec
- Phases accessed via `project.Phases.Get(phaseName)`
- Direct field mutation pattern (no methods, just field assignment)

**Commands to implement**:
- `sow phase set <field-path> <value> [--phase <name>]`

**Key Behaviors**:
- Load project, navigate to phase
- Set status, enabled (direct fields)
- Set metadata.* via field path parser
- Default to active phase if --phase omitted

**Integration tests to write first**:
- `cli/testdata/script/unified_commands/phase/phase_operations.txtar`
- `cli/testdata/script/unified_commands/phase/phase_errors.txtar`

**Files to create**:
- `cli/cmd/phase.go`
- Integration test files

**Acceptance Criteria**:
- [ ] Integration tests written first
- [ ] Phase set modifies fields and metadata
- [ ] --phase defaults to active phase
- [ ] Dot notation routes to metadata
- [ ] Integration tests pass

**Agent**: implementer

**Dependencies**: Task 1, Task 2

---

## Task 4: Implement Input/Output Commands (TDD)

**Description**: Implement phase-level artifact management.

**Context**:
- See `.sow/knowledge/designs/command-hierarchy-design.md` lines 397-738 for complete spec
- Artifacts use collection pattern: `phase.Inputs.Get(index)`, `phase.Outputs.Add(artifact)`
- Index-based operations (not IDs)
- Type validation via project type config

**Commands to implement**:
- `sow input add --type <type> [--phase <name>] [--path <path>] [...]`
- `sow input set --index <n> <field-path> <value> [--phase <name>]`
- `sow input remove --index <n> [--phase <name>]`
- `sow input list [--phase <name>]`
- `sow output add/set/remove/list` (same pattern)

**Key Behaviors**:
- Navigate to phase, then inputs/outputs collection
- Index-based references (0, 1, 2, ...)
- Validate artifact type per project type config
- Field paths work: `approved`, `metadata.assessment`

**Integration tests to write first**:
- `cli/testdata/script/unified_commands/artifacts/input_operations.txtar`
- `cli/testdata/script/unified_commands/artifacts/output_operations.txtar`
- `cli/testdata/script/unified_commands/artifacts/artifact_metadata.txtar`
- `cli/testdata/script/unified_commands/artifacts/artifact_validation.txtar`

**Files to create**:
- `cli/cmd/input.go`
- `cli/cmd/output.go`
- Integration test files

**Acceptance Criteria**:
- [ ] Integration tests written first
- [ ] All input/output operations implemented
- [ ] Field path parser used for set
- [ ] Type validation enforced
- [ ] Integration tests pass

**Agent**: implementer

**Dependencies**: Task 1, Task 3

---

## Task 5: Implement Task Commands (TDD)

**Description**: Implement task management and task-level input/output commands.

**Context**:
- See `.sow/knowledge/designs/command-hierarchy-design.md` lines 388-446, 533-598, 740-875
- Tasks accessed via `phase.Tasks.Get(id)`
- Task state lives in project state (not separate files)
- Task directory still created for description.md, log.md, feedback/
- Gap-numbered IDs: 010, 020, 030, ...

**Commands to implement**:
- `sow task add <name> [--agent <agent>] [--description "..."]`
- `sow task set --id <task-id> <field-path> <value>`
- `sow task abandon --id <task-id>`
- `sow task list`
- `sow task input add/set/remove/list --id <id> [...]`
- `sow task output add/set/remove/list --id <id> [...]`

**Key Behaviors**:
- Generate gap-numbered task IDs
- Create task directory on add
- Task inputs/outputs mirror phase artifact pattern
- Set uses field path parser for status, iteration, metadata

**Integration tests to write first**:
- `cli/testdata/script/unified_commands/tasks/task_operations.txtar`
- `cli/testdata/script/unified_commands/tasks/task_inputs_outputs.txtar`
- `cli/testdata/script/unified_commands/tasks/task_lifecycle.txtar`

**Files to create**:
- `cli/cmd/task.go`
- `cli/cmd/task_input.go`
- `cli/cmd/task_output.go`
- Integration test files

**Acceptance Criteria**:
- [ ] Integration tests written first
- [ ] All task operations implemented
- [ ] Task state in project state
- [ ] Task directory created on add
- [ ] Integration tests pass

**Agent**: implementer

**Dependencies**: Task 1, Task 4

---

## Task 6: Update Advance Command (TDD)

**Description**: Update advance command to use SDK state machine.

**Context**:
- See `.sow/knowledge/designs/project-sdk-implementation.md` lines 217-262 for Advance() flow
- Project.Advance() uses OnAdvance event determiners
- Guards evaluated automatically via machine.CanFire()

**Changes needed**:
- Load project via SDK
- Call `project.Advance()`
- Save updated state
- Display new state

**Integration test to write first**:
- `cli/testdata/script/unified_commands/integration/state_transitions.txtar`

**Files to modify**:
- `cli/cmd/advance.go`

**Files to create**:
- Integration test file

**Acceptance Criteria**:
- [ ] Integration test written first
- [ ] Advance uses SDK Project.Advance()
- [ ] Guards evaluated correctly
- [ ] Integration test passes

**Agent**: implementer

**Dependencies**: Task 2

---

## Task 7: Comprehensive Integration Testing (TDD)

**Description**: Create full lifecycle integration tests and clean up old tests.

**Actions**:
1. Delete all old integration tests using obsolete commands
2. Reorganize worktree tests into subfolder
3. Create comprehensive integration tests

**Tests to delete**:
- `cli/testdata/script/agent_artifact_commands.txtar`
- `cli/testdata/script/agent_phase_commands.txtar`
- `cli/testdata/script/agent_project_lifecycle.txtar`
- `cli/testdata/script/agent_task_commands.txtar`
- `cli/testdata/script/agent_task_feedback_commands.txtar`
- `cli/testdata/script/agent_task_review_workflow.txtar`
- `cli/testdata/script/agent_task_state_commands.txtar`
- `cli/testdata/script/standard_project_full_lifecycle.txtar`
- `cli/testdata/script/standard_project_state_transitions.txtar`
- `cli/testdata/script/review_fail_loop_back.txtar`
- `cli/testdata/script/agent_logging_commands.txtar`
- `cli/testdata/script/agent_commands_error_cases.txtar`

**Tests to reorganize**:
- Move `worktree_*.txtar` → `cli/testdata/script/worktree/`

**New integration tests to create**:

`cli/testdata/script/unified_commands/integration/full_lifecycle.txtar`:
- Complete standard project lifecycle
- Planning: add context inputs, create task_list output, approve, advance
- Implementation: create tasks, add inputs/outputs, set metadata.tasks_approved, advance
- Execution: complete tasks, advance
- Review: add review with metadata.assessment pass, approve, advance
- Finalize: progress through states, complete

`cli/testdata/script/unified_commands/integration/review_fail_loop.txtar`:
- Review with metadata.assessment fail loops back to implementation

`cli/testdata/script/unified_commands/integration/feedback_workflow.txtar`:
- Add feedback as task input
- Address via metadata.status addressed
- Complete task iteration

**Final directory structure**:
```
cli/testdata/script/
├── unified_commands/
│   ├── project/
│   ├── phase/
│   ├── artifacts/
│   ├── tasks/
│   └── integration/
│       ├── full_lifecycle.txtar
│       ├── state_transitions.txtar
│       ├── review_fail_loop.txtar
│       └── feedback_workflow.txtar
└── worktree/
    └── (reorganized tests)
```

**Acceptance Criteria**:
- [ ] All old tests deleted
- [ ] Worktree tests reorganized
- [ ] Full lifecycle test passes
- [ ] Review fail loop test passes
- [ ] Feedback workflow test passes
- [ ] All integration tests passing

**Agent**: implementer

**Dependencies**: Task 2, Task 3, Task 4, Task 5, Task 6

---

## Task 8: Remove Old Commands and Cleanup

**Description**: Clean up deprecated commands and code.

**Files/commands to remove**:
- `cli/cmd/agent/artifact.go`
- `cli/cmd/agent/artifact_add.go`
- `cli/cmd/agent/artifact_approve.go`
- `cli/cmd/agent/artifact_list.go`
- Any old task reference/feedback commands
- Any old phase approval commands
- Old internal project package references (if any)

**Acceptance Criteria**:
- [ ] All old agent commands removed
- [ ] No references to old patterns
- [ ] CLI help text accurate
- [ ] Clean build, no dead code
- [ ] All tests still passing

**Agent**: implementer

**Dependencies**: Task 7

---

## Summary

**Total Tasks**: 8
**Testing**: Integration tests (txtar) for CLI, unit tests only for library code
**Command Changes**: `sow project` → `sow project new` / `sow project continue`

**Key References**:
- Command specs: `.sow/knowledge/designs/command-hierarchy-design.md`
- SDK patterns: `.sow/knowledge/designs/project-sdk-implementation.md`
- Existing SDK: `cli/internal/sdks/project/`
