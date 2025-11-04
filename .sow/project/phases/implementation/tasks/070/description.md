# Task 070: Comprehensive Integration Testing (TDD)

# Task 070: Comprehensive Integration Testing (TDD)

## Overview

Delete obsolete integration tests, reorganize worktree tests, and create comprehensive integration tests for the new unified command structure.

## Context

The old integration tests use commands that no longer exist (e.g., `sow agent artifact add`, `sow agent phase approve-tasks`). These must be deleted and replaced with tests using the new unified commands.

## Design References

- **Test examples**: Existing tests in `cli/testdata/script/` (for txtar format reference)
- **Full lifecycle flow**: `.sow/knowledge/designs/command-hierarchy-design.md` lines 992-1138
- **Standard project spec**: `.sow/knowledge/designs/project-sdk-implementation.md`

## Requirements

### Actions

1. **Delete obsolete tests** - Remove all tests using old command hierarchy
2. **Reorganize worktree tests** - Move to subfolder for better organization
3. **Create comprehensive integration tests** - Full lifecycle and edge cases

## TDD Approach

This task is exclusively about integration testing - no implementation code.

### Step 1: Delete Obsolete Tests

Remove these files (they use old commands):

```bash
cli/testdata/script/agent_artifact_commands.txtar
cli/testdata/script/agent_phase_commands.txtar
cli/testdata/script/agent_project_lifecycle.txtar
cli/testdata/script/agent_task_commands.txtar
cli/testdata/script/agent_task_feedback_commands.txtar
cli/testdata/script/agent_task_review_workflow.txtar
cli/testdata/script/agent_task_state_commands.txtar
cli/testdata/script/standard_project_full_lifecycle.txtar
cli/testdata/script/standard_project_state_transitions.txtar
cli/testdata/script/review_fail_loop_back.txtar
cli/testdata/script/agent_logging_commands.txtar
cli/testdata/script/agent_commands_error_cases.txtar
```

### Step 2: Reorganize Worktree Tests

Create `cli/testdata/script/worktree/` directory and move:

```bash
cli/testdata/script/worktree_*.txtar → cli/testdata/script/worktree/
```

### Step 3: Create New Integration Tests

Create tests in `cli/testdata/script/unified_commands/integration/`.

## New Integration Tests to Create

### 1. Full Lifecycle Test

**File**: `cli/testdata/script/unified_commands/integration/full_lifecycle.txtar`

**Coverage**: Complete standard project lifecycle using new commands.

**Flow**:

```txtar
# Test: Complete Standard Project Lifecycle
# Coverage: Planning → Implementation → Review → Finalize → Complete

# ===== Setup =====
exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test-lifecycle
exec sow init

# ===== Project Creation =====
exec sow project new --branch feat/test-lifecycle --no-launch "Full lifecycle test"
exists .sow/project/state.yaml
exec cat .sow/project/state.yaml
stdout 'current_state: PlanningActive'

# ===== Planning Phase =====

# Add context input
exec mkdir -p .sow/project/context
exec sh -c 'echo "Context document" > .sow/project/context/research.md'
exec sow input add --type context --path context/research.md --phase planning
exec cat .sow/project/state.yaml
stdout 'type: context'

# Create task list output
exec mkdir -p .sow/project/planning
exec sh -c 'echo "# Task List\n\n- Task 1: Implement feature\n- Task 2: Write tests" > .sow/project/planning/tasks.md'
exec sow output add --type task_list --path planning/tasks.md --phase planning
exec cat .sow/project/state.yaml
stdout 'type: task_list'
stdout 'approved: false'

# Approve task list
exec sow output set --index 0 approved true --phase planning
exec cat .sow/project/state.yaml
stdout 'approved: true'

# Advance to implementation
exec sow advance
stdout 'Advanced to: ImplementationPlanning'
exec cat .sow/project/state.yaml
stdout 'current_state: ImplementationPlanning'

# ===== Implementation Planning =====

# Add tasks
exec sow task add "Implement feature" --agent implementer --description "Core feature implementation"
exec sow task add "Write tests" --agent implementer --description "Unit and integration tests"
exists .sow/project/phases/implementation/tasks/010/description.md
exists .sow/project/phases/implementation/tasks/020/description.md
exec cat .sow/project/state.yaml
stdout 'id: "010"'
stdout 'id: "020"'

# Add task inputs (references)
exec mkdir -p .sow/sinks
exec sh -c 'echo "Style guide" > .sow/sinks/style-guide.md'
exec sow task input add --id 010 --type reference --path ../../../sinks/style-guide.md
exec cat .sow/project/state.yaml
stdout 'type: reference'

# Approve tasks
exec sow phase set metadata.tasks_approved true --phase implementation
exec cat .sow/project/state.yaml
stdout 'tasks_approved: true'

# Advance to execution
exec sow advance
stdout 'Advanced to: ImplementationExecuting'

# ===== Implementation Execution =====

# Set tasks in progress
exec sow task set --id 010 status in_progress
exec sow task set --id 020 status in_progress

# Add task outputs (modified files)
exec sow task output add --id 010 --type modified --path src/feature.go
exec sow task output add --id 020 --type modified --path tests/feature_test.go
exec cat .sow/project/state.yaml
stdout 'type: modified'

# Complete tasks
exec sow task set --id 010 status completed
exec sow task set --id 020 status completed
exec cat .sow/project/state.yaml
stdout 'status: completed'

# Advance to review
exec sow advance
stdout 'Advanced to: ReviewActive'

# ===== Review Phase =====

# Create review report
exec mkdir -p .sow/project/review
exec sh -c 'echo "# Review Report\n\nAll checks passed." > .sow/project/review/report.md'
exec sow output add --type review --path review/report.md --phase review
exec sow output set --index 0 metadata.assessment pass --phase review
exec cat .sow/project/state.yaml
stdout 'assessment: pass'

# Approve review
exec sow output set --index 0 approved true --phase review
exec cat .sow/project/state.yaml
stdout 'approved: true'

# Advance to finalize
exec sow advance
stdout 'Advanced to: FinalizeDocumentation'

# ===== Finalize Phase =====

# Documentation state (simplified - would normally update docs)
exec sow advance
stdout 'Advanced to: FinalizeChecks'

# Checks state (simplified - would normally run tests/linters)
exec sow advance
stdout 'Advanced to: FinalizeDelete'

# Delete project
exec sow phase set metadata.project_deleted true --phase finalize
exec sow advance
stdout 'Advanced to: ProjectComplete'

# Verify final state
exec cat .sow/project/state.yaml
stdout 'current_state: ProjectComplete'
```

### 2. Review Fail Loop Test

**File**: `cli/testdata/script/unified_commands/integration/review_fail_loop.txtar`

**Coverage**: Review failure loops back to implementation.

**Flow**:

```txtar
# Test: Review Failure Loop Back to Implementation
# Coverage: Review fails → returns to ImplementationPlanning

# Setup and get to ReviewActive state (abbreviated)
exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test-review-fail
exec sow init
exec sow project new --branch feat/test-review-fail --no-launch "Test review fail"

# Fast-forward to ReviewActive (create required artifacts)
# ... (setup planning output, advance, create tasks, approve, advance, complete tasks, advance)

# ===== Review Phase - Fail Assessment =====

# Create review with fail assessment
exec mkdir -p .sow/project/review
exec sh -c 'echo "# Review Report\n\nIssues found." > .sow/project/review/report.md'
exec sow output add --type review --path review/report.md --phase review
exec sow output set --index 0 metadata.assessment fail --phase review
exec sow output set --index 0 approved true --phase review

# Advance (should loop back to ImplementationPlanning)
exec sow advance
stdout 'Advanced to: ImplementationPlanning'
exec cat .sow/project/state.yaml
stdout 'current_state: ImplementationPlanning'

# Verify we can iterate on tasks
exec sow task set --id 010 status pending
exec sow task set --id 010 iteration 2
exec cat .sow/project/state.yaml
stdout 'iteration: 2'
```

### 3. Feedback Workflow Test

**File**: `cli/testdata/script/unified_commands/integration/feedback_workflow.txtar`

**Coverage**: Add feedback as task input, address it.

**Flow**:

```txtar
# Test: Feedback Workflow
# Coverage: Add feedback, address feedback, track status

# Setup and get to ImplementationExecuting (abbreviated)
exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test-feedback
exec sow init
exec sow project new --branch feat/test-feedback --no-launch "Test feedback"

# Fast-forward to ImplementationExecuting with one task
# ... (setup planning, advance, create task, approve, advance)

# ===== Add Feedback =====

# Create feedback file
exec mkdir -p .sow/project/phases/implementation/tasks/010/feedback
exec sh -c 'echo "Please fix the error handling" > .sow/project/phases/implementation/tasks/010/feedback/001.md'

# Add as task input
exec sow task input add --id 010 --type feedback --path feedback/001.md
exec cat .sow/project/state.yaml
stdout 'type: feedback'
stdout 'path: feedback/001.md'

# List task inputs (verify feedback present)
exec sow task input list --id 010
stdout '\[0\] feedback: feedback/001.md'

# ===== Address Feedback =====

# Mark feedback as addressed via metadata
exec sow task input set --id 010 --index 0 metadata.status addressed
exec cat .sow/project/state.yaml
stdout 'status: addressed'

# Verify feedback tracking
exec cat .sow/project/state.yaml
stdout 'type: feedback'
stdout 'status: addressed'

# Complete task iteration
exec sow task set --id 010 status completed
exec sow task set --id 010 iteration 2
exec cat .sow/project/state.yaml
stdout 'iteration: 2'
```

## Final Directory Structure

After this task:

```
cli/testdata/script/
├── unified_commands/
│   ├── project/
│   │   └── project_lifecycle.txtar (from Task 020)
│   ├── phase/
│   │   ├── phase_operations.txtar (from Task 030)
│   │   └── phase_errors.txtar (from Task 030)
│   ├── artifacts/
│   │   ├── input_operations.txtar (from Task 040)
│   │   ├── output_operations.txtar (from Task 040)
│   │   ├── artifact_metadata.txtar (from Task 040)
│   │   └── artifact_validation.txtar (from Task 040)
│   ├── tasks/
│   │   ├── task_operations.txtar (from Task 050)
│   │   ├── task_inputs_outputs.txtar (from Task 050)
│   │   └── task_lifecycle.txtar (from Task 050)
│   └── integration/
│       ├── state_transitions.txtar (from Task 060)
│       ├── full_lifecycle.txtar (NEW)
│       ├── review_fail_loop.txtar (NEW)
│       └── feedback_workflow.txtar (NEW)
└── worktree/
    ├── worktree_command_basic.txtar
    ├── worktree_management_lifecycle.txtar
    ├── worktree_multi_session_concurrency.txtar
    ├── worktree_project_session.txtar
    ├── worktree_reuse_existing.txtar
    ├── worktree_shared_resources.txtar
    ├── worktree_state_isolation.txtar
    └── worktree_uncommitted_changes.txtar
```

## Files to Delete

All files listed in Step 1 above.

## Files to Create

### Reorganization

- `cli/testdata/script/worktree/` directory
- Move all `worktree_*.txtar` files into it

### New Integration Tests

- `cli/testdata/script/unified_commands/integration/full_lifecycle.txtar`
- `cli/testdata/script/unified_commands/integration/review_fail_loop.txtar`
- `cli/testdata/script/unified_commands/integration/feedback_workflow.txtar`

## Acceptance Criteria

- [ ] All old command-based tests deleted
- [ ] Worktree tests reorganized into subfolder
- [ ] Full lifecycle test created and passes
- [ ] Review fail loop test created and passes
- [ ] Feedback workflow test created and passes
- [ ] All integration tests passing
- [ ] Test organization logical and maintainable
- [ ] No references to old commands in any tests

## Testing Strategy

This task IS the testing - all work is creating/organizing integration tests.

**Full lifecycle test must cover**:
1. Project creation
2. Planning phase (inputs, outputs, approval, advance)
3. Implementation planning (tasks, inputs, approval, advance)
4. Implementation execution (task status, outputs, completion, advance)
5. Review phase (review artifact, metadata.assessment, approval, advance)
6. Finalize phase (all states, metadata flags, advance)
7. Project completion

**Review fail loop must cover**:
1. Review with metadata.assessment fail
2. Transition back to ImplementationPlanning
3. Task iteration increment

**Feedback workflow must cover**:
1. Add feedback as task input
2. Set metadata.status addressed
3. Track feedback resolution

## Dependencies

- Task 020 (Project Commands) - Required for project lifecycle
- Task 030 (Phase Commands) - Required for phase operations
- Task 040 (Input/Output Commands) - Required for artifacts
- Task 050 (Task Commands) - Required for task operations
- Task 060 (Advance Command) - Required for state transitions

All previous tasks must be complete and their integration tests passing before this task.

## References

- **Existing tests**: `cli/testdata/script/` - Reference for txtar format
- **Test runner**: See how existing tests are executed in the codebase
- **Full lifecycle example**: `.sow/knowledge/designs/command-hierarchy-design.md` lines 992-1138

## Notes

- txtar format: `exec` for commands, `stdout`/`stderr` for output validation, `exists`/`! exists` for file checks
- Tests should be self-contained (git init, setup, test, no external dependencies)
- Use `--no-launch` flag for project commands to avoid launching Claude Code
- Tests run in isolated temporary directories
- State validation done via `cat .sow/project/state.yaml` and `stdout` assertions
