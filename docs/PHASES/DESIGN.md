# Design Phase

**Last Updated**: 2025-10-15
**Purpose**: Architecture and planning phase specification

The Design phase transforms discovery findings or human-provided notes into formal architecture decisions and structured design documentation.

---

## Table of Contents

- [Overview](#overview)
- [Purpose and Goals](#purpose-and-goals)
- [Entry Criteria](#entry-criteria)
- [Orchestrator Role](#orchestrator-role)
- [Design Alignment Subphase](#design-alignment-subphase)
- [Architect Agent](#architect-agent)
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

**Duration**: Variable (hours to days, depending on complexity)

**Output**: Formal design documentation (ADRs, design documents, specifications)

---

## Purpose and Goals

Create formal design artifacts within human-defined constraints. Synthesize discovery findings or human-provided notes into structured architecture decisions and documentation. Transform understanding into actionable design that guides implementation. Not discovery (research phase), not implementation (building phase), not evaluation (review phase).

---

## Entry Criteria

Design enabled when: coming from discovery requiring formalization, user has notes but needs formal design, significant architectural decisions must be documented, large scope requiring design breakdown (10+ tasks), or Design Worthiness Rubric scores 6-8.

Design NOT needed when: bug fixes (implement directly), small features (1-5 tasks), minor refactorings (discovery notes sufficient), or user references existing comprehensive design documents.

**Critical Assessment Required**: Orchestrator must actively evaluate whether design documents are necessary. Default to skipping unless clear justification exists.

---

## Orchestrator Role

Subservient mode: Acts as assistant, not decision-maker.

**Responsibilities**: Synthesize data (discovery notes and existing architecture), assist with fit (help human understand design integration), facilitate design alignment (refine discovery into architecture decisions), coordinate documentation (create docs directly or spawn architect agent), take notes (record design decisions continuously).

**Does NOT**: Make architectural decisions unilaterally, implement solutions, proceed to next phase without approval, create unnecessary documentation.

---

## Design Alignment Subphase

Philosophical concept (not tracked in state): Critical work refining raw discovery data into high-level architecture decisions before creating formal documents.

**Analogy**: Mix ingredients to make cake batter, but don't bake the cake yet. Discovery provides raw ingredients, design alignment mixes into batter, design documents bake the cake.

**Activities**: Review discovery artifacts, identify key design decisions, map to existing architecture, clarify constraints and requirements, define documentation scope, reach consensus on approach.

**Output**: Clear, structured notes ready for formalization into design documents.

**Method**: Conversational work between orchestrator and human, captured in notes.md.

---

## Architect Agent

Specialized agent acting as translation layer between design alignment and formal documentation.

**When to Invoke**: User requests explicitly, complex or numerous design docs required, orchestrator uncertain (prompts user for choice).

**When Orchestrator Handles Directly**: Simple short design docs, user prefers direct interaction, single ADR or small design document.

**Responsibilities**: Transform design alignment notes into formal documentation, organize output into one or more documents, structure content appropriately (ADR format, design doc format), fill in details within established constraints, produce camera-ready documentation.

**Feedback Loop**: Architect produces documents → User reviews and provides feedback → Small changes applied by orchestrator directly → Extensive changes require spawning new architect agent with feedback context.

---

## Artifacts and Structure

Artifacts stored in `.sow/project/phases/design/`: log.md (chronological conversation record), notes.md (design alignment notes, living document), requirements.md (formalized functional requirements, optional), adrs/ (Architecture Decision Records, numbered sequentially), design-docs/ (general design documents), diagrams/ (architecture diagrams, flowcharts).

**Artifact Lifecycle**: Created during design phase in project directory. Selected artifacts moved to repository during finalize phase (ADRs to repository ADR folder, architecture docs to docs/, implementation-specific notes remain in project and deleted with project).

---

## Workflow

Phase starts → Human or orchestrator describes goals from discovery → Orchestrator enters design alignment mode (reviews discovery artifacts, identifies design decisions, maps to existing architecture) → Conversational refinement (human sets constraints, orchestrator synthesizes, decisions captured in notes.md) → Documentation decision (orchestrator handles directly for simple docs, spawns architect agent for complex docs) → Document creation (formalized ADRs, design docs, specifications) → Human review and approval → Ready to exit when all documents approved.

---

## Exit Criteria

Design complete when: one or more design documents produced, documents complete and properly formatted, design fits within project constraints, design aligns with existing architecture, all design decisions documented, human has reviewed and approved all documents.

## Success Criteria

Successful when: formal design artifacts created and approved, design provides clear implementation roadmap, architectural decisions documented with rationale, design fits existing system, next steps clear for implementation, human approves phase transition.

## Phase Transition

**To Implementation**: All design documents produced and human-approved, design alignment complete. Orchestrator invokes `/phase:implementation`.

**Note**: Design phase always leads to implementation (required phase). No other transitions from design phase.

---

## Related Documentation

- **[PROJECT_LIFECYCLE.md](../PROJECT_LIFECYCLE.md)** - Design worthiness rubric and truth table
- **[AGENTS.md](../AGENTS.md)** - Architect agent details
- **[PHASES/DISCOVERY.md](./DISCOVERY.md)** - What happens before design
- **[PHASES/IMPLEMENTATION.md](./IMPLEMENTATION.md)** - What happens after design
- **[PHASES/FINALIZE.md](./FINALIZE.md)** - Moving design artifacts to repository
- **[LOGGING_AND_STATE.md](../LOGGING_AND_STATE.md)** - Design artifacts format
