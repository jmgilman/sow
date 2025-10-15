# Task Management

**Last Updated**: 2025-10-15
**Purpose**: Task structure and lifecycle

This document describes the structure, states, and lifecycle of tasks within the implementation phase.

---

## Table of Contents

- [Overview](#overview)
- [Gap Numbering System](#gap-numbering-system)
- [Task States](#task-states)
- [Iteration Counter](#iteration-counter)
- [Parallel Execution](#parallel-execution)
- [Agent Assignment](#agent-assignment)
- [Dependencies](#dependencies)
- [Task Abandonment](#task-abandonment)
- [Related Documentation](#related-documentation)

---

## Overview

Tasks are discrete units of work within the implementation phase. Orchestrator creates initial task breakdown during planning (with human approval). Tasks use gap numbering for flexibility, track iterations for resumability, support parallel execution for efficiency, and can be abandoned rather than deleted for audit trail preservation.

---

## Gap Numbering System

### Purpose

Allow task insertions without renumbering entire list. Provides flexibility as work evolves and new tasks discovered.

### Numbering Scheme

Initial numbering: 010, 020, 030, 040 (increments of 10). Insertions: 011, 012, 021, 031 (fill gaps between existing numbers). Benefits: no renumbering chaos, clear chronological order, always can insert new tasks, similar to database migrations.

### Example Workflow

Initial tasks (010, 020, 030) created during planning → Work reveals need for additional task between 020 and 030 → Insert 021 without renumbering others → Gap numbering accommodates change naturally.

---

## Task States

### State Values

**`pending`**: Not yet started. Initial state for all tasks. Transitions to in_progress when orchestrator spawns worker.

**`in_progress`**: Currently being worked on. Worker actively executing. Transitions to completed (success) or abandoned (no longer relevant).

**`completed`**: Successfully finished. Terminal state. Work meets acceptance criteria.

**`abandoned`**: Started but no longer relevant. Terminal state. Original approach wrong or requirements changed. Preserves audit trail.

### State Transitions

`pending` → `in_progress` (orchestrator spawns worker for task). `in_progress` → `completed` (worker successfully finishes task). `in_progress` → `abandoned` (task no longer applicable, fail-forward to new approach). Never delete tasks (mark as abandoned instead to preserve history).

---

## Iteration Counter

### Purpose

Track attempts at completing a task. Enables resumability and provides audit trail of approaches tried.

### Counter Management

Orchestrator manages iteration counter (not workers). Increments before spawning worker. Worker reads current iteration from task state. Used to construct agent ID: `{agent_role}-{iteration}` (examples: implementer-1, implementer-2, architect-3).

### When Iterations Increment

Worker gets stuck and task paused (new worker assigned) → increment. Feedback added by human (worker resumes with corrections) → increment. Worker completes successfully (same attempt) → no increment.

### Iteration History

Each iteration represents distinct attempt with own agent ID. Task logs show which iteration did what. Enables understanding of what was tried and why. Supports debugging and learning from failures.

---

## Parallel Execution

### Concept

Tasks can be marked for parallel execution when truly independent. Orchestrator spawns multiple workers simultaneously. Each worker works independently. Orchestrator waits for all to complete before advancing.

### Marking Tasks

Tasks marked with `parallel: true` flag in state. Tasks in same parallel group execute concurrently. Orchestrator identifies parallel groups and spawns workers together.

### Guidelines

Only parallel if truly independent (no shared file modifications, no interdependencies, clear separation of concerns). Benefits: faster execution, independent work streams, natural task parallelism.

---

## Agent Assignment

### Assignment at Planning

Orchestrator assigns agent to each task during project planning. Assignment stored in task state (`assigned_agent` field). Based on task type: design work → architect, code work → implementer, testing work → reviewer, documentation → documenter.

### Execution Time

Orchestrator reads `assigned_agent` from state → spawns correct worker type → no guessing or inference needed during execution. Predictable and auditable.

### Flexibility

Can change assignment if needed (update state manually or orchestrator reassigns on retry). Typically assignment remains stable throughout task lifecycle.

---

## Dependencies

### Task Dependencies

Tasks can depend on other tasks completing first. Dependencies tracked in task state. Orchestrator enforces dependency order before spawning workers.

### Sequential Execution

Non-parallel tasks execute sequentially within phase. Orchestrator processes tasks in ID order (010, 020, 030). Ensures logical progression through work.

### Cross-Phase Dependencies

Implementation phase always follows design phase (if enabled). Review phase always follows implementation phase. Enforced by phase ordering.

**See Also**: [PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md#phase-transitions)

---

## Task Abandonment

### Fail-Forward Philosophy

Add tasks instead of reverting when approach changes. Never delete tasks (preserves audit trail). Mark as `abandoned` when no longer relevant. Add new tasks with gap numbering for correct approach.

### Abandonment Process

Task in progress → Discovery reveals approach wrong → Mark task as abandoned → Create new task with correct approach → New task executes → Both tasks preserved in history.

### Benefits

Shows what was tried (learning from mistakes). Transparent decision-making (why approach changed). No lost context (complete history). Supports debugging (understanding evolution of work).

### Completed Task Count

"All tasks complete" does NOT include abandoned tasks. Abandoned tasks documented but don't block completion. Only pending and completed tasks affect phase completion.

**See Also**: [PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md#fail-forward-logic)

---

## Related Documentation

- **[PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md)** - Task execution and fail-forward mechanics
- **[AGENTS.md](./AGENTS.md)** - Agent assignment and worker coordination
- **[FEEDBACK.md](./FEEDBACK.md)** - Human feedback and iteration mechanics
- **[LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)** - Task state management
- **[SCHEMAS.md](./SCHEMAS.md)** - Task state schema specification
