# Project Type Refactor - Summary

## The Problem

The current codebase is structured in a way that **prevents adding new project types** without duplicating CLI commands and breaking architectural abstractions. This document summarizes the core design intent and how the current implementation fails to achieve it.

## Original Design Intent

### Core Principle: CLI Must Be Type-Agnostic

The CLI should have **zero knowledge** of specific project types or phase implementations. It should work through generic interfaces, with phase-specific behavior encoded in:

1. **CUE Schemas** - What fields each phase has
2. **State Machine Guards** - What conditions must be met to transition
3. **Phase Prompt Templates** - What instructions the orchestrator receives

### The Vision: Minimal Generic Commands

The CLI should provide a small set of generic commands that work across all project types and phases:

```bash
# Artifact operations (any phase that supports them)
sow agent artifact add --path <path> --metadata key=value
sow agent artifact approve <path>
sow agent artifact list

# Task operations (implementation phase)
sow agent task add <name>
sow agent task update --status <status>
sow agent task list

# Generic field setters (any phase can have custom fields)
sow agent set <field> <value>
sow agent get <field>

# Lifecycle
sow agent complete
sow agent skip
sow agent enable

# Inspection
sow agent status
sow agent info
```

### How It Should Work: Review Phase Example

Instead of having dedicated `review_add.go`, `review_approve.go` commands, the review phase should:

1. **Use generic artifacts**: Review reports are just artifacts with `metadata.type = "review"`
2. **Instruct via prompts**: The review phase prompt tells the orchestrator:
   ```markdown
   Create a review report and add it as an artifact:

   sow agent artifact add \
     --path project/phases/review/reports/001.md \
     --metadata type=review \
     --metadata assessment=pass

   Wait for human approval, then complete the phase.
   ```

3. **Enforce via guards**: The review phase guard checks for unapproved artifacts:
   ```go
   func (p *ReviewPhase) canComplete() bool {
       for _, artifact := range p.state.Artifacts {
           if artifact.Metadata["type"] == "review" && !artifact.Approved {
               return false
           }
       }
       return true
   }
   ```

This means **no review-specific CLI commands** are needed. The same pattern applies to all phases.

## Current Implementation Problems

### Problem 1: Phase-Specific CLI Commands

The CLI currently has commands tied to specific phases:

**Review Phase:**
- `cmd/agent/review_add.go` - Should use generic `artifact add`
- `cmd/agent/review_approve.go` - Should use generic `artifact approve`
- `cmd/agent/review_increment.go` - Should use generic `set iteration <n>`

**Finalize Phase:**
- `cmd/agent/finalize_complete.go` - Should use generic `complete` with guards
- `cmd/agent/finalize_doc.go` - Should use generic field setters
- `cmd/agent/finalize_move.go` - Could be generic artifact operation

**Discovery/Design Phases:**
- `cmd/agent/enable.go` - Phase-specific enable logic
- `cmd/agent/skip.go` - Phase-specific skip logic

**Impact**: Every new project type would require duplicating or branching these commands.

### Problem 2: Schema-Level Type Specificity

Schemas encode phase-specific types instead of using generic structures:

**Current `#ReviewPhase`:**
```cue
#ReviewPhase: {
    iteration: int
    reports: [...#ReviewReport]  // Custom type!
}
```

**Should be:**
```cue
#ReviewPhase: {
    iteration: int
    artifacts: [...#Artifact]  // Generic with metadata
}
```

**Current `#Artifact`:**
```cue
#Artifact: {
    path: string
    approved: bool
    created_at: time.Time
}
```

**Should be:**
```cue
#Artifact: {
    path: string
    approved: bool
    created_at: time.Time
    metadata?: {[string]: _}  // Free-form phase-specific data
}
```

**Impact**: Adding new artifact types or review formats requires schema changes instead of using metadata.

### Problem 3: Direct State Access in Project Type

The `Project` type in `internal/project/project.go` directly accesses `schemas.ProjectState` fields:

```go
func (p *Project) AddReviewReport(path, assessment string) error {
    state := p.State()
    state.Phases.Review.Reports = append(...)  // Assumes StandardProjectState structure
}
```

**Impact**:
- Cannot add new project types with different phase structures
- Violates abstraction - CLI code knows about StandardProjectState internals
- Requires casting when discriminated unions are introduced

### Problem 4: Missing Interface Abstraction

There is no `Phase` interface. The CLI works directly with concrete phase types through the `Project` aggregate:

```go
// Current: Project has methods for each phase's operations
func (p *Project) AddReviewReport(...)     // Review-specific
func (p *Project) AddTask(...)             // Implementation-specific
func (p *Project) AddArtifact(phase, ...)  // Generic but leaky
```

**Should be:**
```go
// Project returns Phase interfaces
phase := project.CurrentPhase()
phase.AddArtifact(...)  // Works for any phase that supports artifacts
phase.Complete()        // Works for all phases
```

**Impact**: No way to add new phases without modifying `Project` interface.

## Why This Matters

### Current State: Adding a New Project Type Requires

1. Duplicating CLI commands or adding type-checking branches
2. Updating `Project` type to handle new phase structures
3. Adding schema-level types for any phase-specific data
4. Updating every place that accesses `state.Phases.*` fields

**Estimated effort**: Weeks of refactoring across the entire codebase.

### Desired State: Adding a New Project Type Requires

1. Define schemas (compose from existing phase types)
2. Implement `ProjectType` interface with state machine configuration
3. Write phase-specific guards
4. Write phase prompt templates

**Estimated effort**: Days, no CLI changes needed.

## Summary

The current architecture has phase-specific logic **baked into the CLI layer** instead of being isolated in **schemas, guards, and prompts**. This creates tight coupling that prevents the system from scaling to multiple project types.

The refactor must:
1. Make the CLI generic and type-agnostic
2. Use metadata-driven artifacts instead of specialized types
3. Introduce Phase/Project interfaces for abstraction
4. Move phase-specific logic out of CLI commands and into guards/prompts

---

**Next**: See `PLAN.md` for the detailed implementation plan.
