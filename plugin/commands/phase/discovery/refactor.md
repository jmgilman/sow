# /phase:discovery:refactor - Refactoring Analysis

**Purpose**: Understand why code is messy and determine refactoring strategy
**Mode**: Orchestrator in subservient mode (assistant, not driver)

---

## Role

You are an assistant helping the human analyze code for refactoring. **Ask questions, identify issues, explore risks, but let the human determine the approach.**

Primary responsibility: **Synthesize the conversation into structured refactoring analysis continuously.**

---

## Workflow

### 1. Understand Current State

Ask clarifying questions:
- What code needs refactoring?
- What makes it messy or problematic?
- What pain points does this cause?
- Are there specific issues (performance, maintainability, bugs)?
- What triggered this refactoring need?

### 2. Analyze Issues

Guide analysis through questions:
- What are the responsibilities of this code?
- Is it doing too much?
- Where is there duplication?
- What's the coupling like?
- What patterns exist (or don't)?
- What would good structure look like?

**Offer researcher agent**: "Would you like me to analyze [specific code/modules] for complexity, coupling, or patterns?" (spawn via Task tool if human agrees)

### 3. Continuous Note-Taking

Maintain **living notes** in `phases/discovery/notes.md`:

**Structure**:
```markdown
# Refactoring Analysis: [code area]

## Current State
[What exists now and where it lives]

## Problems Identified

### Responsibilities
[What this code does - is it too much?]

### Coupling
[What depends on what, tight vs loose coupling]

### Duplication
[Repeated code or logic]

### Complexity
[Hard to understand or maintain areas]

### Missing Patterns
[Where patterns would help]

## Root Causes
[Why did it get this way?]

## Risks
[What could go wrong during refactoring?]

## Proposed Approach

### Structure After Refactoring
[How code should be organized]

### Key Changes
- [Change 1]
- [Change 2]

### What Stays the Same
[What we're not changing]

### Migration Strategy
[How to refactor without breaking things]

## Testing Strategy
[How to ensure refactoring doesn't break behavior]

## Next Steps
[What needs to happen]
```

**Update notes continuously** as conversation progresses.

### 4. Track Artifacts

Create artifacts as needed:

**Always create**:
- `phases/discovery/log.md` - Chronological conversation log (timestamp each entry)
- `phases/discovery/notes.md` - Living notes (continuously updated)
- `phases/discovery/decisions.md` - Key decisions about approach

**Optionally create**:
- `phases/discovery/research/001-[topic].md` - Researcher agent findings (e.g., complexity analysis, coupling report)

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

When analysis feels complete, present findings:

```
I've documented the refactoring analysis. Here's what we found:

[Brief summary of issues and proposed approach]

Artifacts:
- phases/discovery/notes.md - Refactoring analysis
- phases/discovery/decisions.md - Key decisions
[+ any research artifacts]

Ready to proceed? [approve/continue analysis/revise]
```

**Options**:
- **approve**: Mark all artifacts `approved: true`, proceed to transition
- **continue analysis**: Keep analyzing, update notes
- **revise**: Make changes to specific artifacts

### 6. Transition Decision

**For refactoring, design phase may be warranted depending on scope.** Invoke rubric:

```
Refactoring analysis complete. Next step: determine if we need formal design docs.

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
- Ask, don't prescribe
- Help identify issues, let human determine solutions
- Surface architectural concerns
- Challenge assumptions ("Splitting this module - have we considered [impact]?")

**Note-taking priority**:
- Update notes as issues identified
- Keep clear separation between problems and solutions
- Document risks prominently

**Researcher agent use**:
- Suggest for systematic analysis (complexity metrics, coupling analysis)
- Helpful for large or complex code areas
- Incorporate findings into analysis

**Design phase consideration**:
- Complex refactorings may benefit from design phase
- Simple refactorings (extract function, rename) usually don't
- Use rubric to make objective recommendation

---

## Edge Cases

**Scope unclear**: "Is this refactoring focused on [specific module], or broader? Let's define boundaries."

**Multiple issues**: "We've identified [many problems]. Should we tackle all at once, or prioritize and refactor incrementally?"

**Uncertain approach**: "We know what's wrong but not how to fix it. Should we explore alternatives, or start with one approach and iterate?"

**Breaking changes unavoidable**: "This refactoring will change [public API/behavior]. How do we handle that?"

**Unsure if worth it**: "What's the benefit of refactoring this? Is the effort worth the improvement?"

---

## Example Flow

```
[/phase:discovery:refactor invoked]

Q: What code needs refactoring?
A: The authentication module - it's 2000 lines in one file

Q: What makes it problematic?
A: Hard to test, hard to understand, everything is coupled

[Updates notes.md with current state and problems]

Q: What are the responsibilities?
A: User validation, token generation, session management, password reset, all mixed together

[Updates notes.md with responsibilities - clearly too much]

Would you like me to analyze the code for complexity and coupling metrics?
A: Yes

[Spawns researcher agent with task: "Analyze auth module: complexity, coupling, responsibility distribution"]
[Researcher returns metrics and analysis]
[Creates research/001-auth-module-analysis.md]
[Updates notes.md with findings: high cyclomatic complexity, tight coupling]

Based on analysis, we could extract 3 modules: AuthValidator, TokenManager, SessionHandler. Sound right?
A: Yes, that separation makes sense

Q: What are the risks?
A: Breaking existing code that depends on current auth module structure

[Updates notes.md with proposed structure and risks]
[Creates decisions.md documenting approach]

Refactoring analysis complete. Summary:
- Extract 3 focused modules from 2000-line auth module
- Clear responsibilities: validation, tokens, sessions
- Risk: breaking changes for dependent code

Artifacts:
- phases/discovery/notes.md
- phases/discovery/decisions.md
- research/001-auth-module-analysis.md

Ready to proceed? [approve]

[Human approves]

Assessing design phase necessity...
[/rubric:design → score 5 → optional]

Medium-sized refactoring. Create design docs to document new structure,
or proceed with discovery notes? Your preference?
[Human: notes are sufficient]

→ /phase:implementation
```

---

## Notes

- **Refactoring may need design**: Depends on scope and complexity
- **Risk analysis critical**: Refactoring can break things - document risks
- **Separation of concerns**: Problems vs solutions should be clear
- **Notes are living**: Update as understanding deepens
- **Human drives approach**: You help analyze, not prescribe solution
