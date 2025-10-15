# System of Work (sow) - Overview

**Last Updated**: 2025-10-15
**Status**: Architecture Documentation v2

---

## What is sow?

`sow` is an AI-powered framework for structured software development. It coordinates specialized AI agents through a 5-phase workflow where humans lead planning and AI executes implementation.

**Core capabilities:**
- Structured 5-phase workflow (Discovery → Design → Implementation → Review → Finalize)
- Human-AI collaboration with distinct modes (human-led planning, AI-autonomous execution)
- Multi-agent orchestration with specialized workers
- External knowledge integration from team repositories
- Zero-context resumability from disk state

---

## Core Concepts

### Two-Layer Architecture

Execution layer (`.claude/`) contains AI agents, commands, and hooks distributed via plugin. Data layer (`.sow/`) contains project knowledge, external references, and work state.

**See**: [ARCHITECTURE.md](./ARCHITECTURE.md#two-layer-architecture)

### 5-Phase Lifecycle

Projects follow a fixed 5-phase lifecycle: Discovery (optional, human-led) → Design (optional, human-led) → Implementation (required, AI-autonomous) → Review (required, AI-autonomous) → Finalize (required, AI-autonomous). Phase selection determined by truth table with scoring rubrics.

**See**: [PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md), [PHASES/](./PHASES/)

### Multi-Agent System

Orchestrator coordinates specialized workers (Researcher, Architect, Planner, Implementer, Reviewer, Documenter). Operates in subservient mode (phases 1-2) or autonomous mode (phases 3-5).

**See**: [AGENTS.md](./AGENTS.md), [ARCHITECTURE.md](./ARCHITECTURE.md#multi-agent-system)

### External References

External repositories containing knowledge or code that agents reference during work. Locally cached with automatic staleness detection.

**See**: [REFS.md](./REFS.md), [ARCHITECTURE.md](./ARCHITECTURE.md#external-references-system)

### Single Project Per Branch

One active project per repository branch. Project state committed to feature branches, deleted before merge (CI enforced).

**See**: [ARCHITECTURE.md](./ARCHITECTURE.md#single-project-constraint)

### Zero-Context Resumability

Agents resume work from disk state without conversation history. All context stored in state files, logs, and feedback.

**See**: [ARCHITECTURE.md](./ARCHITECTURE.md#zero-context-resumability), [LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)

---

## Quick Start

Install the CLI and Claude Code plugin, initialize your repository with `sow init`, create a feature branch, and start a project with `/project:new`. The orchestrator guides you through phase selection and manages execution.

**See**: [USER_GUIDE.md](./USER_GUIDE.md) for installation instructions and detailed workflows

---

## Key Terminology

| Term | Definition |
|------|------------|
| **Orchestrator** | Main coordinating agent, user-facing, has subservient and autonomous modes |
| **Worker** | Specialized agent (researcher, architect, implementer, planner, reviewer, documenter) |
| **Phase** | Fixed stage of work (discovery, design, implementation, review, finalize) |
| **Task** | Single unit of work within implementation phase |
| **Ref** | External knowledge or code reference (replaces "sinks" and "repos") |
| **Iteration** | Attempt counter for task completion |
| **Gap numbering** | Task numbering with gaps (010, 020, 030) allowing insertions |
| **Execution layer** | AI behavior (agents, commands, hooks) |
| **Data layer** | Project knowledge and state |
| **Truth table** | Decision flow for determining which phases to enable |
| **Rubric** | Scoring system for objective phase recommendations |
| **Zero-context resumability** | Ability to resume work without conversation history |
| **Subservient mode** | Orchestrator acts as assistant, human leads (discovery/design) |
| **Autonomous mode** | Orchestrator executes independently (implementation/review/finalize) |

---

## Document Navigation

### Understanding the System

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Design decisions, patterns, and architectural philosophy
- **[PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)** - The 5-phase model, truth table, and rubrics
- **[FILE_STRUCTURE.md](./FILE_STRUCTURE.md)** - Complete directory layout and organization

### Phase Details

- **[PHASES/DISCOVERY.md](./PHASES/DISCOVERY.md)** - Research and investigation phase
- **[PHASES/DESIGN.md](./PHASES/DESIGN.md)** - Architecture and planning phase
- **[PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md)** - Building features phase
- **[PHASES/REVIEW.md](./PHASES/REVIEW.md)** - Quality validation phase
- **[PHASES/FINALIZE.md](./PHASES/FINALIZE.md)** - Cleanup and PR creation phase

### System Components

- **[AGENTS.md](./AGENTS.md)** - Multi-agent system, roles, and coordination
- **[REFS.md](./REFS.md)** - External references system (knowledge and code)
- **[TASK_MANAGEMENT.md](./TASK_MANAGEMENT.md)** - Task structure, states, and execution
- **[FEEDBACK.md](./FEEDBACK.md)** - Human feedback and correction system
- **[LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)** - Logging and state management

### Reference

- **[SCHEMAS.md](./SCHEMAS.md)** - File format specifications (CUE schemas)
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - CLI command documentation
- **[USER_GUIDE.md](./USER_GUIDE.md)** - Day-to-day workflows and best practices

---

## Design Philosophy

### Human-AI Collaboration

Humans excel at setting constraints, architectural judgment, and scoping. AI excels at execution within boundaries and systematic implementation. The 5-phase model reflects this: phases 1-2 are human-led (Discovery, Design), phases 3-5 are AI-autonomous (Implementation, Review, Finalize).

### Progressive Over Waterfall

Start with minimal phases and add dynamically when needed. Use fail-forward approach (add tasks, don't revert). Human approval required for significant changes.

### Opinionated but Flexible

Strong opinions on structure: 5 fixed phases, one project per branch, specific logging format, multi-agent architecture. Flexible on phase selection, extensibility through plugins, and customizable knowledge via refs.

### Scope

`sow` coordinates active work (hours to days) with AI orchestration. It complements but does not replace project management tools (JIRA, Linear) or long-term planning systems (roadmaps, sprints).

---

## Common Workflows

**Starting**: Create feature branch, run `/project:new`, answer truth table questions, approve recommended phase plan.

**Resuming**: Switch to feature branch, run `/project:continue`. Orchestrator reads disk state and continues work.

**Feedback**: Provide corrections during any phase. Orchestrator creates feedback file, increments iteration counter, and spawns worker with updated context.

**External References**: Add knowledge or code refs via `sow refs add`. Agents automatically consult refs during work.

**Completing**: Finalize phase updates documentation, runs checks, creates PR, and deletes project folder. CI enforces cleanup before merge.

---

## Benefits

**For Developers**: Structured workflow with phase boundaries, zero-context resumability, specialized agent expertise, human control over planning with AI execution.

**For Teams**: Shared conventions via committed execution layer, centralized knowledge via refs, collaboration through branch-based project state, mandatory quality gates, consistent AI behavior.

**For AI Agents**: Focused context per worker, clear role boundaries, structured state for resumption, mode-based behavior (subservient/autonomous), comprehensive audit trails.

---

## Next Steps

1. **Learn the architecture** → Read [ARCHITECTURE.md](./ARCHITECTURE.md)
2. **Understand the lifecycle** → Read [PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)
3. **Install and try it** → Follow [USER_GUIDE.md](./USER_GUIDE.md)
4. **Explore phases** → Review [PHASES/](./PHASES/) directory
5. **Set up refs** → Check [REFS.md](./REFS.md)

---

## Project Status

This documentation represents the comprehensive design for `sow` v2 with the 5-phase human-AI collaboration model.

**Current State**: Architecture design complete, implementation in progress

**Repository**: https://github.com/your-org/sow

**Feedback**: Issues and discussions welcome
