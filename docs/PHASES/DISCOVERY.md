# Discovery Phase

**Last Updated**: 2025-10-15
**Purpose**: Research and investigation phase specification

The Discovery phase builds context for design or bug investigation through research, exploration, and conversation with the human.

---

## Table of Contents

- [Overview](#overview)
- [Purpose and Goals](#purpose-and-goals)
- [Entry Criteria](#entry-criteria)
- [Orchestrator Role](#orchestrator-role)
- [Researcher Agent](#researcher-agent)
- [Discovery Types](#discovery-types)
- [Artifacts and Structure](#artifacts-and-structure)
- [Workflow](#workflow)
- [Exit Criteria](#exit-criteria)
- [Success Criteria](#success-criteria)
- [Phase Transition](#phase-transition)
- [Related Documentation](#related-documentation)

---

## Overview

**Phase Type**: Optional, Human-Led

**Orchestrator Mode**: Subservient (acts as assistant, human leads)

**Duration**: Variable (minutes to hours, depending on complexity)

**Output**: Research findings, synthesized notes, and key decisions

---

## Purpose and Goals

Build context for design or bug investigation through research-oriented exploration and collaboration between human and orchestrator. Documented findings inform future phases. Not planning (design phase), not implementing (implementation phase), not decision-making alone (human sets constraints), not open-ended (has clear exit criteria).

---

## Entry Criteria

Discovery enabled when: user provides minimal context, bug needs investigation (cause unknown), feature requires domain knowledge, user explicitly requests discovery, or Discovery Worthiness Rubric scores 6-8.

Discovery NOT needed when: user references existing design doc, has detailed notes, or explicitly requests to skip.

---

## Orchestrator Role

Subservient mode: Acts as assistant, not decision-maker.

**Responsibilities**: Ask clarifying questions, point out inconsistencies, help brainstorm, make suggestions, take notes (synthesize conversation into notes.md), log conversation (chronological record in log.md).

**Does NOT**: Make architectural decisions, implement solutions, unilaterally decide research direction, or advance to next phase without approval.

---

## Researcher Agent

Performs focused research to ground discussions. Invoked when orchestrator suggests or human requests research.

**Research Sources**: Refs (.sow/refs/), linked repositories, local codebase, web search.

**Output**: Summarized findings stored in `.sow/project/phases/discovery/research/NNN-topic.md`. Structure includes summary, findings, sources, and recommendations.

---

## Discovery Types

Five common types with focused slash commands:

**Bug Investigation** (`/phase:discovery:bug`): Root cause analysis, reproduction steps, affected systems, impact assessment. Activities: review logs, examine recent changes, test reproduction, identify code paths.

**Feature Exploration** (`/phase:discovery:feature`): Problem space understanding, user needs, existing solutions (internal/external), competitive analysis. Activities: research implementations, compare libraries, review feedback.

**Documentation Gap Analysis** (`/phase:discovery:docs`): Identify gaps between code and documentation, find missing or outdated content. Activities: read docs, review implementation, identify discrepancies.

**Refactoring Analysis** (`/phase:discovery:refactor`): Understand why code is messy, analyze responsibilities and coupling, identify patterns/anti-patterns, assess risks. Activities: analyze structure, identify code smells, assess test coverage, review change frequency.

**General Discovery** (`/phase:discovery:general`): Catch-all for mixed needs, exploratory work, or learning unfamiliar codebase.

---

## Artifacts and Structure

Artifacts stored in `.sow/project/phases/discovery/`: log.md (chronological conversation record), notes.md (orchestrator's synthesized notes, living document), decisions.md (key conclusions and rationale), research/ (researcher agent outputs, numbered sequentially).

---

## Workflow

Phase starts → Human describes goals → Orchestrator categorizes (invokes appropriate discovery type) → Conversation and research (orchestrator suggests research, spawns researcher, summarizes findings) → Human provides constraints → Orchestrator synthesizes (updates notes.md and decisions.md) → Iteration continues → Ready to exit when solid understanding achieved.

---

## Exit Criteria

Discovery complete when: human has enough context, key questions answered, next steps clear, artifacts approved by human. Orchestrator clarifies whether design phase is needed or going straight to implementation.

## Success Criteria

Successful when: human understands problem space, key questions answered, sufficient context for next phase, artifacts approved, next steps clear, human approves phase transition.

## Phase Transition

**To Design**: When human approves artifacts and indicates design doc needed for complex system. Orchestrator invokes `/phase:design`.

**To Implementation (Skip Design)**: When human approves artifacts and indicates sufficient detail exists without formal design. Orchestrator invokes `/phase:implementation`.

---

## Related Documentation

- **[PROJECT_LIFECYCLE.md](../PROJECT_LIFECYCLE.md)** - Discovery worthiness rubric and truth table
- **[AGENTS.md](../AGENTS.md)** - Researcher agent details
- **[PHASES/DESIGN.md](./DESIGN.md)** - What happens after discovery
- **[PHASES/IMPLEMENTATION.md](./IMPLEMENTATION.md)** - Alternative path after discovery
- **[REFS.md](../REFS.md)** - External knowledge sources for research
- **[LOGGING_AND_STATE.md](../LOGGING_AND_STATE.md)** - Discovery artifacts format
