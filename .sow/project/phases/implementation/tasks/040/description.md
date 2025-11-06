# Task 040: Implement Guard Functions

## Context

This task implements guard functions that control state machine transitions. Guards are pure functions that examine project state and return a boolean indicating whether a transition is allowed.

The exploration project type has three transitions, each with a guard:
1. Active → Summarizing: Guard checks if all research topics are resolved
2. Summarizing → Finalizing: Guard checks if all summary artifacts are approved
3. Finalizing → Completed: Guard checks if all finalization tasks are complete

Guards operate on `*state.Project` and access phase data directly. They must be pure functions with no side effects.

## Requirements

### Create Guards File

Create `cli/internal/projects/exploration/guards.go` with:

1. **Package declaration and imports**:
   ```go
   package exploration

   import (
       "github.com/jmgilman/sow/cli/internal/sdks/project/state"
       projschema "github.com/jmgilman/sow/cli/schemas/project"
   )
   ```

2. **allTasksResolved() function**:
   - Checks if all research topics in exploration phase are completed or abandoned
   - Guards Active → Summarizing transition
   - Returns false if:
     - Exploration phase doesn't exist
     - No tasks exist (must have at least one research topic)
     - Any task has status other than "completed" or "abandoned"
   - Returns true if all tasks are completed or abandoned

3. **allSummariesApproved() function**:
   - Checks if at least one summary artifact exists and all summaries are approved
   - Guards Summarizing → Finalizing transition
   - Summary artifacts are identified by type == "summary"
   - Returns false if:
     - Exploration phase doesn't exist
     - No summary artifacts exist
     - Any summary artifact has Approved == false
   - Returns true if at least one summary exists and all are approved

4. **allFinalizationTasksComplete() function**:
   - Checks if all finalization tasks are completed
   - Guards Finalizing → Completed transition
   - Returns false if:
     - Finalization phase doesn't exist
     - No tasks exist
     - Any task has status != "completed"
   - Returns true if all tasks are completed

5. **Helper functions** (optional but recommended):
   ```go
   // countUnresolvedTasks returns count of pending/in_progress tasks
   func countUnresolvedTasks(p *state.Project) int

   // countUnapprovedSummaries returns count of unapproved summary artifacts
   func countUnapprovedSummaries(p *state.Project) int
   ```

## Test-Driven Development

This task follows TDD methodology:

1. **Write tests first** for each guard function:
   - `allTasksResolved()`:
     - Returns false when exploration phase missing
     - Returns false when no tasks exist
     - Returns false when tasks are pending/in_progress
     - Returns true when all tasks completed
     - Returns true when tasks are completed or abandoned
   - `allSummariesApproved()`:
     - Returns false when exploration phase missing
     - Returns false when no summary artifacts exist
     - Returns false when summaries not approved
     - Returns true when all summaries approved
     - Filters only "summary" type artifacts
   - `allFinalizationTasksComplete()`:
     - Returns false when finalization phase missing
     - Returns false when no tasks exist
     - Returns false when tasks not completed
     - Returns true when all tasks completed

2. **Run tests** - they should fail initially (red phase)

3. **Implement functionality** - write guard functions (green phase)

4. **Refactor** - extract helper functions for readability

Place tests in `cli/internal/projects/exploration/guards_test.go`.

## Acceptance Criteria

- [ ] File `guards.go` exists with all required functions
- [ ] **Unit tests written before implementation**
- [ ] Comprehensive test coverage for all guard behaviors
- [ ] All tests pass
- [ ] `allTasksResolved()` correctly validates task completion
- [ ] `allTasksResolved()` requires at least one task to exist
- [ ] `allTasksResolved()` accepts both "completed" and "abandoned" statuses
- [ ] `allSummariesApproved()` correctly validates summary approval
- [ ] `allSummariesApproved()` requires at least one summary
- [ ] `allSummariesApproved()` filters by type == "summary"
- [ ] `allFinalizationTasksComplete()` validates all tasks completed
- [ ] All functions are pure (no side effects)
- [ ] All functions handle missing phases gracefully (return false)
- [ ] Helper functions implemented (optional)
- [ ] Code follows Go formatting standards (gofmt)
- [ ] No compilation errors

## Technical Details

### Guard Function Signature

Guards have the signature:
```go
func(p *state.Project) bool
```

They are called by the state machine during transition validation. If a guard returns false, the transition is blocked.

### Phase Access Pattern

Access phases via map lookup with existence check:
```go
phase, exists := p.Phases["exploration"]
if !exists {
    return false
}
```

This is safer than direct access and handles missing phases gracefully.

### Task Status Values

Valid task status values:
- "pending" - Not started
- "in_progress" - Currently being worked on
- "completed" - Finished successfully
- "abandoned" - Determined not needed

The "needs_review" status is NOT used in exploration projects (no formal review cycle).

### Artifact Approval

Artifacts have an `Approved` boolean field. Summary artifacts require explicit approval before advancing to finalization. Research findings (type="findings") do not require approval.

### Task vs Artifact Distinction

- **Tasks**: Formal work items tracked in phase.Tasks with status, assignee, etc.
- **Artifacts**: Files/documents tracked in phase.Outputs with approval status

Exploration phase uses tasks for research topics. Finalization phase may use tasks for checklist items.

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/guards.go` - Reference guard implementations
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/state/project.go` - Project state structure
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/schemas/project/task.cue` - Task schema
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/schemas/project/artifact.cue` - Artifact schema
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/knowledge/designs/project-modes/exploration-design.md` - Design specification (lines 356-481)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/project/context/issue-36.md` - Requirements

## Examples

### Standard Project Guard (Reference)

From `cli/internal/projects/standard/guards.go:84-102`:

```go
// allTasksComplete checks if all implementation tasks are completed or abandoned.
// Returns false if implementation phase missing or if no tasks exist.
func allTasksComplete(p *state.Project) bool {
    phase, exists := p.Phases["implementation"]
    if !exists {
        return false
    }

    if len(phase.Tasks) == 0 {
        return false
    }

    for _, task := range phase.Tasks {
        if task.Status != "completed" && task.Status != "abandoned" {
            return false
        }
    }
    return true
}
```

### Artifact Filtering Pattern

From exploration design:

```go
summaries := []projschema.ArtifactState{}

// Collect all summary artifacts
for _, artifact := range phase.Outputs {
    if artifact.Type == "summary" {
        summaries = append(summaries, artifact)
    }
}

// Must have at least one summary
if len(summaries) == 0 {
    return false
}

// All summaries must be approved
for _, summary := range summaries {
    if !summary.Approved {
        return false
    }
}
```

### Helper Function Example

```go
func countUnresolvedTasks(p *state.Project) int {
    phase, exists := p.Phases["exploration"]
    if !exists {
        return 0
    }

    count := 0
    for _, task := range phase.Tasks {
        if task.Status != "completed" && task.Status != "abandoned" {
            count++
        }
    }
    return count
}
```

## Dependencies

- Task 010 (Package structure) - Provides package directory
- Task 020 (States and events) - Provides state constants (used in transition config)
- Will be used by Task 050 (Transitions) to configure guards
- Helper functions will be used by Task 070 (Prompts) for status messages

## Constraints

- Guards must be pure functions (no state mutation)
- Must handle missing phases gracefully (return false, not panic)
- Cannot return errors (boolean only)
- Must validate all conditions (existence, counts, status)
- Task status checks must accept both "completed" and "abandoned"
- Summary filtering must check type == "summary" explicitly
- No direct console output or logging in guards
