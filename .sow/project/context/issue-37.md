# Issue #37: Design Project Type Implementation

**URL**: https://github.com/jmgilman/sow/issues/37
**State**: OPEN

## Description

# Design Project Type Implementation

## Overview

Implement the design project type - a workflow for creating architecture and design documentation. Design projects help users track design artifacts (ADRs, design docs, architecture docs, diagrams) from planning through approval and finalization.

**Reference Design**: `.sow/knowledge/designs/project-modes/design-design.md`

## Key Characteristics

- **Input tracking**: Register sources that inform design (explorations, references, existing docs)
- **Task-based document tracking**: Each task represents a document to create
- **Review workflow**: Draft → needs_review → completed (with auto-approval)
- **Metadata-driven**: Tasks store document type, target location, artifact path
- **No GitHub integration**: Pure design workflow, no issue linking
- **SDK-based implementation**: Built using Project SDK builder pattern

## State Machine

The design workflow progresses through **3 states** across 2 phases:

1. **Active** (design phase, status="active")
   - Purpose: Plan outputs, draft documents, iterate through reviews
   - Operations: Create document tasks, draft documents, mark for review, approve
   - Complete condition: All tasks completed or abandoned (at least one completed)

2. **Finalizing** (finalization phase, status="in_progress")
   - Purpose: Move approved documents to targets, create PR, cleanup
   - Operations: Execute finalization checklist
   - Complete condition: All finalization tasks completed

3. **Completed** (terminal state)

## Phases

### Design Phase

- Single Active state (no intra-phase transitions)
- Can plan documents as tasks before drafting
- Cannot add artifacts before creating at least one task
- Task completion auto-approves linked artifact
- Can add input artifacts to track design sources
- Inputs stored separately from outputs (phase.inputs)
- `needs_review` status enables review workflow
- Can transition back from `needs_review` to `in_progress`

### Finalization Phase

- Move approved documents to target locations
- Create PR with design artifacts
- Delete `.sow/project/` on completion

## Implementation Details

**Package**: `cli/internal/projects/design/`

**Files**:
- `design.go` - SDK configuration using builder pattern
- `states.go` - State constants (Active, Finalizing, Completed)
- `events.go` - Event constants (EventCompleteDesign, EventCompleteFinalization)
- `guards.go` - Guard functions and helpers
- `prompts.go` - Prompt generator functions
- `metadata.go` - Embedded CUE metadata schemas
- `cue/design_metadata.cue` - Phase metadata validation

**Guards**:
- `allDocumentsApproved(p)`: All tasks completed or abandoned, at least one completed (guards Active → Finalizing)
- `allFinalizationTasksComplete(p)`: All finalization tasks completed (guards Finalizing → Completed)

**Helper Functions**:
- `validateTaskForCompletion(p, taskID)`: Checks if artifact exists at task.metadata.artifact_path
- `autoApproveArtifact(p, taskID)`: Approves artifact linked to task when task completed

**Task Metadata Structure**:
```yaml
metadata:
  artifact_path: "project/auth-design.md"
  document_type: "design"  # adr, design, architecture, diagram, spec
  target_location: ".sow/knowledge/designs/auth-design.md"
  template: "design-doc"  # Optional
```

## Acceptance Criteria

### Schema and Types
- [ ] ProjectState validates correctly with type="design"
- [ ] Phase status constraints enforced
- [ ] State machine state constraints enforced
- [ ] Can create design project on `design/*` branch

### State Machine
- [ ] All transitions configured correctly
- [ ] Guards prevent invalid transitions
- [ ] Events fire correctly
- [ ] Prompts generate for all states

### Design Phase
- [ ] Can plan documents as tasks before drafting
- [ ] Cannot add artifacts before creating at least one task
- [ ] Task completion auto-approves linked artifact
- [ ] Can add input artifacts to track design sources
- [ ] Inputs stored separately from outputs (phase.inputs)
- [ ] Task metadata tracks artifact paths and document types
- [ ] `needs_review` status enables review workflow
- [ ] Can transition back from `needs_review` to `in_progress`

### Finalization
- [ ] Moves approved documents to target locations
- [ ] Creates PR with design artifacts
- [ ] Deletes `.sow/project/` on completion

### Integration
- [ ] End-to-end workflow works (Active → Finalizing → Completed)
- [ ] Zero-context resumability works
- [ ] All tests pass

## Task Lifecycle

```
pending
  │
  │ Start drafting
  ▼
in_progress
  │
  │ Document drafted, ready for review
  ▼
needs_review
  │
  ├─→ in_progress (changes requested)
  │
  └─→ completed (approved) + artifact auto-approved
```

**Or abandon if not needed**:
```
pending/in_progress/needs_review
  │
  └─→ abandoned (document not needed)
```

## Success Criteria

1. ✅ Can create design project on `design/*` branch
2. ✅ Can register inputs to track design sources
3. ✅ Can create document tasks with metadata
4. ✅ Can link artifacts to tasks via task.outputs or metadata
5. ✅ Can transition tasks through pending → in_progress → needs_review → completed
6. ✅ Can go backward from needs_review → in_progress for revisions
7. ✅ Completing task auto-approves linked artifact
8. ✅ Cannot complete task without artifact existing (validation)
9. ✅ Can advance to Finalizing when all documents approved (at least one completed)
10. ✅ Finalization moves documents to target locations
11. ✅ Zero-context resumability works (project state on disk)
12. ✅ Prompts provide clear guidance at each state
13. ✅ SDK configuration is declarative and testable

## References

- **Design Document**: `.sow/knowledge/designs/project-modes/design-design.md`
- **Core Design**: `.sow/knowledge/designs/project-modes/core-design.md`
- **Project SDK**: `cli/internal/sdks/project/` (builder, config, machine)
- **Standard Project**: `cli/internal/projects/standard/` (reference implementation)
- **State Machine SDK**: `cli/internal/sdks/state/` (underlying state machine)
