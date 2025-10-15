# /project:new - Project Initialization

**Purpose**: Initialize new project using truth table logic to determine phase plan
**Mode**: Orchestrator in subservient mode (asks, recommends, human decides)

---

## Workflow

### 1. Gather Context

**Ask**: "What would you like to accomplish in this project?"

**Listen for**: Work type (bug/feature/refactor/new component), scope indicators, existing context mentions

**Ask**: "Do you have any existing context, documents, or notes about this?"

**If "yes"**: Ask "Are these detailed enough to start implementation, or would you like more discovery/design work first?"

### 2. Assess Project vs One-Off

**If trivial work suspected**: Invoke `/rubric:one-off`
- Score 3-4: Suggest one-off, respect user choice
- Score 0-2: Continue as project

**If one-off chosen**: Exit command, perform work directly (no project created)

### 3. Determine Discovery Phase

**Invoke**: `/rubric:discovery`

**Apply recommendation based on score**, user can override

**If discovery enabled**: Defer design question until discovery completes

### 4. Determine Design Phase

**Timing**:
- If discovery enabled: Ask later (end of discovery phase)
- If discovery disabled: Ask now

**Invoke**: `/rubric:design`

**Apply recommendation based on score**, user can override

### 5. Confirm Phase Plan

```
Based on our conversation, here's the plan:
- Discovery: [enabled/disabled] - [reason]
- Design: [enabled/disabled or TBD] - [reason]
- Implementation: enabled (required)
- Review: enabled (required)
- Finalize: enabled (required)

Does this work? [yes/modify]
```

**Must get explicit approval before proceeding**

### 6. Branch Management

**On main/master**:
```
On main branch. Creating feature branch.
Suggested: [feat|fix|refactor]/[description-kebab-case]
Okay? [yes/suggest different]
```

**On feature branch without project**: "Create project on '[branch]'? [yes/new branch]"

**On feature branch with project**:
```
Branch '[branch]' has active project: [name] ([phase] - [status])
Options: 1) Continue existing (/project:continue), 2) New branch
Choose: [continue/new branch]
```

**If continue**: Exit and invoke `/project:continue`

### 7. Project Name

Infer from description (kebab-case, 2-4 words): "Call this project '[name]'? [yes/suggest different]"

### 8. Initialize State

Create `.sow/project/state.yaml`:
```yaml
project:
  name: [project-name]
  branch: [branch-name]
  description: [user's original description]
  created_at: [ISO 8601]
  updated_at: [ISO 8601]

phases:
  discovery:
    enabled: [true/false]
    status: [pending/skipped]
    created_at: [ISO 8601]
    started_at: null
    completed_at: null
    artifacts: []

  design:
    enabled: [true/false]
    status: [pending/skipped]
    created_at: [ISO 8601]
    started_at: null
    completed_at: null
    artifacts: []

  implementation:
    enabled: true
    status: pending
    created_at: [ISO 8601]
    started_at: null
    completed_at: null
    tasks: []

  review:
    enabled: true
    status: pending
    created_at: [ISO 8601]
    started_at: null
    completed_at: null
    iteration: 1
    reports: []

  finalize:
    enabled: true
    status: pending
    created_at: [ISO 8601]
    started_at: null
    completed_at: null
    project_deleted: false
```

### 9. Create Directories

**Always**: `.sow/project/phases/{implementation,review,finalize}/`

**Conditional**: Add `discovery/` if enabled, `design/` if enabled

### 10. Commit

```bash
git add .sow/
git commit -m "chore: initialize sow project - [project-name]

Project: [project-name]
Phases: [enabled-phases-list]

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

### 11. Transition

**Determine first phase**: discovery.enabled ? `/phase:discovery` : design.enabled ? `/phase:design` : `/phase:implementation`

**Output**:
```
✓ Project initialized: [name]
✓ Branch: [branch]
✓ Phases: [enabled-list]

Starting [first-phase] phase...
```

**Invoke first phase command**

---

## Edge Cases

**Vague description**: "Could you provide more detail? What feature/problem/goal specifically?"

**Conflicting info** (has docs but sounds uncertain): "You mentioned docs, but sounds like unknowns. Review docs in discovery, or proceed as-is?"

**Branch name conflict**: "Branch '[name]' exists. Options: 1) '[name-2]', 2) Suggest different, 3) Use existing. Prefer?"

**Uncommitted changes**: "Uncommitted changes present. Options: 1) Stash and proceed, 2) Wait for commit, 3) Proceed anyway. Prefer?"

---

## Examples

### Example 1: Bug Fix (Skip Optional Phases)
```
Q: What to accomplish?
A: Fix login bug after password reset

Q: Existing context?
A: No, just noticed issue

[/rubric:discovery → score 4 → optional]
Q: Do discovery to investigate root cause, or dive into debugging?
A: Let's do discovery

[Design deferred or assessed as score 0 after bug penalty]

Confirm: Discovery=enabled, Design=disabled (bug fix), Impl/Review/Final=enabled
Branch: fix/login-after-reset
→ /phase:discovery
```

### Example 2: Existing Design (Skip All Optional)
```
Q: What to accomplish?
A: Implement auth system from docs/auth-design.md

Q: Existing context?
A: Yes, docs/auth-design.md has complete design

[/rubric:discovery → score 0 → skip]
[/rubric:design → score 0 → skip]

Confirm: Discovery=disabled (design exists), Design=disabled (exists), Impl/Review/Final=enabled
Branch: feat/implement-auth-system
→ /phase:implementation
```

### Example 3: Large Feature (All Phases, Design TBD)
```
Q: What to accomplish?
A: Add notification system with email, SMS, push

Q: Existing context?
A: No, just high-level idea

[/rubric:one-off → score 0 → project]
[/rubric:discovery → score 7 → recommend strongly]
Q: Recommend discovery to research approaches and requirements. Sound good?
A: Yes

[Design question deferred to end of discovery]

Confirm: Discovery=enabled, Design=TBD after discovery, Impl/Review/Final=enabled
Branch: feat/add-notification-system
→ /phase:discovery
```

---

## Notes

- **Subservient**: Orchestrator recommends, human decides
- **No assumptions**: Always ask unclear points
- **User overrides rubrics**: Always respect user choice
- **Explicit approval required**: Cannot proceed without confirmation of phase plan
- **Design timing**: If discovery enabled, ask about design later (not during init)
