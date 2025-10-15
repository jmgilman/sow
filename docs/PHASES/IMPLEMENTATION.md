# Implementation Phase

**Last Updated**: 2025-10-15
**Purpose**: Building features phase specification

The Implementation phase executes work to achieve project goals through task breakdown and systematic completion by implementer agents.

---

## Table of Contents

- [Overview](#overview)
- [Purpose and Goals](#purpose-and-goals)
- [Entry Criteria](#entry-criteria)
- [Orchestrator Role](#orchestrator-role)
- [Initial Task Breakdown](#initial-task-breakdown)
- [Planner Agent](#planner-agent)
- [Task Execution](#task-execution)
- [Human Approval Scenarios](#human-approval-scenarios)
- [Fail-Forward Logic](#fail-forward-logic)
- [Artifacts and Structure](#artifacts-and-structure)
- [Exit Criteria](#exit-criteria)
- [Success Criteria](#success-criteria)
- [Phase Transition](#phase-transition)
- [Related Documentation](#related-documentation)

---

## Overview

**Phase Type**: Required, AI-Autonomous

**Orchestrator Mode**: Autonomous (executes independently within boundaries)

**Duration**: Variable (hours to days, depending on scope)

**Output**: Completed implementation, passing tests, committed code

---

## Purpose and Goals

Execute implementation work to achieve project goals. Break down requirements into discrete tasks and complete them systematically using implementer agents. Deliver working, tested code that meets original intent. Not planning (design phase), not validating (review phase), not investigating (discovery phase).

---

## Entry Criteria

Implementation always happens (required phase, cannot be skipped).

**Entry Paths**: Directly from project start (user has clear requirements, no discovery/design needed), from discovery phase (skip design, implementation plan clear from discovery), from design phase (design documents provide implementation roadmap), from review phase (loop-back to address issues).

---

## Orchestrator Role

Autonomous mode: Executes independently within established boundaries.

**Responsibilities**: Create initial task breakdown (with human approval), spawn implementer agents with task context, monitor task progress, mark tasks completed, handle implementer issues, add tasks when needed (fail-forward with approval), update project state automatically.

**Maximum Autonomy For**: Normal task execution flow, marking tasks completed, moving to next task, re-invoking implementers with feedback, adjusting task descriptions, iteration management.

**Human Approval Required For**: Adding new tasks (fail-forward), going back to previous phases, implementer stuck needing human help.

---

## Initial Task Breakdown

Orchestrator creates implementation plan from inputs (discovery artifacts, design documents, human requirements).

**Task Breakdown Process**: Analyze inputs (review artifacts, understand constraints and goals), identify work units (what needs building/changing, logical chunks, dependencies), create task list (gap numbering 010/020/030, clear actionable descriptions, identify dependencies, mark parallelizable tasks, assign to implementer agent), validate plan (accomplishes goal, appropriately sized tasks, correct dependencies, nothing missing).

**Human Approval Required**: User must approve initial task breakdown before work begins.

**Task Properties**: Gap-numbered IDs (allows insertions), clear descriptions, dependency tracking, parallel execution flags, agent assignment (typically implementer).

---

## Planner Agent

Optional specialist for large project task breakdown.

**When to Use**: Very large projects (10+ tasks expected), complex task dependencies, orchestrator uncertain about breakdown, user requests explicitly.

**When NOT to Use**: Small projects (1-5 tasks), clear straightforward breakdown, orchestrator confident.

**Responsibilities**: Analyze discovery/design artifacts, suggest logical task breakdown, consider dependencies and ordering, propose parallel vs sequential tasks, output separate planning document.

**Output**: Planner produces `implementation-plan.md` with suggested task breakdown. Orchestrator reviews, adjusts if needed, creates actual state file, presents to human for approval.

**Key Distinction**: Planner does NOT write state file directly (orchestrator controls state).

---

## Task Execution

Maximum orchestrator autonomy: Normal flow requires no human approval.

**Execution Flow**: Identify next pending task (or parallel set), spawn implementer agent(s) with task context, implementer completes work, implementer reports back, orchestrator marks task completed, orchestrator moves to next task, repeat until all tasks done.

**Parallel Execution**: Orchestrator can spawn multiple implementers simultaneously for tasks marked as parallelizable.

**Implementer Issues**: Orchestrator has autonomy to re-invoke implementer with additional feedback, adjust task description, mark task as completed if good enough, continue iterating with implementer.

**No Human Approval Needed For**: Marking tasks completed, moving to next task, re-invoking implementer with feedback, normal task execution flow.

---

## Human Approval Scenarios

Orchestrator does NOT have full autonomy in specific scenarios.

**Implementer Stuck**: Problem occurred that orchestrator cannot resolve, needs human input to unblock. Orchestrator prompts human with situation explanation and waits for guidance.

**Fail-Forward - Add New Tasks**: Problem reveals new work needed, orchestrator wants to add tasks to address it. Orchestrator prompts human with task description and rationale, waits for approval before adding.

**Go Back to Previous Phase**: Problem reveals design is flawed OR more discovery needed. Orchestrator wants to return to discovery or design phase. Orchestrator prompts human with explanation and waits for approval before transitioning.

**Principle**: Human approval only for real problems, not normal task flow.

---

## Fail-Forward Logic

When work reveals new requirements, add tasks and keep moving forward.

**Process**: Problem/gap discovered during implementation → Orchestrator identifies additional work needed → Orchestrator requests approval to add new task(s) → If approved: add tasks to state and continue → Original task may be marked as abandoned if no longer relevant.

**Task Abandonment**: When original task no longer applicable, mark as "abandoned" in state. Does NOT count toward "all tasks complete". Log explains abandonment rationale.

**Purpose**: Adapt to discovered requirements without reverting progress. Maintain forward momentum while documenting path changes.

---

## Artifacts and Structure

Artifacts stored in `.sow/project/phases/implementation/`: log.md (orchestrator coordination log), implementation-plan.md (planner output if used), tasks/ (task directories with gap-numbered IDs).

**Task Directory Structure** (`tasks/NNN/`): state.yaml (task metadata, iteration, status, dependencies), description.md (task requirements and context), log.md (implementer action log), feedback/ (human corrections if any, numbered chronologically).

---

## Exit Criteria

Implementation complete when: all non-abandoned tasks completed, code committed to git, tests passing, no blocking issues remain.

**Note**: "All tasks complete" does NOT include abandoned tasks. Abandoned tasks are documented but don't block completion.

## Success Criteria

Successful when: all required tasks completed successfully, code changes committed and pushed, tests passing, implementer agents completed work within boundaries, ready for review phase.

## Phase Transition

**To Review Phase**: All tasks completed successfully. Automatic transition (no human approval required). Orchestrator invokes `/phase:review`.

**Loop-Back from Review**: Human can request returning to implementation after review if issues found. Review phase manages this transition.

---

## Related Documentation

- **[PROJECT_LIFECYCLE.md](../PROJECT_LIFECYCLE.md)** - Implementation as required phase
- **[AGENTS.md](../AGENTS.md)** - Implementer and planner agent details
- **[PHASES/DISCOVERY.md](./DISCOVERY.md)** - Possible entry point
- **[PHASES/DESIGN.md](./DESIGN.md)** - Possible entry point
- **[PHASES/REVIEW.md](./REVIEW.md)** - What happens after implementation
- **[TASK_MANAGEMENT.md](../TASK_MANAGEMENT.md)** - Task structure and lifecycle
- **[FEEDBACK.md](../FEEDBACK.md)** - Human feedback system
- **[LOGGING_AND_STATE.md](../LOGGING_AND_STATE.md)** - Implementation artifacts format
