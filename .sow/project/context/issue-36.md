# Issue #36: Exploration Project Type Implementation

**URL**: https://github.com/jmgilman/sow/issues/36
**State**: OPEN

## Description

# Exploration Project Type Implementation

## Overview

Implement the exploration project type - a workflow for research, investigation, and knowledge gathering. Exploration projects help users systematically investigate topics, document findings, and synthesize results into comprehensive summaries.

**Reference Design**: `.sow/knowledge/designs/project-modes/exploration-design.md`

## Key Characteristics

- **Flexible topic discovery**: Add research topics dynamically as exploration progresses
- **Task-based lifecycle**: Each topic is a task with independent status tracking
- **Synthesis-focused**: Culminates in approved summary artifact(s)
- **No GitHub integration**: Pure research workflow, no issue linking
- **SDK-based implementation**: Built using Project SDK builder pattern

## State Machine

The exploration workflow progresses through **4 states** across 2 phases:

1. **Active** (exploration phase, status="active")
   - Purpose: Active research - identify topics, investigate, document findings
   - Operations: Add/complete tasks, create research findings
   - Advance condition: All tasks resolved (completed or abandoned)

2. **Summarizing** (exploration phase, status="summarizing")
   - Purpose: Synthesize findings into comprehensive summary document(s)
   - Operations: Create summaries, approve artifacts
   - Complete condition: All summary artifacts approved

3. **Finalizing** (finalization phase, status="in_progress")
   - Purpose: Move artifacts to permanent location, create PR, cleanup
   - Operations: Execute finalization checklist
   - Complete condition: All finalization tasks completed

4. **Completed** (terminal state)

## Phases

### Exploration Phase

- **Active → Summarizing** (intra-phase transition)
- Dynamic topic discovery (can add topics anytime during Active)
- Cannot add topics in Summarizing state
- Research findings don't require approval during Active
- Summary artifacts require approval in Summarizing

### Finalization Phase

- Move approved summaries to `.sow/knowledge/explorations/`
- Support for multiple summary documents with ToC validation
- Single-document explorations work without ToC requirement
- Delete `.sow/project/` on completion

## Implementation Details

**Package**: `cli/internal/projects/exploration/`

**Files**:
- `exploration.go` - SDK configuration using builder pattern
- `states.go` - State constants (Active, Summarizing, Finalizing, Completed)
- `events.go` - Event constants (EventBeginSummarizing, EventCompleteSummarizing, EventCompleteFinalization)
- `guards.go` - Guard functions (allTasksResolved, allSummariesApproved, allFinalizationTasksComplete)
- `prompts.go` - Prompt generator functions
- `metadata.go` - Embedded CUE metadata schemas
- `cue/exploration_metadata.cue` - Phase metadata validation

**Guards**:
- `allTasksResolved(p)`: All tasks completed or abandoned (guards Active → Summarizing)
- `allSummariesApproved(p)`: At least one summary exists, all approved (guards Summarizing → Finalizing)
- `allFinalizationTasksComplete(p)`: All finalization tasks completed (guards Finalizing → Completed)

**Summary Structure**:
- **Single summary**: Can use any filename, no ToC requirement
- **Multiple summaries**: `summary.md` **must** exist as overview/ToC, validated during finalization

## Acceptance Criteria

### Schema and Types
- [ ] ProjectState validates correctly with type="exploration"
- [ ] Phase status constraints enforced
- [ ] State machine state constraints enforced
- [ ] Can create exploration project on `explore/*` branch

### State Machine
- [ ] All transitions configured correctly
- [ ] Guards prevent invalid transitions
- [ ] Events fire correctly
- [ ] Prompts generate for all states

### Exploration Phase
- [ ] Can add/update research topics dynamically in Active state
- [ ] Cannot add topics in Summarizing state
- [ ] Can advance to Summarizing when all tasks resolved
- [ ] Can create and approve multiple summary artifacts
- [ ] Can advance to Finalizing when all summaries approved

### Finalization
- [ ] Multiple summaries without `summary.md` fail validation with clear error
- [ ] Single summary can be finalized without ToC requirement
- [ ] Finalization moves artifacts to appropriate structure (file or folder)
- [ ] Finalization creates PR with link to exploration location
- [ ] Project directory deleted on completion

### Integration
- [ ] End-to-end workflow works (Active → Summarizing → Finalizing → Completed)
- [ ] Zero-context resumability works
- [ ] All tests pass

## Success Criteria

1. ✅ Can create exploration project on `explore/*` branch
2. ✅ Can add/update research topics dynamically during Active state
3. ✅ Cannot add topics in Summarizing state (enforced by SDK)
4. ✅ Can advance to Summarizing when all tasks resolved
5. ✅ Can create and approve multiple summary artifacts
6. ✅ Can advance to Finalizing when all summaries approved
7. ✅ Multiple summaries without `summary.md` fail validation with clear error
8. ✅ Single summary can be finalized without ToC requirement
9. ✅ Finalization moves artifacts to appropriate structure (file or folder)
10. ✅ Finalization creates PR with link to exploration location
11. ✅ Zero-context resumability works (project state on disk)
12. ✅ Prompts provide clear guidance at each state
13. ✅ All guards prevent invalid transitions
14. ✅ SDK configuration is declarative and testable

## References

- **Design Document**: `.sow/knowledge/designs/project-modes/exploration-design.md`
- **Core Design**: `.sow/knowledge/designs/project-modes/core-design.md`
- **Project SDK**: `cli/internal/sdks/project/` (builder, config, machine)
- **Standard Project**: `cli/internal/projects/standard/` (reference implementation)
- **State Machine SDK**: `cli/internal/sdks/state/` (underlying state machine)
