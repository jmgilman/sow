# sow Agent System

**Last Updated**: 2025-10-15
**Purpose**: Multi-agent system, roles, and coordination

This document describes how `sow` uses multiple specialized AI agents to coordinate software development work through a hierarchical orchestrator-worker pattern.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Orchestrator Agent](#orchestrator-agent)
- [Worker Agents](#worker-agents)
- [Context Compilation](#context-compilation)
- [Agent Coordination](#agent-coordination)
- [Agent File Format](#agent-file-format)
- [Related Documentation](#related-documentation)

---

## Architecture Overview

### Hierarchical Pattern

Orchestrator coordinates specialized workers. Users interact with orchestrator. Workers spawned by orchestrator using Task tool. Workers report back to orchestrator. Clear separation: coordination vs execution.

### Key Principles

**One Orchestrator**: User-facing agent (main Claude Code session visible to user), coordinates all work, manages project state, switches modes based on phase.

**Multiple Workers**: Spawned by orchestrator, specialist expertise, separate context windows, receive curated context, report results back.

**Explicit Delegation**: Orchestrator assigns tasks to specific agents, agent type stored in task state, predictable auditable workflow.

---

## Orchestrator Agent

### Role

Main interface to `sow`. Agent users see and interact with directly. Mode switches based on active phase.

### Two Operating Modes

**Subservient Mode** (Discovery, Design phases): Acts as assistant, human leads. Asks questions, points out inconsistencies, helps brainstorm, makes suggestions, takes notes continuously, logs conversation. Never makes unilateral decisions. Waits for approval before advancing phases.

**Autonomous Mode** (Implementation, Review, Finalize phases): Executes independently within boundaries. Makes implementation decisions within constraints, updates state automatically, spawns workers without approval. Requests approval only for adding new tasks, going back to previous phases, or blocking issues.

### Responsibilities by Mode

**Subservient Mode**: Check for existing project and prompt continuation, ask clarifying questions, facilitate design alignment, synthesize conversation into notes, spawn specialist workers when requested (researcher for discovery, architect for design), log all conversations chronologically, request human approval for phase transitions.

**Autonomous Mode**: Read project state and identify next action, compile context for workers, spawn workers with curated context (planner for large task breakdowns, implementer for coding, reviewer for validation, documenter for documentation), update project state after worker completion, mark tasks completed, manage task iteration counters, handle fail-forward task additions (with approval), coordinate phase transitions.

**Trivial Tasks**: Handle directly without project structure (typo fixes, simple edits). Exception to delegation pattern for efficiency.

### Context Compilation Responsibility

Orchestrator acts as context compiler: reads project requirements, refs index, knowledge documents, and task descriptions; filters for task-relevant content; packages into task description with file references in state; workers receive minimal targeted context.

**See Also**: [Context Compilation](#context-compilation)

---

## Worker Agents

### Agent Roster

Six specialized worker agents with focused expertise.

#### Researcher

**Role**: Discovery research and investigation.

**When Invoked**: Discovery phase when orchestrator suggests or human requests research.

**Responsibilities**: Perform focused research from refs, linked repositories, local codebase, web search. Summarize findings for orchestrator/human review. Ground discussions in real sources.

**Output**: Research reports stored in discovery phase (numbered sequentially with topic in filename).

**Typical Phase**: Discovery

**See Also**: [PHASES/DISCOVERY.md](./PHASES/DISCOVERY.md#researcher-agent)

---

#### Architect

**Role**: System design and architecture.

**When Invoked**: Design phase when orchestrator or human determines design documents needed. Used directly by orchestrator for simple docs, spawned for complex documentation.

**Responsibilities**: Transform design alignment notes into formal documentation, create Architecture Decision Records (ADRs), write design documents and API specifications, structure content appropriately within established constraints, produce camera-ready documentation.

**Output**: ADRs, design documents, API specifications in design phase.

**Typical Phase**: Design

**See Also**: [PHASES/DESIGN.md](./PHASES/DESIGN.md#architect-agent)

---

#### Planner

**Role**: Implementation task breakdown for large projects.

**When Invoked**: Implementation phase for very large projects (10+ tasks expected), complex task dependencies, orchestrator uncertain about breakdown, or user requests explicitly.

**Responsibilities**: Analyze discovery/design artifacts, suggest logical task breakdown, consider dependencies and ordering, propose parallel vs sequential tasks, output implementation plan document.

**Output**: `implementation-plan.md` with suggested task breakdown (orchestrator reviews and creates actual state file).

**Typical Phase**: Implementation

**Key Distinction**: Planner does NOT write state file directly (orchestrator controls state).

**See Also**: [PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md#planner-agent)

---

#### Implementer

**Role**: Code implementation with Test-Driven Development.

**When Invoked**: Implementation phase for all coding tasks (assigned during task breakdown).

**Responsibilities**: Implement features using TDD approach, write tests first then implementation, fix bugs (write failing test then fix), integrate with existing code, refactor for functionality, ensure test coverage.

**Output**: Production code, unit tests, task logs.

**Typical Phase**: Implementation

**TDD Enforcement**: Agent prompt requires test-first development (red → green → refactor cycle).

**See Also**: [PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md#task-execution)

---

#### Reviewer

**Role**: Quality validation.

**When Invoked**: Review phase (optional assistance for orchestrator, invoked for large/complex changes or when orchestrator uncertain).

**Responsibilities**: Review requirements (discovery artifacts, design documents, original intent), review implementation (examine file changes, review task logs, check test coverage, analyze code quality), compare and validate (implementation matches requirements, identify gaps/issues, assess quality), document findings (review report with specific issues and recommendations).

**Output**: Review report document (orchestrator presents to human).

**Typical Phase**: Review

**See Also**: [PHASES/REVIEW.md](./PHASES/REVIEW.md#reviewer-agent)

---

#### Documenter

**Role**: Documentation maintenance.

**When Invoked**: Finalize phase when documentation updates needed, or other phases when explicit documentation tasks assigned.

**Responsibilities**: Update README files, write inline documentation, update API documentation, maintain architectural docs, add code comments, create usage examples.

**Output**: Updated documentation files.

**Typical Phase**: Finalize (documentation subphase)

**See Also**: [PHASES/FINALIZE.md](./PHASES/FINALIZE.md#documentation-subphase)

---

### Why Multiple Agents?

**Problem**: Single agent with all capabilities results in massive system prompt covering all scenarios, context explosion (style guides + testing + architecture + deployment), poor performance at each individual task, constant compaction/restarts.

**Solution**: Focused specialized agents provide shorter more effective prompts, distinct context needs per agent, better performance (targeted context), easy extensibility (just Markdown files).

---

## Context Compilation

### Problem

Workers need minimal curated context to avoid window bloat.

### Process

**Orchestrator Gathers** (broad collection): Task requirements (description.md), refs index, knowledge documents, project context, previous task logs.

**Orchestrator Filters** (selective curation): Which refs apply for this specific task? Which knowledge docs relevant? Which code examples needed? Which project decisions matter?

**Orchestrator Packages** (structured handoff): Creates task description.md (requirements, acceptance criteria), creates task state.yaml with references list, all file paths relative to `.sow/` root.

**Worker Receives** (focused context): Reads state.yaml (iteration, assigned agent, references), reads description.md (what to do), reads all referenced files (refs, knowledge, code), reads feedback if any (human corrections). Worker has everything needed, nothing extra.

### Benefits

Performance (workers don't wade through irrelevant info), accuracy (focused context equals better results), efficiency (no context window bloat), scalability (handles large knowledge bases).

---

## Agent Coordination

### Delegation Pattern

**Orchestrator spawns workers**: Uses Task tool with agent type and context. Workers run independently. Workers report results back. Orchestrator updates state and continues.

**User Experience**: User sits in front of orchestrator (main session), sees orchestrator decisions and summaries, workers run in background, results communicated through orchestrator.

**Visibility**: Normal Claude Code interruption works, session history shows orchestrator decisions, task logs show worker actions, post-facto debugging available.

### Error Correction

**User provides feedback to orchestrator**: Orchestrator creates feedback file for current task, increments iteration counter, spawns worker with feedback context.

**Worker reads feedback**: Understands correction, makes changes, updates feedback status to addressed.

### Task Routing

**Simple tasks**: Orchestrator handles directly (avoids overhead, faster for trivial changes, no project structure needed).

**Complex tasks**: Spawn workers (better expertise, separate context, auditable via logs).

**Modern context windows** (200k+): Simple tasks safe to handle inline without delegation overhead.

---

## Agent File Format

### Structure

Agents defined as Markdown files with YAML frontmatter. Located at `.claude/agents/<agent-name>.md`. Installed via Claude Code Plugin.

### Frontmatter Fields

**`name`** (required): Agent identifier used when spawning via Task tool. Lowercase hyphen-separated.

**`description`** (required): When this agent should be used. Helps orchestrator choose appropriate worker. Brief action-oriented.

**`tools`** (optional): Comma-separated list of allowed tools. If omitted agent inherits all tools.

**`model`** (optional): Which Claude model to use (inherit, sonnet, opus, haiku). Default inherit (same as orchestrator).

### Body Content

Agent system prompt defining role, responsibilities, capabilities, guidance. Focused instructions for specialized task type.

---

## Related Documentation

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Multi-agent system rationale and design
- **[PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)** - Orchestrator mode switching across phases
- **[PHASES/DISCOVERY.md](./PHASES/DISCOVERY.md)** - Researcher agent usage
- **[PHASES/DESIGN.md](./PHASES/DESIGN.md)** - Architect agent usage
- **[PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md)** - Planner and implementer agents
- **[PHASES/REVIEW.md](./PHASES/REVIEW.md)** - Reviewer agent usage
- **[PHASES/FINALIZE.md](./PHASES/FINALIZE.md)** - Documenter agent usage
- **[TASK_MANAGEMENT.md](./TASK_MANAGEMENT.md)** - Agent assignment to tasks
- **[LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)** - Context compilation details
