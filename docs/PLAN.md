# Documentation Restructuring Plan

**Status**: In Progress
**Created**: 2025-10-15
**Purpose**: Roadmap for creating new component-based architecture documentation

---

## Overview

This plan outlines the restructuring of sow documentation to support the new 5-phase, human-in-loop design. The new structure organizes documentation around discrete system components with clear cross-references, avoiding duplication.

**Goal**: Create comprehensive, component-based documentation that combines:
- Original design from `docs/`
- New design specifications from `notes/`
- Clear component boundaries with cross-references

---

## New Documentation Structure

### Core Documents (12 total)

#### 1. OVERVIEW.md (Entry Point)
**Purpose**: Introduction and navigation
**Content**:
- What is sow and why it exists
- Core concepts at a glance
- Quick start guide
- Document navigation map
- Key terminology

**Sources**:
- `docs/OVERVIEW.md` (update)
- `notes/NOTES.md` (new concepts)

---

#### 2. ARCHITECTURE.md (Foundational Concepts)
**Purpose**: Core architectural principles
**Content**:
- Two-layer architecture (.claude/ vs .sow/)
- Multi-agent system principles
- Zero-context resumability
- One project per branch constraint
- Design philosophy

**Sources**:
- `docs/ARCHITECTURE.md` (existing principles)
- `notes/NOTES.md` (updated philosophy)

---

#### 3. PROJECT_LIFECYCLE.md (The 5-Phase Model)
**Purpose**: Project initialization and phase management
**Content**:
- Project initialization system
  - Truth table decision flow
  - Scoring rubrics (discovery, design, one-off)
  - Question flow and inference
- The 5 fixed phases (overview)
- Phase enablement vs execution
- State management across lifecycle
- Branch management integration

**Sources**:
- `notes/TRUTH_TABLE.md` (primary source)
- `notes/NOTES.md` (phase model)
- `docs/PROJECT_MANAGEMENT.md` (lifecycle concepts)

**New Content**: Truth table system, rubrics, 5-phase model

---

#### 4. PHASES/ Directory (5 Phase Documents)

##### 4a. PHASES/DISCOVERY.md
**Purpose**: Research and investigation phase
**Content**:
- Purpose and goals
- Entry/exit criteria
- Orchestrator role (subservient)
- Discovery types (bug, feature, docs, refactor, general)
- Researcher agent role
- Artifacts and structure
- Approval mechanics

**Sources**:
- `notes/PHASE_SPECIFICATIONS.md` (primary - discovery section)
- `notes/TRUTH_TABLE.md` (categorization)

**New Content**: Complete phase specification

---

##### 4b. PHASES/DESIGN.md
**Purpose**: Architecture and planning phase
**Content**:
- Purpose and goals
- Entry/exit criteria
- Orchestrator role (subservient)
- Design alignment subphase
- Architect agent role
- When to use/skip design
- Approval mechanics
- Artifact management

**Sources**:
- `notes/PHASE_SPECIFICATIONS.md` (primary - design section)
- `notes/TRUTH_TABLE.md` (design rubric)

**New Content**: Design alignment concept, critical thinking about design necessity

---

##### 4c. PHASES/IMPLEMENTATION.md
**Purpose**: Building features phase
**Content**:
- Purpose and goals
- Entry/exit criteria
- Orchestrator role (autonomous)
- Task breakdown process
- Planner agent role
- Fail-forward logic
- Parallel execution
- Human approval scenarios

**Sources**:
- `notes/PHASE_SPECIFICATIONS.md` (primary - implementation section)
- `docs/PROJECT_MANAGEMENT.md` (task management)

**New Content**: Planner agent, autonomous orchestrator, fail-forward mechanics

---

##### 4d. PHASES/REVIEW.md
**Purpose**: Quality validation phase
**Content**:
- Purpose and goals
- Entry/exit criteria
- Mandatory orchestrator review
- Reviewer agent role
- Review principle (does it meet expectations?)
- Loop-back mechanics
- Iteration tracking
- Human approval process

**Sources**:
- `notes/PHASE_SPECIFICATIONS.md` (primary - review section)

**New Content**: Complete review phase specification with loop-back

---

##### 4e. PHASES/FINALIZE.md
**Purpose**: Cleanup and PR creation phase
**Content**:
- Purpose and goals
- Entry/exit criteria
- Documentation subphase
- Final checks (tests, linters)
- Project deletion (mandatory)
- PR creation workflow
- Handoff to human

**Sources**:
- `notes/PHASE_SPECIFICATIONS.md` (primary - finalize section)
- `docs/PROJECT_MANAGEMENT.md` (completion workflow)

**New Content**: Complete finalize phase specification with mandatory cleanup

---

#### 5. AGENTS.md (Multi-Agent System)
**Purpose**: Agent roles and coordination
**Content**:
- Orchestrator responsibilities
  - Subservient mode (discovery/design)
  - Autonomous mode (implementation/review/finalize)
  - Mode switching
- Worker agents
  - Architect (design work)
  - Implementer (TDD coding)
  - Researcher (discovery research) - NEW
  - Planner (task breakdown) - NEW
  - Reviewer (quality checks)
  - Documenter (documentation)
- Context compilation process
- Agent coordination patterns
- Agent file format

**Sources**:
- `docs/AGENTS.md` (existing agents)
- `notes/PHASE_SPECIFICATIONS.md` (new agents)

**New Agents**: Researcher, Planner

---

#### 6. REFS.md (External References System)
**Purpose**: External knowledge and code reference system
**Content**:
- Purpose (unified system replacing sinks/repos)
- Architecture
  - Local caching strategy
  - Symlink approach (Unix) vs copy (Windows)
  - Two-index system (committed vs cache)
- CLI commands (`sow refs`, `sow cache`)
- Reference types (knowledge vs code)
- Workflows (add, init, update, status)
- Platform differences
- Orchestrator integration

**Sources**:
- `notes/refs.md` (primary source)
- `docs/ARCHITECTURE.md` (sinks/repos concepts)

**New Content**: Complete refs system replacing sinks and repos

---

#### 7. TASK_MANAGEMENT.md (Task System)
**Purpose**: Task structure and lifecycle
**Content**:
- Gap numbering system
- Task states and transitions
- Iteration counter mechanics
- Parallel execution
- Agent assignment
- Dependencies
- Abandonment (fail-forward)

**Sources**:
- `docs/PROJECT_MANAGEMENT.md` (tasks section)
- `notes/PHASE_SPECIFICATIONS.md` (implementation mechanics)

**Extracted From**: PROJECT_MANAGEMENT.md (too large, splitting out)

---

#### 8. FEEDBACK.md (Human Feedback System)
**Purpose**: Feedback mechanism for corrections
**Content**:
- Creating feedback
- Feedback storage and format
- Feedback processing by workers
- Feedback tracking in state
- Multiple feedback iterations
- Status values (pending, addressed, superseded)

**Sources**:
- `docs/PROJECT_MANAGEMENT.md` (feedback section)

**Extracted From**: PROJECT_MANAGEMENT.md (discrete component)

---

#### 9. LOGGING_AND_STATE.md (Technical Implementation)
**Purpose**: Logging and state management
**Content**:
- Structured logging format
- CLI-driven logging (`sow log`)
- Action vocabulary
- Project vs task logs
- State file management
  - Project state schema
  - Task state schema
- State maintenance responsibilities

**Sources**:
- `docs/PROJECT_MANAGEMENT.md` (logging section)
- `notes/SCHEMA_PROPOSAL.md` (state schemas)

**Extracted From**: PROJECT_MANAGEMENT.md (technical details)

---

#### 10. FILE_STRUCTURE.md (Directory Organization)
**Purpose**: Complete directory layout
**Content**:
- Complete directory tree
- `.claude/` contents (execution layer)
  - Plugin installation flow
  - Agents directory
  - Commands directory
  - Hooks and integrations
- `.sow/` contents (data layer)
  - Knowledge (committed)
  - Refs (gitignored, replaces sinks/repos)
  - Project (committed to feature branch)
- Conditional directory creation (based on phase enablement)
- Git versioning strategy

**Sources**:
- `docs/FILE_STRUCTURE.md` (existing structure)
- `notes/refs.md` (refs structure)
- `notes/SCHEMA_PROPOSAL.md` (project structure implications)

**Updates**: Replace sinks/repos with refs, update project structure for 5 phases

---

#### 11. SCHEMAS.md (Data Specifications)
**Purpose**: Complete schema reference
**Content**:
- CUE schema approach (embedded in CLI)
- Project state schema (5-phase model)
  - Fixed phase structure
  - Phase-specific fields
  - Artifact tracking
  - Review iteration
- Task state schema
- Index schemas (refs system)
- Validation rules and constraints
- Example YAML files

**Sources**:
- `docs/SCHEMAS.md` (existing schemas)
- `notes/SCHEMA_PROPOSAL.md` (new 5-phase schemas)

**Major Update**: Complete schema overhaul for 5-phase model

---

#### 12. CLI_REFERENCE.md (Command Documentation)
**Purpose**: Complete CLI command reference
**Content**:
- Core `sow` commands
  - `sow version`
  - `sow init`
  - `sow schema`
  - `sow validate`
- Logging commands
  - `sow log`
  - `sow session-info`
- Refs commands
  - `sow refs add`
  - `sow refs init`
  - `sow refs status`
  - `sow refs update`
  - `sow refs list`
  - `sow refs remove`
- Cache commands
  - `sow cache status`
  - `sow cache prune`
  - `sow cache clear`
- Slash commands (orchestrator-facing)
  - `/project:new` (replaces `/start-project`)
  - `/project:continue` (replaces `/continue`)
  - `/phase:discovery`
  - `/phase:design`
  - `/phase:implementation`
  - `/phase:review`
  - `/phase:finalize`

**Sources**:
- `docs/CLI_REFERENCE.md` (existing commands)
- `notes/refs.md` (refs commands)
- `notes/PHASE_SPECIFICATIONS.md` (slash commands)

**Major Additions**: Refs and cache commands

---

#### 13. USER_GUIDE.md (Day-to-Day Usage)
**Purpose**: Practical workflows for users
**Content**:
- Installation and setup
- Starting projects (truth table flow)
- Working with phases
  - Discovery workflows
  - Design workflows
  - Implementation workflows
  - Review workflows
  - Finalize workflows
- Providing feedback
- Using external references (refs)
- Resuming work
- Completion and cleanup
- Troubleshooting common issues

**Sources**:
- `docs/USER_GUIDE.md` (existing workflows)
- `notes/TRUTH_TABLE.md` (project initialization)
- `notes/PHASE_SPECIFICATIONS.md` (phase workflows)

**Major Updates**: New phase workflows, refs usage

---

## Migration Map

| Old Document | New Location(s) | Notes |
|--------------|-----------------|-------|
| `docs/OVERVIEW.md` | `OVERVIEW.md` | Updated for 5-phase model |
| `docs/ARCHITECTURE.md` | `ARCHITECTURE.md` | Core concepts remain, updated philosophy |
| `docs/FILE_STRUCTURE.md` | `FILE_STRUCTURE.md` | Updated for refs system |
| `docs/AGENTS.md` | `AGENTS.md` | Add researcher and planner agents |
| `docs/PROJECT_MANAGEMENT.md` | `PROJECT_LIFECYCLE.md` + `TASK_MANAGEMENT.md` + `FEEDBACK.md` + `LOGGING_AND_STATE.md` | Split into 4 component documents |
| `docs/SCHEMAS.md` | `SCHEMAS.md` | Complete rewrite for 5-phase model |
| `docs/CLI_REFERENCE.md` | `CLI_REFERENCE.md` | Add refs and cache commands |
| `docs/USER_GUIDE.md` | `USER_GUIDE.md` | Updated workflows |
| `docs/COMMANDS_AND_SKILLS.md` | `CLI_REFERENCE.md` | Merged into CLI reference |
| `docs/HOOKS_AND_INTEGRATIONS.md` | `CLI_REFERENCE.md` or `INTEGRATIONS.md` | TBD - may keep separate |
| `docs/DISTRIBUTION.md` | `DISTRIBUTION.md` or fold into `OVERVIEW.md` | TBD |
| **New from `notes/`** | | |
| `notes/PHASE_SPECIFICATIONS.md` | `PHASES/*.md` (5 docs) | Split into individual phase documents |
| `notes/TRUTH_TABLE.md` | `PROJECT_LIFECYCLE.md` | Truth table and rubrics |
| `notes/SCHEMA_PROPOSAL.md` | `SCHEMAS.md` | New 5-phase schemas |
| `notes/refs.md` | `REFS.md` | External references system |
| `notes/NOTES.md` | Various | Design notes distributed across docs |

---

## Document Dependencies (Cross-Reference Map)

This shows which documents reference which other documents to help maintain consistency:

```
OVERVIEW.md
├─→ ARCHITECTURE.md (foundational concepts)
├─→ PROJECT_LIFECYCLE.md (how to start)
└─→ USER_GUIDE.md (practical usage)

ARCHITECTURE.md
├─→ AGENTS.md (multi-agent system)
├─→ FILE_STRUCTURE.md (two-layer architecture)
└─→ LOGGING_AND_STATE.md (zero-context resumability)

PROJECT_LIFECYCLE.md
├─→ PHASES/*.md (individual phases)
├─→ AGENTS.md (orchestrator modes)
├─→ TASK_MANAGEMENT.md (task creation)
└─→ FILE_STRUCTURE.md (project directory structure)

PHASES/DISCOVERY.md
├─→ AGENTS.md (researcher agent)
├─→ REFS.md (knowledge sources)
└─→ LOGGING_AND_STATE.md (discovery artifacts)

PHASES/DESIGN.md
├─→ AGENTS.md (architect agent)
├─→ FEEDBACK.md (design alignment)
└─→ LOGGING_AND_STATE.md (design artifacts)

PHASES/IMPLEMENTATION.md
├─→ AGENTS.md (implementer, planner)
├─→ TASK_MANAGEMENT.md (task breakdown)
└─→ FEEDBACK.md (human corrections)

PHASES/REVIEW.md
├─→ AGENTS.md (reviewer agent)
└─→ PHASES/IMPLEMENTATION.md (loop-back)

PHASES/FINALIZE.md
├─→ CLI_REFERENCE.md (PR creation)
└─→ FILE_STRUCTURE.md (project deletion)

AGENTS.md
├─→ PHASES/*.md (when agents are used)
├─→ TASK_MANAGEMENT.md (agent assignment)
└─→ LOGGING_AND_STATE.md (context compilation)

REFS.md
├─→ CLI_REFERENCE.md (refs commands)
├─→ FILE_STRUCTURE.md (refs directory)
└─→ PHASES/DISCOVERY.md (knowledge sources)

TASK_MANAGEMENT.md
├─→ AGENTS.md (agent assignment)
├─→ FEEDBACK.md (feedback mechanism)
└─→ LOGGING_AND_STATE.md (task state)

FEEDBACK.md
├─→ TASK_MANAGEMENT.md (iteration counter)
└─→ LOGGING_AND_STATE.md (feedback files)

LOGGING_AND_STATE.md
├─→ CLI_REFERENCE.md (sow log command)
├─→ SCHEMAS.md (state file formats)
└─→ FILE_STRUCTURE.md (log locations)

FILE_STRUCTURE.md
├─→ SCHEMAS.md (file formats)
└─→ All phase docs (directory creation)

SCHEMAS.md
└─→ FILE_STRUCTURE.md (file locations)

CLI_REFERENCE.md
├─→ REFS.md (refs commands)
└─→ LOGGING_AND_STATE.md (log command)

USER_GUIDE.md
├─→ PROJECT_LIFECYCLE.md (starting projects)
├─→ PHASES/*.md (phase workflows)
├─→ REFS.md (using references)
└─→ CLI_REFERENCE.md (commands)
```

---

## Key Benefits of New Structure

1. **Component Isolation**: Each document focuses on a single system component
2. **Progressive Discovery**: Users can read overview → architecture → specifics
3. **Cross-References**: Phases reference agents, agents reference task management, etc.
4. **No Duplication**: Each concept explained once, referenced everywhere else
5. **Scalability**: Easy to add new phase docs or component docs
6. **Clarity**: Clear separation between "what" (components) and "how" (user guide)
7. **Phase Focus**: 5 individual phase documents allow deep dives without overwhelming single doc

---

## Documentation Principles

When writing new documents:

1. **Single Responsibility**: Each document covers one component
2. **Cross-Reference, Don't Duplicate**: Link to other docs instead of repeating
3. **Examples Over Abstraction**: Show concrete examples
4. **Progressive Disclosure**: Start simple, add complexity
5. **Consistent Format**:
   - Purpose section
   - Table of contents
   - Clear headings
   - Code examples
   - Cross-references at end
6. **Audience Awareness**:
   - OVERVIEW/USER_GUIDE: End users
   - ARCHITECTURE/PHASES: Developers and power users
   - SCHEMAS/CLI_REFERENCE: Technical reference

---

## Implementation Checklist

### Phase 1: Core Documents (Foundation)
- [ ] OVERVIEW.md
- [ ] ARCHITECTURE.md
- [ ] FILE_STRUCTURE.md

### Phase 2: Lifecycle and Phases
- [ ] PROJECT_LIFECYCLE.md
- [ ] PHASES/DISCOVERY.md
- [ ] PHASES/DESIGN.md
- [ ] PHASES/IMPLEMENTATION.md
- [ ] PHASES/REVIEW.md
- [ ] PHASES/FINALIZE.md

### Phase 3: Component Documents
- [ ] AGENTS.md
- [ ] REFS.md
- [ ] TASK_MANAGEMENT.md
- [ ] FEEDBACK.md
- [ ] LOGGING_AND_STATE.md

### Phase 4: Reference Documents
- [ ] SCHEMAS.md
- [ ] CLI_REFERENCE.md
- [ ] USER_GUIDE.md

### Phase 5: Review and Polish
- [ ] Review all cross-references
- [ ] Ensure consistency across documents
- [ ] Add navigation aids
- [ ] Proofread and edit
- [ ] Delete old `docs/` and `notes/`
- [ ] Rename `docsv2/` to `docs/`
- [ ] Update any external references to old docs

---

## Notes

- Keep `notes/orchestrator/commands/` directory for now - these will inform slash command implementation
- Some old docs may be archived rather than deleted (DISTRIBUTION.md, HOOKS_AND_INTEGRATIONS.md)
- Focus on clarity and component boundaries over comprehensive coverage in single docs
- Each phase document should be self-contained but reference other components as needed

---

## Status

**Current Phase**: Phase 1 - Planning Complete
**Next Step**: Begin drafting core documents (OVERVIEW.md, ARCHITECTURE.md, FILE_STRUCTURE.md)
