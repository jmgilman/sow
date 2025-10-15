# Human Feedback System

**Last Updated**: 2025-10-15
**Purpose**: Feedback mechanism for corrections

This document describes the structured system for humans to provide corrections and guidance to workers during task execution.

---

## Table of Contents

- [Overview](#overview)
- [Creating Feedback](#creating-feedback)
- [Feedback Storage](#feedback-storage)
- [Feedback Processing](#feedback-processing)
- [Feedback Tracking](#feedback-tracking)
- [Multiple Iterations](#multiple-iterations)
- [Related Documentation](#related-documentation)

---

## Overview

Structured way for humans to provide corrections and guidance during task execution. Orchestrator creates numbered feedback files in task directory. Worker reads feedback on subsequent iteration and incorporates changes. Feedback tracked in task state with status values. Supports multiple feedback rounds per task.

---

## Creating Feedback

### When to Use

Agent made incorrect assumption (implementation doesn't match mental model). Implementation doesn't meet requirements (missing functionality or wrong approach). Code needs specific changes (technical corrections or style issues). Agent got stuck on wrong approach (needs redirection).

### Process

User identifies issue with task work → User explains problem to orchestrator → Orchestrator creates numbered feedback file → Orchestrator increments task iteration counter → Orchestrator spawns worker with feedback context.

### Orchestrator Role

Listens to human feedback and captures intent. Formats feedback in structured way. Creates chronologically numbered feedback file. Updates task state with feedback tracking. Increments iteration counter before spawning worker. Spawns worker with updated context including feedback.

---

## Feedback Storage

### Location

`phases/<phase>/tasks/<id>/feedback/` directory within task. Chronologically numbered: `001.md`, `002.md`, `003.md`. Each feedback file is separate document.

### File Format

Structured markdown with standard sections. Created timestamp and status field. Issue description (what's wrong). Required changes (specific actions needed). Context (additional information, references to docs or standards).

### Chronological Ordering

Always read feedback in ascending order (001 → 002 → 003). Order matters for understanding evolution of requirements. Later feedback may supersede earlier feedback.

---

## Feedback Processing

### Worker Startup with Feedback

Read task state to identify pending feedback. Read all feedback files in chronological order. Incorporate feedback into work approach. Address each feedback item systematically. Update state to mark feedback as addressed. Log actions taken to address feedback.

### Incorporation Strategy

Worker reviews all pending feedback before starting work. Understands corrections needed. Makes required changes to implementation. Runs tests to verify changes. Logs which feedback items addressed and how.

### State Updates

Worker updates feedback status in task state. Changes `pending` to `addressed` when incorporated. Orchestrator validates feedback addressed before considering task complete.

---

## Feedback Tracking

### Status Values

**`pending`**: Not yet addressed by worker. Initial status when feedback created. Worker must address before task considered complete.

**`addressed`**: Worker has incorporated feedback. Changes made and verified. Work now meets corrected requirements.

**`superseded`**: No longer relevant. Newer feedback replaces or invalidates earlier feedback. Preserved for audit trail but not actionable.

### Tracking in State

Each feedback item tracked independently in task state. Includes feedback ID, creation timestamp, current status. Worker can see which feedback items need attention. Orchestrator can validate all feedback addressed.

---

## Multiple Iterations

### Feedback Rounds

Single task can have multiple feedback rounds (iteration 1: initial attempt, feedback 001 created → iteration 2: address feedback 001, discover additional issue, feedback 002 created → iteration 3: address feedback 002, complete successfully).

### Iteration Counter

Each feedback round increments iteration counter. Agent ID changes with each iteration (implementer-1, implementer-2, implementer-3). Logs show which iteration addressed which feedback. Clear audit trail of correction cycles.

### Convergence

Multiple feedback rounds expected and normal. Each round narrows gap between work and requirements. Eventually work converges to acceptable state. Orchestrator validates all feedback addressed.

**See Also**: [TASK_MANAGEMENT.md](./TASK_MANAGEMENT.md#iteration-counter)

---

## Related Documentation

- **[TASK_MANAGEMENT.md](./TASK_MANAGEMENT.md)** - Iteration counter mechanics
- **[PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md)** - Task execution and human approval
- **[AGENTS.md](./AGENTS.md)** - Worker agent coordination
- **[LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)** - Feedback files and state tracking
