# Implementation Plan: Add `needs_review` Status for Task-Level Reviews

## Overview
Add a new `needs_review` task status that creates a quality gate after worker completion and before task completion. The orchestrator performs a lightweight review of each task's changes, either approving (→ completed) or providing feedback (→ in_progress with iteration increment).

## Phase 1: Schema Changes

### 1.1 Update CUE Schemas
**Files:**
- `cli/schemas/task_state.cue` (line 21)
- `cli/schemas/phases/common.cue` (line 52)

**Changes:**
```cue
// Before:
status: "pending" | "in_progress" | "completed" | "abandoned"

// After:
status: "pending" | "in_progress" | "needs_review" | "completed" | "abandoned"
```

### 1.2 Regenerate Go Types
**Command:** `cue exp gengotypes` (or project's equivalent command)
**Files updated:**
- `cli/schemas/cue_types_gen.go`
- `cli/schemas/phases/cue_types_gen.go`

## Phase 2: Core Logic Updates

### 2.1 Update Task Domain Logic
**File:** `cli/internal/project/domain/task.go`

**Location:** Line 57-62 (validStatuses map)
```go
validStatuses := map[string]bool{
    "pending":      true,
    "in_progress":  true,
    "needs_review": true,  // NEW
    "completed":    true,
    "abandoned":    true,
}
```

**Location:** Line 103-127 (SetStatus completion check)
```go
// Current: Checks if all tasks are "completed" or "abandoned"
// Change: Only count "completed" or "abandoned" (NOT "needs_review")
// Rationale: needs_review tasks are not yet complete
```

### 2.2 Update State Machine Guards
**File:** `cli/internal/project/statechart/guards.go`

**Search for:** References to task status checks
**Update:** Ensure guards checking for "all tasks complete" exclude `needs_review` status

## Phase 3: CLI Commands

### 3.1 Create Review Command
**New File:** `cli/cmd/agent/task/review.go`

**Command Structure:**
```go
sow agent task review [task-id] --approve
sow agent task review [task-id] --request-changes
```

**Implementation:**

1. **Shared validation:**
   - Verify `review.md` exists at `project/phases/implementation/tasks/<id>/review.md`
   - Verify task status is `needs_review`
   - Support task ID inference (same as update command)

2. **`--approve` flag:**
   - Read task state
   - Transition status: `needs_review` → `completed`
   - Keep `review.md` file (preserved as approval record)
   - Update timestamps (completed_at)
   - Save state
   - Print success message

3. **`--request-changes` flag:**
   - Read task state
   - Read `review.md` content
   - Generate next feedback ID (using existing Task.AddFeedback pattern)
   - Move `review.md` → `feedback/<id>.md`
   - Add feedback entry to task state
   - Increment iteration counter
   - Transition status: `needs_review` → `in_progress`
   - Clear started_at and completed_at (new iteration)
   - Save state
   - Print message: "Feedback added: feedback/<id>.md. Task returned to worker."

4. **Error handling:**
   - Missing review.md: Clear error with path where it's expected
   - Wrong status: Show current status and expected status
   - Invalid task ID: Standard task not found error

### 3.2 Register Review Command
**File:** `cli/cmd/agent/task/task.go`

**Location:** Line 40-48 (NewTaskCmd function)
```go
cmd.AddCommand(NewReviewCmd())  // Add after line 42
```

### 3.3 Update Update Command Documentation
**File:** `cli/cmd/agent/task/update.go`

**Location:** Line 19 (status list in help text)
```
- Status (pending, in_progress, needs_review, completed, abandoned)
```

**Location:** Line 43 (flag help)
```
"New task status (pending, in_progress, needs_review, completed, abandoned)"
```

## Phase 4: Prompt Updates

### 4.1 Update Implementer Agent
**File:** `.claude/agents/implementer.md`

**Location:** End of Workflow section (after line 141)
```markdown
7. Mark task as needs_review: sow agent task update <id> --status needs_review
8. Control returns to orchestrator for review
```

**Rationale:** Workers no longer mark tasks as completed directly

### 4.2 Update Implementation Executing Prompt
**File:** `cli/internal/prompts/templates/statechart/implementation_executing.md`

**Location:** After line 33 (add new section)
```markdown
TASK REVIEW WORKFLOW:

When a task transitions to "needs_review":

1. Read task requirements from description.md
2. Check task state.yaml for files_modified list
3. Review actual changes: git diff for those files
4. Write review to: project/phases/implementation/tasks/<id>/review.md

   Include in review.md:
   - Summary of requirements
   - What was actually changed
   - Assessment (approve or request changes)
   - If requesting changes: specific issues to address

5. Execute review decision:

   APPROVE:
   sow agent task review <id> --approve
   → Task marked completed, review.md preserved

   REQUEST CHANGES:
   sow agent task review <id> --request-changes
   → Review becomes feedback, worker re-invoked with iteration + 1
```

**Location:** Update line 22 (Autonomy boundaries)
```markdown
Full Autonomy (no approval needed):
  • Reviewing tasks (lightweight sanity check)  ← ADD THIS
  • Approving or rejecting task reviews         ← ADD THIS
  • Marking tasks completed (via review)         ← MODIFY
  • Moving to next task
  • Re-invoking implementers with feedback
  • Adjusting task descriptions
  • Normal task execution flow
```

**Location:** Update line 42 (Next Actions)
```markdown
- For pending tasks: Spawn implementer agent
- When task reaches needs_review: Perform review (see workflow above)
- To approve: sow agent task review <id> --approve
- To reject: sow agent task review <id> --request-changes
- When all done: Auto-transition to review
```

### 4.3 Update Continue Command Template
**File:** `cli/internal/prompts/templates/commands/continue.md`

**Location:** Line 73 (Task Management section)
```bash
sow agent task review <id> --approve          # Approve task review
sow agent task review <id> --request-changes  # Request changes
```

## Phase 5: Testing

### 5.1 Unit Tests
**New File:** `cli/cmd/agent/task/review_test.go`

Test cases:
- Approve with valid review.md
- Request changes with valid review.md
- Error when review.md missing
- Error when task not in needs_review status
- Error when task not found
- Task ID inference works

### 5.2 Integration Tests
**File:** `cli/testdata/script/e2e_task_review_workflow.txtar`

Test scenario:
1. Create project and tasks
2. Worker marks task as needs_review
3. Orchestrator writes review.md (approve case)
4. Orchestrator approves review
5. Verify task status = completed
6. Verify review.md preserved
7. Worker marks second task as needs_review
8. Orchestrator writes review.md (reject case)
9. Orchestrator requests changes
10. Verify feedback file created
11. Verify iteration incremented
12. Verify task status = in_progress

### 5.3 Update Existing Tests
Search for hardcoded status lists and add `needs_review`:
- `cli/internal/project/standard/phases_test.go`
- `cli/internal/project/collections_test.go`
- Any test files validating task statuses

## Phase 6: Documentation

### 6.1 Update Phase Documentation
**File:** `.sow/knowledge/phases/IMPLEMENTATION.md` (if exists)

Add section on task review workflow

### 6.2 Update README/Docs
Document the new review workflow and commands in user-facing documentation

## Implementation Order

1. **Phase 1** (Schema) - Foundation for everything else
2. **Phase 2** (Core Logic) - Enable the new status in code
3. **Phase 3** (CLI Commands) - Make it usable
4. **Phase 4** (Prompts) - Guide orchestrator/worker behavior
5. **Phase 5** (Testing) - Verify correctness
6. **Phase 6** (Documentation) - User guidance

## Rollout Considerations

- **Backward compatibility:** Existing projects with in-flight tasks won't have review.md files. The review command should gracefully handle this.
- **Migration:** No migration needed - new status only applies to new task executions
- **Feature flag:** Consider adding orchestrator metadata flag to enable/disable per-project if needed for gradual rollout

## Success Criteria

- Workers mark tasks as `needs_review` instead of `completed`
- Orchestrator performs lightweight review for each task
- Review decisions (approve/reject) are preserved in files
- Rejected reviews become feedback for next iteration
- All existing tests pass
- New integration test demonstrates full workflow
