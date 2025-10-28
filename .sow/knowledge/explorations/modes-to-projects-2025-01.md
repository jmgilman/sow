# Modes-to-Projects Consolidation Exploration Summary

**Date:** January 2025
**Branch:** explore/modes-to-projects
**Status:** Research complete

## Context

This exploration investigated consolidating sow's three operating modes (exploration, design, breakdown) into the unified project system. The goal: simplify the framework, enable state machine workflows, improve context engineering through continual prompting, and streamline UX—all while preserving mode flexibility.

## What We Researched

- Translation of mode index schemas (exploration, design, breakdown) to project phase schemas
- Schema gaps preventing clean translation
- Required modifications to phase/task/artifact schemas
- State machine integration for workflow guidance
- Feature preservation analysis across all modes

## Key Findings

### All Modes Translate Cleanly to Project Phases

With six targeted schema improvements, every mode feature maps naturally to project constructs without semantic mismatches or feature loss.

#### Translation Mappings

**Exploration mode**:
- Topics → tasks (parking lot becomes task list)
- Files → artifacts (with optional approval)
- Journal → log entries (reusing existing logging infrastructure)
- Natural state flow: gathering → researching → summarizing → completed

**Design mode**:
- Inputs → Phase.inputs (new field, reuses Artifact type)
- Outputs → artifacts (with required approval)
- Perfect semantic match for approval workflow
- Natural state flow: drafting → reviewing → approved → completed

**Breakdown mode**:
- Work units → tasks (including dependencies)
- GitHub metadata → task.metadata
- Substatus tracking → task.metadata.substatus
- Natural state flow: decomposing → documenting → approving → publishing → completed

### Six Schema Improvements Enable Consolidation

All changes are minimal, targeted, and preserve backward compatibility (except #6):

1. **`Artifact.approved?: bool`** (optional) - Enables both exploration files (no approval) and design outputs (approval required) to use the same artifact concept.

2. **`Phase.inputs?: [...#Artifact]`** (new field) - Design and breakdown track input sources (explorations, files, references) that inform the phase's work. Reusing Artifact type provides structured tracking with descriptions and tags. Clear separation: inputs inform work, artifacts are outputs.

3. **`Task.refs?: [...#Artifact]`** (new field) - Exploration topics track related files created during research. Tasks need reference tracking for context compilation. Reusing Artifact type provides richer metadata than simple string paths.

4. **`Task.metadata?: {[string]: _}`** (new field) - Breakdown stores GitHub issue metadata (URLs, numbers) and substatus tracking. Matches existing Phase.metadata and Artifact.metadata patterns. Provides extensibility for future task-specific data.

5. **Use logging for journaling** - Exploration's session journal tracks decisions, insights, questions. Instead of separate journal field, reuse existing logging infrastructure with `action=journal`. Command: `sow agent log --action journal --result <type> --notes "<content>"`. Maintains chronological ordering, zero-context recovery, and single source of truth for orchestrator actions.

6. **`Phase.status: string`** (remove constraint) - Current schema constrains all phases to "pending | in_progress | completed | skipped". Project types should define valid statuses per phase, enabling natural workflow progressions. Example: exploration phases use "gathering | researching | summarizing | completed" instead of generic states. Project type schemas constrain statuses via CUE. **Breaking change** but justifiable—unlocks meaningful state machine design.

### State Machine Benefits

Removing Phase.status constraint enables natural workflow states that guide orchestrator behavior:

**Exploration state flow**:
```
gathering topics
    ↓ (guard: min topics added, user approves)
researching topics
    ↓ (guard: all topics completed)
summarizing findings
    ↓ (guard: summary created)
completed
```

Orchestrator knows exactly what to do at each state: focus on topic discovery during gathering, work through topics during researching, create comprehensive summary during summarizing.

**Design state flow**:
```
drafting
    ↓ (guard: all outputs created)
reviewing
    ↓ (guard: all artifacts approved)
approved
    ↓ (guard: artifacts moved to target locations)
completed
```

State-specific prompts guide artifact creation, approval workflow, and finalization.

**Breakdown state flow**:
```
decomposing
    ↓ (guard: all work units proposed)
documenting
    ↓ (guard: all documents created)
approving
    ↓ (guard: all units approved)
publishing
    ↓ (guard: all issues published)
completed
```

Clear progression from high-level decomposition to published GitHub issues.

### Feature Preservation

**Features that translate cleanly**:
- ✓ Exploration topic parking lot (phase tasks with dependencies)
- ✓ Exploration file tracking (artifacts with optional approval)
- ✓ Exploration session journal (log entries with action=journal)
- ✓ Design input tracking (Phase.inputs reusing Artifact type)
- ✓ Design output approval workflow (artifacts with required approval)
- ✓ Breakdown work unit dependencies (task dependencies)
- ✓ Breakdown GitHub issue tracking (task metadata)
- ✓ Breakdown substatus workflow (task metadata.substatus)
- ✓ State machine guards (check task/artifact status, semantic states)

**Features requiring modification**: None. All mode features preserved with proposed schema changes.

**Features lost**: None identified.

### Architecture Benefits

1. **Consolidation** - Single foundational concept (projects) instead of separate modes. Easier comprehension, more streamlined code, broader standardization.

2. **State machines** - Natural state flows with meaningful guards guide orchestrator through phase lifecycles.

3. **Context engineering** - State changes trigger continual prompting instead of relying on single large prompt that gets "forgotten" in long sessions.

4. **Streamlined UX** - Developers avoid inconsistencies and gotchas between modes and projects.

5. **Flexibility preserved** - Each mode becomes a project type with relaxed guidelines. Exploration/design get flexibility they need; standard projects maintain structure.

### Migration Path

**Non-breaking approach** (gradual):
1. Create new project type schemas (ExplorationProjectState, DesignProjectState, BreakdownProjectState)
2. Implement CLI commands working with both old indexes and new project states
3. Add migration command: `sow migrate-mode <mode>` reads old index, creates project structure
4. Deprecate mode commands with warnings, keep functional during transition
5. Remove mode code after deprecation period

**Breaking change approach** (faster):
1. Create project type schemas
2. Delete mode-specific code and commands
3. Document manual migration (most users won't have active modes in progress)
4. Ship breaking change with clear upgrade guide

**Recommendation**: Breaking change acceptable. Consolidating modes to projects is already a breaking change; combining schema improvements into single migration minimizes disruption.

### Branch Prefix Mapping

Automatically detect project type from branch prefix:
- `explore/` → ExplorationProjectState
- `design/` → DesignProjectState
- `breakdown/` → BreakdownProjectState
- Other branches → StandardProjectState (default)

Preserves existing branch naming conventions while transitioning to unified project system.

## Open Questions

- [ ] CLI command consolidation: keep mode-specific commands (sow exploration add-topic) or unify to project commands (sow agent task create)?
- [ ] State machine implementation: define state machines in code or declaratively in project type schemas?
- [ ] Prompt generation: how to generate state-specific prompts for orchestrator guidance?
- [ ] Migration tooling: automated migration vs manual migration guide?

## Implementation Considerations

**Schema changes** (schemas/phases/common.cue):
```cue
#Phase: {
  status: string  // Remove constraint
  enabled: bool
  created_at: time.Time
  started_at?: time.Time
  completed_at?: time.Time

  inputs?: [...#Artifact]     // NEW: input artifacts
  artifacts: [...#Artifact]   // Output artifacts
  tasks: [...#Task]
  metadata?: {[string]: _}
}

#Artifact: {
  path: string
  approved?: bool  // NEW: optional
  created_at: time.Time
  metadata?: {[string]: _}
}

#Task: {
  id: string & =~"^[0-9]{3,}$"
  name: string & !=""
  status: "pending" | "in_progress" | "completed" | "abandoned"
  parallel: bool
  dependencies?: [...string]

  refs?: [...#Artifact]     // NEW: related files/artifacts
  metadata?: {[string]: _}  // NEW: task-specific data
}
```

**Project type schemas** (schemas/projects/exploration.cue):
```cue
#ExplorationProjectState: {
  statechart: {
    current_state: "Gathering" | "Researching" | "Summarizing" | "Completed"
  }

  project: {
    type: "exploration"
    name: string
    branch: string
    description: string
    created_at: time.Time
    updated_at: time.Time
  }

  phases: {
    research: #Phase & {
      status: "gathering" | "researching" | "summarizing" | "completed"
    }
  }
}
```

**Discriminated union** (schemas/projects/projects.cue):
```cue
#ProjectState:
  | #StandardProjectState
  | #ExplorationProjectState
  | #DesignProjectState
  | #BreakdownProjectState
```

## Participants

**Conducted:** January 28, 2025
**Participants:** Josh Gilman, Claude

## Artifacts Created

1. **mode-to-phase-translation.md** - Detailed analysis of how each mode's index schema translates to project phases, including field-by-field mappings and key considerations.

2. **schema-improvements.md** - Comprehensive documentation of six schema improvements enabling clean translation, with rationale, examples, and impact analysis.

3. **phase-status-constraint-removal.md** - Deep dive on removing Phase.status constraint to enable custom state flows per project type, with example state machines and orchestrator behavior changes.

## Recommendation

**Consolidate modes to projects using the six schema improvements identified.** This consolidation delivers significant architectural benefits (simpler mental model, state machine workflows, improved context engineering) while preserving all mode features and flexibility. Accept as breaking change—the value justifies migration effort.
