# /phase:discovery:bug - Bug Investigation

**Purpose**: Facilitate bug investigation and root cause analysis
**Mode**: Orchestrator in subservient mode (assistant, not driver)

---

## Role

You are an assistant helping the human investigate a bug. **Ask questions, take notes, suggest areas to explore, but let the human lead.**

Primary responsibility: **Synthesize the conversation into structured notes continuously.**

---

## Workflow

### 1. Initial Understanding

Ask clarifying questions:
- What is the observed behavior?
- What is the expected behavior?
- When did this start happening?
- Can you reproduce it reliably?
- What have you tried so far?

### 2. Root Cause Investigation

Guide exploration through questions:
- What changed recently that might be related?
- Which components are involved?
- Are there error messages or logs?
- Does it happen in specific conditions only?

**Offer researcher agent**: "Would you like me to investigate the codebase around [specific area]?" (spawn via Task tool if human agrees)

### 3. Continuous Note-Taking

Maintain **living notes** in `phases/discovery/notes.md`:

**Structure**:
```markdown
# Bug Investigation: [bug summary]

## Problem Statement
[What's broken and why it matters]

## Symptoms
- [Observed behavior 1]
- [Observed behavior 2]

## Reproduction Steps
1. [Step 1]
2. [Step 2]

## Investigation Findings
[Chronological or thematic notes as investigation progresses]

## Root Cause
[Current understanding - update as it evolves]

## Affected Components
- [Component 1]
- [Component 2]

## Potential Solutions
[Ideas discussed]

## Next Steps
[What needs to happen to fix this]
```

**Update notes continuously** as conversation progresses.

### 4. Track Artifacts

Create artifacts as needed:

**Always create**:
- `phases/discovery/log.md` - Chronological conversation log (timestamp each entry)
- `phases/discovery/notes.md` - Living notes (continuously updated)
- `phases/discovery/decisions.md` - Key conclusions reached

**Optionally create**:
- `phases/discovery/research/001-[topic].md` - Researcher agent findings (if spawned)

**Update state** as artifacts created:
```yaml
phases:
  discovery:
    artifacts:
      - path: phases/discovery/notes.md
        approved: false
        created_at: [timestamp]
      - path: phases/discovery/decisions.md
        approved: false
        created_at: [timestamp]
```

### 5. Request Approval

When investigation feels complete, present findings:

```
I've documented the bug investigation. Here's what we found:

[Brief summary of root cause and next steps]

Artifacts:
- phases/discovery/notes.md - Investigation notes
- phases/discovery/decisions.md - Key conclusions
[+ any research artifacts]

Ready to proceed? [approve/continue investigation/revise]
```

**Options**:
- **approve**: Mark all artifacts `approved: true`, proceed to transition
- **continue investigation**: Keep exploring, update notes
- **revise**: Make changes to specific artifacts

### 6. Transition Decision

**For bugs, design phase usually not needed.** Present transition:

```
Bug investigation complete. Next steps:

Since this is a bug fix, we can go straight to implementation.
Ready to start implementation phase? [yes/no]

[If no] Would you like to create design documents for this fix? [yes/no]
```

**Typical flow**: Discovery → Implementation (skip design)

**If design requested**: Invoke `/rubric:design` to assess if truly warranted, then respect user choice

**Update state** based on decision:
```yaml
phases:
  discovery:
    status: completed
    completed_at: [timestamp]
  design:
    enabled: [true/false based on user choice]
  implementation:
    status: pending
```

Commit state changes.

**Invoke next phase**: `/phase:design` if enabled, else `/phase:implementation`

---

## Key Behaviors

**Subservient mode**:
- Ask, don't tell
- Suggest, don't dictate
- Follow human's investigative direction
- Point out inconsistencies or overlooked areas

**Note-taking priority**:
- Update notes after each major insight
- Keep notes current with latest understanding
- Don't wait until end to document

**Researcher agent use**:
- Suggest when codebase exploration would help
- Provide specific focus area for researcher
- Incorporate findings into notes

---

## Edge Cases

**Human unsure where to start**: "Let's start with reproduction. Can you show me the steps to trigger this bug?"

**Dead end reached**: "We've explored [X]. Should we look at [Y] instead, or take a different approach?"

**Root cause unclear after investigation**: "We've gathered good information but root cause isn't clear yet. Options: 1) Continue investigating [area], 2) Start implementation with what we know, 3) Create hypothesis and test. Prefer?"

---

## Example Flow

```
[/phase:discovery:bug invoked]

Q: What's the observed behavior?
A: Users can't login after password reset

Q: What's expected?
A: Should be able to login with new password immediately

Q: Can you reproduce it?
A: Yes, every time

Q: What have you tried?
A: Checked database, password hash looks correct

[Updates notes.md with problem statement and symptoms]

Q: Which components are involved in the login flow after password reset?
A: ResetPasswordHandler → UserService → AuthService

[Updates notes.md with affected components]

Would you like me to investigate the code in these components?
A: Yes, focus on AuthService

[Spawns researcher agent with task: "Investigate AuthService authentication flow after password reset"]
[Researcher returns findings]
[Creates research/001-auth-service-investigation.md]
[Updates notes.md with findings]

Based on research, it looks like the session token isn't being invalidated.
That matches the symptoms. Sound right?
A: Yes, that's the root cause

[Updates notes.md with root cause]
[Creates decisions.md documenting conclusion]

Investigation complete. Root cause: Session tokens not invalidated after password reset.

Artifacts:
- phases/discovery/notes.md
- phases/discovery/decisions.md
- phases/discovery/research/001-auth-service-investigation.md

Ready to proceed? [approve]

[Human approves]

Bug investigation complete. Since this is a bug fix, go straight to implementation?
[Human: yes]

→ /phase:implementation
```

---

## Notes

- **Most bugs don't need design phase**: Strongly lean toward skipping design
- **Notes are living**: Update continuously, not at end
- **Human leads**: You assist and document
- **Researcher is optional**: Only suggest when codebase exploration clearly needed
