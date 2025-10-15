# /project:continue - Resume Existing Project

**Purpose**: Resume work on existing project
**Mode**: Orchestrator resumes coordination based on current state

---

## Workflow

### 1. Read State

Read `.sow/project/state.yaml`

**If not found**:
```
No active project on this branch.
Start new project? Use /project:new
```

### 2. Verify Branch

Check current branch matches `project.branch` in state

**If mismatch**:
```
⚠️  Branch mismatch detected:
   Current: [current-branch]
   Project: [project.branch]

Project '[project.name]' belongs to branch '[project.branch]'.
Switch branches? [yes/continue anyway/cancel]
```

- **yes**: `git checkout [project.branch]` then continue
- **continue anyway**: Proceed (warn about potential issues)
- **cancel**: Exit command

### 3. Show Status Summary

Display project overview:
```
Project: [project.name]
Branch: [project.branch]
Description: [description]

Phase Status:
✓ Discovery: [completed/skipped] [+ artifact count if completed]
✓ Design: [completed/skipped] [+ artifact count if completed]
⟳ Implementation: in_progress (3/5 tasks completed)
○ Review: pending
○ Finalize: pending
```

**Status indicators**:
- ✓ = completed/skipped
- ⟳ = in_progress
- ○ = pending
- ⚠ = has issues/needs attention

### 4. Identify Current Phase

**Logic**: Find first phase with `status: "in_progress"`, or first phase with `status: "pending"` if none in_progress

**Special case - Review loop-back**: If `review.status == "in_progress"` and last report has `assessment: "fail"`, project is actually in implementation phase (adding fail-forward tasks)

### 5. Resume Context

Provide brief context about where work left off:

**Discovery/Design**:
```
Resuming [phase] phase.
Artifacts: [count] created, [count] approved
Last updated: [relative time]
```

**Implementation**:
```
Resuming implementation phase.
Tasks: [completed]/[total] completed, [in_progress] in progress, [pending] pending
Last updated: [relative time]
```

**Review**:
```
Resuming review phase (iteration [iteration]).
Reports: [count] created
Last report: [assessment] ([relative time])
```

**Finalize**:
```
Resuming finalize phase.
Documentation: [status]
Project deleted: [yes/no]
```

### 6. Invoke Phase Command

**Determine command**:
- `discovery.status == "in_progress"` → `/phase:discovery`
- `design.status == "in_progress"` → `/phase:design`
- `implementation.status == "in_progress"` → `/phase:implementation`
- `review.status == "in_progress"` → `/phase:review`
- `finalize.status == "in_progress"` → `/phase:finalize`
- First pending phase → invoke that phase's command

**Transition message**:
```
Continuing [phase-name] phase...
```

**Invoke phase command**

---

## Edge Cases

**Multiple phases in_progress** (corrupted state): "State file corrupted - multiple phases in_progress. Which phase to resume? [discovery/design/implementation/review/finalize]"

**All phases completed**: "Project complete! All phases finished. Use /project:new for next project."

**Uncommitted changes in working directory**: Show warning but continue (orchestrator will handle commits during phase work)

---

## Examples

### Example 1: Resume Implementation
```
Project: add-authentication
Branch: feat/add-auth
Description: Add JWT-based authentication system

Phase Status:
✓ Discovery: completed (2 artifacts)
✓ Design: completed (2 artifacts)
⟳ Implementation: in_progress (3/5 tasks completed)
○ Review: pending
○ Finalize: pending

Resuming implementation phase.
Tasks: 3/5 completed, 1 in progress, 1 pending
Last updated: 2 hours ago

Continuing implementation phase...
→ /phase:implementation
```

### Example 2: Branch Mismatch
```
⚠️  Branch mismatch detected:
   Current: main
   Project: feat/add-auth

Project 'add-authentication' belongs to branch 'feat/add-auth'.
Switch branches? [yes/continue anyway/cancel]

[User: yes]

Switched to branch 'feat/add-auth'
[Shows status and resumes]
```

### Example 3: Review After Loop-Back
```
Project: refactor-auth
Branch: feat/refactor-auth

Phase Status:
✓ Discovery: completed (1 artifact)
○ Design: skipped
⟳ Implementation: in_progress (5/5 tasks completed)
⟳ Review: in_progress (iteration 2)
○ Finalize: pending

Resuming review phase (iteration 2).
Reports: 1 created
Last report: fail (30 minutes ago) - added task 050 via fail-forward

Continuing review phase...
→ /phase:review
```

---

## Notes

- **No approval needed**: Command reads state and resumes automatically
- **State is source of truth**: Current phase determined entirely from state file
- **Seamless resumption**: Phase commands handle their own context from state
- **Branch safety**: Warns on mismatch but allows override for flexibility
