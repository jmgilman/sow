# Task 020: Define States and Events

## Overview

Create state and event constant definitions for the SDK-based standard project implementation. These constants define the state machine structure using SDK types.

**Critical**: Use SDK types (`internal/sdks/state`) NOT old statechart types (`internal/project/statechart`).

## Context

**Design Reference**:
- `.sow/knowledge/designs/project-sdk-implementation.md` (lines 225-262) for universal types
- `.sow/knowledge/designs/command-hierarchy-design.md` (lines 248-610) for state machine diagram

**Existing Implementation Reference**:
- `cli/internal/project/standard/states.go` - Old implementation using `statechart.State`
- `cli/internal/project/standard/events.go` - Old implementation using `statechart.Event`

**Migration Goal**: Preserve the same state/event structure but use SDK types.

## Requirements

### States File

Create `cli/internal/projects/standard/states.go`:

```go
package standard

import (
	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Standard project states for the 5-phase workflow.

const (
	// NoProject indicates no active project (initial and final state)
	NoProject = state.State("NoProject")

	// PlanningActive indicates planning phase in progress
	PlanningActive = state.State("PlanningActive")

	// ImplementationPlanning indicates implementation planning step
	ImplementationPlanning = state.State("ImplementationPlanning")

	// ImplementationExecuting indicates task execution
	ImplementationExecuting = state.State("ImplementationExecuting")

	// ReviewActive indicates review phase in progress
	ReviewActive = state.State("ReviewActive")

	// FinalizeDocumentation indicates documentation update step
	FinalizeDocumentation = state.State("FinalizeDocumentation")

	// FinalizeChecks indicates final validation checks
	FinalizeChecks = state.State("FinalizeChecks")

	// FinalizeDelete indicates project cleanup step
	FinalizeDelete = state.State("FinalizeDelete")
)
```

### Events File

Create `cli/internal/projects/standard/events.go`:

```go
package standard

import (
	"github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Standard project events trigger state transitions.

const (
	// EventProjectInit creates new project
	// Transition: NoProject → PlanningActive
	EventProjectInit = state.Event("project_init")

	// EventCompletePlanning completes planning phase
	// Transition: PlanningActive → ImplementationPlanning
	// Guard: task_list artifact approved
	EventCompletePlanning = state.Event("complete_planning")

	// EventTasksApproved approves implementation tasks
	// Transition: ImplementationPlanning → ImplementationExecuting
	// Guard: tasks_approved metadata flag
	EventTasksApproved = state.Event("tasks_approved")

	// EventAllTasksComplete indicates all tasks done
	// Transition: ImplementationExecuting → ReviewActive
	// Guard: all tasks completed or abandoned
	EventAllTasksComplete = state.Event("all_tasks_complete")

	// EventReviewPass passes review assessment
	// Transition: ReviewActive → FinalizeDocumentation
	// Guard: review artifact approved with assessment=pass
	EventReviewPass = state.Event("review_pass")

	// EventReviewFail fails review (rework loop)
	// Transition: ReviewActive → ImplementationPlanning
	// Guard: review artifact approved with assessment=fail
	EventReviewFail = state.Event("review_fail")

	// EventDocumentationDone completes documentation
	// Transition: FinalizeDocumentation → FinalizeChecks
	EventDocumentationDone = state.Event("documentation_done")

	// EventChecksDone completes final checks
	// Transition: FinalizeChecks → FinalizeDelete
	EventChecksDone = state.Event("checks_done")

	// EventProjectDelete deletes project
	// Transition: FinalizeDelete → NoProject
	// Guard: project_deleted metadata flag
	EventProjectDelete = state.Event("project_delete")
)
```

## State Machine Diagram

```
NoProject
    ↓ EventProjectInit
PlanningActive
    ↓ EventCompletePlanning (guard: task_list approved)
ImplementationPlanning
    ↓ EventTasksApproved (guard: tasks_approved metadata)
ImplementationExecuting
    ↓ EventAllTasksComplete (guard: all tasks completed)
ReviewActive
    ├→ EventReviewPass (assessment="pass") → FinalizeDocumentation
    └→ EventReviewFail (assessment="fail") → ImplementationPlanning (rework)

FinalizeDocumentation
    ↓ EventDocumentationDone
FinalizeChecks
    ↓ EventChecksDone
FinalizeDelete
    ↓ EventProjectDelete (guard: project_deleted metadata)
NoProject
```

## Acceptance Criteria

- [ ] File `cli/internal/projects/standard/states.go` created
- [ ] File `cli/internal/projects/standard/events.go` created
- [ ] All 8 states defined using `state.State` type (SDK type)
- [ ] All 9 events defined using `state.Event` type (SDK type)
- [ ] Each constant has clear documentation comment
- [ ] Events include transition documentation (from/to states, guards)
- [ ] Imports use `internal/sdks/state` NOT `internal/project/statechart`
- [ ] Files compile without errors: `go build ./cli/internal/projects/standard/...`
- [ ] Old package `cli/internal/project/standard/` untouched

## Validation Commands

```bash
# Verify files exist
ls cli/internal/projects/standard/states.go
ls cli/internal/projects/standard/events.go

# Verify compilation
go build ./cli/internal/projects/standard/...

# Verify imports are correct (should reference internal/sdks/state)
grep "internal/sdks/state" cli/internal/projects/standard/states.go
grep "internal/sdks/state" cli/internal/projects/standard/events.go

# Verify NOT using old statechart types
! grep "internal/project/statechart" cli/internal/projects/standard/states.go
! grep "internal/project/statechart" cli/internal/projects/standard/events.go

# Verify old package untouched
git diff cli/internal/project/standard/
```

## Dependencies

- Task 010 (package structure exists)

## Standards

- Use descriptive constant names matching state machine diagram
- Include comprehensive documentation comments for each constant
- Follow Go naming conventions (PascalCase for exported constants)
- Use SDK types exclusively (no old statechart types)

## Notes

- These constants will be used throughout Tasks 4, 5, and 6
- State names match the state machine diagram from design doc
- Event names describe the action that triggers the transition
- Guards (transition conditions) are documented but not implemented here (Task 5)
