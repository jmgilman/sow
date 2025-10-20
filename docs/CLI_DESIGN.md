# CLI Command Design: Project & Task Management

**Status**: Design Document (Pre-Implementation)
**Created**: 2025-10-16
**Purpose**: Comprehensive specification for project and task CLI commands

---

## Table of Contents

- [Overview](#overview)
- [Design Rationale](#design-rationale)
- [Command Hierarchy](#command-hierarchy)
- [Task ID Inference](#task-id-inference)
- [State File Updates](#state-file-updates)
- [Implementation Considerations](#implementation-considerations)

---

## Overview

The project and task management commands provide CLI access to manipulate project and task state files. These commands are primarily designed for AI agents (orchestrator and workers) but can be used manually for debugging or intervention.

### Design Principles

1. **Agent-first**: Optimize for programmatic use by AI agents
2. **Validation**: All state changes validated against CUE schemas before writing
3. **Atomicity**: State file writes are atomic (write to temp, then rename)
4. **Context-aware**: Task commands infer ID from active task or current directory
5. **Simple output**: Machine-readable output (JSON where applicable)

---

## Design Rationale

### Why Separate Project and Task Commands?

The command structure mirrors the ownership model defined in CLAUDE.md:

- **Project commands** (`sow project *`) - Orchestrator-only operations that manage project state, phases, and coordination
- **Task commands** (`sow task *`) - Worker-focused operations for executing and tracking individual tasks

This separation enforces the boundary between coordination (orchestrator) and execution (workers), preventing workers from accidentally modifying project-level state.

### Why Task ID Inference?

Workers operate in a single-task context and shouldn't need to repeatedly specify which task they're working on. The inference strategy (directory → active task → error) balances convenience with safety:

- **Convenience**: Workers in their task directory can run `sow task log ...` without flags
- **Safety**: Ambiguous situations (multiple in-progress tasks, wrong directory) produce clear errors
- **Explicitness**: `--id` flag always available when needed

### Why Atomic Writes?

State files are the single source of truth for project coordination. Corrupted state would break zero-context resumability. The atomic write pattern (write temp, validate, rename) ensures:

- **Consistency**: Invalid states never reach disk
- **Durability**: No partial writes from crashes
- **Simplicity**: No distributed locking needed (single-writer assumption)

### Why CUE Validation?

Embedded CUE schemas provide:

- **Type safety**: Catch invalid state transitions before they hit disk
- **Self-documentation**: Schemas define the contract between CLI and agents
- **Version control**: Schema changes are versioned with the CLI binary
- **Fast feedback**: Validation happens in <1ms, suitable for tight loops

### Command Design Decisions

**Granular vs. Composite Commands**: We chose granular commands (e.g., `project artifact add`, `project artifact approve`) over composite operations because:

- Agents can compose operations in any order
- Each operation has clear success/failure semantics
- Easier to test and maintain individual operations
- Better for fail-forward workflows (retry single step, not entire flow)

**Flags vs. Positional Arguments**: We use positional arguments for required identifiers (name, ID, path) and flags for optional/semantic parameters (--approved, --phase, --format) to make commands readable in agent logs.

---

## Command Hierarchy

### Command Structure Overview

For detailed command reference (usage, arguments, examples), see [CLI_REFERENCE.md](./CLI_REFERENCE.md).

### Project Commands

```bash
# Core project operations
sow project init
sow project status
sow project delete

# Phase management
sow project phase enable <phase> [--type <discovery-type>]
sow project phase status
sow project phase complete <phase>

# Artifact management
sow project artifact add <path> --phase <phase> [--approved]
sow project artifact approve <path> --phase <phase>
sow project artifact list [--phase <phase>]

# Review management
sow project review increment
sow project review add-report <path> --assessment <pass|fail>

# Finalize management
sow project finalize add-document <path>
sow project finalize move-artifact <from> <to>
```

### Task Commands

```bash
# Core task operations
sow task init <name> [--id <id>]
sow task list
sow task show [<id>]
sow task set-status <status> [<id>]
sow task abandon [<id>]

# Task state management
sow task state increment [<id>]
sow task state set-agent <agent> [<id>]
sow task state add-reference <path> [<id>]
sow task state add-file <path> [<id>]

# Task feedback
sow task feedback add <message> [<id>]
sow task feedback mark-addressed <feedback-id> [<id>]
```

---

## Task ID Inference

Task ID is optional for task commands. When omitted, ID is inferred from:

### Inference Order

1. **Current directory**: If running from `.sow/project/phases/implementation/tasks/<id>/`, use that ID
2. **Active task in state**: If project state has a task with status `in_progress`, use that ID
3. **Error**: If neither applies, return error requesting explicit `--id` flag

### Implementation Pattern

```go
func inferTaskID() (string, error) {
    // 1. Check current directory
    if id := inferFromCurrentDir(); id != "" {
        return id, nil
    }

    // 2. Check active task in state
    if id := inferFromActiveTask(); id != "" {
        return id, nil
    }

    // 3. Error - cannot infer
    return "", errors.New("cannot infer task ID: use --id flag or run from task directory")
}
```

---

## State File Updates

### Atomicity Strategy

All state file updates follow this pattern:

1. Read current state file
2. Validate current state against CUE schema
3. Apply modification
4. Validate modified state against CUE schema
5. Write to temporary file (`.state.yaml.tmp`)
6. Atomic rename to actual file
7. Remove temp file on error

### Validation

All state modifications validated against embedded CUE schemas:
- `project_state.cue` - Project state schema
- `task_state.cue` - Task state schema

Invalid states are rejected with detailed error messages.

### Concurrent Access

Not expected in normal operation. If needed:
- File locking can be added
- Optimistic concurrency (check updated_at before write)
- For now: assume single writer (orchestrator or manual intervention)

---

## Implementation Considerations

### File I/O Pattern

All commands that modify state follow this pattern:

```go
func updateProjectState(modify func(*State) error) error {
    // 1. Read current state
    state, err := readProjectState()
    if err != nil {
        return err
    }

    // 2. Validate current state
    if err := validateState(state); err != nil {
        return fmt.Errorf("invalid current state: %w", err)
    }

    // 3. Apply modification
    if err := modify(state); err != nil {
        return err
    }

    // 4. Validate modified state
    if err := validateState(state); err != nil {
        return fmt.Errorf("invalid modified state: %w", err)
    }

    // 5. Atomic write
    return atomicWrite(".sow/project/state.yaml", state)
}

func atomicWrite(path string, data interface{}) error {
    tmpPath := path + ".tmp"

    // Write to temp file
    if err := writeYAML(tmpPath, data); err != nil {
        return err
    }

    // Atomic rename
    if err := os.Rename(tmpPath, path); err != nil {
        os.Remove(tmpPath) // Clean up on error
        return err
    }

    return nil
}
```

### Validation Strategy

Use embedded CUE schemas for validation:

```go
func validateState(state *ProjectState) error {
    // Marshal to YAML
    yamlData, err := yaml.Marshal(state)
    if err != nil {
        return err
    }

    // Validate against embedded CUE schema
    return schemas.ValidateProjectState(yamlData)
}
```

### Error Handling

All commands should:
1. Return non-zero exit codes on error
2. Print clear error messages to stderr
3. Use structured errors for programmatic parsing

### Output Format

Default output optimized for humans, JSON format for programmatic use:

```bash
# Human-readable
sow task list
# ID   STATUS       NAME
# 010  completed    Create User model

# Machine-readable
sow task list --format json
# [{"id":"010","status":"completed","name":"Create User model"}]
```

---

## Related Documentation

- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - Complete command reference for users and agents
- **[CLAUDE.md](../.claude/CLAUDE.md)** - File ownership model and agent responsibilities
- **[SCHEMAS.md](./SCHEMAS.md)** - CUE schema documentation

---

## Implementation Roadmap

Commands should be implemented in phases:

1. **Phase 1: Project Core**
   - `sow project init`
   - `sow project status`
   - `sow project delete`

2. **Phase 2: Phase Management**
   - `sow project phase enable`
   - `sow project phase status`
   - `sow project phase complete`

3. **Phase 3: Task Core**
   - `sow task init`
   - `sow task list`
   - `sow task show`
   - `sow task set-status`

4. **Phase 4: Advanced Features**
   - Artifact management
   - Review management
   - Finalize management
   - Task state management
   - Feedback management
