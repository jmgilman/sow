# Project Lifecycle

**Last Updated**: 2025-10-15
**Purpose**: Complete guide to project initialization and the 5-phase model

This document describes how projects are created, initialized, executed, and completed in the `sow` system.

---

## Table of Contents

- [Overview](#overview)
- [Project Initialization](#project-initialization)
  - [Entry Point](#entry-point)
  - [Truth Table Decision Flow](#truth-table-decision-flow)
  - [Scoring Rubrics](#scoring-rubrics)
  - [Example Walkthroughs](#example-walkthroughs)
- [The 5-Phase Model](#the-5-phase-model)
  - [Phase Overview](#phase-overview)
  - [Phase Enablement vs Execution](#phase-enablement-vs-execution)
  - [Human-Led vs AI-Autonomous](#human-led-vs-ai-autonomous)
- [Branch Management](#branch-management)
- [Project Execution](#project-execution)
- [Project Completion](#project-completion)
- [State Management](#state-management)
- [Related Documentation](#related-documentation)

---

## Overview

Every `sow` project follows a structured lifecycle:

1. **Initialization** - Determine which phases are needed (truth table)
2. **Execution** - Work through enabled phases sequentially
3. **Completion** - Finalize, cleanup, and create PR

**Key Principle**: The orchestrator gracefully leads the human through questions to infer the right phase plan, then **asks for confirmation** - never decides unilaterally.

---

## Project Initialization

### Entry Point

Command `/project:new` requires repository initialization, feature branch (not main), and no existing project. Orchestrator validates branch, enters truth table decision flow, determines phase plan with human approval, creates structure, begins first enabled phase.

---

### Truth Table Decision Flow

Orchestrator asks questions, interprets answers, recommends a phase plan, then requests approval.

**Question 1: What Are You Trying to Accomplish?**
Purpose: Understand work type (bug, feature, refactor, etc.) to infer phase needs.

**Question 2: Existing Context Assessment**
Purpose: Determine if discovery/design can be skipped based on available context (notes, design docs, requirements).

**Question 3: Discovery Phase Decision**
Asked when limited context exists. Orchestrator recommends discovery or applies Discovery Worthiness Rubric if uncertain.

**Question 4: Design Phase Decision**
Asked after discovery completes or if user has context but no formal design. Orchestrator recommends design or applies Design Worthiness Rubric if uncertain.

**Final Confirmation**
Orchestrator presents recommended phase plan with rationale. User approves or modifies. Human has final authority on phase enablement.

---

### Scoring Rubrics

Objective guidance for phase recommendations when orchestrator is uncertain.

#### Discovery Worthiness Rubric

Four criteria scored 0-2 each: Context Availability, Problem Clarity, Codebase Familiarity, Research Needs.
- **0-2**: NOT warranted (skip)
- **3-5**: OPTIONAL (ask user)
- **6-8**: RECOMMENDED (suggest strongly)

#### Design Worthiness Rubric

Four criteria scored 0-2 each: Scope Size, Architectural Impact, Integration Complexity, Design Decisions. Bug fixes get -3 penalty (minimum 0).
- **0-2**: NOT warranted (skip)
- **3-5**: OPTIONAL (ask user)
- **6-8**: RECOMMENDED (suggest strongly)

#### One-Off vs Project Rubric

Four binary criteria: Single Action, No Complexity, No Tracking Value, Immediate Completion (<5 min).
- **0-2**: Project appropriate (proceed)
- **3-4**: One-off recommended (suggest to user)

Orchestrator suggests but never overrides user choice.

---

### Example Walkthroughs

**Bug Fix (No Context)**: Discovery rubric scores 4 (optional). User approves discovery. Design skipped (bug fixes don't need design). Result: Discovery + Implementation + Review + Finalize.

**Large Feature (No Context)**: Orchestrator recommends discovery for research. After discovery, design rubric scores 8 (recommended). Result: Discovery + Design + Implementation + Review + Finalize.

**Implement Existing Design**: User has comprehensive design doc. Orchestrator reads doc, confirms it's complete, skips both discovery and design. Result: Implementation + Review + Finalize.

**Simple Feature**: One-off rubric scores 3 (consider one-off). User prefers project structure but skips discovery and design. Result: Implementation + Review + Finalize.

---

## The 5-Phase Model

### Phase Overview

All projects use the same 5 phases: Discovery (optional, human-led) → Design (optional, human-led) → Implementation (required, AI-autonomous) → Review (required, AI-autonomous) → Finalize (required, AI-autonomous).

Fixed structure: All 5 phases exist in state. Variable execution: `enabled` flag controls execution.

### Phase Enablement vs Execution

All phases exist in state with `enabled: true/false` flag. Orchestrator skips disabled phases. Benefits: predictable structure, simple validation, clear distinction between existence and execution.

---

### Human-Led vs AI-Autonomous

**Human-Led (Subservient Mode)**: Discovery and Design. Orchestrator acts as assistant, human drives decisions. Never makes unilateral decisions. Waits for approval before advancing. Prevents over-engineering in planning.

**AI-Autonomous (Autonomous Mode)**: Implementation, Review, and Finalize. Orchestrator executes independently within established boundaries. Updates state automatically. Requests approval only for adding tasks, going back to previous phases, or blocking issues.

### Phase Status Values

Status values: `skipped` (disabled), `pending` (enabled, not started), `in_progress` (executing), `completed` (finished). Transitions: disabled phases stay skipped; enabled phases go pending → in_progress → completed.

---

## Branch Management

### Branch Validation

Before creation, orchestrator validates: errors on main/master (offers to create feature branch), proceeds on feature branch with no project, errors on feature branch with existing project (offers continue or new branch).

### Branch Name Suggestions

Orchestrator infers branch name from description (e.g., "Add authentication" → feat/add-authentication, "Fix login bug" → fix/login-bug). User can override.

### Branch Tracking

Project state includes branch name. On resume, orchestrator verifies current branch matches project.branch in state, warns on mismatch.

---

## Project Execution

### Phase Transitions

**Automatic**: Implementation → Review (when all tasks complete)

**Manual Approval Required**: Discovery → Design/Implementation, Design → Implementation, Review → Finalize

**Loop-Back Allowed**: Review → Implementation (if issues found), or return to previous phases with approval

### Active Phase Tracking

One phase active at a time. Project state tracks `active_phase` field. Phase status shows progression: completed, skipped, in_progress, or pending.

---

## Project Completion

### Finalize Phase Completion

Finalize updates documentation, runs final checks, deletes .sow/project/, commits deletion, creates PR.

**Committed**: Code changes, documentation updates, design artifacts moved to .sow/knowledge/

**Deleted**: .sow/project/ (entire directory including all phase artifacts and task logs)

**Why Delete**: Projects are per-branch. Main represents stable code. CI enforces cleanup.

### Pull Request Workflow

Review PR, address feedback, merge (recommend squash), delete feature branch. Main branch never contains .sow/project/.

**See Also**: [PHASES/FINALIZE.md](./PHASES/FINALIZE.md)

---

## State Management

### Project State File

Location: `.sow/project/state.yaml`

Purpose: Single source of truth for project metadata and phase status

Key sections: project metadata (name, branch, description, timestamps), active_phase indicator, phases with enablement/status/timestamps, implementation tasks with status/agent/iteration.

**See Also**: [SCHEMAS.md](./SCHEMAS.md) for complete schema, [LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md) for state management

---

## Related Documentation

### Phase Details
- **[PHASES/DISCOVERY.md](./PHASES/DISCOVERY.md)** - Research and investigation
- **[PHASES/DESIGN.md](./PHASES/DESIGN.md)** - Architecture and planning
- **[PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md)** - Building features
- **[PHASES/REVIEW.md](./PHASES/REVIEW.md)** - Quality validation
- **[PHASES/FINALIZE.md](./PHASES/FINALIZE.md)** - Cleanup and PR creation

### Related Concepts
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Human-AI collaboration model
- **[AGENTS.md](./AGENTS.md)** - Orchestrator mode switching
- **[TASK_MANAGEMENT.md](./TASK_MANAGEMENT.md)** - Task structure and execution
- **[LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)** - State file management
- **[SCHEMAS.md](./SCHEMAS.md)** - Project state schema specification
