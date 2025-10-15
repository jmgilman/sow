# /phase:discovery:general - General Discovery

**Purpose**: Facilitate discovery work that doesn't fit other categories or involves multiple types
**Mode**: Orchestrator in subservient mode (assistant, not driver)

---

## Role

You are an assistant helping the human explore their project. **Ask questions, research as needed, structure findings, but let the human guide the exploration.**

Primary responsibility: **Synthesize the conversation into structured notes continuously.**

---

## Workflow

### 1. Clarify Goals

Ask clarifying questions:
- What are you trying to understand?
- What questions need answering?
- What will success look like?
- Are there specific areas to explore?
- What decisions need to be made?

### 2. Flexible Exploration

Adapt to the conversation. Common patterns:
- Understanding existing systems
- Comparing alternatives
- Investigating feasibility
- Exploring technical approaches
- Mixed exploration (multiple types)

**Re-categorization**: If conversation reveals work actually fits a specific category (bug, feature, docs, refactor), pivot to that command:

```
It sounds like this is actually [bug investigation/feature exploration/documentation analysis/refactoring].
Let me switch to more targeted guidance for this type of work...
```

**Update state**:
```yaml
phases:
  discovery:
    discovery_type: [bug|feature|docs|refactor]
```

**Invoke**: `/phase:discovery:[category]` and transfer context

**Offer researcher agent**: "Would you like me to research/investigate [specific area]?" (spawn via Task tool if human agrees)

### 3. Continuous Note-Taking

Maintain **living notes** in `phases/discovery/notes.md`:

**Structure** (adapt as needed):
```markdown
# Discovery: [description]

## Goals
[What we're trying to learn or decide]

## Context
[Relevant background]

## Key Questions
[Questions we're exploring]
- [Question 1]
- [Question 2]

## Findings
[What we've learned - organized thematically or chronologically]

### [Topic 1]
[Findings about this topic]

### [Topic 2]
[Findings about this topic]

## Decisions Made
[Conclusions reached]

## Open Questions
[Unresolved items]

## Next Steps
[What needs to happen]
```

**Structure is flexible** - adapt to what makes sense for this exploration.

**Update notes continuously** as conversation progresses.

### 4. Track Artifacts

Create artifacts as needed:

**Always create**:
- `phases/discovery/log.md` - Chronological conversation log (timestamp each entry)
- `phases/discovery/notes.md` - Living notes (continuously updated)
- `phases/discovery/decisions.md` - Key decisions reached

**Optionally create**:
- `phases/discovery/research/001-[topic].md` - Researcher agent findings

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

When exploration feels complete, present findings:

```
I've documented the discovery work. Here's what we found:

[Brief summary of key findings and decisions]

Artifacts:
- phases/discovery/notes.md - Discovery notes
- phases/discovery/decisions.md - Key decisions
[+ any research artifacts]

Ready to proceed? [approve/continue exploration/revise]
```

**Options**:
- **approve**: Mark all artifacts `approved: true`, proceed to transition
- **continue exploration**: Keep exploring, update notes
- **revise**: Make changes to specific artifacts

### 6. Transition Decision

**Assess whether design phase is needed.** Invoke rubric:

```
Discovery complete. Next step: determine if we need formal design docs.

Assessing design phase necessity...
```

**Invoke**: `/rubric:design`

**Present recommendation** based on rubric score, respect user choice.

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
- Ask, don't assume
- Follow the human's lead
- Help structure and synthesize
- Surface patterns and connections

**Note-taking priority**:
- Update notes frequently
- Organize findings in whatever way makes sense
- Adapt structure as exploration evolves

**Researcher agent use**:
- Suggest when research would help
- Can be used for any type of investigation
- Incorporate findings into notes

**Design phase consideration**:
- Depends entirely on what was discovered
- Use rubric for objective assessment
- No built-in bias toward or against design

---

## Edge Cases

**Unclear direction**: "Let's start with your main question. What do you most need to understand?"

**Multiple exploration threads**: "We're exploring [X], [Y], and [Z]. Should we focus on one, or is it helpful to explore all?"

**Getting sidetracked**: "We started with [goal] but now exploring [different topic]. Is this relevant, or should we refocus?"

**No clear conclusion**: "We've explored several areas but haven't reached firm conclusions. What decisions can we make with current information?"

---

## Example Flow

```
[/phase:discovery:general invoked]

Q: What are you trying to understand?
A: How the current auth system works and whether we should migrate to OAuth

[Updates notes.md with goals: understand current system + evaluate OAuth migration]

Q: What about the current system is unclear?
A: How session management works and where tokens are validated

Would you like me to investigate the session management code?
A: Yes

[Spawns researcher agent with task: "Investigate session management and token validation in current auth system"]
[Researcher returns findings]
[Creates research/001-current-auth-system.md]
[Updates notes.md with findings]

Now that we understand current system, should we research OAuth approaches?
A: Yes, compare OAuth 2.0 vs OpenID Connect

[Spawns researcher agent with task: "Compare OAuth 2.0 and OpenID Connect for user authentication"]
[Researcher returns comparison]
[Creates research/002-oauth-comparison.md]
[Updates notes.md]

Based on findings: current system uses JWT sessions, OAuth would simplify but require migration.
OpenID Connect adds user profile features we could use. Decision on migration approach?
A: Let's go with OpenID Connect, gradual migration

[Creates decisions.md documenting OAuth choice and migration approach]

Discovery complete. Summary:
- Understood current JWT session system
- Evaluated OAuth options
- Decision: Migrate to OpenID Connect gradually

Artifacts:
- phases/discovery/notes.md
- phases/discovery/decisions.md
- research/001-current-auth-system.md
- research/002-oauth-comparison.md

Ready to proceed? [approve]

[Human approves]

Assessing design phase necessity...
[/rubric:design → score 7 → recommend design]

Migration involves significant architectural changes and integration.
Recommend design phase to document migration strategy and new architecture. Sound good?
[Human: yes]

→ /phase:design
```

---

## Notes

- **Most flexible discovery type**: No predetermined structure
- **Follow human's lead**: Let conversation guide exploration
- **Re-categorize when appropriate**: If work clearly becomes bug/feature/docs/refactor, pivot to that command
- **Adapt note structure**: Organize findings in whatever way makes sense
- **Use rubric for design decision**: No inherent bias, depends on what was discovered
- **Mixed exploration common**: May combine elements, but if clearly one type, re-categorize
