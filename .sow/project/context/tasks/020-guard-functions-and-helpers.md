# Task 020: Guard Functions and Helpers

## Context

This task implements the guard functions and helper utilities for the design project type. Guards are pure functions that evaluate whether state transitions are allowed by examining project state. They are crucial for enforcing workflow rules and ensuring data integrity.

The design project type has two critical transitions that require guards:
1. **Active → Finalizing**: Can only advance when all document tasks are approved
2. **Finalizing → Completed**: Can only advance when all finalization tasks are complete

Additionally, the design workflow requires helper functions for task validation and auto-approval of artifacts when tasks are completed. These helpers integrate with the task lifecycle to enforce that documents exist before task completion and to automatically approve linked artifacts.

This builds on Task 010's state and event constants, implementing the business logic that controls workflow progression.

## Requirements

### Guard Functions

Implement the following guard functions in `guards.go`:

1. **allDocumentsApproved(p *state.Project) bool**
   - Guards Active → Finalizing transition
   - Returns `false` if:
     - Design phase doesn't exist
     - No tasks exist (must plan at least one document)
     - Any task has status other than "completed" or "abandoned"
     - All tasks are abandoned (must have at least one completed)
   - Returns `true` if all tasks are completed/abandoned AND at least one is completed
   - This ensures the design has meaningful output before advancing

2. **allFinalizationTasksComplete(p *state.Project) bool**
   - Guards Finalizing → Completed transition
   - Returns `false` if:
     - Finalization phase doesn't exist
     - No tasks exist
     - Any task has status != "completed"
   - Returns `true` if all finalization tasks are completed
   - Note: Unlike design tasks, finalization tasks cannot be abandoned (they must complete)

### Helper Functions

Implement helper functions for task lifecycle management:

1. **countUnresolvedTasks(p *state.Project) int**
   - Returns count of tasks in design phase with status != "completed" and != "abandoned"
   - Used in prompts and error messages
   - Returns 0 if design phase doesn't exist

2. **validateTaskForCompletion(p *state.Project, taskID string) error**
   - Validates that a task can be marked as completed
   - Checks:
     - Task exists in design phase
     - Task has metadata field
     - Task metadata contains "artifact_path" key
     - artifact_path is a valid string
     - Artifact exists at the specified path in phase.Outputs
   - Returns descriptive error if validation fails
   - Returns nil if validation passes

3. **autoApproveArtifact(p *state.Project, taskID string) error**
   - Automatically approves the artifact linked to a completed task
   - Called during task completion (status update to "completed")
   - Finds task by ID in design phase
   - Reads artifact_path from task metadata
   - Finds artifact in phase.Outputs by path
   - Sets artifact.Approved = true
   - Updates project state
   - Returns error if task not found, artifact_path invalid, or artifact not found

### Error Handling

All functions should provide clear, actionable error messages:
- "design phase not found" - Phase doesn't exist
- "task %s not found" - Task doesn't exist
- "task %s has no metadata - set artifact_path before completing" - Missing metadata
- "task %s has no artifact_path in metadata - link artifact to task before completing" - Missing artifact_path
- "artifact not found at %s - add artifact before completing task" - Artifact doesn't exist

## Acceptance Criteria

### Functional Requirements

- [ ] `guards.go` file created with all guard and helper functions
- [ ] `allDocumentsApproved` correctly evaluates all edge cases
- [ ] `allFinalizationTasksComplete` correctly evaluates all edge cases
- [ ] `countUnresolvedTasks` accurately counts unresolved tasks
- [ ] `validateTaskForCompletion` performs comprehensive validation
- [ ] `autoApproveArtifact` correctly finds and approves artifacts
- [ ] All functions are pure (no side effects except autoApproveArtifact)
- [ ] Error messages are clear and actionable

### Test Requirements (TDD)

Write comprehensive unit tests in `guards_test.go`:

**allDocumentsApproved tests**:
- [ ] Returns false when design phase missing
- [ ] Returns false when no tasks exist
- [ ] Returns false when tasks are pending
- [ ] Returns false when tasks are in_progress
- [ ] Returns false when all tasks abandoned (need at least one completed)
- [ ] Returns true when all tasks completed
- [ ] Returns true when mix of completed and abandoned (with at least one completed)

**allFinalizationTasksComplete tests**:
- [ ] Returns false when finalization phase missing
- [ ] Returns false when no tasks exist
- [ ] Returns false when any task is pending
- [ ] Returns false when any task is in_progress
- [ ] Returns false when any task is abandoned
- [ ] Returns true when all tasks completed

**countUnresolvedTasks tests**:
- [ ] Returns 0 when design phase missing
- [ ] Returns 0 when all tasks resolved
- [ ] Returns correct count for mix of statuses

**validateTaskForCompletion tests**:
- [ ] Returns error when design phase missing
- [ ] Returns error when task not found
- [ ] Returns error when task has no metadata
- [ ] Returns error when metadata missing artifact_path
- [ ] Returns error when artifact_path is not a string
- [ ] Returns error when artifact doesn't exist at path
- [ ] Returns nil when validation passes

**autoApproveArtifact tests**:
- [ ] Returns error when design phase missing
- [ ] Returns error when task not found
- [ ] Returns error when artifact_path invalid
- [ ] Returns error when artifact not found at path
- [ ] Sets artifact.Approved to true when successful
- [ ] Updates project state correctly

### Code Quality

- [ ] All functions documented with clear descriptions
- [ ] Guard functions are pure (no mutations except where documented)
- [ ] Helper functions handle nil/missing data gracefully
- [ ] Test coverage for all edge cases
- [ ] Tests use table-driven approach where appropriate

## Technical Details

### Function Signatures

```go
package design

import (
	"fmt"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// Guard functions
func allDocumentsApproved(p *state.Project) bool
func allFinalizationTasksComplete(p *state.Project) bool

// Helper functions
func countUnresolvedTasks(p *state.Project) int
func validateTaskForCompletion(p *state.Project, taskID string) error
func autoApproveArtifact(p *state.Project, taskID string) error
```

### Project State Access

Guards access project state through the `*state.Project` parameter:
- `p.Phases` - Map of phase name to PhaseState
- `phase.Tasks` - Slice of TaskState
- `phase.Outputs` - Slice of ArtifactState
- `task.Status` - Task status string
- `task.Metadata` - Map of metadata key to value
- `artifact.Path` - Artifact file path
- `artifact.Approved` - Artifact approval boolean

### Task Status Values

Standard task statuses used in design project:
- `"pending"` - Task planned but not started
- `"in_progress"` - Actively working on task
- `"needs_review"` - Ready for human review
- `"completed"` - Task approved and finished
- `"abandoned"` - Task no longer needed

### Metadata Structure

Task metadata for design documents:
```yaml
metadata:
  artifact_path: "project/auth-design.md"
  document_type: "design"
  target_location: ".sow/knowledge/designs/auth-design.md"
  template: "design-doc"  # Optional
```

The `artifact_path` is the critical field for validation and auto-approval.

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/.sow/project/context/issue-37.md` - Requirements for guards and helpers
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/.sow/knowledge/designs/project-modes/design-design.md` - Guard specifications (lines 318-489)
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/guards.go` - Reference guard implementation pattern
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/guards_test.go` - Reference guard testing pattern
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/schemas/project/project.go` - Project state schema definitions

## Examples

### Guard Usage in Transition

Guards are passed as closures to transition options:
```go
AddTransition(
    sdkstate.State(Active),
    sdkstate.State(Finalizing),
    sdkstate.Event(EventCompleteDesign),
    project.WithGuard("all documents approved", func(p *state.Project) bool {
        return allDocumentsApproved(p)
    }),
)
```

### Validation Usage

Called before allowing task completion:
```go
// Before marking task as completed
if err := validateTaskForCompletion(project, taskID); err != nil {
    return fmt.Errorf("cannot complete task: %w", err)
}

// Mark task as completed
task.Status = "completed"

// Auto-approve linked artifact
if err := autoApproveArtifact(project, taskID); err != nil {
    return fmt.Errorf("failed to approve artifact: %w", err)
}
```

### Helper Test Structure

```go
func TestAllDocumentsApproved_MixOfStatuses(t *testing.T) {
    p := newTestProject()
    p.Phases["design"] = projschema.PhaseState{
        Tasks: []projschema.TaskState{
            newTask("010", "completed"),
            newTask("020", "abandoned"),
            newTask("030", "completed"),
        },
    }

    result := allDocumentsApproved(p)

    if !result {
        t.Error("Expected true when mix includes completed tasks, got false")
    }
}
```

## Dependencies

- Task 010: Core Structure and Constants - Provides state constants used in guard logic

## Constraints

### Pure Functions

Guards must be pure functions with no side effects:
- They can only read project state
- They cannot modify project state
- They must be deterministic (same input → same output)
- Exception: `autoApproveArtifact` modifies state (documented as helper, not guard)

### Performance

Guards are called frequently during state transitions:
- Should be O(n) where n is number of tasks/artifacts
- Avoid unnecessary iterations
- Cache results within single evaluation if needed

### Validation Philosophy

The design project enforces:
- Plan before draft: Must create tasks before artifacts
- Draft before complete: Must add artifact before completing task
- Auto-approval on completion: Completing task auto-approves artifact
- At least one success: Cannot advance with all abandoned tasks

These rules ensure meaningful output and prevent orphaned artifacts.
