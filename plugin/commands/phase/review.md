# /phase:review - Review Orchestration

**Purpose**: Validate implementation meets requirements
**Mode**: Orchestrator performs mandatory review, then gets human approval

---

## Role

You are the quality gate. **Review all implementation changes against original requirements, identify gaps, present findings to human.**

**Critical**: Review is mandatory - you always perform it, never skip.

---

## Workflow

### 1. Update Phase Status

Update `.sow/project/state.yaml`:
```yaml
phases:
  review:
    status: in_progress
    started_at: [timestamp]
```

Commit state change.

### 2. Perform Review

**Read original requirements**:
- Project description from state file
- Discovery artifacts (if any)
- Design documents (if any)

**Read ALL implementation changes**:
- Every file modified (check git diff)
- All task logs
- All commits made
- Test coverage and results

**Compare and validate**:
- Does implementation match requirements?
- Was original intent achieved?
- Are there gaps or deviations?
- Is quality acceptable?
- Are tests sufficient?

**Decide review approach**:

**Orchestrator performs review** (default):
- Small changes (1-5 files)
- Simple, straightforward work
- High confidence in assessment

**Spawn reviewer agent** (optional):
- Large changes (10+ files)
- Complex changes requiring deep analysis
- Uncertain about quality
- Human requests it

**If spawning reviewer**:
```
This is a substantial change. I'll spawn a reviewer agent to assist with thorough analysis.

Performing detailed review...
```

- Spawn reviewer via Task tool with context (requirements, implementation changes)
- Reviewer performs analysis
- Reviewer returns findings
- Orchestrator synthesizes findings

### 3. Create Review Report

Create `phases/review/reports/00[N]-review.md` (numbered by iteration):

**Report structure**:
```markdown
# Review Report 00[N]

**Date**: [timestamp]
**Reviewer**: [orchestrator/reviewer-agent]
**Implementation Phase**: Completed [timestamp]

## Original Intent

[Summary of what we were trying to achieve]

## Changes Made

[Summary of implementation changes]
- Files modified: [count]
- Tests added: [count]
- Commits: [count]

## Review Findings

### Requirements Met ✓
- [Requirement 1]: Met
- [Requirement 2]: Met

### Issues Identified ✗
- [Issue 1]: Description and severity
- [Issue 2]: Description and severity

### Recommendations
- [Recommendation 1]
- [Recommendation 2]

## Overall Assessment

[Pass/Fail with reasoning]

## Next Steps

[Approve and finalize, or loop back to implementation with specific tasks]
```

**Update state**:
```yaml
phases:
  review:
    reports:
      - id: "001"
        path: phases/review/reports/001-review.md
        created_at: [timestamp]
        assessment: [pass/fail]
```

### 4. Present Findings

```
Review complete. Here's my assessment:

[Brief summary of findings]

Overall: [Pass - ready to finalize / Fail - issues need addressing]

Review report: phases/review/reports/00[N]-review.md

Ready to proceed? [approve/feedback/loop-back]
```

**Options**:
- **approve**: Proceed to finalize phase
- **feedback**: Human provides additional feedback, update notes
- **loop-back**: Return to implementation to address issues

### 5. Handle Approval

**If approved**:
```yaml
phases:
  review:
    status: completed
    completed_at: [timestamp]
  finalize:
    status: pending
```

Commit state changes.

**Output**:
```
✓ Review phase complete
✓ Implementation approved

Starting finalize phase...
```

**Invoke**: `/phase:finalize`

### 6. Handle Loop-Back

**If loop-back requested**:

**Identify needed changes**:
- What specific issues need fixing?
- What tasks needed to address them?

**Create fail-forward tasks**:
```
Based on review findings, I propose adding these tasks:

050: [Task to address issue 1]
060: [Task to address issue 2]

These will be added to implementation phase. Approve? [yes/adjust]
```

**If approved**:

**Update state**:
```yaml
phases:
  review:
    status: in_progress  # Stays in_progress until next review
    iteration: [increment by 1]
  implementation:
    status: in_progress  # Reopened
    tasks:
      # Existing completed tasks stay completed
      - id: "050"
        name: [new task]
        status: pending
        # ...
```

**Add tasks to implementation** using same process as fail-forward.

Commit state changes.

**Output**:
```
Returning to implementation phase to address review findings.
Tasks added: [list]

Continuing implementation phase...
```

**Invoke**: `/phase:implementation`

**When implementation completes again**: Automatically return to review phase (this command), increment iteration, create new review report.

---

## Key Behaviors

**Mandatory review**:
- Never skip review
- Always perform assessment
- Create report every iteration

**Be thorough**:
- Review every file change
- Check against ALL requirements
- Identify gaps honestly
- Don't rubber-stamp

**Human gates transition**:
- Only proceed on explicit human approval
- Human can override your assessment (approve despite issues, or reject despite clean review)
- Respect human's decision

**Iterative loop-back**:
- Can return to implementation multiple times
- Each iteration gets new review report (002, 003, etc.)
- Previous reports remain for reference

**Reviewer agent optional**:
- Use judgment on when to spawn
- More eyes helpful for complex changes
- Not required for simple work

---

## Edge Cases

**No issues found but human wants changes**: "Review looks clean, but you'd like changes. What should we address?" → Loop back with new tasks

**Issues found but human approves anyway**: Accept decision, document in log, proceed to finalize

**Unclear if requirement met**: "Requirement [X] is ambiguous. Was [Y] the intended outcome?" → Clarify with human

**Test failures**: Major issue - "Tests are failing. Must return to implementation to fix. Approve loop-back?"

**Massive changes**: "Implementation touched [many] files. I'll use reviewer agent for thorough analysis."

---

## Example Flow

```
[/phase:review invoked]

[Reads requirements from discovery/design]
[Reads all implementation changes via git diff]
[Reviews 8 files modified, 12 tests added, 5 commits]

Implementation added JWT authentication as required. Checking coverage...

Found potential issue: Login endpoint missing error handling for expired tokens.

Creating review report...
[Creates phases/review/reports/001-review.md]
[Updates state with report, assessment: fail]

Review complete. Assessment:

Requirements mostly met:
✓ User model with validation
✓ JWT token generation
✓ Login endpoint functional
✓ Authentication middleware

Issue identified:
✗ Login endpoint doesn't handle expired token scenario

Overall: Fail - missing error handling

Review report: phases/review/reports/001-review.md

Ready to proceed? [loop-back]

[Human: loop-back]

Based on review, I propose adding:
- 050: Add error handling for expired tokens in login endpoint

Approve? [yes]

[Adds task 050 to implementation]
[Updates state: review iteration: 2, implementation reopened]

Returning to implementation phase to address review findings.
Tasks added: 050

Continuing implementation phase...
→ /phase:implementation

[Task 050 completes]
[Implementation phase completes]
[Automatic transition back to review]

[/phase:review invoked again]

[Reviews changes from task 050]
[Creates phases/review/reports/002-review.md]

Review complete (iteration 2). Assessment:

All requirements now met:
✓ Error handling added for expired tokens
✓ All previous requirements still met

Overall: Pass - ready to finalize

Review report: phases/review/reports/002-review.md

Ready to proceed? [approve]

[Human: approve]

✓ Review phase complete
✓ Implementation approved

Starting finalize phase...
→ /phase:finalize
```

---

## Notes

- **Review is mandatory**: Always happens, never skip
- **Human gates finalize**: Only proceed on explicit approval
- **Iterative by design**: Loop back as many times as needed
- **Each iteration tracked**: New report, incremented counter
- **Honest assessment**: Don't rubber-stamp, identify real issues
- **Automatic transitions**: Impl → Review (automatic), Review → Finalize (human approval)
